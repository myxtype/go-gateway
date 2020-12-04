package gateway

import (
	"encoding/json"
	"errors"
	"github.com/myxtype/go-gateway/client"
	"github.com/myxtype/go-gateway/pkg/timer"
	"github.com/myxtype/go-gateway/protocol"
	"github.com/myxtype/go-gateway/worker"
	"log"
	"net"
	"sync"
	"time"
)

type Gateway struct {
	c                 *GatewayConfig
	clientConnections sync.Map
	uidConnections    sync.Map
	groupConnections  sync.Map
	workerConnections sync.Map
}

type GatewayConfig struct {
	Addr            string        // 对外Gateway监听地址
	InnerAddr       string        // 内部Gateway监听地址
	RegisterAddress string        // 注册服务地址
	PingInterval    time.Duration // 心跳频率
	Certificate     string        // 连接凭证，为空表示无凭证
}

func NewGateway(conf *GatewayConfig) *Gateway {
	return &Gateway{
		c: conf,
	}
}

func (g *Gateway) Start() error {
	w := worker.NewWorker(g.c.Addr, g)

	if err := w.Start(); err != nil {
		return err
	}
	return nil
}

func (r *Gateway) OnWorkerStart() {
	// 客户端心跳
	go timer.NewTimer(25*time.Second, r.ping).Start()
	// business心跳
	go timer.NewTimer(25*time.Second, r.pingBusinessWorker).Start()

	// 内部Worker通信
	inner := worker.NewWorker(r.c.InnerAddr, &worker.WorkerEventProxy{
		ProxyOnConnect: r.onWorkerConnect,
		ProxyOnMessage: r.onWorkerMessage,
		ProxyOnClose:   r.onWorkerClose,
	})
	go func() {
		if err := inner.Start(); err != nil {
			log.Panic(err)
		}
	}()

	// 注册地址
	go r.registerAddress()
}

func (g *Gateway) OnWorkerStop() {

}

// 客户端连接
func (g *Gateway) OnConnect(conn *worker.Connection) {
	g.clientConnections.Store(conn.Id(), conn)

	g.sendToWorker(protocol.CMD_ON_CONNECT, conn, nil)
}

// 客户端消息
func (g *Gateway) OnMessage(conn *worker.Connection, message []byte) {
	g.sendToWorker(protocol.CMD_ON_MESSAGE, conn, message)
}

// 客户端断开连接
func (g *Gateway) OnClose(conn *worker.Connection) {
	g.sendToWorker(protocol.CMD_ON_CLOSE, conn, nil)

	g.clientConnections.Delete(conn.Id())

	// 清理 uid
	if v, found := conn.Payload.Load("uid"); found {
		uid := v.(string)
		if vv, found := g.uidConnections.Load(uid); found {
			uidPools := vv.(*MapString)
			uidPools.Delete(conn.Id())

			if uidPools.Length() == 0 {
				g.uidConnections.Delete(uid)
			}
		}
	}

	// 清理 group
	if v, found := conn.Payload.Load("groups"); found {
		groups := v.([]string)
		for _, groupName := range groups {
			if vv, found := g.groupConnections.Load(groupName); found {
				group := vv.(*MapString)
				group.Delete(conn.Id())

				if group.Length() == 0 {
					g.groupConnections.Delete(groupName)
				}
			}
		}
	}
}

// 向注册中心注册Gateway的地址
func (g *Gateway) registerAddress() {
	c := client.NewAsyncTcpConnection(g.c.RegisterAddress)

	var ping *timer.Timer

	c.OnConnect = func(conn *net.TCPConn) {
		buffer := append((&RegisterMessage{
			Event:       "gateway_connect",
			Certificate: g.c.Certificate,
			Address:     g.c.InnerAddr,
		}).Bytes(), '\n')

		if _, err := conn.Write(buffer); err != nil {
			log.Println(err)
		}

		ping = timer.NewTimer(time.Second*25, func() {
			conn.Write([]byte("{\"event\":\"ping\"}\n"))
		})
		go ping.Start()
	}

	c.OnClose = func(conn *net.TCPConn) {
		if ping != nil {
			ping.Stop()
		}
		// todo
	}

	if err := c.Connect(); err != nil {
		log.Panic(err)
	}
}

// onWorkerConnect business连接
func (g *Gateway) onWorkerConnect(conn *worker.Connection) {

}

// onWorkerClose business断开
func (g *Gateway) onWorkerClose(conn *worker.Connection) {
	g.workerConnections.Delete(conn.Id())
}

// onWorkerMessage business消息
func (g *Gateway) onWorkerMessage(conn *worker.Connection, data []byte) {
	var msg BusinessMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Println(err)
		return
	}

	// 判断否是否认证
	if _, found := conn.Payload.Load("authorized"); !found {
		if msg.Cmd != protocol.CMD_WORKER_CONNECT && msg.Cmd != protocol.CMD_GATEWAY_CLIENT_CONNECT {
			log.Printf("Unauthorized request from %v \n", conn.RemoteAddr().String())
			conn.Close()
			return
		}
	}

	switch msg.Cmd {
	case protocol.CMD_WORKER_CONNECT: // BusinessWorker连接Gateway
		var certificate string
		if err := msg.UnmarshalBody(&certificate); err == nil {
			if certificate != g.c.Certificate {
				log.Printf("Gateway: Worker key does not match %v != %v \n", certificate, g.c.Certificate)
				conn.Close()
				return
			}
			g.workerConnections.Store(conn.Id(), conn)
			conn.Payload.Store("authorized", true)
		}

	case protocol.CMD_GATEWAY_CLIENT_CONNECT: // GatewayClient连接Gateway
		var certificate string
		if err := msg.UnmarshalBody(&certificate); err == nil {
			if certificate != g.c.Certificate {
				log.Printf("Gateway: GatewayClient key does not match %v != %v \n", certificate, g.c.Certificate)
				conn.Close()
				return
			}
			conn.Payload.Store("authorized", true)
		}

	case protocol.CMD_SEND_TO_ONE: // 向某客户端发送数据，Gateway::sendToClient($client_id, $message);
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			v.(*worker.Connection).Write(append(msg.Body, '\n'))
		}

	default:
		log.Printf("Gateway inner pack err cmd=%v \n", msg.Cmd)
	}
}

// sendToWorker 向BusinessWorker发送指令
func (g *Gateway) sendToWorker(cmd protocol.Protocol, conn *worker.Connection, data []byte) {
	workerConn, err := g.router()
	if err != nil {
		log.Println(err)
		return
	}

	var session = map[string]interface{}{}
	if v, found := conn.Payload.Load("session"); found {
		session = v.(map[string]interface{})
	}
	b, _ := json.Marshal(session)
	msg := &BusinessMessage{
		Cmd:     cmd,
		Body:    data,
		ExtData: b,
	}

	if err := workerConn.Send(msg.Bytes()); err != nil {
		log.Println(err)
	}
}

// router 选择一个BusinessWorker
func (g *Gateway) router() (*worker.Connection, error) {
	var conns []*worker.Connection
	g.workerConnections.Range(func(key, value interface{}) bool {
		conns = append(conns, value.(*worker.Connection))
		return true
	})
	if len(conns) == 0 {
		return nil, errors.New("no business worker online")
	}
	return conns[0], nil
}

// pingBusinessWorker 向 BusinessWorker 发送心跳数据，用于保持长连接
func (g *Gateway) pingBusinessWorker() {
	msg := (&BusinessMessage{
		Cmd: protocol.CMD_PING,
	}).Bytes()
	g.workerConnections.Range(func(key, value interface{}) bool {
		if err := value.(*worker.Connection).Send(msg); err != nil {
			log.Println(err)
		}
		return true
	})
}

// ping 检查客户端的心跳
func (g *Gateway) ping() {
	g.clientConnections.Range(func(key, value interface{}) bool {
		return true
	})
}
