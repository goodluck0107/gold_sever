package user

import (
	"github.com/open-source/game/chess.git/pkg/static"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"math/rand"
	"time"
)

//////////////////////////////////////////////////////////////
//! 玩家管理者
type PersonMgr struct {
	MapPerson map[int64]*static.Person
	maxUid    int64 // 当前最大的用户数字id

	lock *lock2.RWMutex
}

var (
	personmgrSingleton *PersonMgr
	_                  = rand.New(rand.NewSource(time.Now().UnixNano()))
)

//! 得到服务器指针
func GetPersonMgr() *PersonMgr {
	if personmgrSingleton == nil {
		personmgrSingleton = new(PersonMgr)
		personmgrSingleton.MapPerson = make(map[int64]*static.Person)
		personmgrSingleton.lock = new(lock2.RWMutex)
	}

	return personmgrSingleton
}

//! 加入玩家
func (self *PersonMgr) AddPerson(person *static.Person) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()

	self.MapPerson[person.Uid] = person
}

//! 删玩家
func (self *PersonMgr) DelPerson(uid int64) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()

	delete(self.MapPerson, uid)
}

//! 该玩家是否存在
func (self *PersonMgr) GetPerson(uid int64) *static.Person {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	person, ok := self.MapPerson[uid]
	if ok {
		return person
	}
	return nil
}

//! 该玩家是否存在
func (self *PersonMgr) GetPersonbyOpenid(openid string) *static.Person {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	for _, person := range self.MapPerson {
		if person.Openid == openid {
			return person
		}
	}
	return nil
}

//! 该玩家是否存在
func (self *PersonMgr) GetPersonbyGuestid(guestid string) *static.Person {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	for _, person := range self.MapPerson {
		if person.Guestid == guestid {
			return person
		}
	}
	return nil
}

//! 该玩家是否存在
func (self *PersonMgr) GetPersonbyMobile(mobile string) *static.Person {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	for _, person := range self.MapPerson {
		if person.Tel == mobile {
			return person
		}
	}
	return nil
}

// 设置最大uid
func (self *PersonMgr) SetMaxUid(uid int64) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()

	self.maxUid = uid
}

// 获取下一个uid
//func (self *PersonMgr) GetNextUid() int64 {
//	self.lock.CustomLock()
//	defer self.lock.CustomUnLock()
//
//	if self.maxUid == 0 {
//		// 初始化maxUid
//		self.maxUid = 10000000
//	}
//
//	// uid随机增加1-5
//	self.maxUid = self.maxUid + int64(_rand.Intn(5)+1)
//
//	return self.maxUid
//}
