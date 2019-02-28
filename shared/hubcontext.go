package shared

import (
	"encoding/json"
)

type HubContext struct {
	Sender     string     `json:"sender"`
	SenderID   int64      `json:"sender_id"`
	Operation  Operation  `json:"operation"`
	Receiver   string     `json:"receiver"`
	ReceiverID int64      `json:"receiver_id"`
	Properties Properties `json:"properties"`
}

func (p *HubContext) Match(

	operation Operation,
	sender string,
	conditions EvalFuncs,

) bool {
	matcher := NewMatcher(
		func() (bool, EvalFunc) {
			return sender != "",
				func(_ interface{}) bool {
					return p.Sender == sender
				}
		},
		func() (bool, EvalFunc) {
			return operation != "",
				func(_ interface{}) bool {
					return p.Operation == operation
				}
		},
		MatchConditions(conditions...),
	)

	return matcher.Eval(*p)
}

type HubContextCodec struct{}

func (p *HubContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *HubContextCodec) Decode(data []byte) (interface{}, error) {
	var m HubContext
	return &m, json.Unmarshal(data, &m)
}
