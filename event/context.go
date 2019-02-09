package event

import (
	"encoding/json"
	"time"

	"github.com/denkhaus/nksh/shared"
)

type ChangeInfo struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

type ChangeInfos map[string]ChangeInfo

func (p ChangeInfos) Created(field string) bool {
	if field == "" {
		return false
	}
	if field == "*" {
		for _, info := range p {
			if info.Before == nil && info.After != nil {
				return true
			}
		}
	} else {
		if info, ok := p[field]; ok {
			if info.Before == nil && info.After != nil {
				return true
			}
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
			if info.Before == nil || info.After == nil {
				continue
			}
			if info.Before != info.After {
				return true
			}
		}
	} else {
		if info, ok := p[field]; ok {
			if info.Before == nil || info.After == nil {
				return false
			}
			return info.Before != info.After
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
			if info.Before != nil && info.After == nil {
				return true
			}
		}
	} else {
		if info, ok := p[field]; ok {
			if info.Before != nil && info.After == nil {
				return true
			}
		}
	}

	return false
}

type Context struct {
	TimeStamp   time.Time         `json:"time_stamp"`
	Operation   shared.Operation  `json:"operation"`
	NodeID      int64             `json:"node_id"`
	ChangeInfos ChangeInfos       `json:"change_infos"`
	Properties  shared.Properties `json:"properties"`
}

func (p *Context) Match(

	operation shared.Operation,
	fieldName string,
	fieldOperation shared.Operation,

) bool {

	if p.Operation == shared.UpdatedOperation && p.Operation == operation {
		switch fieldOperation {
		case shared.CreatedOperation:
			return p.ChangeInfos.Created(fieldName)
		case shared.UpdatedOperation:
			return p.ChangeInfos.Updated(fieldName)
		case shared.DeletedOperation:
			return p.ChangeInfos.Deleted(fieldName)
		default:
			return false
		}
	}

	return p.Operation == operation
}

func (p *Context) buildChanges(before bool, props map[string]interface{}) {
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

type ContextCodec struct{}

func (p *ContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *ContextCodec) Decode(data []byte) (interface{}, error) {
	var m Context
	return &m, json.Unmarshal(data, &m)
}
