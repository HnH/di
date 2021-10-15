package di

import (
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
}

func (self *container) getResolver() *resolver {
	self.lock.RLock()
	defer self.lock.RUnlock()

	return &resolver{
		containers: []Container{
			self,
		},
	}
}

// bind creates a binding for an abstraction
func (self *container) bind(constructor interface{}, opts bindOptions) (err error) {
	var ref = reflect.TypeOf(constructor)
	if ref.Kind() != reflect.Func {
		return errors.New("di: the constructor must be a function")
	}

	// if constructor returns no useful values
	if ref.NumOut() == 0 || ref.NumOut() == 1 && isError(ref.Out(0)) {
		return errors.New("di: the constructor must return useful values")
	}

	var numRealImplementations = ref.NumOut()
	if isError(ref.Out(numRealImplementations - 1)) {
		numRealImplementations--
	}

	var implementations []reflect.Value
	switch {
	case !opts.factory:
		if numRealImplementations > 1 && len(opts.names) > 1 && numRealImplementations != len(opts.names) {
			return errors.New("di: the constructor that returns multiple values must be called with either one name or number of names equal to number of values")
		}

		if implementations, err = self.getResolver().invoke(constructor); err != nil {
			return
		}

	case opts.factory && (ref.NumOut() == 2 && !isError(ref.Out(1)) || ref.NumOut() > 2):
		return errors.New("di: factory resolvers must return exactly one value and optionally one error")
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	for i := 0; i < numRealImplementations; i++ {
		// we are not interested in returned errors
		if isError(ref.Out(i)) {
			continue
		}

		if _, ok := self.bindings[ref.Out(i)]; !ok {
			self.bindings[ref.Out(i)] = make(map[string]Binding)
		}

		if opts.names == nil {
			opts.names = []string{DefaultBindName}
		}

		var name = opts.names[0]

		// Factory method
		if opts.factory {
			self.bindings[ref.Out(i)][name] = Binding{factory: constructor}
			continue
		}

		// Singleton instances
		// if there is more than one instance returned from constructor - use appropriate name for it
		if numRealImplementations > 1 {
			if len(opts.names) > 1 {
				name = opts.names[i]
			}

			self.bindings[ref.Out(i)][name] = Binding{instance: implementations[i].Interface()}
			continue
		}

		// if only one instance is returned from constructor - bind it under all provided names
		for _, name = range opts.names {
			self.bindings[ref.Out(i)][name] = Binding{instance: implementations[i].Interface()}
		}
	}

	return nil
}

// Singleton binds value(s) returned from constructor as a singleton objects of related types.
func (self *container) Singleton(constructor interface{}, opts ...Option) error {
	return self.bind(constructor, newBindOptions(opts))
}

// Factory binds constructor as a factory method of related type.
func (self *container) Factory(constructor interface{}, opts ...Option) error {
	var options = newBindOptions(opts)
	options.factory = true

	return self.bind(constructor, options)
}

// Implementation receives ready instance and binds it to its REAL type, which means that declared abstract variable type (interface) is ignored
func (self *container) Implementation(implementation interface{}, opts ...Option) error {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var ref = reflect.TypeOf(implementation)
	if _, ok := self.bindings[ref]; !ok {
		self.bindings[ref] = make(map[string]Binding)
	}

	var options = newBindOptions(opts)
	if len(options.names) == 0 {
		options.names = []string{DefaultBindName}
	}

	self.bindings[ref][options.names[0]] = Binding{instance: implementation}

	return nil
}

func (self *container) ListBindings(abstraction reflect.Type) (map[string]Binding, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var bnds, ok = self.bindings[abstraction]
	if !ok {
		return bnds, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	return bnds, nil
}

// Reset deletes all the existing bindings and empties the container instance.
func (self *container) Reset() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for k := range self.bindings {
		delete(self.bindings, k)
	}
}
