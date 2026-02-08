package object

import (
	"sort"
)

type object = interface{}

type any = object

type objects []object

type objSorter struct {
	data objects
	less func(i, j object) bool
}

func (os *objSorter) Len() int { return len(os.data) }

func (os *objSorter) Less(i, j int) bool { return os.less(os.data[i], os.data[j]) }

func (os *objSorter) Swap(i, j int) { os.data[i], os.data[j] = os.data[j], os.data[i] }

func SortObj(data objects, less func(i, j object) bool) {
	// parameters are verified
	if data == nil {
		panic("first arg must be a []interface{}, not nil.")
	}
	if less == nil {
		panic("second arg (less) must be a valid comparison function, not nil.")
	}
	// start sorting
	sort.Sort(&objSorter{data, less})
}
