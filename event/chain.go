package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type Handler func(ctx goka.Context, m *Context) error

type ActionData struct {
	Operation      shared.Operation
	FieldOperation shared.Operation
	FieldName      string
	HandleEvent    Handler
	Or             []ActionData
	And            []ActionData
	Not            []ActionData
}

func (p *ActionData) Match(m *Context) bool {
	result := m.Match(
		p.Operation,
		p.FieldName,
		p.FieldOperation,
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
	Do(fn Handler) Action
}

type Action interface {
	ApplyMessage(ctx goka.Context, m *Context) (bool, error)
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

func (b chain) Do(fn Handler) Action {
	return builder.Set(b, "HandleEvent", fn).(Action)
}

func (b chain) ApplyMessage(ctx goka.Context, m *Context) (bool, error) {
	data := builder.GetStruct(b).(ActionData)
	if data.HandleEvent == nil {
		return false, errors.New("EventChain: handler func undefined")
	}
	
	if data.Match(m) {
		if err := data.HandleEvent(ctx, m); err != nil {
			return false, errors.Annotate(err, "HandleEvent")
		}
		return true, nil
	}

	return false, nil
}

var Chain = builder.Register(chain{}, ActionData{}).(Stage1)
