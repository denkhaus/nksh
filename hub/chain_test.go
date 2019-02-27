package hub

import (
	"testing"

	"github.com/lovoo/goka"
	"github.com/stretchr/testify/assert"
)

var update = `
{	
	  "sender": "Photo",
	  "sender_id": 2336,
	  "receiver": "PersonConnector",
	  "receiver_id": 4587,
	  "operation": "updated", 
	  "properties":{
		  "name": "denkhaus",
		  "visible": false
	  }	
}`

func nodeInvisible(arg interface{}) bool {
	m:= arg.(Context)
	return m.Properties.MustBool("visible") == false
}

func TestChain(t *testing.T) {

	codec := ContextCodec{}

	m, err := codec.Decode([]byte(update))
	assert.NoError(t, err, "decode raw message")

	ctx, ok := m.(*Context)
	assert.True(t, ok)

	handlerTriggered := 0
	onUpdated := Chain.OnNodeUpdated()

	var subordinates = Chain.From("Person").Or(
		Chain.From("PersonPosition"),
		Chain.From("Photo"),
	)

	subordinatesUpdated := onUpdated.And(subordinates)

	condition := subordinatesUpdated.Then(
		func(ctx goka.Context, m *Context) error {
			handlerTriggered++
			return nil
		})

	hit, err := condition.ApplyMessage(nil, ctx)
	assert.NoError(t, err, "apply message")
	assert.Equal(t, 1, handlerTriggered, "handler triggered")

	assert.True(t, hit, "condition hit")

	handlerTriggered = 0
	var onSubordinatesInvisibled = subordinatesUpdated.With(nodeInvisible).
		Then(func(ctx goka.Context, m *Context) error {
			handlerTriggered++
			return nil
		})

	hit, err = onSubordinatesInvisibled.ApplyMessage(nil, ctx)
	assert.NoError(t, err, "apply message")
	assert.Equal(t, 1, handlerTriggered, "handler triggered")
	assert.True(t, hit, "condition hit")
}
