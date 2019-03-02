package hub

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type Handler func(ctx goka.Context, descr shared.EntityDescriptor, m *shared.HubContext) error

type ActionData struct {
	EntityDescriptor shared.EntityDescriptor
	Sender           string
	Operation        shared.Operation
	Conditions       shared.EvalFuncs
	Handlers         []Handler
	Or               []ActionData
	And              []ActionData
	Not              []ActionData
}

func (p *ActionData) Match(m *shared.HubContext) bool {
	result := m.Match(
		p.Operation,
		p.Sender,
		p.Conditions,
	)
	for _, data := range p.Or {
		result = result || data.Match(m)
	}
	for _, data := range p.And {
		result = result && data.Match(m)
	}
	for _, data := range p.Not {
		result = result && !data.Match(m)
	}
	return result
}

type Stage1 interface {
	From(sender string) Stage2
	OnNodeCreated() Stage2
	OnNodeUpdated() Stage2
	OnNodeDeleted() Stage2
}

type Stage2 interface {
	Or(or ...Stage2) Stage2
	And(and ...Stage2) Stage2
	Not(not ...Stage2) Stage2
	With(fn shared.EvalFunc) Stage2
	Then(fn Handler) Action
}

type Action interface {
	Then(fn Handler) Action
	applyMessage(ctx goka.Context, m *shared.HubContext) (bool, error)
	setDescriptor(shared.EntityDescriptor) Action
}

type chain builder.Builder

func (b chain) From(sender string) Stage2 {
	return builder.Set(b, "Sender", sender).(Stage2)
}

func (b chain) OnNodeCreated() Stage2 {
	return builder.Set(b, "Operation", shared.CreatedOperation).(Stage2)
}

func (b chain) OnNodeUpdated() Stage2 {
	return builder.Set(b, "Operation", shared.UpdatedOperation).(Stage2)
}

func (b chain) OnNodeDeleted() Stage2 {
	return builder.Set(b, "Operation", shared.DeletedOperation).(Stage2)
}

func (b chain) With(fn shared.EvalFunc) Stage2 {
	return builder.Append(b, "Conditions", fn).(Stage2)
}

func (b chain) Or(or ...Stage2) Stage2 {
	data := []interface{}{}
	for _, o := range or {
		data = append(data, builder.GetStruct(o))
	}
	return builder.Append(b, "Or", data...).(Stage2)
}

func (b chain) And(and ...Stage2) Stage2 {
	data := []interface{}{}
	for _, a := range and {
		data = append(data, builder.GetStruct(a))
	}
	return builder.Append(b, "And", data...).(Stage2)
}

func (b chain) Not(not ...Stage2) Stage2 {
	data := []interface{}{}
	for _, n := range not {
		data = append(data, builder.GetStruct(n))
	}
	return builder.Append(b, "Not", data...).(Stage2)
}

func (b chain) Then(fn Handler) Action {
	return builder.Append(b, "Handlers", fn).(Action)
}

func (b chain) setDescriptor(descr shared.EntityDescriptor) Action {
	return builder.Set(b, "EntityDescriptor", descr).(Action)
}

func (b chain) applyContext(ctx goka.Context, m *shared.HubContext) (bool, error) {
	data := builder.GetStruct(b).(ActionData)
	if len(data.Handlers) == 0 {
		return false, errors.New("HubChain: no handler defined")
	}

	if data.Match(m) {
		for _, handle := range data.Handlers {
			if err := handle(ctx, data.EntityDescriptor, m); err != nil {
				return false, errors.Annotate(err, "HandleEvent")
			}
		}

		return true, nil
	}

	return false, nil
}

var Chain = builder.Register(chain{}, ActionData{}).(Stage1)
