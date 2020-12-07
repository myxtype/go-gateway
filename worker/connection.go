package worker

import (
	"encoding/json"
	"github.com/twinj/uuid"
	"net"
	"sync"
)

type Connection struct {
	id   string
	conn *net.TCPConn

	// 存一些此连接相关的数据
	Payload sync.Map
}

func NewConnection(conn *net.TCPConn) *Connection {
	return &Connection{
		id:   uuid.NewV4().String(),
		conn: conn,
	}
}

func (c *Connection) Id() string {
	return c.id
}

func (c *Connection) Conn() *net.TCPConn {
	return c.conn
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) Send(b []byte) error {
	_, err := c.conn.Write(append(b, '\n'))
	return err
}

func (c *Connection) SendJson(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(b)
}
