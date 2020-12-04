package main

import (
	"github.com/myxtype/go-gateway"
	"log"
)

func main() {
	reg := gateway.NewRegister(&gateway.RegisterConfig{
		Addr:        "127.0.0.1:1234",
		Certificate: "",
	})

	if err := reg.Start(); err != nil {
		log.Panic(err)
	}
}
