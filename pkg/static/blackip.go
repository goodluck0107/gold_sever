//! ip黑名单
package static

import (
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
)

const TIMESOUT = 10000 //!黑名单次数

type BlackIpMgr struct {
	Black map[string]int
	lock  *lock2.RWMutex
}

var blackipSingleton *BlackIpMgr = nil

//! 得到服务器指针
func GetBlackIpMgr() *BlackIpMgr {
	if blackipSingleton == nil {
		blackipSingleton = new(BlackIpMgr)
		blackipSingleton.lock = new(lock2.RWMutex)
		blackipSingleton.Black = make(map[string]int)
	}

	return blackipSingleton
}

//! 加入一个黑名单
func (self *BlackIpMgr) AddIp(ip string, info string) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()

	value, ok := self.Black[ip]
	if ok {
		self.Black[ip] = value + 1
	} else {
		self.Black[ip] = 1
	}
}

//! 是否在黑名单
func (self *BlackIpMgr) IsIp(ip string) bool {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()
	value, _ := self.Black[ip]
	return value >= TIMESOUT
}
