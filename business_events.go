package gateway

import (
	"encoding/json"
)

type BusinessEventsInterface interface {
	OnConnect(connId string)
	OnMessage(connId string, msg *BusinessEventsMessage)
	OnClose(connId string)
}

type BusinessEventsMessage struct {
	Data    json.RawMessage
	Session map[string]interface{}
	Remote  map[string]interface{}
}

func (bm *BusinessEventsMessage) UnmarshalData(v interface{}) error {
	return json.Unmarshal(bm.Data, v)
}

func NewBusinessEventsMessage(msg *GatewayMessage) *BusinessEventsMessage {
	return &BusinessEventsMessage{
		Data:    msg.Body,
		Session: msg.Session,
		Remote:  msg.Remote,
	}
}
