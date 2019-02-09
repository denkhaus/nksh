package hub

import (
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type ConditionFunc func(c Context) bool
type Handler func(ctx goka.Context, m *Context) error

type ActionData struct {
	Operation   string
	Sender      string
	Condition   ConditionFunc
	Or          []ActionData
	And         []ActionData
	Not         []ActionData
	HandleEvent Handler
}

func (p *ActionData) Match(m *Context) bool {
	result := m.Match(
		p.Operation,
		p.Sender,
		p.Condition,
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
}

type Stage3 interface {
	With(fn ConditionFunc) Stage4
	Do(fn Handler) Action
}

type Stage4 interface {
	Or(or ...Stage2) Stage4
	And(or ...Stage2) Stage4
	Not(not ...Stage2) Stage4
	Do(fn Handler) Action
}

type Action interface {
	ApplyMessage(ctx goka.Context, m *Context) error
}

type chain builder.Builder

func (b chain) From(sender string) Stage2 {
	return builder.Set(b, "Sender", sender).(Stage2)
}

func (b chain) OnNodeCreated() Stage3 {
	return builder.Set(b, "Operation", "created").(Stage3)
}

func (b chain) OnNodeUpdated() Stage3 {
	return builder.Set(b, "Operation", "updated").(Stage3)
}

func (b chain) OnNodeDeleted() Stage3 {
	return builder.Set(b, "Operation", "deleted").(Stage3)
}

func (b chain) With(fn ConditionFunc) Stage4 {
	return builder.Set(b, "Condition", fn).(Stage4)
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

func (b chain) ApplyMessage(ctx goka.Context, m *Context) error {
	data := builder.GetStruct(b).(ActionData)
	if data.Match(m) {
		if data.HandleEvent == nil {
			return errors.New("EventChain: handler func undefined")
		}
		if err := data.HandleEvent(ctx, m); err != nil {
			return errors.Annotate(err, "HandleEvent")
		}
	}

	return nil
}

var Chain = builder.Register(chain{}, ActionData{}).(Stage1)
