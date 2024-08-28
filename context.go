package di

import (
	"context"
	"fmt"
)

// Context describe DI context propagator capabilities
type Context interface {
	SetContainer(Container) Context
	Container() Container
	SetResolver(Resolver) Context
	Resolver() Resolver
	Visualize() []string
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

func (self *ctx) Visualize() []string {
	var out = make([]string, 0, 100)

	var r, ok = self.Resolver().(*resolver)
	if !ok {
		return out
	}

	out = append(out, fmt.Sprintf("resolver has [%d] containers", len(r.containers)))
	for i, c := range r.containers {
		var cnt *container
		if cnt, ok = c.(*container); !ok {
			continue
		}

		out = append(out, fmt.Sprintf("  -> container [%d] has [%d] type binding(s)", i, len(cnt.bindings)))

		for t, bindingList := range cnt.bindings {
			out = append(out, fmt.Sprintf("    -> [%s] has [%d] binding(s)", t.String(), len(bindingList)))

			for name, binding := range bindingList {
				out = append(out, fmt.Sprintf("     • [%s] %s declared at [%s]", name, func() string {
					if binding.factory != nil {
						return "factory"
					}

					return "instance"
				}(), binding.caller))
			}
		}
	}

	return out
}

// Raw returns raw context.Context
func (self *ctx) Raw() context.Context {
	return self.Context
}
