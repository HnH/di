[![GoDoc](https://godoc.org/github.com/HnH/di?status.svg)](https://godoc.org/github.com/HnH/di)
[![Go Report Card](https://goreportcard.com/badge/github.com/HnH/di)](https://goreportcard.com/report/github.com/HnH/di)

# Dependency injection
DI is a dependency injection library that is focused on clean API and flexibility. DI has two types of top-level abstractions: Container and Resolver.
First one is responsible for accepting constructors and implementations and creating abstraction bindings out of them.
Second implements different implementation resolution scenarios against one or more Containers.

Initially this library was heavily inspired by [GoLobby Container](https://github.com/golobby/container) but since then 
had a lot of backwards incompatible changes in structure, functionality and API.
To install DI simply run in your project directory:
```bash
go get github.com/HnH/di
```

### Constructor
Constructor implements a `Construct()` method which is called either after binding to container in case of singleton, either after factory method was called.

```go
type Constructor interface {
    Construct() error
}
```

### Container
```go
type Container interface {
    Singleton(constructor interface{}, opts ...Option) error
    Factory(constructor interface{}, opts ...Option) error
    Implementation(implementation interface{}, opts ...Option) error
    ListBindings(reflect.Type) (map[string]Binding, error)
    Reset()
}
```

#### Singleton
`Singleton()` method requires a constructor which will return Implementation(s) of Abstraction(s). Constructor will be called once 
and returned Implementations(s) will later always bound to Abstraction(s) on resolution requests.

```go
err = di.Singleton(func() (Abstraction, SecondAbstraction) {
    return Implementation, SecondImplementation
})

// Singleton may also accept naming option which means that returned Implementation will be available only under provided name.
err = di.Singleton(func() (Abstraction) {
    return Implementation
}, di.WithName("customName"))

// Name can be provided for each of the Implementations if there are more than one.
err = di.Singleton(func() (Abstraction, SecondAbstraction) {
    return Implementation, SecondImplementation
}, di.WithName("customName", "secondCustomName"))

// If there is only one Implementation returned you may give multiple aliases for it.
err = di.Singleton(func() (Abstraction) {
    return Implementation
}, di.WithName("customName", "secondCustomName"))


// WithFill() option calls `resolver.Fill()` on an instance right after it is created.
err = di.Singleton(func() (Abstraction) {
    return Implementation
}, di.WithFill()) // di.resolver.Fill(Implementation) will be called under the hood
```

#### Factory
`Factory()` method requires a constructor which will return exactly one Implementation of exactly one Abstraction.
Constructor will be called on each Abstraction resolution request.

```go
err = di.Factory(func() (Abstraction) {
    return Implementation
})

// Factory also optionally accepts naming option which means that returned Implementation will be available only under provided name.
err := di.Factory(func() (Abstraction) {
    return Implementation
}, di.WithName("customName"))

// Similarly to Singleton binding WithFill() option can be provided
err = di.Factory(func() (Abstraction) {
return Implementation
}, di.WithFill()) // di.resolver.Fill(Implementation) will be called under the hood
```

#### Implementation
`Implementation()` receives ready instance and binds it to its **real** type, which means that declared abstract variable type (interface) is ignored.

```go
var circle Shape = newCircle()
err = di.Implementation(circle)

// Will return error di: no binding found for: di_test.Shape
var a Shape
err = di.Resolve(&a)

// Will resolve circle.
var c *Circle
err = di.Resolve(&a)

// Also naming options can be used as everywhere.
err = di.Implementation(circle, di.WithName("customName"))
err = di.Resolve(&c, di.WithName("customName"))
```

### Resolver
```go
type Resolver interface {
    With(implementations ...interface{}) Resolver
    Resolve(receiver interface{}, opts ...Option) error
    Call(function interface{}, opts ...Option) error
    Fill(receiver interface{}) error
}
```
#### With
`With()` takes a list of instantiated implementations and tries to use them in resolving scenarios.
In the opposite to Container's `Implementation()` method `With()` does not put instances into container and does not reflect a type on a binding time.
Instead of this it reuses `reflect.Type.AssignableTo()` method capabilities on abstraction resolution time.

```go
var circle Shape = newCircle()
err = di.Implementation(circle)

// di: no binding found for: di_test.Shape
di.Call(func(s Shape) { return })

// ok
di.With(circle).Call(func(s Shape) { return }))
```

#### Resolve
`Resolve()` requires a receiver (pointer) of an Abstraction and fills it with appropriate Implementation.

```go
var abs Abstraction
err = di.Resolve(&a)

// Resolution can be done with previously registered names as well
err = di.Resolve(&a, di.WithName("customName"))
```

#### Call
The `Call()` executes as `function` with resolved Implementation as a arguments.

```go
err = di.Call(func(a Abstraction) {
    // `a` will be an implementation of the Abstraction
})

// Returned values can be bound to variables by providing an option.
var db Database
err = di.Call(func(a Abstraction) Database {
    return &MySQL{a}
}, di.WithReturn(&db))
// db == &MySQL{a}
```

#### Fill
The `Fill()` method takes a struct (pointer) and resolves its fields. The example below expresses how the `Fill()` method works.

```go
err = di.Singleton(func() Mailer { return &someMailer{} })

err = di.Singleton(func() (Database, Database) {
    return &MySQL{}, &Redis{} 
}, di.WithName("data", "cache"))

type App struct {
    mailer Mailer    `di:"type"`
    data   Database  `di:"name"`
    cache  Database  `di:"name"`
    x int
}

var App = App{}
err = container.Fill(&myApp)

// [Typed Bindings]
// `App.mailer` will be an implementation of the Mailer interface

// [Named Bindings]
// `App.data` will be a MySQL implementation of the Database interface
// `App.cache` will be a Redis implementation of the Database interface

// `App.x` will be ignored since it has no `di` tag
```

Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.

```go
var list []Shape
container.Fill(&list)

// []Shape{&Rectangle{}, &Circle{}}

var list map[string]Shape
container.Fill(&list)

// map[string]Shape{"square": &Rectangle{}, "rounded": &Circle{}} 
```
