package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Container is responsible for abstraction binding
type Container interface {
	Singleton(constructor interface{}, opts ...Option) error
	Factory(constructor interface{}, opts ...Option) error
	Implementation(implementation interface{}, opts ...Option) error
	ListBindings(reflect.Type) (map[string]Binding, error)
	Reset()
}

// Provider is an abstraction of an entity that provides something to Container
type Provider interface {
	Provide(Container) error
}

// Constructor implements a `Construct()` method which is called either after binding to container in case of singleton or after factory method was called.
type Constructor interface {
	Construct(context.Context) error
}

// NewContainer creates a new instance of the Container
func NewContainer() Container {
	return &container{
		bindings: make(map[reflect.Type]map[string]Binding),
	}
}

type container struct {
	bindings map[reflect.Type]map[string]Binding
	lock     sync.RWMutex
}

// DefaultBindName is the name that is used in containers by default when binding values.
const DefaultBindName = "default"

// Binding holds either singleton instance or factory method for a binding
type Binding struct {
	factory  interface{} // factory method that creates the appropriate implementation of the abstraction
	instance interface{} // instance stored for reusing in singleton bindings
	fill     bool        // call Fill() on a returned instance after it's resolution
}

func (cnt *container) getResolver() *resolver {
	cnt.lock.RLock()
	defer cnt.lock.RUnlock()

	return &resolver{
		containers: []Container{
			cnt,
		},
	}
}

// bind creates a binding for an abstraction
func (cnt *container) bind(constructor interface{}, opts bindOptions) (err error) {
	var ref = reflect.TypeOf(constructor)
	if ref.Kind() != reflect.Func {
		return errors.New("di: the constructor must be a function")
	}

	// if constructor returns no useful values
	if ref.NumOut() == 0 || ref.NumOut() == 1 && isError(ref.Out(0)) {
		return errors.New("di: the constructor must return useful values")
	}

	var numRealInstances = ref.NumOut()
	if isError(ref.Out(numRealInstances - 1)) {
		numRealInstances--
	}

	var instances []reflect.Value
	switch {
	case !opts.factory:
		if numRealInstances > 1 && len(opts.names) > 1 && numRealInstances != len(opts.names) {
			return errors.New("di: the constructor that returns multiple values must be called with either one name or number of names equal to number of values")
		}

		if instances, err = cnt.getResolver().invoke(constructor); err != nil {
			return
		}

		for i := 0; i < numRealInstances; i++ {
			if opts.fill {
				if err = cnt.getResolver().Fill(instances[i].Interface()); err != nil {
					return
				}
			}

			if t, ok := instances[i].Interface().(Constructor); ok {
				if _, err = cnt.getResolver().invoke(t.Construct); err != nil {
					return
				}
			}
		}

	case opts.factory && (ref.NumOut() == 2 && !isError(ref.Out(1)) || ref.NumOut() > 2):
		return errors.New("di: factory resolvers must return exactly one value and optionally one error")
	}

	cnt.lock.Lock()
	defer cnt.lock.Unlock()

	for i := 0; i < numRealInstances; i++ {
		if _, ok := cnt.bindings[ref.Out(i)]; !ok {
			cnt.bindings[ref.Out(i)] = make(map[string]Binding)
		}

		if opts.names == nil {
			opts.names = []string{DefaultBindName}
		}

		var name = opts.names[0]

		// Factory method
		if opts.factory {
			cnt.bindings[ref.Out(i)][name] = Binding{factory: constructor, fill: opts.fill}
			continue
		}

		// Singleton instances
		// if there is more than one instance returned from constructor - use appropriate name for it
		if numRealInstances > 1 {
			if len(opts.names) > 1 {
				name = opts.names[i]
			}

			cnt.bindings[ref.Out(i)][name] = Binding{instance: instances[i].Interface(), fill: opts.fill}
			continue
		}

		// if only one instance is returned from constructor - bind it under all provided names
		for _, name = range opts.names {
			cnt.bindings[ref.Out(i)][name] = Binding{instance: instances[i].Interface(), fill: opts.fill}
		}
	}

	return nil
}

// Singleton binds value(s) returned from constructor as a singleton objects of related types.
func (cnt *container) Singleton(constructor interface{}, opts ...Option) error {
	return cnt.bind(constructor, newBindOptions(opts))
}

// Factory binds constructor as a factory method of related type.
func (cnt *container) Factory(constructor interface{}, opts ...Option) error {
	var options = newBindOptions(opts)
	options.factory = true

	return cnt.bind(constructor, options)
}

// Implementation receives ready instance and binds it to its REAL type, which means that declared abstract variable type (interface) is ignored
func (cnt *container) Implementation(implementation interface{}, opts ...Option) error {
	cnt.lock.RLock()
	defer cnt.lock.RUnlock()

	var ref = reflect.TypeOf(implementation)
	if _, ok := cnt.bindings[ref]; !ok {
		cnt.bindings[ref] = make(map[string]Binding)
	}

	var options = newBindOptions(opts)
	if len(options.names) == 0 {
		options.names = []string{DefaultBindName}
	}

	cnt.bindings[ref][options.names[0]] = Binding{instance: implementation}

	return nil
}

func (cnt *container) ListBindings(abstraction reflect.Type) (map[string]Binding, error) {
	cnt.lock.RLock()
	defer cnt.lock.RUnlock()

	var bnds, ok = cnt.bindings[abstraction]
	if !ok {
		return bnds, fmt.Errorf("di: no binding found for %s", abstraction.String())
	}

	return bnds, nil
}

// Reset deletes all the existing bindings and empties the container instance.
func (cnt *container) Reset() {
	cnt.lock.Lock()
	defer cnt.lock.Unlock()

	for k := range cnt.bindings {
		delete(cnt.bindings, k)
	}
}
