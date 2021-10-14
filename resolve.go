package di

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

// arguments returns container-resolved arguments of a function.
func (c Container) arguments(function interface{}) ([]reflect.Value, error) {
	var (
		ref  = reflect.TypeOf(function)
		args = make([]reflect.Value, ref.NumIn())
	)

	for i := 0; i < ref.NumIn(); i++ {
		var bnd, ok = c[ref.In(i)][defaultBindName]
		if !ok {
			return nil, errors.New("di: no binding found for: " + ref.In(i).String())
		}

		var instance, err = bnd.resolve(c)
		if err != nil {
			return nil, err
		}

		args[i] = reflect.ValueOf(instance)
	}

	return args, nil
}

// invoke calls a function and returns the yielded values.
func (c Container) invoke(function interface{}) (out []reflect.Value, err error) {
	var args []reflect.Value
	if args, err = c.arguments(function); err != nil {
		return
	}

	out = reflect.ValueOf(function).Call(args)
	// if there is more than one returned value and the last one is error and it's not nil then return it
	if len(out) > 1 && c.isError(out[len(out)-1].Type()) && !out[len(out)-1].IsNil() {
		return nil, out[len(out)-1].Interface().(error)
	}

	return
}

func (c Container) isError(v reflect.Type) bool {
	return v.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

// Call takes a function (receiver) with one or more arguments of the abstractions (interfaces).
// It invokes the function (receiver) and passes the related implementations.
func (c Container) Call(function interface{}) error {
	var receiverType = reflect.TypeOf(function)
	if receiverType == nil || receiverType.Kind() != reflect.Func {
		return errors.New("di: invalid function")
	}

	var args, err = c.arguments(function)
	if err != nil {
		return err
	}

	var out = reflect.ValueOf(function).Call(args)
	// if there is something returned from a function and the last value is error and it's not nil then return it
	if len(out) > 0 && out[len(out)-1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) && !out[len(out)-1].IsNil() {
		return out[len(out)-1].Interface().(error)
	}

	return nil
}

// Resolve takes an abstraction (interface reference) and fills it with the related implementation.
func (c Container) Resolve(abstraction interface{}, opts ...Option) error {
	var receiverType = reflect.TypeOf(abstraction)
	if receiverType == nil {
		return errors.New("di: invalid abstraction")
	}

	if receiverType.Kind() != reflect.Ptr {
		return errors.New("di: invalid abstraction")
	}

	var (
		options = newResolveOptions(opts)
		bnd, ok = c[receiverType.Elem()][options.name]
	)

	if !ok {
		return fmt.Errorf("di: no binding found for: %s", receiverType.Elem().String())
	}

	var instance, err = bnd.resolve(c)
	if err != nil {
		return err
	}

	reflect.ValueOf(abstraction).Elem().Set(reflect.ValueOf(instance))

	return nil
}

// Fill takes a struct and resolves the fields with the tag `container:"inject"`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func (c Container) Fill(receiver interface{}) error {
	var receiverType = reflect.TypeOf(receiver)
	if receiverType == nil {
		return errors.New("di: invalid receiver")
	}

	if receiverType.Kind() != reflect.Ptr {
		return errors.New("di: receiver is not a pointer")
	}

	switch receiverType.Elem().Kind() {
	case reflect.Struct:
		return c.fillStruct(receiver)

	case reflect.Slice:
		return c.fillSlice(receiver)

	case reflect.Map:
		if receiverType.Elem().Key().Name() != "string" {
			break
		}

		return c.fillMap(receiver)
	}

	return errors.New("di: invalid receiver")
}

func (c Container) fillStruct(receiver interface{}) error {
	var elem = reflect.ValueOf(receiver).Elem()
	for i := 0; i < elem.NumField(); i++ {
		var tag, ok = elem.Type().Field(i).Tag.Lookup("di")
		if !ok {
			continue
		}

		var name string
		switch tag {
		case "type":
			name = defaultBindName

		case "name":
			name = elem.Type().Field(i).Name

		default:
			return fmt.Errorf("di: %v has an invalid struct tag", elem.Type().Field(i).Name)
		}

		var bnd binding
		if bnd, ok = c[elem.Field(i).Type()][name]; !ok {
			return fmt.Errorf("di: no binding found for: %v", elem.Field(i).Type().Name())
		}

		var instance, err = bnd.resolve(c)
		if err != nil {
			return err
		}

		var ptr = reflect.NewAt(elem.Field(i).Type(), unsafe.Pointer(elem.Field(i).UnsafeAddr())).Elem()
		ptr.Set(reflect.ValueOf(instance))
	}

	return nil
}

func (c Container) fillSlice(receiver interface{}) error {
	var elem = reflect.TypeOf(receiver).Elem()
	if _, ok := c[elem.Elem()]; !ok {
		return fmt.Errorf("di: no binding found for: %v", elem.Elem().String())
	}

	var result = reflect.MakeSlice(reflect.SliceOf(elem.Elem()), 0, len(c[elem.Elem()]))
	for _, bnd := range c[elem.Elem()] {
		var instance, err = bnd.resolve(c)
		if err != nil {
			return err
		}

		result = reflect.Append(result, reflect.ValueOf(instance))
	}

	reflect.ValueOf(receiver).Elem().Set(result)

	return nil
}

func (c Container) fillMap(receiver interface{}) error {
	var elem = reflect.TypeOf(receiver).Elem()
	if _, ok := c[elem.Elem()]; !ok {
		return fmt.Errorf("di: no binding found for: %v", elem.Elem().String())
	}

	var result = reflect.MakeMapWithSize(reflect.MapOf(elem.Key(), elem.Elem()), len(c[elem.Elem()]))
	for name, bnd := range c[elem.Elem()] {
		var instance, err = bnd.resolve(c)
		if err != nil {
			return err
		}

		result.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(instance))
	}

	reflect.ValueOf(receiver).Elem().Set(result)

	return nil
}
