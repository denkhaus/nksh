package shared

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/lovoo/goka"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

type DispatcherFunc func(ctx context.Context, kServers, zServers []string) func() error

type Operation string
type Operations []Operation

var (
	CreatedOperation = Operation("created")
	UpdatedOperation = Operation("updated")
	DeletedOperation = Operation("deleted")
)

var (
	HubStream  = goka.Stream("Hub")  // Hubmessages Entity-> Hub
	HubGroup   = goka.Group("Hub")   // Hubmessages Entity-> Hub
	InputGroup = goka.Group("Input") // Neo4jMessages  -> Incoming
)

var (
	Neo4jDriver neo4j.Driver
)

type Properties map[string]interface{}

func (p Properties) MustGet(field string) interface{} {
	if val, ok := p[field]; ok {
		return val
	}
	panic(fmt.Sprintf("Properties:MustGet: field %s undefined", field))
}

func (p Properties) MustString(field string) string {
	if value, ok := p.MustGet(field).(string); ok {
		return value
	}

	panic(fmt.Sprintf("Properties:MustString: field %s not of type string", field))
}

func (p Properties) MustBool(field string) bool {
	if value, ok := p.MustGet(field).(bool); ok {
		return value
	}

	panic(fmt.Sprintf("Properties:MustBool: field %s not of type bool", field))
}

func (p Properties) MustInt64(field string) int64 {
	if value, ok := p.MustGet(field).(int64); ok {
		return value
	}

	panic(fmt.Sprintf("Properties:MustInt64: field %s not of type int64", field))
}

func ComposeKey(label string, id int64) string {
	return fmt.Sprintf("%s-%d-%s", label, id, RandStringBytes(4))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
