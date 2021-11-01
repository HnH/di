package di

// Option represents single option type
type Option func(Options)

// Options represents a target for applying an Option
type Options interface {
	Apply(Option)
}

// NamingOption supports setting a name
type NamingOption interface {
	SetName(...string)
}

// ReturnOption supports setting a return target
type ReturnOption interface {
	SetReturn(...interface{})
}

// FillingOption supports setting a fill flag
type FillingOption interface {
	SetFill(bool)
}

// WithName returns a NamingOption
func WithName(names ...string) Option {
	return func(o Options) {
		if opt, ok := o.(NamingOption); ok {
			opt.SetName(names...)
		}
	}
}

// WithReturn returns a ReturnOption
func WithReturn(returns ...interface{}) Option {
	return func(o Options) {
		if opt, ok := o.(ReturnOption); ok {
			opt.SetReturn(returns...)
		}
	}
}

// WithFill returns a FillingOption
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

// Apply implements Options interface
func (self *bindOptions) Apply(opt Option) {
	opt(self)
}

// SetName implements NamingOption interface
func (self *bindOptions) SetName(names ...string) {
	self.names = names
}

// SetFill implements FillingOption interface
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

// Apply implements Options interface
func (self *resolveOptions) Apply(opt Option) {
	opt(self)
}

// SetName implements NamingOption interface
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

// Apply implements Options interface
func (self *callOptions) Apply(opt Option) {
	opt(self)
}

// SetReturn implements ReturnOption interface
func (self *callOptions) SetReturn(returns ...interface{}) {
	self.returns = returns
}
