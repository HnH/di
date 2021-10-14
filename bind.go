package di

import (
	"errors"
	"fmt"
	"reflect"
)

const defaultBindName = "default"

// binding holds a binding resolver and an instance (for singleton bindings).
type binding struct {
	factory  interface{} // factory method that creates the appropriate implementation of the abstraction
	instance interface{} // instance stored for reusing in singleton bindings
}

func (c Container) getAllBindings(abstraction reflect.Type) (bnds map[string]binding, err error) {
	var ok bool
	if bnds, ok = c[abstraction]; !ok {
		return bnds, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	return
}

func (c Container) getBinding(abstraction reflect.Type, name string) (bnd binding, err error) {
	var ok bool
	if _, ok = c[abstraction]; !ok {
		return bnd, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	if bnd, ok = c[abstraction][name]; !ok {
		return bnd, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	return
}

func (c Container) getResolver() *Resolver {
	return &Resolver{
		list: []Container{
			c,
		},
	}
}

// bind creates a binding for an abstraction
func (c Container) bind(resolver interface{}, opts bindOptions) (err error) {
	var ref = reflect.TypeOf(resolver)
	if ref.Kind() != reflect.Func {
		return errors.New("di: the resolver must be a function")
	}

	// if resolver returns no useful values
	if ref.NumOut() == 0 || ref.NumOut() == 1 && isError(ref.Out(0)) {
		return errors.New("di: the resolver must return useful values")
	}

	var numRealInstances = ref.NumOut()
	if isError(ref.Out(numRealInstances - 1)) {
		numRealInstances--
	}

	var instances []reflect.Value
	switch {
	case !opts.factory:
		if numRealInstances > 1 && len(opts.names) > 1 && numRealInstances != len(opts.names) {
			return errors.New("di: the resolver that returns multiple values must be called with either one name or number of names equal to number of values")
		}

		if instances, err = c.getResolver().invoke(resolver); err != nil {
			return
		}

	case opts.factory && (ref.NumOut() == 2 && !isError(ref.Out(1)) || ref.NumOut() > 2):
		return errors.New("di: factory resolvers must return exactly one value and optionally one error")
	}

	for i := 0; i < numRealInstances; i++ {
		// we are not interested in returned errors
		if isError(ref.Out(i)) {
			continue
		}

		if _, ok := c[ref.Out(i)]; !ok {
			c[ref.Out(i)] = make(map[string]binding)
		}

		if opts.names == nil {
			opts.names = []string{defaultBindName}
		}

		var name = opts.names[0]

		// Factory method
		if opts.factory {
			c[ref.Out(i)][name] = binding{factory: resolver}
			continue
		}

		// Singleton instances
		// if there is more than one instance returned from resolver - use appropriate name for it
		if numRealInstances > 1 {
			if len(opts.names) > 1 {
				name = opts.names[i]
			}

			c[ref.Out(i)][name] = binding{instance: instances[i].Interface()}
			continue
		}

		// if only one instance is returned from resolver - bind it under all provided names
		for _, name = range opts.names {
			c[ref.Out(i)][name] = binding{instance: instances[i].Interface()}
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
