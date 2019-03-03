package shared

import "github.com/lovoo/goka"

type HandlerContext struct {
	GokaContext      goka.Context
	EventContext     *EventContext
	EntityContext    *EntityContext
	HubContext       *HubContext
	EntityDescriptor EntityDescriptor
}

type Handler func(ctx *HandlerContext) error
type Handlers []Handler
