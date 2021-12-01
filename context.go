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

const (
	ctxKeyContainer = "di.ctx.container"
	ctxKeyResolver  = "di.ctx.resolver"
)

type ctx struct {
	context.Context
}

// SetContainer puts container to a context
func (ct *ctx) SetContainer(c Container) Context {
	ct.Context = context.WithValue(ct.Context, ctxKeyContainer, c)
	return ct
}

// Container returns container from context or returns a global container
func (ct *ctx) Container() Container {
	if c, has := ct.Context.Value(ctxKeyContainer).(Container); has {
		return c
	}

	return globalContext.Container()
}

// SetResolver puts container to a context
func (ct *ctx) SetResolver(r Resolver) Context {
	ct.Context = context.WithValue(ct.Context, ctxKeyResolver, r)
	return ct
}

// Resolver returns a resolver instance either preset or against a Container() output
func (ct *ctx) Resolver() Resolver {
	if r, has := ct.Context.Value(ctxKeyResolver).(Resolver); has {
		return r
	}

	return NewResolver(ct.Container())
}

// Raw returns raw context.Context
func (ct *ctx) Raw() context.Context {
	return ct.Context
}
