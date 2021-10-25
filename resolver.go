package di

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

func NewResolver(containers ...Container) Resolver {
	if containers == nil {
		containers = make([]Container, 0)
	}

	return &resolver{
		containers: containers,
	}
}

type Resolver interface {
	With(implementations ...interface{}) Resolver
	Resolve(receiver interface{}, opts ...Option) error
	Call(function interface{}, opts ...Option) error
	Fill(receiver interface{}) error
}

type Constructor interface {
	Construct() error
}

type resolver struct {
	containers      []Container
	implementations []interface{}
}

func (self *resolver) getBinding(abstraction reflect.Type, name string) (bnd Binding, err error) {
	// look in with() implementation list
	for _, inst := range self.implementations {
		if reflect.TypeOf(inst).AssignableTo(abstraction) && name == DefaultBindName {
			return Binding{
				instance: inst,
			}, nil
		}
	}

	// look in containers
	var list map[string]Binding
	for _, cnt := range self.containers {
		if list, err = cnt.ListBindings(abstraction); err != nil {
			continue
		}

		var ok bool
		if bnd, ok = list[name]; ok {
			return bnd, nil
		}
	}

	return bnd, fmt.Errorf("di: no binding found for: %s", abstraction.String())
}

func (self *resolver) resolveBinding(abstraction reflect.Type, name string) (interface{}, error) {
	var bnd, err = self.getBinding(abstraction, name)
	if err != nil {
		return nil, err
	}

	return self.resolveBindingInstance(bnd)
}

func (self *resolver) resolveBindingInstance(bnd Binding) (interface{}, error) {
	// Is binding already instantiated?
	if bnd.instance != nil {
		return bnd.instance, nil
	}

	// Or we need to call a factory method?
	var out, err = self.invoke(bnd.factory)
	if err != nil {
		return nil, err
	}

	if t, ok := out[0].Interface().(Constructor); ok {
		if err = t.Construct(); err != nil {
			return nil, err
		}
	}

	return out[0].Interface(), nil
}

// arguments returns container-resolved arguments of a function.
func (self *resolver) arguments(function interface{}) ([]reflect.Value, error) {
	var (
		ref  = reflect.TypeOf(function)
		args = make([]reflect.Value, ref.NumIn())
	)

	for i := 0; i < ref.NumIn(); i++ {
		var instance, err = self.resolveBinding(ref.In(i), DefaultBindName)
		if err != nil {
			return nil, err
		}

		args[i] = reflect.ValueOf(instance)
	}

	return args, nil
}

// invoke calls a function and returns the yielded values.
func (self *resolver) invoke(function interface{}) (out []reflect.Value, err error) {
	var args []reflect.Value
	if args, err = self.arguments(function); err != nil {
		return
	}

	out = reflect.ValueOf(function).Call(args)
	// if there is more than one returned value and the last one is error and it's not nil then return it
	if len(out) > 1 && isError(out[len(out)-1].Type()) && !out[len(out)-1].IsNil() {
		return nil, out[len(out)-1].Interface().(error)
	}

	return
}

// With takes a list of instantiated implementations and tries to use them in resolving scenarios
func (self *resolver) With(implementations ...interface{}) Resolver {
	var res = &resolver{
		containers:      make([]Container, len(self.containers)),
		implementations: implementations, // this is required for us to be able to resolve already existing implementations to abstract types (interfaces)
	}

	copy(res.containers, self.containers)

	return res
}

// Call takes a function, builds a list of arguments for it from the available bindings, calls it and returns a result.
func (self *resolver) Call(function interface{}, opts ...Option) error {
	var ref = reflect.TypeOf(function)
	if ref == nil || ref.Kind() != reflect.Func {
		return errors.New("di: invalid function")
	}

	// not boolean to make further logic easier
	var returnsAnError int
	if ref.NumOut() > 0 && isError(ref.Out(ref.NumOut()-1)) {
		returnsAnError = 1
	}

	var options = newCallOptions(opts)
	if options.returns != nil && ref.NumOut()-returnsAnError-len(options.returns) != 0 {
		return fmt.Errorf("di: cannot assign %d returned values to %d receivers", ref.NumOut()-returnsAnError, len(options.returns))
	}

	var args, err = self.arguments(function)
	if err != nil {
		return err
	}

	var out = reflect.ValueOf(function).Call(args)
	// if there is something returned from a function and the last value is error and it's not nil then return it
	if returnsAnError == 1 && !out[len(out)-1].IsNil() {
		return out[len(out)-1].Interface().(error)
	}

	for i, ret := range options.returns {
		var t = reflect.TypeOf(ret)
		if t == nil || t.Kind() != reflect.Ptr || !reflect.ValueOf(ret).Elem().CanSet() || ref.Out(i) != t.Elem() {
			return fmt.Errorf("di: cannot assign returned value of type %s to %s", ref.Out(i).Name(), t.Elem().Name())
		}

		reflect.ValueOf(ret).Elem().Set(out[i])
	}

	return nil
}

// Resolve takes a receiver and fills it with the related implementation.
func (self *resolver) Resolve(receiver interface{}, opts ...Option) error {
	var ref = reflect.TypeOf(receiver)
	if ref == nil || ref.Kind() != reflect.Ptr {
		return errors.New("di: invalid receiver")
	}

	var (
		options   = newResolveOptions(opts)
		inst, err = self.resolveBinding(ref.Elem(), options.name)
	)

	if err != nil {
		return err
	}

	reflect.ValueOf(receiver).Elem().Set(reflect.ValueOf(inst))

	return nil
}

// Fill takes a struct and resolves the fields with the tag `di:"..."`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func (self *resolver) Fill(receiver interface{}) error {
	var ref = reflect.TypeOf(receiver)
	if ref == nil {
		return errors.New("di: invalid receiver")
	}

	if ref.Kind() != reflect.Ptr {
		return errors.New("di: receiver is not a pointer")
	}

	switch ref.Elem().Kind() {
	case reflect.Struct:
		return self.fillStruct(receiver)

	case reflect.Slice:
		return self.fillSlice(receiver)

	case reflect.Map:
		if ref.Elem().Key().Name() != "string" {
			break
		}

		return self.fillMap(receiver)
	}

	return errors.New("di: invalid receiver")
}

func (self *resolver) fillStruct(receiver interface{}) error {
	var elem = reflect.ValueOf(receiver).Elem()
	for i := 0; i < elem.NumField(); i++ {
		var tag, ok = elem.Type().Field(i).Tag.Lookup("di")
		if !ok {
			continue
		}

		var name string
		switch tag {
		case "type":
			name = DefaultBindName

		case "name":
			name = elem.Type().Field(i).Name

		default:
			return fmt.Errorf("di: %v has an invalid struct tag", elem.Type().Field(i).Name)
		}

		var instance, err = self.resolveBinding(elem.Field(i).Type(), name)
		if err != nil {
			return err
		}

		var ptr = reflect.NewAt(elem.Field(i).Type(), unsafe.Pointer(elem.Field(i).UnsafeAddr())).Elem()
		ptr.Set(reflect.ValueOf(instance))
	}

	return nil
}

func (self *resolver) fillSlice(receiver interface{}) error {
	var (
		elem   = reflect.TypeOf(receiver).Elem()
		result = reflect.MakeSlice(reflect.SliceOf(elem.Elem()), 0, 3)
	)

	for _, cnt := range self.containers {
		var bindings, err = cnt.ListBindings(elem.Elem())
		if err != nil {
			continue
		}

		for _, bnd := range bindings {
			var instance interface{}
			if instance, err = self.resolveBindingInstance(bnd); err != nil {
				return err
			}

			result = reflect.Append(result, reflect.ValueOf(instance))
		}
	}

	if result.Len() == 0 {
		return fmt.Errorf("di: no binding found for: %v", elem.Elem().String())
	}

	reflect.ValueOf(receiver).Elem().Set(result)

	return nil
}

func (self *resolver) fillMap(receiver interface{}) error {
	var (
		elem   = reflect.TypeOf(receiver).Elem()
		result = reflect.MakeMapWithSize(reflect.MapOf(elem.Key(), elem.Elem()), 3)
	)

	for _, cnt := range self.containers {
		var bindings, err = cnt.ListBindings(elem.Elem())
		if err != nil {
			continue
		}

		for name, bnd := range bindings {
			var instance interface{}
			if instance, err = self.resolveBindingInstance(bnd); err != nil {
				return err
			}

			result.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(instance))
		}
	}

	if result.Len() == 0 {
		return fmt.Errorf("di: no binding found for: %v", elem.Elem().String())
	}

	reflect.ValueOf(receiver).Elem().Set(result)

	return nil
}
