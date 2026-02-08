package components

import (
	"errors"
	"fmt"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"strconv"
	"time"

	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

//mahjioncy "mahjioncy"

type Player struct {
	Uid                    int64               `json:"uid"`                    //!
	Name                   string              `json:"name"`                   //! 名字
	ImgUrl                 string              `json:"imgurl"`                 //! 头像
	Sex                    int                 `json:"sex"`                    //! 性别
	Ontable                bool                `json:"ontable"`                //! 是否已坐下
	ETime                  int                 `json:"etime"`                  //! 退出时间
	Seat                   uint16              `json:"seat"`                   //! 玩家桌位号
	Ready                  bool                `json:"ready"`                  //! 玩家是否准备
	TableId                uint16              `json:"tableid"`                //桌号
	Ip                     string              `json:"ip"`                     //IP地址
	FaceID                 int                 `json:"faceid"`                 //玩家默认头像
	FaceUrl                string              `json:"faceurl"`                //玩家头像
	GameID                 int                 `json:"gameid"`                 //游戏ID
	GroupID                int                 `json:"groupid"`                //社团ID
	UserRight              int                 `json:"userright"`              //用户等级
	Loveliness             int                 `json:"loveliness"`             //用户魅力
	MasterRight            int                 `json:"masterright"`            //管理权限
	MemberOrder            byte                `json:"memberorder"`            //会员等级
	MasterOrder            byte                `json:"masterorder"`            //管理等级
	UserScoreInfo          static.TagUserScore `json:"userscoreinfo"`          //积分信息
	UserStatus             int                 `json:"userstatus"`             //w玩家状态
	UserStatus_ex          uint64              `json:"userstatus_ex"`          //w玩家状态扩展
	TRUSTNum_g             uint64              `json:"trustnum_g"`             //w玩家总托管次数
	UserOfflineTag         int64               `json:"userofflinetag"`         //断线标记
	LaunchDismissTag       int64               `json:"launchdismisstag"`       //20191127 苏大强 申请结散标记
	UserVitaminLowPauseTag int64               `json:"uservitaminlowpausetag"` //竞技点过低游戏暂停标记
	AllowLookon            bool                `json:"allowlookon"`            //允许旁观标志 是否允许别人旁观我
	LookonFlag             bool                `json:"lookonflag"`             //是否正在旁观别人。UserStatus的public.US_LOOKON状态不准确，因为有的游戏进入桌子还没有坐下时是这个状态，和现在的旁观不是一个概念
	LookonTableId          int                 `json:"ltableid"`               //我旁观的桌号
	UserReady              bool                `json:"userready"`              //下一局准备好的用户（点击准备好的用户）
	Addr                   string              `json:"addr"`                   //地址
	GPS_type               int                 `json:"gpstype"`                //GPS类型
	Latitude               float32             `json:"latitude"`               //经纬度
	Longitude              float32             `json:"longitude"`              //经纬度
	Acitve                 bool                `json:"acitve"`                 //激活标志
	GameCount              int                 `json:"gamecount"`              //游戏次数
	WinCount               int                 `json:"wincount"`               //赢的次数
	DissmisReqCount        int                 `json:"dissmisreqcount"`        //玩家当前大局申请解散的次数
	MemberId               int                 `json:"memberid"`               //即时语音id
	Ctx                    PlayerMeta          `json:"ctx"`                    //游戏逻辑相关数据
	TRUSTInfos             [][]int64           `json:"trustinfos "`            //20201111 记录托管起始时间，以切片记录，因为局数顺序是有的
}

// 初始化玩家数据
func (self *Player) Init(person base2.PersonBase, gameType int) {
	self.Name = person.GetInfo().Nickname
	self.Sex = person.GetInfo().Sex
	self.ImgUrl = person.GetInfo().Imgurl
	if gameType != static.GAME_TYPE_FRIEND {
		self.UserScoreInfo.Score = person.GetInfo().Gold
	}

	self.TableId = uint16(person.GetInfo().TableId)
	self.Ip = person.GetIp()
	self.Uid = person.GetInfo().Uid
	self.Acitve = true
	self.AllowLookon = false
	self.LookonTableId = person.GetInfo().WatchTable
	if self.LookonTableId > 0 {
		self.LookonFlag = true
	}
	self.MemberId = -1
	self.Ctx.Init()
	self.CleanTask()
	self.TRUSTInfos = make([][]int64, 1)
}

// 初始化玩家数据
func (self *Player) ReInit(person base2.PersonBase) {
	self.Name = person.GetInfo().Nickname
	self.Sex = person.GetInfo().Sex
	self.ImgUrl = person.GetInfo().Imgurl
	self.TableId = uint16(person.GetInfo().TableId)
	self.Ip = person.GetIp()
	self.Uid = person.GetInfo().Uid
	if person.GetInfo().WatchTable != 0 {
		self.LookonFlag = true
		self.LookonTableId = person.GetInfo().WatchTable
	}
	self.Acitve = true
	//self.MemberId = -1
	self.Ctx.Init()
}

func (self *Player) CleanTask() {
	self.GameCount = 0
	self.WinCount = 0
}

// 重置
func (self *Player) ResetDissmiss() {
	self.DissmisReqCount = 0 //玩家当前大局申请解散的次数
	//self.DissmisReqMax	= 0//最大可申请解散的次数
}

// 获取玩家游戏逻辑数据
func (self *Player) GetCtx() *PlayerMeta {
	return &self.Ctx
}

// 数据重置
func (self *Player) Reset() {
	self.Ctx.Reset()
	//self.UserReady = false //在OnNextGame()和UserReadyReset()里面reset
	self.Ctx.CleanPiao()
	self.LaunchDismissTag = 0
	self.LookonFlag = false
}

// 数据重置
func (self *Player) UserReadyReset() {
	self.UserReady = false
}

// 开始下一局
func (self *Player) OnNextGame() {
	self.Ctx.OnNextGame()
	self.UserReady = false
	//20200512 恩施用了这个东西，那么在下一局的时候修改状态，为玩的状态
	if self.UserStatus_ex&static.US_READY != 0 {
		self.UserStatus_ex ^= static.US_READY
	}
	//这个可以不加，先放着 //竟然不是按位的。。。
	self.UserStatus_ex |= static.US_PLAY
}

func (self *Player) OnBegin() {
	self.Ctx.OnBegin()
}

func (self *Player) OnEnd() {
	self.Ctx.OnEnd()
}

func (self *Player) GetUserID() int64 {
	return self.Uid
}

func (self *Player) GetUserStatus() int {
	return self.UserStatus
}
func (self *Player) GetUserStatus_ex() uint64 {
	return self.UserStatus_ex
}
func (self *Player) GetTableID() uint16 {
	return self.TableId
}

func (self *Player) GetChairID() uint16 {
	return self.Seat
}

func (self *Player) UpdateTaskInfo(ScoreInfo *rule2.TagScoreInfo, _time int64) bool {
	//修改属性
	switch ScoreInfo.ScoreKind {
	case static.ScoreKind_Win:
		//self.UserScoreInfo.WinCount++
		self.WinCount++
		self.GameCount++
	case static.ScoreKind_Lost:
		//self.UserScoreInfo.LostCount++
		self.GameCount++
	case static.ScoreKind_Draw:
		//self.UserScoreInfo.DrawCount++
		self.GameCount++
	case static.ScoreKind_Flee:
		//self.UserScoreInfo.FleeCount++
		self.GameCount++
	}
	return true
}

func (self *Player) WriteScore(ScoreInfo *rule2.TagScoreInfo, _time int64) bool {
	//效验参数
	if self.Acitve == false {
		return false
	}

	//修改属性
	switch ScoreInfo.ScoreKind {
	case static.ScoreKind_Win:
		self.UserScoreInfo.WinCount++
	case static.ScoreKind_Lost:
		self.UserScoreInfo.LostCount++
	case static.ScoreKind_Draw:
		self.UserScoreInfo.DrawCount++
	case static.ScoreKind_Flee:
		self.UserScoreInfo.FleeCount++
	}
	// TODO 20190809 hex 这里不再修改玩家积分 移动到GameCommon.go统一的OnSettle()接口管理
	// self.UserScoreInfo.Score += ScoreInfo.Score
	// 更新战绩
	self.updateUserBattleRecord(self.Uid, ScoreInfo.ScoreKind)

	return true
}

// 更新用户战绩数据
func (self *Player) updateUserBattleRecord(uid int64, recordType int) error {
	p, err := server2.GetDBMgr().GetDBrControl().GetPerson(uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	//dateStr := time.Now().Format(time.DateOnly)
	//rankWinKey := fmt.Sprintf("rank:winround:%s", dateStr)
	//rankTotalKey := fmt.Sprintf("rank:totalround:%s", dateStr)
	//pipe := server2.GetDBMgr().GetDBrControl().RedisV2.Pipeline()
	//defer pipe.Close()
	//pipe.ZIncrBy(rankTotalKey, 1, static.HF_I64toa(uid))
	switch recordType {
	case static.ScoreKind_Win: // 胜利
		p.TotalCount++
		p.WinCount++
		//pipe.ZIncrBy(rankWinKey, 1, static.HF_I64toa(uid))

		server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "WinCount", p.WinCount, "TotalCount", p.TotalCount)

		// 更新内存
		person := server2.GetPersonMgr().GetPerson(uid)
		if person != nil {
			person.Info.TotalCount = p.TotalCount
			person.Info.WinCount = p.WinCount
		}
	case static.ScoreKind_Lost: // 失败
		p.TotalCount++
		p.LostCount++
		server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "LostCount", p.LostCount, "TotalCount", p.TotalCount)

		// 更新内存
		person := server2.GetPersonMgr().GetPerson(uid)
		if person != nil {
			person.Info.TotalCount = p.TotalCount
			person.Info.LostCount = p.LostCount
		}
	case static.ScoreKind_Draw: // 平局
		p.TotalCount++
		p.DrawCount++
		server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "DrawCount", p.DrawCount, "TotalCount", p.TotalCount)

		// 更新内存
		person := server2.GetPersonMgr().GetPerson(uid)
		if person != nil {
			person.Info.TotalCount = p.TotalCount
			person.Info.DrawCount = p.DrawCount
		}
	case static.ScoreKind_Flee: // 逃跑
		p.TotalCount++
		p.FleeCount++
		server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "FleeCount", p.FleeCount, "TotalCount", p.TotalCount)

		// 更新内存
		person := server2.GetPersonMgr().GetPerson(uid)
		if person != nil {
			person.Info.TotalCount = p.TotalCount
			person.Info.FleeCount = p.FleeCount
		}
	default:
		err = errors.New(fmt.Sprintf("unknown record type: %d", recordType))
		xlog.Logger().Errorln(err.Error())
		return err
	}

	//pipe.Expire(rankTotalKey, time.Hour*48)
	//pipe.Expire(rankWinKey, time.Hour*48)
	//
	//_, err = pipe.Exec()
	//if err != nil {
	//	xlog.Logger().Error(err)
	//} else {
	//	xlog.Logger().Infof("%d 更新排行榜成功", uid)
	//}
	return nil
}

func (self *Player) SetUserStatus(cbUserStatus byte, wTableID uint16, wChairID uint16) bool {
	xlog.Logger().Debug("setuserstatus ; tableid :" + strconv.Itoa(int(wTableID)) + " chair :" + strconv.Itoa(int(wChairID)))
	// 效验状态
	if self.Acitve == false {
		return false
	}
	// 设置变量
	self.TableId = wTableID
	self.Seat = wChairID
	self.UserStatus = int(cbUserStatus)

	// 如果当前玩家的状态为坐下，而玩家之前又准备，此时置状态为准备状态
	if self.UserStatus == static.US_SIT && self.IsReady() {
		self.UserStatus = static.US_READY
	}
	return true
}

func (self *Player) IsActive() bool {
	return self.Acitve
}

func (self *Player) IsReady() bool {
	return self.Ready
}

func (self *Player) GetClientIP() string {
	return self.Ip
}

type PlayerMgr struct {
	player         map[int64]*Player //玩家数据
	m_LookonPlayer map[int64]*Player //旁观用户
}

// 20191219 苏大强 托管计数
func (self *Player) TRUSTRecord() {
	self.TRUSTNum_g++
	self.Ctx.TrustRecord()
}

// 20201111 苏大强 托管信息,武汉需要记录所有托管进入时间，这样随时都能H
func (self *Player) TRUSTRecordInfo() {
	currentcount := len(self.TRUSTInfos)
	if self.GameCount+1 > currentcount {
		//要创建了 current一定小于self.GameCount
		for i := currentcount; i < self.GameCount+1; i++ {
			currenInfo := []int64{}
			self.TRUSTInfos = append(self.TRUSTInfos, currenInfo)
		}
	}
	self.TRUSTInfos[int64(self.GameCount)] = append(self.TRUSTInfos[uint64(self.GameCount)], time.Now().Unix())
}

// 20201111 苏大强 獲取最後托管時間,可能是0
func (self *Player) GetlastTrustTime() int64 {
	count := len(self.TRUSTInfos)
	if count == 0 {
		return 0
	}
	count -= 1
	num := len(self.TRUSTInfos[count])
	return self.TRUSTInfos[count][num-1]
}

func (self *Player) ChangeTRUST(set bool) {
	if set {
		self.UserStatus_ex |= static.US_TRUST
		self.TRUSTRecord()
		self.TRUSTRecordInfo()
	} else {
		//不是设置就是关闭托管
		if self.UserStatus_ex&static.US_TRUST != 0 {
			self.UserStatus_ex &= (^uint64(static.US_TRUST))
		}
	}
}
func (self *Player) CheckTRUST() bool {

	return self.UserStatus_ex&static.US_TRUST != 0
}
