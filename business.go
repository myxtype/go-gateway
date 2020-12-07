package gateway

import (
	"encoding/json"
	"github.com/myxtype/go-gateway/client"
	"github.com/myxtype/go-gateway/pkg/logger"
	"github.com/myxtype/go-gateway/pkg/timer"
	"github.com/myxtype/go-gateway/protocol"
	"sync"
	"time"
)

type Business struct {
	c       *BusinessConfig
	handler BusinessEventsInterface

	gatewayConnections             sync.Map
	gatewayAddresses               map[string]struct{}
	connectingGatewayAddresses     map[string]struct{}
	waitingConnectGatewayAddresses map[string]struct{}
}

type BusinessConfig struct {
	RegisterAddress string        // 注册中心地址
	PingInterval    time.Duration // 心跳时间
	Certificate     string        // 连接凭证，为空表示无凭证
}

func NewBusiness(handler BusinessEventsInterface, conf *BusinessConfig) *Business {
	return &Business{
		c:       conf,
		handler: handler,

		gatewayConnections:             sync.Map{},
		gatewayAddresses:               map[string]struct{}{},
		connectingGatewayAddresses:     map[string]struct{}{},
		waitingConnectGatewayAddresses: map[string]struct{}{},
	}
}

func (b *Business) Start() {
	go b.connectToRegister()
}

func (b *Business) connectToRegister() {
	c := client.NewAsyncTcpConnection(b.c.RegisterAddress)

	var ping *timer.Timer

	c.OnConnect = func(conn *client.AsyncTcpConnection) {
		buffer := (&RegisterMessage{
			Event:       "worker_connect",
			Certificate: b.c.Certificate,
		}).Bytes()

		if err := conn.Send(buffer); err != nil {
			logger.Sugar.Panic(err)
		}

		logger.Sugar.Infof("register %v connected", b.c.RegisterAddress)

		ping = timer.NewTimer(b.c.PingInterval, func() {
			conn.Send(PingData)
		})
		go ping.Start()
	}

	c.OnClose = func(conn *client.AsyncTcpConnection) {
		if ping != nil {
			ping.Stop()
		}
		logger.Sugar.Infof("register %v disconnected", b.c.RegisterAddress)
		// 重新连接
		time.AfterFunc(2*time.Second, func() {
			for {
				logger.Sugar.Infof("register %v attempting to reconnect", b.c.RegisterAddress)

				if err := c.Connect(); err != nil {
					logger.Sugar.Error(err)
					time.Sleep(2 * time.Second)
					continue
				}

				break
			}
		})
	}
	c.OnMessage = b.onRegisterConnectionMessage

	if err := c.Connect(); err != nil {
		logger.Sugar.Panic(err)
	}
}

func (b *Business) onRegisterConnectionMessage(conn *client.AsyncTcpConnection, data []byte) {
	var msg RegisterMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.Event {
	case "broadcast_addresses":
		if len(msg.Addresses) == 0 {
			logger.Sugar.Info("received bad data from register. addresses empty")
			return
		}

		var temp = map[string]struct{}{}
		for _, addr := range msg.Addresses {
			temp[addr] = struct{}{}
		}
		b.gatewayAddresses = temp

		b.checkGatewayConnections(msg.Addresses)
	default:
		logger.Sugar.Infof("receive bad event: %v from register.", msg.Event)
	}
}

func (b *Business) checkGatewayConnections(addresses []string) {
	for _, addr := range addresses {
		b.tryToConnectGateway(addr)
	}
}

func (b *Business) tryToConnectGateway(addr string) {
	_, isConnecting := b.connectingGatewayAddresses[addr]
	_, isIn := b.gatewayAddresses[addr]

	if _, found := b.gatewayConnections.Load(addr); !found && !isConnecting && isIn {
		logger.Sugar.Infof("try to connect to gateway %v", addr)

		conn := client.NewAsyncTcpConnection(addr)

		conn.OnConnect = b.onGatewayConnect
		conn.OnClose = b.onGatewayClose
		conn.OnMessage = b.onGatewayMessage
		b.connectingGatewayAddresses[addr] = struct{}{}
		b.gatewayConnections.Store(addr, conn)

		go conn.Connect()
	}

	delete(b.waitingConnectGatewayAddresses, addr)
}

func (b *Business) onGatewayConnect(conn *client.AsyncTcpConnection) {
	addr := conn.Addr()
	b.gatewayConnections.Store(addr, conn)

	// 发送认证
	if err := conn.Send((&GatewayMessage{
		Cmd:     protocol.CMD_WORKER_CONNECT,
		Body:    []byte(b.c.Certificate),
		ExtData: nil,
	}).Bytes()); err != nil {
		logger.Sugar.Error(err)
	}

	delete(b.connectingGatewayAddresses, addr)
	delete(b.waitingConnectGatewayAddresses, addr)

	logger.Sugar.Infof("gateway %v connected", addr)
}

func (b *Business) onGatewayClose(conn *client.AsyncTcpConnection) {
	addr := conn.Addr()

	b.gatewayConnections.Delete(addr)
	delete(b.connectingGatewayAddresses, addr)
	if _, found := b.gatewayAddresses[addr]; found {
		if _, found := b.waitingConnectGatewayAddresses[addr]; !found {
			b.waitingConnectGatewayAddresses[addr] = struct{}{}
		}
	}

	logger.Sugar.Infof("gateway %v disconnected", addr)
}

func (b *Business) onGatewayMessage(conn *client.AsyncTcpConnection, data []byte) {
	var msg GatewayMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Sugar.Error(err)
		return
	}

	switch msg.Cmd {
	case protocol.CMD_ON_CONNECT:
		b.handler.OnConnect(msg.ConnId)
	case protocol.CMD_ON_MESSAGE:
		b.handler.OnMessage(msg.ConnId, NewBusinessEventsMessage(&msg))
	case protocol.CMD_ON_CLOSE:
		b.handler.OnClose(msg.ConnId)
	case protocol.CMD_PING:

	default:
		logger.Sugar.Warnf("unknown gateway message cmd: %v", msg.Cmd)
	}
}
