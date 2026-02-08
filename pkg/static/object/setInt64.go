package object

import (
	"sync"
)

type SetInt64 struct {
	mu  sync.RWMutex
	set map[int64]struct{}
}

func (si *SetInt64) Add(elem int64) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.set == nil {
		si.set = make(map[int64]struct{})
	}
	si.set[elem] = struct{}{}
}

func (si *SetInt64) Remove(elem int64) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.set == nil {
		return
	}
	delete(si.set, elem)
}

func (si *SetInt64) Count() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.set)
}

func (si *SetInt64) Contains(elem int64) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	_, ok := si.set[elem]
	return ok
}
