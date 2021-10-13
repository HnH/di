package di

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

// arguments returns container-resolved arguments of a function.
func (c Container) arguments(function interface{}) ([]reflect.Value, error) {
	reflectedFunction := reflect.TypeOf(function)
	argumentsCount := reflectedFunction.NumIn()
	arguments := make([]reflect.Value, argumentsCount)

	for i := 0; i < argumentsCount; i++ {
		abstraction := reflectedFunction.In(i)

		if concrete, exist := c[abstraction][""]; exist {
			instance, _ := concrete.resolve(c)

			arguments[i] = reflect.ValueOf(instance)
		} else {
			return nil, errors.New("container: no concrete found for: " + abstraction.String())
		}
	}

	return arguments, nil
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

// resolve creates an appropriate implementation of the related abstraction
func (b binding) resolve(c Container) (interface{}, error) {
	if b.instance != nil {
		return b.instance, nil
	}

	out, err := c.invoke(b.resolver)
	if err != nil {
		return nil, err
	}

	return out[0].Interface(), nil
}

func (c Container) isError(v reflect.Type) bool {
	return v.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

// Call takes a function (receiver) with one or more arguments of the abstractions (interfaces).
// It invokes the function (receiver) and passes the related implementations.
func (c Container) Call(function interface{}) error {
	receiverType := reflect.TypeOf(function)
	if receiverType == nil || receiverType.Kind() != reflect.Func {
		return errors.New("container: invalid function")
	}

	args, err := c.arguments(function)
	if err != nil {
		return err
	}

	out := reflect.ValueOf(function).Call(args)
	// if there is something returned from a function and the last value is error and it's not nil then return it
	if len(out) > 0 && out[len(out)-1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) && !out[len(out)-1].IsNil() {
		return out[len(out)-1].Interface().(error)
	}

	return nil
}

// Resolve takes an abstraction (interface reference) and fills it with the related implementation.
func (c Container) Resolve(abstraction interface{}) error {
	return c.NamedResolve(abstraction, "")
}

// NamedResolve resolves like the Resolve method but for named bindings.
func (c Container) NamedResolve(abstraction interface{}, name string) error {
	receiverType := reflect.TypeOf(abstraction)
	if receiverType == nil {
		return errors.New("container: invalid abstraction")
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()

		if concrete, exist := c[elem][name]; exist {
			if instance, err := concrete.resolve(c); err != nil {
				return err
			} else {
				reflect.ValueOf(abstraction).Elem().Set(reflect.ValueOf(instance))
			}

			return nil
		}

		return errors.New("container: no concrete found for: " + elem.String())
	}

	return errors.New("container: invalid abstraction")
}

// Fill takes a struct and resolves the fields with the tag `container:"inject"`.
// Alternatively map[string]Type or []Type can be provided. It will be filled with all available implementations of provided Type.
func (c Container) Fill(receiver interface{}) error {
	receiverType := reflect.TypeOf(receiver)
	if receiverType == nil {
		return errors.New("container: invalid receiver")
	}

	if receiverType.Kind() != reflect.Ptr {
		return errors.New("container: receiver is not a pointer")
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

	return errors.New("container: invalid receiver")
}

func (c Container) fillStruct(receiver interface{}) error {
	s := reflect.ValueOf(receiver).Elem()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if t, exist := s.Type().Field(i).Tag.Lookup("container"); exist {
			var name string

			if t == "type" {
				name = ""
			} else if t == "name" {
				name = s.Type().Field(i).Name
			} else {
				return errors.New(
					fmt.Sprintf("container: %v has an invalid struct tag", s.Type().Field(i).Name),
				)
			}

			if concrete, exist := c[f.Type()][name]; exist {
				instance, _ := concrete.resolve(c)

				ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
				ptr.Set(reflect.ValueOf(instance))

				continue
			}

			return errors.New(fmt.Sprintf("container: cannot resolve %v field", s.Type().Field(i).Name))
		}
	}

	return nil
}

func (c Container) fillSlice(receiver interface{}) error {
	elem := reflect.TypeOf(receiver).Elem()

	if _, exist := c[elem.Elem()]; exist {
		result := reflect.MakeSlice(reflect.SliceOf(elem.Elem()), 0, len(c[elem.Elem()]))

		for _, concrete := range c[elem.Elem()] {
			instance, _ := concrete.resolve(c)

			result = reflect.Append(result, reflect.ValueOf(instance))
		}

		reflect.ValueOf(receiver).Elem().Set(result)
	}

	return nil
}

func (c Container) fillMap(receiver interface{}) error {
	elem := reflect.TypeOf(receiver).Elem()

	if _, exist := c[elem.Elem()]; exist {
		result := reflect.MakeMapWithSize(reflect.MapOf(elem.Key(), elem.Elem()), len(c[elem.Elem()]))

		for name, concrete := range c[elem.Elem()] {
			instance, _ := concrete.resolve(c)

			result.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(instance))
		}

		reflect.ValueOf(receiver).Elem().Set(result)
	}

	return nil
}
