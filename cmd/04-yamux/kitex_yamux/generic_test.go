package kitex_yamux

import (
	"fmt"
	"reflect"
	"testing"
)

type TestStruct[T any] struct {
	A T
}

func (t *TestStruct[T]) AIsNil() bool {
	return IsZero(t.A)
}

func TestNilAny(t *testing.T) {
	tt0 := &TestStruct[any]{}
	fmt.Println(tt0.AIsNil())
	tt1 := &TestStruct[*string]{}
	fmt.Println(tt1.AIsNil())
	tt2 := &TestStruct[string]{A: ""}
	fmt.Println(tt2.AIsNil())
	strEmpty := ""
	tt3 := &TestStruct[*string]{A: &strEmpty}
	fmt.Println(tt3.AIsNil())

	reflect.ValueOf(nil)
}
