[![CircleCI](https://circleci.com/gh/HnH/di/tree/master.svg?style=svg&circle-token=cd6ef5c602e0f89a80488349a1e4fbe034b8d717)](https://circleci.com/gh/HnH/di/tree/master)
[![codecov](https://codecov.io/gh/HnH/di/branch/master/graph/badge.svg)](https://codecov.io/gh/HnH/di)
[![Go Report Card](https://goreportcard.com/badge/github.com/HnH/di)](https://goreportcard.com/report/github.com/HnH/di)
[![GoDoc](https://godoc.org/github.com/HnH/di?status.svg)](https://godoc.org/github.com/HnH/di)

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

### Container
```go
type Container interface {
    Singleton(constructor any, opts ...Option) error
    Factory(constructor any, opts ...Option) error
    Implementation(implementation any, opts ...Option) error
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

// Will return error di: no binding found for di_test.Shape
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
    With(implementations ...any) Resolver
    Resolve(receiver any, opts ...Option) error
    Call(function any, opts ...Option) error
    Fill(receiver any) error
}
```
#### With
`With()` takes a list of instantiated implementations and tries to use them in resolving scenarios.
In the opposite to Container's `Implementation()` method `With()` does not put instances into container and does not reflect a type on a binding time.
Instead of this it reuses `reflect.Type.AssignableTo()` method capabilities on abstraction resolution time.

```go
var circle Shape = newCircle()
err = di.Implementation(circle)

// di: no binding found for di_test.Shape
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
    mailer  Mailer     `di:"type"` // fills by field type (Mailer)
    data    Database   `di:"name"` // fills by field type (Mailer) and requires binding name to be field name (data)
    cache   Database   `di:"name"`
    inner   struct {
        cache Database `di:"name"`	
    } `di:"recursive"`             // instructs DI to fill struct recursively
    another struct {
        cache Database `di:"name"` // won't have any affect as long as outer field in App struct won't have `di:"recursive"` tag
    }
}

var App = App{}
err = container.Fill(&myApp)

// [Typed Bindings]
// `App.mailer` will be an implementation of the Mailer interface

// [Named Bindings]
// `App.data` will be a MySQL implementation of the Database interface
// `App.cache` will be a Redis implementation of the Database interface
// `App.inner.cache` will be a Redis implementation of the Database interface

// `App.another` will be ignored since it has no `di` tag
```
Notice that by default `Fill()` method returns error if unable to resolve any struct fields.
If one of the fields if optional, omitempty suffix should be added to the di tag.
```go
type App struct {
    mailer  Mailer `di:"type,omitempty"` // Fill will not return error if Mailer was not provided
}
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

### Provider
Provider is an abstraction of an entity that provides something to Container

```go
type Provider interface {
    Provide(Container) error
}
```

### Constructor
Constructor implements a `Construct()` method which is called either after binding to container in case of singleton, either after factory method was called.
Note that `context.Context` must be provided in container before Constructor method can be called.

```go
type Constructor interface {
    Construct(context.Context) error
}
```

### Context propagation
```go
type Context interface {
    Put(Container) Context
    Container() Container
    Resolver() Resolver
    Raw() context.Context
}
```
Context propagation is possible via `di.Context` abstraction. Quick example:
```go
var container = di.NewContainer()
container.Implementation(newCircle())

var (
    ctx = di.Ctx(context.Background).Put(container)
    shp Shape
)

err = ctx.Resolver().Resolve(&shp) // err == nil
```
