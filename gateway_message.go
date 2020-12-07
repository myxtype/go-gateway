package gateway

import (
	"encoding/json"
	"github.com/myxtype/go-gateway/pkg/logger"
	"github.com/myxtype/go-gateway/protocol"
)

type GatewayMessage struct {
	Cmd     protocol.Protocol      `json:"cmd,omitempty"`
	Body    json.RawMessage        `json:"body,omitempty"`
	ConnId  string                 `json:"conn_id,omitempty"`
	Flag    bool                   `json:"flag,omitempty"`
	ExtData json.RawMessage        `json:"ext_data,omitempty"`
	Session map[string]interface{} `json:"session,omitempty"`
	Remote  map[string]interface{} `json:"remote,omitempty"`
}

func (m *GatewayMessage) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(m.Body, v)
}

func (m *GatewayMessage) UnmarshalExtData(v interface{}) error {
	return json.Unmarshal(m.ExtData, v)
}

func (m *GatewayMessage) Bytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		logger.Sugar.Error(err)
	}
	return b
}
