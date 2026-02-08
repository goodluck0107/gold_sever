package object

import (
	"fmt"
	"reflect"
)

// Map encapsulation a thread safe dictionary
type Map struct {
	iList   map[object]object // generic collections
	keyType reflect.Kind      // kind of data
	mu      Mu                // lock
}

// set tag set a name of your iList
func (mp *Map) SetTag(t string) {
	mp.mu.setTag(t)
}

// onWrite when iList add or set same value to lock it
func (mp *Map) onWrite(key, value object) func() {
	if mp.iList == nil {
		mp.iList = make(map[object]object)
		mp.keyType = reflect.TypeOf(key).Kind()
	} else {
		if kk := reflect.TypeOf(key).Kind(); kk != mp.keyType {
			panic(fmt.Sprintf("wrong type of key, expected type:%v, got type:%v", mp.keyType, kk))
		}
	}
	return mp.mu.Lock()
}

// contain return key is existing
func (mp *Map) Contains(key object) bool {
	defer mp.mu.RLock()()
	_, ok := mp.iList[key]
	return ok
}

// get return your set
func (mp *Map) Get(key object) object {
	defer mp.mu.RLock()()
	if val, ok := mp.iList[key]; ok {
		return val
	}
	return nil
}

// set your want, may override the original value
func (mp *Map) Set(key, value object) {
	defer mp.onWrite(key, value)()
	mp.iList[key] = value
}

// add your want, does not override the original value
func (mp *Map) Add(key, value object) {
	defer mp.onWrite(key, value)()
	if _, ok := mp.iList[key]; ok {
		return
	}
	mp.iList[key] = value
}

// remove delete some data you don't want
func (mp *Map) Remove(key object) {
	defer mp.mu.RLock()()
	delete(mp.iList, key)
}

// len return length of your set.
func (mp *Map) Len() int {
	defer mp.mu.RLock()()
	return len(mp.iList)
}

// to slice convert your set data to a slice
func (mp *Map) ToSlice() (keys, values []object) {
	defer mp.mu.RLock()()
	for key, value := range mp.iList {
		keys = append(keys, key)
		values = append(values, value)
	}
	return
}

// range iterate through your data
// filter: if this function returns false, the loop is ignored
// do: somethings you want to do, will given you data of this loop
func (mp *Map) Range(filter func(key, value object) bool, do func(key, value object)) {
	defer mp.mu.Lock()()
	for k, v := range mp.iList {
		if k == nil || v == nil {
			continue
		}
		if filter != nil {
			if !filter(k, v) {
				continue
			}
		}
		if do != nil {
			do(k, v)
		}
	}
}

// sort return a key slice and a value slice of data sorted in the way you specify
// less: you specify sort func.
func (mp *Map) SortKey(less func(i, j object) bool) (keys, values []object) {
	gotKeys := make([]object, 0)
	defer mp.mu.RLock()()
	for key := range mp.iList {
		gotKeys = append(gotKeys, key)
	}
	SortAny(gotKeys, less)
	for _, key := range gotKeys {
		if key == nil {
			continue
		}
		value := mp.iList[key]
		if value == nil {
			continue
		}
		values = append(values, value)
	}
	return gotKeys, values
}

// sort return a value slice of data sorted in the way you specify
// less: you specify sort func.
func (mp *Map) SortVal(less func(i, j object) bool) []object {
	_, values := mp.ToSlice()
	SortAny(values, less)
	return values
}
