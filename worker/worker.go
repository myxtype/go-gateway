package worker

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Worker struct {
	c       *WorkerConfig
	handler WorkerEventInterface

	connected int64
}

type WorkerConfig struct {
	Addr    string
	MaxConn int64
}

func NewWorker(handler WorkerEventInterface, conf *WorkerConfig) *Worker {
	return &Worker{c: conf, handler: handler}
}

func (w *Worker) Start() error {
	netAddr, err := net.ResolveTCPAddr("tcp", w.c.Addr)
	if err != nil {
		return err
	}
	tcpListener, err := net.ListenTCP("tcp", netAddr)
	if err != nil {
		return err
	}
	defer func() {
		tcpListener.Close()
		w.handler.OnWorkerStop()
	}()

	w.handler.OnWorkerStart()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		go w.tcpPipe(tcpConn)
	}
}

func (w *Worker) tcpPipe(conn *net.TCPConn) {
	connection := NewConnection(conn)

	defer func() {
		conn.Close()
		atomic.AddInt64(&w.connected, -1)
		w.handler.OnClose(connection)
	}()

	atomic.AddInt64(&w.connected, 1)
	// 超过最大连接数
	if w.c.MaxConn > 0 && w.connected > w.c.MaxConn {
		return
	}

	w.handler.OnConnect(connection)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			break
		}
		w.handler.OnMessage(connection, message)
	}
}
