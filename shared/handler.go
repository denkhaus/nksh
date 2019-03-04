package shared

import (
	"sync"

	"github.com/lovoo/goka"
)

type HandlerContext struct {
	GokaContext      goka.Context
	EntityDescriptor EntityDescriptor
	EventContext     *EventContext
	HubContext       *HubContext
	store            map[string]interface{}
	mu               sync.Mutex
}

func (p *HandlerContext) Set(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.store[key] = value
}

func (p *HandlerContext) Get(key string) interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	if val, ok := p.store[key]; ok {
		return val
	}

	return nil
}

type Handler func(ctx *HandlerContext) error
type Handlers []Handler
