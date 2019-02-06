package nksh

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/juju/errors"
)

type Neo4jProperties map[string]interface{}

func (p Neo4jProperties) MustGet(field string) interface{} {
	if val, ok := p[field]; ok {
		return val
	}
	panic(fmt.Sprintf("Neo4jProperties:MustGet: field %s undefined", field))
}

type Neo4jBeforeOrAfter struct {
	Labels     []string        `json:"labels"`
	Properties Neo4jProperties `json:"properties"`
}

type Neo4jStartOrEnd struct {
	Labels []string `json:"labels"`
	ID     string   `json:"id"`
}

type Neo4jSource struct {
	Hostname string `json:"hostname"`
}

type Neo4jPayload struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	RelLabel string             `json:"label"`
	Start    Neo4jStartOrEnd    `json:"start"`
	End      Neo4jStartOrEnd    `json:"end"`
	After    Neo4jBeforeOrAfter `json:"after"`
	Before   Neo4jBeforeOrAfter `json:"before"`
}

type Neo4jMeta struct {
	Timestamp     int64       `json:"timestamp"`
	Username      string      `json:"username"`
	TxID          int         `json:"tx_id"`
	TxEventID     int         `json:"tx_event_id"`
	TxEventsCount int         `json:"tx_events_count"`
	Operation     string      `json:"operation"`
	Source        Neo4jSource `json:"source"`
}

type Neo4jMessage struct {
	Meta    Neo4jMeta    `json:"meta"`
	Payload Neo4jPayload `json:"payload"`
}

func (p *Neo4jMessage) ToNodeContext() (*NodeContext, error) {
	id, err := strconv.ParseUint(p.Payload.ID, 10, 64)
	if err != nil {
		return nil, errors.Annotate(err, "ParseUint [id]")
	}

	n := NodeContext{
		NodeID:      id,
		ChangeInfos: make(ChangeInfos),
		Operation:   p.Meta.Operation,
		TimeStamp: time.Unix(0,
			p.Meta.Timestamp*int64(time.Millisecond),
		),
	}

	switch p.Meta.Operation {
	case "deleted":
		n.Properties = p.Payload.Before.Properties
		n.buildChanges(true, p.Payload.Before.Properties)
	case "created":
		n.Properties = p.Payload.After.Properties
		n.buildChanges(false, p.Payload.After.Properties)
	case "updated":
		n.Properties = p.Payload.After.Properties
		n.buildChanges(false, p.Payload.After.Properties)
		n.buildChanges(true, p.Payload.Before.Properties)
	}

	// remove unchanged properties
	for field, info := range n.ChangeInfos {
		if info.After == info.Before {
			delete(n.ChangeInfos, field)
		}
	}

	return &n, nil
}

type Neo4jMessageCodec struct{}

func (p *Neo4jMessageCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (p *Neo4jMessageCodec) Decode(data []byte) (interface{}, error) {
	var m Neo4jMessage
	return &m, json.Unmarshal(data, &m)
}
