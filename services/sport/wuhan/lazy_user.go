package wuhan

import (
	"sync"
)

var (
	lazyUserOnce sync.Once
	lazyUserLock sync.Mutex
	member       map[int64]*LazyUser
)

func init() {
	lazyUserOnce.Do(func() {
		member = make(map[int64]*LazyUser)
	})
}

type LazyUser struct {
	Uid      int64
	Sex      int
	Name     string
	ImageUrl string
}

func GetLazyUser(uid int64) *LazyUser {
	lazyUserLock.Lock()
	defer lazyUserLock.Unlock()

	user, ok := member[uid]
	if !ok {
		user = new(LazyUser)
		user.Uid = uid
		person, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if err == nil {
			user.Name = person.Nickname
			user.ImageUrl = person.Imgurl
			user.Sex = person.Sex
		}
		member[uid] = user
	}
	return user
}

func UpdateLazyUser(uid int64, name string, sex int, imageUrl string) {
	lazyUserLock.Lock()
	defer lazyUserLock.Unlock()

	user := new(LazyUser)
	user.Uid = uid
	user.Name = name
	user.ImageUrl = imageUrl
	user.Sex = sex

	member[uid] = user
}
