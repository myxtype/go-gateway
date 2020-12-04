package main

import (
	"github.com/myxtype/go-gateway"
	"time"
)

func main() {
	b := gateway.NewBusiness(&gateway.BusinessConfig{
		Addr:            "127.0.0.1:1999",
		RegisterAddress: "127.0.0.1:1234",
		PingInterval:    25 * time.Second,
		Certificate:     "",
	})

	if err := b.Start(); err != nil {
		panic(err)
	}
}
