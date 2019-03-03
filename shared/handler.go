package shared

import "github.com/lovoo/goka"

type HandlerContext struct {
	GokaContext      goka.Context
	EntityDescriptor EntityDescriptor
	EventContext     *EventContext
	EntityContext    *EntityContext
	HubContext       *HubContext
}

type Handler func(ctx *HandlerContext) error
type Handlers []Handler
