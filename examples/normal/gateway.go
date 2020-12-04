package main

import (
	"github.com/myxtype/go-gateway"
	"log"
)

func main() {
	g := gateway.NewGateway(&gateway.GatewayConfig{
		Addr:            "127.0.0.1:1235",
		RegisterAddress: "127.0.0.1:1234",
		PingInterval:    0,
		Certificate:     "",
	})

	if err := g.Start(); err != nil {
		log.Panic(err)
	}
}
