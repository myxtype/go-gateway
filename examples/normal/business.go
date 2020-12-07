package main

import (
	"github.com/myxtype/go-gateway"
	"github.com/myxtype/go-gateway/pkg/logger"
	"time"
)

type BusinessHandler struct {
}

func (h *BusinessHandler) OnConnect(connId string) {
	logger.Sugar.Infof("new client connect %v", connId)
}

func (h *BusinessHandler) OnMessage(connId string, msg *gateway.BusinessEventsMessage) {
	logger.Sugar.Infof("new message %v from %v", msg.Data, connId)
}

func (h *BusinessHandler) OnClose(connId string) {
	logger.Sugar.Infof("client disconnect %v", connId)
}

func main() {
	b := gateway.NewBusiness(&BusinessHandler{}, &gateway.BusinessConfig{
		RegisterAddress: "127.0.0.1:1234",
		PingInterval:    25 * time.Second,
		Certificate:     "",
	})

	if err := b.Start(); err != nil {
		panic(err)
	}
}
