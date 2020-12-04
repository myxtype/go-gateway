package gateway

import (
	"encoding/json"
	"github.com/myxtype/go-gateway/protocol"
)

type BusinessMessage struct {
	Cmd     protocol.Protocol `json:"c"`
	Body    json.RawMessage   `json:"b"`
	ConnId  string            `json:"ci"`
	Flag    bool              `json:"f"`
	ExtData json.RawMessage   `json:"e"`
}

func (bm *BusinessMessage) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(bm.Body, v)
}

func (bm *BusinessMessage) UnmarshalExtData(v interface{}) error {
	return json.Unmarshal(bm.ExtData, v)
}

func (bm *BusinessMessage) Bytes() []byte {
	b, _ := json.Marshal(bm)
	return b
}
