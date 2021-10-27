package di

type Option func(Options)

type Options interface {
	Apply(Option)
}

type NamingOption interface {
	SetName(...string)
}

type ReturnOption interface {
	SetReturn(...interface{})
}

type FillingOption interface {
	SetFill(bool)
}

func WithName(names ...string) Option {
	return func(o Options) {
		if opt, ok := o.(NamingOption); ok {
			opt.SetName(names...)
		}
	}
}

func WithReturn(returns ...interface{}) Option {
	return func(o Options) {
		if opt, ok := o.(ReturnOption); ok {
			opt.SetReturn(returns...)
		}
	}
}

func WithFill() Option {
	return func(o Options) {
		if opt, ok := o.(FillingOption); ok {
			opt.SetFill(true)
		}
	}
}

// options for binding implementations into container
type bindOptions struct {
	factory bool
	fill    bool
	names   []string
}

func newBindOptions(opts []Option) (out bindOptions) {
	for _, o := range opts {
		out.Apply(o)
	}

	return
}

func (self *bindOptions) Apply(opt Option) {
	opt(self)
}

func (self *bindOptions) SetName(names ...string) {
	self.names = names
}

func (self *bindOptions) SetFill(f bool) {
	self.fill = f
}

// options for resolving abstractions
type resolveOptions struct {
	name string
}

func newResolveOptions(opts []Option) (out resolveOptions) {
	out.name = DefaultBindName
	for _, o := range opts {
		out.Apply(o)
	}

	return
}

func (self *resolveOptions) Apply(opt Option) {
	opt(self)
}

func (self *resolveOptions) SetName(names ...string) {
	if len(names) > 0 {
		self.name = names[0]
	}
}

// options for resolving abstractions
type callOptions struct {
	returns []interface{}
}

func newCallOptions(opts []Option) (out callOptions) {
	for _, o := range opts {
		out.Apply(o)
	}

	return
}

func (self *callOptions) Apply(opt Option) {
	opt(self)
}

func (self *callOptions) SetReturn(returns ...interface{}) {
	self.returns = returns
}
