package kitex_yamux

import "reflect"

func IsZero(t any) bool {
	v := reflect.ValueOf(t)
	return v == reflect.Value{} || v.IsZero()
}
