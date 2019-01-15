package gosocket

import (
	"fmt"
	"reflect"
)

type caller struct {
	Func reflect.Value
}

func newCaller(f interface{}) (*caller, error) {
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		return nil, fmt.Errorf("f is not func")
	}
	return &caller{
		Func: fv,
	}, nil
}

func (c *caller) Call(so *Socket, message Message) []reflect.Value {
	var a []reflect.Value
	a = make([]reflect.Value, 2)
	a[0] = reflect.ValueOf(so)
	a[1] = reflect.ValueOf(message)
	return c.Func.Call(a)
}
