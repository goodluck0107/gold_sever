package util

import (
	"fmt"
	"reflect"
	"sort"
)

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type SortItem struct {
	key   int
	param interface{}
}
type SortItemSlice []SortItem

func (self SortItemSlice) Len() int {
	return len(self)
}
func (self SortItemSlice) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self SortItemSlice) Less(i, j int) bool {
	return self[i].key < self[j].key
}

func SortAny(iSlice interface{}, less func(i, j interface{}) bool) {
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
	less func(i, j interface{}) bool
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
