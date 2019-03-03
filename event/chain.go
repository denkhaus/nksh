package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lann/builder"
	"github.com/lovoo/goka"
)

type ActionData struct {
	EntityDescriptor shared.EntityDescriptor
	Operation        shared.Operation
	FieldOperation   shared.Operation
	ErrorHandlers    shared.ErrorHandlers
	Handlers         shared.Handlers
	Conditions       shared.EvalFuncs
	FieldName        string
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

type chain builder.Builder

type Selectable interface {
	OnNodeCreated() Combinable
	OnNodeUpdated() Combinable
	OnNodeDeleted() Combinable
	OnFieldCreated(field string) Combinable
	OnFieldUpdated(field string) Combinable
	OnFieldDeleted(field string) Combinable
	With(fn shared.EvalFunc) Combinable
}

type Combinable interface {
	Or(or ...Combinable) Combinable
	And(or ...Combinable) Combinable
	Not(not ...Combinable) Combinable
}

type Action interface {
	Do(fns ...shared.Handler) Catchable
	LoadEntityContext(entityID int64) Action
}

type Catchable interface {
	Catch(fn shared.ErrorHandler) Executable
}

type Executable interface {
	Execute(ctx goka.Context, m *shared.EventContext) bool
	SetDescriptor(descr shared.EntityDescriptor) Executable
}

type Proceedable interface {
	Then() Action
}

func (b chain) OnNodeCreated() Combinable {
	return builder.Set(b, "Operation", "created").(Combinable)
}

func (b chain) OnNodeUpdated() Combinable {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.UpdatedOperation)
	return builder.Set(c, "FieldName", "*").(Combinable)
}

func (b chain) OnNodeDeleted() Combinable {
	return builder.Set(b, "Operation", shared.DeletedOperation).(Combinable)
}

func (b chain) OnFieldCreated(field string) Combinable {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.CreatedOperation)
	return builder.Set(c, "FieldName", field).(Combinable)
}

func (b chain) OnFieldUpdated(field string) Combinable {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.UpdatedOperation)
	return builder.Set(c, "FieldName", field).(Combinable)
}

func (b chain) OnFieldDeleted(field string) Combinable {
	c := builder.Set(b, "Operation", shared.UpdatedOperation)
	c = builder.Set(c, "FieldOperation", shared.DeletedOperation)
	return builder.Set(c, "FieldName", field).(Combinable)
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

func (b chain) With(fn shared.EvalFunc) Combinable {
	return builder.Append(b, "Conditions", fn).(Combinable)
}

func (b chain) Do(fns ...shared.Handler) Catchable {
	data := []interface{}{}
	for _, fn := range fns {
		data = append(data, fn)
	}
	return builder.Append(b, "Handlers", data...).(Catchable)
}

func (b chain) LoadEntityContext(entityID int64) Action {
	build := func(ctx *shared.HandlerContext) (err error) {
		exec := shared.NewExecutor(ctx)
		ctx.EntityContext, err = exec.BuildEntityContext(entityID)
		if err != nil {
			return errors.Annotate(err, "BuildEntityContext")
		}
		return nil
	}

	return builder.Append(b, "Handlers", build).(Action)
}

func (b chain) Catch(fn shared.ErrorHandler) Executable {
	return builder.Append(b, "ErrorHandlers", fn).(Executable)
}

func (b chain) Then() Action {
	return Action(b)
}

func (b chain) SetDescriptor(descr shared.EntityDescriptor) Executable {
	return builder.Set(b, "EntityDescriptor", descr).(Executable)
}

func (b chain) handleError(err error) {
	if ehs, ok := builder.Get(b, "ErrorHandlers"); ok {
		handlers := ehs.(shared.ErrorHandlers)
		for _, handle := range handlers {
			handle(err)
		}
	} else {
		panic(errors.Annotate(err, "EventChain: no catch handler found"))
	}
}

func (b chain) Execute(ctx goka.Context, m *shared.EventContext) bool {
	data := builder.GetStruct(b).(ActionData)
	if len(data.Handlers) == 0 {
		b.handleError(errors.New("EventChain: no handler defined"))
		return false
	}

	if data.Match(m) {
		hCtx := shared.HandlerContext{
			GokaContext:      ctx,
			EntityDescriptor: data.EntityDescriptor,
			EventContext:     m,
		}

		for _, handle := range data.Handlers {
			if err := handle(&hCtx); err != nil {
				b.handleError(errors.Annotate(err, "HandleEvent"))
				return false
			}
		}

		return true
	}

	return false
}

var actionChain = builder.Register(chain{}, ActionData{})

func If(comb Combinable) Proceedable {
	return comb.(Proceedable)
}
func Do(fns ...shared.Handler) Catchable {
	return actionChain.(Action).Do(fns...)
}
func LoadEntityContext(entityID int64) Action {
	return actionChain.(Action).LoadEntityContext(entityID)
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
func OnFieldCreated(field string) Combinable {
	return actionChain.(Selectable).OnFieldCreated(field)
}
func OnFieldUpdated(field string) Combinable {
	return actionChain.(Selectable).OnFieldUpdated(field)
}
func OnFieldDeleted(field string) Combinable {
	return actionChain.(Selectable).OnFieldDeleted(field)
}
