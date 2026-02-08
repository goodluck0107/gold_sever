package gameserver

import (
	"github.com/open-source/game/chess.git/pkg/static"
)

// 获取person信息
func getPersonInfo(person *static.Person) *static.Msg_S2S_Person {
	p := new(static.Msg_S2S_Person)
	p.Uid = person.Uid
	p.Nickname = person.Nickname
	p.Imgurl = person.Imgurl
	p.Card = person.Card
	p.Sex = person.Sex
	p.Tel = person.Tel
	return p
}
