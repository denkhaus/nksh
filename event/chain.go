package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type HandlerContext struct {
	GokaContext      goka.Context
	EventContext     *shared.EventContext
	EntityDescriptor shared.EntityDescriptor
}

type Handler func(ctx *HandlerContext) error

type ActionData struct {
	EntityDescriptor shared.EntityDescriptor
	Operation        shared.Operation
	FieldOperation   shared.Operation
	FieldName        string
	Handlers         []Handler
	Conditions       []shared.EvalFunc
	Or               []ActionData
	And              []ActionData
	Not              []ActionData
}

func (p *ActionData) Match(m *shared.EventContext) bool {
	result := m.Match(
		p.Operation,
		p.FieldName,
		p.FieldOperation,
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
	OnNodeCreated() Stage2
	OnNodeUpdated() Stage2
	OnNodeDeleted() Stage2
	OnFieldCreated(field string) Stage2
	OnFieldUpdated(field string) Stage2
	OnFieldDeleted(field string) Stage2
}

type Stage2 interface {
	Or(or ...Stage2) Stage2
	And(or ...Stage2) Stage2
	Not(not ...Stage2) Stage2
	With(fn shared.EvalFunc) Stage2
	Then(fn Handler) Action
}

type Action interface {
	Then(fn Handler) Action
	applyContext(ctx goka.Context, m *shared.EventContext) (bool, error)
	setDescriptor(descr shared.EntityDescriptor) Action
}

type chain builder.Builder

func (b chain) OnNodeCreated() Stage2 {
	return builder.Set(b, "Operation", "created").(Stage2)
}

func (b chain) OnNodeUpdated() Stage2 {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.UpdatedOperation)
	return builder.Set(c, "FieldName", "*").(Stage2)
}

func (b chain) OnNodeDeleted() Stage2 {
	return builder.Set(b, "Operation", shared.DeletedOperation).(Stage2)
}

func (b chain) OnFieldCreated(field string) Stage2 {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.CreatedOperation)
	return builder.Set(c, "FieldName", field).(Stage2)
}

func (b chain) OnFieldUpdated(field string) Stage2 {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.UpdatedOperation)
	return builder.Set(c, "FieldName", field).(Stage2)
}

func (b chain) OnFieldDeleted(field string) Stage2 {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.DeletedOperation)
	return builder.Set(c, "FieldName", field).(Stage2)
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

func (b chain) With(fn shared.EvalFunc) Stage2 {
	return builder.Append(b, "Conditions", fn).(Stage2)
}

func (b chain) Then(fn Handler) Action {
	return builder.Append(b, "Handlers", fn).(Action)
}

func (b chain) setDescriptor(descr shared.EntityDescriptor) Action {
	return builder.Set(b, "EntityDescriptor", descr).(Action)
}

func (b chain) applyContext(ctx goka.Context, m *shared.EventContext) (bool, error) {
	data := builder.GetStruct(b).(ActionData)
	if len(data.Handlers) == 0 {
		return false, errors.New("EventChain: no handler defined")
	}

	if data.Match(m) {
		hCtx := HandlerContext{
			GokaContext:      ctx,
			EntityDescriptor: data.EntityDescriptor,
			EventContext:     m,
		}

		for _, handle := range data.Handlers {
			if err := handle(&hCtx); err != nil {
				return false, errors.Annotate(err, "HandleEvent")
			}
		}

		return true, nil
	}

	return false, nil
}

var Chain = builder.Register(chain{}, ActionData{}).(Stage1)
