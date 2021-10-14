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
	GetBinding(abstraction reflect.Type, name string) (Binding, error)
	GetAllBindings(reflect.Type) (map[string]Binding, error)
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

const defaultBindName = "default"

// Binding holds either singleton instances or factory methods for bindings
type Binding struct {
	factory  interface{} // factory method that creates the appropriate implementation of the abstraction
	instance interface{} // instance stored for reusing in singleton bindings
}

func (self *container) GetAllBindings(abstraction reflect.Type) (bnds map[string]Binding, err error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var ok bool
	if bnds, ok = self.bindings[abstraction]; !ok {
		return bnds, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	return
}

func (self *container) GetBinding(abstraction reflect.Type, name string) (bnd Binding, err error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var ok bool
	if _, ok = self.bindings[abstraction]; !ok {
		return bnd, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	if bnd, ok = self.bindings[abstraction][name]; !ok {
		return bnd, fmt.Errorf("di: no binding found for: %s", abstraction.String())
	}

	return
}

func (self *container) getResolver() *resolver {
	self.lock.RLock()
	defer self.lock.RUnlock()

	return &resolver{
		list: []Container{
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

		if instances, err = self.getResolver().invoke(constructor); err != nil {
			return
		}

	case opts.factory && (ref.NumOut() == 2 && !isError(ref.Out(1)) || ref.NumOut() > 2):
		return errors.New("di: factory resolvers must return exactly one value and optionally one error")
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	for i := 0; i < numRealInstances; i++ {
		// we are not interested in returned errors
		if isError(ref.Out(i)) {
			continue
		}

		if _, ok := self.bindings[ref.Out(i)]; !ok {
			self.bindings[ref.Out(i)] = make(map[string]Binding)
		}

		if opts.names == nil {
			opts.names = []string{defaultBindName}
		}

		var name = opts.names[0]

		// Factory method
		if opts.factory {
			self.bindings[ref.Out(i)][name] = Binding{factory: constructor}
			continue
		}

		// Singleton instances
		// if there is more than one instance returned from constructor - use appropriate name for it
		if numRealInstances > 1 {
			if len(opts.names) > 1 {
				name = opts.names[i]
			}

			self.bindings[ref.Out(i)][name] = Binding{instance: instances[i].Interface()}
			continue
		}

		// if only one instance is returned from constructor - bind it under all provided names
		for _, name = range opts.names {
			self.bindings[ref.Out(i)][name] = Binding{instance: instances[i].Interface()}
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

// Reset deletes all the existing bindings and empties the container instance.
func (self *container) Reset() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for k := range self.bindings {
		delete(self.bindings, k)
	}
}
