package static

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sync"
)

// 牌桌信息
type Table struct {
	Id                int                       `json:"id"`                  // 牌桌号
	NTId              int                       `json:"ntid"`                // 牌桌索引
	HId               int                       `json:"hid"`                 // 包厢id(6位)
	IsVitamin         bool                      `json:"is_vitamin"`          // 包厢是否开启了疲劳值
	IsHidHide         bool                      `json:"ishidhide"`           // 包厢是否开启了圈号隐藏
	IsAiSuper         bool                      `json:"isaisuper"`           // 是否开启了超级防作弊
	IsAnonymous       bool                      `json:"isanonymous"`         // 是否开启了匿名游戏功能(与超级防作弊一样，但是每个游戏自己设置，不是包厢设置的)
	DHId              int64                     `json:"dhid"`                // 包厢id(数据库自增id)
	FId               int64                     `json:"fid"`                 // 楼层id(数据库自增id)
	NFId              int                       `json:"nfid"`                // 楼层索引
	IsCost            bool                      `json:"iscost"`              // 是否已支付
	GameNum           string                    `json:"gamenum"`             // 牌桌游戏唯一标识
	Creator           int64                     `json:"creator"`             // 创建者uid(若牌桌为包厢牌桌则创建者为包厢创建者)
	CreateType        int                       `json:"createtype"`          // 牌桌创建类型
	GameId            int                       `json:"gameid"`              // 游戏服ip
	KindId            int                       `json:"kindid"`              // 游戏类型
	KindVersion       int                       `json:"kindVersion"`         // 游戏类型版本号
	SiteType          int                       `json:"sitetype"`            // 游戏场次类型
	userLock          sync.RWMutex              `json:"-"`                   // 座位锁
	Users             []*TableUser              `json:"users"`               // 牌桌上坐着的用户(用来分配座位号)
	Config            *TableConfig              `json:"config"`              // 游戏相关参数配置
	Begin             bool                      `json:"begin"`               // 游戏是否已经开始
	BXiaPaoIng        bool                      `json:"xiapaoing"`           // 是否在下跑过程中
	FewerBegin        bool                      `json:"fewer_begin"`         // 是否为少人开局
	Step              int                       `json:"step"`                // 当前第几局
	GameInfo          string                    `json:"gameinfo"`            // 游戏场景恢复数据
	LeagueID          int64                     `json:"league_id"`           // 房间加盟商信息
	NotPool           bool                      `json:"not_pool"`            // 是否被限制数量
	CreateStamp       int64                     `json:"create_stamp"`        // 桌子创建时间戳
	JoinType          consts.HouseTableJoinType `json:"join_type"`           // 包厢桌子加入类型
	CurrentMappingNum int                       `json:"current_mapping_num"` // 当前匹配中人数
	TotalMappingNum   int                       `json:"total_mapping_num"`   // 匹配总数
	IsFloorHideImg    bool                      `json:"isfloorhideimg"`      // 楼层是否隐藏头像
	IsMemUidHide      bool                      `json:"is_mem_uid_hide"`     // 茶楼是否开启隐藏玩家ID
	Status            int                       `json:"-"`                   // 桌子状态
	InviteTimes       map[int64]int             `json:"invite_times"`        // 邀请次数
	InviteLock        sync.RWMutex              `json:"-"`                   // 邀请锁
	IsForbidWX        bool                      `json:"isforbidwx"`          // 是否禁用微信分享
	Channel           int                       `json:"channel"`             // 渠道标识
}

func (self *Table) RollbackCard() bool {
	if self.IsCost == false && self.Step == 0 {
		return true
	}
	return false
}

func (self *Table) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, self)
}

func (self *Table) MarshalBinary() (data []byte, err error) {
	return json.Marshal(self)
}

// 牌桌用户
type TableUser struct {
	Uid int64 `json:"uid"` // 用户uid
	// Restart bool  `json:"restart"` // 用户是否通过再来一局加入
	JoinAt int64 `json:"join_at"` // 大厅tablein时间戳
	Ready  bool  `json:"ready"`   //准备状态
	Payer  int64 `json:"payer"`
}

// 牌桌配置
type TableConfig struct {
	GameType       int         `json:"gametype"`       // 2好友房 or 0 金币房 or 1 比赛房
	MinPlayerNum   int         `json:"minplayernum"`   // 游戏开始最少人数
	MaxPlayerNum   int         `json:"maxplayernum"`   // 游戏开始最大人数
	RoundNum       int         `json:"roundnum"`       // 游戏局数
	CardCost       int         `json:"cardcost"`       // 房卡消耗
	CostType       int         `json:"costtype"`       // 支付方式
	Restrict       bool        `json:"restrict"`       // ip限制
	View           bool        `json:"view"`           // 是否允许观看
	FewerStart     string      `json:"fewerstart"`     // 是否可以少人开局
	GVoice         string      `json:"gvoice"`         // 是否开启实时语音
	GameConfig     string      `json:"gameconfig"`     // 游戏玩法特色配置
	MatchConfig    MatchConfig `json:"matchconfig"`    // 排位赛配置
	IsLimitChannel bool        `json:"islimitchannel"` // 是否限制渠道同服
	Paybyend       bool        `json:"paybyend"`       // 20210324 苏大强 是不是最后扣茶水，目前的玩法是入座就扣，设置这个是为了数据库写分，多算了一次茶水
	Difen          int64       `json:"difen"`
}

// ! 牌桌信息
type TableInfoDetail struct {
	TableId      int                `json:"tableid"`      // 牌桌id
	HId          int                `json:"hid"`          //! 包厢id
	NFId         int                `json:"nfid"`         //! 楼层id索引 [0-max] 第几层楼
	NTId         int                `json:"ntid"`         //! 牌桌id索引 [0-max] 第几张桌子
	Creator      int64              `json:"creator"`      //! 创建者uid(若牌桌为包厢牌桌则创建者为包厢创建者)
	Person       []Son_PersonInfo   `json:"person"`       // 牌桌上的人
	LookonPerson [][]Son_PersonInfo `json:"lookonperson"` // 牌桌上的旁观者
	//DelTime int64            `json:"time"`
	Step           int                       `json:"step"`        // 当前第几局
	KindId         int                       `json:"kindid"`      // 游戏玩法
	IsVitamin      bool                      `json:"isvitamin"`   // 是否开启了疲劳值功能
	IsHidHide      bool                      `json:"ishidhide"`   // 是否开启了圈号隐藏
	IsAiSuper      bool                      `json:"isaisuper"`   // 是否开启了超级防作弊
	IsAnonymous    bool                      `json:"isanonymous"` // 是否开启了匿名游戏功能(与超级防作弊一样，但是每个游戏自己设置，不是包厢设置的)
	GameConfig     map[string]interface{}    `json:"gameconfig"`  // 游戏特色玩法配置参数
	JoinType       consts.HouseTableJoinType `json:"join_type"`
	CurrentAiNum   int                       `json:"current_ai_num"`
	AiSuperNum     int                       `json:"ai_super_num"`
	AiSuperPercent int                       `json:"ai_super_percent"`
	DistanceLimit  int                       `json:"distancelimit"` // 牌桌距离限制
	IsFloorHideImg bool                      `json:"isfloorhideimg"`
	PrivateGPS     bool                      `json:"privategps"`      // 隐藏地址位置
	IsForbidWX     bool                      `json:"isforbidwx"`      // 是否禁用微信分享
	IsMemUidHide   bool                      `json:"is_mem_uid_hide"` // 茶楼是否开启了uid隐藏
}

type Son_PersonInfo struct {
	Uid        int64  `json:"uid"`
	Card       int    `json:"card"`
	Gold       int    `json:"gold"`
	Name       string `json:"name"`
	ImgUrl     string `json:"imgurl"`
	Sex        int    `json:"sex"`
	Ip         string `json:"ip"`
	Address    string `json:"address"`
	Longitude  string `json:"longitude"`   //! 经度
	Latitude   string `json:"latitude"`    //! 纬度
	Seat       int    `json:"seat"`        //! 座位号, -1时没有坐下
	WinCount   int    `json:"win_count"`   // 胜利局数
	LostCount  int    `json:"lost_count"`  // 失败局数
	DrawCount  int    `json:"draw_count"`  // 和局局数
	FleeCount  int    `json:"flee_count"`  // 逃跑局数
	TotalCount int    `json:"total_count"` // 总局数
	Diamond    int    `json:"diamond"`     // 钻石
}

// 序列化
func (self *Table) ToBytes() []byte {
	return HF_JtoB(self)
}

// 获取redis key
func (self *Table) GetRedisKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_TABLEINFO, self.GameId, self.Id)
}

// 获取用户座位号
func (self *Table) GetSeat(id int64) (int, bool) {
	for i, u := range self.Users {
		if u != nil && u.Uid == id {
			return i, true
		}
	}
	return -1, false
}

// ! 获取空座位号
func (self *Table) GetEmptySeat() int {
	for seat, person := range self.Users {
		if person == nil {
			return seat
		}
	}
	return -1
}

// ! 是否是包厢桌子
func (self *Table) IsTeaHouse() bool {
	return self.HId > 0
}

// 在创建桌子的时候 是不是疲劳值模式
func (self *Table) IsVitaminOnCreate() bool {
	return self.HId > 0 && self.IsVitamin
}

// ! 游戏是否开始
func (self *Table) IsBegin() bool {
	return self.Begin
}

// ! 获取游戏配置
func (self *Table) GetGameConfig() map[string]interface{} {
	info := make(map[string]interface{})

	err := json.Unmarshal([]byte(self.Config.GameConfig), &info)
	if err != nil {
		xlog.Logger().Errorln(self.Config.GameConfig, err)
	} else {
		info["hid"] = self.HId
		info["nfid"] = self.NFId
		info["ntid"] = self.NTId
		info["roundnum"] = self.Config.RoundNum
		info["playernum"] = self.Config.MaxPlayerNum
		info["fewerstart"] = self.Config.FewerStart
		info["isvitamin"] = self.IsVitaminOnCreate()
		info["ishidhide"] = self.IsHidHide
		info["gvoice"] = self.Config.GVoice
	}
	return info
}

// 得到桌上其他玩家的个数
func (self *Table) GetOtherUserCount(uid int64) (num uint16) {
	for _, u := range self.Users {
		if u == nil || u.Uid == uid {
			continue
		}
		num++
	}
	return
}

// 释放掉无效的玩家
func (self *Table) InvalidUserFree(isInArray func(int64) bool) {
	for i := 0; i < len(self.Users); {
		u := self.Users[i]
		if u == nil || !isInArray(u.Uid) {
			copy(self.Users[i:], self.Users[i+1:])
			self.Users = self.Users[:len(self.Users)-1]
			continue
		}
		i++
	}
}

func (self *Table) OnFewerStart() {
	self.FewerBegin = true
	self.Config.MaxPlayerNum -= 1
	if self.Config.MinPlayerNum > self.Config.MaxPlayerNum {
		self.Config.MinPlayerNum = self.Config.MaxPlayerNum
	}
}

func (self *Table) UserReady(uid int64, ready bool) {
	for _, u := range self.Users {
		if u != nil && u.Uid == uid {
			u.Ready = ready
			break
		}
	}
}

func (self *Table) AddUser(curSeat int, uid int64, timeNow int64, payer int64) {
	self.userLock.Lock()
	defer self.userLock.Unlock()
	self.Users[curSeat] = &TableUser{
		Uid:    uid,
		JoinAt: timeNow,
		Ready:  false,
		Payer:  payer,
	}
}

func (self *Table) DelUsers(uid int64) {
	self.userLock.Lock()
	defer self.userLock.Unlock()
	for k, v := range self.Users {
		if v != nil && v.Uid == uid {
			self.Users[k] = nil
		}
	}
}

func (self *Table) GetUser(seat int) *TableUser {
	self.userLock.RLock()
	defer self.userLock.RUnlock()
	return self.Users[seat]
}

func (self *TableConfig) IsLeastNum(min int) bool {
	return self.MaxPlayerNum <= min
}

func (self *TableConfig) IsFriendMode() bool {
	return self.GameType == GAME_TYPE_FRIEND
}

func (self *TableConfig) CanFewer() bool {
	return self.FewerStart == "true"
}

func (self *Table) GetUsers() []*TableUser {
	self.userLock.RLock()
	defer self.userLock.RUnlock()
	return self.Users
}

func (self *Table) GetUserInviteTimes(uid int64) int {
	self.InviteLock.RLock()
	defer self.InviteLock.RUnlock()
	return self.InviteTimes[uid]
}

func (self *Table) AddUserInviteTimes(uid int64) {
	self.InviteLock.Lock()
	defer self.InviteLock.Unlock()
	if self.InviteTimes == nil {
		self.InviteTimes = make(map[int64]int)
	}
	self.InviteTimes[uid]++
}
