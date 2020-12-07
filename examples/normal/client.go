package main

import (
	"github.com/myxtype/go-gateway/client"
	"github.com/myxtype/go-gateway/pkg/logger"
)

func main() {
	for i := 0; i < 1000; i++ {
		go func(i int) {
			logger.Sugar.Info(i)
			conn := client.NewAsyncTcpConnection("127.0.0.1:2000")
			conn.Connect()
		}(i)
	}

	select {}
}
