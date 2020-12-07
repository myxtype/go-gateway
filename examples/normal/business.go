package main

import (
	"github.com/myxtype/go-gateway"
	"time"
)

type BusinessHandler struct {
}

func (h *BusinessHandler) OnConnect(connId string)                                     {}
func (h *BusinessHandler) OnMessage(connId string, msg *gateway.BusinessEventsMessage) {}
func (h *BusinessHandler) OnClose(connId string)                                       {}

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
