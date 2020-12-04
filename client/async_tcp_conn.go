package client

import (
	"bufio"
	"io"
	"net"
)

type AsyncTcpConnection struct {
	addr string
	conn *net.TCPConn

	OnConnect func(*AsyncTcpConnection)
	OnClose   func(*AsyncTcpConnection)
	OnMessage func(*AsyncTcpConnection, []byte)
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
			tc.OnClose(tc)
		}
	}()

	if tc.OnConnect != nil {
		tc.OnConnect(tc)
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
		msg, err := reader.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}

		if tc.OnMessage != nil {
			tc.OnMessage(tc, msg)
		}
	}
}

func (tc *AsyncTcpConnection) Send(b []byte) error {
	_, err := tc.conn.Write(append(b, '\n'))
	return err
}
