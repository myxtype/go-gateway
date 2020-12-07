package gateway

import (
	"encoding/json"
	"github.com/myxtype/go-gateway/pkg/logger"
	"github.com/myxtype/go-gateway/protocol"
)

type BusinessMessage struct {
	Cmd     protocol.Protocol      `json:"cmd,omitempty"`
	Body    json.RawMessage        `json:"body,omitempty"`
	ConnId  string                 `json:"conn_id,omitempty"`
	Flag    bool                   `json:"flag,omitempty"`
	ExtData json.RawMessage        `json:"ext_data,omitempty"`
	Session map[string]interface{} `json:"session"`
	Remote  map[string]interface{} `json:"remote"`
}

func (bm *BusinessMessage) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(bm.Body, v)
}

func (bm *BusinessMessage) UnmarshalExtData(v interface{}) error {
	return json.Unmarshal(bm.ExtData, v)
}

func (bm *BusinessMessage) Bytes() []byte {
	b, err := json.Marshal(bm)
	if err != nil {
		logger.Sugar.Error(err)
	}
	return b
}
