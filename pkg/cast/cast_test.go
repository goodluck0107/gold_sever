package cast_test

import (
	"reflect"
)

type testStep struct {
	input  any
	expect any
	iserr  bool
}

type testValues struct {
	zero, one, eight, eightnegative, eightpoint31, eightpoint31negative, eightpoint31_32, eightpoint31negative_32 any
}

func isUint(a any) bool {
	kind := reflect.TypeOf(a).Kind()
	return kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64
}
