package object

// Set encapsulation an unordered collection of generics
type Set struct {
	iList map[object]struct{} // generic collections
	mu    Mu                  // lock
}

// set tag set a name of your iList
func (st *Set) SetTag(t string) {
	st.mu.setTag(t)
}

// onWrite when iList add or set same value to lock it
func (st *Set) onWrite() func() {
	if st.iList == nil {
		st.iList = make(map[object]struct{})
	}
	return st.mu.Lock()
}

// contain return key is existing
func (st *Set) Contains(key object) bool {
	defer st.mu.RLock()()
	_, ok := st.iList[key]
	return ok
}

// add your want, does not override the original value
func (st *Set) Add(key object) {
	defer st.onWrite()()
	if _, ok := st.iList[key]; ok {
		return
	}
	st.iList[key] = struct{}{}
}

// remove delete some data you don't want
func (st *Set) Remove(key object) {
	defer st.mu.Lock()()
	delete(st.iList, key)
}

// len return length of your set.
func (st *Set) Len() int {
	defer st.mu.RLock()()
	return len(st.iList)
}

// to slice convert your set data to a slice
func (st *Set) ToSlice() (keys []object) {
	defer st.mu.RLock()()
	for key := range st.iList {
		keys = append(keys, key)
	}
	return
}

// range iterate through your data
// filter: if this function returns false, the loop is ignored
// do: somethings you want to do, will given you data of this loop
func (st *Set) Range(filter func(key object) bool, do func(key object)) {
	defer st.mu.RLock()()
	for k := range st.iList {
		if k == nil {
			continue
		}
		if filter != nil {
			if !filter(k) {
				continue
			}
		}
		if do != nil {
			do(k)
		}
	}
}

// sort return a key slice and a value slice of data sorted in the way you specify
// less: you specify sort func.
func (st *Set) SortKey(less func(i, j object) bool) (keys []object) {
	defer st.mu.RLock()()
	for key := range st.iList {
		keys = append(keys, key)
	}
	SortAny(keys, less)
	return
}
