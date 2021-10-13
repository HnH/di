package di

import (
	"errors"
	"reflect"
)

// binding holds a binding resolver and an instance (for singleton bindings).
type binding struct {
	resolver interface{} // resolver function that creates the appropriate implementation of the related abstraction
	instance interface{} // instance stored for reusing in singleton bindings
}

// resolve creates an appropriate implementation of the related abstraction
func (b binding) resolve(c Container) (interface{}, error) {
	if b.instance != nil {
		return b.instance, nil
	}

	var out, err = c.invoke(b.resolver)
	if err != nil {
		return nil, err
	}

	return out[0].Interface(), nil
}

// bind creates a binding for an abstraction
func (c Container) bind(resolver interface{}, opts bindOptions) (err error) {
	var reflectedResolver = reflect.TypeOf(resolver)
	if reflectedResolver.Kind() != reflect.Func {
		return errors.New("container: the resolver must be a function")
	}

	var instances []reflect.Value
	switch {
	case !opts.factory:
		if instances, err = c.invoke(resolver); err != nil {
			return
		}

	case opts.factory && reflectedResolver.NumOut() > 2,
		opts.factory && reflectedResolver.NumOut() == 2 && !c.isError(reflectedResolver.Out(1)),
		opts.factory && reflectedResolver.NumOut() == 1 && c.isError(reflectedResolver.Out(0)):
		return errors.New("container: transient value resolvers must return exactly one value and optionally one error")
	}

	for i := 0; i < reflectedResolver.NumOut(); i++ {
		// we are not interested in returned errors
		if c.isError(reflectedResolver.Out(i)) {
			continue
		}

		if _, ok := c[reflectedResolver.Out(i)]; !ok {
			c[reflectedResolver.Out(i)] = make(map[string]binding)
		}

		// TODO: match name and returned instances cound
		if opts.factory {
			c[reflectedResolver.Out(i)][opts.names[0]] = binding{resolver: resolver}
		} else {
			c[reflectedResolver.Out(i)][opts.names[0]] = binding{resolver: resolver, instance: instances[i].Interface()}
		}
	}

	return nil
}

// Singleton binds an abstraction to concrete for further singleton resolves.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func (c Container) Singleton(resolver interface{}, opts ...Option) error {
	return c.bind(resolver, newBindOptions(opts))
}

// Factory binds an abstraction to concrete for further transient resolves.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func (c Container) Factory(resolver interface{}, opts ...Option) error {
	var options = newBindOptions(opts)
	options.factory = true

	return c.bind(resolver, options)
}

// Reset deletes all the existing bindings and empties the container instance.
func (c Container) Reset() {
	for k := range c {
		delete(c, k)
	}
}
