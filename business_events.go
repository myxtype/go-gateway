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
}

func (bm *BusinessEventsMessage) UnmarshalData(v interface{}) error {
	return json.Unmarshal(bm.Data, v)
}

func NewBusinessEventsMessage(msg *BusinessMessage) (*BusinessEventsMessage, error) {
	bem := &BusinessEventsMessage{Data: msg.Body}
	if err := msg.UnmarshalExtData(&bem.Session); err != nil {
		return nil, err
	}
	return bem, nil
}
