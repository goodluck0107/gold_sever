package infrastructure

import (
	"github.com/open-source/game/chess.git/pkg/static"
)

// 游戏协议
type SportInterface interface {
	OnInit(table TableBase) // 初始化游戏

	OnSendInfo(person PersonBase) // 游戏数据同步

	OnSendInfoLookon(person PersonBase) // 游戏数据同步

	OnSwitchLookon(person PersonBase, seat uint16) // 切换旁观位置

	OnMsg(msg *TableMsg) bool // 收到消息

	OnBaseMsg(msg *TableMsg) // gamecommon收到消息

	OnBegin() // 牌局开始

	OnEnd() // 牌局解散

	OnGameOver(chair uint16, cbReason byte) bool // 单局结算

	OnGameStart() // 小局开始

	OnExit(uid int64) // 玩家退出

	OnIsDealer(uid int64) bool // 判断庄家

	OnSeat(uid int64, seat int) bool // 玩家坐下

	OnStandup(uid, by int64) bool // 玩家站起

	OnStandupLookon(uid, by int64) bool // 玩家站起

	SendWatcherList(uid int64) //! 推送观战消息列表

	OnLine(uid int64, line bool) // 玩家上下线

	OnReady(uid int64, readystaus bool) // 玩家准备或取消准备

	OnTime() // 调度周期

	OnEventTimer(dwTimerID uint16, wBindParam int64) bool // 计时器

	SendGameScene(uid int64, status byte, secret bool) // 发送游戏场景

	GetGameConfig() *static.GameConfig // 获取游戏相关配置

	CheckExit(uid int64) bool // 判断是否可以强退, false:不能强退；true可以强退

	SendUpdateScore(wChairID uint16) // 发送玩家积分数据

	Tojson() string // 场景保存

	Unmarsha(data string) // 场景恢复

	IsPausing() bool // 是否是暂停状态

	OnScoreOffset(total []float64) // 分数变化事件

	OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool //开始下一局

	AiSupperLow() bool

	GetSeatStr(Seat uint16) (fengSeat string)

	GetRepertoryCards() (res [static.MAX_REPERTORY]byte)
}
