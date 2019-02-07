package nksh

import "encoding/json"

type HubMessage struct {
	Sender     string                 `json:"sender"`
	Receiver   string                 `json:"receiver"`
	Properties map[string]interface{} `json:"properties"`
}

type HubMessageCodec struct{}

func (p *HubMessageCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *HubMessageCodec) Decode(data []byte) (interface{}, error) {
	var m HubMessage
	return &m, json.Unmarshal(data, &m)
}
