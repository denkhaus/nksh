package hub

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type ActionData struct {
	EntityDescriptor shared.EntityDescriptor
	Sender           string
	Operation        shared.Operation
	Conditions       shared.EvalFuncs
	ErrorHandlers    shared.ErrorHandlers
	Then             shared.Handlers
	Else             shared.Handlers
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

type Selectable interface {
	From(sender string) Combinable
	With(fn shared.EvalFunc) Combinable
	OnNodeCreated() Combinable
	OnNodeUpdated() Combinable
	OnNodeDeleted() Combinable
}

type Combinable interface {
	Or(or ...Combinable) Combinable
	And(or ...Combinable) Combinable
	Not(not ...Combinable) Combinable
}
type Catchable interface {
	Catch(fn shared.ErrorHandler) Executable
}

type Executable interface {
	Execute(ctx goka.Context, m *shared.HubContext) shared.ChainHandledState
	SetDescriptor(descr shared.EntityDescriptor) Executable
}

type Proceedable interface {
	Then(fns ...shared.Handler) Alternative
}

type Alternative interface {
	Else(fns ...shared.Handler) Catchable
	Catch(fn shared.ErrorHandler) Executable
}

type chain builder.Builder

func (b chain) From(sender string) Combinable {
	return builder.Set(b, "Sender", sender).(Combinable)
}

func (b chain) OnNodeCreated() Combinable {
	return builder.Set(b, "Operation", shared.CreatedOperation).(Combinable)
}

func (b chain) OnNodeUpdated() Combinable {
	return builder.Set(b, "Operation", shared.UpdatedOperation).(Combinable)
}

func (b chain) OnNodeDeleted() Combinable {
	return builder.Set(b, "Operation", shared.DeletedOperation).(Combinable)
}

func (b chain) With(fn shared.EvalFunc) Combinable {
	return builder.Append(b, "Conditions", fn).(Combinable)
}

func (b chain) Or(or ...Combinable) Combinable {
	data := []interface{}{}
	for _, o := range or {
		data = append(data, builder.GetStruct(o))
	}
	return builder.Append(b, "Or", data...).(Combinable)
}

func (b chain) And(and ...Combinable) Combinable {
	data := []interface{}{}
	for _, a := range and {
		data = append(data, builder.GetStruct(a))
	}
	return builder.Append(b, "And", data...).(Combinable)
}

func (b chain) Not(not ...Combinable) Combinable {
	data := []interface{}{}
	for _, n := range not {
		data = append(data, builder.GetStruct(n))
	}
	return builder.Append(b, "Not", data...).(Combinable)
}

func (b chain) SetDescriptor(descr shared.EntityDescriptor) Executable {
	return builder.Set(b, "EntityDescriptor", descr).(Executable)
}

func (b chain) Catch(fn shared.ErrorHandler) Executable {
	return builder.Append(b, "ErrorHandlers", fn).(Executable)
}

func (b chain) Then(fns ...shared.Handler) Alternative {
	data := []interface{}{}
	for _, fn := range fns {
		data = append(data, fn)
	}
	return builder.Append(b, "Then", data...).(Alternative)
}

func (b chain) Else(fns ...shared.Handler) Catchable {
	data := []interface{}{}
	for _, fn := range fns {
		data = append(data, fn)
	}
	return builder.Append(b, "Else", data...).(Catchable)
}

func (b chain) handleError(err error) {
	if ehs, ok := builder.Get(b, "ErrorHandlers"); ok {
		handlers := ehs.(shared.ErrorHandlers)
		for _, handle := range handlers {
			handle(err)
		}
	} else {
		panic(errors.Annotate(err, "HubChain: no catch handler found"))
	}
}

func (b chain) Execute(ctx goka.Context, m *shared.HubContext) shared.ChainHandledState {
	data := builder.GetStruct(b).(ActionData)
	if len(data.Then) == 0 {
		b.handleError(errors.New("HubChain: no handler defined"))
		return shared.ChainHandledStateThenFailed
	}

	hCtx := shared.HandlerContext{
		GokaContext:      ctx,
		EntityDescriptor: data.EntityDescriptor,
		HubContext:       m,
	}

	if data.Match(m) {
		for _, handle := range data.Then {
			if err := handle(&hCtx); err != nil {
				b.handleError(errors.Annotate(err, "HandleEvent [then]"))
				return shared.ChainHandledStateThenFailed
			}
		}

		return shared.ChainHandledStateThen
	}

	if len(data.Else) == 0 {
		return shared.ChainHandledStateUnhandled
	}

	for _, handle := range data.Else {
		if err := handle(&hCtx); err != nil {
			b.handleError(errors.Annotate(err, "HandleEvent [else]"))
			return shared.ChainHandledStateElseFailed
		}
	}

	return shared.ChainHandledStateElse
}

var actionChain = builder.Register(chain{}, ActionData{})

func If(comb Combinable) Proceedable {
	return comb.(Proceedable)
}
func OnNodeCreated() Combinable {
	return actionChain.(Selectable).OnNodeCreated()
}
func OnNodeUpdated() Combinable {
	return actionChain.(Selectable).OnNodeUpdated()
}
func OnNodeDeleted() Combinable {
	return actionChain.(Selectable).OnNodeDeleted()
}
func From(sender string) Combinable {
	return actionChain.(Selectable).From(sender)
}
func With(fn shared.EvalFunc) Combinable {
	return actionChain.(Selectable).With(fn)
}
