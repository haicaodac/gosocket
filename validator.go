package gosocket

import "reflect"

// IsBlank ...
func IsBlank(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}
