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
	conditions shared.EvalFuncs,

) bool {
	matcher := shared.NewMatcher(
		func() (bool, shared.EvalFunc) {
			return sender != "",
				func(_ interface{}) bool {
					return p.Sender == sender
				}
		},
		func() (bool, shared.EvalFunc) {
			return operation != "",
				func(_ interface{}) bool {
					return p.Operation == operation
				}
		},
		shared.MatchConditions(conditions...),
	)

	return matcher.Eval(*p)
}

type ContextCodec struct{}

func (p *ContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *ContextCodec) Decode(data []byte) (interface{}, error) {
	var m Context
	return &m, json.Unmarshal(data, &m)
}
