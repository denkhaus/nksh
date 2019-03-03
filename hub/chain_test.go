package hub

import (
	"testing"

	"github.com/denkhaus/nksh/shared"
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

	var handledError error
	thenTriggered := 0
	elseTriggered := 0

	onUpdated := OnNodeUpdated()

	var subordinates = From("Person").Or(
		From("PersonPosition"),
		From("Photo"),
	)

	subordinatesUpdated := onUpdated.And(subordinates)

	condition := If(
		subordinatesUpdated,
	).Then(func(ctx *shared.HandlerContext) error {
		thenTriggered++
		return nil
	}).Else(func(_ *shared.HandlerContext) error {
		elseTriggered++
		return nil
	}).Catch(func(err error) {
		handledError = err
	})

	state := condition.Execute(nil, ctx)
	assert.NoError(t, handledError, "handled error")
	assert.Equal(t, 1, thenTriggered, "then triggered")
	assert.Equal(t, 0, elseTriggered, "else triggered")
	assert.Equal(t, shared.ChainHandledStateThen, state, "condition hit")

	thenTriggered = 0
	var onSubordinatesInvisibled = If(
		subordinatesUpdated.And(
			With(IsNodeInvisible),
		),
	).Then(func(ctx *shared.HandlerContext) error {
		thenTriggered++
		return nil
	}).Catch(func(err error) {
		handledError = err
	})

	state = onSubordinatesInvisibled.Execute(nil, ctx)
	assert.NoError(t, handledError, "handled error")
	assert.Equal(t, 1, thenTriggered, "then triggered")
	assert.Equal(t, shared.ChainHandledStateThen, state, "condition hit")
}
