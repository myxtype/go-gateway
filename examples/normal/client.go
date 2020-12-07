package main

import "github.com/myxtype/go-gateway/client"

func main() {
	for i := 0; i < 1000; i++ {
		conn := client.NewAsyncTcpConnection("127.0.0.1:2000")
		go conn.Connect()
	}

	select {}
}
