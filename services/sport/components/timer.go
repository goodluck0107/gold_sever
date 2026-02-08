package components

import "sync"

/**

游戏计时器，计时池

*/

// 计算器事件
const (
	GameTime_Nine     = iota + 1 //9秒场
	GameTime_1                   //1秒场
	GameTime_12                  //12秒场
	GameTime_15                  //15秒场
	GameTime_30                  //30秒场
	GameTime_AutoNext            //自动开始下一局
	GameTime_60                  //60秒场
	GameTime_TuoGuan             //托管定时间
	GameTime_Delayer
	GameTime_Ready //准备超时
	GameTime_Null  //无效事件
)

const (
	GameTime_Type_All  = 1 //大局生效
	GameTime_Type_Only = 2 //小局内部及时
)

const (
	GAME_OPERATION_TIME_AUTONEXT = 14 //自动开始下一局,操作超时时间
	GAME_OPERATION_TIME_READY    = 15 //准备超时时间

)

type TagTimerItem struct {
	Type      int    //类型
	TimerID   uint16 //时间标识, 计算器事件
	TimeLeave int    //剩余时间 倒计时数 服务器用
	Id        int64  //玩家id
}

type GameTime = Time

type Time struct {
	//	limitOP   *TagTimerItem           //20181213 限时操作用
	TimeArray    map[int64]*TagTimerItem //计时池
	OperateMutex sync.Mutex
}

func (self *Time) Init() {
	//初始化扔2个先
	if self.TimeArray == nil {
		self.TimeArray = make(map[int64]*TagTimerItem, 2)
	}
}

func (self *Time) Clean() {
	self.TimeArray = make(map[int64]*TagTimerItem, 2)
}

//timeId:计时器事件
//lefttime:时间
//func (self *Time) SetTimer(timeId int64, lefttime int) (result bool) {
//	switch timeId {
//	//case GameTime_AutoNext:
//	//	result = self.SetTimer(timeId, public.XTMJ_OPERATION_TIME)
//	default:
//		result = self.SetTimer(timeId, lefttime)
//	}
//	return
//}

// 20181213 苏大强 添加的新处理
func (self *Time) SetLimitTimer(lefttime int) {
	//if self.limitOP == nil {
	//	self.limitOP = &TagTimerItem{}
	//}
	//	self.limitOP.TimeLeave = lefttime

	self.SetTimer(GameTime_Nine, lefttime)

}

func (self *Time) SetDelayLimitTimer(timerID int64, lefttime int) {

	self.SetTimer(timerID, lefttime)
}

func (self *Time) KillLimitTimer() {
	//self.limitOP = nil
	self.KillTimer(GameTime_Nine)
}

// 20191030 苏大强 添加的新处理
func (self *Time) SetLimitTimer_safe(lefttime int) {
	self.SetTimer_safe(GameTime_Nine, lefttime, true)

}

// 设置准备超时
func (self *Time) SetReadyTimer(lefttime int) {
	if lefttime > 0 {
		self.SetTimer(GameTime_Ready, lefttime)
		return
	}
	self.SetTimer(GameTime_Ready, GAME_OPERATION_TIME_READY)
}

func (self *Time) KillReadyTimer() {
	self.KillTimer(GameTime_Ready)
}

/*
功能开启计时器
参数 :
timeId :计时器事件ID,唯一标示当前计算事件:
[
9秒场：GameTime_Nine
]
lefttime:计时器相应剩余时间

说明：小局内部计时器，下一局时会清除，建议游戏内部规则相关计时器，

例如：自动出牌
*/
func (self *Time) SetTimer(timeId int64, lefttime int) bool {
	return self.SetTimerByType(timeId, lefttime, GameTime_Type_Only)

}

/*
说明：大局内部计时器，下一局时不会清除，建议框架相关计时器,

例如：离线超时
*/
func (self *Time) SetWholeTime(timeId int64, lefttime int) bool {
	return self.SetTimerByType(timeId, lefttime, GameTime_Type_All)
}

func (self *Time) SetTimerByType(timeId int64, lefttime int, _type int) bool {
	if self.TimeArray == nil {
		self.Init()
	}
	if timeritem, ok := self.TimeArray[timeId]; ok {
		//20181224 苏大强 场景恢复后，TimerID没有，导致服务器不动
		timeritem.TimerID = uint16(timeId)
		timeritem.TimeLeave = lefttime
		return false
	}
	var timeItem TagTimerItem
	//20181224 苏大强 判断是根据timerID来做的，Id只是玩家的ID，现在无用了
	timeItem.Id = timeId
	timeItem.TimerID = uint16(timeId)
	timeItem.TimeLeave = lefttime
	timeItem.Type = _type
	// fmt.Println(fmt.Sprintf("添加新的计时器TimerID（%d）id（%d）TimeLeave（%d）周期", timeItem.TimerID, timeItem.Id, timeItem.TimeLeave))
	//以9秒做id
	self.TimeArray[timeId] = &timeItem
	return true

}

func (self *Time) SetTimer_safe(timeId int64, lefttime int, needlock bool) bool {
	if self.TimeArray == nil {
		self.Init()
	}
	if timeritem, ok := self.TimeArray[timeId]; ok {
		//20181224 苏大强 场景恢复后，TimerID没有，导致服务器不动
		timeritem.TimerID = uint16(timeId)
		if needlock {
			self.OperateMutex.Lock()
			timeritem.TimeLeave = lefttime
			self.OperateMutex.Unlock()
		} else {
			timeritem.TimeLeave = lefttime
		}
		return false
	}
	var timeItem TagTimerItem
	//20181224 苏大强 判断是根据timerID来做的，Id只是玩家的ID，现在无用了
	timeItem.Id = timeId
	timeItem.TimerID = uint16(timeId)
	timeItem.Type = GameTime_Type_Only
	if needlock {
		self.OperateMutex.Lock()
		timeItem.TimeLeave = lefttime
		self.OperateMutex.Unlock()
	} else {
		timeItem.TimeLeave = lefttime
	}
	// fmt.Println(fmt.Sprintf("添加新的计时器TimerID（%d）id（%d）TimeLeave（%d）周期", timeItem.TimerID, timeItem.Id, timeItem.TimeLeave))
	//以9秒做id
	self.TimeArray[timeId] = &timeItem
	return true

}

/*
//删除固定计时器
timeId :计算器时间ID,唯一标示当前计算事件
*/
func (self *Time) KillTimer(timeId int64) bool {
	for k, v := range self.TimeArray {
		if v.TimerID == uint16(timeId) {
			// fmt.Println(fmt.Sprintf("删除（%d）的定时器", timeId))
			delete(self.TimeArray, k)
			break
		}
	}
	return true
}

func (self *Time) OnNextClean() {
	for k, v := range self.TimeArray {
		if v.Type == GameTime_Type_Only {
			delete(self.TimeArray, k)
		}
	}
}

// 清理
func (self *Time) Clear() {
	for k, v := range self.TimeArray {
		if v.TimeLeave <= 0 {
			delete(self.TimeArray, k)
		}
	}
}
