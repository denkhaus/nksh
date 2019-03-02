package hub

import (
	"testing"

	"github.com/denkhaus/nksh/shared"
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

func TestChain(t *testing.T) {

	codec := shared.HubContextCodec{}
	m, err := codec.Decode([]byte(update))
	assert.NoError(t, err, "decode raw message")

	ctx, ok := m.(*shared.HubContext)
	assert.True(t, ok)

	handlerTriggered := 0
	onUpdated := Chain.OnNodeUpdated()

	var subordinates = Chain.From("Person").Or(
		Chain.From("PersonPosition"),
		Chain.From("Photo"),
	)

	subordinatesUpdated := onUpdated.And(subordinates)

	condition := subordinatesUpdated.Then(
		func(ctx goka.Context, descr shared.EntityDescriptor, m *shared.HubContext) error {
			handlerTriggered++
			return nil
		})

	hit, err := condition.applyMessage(nil, ctx)
	assert.NoError(t, err, "apply message")
	assert.Equal(t, 1, handlerTriggered, "handler triggered")

	assert.True(t, hit, "condition hit")

	handlerTriggered = 0
	var onSubordinatesInvisibled = subordinatesUpdated.With(IsNodeInvisible).
		Then(func(ctx goka.Context, descr shared.EntityDescriptor, m *shared.HubContext) error {
			handlerTriggered++
			return nil
		})

	hit, err = onSubordinatesInvisibled.applyMessage(nil, ctx)
	assert.NoError(t, err, "apply message")
	assert.Equal(t, 1, handlerTriggered, "handler triggered")
	assert.True(t, hit, "condition hit")
}
