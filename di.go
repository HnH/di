// Package di is a dependency injection library that is focused on clean API and flexibility.
// DI has two types of top-level abstractions: Container and Resolver.
// First one is responsible for accepting constructors and implementations and creating abstraction bindings out of them.
// Second implements different implementation resolution scenarios against one or more Containers.
//
// Initially this library was heavily inspired by GoLobby Container (https://github.com/golobby/container) but since then
// had a lot of backwards incompatible changes in structure, functionality and API.
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

// Factory binds constructor as a factory method of related type.
func Factory(constructor interface{}, opts ...Option) error {
	return globalContainer.Factory(constructor, opts...)
}

// Implementation receives ready instance and binds it to its REAL type, which means that declared abstract variable type (interface) is ignored
func Implementation(implementation interface{}, opts ...Option) error {
	return globalContainer.Implementation(implementation, opts...)
}

// Reset deletes all the existing bindings and empties the container instance.
func Reset() {
	globalContainer.Reset()
}

// With takes a list of instantiated implementations and tries to use them in resolving scenarios
func With(implementations ...interface{}) Resolver {
	return globalResolver.With(implementations...)
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
