package shared

import (
	"fmt"

	"github.com/lovoo/goka"
)

type CypherQuery string

func (p CypherQuery) String() string {
	return string(p)
}

type ContextDefinition map[string]CypherQuery

type EntityDescriptor interface {
	HubInputStream() goka.Stream
	HubOutputStream() goka.Stream
	HubGroup() goka.Group
	EventInputStream() goka.Stream
	EventOutputStream() goka.Stream
	EventGroup() goka.Group
	ContextDef() ContextDefinition
	Label() string
}

type BaseDescriptor struct {
	label string
}

func (p *BaseDescriptor) Label() string {
	return p.label
}

func (p *BaseDescriptor) EventGroup() goka.Group {
	return goka.Group(fmt.Sprintf("%s_Input", p.label))
}

func (p *BaseDescriptor) EventInputStream() goka.Stream {
	return goka.Stream(goka.Stream(fmt.Sprintf("Input2%s", p.label)))
}

func (p *BaseDescriptor) EventOutputStream() goka.Stream {
	return HubStream
}

func (p *BaseDescriptor) HubGroup() goka.Group {
	return goka.Group(fmt.Sprintf("%s_Hub", p.label))
}

func (p *BaseDescriptor) HubInputStream() goka.Stream {
	return goka.Stream(goka.Stream(fmt.Sprintf("Hub2%s", p.label)))
}

func (p *BaseDescriptor) HubOutputStream() goka.Stream {
	return HubStream
}

func NewBaseDescriptor(label string) *BaseDescriptor {
	d := &BaseDescriptor{
		label: label,
	}
	return d
}
