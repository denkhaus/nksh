package event

import (
	"testing"

	"github.com/denkhaus/nksh/shared"

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
			"street":"Jordan ave",
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

	var handledError error
	handlerTriggered := 0
	condition := Chain.OnNodeUpdated().And(
		Chain.OnFieldUpdated("first_name"),
		Chain.OnFieldCreated("geo"),
		Chain.OnFieldDeleted("street"),
	).Not(
		Chain.OnFieldUpdated("email"),
	).Then().Do(func(_ *shared.HandlerContext) error {
		handlerTriggered++
		return nil
	}).Catch(func(err error) {
		handledError = err
	})

	hit := condition.Execute(nil, ctx)
	assert.NoError(t, handledError, "handled error")
	assert.Equal(t, 1, handlerTriggered, "handler triggered")
	assert.True(t, hit, "condition hit")
}
