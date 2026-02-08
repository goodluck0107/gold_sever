package center

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

const (
	MAX_TABLE_ID = 999999
	MIN_TABLE_ID = 100000
)

//! 牌桌的数据结构
type Table struct {
	*static.Table
	lock *lock2.RWMutex
}

func (self *Table) UserCount() int {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	count := 0
	for _, u := range self.Users {
		if u != nil {
			count++
		}
	}
	return count
}

func (self *Table) GetUser(uid int64) *static.TableUser {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()
	for _, u := range self.Users {
		if u == nil {
			continue
		}
		if u.Uid == uid {
			return u
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////
//! 玩家管理者
type TableMgr struct {
	MapTable     map[int]*Table
	GameTableMap map[int]int // 记录游戏服对应的桌子数量
	lock         *lock2.RWMutex
	TableIds     *static.LinkedList
}

var tablemgrSingleton *TableMgr = nil

//! 得到牌桌管理单例
func GetTableMgr() *TableMgr {

	if tablemgrSingleton == nil {
		tablemgrSingleton = new(TableMgr)
		tablemgrSingleton.MapTable = make(map[int]*Table)
		tablemgrSingleton.GameTableMap = make(map[int]int)
		tablemgrSingleton.lock = new(lock2.RWMutex)

		tablemgrSingleton.TableIds = static.NewLinkedList()
		for i := MIN_TABLE_ID; i <= MAX_TABLE_ID; i++ {
			tablemgrSingleton.TableIds.Prepend(static.NewINode(static.ElementType(i), nil))
		}
	}

	return tablemgrSingleton
}

//! 恢复牌桌数据
func (self *TableMgr) Restore() {
	//self.lock.CustomLock()
	//defer self.lock.CustomUnLock()

	// 获取所有table key
	keys, err := GetDBMgr().db_R.Keys(consts.REDIS_KEY_TABLEINFO_ALL)
	if err != nil {
		xlog.Logger().Errorln("restore hall table failed: ", err.Error())
		return
	}
	for _, key := range keys {
		data, err := GetDBMgr().db_R.Get(key)
		if err != nil {
			xlog.Logger().Errorln("get hall tableinfo failed:%v ", err.Error())
			continue
		}
		var table Table
		err = json.Unmarshal(data, &table)
		if err != nil {
			xlog.Logger().Errorln("unmarshal hall tableinfo failed: ", err.Error())
			continue
		}

		// 只恢复好友房的桌子
		if table.Config.GameType == static.GAME_TYPE_FRIEND {
			table.lock = new(lock2.RWMutex)

			flag := false
			for _, u := range table.Users {
				if u == nil {
					continue
				}

				person, err := GetDBMgr().db_R.GetPerson(u.Uid)
				if err != nil {
					xlog.Logger().Errorln("person not exists: ", u.Uid)
					break
				}
				flag = true
				hp := new(PlayerCenterMemory)
				hp.Info = *person
				// 添加person
				//if person.Online {
				GetPlayerMgr().AddPerson(hp)
				//}
			}

			if flag {
				// 添加桌子
				self.AddTable(&table)
				self.TableIds.Remove(static.ElementType(table.Id))
			} else {
				xlog.Logger().Errorf("table data error: %d,%s", table.Id, data)
			}
		}
	}

	//大厅启动等待游戏服务牌桌数据恢复,开始计算在线人数
	go GetPlayerMgr().calculateOnline()
}

// 添加游戏服牌桌数量
func (self *TableMgr) UpdateGameTableNum(gameId int, num int) {
	v, _ := self.GameTableMap[gameId]
	v = v + num
	self.GameTableMap[gameId] = v
}

// 获取游戏服牌桌数量
func (self *TableMgr) GetGameTableNum(gameId int) int {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()
	return self.GameTableMap[gameId]
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

//! 添加牌桌
func (self *TableMgr) AddTable(table *Table) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()
	if table.CreateStamp == 0 {
		table.Table.CreateStamp = time.Now().UnixNano()
	}
	self.MapTable[table.Id] = table
	self.UpdateGameTableNum(table.GameId, 1)
}

//! 删除牌桌
func (self *TableMgr) DelTable(table *Table) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()
	table.remove()
	// 还原链表
	self.TableIds.Prepend(static.NewINode(static.ElementType(table.Id), nil))
	self.UpdateGameTableNum(table.GameId, -1)
	delete(self.MapTable, table.Id)
}

//! 获取牌桌
func (self *TableMgr) GetTable(tableid int) *Table {
	if tableid <= 0 {
		return nil
	}
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()
	table, ok := self.MapTable[tableid]
	if ok {
		return table
	}
	return nil
}

//! 创建牌桌
func (self *TableMgr) CreateTable(table *static.Table) *Table {

	hallTable := new(Table)
	hallTable.Table = table
	hallTable.lock = new(lock2.RWMutex)

	self.AddTable(hallTable)

	return hallTable
}

// 写入redis
func (self *Table) flush() {
	GetDBMgr().db_R.Set(self.GetRedisKey(), self.ToBytes())
}

//! del table
func (self *Table) remove() {
	GetDBMgr().db_R.Remove(self.GetRedisKey())
}

// 用户离开桌子
func (self *Table) UserLeaveTable(uid int64) {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	for i, u := range self.Users {
		if u != nil && u.Uid == uid {
			self.Users[i] = nil
			// 更新内存
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(u.Uid, "TableId", 0, "GameId", 0)
			GetClubMgr().UserOffLine(u.Uid)
			person := GetPlayerMgr().GetPlayer(uid)
			if person != nil {
				person.Info.TableId = 0
				person.Info.GameId = 0
				if !self.IsTeaHouse() && self.CreateType == consts.CreateTypeSelf && self.Creator == person.Info.Uid {
					person.Info.CreateTable = 0
				}
			}
		}
	}
}

func (self *Table) CheckRestartJoin(p *static.Person, ip string, seat int) *xerrors.XError {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()
	// 校验seat
	if seat >= self.Config.MaxPlayerNum {
		return xerrors.ArgumentError
	}
	// 判断用户有没有坐下
	for _, u := range self.Users {
		if u != nil && u.Uid == p.Uid {
			// 已经坐下的用户不做任何处理
			return nil
		}
	}

	isFull := true
	for _, u := range self.Users {
		//fmt.Println(fmt.Sprintf("%+v", u))
		if u == nil {
			isFull = false
			break
		}
	}
	if isFull {
		return xerrors.TableIsFullError
	}

	//str := fmt.Sprintf("这里检查入桌限制 CheckRestartJoin 桌子ID：%d , 最大开始人数：%d , 是否需要验证IP：%v", self.Id, self.Config.MaxPlayerNum, self.Config.Restrict)
	//syslog.Logger().Errorf(str)

	// 判断ip限制
	if self.Config.Restrict { //2人玩 不做IP  GPS 限制
		//
		if self.Config.MaxPlayerNum != 2 {
			for _, u := range self.Users {
				if u != nil {
					tperson := GetPlayerMgr().GetPlayer(u.Uid)
					var tIp string
					if tperson == nil {
						xlog.Logger().Errorf("person not online:%d ", u.Uid)
						tp, _ := GetDBMgr().GetDBrControl().GetPerson(u.Uid)
						tIp = tp.Ip
					} else {
						tIp = tperson.Ip
					}
					if tIp == ip {
						cuserror := xerrors.NewXError("同IP限制，无法进入房间")
						return cuserror
					}
				}
			}
		}
	}
	return nil
}

// 用户加入桌子 seat=-1时顺序落座
func (self *Table) CheckUserJoinTable(p *static.Person, seat int) *xerrors.XError {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	cp := GetPlayerMgr().GetPlayer(p.Uid)
	if cp == nil {
		xlog.Logger().Warnf("person not online:%d ", p.Uid) // 可能来自换桌
		// return xerrors.TableInError
	}

	// 校验seat
	if seat >= self.Config.MaxPlayerNum {
		return xerrors.ArgumentError
	}

	// 判断用户有没有坐下
	for _, u := range self.Users {
		if u != nil && u.Uid == p.Uid {
			// 已经坐下的用户不做任何处理
			return nil
		}
	}

	// 判断桌子有没有坐满
	isFull := true
	for _, u := range self.Users {
		// fmt.Println(fmt.Sprintf("%+v", u))
		if u == nil {
			isFull = false
			break
		}
	}
	if isFull {
		return xerrors.TableIsFullError
	}

	//str := fmt.Sprintf("这里检查入桌限制 CheckUserJoinTable 桌子ID：%d , 最大开始人数：%d , 是否需要验证IP：%v", self.Id, self.Config.MaxPlayerNum, self.Config.Restrict)
	//syslog.Logger().Errorf(str)

	// 判断ip限制
	if self.Config.Restrict {
		if self.Config.MaxPlayerNum != 2 { //2人玩 不做IP GPS 限制
			for _, u := range self.Users {
				if u != nil {
					tperson := GetPlayerMgr().GetPlayer(u.Uid)
					var tIp string
					if tperson == nil {
						xlog.Logger().Warnf("person not online:%d ", u.Uid)
						tp, _ := GetDBMgr().GetDBrControl().GetPerson(u.Uid)
						tIp = tp.Ip
					} else {
						tIp = tperson.Ip
					}
					var ip string
					if cp == nil {
						ip = p.Ip
					} else {
						ip = cp.Ip
					}
					if tIp == ip {
						cuserror := xerrors.NewXError("同IP限制，无法进入房间")
						return cuserror
					}
				}
			}
		}
	}
	return nil
}

//! 获取牌桌信息
func (self *Table) getTableMsg() *static.TableInfoDetail {
	var msg static.TableInfoDetail
	msg.TableId = self.Id
	for i := 0; i < len(self.Users); i++ {
		// 跳过空位
		if self.Users[i] == nil {
			continue
		}

		var son static.Son_PersonInfo
		person := GetPlayerMgr().GetPlayer(self.Users[i].Uid)
		if person != nil {
			son.Uid = person.Info.Uid
			son.Name = person.Info.Nickname
			son.ImgUrl = person.Info.Imgurl
			son.Sex = person.Info.Sex
			son.Seat = i
		} else {
			p, err := GetDBMgr().GetDBrControl().GetPerson(self.Users[i].Uid)
			if err != nil {
				xlog.Logger().Errorln(err)
				continue
			}
			son.Uid = p.Uid
			son.Name = p.Nickname
			son.ImgUrl = p.Imgurl
			son.Sex = p.Sex
			son.Seat = i
		}

		msg.Person = append(msg.Person, son)
	}
	msg.Creator = self.Creator
	msg.HId = self.HId
	msg.NFId = self.NFId
	msg.NTId = self.NTId
	msg.Step = self.Step
	msg.KindId = self.KindId
	msg.GameConfig = self.GetGameConfig()

	return &msg
}

func (self *Table) Kick(uid, opt int64) error {
	if self.Begin {
		return fmt.Errorf("游戏已开始。")
	}
	if self.Step > 0 {
		return fmt.Errorf("游戏已开始。")
	}
	u := self.GetUser(uid)
	if u == nil {
		return fmt.Errorf("玩家不存在。")
	}
	msg := static.MsgHouseTableUserKick{
		Uid: uid,
		Tid: self.Id,
		Opt: opt,
	}
	// 通知游戏服退出牌桌
	reply, err := GetServer().CallGame(self.GameId, uid, "NewServerMsg", consts.MsgTypeTableUserKick, xerrors.SuccessCode, &msg)
	if err != nil {
		return err
	}
	if res := string(reply); res != "SUC" {
		return fmt.Errorf(res)
	}
	// 删除内存数据
	self.UserLeaveTable(uid)
	if house := GetClubMgr().GetClubHouseById(self.DHId); house != nil {
		msgExit := new(static.GH_TableExit_Ntf)
		msgExit.Uid = uid
		msgExit.GameId = self.GameId
		msgExit.TableId = self.Id
		msgExit.KindId = self.KindId
		oldFloor := house.GetFloorByFId(self.FId)
		ChOptTableOut_Ntf(oldFloor, nil, nil, msgExit)
	}
	return nil
}
