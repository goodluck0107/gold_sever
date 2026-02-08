package center

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xerrors"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

//////////////////////////////////////////////////////////////
//! 玩家管理者
type PlayerMgr struct {
	MapPerson       map[int64]*PlayerCenterMemory
	onlineNumberMap map[int]*GameServerOnline // 在线人数结构 game_id - kind_id - site_type - 在线人数
	offlineMap      map[int64]int64           // 记录将要离线的user key为uid, value为最后活跃时间戳
	lock            *lock2.RWMutex
	userOnlineChan  chan int64 // 用户上线channel
	userOfflineChan chan int64 // 用户下线channel
}

// 游戏服在线
type GameServerOnline struct {
	Online    int                   // 总在线人数
	KindIdMap map[int]*KindIdOnline // 子游戏在线
}

// 子游戏在线
type KindIdOnline struct {
	Online      int         // 总在线人数
	SiteTypeMap map[int]int // 游戏场次在线
}

const (
	offlineTimeout = 3 // 换服 离线超时10s
)

var playerMgrSingleton *PlayerMgr = nil

func GetPlayerMgr() *PlayerMgr {
	if playerMgrSingleton == nil {
		playerMgrSingleton = new(PlayerMgr)
		playerMgrSingleton.MapPerson = make(map[int64]*PlayerCenterMemory)
		playerMgrSingleton.onlineNumberMap = make(map[int]*GameServerOnline)
		playerMgrSingleton.offlineMap = make(map[int64]int64)
		playerMgrSingleton.lock = new(lock2.RWMutex)
		playerMgrSingleton.userOnlineChan = make(chan int64, 999)
		playerMgrSingleton.userOfflineChan = make(chan int64, 999)

		// 判断用户在线离线状态
		//go playerMgrSingleton.checkUserOnline()
		// 定时获取在线人数
		//go playerMgrSingleton.calculateOnline()

		//go playerMgrSingleton.HeartBeating()
	}

	return playerMgrSingleton
}

// 检测用户离桌定时器
func (self *PlayerMgr) checkUserOnline() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			// 用户超时没响应则判定用户离线
			timeNow := time.Now().Unix()

			for k, v := range self.offlineMap {
				if timeNow-v > offlineTimeout {
					// user map清除
					p := self.GetPlayer(k)
					if p != nil && p.session == nil {
						GetClubMgr().UserOffLine(p.Info.Uid)
						if p.Info.TableId == 0 {
							self.DelPerson(k)
						}
					}
					delete(self.offlineMap, k)
				}
			}
		case uid := <-self.userOnlineChan:
			// 用户加入游戏房间, 判定用户在线
			xlog.Logger().Infof("[玩家上线:%d]", uid)
			delete(self.offlineMap, uid)
			self.userOnline(uid)
		case uid := <-self.userOfflineChan:
			// 用户下线开始计时,延迟判定用户离线
			xlog.Logger().Infof("[玩家离线:%d]", uid)
			self.offlineMap[uid] = time.Now().Unix()
			self.userOffline(uid)
		}
	}
}

// 用户上线
func (self *PlayerMgr) userOnline(uid int64) {
	p, err := GetDBMgr().GetDBrControl().GetPerson(uid)
	if p == nil || err != nil {
		return
	}
	GetPlayerMgr().AddPerson(&PlayerCenterMemory{Info: *p})
}

// 用户离线
func (self *PlayerMgr) userOffline(uid int64) {
	//syslog.Logger().Debug("=========================================================userOffline")
	//self.userOfflineChan <- uid
	GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "LastOffLineTime", time.Now().Unix(), "Online", false)
}

// 判断person是否在线
func (self *PlayerMgr) IsUserOnline(uid int64) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	if _, ok := self.MapPerson[uid]; ok {
		return true
	}
	return false
}

//! 加入玩家
func (self *PlayerMgr) AddPerson(person *PlayerCenterMemory) {
	self.lock.CustomLock()
	self.MapPerson[person.Info.Uid] = person
	self.lock.CustomUnLock()
	// GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Info.Uid, "Online", true)
	// GetClubMgr().UserOnline(&person.Info)
}

//! 删玩家
func (self *PlayerMgr) DelPerson(uid int64) {
	self.lock.CustomLock()
	delete(self.MapPerson, uid)
	self.lock.CustomUnLock()
	GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "LastOffLineTime", time.Now().Unix(), "Online", false)
	GetClubMgr().UserOffLine(uid)
}

//! 该玩家是否存在
func (self *PlayerMgr) GetPlayer(uid int64) *PlayerCenterMemory {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	person, ok := self.MapPerson[uid]
	if ok {
		return person
	}
	return nil
}

//! 该玩家是否存在e
func (self *PlayerMgr) GetPlayerWithOpenId(openid string) *PlayerCenterMemory {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	for _, person := range self.MapPerson {
		if person.Info.Openid == openid {
			return person
		}
	}
	return nil
}

// 为所有在线用户推送消息
func (self *PlayerMgr) Broadcast(head string, msg interface{}) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, person := range self.MapPerson {
		if person != nil {
			s := person
			go s.SendMsg(head, msg)
		}
	}
}

func (self *PlayerMgr) BroadcastV2(head string, msg interface{}) {
	buf := static.HF_EncodeMsg(head, xerrors.SuccessCode, msg, GetServer().Con.Encode, GetServer().Con.EncodeClientKey, 0)
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, person := range self.MapPerson {
		if person != nil {
			go person.SendBuf(buf)
		}
	}
}

//! 获取在线人数
func (self *PlayerMgr) GetOnlineNumber() {
	self.lock.RLock()
	// 获取各游戏在线人数
	onlineMap := make(map[int]*GameServerOnline)
	for _, person := range self.MapPerson {
		if person != nil {
			// 游戏服人数
			if person.Info.GameId != 0 {
				gameObj, ok := onlineMap[person.Info.GameId]
				if !ok {
					// 初始化
					gameObj = new(GameServerOnline)
					gameObj.KindIdMap = make(map[int]*KindIdOnline)
				}
				// 游戏服总人数+1
				gameObj.Online = gameObj.Online + 1

				var kindId int
				var siteType int
				// 先判断场次信息, 再判断桌子信息
				if person.Info.SiteId != 0 {
					kindId = person.Info.SiteId / 100
					siteType = person.Info.SiteId % 100
				} else {
					if person.Info.TableId != 0 {
						table := GetTableMgr().GetTable(person.Info.TableId)
						if table != nil {
							kindId = table.KindId
							siteType = table.SiteType
						}
					}
				}

				// 更新人数
				if kindId != 0 {
					kindidObj, ok := gameObj.KindIdMap[kindId]
					if !ok {
						kindidObj = new(KindIdOnline)
						kindidObj.SiteTypeMap = make(map[int]int)
						gameObj.KindIdMap[kindId] = kindidObj
					}
					// 子游戏人数+1
					kindidObj.Online = kindidObj.Online + 1
					// 场次人数+1
					kindidObj.SiteTypeMap[siteType] = kindidObj.SiteTypeMap[siteType] + 1
					gameObj.KindIdMap[kindId] = kindidObj
				}

				onlineMap[person.Info.GameId] = gameObj
			}
		}
	}
	self.lock.RUnlock()

	// 写数据
	self.onlineNumberMap = onlineMap
}

// 获取所有子游戏在线人数
func (self *PlayerMgr) GetAllKindIdOnlineNumber() map[int]int {
	self.lock.RLock()
	defer self.lock.RUnlock()

	result := make(map[int]int)

	// 获取子游戏人数
	for _, gameObj := range self.onlineNumberMap {
		for kindid, kindidObj := range gameObj.KindIdMap {
			if v, ok := result[kindid]; ok {
				result[kindid] = v + kindidObj.Online
			} else {
				result[kindid] = kindidObj.Online
			}
		}
	}

	// 获取总人数
	result[0] = 0
	for _, person := range self.MapPerson {
		if person.session != nil || person.Info.TableId != 0 {
			result[0]++
		}
	}

	return result
}

// 获取子游戏某场次的在线人数(当siteType=0时获取该游戏下的所有人数)
func (self *PlayerMgr) GetOnlineNumberByKindId(kindId int, siteType int) int {
	result := 0
	for _, gameObj := range self.onlineNumberMap {
		if kindidObj, ok := gameObj.KindIdMap[kindId]; ok {
			if siteType == 0 {
				// 获取子游戏的所有人数
				result = result + kindidObj.Online
			} else {
				// 获取子游戏该场次的人数
				result = result + kindidObj.SiteTypeMap[siteType]
			}
		}
	}
	return result
}

//! 计算在线人数
func (self *PlayerMgr) calculateOnline() {
	for {
		// 每隔一段时间计算当前在线人数
		self.GetOnlineNumber()
		//// 广播在线人数变化(小鱼游使用,chess注释掉)
		//self.notifyGameOnlineChange()
		// 一分钟后在计算
		time.Sleep(1 * time.Minute)
	}
}

// 推送消息告知用户在线人数变化
func (self *PlayerMgr) notifyGameOnlineChange() {
	// 先计算好各个区域下游戏的人数情况
	areaOnline := make(map[string][]*static.Msg_S2C_GameOnline)
	areas := GetAreaPackagesByKind(static.AreaPackageKindGold)
	for _, area := range areas {
		onlineArr := make([]*static.Msg_S2C_GameOnline, 0)
		// 获取区域下所有游戏
		for _, game := range area.Games {
			obj := new(static.Msg_S2C_GameOnline)
			obj.KindId = game.KindId
			siteTypes := GetServer().GetSiteTypeByKindId(static.GAME_TYPE_GOLD, game.KindId)
			totalOnline := 0
			for _, item := range siteTypes {
				v := &static.Msg_S2C_SiteTypeOnline{
					Type: item,
				}
				// 获取房间配置
				c := GetServer().GetRoomConfig(obj.KindId, item)
				if c != nil {
					// 获取该游戏该场次的在线人数(加上基础人数)
					v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type) + c.BaseNum
				} else {
					// 获取该游戏该场次的在线人数(加上基础人数)
					v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type)
				}
				totalOnline = totalOnline + v.Online
				obj.SiteType = append(obj.SiteType, v)
			}
			if len(obj.SiteType) > 0 {
				//obj.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, 0)
				obj.Online = totalOnline
				onlineArr = append(onlineArr, obj)
			}
		}
		areaOnline[area.Code] = onlineArr
	}

	self.lock.RLock()
	defer self.lock.RUnlock()

	// 只给进入区域的用户推送人数变化
	for _, p := range self.MapPerson {
		if p == nil || p.Info.Area == "" {
			continue
		}
		v := areaOnline[p.Info.Area]
		if len(v) > 0 {
			p.SendMsg(consts.MsgTypeGameOnline, v)
		}
	}
}

// 服务器变化通知客户端(area为空时, 通知所有区域的用户变化)
func (self *PlayerMgr) NotifyGameServerChange(area []string) {
	// // 先计算一次游戏列表
	// gameList := make(map[string][]*public.Msg_S2C_GameList)
	// areaList := GetAreaPackages()
	// for _, area := range areaList {
	// 	arr := getGameList(area.Code)
	// 	if len(arr) > 0 {
	// 		gameList[area.Code] = arr
	// 	}
	// }
	//
	// isInArray := func(str string, arr []string) bool {
	// 	for _, item := range arr {
	// 		if item == str {
	// 			return true
	// 		}
	// 	}
	// 	return false
	// }
	//
	// self.lock.RLock()
	// defer self.lock.RUnlock()
	// for _, p := range self.MapPerson {
	// 	if p == nil || p.Info.Area == "" {
	// 		continue
	// 	}
	// 	if len(area) == 0 || isInArray(p.Info.Area, area) {
	// 		if v, ok := gameList[p.Info.Area]; ok {
	// 			p.SendMsg(constant.MsgTypeGameList_Ntf, v)
	// 		}
	// 	}
	// }
}

// 区域用户广播(area不可为空)
func (self *PlayerMgr) SendAreaBroadcast(area string, head string, v interface{}) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	for _, p := range self.MapPerson {
		if p == nil || p.Info.Area == "" {
			continue
		}
		if p.Info.Area == area {
			p.SendMsg(head, v)
		}
	}
}

//
func (self *PlayerMgr) GetPersonWithLoginTime(uid int64) int64 {
	user := models.User{Id: uid}
	err := GetDBMgr().GetDBmControl().Model(user).First(&user).Error
	if err != nil {
		return 0
	}
	return user.LastLoginAt.Unix()
}

// 公用的给玩家发送通知，客户端会弹出面板显示且只有确定按钮
// save 如果玩家不在线 是否保存在redis等待玩家上线的时候推送
func (self *PlayerMgr) SendNotify(uid int64, save bool, title, content string) {
	var msg static.Msg_HC_NotifyPush
	msg.Msg = fmt.Sprintf("%s\n\n%s", title, content)
	p := self.GetPlayer(uid)
	if p != nil {
		p.SendMsg(consts.MsgTypeHallNotifyPush, &msg)
	} else {
		if save {
			if err := AddUnreadNotify(uid, msg.Msg); err != nil {
				xlog.Logger().Error("AddUnreadNotify", err)
			}
		}
	}
}
