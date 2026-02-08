package center

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sync"
	"time"
)

func GetMsg() {
	go GetTableMsg()

	workPool := NewWorkPool()

	cli := GetDBMgr().Redis
	for !GetServer().ShutDown {
		result, err := cli.BRPop(time.Second, consts.MsgTypeTableDel_Ntf,
			consts.MsgTypeTableExit_Ntf, consts.MsgTypeTableIn_Ntf, consts.MsgTypeGameReady).Result()
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if len(result) > 0 {
			if (len(result) % 2) != 0 {
				xlog.Logger().Warnf("blpop return length error. redis return: %+v", result)
				continue
			}
			xlog.Logger().Infof("【GET REDIS MSG】 head:%s ,params:%s", result[0], result[1])
			switch result[0] {
			case consts.MsgTypeTableDel_Ntf:
				msg := static.GH_TableDel_Ntf{}
				if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
					xlog.Logger().Errorf("加载任务内容失败。 %s  ", err)
					continue
				}
				//AsyncTableDel_Ntf(&task)
				workPool.AddWok(msg.TableId%1000, func(w *Work) {
					w.Job <- result
				})
			case consts.MsgTypeTableExit_Ntf:
				msg := static.GH_TableExit_Ntf{}
				if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
					xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
					continue
				}
				//AsyncProtocolTableOutNtf(&msg)
				workPool.AddWok(msg.TableId%1000, func(w *Work) {
					w.Job <- result
				})
			case consts.MsgTypeTableIn_Ntf:
				msg := static.GH_HTableIn_Ntf{}
				if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
					xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
					continue
				}
				//AsyncTableInNtf(&msg)
				workPool.AddWok(msg.TableId%1000, func(w *Work) {
					w.Job <- result
				})
			case consts.MsgTypeGameReady:
				msg := static.UserReadyState{}
				if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
					xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
					continue
				}
				//AsyncUserReady(&msg)
				workPool.AddWok(msg.Tid%1000, func(w *Work) {
					w.Job <- result
				})
			}

		}
	}
}

func AsyncTableDel_Ntf(msg *static.GH_TableDel_Ntf) {
	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Warnf("header:%s tid:%d not exist", consts.MsgTypeTableDel_Ntf, msg.TableId) //用户离开桌子已经清理掉桌子数据，游戏服30分钟之后再次通知解散导致
		return
	}

	if table.IsTeaHouse() {

		house := GetClubMgr().GetClubHouseByHId(table.HId)
		if house == nil {
			xlog.Logger().Errorln("header:", consts.MsgTypeTableDel_Ntf, " hid:", table.HId, " not exist")
			return
		}
		floor := house.GetFloorByFId(table.FId)
		if floor == nil {
			xlog.Logger().Errorln("header:", consts.MsgTypeTableDel_Ntf, " fid:", table.FId, " not exist")
			return
		}
		// 队列数据
		floortable := new(static.FloorTable)
		floortable.NTId = table.NTId
		floortable.TId = table.Id
		floortable.DHId = table.DHId
		floortable.FId = table.FId
		code, _ := ChOptTableDel_Ntf(floor, nil, nil, floortable)
		if code == xerrors.SuccessCode {
			// 更新玩家数据 用户退出牌桌
			for _, u := range table.Users {
				if u != nil {
					table.UserLeaveTable(u.Uid)
					GetPlayerMgr().DelPerson(u.Uid)
				}
			}
			// 删除内存数据
			GetTableMgr().DelTable(table)
		}
		return
	}

	// 更新玩家数据 用户退出牌桌
	for _, u := range table.Users {
		if u != nil {
			table.UserLeaveTable(u.Uid)
			GetPlayerMgr().DelPerson(u.Uid)
		}
	}
	// 删除内存数据
	GetTableMgr().DelTable(table)
	return
}

func AsyncProtocolTableOutNtf(msg *static.GH_TableExit_Ntf) {
	defer func() {
		GetDBMgr().Redis.Del(fmt.Sprintf("userstatus_doing_exit_%d", msg.Uid))
		GetPlayerMgr().DelPerson(msg.Uid)
	}()

	// 消息接收者
	ph := GetPlayerMgr().GetPlayer(msg.Uid)
	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Errorln(fmt.Sprintf("header:%s tid:%d not exist", consts.MsgTypeTableExit_Ntf, msg.TableId))
		return
	}

	user_lock_key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_LOCK, msg.Uid)
	if GetDBMgr().Redis.Exists(user_lock_key).Val() == 1 {
		return
	}

	if !table.IsTeaHouse() {
		table.UserLeaveTable(msg.Uid)
		return
	}

	// 包厢
	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeTableExit_Ntf, " hid:", msg.TableId, " not exist")
		return
	}
	// 包厢楼层
	floor := house.GetFloorByFId(table.FId)
	if floor == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeTableExit_Ntf, " fid:", msg.TableId, " not exist")
		return
	}
	// 包厢牌桌事件
	var pPerson *static.Person
	var pSession *Session
	if ph != nil {
		pPerson = &ph.Info
		pSession = ph.session
	} else {
		p, err := GetDBMgr().GetDBrControl().GetPerson(msg.Uid)
		if err != nil {
			pPerson = p
		} else {
			pPerson = nil
		}
		pSession = nil
	}
	ChOptTableOut_Ntf(floor, pSession, pPerson, msg)
}

func AsyncTableInNtf(ntf *static.GH_HTableIn_Ntf) {
	// 获取房间信息
	table := GetTableMgr().GetTable(ntf.TableId)
	if table == nil {
		// 牌桌不存在
		xlog.Logger().Errorf("header: %s tid %d not exists", consts.MsgTypeTableIn_Ntf, ntf.TableId)
		return
	}
	// 入桌成功
	if ntf.Result {
		hp := GetPlayerMgr().GetPlayer(ntf.Uid)
		if hp != nil {
			// 更新内存
			hp.CloseSession()
			GetPlayerMgr().AddPerson(hp)
			go GetClubMgr().UserInToHouse(&hp.Info, hp.Info.HouseId)
			_, floor, _, cuserr := inspectClubFloorMemberWithRight(hp.Info.HouseId, table.FId, hp.Info.Uid, consts.ROLE_MEMBER, MinorRightNull)
			if cuserr == xerrors.RespOk {
				hft := floor.GetHftByTid(ntf.TableId)
				if hft != nil {
					hft.UserOnlineChange(hp.Info.Uid, true)
				}
			}
		} else {
			p, err := GetDBMgr().GetDBrControl().GetPerson(ntf.Uid)
			if p == nil || err != nil {
				return
			}
			GetPlayerMgr().AddPerson(&PlayerCenterMemory{Info: *p})
			go GetClubMgr().UserInToHouse(p, p.HouseId)

			_, floor, _, cuserr := inspectClubFloorMemberWithRight(p.HouseId, table.FId, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
			if cuserr == xerrors.RespOk {
				hft := floor.GetHftByTid(ntf.TableId)
				if hft != nil {
					hft.UserOnlineChange(p.Uid, true)
				}
			}

		}

	} else {
		// 入桌失败
		table.UserLeaveTable(ntf.Uid)
	}

	return
}

func AsyncUserReady(info *static.UserReadyState) {
	// 获取房间信息
	table := GetTableMgr().GetTable(info.Tid)
	if table == nil {
		// 牌桌不存在
		xlog.Logger().Errorf("header: %s tid %d not exists", consts.MsgTypeGameReady, info.Tid)
		return
	}
	_, floor, _, cuserr := inspectClubFloorMemberWithRight(table.HId, table.FId, info.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cuserr == xerrors.RespOk {
		hft := floor.GetHftByTid(info.Tid)
		if hft != nil {
			hft.UserReadyChange(info.Uid, info.Ready)
		}
	}
	return
}

type Work struct {
	Job chan []string
}

func NewWork() *Work {
	return &Work{
		Job: make(chan []string, 1000),
	}
}

func (w *Work) Run() {
	go func() {
		for {
			select {
			case result := <-w.Job:
				if len(result) != 3 {
					continue
				}
				hid := static.HF_Atoi(result[2])

				// 慢操作丢到慢操作的携程
				if CheckSlowWork(hid, result) {
					continue
				}

				timeAfter := time.AfterFunc(5*time.Second, func() {
					// 超时5秒丢入满携程
					AddSlowWork(hid, result)

					// 开启新的携程处理原来队列的消息
					w.Run()
				})
				w.RunAction(result)
				if !timeAfter.Stop() {
					// 超时处理完,超时后已经开辟新的携程处理,此携程关闭
					// 若RunAction卡死处理不完,超时后已经开辟新的携程处理,此携程卡死
					xlog.Logger().Errorf("RunAction TimeOut, Data:%v", result)
					return
				}
			}
		}
	}()
}

func (w *Work) SlowRun() {
	go func() {
		for {
			select {
			case result := <-w.Job:
				// 慢操作10秒还未执行完,可能是卡圈,手动重置包厢
				timeAfter := time.AfterFunc(10*time.Second, func() {
					ResetHouse(nil, &static.Msg_Http_HouseTableReset{
						Hid:      static.HF_Atoi(result[2]),
						PassCode: "!@#wqsae#%^*",
					})
					xlog.Logger().Errorf("SlowRun RunAction TimeOut, ResetHouse, Data:%v", result)
					w.RunAction(result)
					w.SlowRun()
				})
				w.RunAction(result)
				if !timeAfter.Stop() {
					// 超时处理完,超时后已经开辟新的携程处理,此携程关闭
					// 若RunAction卡死处理不完,超时后已经开辟新的携程处理,此携程卡死
					xlog.Logger().Errorf("SlowRun RunAction TimeOut, Data:%v", result)
					return
				}
			}
		}
	}()
}

func (w *Work) RunAction(result []string) {
	if len(result) > 0 {
		if (len(result) % 3) != 0 {
			xlog.Logger().Warnf("blpop return length error. redis return: %+v", result)
			return
		}
		xlog.Logger().Infof("【GET REDIS MSG】 head:%s ,params:%s", result[0], result[1])
		switch result[0] {
		case consts.MsgTypeTableDel_Ntf:
			msg := static.GH_TableDel_Ntf{}
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				xlog.Logger().Errorf("加载任务内容失败。 %s  ", err)
				return
			}
			AsyncTableDel_Ntf(&msg)
		case consts.MsgTypeTableExit_Ntf:
			msg := static.GH_TableExit_Ntf{}
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
				return
			}
			AsyncProtocolTableOutNtf(&msg)
		case consts.MsgTypeTableIn_Ntf:
			msg := static.GH_HTableIn_Ntf{}
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
				return
			}
			AsyncTableInNtf(&msg)
		case consts.MsgTypeGameReady:
			msg := static.UserReadyState{}
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
				return
			}
			AsyncUserReady(&msg)
		}
	}
}

type WorkPool struct {
	Works map[int]*Work
}

func NewWorkPool() *WorkPool {
	wp := new(WorkPool)
	wp.Works = make(map[int]*Work)
	for i := 0; i < 1000; i++ {
		w := NewWork()
		wp.Works[i] = w
		w.Run()
	}
	return wp
}

func (wp *WorkPool) AddWok(index int, do func(w *Work)) {
	w, ok := wp.Works[index]
	if ok {
		do(w)
	}
}

func GetTableMsg() {
	workPool := NewWorkPool()

	cli := GetDBMgr().Redis
	for !GetServer().ShutDown {
		result, err := cli.RPop(consts.MsgTypeTableStatus_Ntf).Result()
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if result != "" {
			msg := static.G2HTableStatsChangeNtf{}
			if err := json.Unmarshal([]byte(result), &msg); err != nil {
				xlog.Logger().Errorf("加载任务内容失败。 %s ", err)
				continue
			}

			var job []string
			job = append(job, msg.Head)
			job = append(job, msg.Data)
			job = append(job, fmt.Sprintf("%d", msg.Hid))
			//AsyncUserReady(&msg)
			workPool.AddWok(int(msg.Hid)%1000, func(w *Work) {
				w.Job <- job
			})
		}
	}
}

var SlowWorks map[int]*Work
var SlowWorksOnce sync.Once
var SlowWorksLock sync.RWMutex

func init() {
	SlowWorksOnce.Do(func() {
		SlowWorks = make(map[int]*Work)
	})
}

func AddSlowWork(hid int, result []string) {
	SlowWorksLock.Lock()
	defer SlowWorksLock.Unlock()
	w := NewWork()
	SlowWorks[hid] = w
	w.SlowRun()

	var keys []int
	for key, _ := range SlowWorks {
		keys = append(keys, key)
	}
	xlog.Logger().Errorf("AddSlowWork, Now slow hid = %v", keys)

	w.Job <- result
}

func CheckSlowWork(hid int, result []string) bool {
	SlowWorksLock.RLock()
	defer SlowWorksLock.RUnlock()
	sWork, ok := SlowWorks[hid]
	if ok {
		sWork.Job <- result
		return true
	}
	return false
}
