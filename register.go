package gateway

import (
	"encoding/json"
	"github.com/myxtype/go-gateway/pkg/logger"
	"github.com/myxtype/go-gateway/worker"
	"sync"
)

type Register struct {
	c                  *RegisterConfig
	connId             int64
	gatewayConnections sync.Map
	workerConnections  sync.Map
}

type RegisterConfig struct {
	Addr        string // 监听地址
	Certificate string // 连接凭证，为空表示无凭证
}

func NewRegister(conf *RegisterConfig) *Register {
	return &Register{
		c: conf,
	}
}

func (r *Register) Start() error {
	w := worker.NewWorker(r, &worker.WorkerConfig{Addr: r.c.Addr})

	if err := w.Start(); err != nil {
		return err
	}
	return nil
}

func (r *Register) OnWorkerStart() {}

func (r *Register) OnWorkerStop() {}

func (r *Register) OnConnect(conn *worker.Connection) {
	logger.Sugar.Infof("connected %v id: %v", conn.RemoteAddr().String(), conn.Id())
}

func (r *Register) OnMessage(conn *worker.Connection, message []byte) {
	var msg RegisterMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		conn.Close()
		return
	}

	switch msg.Event {
	case "gateway_connect":
		if msg.Certificate != r.c.Certificate {
			logger.Sugar.Infof("certificate invalid, gateway address: %v", conn.RemoteAddr().String())
			conn.Close()
			return
		}
		logger.Sugar.Infof("new gateway %v id: %v", conn.RemoteAddr().String(), conn.Id())
		r.gatewayConnections.Store(conn.Id(), msg.Address)
		r.broadcastAddresses(nil)
	case "worker_connect":
		if msg.Certificate != r.c.Certificate {
			logger.Sugar.Infof("certificate invalid, worker address: %v", conn.RemoteAddr().String())
			conn.Close()
			return
		}
		logger.Sugar.Infof("new worker %v id: %v", conn.RemoteAddr().String(), conn.Id())
		r.workerConnections.Store(conn.Id(), conn)
		r.broadcastAddresses(conn)
	case "ping":
		logger.Sugar.Infof("ping from %v id: %v", conn.RemoteAddr().String(), conn.Id())
	default:
		conn.Close()
		return
	}
}

func (r *Register) OnClose(conn *worker.Connection) {
	logger.Sugar.Infof("disconnected from %v id: %v", conn.RemoteAddr().String(), conn.Id())

	if _, found := r.gatewayConnections.Load(conn.Id()); found {
		r.gatewayConnections.Delete(conn.Id())
		r.broadcastAddresses(nil)
	}
	if _, found := r.workerConnections.Load(conn.Id()); found {
		r.workerConnections.Delete(conn.Id())
	}
}

// 广播地址
func (r *Register) broadcastAddresses(conn *worker.Connection) {
	var addresses []string
	r.gatewayConnections.Range(func(key, value interface{}) bool {
		addresses = append(addresses, value.(string))
		return true
	})

	msg := (&RegisterMessage{
		Event:     "broadcast_addresses",
		Addresses: addresses,
	}).Bytes()

	if conn != nil {
		if err := conn.Send(msg); err != nil {
			logger.Sugar.Error(err)
		}
		return
	}

	r.workerConnections.Range(func(key, value interface{}) bool {
		if err := value.(*worker.Connection).Send(msg); err != nil {
			logger.Sugar.Error(err)
		}
		return true
	})
}
