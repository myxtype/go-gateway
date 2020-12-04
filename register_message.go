package gateway

import "encoding/json"

type RegisterMessage struct {
	Event       string   `json:"event"`
	Certificate string   `json:"certificate"`
	Address     string   `json:"address"`
	Addresses   []string `json:"addresses"`
}

func (msg *RegisterMessage) Bytes() []byte {
	b, _ := json.Marshal(msg)
	return b
}
