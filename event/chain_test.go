package event

import (
	"testing"

	"github.com/lovoo/goka"
	"github.com/stretchr/testify/assert"
)

var update = `
{
	"meta": {
	  "timestamp": 1532597182604,
	  "username": "neo4j",
	  "tx_id": 3,
	  "tx_event_id": 0,
	  "tx_events_count": 2,
	  "operation": "updated",
	  "source": {
		"hostname": "neo4j.mycompany.com"
	  }
	},
	"payload": {
	  "id": "1004",
	  "type": "node",
	  "before": {
		"labels": ["Person", "Tmp"],
		"properties": {
		  "email": "annek@noanswer.org",
		  "last_name": "Kretchmar",
		  "first_name": "Anne"
		}
	  },
	  "after": {
		"labels": ["Person"],
		"properties": {
		  "last_name": "Kretchmar",
		  "email": "annek@noanswer.org",
		  "first_name": "Anne Marie",
		  "geo":[0.123, 46.2222, 32.11111]
		}
	  }
	}
  }
`

func TestChain(t *testing.T) {

	codec := Neo4jMessageCodec{}

	m, err := codec.Decode([]byte(update))
	assert.NoError(t, err, "decode raw message")

	msg, ok := m.(*Neo4jMessage)
	assert.True(t, ok)

	ctx, err := msg.ToContext()
	assert.NoError(t, err, "create context")

	condition := Chain.OnFieldUpdated("first_name").Do(func(ctx goka.Context, m *Context) error {
		return nil
	})

	hit, err := condition.ApplyMessage(nil, ctx)
	assert.NoError(t, err, "apply message")
	assert.True(t, hit, "condition hit")
}
