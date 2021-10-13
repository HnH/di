package di

type Option func(Options)

type Options interface {
	Apply(Option)
}

type NamingOption interface {
	SetName(...string)
}

func WithName(names ...string) Option {
	return func(o Options) {
		if opt, ok := o.(NamingOption); ok {
			opt.SetName(names...)
		}
	}
}

// options for binding instances into container
type bindOptions struct {
	factory bool
	names   []string
}

func newBindOptions(opts []Option) (out bindOptions) {
	out.names = []string{""}
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

// options for resolving abstractions
type resolveOptions struct {
	name string
}

func newResolveOptions(opts []Option) (out resolveOptions) {
	for _, o := range opts {
		out.Apply(o)
	}

	return
}

func (self *resolveOptions) Apply(opt Option) {
	opt(self)
}

func (self *resolveOptions) SetName(names ...string) {
	self.name = names[0]
}
