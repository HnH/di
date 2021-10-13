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

// bind maps an abstraction to a concrete and sets an instance if it's a singleton binding.
func (c Container) bind(resolver interface{}, name string, singleton bool) (err error) {
	reflectedResolver := reflect.TypeOf(resolver)
	if reflectedResolver.Kind() != reflect.Func {
		return errors.New("container: the resolver must be a function")
	}

	var instances []reflect.Value
	switch {
	case singleton:
		if instances, err = c.invoke(resolver); err != nil {
			return
		}

	case !singleton && reflectedResolver.NumOut() > 2,
		!singleton && reflectedResolver.NumOut() == 2 && !c.isError(reflectedResolver.Out(1)),
		!singleton && reflectedResolver.NumOut() == 1 && c.isError(reflectedResolver.Out(0)):
		return errors.New("container: transient value resolvers must return exactly one value and optionally one error")
	}

	for i := 0; i < reflectedResolver.NumOut(); i++ {
		// we are not interested in returned errors
		if c.isError(reflectedResolver.Out(i)) {
			continue
		}

		if _, exist := c[reflectedResolver.Out(i)]; !exist {
			c[reflectedResolver.Out(i)] = make(map[string]binding)
		}

		if singleton {
			c[reflectedResolver.Out(i)][name] = binding{resolver: resolver, instance: instances[i].Interface()}
		} else {
			c[reflectedResolver.Out(i)][name] = binding{resolver: resolver}
		}
	}

	return nil
}

// Singleton binds an abstraction to concrete for further singleton resolves.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func (c Container) Singleton(resolver interface{}) error {
	return c.bind(resolver, "", true)
}

// NamedSingleton binds like the Singleton method but for named bindings.
func (c Container) NamedSingleton(name string, resolver interface{}) error {
	return c.bind(resolver, name, true)
}

// Transient binds an abstraction to concrete for further transient resolves.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func (c Container) Transient(resolver interface{}) error {
	return c.bind(resolver, "", false)
}

// NamedTransient binds like the Transient method but for named bindings.
func (c Container) NamedTransient(name string, resolver interface{}) error {
	return c.bind(resolver, name, false)
}

// Reset deletes all the existing bindings and empties the container instance.
func (c Container) Reset() {
	for k := range c {
		delete(c, k)
	}
}
