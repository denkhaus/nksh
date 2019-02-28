package shared

import (
	"encoding/json"
	"time"
)

type ChangeInfo struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

func (p ChangeInfo) Created() bool {
	return p.Before == nil && p.After != nil
}

func (p ChangeInfo) Updated() bool {
	if p.Before == nil && p.After == nil {
		return false
	}
	return p.Before != p.After
}

func (p ChangeInfo) Deleted() bool {
	return p.Before != nil && p.After == nil
}

type ChangeInfos map[string]ChangeInfo

func (p ChangeInfos) Created(field string) bool {
	if field == "" {
		return false
	}
	if field == "*" {
		for _, info := range p {
			if info.Created() {
				return true
			}
		}
	} else {
		if info, ok := p[field]; ok {
			return info.Created()
		}
	}

	return false
}

func (p ChangeInfos) Updated(field string) bool {
	if field == "" {
		return false
	}
	if field == "*" {
		for _, info := range p {
			if info.Updated() {
				return true
			}
		}
	} else {
		if info, ok := p[field]; ok {
			return info.Updated()
		}
	}

	return false
}

func (p ChangeInfos) Deleted(field string) bool {
	if field == "" {
		return false
	}
	if field == "*" {
		for _, info := range p {
			if info.Deleted() {
				return true
			}
		}
	} else {
		if info, ok := p[field]; ok {
			return info.Deleted()
		}
	}

	return false
}

type EventContext struct {
	TimeStamp   time.Time   `json:"time_stamp"`
	Operation   Operation   `json:"operation"`
	NodeID      int64       `json:"node_id"`
	ChangeInfos ChangeInfos `json:"change_infos"`
	Properties  Properties  `json:"properties"`
}

func (p *EventContext) Match(

	operation Operation,
	fieldName string,
	fieldOperation Operation,
	conditions EvalFuncs,

) bool {

	matcher := NewMatcher(
		func() (bool, EvalFunc) {
			return p.Operation == UpdatedOperation &&
					p.Operation == operation &&
					fieldName != "" && fieldOperation != "",
				func(_ interface{}) bool {
					switch fieldOperation {
					case CreatedOperation:
						return p.ChangeInfos.Created(fieldName)
					case UpdatedOperation:
						return p.ChangeInfos.Updated(fieldName)
					case DeletedOperation:
						return p.ChangeInfos.Deleted(fieldName)
					default:
						return false
					}
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

func (p *EventContext) BuildChanges(before bool, props map[string]interface{}) {
	for field, value := range props {
		if info, ok := p.ChangeInfos[field]; ok {
			if before {
				info.Before = value
			} else {
				info.After = value
			}

			p.ChangeInfos[field] = info
		} else {
			if before {
				p.ChangeInfos[field] = ChangeInfo{
					Before: value,
				}
			} else {
				p.ChangeInfos[field] = ChangeInfo{
					After: value,
				}
			}
		}
	}
}

type EventContextCodec struct{}

func (p *EventContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *EventContextCodec) Decode(data []byte) (interface{}, error) {
	var m EventContext
	return &m, json.Unmarshal(data, &m)
}
