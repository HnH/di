// Package di is a lightweight yet powerful IoC container for Go projects.
// It provides an easy-to-use interface and performance-in-mind container to be your ultimate requirement.
package di

import (
	"reflect"
)

var (
	// globalContainer is the global repository for bindings
	globalContainer = NewContainer().(*container)
	// globalResolver is the resolver against globalContainer
	globalResolver = globalContainer.getResolver()
)

// Singleton binds value(s) returned from constructor as a singleton objects of related types.
func Singleton(constructor interface{}, opts ...Option) error {
	return globalContainer.Singleton(constructor, opts...)
}

// Instance receives ready instance and bind it to it's REAL type, which means that declared abstract variable type (interface) is ignored
func Instance(instance interface{}, name string) error {
	return globalContainer.Instance(instance, name)
}

// Factory binds constructor as a factory method of related type.
func Factory(constructor interface{}, opts ...Option) error {
	return globalContainer.Factory(constructor, opts...)
}

// Reset deletes all the existing bindings and empties the container instance.
func Reset() {
	globalContainer.Reset()
}

// With takes a list of ready instances and tries to use them in resolving scenarios
func With(instances ...interface{}) Resolver {
	return globalResolver.With(instances...)
}

// Call takes a function, builds a list of arguments for it from the available bindings, calls it and returns a result.
func Call(function interface{}, opts ...Option) error {
	return globalResolver.Call(function, opts...)
}

// Resolve takes a receiver and fills it with the related implementation.
func Resolve(receiver interface{}, opts ...Option) error {
	return globalResolver.Resolve(receiver, opts...)
}

// Fill takes a struct and resolves the fields with the tag `di:"..."`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func Fill(receiver interface{}) error {
	return globalResolver.Fill(receiver)
}

func isError(v reflect.Type) bool {
	return v.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
