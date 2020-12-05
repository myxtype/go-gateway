package worker

import (
	"bufio"
	"io"
	"log"
	"net"
)

type Worker struct {
	addr    string
	handler WorkerEventInterface
}

func NewWorker(addr string, handler WorkerEventInterface) *Worker {
	return &Worker{addr: addr, handler: handler}
}

func (w *Worker) Start() error {
	netAddr, err := net.ResolveTCPAddr("tcp", w.addr)
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
	defer func() {
		conn.Close()
	}()

	connection := NewConnection(conn)
	w.handler.OnConnect(connection)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}
		w.handler.OnMessage(connection, message)
	}
	w.handler.OnClose(connection)
}