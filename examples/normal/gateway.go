package main

import (
	"github.com/myxtype/go-gateway"
	"log"
	"time"
)

func main() {
	g := gateway.NewGateway(&gateway.GatewayConfig{
		Addr:            "127.0.0.1:2000",
		InnerAddr:       "127.0.0.1:2001",
		RegisterAddress: "127.0.0.1:1234",
		PingInterval:    25 * time.Second,
		Certificate:     "",
	})

	if err := g.Start(); err != nil {
		log.Panic(err)
	}
}
