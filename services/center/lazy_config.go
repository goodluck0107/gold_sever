package center

import (
	"sync"
)

var (
	lazyWatchUserLock sync.Mutex
	watchGameMap      map[int]bool
)

func IsGameSupportWatch(kindId int) bool {
	lazyWatchUserLock.Lock()
	defer lazyWatchUserLock.Unlock()

	support, ok := watchGameMap[kindId]
	if ok {
		return support
	}
	return false
}

func LoadGameSupportWatch() {
	lazyWatchUserLock.Lock()
	defer lazyWatchUserLock.Unlock()

	watchGameMap = make(map[int]bool)
	configs := GetServer().ConGame
	for _, con := range configs {
		watchGameMap[con.KindId] = con.IsSupportWatch
	}
}
