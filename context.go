package di

import "context"

// Context describe DI context propagator capabilities
type Context interface {
	SetContainer(Container) Context
	Container() Container
	SetResolver(Resolver) Context
	Resolver() Resolver
	Raw() context.Context
}

// Ctx returns context propagator
func Ctx(ctxt context.Context) Context {
	return &ctx{
		Context: ctxt,
	}
}

type ctxKey string

const (
	ctxKeyContainer ctxKey = "di.ctx.container"
	ctxKeyResolver  ctxKey = "di.ctx.resolver"
)

type ctx struct {
	context.Context
}

// SetContainer puts container to a context
func (self *ctx) SetContainer(c Container) Context {
	self.Context = context.WithValue(self.Context, ctxKeyContainer, c)
	return self
}

// Container returns container from context or returns a global container
func (self *ctx) Container() Container {
	if c, has := self.Context.Value(ctxKeyContainer).(Container); has {
		return c
	}

	return globalContext.Container()
}

// SetResolver puts container to a context
func (self *ctx) SetResolver(r Resolver) Context {
	self.Context = context.WithValue(self.Context, ctxKeyResolver, r)
	return self
}

// Resolver returns a resolver instance either preset or against a Container() output
func (self *ctx) Resolver() Resolver {
	if r, has := self.Context.Value(ctxKeyResolver).(Resolver); has {
		return r
	}

	return NewResolver(self.Container())
}

// Raw returns raw context.Context
func (self *ctx) Raw() context.Context {
	return self.Context
}
