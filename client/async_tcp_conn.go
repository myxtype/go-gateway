package client

import (
	"bufio"
	"io"
	"log"
	"net"
)

type AsyncTcpConnection struct {
	addr string
	conn *net.TCPConn

	OnConnect func(conn *net.TCPConn)
	OnClose   func(conn *net.TCPConn)
}

func NewAsyncTcpConnection(addr string) *AsyncTcpConnection {
	return &AsyncTcpConnection{
		addr: addr,
	}
}

func (tc *AsyncTcpConnection) Connect() error {
	netAddr, err := net.ResolveTCPAddr("tcp", tc.addr)
	if err != nil {
		return err
	}

	tc.conn, err = net.DialTCP("tcp", nil, netAddr)
	if err != nil {
		return err
	}
	defer func() {
		tc.conn.Close()
		if tc.OnClose != nil {
			tc.OnClose(tc.conn)
		}
	}()

	if tc.OnConnect != nil {
		tc.OnConnect(tc.conn)
	}

	tc.onMessageReceived(tc.conn)

	return nil
}

func (tc *AsyncTcpConnection) Conn() *net.TCPConn {
	return tc.conn
}

func (tc *AsyncTcpConnection) onMessageReceived(conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		log.Println(msg)
	}
}
