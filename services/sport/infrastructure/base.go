package infrastructure

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sync"
)

type GameHandler func(kindid int) (kan bool, v SportInterface)

//游戏入口handler
var CreateGameFunc GameHandler = func(int) (bool, SportInterface) {
	xlog.Logger().Error("CreateGameFunc not implement yet...")
	return false, nil
}

//批量测试时，游戏局数索引
var DebugGameStep = 1
var userMutex sync.Mutex
var DebugOff = true //debug开关

func GameStep(addstep bool) int {
	if DebugOff {
		return 0
	}

	defer userMutex.Unlock()
	userMutex.Lock()
	if addstep {
		DebugGameStep++
	}
	return DebugGameStep - 1
}

/////////////////////////////////////////////////////////

type TableMsg struct {
	Head string
	Data string
	Uid  int64
	V    interface{}
}

func NewTableMsg(head string, data string, uid int64, v interface{}) *TableMsg {
	tablemsg := new(TableMsg)
	tablemsg.Head = head
	tablemsg.Data = data
	tablemsg.Uid = uid
	tablemsg.V = v

	return tablemsg
}

//! 牌桌上的人
type TablePerson struct {
	Uid       int64    `json:"uid"`        //!
	Name      string   `json:"name"`       //! 名字
	ImgUrl    string   `json:"imgurl"`     //! 头像
	Sex       int      `json:"sex"`        //! 性别
	Seat      int      `json:"seat"`       //! 座位编号
	IP        string   `json:"IP"`         //! 座位编号
	SubMsg    bool     `json:"sub_msg"`    // 是否订阅大厅消息
	SubDetail []string `json:"sub_detail"` //具体订阅哪些消息
}

func (self *TablePerson) Copy(person *TablePerson) {
	//self.Seat = person.Seat
	self.Uid = person.Uid
	self.IP = person.IP
	self.Name = person.Name
	self.ImgUrl = person.ImgUrl
	self.Sex = person.Sex
	self.SubMsg = person.SubMsg
	self.SubDetail = person.SubDetail
}
