package gateway

import "encoding/json"

type RegisterMessage struct {
	Event       string   `json:"event,omitempty"`
	Certificate string   `json:"certificate,omitempty"`
	Address     string   `json:"address,omitempty"`
	Addresses   []string `json:"addresses,omitempty"`
}

func (msg *RegisterMessage) Bytes() []byte {
	b, _ := json.Marshal(msg)
	return b
}
