package object

import (
	"fmt"
	"reflect"
	"sort"
)

// see https://github.com/golang/go/wiki/InterfaceSlice
// because non-interface-type slices have a different spatial layout in memory than interface-type slices,
// it would be time-consuming to do such implicit conversions, so go does not support such conversions.
// We solved it with reflection.(ps: although the performance was a little bit worse)
func SortAny(iSlice any, less func(i, j any) bool) {
	// parameters are verified
	if iSlice == nil {
		panic("first arg must be a slice, not nil.")
	}
	if less == nil {
		panic("second arg (less) must be a valid comparison function, not nil.")
	}
	if k := reflect.TypeOf(iSlice).Kind(); k != reflect.Slice {
		panic(fmt.Errorf("wrong type: first arg must be a slice, given %v", k))
	}
	// start sorting
	sort.Sort(&anySorter{
		data: reflect.ValueOf(iSlice),
		less: less,
	})
}

type anySorter struct {
	data reflect.Value
	less func(i, j any) bool
}

func (ms *anySorter) Len() int {
	return ms.data.Len()
}

func (ms *anySorter) Less(i, j int) bool {
	return ms.less(ms.data.Index(i).Interface(), ms.data.Index(j).Interface())
}

func (ms *anySorter) Swap(i, j int) {
	bottle := ms.data.Index(i).Interface()
	ms.data.Index(i).Set(reflect.ValueOf(ms.data.Index(j).Interface()))
	ms.data.Index(j).Set(reflect.ValueOf(bottle))
}
