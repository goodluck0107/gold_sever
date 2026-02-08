package wuhan

import "github.com/open-source/game/chess.git/pkg/xlock"

//////////////////////////////////////////////////////////////
//! 玩家管理者
type PersonMgr struct {
	MapPerson map[int64]*PersonGame

	lock *xlock.RWMutex

	MapLookonPerson map[int64]*PersonGame
	Lookonlock      *xlock.RWMutex
}

var personmgrSingleton *PersonMgr = nil

//! 得到服务器指针
func GetPersonMgr() *PersonMgr {
	if personmgrSingleton == nil {
		personmgrSingleton = new(PersonMgr)
		personmgrSingleton.MapPerson = make(map[int64]*PersonGame)
		personmgrSingleton.lock = new(xlock.RWMutex)
		personmgrSingleton.MapLookonPerson = make(map[int64]*PersonGame)
		personmgrSingleton.Lookonlock = new(xlock.RWMutex)

		//go personmgrSingleton.HeartBeating()
	}

	return personmgrSingleton
}

//! 加入玩家
func (self *PersonMgr) AddPerson(person *PersonGame) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()
	self.MapPerson[person.Info.Uid] = person
}

//! 删玩家
func (self *PersonMgr) DelPerson(uid int64) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()
	delete(self.MapPerson, uid)
}

//! 该玩家是否存在
func (self *PersonMgr) GetPerson(uid int64) *PersonGame {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	person, ok := self.MapPerson[uid]
	if ok {
		return person
	}

	return nil
}

//! 加入旁观玩家
func (self *PersonMgr) AddLookonPerson(person *PersonGame) {
	self.Lookonlock.CustomLock()
	defer self.Lookonlock.CustomUnLock()
	self.MapLookonPerson[person.Info.Uid] = person
}

//! 删旁观玩家
func (self *PersonMgr) DelLookonPerson(uid int64) {
	self.Lookonlock.CustomLock()
	defer self.Lookonlock.CustomUnLock()
	delete(self.MapLookonPerson, uid)
}

//! 该旁观玩家是否存在
func (self *PersonMgr) GetLookonPerson(uid int64) *PersonGame {
	self.Lookonlock.RLockWithLog()
	defer self.Lookonlock.RUnlock()

	person, ok := self.MapLookonPerson[uid]
	if ok {
		return person
	}

	return nil
}
