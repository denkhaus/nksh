package nksh

import (
	"encoding/json"
	"time"
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

type NodeContext struct {
	TimeStamp   time.Time       `json:"time_stamp"`
	Operation   string          `json:"operation"`
	NodeID      uint64          `json:"node_id"`
	ChangeInfos ChangeInfos     `json:"change_infos"`
	Properties  Neo4jProperties `json:"properties"`
}

func (p *NodeContext) Match(

	operation string,
	fieldName string,
	fieldOperation string,

) bool {

	if p.Operation == "updated" && p.Operation == operation {
		switch fieldOperation {
		case "created":
			return p.ChangeInfos.Created(fieldName)
		case "updated":
			return p.ChangeInfos.Updated(fieldName)
		case "deleted":
			return p.ChangeInfos.Deleted(fieldName)
		default:
			return false
		}
	}

	return p.Operation == operation
}

func (p *NodeContext) buildChanges(before bool, props map[string]interface{}) {
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

type NodeContextCodec struct{}

func (p *NodeContextCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *NodeContextCodec) Decode(data []byte) (interface{}, error) {
	var m NodeContext
	return &m, json.Unmarshal(data, &m)
}
