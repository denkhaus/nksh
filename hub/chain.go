package hub

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type Handler func(ctx goka.Context, m *Context) error

type ActionData struct {
	Operation  shared.Operation
	Sender     string
	Conditions []shared.EvalFunc
	Handlers   []Handler
	Or         []ActionData
	And        []ActionData
	Not        []ActionData
}

func (p *ActionData) Match(m *Context) bool {
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
}

type Stage2 interface {
	OnNodeCreated() Stage3
	OnNodeUpdated() Stage3
	OnNodeDeleted() Stage3
	Or(or ...Stage2) Stage4
	And(or ...Stage2) Stage4
	Not(not ...Stage2) Stage4
}

type Stage3 interface {
	With(fn shared.EvalFunc) Stage4
	Then(fn Handler) Action
}

type Stage4 interface {
	Stage2
	Then(fn Handler) Action
}

type Action interface {
	Then(fn Handler) Action
	ApplyMessage(ctx goka.Context, m *Context) (bool, error)
}

type chain builder.Builder

func (b chain) From(sender string) Stage2 {
	return builder.Set(b, "Sender", sender).(Stage2)
}

func (b chain) OnNodeCreated() Stage3 {
	return builder.Set(b, "Operation", shared.CreatedOperation).(Stage3)
}

func (b chain) OnNodeUpdated() Stage3 {
	return builder.Set(b, "Operation", shared.UpdatedOperation).(Stage3)
}

func (b chain) OnNodeDeleted() Stage3 {
	return builder.Set(b, "Operation", shared.DeletedOperation).(Stage3)
}

func (b chain) With(fn shared.EvalFunc) Stage4 {
	return builder.Append(b, "Conditions", fn).(Stage4)
}

func (b chain) Or(or ...Stage2) Stage4 {
	data := []interface{}{}
	for _, o := range or {
		data = append(data, builder.GetStruct(o))
	}
	return builder.Append(b, "Or", data...).(Stage4)
}

func (b chain) And(and ...Stage2) Stage4 {
	data := []interface{}{}
	for _, a := range and {
		data = append(data, builder.GetStruct(a))
	}
	return builder.Append(b, "And", data...).(Stage4)
}

func (b chain) Not(not ...Stage2) Stage4 {
	data := []interface{}{}
	for _, n := range not {
		data = append(data, builder.GetStruct(n))
	}
	return builder.Append(b, "Not", data...).(Stage4)
}

func (b chain) Then(fn Handler) Action {
	return builder.Append(b, "Handlers", fn).(Action)
}

func (b chain) ApplyMessage(ctx goka.Context, m *Context) (bool, error) {
	data := builder.GetStruct(b).(ActionData)
	if len(data.Handlers) == 0 {
		return false, errors.New("HubChain: no handler defined")
	}

	if data.Match(m) {
		for _, handle := range data.Handlers {
			if err := handle(ctx, m); err != nil {
				return false, errors.Annotate(err, "HandleEvent")
			}
		}

		return true, nil
	}

	return false, nil
}

var Chain = builder.Register(chain{}, ActionData{}).(Stage1)
