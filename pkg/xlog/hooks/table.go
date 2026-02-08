package hooks

import (
	"github.com/sirupsen/logrus"
)

type tableLog struct {
	tableGame string //! 当期桌子游戏名称
	gameIndex string //! 所在服务器编号
	tableId   int    //! 当前桌子日志桌子Id
	gameNum   string //! 当前桌子的唯一标识
	hid       int    //! 当前包厢六位编码
	floor     int    //! 当前桌子包厢楼层索引
	dHid      int64  //! 当前桌子包厢数据库Id
	dFid      int64  //! 当前桌子包厢楼层数据库Id
}

func (self *tableLog) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
	}
}

func (self *tableLog) Fire(entry *logrus.Entry) error {
	entry.Data["gamename"] = self.tableGame
	entry.Data["gameindex"] = self.gameIndex
	entry.Data["tableid"] = self.tableId
	entry.Data["gamenum"] = self.gameNum
	entry.Data["hid"] = self.hid
	entry.Data["floor"] = self.floor
	entry.Data["dhid"] = self.dHid
	entry.Data["dfid"] = self.dFid
	return nil
}

func NewTableHook(tId int, tGame, gIndex, gNum string, tHid, tFloor int, tdhid, tdfid int64) *tableLog {
	return &tableLog{tableGame: tGame, gameIndex: gIndex, tableId: tId, gameNum: gNum, hid: tHid, floor: tFloor, dHid: tdhid, dFid: tdfid}
}

type tableExtraLog struct {
	uid   int64
	round int
	line  int
}

func (self *tableExtraLog) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
	}
}

func (self *tableExtraLog) Fire(entry *logrus.Entry) error {
	entry.Data["uid"] = self.uid
	entry.Data["round"] = self.round
	entry.Data["line"] = self.line
	return nil
}

func NewTableExtraHook(uid int64, round int, line int) *tableExtraLog {
	return &tableExtraLog{uid: uid, round: round, line: line}
}
