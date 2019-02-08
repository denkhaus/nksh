package nksh

import "encoding/json"

type HubMessage struct {
	SenderLabel string     `json:"sender_label"`
	SenderID    int64      `json:"sender_id"`
	Receivers   []string   `json:"receivers"`
	Properties  Properties `json:"properties"`
}

type HubMessageCodec struct{}

func (p *HubMessageCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *HubMessageCodec) Decode(data []byte) (interface{}, error) {
	var m HubMessage
	return &m, json.Unmarshal(data, &m)
}
