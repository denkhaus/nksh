package hub

import (
	"encoding/json"

	"github.com/denkhaus/nksh/shared"
)

type Context struct {
	Sender     string            `json:"sender"`
	SenderID   int64             `json:"sender_id"`
	Operation  shared.Operation  `json:"operation"`
	Receiver   string            `json:"receiver"`
	ReceiverID int64             `json:"receiver_id"`
	Properties shared.Properties `json:"properties"`
}

func (p *Context) Match(
	operation shared.Operation,
	sender string,
	condition ConditionFunc,

) bool {
	matcher := shared.NewMatcher(
		func() (bool, func() bool) {
			return sender != "",
				func() bool {
					return p.Sender == sender
				}
		},
		func() (bool, func() bool) {
			return operation != "",
				func() bool {
					return p.Operation == operation
				}
		},
		func() (bool, func() bool) {
			return condition != nil,
				func() bool {
					return condition(*p)
				}
		},
	)

	return matcher.Eval()
}

type ContextCodec struct{}

func (p *ContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *ContextCodec) Decode(data []byte) (interface{}, error) {
	var m Context
	return &m, json.Unmarshal(data, &m)
}
