// Package di is a dependency injection library that is focused on clean API and flexibility.
// DI has two types of top-level abstractions: Container and Resolver.
// First one is responsible for accepting constructors and implementations and creating abstraction bindings out of them.
// Second implements different implementation resolution scenarios against one or more Containers.
//
// Initially this library was heavily inspired by GoLobby Container (https://github.com/golobby/container) but since then
// had a lot of backwards incompatible changes in structure, functionality and API.
package di

import (
	"context"
	"reflect"
)

// globalContext holds a reference to global Container and Resolver
var globalContext = Ctx(context.Background()).SetContainer(NewContainer())

// Singleton binds value(s) returned from constructor as a singleton objects of related types.
func Singleton(ctx context.Context, constructor any, opts ...Option) error {
	return Ctx(ctx).Container().Singleton(constructor, opts...)
}

// Factory binds constructor as a factory method of related type.
func Factory(ctx context.Context, constructor any, opts ...Option) error {
	return Ctx(ctx).Container().Factory(constructor, opts...)
}

// Implementation receives ready instance and binds it to its REAL type, which means that declared abstract variable type (interface) is ignored
func Implementation(ctx context.Context, implementation any, opts ...Option) error {
	return Ctx(ctx).Container().Implementation(implementation, opts...)
}

// Reset deletes all the existing bindings and empties the container instance.
func Reset(ctx context.Context) {
	Ctx(ctx).Container().Reset()
}

// With takes a list of instantiated implementations and tries to use them in resolving scenarios
func With(ctx context.Context, implementations ...any) Resolver {
	return Ctx(ctx).Resolver().With(implementations...)
}

// Call takes a function, builds a list of arguments for it from the available bindings, calls it and returns a result.
func Call(ctx context.Context, function any, opts ...Option) error {
	return Ctx(ctx).Resolver().Call(function, opts...)
}

// Resolve takes a receiver and fills it with the related implementation.
func Resolve(ctx context.Context, receiver any, opts ...Option) error {
	return Ctx(ctx).Resolver().Resolve(receiver, opts...)
}

// Fill takes a struct and resolves the fields with the tag `di:"..."`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func Fill(ctx context.Context, receiver any) error {
	return Ctx(ctx).Resolver().Fill(receiver)
}

func isError(v reflect.Type) bool {
	return v.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
