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
			return nil, fmt.Errorf("di: no binding found for: %s", ref.In(i).String())
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
	if len(out) > 1 && isError(out[len(out)-1].Type()) && !out[len(out)-1].IsNil() {
		return nil, out[len(out)-1].Interface().(error)
	}

	return
}

// Call takes a function (receiver) with one or more arguments of the abstractions (interfaces).
// It invokes the function (receiver) and passes the related implementations.
func (c Container) Call(function interface{}, opts ...Option) error {
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

	var args, err = c.arguments(function)
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

// Resolve takes an abstraction (interface reference) and fills it with the related implementation.
func (c Container) Resolve(receiver interface{}, opts ...Option) error {
	var ref = reflect.TypeOf(receiver)
	if ref == nil || ref.Kind() != reflect.Ptr {
		return errors.New("di: invalid receiver")
	}

	var (
		options = newResolveOptions(opts)
		bnd, ok = c[ref.Elem()][options.name]
	)

	if !ok {
		return fmt.Errorf("di: no binding found for: %s", ref.Elem().String())
	}

	var instance, err = bnd.resolve(c)
	if err != nil {
		return err
	}

	reflect.ValueOf(receiver).Elem().Set(reflect.ValueOf(instance))

	return nil
}

// Fill takes a struct and resolves the fields with the tag `container:"inject"`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func (c Container) Fill(receiver interface{}) error {
	var ref = reflect.TypeOf(receiver)
	if ref == nil {
		return errors.New("di: invalid receiver")
	}

	if ref.Kind() != reflect.Ptr {
		return errors.New("di: receiver is not a pointer")
	}

	switch ref.Elem().Kind() {
	case reflect.Struct:
		return c.fillStruct(receiver)

	case reflect.Slice:
		return c.fillSlice(receiver)

	case reflect.Map:
		if ref.Elem().Key().Name() != "string" {
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
