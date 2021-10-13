// Package di is a lightweight yet powerful IoC container for Go projects.
// It provides an easy-to-use interface and performance-in-mind container to be your ultimate requirement.
package di

import (
	"reflect"
)

// Container holds all of the declared bindings
type Container map[reflect.Type]map[string]binding

// New creates a new instance of the Container
func New() Container {
	return make(Container)
}

// container is the global repository of bindings
var container = New()

// Singleton binds an abstraction to concrete for further singleton resolves.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func Singleton(resolver interface{}, opts ...Option) error {
	return container.Singleton(resolver, opts...)
}

// Factory binds an abstraction to concrete for further transient resolves.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func Factory(resolver interface{}, opts ...Option) error {
	return container.Factory(resolver, opts...)
}

// Reset deletes all the existing bindings and empties the container instance.
func Reset() {
	container.Reset()
}

// Call takes a function (receiver) with one or more arguments of the abstractions (interfaces).
// It invokes the function (receiver) and passes the related implementations.
func Call(receiver interface{}) error {
	return container.Call(receiver)
}

// Resolve takes an abstraction (interface reference) and fills it with the related implementation.
func Resolve(abstraction interface{}, opts ...Option) error {
	return container.Resolve(abstraction, opts...)
}

// Fill takes a struct and resolves the fields with the tag `container:"inject"`
func Fill(receiver interface{}) error {
	return container.Fill(receiver)
}
