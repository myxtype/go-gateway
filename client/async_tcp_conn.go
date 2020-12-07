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

func (tc *AsyncTcpConnection) Addr() string {
	return tc.addr
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

	tc.onMessageReceived()

	return nil
}

func (tc *AsyncTcpConnection) onMessageReceived() {
	reader := bufio.NewReader(tc.conn)
	for {
		msg, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			break
		}

		if tc.OnMessage != nil {
			tc.OnMessage(tc, msg)
		}
	}
}

func (tc *AsyncTcpConnection) Conn() *net.TCPConn {
	return tc.conn
}

func (tc *AsyncTcpConnection) Close() error {
	return tc.conn.Close()
}

func (tc *AsyncTcpConnection) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

func (tc *AsyncTcpConnection) Write(b []byte) (int, error) {
	return tc.conn.Write(b)
}

func (tc *AsyncTcpConnection) Send(b []byte) error {
	_, err := tc.conn.Write(append(b, '\n'))
	return err
}
