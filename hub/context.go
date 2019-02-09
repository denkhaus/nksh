package hub

import (
	"encoding/json"

	"github.com/denkhaus/nksh/shared"
)

type Context struct {
	SenderLabel   string            `json:"sender_label"`
	SenderID      int64             `json:"sender_id"`
	Operation     string            `json:"operation"`
	ReceiverLabel string            `json:"receiver_label"`
	ReceiverID    int64             `json:"receiver_id"`
	Properties    shared.Properties `json:"properties"`
}

func (p *Context) Match(
	operation string,
	sender string,
	condition ConditionFunc,

) bool {

	if p.Operation == operation &&
		sender == p.SenderLabel {
		if condition != nil {
			return condition(*p)
		}
	}

	return false
}

type ContextCodec struct{}

func (p *ContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *ContextCodec) Decode(data []byte) (interface{}, error) {
	var m Context
	return &m, json.Unmarshal(data, &m)
}
