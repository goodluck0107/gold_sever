package components

/**

游戏公共部分

注意：所有子游戏公用部分，修改时必须保证兼容

*/

import (
	"encoding/json"
	"errors"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/sirupsen/logrus"
)

type GameCommon = Common

type Common struct {
	BaseFunc
	// 20181124 苏大强 添加一个信号，告诉上层拒绝解散了
	RestTime  chan byte //消息管道
	GameTimer Time      //时间计时器
	GameTask  Task      //金豆任务

	//szGameRoomRule    string            //房间规则
	Debug int //debug开关
	//g_iLog            int               //日志开关

	//属性变量
	RenWuAble bool //任务是否开启

	//Config       *static.ConfigGame //! csv
	FriendInfo info2.TagTableFriend //好友房信息
	//游戏变量
	SiceCount     uint16 //骰子点数
	IsTableLocked bool   //锁定标志
	//m_GameSate int //当前游戏状态

	//设置变量
	HaveDeleteFangKa bool  //是否已经扣过卡了
	HaveGetHongbao   bool  //是否已经获得红包
	CurCompleteCount byte  //当前游戏局数，局数基本标识，不要做其他用途
	TimeStart        int64 //游戏开始时间
	TimeEnd          int64 //游戏大局结束时间
	RoundTimeStart   int64 //一局游戏开始时间

	//记录分数
	GameTaxScore int

	HuType static.TagHuType //胡牌配置

	GameBeginTime time.Time //游戏开始时间
	RoundReplayId int64     //单局回放Id

	//游戏算番算分
	FanScore        [4]meta2.Game_mj_fan_score
	LimitTime       int64 //20181212 限定时间，和定时器计时不放一起算了
	userMutex       sync.Mutex
	lookonuserMutex sync.Mutex

	InvalidGangCards    []byte //无效的杠牌
	BeFirstGang         bool
	DaoChePai           [4][2]byte             // 倒车牌,吃牌后不能明倒车,不能打出
	DismissRoomTime     int                    // 自动解散房间时间
	OfflineRoomTime     int                    // 离线解散房间时间
	LaunchDismissTime   int                    //20191127 苏大强 自动发起解散房间时间
	VitaminLowPauseTime int                    // 竞技过低超时解散房间时间
	ReBanker            [4]int                 // 20190708 记录连庄信息
	PauseStatus         int                    // 是否暂停中
	PauseUsers          []uint16               // 被暂停的玩家座位号
	SendCardOpt         static.TagSendCardInfo // 当前发牌信息

	IsSetDissmiss          bool      // 标识当前是否设置过了大局最大解散次数
	DissmisReqMax          int       // 大局每个玩家可申请解散的最大次数
	DismissTrustee         []uint16  // 记录解散房间时托管玩家的座位号
	LianGangCount          int       //连杠
	GuoHuCount             [4]int    //玩家过胡次数
	KanCount               [4]int    //玩家数坎个数
	HasSendNo13Tip         bool      //不够13张不能亮牌
	FirstXuanPiaoSure      bool      //首局定漂是否选完了
	FirstPiaoNum           [4]int    //首局定漂是几
	GameUserTing           [4]bool   //玩家是否听牌,跟赔庄相关
	GameUserTingType       [4]uint64 //玩家听牌的最大牌型
	QiangGangScoreSend     bool
	QiangGangOperateScore  meta2.Msg_S_OperateScore_K5X
	K5xReplayRecord        meta2.K5x_Replay_Record //回放记录
	GameShowCard           [4]meta2.ShowCard
	GameHuCards            []byte //当前这局已经胡过的牌
	QiangGangOperateResult static.Msg_S_OperateResult
	GameNextBanker         uint16            //下一局庄家
	Chatlog                []static.ChatLogs //聊天数据
	HasFanLaiZi            bool              //当前这局有没有翻癞子
	StartMengQuan          bool              //开始计算焖癞子圈数
	NormalDispatchRound    int               //第几轮摸牌(每个人都正常摸牌一次,明杠抓牌不算，暗杠抓牌算抓过一次)
	NormalDispatchStatus   [4]bool           //玩家是否正常摸牌(明杠抓牌不算，暗杠抓牌算抓过一次)
	AlarmStatus            [4]int            //玩家报警标识(0未报警)
	BAlreadyAddZongZha     bool              //是否已经加了总炸积分
	BSendDissmissReq       bool              //每小局是否已经发送了申请解散
	FirstLiang             uint16            //第一个亮倒的玩家
	IsLastRoundHZ          bool              //上一局是否荒庄
	BShangLou              bool
	TrustPunish            bool                     //20201112 苏大强 第一局打完就可以了托管扣分了，不管是不是流局
	SendGameEnd            bool                     //20201112 苏大强 发送过小结算信息
	CardRecorder           [10][]meta2.CardRecorder //记牌器数据 ,假设最多10个玩家
	BCardRcdNextAble       [10]bool                 //记牌器数据 ,true表示新购买的记牌器下局生效,false表示记牌器本局已经生效
	CardRecordFlag         bool                     //记牌器数据 true表示可以购买记牌器也可以使用记牌器
	GameUserHuaZhu         [4]bool                  //玩家是否花猪   20210129 苏大强 宜昌血流用
	SuperMap               map[int64]int
}

func (self *Common) Init() {
	self.ReBanker = [4]int{0, 0, 0, 0}
	self.FriendInfo.CreateUserID = self.GetTableInfo().Creator
	if self.Plock == nil {
		self.Plock = new(sync.RWMutex)
	}
	//立即把PlayerInfo 保护起来，防止桌子初始化一半时定时检查任务启动了
	self.Plock.Lock()
	self.PlayerInfo = make(map[int64]*Player)
	self.Plock.Unlock()
	if self.Lplock == nil {
		self.Lplock = new(sync.RWMutex)
	}
	//立即把LookonPlayer 保护起来，防止桌子初始化一半时定时检查任务启动了
	self.Lplock.Lock()
	self.LookonPlayer = make(map[int64]*Player)
	self.Lplock.Unlock()

	// 如果是少人开局就减少一个人
	self.InitPlayerCount()
	//初始化扔10个先
	//self.GameTimer.TimeArray = make(map[int64]*public.TagTimerItem, 10)
	//self.GameTimer.mu = new(Lock.RWMutex)
	self.Debug = consts.Debug

	self.DismissRoomTime = static.DISMISS_ROOM_TIMER
	self.OfflineRoomTime = static.OFFLINE_ROOM_TIMER
	self.VitaminLowPauseTime = static.VITAMIN_LOW_DISMISS_TIMER
	self.IsSetDissmiss = false
	self.Rule.Radix = 1
	self.CardRecorder = [10][]meta2.CardRecorder{}
	self.BCardRcdNextAble = [10]bool{false} //默认本局及时生效
	self.CardRecordFlag = false
	// self.checkVitamin()
}

func (self *Common) InitPlayerCount() {
	self.GetConfig().PlayerCount = uint16(self.GetTableInfo().Config.MinPlayerNum)
	self.GetConfig().ChairCount = uint16(self.GetTableInfo().Config.MaxPlayerNum)
	self.GetConfig().LookonCount = 100 //公共的定义为100个，每个游戏可以重新赋值
}

// 设置游戏状态
func (self *Common) onBegin() {

	xlog.Logger().Debug("onbegin")
	//效验状态
	if self.m_bGameStarted == true {
		return
	}

	//检查程序是否出错，程序出错后无法恢复房间，需要强制解散 对于入桌自动准备的游戏，需要在onbegin消息里面检查一下状态是否异常
	if self.CheckErrorDismissAuto() {
		return
	}

	//设置变量
	self.HaveDeleteFangKa = false
	self.HaveGetHongbao = false
	self.CurCompleteCount = 0
	self.m_bGameStarted = true
	self.TimeStart = time.Now().Unix() //(DWORD)time(NULL);
	self.RoundTimeStart = self.TimeStart

	//记录分数
	self.GameTaxScore = 0

	//	//设置玩家
	//	tagServerUserData * pUserData=NULL;
	//	for (WORD i=0;i<m_wChairCount;i++) {
	//		if (m_pIUserItem[i]!=NULL)
	//		{
	//			pUserData=m_pIUserItem[i]->GetUserData();
	//			pUserData->cbUserStatus=US_PLAY;
	//			m_dwPlayerID[i]=pUserData->dwUserID;
	//		}
	//		self.m_UserOfflineTag[i] = -1;	//清除时间戳
	//	}

	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if self.GetConfig().StartIgnoreOffline && v.UserStatus == static.US_OFFLINE {
			} else {
				v.UserStatus = static.US_PLAY
				v.UserOfflineTag = -1
				v.UserVitaminLowPauseTag = -1
				//用户状态
				self.SendUserStatus(int(v.GetChairID()), static.US_PLAY)
			}
			self.sendUserStartGame(v)
			v.Ctx.LimitedTime = 0 //累计时间清理
			v.Ctx.RecordbeginTime = 0
		}
	})

	//发送状态
	self.SendTableStatus(self.GetTableInfo().Id)

	//通知事件
	//m_pITableFrameSink->OnEventGameStart();
	self.getServiceFrame().OnBegin()

	self.ReflushUserItem()

	if self.FriendInfo.CreateType != 0 {
		var _info rule2.Tag_OtherFriendCreate
		_info.UserID = self.FriendInfo.CreateUserID //req->dwUserID;
		_info.RoomNum = self.FriendInfo.RoomID      //req->dwRoomID;
		_info.Type = 1
		_info.Count = self.FriendInfo.JuShu
		_info.CreateType = self.FriendInfo.CreateType

		for i := 0; i < self.GetChairCount(); i++ {
			item := self.GetUserItemByChair(uint16(i))
			if item == nil {
				continue
			}
			if item.GetChairID() != static.INVALID_CHAIR {
				_info.UserPlayerID[i] = item.Uid
				_info.UserName[i] = item.Name
			}
		}
		_info.GameNum = self.FriendInfo.GameNum
		_info.Rule = self.FriendInfo.Rule
		self.AddCreateOtherFriend(&_info)
	}

	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("游戏规则：%s,restrict:%t", self.GetTableInfo().Config.GameConfig, self.GetTableInfo().Config.Restrict))
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if person := server2.GetPersonMgr().GetPerson(v.Uid); person != nil {
				self.OnWriteGameRecord(v.GetChairID(), fmt.Sprintf("地址详情,Ip:%s -> Longitude:%s -> Latitude:%s  -> Address:%s", person.Info.Ip, person.Info.Longitude, person.Info.Latitude, person.Info.Address))
			}
		}
	})

	//启动定时器，每分钟一次
	//	bool res = m_pIGameServiceFrame->SetTableTimer(TableID, IDI_OFFLINE+m_wChairCount, 48000L, 99999, 0);
	//	if (!res)
	//	{
	//		CTraceService::TraceString(TEXT("设置计时器失败5"), TraceLevel_Warning);
	//	}

	//	//日志
	//	if (g_iLogDetails >= 1)
	//	{
	//		char buf[128];
	//		memset(buf, 0, sizeof(buf));
	//		sprintf(buf, "【StartGame】：开启定时器 wTableID(%d) dwTimerID(%d) dwTableTimeID(%d) dwElapse(%d) wBindParam(%d)", TableID, IDI_OFFLINE+m_wChairCount, 60000L, 99999, 0);
	//		TraceMessage(m_pGameServiceOption->wServerID, buf);
	//	}

	self.AAPay()
}

// 他人开房
func (self *Common) AddCreateOtherFriend(_info *rule2.Tag_OtherFriendCreate) {

	switch _info.CreateType {
	case 1: //提他人开房
	case 2: //VIP开房
	case 4: //机器人开房
	}

}

func WithBranch(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// 是否位于两个玩家之间 可以跨越
func (self *Common) IsBetween(left, center, right uint16) bool {
	fmt.Println("IsBetween：", left, center, right)
	ci, ri := static.INVALID_CHAIR, static.INVALID_CHAIR
	var i uint16
	for {
		i++
		nextUserOfLeft := self.GetNextSeat(left)
		if nextUserOfLeft == int(center) {
			ci = i
		}
		if nextUserOfLeft == int(right) {
			ri = i
		}
		left = uint16(nextUserOfLeft)
		if i >= uint16(self.GetPlayerCount()) {
			break
		}
	}
	if ci == static.INVALID_CHAIR || ri == static.INVALID_CHAIR {
		fmt.Println("找位置出错")
		return false
	}
	if ci < ri {
		fmt.Println("在中间", ci, ri)
		return true
	}
	fmt.Println("不在中间", ci, ri)
	return false
}

// ! 发送消息
func (self *Common) OnBaseMsg(msg *base2.TableMsg) {

	xlog.Logger().Debug("gameserver onmsg:")
	xlog.Logger().Debug(msg)

	switch msg.Head {
	case consts.MsgTypeGameinfo: //! 请求信息
		var _msg static.Msg_C_Info
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if _msg.AllowLookon == 0 {
				self.onUserInfo(msg.Uid, &_msg)
			} else {
				self.onLookonUserInfo(msg.Uid, &_msg)
			}
		}

	case consts.MsgTypeGameDismissFriendReq: //申请解散房间
		var _msg static.Msg_C_DismissFriendReq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if self.RestTime != nil {
				self.RestTime <- 0
			}
			if !self.checkDissmissTag() && self.IsGameStarted() {
				_msg := "游戏中不能申请解散，请联系盟主处理!"
				wChairSeat := self.GetChairByUid(msg.Uid)
				self.SendGameNotificationMessage(wChairSeat, _msg)
				return
			}
			self.onDismissFriendMsg(msg.Uid, &_msg)
			// 通知玩法规则中谁发起了解散
			self.getServiceFrame().OnMsg(msg)
		}
		//20200217 苏大强 托管玩家自动同意解散房间
		self.TrustUserAutoOperDismiss(true)
	case consts.MsgTypeGameDismissFriendResult: //申请解散玩家选择
		var _msg static.Msg_C_DismissFriendResult
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if !self.getServiceFrame().OnMsg(msg) {
				self.FlashClient(msg.Uid, "")
			}
			self.onDismissResult(msg.Uid, &_msg)
		}
	case consts.MsgTypeGameReady: //！准备
		var _msg static.Msg_C_Ready
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.onUserReady(msg.Uid, &_msg)
		}
	case consts.MsgTypeGameCancelReady: //！取消准备
		var _msg static.Msg_C_CancelReady
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.onUserCancelReady(msg.Uid, &_msg)
		}
	case consts.MsgTypeGameUserChat: //聊天信息
		var _msg static.Msg_C_UserChat
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.onUserChat(msg.Uid, &_msg)
			self.recordChat(&_msg, nil)
		}
	case consts.MsgTypeGameUserYYInfo: //语音聊天
		var _msg static.Msg_C_UserYYInfo
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.onUserYVChat(msg.Uid, &_msg)
			self.recordChat(nil, &_msg)
		}
	case consts.MsgTypeGameGPSReq: //GPS数据
		var _msg static.Msg_S_UserGPSReq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.onGPSMsg(msg.Uid, &_msg)
		}
	case consts.MsgTypeGameChatLogs: //请求聊天记录
		var _msg static.Msg_C_ChatLog
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.SendPersonMsg(consts.MsgTypeGameChatLogs, self.Chatlog, _msg.Seat)
		}
	case consts.MsgTypeGameForceExit: //强退
		var _msg static.Msg_S_ForceExit
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if _item := self.GetUserItemByUid(msg.Uid); _item != nil {
				self.OnStandup(_item.Uid, _item.Uid)
				//self.getServiceFrame().OnGameOver(_item.Seat, public.GER_USER_LEFT)
				//self.GetTable().TableExit(msg.Uid)
				//self.GetTable().NotifyTableChange()
			}
		}
	case consts.MsgTypeGameFewerStartReq:
		self.OnFewerApply(msg.Uid)
	case consts.MsgTypeGameFewerStartResult:
		var _msg static.Msg_C_DismissFriendResult
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.OnFewerResult(msg.Uid, &_msg)
		}
	case consts.MsgTypeHouseTableInviteSend:
		var _msg static.Msg_GetUserInfo
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.SendUserMsg(consts.MsgTypeHouseTableInviteSend, nil, msg.Uid)
			self.HouseTableInvite(msg.Uid, _msg.Uid)
		} else {
			xlog.Logger().Errorln("在线邀请：消息解析错误：", err)
		}
	// case constant.MsgCommonToGameContinue:
	//
	// 	self.OnWriteGameRecord(public.INVALID_CHAIR, "子游戏未实现恢复发牌接口。")

	case consts.MsgTypeHouseVitaminSet_Ntf:
		self.OnHouseMemVitaminUpdate(msg.V.(*static.Msg_UserTableId).Uid)
	case consts.MsgTypeGameTimeMsg:
		self.OnTimerEvent(msg.V.(*static.Msg_TimeEventMsg).Id, msg.V.(*static.Msg_TimeEventMsg).Uid)
	case consts.MsgTypeGameGVoiceMember: //即时语音
		var _msg static.Msg_C_GVoiceMember
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.SetGVoiceMember(msg.Head, &_msg)
		}

	case consts.MsgTypeToolList: //用户可以兑换的记牌器列表
		var _msgRet static.Msg_S_Tool_ToolList
		//查询当前玩家的记牌器失效时间
		retb, retTime := self.GetUserToolInfo(msg.Uid, 0)
		_msgRet.DeadAt = retTime
		if !retb {
			_msgRet.DeadAt = 0 //过期显示成未开通
		}

		self.SendPersonMsg(consts.MsgTypeToolList, _msgRet, self.GetChairByUid(msg.Uid))
	case consts.MsgTypeToolExchange: //用户请求兑换记牌器
		var _msg static.Msg_C_Tool_ToolExchange
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			var _msgRet static.Msg_S_Tool_ToolExchange
			p, errUpr := server2.GetDBMgr().GetDBrControl().GetPerson(msg.Uid)
			if errUpr != nil {
				xlog.Logger().Errorln("Set user data to redis error: ", errUpr.Error())
			}
			_msgRet.Num = _msg.Num
			_msgRet.Price = _msg.Price
			if p != nil {
				_msgRet.DeadAt = p.CardRecorderDeadAt
			}
			self.SendPersonMsg(consts.MsgTypeToolExchange, _msgRet, self.GetChairByUid(msg.Uid))
			//购买成功立即触发记牌器
			self.SendCardRecorder(self.GetChairByUid(msg.Uid), 1)
		}
	case consts.MsgTypeToolExchangeHall: //用户请求兑换记牌器,大厅转发
		var _msg static.Msg_S_Tool_ToolExchange
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.SendPersonMsg(consts.MsgTypeToolExchange, _msg, self.GetChairByUid(msg.Uid))
			//购买成功立即触发记牌器
			self.SendCardRecorder(self.GetChairByUid(msg.Uid), 1)
		}
	case consts.MsgTypeGameLeftCards:
		pg := server2.GetPersonMgr().GetPerson(msg.Uid)
		if pg != nil {
			pg.SendMsg(msg.Head, xerrors.SuccessCode, self.GetLeftCards(pg.Info.Uid))
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("看牌库，玩家%d不存在游戏服。", msg.Uid))
		}
	case consts.MsgTypeGameWantGood:
		pg := server2.GetPersonMgr().GetPerson(msg.Uid)
		if pg != nil {
			userItem := self.GetUserItemByUid(msg.Uid)
			if userItem != nil && server2.GetDBMgr().CheckIsHigher(userItem.Uid) {
				userItem.Ctx.WantGood = true
				pg.SendMsg(msg.Head, xerrors.SuccessCode, &static.Msg_S2C_WantGood{Ok: true})
				self.OnWriteGameRecord(userItem.GetChairID(), "想要好牌")
			} else {
				pg.SendMsg(msg.Head, xerrors.SuccessCode, &static.Msg_S2C_WantGood{Ok: false})
			}
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("要好牌，玩家%d不存在游戏服。", msg.Uid))
		}
	case consts.MsgTypeGameWantCard:
		var _msg static.Msg_Card
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.SetPlayerWantCard(msg.Uid, msg.Head, _msg.Card)
		} else {
			xlog.Logger().Error(err)
		}
	case consts.MsgTypeGamePeepCard:
		var _msg static.Msg_Userid
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			userItem := self.GetUserItemByUid(_msg.Uid)
			if userItem != nil {
				pg := server2.GetPersonMgr().GetPerson(msg.Uid)
				if pg != nil {
					pg.SendMsg(msg.Head, xerrors.SuccessCode, self.PeepPlayerCard(pg.Info.Uid, userItem.GetChairID()))
				} else {
					self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("看牌，玩家%d不存在游戏服。", msg.Uid))
				}
			} else {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("看牌，对方玩家%d 不存在", _msg.Uid))
			}
		} else {
			xlog.Logger().Error(err)
		}
	default:
		//syslog.Logger().Errorln("未受理的协议请求", msg.Head)
		if !self.getServiceFrame().OnMsg(msg) {
			self.FlashClient(msg.Uid, "")
		}
	}
}

// ! 包厢邀请
func (self *Common) HouseTableInvite(inviter int64, invitee int64) {
	var ntfMsg string
	if self.m_bGameStarted || self.GetTable().IsBegin() {
		ntfMsg = consts.MsgContentHGameInviteOnBegin
	} else if self.GetSeatedNum(false) >= self.GetChairCount() {
		ntfMsg = consts.MsgContentHGameInviteOnFull
	} else {
		if userItem := self.GetUserItemByUid(inviter); userItem != nil {
			if invitee == 0 && self.GetTableInfo().GetUserInviteTimes(userItem.Uid) >= consts.MaxHouseInviteTimes {
				ntfMsg = consts.MsgContentHGameInviteOverFlow
				self.GetTableInfo().GetEmptySeat()
			} else {
				if err := self.GetTable().OnInvite(inviter, invitee); err == nil {
					ntfMsg = consts.MsgContentHGameInviteSucceed
					if invitee == 0 {
						self.GetTableInfo().AddUserInviteTimes(userItem.Uid)
					}
				} else {
					ntfMsg = consts.MsgContentHGameInviteFailed
				}
			}
		} else {
			ntfMsg = consts.MsgContentHGameInviteUserNil
		}
	}
	self.SendGameNotificationMessage(self.GetChairByUid(inviter), ntfMsg)
}

// ! 设置即时语音memberid
func (self *Common) SetGVoiceMember(head string, opt *static.Msg_C_GVoiceMember) {
	_user := self.GetUserItemByUid(opt.Uid)
	if _user != nil {
		_user.MemberId = opt.Id
	}
	//var _msg public.Msg_C_GVoiceMember
	self.SendGVoiceMember()
}

func (self *Common) SendGVoiceMember() {
	var _msg static.Msg_S2C_GVoiceMember
	_msg.User = make([]*static.Item_GVoiceMember, 0)
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			data := new(static.Item_GVoiceMember)
			data.Uid = v.Uid
			data.Id = v.MemberId
			_msg.User = append(_msg.User, data)
		}
	})
	self.SendTableMsg(consts.MsgTypeGameGVoiceMember, &_msg)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameGVoiceMember, _msg)
}

func (self *Common) SetSendCardOpt(opt static.TagSendCardInfo) {
	user := self.GetUserItemByChair(opt.CurrentUser)
	if user == nil {
		self.OnWriteGameRecord(opt.CurrentUser, "SetSendCardOpt无效的用户")
		return
	}
	opt.Uid = user.Uid
	opt.Status = true
	self.SendCardOpt = opt
}

func (self *Common) onGPSMsg(uid int64, msg *static.Msg_S_UserGPSReq) {
	item := self.GetUserItemByUid(uid)
	if item != nil {
		item.Latitude = msg.Latitude
		item.Longitude = msg.Longitude
		item.Addr = msg.Addr
		item.GPS_type = msg.Type

		self.sendUserStartGame(item)
	}
}

func (self *Common) checkDissmissTag() bool {
	//20200806 苏大强 嘉鱼是个新需求，其他游戏还是走以前的
	if self.Rule.DissMissMask == 0 {
		if self.Rule.DissMissTeaTag != 0 && self.GetTableInfo().IsTeaHouse() {
			return false
		}
	} else {
		//嘉鱼的要检查mask
		return self.Rule.CheckCanDissMiss(self.GetGameRoundStatus())
	}

	return true
}

// 语音
func (self *Common) onUserYVChat(uid int64, msg *static.Msg_C_UserYYInfo) {
	self.SendTableMsg(consts.MsgTypeGameUserYYInfo, msg)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameUserYYInfo, msg)
}

// 聊天
func (self *Common) onUserChat(uid int64, msg *static.Msg_C_UserChat) {
	_userItem := self.GetUserItemByUid(uid)
	if _userItem == nil {
		return
	}

	var _minGold int
	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND {
		_minGold = server2.GetSiteMgr().GetSiteByType(self.GetTableInfo().KindId, self.GetTableInfo().SiteType).MinGold
	} else {
		houseApi := self.GetHouseApi()
		if houseApi == nil {
			return
		}
		fo, _ := houseApi.GetFloorVitaminOption()
		if fo == nil {
			return
		}
		if !fo.IsVitamin {
			return
		}
		_minGold = int(fo.VitaminLowLimit)
	}

	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND {

		if msg.Color == meta2.CHAT_COLOR_MOFA { //魔法表情
			if _userItem.UserScoreInfo.Score < meta2.CHAT_COLOR_MOFA_COST+_minGold {
				//self.SendGameNotificationMessage(_userItem.Seat, "金币不足")
				self.GameTable.SendMsg(_userItem.Uid, consts.MsgTypeGameUserChat, xerrors.ChatErrorCode, xerrors.ChatEerrorCodeError.Msg)
				return
			}

			self.writeScore(_userItem, -meta2.CHAT_COLOR_MOFA_COST, static.ScoreKind_pass, models.CostTypeSticker)
		} else if msg.Color == meta2.CHAT_COLOR_MOFA_10 { //魔法表情10连
			if _userItem.UserScoreInfo.Score < meta2.CHAT_COLOR_MOFA_10_COST+_minGold {
				self.GameTable.SendMsg(_userItem.Uid, consts.MsgTypeGameUserChat, xerrors.ChatErrorCode, xerrors.ChatEerrorCodeError.Msg)
				return
			}
			self.writeScore(_userItem, -meta2.CHAT_COLOR_MOFA_10_COST, static.ScoreKind_pass, models.CostTypeSticker)
		}
	} else {
		if msg.Color == meta2.CHAT_COLOR_MOFA { //魔法表情
			if _userItem.UserScoreInfo.Vitamin < static.SwitchVitaminToF64(int64(meta2.CHAT_COLOR_MOFA_COST+_minGold)) {
				//self.SendGameNotificationMessage(_userItem.Seat, "金币不足")
				self.GameTable.SendMsg(_userItem.Uid, consts.MsgTypeGameUserChat, xerrors.ChatErrorCode, xerrors.ChatEerrorCodeError.Msg)
				return
			}

			self.writeScore(_userItem, -meta2.CHAT_COLOR_MOFA_COST, static.ScoreKind_pass, models.CostTypeSticker)
		} else if msg.Color == meta2.CHAT_COLOR_MOFA_10 { //魔法表情10连
			if _userItem.UserScoreInfo.Vitamin < static.SwitchVitaminToF64(int64(meta2.CHAT_COLOR_MOFA_10_COST+_minGold)) {
				self.GameTable.SendMsg(_userItem.Uid, consts.MsgTypeGameUserChat, xerrors.ChatErrorCode, xerrors.ChatEerrorCodeError.Msg)
				return
			}
			self.writeScore(_userItem, -meta2.CHAT_COLOR_MOFA_10_COST, static.ScoreKind_pass, models.CostTypeSticker)
		}
	}

	self.SendTableMsg(consts.MsgTypeGameUserChat, msg)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameUserChat, msg)
}

func (self *Common) onUserReady(uid int64, msg *static.Msg_C_Ready) bool {
	//变量定义
	pUserData := self.GetUserItemByUid(uid)
	if pUserData == nil {
		xlog.Logger().Debug("ready fail ,no uid!")
		return false
	}
	bLookonUser := (pUserData.GetUserStatus() == static.US_LOOKON)
	pUserData.Ready = true

	//状态效验
	if bLookonUser == true {
		xlog.Logger().Debug("ready fail ,bLookonUser!")
		return false
	}
	if pUserData.GetUserStatus() >= static.US_PLAY {
		xlog.Logger().Debug("ready fail ,US_PLAY!")
		return true
	}
	//设置变量
	pUserData.UserStatus = static.US_READY
	//20200702 苏大强 这个情况发生在开局的时候前一个玩家准备了离线，下一个玩家看不到他准备了
	pUserData.UserStatus_ex |= static.US_READY
	pUserData.UserReady = true
	//同意处理
	self.OnActionUserReady(pUserData)

	//发送同意消息
	t := self.GetTableInfo()
	if t != nil {
		t.UserReady(uid, true)
		server2.GetDBMgr().GetDBrControl().Set(t.GetRedisKey(), t.ToBytes())
		server2.PushTableStatusMsg(t.HId, consts.MsgTypeGameReady, &static.UserReadyState{Ready: true, Uid: uid, Hid: t.DHId, Fid: t.FId, Tid: t.Id})
	}

	//变量定义
	//bool bMatchServer=((m_pGameServiceOption->wServerType&public..GAME_GENRE_MATCH)!=0);
	bControlStart := false //( m_pGameServiceOption- > cbControlStart == TRUE);

	//发送状态
	self.SendPlayStatus(pUserData)
	self.getServiceFrame().OnReady(uid, pUserData.Ready)

	//开始判断
	if (bControlStart == false) && (self.StartVerdict() == true) {
		//syslog.Logger().Debug("[游戏开始]")
		xlog.Logger().Debug("read succes ,begin!")
		self.onBegin()
		return true
	}

	return true
}

func (self *Common) onUserCancelReady(uid int64, msg *static.Msg_C_CancelReady) bool {
	//变量定义
	pUserData := self.GetUserItemByUid(uid)
	if pUserData == nil {
		xlog.Logger().Debug("read fail ,no uid!")
		return false
	}
	bLookonUser := (pUserData.GetUserStatus() == static.US_LOOKON)
	pUserData.Ready = false

	//状态效验
	if bLookonUser == true {
		xlog.Logger().Debug("read fail ,bLookonUser!")
		return false
	}
	if pUserData.GetUserStatus() >= static.US_PLAY {
		xlog.Logger().Debug("read fail ,US_PLAY!")
		return true
	}
	//设置变量
	pUserData.UserStatus = static.US_SIT
	pUserData.Ready = false

	//开始准备倒计时
	self.OnStartReadyTime(uid)

	//发送状态
	self.SendPlayStatus(pUserData)
	self.getServiceFrame().OnReady(uid, pUserData.Ready)

	//发送同意消息
	t := self.GetTableInfo()
	if t != nil {
		t.UserReady(uid, false)
		server2.GetDBMgr().GetDBrControl().Set(t.GetRedisKey(), t.ToBytes())
		server2.PushTableStatusMsg(t.HId, consts.MsgTypeGameReady, &static.UserReadyState{Ready: false, Uid: uid, Hid: t.DHId, Fid: t.FId, Tid: t.Id})
	}

	return true
}

func (self *Common) OnActionUserReady(user *Player) {
	//do nothing
	user.Ctx.Timer.KillReadyTimer()
	//self.Rule.ReadyTimeTag = -1

}

func (self *Common) OnDismissFriendMsg(uid int64, msg *static.Msg_C_DismissFriendReq) bool {
	//return self.onDismissFriendMsg(uid, msg)
	self.GetTable().Operator(base2.NewTableMsg(consts.MsgTypeGameDismissFriendReq, static.HF_JtoA(*msg), uid, nil))
	self.BSendDissmissReq = true
	return true
}

func (self *Common) OnDismissResult(uid int64, msg *static.Msg_C_DismissFriendResult) bool {
	//return self.onDismissResult(uid, msg)
	self.GetTable().Operator(base2.NewTableMsg(consts.MsgTypeGameDismissFriendResult, static.HF_JtoA(*msg), uid, nil))
	return true
}
func (self *Common) onDismissResult(uid int64, msg *static.Msg_C_DismissFriendResult) bool {
	xlog.Logger().Debug("onDismissResult  ")
	//获取用户
	UserItem := self.GetUserItemByUid(uid)
	if UserItem == nil {
		return false
	}

	var result static.Msg_C_DismissFriendResult
	result.Id = msg.Id
	result.Flag = msg.Flag
	//发送消息给本桌的所有人
	if UserItem.GetTableID() != static.INVALID_TABLE {
		dissmisscause := ""
		//判断是否为重复选择
		char_id := UserItem.GetChairID()
		if char_id < static.MAX_CHAIR_NORMAL && char_id >= 0 && self.FriendInfo.MissItem[char_id] != info2.WATING {
			return true
		}
		//通知其他玩家
		iOffLineCnt := 0
		for j := 0; j < self.GetChairCount(); j++ {

			userItme := self.GetUserItemByChair(uint16(j))
			if userItme == nil {
				iOffLineCnt++
				continue
			}
			if self.FriendInfo.MissItem[j] == info2.DISMISS_CREATOR {
				dissmisscause = self.FriendInfo.MissCause
			}
			if userItme.GetUserStatus() == static.US_OFFLINE {
				iOffLineCnt++
				continue
			}
		}
		//syslog.Logger().Debug("发送解散响应...")
		self.SendTableMsg(consts.MsgTypeGameDismissFriendResult, result)
		//发送旁观数据
		self.SendTableLookonMsg(consts.MsgTypeGameDismissFriendResult, result)
		//T处理逻辑，只有2人同意，才可以解散;1人拒绝，则立即拒绝。
		//如果其他人都断线了，也可以结束
		//记录一下申请的时候的时间
		t := time.Now()
		disRoomTimeStr := fmt.Sprintf("%d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
		self.FriendInfo.DismissOptTime[char_id] = disRoomTimeStr
		if msg.Flag {
			//日志
			self.GameTable.WriteTableLog(char_id, "同意了解散房间")

			if self.GetTableInfo().IsTeaHouse() {
				self.FriendInfo.AgreeDismiss++
				//玩家同意解散
				char_id := UserItem.GetChairID()
				if char_id < static.MAX_CHAIR_NORMAL && char_id >= 0 {
					self.FriendInfo.MissItem[char_id] = info2.AGREE
				}
				xlog.Logger().Warningf("玩家（%d）座位号（%d）同意解散", uid, char_id)
				if self.FriendInfo.DismissType == 1 {
					if ((iOffLineCnt == 0 && self.GetChairCount() == 4) && (int(self.FriendInfo.AgreeDismiss) >= self.GetChairCount()-2)) || (iOffLineCnt+int(self.FriendInfo.AgreeDismiss) >= self.GetChairCount()-1) {
						//同意
						tempChairId := static.INVALID_CHAIR //解散房间发起者
						//要通知游戏逻辑，解散桌子
						for j := 0; j < self.GetChairCount(); j++ {
							if self.FriendInfo.MissItem[j] == info2.DISMISS_CREATOR {
								tempChairId = uint16(j)
								break
							}
						}
						xlog.Logger().Println("self.FriendInfo.MissItem  : ", self.FriendInfo.MissItem)
						//syslog.Logger().Debug("房间解散:解散房间发起者->", tempChairId)
						self.DismissGame(tempChairId)

						//立即解散了，就不需要存储离线过程中的房间解散信息了
						self.FriendInfo.InitMissItem()
					}
				} else {

					if self.FriendInfo.AgreeDismiss == byte(self.GetChairCount()-1) {
						//同意
						tempChairId := static.INVALID_CHAIR //解散房间发起者
						//要通知游戏逻辑，解散桌子
						for j := 0; j < self.GetChairCount(); j++ {
							if self.FriendInfo.MissItem[j] == info2.DISMISS_CREATOR {
								tempChairId = uint16(j)
								break
							}
						}
						//syslog.Logger().Println("self.FriendInfo.MissItem  : ", self.FriendInfo.MissItem)
						//syslog.Logger().Debug("房间解散:解散房间发起者->", tempChairId)
						self.DismissGame(tempChairId)
						//立即解散了，就不需要存储离线过程中的房间解散信息了
						self.FriendInfo.InitMissItem()
					}

				}

			} else {
				//同意
				if self.FriendInfo.AgreeDismiss > 0 || (self.GetChairCount() == 4 && iOffLineCnt >= 2) || (self.GetChairCount() == 3 && iOffLineCnt >= 1) || (self.GetChairCount() == 2 && iOffLineCnt >= 0) {
					//同意
					tempChairId := static.INVALID_CHAIR //解散房间发起者
					//要通知游戏逻辑，解散桌子
					for j := 0; j < self.GetChairCount(); j++ {
						if self.FriendInfo.MissItem[j] == info2.DISMISS_CREATOR {
							tempChairId = uint16(j)
							break
						}
					}
					// syslog.Logger().Debug("房间解散:解散房间发起者->", tempChairId)
					self.DismissGame(tempChairId)
					//立即解散了，就不需要存储离线过程中的房间解散信息了
					self.FriendInfo.InitMissItem()
				} else {
					self.FriendInfo.AgreeDismiss++
					//玩家同意解散
					char_id := UserItem.GetChairID()
					if char_id < static.MAX_CHAIR_NORMAL && char_id >= 0 {
						self.FriendInfo.MissItem[char_id] = info2.AGREE
					}
					xlog.Logger().Warningf("玩家（%d）座位号（%d）同意解散", uid, char_id)
				}
			}
			if len(dissmisscause) != 0 {
				self.GameTable.WriteTableLog(static.INVALID_CHAIR, fmt.Sprintf("解散房间(%s)", dissmisscause))
			}
		} else {
			if self.RestTime != nil {
				fmt.Sprintln(fmt.Sprintf("玩家（%d）座位号（%d）拒绝解散，恢复定时器", uid, char_id))
				self.RestTime <- 1
			}

			//拒绝，nothing,继续游戏
			self.FriendInfo.AgreeDismiss = 0
			//拒绝了，就不需要存储离线过程中的房间解散信息了
			self.FriendInfo.InitMissItem()
			//日志
			self.GameTable.WriteTableLog(char_id, "拒绝了解散房间")
			self.FriendInfo.MissCause = ""
			//20191127 苏大强 崇阳新规，如果这个时候有离线玩家申请的话，就把-1改为0，这样，有新的离线玩家会重新申请，但是老的离线玩家就不会响应了
			if self.LaunchDismissTime > 0 {
				self.PlayerInfoRead(func(m *map[int64]*Player) {
					for _, v := range *m {
						if v.LaunchDismissTag == -1 {
							v.LaunchDismissTag = 0
						}
					}
				})
			}
			//---------------------
		}

	} else {
		self.SendPersonMsg(consts.MsgTypeGameDismissFriendResult, result, UserItem.GetChairID())
		//发送旁观数据
		self.SendTableLookonMsg(consts.MsgTypeGameDismissFriendResult, result)
	}

	return true
}

// 申请解散房间
func (self *Common) onDismissFriendMsg(uid int64, msg *static.Msg_C_DismissFriendReq) bool {
	//消息处理
	wChairId := self.GetChairByUid(uid)
	if wChairId == static.INVALID_CHAIR { //本桌没有这个玩家
		return false
	}

	//判断一下是否有人已经申请解散了（一般情况下不会出现，除非两个手机放在一起，玩家同时非常短的时间内点击了申请解散，界面上会出现两个申请解散）忽略掉第二个申请解散
	for i := 0; i < len(self.FriendInfo.MissItem); i++ {
		//如果已经有人在前面申请解散了，后面来的就不处理了
		if self.FriendInfo.MissItem[i] == info2.DISMISS_CREATOR {
			xlog.Logger().Debug("已经有人在前面申请解散了，后面来的就不处理了")
			return true
		}
	}

	//如果玩家已经达到了当局最大可申请解散次数,就不予处理了,这个是荆门麻将的,其它游戏暂时未添加这个功能
	//托管解散玩家发起的解散不受解散次数影响
	if requser := self.GetUserItemByChair(wChairId); requser != nil && requser.CheckTRUST() == false {
		if !self.BSendDissmissReq {
			//不是自动发起的解散有解散次数限制
			if self.DissmisReqMax > 0 && requser.DissmisReqCount >= self.DissmisReqMax {
				self.SendGameNotificationMessage(wChairId, fmt.Sprintf("一局游戏只能申请%d次解散", self.DissmisReqMax))
				//日志
				self.GameTable.WriteTableLog(wChairId, fmt.Sprintf("一局游戏只能申请%d次解散", self.DissmisReqMax))
				return true
			}
			if self.DissmisReqMax == -1 {
				self.SendGameNotificationMessage(wChairId, "本桌游玩家金币不足，房间将在15秒后自动解散。")
				//日志
				self.GameTable.WriteTableLog(wChairId, "不能申请解散，请联系盟主处理")
				return true
			}
		}
		requser.DissmisReqCount++
	}

	//计数器归0
	self.FriendInfo.AgreeDismiss = 0
	//清空解散状态记录
	self.FriendInfo.InitMissItem()
	//记录下解散倒计时的最后时间
	auto_dissmiss_time := self.GetNowSystemTimerSecond() + int64(self.DismissRoomTime)
	self.FriendInfo.Dismiss_end_time = auto_dissmiss_time

	//日志
	self.GameTable.WriteTableLog(wChairId, "申请解散房间")

	//发送自动解散的截至时间（不能混合在上面的协议中，防止影响现有老客户端）
	var dissmissRsp static.Msg_S_DismissFriendRep
	//倒计时
	dissmissRsp.Timer = self.DismissRoomTime
	dissmissRsp.Id = uid

	//通知其他玩家
	iOffLineCnt := 0
	for j := 0; j < self.GetChairCount(); j++ {
		//是否同意其他玩家等待中
		if j < static.MAX_CHAIR_NORMAL {
			self.FriendInfo.Dismiss_end_time = auto_dissmiss_time
			//自己是解散房间发起者
			if int(wChairId) == j {
				self.FriendInfo.MissItem[j] = info2.DISMISS_CREATOR
				self.FriendInfo.MissCause = msg.Reason
				//记录一下申请的时候的时间
				t := time.Now()
				disRoomTimeStr := fmt.Sprintf("%d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
				self.FriendInfo.DismissOptTime[j] = disRoomTimeStr
			} else { //其他人等待
				self.FriendInfo.MissItem[j] = info2.WATING
			}
		}

		userItem := self.GetUserItemByChair(uint16(j))
		if userItem == nil {
			iOffLineCnt++
			continue
		}

		// 记录发起解散时 每个玩家的状态
		if j < static.MAX_CHAIR_NORMAL {
			self.FriendInfo.MissStatus[j] = userItem.UserStatus_ex
		}

		if userItem.GetUserStatus() == static.US_OFFLINE {
			iOffLineCnt++
			continue
		}
	}
	dissmissRsp.Reason = msg.Reason
	fmt.Sprintln(fmt.Sprintf("玩家（%d）座位号（%d）申请解散房间", uid, wChairId))
	self.SendTableMsg(consts.MsgTypeGameDismissFriendRep, dissmissRsp)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameDismissFriendRep, dissmissRsp)
	//---------------
	if self.LaunchDismissTime > 0 {
		if self.Rule.Overtime_dismiss > 0 {
			areadyover := 0
			self.PlayerInfoRead(func(m *map[int64]*Player) {
				for _, v := range *m {
					if v.LaunchDismissTag == 0 && v.GetUserStatus() == static.US_OFFLINE {
						areadyover++
					}
				}
			})
			//两人玩，申请解散的是离线的玩家，导致正常玩家不选择就同意解散房间了
			if areadyover >= self.GetChairCount()-1 && self.GetUserItemByChair(wChairId).GetUserStatus() != static.US_OFFLINE {
				//要通知游戏逻辑，解散桌子
				self.DismissGame(wChairId)
				//立即解散了，就不需要存储离线过程中的房间解散信息了
				self.FriendInfo.InitMissItem()
			}
		}
	}

	//-------------------
	if !self.GetTableInfo().IsTeaHouse() || self.FriendInfo.DismissType == 1 {
		//普通房间逻辑,包厢房间必须要所有人同意解散才能解散房间
		//如果其他玩家都已断线，立即解散
		//20191127 苏大强 为防万一
		if iOffLineCnt >= self.GetChairCount()-1 && self.GetUserItemByChair(wChairId).GetUserStatus() != static.US_OFFLINE {
			//要通知游戏逻辑，解散桌子
			self.DismissGame(wChairId)
			//立即解散了，就不需要存储离线过程中的房间解散信息了
			self.FriendInfo.InitMissItem()
		}
	}

	return true
}

// 发送数据
func (self *Common) SendUserStatus(wChairID int, cbStatus byte) bool {
	//把我的状态广播给其他人
	userItem := self.GetUserItemByChair(uint16(wChairID))
	if userItem != nil {
		//if self.IsClientReady(self.PlayerInfo[userItem.Uid]) == true {
		self.SendUserStatusV2(userItem, cbStatus)
		return true
		//}
	}

	return true
}

// 发送状态
func (self *Common) SendUserStatusV2(userItem *Player, cbStatus byte) bool {
	//效验参数
	if userItem == nil {
		return false
	}
	//变量定义
	var UserStatus static.Msg_S_UserStatus
	//构造数据
	UserStatus.UserID = userItem.Uid
	UserStatus.TableID = self.GetTableInfo().Id
	UserStatus.ChairID = int(userItem.Seat)
	UserStatus.UserStatus = cbStatus //写死
	UserStatus.UserReady = userItem.UserReady

	if self.IsMtReady() && !self.Rule.Endready && self.GameTable.GetTableInfo().Begin && self.GameTable.GetTableInfo().Step > 0 {
		UserStatus.UserStatus = static.US_PLAY
	}

	//syslog.Logger().Debug("game SendTableMsg  UserID :" + strconv.Itoa( int(UserStatus.UserID) )  + "status:" + strconv.Itoa( int( UserStatus.UserStatus)) )
	xlog.Logger().Debug(UserStatus)
	//发给同桌的人
	self.SendTableMsg(consts.MsgTypeGameUserStatus, UserStatus)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameUserStatus, UserStatus)

	return true
}

// ! 请求信息
func (self *Common) onUserInfo(uid int64, msg *static.Msg_C_Info) bool {
	//效验状态
	person := self.GetUserItemByUid(uid)
	if person == nil {
		return false
	}

	wChairID := person.Seat
	bLookonUser := (person.UserStatus == static.US_LOOKON)
	//设置用户
	//self.m_ClientReadyUser[pUserData->dwUserID]=pUserData->dwUserID;
	//person.Ready = true

	//设置变量
	person.LookonFlag = false
	if person.UserStatus == static.US_LOOKON {
		person.LookonFlag = true
	}

	//发送配置
	var Option static.Msg_S_Option
	Option.GameStatus = self.GameStatus
	Option.AllowLookon = msg.AllowLookon
	Option.GameConfig = self.GetTableInfo().Config.GameConfig
	for i := 0; i < self.GetChairCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		Option.UserReady = append(Option.UserReady, false)
		Option.UserStatus = append(Option.UserStatus, static.US_FREE)
		if _item != nil {
			//Option.UserReady[i] = _item.UserReady
			//20200702 苏大强 这里修改会导致玩家准备后离线了，也会照旧发牌
			//Option.UserReady[i] = _item.UserStatus_ex&public.US_READY == 0
			//20200903 zwj for bug  20392 蕲春麻将（应该是公共bug）-小结算结束以后断线重连以后其他玩家显示全部准备了
			if _item.UserReady {
				Option.UserReady[i] = _item.UserReady
			} else {
				Option.UserReady[i] = _item.UserStatus_ex^static.US_READY == 0
			}
			Option.UserStatus[i] = _item.GetUserStatus()
		}
	}

	Option.GameStatus = self.GameEndStatus
	self.SendUserMsg(consts.MsgTypeGameinfo, Option, uid)

	//用户状态
	self.sendUserStartGame(person)

	self.OnSeat(uid, int(wChairID))

	//发送场景
	bSendSecret := ((bLookonUser == false) || (person.LookonFlag == true))
	self.sendTestMsg(false)
	self.getServiceFrame().SendGameScene(person.Uid, self.GameStatus, bSendSecret)

	// 用户上线事件
	self.getServiceFrame().OnLine(uid, true)

	// 同步少人开局信息
	self.OnFewerShow(person.GetChairID())

	// 同步游戏暂停信息
	self.OnSyncPauseInfo(person.GetChairID())

	//开始准备倒计时
	self.OnStartReadyTime(uid)

	//发送即时语音数据
	self.SendGVoiceMember()

	return true
}

// ! 请求信息
func (self *Common) onLookonUserInfo(uid int64, msg *static.Msg_C_Info) bool {
	//效验状态
	person := self.GetLookonUserItemByUid(uid)
	if person == nil {
		return false
	}

	wChairID := person.Seat
	bLookonUser := (person.UserStatus == static.US_LOOKON)
	//设置用户
	//self.m_ClientReadyUser[pUserData->dwUserID]=pUserData->dwUserID;
	//person.Ready = true

	//设置变量
	person.LookonFlag = false
	if person.UserStatus == static.US_LOOKON {
		person.LookonFlag = true
	}

	//发送配置
	var Option static.Msg_S_Option
	Option.GameStatus = self.GameStatus
	Option.AllowLookon = msg.AllowLookon
	Option.GameConfig = self.GetTableInfo().Config.GameConfig
	for i := 0; i < self.GetChairCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		Option.UserReady = append(Option.UserReady, false)
		Option.UserStatus = append(Option.UserStatus, static.US_FREE)
		if _item != nil {
			//Option.UserReady[i] = _item.UserReady
			//20200702 苏大强 这里修改会导致玩家准备后离线了，也会照旧发牌
			//Option.UserReady[i] = _item.UserStatus_ex&public.US_READY == 0
			//20200903 zwj for bug  20392 蕲春麻将（应该是公共bug）-小结算结束以后断线重连以后其他玩家显示全部准备了
			if _item.UserReady {
				Option.UserReady[i] = _item.UserReady
			} else {
				Option.UserReady[i] = _item.UserStatus_ex^static.US_READY == 0
			}
			Option.UserStatus[i] = _item.GetUserStatus()
		}
	}

	Option.GameStatus = self.GameEndStatus
	self.SendLookonUserMsg(consts.MsgTypeGameinfo, Option, uid)

	//用户状态
	self.sendUserStartGame(person)

	self.OnLookonSeat(uid, int(wChairID))

	//发送场景
	bSendSecret := ((bLookonUser == false) || (person.LookonFlag == true))
	self.sendTestMsg(false)
	self.getServiceFrame().SendGameScene(person.Uid, self.GameStatus, bSendSecret)

	self.SendOfflineRemainTime()

	// 同步游戏暂停信息
	self.OnSyncPauseInfoLookon(person.Uid)

	//发送即时语音数据
	self.SendGVoiceMember()

	return true
}

func (self *Common) OnStartReadyTime(uid int64) {

	person := self.GetUserItemByUid(uid)
	if person == nil {
		return
	}

	if self.IsUserPlaying(person) {
		return
	}

	if !person.Ready && self.Rule.ReadyTimeTag > 0 {
		person.Ctx.Timer.SetReadyTimer(self.Rule.ReadyTimeTag)
	}

}

// 游戏开始时，发送相关游戏消息
func (self *Common) sendUserStartGame(player *Player) {

	var rep static.Msg_S_UserGPSReq
	rep.UserID = player.Uid
	rep.Type = player.GPS_type
	rep.Latitude = player.Latitude
	rep.Longitude = player.Longitude
	rep.Addr = player.Addr

	//获取桌子
	if player.TableId != static.INVALID_TABLE {
		//通知其他玩家
		self.SendTableMsg(consts.MsgTypeGameGPSReq, rep)
	} else {
		self.SendPersonMsg(consts.MsgTypeGameGPSReq, rep, player.GetChairID())
	}
}

func (self *Common) SendAllPlayerDissmissInfo(play *Player) {
	var dissmiss_room_info static.Msg_S_DisMissRoom

	dissmiss_room_info.Reason = self.FriendInfo.MissCause

	//椅子数量跟人数量一样
	for i := uint16(0); i < uint16(self.GetChairCount()) && i < static.MAX_CHAIR; i++ {
		//item := self.GetUserItemByChair(i)
		//if item == nil {
		//	continue
		//}
		dissmiss_room_info.Situation[i] = self.FriendInfo.MissItem[i]

		if self.FriendInfo.MissItem[i] >= info2.AGREE && self.FriendInfo.MissItem[i] < info2.DISSMISS_MAX {
		}
	}

	last_time := self.FriendInfo.Dismiss_end_time - self.GetNowSystemTimerSecond()
	if last_time <= 0 {
		last_time = 0
	}
	dissmiss_room_info.Timer = last_time

	if play.LookonTableId > 0 {
		self.SendPersonLookonMsg(consts.MsgTypeGameDissMissRoom, dissmiss_room_info, play.Uid)
	} else {
		self.SendPersonMsg(consts.MsgTypeGameDissMissRoom, dissmiss_room_info, play.GetChairID())
	}
	// 发送少人开局申请信息
	// self.OnFewerApplyInfo(play)
}

func (self *Common) SendGameSceneStatusFree(player *Player) bool {
	//变量定义
	var StatusFree static.Msg_S_StatusFree
	//构造数据
	StatusFree.BankerUser = self.BankerUser
	StatusFree.CellScore = self.GetCellScore() //self.m_pGameServiceOption->lCellScore;
	StatusFree.KaiKou = self.Rule.KouKou
	StatusFree.TaskAble = self.RenWuAble
	//发送场景
	//	self.SendPersonMsg(constant.MsgTypeGameStatusFree, StatusFree, PlayerInfo.GetChairID())
	self.SendUserMsg(consts.MsgTypeGameStatusFree, StatusFree, player.Uid)
	return true
}

// ! 玩家坐下，这个坐下消息目前取消
func (self *Common) OnSeat(uid int64, seat int) bool {
	xlog.Logger().Debug("game onseat uid:" + strconv.Itoa(int(uid)) + "  seat:" + strconv.Itoa(seat))

	wTheChairID := seat //public.INVALID_CHAIR;
	var strThePassword string

	//判断位置
	if wTheChairID >= self.GetChairCount() {
		return false
	}

	//获取用户
	pUserData := self.GetUserItemByUid(uid)
	if pUserData == nil {
		return false
	}

	//发送玩家信息
	//发送自己信息
	//self.SendUserItem(pUserData, pUserData.Uid)

	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v.Seat != static.INVALID_CHAIR {
				self.SendUserItem(pUserData, v.Uid, 0)
				if v.Uid != pUserData.Uid {
					self.SendUserItem(v, pUserData.Uid, 0)
				}
			}
		}
	})
	for _, v := range self.LookonPlayer {
		if v.Seat != static.INVALID_CHAIR {
			self.SendUserItem(pUserData, v.Uid, 1)
		}
	}

	//真正坐下
	self.PerformSitDownAction(pUserData.Seat, pUserData, strThePassword)

	//庄家设置
	//	for _, v := range self.PlayerInfo {
	//		if v.Uid == uid {
	//			v.Ready = true
	//		}
	//	}

	//	if self.BankerUser == uint16(INVALID_CHAIR) {
	//		self.BankerUser = uint16(seat)
	//	}
	return true
}

// ! 玩家坐下，这个坐下消息目前取消
func (self *Common) OnLookonSeat(uid int64, seat int) bool {
	xlog.Logger().Debug("game onlookonseat uid:" + strconv.Itoa(int(uid)) + "  seat:" + strconv.Itoa(seat))

	//wTheChairID := seat //public.INVALID_CHAIR;
	//var strThePassword string

	////判断位置
	//if wTheChairID >= self.GetChairCount() {
	//	return false
	//}

	//获取用户
	pUserData := self.GetLookonUserItemByUid(uid)
	if pUserData == nil {
		return false
	}
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v.Seat != static.INVALID_CHAIR {
				self.SendUserItem(v, pUserData.Uid, 1)
			}
		}
	})

	//真正坐下
	//self.PerformSitDownAction(pUserData.Seat, pUserData, strThePassword)

	return true
}

func (self *Common) AiSupperLow() bool {
	return self.IsAiSupperLow
}

func (self *Common) ReflushUserItem() {
	if !self.IsAiSupperLow {
		return
	}
	self.BroadcastTableInfo(false)
}

// 坐下动作
func (self *Common) PerformSitDownAction(wChairID uint16, userItem *Player, szPassword string) bool {

	//设置玩家
	//userItem.AllowLookon = false
	userItem.SetUserStatus(static.US_SIT, userItem.GetTableID(), userItem.Seat)

	//发送状态
	self.SendTableStatus(self.GetTableInfo().Id)
	self.SendUserStatusV2(userItem, byte(userItem.GetUserStatus()))

	_creatType := self.FriendInfo.CreateType
	if _creatType == 1 || _creatType == 2 || _creatType == 4 {
		var _othercreateInfo rule2.Tag_OtherFriendCreate

		_othercreateInfo.UserPlayerID[wChairID] = userItem.GetUserID()
		_othercreateInfo.Type = 6
		_othercreateInfo.CreateType = self.FriendInfo.CreateType
		//_snprintf(_othercreateInfo.szGameNum, sizeof(_othercreateInfo.szGameNum),TEXT("%s"),FriendInfo.GameNum);
		_othercreateInfo.GameNum = self.FriendInfo.GameNum

		self.AddCreateOtherFriend(&_othercreateInfo)
	}

	//坐下处理
	self.OnActionUserSitDown(wChairID, userItem, false)

	return true
}

// 用户坐下
func (self *Common) OnActionUserSitDown(wChairID uint16, userItem *Player, bLookonUser bool) bool {
	//庄家设置
	if (bLookonUser == false) && (self.BankerUser == static.INVALID_CHAIR) {
		self.BankerUser = userItem.GetChairID()
	}
	return true
}

// 20181124 检查所有用户是不是都在线
func (self *Common) checkAllOnLine() bool {
	retFlag := true
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v.UserStatus == static.US_OFFLINE {
				retFlag = false
				break
			}
		}
	})
	return retFlag
}

// 玩家离线状态变更
func (self *Common) OnLine(uid int64, line bool) {
	if !self.IsGameStarted() { //游戏未开始，上下线player数据加锁
		//游戏未开始，金币场离线，踢出桌子
		if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND && !line {
			if self.OnStandup(uid, uid) {
				return
			}
		}

		self.userMutex.Lock()
		defer self.userMutex.Unlock()
	}

	item := self.GetUserItemByUid(uid)

	if item == nil {
		return
	}

	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND && !line && item.LookonFlag { //旁观玩家离线，直接踢出
		if self.OnStandup(uid, uid) {
			return
		}
	}

	if line {
		// 日志,玩家离线在线
		self.GameTable.WriteTableLog(item.Seat, "该玩家上线")
		item.UserOfflineTag = -1
		//item.LaunchDismissTag = -1
		if self.IsGameStarted() && !item.LookonFlag {
			item.UserStatus = static.US_PLAY
		} else {
			// fmt.Println("-----------------------online---------------------")
			if self.GetTableInfo().Config.GameType == static.GAME_TYPE_FRIEND {
				// if item.IsReady() {
				// 	item.UserReady = true
				// 	item.UserStatus = public.US_READY
				// } //else {
				//item.UserStatus = public.US_SIT
				//}
				// 判断开局
				if !item.LookonFlag && !self.IsGameStarted() && self.GetTable().IsNew() {
					//开始判断
					bControlStart := false
					if (bControlStart == false) && (self.StartVerdict() == true) {
						//syslog.Logger().Debug("[游戏开始]")
						xlog.Logger().Debug("read succes ,begin!")
						self.onBegin()
					}
				}
			}
		}
		//if self.checkAllOnLine() && self.RestTime != nil {
		//	self.RestTime <- 1
		//}
	} else {
		//if self.RestTime != nil {
		//	self.RestTime <- 0
		//}
		item.UserStatus = static.US_OFFLINE
		//item.UserReady = false
		if self.IsGameStarted() {
			item.UserOfflineTag = self.GetNowSystemTimerSecond()
			if self.Rule.Overtime_dismiss > 0 && self.LaunchDismissTime > 0 {
				item.LaunchDismissTag = item.UserOfflineTag
			}

			//日志
		} else {
			//游戏未开始，5分钟算断线。房主按10分钟
			if uid == self.FriendInfo.CreateUserID {
				item.UserOfflineTag = self.GetNowSystemTimerSecond()
			} else {
				item.UserOfflineTag = self.GetNowSystemTimerSecond()
			}

		}

		// 日志,玩家离线在线
		self.GameTable.WriteTableLog(item.Seat, "该玩家离线")
	}

	self.SendOfflineRemainTime()
}
func (self *Common) OnReady(uid int64, readystaus bool) { // 玩家准备或取消准备

}

// 玩家站起
func (self *Common) OnStandup(uid, by int64) bool {
	//20200317 少人开局的时候，玩家离线了，还是要踢人的
	//if self.GetUserItemByUid(uid).GetUserStatus() != public.US_OFFLINE{
	if self.isFewerApplying() {
		//如果是少人開局的情況下，不准退
		self.OnFewerClose(consts.ApplyIng, true)
	}
	//}
	//self.userMutex.Lock()
	//defer self.userMutex.Unlock()
	//效验参数
	userItem := self.GetUserItemByUid(uid)
	if userItem == nil {
		return false
	}
	//变量定义
	wChairID := userItem.GetChairID()
	//cbUserStatus := userItem.GetUserStatus()

	//用户处理
	if true { //int(userItem.GetTableID()) == self.GameTable.Id {
		/*****************************************Begin:包厢模块**********************************************************/
		// 包厢桌子，有人离开就同步下数据
		/******************************************End:包厢模块*********************************************************/

		//变量定义
		bTableLocked := self.IsTableLocked
		bGameStarted := self.m_bGameStarted

		//机器人开房，游戏没开始时，用户退出，清除用户数据
		if self.FriendInfo.CreateType == 4 && !bGameStarted {
			var _info rule2.Tag_OtherFriendCreate
			self.AddCreateOtherFriend(&_info)
		}

		//设置变量
		//userItem.AllowLookon = false
		//Delete(self.m_ClientReadyUser, userItem.Uid)

		//结束游戏
		if self.IsUserPlaying(userItem) {
			//结束游戏
			self.getServiceFrame().OnGameOver(wChairID, static.GER_USER_LEFT)
			if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND {
				self.GetTable().TableExit(userItem.Uid, by)
				self.GetTable().NotifyTableChange()
				return true
			}
		}

		//变量定义
		var LookonControl static.Msg_S_LookonControl
		LookonControl.Id = 0
		LookonControl.AllowLookon = 0

		//发送消息
		for i, _ := range self.LookonPlayer {
			//获取用户
			pILookonUserItem := self.LookonPlayer[i]

			//发送消息
			if (pILookonUserItem.GetChairID() == wChairID) && (self.IsClientReady(userItem) == true) {
				//m_pIGameServiceFrame->SendData(pILookonUserItem,MDM_GF_FRAME,SUB_GF_LOOKON_CONTROL,&LookonControl,sizeof(LookonControl));
			}
		}

		////设置用户
		//userItem.SetUserStatus(public.US_FREE, userItem.GetTableID(), wChairID)
		//self.SendPlayStatus(userItem) //要用到桌子
		//userItem.SetUserStatus(public.US_FREE, public.INVALID_TABLE, public.INVALID_CHAIR)

		ok := self.GetTable().TableExit(userItem.Uid, by) //清理框架数据
		if ok {
			if self.GetTableInfo().Config.GameType == static.GAME_TYPE_FRIEND {
				// 通大厅
				_msg := new(static.GH_TableExit_Ntf)
				_msg.TableId = self.GetTableInfo().Id
				_msg.Uid = userItem.Uid
				_msg.KindId = self.GetTableInfo().KindId
				_msg.GameId = self.GetTableInfo().GameId
				_msg.Hid = self.GetTableInfo().HId
				_msg.Fid = self.GetTableInfo().FId

				// gsvr.Protocolworkers.CallHall(constant.MsgTypeTableExit_Ntf, xerrors.SuccessCode, &_msg, userItem.Uid)
				server2.PushTableStatusMsg(self.GetTableInfo().HId, consts.MsgTypeTableExit_Ntf, &_msg)

				if !self.GetTableInfo().IsTeaHouse() {
					// 如果游戏未开始 && 不是包厢牌桌 && 房主退出, 则解散牌桌
					if !self.GetTableInfo().Begin && self.GetTableInfo().Creator == userItem.Uid {
						msg := new(static.Msg_S2C_TableDel)
						msg.Type = server2.ForceCloseByOwner
						msg.Msg = "牌桌被房主解散"
						self.GetTable().Operator(base2.NewTableMsg(consts.MsgTypeTableDel, "now", 0, msg))
					}
				}
			} else {
				self.GetTable().NotifyTableChange()
			}
		} else {
			return false
		}
		//统计人数
		wUserCount := 0
		for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
			if self.GetUserItemByChair(i) != nil {
				wUserCount++
			}
		}

		//踢走旁观
		if wUserCount == 0 {
			for i, _ := range self.LookonPlayer {
				self.SendGameMessage(self.LookonPlayer[i], "此游戏桌的所有玩家已经离开了！", static.SMT_CLOSE_GAME|static.SMT_EJECT|static.SMT_INFO)
			}
		}

		//发送状态
		if bTableLocked != self.IsTableLocked || bGameStarted != self.m_bGameStarted {
			self.SendTableStatus(self.GetTableInfo().Id)
		}

		//起立处理
		self.standupReSetBanker(wChairID)

		//变量定义
		//bMatchServer := ((self.GameTable.Config.GameType & public.GAME_GENRE_MATCH) != 0)
		// 此处GameType定义有歧义, 需要讨论后更改
		bMatchServer := false                   //((self.GameTable.Type & public.GAME_GENRE_MATCH) != 0)
		bControlStart := (bMatchServer == true) //&&m_pGameServiceOption->cbControlStart=

		//开始判断
		if (bControlStart == false) && (self.StartVerdict() == true) {
			self.onBegin()
			return true
		}
	} else {
		//旁观用户
		for i, _ := range self.LookonPlayer {
			if userItem.Uid == self.LookonPlayer[i].Uid {
				//设置变量
				//				Delete(self.m_ClientReadyUser, userItem.Uid)
				//设置用户
				delete(self.LookonPlayer, i)
				userItem.SetUserStatus(static.US_FREE, static.INVALID_TABLE, static.INVALID_CHAIR)
				self.SendPlayStatus(userItem)

				//起立处理
				self.standupReSetBanker(wChairID)
				return true
			}
		}
	}
	return true
}

// 玩家站起
func (self *Common) OnStandupLookon(uid, by int64) bool {

	userItem := self.GetLookonUserItemByUid(uid)
	if userItem == nil {
		return false
	}

	//旁观用户
	for i, v := range self.LookonPlayer {
		if v != nil && userItem.Uid == v.Uid {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%s 当前该旁观玩家【离开】桌子,gamecount:%d, wincount:%d", v.Name, v.GameCount, v.WinCount))
			delete(self.LookonPlayer, i)
			userItem.SetUserStatus(static.US_FREE, static.INVALID_TABLE, static.INVALID_CHAIR)
			break
		}
	}

	return true
}

// 发送游戏消息
func (self *Common) SendGameMessage(player *Player, lpszMessage string, wMessageType int) bool {
	//用户判断
	if self.IsClientReady(player) == false {
		if player != nil && player.LookonTableId == 0 {
			return false
		}
	}

	//构造数据包
	var Message static.Msg_S_Message
	Message.Type = wMessageType
	Message.Content = lpszMessage

	//发送数据
	if player != nil && player.LookonTableId != 0 {
		self.SendPersonLookonMsg(consts.MsgTypeGameMessage, Message, player.Uid)
	} else {
		self.SendPersonMsg(consts.MsgTypeGameMessage, Message, player.GetChairID())
	}

	return true
}

// 发送游戏通知消息
func (self *Common) SendGameNotificationMessage(seadid uint16, content string) {
	if seadid == static.INVALID_CHAIR {
		return
	}
	if seadid >= uint16(self.GetPlayerCount()) {
		return
	}
	//构造数据包
	var Message static.Msg_S_NotificationMessage
	Message.Content = content

	//发送数据
	self.SendPersonMsg(consts.MsgTypeGameNotificationMessage, Message, seadid)
}

//20191218 苏大强 由恩施麻将要拓展的消息
/*
	1.目前就是标志是癞子禁胡还是痞子禁胡
	CardClass:普通牌(0)，痞子牌(1)，癞子牌(2)
	Ctx:游戏过程（0），杠开（1），热冲（2）。。。
*/
func (self *Common) CreateNotification_ex(operateAction_ex uint64, card byte, cardClass byte, seat uint16, ctx byte) (result *static.Msg_S_Notification_ex) {
	result = &static.Msg_S_Notification_ex{
		OperateAction_ex: operateAction_ex,
		Card:             card,
		CardClass:        cardClass,
		Seat:             seat,
		Ctx:              ctx,
	}
	return result
}

// 发送游戏操作结果
func (self *Common) SendGameNotification_ex(seadid uint16, opearteuser uint16, opearteaction int, opearteresult string, exinfo *static.Msg_S_Notification_ex) {
	//构造数据包
	var Message static.Msg_S_NotificationOperateStatus
	Message.OperateUser = opearteuser
	Message.OperateAction = opearteaction
	Message.OperateResult = opearteresult
	Message.ExInfo = *exinfo
	if seadid == static.INVALID_CHAIR {
		//发送数据
		self.SendTableMsg(consts.MsgTypeGameOpStatusMessage, Message)
	} else {
		if seadid >= uint16(self.GetPlayerCount()) {
			return
		}
		//发送数据
		self.SendPersonMsg(consts.MsgTypeGameOpStatusMessage, Message, seadid)
	}
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameOpStatusMessage, Message)
}

// 发送游戏操作结果
func (self *Common) SendGameNotificationOpearteStatus(seadid uint16, opearteuser uint16, opearteaction int, opearteresult string) {
	//构造数据包
	var Message static.Msg_S_NotificationOperateStatus
	Message.OperateUser = opearteuser
	Message.OperateAction = opearteaction
	Message.OperateResult = opearteresult

	if seadid == static.INVALID_CHAIR {
		//发送数据
		self.SendTableMsg(consts.MsgTypeGameOpStatusMessage, Message)
	} else {
		if seadid >= uint16(self.GetPlayerCount()) {
			return
		}
		//发送数据
		self.SendPersonMsg(consts.MsgTypeGameOpStatusMessage, Message, seadid)
	}
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameOpStatusMessage, Message)
}

// ! 游戏数据同步
func (self *Common) OnSendInfo(person base2.PersonBase) {
	xlog.Logger().Debug("game OnSendInfo")
	var _msg static.Msg_C_Info
	_msg.AllowLookon = 0

	okFlag := false
	self.PlayerInfoWrite(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v.Uid == person.GetInfo().Uid {
				v.ReInit(person)
				okFlag = true
				break
			}
		}
	})
	if okFlag {
		self.OperateUserInfo(person.GetInfo().Uid, 0)
		return
	}

	self.PlayerInfoWrite(func(m *map[int64]*Player) {
		//self.userMutex.Lock()
		// syslog.Logger().Debug("PlayerInfo Add one ...")
		item := new(Player)
		item.Init(person, self.GetTableInfo().Config.GameType)
		item.Seat = uint16(self.GameTable.GetTableSeat(item.Uid)) //public.INVALID_CHAIR //self.getRandChair()
		// item.UserScoreInfo.Vitamin = public.SwitchVitaminToF64(self.GetUserVitamin(item, nil))
		if self.IsGameStarted() && self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND { //游戏已经开始，新用户为旁观状态
			item.UserStatus = static.US_LOOKON
		}
		(*m)[item.Uid] = item
		//self.userMutex.Unlock()
	})
	//self.onUserInfo(person.GetInfo().Uid, &_msg)
	self.OperateUserInfo(person.GetInfo().Uid, 0)
	// 清掉他的最后一局记录
	// _ = GetDBMgr().GetDBrControl().UpdateHRecordPlayers(item.Uid, nil)
}

// ! 游戏数据同步
func (self *Common) OnSendInfoLookon(person base2.PersonBase) {
	xlog.Logger().Debug("game OnSendInfo2Lookon")
	var _msg static.Msg_C_Info
	_msg.AllowLookon = 0

	for _, v := range self.LookonPlayer {
		if v.Uid == person.GetInfo().Uid {
			v.ReInit(person)
			//self.onUserInfo(person.GetInfo().Uid, &_msg)
			self.OperateUserInfo(person.GetInfo().Uid, 1)
			return
		}
	}
	self.lookonuserMutex.Lock()
	// syslog.Logger().Debug("PlayerInfo Add one ...")
	item := new(Player)
	item.Init(person, self.GetTableInfo().Config.GameType)
	//em.Uid)) //public.INVALID_CHAIR //self.getRandChair()
	// item.UserScoreInfo.Vitamin = public.SwitchVitaminToF64(self.GetUserVitamin(item, nil))
	if self.IsGameStarted() && self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND { //游戏已经开始，新用户为旁观状态
		item.UserStatus = static.US_LOOKON
	}
	if nil != self.LookonPlayer {
		self.LookonPlayer[item.Uid] = item
	} else {
		//保护一下，旧服务器没有旁观，LookonPlayer=nil，游戏中时重启服务器恢复数据时会把nil恢复到LookonPlayer中去导致问题。(实际在JsonToStruct做了保护，但不知为什么有的游戏没有作用)
		self.OnWriteGameRecord(static.INVALID_CHAIR, "error LookonPlayer == nil，请检查")
		xlog.Logger().Errorln("error LookonPlayer == nil，请检查")
		self.LookonPlayer = make(map[int64]*Player)
		self.LookonPlayer[item.Uid] = item
	}

	self.lookonuserMutex.Unlock()
	//self.onUserInfo(person.GetInfo().Uid, &_msg)
	self.OperateUserInfo(person.GetInfo().Uid, 1)
	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%s 当前该旁观玩家【进入】桌子,TotalCount:%d, wincount:%d", person.GetInfo().Nickname, person.GetInfo().TotalCount, person.GetInfo().WinCount))
	// 清掉他的最后一局记录
	// _ = GetDBMgr().GetDBrControl().UpdateHRecordPlayers(item.Uid, nil)
}

// ! 旁观玩家切换位置
func (self *Common) OnSwitchLookon(person base2.PersonBase, seat uint16) {
	xlog.Logger().Debug("game OnSwitchLookon")
	var _msg static.Msg_C_Info
	_msg.AllowLookon = 1

	for _, v := range self.LookonPlayer {
		if v.Uid == person.GetInfo().Uid {
			v.ReInit(person)
			v.Seat = seat
			self.onLookonUserInfo(v.Uid, &_msg)
			return
		}
	}
	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%s 当前该旁观玩家【切换座位号】,newseat:%d", person.GetInfo().Nickname, seat))
}

// 获取空闲的椅子
func (self *Common) GetNullChairID() uint16 {

	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		if self.GetUserItemByChair(i) == nil {
			return i
		}
	}
	return 0
}

// ! 判断庄家
func (self *Common) OnIsDealer(uid int64) bool {
	//庄家设置
	retFlag := false
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Uid == uid {
				retFlag = (v.Seat == self.BankerUser)
				break
			}
		}
	})
	return retFlag
}

// 获得基础分
func (self *Common) GetCellScore() int {
	///self.m_pGameServiceOption.lCellScore
	return 10
}

// 扣卡
func (self *Common) TableDeleteFangKa(count byte) {
	return

	if self.FriendInfo.CreateType == 1 || self.FriendInfo.CreateType == 2 || self.FriendInfo.CreateType == 4 { //为他人开房，开始时已经扣卡
		self.HaveDeleteFangKa = true
		return
	}
	// 包厢桌子让楼主支付
	bTeaHouse := false
	if self.GetTableInfo().HId > 0 {
		bTeaHouse = true
	}

	//好友房，扣房卡
	if self.GetTableInfo().Config.GameType == static.GAME_TYPE_FRIEND && !self.HaveDeleteFangKa {
		for i := 0; i < self.GetChairCount(); i++ {
			// lifei 包厢桌子让楼主支付，先找一个桌上玩家传进去
			pUserItem := self.GetUserItemByChair(uint16(i))
			if pUserItem != nil && bTeaHouse {
				self.HaveDeleteFangKa = true
				//扣房卡
				self.DeleteFangKa(pUserItem, self.FriendInfo.GameNum, self.FriendInfo.RoomID, self.FriendInfo.JuShu, self.GetPlayerCount(), self.FriendInfo.Rule, count)
				return
			}

			if pUserItem != nil && pUserItem.GetUserID() == self.FriendInfo.CreateUserID {
				self.HaveDeleteFangKa = true
				//扣房卡
				self.DeleteFangKa(pUserItem, self.FriendInfo.GameNum, self.FriendInfo.RoomID, self.FriendInfo.JuShu, self.GetPlayerCount(), self.FriendInfo.Rule, count)
				return
			}
		}
	}
}

func (self *Common) DeleteFangKa(userItem *Player, pszGameNum string, dwTableNum int, nCount uint32, nPlayers int, pszRule string, nPlayCount byte) {
	if userItem == nil {
		return
	}

	uid := userItem.GetUserID()
	// 检查看是不是包厢桌子
	if self.GetTableInfo().IsTeaHouse() {
		uid = self.GetTableInfo().Creator //self.FriendInfo.teaHouseCaptionId
	}
	//扣卡
	card := -1 * self.GetTableInfo().Config.CardCost

	//打0局扣0张房卡
	if nPlayCount == 0 {
		card = 0
	}
	if self.GetTableInfo().IsTeaHouse() && consts.ClubHouseOwnerPay == false {
		for _, tableUser := range self.GetTableInfo().Users {
			if tableUser != nil && tableUser.Uid > 0 && tableUser.Payer > 0 {
				xlog.Logger().Debugf("玩家:%d, 支付:%d, 扣卡:%d", tableUser.Uid, tableUser.Payer, card)
				after, err := self.GetTable().CostUserWealth(tableUser.Payer, consts.WealthTypeCard, card, models.CostTypeGame, self.GetProperPNum())
				if err != nil {
					xlog.Logger().Errorf("玩家:%d, 支付:%d, 扣卡失败err=%s", tableUser.Uid, tableUser.Payer, err)
				} else {
					xlog.Logger().Errorf("玩家:%d, 支付:%d, afterCard=%d", tableUser.Uid, tableUser.Payer, after)
				}
			}
		}
		self.GetTable().SetCostPaid()
	} else {
		xlog.Logger().Debugf("玩家:%d, 支付:%d, 扣卡:%d", userItem.Uid, uid, card)
		after, err := self.GetTable().CostUserWealth(uid, consts.WealthTypeCard, card, models.CostTypeGame, self.GetProperPNum())
		if err != nil {
			xlog.Logger().Errorf("玩家:%d, 支付:%d, 扣卡失败err=%s", userItem.Uid, uid, err)
		} else {
			xlog.Logger().Errorf("玩家:%d, 支付:%d, afterCard=%d", userItem.Uid, uid, after)
		}
		self.GetTable().SetCostPaid()
	}
}

// 通知框架结束游戏
func (self *Common) ConcludeGame() {

	self.OnGameEnd()
	self.SetGameStatus(static.GS_MJ_FREE)

	self.m_bGameStarted = false
	self.DissmisReqMax = 0
	self.IsSetDissmiss = false
	self.GameTimer.Clean() //清除计时器
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil {
				v.Ctx.Timer.Clean()
				v.ResetDissmiss() //清除玩家当局申请解散次数
			}
		}
	})

	if self.GetTableInfo().Config.GameType == static.GAME_TYPE_FRIEND {
		// 预判作弊玩家
		self.AnalyticalCribber()
		// 释放作弊玩家
		self.AnalyticalReleaseCribber()
		self.GameTable.Bye()
	} else {
		self.GetTable().SetBegin(false)
		_siteTmp := server2.GetSiteMgr().GetSiteByType(self.GetTableInfo().KindId, self.GetTableInfo().SiteType)
		self.PlayerInfoRead(func(m *map[int64]*Player) {
			for _, v := range *m {
				if (_siteTmp.MinGold > 0 && v.UserScoreInfo.Score < _siteTmp.MinGold) || (_siteTmp.MaxGold > 0 && v.UserScoreInfo.Score > _siteTmp.MaxGold) || v.UserStatus == static.US_OFFLINE {
					if v.UserScoreInfo.Score < _siteTmp.MinGold {
						self.SendGameMessage(v, xerrors.GoldNotEnoughError.Msg, static.SMT_CLOSE_GAME)
					} else if v.UserScoreInfo.Score > _siteTmp.MaxGold {
						self.SendGameMessage(v, xerrors.GoldExceedingError.Msg, static.SMT_CLOSE_GAME)
					}

					self.GetTable().TableExit(v.Uid, v.Uid)
					self.GetTable().NotifyTableChange()
				} else {
					v.Ready = false
					v.LookonFlag = false
					v.SetUserStatus(static.US_SIT, v.GetTableID(), v.GetChairID())
					self.SendPlayStatus(v) //用户状态发送
				}
			}
		})

		self.SendTableStatus(self.GetTableId())
		self.BroadcastTableInfo(true) //桌子数据更新
	}

	self.StatisticsUserGameOpt()
}

// 校验欢乐豆是否充足
func (self *Common) CheckCoin(seat uint16) bool {
	_item := self.GetUserItemByChair(seat)
	if _item == nil {
		return false
	}
	_siteTmp := server2.GetSiteMgr().GetSiteByType(self.GetTableInfo().KindId, self.GetTableInfo().SiteType)

	if (_siteTmp.MinGold > 0 && _item.UserScoreInfo.Score < _siteTmp.MinGold) || (_siteTmp.MaxGold > 0 && _item.UserScoreInfo.Score > _siteTmp.MaxGold) {
		if _item.UserScoreInfo.Score < _siteTmp.MinGold {
			self.SendGameMessage(_item, xerrors.GoldNotEnoughError.Msg, static.SMT_CLOSE_GAME)
		} else if _item.UserScoreInfo.Score > _siteTmp.MaxGold {
			self.SendGameMessage(_item, xerrors.GoldExceedingError.Msg, static.SMT_CLOSE_GAME)
		}
		self.GetTable().TableExit(_item.Uid, _item.Uid)
		self.GetTable().NotifyTableChange()
		return false
	}
	return true
}

// 获取系统时间
func (self *Common) GetNowSystemTimerSecond() int64 {
	return time.Now().Unix()
}

// 自动解散房间
func (self *Common) DissmissRoomAuto() {

	if self.FriendInfo.Dismiss_end_time != 0 {
		if self.GetNowSystemTimerSecond() > self.FriendInfo.Dismiss_end_time {
			var buf string = "解散房间倒计时到了，解散房间： wTableID:" + strconv.Itoa(self.GetTableId()) + " GetChairCount:" + strconv.Itoa(self.GetChairCount())
			xlog.Logger().Debug(buf)

			// 游戏日志
			self.GameTable.WriteTableLog(static.INVALID_CHAIR, "申请解散超时，解散游戏房间")

			//要通知游戏逻辑，解散桌子
			tempChairId := static.INVALID_CHAIR //解散房间发起者
			//要通知游戏逻辑，解散桌子
			for j := 0; j < self.GetChairCount(); j++ {
				if self.FriendInfo.MissItem[j] == info2.DISMISS_CREATOR {
					tempChairId = uint16(j)
					break
				}
			}
			self.DismissGame(tempChairId)
			//清空解散房间状态记录数据
			self.FriendInfo.InitMissItem()
		}
	}
}

func (self *Common) DismissGame(wChairID uint16) bool {
	// syslog.Logger().Warningln("DismissGame")
	//状态判断
	/*存在一个bug 当有吓跑的时候，如果这时候申请解散，会解散不了，原因是游戏没有开始直接返回了  self.GetTable().OnBegin()未调用
	但是这个又不能调用，因为CheckErrorDismissAuto()这个函数会检查局数为0（局数确实为0因为还没发牌）并且游戏未开始的这
	种，把这种桌子视为非正常情况的桌子会直接解散掉桌子。添加IsXiaPaoIng()函数过滤掉第一局吓跑中不能解散的问题。
	*/
	if self.GetTable().IsNew() && (!self.GetTable().IsXiaPaoIng()) {
		xlog.Logger().Warningln("DismissGame,游戏没开始")
		return false
	}

	// 在真正解散前分析一波解散
	//self.AnalyzeDismissRoom()

	//todo  在这里记录解散的房间的 解散类型
	self.AnalyzeDismissRoomV2()

	//结束游戏
	if self.getServiceFrame().OnGameOver(wChairID, static.GER_DISMISS) == false {
		//ASSERT(FALSE);
		//syslog.Logger().Debug("DismissGame,结束游戏失败")
		return false
	}

	//设置状态
	if self.m_bGameStarted != false {
		//设置变量
		self.m_bGameStarted = false
		//发送状态
		self.SendTableStatus(self.GetTableInfo().Id)
	}
	/*****************************************Begin:包厢模块**********************************************************/
	// 包厢桌子解散时，同步一下数据
	if len(self.FriendInfo.TeaHouseID) > 0 {
		self.SendTeaHouseTableDismissMsg(self.FriendInfo.TeaHouseTableIndex, self.FriendInfo.TeaHouseID)
	}
	/*****************************************End:包厢模块**********************************************************/
	return true
}

func (self *Common) AnalyzeDismissRoom() {
	var (
		checkTrustFunc = func(status uint64) bool { // 查托管状态
			return status&static.US_TRUST != 0
		}
		flag bool   // 托管解散标识
		log  string // 输出日志
	)
	// 分析是否为托管者发起解散的
	for i := 0; i < self.GetChairCount(); i++ {
		if i < static.MAX_CHAIR_NORMAL {
			// 如果解散房间申请者 申请解散时 是托管状态，则认为该游戏是因为托管才解散的
			if self.FriendInfo.MissItem[i] == info2.DISMISS_CREATOR && checkTrustFunc(self.FriendInfo.MissStatus[i]) {
				flag = true
			}
		}
	}
	if flag {
		self.DismissTrustee = make([]uint16, 0)
		for i := 0; i < self.GetChairCount(); i++ {
			if i < static.MAX_CHAIR_NORMAL {
				if checkTrustFunc(self.FriendInfo.MissStatus[i]) {
					if user := self.GetUserItemByChair(uint16(i)); user != nil {
						self.DismissTrustee = append(self.DismissTrustee, user.GetChairID())
						if self.FriendInfo.MissItem[i] == info2.DISMISS_CREATOR {
							log += fmt.Sprintf("托管用户(发起者):%s[%d],", user.Name, user.Uid)
						} else {
							log += fmt.Sprintf("托管用户:%s[%d],", user.Name, user.Uid)
						}
					}
				}
			}
		}
		if self.IsTrusteeDismiss() {
			xlog.Logger().Infof("托管解散：%s", log)
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("托管解散：%s", log))
		}
	}
}

func (self *Common) AnalyzeDismissRoomV2() { //30需求 中途解散详情 在这个函数记录
	var (
		checkTrustFunc = func(status uint64) bool { // 查托管状态
			return status&static.US_TRUST != 0
		}
		flag bool     // 托管解散标识
		log  []string // 输出日志
	)
	dismissType := 4 // 0：无,  1：盟主解散，2：管理员解散 3：队长解散 4：申请解散 5：超时解散  6：托管解散
	t := time.Now()
	disRoomTimeStr := fmt.Sprintf("%d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	// 分析是否为托管者发起解散的
	for i := 0; i < self.GetChairCount(); i++ {
		if i < static.MAX_CHAIR_NORMAL {
			// 如果解散房间申请者 申请解散时 是托管状态，则认为该游戏是因为托管才解散的
			if self.FriendInfo.MissItem[i] == info2.DISMISS_CREATOR && checkTrustFunc(self.FriendInfo.MissStatus[i]) {
				flag = true
			}
		}
	}
	if flag {
		self.DismissTrustee = make([]uint16, 0)
		for i := 0; i < self.GetChairCount(); i++ {
			if i < static.MAX_CHAIR_NORMAL {
				if checkTrustFunc(self.FriendInfo.MissStatus[i]) {
					if user := self.GetUserItemByChair(uint16(i)); user != nil {
						self.DismissTrustee = append(self.DismissTrustee, user.GetChairID())
						if self.FriendInfo.MissItem[i] == info2.DISMISS_CREATOR {
							rear := append([]string{}, log[0:]...)
							log = append(log[0:0], fmt.Sprintf("%s申请解散 %s", user.Name, self.FriendInfo.DismissOptTime[i]))
							log = append(log, rear...)
							//log = append([]string{""}, fmt.Sprintf("%s申请解散 %s", user.Name, self.FriendInfo.DismissOptTime[i]))
							//log += fmt.Sprintf("%s申请解散 %s,", user.Name, self.FriendInfo.DismissOptTime[i])
						} else {
							log = append(log, fmt.Sprintf("%s 托管", user.Name))
							//log += fmt.Sprintf("%s托管用户,", user.Name)
						}
					}
				}
			}
		}
		if !self.IsTrusteeDismiss() {
			log = []string{""}
		} else {
			//log = []string{fmt.Sprintf("此局游戏为托管解散 %s", disRoomTimeStr)}
			dismissType = 6
		}
	} else {
		for i := 0; i < self.GetChairCount(); i++ {
			if i < static.MAX_CHAIR_NORMAL {
				if user := self.GetUserItemByChair(uint16(i)); user != nil {
					if self.FriendInfo.MissItem[i] == info2.DISMISS_CREATOR {
						rear := append([]string{}, log[0:]...)
						log = append(log[0:0], fmt.Sprintf("%s申请解散 %s", user.Name, self.FriendInfo.DismissOptTime[i]))
						log = append(log, rear...)
						//log = append([]string{""}, fmt.Sprintf("%s申请解散 %s", user.Name, self.FriendInfo.DismissOptTime[i]))
						//log += fmt.Sprintf("%s申请解散 %s,", user.Name, self.FriendInfo.DismissOptTime[i])
					} else if self.FriendInfo.MissItem[i] == info2.AGREE {
						log = append(log, fmt.Sprintf("%s同意解散 %s", user.Name, self.FriendInfo.DismissOptTime[i]))
						//log += fmt.Sprintf("%s同意解散 %s,", user.Name, self.FriendInfo.DismissOptTime[i])
					} else if self.FriendInfo.MissItem[i] == info2.WATING {
						log = append(log, fmt.Sprintf("%s 超时", user.Name))
						//log += fmt.Sprintf("%s同意解散 %s,", user.Name, self.FriendInfo.DismissOptTime[i])
					}
				}
			}
		}
		if self.FriendInfo.Dismiss_end_time != 0 {
			if self.GetNowSystemTimerSecond() > self.FriendInfo.Dismiss_end_time {
				dismissType = 5
			}
		}
	}
	templog := ""
	for i := 0; i < len(log); i++ {
		templog += fmt.Sprintf("%s,", log[i])
	}

	templog = strings.TrimRight(templog, ",")
	recordDismiss := new(models.RecordDismiss)
	recordDismiss.GameNum = self.GetTableInfo().GameNum
	recordDismiss.DismissTime = disRoomTimeStr
	recordDismiss.DismissType = dismissType
	recordDismiss.DismissDet = templog
	// 写数据库记录
	if err := server2.GetDBMgr().GetDBmControl().Create(&recordDismiss).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("战绩中途解散写入详情出错：（%v）", err))
	}

	//syslog.Logger().Println("AnalyzeDismissRoomV2现在开始解散:", templog, dismissType)
}

// 记录出牌超时解散详情
func (self *Common) OnWriteOutCardOverTimeDismissRecord(userName string, timeSec int) {
	t := time.Now()
	dismissTime := fmt.Sprintf("%d:%02d:%02d", t.Hour(), t.Minute(), t.Second())

	recordDismiss := new(models.RecordDismiss)
	recordDismiss.GameNum = self.GetTableInfo().GameNum
	recordDismiss.DismissTime = dismissTime
	recordDismiss.DismissType = 8
	recordDismiss.DismissDet = fmt.Sprintf("%s超过%d分钟未操作，房间解散", userName, timeSec/60)

	// 写数据库记录
	if err := server2.GetDBMgr().GetDBmControl().Create(&recordDismiss).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("战绩中途解散写入详情出错：（%v）", err))
	}
}

// 是否为托管解散
func (self *Common) IsTrusteeDismiss() bool {
	return self.DismissTrustee != nil && len(self.DismissTrustee) > 0
}

// 发送状态
func (self *Common) SendTableStatus(wTableID int) bool {
	//构造变量
	var TableStatus static.Msg_S_TableStatus
	TableStatus.TableID = uint16(wTableID)
	TableStatus.TableLock = self.IsTableLocked
	TableStatus.PlayStatus = self.m_bGameStarted

	//发送数据
	//m_pITCPNetworkEngine->SendDataBatch(MDM_GR_STATUS,SUB_GR_TABLE_STATUS,&TableStatus,sizeof(TableStatus));
	//m_AndroidUserManager.SendDataToClient(MDM_GR_STATUS,SUB_GR_TABLE_STATUS,&TableStatus,sizeof(TableStatus));
	//syslog.Logger().Debug("发送桌子状态")
	self.SendTableMsg(consts.MsgTypeGameTableStatus, TableStatus)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameTableStatus, TableStatus)

	return true
}

// 发送准备消息
func (self *Common) sendUserReadyState(player *Player, states []bool, n uint16) bool {
	var redy_info static.Msg_S_PLAYER_REDEAY_STATE

	b_find := false
	//椅子数量跟人数量一样
	for i := uint16(0); i < uint16(self.GetChairCount()) && i < static.MAX_CHAIR; i++ {
		//if (self.m_pIUserItem[i]==NULL)  continue;
		item := self.GetUserItemByChair(i)
		//int chairid = m_pIUserItem[i]->GetChairID();
		if item == nil {
			continue
		}
		if i >= 0 && (i < uint16(self.GetChairCount()) && i < static.MAX_CHAIR_NORMAL && i < n) {
			redy_info.Situation[i] = states[i]
			if states[i] {
				b_find = true
			}
		}
	}

	if !b_find { //如果没有，就不要发送了
		return false
	}

	self.SendPersonMsg(consts.MsgTypeGameReadyState, redy_info, player.GetChairID())
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameReadyState, redy_info)

	return true
}

func (self *Common) SendTeaHouseTableDismissMsg(tableid uint32, houserid string) {

}

func (self *Common) IsGameStarted() bool {
	return self.m_bGameStarted
}

// 计时器事件走table消息信道，防止阻塞计时器，造成并发问题
func (self *Common) SendTimeMsg(dwTimerID uint16, uid int64) {
	msg := new(static.Msg_TimeEventMsg)
	msg.Uid = uid
	msg.Id = dwTimerID
	self.GameTable.Operator(base2.NewTableMsg(consts.MsgTypeGameTimeMsg, "now", msg.Uid, msg))
}

func (self *Common) OperateUserInfo(_uid int64, _lookon int) {
	var _msg static.Msg_C_Info
	_msg.AllowLookon = _lookon

	self.GameTable.Operator(base2.NewTableMsg(consts.MsgTypeGameinfo, static.HF_JtoA(&_msg), _uid, nil))
}

// 计时器事件
func (self *Common) OnTimerEvent(dwTimerID uint16, wBindParam int64) bool {
	//桌子定时器
	// if dwTimerID > public.IDI_MAX_TIME_ID {
	// }
	switch dwTimerID {
	case GameTime_AutoNext: //自动开始下一局
		self.AutoNextGame(wBindParam)
	case GameTime_Ready: //准备超时
		self.ReadyTimeEvent(wBindParam)
	default:
		self.getServiceFrame().OnEventTimer(dwTimerID, wBindParam)
	}

	//游戏定时器
	//return m_pITableFrameSink->OnTimerMessage((WORD)dwTimerID,wBindParam);
	return true
}

func (self *Common) ReadyTimeEvent(userId int64) {
	_userItem := self.GetUserItemByUid(userId)
	if _userItem != nil && !self.IsUserPlaying(_userItem) { //玩家没有准备,自动准备
		//_msg := public.Msg_C_GoOnNextGame{Id: id}
		//self.getServiceFrame().OnUserClientNextGame(&_msg)
		_msg := "你长时间未准备，被踢出房间"
		self.OnWriteGameRecord(_userItem.Seat, _msg)
		self.SendGameNotificationMessage(_userItem.Seat, _msg)
		self.OnStandup(userId, userId)
		self.GetTable().Bye()
	}
}

func (self *Common) AutoNextGame(id int64) {
	if !self.CanContinue() {
		return
	}
	// 局数校验 大于当前总局数，不处理此消息，且没有开始游戏
	if (!self.Rule.Always1Round && int(self.CurCompleteCount) >= self.Rule.JuShu) || !self.m_bGameStarted {
		return
	}
	//没有开启一局
	if self.GameEndStatus == static.GS_MJ_PLAY {
		return
	}

	_userItem := self.GetUserItemByUid(id)
	if _userItem != nil && !_userItem.UserReady { //玩家没有准备,自动准备
		_msg := static.Msg_C_GoOnNextGame{Id: id}
		self.getServiceFrame().OnUserClientNextGame(&_msg)
	}

}

func (self *Common) SetAutoNextTimer(leftTime int) {
	if !self.Rule.Always1Round && int(self.CurCompleteCount) >= self.Rule.JuShu { //局数够了
		return
	}
	if leftTime <= 0 {
		leftTime = GAME_OPERATION_TIME_AUTONEXT
	}
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil {
				v.Ctx.EndAutoReadyTimeOut = time.Now().Unix() + int64(leftTime)
				v.Ctx.Timer.SetTimer(GameTime_AutoNext, leftTime)
			}
		}
	})
}

// 指定玩家自动准备
func (self *Common) SetPlayerAutoNextTimer(uSeat uint16, leftTime int) {
	if (!self.Rule.Always1Round && int(self.CurCompleteCount) >= self.Rule.JuShu) || int(uSeat) >= self.GetPlayerCount() { //局数够了
		return
	}
	if leftTime <= 0 {
		leftTime = GAME_OPERATION_TIME_AUTONEXT
	}
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v.GetChairID() == uSeat {
				v.Ctx.EndAutoReadyTimeOut = time.Now().Unix() + int64(leftTime)
				v.Ctx.Timer.SetTimer(GameTime_AutoNext, leftTime)
				break
			}
		}
	})
}

// 广播断线剩余时间
func (self *Common) SendOfflineRemainTime() {
	//检查是否有人断线
	var Message static.Msg_S_UserOfflineTime
	_nowTime := self.GetNowSystemTimerSecond()
	_tmpTime := 0
	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		userItem := self.GetUserItemByChair(i)
		_tmpTime = 0
		if userItem != nil {
			Message.Code[i] = userItem.GetUserStatus()
			//20210121 苏大强 恩施玩法累计时
			if userItem.GetUserStatus() == static.US_OFFLINE {
				//剩余秒数
				if userItem.UserOfflineTag > 0 {
					_tmpTime = self.OfflineRoomTime - int(_nowTime-userItem.UserOfflineTag)
					if _tmpTime < 0 {
						_tmpTime = 0
					}
					Message.Time[i] = _tmpTime
				}
				//20191127追加离线申请解散房间时间
				if userItem.LaunchDismissTag > 0 {
					_tmpTime = self.LaunchDismissTime - int(_nowTime-userItem.LaunchDismissTag)
					if _tmpTime < 0 {
						_tmpTime = 0
					}
					Message.ShortOfflineTime[i] = _tmpTime
				} else {
					Message.ShortOfflineTime[i] = Message.Time[i]
				}
			}
			self.OnWriteGameRecord(i, fmt.Sprintf("记录所有玩家的状态 status = %d 【6是断线，5是上线】", Message.Code[i]))
		}
		//Message.Time[i] = _tmpTime
	}
	//这里创建一下累计时间
	Message.TotalTimerUser, Message.TotalTime = self.ModifyTotalInfo()
	//广播通知
	self.SendTableMsg(consts.MsgTypeGameUserOfflineTime, Message)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameUserOfflineTime, Message)
}

// 自动发起解散房间
func (self *Common) LaunchDismissAuto() {
	//////////////////////////////////////////////////////////////////////////
	if self.LaunchDismissTime < 1 {
		return
	}
	//离线超时处理
	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND { //非好友房，不处理
		return
	}
	_lTime := self.GetNowSystemTimerSecond() //sys.wYear*365*24*60+sys.wMonth *30*24*60 + sys.wDay*24*60 + sys.wHour*60 + sys.wMinute;
	oldOffline := _lTime
	wChairID := static.INVALID_CHAIR
	//查找离线时间最长的玩家
	areadyLaunch := false
	overtimeuser := []uint16{}
	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		userItem := self.GetUserItemByChair(i)
		if userItem == nil {
			continue
		}
		switch userItem.LaunchDismissTag {
		case -1:
			//已经申请或者申请完了
			if !areadyLaunch {
				//fmt.Println(fmt.Sprintf("玩家（%d）LaunchDismissTag(%d)",userItem.Seat,userItem.LaunchDismissTag))
				areadyLaunch = true
			}
		case 0:
			//在线或者已经选择
			if userItem.GetUserStatus() == static.US_OFFLINE {
				//如果是离线的情况下，已经走完时间的玩家要选择同意，避免有非托管玩家不小心在第一次申请解散的时候选了拒绝，导致最早申请的离线玩家不反应
				overtimeuser = append(overtimeuser, i)
			} else {
				continue
			}
		default:
			if userItem != nil && userItem.LaunchDismissTag < oldOffline && _lTime-userItem.LaunchDismissTag >= int64(self.LaunchDismissTime) {
				xlog.Logger().Debug(fmt.Sprintf("玩家离线超时申请解散房间：%+v", userItem))
				oldOffline = userItem.LaunchDismissTag
				wChairID = i
			}
		}
	}
	//找到最早的吧
	if wChairID != static.INVALID_CHAIR {
		pIServerUserItem := self.GetUserItemByChair(wChairID)
		//断线处理
		if pIServerUserItem != nil {
			//状态判断
			if pIServerUserItem.GetUserStatus() != static.US_OFFLINE {
				return
			}
			pIServerUserItem.LaunchDismissTag = -1
			if areadyLaunch {
				var _msg = &static.Msg_C_DismissFriendResult{
					Id:   pIServerUserItem.Uid,
					Flag: true,
				}
				self.OnDismissResult(pIServerUserItem.Uid, _msg)
				self.GameTable.WriteTableLog(pIServerUserItem.Seat, "离线超时，自动同意解散游戏房间")
			} else {
				var msg = &static.Msg_C_DismissFriendReq{
					Id: pIServerUserItem.Uid,
				}
				self.SetDismissRoomTime(self.Rule.Overtime_dismiss)
				self.OnDismissFriendMsg(pIServerUserItem.Uid, msg)
				self.GameTable.WriteTableLog(pIServerUserItem.Seat, "离线超时，申请解散游戏房间")
			}
		}
		if len(overtimeuser) != 0 {
			for _, v := range overtimeuser {
				overItem := self.GetUserItemByChair(v)
				//已经被忽略的离线玩家
				var _msg = &static.Msg_C_DismissFriendResult{
					Id:   overItem.Uid,
					Flag: true,
				}
				self.OnDismissResult(overItem.Uid, _msg)
				self.GameTable.WriteTableLog(overItem.Seat, "离线超时，自动同意解散游戏房间")
			}
		}
	}
}

// 检测离线
func (self *Common) CheckOfflineTime() {
	//////////////////////////////////////////////////////////////////////////
	//离线超时处理
	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND { //非好友房，不处理
		return
	}
	var log []string // 输出日志
	// t := time.Now()
	// disRoomTimeStr := fmt.Sprintf("%d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	_lTime := self.GetNowSystemTimerSecond() //sys.wYear*365*24*60+sys.wMonth *30*24*60 + sys.wDay*24*60 + sys.wHour*60 + sys.wMinute;

	oldOffline := _lTime
	wChairID := static.INVALID_CHAIR
	//查找离线时间最长的玩家
	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		userItem := self.GetUserItemByChair(i)
		if userItem != nil && userItem.UserOfflineTag > 0 && _lTime-userItem.UserOfflineTag >= int64(self.OfflineRoomTime) && userItem.UserOfflineTag < oldOffline {
			xlog.Logger().Debug(fmt.Sprintf("玩家离线超时：%+v", userItem))
			oldOffline = userItem.UserOfflineTag
			log = append(log, fmt.Sprintf("%s 离线超时", userItem.Name))
			//fmt.Println(fmt.Sprintf("参数1（%d）参数2（%d）参数3（%d）参数4（%d）参数5（%d）", _lTime, userItem.UserOfflineTag, _lTime-userItem.UserOfflineTag, self.OfflineRoomTime, oldOffline))
			wChairID = i
		}
	}

	//有人离线超时
	if wChairID != static.INVALID_CHAIR {
		pIServerUserItem := self.GetUserItemByChair(wChairID)
		//断线处理
		if pIServerUserItem != nil {
			pIServerUserItem.UserOfflineTag = -1
			//状态判断
			if pIServerUserItem.GetUserStatus() != static.US_OFFLINE {
				return
			}

			templog := ""
			for i := 0; i < len(log); i++ {
				templog += fmt.Sprintf("%s,", log[i])
			}
			templog = strings.TrimRight(templog, ",")
			status := self.GetGameStatus()
			if status != static.GS_MJ_FREE {
				self.getServiceFrame().OnEventTimer(GameTime_TuoGuan, pIServerUserItem.Uid)
				//recordDismiss := new(models.RecordDismiss)
				//recordDismiss.GameNum = self.GetTableInfo().GameNum
				//recordDismiss.DismissTime = disRoomTimeStr
				//recordDismiss.DismissType = 7 // 7 为离线解散
				//recordDismiss.DismissDet = templog
				//// 写数据库记录
				//if err := server2.GetDBMgr().GetDBmControl().Create(&recordDismiss).Error; err != nil {
				//	xlog.Logger().Errorln(fmt.Sprintf("战绩中途解散写入详情出错：（%v）", err))
				//} else {
				//	// 日志
				//	self.GameTable.WriteTableLog(pIServerUserItem.Seat, "离线超时，解散游戏房间")
				//}
			} else {
				////用户起来
				self.OnStandup(pIServerUserItem.Uid, pIServerUserItem.Uid)
			}
		}
	}
}

// 检查房间是否出错
// 注意：游戏服务器中 CurCompleteCount++要在self.GetTable().SetBegin(true)之前调用，建议它们一起在OnGameStart()中被调用，
//
//	 self.CurCompleteCount++
//		self.GetTable().SetBegin(true)
func (self *Common) CheckErrorDismissAuto() bool {
	//////////////////////////////////////////////////////////////////////////
	//离线超时处理
	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND { //非好友房，不处理
		return false
	}
	if self.GameTable.GetShutDownFlag() {
		return false
	}
	//程序出错时，找个时机把游戏解散掉
	if self.CurCompleteCount == 0 && self.GameTable.IsBegin() {
		//恢复失败 ,需要先等客户端接收到tablein，否则客户端不会处理解散消息
		//self.OnWriteGameRecord(public.INVALID_CHAIR, "异常，无法恢复牌局，系统解散")
		//结束游戏
		self.getServiceFrame().OnGameOver(static.INVALID_CHAIR, static.GER_GAME_ERROR)
		return true
	}
	return false
}

// //手动准备的游戏
func (self *Common) IsMtReady() bool {
	if self.GetTableInfo().KindId == base2.KIND_ID_DY_DYHH || self.GetTableInfo().KindId == base2.KIND_ID_DAYEKAIKOUFAN {
		return true
	}
	return false
}

// 检测竞技点过低游戏暂停
func (self *Common) CheckVitaminLowPauseTime() {
	//////////////////////////////////////////////////////////////////////////
	//离线超时处理
	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND { //非好友房，不处理
		return
	}

	_lTime := self.GetNowSystemTimerSecond() //sys.wYear*365*24*60+sys.wMonth *30*24*60 + sys.wDay*24*60 + sys.wHour*60 + sys.wMinute;

	oldOffline := _lTime
	wChairID := static.INVALID_CHAIR
	//查找竞技点过低暂停游戏时间最长的玩家
	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		userItem := self.GetUserItemByChair(i)
		if userItem != nil && userItem.UserVitaminLowPauseTag > 0 && _lTime-userItem.UserVitaminLowPauseTag >= int64(self.VitaminLowPauseTime-1) && userItem.UserVitaminLowPauseTag < oldOffline {
			xlog.Logger().Debug(fmt.Sprintf("玩家竞技点过低,游戏暂停超时：%+v", userItem))
			oldOffline = userItem.UserVitaminLowPauseTag
			wChairID = i
		}
	}

	//有人超时
	if wChairID != static.INVALID_CHAIR {

		pIServerUserItem := self.GetUserItemByChair(wChairID)

		//断线处理
		if pIServerUserItem != nil {

			pIServerUserItem.UserVitaminLowPauseTag = -1

			////状态判断
			//if pIServerUserItem.GetUserStatus() != public.US_OFFLINE {
			//	return
			//}
			// 日志
			self.GameTable.WriteTableLog(pIServerUserItem.Seat, "玩家竞技点过低,游戏暂停超时，解散游戏房间")

			if self.IsUserPlaying(pIServerUserItem) {
				//结束游戏
				self.getServiceFrame().OnGameOver(wChairID, static.GER_DISMISS)
			} else {
				//用户起来
				self.OnStandup(pIServerUserItem.Uid, pIServerUserItem.Uid)
			}
		}
	}
}

// 广播断线剩余时间
func (self *Common) SendVitaminLowRemainTime() {
	//检查是否有人断线
	var Message static.Msg_S_UserVitaminLowTime
	_nowTime := self.GetNowSystemTimerSecond()
	_tmpTime := 0
	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		userItem := self.GetUserItemByChair(i)
		_tmpTime = 0
		if userItem != nil {
			Message.Code[i] = userItem.GetUserStatus()
			if userItem.GetUserStatus() == static.US_OFFLINE {

				//剩余秒数
				if userItem.UserVitaminLowPauseTag > 0 {
					_tmpTime = self.VitaminLowPauseTime - int(_nowTime-userItem.UserVitaminLowPauseTag)
				}
				if _tmpTime < 0 {
					_tmpTime = 0
				}
			}
		}
		Message.Time[i] = _tmpTime

	}

	//广播通知
	self.SendTableMsg(consts.MsgTypeGameUserVitaminLowTime, Message)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameUserVitaminLowTime, Message)
}

func (self *Common) OnTime() {
	//玩家计时器事件
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, _user := range *m {
			if _user == nil {
				continue
			}
			for _, v := range _user.Ctx.Timer.TimeArray {
				if v == nil {
					continue
				}
				v.TimeLeave--
				//fmt.Println(fmt.Sprintf("玩家(%d)当前计时（%d）", v.TimerID, v.TimeLeave))
				if v.TimeLeave <= 0 {
					//fmt.Println(fmt.Sprintf("玩家(%d)计时到了", v.Id))
					self.SendTimeMsg(v.TimerID, _user.Uid)
				}
			}
			_user.Ctx.Timer.Clear()
		}
	})
	//游戏计算时器事件
	for _, v := range self.GameTimer.TimeArray {
		if v == nil {
			continue
		}
		v.TimeLeave--
		//fmt.Println(fmt.Sprintf("玩家(%d)当前计时（%d）", v.TimerID, v.TimeLeave))
		if v.TimeLeave <= 0 {
			//fmt.Println(fmt.Sprintf("玩家(%d)计时到了", v.Id))
			self.SendTimeMsg(v.TimerID, v.Id)
		}
		self.GameTimer.Clear()
	}
	//检查是不是要自动申请解散
	self.LaunchDismissAuto()
	//检查是否离线时间超过5分钟，超过5分钟则自动解散房间
	self.CheckOfflineTime()
	//判断是否解散房间
	self.DissmissRoomAuto()
	//检查是否竞技点过低游戏暂停5分钟,超过5分钟则自动解散房间
	self.CheckVitaminLowPauseTime()
	//检查程序是否出错，程序出错后无法恢复房间，需要强制解散
	self.CheckErrorDismissAuto()
}

// ! 游戏退出
func (self *Common) OnExit(uid int64) {
	self.userMutex.Lock()
	defer self.userMutex.Unlock()

	self.PlayerInfoWrite(func(m *map[int64]*Player) {
		for k, v := range *m {
			if v.Uid == uid {
				//TODO 给客户端同步数据
				self.OnWriteGameRecord(v.Seat, fmt.Sprintf("当前该玩家离开桌子,gamecount:%d, wincount:%d", v.GameCount, v.WinCount))

				//同步任务数据
				if v.GameCount > 0 {
					server2.GetServer().SendTaskCompleteState(v.Uid, consts.TASK_KIND_GAME_ROUND, v.GameCount, self.GetTableInfo().KindId, consts.TASK_KIND_CARDTYPE_NONE)
				}
				if v.WinCount > 0 {
					server2.GetServer().SendTaskCompleteState(v.Uid, consts.TASK_KIND_GAME_WIN, v.WinCount, self.GetTableInfo().KindId, consts.TASK_KIND_CARDTYPE_NONE)
				}

				v.CleanTask()

				delete((*m), k)
				break
			}
		}
	})
	self.OnFewerHide(static.INVALID_CHAIR)
}

// ! 游戏退出
func (self *Common) OnExitLookon(uid int64) {
	self.lookonuserMutex.Lock()
	defer self.lookonuserMutex.Unlock()

	for k, v := range self.LookonPlayer {
		if v.Uid == uid {
			//TODO 给客户端同步数据
			self.OnWriteGameRecord(v.Seat, fmt.Sprintf("当前该旁观玩家【离开】桌子,gamecount:%d, wincount:%d", v.GameCount, v.WinCount))

			delete(self.LookonPlayer, k)
			break
		}
	}
	self.OnFewerHide(static.INVALID_CHAIR)
}

func (self *Common) FlashClient(userid int64, msg string) {
	//self.onUserInfo(userid, new(public.Msg_C_Info))
	_msg := "网络异常，重连中。。。"
	if msg != "" {
		_msg = msg
	}
	wChairSeat := self.GetChairByUid(userid)
	self.OnWriteGameRecord(wChairSeat, _msg)
	self.SendGameNotificationMessage(wChairSeat, _msg)
	self.DisconnectOnMisoperation(wChairSeat)
}

// 桌子有效牌局大结算事件
// 对于已经扣了房卡的桌子 总结算时 会走这里
func (self *Common) OnBalance(houseApi *HouseApi, playCount int /*对局次数*/, bigWinScore float64 /*大赢家分数*/, gameScore []float64 /*总成绩*/, floorPayInfo *models.HouseFloorGearPay) /*返回被扣了房费的玩家已经扣了多少*/ {
	self.OnWriteGameRecord(static.INVALID_CHAIR, "OnBalance")
	if houseApi != nil {
		// 支付房费
		self.PayRate(houseApi, bigWinScore, gameScore, floorPayInfo)
		//// 推送闲聊战绩
		//self.PushRecord(playCount, bigWinScore, gameScore)
	} else {

	}
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		if user := self.GetUserItemByChair(uint16(i)); user != nil {
			server2.GetServer().OnBettleDone(user.Uid, gameScore[i])
		}
	}
	return
}

func (self *Common) AAPay() {
	houseApi := self.GetHouseApi()
	floorPayInfo := models.NewHouseFloorGearPay(self.GetTableInfo().DHId, self.GetTableInfo().FId, self.GetProperPNum())
	if houseApi != nil {
		floorPayInfo = houseApi.GetFloorPayInfo()
	}
	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("楼层房费支付信息：%+v", floorPayInfo))
	if floorPayInfo == nil {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("楼层房费支付信息为空"))
		return
	}
	baseCost := floorPayInfo.BaseCost()
	if baseCost != models.InvalidPay {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局参考合伙人收益为：%d", baseCost))
		// 结算房费 得到本局真实的合伙人 可配置额度
		realRoyalty, usersCost, payType, isValidRound, err := self.CostVitaminRoomRateAA(houseApi, floorPayInfo)
		if err != nil {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("扣除房费错误：%v", err))
		} else {
			// 统计玩家给上级提供的收益
			if isValidRound {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局为有效局(合伙人收益:%d)", realRoyalty))
				err = houseApi.StatisticsPartnerProfit(realRoyalty, usersCost, payType, self)
				if err != nil {
					self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("统计合伙人收益错误:%v", err))
				}
			} else {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局为无效局(合伙人收益:%d)", realRoyalty))
			}
		}
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("楼层并未配置基础房费：%d", baseCost))
	}
}

func (self *Common) PayRate(houseApi *HouseApi, bigWinScore float64 /*大赢家分数*/, gameScore []float64 /*总成绩*/, floorPayInfo *models.HouseFloorGearPay) {
	return
	if FloorDeductCompatible {
		// 结算房费 得到本局真实的合伙人 可配置额度
		realRoyalty, usersCost, payType, isValidRound, err := self.CostVitaminRoomRate(houseApi, bigWinScore, gameScore, floorPayInfo)
		if err != nil {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("扣除房费错误：%v", err))
		} else {
			// 统计玩家给上级提供的收益
			if isValidRound {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局为有效局(合伙人收益:%d)", realRoyalty))
				err = houseApi.StatisticsPartnerProfit(realRoyalty, usersCost, payType, self)
				if err != nil {
					self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("统计合伙人收益错误:%v", err))
				}
			} else {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局为无效局(合伙人收益:%d)", realRoyalty))
			}
		}

	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("楼层房费支付信息：%+v", floorPayInfo))
		if floorPayInfo == nil {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("楼层房费支付信息为空"))
			return
		}
		baseCost := floorPayInfo.BaseCost()
		if baseCost != models.InvalidPay {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局参考合伙人收益为：%d", baseCost))
			// 结算房费 得到本局真实的合伙人 可配置额度
			realRoyalty, usersCost, payType, isValidRound, err := self.CostVitaminRoomRate(houseApi, bigWinScore, gameScore, floorPayInfo)
			if err != nil {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("扣除房费错误：%v", err))
			} else {
				// 统计玩家给上级提供的收益
				if isValidRound {
					self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局为有效局(合伙人收益:%d)", realRoyalty))
					err = houseApi.StatisticsPartnerProfit(realRoyalty, usersCost, payType, self)
					if err != nil {
						self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("统计合伙人收益错误:%v", err))
					}
				} else {
					self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("本局为无效局(合伙人收益:%d)", realRoyalty))
				}
			}
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("楼层并未配置基础房费：%d", baseCost))
		}
	}

}

// 无效的对局
func (self *Common) OnInvalid() {
	self.OnWriteGameRecord(static.INVALID_CHAIR, "OnInvalid")
	self.OnRestoreVitamin()
}

// 还原疲劳值
func (self *Common) OnRestoreVitamin() {
	houseApi := self.GetHouseApi()
	if houseApi == nil {
		return
	}
	fo, err := houseApi.GetFloorVitaminOption()
	if err != nil {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "Get House Floor Form Redis Error: "+err.Error())
		return
	}
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		user := self.GetUserItemByChair(uint16(i))
		if user == nil {
			continue
		}
		self.UpdateUserVitamin(fo, user, -static.SwitchF64ToVitamin(self.GetRealScore(user.Ctx.GangScore+user.Ctx.MagicScore)), models.GameCost)
	}
}

// 高低位设置
func MAKEWORD(a byte, b byte) uint16 {
	return uint16(a) | (uint16(b) << 8)
}

// ! 加入无效杠牌
func (self *Common) AddGangCardToInvalidArea(GangCard byte) {
	for _, CardData := range self.InvalidGangCards {
		if CardData == GangCard {
			return
		}
	}
	self.InvalidGangCards = append(self.InvalidGangCards, GangCard)
}

// ! 检查是否为无效杠牌
func (self *Common) CheckGangCardInvalid(GangCard byte) bool {
	for _, CardData := range self.InvalidGangCards {
		if CardData == GangCard {
			return false
		}
	}
	return true
}

// ! 麻将做牌测试，比如潜江有“少三张（发牌时少发3张牌，发10张牌的功能）”
func (self *Common) InitDebugCardsByFixHandCardNum(configName string, cbRepertoryCard *[]byte, wBankerUser *uint16, handCardNum int) (newlefcount byte, err error) {
	defer func() {
		if err != nil {
			self.OnWriteGameRecord(static.INVALID_CHAIR, err.Error())
		}
	}()
	//! 做牌文件配置
	var debugCardConfig *meta2.CardConfig = new(meta2.CardConfig)

	if ok, fileName := self.readDebugCards(configName, debugCardConfig); !ok {
		return 0, fmt.Errorf("根据房主id查找做牌文件失败！做牌文件不存在:%s", fileName)
	}

	// 是否开启做牌
	if debugCardConfig.IsAble == 1 {
		//检查做牌文件是否做牌异常
		for _, handCards := range debugCardConfig.UserCards {
			if len(strings.Split(handCards, ",")) != handCardNum {
				return 0, fmt.Errorf(fmt.Sprintf("做牌文件:手牌长度不为%d", handCardNum))
			}
		}
		//检查牌堆牌是否正常
		if len(debugCardConfig.RepertoryCard) != debugCardConfig.RepertoryCardCount*5-1 {
			//fmt.Println(fmt.Sprintf("做牌文件:牌库牌数量不一致:::RepertoryCard:[%d]>>>实际做牌牌库数量:[%d]", len(debugCardConfig.RepertoryCard), debugCardConfig.RepertoryCardCount*5-1))
			return 0, fmt.Errorf("做牌文件:牌库牌数量不一致:::RepertoryCardCount:[%d]>>>牌库数量:[%d]", debugCardConfig.RepertoryCardCount, (len(debugCardConfig.RepertoryCard)+1)/5)
		}
		// 设置玩家手牌
		for userIndex, handCards := range debugCardConfig.UserCards {
			_item := self.GetUserItemByChair(uint16(userIndex))
			if _item != nil {
				_item.Ctx.InitCardIndex() //清理手牌
				handCardsSlice := strings.Split(handCards, ",")
				for _, cardStr := range handCardsSlice {
					if cardValue := self.getCardDataByName(cardStr, 0); cardValue == static.INVALID_BYTE {
						return 0, fmt.Errorf("做牌文件:第%d号玩家做牌异常：%s", userIndex, cardStr)
					} else {
						_item.Ctx.DispatchCard(cardValue)
					}
				}
			}
		}
		//设置牌堆牌
		repertoryCards := strings.Split(debugCardConfig.RepertoryCard, ",")
		for cardIndex, cardStr := range repertoryCards {
			if cardValue := self.getCardDataByName(cardStr, 0); cardValue == static.INVALID_BYTE {
				return 0, fmt.Errorf("做牌文件:牌堆第%d个做牌异常：%s", cardIndex, cardStr)
			} else {
				(*cbRepertoryCard)[cardIndex] = byte(cardValue)
			}
		}
		// 设置庄家
		if debugCardConfig.BankerUserSeatId != -1 {
			(*wBankerUser) = uint16(debugCardConfig.BankerUserSeatId)
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("做牌文件:设置庄家,座位号:%d", debugCardConfig.BankerUserSeatId))
		}

		if debugCardConfig.Max > 0 {
			self.Rule.JuShu = debugCardConfig.Max
			base2.DebugOff = false
		}
	} else {
		return 0, fmt.Errorf("做牌文件:开关未开启")
	}
	return debugCardConfig.LeftCardCount, err
}

// ! 加载测试麻将数据
func (self *Common) InitDebugCards_ex(configName string, cbRepertoryCard *[]byte, wBankerUser *uint16) (newlefcount byte, err error) {
	defer func() {
		if err != nil {
			self.OnWriteGameRecord(static.INVALID_CHAIR, err.Error())
		}
	}()
	//! 做牌文件配置
	var debugCardConfig *meta2.CardConfig = new(meta2.CardConfig)

	if ok, fileName := self.readDebugCards(configName, debugCardConfig); !ok {
		return 0, fmt.Errorf("根据房主id查找做牌文件失败！做牌文件不存在:%s", fileName)
	}

	// 是否开启做牌
	if debugCardConfig.IsAble == 1 {
		//检查做牌文件是否做牌异常
		for _, handCards := range debugCardConfig.UserCards {
			if len(strings.Split(handCards, ",")) != 13 {
				return 0, fmt.Errorf("做牌文件:手牌长度不为13")
			}
		}
		//检查牌堆牌是否正常
		if len(debugCardConfig.RepertoryCard) != debugCardConfig.RepertoryCardCount*5-1 {
			//fmt.Println(fmt.Sprintf("做牌文件:牌库牌数量不一致:::RepertoryCard:[%d]>>>实际做牌牌库数量:[%d]", len(debugCardConfig.RepertoryCard), debugCardConfig.RepertoryCardCount*5-1))
			return 0, fmt.Errorf("做牌文件:牌库牌数量不一致:::RepertoryCardCount:[%d]>>>牌库数量:[%d]", debugCardConfig.RepertoryCardCount, (len(debugCardConfig.RepertoryCard)+1)/5)
		}
		// 设置玩家手牌
		for userIndex, handCards := range debugCardConfig.UserCards {
			_item := self.GetUserItemByChair(uint16(userIndex))
			if _item != nil {
				_item.Ctx.InitCardIndex() //清理手牌
				handCardsSlice := strings.Split(handCards, ",")
				for _, cardStr := range handCardsSlice {
					if cardValue := self.getCardDataByName(cardStr, 0); cardValue == static.INVALID_BYTE {
						return 0, fmt.Errorf("做牌文件:第%d号玩家做牌异常：%s", userIndex, cardStr)
					} else {
						_item.Ctx.DispatchCard(cardValue)
					}
				}
			}
		}
		//设置牌堆牌
		repertoryCards := strings.Split(debugCardConfig.RepertoryCard, ",")
		for cardIndex, cardStr := range repertoryCards {
			if cardValue := self.getCardDataByName(cardStr, 0); cardValue == static.INVALID_BYTE {
				return 0, fmt.Errorf("做牌文件:牌堆第%d个做牌异常：%s", cardIndex, cardStr)
			} else {
				if cardIndex < len(*cbRepertoryCard) {
					(*cbRepertoryCard)[cardIndex] = byte(cardValue)
				}
			}
		}
		// 设置庄家
		if debugCardConfig.BankerUserSeatId != -1 {
			(*wBankerUser) = uint16(debugCardConfig.BankerUserSeatId)
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("做牌文件:设置庄家,座位号:%d", debugCardConfig.BankerUserSeatId))
		}

		if debugCardConfig.Max > 0 {
			self.Rule.JuShu = debugCardConfig.Max
			base2.DebugOff = false
		}
	} else {
		return 0, fmt.Errorf("做牌文件:开关未开启")
	}
	return debugCardConfig.LeftCardCount, err
}
func (self *Common) InitDebugCards(configName string, cbRepertoryCard *[]byte, wBankerUser *uint16) (err error) {
	_, err = self.InitDebugCards_ex(configName, cbRepertoryCard, wBankerUser)
	return err
}

func (self *Common) readDebugCards(configName string, debugCardConfig *meta2.CardConfig) (bool, string) {
	fileName := fmt.Sprintf("./%s%d_%d", configName, self.GetPlayerCount(), self.GetTableInfo().Step-1)
	self.OnWriteGameRecord(static.INVALID_CHAIR, "开始根据小局读取做牌文件，文件名："+fileName)
	if !static.GetJsonMgr().ReadData("./json/check", fileName, debugCardConfig) {
		fileName = fmt.Sprintf("./%s%d_%d", configName, self.GetPlayerCount(), self.Rule.FangZhuID)
		self.OnWriteGameRecord(static.INVALID_CHAIR, "开始根据房主id读取做牌文件，文件名："+fileName)
		if !static.GetJsonMgr().ReadData("./json", fileName, debugCardConfig) {
			fileName = fmt.Sprintf("./%s%d", configName, self.GetPlayerCount())
			if !static.GetJsonMgr().ReadData("./json", fileName, debugCardConfig) {
				return false, fileName
			}
		}
	}

	return true, fileName
}

// ！游戏小结算
func (self *Common) TableWriteGameDate(playNum int, _user *Player, scoreKind int, winScore int) {
	if _user == nil {
		xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDate，_user为空,游戏类型（%d）包厢id（%d）牌桌号（%d）", self.GetTableInfo().KindId, self.GetTableInfo().HId, self.GetTableInfo().Id))
		return
	}

	if self.GetTableInfo().Config.GameType == static.GAME_TYPE_GOLD { // 金币场写战绩
		record := new(models.GameResultDetail)
		record.GameNum = self.GetTableInfo().GameNum
		record.UId = _user.Uid
		record.BeforeScore = _user.UserScoreInfo.Score
		record.AfterScore = _user.UserScoreInfo.Score + winScore
		if record.AfterScore < 0 {
			record.AfterScore = 0
		}
		record.Score = winScore + self.GetTableInfo().Config.CardCost // 注: 此处winscore是计算茶水费后的分数加减, 此处需还原原始得分
		record.Revenue = self.GetTableInfo().Config.CardCost
		record.Result = scoreKind
		record.KindId = self.GetTableInfo().KindId
		record.SiteType = self.GetTableInfo().SiteType
		record.GameId = server2.GetServer().Con.Id
		record.Ntid = self.GetTableInfo().NTId
		record.ClientIp = _user.Ip
		// 获取底分
		configMap := make(map[string]interface{})
		json.Unmarshal([]byte(self.GetTableInfo().Config.GameConfig), &configMap)
		record.Difen = int(configMap["difen"].(float64))
		// 写数据库记录
		if err := server2.GetDBMgr().GetDBmControl().Create(&record).Error; err != nil {
			xlog.Logger().Errorln(fmt.Sprintf("金币场写入小结算出错：（%v）", err))
		}
		// 更新用户个人战绩
		self.writeScore(_user, winScore, scoreKind, models.CostTypeGame)
		//排位赛正在进行中
		if self.IsMatching() {
			matchrecord := new(models.GameMatchDetail)
			matchrecord.GameNum = self.GetTableInfo().GameNum
			keyHeader := fmt.Sprintf("Match_%03d_%02d_", self.GetTableInfo().KindId, self.GetTableInfo().SiteType)
			matchrecord.UId = _user.Uid
			matchrecord.KindId = self.GetTableInfo().KindId
			matchrecord.SiteType = self.GetTableInfo().SiteType
			matchrecord.Ntid = self.GetTableInfo().NTId
			matchrecord.ClientIp = _user.Ip
			matchrecord.Score = winScore
			matchrecord.Result = scoreKind
			matchrecord.BeginDate = time.Unix(self.GetTableInfo().Config.MatchConfig.BeginDate, 0)
			matchrecord.BeginTime = time.Unix(self.GetTableInfo().Config.MatchConfig.BeginTime, 0)
			matchrecord.EndDate = time.Unix(self.GetTableInfo().Config.MatchConfig.EndDate, 0)
			matchrecord.EndTime = time.Unix(self.GetTableInfo().Config.MatchConfig.EndTime, 0)
			matchrecord.MatchKey = keyHeader + time.Now().Format("20060102") + matchrecord.BeginTime.Format("150405")
			// 写数据库记录
			if err := server2.GetDBMgr().GetDBmControl().Create(&matchrecord).Error; err != nil {
				xlog.Logger().Errorln(fmt.Sprintf("金币场写入排位赛数据出错：（%v）", err))
			}
			//排位赛统计
			matchtotal := new(models.GameMatchTotal)
			matchtotal.MatchKey = matchrecord.MatchKey
			matchtotal.UId = matchrecord.UId
			matchtotal.KindId = matchrecord.KindId
			matchtotal.CreatedAt = time.Now()
			matchtotal.SiteType = matchrecord.SiteType
			matchtotal.Score = matchrecord.Score
			matchtotal.TotalCount = 1 //新增一次参入场次
			if matchrecord.Result == 0 {
				matchtotal.WinCount = 1 //新增一次胜利场次
			} else if matchrecord.Result == static.ScoreKind_Flee || matchrecord.Result == static.ScoreKind_pass {
				matchtotal.TotalCount = 0 //逃跑不算场次
			}
			matchtotal.BeginDate = matchrecord.BeginDate
			matchtotal.BeginTime = matchrecord.BeginTime
			matchtotal.EndDate = matchrecord.EndDate
			matchtotal.EndTime = matchrecord.EndTime
			//! 插入数据库
			if id, err := server2.GetDBMgr().UpdataGameMatchTotal(matchtotal); err != nil {
				xlog.Logger().Errorln(fmt.Sprintf("id（%d）写入排位赛统计出错：%v", id, err))

			}
		}

	} else { // 好友场写战绩
		record := new(models.RecordGameRound)
		record.UId = _user.Uid
		record.UName = _user.Name
		record.Ip = _user.Ip
		record.UUrl = _user.ImgUrl
		record.UGenber = _user.Sex
		record.GameNum = self.GetTableInfo().GameNum
		record.RoomNum = self.GetTableInfo().Id
		record.PlayNum = playNum
		record.ServerId = server2.GetServer().Con.Id
		record.SeatId = int(_user.Seat)
		record.ReplayId = self.RoundReplayId
		record.ScoreKind = scoreKind
		record.WinScore = winScore
		record.BeginDate = self.GameBeginTime
		record.CreatedAt = time.Now()
		record.EndDate = time.Now()
		record.Radix = self.Rule.Radix

		//写入小结算
		server2.GetDBMgr().InsertGameRecord(record)

		// 更新用户个人战绩
		self.writeScore(_user, winScore, scoreKind, models.CostTypeGame)

		//单局结算写入总结算
		self.TableWriteGameDateTotalEveryRound(_user.Uid, playNum, int(_user.Seat), scoreKind, winScore)
	}

	self.updateTaskInfo(_user, winScore, scoreKind, models.CostTypeGame)
}

func (self *Common) IsMatching() bool {

	if self.GetTableInfo().Config.MatchConfig.State == 1 && self.GetTableInfo().Config.MatchConfig.Flag == 1 {
		return true
	}
	return false
}

func (self *Common) writeScore(_user *Player, winScore int, scoreKind int, costType int8) {
	if costType == models.CostTypeGame {
		var scoreInfo rule2.TagScoreInfo
		scoreInfo.Score = winScore
		scoreInfo.ScoreKind = scoreKind
		_user.WriteScore(&scoreInfo, self.GetNowSystemTimerSecond())
	}

	if self.GetTableInfo().Config.GameType != static.GAME_TYPE_FRIEND {
		// 分数结算
		afterScore, err := self.GetTable().CostUserWealth(_user.Uid, consts.WealthTypeGold, winScore, costType, self.GetProperPNum())
		if err == nil {
			// 更新分数
			_user.UserScoreInfo.GameGold = afterScore

			self.getServiceFrame().SendUpdateScore(_user.Seat)
		}
	} else {
		if costType == models.CostTypeSticker {
			var offset [meta2.MAX_PLAYER]int
			offset[_user.GetChairID()] = winScore
			self.OnSettle(offset, consts.EventChat)
		}
	}
}

func (self *Common) updateTaskInfo(_user *Player, winScore int, scoreKind int, costType int8) {
	//修改属性
	if models.CostTypeGame == costType {
		var scoreInfo rule2.TagScoreInfo
		scoreInfo.Score = winScore
		scoreInfo.ScoreKind = scoreKind
		_user.UpdateTaskInfo(&scoreInfo, self.GetNowSystemTimerSecond())
	}
}

func (self *Common) SendUpdateScore(wChairID uint16) {
	var _userScore static.Msg_S_UserScore
	for i := 0; i < self.GetTableInfo().Config.MaxPlayerNum; i++ {
		_userScore.Score = append(_userScore.Score, 0)
		if _item := self.GetUserItemByChair(uint16(i)); _item != nil {
			_userScore.Score[i] = _item.UserScoreInfo.Score
		}

	}
	self.SendTableMsg(consts.MsgTypeUserScoreUpdate, _userScore)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeUserScoreUpdate, _userScore)
}

// ! 游戏大结算小结算写入
func (self *Common) TableWriteGameDateTotalEveryRound(uid int64, playCount int, seatId int, scoreKind int, winScore int) {
	recordTotal, err := server2.GetDBMgr().SelectGameRecordTotal(self.GetTableInfo().GameNum, uid)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			xlog.Logger().Errorln(fmt.Sprintf("查询总战绩失败，gamenum = %s, seatid = %d", self.GetTableInfo().GameNum, seatId))
			return
		}

		recordTotal := new(models.RecordGameTotal)
		recordTotal.KindId = self.GetTableInfo().KindId
		recordTotal.GameNum = self.GetTableInfo().GameNum
		recordTotal.RoomNum = self.GetTableInfo().Id
		recordTotal.PlayCount = playCount
		recordTotal.Round = self.GetTableInfo().Config.RoundNum
		recordTotal.ServerId = server2.GetServer().Con.Id
		recordTotal.SeatId = seatId
		recordTotal.ScoreKind = static.ScoreKind_pass
		recordTotal.WinScore = winScore
		recordTotal.HId = self.GetTableInfo().DHId
		recordTotal.IsHeart = 0
		recordTotal.FId = int(self.GetTableInfo().FId)
		recordTotal.DFId = self.GetTableInfo().NFId
		recordTotal.Radix = self.Rule.Radix
		recordTotal.IsValidRound = false
		recordTotal.IsBigWinner = false

		GamePerson := self.GetPlayerByChair(uint16(seatId))

		if GamePerson == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，第（%d）局，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）", playCount, seatId, self.GetTableInfo().KindId, self.GetTableInfo().HId, self.GetTableInfo().Id))
			return
		}
		recordTotal.Uid = GamePerson.Uid
		recordTotal.UName = GamePerson.Name
		recordTotal.Ip = GamePerson.Ip
		recordTotal.HalfWayDismiss = false
		recordTotal.CreatedAt = time.Now()

		//! 插入数据库
		server2.GetDBMgr().InsertGameRecordTotal(recordTotal)

		return
	}

	recordTotal.WinScore += winScore
	recordTotal.ScoreKind = static.ScoreKind_pass

	recordTotal.PlayCount = playCount
	recordTotal.CreatedAt = time.Now()

	if scoreKind == static.ScoreKind_pass {
		recordTotal.HalfWayDismiss = true
	} else {
		recordTotal.HalfWayDismiss = false
	}

	// recordTotal.IsValidRound = false
	// recordTotal.IsBigWinner = false

	updateMap := make(map[string]interface{})
	updateMap["win_score"] = recordTotal.WinScore
	updateMap["score_kind"] = recordTotal.ScoreKind
	updateMap["play_count"] = recordTotal.PlayCount
	updateMap["created_at"] = recordTotal.CreatedAt
	updateMap["halfwaydismiss"] = recordTotal.HalfWayDismiss
	// updateMap["is_big_winner"] = recordTotal.IsBigWinner
	// updateMap["is_valid_round"] = recordTotal.IsValidRound

	//mysql
	server2.GetDBMgr().UpdataGameRecordTotal(recordTotal.Id, updateMap)
}

// ! 游戏大结算
func (self *Common) TableWriteGameDateTotal(playCount int, seatId int, scoreKind int, winScore int, isBigWin int) {
	finishTime := time.Now().Unix()

	GamePerson := self.GetPlayerByChair(uint16(seatId))

	if GamePerson == nil {
		xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，第（%d）局，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）", playCount, seatId, self.GetTableInfo().KindId, self.GetTableInfo().HId, self.GetTableInfo().Id))
		return
	}

	// 结算通知
	if self.IsHouse() {
		// 更新玩家包厢上局对战记录
		recordPlayer := new(static.HTableRecordPlayers)
		recordPlayer.HId = self.GetTableInfo().HId
		recordPlayer.FId = self.GetTableInfo().FId
		recordPlayer.TId = self.GetTableId()
		recordPlayer.KId = self.KIND_ID
		recordPlayer.Users = self.GetOtherUids(GamePerson.Uid)
		if err := server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(GamePerson.Uid, recordPlayer); err != nil {
			xlog.Logger().Errorf("更新玩家(%d)的包厢桌子最新的对战玩家记录失败:%v", GamePerson.Uid, err)
		}
		// ntf := new(public.GH_TableRes_Ntf)
		// ntf.HId = self.GetTableInfo().HId
		// ntf.UId = GamePerson.Uid
		// ntf.TId = self.GetTableInfo().Id
		// ntf.KindId = self.GetTableInfo().KindId
		// ntf.GameId = self.GetTableInfo().GameId
		// ntf.WinScore = winScore
		// ntf.IsBigWin = isBigWin
		// _ = protocolworkers.CallHall(constant.MsgTypeTableRes_Ntf, xerrors.SuccessCode, ntf, GamePerson.Uid)
		// _ = protocolworkers.CallHall(constant.MsgTypeTableRes_Ntf, xerrors.SuccessCode, ntf, GamePerson.Uid)

	} else {
		// 如果不是包厢玩法，则删掉包厢上局对战记录，以免影响再来一局功能
		_ = server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(GamePerson.Uid, nil)
	}

	recordTotal, err := server2.GetDBMgr().SelectGameRecordTotal(self.GetTableInfo().GameNum, GamePerson.Uid)
	if err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("查询总战绩失败，gamenum = %s, seatid = %d", self.GetTableInfo().GameNum, seatId))
		return
	}

	recordTotal.WinScore = winScore
	recordTotal.ScoreKind = scoreKind
	recordTotal.PlayCount = playCount
	recordTotal.CreatedAt = time.Now()

	var partnerId, superiorId int64
	if houseApi := self.GetHouseApi(); houseApi != nil {
		partnerId, superiorId = houseApi.GetHouseMemberPartnerAndSuperiorId(GamePerson.Uid)
	} else {
		partnerId = 0
		superiorId = 0
	}
	recordTotal.Partner = partnerId
	recordTotal.SuperiorId = superiorId

	updateMap := make(map[string]interface{})
	updateMap["win_score"] = recordTotal.WinScore
	updateMap["score_kind"] = recordTotal.ScoreKind
	updateMap["play_count"] = recordTotal.PlayCount
	updateMap["created_at"] = recordTotal.CreatedAt
	updateMap["partner"] = recordTotal.Partner
	updateMap["superiorid"] = recordTotal.SuperiorId

	//mysql
	server2.GetDBMgr().UpdataGameRecordTotal(recordTotal.Id, updateMap)

	//mysql
	if scoreKind != 6 {
		server2.GetDBMgr().UpdataUserRateOfWinning(GamePerson.Uid, recordTotal.WinScore)
	}

	//第一局解散,不判断包厢活动
	if playCount < 1 || (playCount == 1 && scoreKind == 6) {
		if seatId == 0 {
			self.OnInvalid()
			//未打完解散房间,扣0房卡,统计一次扣卡
			self.TableDeleteFangKa(0)
		}
		return
	}

	// 活动数据
	dhid := self.GetTableInfo().DHId
	if dhid != 0 {
		hact, err := server2.GetDBMgr().GetDBrControl().HouseActivityList(dhid, true)
		if err != nil {
			xlog.Logger().Errorln(err)
			return
		}
		for _, act := range hact {
			// 楼层活动
			InActFloor := false
			fidstrs := strings.Split(act.FId, ",")
			for i := 0; i < len(fidstrs); i++ {
				if self.GetTableInfo().FId == static.HF_Atoi64(fidstrs[i]) {
					InActFloor = true
					break
				}
			}
			if !InActFloor {
				continue
			}
			// 活动时间
			if finishTime < act.BegTime || finishTime > act.EndTime {
				continue
			}
			// 局数累计 活动
			if act.Kind == consts.HFACT_ROUNDS {
				err := server2.GetDBMgr().GetDBrControl().HouseActivityRecordInsert(act.Id, self.GetTableInfo().Users[seatId].Uid, 1)
				if err != nil {
					xlog.Logger().Errorln(err)
				}
				if act.Type == 1 { //抽奖活动
					err := server2.AddUserTicket(dhid, self.GetTableInfo().Users[seatId].Uid, act.Id, 1)
					if err != nil {
						xlog.Logger().Errorln(err)
						continue
					}
				}
			}
			// 活动 大赢家统计
			if act.Kind == consts.HFACT_BW {
				if isBigWin > 0 {
					err := server2.GetDBMgr().GetDBrControl().HouseActivityRecordInsert(act.Id, self.GetTableInfo().Users[seatId].Uid, 1)
					if err != nil {
						xlog.Logger().Errorln(err)
						continue
					}
				}
			}
			// 活动 积分统计
			if act.Kind == consts.HFACT_SCORE {
				err := server2.GetDBMgr().GetDBrControl().HouseActivityRecordInsert(act.Id, self.GetTableInfo().Users[seatId].Uid, self.GetRealScore(winScore))
				if err != nil {
					xlog.Logger().Errorln(err)
					continue
				}
			}
		}

	}

}
func (self *Common) Updatetrustpunish(gamenum string, CompleteCount int, winScore [4]int) {
	if gamenum == "" {
		return
	}
	for i := 0; i < self.GetPlayerCount(); i++ {
		GamePerson := self.GetPlayerByChair(uint16(i))

		if GamePerson == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）", i, self.GetTableInfo().KindId, self.GetTableInfo().HId, self.GetTableInfo().Id))
			continue
		}
		// 结算通知
		if self.IsHouse() {
			// 更新玩家包厢上局对战记录
			recordPlayer := new(static.HTableRecordPlayers)
			recordPlayer.HId = self.GetTableInfo().HId
			recordPlayer.FId = self.GetTableInfo().FId
			recordPlayer.TId = self.GetTableId()
			recordPlayer.KId = self.KIND_ID
			recordPlayer.Users = self.GetOtherUids(GamePerson.Uid)
			if err := server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(GamePerson.Uid, recordPlayer); err != nil {
				xlog.Logger().Errorf("更新玩家(%d)的包厢桌子最新的对战玩家记录失败:%v", GamePerson.Uid, err)
			}
		} else {
			// 如果不是包厢玩法，则删掉包厢上局对战记录，以免影响再来一局功能
			_ = server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(GamePerson.Uid, nil)
		}
	}
	var records []models.RecordGameRound
	err := server2.GetDBMgr().GetDBmControl().Model(models.RecordGameRound{}).Where("gamenum = ?", gamenum).Find(&records).Error
	if err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("查询战绩失败, game_num = %s", gamenum))
		return
	}
	updateMap := make(map[string]interface{})
	for _, record := range records {
		if record.PlayNum != CompleteCount {
			continue
		}
		updateMap["winscore"] = record.WinScore + winScore[record.SeatId]
		err = server2.GetDBMgr().GetDBmControl().Model(models.RecordGameRound{}).Where("id = ?", record.Id).Update(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(fmt.Sprintf("更新战绩失败, gamenum = %s, ScoreKind = %d", gamenum, record.ScoreKind))
		}
	}

}
func (self *Common) UpdateErrGameTotal(gamenum string) {
	if gamenum == "" {
		return
	}

	for i := 0; i < self.GetPlayerCount(); i++ {
		GamePerson := self.GetPlayerByChair(uint16(i))

		if GamePerson == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）", i, self.GetTableInfo().KindId, self.GetTableInfo().HId, self.GetTableInfo().Id))
			continue
		}
		// 结算通知
		if self.IsHouse() {
			// 更新玩家包厢上局对战记录
			recordPlayer := new(static.HTableRecordPlayers)
			recordPlayer.HId = self.GetTableInfo().HId
			recordPlayer.FId = self.GetTableInfo().FId
			recordPlayer.TId = self.GetTableId()
			recordPlayer.KId = self.KIND_ID
			recordPlayer.Users = self.GetOtherUids(GamePerson.Uid)
			if err := server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(GamePerson.Uid, recordPlayer); err != nil {
				xlog.Logger().Errorf("更新玩家(%d)的包厢桌子最新的对战玩家记录失败:%v", GamePerson.Uid, err)
			}
		} else {
			// 如果不是包厢玩法，则删掉包厢上局对战记录，以免影响再来一局功能
			_ = server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(GamePerson.Uid, nil)
		}
	}

	var records []models.RecordGameTotal
	err := server2.GetDBMgr().GetDBmControl().Model(models.RecordGameTotal{}).Where("game_num = ?", gamenum).Find(&records).Error
	if err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("查询战绩失败, game_num = %s", gamenum))
		return
	}
	for _, record := range records {
		if record.PlayCount < 1 {
			return
		}
		if record.WinScore > 0 {
			record.ScoreKind = 0
		} else {
			record.ScoreKind = 1
		}

		updateMap := make(map[string]interface{})
		updateMap["score_kind"] = record.ScoreKind
		err = server2.GetDBMgr().GetDBmControl().Model(models.RecordGameTotal{}).Where("id = ?", record.Id).Update(updateMap).Error
		if err != nil {
			xlog.Logger().Errorln(fmt.Sprintf("更新战绩失败, gamenum = %s, ScoreKind = %d", gamenum, record.ScoreKind))
			continue
		}
	}
}

// 总结算信息推送至闲聊
func (self *Common) PushRecord(playCount int, bigWinScore float64, gameScore []float64) {
	if playCount < 1 {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "一局都未打完，不推送闲聊战绩")
		return
	}
	if !self.GetTableInfo().IsTeaHouse() {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "不是包厢牌桌，不推送闲聊战绩")
		return
	}
	self.OnWriteGameRecord(static.INVALID_CHAIR, "开始推送闲聊战报。")
	var msg static.MsgXianTalkRecord
	msg.HouseId = self.GetTableInfo().HId
	msg.Room = static.HF_Itoa(self.GetTableId())
	msg.GameNumber = fmt.Sprintf("%d/%d", playCount, self.Rule.JuShu)
	msg.Ante = static.HF_F64toa(float64(self.Rule.DiFen) / float64(self.Rule.Radix))
	msg.Time = fmt.Sprintf("%s-%s",
		time.Unix(self.TimeStart, 0).Format("01-02 15:04"),
		time.Now().Format("15:04"),
	)
	msg.User = make([]static.MsgXianTalkUser, 0)
	for i := 0; i < self.GetPlayerCount() && i < meta2.MAX_PLAYER; i++ {
		userItem := self.GetUserItemByChair(uint16(i))
		if userItem == nil {
			continue
		}
		var userMsg static.MsgXianTalkUser
		userMsg.UserId = userItem.Uid
		userMsg.UserName = userItem.Name
		userMsg.ImgUrl = userItem.ImgUrl
		userMsg.Number = gameScore[i]
		userMsg.Win = bigWinScore == userMsg.Number
		// 从redis取二维码信息
		if p := server2.GetPersonMgr().GetPerson(userItem.Uid); p != nil {
			userMsg.Qrcode = p.Info.DeliveryImg
		} else {
			self.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("获取收款二维码信息时玩家不存在。"))
		}
		msg.User = append(msg.User, userMsg)
	}
	// syslog.Logger().Warningln("闲聊数据：", public.HF_JtoA(&msg))
	if err := service2.PushRecordToXianTalk(server2.GetServer().ConServers.XianliaoHost, &msg); err != nil {
		maxLen := 20
		errStr := err.Error()
		if len(errStr) > maxLen {
			errStr = errStr[:maxLen]
		}
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("推送闲聊战报失败:%s", errStr))
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "推送闲聊战绩成功。")
	}

}

// 记录单个用户好友房历史战绩(最近24小时的8条记录)
func (self *Common) TableWriteHistoryRecord(balanceGame *static.Msg_S_BALANCE_GAME) {
	self.TableWriteHistoryRecordWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
}

// 记录单个用户好友房历史战绩(最近24小时的8条记录)
func (self *Common) TableWriteHistoryRecordWith(playCount int, gameScore []int) {

	var realScore [8]float64
	for i := 0; i < len(gameScore) && i < len(realScore); i++ {
		realScore[i] = self.GetRealScore(gameScore[i])
	}

	//! 计算大赢家的座位号
	bigWinScore := float64(0)
	for i := 0; i < self.GetPlayerCount(); i++ {
		if realScore[i] >= bigWinScore {
			bigWinScore = realScore[i]
		}
	}

	IsValidRound := false
	IsBigValidRound := false

	houseApi := self.GetHouseApi()
	payInfo := models.NewHouseFloorGearPay(self.GetTableInfo().DHId, self.GetTableInfo().FId, self.GetProperPNum())

	if FloorDeductCompatible {
		ValidScore, ValidBigScore, _ := server2.GetDBMgr().SelectHouseValidRound(self.GetTableInfo().DHId, self.GetTableInfo().FId)
		if bigWinScore >= float64(ValidScore) {
			IsValidRound = true

			if bigWinScore >= float64(ValidBigScore) && ValidBigScore > ValidScore {
				IsBigValidRound = true
			}
		}
	} else {
		if houseApi != nil {
			payInfo = houseApi.GetFloorPayInfo()
			IsValidRound = payInfo.IsValidRound(static.SwitchF64ToVitamin(bigWinScore))
		}
	}

	// 调试日志
	self.OnWriteGameRecord(static.INVALID_CHAIR, "TableWriteHistoryRecordWith....")
	for i := 0; i < self.GetPlayerCount(); i++ {
		if i >= len(gameScore) {
			xlog.Logger().Errorf("TableWriteGameDayRecord error: i = %d, GetPlayerCount() = %d, len(gameScore) = %d", i, self.GetPlayerCount(), len(gameScore))
			continue
		}
		if IsValidRound && bigWinScore > 0 && bigWinScore <= realScore[i] {
			self.OnWriteGameRecord(uint16(i), "大赢家")
			self.TableWriteGameDayRecord(i, 1, 1, gameScore[i], IsValidRound, IsBigValidRound)
		} else {
			self.TableWriteGameDayRecord(i, 1, 0, gameScore[i], IsValidRound, IsBigValidRound)
		}
	}

	// 跳转至总结算事件
	if playCount > 0 {
		self.OnBalance(houseApi, playCount, bigWinScore, realScore[:], payInfo)
	} else {
		self.OnInvalid()
	}

	// record := new(public.GameRecordHistory)
	// record.KindId = self.GameTable.KindId
	// record.GameNum = self.GameTable.GameNum
	// record.RoomNum = self.GameTable.Id
	// record.PlayedAt = time.Now().Unix()
	// record.HId = self.GameTable.DHId
	// record.FId = self.GameTable.NFId
	// record.PlayCount = playCount
	// record.Round = self.GetTable().Config.RoundNum
	// record.IsHeart = 0
	//
	// record.Player = make([]*public.GameRecordHistoryPlayer, 0)
	// for i := 0; i < self.GetPlayerCount(); i++ {
	// 	p := self.GameTable.getPersonWithSeatId(i)
	// 	if p == nil {
	// 		syslog.Logger().Errorln(fmt.Sprintf("TableWriteHistoryRecord获取（%d）玩家信息为空,游戏类型（%d）包厢id（%d）牌桌号（%d）", i, self.GameTable.KindId, self.GameTable.HId, self.GameTable.Id))
	// 		continue
	// 	}
	// 	person := new(public.GameRecordHistoryPlayer)
	// 	person.Uid = p.Uid
	// 	person.Nickname = p.Name
	// 	person.Score = gameScore[i]
	// 	record.Player = append(record.Player, person)
	// }
	//
	// if self.IsHouse() {
	// 	//记录一条包厢战绩(包厢战绩)
	// 	GetDBMgr().db_R.InsertHouseHallRecord(record.HId, record)
	// 	for _, p := range record.Player {
	// 		// 每人记录一条(包厢我的战绩)
	// 		GetDBMgr().db_R.InsertHouseHallMyRecord(record.HId, p.Uid, record)
	// 	}
	// } else {
	// 	for _, p := range record.Player {
	// 		// 每人记录一条
	// 		GetDBMgr().db_R.InsertNormalHallRecord(p.Uid, record)
	// 	}
	// }
}

// 记录战绩详情数据
func (self *Common) TableWriteHistoryRecordDetail(balanceGame *static.Msg_S_BALANCE_GAME) {
	//self.TableWriteHistoryRecordDetailWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
}

// 记录战绩详情数据
func (self *Common) TableWriteHistoryRecordDetailWith(playCount int, gameScore []int) {
	//// 记录对局详情
	//roundRecord := new(public.Msg_S2C_GameRecordInfo)
	//roundRecord.GameNum = self.GameTable.GameNum
	//roundRecord.KindId = self.GameTable.KindId
	//roundRecord.RoomId = self.GameTable.Id
	//roundRecord.Time = time.Now().Unix()
	//// 玩家列表
	//for i := 0; i < self.GetPlayerCount(); i++ {
	//	p := self.GameTable.getPersonWithSeatId(i)
	//	if p == nil {
	//		syslog.Logger().Errorln(fmt.Sprintf("TableWriteHistoryRecordDetail,获取（%d）玩家信息为空,游戏类型（%d）包厢id（%d）牌桌号（%d）", i, self.GameTable.KindId, self.GameTable.HId, self.GameTable.Id))
	//		continue
	//	}
	//	roundRecord.UserArr = append(roundRecord.UserArr, &public.Msg_S2C_GameRecordInfoUser{
	//		Uid:      p.Uid,
	//		Nickname: p.Name,
	//		Imgurl:   p.ImgUrl,
	//		Sex:      p.Sex,
	//		Score:    gameScore[i],
	//	})
	//}
	//// 每局积分
	//scoreArr := make([][]int, 0)
	//
	//// 初始化二维数组
	//for i := 0; i < int(playCount); i++ {
	//	arr := make([]int, self.GetPlayerCount())
	//	scoreArr = append(scoreArr, arr)
	//}
	//
	//replayIdArr := make([]int64, 0)
	//endTimeArr := make([]int64, 0)
	//startTimeArr := make([]int64, 0)
	//for i := 0; i < self.GetPlayerCount(); i++ {
	//	p := self.GameTable.getPersonWithSeatId(i)
	//	if p == nil {
	//		syslog.Logger().Errorln(fmt.Sprintf("TableWriteHistoryRecordDetail，获取（%d）玩家信息为空,游戏类型（%d）包厢id（%d）牌桌号（%d）", i, self.GameTable.KindId, self.GameTable.HId, self.GameTable.Id))
	//		continue
	//	}
	//	list, err := GetDBMgr().db_R.SelectGameRecord(self.GameTable.GameNum, p.Uid)
	//	if err != nil {
	//		syslog.Logger().Debug("get game record from redis failed: ", err.Error())
	//		return
	//	}
	//	for j, item := range list {
	//		// 测试时屏蔽, 为避免异常情况, 上线时应打开
	//		if j >= int(playCount) {
	//			continue
	//		}
	//		scoreArr[j][i] = item.WinScore
	//		if i == 0 {
	//			replayIdArr = append(replayIdArr, item.ReplayId)
	//			endTimeArr = append(endTimeArr, item.WriteDate)
	//			startTimeArr = append(startTimeArr, item.BeginDate)
	//		}
	//	}
	//
	//	//使用玩删除Redis数据，不需要长期存放到Redis
	//	GetDBMgr().db_R.DeleteGameRecord(self.GameTable.GameNum, p.Uid)
	//}
	//
	//for i := 0; i < int(playCount); i++ {
	//	if i >= len(replayIdArr) {
	//		break
	//	}
	//	roundRecord.ScoreArr = append(roundRecord.ScoreArr, &public.Msg_S2C_GameRecordInfoScore{
	//		ReplayId:  replayIdArr[i],
	//		StartTime: startTimeArr[i],
	//		EndTime:   endTimeArr[i],
	//		Score:     scoreArr[i],
	//	})
	//}
	//
	////! 插入redis
	//GetDBMgr().db_R.InsertGameRecordInfo(roundRecord)
}

// ! 游戏出牌记录
func (self *Common) TableWriteOutDate(playNum int, replayRecord meta2.Replay_Record) {
	recordReplay := new(models.RecordGameReplay)
	recordReplay.GameNum = self.GetTableInfo().GameNum
	recordReplay.RoomNum = self.GetTableInfo().Id
	recordReplay.PlayNum = playNum
	recordReplay.ServerId = server2.GetServer().Con.Id
	recordReplay.HandCard = self.getWriteHandReplayRecordCString(replayRecord)
	recordReplay.OutCard = self.getWriteOutReplayRecordCString(replayRecord)
	recordReplay.KindID = self.GetTableInfo().KindId
	recordReplay.CardsNum = int(replayRecord.LeftCardCount)
	recordReplay.UVitaminMap = replayRecord.UVitamin
	recordReplay.CreatedAt = time.Now()
	if replayRecord.EndInfo != nil {
		recordReplay.EndInfo = static.HF_JtoA(replayRecord.EndInfo)
	}

	server2.GetDBMgr().InsertGameRecordReplay(recordReplay)

	self.RoundReplayId = recordReplay.Id
}

func (self *Common) TableUpdateOutDate(replayRecord meta2.Replay_Record) {
	recordReplay := new(models.RecordGameReplay)
	recordReplay.Id = self.RoundReplayId
	recordReplay.OutCard = self.getWriteOutReplayRecordCString(replayRecord)

	server2.GetDBMgr().UpdataGameRecordReplay(recordReplay)
}

func (self *Common) getWriteHandReplayRecordCString(replayRecord meta2.Replay_Record) string {
	handCardStr := ""
	for i := 0; i < self.GetPlayerCount(); i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < static.MAX_COUNT; j++ {
			handCardStr += fmt.Sprintf("%02x,", replayRecord.RecordHandCard[i][j])
		}
	}

	//写入分数
	handCardStr += "S:"

	for i := 0; i < self.GetPlayerCount(); i++ {
		handCardStr += fmt.Sprintf("%d,", replayRecord.Score[i])
	}

	return handCardStr
}

func (self *Common) getWriteOutReplayBigHuCString(huKind byte, ProvideUser uint16) string {
	var bighuStr string
	switch huKind {
	case 11:
		bighuStr = fmt.Sprintf("JYS%d,", ProvideUser)
		break
	case 4:
		bighuStr = fmt.Sprintf("QGH%d,", ProvideUser)
		break
	case 13:
		bighuStr = fmt.Sprintf("QSB%d,", ProvideUser)
		break
	case 5:
		bighuStr = fmt.Sprintf("HDL%d,", ProvideUser)
		break
	case 6:
		bighuStr = fmt.Sprintf("FYS%d,", ProvideUser)
		break
	case 7:
		bighuStr = fmt.Sprintf("GSK%d,", ProvideUser)
		break
	case 8:
		bighuStr = fmt.Sprintf("QQR%d,", ProvideUser)
		break
	case 9:
		bighuStr = fmt.Sprintf("PPH%d,", ProvideUser)
		break
	case 10:
		bighuStr = fmt.Sprintf("QYS%d,", ProvideUser)
		break
	case 12:
		bighuStr = fmt.Sprintf("MQQ%d,", ProvideUser)
		break
	case 14:
		bighuStr = fmt.Sprintf("QID%d,", ProvideUser)
		break
	case static.GameNoMagicHu:
		bighuStr = fmt.Sprintf("YHM%d,", ProvideUser)
		break
	case static.GameMagicHu:
		bighuStr = fmt.Sprintf("RHM%d,", ProvideUser)
		break
	case static.GameYgk:
		bighuStr += fmt.Sprintf("YGK%d,", ProvideUser)
		break
	case static.GameRgk:
		bighuStr += fmt.Sprintf("RGK%d,", ProvideUser)
		break
	default:
		bighuStr = fmt.Sprintf("NIL%d,", ProvideUser)
		break
	}
	return bighuStr
}

func (self *Common) getWriteOutReplayRecordCString(replayRecord meta2.Replay_Record) string {
	ourCardStr := ""
	ourCardStr += fmt.Sprintf("P:%02x,", replayRecord.PiziCard)
	if replayRecord.Fengquan > 0 {
		ourCardStr += fmt.Sprintf("f:%02d,", replayRecord.Fengquan)
	}
	// 把胡牌的U拿出来
	var hasHu bool
	endMsgUpdateScore := [meta2.MAX_PLAYER]float64{}
	for i := 0; i < len(replayRecord.VecOrder); i++ {
		recordI := replayRecord.VecOrder[i]
		if recordI.Operation == info2.E_Hu {
			for j, count := i, 0; j < len(replayRecord.VecOrder); j++ {
				recordJ := replayRecord.VecOrder[j]
				if recordJ.Operation == info2.E_GameScore {
					count++
					if count > self.GetPlayerCount() {
						break
					}
					recordJ.Operation = -1 // 置为无效
					endMsgUpdateScore[recordJ.Chair_id] = recordJ.UserScore
					replayRecord.VecOrder[j] = recordJ
				}
			}
			hasHu = true
			break
		}
	}

	for _, record := range replayRecord.VecOrder {
		if record.Operation < 0 {
			continue
		}
		if len(record.Value) == 0 && record.Operation != info2.E_GameScore {
			xlog.Logger().Errorf("(Common)记录数据牌对象异常（空牌值）:玩家ID(%d)发动操作（%d）\n", record.Chair_id, record.Operation)
			continue
		}
		ourCardStr += fmt.Sprintf("%d:", record.Chair_id)
		switch record.Operation {
		case info2.E_SendCard:
			ourCardStr += fmt.Sprintf("S%02x,", record.Value[0])
			break
		case info2.E_OutCard:
			ourCardStr += fmt.Sprintf("O%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG:
			ourCardStr += fmt.Sprintf("o%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG_Magic:
			ourCardStr += fmt.Sprintf("m%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG_PiZi:
			ourCardStr += fmt.Sprintf("z%02x,", record.Value[0])
			break
		case info2.E_Wik_Left:
			ourCardStr += fmt.Sprintf("L%02x,", record.Value[0])
			break
		case info2.E_Wik_Center:
			ourCardStr += fmt.Sprintf("C%02x,", record.Value[0])
			break
		case info2.E_Wik_Right:
			ourCardStr += fmt.Sprintf("R%02x,", record.Value[0])
			break
		case info2.E_Peng:
			ourCardStr += fmt.Sprintf("P%02x,", record.Value[0])
			break
		case info2.E_Gang:
			ourCardStr += fmt.Sprintf("G%02x,", record.Value[0])
			break
		case info2.E_Gang_HongZhongGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_FaCaiGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_PiziGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_LaiziGand:
			ourCardStr += fmt.Sprintf("M%02x,", record.Value[0])
			break
		case info2.E_Gang_XuGand:
			ourCardStr += fmt.Sprintf("X%02x,", record.Value[0])
			break
		case info2.E_Gang_AnGang:
			ourCardStr += fmt.Sprintf("A%02x,", record.Value[0])
			break
		case info2.E_Qiang:
			ourCardStr += fmt.Sprintf("Q%02x,", record.Value[0])
			break
		case info2.E_Hu:
			ourCardStr += fmt.Sprintf("H%02x,", record.Value[0])
			break
		case info2.E_HardHu: //硬胡还是软胡（回放算分可能会用到）
			ourCardStr += fmt.Sprintf("d%02x,", record.Value[0])
			break
		case info2.E_HuangZhuang:
			ourCardStr += fmt.Sprintf("N%02x,", record.Value[0])
			break
		case info2.E_Bird:
			ourCardStr += fmt.Sprintf("B%02x,", record.Value[0])
			break
		case info2.E_Li_Xian:
			ourCardStr += fmt.Sprintf("l%02x,", record.Value[0])
			break
		case info2.E_Jie_san:
			ourCardStr += fmt.Sprintf("j%02x,", record.Value[0])
			break
		case info2.E_Pao:
			ourCardStr += fmt.Sprintf("K%02x,", record.Value[0])
			break
		case info2.E_SendCardRight:
			ourCardStr += fmt.Sprintf("s%02x,", record.Value[0])
			break
		case info2.E_JZSendCardRight:
			ourCardStr += fmt.Sprintf("s%02xHR%d,", record.Value[0], record.OpreateRight)
			break
		case info2.E_HandleCardRight:
			ourCardStr += fmt.Sprintf("h%02x,", record.Value[0])
			break
		case info2.E_Gang_ChaoTianGand:
			ourCardStr += fmt.Sprintf("T%02x,", record.Value[0])
			break
		case info2.E_Gang_SmallChaoTianGand:
			ourCardStr += fmt.Sprintf("a%02x,", record.Value[0]) //小朝天
			break
		case info2.E_Baoqing:
			ourCardStr += fmt.Sprintf("q%02x,", record.Value[0])
			break
		case info2.E_Baojiang:
			ourCardStr += fmt.Sprintf("J%02x,", record.Value[0])
			break
		case info2.E_Baojing: //报警
			ourCardStr += fmt.Sprintf("c%02x,", record.Value[0])
			break
		case info2.E_Baofeng:
			ourCardStr += fmt.Sprintf("F%02x,", record.Value[0])
			break
		case info2.E_Baoqi:
			ourCardStr += fmt.Sprintf("D%02x,", record.Value[0])
			break
		case info2.E_BaoTing:
			ourCardStr += fmt.Sprintf("t%02x,", record.Value[0])
			break
		case info2.E_ExchageThreeCard: //换三张换的牌
			ourCardStr += fmt.Sprintf("E%02x,", record.Value[0])
			break
		case info2.E_ExchageThreeEnd: //换三张结束
			ourCardStr += fmt.Sprintf("e%02x,", record.Value[0])
			break
		case info2.E_LastCard: //牌堆最后一张牌
			ourCardStr += fmt.Sprintf("r%02x,", record.Value[0])
			break
		case info2.E_GameScore:
			if fs := strings.Split(fmt.Sprintf("%v", record.UserScore), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s,", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f,", record.UserScore)
			}
		case info2.E_BuHua:
			ourCardStr += fmt.Sprintf("BH%02x,", record.Value[0])
			break
		case info2.E_BuHua_TG:
			ourCardStr += fmt.Sprintf("bh%02x,", record.Value[0])
		case info2.E_Liang:
			ourCardStr += fmt.Sprintf("b%02x,", record.Value[0])
			break
		case info2.E_Change_Pizhi:
			ourCardStr += fmt.Sprintf("p%02x,", record.Value[0])
			break
		default:
			break
		}
	}
	//一炮多响的话有多个人胡牌，回放有几个人胡就应该播放几个胡
	if 0 != len(replayRecord.BigHuKindArray) {
		for _, huKind := range replayRecord.BigHuKindArray {
			ourCardStr += self.getWriteOutReplayBigHuCString(huKind, replayRecord.ProvideUser)
		}
	} else {
		ourCardStr += self.getWriteOutReplayBigHuCString(replayRecord.BigHuKind, replayRecord.ProvideUser)
	}

	//汉川搓虾子 回放需要显示最后一张牌(其他规则这个未赋值的字段编译器会自动初始化为0，不会走到这个逻辑)
	if replayRecord.LeftCard != 0x00 && replayRecord.LeftCard != 0xff {
		ourCardStr += fmt.Sprintf("%d:", 0)
		ourCardStr += fmt.Sprintf("r%02x,", replayRecord.LeftCard)
	}

	if hasHu {
		// 最后补上胡牌U
		for i, s := range endMsgUpdateScore {
			if i >= self.GetPlayerCount() {
				break
			}
			ourCardStr += fmt.Sprintf("%d:", i)
			if fs := strings.Split(fmt.Sprintf("%v", s), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s,", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f,", s)
			}
		}
	}
	return ourCardStr
}

func (self *Common) IsBigHu(bh byte) bool {
	s := [...]byte{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	in := func() bool {
		for _, b := range s {
			if b == bh {
				return true
			}
		}
		return false
	}
	return in()
}

func (self *Common) getCardDataByName(cbCardStr string, msgType byte) (carddata byte) {
	defer func() {
		if carddata != static.INVALID_BYTE {
			carddata = ((carddata / 9) << 4) | (carddata%9 + 1)
		}
	}()
	switch msgType {
	case 0:
		for k, v := range logic2.StrCardsMessage {
			if v == cbCardStr {
				carddata = byte(k)
				return
			}
		}
	default:
		for k, v := range logic2.StrCardsMessage1 {
			if v == cbCardStr {
				carddata = byte(k)
				return
			}
		}
	}
	carddata = static.INVALID_BYTE
	return
}

func (self *Common) GetCardDataByStr(cbCardStr string) (carddataHex byte, carddata byte) {
	var card byte
	defer func() {
		if card != static.INVALID_BYTE {
			carddataHex = ((card / 13) << 4) | (card%13 + 1) //十六进制的牌
			carddata = card + 1                              //十进制的牌
		} else {
			carddataHex = static.INVALID_BYTE
			carddata = static.INVALID_BYTE
		}
	}()

	for k, v := range logic2.PokerCardsMessageUP {
		if v == cbCardStr {
			card = byte(k)
			return
		}
	}
	for k, v := range logic2.PokerCardsMessageLW {
		if v == cbCardStr {
			card = byte(k)
			return
		}
	}

	card = static.INVALID_BYTE
	return
}

// ! 记录每日统计数据
func (self *Common) TableWriteGameDayRecord(SeatId int, Play_Times int, Bw_Times int, Total_Score int, IsValidRound bool, IsBigVaildRound bool) {
	// 包厢的桌子统计每日统计数据
	if self.IsHouse() {
		GamePerson := self.GetPlayerByChair(uint16(SeatId))
		if GamePerson == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDayRecord，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）", SeatId, self.GetTableInfo().KindId, self.GetTableInfo().HId, self.GetTableInfo().Id))
			return
		}

		isBigWinner := Bw_Times > 0

		ValidRound := 0
		ValidBigRound := 0
		if IsValidRound {
			ValidRound = 1
		}
		if IsBigVaildRound {
			ValidBigRound = 1
		}

		var partnerId, superiorId int64
		if houseApi := self.GetHouseApi(); houseApi != nil {
			partnerId, superiorId = houseApi.GetHouseMemberPartnerAndSuperiorId(GamePerson.Uid)
		} else {
			partnerId = 0
			superiorId = 0
		}

		server2.GetDBMgr().HouseUpdataGameDayRecord(self.GetTableInfo().DHId, self.GetTableInfo().FId, self.GetTableInfo().NFId, GamePerson.Uid, Play_Times, Bw_Times, Total_Score, ValidRound, ValidBigRound, partnerId, superiorId, self.Rule.Radix)
		server2.GetDBMgr().UpdataGameHouseMemberRecord(self.GetTableInfo().DHId, GamePerson.Uid, Play_Times, Bw_Times)
		self.OnWriteGameRecord(GamePerson.GetChairID(), fmt.Sprintf("uid:%d,score: %d，is_valid_round:%t,is_big_winner:%t", GamePerson.Uid, Total_Score, IsValidRound, isBigWinner))
		server2.GetDBMgr().UpdateGameRecordTotalByGameNum(self.GetTableInfo().GameNum, GamePerson.Uid, map[string]interface{}{
			"is_valid_round": IsValidRound,
			"is_big_winner":  isBigWinner,
		})
	}
}

// ！由于客户端非法操作，服务器主动与客服端断开连接，使客户端断线重连
func (self *Common) DisconnectOnMisoperation(seatId uint16) error {
	if person := server2.GetPersonMgr().GetPerson(self.GetUidByChair(seatId)); person == nil {
		return errors.New("disconnect fail: person is nil")
	} else {
		person.CloseSession(consts.SESSION_CLOED_BYGAME)
	}
	return nil
}

// 核对少人开局的申请是否能发起
func (self *Common) checkFewerApply() bool {
	// if self.isFewerApplying() {
	// 	syslog.Logger().Debugln("正在申请少人开局过程中")
	// 	return false
	// }
	// 已经坐下的人和最大可以坐下的人相差为1时才可以申请少人开局
	if self.GetChairCount()-self.GetSeatedNum(false) != 1 {
		xlog.Logger().Debugf("已经坐下的人[%d]和最大可以坐下的人[%d]相差为1时才可以申请少人开局", self.GetSeatedNum(false), self.GetChairCount())
		return false
	}
	return self.GetTable().CanFewer()
}

// 核对是否已经有人发起了少人申请
func (self *Common) checkFewerApplied() bool {
	for i, l := 0, len(self.FriendInfo.FewerItem); i < l; i++ {
		if self.FriendInfo.FewerItem[i] == info2.DISMISS_CREATOR {
			xlog.Logger().Debug("已经有人在前面申请少人开局了，后面来的就不处理。")
			return true
		}
	}
	return false
}

// 得到当前座子上玩家个数
func (self *Common) GetSeatedNum(filterOff_line bool) (num int) {
	for i, l := 0, self.GetChairCount(); i < l; i++ {
		userItem := self.GetUserItemByChair(uint16(i))
		if userItem == nil {
			continue
		}
		if filterOff_line && userItem.GetUserStatus() == static.US_OFFLINE {
			continue
		}
		num++
	}
	return
}

// 申请少人开局
func (self *Common) OnFewerApply(uid int64) bool {
	wChairId := self.GetChairByUid(uid)
	if wChairId == static.INVALID_CHAIR { //本桌没有这个玩家
		xlog.Logger().Errorln("申请者不存在")
		return false
	}
	if !self.checkFewerApply() {
		xlog.Logger().Errorln("条件不满足，无法申请少人开局")
		return false
	}
	if self.checkFewerApplied() {
		xlog.Logger().Debugln("已经有人申请的少人开局")
		return true
	}
	// 清空少人开局的申请记录
	self.FriendInfo.InitFewer()
	// 此时可选项应为勾选状态
	self.FriendInfo.FewerShow = true
	// 记录
	self.OnWriteGameRecord(wChairId, "申请少人开局")
	// 状态改变
	for i := 0; i < static.MAX_CHAIR_NORMAL; i++ {
		//自己是解散房间发起者
		if int(wChairId) == i {
			self.FriendInfo.FewerItem[i] = info2.DISMISS_CREATOR
		} else { //其他人等待
			self.FriendInfo.FewerItem[i] = info2.WATING
		}
	}
	xlog.Logger().Debugf("[玩家%d 座位号%d]申请少人开局", uid, wChairId)
	self.SendTableMsg(consts.MsgTypeGameFewerStartReq, &static.Msg_S_FewerNum{
		Uid:      uid,
		FewerNum: self.GetChairCount() - 1,
	})
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameFewerStartReq, static.Msg_S_FewerNum{
		Uid:      uid,
		FewerNum: self.GetChairCount() - 1,
	})
	return true
}

// 其他人响应这个申请
func (self *Common) OnFewerResult(uid int64, msg *static.Msg_C_DismissFriendResult) bool {
	if !self.isFewerApplying() {
		xlog.Logger().Debugln("不在申请过程中")
		return false
	}
	userItem := self.GetUserItemByUid(uid)
	if userItem == nil {
		xlog.Logger().Debugf("OnFewerResult:玩家[%d]不存在.", uid)
		return false
	}
	msg.Id = uid
	msg.FewerNum = self.GetChairCount() - 1
	if userItem.GetTableID() == static.INVALID_TABLE {
		self.SendPersonMsg(consts.MsgTypeGameDismissFriendResult, msg, userItem.GetChairID())
		return true
	}
	// 判断是否为重复选择
	if userItem.GetChairID() < static.MAX_CHAIR_NORMAL && userItem.GetChairID() >= 0 && self.FriendInfo.FewerItem[userItem.GetChairID()] != info2.WATING {
		return true
	}
	// 广播操作
	self.SendTableMsg(consts.MsgTypeGameFewerStartResult, msg)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameFewerStartResult, msg)
	// 状态改变
	if msg.Flag {
		self.OnWriteGameRecord(userItem.GetChairID(), "同意了少人开局")
		if int(self.FriendInfo.AgreeFewer) >= self.GetChairCount()-3 || // 已同意人数 >= 座位数量 - 1(少人) - 1(新同意用户) - 1(申请者)
			self.GetChairCount() == 3 { // 3人转2人 座子上只有两人 一个发起  当有人同意就是都同意了
			proposer := static.INVALID_CHAIR // 申请人
			for i, l := 0, self.GetChairCount(); i < l; i++ {
				if self.FriendInfo.FewerItem[i] == info2.DISMISS_CREATOR {
					proposer = uint16(i)
				}
			}
			// 少人开局申请成功
			self.OnFewerStart(proposer)
		} else {
			self.FriendInfo.AgreeFewer++
			if userItem.GetChairID() < static.MAX_CHAIR_NORMAL && userItem.GetChairID() >= 0 {
				self.FriendInfo.FewerItem[userItem.GetChairID()] = info2.AGREE
			}
		}
	} else {
		self.OnWriteGameRecord(userItem.GetChairID(), "拒绝了少人开局")
		self.OnFewerClose(0, false)
	}

	return true
}

// 少人开局申请通过
func (self *Common) OnFewerStart(originator uint16) {
	xlog.Logger().Debugln("OnFewerStart...")
	self.OnFewerClose(consts.ApplyOk, true)
	// 改变座子数量
	self.GetTable().OnFewerStart(self.GetUids()...)
	self.InitPlayerCount()
	// 修复座位号
	if !self.FixSeat() {
		return
	}
	self.BroadcastTableInfo(true)
	if self.StartVerdict() {
		xlog.Logger().Debugln("少人开局-游戏开始")
		xlog.Logger().WithFields(logrus.Fields{
			"user长度":   len(self.GetTableInfo().Users),
			"user":     &self.GetTableInfo().Users,
			"person长度": len(self.GetTable().GetPersons()),
			"person":   self.GetTable().GetPersons(),
			"player长度": len(self.PlayerInfo),
			"player":   &self.PlayerInfo,
		}).Warningln("少人开局开始后，房间人数信息.")
		self.onBegin()
	}
}

// 换座位
func (self *Common) ReSeat(id1 int64, seat int) bool {
	if self.GetTable().ReSeat(id1, seat) {
		self.FixSeat()
		self.BroadcastTableInfo(true)
		return true
	}
	return false
}

// 换座位
func (self *Common) ExChangSeat(id1, id2 int64) bool {
	if self.GetTable().ExChangeSeat(id1, id2) {
		self.FixSeat()
		self.BroadcastTableInfo(true)
		return true
	}
	return false
}

// 修正玩家的座位号
func (self *Common) FixSeat() bool {
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, player := range *m {
			if player == nil {
				continue
			}
			if seat, ok := self.GetTableInfo().GetSeat(player.Uid); ok {
				player.Seat = uint16(seat)
			}
		}
	})
	return true
}

// 少人开局申请失败
func (self *Common) OnFewerClose(state int, sendMsg bool) {
	self.FriendInfo.InitFewer()
	//self.OnWriteGameRecord(originator, "在少人开局申请流程中途加入或拒绝申请")
	if sendMsg {
		self.SendTableMsg(consts.MsgTypeGameFewerStartClose, &static.Msg_S_FewerClose{
			FewerNum: self.GetSeatedNum(false),
			State:    state,
		})
	}
}

// 少人开局申请可选项显示
func (self *Common) OnFewerShow(originator uint16) {
	// if !self.GetTable().IsNew() {
	// 	return
	// }

	if !self.checkFewerApply() {
		// 如果不能发起少人申请 就可能要隐藏按钮和关闭申请面板
		self.OnFewerHide(originator)
		return
	}

	if self.FriendInfo.FewerShow {
		self.OnFewerApplyInfo(originator)
		return
	}
	self.FriendInfo.FewerShow = true
	self.SendTableMsg(consts.MsgTypeGameFewerStartShow, &static.Msg_S_FewerNum{
		FewerNum: self.GetChairCount() - 1,
	})
	// test:
	// self.OnFewerStart(originator)
}

// 少人开局声请可选项隐藏
func (self *Common) OnFewerHide(originator uint16) {
	// if !self.GetTable().IsNew() {
	// 	return
	// }
	if self.checkFewerApply() {
		// 如果此时能申请少人开局 就要显示按钮
		self.OnFewerShow(originator)
		return
	}

	if self.isFewerApplying() {
		self.OnFewerClose(consts.ApplyIng, true)
	} else if self.FriendInfo.FewerShow {
		if self.GetSeatedNum(false) == self.GetPlayerCount() {
			self.OnFewerClose(consts.ApplyOk, true)
		} else {
			self.OnFewerClose(consts.ApplyBef, true)
		}
	}
	// self.FriendInfo.InitFewer()
	// self.SendTableMsg(constant.MsgTypeGameFewerStartHide, nil)
}

// 断线重连玩家显示少人开局按钮
func (self *Common) onFewerSynch(originator uint16) {
	// if !self.GetTable().IsNew() {
	// 	return
	// }
	if self.FriendInfo.FewerShow {
		self.SendPersonMsg(consts.MsgTypeGameFewerStartShow, &static.Msg_S_FewerNum{
			FewerNum: self.GetChairCount() - 1,
		}, originator)
	}
}

// 检查是否为正在少人开局申请中
func (self *Common) isFewerApplying() bool {
	return !self.FriendInfo.FewerIsNil()
}

// 一般用在断线重连后 给玩法发送少局开始的申请信息
func (self *Common) OnFewerApplyInfo(originator uint16) {
	self.onFewerSynch(originator)
	if !self.isFewerApplying() {
		return
	}
	var fewer_apply_info static.Msg_S_DisMissRoom
	fewer_apply_info.FewerNum = self.GetChairCount() - 1
	//椅子数量跟人数量一样
	xlog.Logger().Debugln("断线重连少人开局信息：", self.FriendInfo.FewerItem)
	for i := uint16(0); i < uint16(self.GetChairCount()) && i < static.MAX_CHAIR; i++ {
		fewer_apply_info.Situation[i] = self.FriendInfo.FewerItem[i]
	}
	self.SendPersonMsg(consts.MsgTypeGameFewerInfo, &fewer_apply_info, originator)
}

// 得到真实游戏玩家uid s
func (self *Common) GetUids() []int64 {
	uids := make([]int64, 0)
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for k, v := range *m {
			if v == nil {
				continue
			}
			uids = append(uids, k)
		}
	})
	return uids
}

// 得到真实游戏玩家信息
func (self *Common) GetLastGamePlayersInfo() []static.LastGamePlayersInfo {
	players := make([]static.LastGamePlayersInfo, 0)
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v == nil {
				continue
			}
			var player static.LastGamePlayersInfo
			player.Uid = v.Uid
			player.Name = v.Name
			player.ImgUrl = v.ImgUrl
			player.Sex = v.Sex
			player.Ip = v.Ip
			player.FaceID = v.FaceID
			player.FaceUrl = v.FaceUrl
			player.UserRight = v.UserRight
			player.Loveliness = v.Loveliness
			player.UserScoreInfo = v.UserScoreInfo
			players = append(players, player)
		}
	})
	return players
}
func (self *Common) GetOtherUids(uid int64) []int64 {
	uids := make([]int64, 0)
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for k, v := range *m {
			if v == nil {
				continue
			}
			if uid == k {
				continue
			}
			uids = append(uids, k)
		}
	})
	return uids
}

func (self *Common) SetDismissRoomTime(dissmisstime int) {
	if dissmisstime > 0 {
		self.DismissRoomTime = dissmisstime
	} else {
		self.DismissRoomTime = static.DISMISS_ROOM_TIMER
	}
	self.DismissRoomTime = static.DISMISS_ROOM_TIMER
}

func (self *Common) SetOfflineRoomTime(offlinetime int) {
	if offlinetime > 0 {
		self.OfflineRoomTime = offlinetime
	} else {
		self.OfflineRoomTime = static.OFFLINE_ROOM_TIMER
	}
	self.OfflineRoomTime = static.OFFLINE_ROOM_TIMER
}

// 设置超级反作弊时，游戏开始后数据正常显示
func (self *Common) SetAiSupperLow(low bool) {
	self.IsAiSupperLow = low
}

// 准备超时
func (self *Common) SetReadyTime(t int) {
	self.Rule.ReadyTimeTag = t
	if t <= 0 {
		self.Rule.ReadyTimeTag = GAME_OPERATION_TIME_READY
	}
	self.Rule.ReadyTimeTag = GAME_OPERATION_TIME_READY
}

func (self *Common) SetVitaminLowPauseTime(pausetime int) {
	if pausetime > 0 {
		self.VitaminLowPauseTime = pausetime
	} else {
		self.VitaminLowPauseTime = static.VITAMIN_LOW_DISMISS_TIMER
	}
	self.VitaminLowPauseTime = static.VITAMIN_LOW_DISMISS_TIMER
}

// 是否为及时的暂停事件
func (self *Common) IsPauseTimelyEvent(eventType int) bool {
	switch eventType {
	case
		consts.EventSettleGaming,
		consts.EventSettleGang,
		consts.EventSettleMagic,
		consts.EventChat:
		return true
	default:
		return false
	}
}

// TODO 游戏服务器偏移玩家积分信息唯一接口
// 不管是小结算算分 还是游戏中途及时算分 传入分数的偏移量
// 游戏算分事件, 如果有及时算分的情况 小结算时 请传入胡牌分 不要传入杠分和飘癞子分
// 返回结算后的玩家剩余积分
// 如果是包厢并且开启了疲劳值 返回玩家剩余的疲劳值
// 其他返回玩家当前剩余积分和当前剩余疲劳值
// 另外，Js客户端后面不再计算分数，只用拿到服务器计算的结果去刷新就OK
func (self *Common) OnSettle(scoreOffset [meta2.MAX_PLAYER]int /*积分偏移量*/, eventType int /*事件类型*/) (score [meta2.MAX_PLAYER]int, vitamin [meta2.MAX_PLAYER]float64) {
	// 如果没有变化 则取内存值下发
	score, vitamin = self.GetTableUserScoreInfo()
	if !IsOffset(scoreOffset[:]) {
		return
	}
	// 开始计算
	var (
		floorOptions *static.FloorVitaminOptions // 包厢疲劳值选项
		err          error                       // 处理错误
	)
	// 得到楼层疲劳值选项
	if houseApi := self.GetHouseApi(); houseApi != nil {
		floorOptions, err = houseApi.GetFloorVitaminOption()
		if err != nil {
			errStr := fmt.Sprintf("get house floor vitamin error: %v", err)
			xlog.Logger().Error(errStr)
			self.OnWriteGameRecord(static.INVALID_CHAIR, errStr)
			return
		}

		if floorOptions != nil && floorOptions.IsVitamin {
			// 防爆庄
			var winIndex []int
			for index, offset := range scoreOffset {
				if offset > 0 {
					winIndex = append(winIndex, index)
				}
			}
			var winOver int
			for index, offset := range scoreOffset {
				if offset >= 0 {
					continue
				}
				vitaminCost := static.SwitchF64ToVitamin(self.GetRealScore(offset))
				chair := uint16(index)
				if user := self.GetUserItemByChair(chair); user != nil {
					if losemem, _ := houseApi.GetMember(user.Uid); losemem != nil {
						if overflow := -vitaminCost - losemem.UVitamin; overflow > 0 { // 不够输
							self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("当前疲劳值:%d，输:%d, offset:%d,不够输", losemem.UVitamin, vitaminCost, offset))
							winOver += int(overflow)
						} else {
							self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("当前疲劳值:%d，输:%d, offset:%d,够输", losemem.UVitamin, vitaminCost, offset))
						}
					} else {
						self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("玩家茶楼成员未找到，输:%d, offset:%d,不够输", vitaminCost, offset))
					}
				} else {
					self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("玩家椅子没找到，输:%d, offset:%d,不够输", vitaminCost, offset))
				}
			}
			winOver = static.SwitchVitaminInt(int64(winOver))
			if winOver > 0 && len(winIndex) > 0 { // 赢钱变少
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("赢钱变少:%d, 赢家index:%v", winOver, winIndex))
				overAvg := winOver / len(winIndex)
				for _, index := range winIndex {
					origin := scoreOffset[index]
					scoreOffset[index] -= overAvg
					if scoreOffset[index] < 0 {
						scoreOffset[index] = 0
					}
					self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("赢钱变少:%d, 赢家index:%v, 原本:%d，变少后:%d", overAvg, index, origin, scoreOffset[index]))
				}
			} else {
				self.OnWriteGameRecord(static.INVALID_CHAIR, "没有输超的玩家")
			}
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, "楼层未开启疲劳值")
		}
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "非茶楼房间不检查爆庄")
	}
	// 效验偏移量 计算分数偏移
	for index, offset := range scoreOffset {
		chair := uint16(index)
		if user := self.GetUserItemByChair(chair); user != nil {
			user.UserScoreInfo.Score += offset
			score[chair] = user.UserScoreInfo.Score
		}
	}
	// 如果是包厢房间 并且开启了疲劳值功能
	if floorOptions != nil && floorOptions.IsVitamin {
		// TODO 注意：这里如果是“及时算分”并且“及时暂停 ” 就会效验疲劳值 并让游戏暂停
		// 计算疲劳值偏移
		// 这里修改后，不再使用redis数据OnScoreOffset作为最终结果，会在发生偏移后读一遍mysql
		// 镜像线实测：四个人查询疲劳值sql共耗时在40-50ms左右，预估正式线也差不了多少。
		self.settleVitamin(floorOptions, scoreOffset, eventType)
		_, vitamin = self.SyncTableUserVitamin()
		self.getServiceFrame().OnScoreOffset(vitamin[:self.GetPlayerCount()])
	} else {
		userScores := [meta2.MAX_PLAYER]float64{}
		for i, s := range score {
			userScores[i] = self.GetRealScore(s)
		}
		self.getServiceFrame().OnScoreOffset(userScores[:self.GetPlayerCount()])
	}
	// 有偏移 就打印下牌桌日志记录下
	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("分数记录:o:%v -> s:%v -> v:%v", scoreOffset, score, vitamin))
	return
}

// ! 游戏小局开始调用
func (self *Common) OnStartNextGame() {
	self.GameEndStatus = static.GS_MJ_PLAY
	//self.m_GameSate = public.GS_MJ_PLAY
	//self.GameTimer.OnNextClean()
	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("游戏第 %d 局开始", self.CurCompleteCount+1))

	self.BSendDissmissReq = false
	//发牌了，那下跑也就结束了
	self.GetTable().SetXiaPaoIng(false)

	//托管状态，自动开始下一局，发送玩家游戏状态，
	//客户端离线状态刷新
	go func() {
		time.Sleep(100 * time.Millisecond)
		self.SendOfflineRemainTime()
	}()
}

// 判断是不是在立即解散的房间已经有人申请了解散
func (self *Common) Has1sDismissReq() bool {
	//是不是理解解散
	if self.Rule.Overtime_dismiss != 1 {
		return false
	}
	if self.BSendDissmissReq {
		return true
	}
	return false
}

// ! 游戏小局结算开始调用
func (self *Common) OnGameRoundBalance() {
	self.GameEndStatus = static.GS_MJ_END
}

func (self *Common) sendTestMsg(addstep bool) {
	if base2.DebugOff {
		return
	}
	//构造数据,发送开始信息
	type StartType struct {
		Step int `json:"step"`
	}
	var GameStart StartType

	//if self.GetTableInfo().Step ==5 {
	//infrastructure.DebugGameStep= 158
	//}

	self.GetTableInfo().Step = base2.GameStep(addstep)
	GameStart.Step = self.GetTableInfo().Step
	self.SendTableMsg("startfortest", GameStart)
}

// ! 游戏小局结算完成调用
func (self *Common) OnGameEnd() {
	self.GameEndStatus = static.GS_MJ_FREE
	//self.m_GameSate = public.GS_MJ_FREE
	self.sendTestMsg(true)
	self.GameTimer.OnNextClean()
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			v.Ctx.Timer.OnNextClean()
		}
	})
}

// 保存最后一局的结算信息到redis
func (self *Common) SaveLastGameinfo(uid int64, gameendString string, resultTotalString string, cardsDataString string) bool {
	if len(gameendString) == 0 || len(resultTotalString) == 0 {
		return false
	}
	recordLastGameInfo := new(static.LastGameInfo)
	recordLastGameInfo.HId = self.GetTableInfo().HId
	recordLastGameInfo.FId = self.GetTableInfo().FId
	recordLastGameInfo.TId = self.GetTableId()
	recordLastGameInfo.KId = self.KIND_ID
	recordLastGameInfo.Users = self.GetUids()
	recordLastGameInfo.PlayersInfo = self.GetLastGamePlayersInfo()
	recordLastGameInfo.GameendInfo = gameendString
	recordLastGameInfo.ResultTotalInfo = resultTotalString
	recordLastGameInfo.CardsDataInfo = cardsDataString
	if err := server2.GetDBMgr().GetDBrsControl().UpdateLastGameInfo(uid, recordLastGameInfo); err == nil {
		server2.GetDBMgr().GetDBrsControl().SetLastGameInfoExpire(uid)
		return true
	}
	return false
}

// 保存道具信息
func (self *Common) GetUserToolInfo(uid int64, toolType int16) (bool, int64) {
	person, err := server2.GetDBMgr().GetDBrControl().GetPerson(uid)
	if err != nil {
		xlog.Logger().Errorln("person not exists: ", uid)
		return false, 0
	}
	return person.CardRecorderDeadAt >= time.Now().Unix(), person.CardRecorderDeadAt
}

// DB -> REDIS -> MEM
// 从数据库/redis同步玩家疲劳值信息，并返回最新的疲劳值数据
func (self *Common) SyncTableUserVitamin() (res [meta2.MAX_PLAYER]int64, vitamin [meta2.MAX_PLAYER]float64) {
	houseApi := self.GetHouseApi()
	if houseApi == nil {
		return
	}
	// 忽略错误的原因是函数里面已经打印了错误，另外，该函数不能因为错误就return
	// 设计原因是mysql的玩家疲劳值信息永远是最新的（理论上也是没有问题的）
	// 所以后面遍历结果时：如果哪个玩家的数据没有取到，那就只能从redis取（目前redis疲劳值只能保证绝大多数情况下的正常）
	vitaminMap, _ := server2.GetDBMgr().GetUsersLatestVitaminFromDataBase(self.GetTableInfo().DHId, self.GetUids()...)
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		if user := self.GetUserItemByChair(uint16(i)); user != nil {
			member := &static.HouseMember{
				DHId: self.GetTableInfo().DHId,
				UId:  user.Uid,
			}
			cli := server2.GetDBMgr().Redis
			member.Lock(cli)
			member, err := houseApi.GetMember(user.Uid)
			if err != nil {
				xlog.Logger().Errorln("invalid house member.", user.Uid, err)
				member.Unlock(cli)
				return
			}
			var flush bool
			if vitamin, ok := vitaminMap[user.Uid]; ok {
				res[user.GetChairID()] = vitamin
				if vitamin != member.UVitamin {
					flush = true
				}
			} else {
				res[user.GetChairID()] = member.UVitamin
			}
			if flush {
				member.UVitamin = res[user.GetChairID()]
				err = houseApi.SetMember(member)
				if err != nil {
					xlog.Logger().Errorln("failed to write house member.", user.Uid, err)
				}
			}
			member.Unlock(cli)
			vitamin[user.GetChairID()] = static.SwitchVitaminToF64(res[user.GetChairID()])
			user.UserScoreInfo.Vitamin = vitamin[user.GetChairID()]
		}
	}
	return
}

func (self *Common) GetTableUserScoreInfo() (score [meta2.MAX_PLAYER]int, vitamin [meta2.MAX_PLAYER]float64) {
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		if user := self.GetUserItemByChair(uint16(i)); user != nil {
			score[user.GetChairID()] = user.UserScoreInfo.Score
			vitamin[user.GetChairID()] = user.UserScoreInfo.Vitamin
		}
	}
	return
}

func IsOffset(offset []int) bool {
	for i := 0; i < len(offset); i++ {
		if offset[i] != 0 {
			return true
		}
	}
	return false
}

// 游戏疲劳值结算
func (self *Common) settleVitamin(fo *static.FloorVitaminOptions, scoreOffset [meta2.MAX_PLAYER]int /*积分偏移量*/, eventType int /*事件类型*/) (vitamin [meta2.MAX_PLAYER]float64) {
	seats := make([]uint16, 0)
	for index, offset := range scoreOffset {
		user := self.GetUserItemByChair(uint16(index))
		if user == nil {
			continue
		}
		// vitaminCost := int64(self.GetRealScore(offset) * constant.VitaminExchangeRate)
		vitaminCost := static.SwitchF64ToVitamin(self.GetRealScore(offset))
		after := self.UpdateUserVitamin(fo, user, vitaminCost, models.GameCost)
		vitamin[user.GetChairID()] = static.SwitchVitaminToF64(after)
		if fo.IsGamePause && fo.ConfiguredLowLimitPause() && after < fo.VitaminLowLimitPause {
			seats = append(seats, user.GetChairID())
		}
	}
	if len(seats) > 0 {
		self.PauseStatus = eventType
	}
	// 暂停校验
	if self.IsPauseTimelyEvent(eventType) {
		self.OnPause(eventType, seats...)
	}
	return vitamin
}

// 游戏是否能继续
func (self *Common) CanContinue() bool {
	if !self.GetTableInfo().IsVitamin {
		return true
	}
	// 校验疲劳值
	if !self.checkVitamin(consts.EventSettleOnBegin) {
		xlog.Logger().Errorln("疲劳值不足 无法开始游戏")
		return false
	}
	return true
}

// 校验疲劳值
func (self *Common) checkVitamin(pauseStatus int) bool {
	if !self.GetTableInfo().Config.IsFriendMode() {
		return true
	}

	houseApi := self.GetHouseApi()

	if houseApi == nil {
		return true
	}

	self.OnWriteGameRecord(static.INVALID_CHAIR, "【校验疲劳值】")
	fo, err := houseApi.GetFloorVitaminOption()
	if err != nil {
		s := fmt.Sprintf("效验疲劳值 从redis取得包厢信息失败::%s", err.Error())
		xlog.Logger().Error(s)
		self.OnWriteGameRecord(static.INVALID_CHAIR, s)
		return true
	}
	if !fo.IsVitamin {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "House Floor Vitamin False")
		return true
	}

	// vitaminMap, err := gsvr.GetDBMgr().GetUsersLatestVitaminFromDataBase(self.GetTableInfo().DHId, self.GetUids()...)
	// if !house.IsGamePause {
	// 	return true
	// }
	afterUpdate := false
	seats := make([]uint16, 0)
	total := [meta2.MAX_PLAYER]float64{}
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		chair := uint16(i)
		user := self.GetUserItemByChair(chair)
		if user == nil {
			continue
		}

		beforVitamin := user.UserScoreInfo.Vitamin

		vitamin := self.GetUserVitamin(user, fo)

		total[i] = static.SwitchVitaminToF64(vitamin)
		afterVitamin := user.UserScoreInfo.Vitamin

		if beforVitamin != afterVitamin {
			afterUpdate = true
			self.SendTableMsg(consts.MsgTypeHouseVitaminSet_Ntf, &static.Msg_Game_UserVitaminUpd{
				Uid:     user.Uid,
				Seat:    user.GetChairID(),
				Vitamin: afterVitamin,
			})
			//发送旁观数据
			self.SendTableLookonMsg(consts.MsgTypeHouseVitaminSet_Ntf, static.Msg_Game_UserVitaminUpd{
				Uid:     user.Uid,
				Seat:    user.GetChairID(),
				Vitamin: afterVitamin,
			})
		}

		if fo.ConfiguredLowLimitPause() && vitamin < fo.VitaminLowLimitPause {
			seats = append(seats, user.GetChairID())
		}
	}

	if afterUpdate {
		self.getServiceFrame().OnScoreOffset(total[:self.GetPlayerCount()])
	}
	if fo.IsGamePause {
		self.OnPause(pauseStatus, seats...)
	}
	return !fo.IsGamePause || len(seats) <= 0
}

// 游戏暂停
func (self *Common) OnPause(eventType int, seats ...uint16 /*被暂停的玩家*/) {
	pausing := self.IsPausing()

	if len(seats) == 0 {
		return
	}

	if eventType == consts.EventSettleNotPause {
		return
	}

	if !self.IsPauseUserUpd(seats...) {
		return
	} else {
		//有人充值恢复,就取消他的竞技点过低充值超时解散的时间标记
		if pausing {
			//原来在self.PauseUsers 中有的玩家,如果在seats里面没有,就证明该玩家充值过了,需要解除标记
			freeUsers := self.GetUpdatePauseUsers(self.PauseUsers, seats)
			if freeUsers != nil {
				for _, chair := range freeUsers {
					pauseuser := self.GetUserItemByChair(chair)
					if pauseuser == nil {
						continue
					}
					pauseuser.UserVitaminLowPauseTag = -1
				}
			}
			//原来在self.PauseUsers 中没有的玩家,如果在seats里面有,就证明该玩家是新加入防沉迷的,需要增加离线标记
			newAddUsers := self.GetUpdatePauseUsers(seats, self.PauseUsers)
			if newAddUsers != nil {
				for _, chair := range newAddUsers {
					pauseuser := self.GetUserItemByChair(chair)
					if pauseuser == nil {
						continue
					}
					pauseuser.UserVitaminLowPauseTag = self.GetNowSystemTimerSecond()
				}
			}
		}
	}

	self.PauseStatus = eventType

	if self.IsPauseTimelyEvent(self.PauseStatus) {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "【游戏过程中暂停游戏】")
		self.OnAutomaticPause()
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "【小局开始前暂停游戏】")
	}
	self.PauseUsers = make([]uint16, 0)
	self.PauseUsers = append(self.PauseUsers, seats...)
	// 广播消息
	{
		if !pausing {
			for _, seat := range self.PauseUsers {
				pauseuser := self.GetUserItemByChair(seat)
				if pauseuser == nil {
					continue
				}
				pauseuser.UserVitaminLowPauseTag = self.GetNowSystemTimerSecond()
			}
		}

		UserNames := self.getPauseUserNames()
		for i := 0; i < meta2.MAX_PLAYER; i++ {
			chair := uint16(i)
			self.SendPersonMsg(consts.MsgTypeGamePause, self.getPausingMsg(chair, UserNames), chair)
			//发送旁观数据
			self.SendTableLookonMsg(consts.MsgTypeGamePause, self.getPausingMsg(chair, UserNames))
		}

		if !pausing {
			// go self.OnPauseCheckScheduler()
		}
	}
}

// 游戏继续
func (self *Common) OnContinue() {
	if !self.IsPausing() {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "【游戏继续】非暂停状态")
		return
	}

	//继续开始游戏,需要对所有人的竞技点过低超时不充值解散标记进行清除
	{
		for _, seat := range self.PauseUsers {
			pauseuser := self.GetUserItemByChair(seat)
			if pauseuser == nil {
				continue
			}
			pauseuser.UserVitaminLowPauseTag = -1
		}
	}

	curStatus := self.PauseStatus
	self.PauseStatus = consts.EventSettleNotPause
	self.PauseUsers = make([]uint16, 0)
	self.SendTableMsg(consts.MsgTypeGameContinue, nil)
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameContinue, nil)
	if self.IsPauseTimelyEvent(curStatus) {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "【游戏继续】游戏中")
		self.OnAutomaticContinue()
		if self.SendCardOpt.Status {
			self.GetTable().Operator(&base2.TableMsg{
				Uid:  self.SendCardOpt.Uid,
				Head: consts.MsgCommonToGameContinue,
				V:    &self.SendCardOpt,
			})
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, "【游戏继续】游戏中,但不是通过发牌触发。")
		}
		self.SendCardOpt.Status = false
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "【游戏继续】游戏小结算")
		self.getServiceFrame().OnGameStart()
	}
}

// 是否在暂停中
func (self *Common) IsPausing() bool /*游戏是否暂停中*/ {
	return self.PauseStatus != consts.EventSettleNotPause && len(self.PauseUsers) > 0
}

// 暂停9秒场
func (self *Common) OnAutomaticPause() {
	if self.Rule.NineSecondRoom {
		self.GameTimer.KillLimitTimer()
	}
}

// 恢复9秒场
func (self *Common) OnAutomaticContinue() {
	if len(self.PlayerInfo) == 0 {
		return
	}
	if self.Rule.NineSecondRoom {
		self.LimitTime = time.Now().Unix() + static.GAME_OPERATION_TIME
		self.GameTimer.SetLimitTimer(static.GAME_OPERATION_TIME)
	}
}

// 包厢成员疲劳值更新
func (self *Common) OnHouseMemVitaminUpdate(mem int64) {
	user := self.GetUserItemByUid(mem)
	if user == nil {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("OnHouseMemVitaminUpdate 玩家不存在:%d", mem))
		return
	}
	self.OnWriteGameRecord(user.GetChairID(), "OnHouseMemVitaminUpdate")

	if self.IsPausing() {
		if self.checkVitamin(self.PauseStatus) {
			self.OnWriteGameRecord(static.INVALID_CHAIR, "推送校验疲劳值通过 开始游戏。")
			self.OnContinue()
		}
	} else {
		self.SendTableMsg(consts.MsgTypeHouseVitaminSet_Ntf, &static.Msg_Game_UserVitaminUpd{
			Uid:     user.Uid,
			Seat:    user.GetChairID(),
			Vitamin: static.SwitchVitaminToF64(self.GetUserVitamin(user, nil)),
		})
		//发送旁观数据
		self.SendTableLookonMsg(consts.MsgTypeHouseVitaminSet_Ntf, &static.Msg_Game_UserVitaminUpd{
			Uid:     user.Uid,
			Seat:    user.GetChairID(),
			Vitamin: static.SwitchVitaminToF64(self.GetUserVitamin(user, nil)),
		})
	}
}

// 断线重连同步暂停消息
func (self *Common) OnSyncPauseInfo(chair uint16) {
	if !self.IsPausing() {
		return
	}
	self.SendPersonMsg(consts.MsgTypeGamePause, self.getPausingMsg(chair, self.getPauseUserNames()), chair)
}

// 断线重连同步暂停消息
func (self *Common) OnSyncPauseInfoLookon(uid int64) {
	if !self.IsPausing() {
		return
	}
	self.SendPersonLookonMsg(consts.MsgTypeGamePause, self.getPausingMsgLookon(self.getPauseUserNames()), uid)
}

// 被暂停效验
func (self *Common) IsPauseUser(chair uint16) bool {
	for _, c := range self.PauseUsers {
		if c == chair {
			return true
		}
	}
	return false
}

// 检查暂停用户是否有更新
func (self *Common) IsPauseUserUpd(seats ...uint16) bool {
	// 判断老用户是不是都存在于新用户中
	if len(seats) != len(self.PauseUsers) {
		return true
	}
	for i := 0; i < len(self.PauseUsers); i++ {
		has := false
		for j := 0; j < len(seats); j++ {
			if self.PauseUsers[i] == seats[j] {
				has = true
				break
			}
		}
		if !has {
			return true
		}
	}
	return false
}

// 检查被暂停游戏后充值恢复的玩家或者新被加入暂停的玩家
func (self *Common) GetUpdatePauseUsers(oldUsers []uint16, newUsers []uint16) []uint16 {
	updateUsers := make([]uint16, 0)
	for i := 0; i < len(oldUsers); i++ {
		bFind := false
		for j := 0; j < len(newUsers); j++ {
			if oldUsers[i] == newUsers[j] {
				bFind = true
				break
			}
		}
		if !bFind {
			updateUsers = append(updateUsers, oldUsers[i])
		}
	}

	if len(updateUsers) > 0 {
		return updateUsers
	}

	return nil
}

// 暂停玩家名字拼接
func (self *Common) getPauseUserNames() string {
	_userNames := make([]string, 0)
	for _, seat := range self.PauseUsers {
		user := self.GetUserItemByChair(seat)
		if user == nil {
			continue
		}
		_userNames = append(_userNames, user.Name)
	}
	return strings.Join(_userNames, "，")
}

// 构造暂停消息
func (self *Common) getPausingMsg(chair uint16, UserNames string) *static.Msg_Game_Addiction {

	var msg static.Msg_Game_Addiction

	msg.Status = make(map[int64]int)

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		user := self.GetUserItemByChair(uint16(i))
		if user == nil {
			continue
		}
		if self.IsPauseUser(user.GetChairID()) {
			msg.Status[user.Uid] = consts.PlayerStatusAnti
			//msg.Remaintime[user.Seat] = public.VITAMIN_LOW_DISMISS_TIMER - (self.GetNowSystemTimerSecond() - user.UserVitaminLowPauseTag)
			msg.Remaintime[user.Seat] = int64(self.VitaminLowPauseTime) - (self.GetNowSystemTimerSecond() - user.UserVitaminLowPauseTag)
		} else {
			msg.Status[user.Uid] = consts.PlayerStatusNormal
			msg.Remaintime[user.Seat] = 0
		}
	}

	if self.IsPauseUser(chair) {
		if self.IsPauseTimelyEvent(self.PauseStatus) {
			msg.Content = consts.MsgContentGamePausePlaying
		} else {
			msg.Content = consts.MsgContentGamePauseOnNext
		}
	} else {
		msg.Content = fmt.Sprintf(consts.MsgContentGamePauseOtherPlayer, UserNames)
	}

	return &msg
}

// 构造暂停消息
func (self *Common) getPausingMsgLookon(UserNames string) *static.Msg_Game_Addiction {

	var msg static.Msg_Game_Addiction

	msg.Status = make(map[int64]int)

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		user := self.GetUserItemByChair(uint16(i))
		if user == nil {
			continue
		}
		if self.IsPauseUser(user.GetChairID()) {
			msg.Status[user.Uid] = consts.PlayerStatusAnti
			//msg.Remaintime[user.Seat] = public.VITAMIN_LOW_DISMISS_TIMER - (self.GetNowSystemTimerSecond() - user.UserVitaminLowPauseTag)
			msg.Remaintime[user.Seat] = int64(self.VitaminLowPauseTime) - (self.GetNowSystemTimerSecond() - user.UserVitaminLowPauseTag)
		} else {
			msg.Status[user.Uid] = consts.PlayerStatusNormal
			msg.Remaintime[user.Seat] = 0
		}
	}
	msg.Content = fmt.Sprintf(consts.MsgContentGamePauseOtherPlayer, UserNames)

	return &msg
}

func (self *Common) UpdateUserVitamin(fo *static.FloorVitaminOptions, user *Player, offset int64, vType models.VitaminChangeType) int64 {
	houseApi := self.GetHouseApi()
	if houseApi == nil || fo == nil {
		return 0
	}

	if !fo.IsVitamin {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "House Floor Vitamin False")
		return 0
	}

	//// cli := GetDBMgr().RedisLock
	//// memLock := public.NewMemLock(self.GetTable().DHId, user.Uid)
	//// memLock.Lock(cli)
	//// defer memLock.Unlock(cli)
	//var err error
	//mem := &public.HouseMember{
	//	DHId: self.GetTableInfo().DHId,
	//	UId:  user.Uid,
	//}
	//cli := gsvr.GetDBMgr().Redis
	//mem.Lock(cli)
	//
	//mem, err = houseApi.GetMember(user.Uid)
	//if err != nil {
	//	mem.Unlock(cli)
	//	self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("UpdateUserVitamin获取桌子包厢成员%d信息失败:%s", user.Uid, err.Error()))
	//	return 0
	//}
	//if offset != 0 {
	//	// 得到前后值
	//	before := mem.UVitamin
	//	after := before + offset
	//	self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("【扣除疲劳值】扣除前:%d, 扣除后：%d, 偏移量：%d.", before, after, offset))
	//
	//	// 得到扣除类型
	//	vType := model.GameCost
	//	if costType == constant.VitaminCostTypeTotal {
	//		if fo.VitaminDeductType == 0 {
	//			vType = model.BigWinCost
	//		} else {
	//			vType = model.GamePay
	//		}
	//	}
	//
	//	// 定义疲劳值日志函数
	//	addVitaminLogFunc := func(uid int64, befvitamin int64, aftvitamin int64, vtype model.VitaminChangeType) error {
	//		tx := gsvr.GetDBMgr().GetDBmControl().Begin()
	//		err := model.AddVitaminLog(self.GetTableInfo().DHId, uid, uid, befvitamin, aftvitamin, vtype, tx)
	//		if err != nil {
	//			return err
	//		}
	//		err = tx.Commit().Error
	//		if err != nil {
	//			return err
	//		}
	//		return nil
	//	}
	//
	//	// 执行疲劳值日志函数
	//	tryNum := 5
	//	ok := false
	//	for i := 0; i < tryNum; i++ {
	//		err := addVitaminLogFunc(user.Uid, before, after, vType)
	//		if err == nil {
	//			ok = true
	//			break
	//		} else {
	//			syslog.Logger().Error("AddVitaminLog error:", err)
	//		}
	//	}
	//
	//	if ok {
	//		// 查询
	//		mem.UVitamin = after
	//		err = houseApi.SetMember(mem)
	//		if err != nil {
	//			s := fmt.Sprintf("UpdateUserVitamin更新包厢成员信息失败:%s", err.Error())
	//			self.OnWriteGameRecord(user.GetChairID(), s)
	//			syslog.Logger().Error(s)
	//		}
	//		mem.Unlock(cli)
	//	} else {
	//		mem.Unlock(cli)
	//		syslog.Logger().Error("AddVitaminLog接口调用错误。")
	//		return before
	//	}
	//
	//	var winLose, Aa, Bw int64
	//
	//	if costType == constant.VitaminCostTypeTotal {
	//		if fo.VitaminDeductType == 0 {
	//			Bw = offset
	//		} else {
	//			Aa = offset
	//		}
	//		err = gsvr.GetDBMgr().InsertVitaminCostClear(self.GetTableInfo().DHId, Aa, Bw)
	//		if err != nil {
	//			s := fmt.Sprintf("AA InsertVitaminCostClear.error:%s", err.Error())
	//			self.OnWriteGameRecord(user.GetChairID(), s)
	//			syslog.Logger().Error(s)
	//		}
	//	} else {
	//		winLose = offset
	//	}
	//
	//	err = gsvr.GetDBMgr().InsertVitaminCostFromLastNode(self.GetTableInfo().DHId, user.Uid, before, mem.UVitamin, winLose, Aa, Bw)
	//	if err != nil {
	//		s := fmt.Sprintf("UpdateUserVitamin.InsertVitaminCostFromLastNode.error:%s", err.Error())
	//		self.OnWriteGameRecord(user.GetChairID(), s)
	//		syslog.Logger().Error(s)
	//	}
	//
	//	err = gsvr.GetDBMgr().InsertVitaminCost(self.GetTableInfo().DHId, self.GetTableInfo().FId, int64(self.GetTableInfo().NFId), user.Uid, Aa, Bw, winLose, mem.Partner)
	//	if err != nil {
	//		s := fmt.Sprintf("AA InsertVitaminCost.error:%s", err.Error())
	//		self.OnWriteGameRecord(user.GetChairID(), s)
	//		syslog.Logger().Error(s)
	//	}
	//} else {
	//	mem.Unlock(cli)
	//}
	vitamin, err := houseApi.UpdateUserVitamin(user.Uid, offset, vType)
	if err != nil {
		errStr := fmt.Sprintf("更新疲劳值错误: %s", err.Error())
		self.OnWriteGameRecord(user.GetChairID(), errStr)
		xlog.Logger().Error(errStr)
	}
	user.UserScoreInfo.Vitamin = static.SwitchVitaminToF64(vitamin)
	return vitamin
}

// 扣除疲劳值房费
func (self *Common) CostVitaminRoomRateAA(houseApi *HouseApi,
	payInfo *models.HouseFloorGearPay) (int64 /*本局合伙人收益*/, map[int64]int64, /*每个玩家的扣除*/
	int /*是否未托管玩家扣除*/, bool /*本局是否为有效局*/, error /*错误处理*/) {

	if houseApi == nil {
		return 0, nil, 0, false, fmt.Errorf("house api is nil")
	}

	if payInfo == nil {
		return 0, nil, 0, false, fmt.Errorf("[room rate]pay eve is nil")
	}

	fo, err := houseApi.GetFloorVitaminOption()
	if err != nil {
		return 0, nil, 0, false, fmt.Errorf("[room rate] get floor option from redis error:%s", err.Error())
	}
	if !fo.IsVitamin {
		return 0, nil, 0, false, fmt.Errorf("[room rate] house floor vitamin is false")
	}

	usersRealCost := make(map[int64]int64)
	var payType int // 0=正常扣除 1=托管扣除 2=奖励均衡扣除

	var cost, SUM, royalty int64
	cost = payInfo.Gear1Cost

	// 托管平摊房费
	var vitaminType models.VitaminChangeType
	//if payInfo.AAPay {
	//	vitaminType = models.GamePay
	//	SUM = cost * int64(self.GetPlayerCount())
	//	royalty = cost
	//} else {
	//	vitaminType = models.BigWinCost
	//	SUM = cost
	//	royalty = cost / int64(self.GetPlayerCount())
	//}

	var usersCost [meta2.MAX_PLAYER]int64

	vitaminType = models.GamePay
	SUM = cost * int64(self.GetPlayerCount())
	royalty = cost
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		usersCost[i] = cost
	}
	self.OnWriteGameRecord(static.INVALID_CHAIR, "[room rate]AA PAY")

	self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate] players cost detail: %+v", usersCost))

	var (
		isValidRound     bool
		isRewardBalanced bool
		house            *models.House
	)
	isValidRound = true
	if self.Rule.TrusteeCostSharing && self.IsTrusteeDismiss() {
		payType = 1
		count := len(self.DismissTrustee)
		if SUM > 0 {
			TCost := SUM / int64(count) / 10 * 10
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate]trustee cost sharing: tuser:%+v, SUM:%d, tcost:%d", self.DismissTrustee, SUM, TCost))
			for i := 0; i < count; i++ {
				if user := self.GetUserItemByChair(self.DismissTrustee[i]); user != nil {
					self.UpdateUserVitamin(fo, user, -TCost, vitaminType)
					usersRealCost[user.Uid] -= TCost
				}
			}
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate]trustee cost sharing: tuser:%+v, SUM:%d, tcost:%d", self.DismissTrustee, SUM, 0))
		}
	} else {
		if !isValidRound {
			house, err = houseApi.GetHouse()
			if err != nil {
				// self.OnWriteGameRecord(public.INVALID_CHAIR, fmt.Sprintf("[room rate] get house err %v", err))
				return 0, nil, 0, false, fmt.Errorf("[room rate] get house err %v", err)
			}
			isRewardBalanced = house.RewardBalanced
			royalty = 0
		}
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate] is valid round: %t, is reward balanced: %t.", isValidRound, isRewardBalanced))

		for i := 0; i < meta2.MAX_PLAYER; i++ {
			if user := self.GetUserItemByChair(uint16(i)); user != nil {
				offset := usersCost[i]
				if isValidRound {
					if offset != 0 {
						self.UpdateUserVitamin(fo, user, -offset, vitaminType)
						usersRealCost[user.Uid] -= offset
					}
				} else if isRewardBalanced {
					offset = static.SwitchIntVitamin(self.GetTableInfo().Config.CardCost) / int64(self.GetPlayerCount()) / 10
					mem, err1 := houseApi.GetMember(user.Uid)
					if err1 != nil {
						self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("[room rate] get member error:%v", err1))
						continue
					}
					var uid int64
					if mem.IsPartner() {
						uid = mem.UId
					} else if mem.Partner > 0 {
						uid = mem.Partner
					} else {
						continue
					}
					payType = 2
					before := mem.UVitamin
					after, err2 := houseApi.UpdateUserVitamin(uid, -offset, vitaminType)
					usersRealCost[uid] -= offset
					self.OnWriteGameRecord(static.INVALID_CHAIR,
						fmt.Sprintf("[room rate] invalid round cost: player id[%d],partner id[%d],offset[%d],before[%d],after[%d],err[%v].",
							user.Uid, uid, -offset, before, after, err2))
				}
			}
		}
	}
	return royalty, usersRealCost, payType, isValidRound, nil
}

// 扣除疲劳值房费
func (self *Common) CostVitaminRoomRate(houseApi *HouseApi, bigWinScore float64, gameScore []float64,
	payInfo *models.HouseFloorGearPay) (int64 /*本局合伙人收益*/, map[int64]int64, /*每个玩家的扣除*/
	int /*是否未托管玩家扣除*/, bool /*本局是否为有效局*/, error /*错误处理*/) {

	if houseApi == nil {
		return 0, nil, 0, false, fmt.Errorf("house api is nil")
	}

	if payInfo == nil {
		return 0, nil, 0, false, fmt.Errorf("[room rate]pay eve is nil")
	}

	fo, err := houseApi.GetFloorVitaminOption()
	if err != nil {
		return 0, nil, 0, false, fmt.Errorf("[room rate] get floor option from redis error:%s", err.Error())
	}
	if !fo.IsVitamin {
		return 0, nil, 0, false, fmt.Errorf("[room rate] house floor vitamin is false")
	}

	usersRealCost := make(map[int64]int64)
	var payType int // 0=正常扣除 1=托管扣除 2=奖励均衡扣除

	if FloorDeductCompatible {
		var isValidRound bool
		ValidScore, _, _ := server2.GetDBMgr().SelectHouseValidRound(self.GetTableInfo().DHId, self.GetTableInfo().FId)
		if bigWinScore >= float64(ValidScore) {
			isValidRound = true
		}
		if isValidRound {
			deductInfo, err1 := houseApi.GetFloorDeductInfo()
			if err1 != nil {
				return 0, nil, 0, false, fmt.Errorf("[room rate]get floor deduct config err: %v", err1)
			}
			var cost, SUM, royalty int64
			// 旧版房费收益始终取基础值
			royalty = deductInfo.BaseDeduct()
			if royalty < 0 {
				royalty = 0
			}

			bw := static.SwitchF64ToVitamin(bigWinScore)
			cost, err1 = deductInfo.GetDeduct(bw)
			if err1 != nil {
				return 0, nil, 0, false, fmt.Errorf("[room rate]get house deduct err: %v", err1)
			}

			// 托管平摊房费
			var vitaminType models.VitaminChangeType
			if deductInfo.AADeduct() {
				vitaminType = models.GamePay
				SUM = cost * int64(self.GetPlayerCount())
			} else {
				vitaminType = models.BigWinCost
				SUM = cost
			}

			var usersCost [meta2.MAX_PLAYER]int64

			if deductInfo.AADeduct() {
				for i := 0; i < meta2.MAX_PLAYER; i++ {
					usersCost[i] = cost
				}
				self.OnWriteGameRecord(static.INVALID_CHAIR, "[room rate]AA PAY")
			} else {
				var bwCount int64
				for i := 0; i < len(gameScore); i++ {
					if i < self.GetPlayerCount() {
						if gameScore[i] == bigWinScore {
							bwCount++
						}
					}
				}
				for i := 0; i < len(gameScore); i++ {
					if i < self.GetPlayerCount() {
						if gameScore[i] == bigWinScore {
							usersCost[i] = cost / bwCount
						}
					}
				}
				self.OnWriteGameRecord(static.INVALID_CHAIR, "[room rate]BW PAY")
			}

			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate] players cost detail: %+v", usersCost))

			if self.Rule.TrusteeCostSharing && self.IsTrusteeDismiss() {
				payType = 1
				count := len(self.DismissTrustee)
				TCost := SUM / int64(count) / 10 * 10
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate]trustee cost sharing: tuser:%+v, SUM:%d, tcost:%d", self.DismissTrustee, SUM, TCost))
				for i := 0; i < count; i++ {
					if user := self.GetUserItemByChair(self.DismissTrustee[i]); user != nil {
						self.UpdateUserVitamin(fo, user, -TCost, vitaminType)
						usersRealCost[user.Uid] -= TCost
					}
				}
			} else {
				for i := 0; i < meta2.MAX_PLAYER; i++ {
					if user := self.GetUserItemByChair(uint16(i)); user != nil {
						offset := usersCost[i]
						if offset != 0 {
							self.UpdateUserVitamin(fo, user, -offset, vitaminType)
							usersRealCost[user.Uid] -= offset
						}
					}
				}
			}
			return royalty, usersRealCost, payType, isValidRound, nil
		} else {
			return 0, usersRealCost, payType, isValidRound, nil
		}
	} else {
		var cost, SUM, royalty int64
		bw := static.SwitchF64ToVitamin(bigWinScore)
		cost, err = payInfo.GetPay(bw)
		if err != nil {
			return 0, nil, 0, false, fmt.Errorf("[room rate]get house pay err: %v", err)
		}

		// 托管平摊房费
		var vitaminType models.VitaminChangeType
		//if payInfo.AAPay {
		//	vitaminType = models.GamePay
		//	SUM = cost * int64(self.GetPlayerCount())
		//	royalty = cost
		//} else {
		//	vitaminType = models.BigWinCost
		//	SUM = cost
		//	royalty = cost / int64(self.GetPlayerCount())
		//}

		var usersCost [meta2.MAX_PLAYER]int64

		if payInfo.AAPay {
			vitaminType = models.GamePay
			SUM = cost * int64(self.GetPlayerCount())
			royalty = cost
			for i := 0; i < meta2.MAX_PLAYER; i++ {
				usersCost[i] = cost
			}
			self.OnWriteGameRecord(static.INVALID_CHAIR, "[room rate]AA PAY")
		} else {
			var bwCount int64
			for i := 0; i < len(gameScore); i++ {
				if i < self.GetPlayerCount() {
					if gameScore[i] == bigWinScore {
						bwCount++
					}
				}
			}
			for i := 0; i < len(gameScore); i++ {
				if i < self.GetPlayerCount() {
					if gameScore[i] == bigWinScore {
						usersCost[i] = cost
					}
				}
			}
			vitaminType = models.BigWinCost
			if cost > 0 {
				SUM = cost * bwCount
				royalty = SUM / int64(self.GetPlayerCount())
			}
			self.OnWriteGameRecord(static.INVALID_CHAIR, "[room rate]BW PAY")
		}

		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate] players cost detail: %+v", usersCost))

		var (
			isValidRound     bool
			isRewardBalanced bool
			house            *models.House
		)
		isValidRound = payInfo.IsValidRound(bw)
		if self.Rule.TrusteeCostSharing && self.IsTrusteeDismiss() {
			payType = 1
			count := len(self.DismissTrustee)
			if SUM > 0 {
				TCost := SUM / int64(count) / 10 * 10
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate]trustee cost sharing: tuser:%+v, SUM:%d, tcost:%d", self.DismissTrustee, SUM, TCost))
				for i := 0; i < count; i++ {
					if user := self.GetUserItemByChair(self.DismissTrustee[i]); user != nil {
						self.UpdateUserVitamin(fo, user, -TCost, vitaminType)
						usersRealCost[user.Uid] -= TCost
					}
				}
			} else {
				self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate]trustee cost sharing: tuser:%+v, SUM:%d, tcost:%d", self.DismissTrustee, SUM, 0))
			}
		} else {
			if !isValidRound {
				house, err = houseApi.GetHouse()
				if err != nil {
					// self.OnWriteGameRecord(public.INVALID_CHAIR, fmt.Sprintf("[room rate] get house err %v", err))
					return 0, nil, 0, false, fmt.Errorf("[room rate] get house err %v", err)
				}
				isRewardBalanced = house.RewardBalanced
				royalty = 0
			}
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("[room rate] is valid round: %t, is reward balanced: %t.", isValidRound, isRewardBalanced))

			for i := 0; i < meta2.MAX_PLAYER; i++ {
				if user := self.GetUserItemByChair(uint16(i)); user != nil {
					offset := usersCost[i]
					if isValidRound {
						if offset != 0 {
							self.UpdateUserVitamin(fo, user, -offset, vitaminType)
							usersRealCost[user.Uid] -= offset
						}
					} else if isRewardBalanced {
						offset = static.SwitchIntVitamin(self.GetTableInfo().Config.CardCost) / int64(self.GetPlayerCount()) / 10
						mem, err1 := houseApi.GetMember(user.Uid)
						if err1 != nil {
							self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("[room rate] get member error:%v", err1))
							continue
						}
						var uid int64
						if mem.IsPartner() {
							uid = mem.UId
						} else if mem.Partner > 0 {
							uid = mem.Partner
						} else {
							continue
						}
						payType = 2
						before := mem.UVitamin
						after, err2 := houseApi.UpdateUserVitamin(uid, -offset, vitaminType)
						usersRealCost[uid] -= offset
						self.OnWriteGameRecord(static.INVALID_CHAIR,
							fmt.Sprintf("[room rate] invalid round cost: player id[%d],partner id[%d],offset[%d],before[%d],after[%d],err[%v].",
								user.Uid, uid, -offset, before, after, err2))
					}
				}
			}
		}
		return royalty, usersRealCost, payType, isValidRound, nil
	}
}

func (self *Common) CheckCostSpecial(cbReason byte) bool {
	if self.GetTableInfo().IsVitamin {
		return IsGameEndNormal(cbReason) || self.GetTableInfo().IsCost
	}
	return true
}

func IsGameEndNormal(cbReason byte) bool {
	return cbReason == meta2.GOT_NORMAL ||
		cbReason == meta2.GOT_ZHONGTU ||
		cbReason == meta2.GOT_DOUBLEKILL
}

func (self *Common) OnPauseCheckScheduler() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C
		if !self.IsPausing() {
			break
		}
		if self.checkVitamin(self.PauseStatus) {
			self.OnWriteGameRecord(static.INVALID_CHAIR, "定时校验疲劳值通过 开始游戏。")
			self.OnContinue()
			break
		}
	}
	ticker.Stop()
}

// 大局打完后预判作弊玩家
func (self *Common) AnalyticalCribber() {
	if !self.IsHouse() {
		return
	}

	// 得到包厢信息
	house, err := server2.GetDBMgr().GetDBrControl().GetHouseInfoById(self.GetTableInfo().DHId)
	if err != nil {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("从redis获取包厢信息失败:%s", err.Error()))
		return
	}
	if !house.MixActive || house.TableJoinType != consts.NoCheat || !house.AICheck {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "包厢未开启混排智能筛选")
		return
	}

	// 得到所有人的作弊名单
	userCribberList, err := server2.GetDBMgr().GetDBrControl().GetMultiMemCribberList(
		self.GetTableInfo().DHId,
		self.GetUids()...,
	)

	if err != nil {
		e := fmt.Sprintf("从redis获取所有玩家作弊名单失败:%s", err.Error())
		xlog.Logger().Error(e)
		self.OnWriteGameRecord(static.INVALID_CHAIR, e)
		return
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		user := self.GetUserItemByChair(uint16(i))
		if user == nil {
			continue
		}
		res, err := server2.GetDBMgr().GetDBrControl().GetTeamMemGameRecord(self.GetTableInfo().DHId, user.Uid)
		if err != nil {
			e := fmt.Sprintf("从redis获取对战玩家记录失败:%s", err.Error())
			xlog.Logger().Error(e)
			self.OnWriteGameRecord(user.GetChairID(), e)
			return
		}

		for j := 0; j < meta2.MAX_PLAYER; j++ {
			if i == j {
				continue
			}
			teammate := self.GetUserItemByChair(uint16(j))
			if teammate == nil {
				continue
			}
			res[teammate.Uid]++
			// 玩家次数满足
			if res[teammate.Uid] >= server2.GetServer().ConHouse.MixAIOnRoundNum {
				if pt, ok := userCribberList[user.Uid][teammate.Uid]; ok {
					self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("玩家%d和玩家%d已经是作弊玩家，pt:%d，不用再次计算是否拉黑。",
						user.Uid, teammate.Uid, pt))
				} else {
					total := server2.GetDBMgr().SumHouseMemTodayTotalScore(self.GetTableInfo().DHId, user.Uid, teammate.Uid)
					// 满足作弊嫌疑
					if total >= float64(house.AITotalScoreLimit) {
						// 双方拉黑
						userCribberList[user.Uid][teammate.Uid] = 0
						userCribberList[teammate.Uid][user.Uid] = 0
					}
				}
			}
		}
		// syslog.Logger().Error("存进去的res", res)
		// 保存最新记录
		err = server2.GetDBMgr().GetDBrControl().SetTeamMemGameRecord(self.GetTableInfo().DHId, user.Uid, res)
		if err != nil {
			e := fmt.Sprintf("从redis插入对战玩家记录失败:%s", err.Error())
			xlog.Logger().Error(e)
			self.OnWriteGameRecord(user.GetChairID(), e)
			return
		}
	}
	// 保存所有玩家作弊名单
	err = server2.GetDBMgr().GetDBrControl().SetMultiMemCribberList(self.GetTableInfo().DHId, userCribberList)
	if err != nil {
		e := fmt.Sprintf("从redis插入所有玩家作弊名单失败:%s", err.Error())
		xlog.Logger().Error(e)
		self.OnWriteGameRecord(static.INVALID_CHAIR, e)
		return
	}
}

// 大局打完后预判是否释放掉作弊玩家
func (self *Common) AnalyticalReleaseCribber() {
	if !self.IsHouse() {
		return
	}
	// 得到包厢信息
	house, err := server2.GetDBMgr().GetDBrControl().GetHouseInfoById(self.GetTableInfo().DHId)
	if err != nil {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("从redis获取包厢信息失败:%s", err.Error()))
		return
	}

	if !house.MixActive || house.TableJoinType != consts.NoCheat || !house.AICheck {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "Release包厢未开启混排智能筛选")
		return
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		user := self.GetUserItemByChair(uint16(i))
		if user == nil {
			continue
		}

		res, err := server2.GetDBMgr().GetDBrControl().GetMemCribberList(self.GetTableInfo().DHId, user.Uid)
		if err != nil {
			e := fmt.Sprintf("从redis获取对战玩家%d作弊列表失败:%s", user.Uid, err.Error())
			xlog.Logger().Error(e)
			self.OnWriteGameRecord(user.GetChairID(), e)
			continue
		}

		for key, val := range res {
			// 如果已经在作弊黑名单里面的玩家还是排到了一起
			// 可能是因为他们在非智能模式下手动点房间入桌 也可能是因为程序逻辑错误
			// 此时认为异常数据，不做处理，保留原有的数据
			if self.ExistsUser(key) {
				continue
			}

			teammate, err := server2.GetDBMgr().GetDBrControl().GetMemCribberList(self.GetTableInfo().DHId, key)
			if err != nil {
				e := fmt.Sprintf("从redis获取对战玩家%d队友%d作弊列表失败:%s", user.Uid, key, err.Error())
				xlog.Logger().Error(e)
				self.OnWriteGameRecord(user.GetChairID(), e)
				return
			}

			val++

			if val >= server2.GetServer().ConHouse.MixAIOffRoundNum {
				self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("玩家%d 和 玩家%d 满足释放条件%d - %d",
					user.Uid, key, val, server2.GetServer().ConHouse.MixAIOffRoundNum))
				delete(res, key)
				delete(teammate, user.Uid)
			} else {
				self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("玩家%d 和 玩家%d 不满足释放条件%d - %d",
					user.Uid, key, val, server2.GetServer().ConHouse.MixAIOffRoundNum))
				res[key] = val
				teammate[user.Uid] = val
			}

			err = server2.GetDBMgr().GetDBrControl().SetMemCribberList(self.GetTableInfo().DHId, key, teammate)
			if err != nil {
				e := fmt.Sprintf("从redis插入对战玩家%d队友%d作弊列表失败:%s", user.Uid, key, err.Error())
				xlog.Logger().Error(e)
				self.OnWriteGameRecord(user.GetChairID(), e)
				return
			}
		}

		err = server2.GetDBMgr().GetDBrControl().SetMemCribberList(self.GetTableInfo().DHId, user.Uid, res)
		if err != nil {
			e := fmt.Sprintf("从redis插入对战玩家%d作弊列表失败:%s", user.Uid, err.Error())
			xlog.Logger().Error(e)
			self.OnWriteGameRecord(user.GetChairID(), e)
		}
	}
}

func (self *Common) SetDissmissCount(dissmisscount int) {
	if self.IsSetDissmiss {
		return
	}
	if dissmisscount >= 999 {
		dissmisscount = -1
	}
	self.DissmisReqMax = -1
	self.IsSetDissmiss = true
}

// 20191210 苏大强  游戏过程中，托管玩家自动同意解散房间 调用条件是有人已经申请解散房间了（游戏过程中）
func (self *Common) TrustUserAutoOperDismiss(dissmiss bool) {
	for j := 0; j < self.GetChairCount(); j++ {
		if j < static.MAX_CHAIR_NORMAL {
			if self.FriendInfo.MissItem[j] == info2.WATING {
				userItme := self.GetUserItemByChair(uint16(j))
				if userItme == nil {
					continue
				}
				//托管玩家自动同意
				if userItme.CheckTRUST() {
					var _msg = &static.Msg_C_DismissFriendResult{
						Id:   userItme.Uid,
						Flag: dissmiss,
					}
					self.OnDismissResult(userItme.Uid, _msg)
					continue
				}
			}
		}
	}
}

// 推送table eve
func (self *Common) BroadcastTableInfo(private bool) {
	self.GetTable().BroadcastTableInfo(private)
}

// 在结算前广播桌子消息
func (self *Common) BeforeBalance() {
	// 如果是超级防作弊模式下，总结算之前就推送一波真实数据给所有人
	if self.GetTableInfo().IsAiSuper || self.GetTableInfo().IsAnonymous {
		self.BroadcastTableInfo(false)
	}
}

// ! 解析规则difen
func (self *Common) ParseFloatDiFen(diFen int, scoreRadix int) {
	// 强制转换低分1,真实分数为Radix/100
	if scoreRadix == 0 {
		self.Rule.DiFen = diFen
		self.Rule.Radix = 1
	} else {
		self.Rule.DiFen = diFen
		self.Rule.Radix = scoreRadix
	}
}

// 20200416 苏大强 重新写消息
func (self *Common) recordChat(chat *static.Msg_C_UserChat, yy *static.Msg_C_UserYYInfo) {
	newrecord := static.ChatLogs{}
	var rank byte = 0
	if chat != nil {
		//文字记录
		newrecord.UserChat = *chat
		rank = chat.Rank
	}
	if yy != nil {
		//yy要调整一下
		newrecord.UserChat.UserID = yy.UserID
		newrecord.UserChat.Color = 3
		newrecord.UserChat.TargetUserID = -1
		newrecord.UserChat.Index = -1
		newrecord.UserChat.Message = yy.Addr
		rank = yy.Rank
	}
	if rank == 1 {
		newrecord.ExtInfo.CurCompleteCount = self.CurCompleteCount
		newrecord.ExtInfo.GameEndStatus = self.GameEndStatus
		newrecord.ExtInfo.Sendtime = time.Now().Unix()
	}
	self.Chatlog = append(self.Chatlog, newrecord)
}

//20200604 苏大强 检查是不是要修改离线踢人时间
/*
static.GS_MJ_FREE 目前暂时认为这个就是下跑前的阶段，通过局数来判断是不是开桌
static.GS_MJ_PLAY 目前下跑阶段也是游戏阶段，但是为了将来可能出现的变化，我这里还是分开了
static.GS_MJ_PAO 下跑
*/
func (self *Common) CheckOfflineKickTime(gameStatus int) {
	//目前除了第一局会出现更替，下跑和开始游戏都是一个值，为了将来可能出现的变化这里还是包含一下吧
	switch gameStatus {
	case static.GS_MJ_FREE:
		//目前这个在重置桌子的时候会复位，但是第一局的时候要判断下和OfflineKick_init是不是不同的，所以这里要判断下局数，如果是0局就要判断一下
		if self.CurCompleteCount == 0 {
			self.SetOfflineRoomTime(self.Rule.OfflineKick.OfflineKick_init)
		}
	case static.GS_MJ_PAO:
		//跑阶段目前是被定义为游戏阶段，为将来的可能，如果是首局，那么就和OfflineKick_init比较，如果不是就和OfflineKick_start比较
		if self.CurCompleteCount == 0 {
			//首开局 和OfflineKick_init比较
			if self.Rule.OfflineKick.OfflineKick_init != self.Rule.OfflineKick.OfflineKick_Piao {
				self.SetOfflineRoomTime(self.Rule.OfflineKick.OfflineKick_Piao)
			}
		} else {
			//不是首局就和OfflineKick_start比較,如果兩個相等就不改了，因爲首局已經設定過了
			if self.Rule.OfflineKick.OfflineKick_start != self.Rule.OfflineKick.OfflineKick_Piao {
				self.SetOfflineRoomTime(self.Rule.OfflineKick.OfflineKick_Piao)
			}
		}
	case static.GS_MJ_PLAY:
		//发牌情况
		if self.CurCompleteCount == 0 {
			//首局的情况下有下跑和没下跑的情况
			if self.Rule.HasPao {
				if self.Rule.OfflineKick.OfflineKick_start != self.Rule.OfflineKick.OfflineKick_Piao {
					self.SetOfflineRoomTime(self.Rule.OfflineKick.OfflineKick_start)
				}
			} else {
				if self.Rule.OfflineKick.OfflineKick_start != self.Rule.OfflineKick.OfflineKick_init {
					self.SetOfflineRoomTime(self.Rule.OfflineKick.OfflineKick_start)
				}
			}
		} else {
			//不是首局就看有没有下跑的情况
			if self.Rule.HasPao {
				if self.Rule.OfflineKick.OfflineKick_start != self.Rule.OfflineKick.OfflineKick_Piao {
					self.SetOfflineRoomTime(self.Rule.OfflineKick.OfflineKick_start)
				}
			}
		}
	}
}

// 统计用户游戏中的操作次数
func (self *Common) StatisticsUserGameOpt() {
	statistics := make([]*models.StatisticsUserGameOpt, 0)
	//调用统计接口 写入这一局 的统计数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		statistics = append(statistics,
			&models.StatisticsUserGameOpt{
				Uid:   _userItem.Uid,
				Times: _userItem.Ctx.StatisticsOpeningUserHu,
				Type:  models.UserGameOpeningHu,
			},
			&models.StatisticsUserGameOpt{
				Uid:   _userItem.Uid,
				Times: _userItem.Ctx.StatisticsGangKaiCount,
				Type:  models.UserGameGangKai,
			},
			&models.StatisticsUserGameOpt{
				Uid:   _userItem.Uid,
				Times: _userItem.Ctx.StatisticsHave4MagicCount,
				Type:  models.UserGameHave4Magic,
			},
			&models.StatisticsUserGameOpt{
				Uid:   _userItem.Uid,
				Times: _userItem.Ctx.StatisticsOpening4MagicCount,
				Type:  models.UserGameOpening4Magic,
			},
		)
	}

	kindId, gameNum, now := self.KIND_ID, self.GetTableInfo().GameNum, time.Now()
	for i := 0; i < len(statistics); i++ {
		if stc := statistics[i]; stc != nil {
			if stc.Times > 0 {
				stc.KindID = kindId
				stc.GameNum = gameNum
				stc.CreatedAt = now
				if err := server2.GetDBMgr().GetDBmControl().Create(stc).Error; err != nil {
					xlog.Logger().Errorf("Uid:%d:StatisticsUserGameOpt.Error:%v", stc.Uid, err)
				}
			}
		}
	}
	self.initStatisticsUserGameOptData()
}

// 初始化玩家的统计数据
func (self *Common) initStatisticsUserGameOptData() {
	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		_userItem.Ctx.StatisticsOpeningUserHu = 0
		_userItem.Ctx.StatisticsGangKaiCount = 0
		_userItem.Ctx.StatisticsHave4MagicCount = 0
		_userItem.Ctx.StatisticsOpening4MagicCount = 0
	}
}

// 20200916 苏大强 大结算要知道队长信息
func (self *Common) GetPartnerInfo(seatId int) (partnerInfo *static.HouseMember) {
	GamePerson := self.GetPlayerByChair(uint16(seatId))
	if GamePerson == nil {
		return
	}
	if houseApi := self.GetHouseApi(); houseApi != nil {
		partnerInfo, _ = houseApi.GetHouseMemberPartnerInfo(GamePerson.Uid)
	}
	return
}

// ! 推送观战消息列表
func (self *Common) SendWatcherList(uid int64) {
	uidList := server2.GetDBMgr().GetDBrControl().GetWatchTablePlayer(self.GetTableId())
	var ack static.Msg_S_WatcherList
	//ack.Items = make([]public.WatcherListItem, len(uidList))
	for _, tempuid := range uidList {
		person := server2.GetLazyUser(tempuid)
		item := static.WatcherListItem{}
		item.Uid = person.Uid
		item.Name = person.Name
		item.Sex = person.Sex
		item.ImageUrl = person.ImageUrl
		ack.Items = append(ack.Items, item)
	}
	//支持旁观玩家和游戏玩家都能查询
	if useritme := self.GetUserItemByUid(uid); useritme != nil {
		self.SendUserMsg(consts.MstTypeWatcherList, ack, uid)
	} else {
		self.SendLookonUserMsg(consts.MstTypeWatcherList, ack, uid)
	}
}

//20201111 苏大强 查找托管时间最长的用户，只比较最后一次，如果有多个就按照座位号来，选择靠前还是靠后，默认是最前
/*
其实没想好排序的问题,但是一定是有先后的
*/
func (self *Common) FindTrustInfo(info []int64, first bool) (user uint16) {
	if len(info) == 0 {
		return static.INVALID_CHAIR
	}
	checktime := time.Now().Unix()
	var check int64 = 0
	user = static.INVALID_CHAIR
	//可能有相同时间的
	for index, v := range info {
		if v == 0 {
			continue
		}
		if first {
			if checktime-v > check {
				user = uint16(index)
				check = checktime - v
			}
		} else {
			if checktime-v < check || check == 0 {
				user = uint16(index)
				check = checktime - v
			}
		}
	}
	return
}

// 最多支持10个玩家的记牌器，Points牌值列表（每个游戏不同），PlayerCards手牌列表，DownCards底牌列表（比如3人拱，升级，斗地主）
func (self *Common) InitCardRecorder(Points []int, PlayerCards [10][]byte, DownCards []byte) {
	self.CardRecorder = [10][]meta2.CardRecorder{}
	if len(Points) == 0 {
		cardPoint := []int{logic2.CP_SKY_S, logic2.CP_RJ_S, logic2.CP_BJ_S, logic2.CP_2_S, logic2.CP_A_S, logic2.CP_K_S, logic2.CP_Q_S,
			logic2.CP_J_S, logic2.CP_10_S, logic2.CP_9_S, logic2.CP_8_S, logic2.CP_7_S, logic2.CP_6_S, logic2.CP_5_S, logic2.CP_4_S, logic2.CP_3_S}
		Points = cardPoint
	}
	//统计每个玩家的牌数目
	pCardsNum := [10][]int{}
	dCardsNum := []int{}
	for i := 0; int(i) < self.GetPlayerCount() && i < 10; i++ {
		for _, Point := range Points {
			num := 0
			for k := 0; k < len(PlayerCards[i]); k++ {
				if Point > 0 && Point == int(logic2.GetCardPoint(PlayerCards[i][k])) {
					num++
				}
			}
			pCardsNum[i] = append(pCardsNum[i], num)
		}
	}
	for _, Point := range Points {
		num := 0
		for k := 0; k < len(DownCards); k++ {
			if Point > 0 && Point == int(logic2.GetCardPoint(DownCards[k])) {
				num++
			}
		}
		dCardsNum = append(dCardsNum, num)
	}
	//构造数据
	for i := 0; int(i) < self.GetPlayerCount(); i++ {
		for pointIndex, Point := range Points {
			var CRItem meta2.CardRecorder
			CRItem.Point = Point
			Num := 0
			for ii := 0; int(ii) < self.GetPlayerCount(); ii++ {
				//自己的手牌不算
				if ii != i && pointIndex < len(pCardsNum[ii]) {
					Num += pCardsNum[ii][pointIndex]
				}
			}
			//牌堆的牌要算
			if pointIndex < len(dCardsNum) {
				Num += dCardsNum[pointIndex]
			}
			CRItem.Num = Num

			self.CardRecorder[i] = append(self.CardRecorder[i], CRItem)
		}
	}
}

func (self *Common) UpdateCardRecorder(seat uint16, outCards []byte) {
	if int(seat) > self.GetPlayerCount() {
		self.OnWriteGameRecord(seat, fmt.Sprintf("UpdateCardRecorder 更新记牌器数据时，玩家座位号有误:%d", seat))
		return
	}

	for crI, CRItem := range self.CardRecorder[seat] {
		num := 0
		for k := 0; k < len(outCards); k++ {
			if CRItem.Point > 0 && CRItem.Point == int(logic2.GetCardPoint(outCards[k])) {
				num++
			}
		}
		if self.CardRecorder[seat][crI].Num >= num {
			self.CardRecorder[seat][crI].Num -= num
		} else {
			self.CardRecorder[seat][crI].Num = 0
		}
	}
}

// mustOpen == 1表示强制展开记牌器列表
func (self *Common) SendCardRecorder(wChairID uint16, mustOpen int) {
	if !self.GetTableInfo().IsBegin() {
		//游戏大局没有开始，兑换记牌器也不发记牌器数据
		return
	}
	if self.GameEndStatus == static.GS_MJ_END || self.GameEndStatus == static.GS_MJ_FREE {
		//小结算，小结算中兑换记牌器也不发记牌器数据
		return
	}
	if !self.CardRecordFlag {
		//不能使用记牌器道具功能时，不能发送记牌器数据
		return
	}
	//wChairID超过了玩家数目时，表示所有玩家都要发
	for i := uint16(0); int(i) < self.GetPlayerCount(); i++ {
		if i == wChairID || int(wChairID) >= self.GetPlayerCount() {
			if self.CheckCardRecorderEnable(i) {
				//
				continue
			}
			//检查记牌器的失效时间
			if flag, deadAt := self.GetUserToolInfo(self.GetUidByChair(i), 0); !flag {
				//没有记牌器或记牌器已经失效
				if self.RoundTimeStart > deadAt {
					//失效时间比游戏开始时间小，说明在游戏开始时就已经过期，
					continue
				}
			}

			var sendCR static.Msg_S_Tool_SendCardRecorder
			for _, Item := range self.CardRecorder[i] {
				var CrItem static.CardRecorderItem
				CrItem.Num = Item.Num
				CrItem.Point = Item.Point
				sendCR.CardRecorderItem = append(sendCR.CardRecorderItem, CrItem)
			}
			sendCR.MustOpen = mustOpen
			self.SendPersonMsg(consts.MsgTypeSendCardRecorder, sendCR, i)
		}
	}
}

// 在每小局开始时调用，如果是本局及时生效，不需要调用此函数，如果是下局生效则需要调用
func (self *Common) InitCardRecorderEnable() {
	self.BCardRcdNextAble = [10]bool{false}
	//wChairID超过了玩家数目时，表示所有玩家都要发
	for i := uint16(0); int(i) < self.GetPlayerCount(); i++ {
		//检查记牌器的失效时间
		if flag, _ := self.GetUserToolInfo(self.GetUidByChair(i), 0); !flag {
			self.BCardRcdNextAble[i] = true // 下局生效, 即本局没有生效
		}
	}
}

// 记牌器是否生效
func (self *Common) CheckCardRecorderEnable(seat uint16) bool {
	if !self.GetTableInfo().IsBegin() {
		//游戏大局没有开始，兑换记牌器也不发记牌器数据
		return false
	}
	if int(seat) >= self.GetPlayerCount() {
		return false
	}

	return self.BCardRcdNextAble[seat]

}

// 随机换座 沈强的代码
func (self *Common) RandChangeSeat(changeSeat bool, changeBank bool) {
	if changeSeat {
		//目前出功能先，这里庄家位置不换，那就把现在能换的座位号放进去混乱
		randslice := []uint16{}
		self.PlayerInfoRead(func(m *map[int64]*Player) {
			for _, v := range *m {
				if v != nil {
					if v.Seat == self.BankerUser && !changeBank {
						continue
					}
					randslice = append(randslice, v.Seat)
				}
			}
		})
		oldrandslice := make([]uint16, len(randslice))
		copy(oldrandslice, randslice[:])
		mahlib2.RandSlice(randslice)
		changeGathers := self.MakechangeGathers(oldrandslice, randslice)
		if len(changeGathers) != 0 {
			for _, v := range changeGathers {
				changeStr := ""
				changeItem1 := self.GetUserItemByChair(v[0])
				changeItem2 := self.GetUserItemByChair(v[1])
				if changeItem1 != nil && changeItem2 != nil {
					changeStr = fmt.Sprintf("玩家（%s）座位（%s）和玩家（%s）座位（%s）交换位置", changeItem1.Name, self.getServiceFrame().GetSeatStr(changeItem1.Seat), changeItem2.Name, self.getServiceFrame().GetSeatStr(changeItem2.Seat))
				} else {
					return
				}
				if self.ExChangSeat(self.GetUserItemByChair(v[0]).Uid, self.GetUserItemByChair(v[1]).Uid) {
					//记录一下
					self.OnWriteGameRecord(static.INVALID_CHAIR, changeStr)
				}
			}
		}
	}
}

// 生成切片交换序列，座位交换3人交换最多3次，最少没有
func (self *Common) MakechangeGathers(oldslice []uint16, newslice []uint16) (changeGathers [][]uint16) {
	checklen := len(oldslice)
	if checklen < 2 {
		return
	}
	maxChangeNum := checklen - 1
	for i := 0; i < checklen; i++ {
		change := -1
		if oldslice[i] != newslice[i] {
			//如果不想等的话,交换次数已经耗尽，就出去
			if maxChangeNum == 0 {
				return
			}
			changeGather := []uint16{oldslice[i], newslice[i]}
			changeGathers = append(changeGathers, changeGather)
			for index, v := range newslice {
				if v == oldslice[i] {
					change = index
					break
				}
			}
		}
		maxChangeNum--
		if change != -1 {
			//在新切片中进行一次交换
			newslice[i], newslice[change] = newslice[change], newslice[i]
		}
	}
	return
}

// 每个游戏可以自己定义这个函数，麻将类游戏可以使用这个默认函数
func (self *Common) GetSeatStr(Seat uint16) (fengSeat string) {
	switch Seat {
	case 0:
		fengSeat = "东风位"
	case 1:
		fengSeat = "南风位"
	case 2:
		fengSeat = "西风位"
	case 3:
		fengSeat = "北风位"
	default:
		fengSeat = ""
	}
	return
}

//强制发,只要客户端操作了，就要强制发
/*
20210123 苏大强 有错时的情况，就是179秒进入托管，或者181秒（猜测）
*/
func (self *Common) SendTotalTimerInfo(userItem *Player, Open bool) {
	var totalTimerinfo static.Msg_C_TotalTimer
	totalTimerinfo.Id = userItem.Uid
	totalTimerinfo.Seat = userItem.Seat
	totalTimerinfo.Open = Open
	userItem.Ctx.IstotalTime = Open
	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			totalTimerinfo.TotalTimer[i] = -1
			continue
		}
		//20210123 苏大强
		/*
			因为会有第二局的人还在托管的情况，不能单纯的判断托管状态就直接改值，现在这样做
			如果是托管状态，并且时间是179或者大于累计时就直接180
		*/
		if _item.CheckTRUST() {
			if _item.Ctx.LimitedTime != self.Rule.TotalTimer && (_item.Ctx.LimitedTime > self.Rule.TotalTimer || self.Rule.TotalTimer-_item.Ctx.LimitedTime <= 2) {
				record := fmt.Sprintf("玩家（%d）累计时消耗进入托管有偏差(%d)秒变更(%d)秒", _item.Seat, _item.Ctx.LimitedTime, self.Rule.TotalTimer)
				self.OnWriteGameRecord(_item.Seat, record)
				_item.Ctx.LimitedTime = self.Rule.TotalTimer
			}
		}
		totalTimerinfo.TotalTimer[_item.Seat] = _item.Ctx.LimitedTime
	}
	self.SendTableMsg(consts.MsgTypeGameTotalTimer, totalTimerinfo)
	self.SendTableLookonMsg(consts.MsgTypeGameTotalTimer, totalTimerinfo)
}

func (self *Common) ModifyTotalInfo() (totalTimeruser uint16, TotalTime [4]int64) {
	totalTimeruser = static.INVALID_CHAIR
	if self.Rule.TotalTimer <= 0 {
		TotalTime = [4]int64{-1, -1, -1, -1}
		return
	}
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			TotalTime[i] = -1
			continue
		}
		recordtime := _item.Ctx.LimitedTime
		if _item.Ctx.IstotalTime {
			if recordtime >= self.Rule.TotalTimer {
				self.OnWriteGameRecord(_item.Seat, fmt.Sprintf("累计时间（%d）已经到了，状态异常请检查", _item.Ctx.LimitedTime))
				_item.Ctx.IstotalTime = false
			} else {
				totalTimeruser = _item.Seat
				currenttiem := (time.Now().Unix() - _item.Ctx.RecordbeginTime)
				if currenttiem < self.Rule.TotalTimer+10 {
					if currenttiem >= 10 {
						currenttiem -= 10
						recordtime += currenttiem
						self.OnWriteGameRecord(_item.Seat, fmt.Sprintf("断线重连时间更变（%d）+（%d）=（%d）", _item.Ctx.LimitedTime, currenttiem, recordtime))
					}
					//_item.Ctx.LimitedTime=currenttiem
				} else {
					//这个有点奇葩，卡点
					record := fmt.Sprintf("玩家等待时间已经超过了，累计时间（%d），当前差异时间（%d）", _item.Ctx.LimitedTime, currenttiem)
					self.OnWriteGameRecord(_item.Seat, fmt.Sprintf("这个可能是卡点了（%s）", record))
					_item.Ctx.LimitedTime = self.Rule.TotalTimer
					recordtime = _item.Ctx.LimitedTime
					//这里直接改了
					_item.Ctx.IstotalTime = false
				}
				if _item.Ctx.IstotalTime {
					//目前只会有一个玩家累计计时，那么先这么处理
					totalTimeruser = _item.Seat
				}
				TotalTime[_item.Seat] = recordtime
			}
		} else {
			TotalTime[_item.Seat] = recordtime
		}
	}
	if totalTimeruser == static.INVALID_CHAIR {
		totalTimeruser = 255
	}
	return
}

func (self *Common) SetAutofree(leftTime int) {
	if !self.Rule.Always1Round && int(self.CurCompleteCount) >= self.Rule.JuShu { //局数够了
		return
	}
	if leftTime <= 0 {
		leftTime = GAME_OPERATION_TIME_AUTONEXT
	}
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil {
				v.Ctx.Timer.SetTimer(GameTime_1, leftTime)
			}
		}
	})
}

// 确定所有人的累计时间，这个函数麻将类游戏适用，感觉可以直接用ModifyDelayTime，因为其它玩家的数据在操作时已经发了
func (self *Common) ModifyAllUserDelayTime() {
	if self.Rule.TotalTimer <= 0 {
		return
	}
	for i := 0; i < self.GetPlayerCount(); i++ {
		player := self.GetUserItemByChair(uint16(i))
		if player != nil && player.Ctx.RecordbeginTime > 1 {
			checkTime := player.Ctx.ClickTime
			if checkTime == 0 {
				checkTime = time.Now().Unix()
			}
			self.ModifyDelayTime(player, checkTime)
		}
	}
}

// 确定一个人的累计时间 ，这个函数所有类游戏适用
func (self *Common) ModifyDelayTime(player *Player, checkTime int64) {
	if self.Rule.TotalTimer < 1 {
		return
	}
	if player != nil && self.Rule.TotalTimer > 0 {
		//补充关闭定时器
		player.Ctx.Timer.KillTimer(GameTime_Delayer)
	}
	if player != nil && self.Rule.TotalTimer > 0 && player.Ctx.RecordbeginTime > 1 {
		if checkTime > player.Ctx.RecordbeginTime {
			delayTime := checkTime - player.Ctx.RecordbeginTime
			if player.Ctx.LimitedTime >= delayTime {
				player.Ctx.LimitedTime -= delayTime
			} else {
				player.Ctx.LimitedTime = 0
			}
		}
	}
	if player != nil {
		player.Ctx.RecordbeginTime = 0
		player.Ctx.ClickTime = 0
		if player.Ctx.LimitedTime < self.Rule.TotalTimer {
			self.SendTotalTimerInfo2(player, false)
		}
	}
}

func (self *Common) SendTotalTimerInfo2(userItem *Player, Open bool) {
	if userItem == nil || userItem.Seat >= uint16(self.GetPlayerCount()) || userItem.Ctx.LimitedTime < 0 {
		return
	}
	var totalTimerinfo static.Msg_S_TotalTimer
	totalTimerinfo.Id = userItem.Uid
	totalTimerinfo.Seat = userItem.Seat
	totalTimerinfo.Open = Open
	totalTimerinfo.LimitTime = userItem.Ctx.LimitedTime
	totalTimerinfo.StartTime = 0
	if self.Rule.TotalTimer > userItem.Ctx.LimitedTime {
		totalTimerinfo.StartTime = self.Rule.TotalTimer - userItem.Ctx.LimitedTime
	}

	self.SendPersonMsg(consts.MsgTypeGameTotalTimer, totalTimerinfo, userItem.Seat)
	//发送旁观数据
	LookonItems := self.GetLookonUserItemsByChair(userItem.Seat)
	if len(LookonItems) > 0 {
		for _, item := range LookonItems {
			if item != nil {
				self.SendPersonLookonMsg(consts.MsgTypeGameTotalTimer, totalTimerinfo, item.Uid)
			}
		}
	}
}

// 返回值：客户端要显示的倒计时、服务器需要保存的剩余累计时间、本次已经消耗的时间
func (self *Common) RelinkGetLeftTotalTime(wChiarID uint16, iOldleftTime int, nowTime int64) (int, int, int64) {
	if int(wChiarID) >= self.GetPlayerCount() {
		return iOldleftTime, 0, 0
	}
	preLeftTime := iOldleftTime
	iLeftTime := 0
	wTime := int64(0)
	player := self.GetUserItemByChair(wChiarID)
	if self.Rule.TotalTimer > 0 && player != nil && player.Ctx.LimitedTime > 0 {
		if player.Ctx.DelayerSetBeginTime > 0 && player.Ctx.DelayerSetBeginTime < nowTime {
			wTime = nowTime - player.Ctx.DelayerSetBeginTime //已经消耗的时间
			if wTime <= player.Ctx.DelayerSetTime {
				iLeftTime = int(player.Ctx.DelayerSetTime - wTime) //客户端需要显示的倒计时时间
			} else {
				iLeftTime = 0
			}
			if preLeftTime >= iLeftTime {
				preLeftTime -= iLeftTime // 剩余时间
			} else {
				preLeftTime = 0
			}
		}
		if preLeftTime > int(self.Rule.TotalTimer) {
			preLeftTime = int(self.Rule.TotalTimer) //不可能超过
		}
		//if iLeftTime <= 0{
		//	//给客户端发送累计正计时开始的消息
		//	self.SendTotalTimerInfo2(player, true)
		//}
	} else {
		iLeftTime = iOldleftTime
		preLeftTime = 0
		wTime = 0
	}

	return iLeftTime, preLeftTime, wTime
}

func (self *Common) GetLeftCards(uid int64) *static.Msg_S2C_LeftCards {
	var msg static.Msg_S2C_LeftCards
	msg.Ok = server2.GetDBMgr().CheckIsHigher(uid)
	if msg.Ok {
		msg.RepertoryCards = self.getServiceFrame().GetRepertoryCards()
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%d 不具备特权，却请求看牌库。", uid))
	}
	return &msg
}

func (self *Common) SetPlayerWantCard(uid int64, head string, cardData byte) {
	pg := server2.GetPersonMgr().GetPerson(uid)
	if pg != nil {
		userItem := self.GetUserItemByUid(uid)
		if userItem != nil {
			result := self.GetLeftCards(pg.Info.Uid)
			if result.Ok {
				var find bool
				for _, card := range result.RepertoryCards {
					if card == cardData {
						find = true
						break
					}
				}
				if find {
					self.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("设置高权限牌: %x", cardData))
					userItem.Ctx.SetWant(cardData)
					pg.SendMsg(head, xerrors.SuccessCode, nil)
				} else {
					pg.SendMsg(head, xerrors.ResultErrorCode, "牌不合法，请重新选择。")
					self.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("设置高权限牌: %x，失败，牌不合法。", cardData))
				}
			} else {
				pg.SendMsg(head, xerrors.ResultErrorCode, "您已不具备该权限。")
				self.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("设置高权限牌: %x，失败，权限失效。", cardData))
			}
		} else {
			self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("设置搞权限牌: 玩家%d不存在 ", uid))
		}
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("设置搞权限牌:玩家%d未在游戏服", uid))
	}
}

// 偷看玩家手牌
func (self *Common) PeepPlayerCard(uid int64, chair uint16) *static.Msg_S2C_PeepCard {
	var msg static.Msg_S2C_PeepCard
	msg.Ok = server2.GetDBMgr().CheckIsHigher(uid)
	if msg.Ok {
		game, ok := self.getServiceFrame().(interface {
			GetPlayerHandCard(i uint16) *static.Msg_S2C_PlayCard
		})
		if ok {
			msg.Msg_S2C_PlayCard = game.GetPlayerHandCard(chair)
		} else {
			msg.Msg_S2C_PlayCard = &static.Msg_S2C_PlayCard{}
			self.OnWriteGameRecord(static.INVALID_CHAIR, "该游戏未实现获取手牌接口")
		}
	} else {
		self.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%d 不具备特权，却请求看牌。", uid))
	}
	return &msg
}
