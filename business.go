package gateway

import (
	"github.com/myxtype/go-gateway/client"
	"github.com/myxtype/go-gateway/pkg/timer"
	"github.com/myxtype/go-gateway/worker"
	"log"
	"sync"
	"time"
)

type Business struct {
	c                  *BusinessConfig
	gatewayConnections sync.Map
}

type BusinessConfig struct {
	Addr            string        // 监听地址
	RegisterAddress string        // 注册中心地址
	PingInterval    time.Duration // 心跳时间
	Certificate     string        // 连接凭证，为空表示无凭证
}

func NewBusiness(conf *BusinessConfig) *Business {
	return &Business{
		c: conf,
	}
}

func (e *Business) Start() error {
	w := worker.NewWorker(e.c.Addr, e)

	if err := w.Start(); err != nil {
		return err
	}
	return nil
}

func (b *Business) OnWorkerStart() {

	go b.connectToRegister()
}

func (b *Business) OnConnect(conn *worker.Connection) {

}

func (b *Business) OnMessage(conn *worker.Connection, message []byte) {

}

func (b *Business) OnClose(conn *worker.Connection) {

}

func (b *Business) OnWorkerStop() {

}

func (b *Business) connectToRegister() {
	c := client.NewAsyncTcpConnection(b.c.RegisterAddress)

	var ping *timer.Timer

	c.OnConnect = func(conn *client.AsyncTcpConnection) {
		buffer := (&RegisterMessage{
			Event:       "worker_connect",
			Certificate: b.c.Certificate,
			Address:     b.c.Addr,
		}).Bytes()

		if err := conn.Send(buffer); err != nil {
			log.Panic(err)
		}

		ping = timer.NewTimer(b.c.PingInterval, func() {
			conn.Send([]byte("{\"event\":\"ping\"}"))
		})
		go ping.Start()
	}

	c.OnClose = func(conn *client.AsyncTcpConnection) {
		if ping != nil {
			ping.Stop()
		}
		// todo
	}

	if err := c.Connect(); err != nil {
		log.Panic(err)
	}
}
