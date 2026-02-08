package wuhan

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////////
//! 管理者
type TableMgr struct {
	MapTable map[int]*Table
	lock     *sync.RWMutex
	TableIds *static.LinkedList
}

var tablemgrSingleton *TableMgr = nil

//! 得到牌桌管理
func GetTableMgr() *TableMgr {
	if tablemgrSingleton == nil {
		tablemgrSingleton = new(TableMgr)
		tablemgrSingleton.MapTable = make(map[int]*Table)
		tablemgrSingleton.lock = new(sync.RWMutex)
		tablemgrSingleton.TableIds = static.NewLinkedList()
		for i := MIN_TABLE_ID; i <= MAX_TABLE_ID; i++ {
			tablemgrSingleton.TableIds.Prepend(static.NewINode(static.ElementType(i), nil))
		}
	}
	return tablemgrSingleton
}

//! 获取牌桌
func (self *TableMgr) GetRandomTableId() int {
	index := static.HF_GetRandom(self.TableIds.Length())
	value, suc := self.TableIds.FindByIndex(index)
	if suc {
		return int(value.X)
	}

	return 0
}

//！得到当前的牌卓数量
//func (self *TableMgr) GetTableNum() int {
//	self.lock.Lock()
//	defer self.lock.Unlock()
//	for key, _ := range self.MapTable {
//		syslog.Logger().Debug("当前房间:", key)
//	}
//	return len(self.MapTable)
//}

// 场次新建桌子
func (self *TableMgr) NewTable(c *static.Msg_HG_CreateTable) *static.Table {
	nowTime := time.Now()
	table := new(static.Table)
	//获取牌桌id
	table.Id = c.Id
	table.NTId = c.NTId
	table.HId = c.HId
	table.DHId = c.DHId
	table.FId = c.FId
	table.NFId = c.NFId
	table.IsCost = false
	table.BXiaPaoIng = false
	table.Creator = c.Creator
	table.CreateType = c.CreateType
	table.GameId = GetServer().Con.Id
	table.KindId = c.KindId
	table.Users = make([]*static.TableUser, c.MaxPlayerNum)
	table.GameNum = fmt.Sprintf("%d%02d%02d%02d%02d%02d_%d_%d", nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), nowTime.Minute(), nowTime.Second(), GetServer().Con.Id, table.Id)
	table.Config = &static.TableConfig{
		MaxPlayerNum: c.MaxPlayerNum,
		MinPlayerNum: c.MinPlayerNum,
		RoundNum:     c.RoundNum,
		CardCost:     c.CardCost,
		CostType:     c.CostType,
		View:         c.View,
		Restrict:     c.Restrict,
		GameConfig:   c.GameConfig,
		MatchConfig:  c.MatchConfig,
		GVoice:       c.GVoice,
		GameType:     2, // 默认好友房
	}
	if CanBeFewer(table) {
		table.Config.FewerStart = c.FewerStart
	} else {
		table.Config.FewerStart = "false"
	}
	return table
}

//! 创建个房间
func (self *TableMgr) CreateTable(c *static.Table) *static.Table {
	self.lock.Lock()
	defer self.lock.Unlock()

	gameTable := new(Table)
	gameTable.Table = c
	//if gameTable.IsTeaHouse() {
	//	house, err := gameTable.GetHouse()
	//	if err == nil {
	//		gameTable.IsHidHide = house.IsHidHide
	//		floor, err := gameTable.GetFloor()
	//		if err == nil {
	//			gameTable.IsVitamin = house.IsVitamin && floor.IsVitamin
	//			gameTable.IsAiSuper =
	//				// 混排激活总开关
	//				house.MixActive &&
	//					// 楼层是否为混排楼层
	//					floor.IsMix &&
	//					// 防作弊模式
	//					house.TableJoinType == constant.NoCheat &&
	//					// 超级防作弊模式
	//					house.AiSuper
	//		}
	//	}
	//}
	gameTable.Init()
	if _, ok := self.MapTable[c.Id]; ok {
	}
	self.MapTable[c.Id] = gameTable
	self.TableIds.Remove(static.ElementType(c.Id))

	// 写入redis
	gameTable.flush()
	return c
}

//! 添加牌桌
func (self *TableMgr) AddTable(table *Table) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.MapTable[table.Id] = table
	self.TableIds.Remove(static.ElementType(table.Id))
}

//! 删除房间
func (self *TableMgr) DelTable(table *Table) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if table.Config.GameType != static.GAME_TYPE_FRIEND { //金币场，桌子不释放
		return
	}

	table.WriteTableLog(static.INVALID_CHAIR, "DelTable 删除牌桌的函数被调用") //大厅也在调用这个函数，避免漏掉
	xlog.Logger().Warnln("DelTable 删除牌桌的函数被调用", table.Id)
	//begin：牌桌无法进入的问题定位日志
	namestr := fmt.Sprintf("Tid:%d ", table.Id)
	namestr += "游戏者有："
	for i := 0; i < len(table.Person); i++ {
		// 跳过空位
		if table.Person[i] == nil {
			continue
		}
		namestr += fmt.Sprintf("%s ", table.Person[i].Name)
	}
	namestr += "旁观者有："
	for i := 0; i < len(table.LookonPerson[0]); i++ {
		// 跳过空位
		if table.LookonPerson[0][i] == nil {
			continue
		}
		namestr += fmt.Sprintf("%s ", table.LookonPerson[0][i].Name)
	}
	table.WriteTableLog(static.INVALID_CHAIR, namestr)
	xlog.Logger().Warnln(namestr)

	// 冻结房卡还原
	if !table.IsCost {
		// 更新房卡
		//tx := GetDBMgr().db_M.Begin()
		tx := GetDBMgr().GetDBmControl() //20210223 这里经常锁死，去掉事务，
		var aftfka int
		var err error
		if table.Table.LeagueID > 0 {
			leagueID := table.Table.LeagueID
			cost := -1 * table.Config.CardCost
			_, _, _, aftfka, err = updLeagueCard(table.Table.LeagueID, table.Creator, 0, -1*table.Config.CardCost, 0, tx)
			cli := GetDBMgr().Redis
			err := cli.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID), "freeze_card", int64(cost)).Err()
			if err != nil {
				//tx.Rollback()
				return
			}
			if table.Table.NotPool {
				sql := `update league_user set freeze_card = freeze_card +? where league_id = ? and uid = ?`
				tx.Exec(sql, cost, leagueID, table.Creator)
				err := cli.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER_H, table.Table.Creator), "freeze_card", int64(cost)).Err()
				if err != nil {
					//tx.Rollback()
					return
				}
			}
			//err = tx.Commit().Error
			if err != nil {
				xlog.Logger().Errorln(err)
				sql := `update league_user set freeze_card = freeze_card +? where league_id = ? and uid = ?`
				if err = GetDBMgr().GetDBmControl().Exec(sql, cost, leagueID, table.Creator).Error; err != nil {
					xlog.Logger().Errorf("error:%+v", err)
				}
				err := cli.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID), "freeze_card", int64(-1*cost)).Err()
				if err != nil {
					return
				}
				if table.Table.NotPool {
					err := cli.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER_H, table.Table.Creator), "freeze_card", int64(-1*cost)).Err()
					if err != nil {
						return
					}
				}
				return
			}
			msg := static.MsgLeagueCardAdd{LeagueID: leagueID, Uid: table.Table.Creator, NotPool: table.Table.NotPool, UpdCount: int64(table.Config.CardCost)}
			if _, err := GetServer().CallHall("ServerMethod.ServerMsg", consts.MsgTypeLeagueCardAdd, xerrors.SuccessCode, &msg, 0); err != nil {
				xlog.Logger().Errorf("通知大厅更新卡池状态失败%v", err)
			}
		} else {
			if consts.ClubHouseOwnerPay {
				_, _, _, aftfka, err = updcard(table.Creator, 0, -1*table.Config.CardCost, tx, 0)
				if err != nil {
					xlog.Logger().Errorln(err)
					//tx.Rollback()
				} else {
					//tx.Commit()
				}

				// 更新redis
				GetDBMgr().db_R.UpdatePersonAttrs(table.Creator, "FrozenCard", aftfka)
			} else {
				for _, tableUser := range table.Users {
					if tableUser != nil {
						if tableUser.Uid > 0 && tableUser.Payer > 0 {
							_, _, _, aftfka, err = updcard(tableUser.Payer, 0, -1*table.Config.CardCost, tx, 0)
							if err != nil {
								xlog.Logger().Errorln(err)
							}
							// 更新redis
							GetDBMgr().db_R.UpdatePersonAttrs(table.Creator, "FrozenCard", aftfka)
						}
					}
				}
			}
		}
	}
	//end：牌桌无法进入的问题定位日志
	// 删除内存
	delete(self.MapTable, table.Id)
	// 删除redis
	table.remove()
	// 还原链表
	self.TableIds.Prepend(static.NewINode(static.ElementType(table.Id), nil))
	// fmt.Println("删除一个房间：", table.Id)
}

//! 获取牌桌
func (self *TableMgr) GetTable(id int) *Table {
	self.lock.RLock()
	defer self.lock.RUnlock()
	table, ok := self.MapTable[id]
	if !ok {
		return nil
	}
	return table
}

//! 随机获取牌桌
func (self *TableMgr) GetRandTable(site *Site, id int) *Table {
	self.lock.RLock()
	defer self.lock.RUnlock()

	index := -1
	players := -1
	temp := 0

	for i := 0; i < site.tableNum && i < len(site.TableIds); i++ {
		if site.TableIds[i] == id { //和上一局不同桌，不同场次的排除
			continue
		}
		v, ok := self.MapTable[site.TableIds[i]]
		if ok {
			temp = v.GetPersonNumber()
			if temp < v.Config.MaxPlayerNum && players < temp {
				index = site.TableIds[i]
				players = temp

				if players == v.Config.MaxPlayerNum-1 { //轮循人数最多的桌子
					break
				}
			}
		}
	}

	if index > 0 {
		return self.MapTable[index]
	}

	return nil
}

//！应要求，根据uid强制解散牌桌
func (self *TableMgr) CompulsoryDissmiss(msg *static.Msg_UserSeat) error {
	// 根据tableid从redis中找到桌子及相应的玩家
	tablekey := fmt.Sprintf(consts.REDIS_KEY_TABLEINFO, GetServer().Con.Id, msg.TableId)
	data, err := GetDBMgr().GetDBrControl().Get(tablekey)

	if err != nil {
		return fmt.Errorf("get table from redis error: key:%s,err:%s", tablekey, err.Error())
	}
	table := new(Table)

	err = json.Unmarshal(data, table)
	if err != nil {
		return fmt.Errorf("unmarshal game tableinfo failed: %s", err.Error())
	}
	//旁观玩家需要处理
	if tablelookon := self.GetTable(msg.TableId); tablelookon != nil {
		for _, tableLookonUser := range tablelookon.LookonPerson[0] {
			if tableLookonUser == nil {
				continue
			}
			r_person, err := GetDBMgr().GetDBrControl().GetPerson(tableLookonUser.Uid)
			if err == nil && r_person != nil {
				//syslog.Logger().Debug("CompulsoryDissmiss:清掉玩家redis游戏信息：",r_person.Uid)
				r_person.WatchTable = 0
				if tablelookon.HId == 0 && tablelookon.CreateType == consts.CreateTypeSelf && tablelookon.Creator == tableLookonUser.Uid {
					r_person.CreateTable = 0
				}
				err = GetDBMgr().db_R.UpdatePersonAttrs(r_person.Uid, "WatchTable", r_person.WatchTable, "TableId", r_person.TableId, "GameId", r_person.GameId, "CreateTable", r_person.CreateTable)
				if err != nil {
					return fmt.Errorf("Clear PlayerInfo[uid:%d] game information from redis error:%s", r_person.Uid, err.Error())
				}
			}
			mem_person := GetPersonMgr().GetLookonPerson(tableLookonUser.Uid)
			if mem_person != nil {
				//syslog.Logger().Debug("CompulsoryDissmiss:清掉玩家内存游戏信息：",mem_person.Info.Uid)
				mem_person.SendMsg("forcetabledel", xerrors.SuccessCode, &static.Msg_S2C_TableDel{forceCloseByServerGM, "您所在的牌桌因数据异常被强制解散。"})
				mem_person.CloseSession(consts.SESSION_CLOED_FORCE)
				GetPersonMgr().DelLookonPerson(tableLookonUser.Uid)
			}
		}
	}

	for _, tableUser := range table.Users {
		if tableUser == nil {
			continue
		}
		r_person, err := GetDBMgr().GetDBrControl().GetPerson(tableUser.Uid)
		if err == nil && r_person != nil {
			//syslog.Logger().Debug("CompulsoryDissmiss:清掉玩家redis游戏信息：",r_person.Uid)
			r_person.TableId = 0
			r_person.GameId = 0
			if table.HId == 0 && table.CreateType == consts.CreateTypeSelf && table.Creator == tableUser.Uid {
				r_person.CreateTable = 0
			}
			err = GetDBMgr().db_R.UpdatePersonAttrs(r_person.Uid, "TableId", r_person.TableId, "GameId", r_person.GameId, "CreateTable", r_person.CreateTable)
			if err != nil {
				return fmt.Errorf("Clear PlayerInfo[uid:%d] game information from redis error:%s", r_person.Uid, err.Error())
			}
		}
		mem_person := GetPersonMgr().GetPerson(tableUser.Uid)
		if mem_person != nil {
			//syslog.Logger().Debug("CompulsoryDissmiss:清掉玩家内存游戏信息：",mem_person.Info.Uid)
			mem_person.SendMsg("forcetabledel", xerrors.SuccessCode, &static.Msg_S2C_TableDel{forceCloseByServerGM, "您所在的牌桌因数据异常被强制解散。"})
			mem_person.CloseSession(consts.SESSION_CLOED_FORCE)
			GetPersonMgr().DelPerson(tableUser.Uid)
		}
	}
	// redis成员清理完毕后 清掉房间key
	_, _ = GetDBMgr().GetDBrControl().Remove(tablekey)

	if table = self.GetTable(msg.TableId); table != nil {
		// if table.IsBegin() {
		// 	table.game.OnEnd()
		// }
		table.WriteTableLog(uint16(msg.Seat), "CompulsoryDissmiss 牌桌被手动解散")
		xlog.Logger().Warnln("CompulsoryDissmiss 牌桌被手动解散", table.Id)
		self.DelTable(table)
	}

	return nil
}

// 强制解散所有未开始的游戏房间(当kindId=0时解散所有)
func (self *TableMgr) ForceCloseAllTable() {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, table := range self.MapTable {
		// if kindId == 0 || table.KindId == kindId {
		if !table.IsBegin() {
			msg := new(static.Msg_S2C_TableDel)
			msg.Type = forceCloseByServerMaintain
			msg.Msg = "服务器维护, 强制解散房间"
			table.Operator(base2.NewTableMsg(consts.MsgTypeTableDel, "now", 0, msg))
		}
		// }
	}
}

func (self *TableMgr) reflushTable() {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, v := range self.MapTable {
		v.pause()
	}

	time.Sleep(time.Millisecond * 200)

	for _, v := range self.MapTable {
		if v.Config.GameType == static.GAME_TYPE_FRIEND {
			// 好友房牌桌数据保存
			v.flushInfo()
		} else if v.Config.GameType == static.GAME_TYPE_GOLD {
			// 金币房数据清除
			v.remove()
			// 用户数据清除
			for _, p := range v.Users {
				if p != nil {
					GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "SiteId", 0, "TableId", 0, "GameId", 0)
				}
			}
		}
	}
}
