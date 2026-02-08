package infrastructure

import (
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
)

type TableBase interface {
	//Init()
	GetTableInfo() *static.Table
	GetPersonByChair(seat int) *TablePerson
	GetPersons() []*TablePerson
	GetGame() SportInterface
	//! 发送操作
	Operator(op *TableMsg)
	//! 设置开始结束
	SetBegin(begin bool)
	OnBegin()
	SetXiaPaoIng(xiapaoing bool)
	//! 牌桌结束
	IsBye() bool
	//run()
	// 判断用户是否在桌上
	HasIn(uid int64) bool
	CanBeIn(uid int64) *xerrors.XError
	//Tablein(person *PersonGame)
	//runLive()
	//! 检测用户连接状态
	CheckUserConnection(uid int64, seat int)
	//! 清理房间
	//clear(promptMsg *static.Msg_S2C_TableDel)
	CostUserWealth(uid int64, wealthType int8, cost int, costType int8, playernum int) (int, error)
	//! 广播消息
	//broadCastMsg(head string, errCode int16, v interface{})
	//! 发送消息
	SendMsg(uid int64, head string, errCode int16, v interface{})
	//! 发送消息
	SendLookonMsg(uid int64, head string, errCode int16, v interface{})
	//! 加入人
	//addPerson(person *PersonGame, seat int)
	// 获取座位号
	GetTableSeat(uid int64) int
	//! 获取牌桌信息
	BroadcastTableInfo(bool)
	// 写入redis
	//flush()
	//写入详细数据
	//flushInfo()
	// 清除 redis
	//remove()
	// 用户加入桌子 seat=-1时顺序落座
	//UserJoinTable(uid int64, seat int) (int, int64, *xerrors.XError)
	// 退出牌桌
	TableExit(uid, by int64) bool
	// 换桌清理数据
	//tableChange(head string, person *PersonGame) (bool, int)
	Bye()
	DelPerson(uid int64)
	//getPerson(uid int64) *TablePerson
	//! 通过座位编号获取玩家信息
	//getPersonWithSeatId(seatId int) *TablePerson
	//! 是否活动牌桌
	IsLive() bool
	IsNew() bool
	IsXiaPaoIng() bool
	CanFewer() bool
	IsBegin() bool
	//dissmissThread()
	//! 写桌子日志
	WriteTableLog(seatId uint16, recordStr string)
	//! 是否是空桌子
	IsEmpty() bool
	//! 获取当前桌子玩家人数
	GetPersonNumber() int
	// 推送牌桌变化消息至场次
	NotifyTableChange()
	//CallHall(header string, code int16, data interface{}, uid int64)
	OnFewerStart(players ...int64)
	OnInvite(inviter, invitee int64) error
	ExChangeSeat(id1, id2 int64) bool
	ReSeat(id1 int64, seat int) bool
	GetShutDownFlag() bool
	SetCostPaid()
}
