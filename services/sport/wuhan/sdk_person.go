package wuhan

import (
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

//! 玩家信息
type PersonGame struct {
	Info    static.Person
	session *Session       //! session
	Minfo   static.MapInfo // 位置信息
	Ip      string         // ip
}

func (self *PersonGame) SendNullMsg() {
	var msg static.Msg_Null
	self.SendMsg("nothing", xerrors.SuccessCode, &msg)
}

func (self *PersonGame) SendMsg(head string, errCode int16, v interface{}) {
	if self == nil || self.session == nil {
		xlog.Logger().Warnf("persongame is nil:%s", head)
		return
	}
	self.session.SendMsg(head, errCode, v, self.Info.Uid)
}

func (self *PersonGame) GetInfo() static.Person {
	return self.Info
}

func (self *PersonGame) GetIp() string {
	return self.Ip
}

func (self *PersonGame) IsVip() bool {
	return self.Info.IsVip
}

//func (self *PersonGame) SetSesson(session *Session) bool {
//	if self.session == nil {
//		return false
//	}
//	self.session.SafeClose(self.Info.Uid, constant.SESSION_CLOED_FORCE)
//	self.session = session
//	return true
//}

// 由服务器主动断开客户端的session连接, 客户端不重连
func (self *PersonGame) CloseSession(origin uint8) {

	if self.session == nil {
		return
	}

	self.session.SafeClose(origin)
}
