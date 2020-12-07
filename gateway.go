package gateway

import (
	"encoding/json"
	"errors"
	"github.com/myxtype/go-gateway/client"
	"github.com/myxtype/go-gateway/pkg/logger"
	"github.com/myxtype/go-gateway/pkg/timer"
	"github.com/myxtype/go-gateway/protocol"
	"github.com/myxtype/go-gateway/worker"
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
	w := worker.NewWorker(g, &worker.WorkerConfig{Addr: g.c.Addr})

	if err := w.Start(); err != nil {
		return err
	}
	return nil
}

func (g *Gateway) OnWorkerStart() {
	// 客户端心跳
	go timer.NewTimer(g.c.PingInterval, g.ping).Start()
	// business心跳
	go timer.NewTimer(g.c.PingInterval, g.pingBusinessWorker).Start()

	// 内部Worker通信
	inner := worker.NewWorker(&worker.WorkerEventProxy{
		ProxyOnConnect: g.onWorkerConnect,
		ProxyOnMessage: g.onWorkerMessage,
		ProxyOnClose:   g.onWorkerClose,
	}, &worker.WorkerConfig{Addr: g.c.InnerAddr})
	go func() {
		if err := inner.Start(); err != nil {
			logger.Sugar.Panic(err)
		}
	}()

	// 注册地址
	go g.registerAddress()
}

func (g *Gateway) OnWorkerStop() {

}

// 客户端连接
func (g *Gateway) OnConnect(conn *worker.Connection) {
	g.clientConnections.Store(conn.Id(), conn)

	// 初始化客户端信息
	remote := map[string]interface{}{
		"gatewayAddr":      g.c.Addr,
		"gatewayInnerAddr": g.c.InnerAddr,
		"clientAddr":       conn.RemoteAddr().String(),
	}
	conn.Payload.Store("remote", remote)

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
		if v, found = g.uidConnections.Load(uid); found {
			uidPools := v.(*MapString)
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
			if v, found = g.groupConnections.Load(groupName); found {
				group := v.(*MapString)
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

	c.OnConnect = func(conn *client.AsyncTcpConnection) {
		buffer := (&RegisterMessage{
			Event:       "gateway_connect",
			Certificate: g.c.Certificate,
			Address:     g.c.InnerAddr,
		}).Bytes()

		if err := conn.Send(buffer); err != nil {
			logger.Sugar.Panic(err)
		}

		logger.Sugar.Infof("register %v connected", g.c.RegisterAddress)

		ping = timer.NewTimer(g.c.PingInterval, func() {
			conn.Send(PingData)
		})
		go ping.Start()
	}

	c.OnClose = func(conn *client.AsyncTcpConnection) {
		if ping != nil {
			ping.Stop()
		}
		logger.Sugar.Infof("register %v disconnected", g.c.RegisterAddress)
		// 重新连接
		time.AfterFunc(2*time.Second, func() {
			for {
				logger.Sugar.Infof("register %v attempting to reconnect", g.c.RegisterAddress)

				if err := c.Connect(); err != nil {
					logger.Sugar.Error(err)
					time.Sleep(2 * time.Second)
					continue
				}

				break
			}
		})
	}

	if err := c.Connect(); err != nil {
		logger.Sugar.Panic(err)
	}
}

// onWorkerConnect business连接
func (g *Gateway) onWorkerConnect(conn *worker.Connection) {
	logger.Sugar.Infof("worker %v connected", conn.Id())
}

// onWorkerClose business断开
func (g *Gateway) onWorkerClose(conn *worker.Connection) {
	g.workerConnections.Delete(conn.Id())

	logger.Sugar.Infof("worker %v disconnected", conn.Id())
}

// onWorkerMessage business消息
func (g *Gateway) onWorkerMessage(conn *worker.Connection, data []byte) {
	var msg GatewayMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Sugar.Errorf("invalid message body %v, error: %v", conn.Id(), string(data), err.Error())
		return
	}

	// 判断否是否认证
	if _, found := conn.Payload.Load("authorized"); !found {
		if msg.Cmd != protocol.CMD_WORKER_CONNECT && msg.Cmd != protocol.CMD_GATEWAY_CLIENT_CONNECT {
			logger.Sugar.Infof("unauthorized request from %v", conn.RemoteAddr().String())
			conn.Close()
			return
		}
	}

	switch msg.Cmd {
	case protocol.CMD_WORKER_CONNECT:
		if string(msg.Body) != g.c.Certificate {
			logger.Sugar.Infof("worker key does not match %s != %v", msg.Body, g.c.Certificate)
			conn.Close()
			return
		}
		logger.Sugar.Infof("worker %v authorized", conn.Id())
		g.workerConnections.Store(conn.Id(), conn)
		conn.Payload.Store("authorized", true)

	case protocol.CMD_GATEWAY_CLIENT_CONNECT:
		var certificate string
		if err := msg.UnmarshalBody(&certificate); err == nil {
			if certificate != g.c.Certificate {
				logger.Sugar.Infof("gateway client key does not match %v != %v", certificate, g.c.Certificate)
				conn.Close()
				return
			}
			conn.Payload.Store("authorized", true)
		}

	case protocol.CMD_SEND_TO_ONE: // 给一个客户端发送数据
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			v.(*worker.Connection).Send(msg.Body)
		}

	case protocol.CMD_KICK: // 踢出
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			v.(*worker.Connection).Close()
		}

	case protocol.CMD_DESTROY: // 销毁
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			v.(*worker.Connection).Close()
		}

	case protocol.CMD_SEND_TO_ALL: // 广播
		var connIds []string
		if err := msg.UnmarshalExtData(&connIds); err != nil {
			logger.Sugar.Error(err)
			return
		}

		if len(connIds) > 0 {
			for _, connId := range connIds {
				if v, found := g.clientConnections.Load(connId); found {
					v.(*worker.Connection).Send(msg.Body)
				}
			}
		} else {
			g.clientConnections.Range(func(key, v interface{}) bool {
				v.(*worker.Connection).Send(msg.Body)
				return true
			})
		}

	case protocol.CMD_GET_GROUP_ID_LIST: // 获取在线群组列表
		var groups []string
		g.groupConnections.Range(func(key, value interface{}) bool {
			groups = append(groups, key.(string))
			return true
		})
		if err := conn.SendJson(groups); err != nil {
			logger.Sugar.Error(err)
		}

	case protocol.CMD_SET_SESSION: // 设置SESSION
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			v.(*worker.Connection).Payload.Store("session", msg.Session)
		}

	case protocol.CMD_UPDATE_SESSION: // 合并SESSION
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			c := v.(*worker.Connection)
			var old = map[string]interface{}{}
			if v, found = c.Payload.Load("session"); found {
				old = v.(map[string]interface{})
			}
			for key, val := range msg.Session {
				old[key] = val
			}
			c.Payload.Store("session", old)
		}

	case protocol.CMD_GET_SESSION_BY_CLIENT_ID: // 通过ClientId获取Session
		var session = map[string]interface{}{}
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			if v, found = v.(*worker.Connection).Payload.Load("session"); found {
				session = v.(map[string]interface{})
			}
		}
		if err := conn.SendJson(session); err != nil {
			logger.Sugar.Error(err)
		}

	case protocol.CMD_GET_ALL_CLIENT_SESSIONS: // 获取所有Session
		var sessions = map[string]map[string]interface{}{}
		g.clientConnections.Range(func(key, value interface{}) bool {
			if v, found := value.(*worker.Connection).Payload.Load("session"); found {
				sessions[key.(string)] = v.(map[string]interface{})
			}
			return true
		})
		if err := conn.SendJson(sessions); err != nil {
			logger.Sugar.Error(err)
		}

	case protocol.CMD_IS_ONLINE: // 判断是否在线
		var isOnline = "0"
		if _, found := g.clientConnections.Load(msg.ConnId); found {
			isOnline = "1"
		}
		conn.Send([]byte(isOnline))

	case protocol.CMD_BIND_UID: // 将clientId与Uid进行绑定
		var uid string
		if err := msg.UnmarshalExtData(&uid); err != nil {
			logger.Sugar.Error(err)
			return
		}
		if v, found := g.clientConnections.Load(msg.ConnId); found {
			c := v.(*worker.Connection)

			// 此连接已绑定过UID
			if v, found = c.Payload.Load("uid"); found {
				if v, found = g.uidConnections.Load(v.(string)); found {
					uidPools := v.(*MapString)
					uidPools.Delete(msg.ConnId)
					if uidPools.Length() == 0 {
						g.uidConnections.Delete(v.(string))
					}
				}
			}

			var uidPools = &MapString{}
			if v, found = g.uidConnections.Load(uid); found {
				uidPools = v.(*MapString)
			} else {
				g.uidConnections.Store(uid, uidPools)
			}

			c.Payload.Store("uid", uid)
			uidPools.Load(msg.ConnId)
		}

	case protocol.CMD_UNBIND_UID: // 将clientId与Uid解除绑定
		v, found := g.clientConnections.Load(msg.ConnId)
		if !found {
			return
		}
		c := v.(*worker.Connection)
		v, found = c.Payload.Load("uid")
		if !found {
			return
		}
		uid := v.(string)
		if v, found = g.uidConnections.Load(uid); found {
			uidPools := v.(*MapString)
			uidPools.Delete(msg.ConnId)
			if uidPools.Length() == 0 {
				g.uidConnections.Delete(uid)
			}
		}
		c.Payload.Delete("uid")
	default:
		logger.Sugar.Infof("inner pack err cmd=%v", msg.Cmd)
	}
}

// sendToWorker 向BusinessWorker发送指令
func (g *Gateway) sendToWorker(cmd protocol.Protocol, conn *worker.Connection, data []byte) {
	workerConn, err := g.router()
	if err != nil {
		logger.Sugar.Info(err)
		return
	}

	msg := &GatewayMessage{
		Cmd:  cmd,
		Body: data,
	}
	if v, found := conn.Payload.Load("session"); found {
		msg.Session = v.(map[string]interface{})
	}
	if v, found := conn.Payload.Load("remote"); found {
		msg.Remote = v.(map[string]interface{})
	}

	if err := workerConn.Send(msg.Bytes()); err != nil {
		logger.Sugar.Info(err)
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
		return nil, errors.New("no pkg worker online")
	}
	return conns[0], nil
}

// pingBusinessWorker 向 BusinessWorker 发送心跳数据，用于保持长连接
func (g *Gateway) pingBusinessWorker() {
	msg := (&GatewayMessage{
		Cmd: protocol.CMD_PING,
	}).Bytes()
	g.workerConnections.Range(func(key, value interface{}) bool {
		if err := value.(*worker.Connection).Send(msg); err != nil {
			logger.Sugar.Info(err)
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
