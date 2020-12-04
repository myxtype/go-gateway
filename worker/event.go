package worker

// 事件
type WorkerEventInterface interface {
	OnWorkerStart()
	OnConnect(conn *Connection)
	OnMessage(conn *Connection, message []byte)
	OnClose(conn *Connection)
	OnWorkerStop()
}

type WorkerEventProxy struct {
	ProxyOnWorkerStart func()
	ProxyOnConnect     func(conn *Connection)
	ProxyOnMessage     func(conn *Connection, message []byte)
	ProxyOnClose       func(conn *Connection)
	ProxyOnWorkerStop  func()
}

func (e *WorkerEventProxy) OnWorkerStart() {
	if e.ProxyOnWorkerStart != nil {
		e.ProxyOnWorkerStart()
	}
}

func (e *WorkerEventProxy) OnConnect(conn *Connection) {
	if e.ProxyOnConnect != nil {
		e.ProxyOnConnect(conn)
	}
}

func (e *WorkerEventProxy) OnMessage(conn *Connection, message []byte) {
	if e.ProxyOnMessage != nil {
		e.ProxyOnMessage(conn, message)
	}
}

func (e *WorkerEventProxy) OnClose(conn *Connection) {
	if e.ProxyOnClose != nil {
		e.ProxyOnClose(conn)
	}
}

func (e *WorkerEventProxy) OnWorkerStop() {
	if e.ProxyOnWorkerStop != nil {
		e.ProxyOnWorkerStop()
	}
}
