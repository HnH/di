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
func Singleton(resolver interface{}, opts ...Option) error {
	return globalContainer.Singleton(resolver, opts...)
}

// Factory binds constructor as a factory method of related type.
func Factory(resolver interface{}, opts ...Option) error {
	return globalContainer.Factory(resolver, opts...)
}

// Reset deletes all the existing bindings and empties the container instance.
func Reset() {
	globalContainer.Reset()
}

// Call takes a function, builds a list of arguments for it from the available bindings, calls it and returns a result.
func Call(receiver interface{}, opts ...Option) error {
	return globalResolver.Call(receiver, opts...)
}

// Resolve takes a receiver and fills it with the related implementation.
func Resolve(abstraction interface{}, opts ...Option) error {
	return globalResolver.Resolve(abstraction, opts...)
}

// Fill takes a struct and resolves the fields with the tag `di:"..."`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func Fill(receiver interface{}) error {
	return globalResolver.Fill(receiver)
}

func isError(v reflect.Type) bool {
	return v.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
