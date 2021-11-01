package di

import "context"

// Context describe DI context propagator capabilities
type Context interface {
	Put(Container) Context
	Container() Container
	Resolver() Resolver
	Raw() context.Context
}

// Ctx returns context propagator
func Ctx(c context.Context) Context {
	return &ctx{
		ctx: c,
	}
}

const contextKey = "di.ctx"

type ctx struct {
	ctx context.Context
}

// Put sets container in context
func (self *ctx) Put(c Container) Context {
	self.ctx = context.WithValue(self.ctx, contextKey, c)
	return self
}

// Container returns container from context or creates a new one
func (self *ctx) Container() Container {
	if c, has := self.ctx.Value(contextKey).(Container); has {
		return c
	}

	return NewContainer()
}

// Resolver returns a resolver instance against a Container() output
func (self *ctx) Resolver() Resolver {
	return NewResolver(self.Container())
}

// Raw returns raw context.Context
func (self *ctx) Raw() context.Context {
	return self.ctx
}
