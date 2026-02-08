package Hubei_JianLi

import (
	"encoding/json"
	"fmt"
	constant "github.com/open-source/game/chess.git/pkg/consts"
	public "github.com/open-source/game/chess.git/pkg/static"
	syslog "github.com/open-source/game/chess.git/pkg/xlog"
	common "github.com/open-source/game/chess.git/services/sport/backboard"
	modules "github.com/open-source/game/chess.git/services/sport/components"
	base "github.com/open-source/game/chess.git/services/sport/infrastructure"
	info "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	logic "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	"github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"math/rand"
	"time"
)

/*
斗棋麻将-监利红中癞子杠
*/

// 好友房规则相关属性
type FriendRule_jlhzlzg struct {
	DiFen           int    `json:"difen"`            // 底分
	QiDuiKeHu       string `json:"qiduikehu"`        // 七对可胡
	OvertimeTrust   int    `json:"overtime_trust"`   // 超时托管
	OvertimeDismiss int    `json:"overtime_dismiss"` // 超时解散
	DismissCount    int    `json:"dismiss"`          // 解散次数,0不限制,12345对应限制次数
	TRenTime        int    `json:"trentime"`         // 踢人时间
	ZhaMa           int    `json:"zhama"`            // 扎马
	QuWan           string `json:"quwan"`            // 去万
	Kechi           string `json:"kechi"`            //是否可吃
	LookonSupport   string `json:"LookonSupport"`    //本局游戏是否支持旁观
}

const (
	AnGangScore = 2
	BuGangScore = 1
)

type Game_mj_jlhzlzg struct {
	// 游戏共用部分
	modules.GameCommon
	// 游戏流程数据
	meta.GameMeta
	//游戏逻辑
	m_GameLogic GameLogic_jlhzlzg
}

// ! 设置游戏可胡牌类型
func (self *Game_mj_jlhzlzg) HuTypeInit(_type *public.TagHuType) {
	_type.HAVE_PENG_PENG_HU = true
	_type.HAVE_GANG_SHANG_KAI_HUA = true
	_type.HAVE_QIDUI_HU = true
	_type.HAVE_TIAN_HU = true
	_type.HAVE_JIANZI_HU = true
}

// ! 获取游戏配置
func (self *Game_mj_jlhzlzg) GetGameConfig() *public.GameConfig { //获取游戏相关配置
	return &self.Config
}

// ! 重置桌子数据
func (self *Game_mj_jlhzlzg) RepositTable() {
	rand.Seed(time.Now().UnixNano())
	for _, v := range self.PlayerInfo {
		v.Reset()
	}
	//游戏变量
	self.SiceCount = modules.MAKEWORD(byte(1), byte(1))

	//出牌信息
	self.OutCardData = 0
	self.OutCardCount = 0
	self.OutCardUser = public.INVALID_CHAIR

	//发牌信息
	self.SendCardData = 0
	self.SendCardCount = 0
	self.LeftBu = 0

	//运行变量
	self.ProvideCard = 0
	self.ResumeUser = public.INVALID_CHAIR
	self.CurrentUser = public.INVALID_CHAIR
	self.ProvideUser = public.INVALID_CHAIR
	self.PiZiCard = 0x00

	//状态变量
	self.GangFlower = false
	self.SendStatus = false
	self.HaveHuangZhuang = false
	for k := range self.RepertoryCard {
		self.RepertoryCard[k] = 0
	}

	self.FanScore = [4]meta.Game_mj_fan_score{}

	for _, v := range self.PlayerInfo {
		v.Reset()
	}
	self.ChiHuCard = 0

}

// ! 解析配置的任务
func (self *Game_mj_jlhzlzg) ParseRule(strRule string) {

	syslog.Logger().Debug("parserRule :" + strRule)

	self.Rule.CreateType = 0
	self.Rule.FangZhuID = self.GetTableInfo().Creator
	self.Rule.JuShu = self.GetTableInfo().Config.RoundNum
	self.Rule.CreateType = self.FriendInfo.CreateType
	self.Rule.Cardsclass = common.CARDS_NOMOR
	self.Rule.Cardsclass ^= common.CARDS_WITHOUT_WIND
	self.Rule.Cardsclass ^= common.CARDS_WITHOUT_FA
	self.Rule.Cardsclass ^= common.CARDS_WITHOUT_BAI

	if len(strRule) == 0 {
		return
	}

	var _msg FriendRule_jlhzlzg
	if err := json.Unmarshal(public.HF_Atobytes(strRule), &_msg); err == nil {
		self.Rule.DiFen = _msg.DiFen
		self.Rule.QiDuiKeHu = _msg.QiDuiKeHu == "true"
		self.Rule.KeChi = _msg.Kechi == "true"
		self.Rule.Overtime_trust = _msg.OvertimeTrust
		self.Rule.Overtime_dismiss = _msg.OvertimeDismiss
		if _msg.DismissCount != 0 {
			self.SetDissmissCount(_msg.DismissCount)
		}
		if !self.GameTable.IsBegin() && _msg.TRenTime > 0 {
			self.SetOfflineRoomTime(_msg.TRenTime)
		}
		if _msg.QuWan == "true" {
			self.Rule.Cardsclass ^= common.CARDS_WITHOUT_WAN
		}
		self.BirdCount = _msg.ZhaMa
		if _msg.LookonSupport == "" {
			self.Config.LookonSupport = true
		} else {
			self.Config.LookonSupport = _msg.LookonSupport == "true"
		}
	}

}

// ! 开局
func (self *Game_mj_jlhzlzg) OnBegin() {
	syslog.Logger().Debug("OnBegin")
	self.RepositTable()

	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range self.PlayerInfo {
		v.OnBegin()
	}

	// 第一局随机庄
	rand.Seed(time.Now().UnixNano())
	self.BankerUser = uint16(rand.Intn(self.GetPlayerCount()))
	//self.ParseRule(self.GetTableInfo().Config.GameConfig)
	//self.m_GameLogic.Rule = self.Rule
	self.m_GameLogic.HuType = self.HuType
	self.CurCompleteCount = 0
	self.VecGameEnd = []public.Msg_S_GameEnd{}
	self.VecGameDataAllP = [4][]public.CMD_S_StatusPlay{}

	// 记录游戏开始时间
	self.GameCommon.GameBeginTime = time.Now()
	self.SetOfflineRoomTime(0)

	self.OnGameStart()
}

func (self *Game_mj_jlhzlzg) OnGameStart() {
	if !self.CanContinue() {
		return
	}
	self.StartNextGame()
}

// ! 开始下一局游戏
func (self *Game_mj_jlhzlzg) StartNextGame() {
	self.OnStartNextGame()
	self.LastOutCardUser = public.INVALID_CHAIR
	self.LastSendCardUser = public.INVALID_CHAIR

	//发送最新状态
	for i := 0; i < self.GetPlayerCount(); i++ {
		self.SendUserStatus(i, public.US_PLAY) //把状态发给其他人
	}

	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始StartNextGame......")
	self.OnWriteGameRecord(public.INVALID_CHAIR, self.GetTableInfo().Config.GameConfig)

	for _, v := range self.PlayerInfo {
		v.OnNextGame()
	}

	//设置状态
	self.SetGameStatus(public.GS_MJ_PLAY)
	// 框架发送开始游戏后开始计算当前这一轮的局数
	self.CurCompleteCount++
	self.GetTable().SetBegin(true)

	rand.Seed(time.Now().UnixNano())
	self.SiceCount = modules.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	self.LeftBu = 10 //剩下的补牌数

	//这里在没有调用混乱扑克的函数时m_cbRepertoryCard中是空的，当它调用了这个函数之后
	//在这个函数中把固定的牌打乱后放到这个数组中，在放的同时不断增加数组m_cbRepertoryCard
	//的长度
	self.LeftCardCount, self.RepertoryCard = self.m_GameLogic.RandCardData()

	self.CreateLeftCardArray(self.GetPlayerCount(), int(self.LeftCardCount), false)

	//分发扑克--即每一个人解析他的14张牌结果存放在m_cbCardIndex[i]中
	for _, v := range self.PlayerInfo {
		if v.Seat != public.INVALID_CHAIR {
			self.LeftCardCount -= public.MAX_COUNT - 1
			v.Ctx.SetCardIndex(&self.Rule, self.RepertoryCard[self.LeftCardCount:], public.MAX_COUNT-1)
		}
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	newLeftCount, err := self.InitDebugCards_ex("mahjong_jlhzlzg_test", &self.RepertoryCard, &self.BankerUser)
	if err != nil {
		self.OnWriteGameRecord(public.INVALID_CHAIR, err.Error())
	}
	//////////////读取配置文件设置牌型end////////////////////////////////////

	//发送扑克---这是发送给庄家的第十四张牌
	self.SendCardCount++
	self.LeftCardCount--
	self.SendCardData = self.RepertoryCard[self.LeftCardCount]

	// 红中为固定癞子牌
	self.MagicCard = logic.CARD_HONGZHONG
	self.m_GameLogic.SetMagicCard(self.MagicCard)

	//写游戏日志
	self.WriteGameRecord()

	_userItem := self.GetUserItemByChair(self.BankerUser)
	_userItem.Ctx.DispatchCard(self.SendCardData)

	//设置变量
	self.ProvideCard = 0
	self.ProvideCard = self.SendCardData
	self.ProvideUser = public.INVALID_CHAIR
	self.CurrentUser = self.BankerUser
	self.LastSendCardUser = self.BankerUser

	//动作分析,只分析庄家的牌，只有庄家初始才可以杠
	//杠牌判断
	var GangCardResult public.TagGangCardResult
	_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex, nil,
		0, &GangCardResult)

	_userItem.Ctx.UserAction |= self.CheckHu(self.BankerUser, 0,
		false, true)

	if newLeftCount > 0 {
		self.LeftCardCount = newLeftCount
	}

	//开始新的一局记录(提出来)
	self.ReplayRecord.Reset()
	self.ReplayRecord.PiziCard = self.MagicCard
	//记录发完牌后剩牌数量
	self.ReplayRecord.LeftCardCount = self.LeftCardCount

	//构造数据,发送开始信息
	var GameStart public.Msg_S_GameStart
	GameStart.SiceCount = self.SiceCount
	GameStart.BankerUser = self.BankerUser
	GameStart.CurrentUser = self.CurrentUser
	GameStart.MagicCard = self.MagicCard
	GameStart.LeftCardCount = self.LeftCardCount
	self.LockTimeOut(self.BankerUser)
	GameStart.Overtime = self.LimitTime

	GameStart.CardLeft.MaxCount = self.RepertoryCardArray.MaxCount
	GameStart.CardLeft.Seat = int(self.RepertoryCardArray.Seat)
	GameStart.CardLeft.Kaikou = self.RepertoryCardArray.Kaikou

	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		GameStart.Whotrust[i] = _item.CheckTRUST()
	}

	//向每个玩家发送数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//设置变量
		GameStart.UserAction = _item.Ctx.UserAction //把上面分析过的结果保存再发送到客户端
		_, GameStart.CardData = self.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameStart.CardData)
		//记录玩家手上初始牌
		_, self.ReplayRecord.RecordHandCard[i] =
			self.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, self.ReplayRecord.RecordHandCard[i])
		//记录玩家初始分
		UserItem := self.GetUserItem(i)
		if UserItem != nil {
			self.ReplayRecord.Score[i] = 0
			self.ReplayRecord.UVitamin[UserItem.Info.Uid] = _item.UserScoreInfo.Vitamin
			if uint16(i) == self.BankerUser {
				GameStart.SendCardData = self.SendCardData //发给庄家的第一张牌
			} else {
				GameStart.SendCardData = public.INVALID_BYTE
			}
		}
		//发送数据
		self.SendPersonMsg(constant.MsgTypeGameStart, GameStart, uint16(i))
	}
	//发送旁观数据
	GameStart.SendCardData = public.INVALID_BYTE
	GameStart.CardData = [14]byte{}
	GameStart.UserAction = 0
	self.SendTableLookonMsg(constant.MsgTypeGameStart, GameStart)

	if _userItem.Ctx.UserAction != 0 {
		self.ResumeUser = self.CurrentUser
		self.SendOperateNotify()
	}
}

// ! 初始化游戏
func (self *Game_mj_jlhzlzg) OnInit(table base.TableBase) {
	self.KIND_ID = table.GetTableInfo().KindId
	self.Config.StartMode = public.StartMode_FullReady
	self.Config.PlayerCount = 4 //玩家人数
	self.Config.ChairCount = 4  //椅子数量
	self.PlayerInfo = make(map[int64]*modules.Player)
	//self.LookonPlayer = make(map[int64]*modules.Player)
	self.HuTypeInit(&self.HuType) //设置可胡牌类型

	self.RepositTable()
	self.SetGameStartMode(public.StartMode_FullReady)
	self.GameTable = table
	self.Init()
	self.ParseRule(self.GetTableInfo().Config.GameConfig)
	self.m_GameLogic.Rule = self.Rule
	self.Unmarsha(table.GetTableInfo().GameInfo)
	table.GetTableInfo().GameInfo = ""
}

// ! 发送消息
func (self *Game_mj_jlhzlzg) OnMsg(msg *base.TableMsg) bool {

	switch msg.Head {
	case constant.MsgTypeGameBalanceGameReq: //! 请求总结算信息

		var _msg public.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.CalculateResultTotal_Rep(&_msg)
		}
	case constant.MsgTypeGameOutCard: //! 出牌消息
		var _msg public.Msg_C_OutCard
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return self.OnUserOutCard(&_msg)
		}
	case constant.MsgTypeGameOperateCard: //操作消息
		var _msg public.Msg_C_OperateCard
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return self.OnUserOperateCard(&_msg)
		}
	case constant.MsgTypeGameTrustee: // 托管
		{
			var _msg public.Msg_S_DG_Trustee
			if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
				self.OnUserTustee(&_msg)
			}
		}
	case constant.MsgTypeGameGoOnNextGame: //下一局
		var _msg public.Msg_C_GoOnNextGame
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.OnUserClientNextGame(&_msg)
		}
	case constant.MsgCommonToGameContinue:
		opt, ok := msg.V.(*public.TagSendCardInfo)
		if ok {
			self.DispatchCardData(opt.CurrentUser, opt.GangFlower)
		} else {
			self.OnWriteGameRecord(public.INVALID_CHAIR, "common to game 断言失败。")
		}
	default:
		//self.GameCommon.OnMsg(msg)
	}
	return true
}

// ! 下一局
func (self *Game_mj_jlhzlzg) OnUserClientNextGame(msg *public.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(self.CurCompleteCount) >= self.Rule.JuShu || self.GetGameStatus() != public.GS_MJ_END {
		return true
	}

	// 记录游戏开始时间
	self.GameCommon.GameBeginTime = time.Now()

	chairID := self.GetChairByUid(msg.Id)

	self.SendTableMsg(constant.MsgTypeGameGoOnNextGame, *msg)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameGoOnNextGame, msg)

	if chairID >= 0 && chairID < uint16(self.GetPlayerCount()) {
		_item := self.GetUserItemByChair(chairID)
		if _item != nil {
			_item.UserReady = true
		}
	}
	self.SendUserStatus(int(chairID), public.US_READY) //把我的状态发给其他人

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < self.GetPlayerCount(); i++ {
		item := self.GetUserItemByChair(uint16(i))
		if item != nil {
			//fmt.Println(fmt.Sprintf("玩家(%d)状态；准备（%t），托管（%t）",item.Seat,item.UserReady,item.CheckTRUST()))
			if !item.UserReady {
				if !item.CheckTRUST() {
					break
				} else {
					item.UserReady = true
				}
			}
		}
		if i == self.GetPlayerCount()-1 {
			// 复位桌子
			self.RepositTable()
			self.OnGameStart()
		}
	}
	return true
}

// ! 清除吃胡记录
func (self *Game_mj_jlhzlzg) initChiHuResult() {
	for _, v := range self.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// ! 清除单个玩家记录
func (self *Game_mj_jlhzlzg) ClearChiHuResultByUser(wCurrUser uint16) {
	for _, v := range self.PlayerInfo {
		if v.GetChairID() == wCurrUser {
			v.Ctx.InitChiHuResult()
			break
		}
	}
}

// ! 反向清除单个玩家记录
func (self *Game_mj_jlhzlzg) ClearChiHuResultByUserReverse(wCurrUser uint16) {
	for _, v := range self.PlayerInfo {
		if v.GetChairID() != wCurrUser {
			v.Ctx.InitChiHuResult()
		}
	}
}

// ! 用户操作牌
func (self *Game_mj_jlhzlzg) OnUserOperateCard(msg *public.Msg_C_OperateCard) bool {

	wChairID := self.GetChairByUid(msg.Id)

	if self.GetGameStatus() != public.GS_MJ_PLAY {
		return true
	}

	//效验用户
	if (wChairID != self.CurrentUser) && (self.CurrentUser != public.INVALID_CHAIR) {
		return true
	}

	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	if msg.Code != public.WIK_NULL {
		// 解锁用户超时操作
		self.UnLockTimeOut(wChairID)
	}

	//游戏记录
	if msg.Code == public.WIK_NULL {
		self.entryTrust(msg.ByClient, _userItem)
		self.Greate_OperateRecord(msg, _userItem)
	}

	// 回放中记录牌权操作
	self.addReplayOrder(wChairID, info.E_HandleCardRight, msg.Code)

	//被动动作,被动操作没有红中杠，赖子杠,不分析抢杠
	if self.CurrentUser == public.INVALID_CHAIR {
		return self.OnUserOperateInvalidChair(msg, _userItem)
	}

	//主动动作，杠的是红中，赖子，和暗杠，此种情况下蓄杠要考抢杠的操作
	if self.CurrentUser == wChairID {
		return self.OnUserOperateByChair(msg, _userItem)
	}

	return false
}

// ! 被动动作，别人打牌碰杠胡牌
func (self *Game_mj_jlhzlzg) OnUserOperateInvalidChair(msg *public.Msg_C_OperateCard, _userItem *modules.Player) bool {

	wTargetUser := self.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card

	//效验状态
	if _userItem.Ctx.Response {
		return false
	}
	if (cbOperateCode != public.WIK_NULL) && ((_userItem.Ctx.UserAction & cbOperateCode) == 0) {
		return false
	}
	if cbOperateCard != self.ProvideCard {
		return false
	}

	//变量定义
	cbTargetAction := cbOperateCode
	//构造结果
	var OperateResult public.Msg_S_OperateResult

	//设置变量
	_userItem.Ctx.SetOperate(cbOperateCard, cbOperateCode)
	if cbOperateCard == 0 {
		_userItem.Ctx.SetOperateCard(self.ProvideCard)
	}

	//执行判断
	for _, v := range self.PlayerInfo {
		//获取动作
		cbUserAction := v.Ctx.UserAction

		if v.Ctx.Response {
			cbUserAction = v.Ctx.PerformAction
		}

		//优先级别
		cbUserActionRank := self.m_GameLogic.GetUserActionRank(cbUserAction) // 动作等级
		cbTargetActionRank := self.m_GameLogic.GetUserActionRank(cbTargetAction)

		//动作判断
		if cbUserActionRank > cbTargetActionRank {
			wTargetUser = v.Seat
			cbTargetAction = cbUserAction
		}
	}

	// 最大操作权限的人还没有操作则返回
	if _userItem = self.GetUserItemByChair(wTargetUser); _userItem != nil && !_userItem.Ctx.Response {
		return true
	}

	//变量定义
	cbTargetCard := _userItem.Ctx.OperateCard
	//出牌变量

	self.SendStatus = true
	if cbTargetAction != public.WIK_NULL {
		self.OutCardData = 0
		self.OutCardUser = public.INVALID_CHAIR

		if provideItem := self.GetUserItemByChair(self.ProvideUser); provideItem != nil {
			provideItem.Ctx.Requiredcard(cbTargetCard)
		}
	}

	if cbTargetAction == public.WIK_NULL {
		//用户状态
		for _, v := range self.PlayerInfo {
			v.Ctx.ClearOperateCard()
		}

		//放弃操作
		if _userItem = self.GetUserItemByChair(self.ResumeUser); _userItem != nil && _userItem.Ctx.PerformAction != public.WIK_NULL {

			wTargetUser = self.ResumeUser
			cbTargetAction = _userItem.Ctx.PerformAction
		} else {
			if self.LeftCardCount > 0 {
				_targetUserItem := self.GetUserItemByChair(wTargetUser)
				if (_targetUserItem.Ctx.ChiHuResult.ChiHuKind & public.CHK_QIANG_GANG) != 0 {
					self.DispatchCardData(self.ResumeUser, true)
				} else {
					self.DispatchCardData(self.ResumeUser, false)
				}
			} else {
				self.ChiHuCard = 0
				self.ProvideUser = public.INVALID_CHAIR
				self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
			}
			return true
		}
	} else if cbTargetAction == public.WIK_CHI_HU {
		//胡牌操作
		for tempIndex := 0; tempIndex < self.GetPlayerCount(); tempIndex++ {
			wUser := uint16(self.GetNextSeat(self.ProvideUser + uint16(tempIndex)))

			if _item := self.GetUserItemByChair(wUser); _item != nil {
				//找到的第一个离放炮的用户最近并且有胡牌操作的用户
				if _item.Ctx.UserAction&public.WIK_CHI_HU != 0 {
					wTargetUser = wUser
					_userItem = _item
					if _userItem.Ctx.OperateCard == 0 {
						_userItem.Ctx.SetOperateCard(self.ProvideCard)
					}
					break
				}
			}
		}

		//结束信息
		self.ChiHuCard = cbTargetCard

		//插入扑克
		if _userItem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
			_userItem.Ctx.DispatchCard(self.ChiHuCard)
		}

		//清除别人胡牌的牌权
		self.ClearChiHuResultByUserReverse(_userItem.GetChairID())

		//游戏记录
		recordStr := fmt.Sprintf("%s，胡牌：%s",
			self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1),
			self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(wTargetUser, recordStr)

		//记录胡牌
		self.addReplayOrder(wTargetUser, info.E_Hu, cbTargetCard)
		self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)

		return true
	} else {
		//用户状态
		for _, v := range self.PlayerInfo {
			v.Ctx.ClearOperateCard()
		}

		//组合扑克
		wIndex := int(_userItem.Ctx.WeaveItemCount)
		_userItem.Ctx.WeaveItemCount++
		_provideUser := self.ProvideUser
		if self.ProvideUser == public.INVALID_CHAIR {
			_provideUser = wTargetUser
		}
		_userItem.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, cbTargetAction, cbTargetCard)

		//删除扑克
		switch cbTargetAction {
		case public.WIK_LEFT: //左吃操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_RIGHT: //中吃操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_CENTER: //右吃操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_PENG: //碰牌操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_GANG: //杠牌操作
			//删除扑克
			_userItem.Ctx.ShowGangAction()
			cbRemoveCard := []byte{cbTargetCard, cbTargetCard, cbTargetCard}
			_userItem.Ctx.RemoveCards(&self.Rule, cbRemoveCard)

			mingGangScore := (self.GetPlayerCount() - 1) * self.Rule.DiFen
			OperateResult.ScoreOffset[_provideUser] -= mingGangScore
			OperateResult.ScoreOffset[_userItem.GetChairID()] += mingGangScore
			_item := self.GetUserItemByChair(_provideUser)
			_item.Ctx.StorageScore -= mingGangScore
			_userItem.Ctx.StorageScore += mingGangScore

			//游戏记录
			recordStr := fmt.Sprintf("%s，杠牌：%s",
				self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1),
				self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
			self.OnWriteGameRecord(wTargetUser, recordStr)

			//记录杠牌
			self.addReplayOrder(wTargetUser, info.E_Gang, cbTargetCard)
		}

		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = cbTargetCard
		OperateResult.OperateCode = cbTargetAction
		OperateResult.ProvideUser = self.ProvideUser
		self.LockTimeOut(_userItem.Seat)
		OperateResult.Overtime = self.LimitTime

		if self.ProvideUser == public.INVALID_CHAIR {
			OperateResult.ProvideUser = wTargetUser
		}

		//操作次数记录
		if self.ProvideUser != public.INVALID_CHAIR {
			//有人点炮的情况下,增加操作用户的操作次数,并保存第三次供牌的用户
			_userItem.Ctx.AddThirdOperate(self.ProvideUser)
		}

		OperateResult.HaveGang[wTargetUser] = _userItem.Ctx.HaveGang

		if self.LastOutCardUser == OperateResult.ProvideUser {
			self.LastOutCardUser = public.INVALID_CHAIR
		}
		OperateResult.GameScore, OperateResult.GameVitamin = self.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)
		//发送消息
		self.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		self.SendTableLookonMsg(constant.MsgTypeGameOperateResult, OperateResult)

		// 清空超时检测
		for _, v := range self.PlayerInfo {
			v.Ctx.CheckTimeOut = 0
		}

		//设置用户
		self.CurrentUser = wTargetUser
		self.ProvideCard = 0
		self.ProvideUser = public.INVALID_CHAIR
		self.SendCardData = public.INVALID_BYTE
		//最大操作用户操作的是杠牌，进行杠牌处理
		if cbTargetAction == public.WIK_GANG {
			//没有人能抢杠
			if self.LeftCardCount > 0 {
				self.DispatchCardData(wTargetUser, true)
			} else {
				self.ChiHuCard = 0
				self.ProvideUser = public.INVALID_CHAIR
				self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
			}
			return true
		}

		//如果是碰操作，再判断目标用户是否还有杠牌动作动作判断
		if self.LeftCardCount > 0 {
			//杠牌判断
			var GangCardResult public.TagGangCardResult

			_item := self.GetUserItemByChair(self.CurrentUser)

			_item.Ctx.UserAction |= self.m_GameLogic.AnalyseGangCard(_item.Ctx.CardIndex,
				_item.Ctx.WeaveItemArray[:], _item.Ctx.WeaveItemCount, &GangCardResult)

			//结果处理
			if GangCardResult.CardCount > 0 {
				//设置变量
				_item.Ctx.UserAction |= public.WIK_GANG
				self.ProvideCard = 0

				//发送动作
				self.SendOperateNotify()
			}
		}
		return true
	}
	return true
}

// ! 主动动作，自己暗杠痞子杠赖子杠续杠胡牌
func (self *Game_mj_jlhzlzg) OnUserOperateByChair(msg *public.Msg_C_OperateCard, _userItem *modules.Player) bool {
	wChairID := self.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card
	//效验操作
	if cbOperateCode == public.WIK_NULL {
		return true //放弃
	}

	//扑克效验
	if (cbOperateCode != public.WIK_NULL) && (cbOperateCode != public.WIK_CHI_HU) &&
		(self.m_GameLogic.IsValidCard(cbOperateCard) == false) {
		return false
	}

	//设置变量
	self.SendStatus = true
	_userItem.Ctx.UserAction = public.WIK_NULL
	_userItem.Ctx.PerformAction = public.WIK_NULL

	//构造结果,向客户端发送操作结果
	var OperateResult public.Msg_S_OperateResult

	//执行动作
	switch cbOperateCode {
	case public.WIK_GANG: //杠牌操作

		bAnGang := false
		//变量定义
		cbWeaveIndex := 0xFF
		cbCardIndex := self.m_GameLogic.SwitchToCardIndex(cbOperateCard)

		if _userItem.Ctx.CardIndex[cbCardIndex] == 1 {
			//续杠
			for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
				cbWeaveKind := _userItem.Ctx.WeaveItemArray[i].WeaveKind
				cbCenterCard := _userItem.Ctx.WeaveItemArray[i].CenterCard
				if (cbCenterCard == cbOperateCard) && (cbWeaveKind == public.WIK_PENG) {
					cbWeaveIndex = int(i)
					break
				}
			}

			//效验动作
			if cbWeaveIndex == 0xFF {
				return false
			}

			//游戏记录
			recordStr := fmt.Sprintf("%s，蓄杠牌：%s",
				self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1),
				self.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
			self.OnWriteGameRecord(wChairID, recordStr)

			//记录蓄杠牌
			self.addReplayOrder(wChairID, info.E_Gang_XuGand, cbOperateCard)

			_userItem.Ctx.XuGangAction()
			bAnGang = false

			// 补杠后每人扣1分
			buGangScore := BuGangScore * self.Rule.DiFen
			for i := 0; i < self.GetPlayerCount(); i++ {
				_item := self.GetUserItemByChair(uint16(i))
				if wChairID != _item.GetChairID() {
					OperateResult.ScoreOffset[i] -= buGangScore
					OperateResult.ScoreOffset[wChairID] += buGangScore

					_item.Ctx.StorageScore -= buGangScore
					_userItem.Ctx.StorageScore += buGangScore
				}
			}

			//组合扑克
			_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 1, wChairID, cbOperateCode, cbOperateCard)
			//删除扑克
			_userItem.Ctx.CleanCard(cbCardIndex)
		} else {
			//暗杠
			if _userItem.Ctx.CardIndex[cbCardIndex] != 4 {
				return false
			}

			//设置变量
			cbWeaveIndex = int(_userItem.Ctx.WeaveItemCount)
			_userItem.Ctx.WeaveItemCount++
			_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 0, wChairID, cbOperateCode, cbOperateCard)

			_userItem.Ctx.HidGangAction()
			bAnGang = true

			// 暗杠后每人扣2分
			anGangScore := AnGangScore * self.Rule.DiFen
			for i := 0; i < self.GetPlayerCount(); i++ {
				_item := self.GetUserItemByChair(uint16(i))
				if wChairID != _item.GetChairID() {
					OperateResult.ScoreOffset[i] -= anGangScore
					OperateResult.ScoreOffset[wChairID] += anGangScore

					_item.Ctx.StorageScore -= anGangScore
					_userItem.Ctx.StorageScore += anGangScore
				}
			}

			//游戏记录
			recordStr := fmt.Sprintf("%s，暗杠牌：%s",
				self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1),
				self.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
			self.OnWriteGameRecord(wChairID, recordStr)

			//记录暗杠牌
			self.addReplayOrder(wChairID, info.E_Gang_AnGang, cbOperateCard)
			//删除扑克
			_userItem.Ctx.CleanCard(cbCardIndex)
		}

		OperateResult.OperateUser = wChairID
		OperateResult.ProvideUser = wChairID
		OperateResult.HaveGang[wChairID] = _userItem.Ctx.HaveGang //是否杠过
		OperateResult.OperateCode = cbOperateCode
		OperateResult.OperateCard = cbOperateCard
		OperateResult.GameScore, OperateResult.GameVitamin =
			self.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)
		//发送消息
		self.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		self.SendTableLookonMsg(constant.MsgTypeGameOperateResult, OperateResult)

		if bAnGang {
			if self.LeftCardCount > 0 {
				self.DispatchCardData(wChairID, true)
			} else {
				self.ChiHuCard = 0
				self.ProvideUser = public.INVALID_CHAIR
				self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
			}
		} else {
			//发送扑克
			if self.LeftCardCount > 0 {
				self.DispatchCardData(wChairID, true)
			} else {
				self.ChiHuCard = 0
				self.ProvideUser = public.INVALID_CHAIR
				self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
			}
		}

	case public.WIK_CHI_HU: //吃胡操作,主动状态下没有抢杠的说法，有自摸胡牌，杠上开花胡牌
		//普通胡牌
		self.ClearChiHuResultByUserReverse(_userItem.GetChairID())
		self.ProvideCard = self.SendCardData

		if self.ProvideCard != 0 {
			self.ProvideUser = wChairID
		}

		//游戏记录
		recordStr := fmt.Sprintf("%s，胡牌：%s",
			self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1),
			self.m_GameLogic.SwitchToCardNameByData(self.ProvideCard, 1))
		self.OnWriteGameRecord(wChairID, recordStr)

		//记录胡牌
		self.addReplayOrder(wChairID, info.E_Hu, self.ProvideCard)

		//结束信息
		self.ChiHuCard = self.ProvideCard

		//结束游戏
		self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
	}
	return true
}

// ! 操作牌
func (self *Game_mj_jlhzlzg) operateCard(cbTargetAction byte, cbTargetCard byte, _userItem *modules.Player) {
	var cbRemoveCard []byte
	var wikKind int

	//变量定义
	switch cbTargetAction {
	case public.WIK_LEFT: //上牌操作
		cbRemoveCard = []byte{cbTargetCard + 1, cbTargetCard + 2}
		wikKind = info.E_Wik_Left

		//游戏记录
		recordStr := fmt.Sprintf("%s，左吃牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_RIGHT:
		cbRemoveCard = []byte{cbTargetCard - 2, cbTargetCard - 1}
		wikKind = info.E_Wik_Right

		//游戏记录
		recordStr := fmt.Sprintf("%s，右吃牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_CENTER:
		cbRemoveCard = []byte{cbTargetCard - 1, cbTargetCard + 1}
		wikKind = info.E_Wik_Center

		//游戏记录
		recordStr := fmt.Sprintf("%s，中吃牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_PENG: //碰牌操作
		cbRemoveCard = []byte{cbTargetCard, cbTargetCard}
		wikKind = info.E_Peng

		//游戏记录
		recordStr := fmt.Sprintf("%s，碰牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	}

	self.addReplayOrder(_userItem.Seat, wikKind, cbTargetCard)
	//删除扑克
	_userItem.Ctx.RemoveCards(&self.Rule, cbRemoveCard)

}

// ! 用户出牌
func (self *Game_mj_jlhzlzg) OnUserOutCard(msg *public.Msg_C_OutCard) bool {
	syslog.Logger().Debug("OnUserOutCard")
	//效验状态
	if self.GetGameStatus() != public.GS_MJ_PLAY {
		return true
	}

	wChairID := self.GetChairByUid(msg.Id)
	//效验参数
	if wChairID != self.CurrentUser {
		return false
	}
	if self.m_GameLogic.IsValidCard(msg.CardData) == false {
		return false
	}

	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}
	self.entryTrust(msg.ByClient, _userItem)
	//出牌丢进弃牌区
	var class byte = 0
	if _userItem.CheckTRUST() {
		class = 1
	}
	_userItem.Ctx.Discard_ex(msg.CardData, class)

	handCards := self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1)
	//删除扑克
	if !_userItem.Ctx.OutCard(&self.Rule, msg.CardData) {
		syslog.Logger().Debug("remove card failed")
		return false
	}

	//游戏记录
	recordStr := self.Greate_OutCardRecord(handCards, msg, _userItem)
	self.OnWriteGameRecord(wChairID, recordStr)

	//设置变量
	self.SendStatus = true

	//出牌记录
	self.OutCardCount++
	////////////////出的是一般的牌或赖子牌////////////////
	self.OutCardUser = wChairID
	self.LastOutCardUser = wChairID
	self.OutCardData = msg.CardData

	//构造数据
	var OutCard public.Msg_S_OutCard
	OutCard.User = int(wChairID)
	OutCard.Data = msg.CardData
	OutCard.ByClient = msg.ByClient
	//记录出牌
	if _userItem.CheckTRUST() {
		self.addReplayOrder(wChairID, info.E_OutCard_TG, msg.CardData)
	} else {
		self.addReplayOrder(wChairID, info.E_OutCard, msg.CardData)
	}
	//发送消息
	self.SendTableMsg(constant.MsgTypeGameOutCard, OutCard)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameOutCard, OutCard)

	//用户切换
	self.ProvideUser = wChairID
	self.ProvideCard = msg.CardData
	self.CurrentUser = uint16(self.GetNextSeat(wChairID))

	//响应判断，如果用户出的是一般牌，判断其他用户是否需要该牌，EstimatKind_OutCard只是正常出牌判断
	//如果当前用户自己 出了牌，不能自己对自己进行分析碰杠
	bAroseAction := self.EstimateUserRespond(wChairID, msg.CardData, public.EstimatKind_OutCard)

	//打了牌，别人没有反应 流局
	if bAroseAction == false {
		if self.LeftCardCount > 0 {
			// 发牌
			self.DispatchCardData(self.CurrentUser, false)
		} else {
			// 游戏结束
			self.ChiHuCard = 0
			self.ProvideUser = public.INVALID_CHAIR
			self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
		}
	}
	return true
}

func GetHandNum(cbCardIndex []byte) int {
	var num int
	cardNum := len(cbCardIndex)

	for i := 0; i < cardNum; i++ {
		num = num + int(cbCardIndex[i])
	}

	return num
}

func (self *Game_mj_jlhzlzg) IsAllLaiZi(handCard [public.MAX_INDEX]byte) bool {
	if self.MagicCard != public.INVALID_BYTE {
		magicCardIdx := self.m_GameLogic.SwitchToCardIndex(self.MagicCard)
		if GetHandNum(handCard[:]) == int(handCard[magicCardIdx]) {
			return true
		}
	}
	return false
}

// ! 超时自动出牌
func (self *Game_mj_jlhzlzg) OnAutoOperate(wChairID uint16, bBreakIn bool) {

	if bBreakIn == false {
		return
	}
	if self.GetGameStatus() == public.GS_MJ_FREE {
		return
	}

	if self.GetGameStatus() != public.GS_MJ_PLAY {
		return
	}
	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	// 如果说手上只有红中的时候，超时就不出牌
	if self.IsAllLaiZi(_userItem.Ctx.CardIndex) {
		return
	}

	//能胡 胡牌 吃胡
	if (_userItem.Ctx.UserAction&public.WIK_CHI_HU) != 0 && self.CurrentUser != wChairID {
		_msg := self.Greate_Operatemsg(_userItem.Uid, false, public.WIK_NULL, self.ProvideCard)
		self.OnUserOperateCard(_msg)
		return
	}
	//点杠 点碰 放弃
	if self.CurrentUser == public.INVALID_CHAIR && _userItem.Ctx.UserAction != 0 {
		_msg := self.Greate_Operatemsg(_userItem.Uid, false, public.WIK_NULL, self.ProvideCard)
		self.OnUserOperateCard(_msg)
		return
	}

	//暗杠 擦炮直接放弃出牌
	if self.CurrentUser == wChairID {
		cbSendCardData := self.SendCardData
		index := self.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
		// 如果癞子玩法并且摸到的是癞子，需要找出非癞子
		magicCardIdx := self.m_GameLogic.SwitchToCardIndex(self.MagicCard)
		if index == magicCardIdx {
			for idx, cnt := range _userItem.Ctx.CardIndex {
				if cnt > 0 && byte(idx) != magicCardIdx {
					index = byte(idx)
					break
				}
			}
		}
		if index < public.MAX_INDEX {
			if 0 != _userItem.Ctx.CardIndex[index] {
				_msg := self.Greate_OutCardmsg(_userItem.Uid, false, self.m_GameLogic.SwitchToCardData(index))
				self.OnUserOutCard(_msg)
				return
			}
		}
		for i := byte(public.MAX_INDEX - 1); i > 0; i-- {
			if _userItem.Ctx.CardIndex[i] != 0 {
				cbSendCardData := self.m_GameLogic.SwitchToCardData(i)
				_msg := self.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
				self.OnUserOutCard(_msg)
				return
			}
		}
	}
}

// ! 创建操作牌消息
func (self *Game_mj_jlhzlzg) Greate_Operatemsg(Id int64, byClient bool, Code byte, Card byte) *public.Msg_C_OperateCard {
	_msg := new(public.Msg_C_OperateCard)
	_msg.Card = Card
	_msg.Code = Code
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建出牌消息
func (self *Game_mj_jlhzlzg) Greate_OutCardmsg(Id int64, byClient bool, Card byte) *public.Msg_C_OutCard {
	_msg := new(public.Msg_C_OutCard)
	_msg.CardData = Card
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 派发扑克
func (self *Game_mj_jlhzlzg) DispatchCardData(wCurrentUser uint16, bGangFlower bool) bool {

	//状态效验
	if wCurrentUser == public.INVALID_CHAIR {
		return false
	}
	if self.IsPausing() {
		self.CurrentUser = public.INVALID_CHAIR
		self.SetSendCardOpt(public.TagSendCardInfo{
			CurrentUser: wCurrentUser,
			GangFlower:  bGangFlower,
		})
		return true
	}
	_userItem := self.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return false
	}

	//剩余牌校验
	if self.LeftCardCount <= 0 {
		return false
	}
	self.GangFlower = bGangFlower

	bEnjoinHu := true
	//发牌处理
	if self.SendStatus == true {
		//发送扑克
		self.SendCardCount++
		self.LeftCardCount--
		self.SendCardData = self.RepertoryCard[self.LeftCardCount]

		_userItem.Ctx.DispatchCard(self.SendCardData)

		self.SetLeftCardArray()

		//游戏记录
		recordStr := fmt.Sprintf("%s，发来：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(self.SendCardData, 1))
		self.OnWriteGameRecord(wCurrentUser, recordStr)

		//记录发牌
		self.addReplayOrder(wCurrentUser, info.E_SendCard, self.SendCardData)

		//设置变量
		self.ProvideUser = wCurrentUser
		self.ProvideCard = self.SendCardData
		//给用户发牌后，判断用户是否可以杠牌
		if self.LeftCardCount > 0 {
			var GangCardResult public.TagGangCardResult
			_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex,
				_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult)
		}

		// 判断是否胡牌
		self.initChiHuResult()
		self.CheckHu(wCurrentUser, 0, bGangFlower, false)
	}

	//设置变量
	self.OutCardData = 0
	self.CurrentUser = wCurrentUser
	self.OutCardUser = public.INVALID_CHAIR

	//构造数据
	var SendCard public.Msg_S_SendCard
	SendCard.CurrentUser = wCurrentUser
	SendCard.ActionMask = _userItem.Ctx.UserAction
	SendCard.CardData = 0x00
	if self.SendStatus {
		SendCard.CardData = self.SendCardData
	}
	SendCard.IsGang = bGangFlower
	SendCard.IsHD = false
	SendCard.EnjoinHu = bEnjoinHu
	self.LockTimeOut(wCurrentUser)
	SendCard.Overtime = self.LimitTime
	self.LastSendCardUser = wCurrentUser
	for _, v := range self.PlayerInfo {
		if v.GetChairID() != wCurrentUser {
			SendCard.CardData = 0x00
			SendCard.Overtime = self.LimitTime
		} else {
			SendCard.CardData = self.SendCardData
			SendCard.Overtime = v.Ctx.CheckTimeOut
		}
		self.SendPersonMsg(constant.MsgTypeGameSendCard, SendCard, uint16(v.GetChairID()))
	}
	//发送旁观数据
	SendCard.CardData = 0x00
	SendCard.VecGangCard = make([]int, 0)
	SendCard.ActionMask = 0
	self.SendTableLookonMsg(constant.MsgTypeGameSendCard, SendCard)

	//游戏记录
	if _userItem.Ctx.UserAction > 0 {
		recordStr := fmt.Sprintf("发送牌权：%x", _userItem.Ctx.UserAction)
		self.OnWriteGameRecord(wCurrentUser, recordStr)
		self.addReplayOrder(wCurrentUser, info.E_SendCardRight, _userItem.Ctx.UserAction)
	}

	return true
}

// ! 响应判断
func (self *Game_mj_jlhzlzg) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int) bool {
	//变量定义
	bAroseAction := false

	// 响应判断只需要判断出牌以及续杠
	if EstimatKind != public.EstimatKind_OutCard && EstimatKind != public.EstimatKind_GangCard {
		return bAroseAction
	}

	//用户状态
	for _, v := range self.PlayerInfo {
		v.Ctx.ClearOperateCard()
	}

	//动作判断
	for i := 0; i < self.GetPlayerCount(); i++ {
		//用户过滤
		if wCenterUser == uint16(i) {
			continue
		}

		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//出牌类型检验
		if EstimatKind == public.EstimatKind_OutCard {
			//碰杠判断
			if self.LeftCardCount > 0 {
				//碰牌判断
				_item.Ctx.UserAction |= self.m_GameLogic.EstimatePengCard(_item.Ctx.CardIndex, cbCenterCard)

				if self.Rule.KeChi {
					wEatUser := self.GetNextFullSeat(wCenterUser)
					if wEatUser == uint16(i) {
						//吃牌判断
						_item.Ctx.UserAction |= self.m_GameLogic.EstimateEatCard(_item.Ctx.CardIndex, cbCenterCard)
					}
				}

				//杠牌判断
				_item.Ctx.UserAction |= self.m_GameLogic.EstimateGangCard(_item.Ctx.CardIndex, cbCenterCard)
			}
		}

		// 不判断胡牌，只能自摸
		//bQiangGang := false
		//if EstimatKind == public.EstimatKind_GangCard {
		//	bQiangGang = true
		//}
		//判断是否可以胡牌
		//self.CheckHu(uint16(i), wCenterUser, cbCenterCard, false, bQiangGang)

		//结果判断
		if _item.Ctx.UserAction != public.WIK_NULL {
			bAroseAction = true
		}
	}

	//结果处理，标志为真说明上一个用户所出的牌有人可以所以要发送一个提示信息的框
	if bAroseAction == true {
		//设置变量
		self.ProvideUser = wCenterUser
		self.ProvideCard = cbCenterCard
		self.ResumeUser = self.CurrentUser
		self.CurrentUser = public.INVALID_CHAIR

		//发送提示
		self.SendOperateNotify()

		return true
	}

	return false
}

// ! 发送操作
func (self *Game_mj_jlhzlzg) SendOperateNotify() bool {
	//发送提示
	for _, v := range self.PlayerInfo {
		if v.Ctx.UserAction != public.WIK_NULL {
			//构造数据
			var OperateNotify public.Msg_S_OperateNotify
			OperateNotify.ResumeUser = self.ResumeUser
			//抢暗杠时，复用此字段，表示轮到谁抢了
			OperateNotify.ActionCard = self.ProvideCard
			OperateNotify.ActionMask = v.Ctx.UserAction
			OperateNotify.EnjoinHu = false
			self.LockTimeOut(v.Seat)
			OperateNotify.Overtime = self.LimitTime
			//发送数据
			//抢的牌权需要发送给所有玩家，因为其他玩家需要知道轮到谁抢暗杠了
			if v.Ctx.UserAction == public.WIK_QIANG {
				OperateNotify.ActionCard = byte(v.Seat)
				self.SendTableMsg(constant.MsgTypeGameOperateNotify, OperateNotify)
			} else {
				self.SendPersonMsg(constant.MsgTypeGameOperateNotify, OperateNotify, v.Seat)
			}

			// 游戏记录
			if v.Ctx.UserAction > 0 {
				recordStr := fmt.Sprintf("发送牌权：%x", v.Ctx.UserAction)
				self.OnWriteGameRecord(v.Seat, recordStr)

				// 回放记录中记录牌权显示
				self.addReplayOrder(v.Seat, info.E_SendCardRight, v.Ctx.UserAction)
			}

		}
	}

	return true
}

// ! 增加回放操作记录
func (self *Game_mj_jlhzlzg) addReplayOrder(chairId uint16, operation int, card byte) {
	var order meta.Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	self.ReplayRecord.VecOrder = append(self.ReplayRecord.VecOrder, order)
}

// ! 检查是否能胡
func (self *Game_mj_jlhzlzg) CheckHu(wCurrentUser uint16, cbCurrentCard byte, bGangFlower, start bool) byte {
	self.ClearChiHuResultByUser(wCurrentUser)

	_userItem := self.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return public.WIK_NULL
	}

	//牌型权位
	wChiHuRight := uint16(0)
	//杠开权限判断
	if bGangFlower {
		wChiHuRight |= public.CHR_GANG_SHANG_KAI_HUA
	}

	if start {
		wChiHuRight |= public.CHR_TIAN_HU
	}

	tingCount := self.m_GameLogic.AnalyseTingCardCount(_userItem.Ctx.CardIndex[:], _userItem.Ctx.WeaveItemArray[:],
		_userItem.Ctx.WeaveItemCount, 0)
	if tingCount > 20 {
		wChiHuRight |= public.CHR_JIAN_ZI_HU
	}

	//给用户发牌后，胡牌判断
	_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:],
		_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, cbCurrentCard, wChiHuRight, &_userItem.Ctx.ChiHuResult)

	return _userItem.Ctx.UserAction
}

// ! 单局结算
func (self *Game_mj_jlhzlzg) OnGameOver(wChairID uint16, cbReason byte) bool {
	self.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 游戏结束, 流局结束，统计积分
func (self *Game_mj_jlhzlzg) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if self.GetGameStatus() == public.GS_MJ_END && cbReason == public.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	// 清除超时检测
	for _, v := range self.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}

	switch cbReason {
	case public.GER_NORMAL: //常规结束
		return self.OnGameEndNormal(wChairID, cbReason)
	case public.GER_USER_LEFT: //用户强退
		return self.OnGameEndUserLeft(wChairID, cbReason)
	case public.GER_DISMISS: //解散游戏
		return self.OnGameEndDissmiss(wChairID, cbReason, 0)
	case public.GER_GAME_ERROR:
		return self.OnGameEndDissmiss(wChairID, cbReason, 1)
	}
	return false
}

func (self *Game_mj_jlhzlzg) CalcWinScore(winner uint16, GameEnd *public.Msg_S_GameEnd, horseScore int) {
	//记录玩家分数
	_userItem := self.GetUserItemByChair(winner)
	if _userItem != nil {
		// 自摸三家给分
		huScore := self.GetHuScore(winner)
		winScore := huScore + horseScore*2
		for i := 0; i < self.GetPlayerCount(); i++ {
			if uint16(i) == winner {
				continue
			}
			GameEnd.GameScore[i] -= winScore * self.Rule.DiFen
			GameEnd.GameScore[winner] += winScore * self.Rule.DiFen
		}
	} else {
		// 慌庄所有人都是0分
		for i := 0; i < self.GetPlayerCount(); i++ {
			GameEnd.GameScore[i] = 0
		}
	}
}

func (self *Game_mj_jlhzlzg) GetHuScore(winner uint16) int {
	userItem := self.GetUserItemByChair(winner)
	if userItem == nil {
		return 0
	}
	huKind := userItem.Ctx.ChiHuResult.ChiHuKind
	huRight := userItem.Ctx.ChiHuResult.ChiHuRight

	if huKind == public.WIK_NULL {
		return 0
	}

	if huKind&public.CHK_PING_HU_NOMAGIC != 0 &&
		huKind&public.CHK_7_DUI != 0 {
		return 8
	}
	if huRight&public.CHR_TIAN_HU != 0 {
		return 4
	}
	if huKind&public.CHK_PING_HU_MAGIC != 0 &&
		huKind&public.CHK_7_DUI != 0 {
		return 4
	}
	if huKind&public.CHK_PING_HU_NOMAGIC != 0 {
		return 4
	}
	if huRight&public.CHR_JIAN_ZI_HU != 0 {
		return 2
	}
	if huKind&public.CHK_PING_HU_MAGIC != 0 {
		return 2
	}

	return 0
}

// 得到玩家中马详情，传入一个gameend需要赋值的字段地址，直接赋值
func (self *Game_mj_jlhzlzg) CalcBingoHorse(winner uint16, cbHouseData *[meta.MAX_PLAYER][meta.MAX_HORSE]byte, bingoHorseDetail *[meta.MAX_PLAYER]byte) int {
	horseList := make([]byte, 0)
	// 得到买马个数
	sum := self.BirdCount
	if vc := int(self.LeftCardCount); sum > vc {
		sum = vc
	}
	// 没牌了 就不买马
	if sum <= 0 {
		return 0
	}

	// 从牌堆前面翻
	for i := 0; i < sum; i++ {
		self.LeftCardCount--
		self.SendCardData = self.RepertoryCard[self.LeftCardCount]
		horseList = append(horseList, self.SendCardData)
	}

	// 轮询得到中马
	for a := 0; a < len(horseList); a++ {
		v := horseList[a] & public.MASK_VALUE
		if v == 1 || v == 5 || v == 9 || horseList[a] == logic.CARD_HONGZHONG {
			bingoHorseDetail[winner]++
		}
	}

	for index, card := range horseList {
		cbHouseData[winner][index] = card
	}
	return int(bingoHorseDetail[winner])

}

// ! 结束，结束游戏
func (self *Game_mj_jlhzlzg) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	//设置游戏结束状态
	self.SetGameStatus(public.GS_MJ_END)

	//定义变量
	var GameEnd public.Msg_S_GameEnd
	GameEnd.LastSendCardUser = self.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.IsQuit = false
	GameEnd.TheOrder = self.CurCompleteCount
	GameEnd.MagicCard = self.MagicCard
	GameEnd.Contractor = public.INVALID_CHAIR
	GameEnd.ProvideUser = wChairID
	GameEnd.ChiHuCard = self.ChiHuCard
	GameEnd.ChiHuUserCount = 1

	for i := 0; i < int(self.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = self.RepertoryCard[i]
	}

	nWinner := public.INVALID_CHAIR
	//只有一个赢家，循环判断找出赢家
	for _, v := range self.PlayerInfo {
		if v.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
			nWinner = v.Seat
			break
		}
	}
	//胡牌玩家
	GameEnd.Winner = nWinner

	winnerHorseScore := self.CalcBingoHorse(nWinner, &GameEnd.CbBirdData, &GameEnd.BingoBirdCount)
	// 记录买马
	if nWinner != public.INVALID_CHAIR {
		for i := 0; i < 8; i++ { //最多8个鸟
			cbBird := GameEnd.CbBirdData[nWinner][i]
			if cbBird != 0 {
				self.addReplayOrder(nWinner, info.E_Bird, cbBird)
			}
		}
	}

	self.CalcWinScore(nWinner, &GameEnd, winnerHorseScore)

	for i := 0; i < self.GetPlayerCount(); i++ {
		if GameEnd.Winner == uint16(i) {
			GameEnd.WWinner[i] = true
		} else {
			GameEnd.WWinner[i] = false
		}

		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		//保存四家明杠的次数
		GameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang + _userItem.Ctx.XuGang)
		//保存四家暗杠的次数
		GameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		GameEnd.WeaveItemArray[i] = _userItem.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _userItem.Ctx.WeaveItemCount
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] =
			self.m_GameLogic.SwitchToCardData2(_userItem.Ctx.CardIndex, GameEnd.CardData[i])
	}

	_userItem := self.GetUserItemByChair(nWinner)
	var huDetail modules.TagHuCostDetail
	if self.ProvideUser != public.INVALID_CHAIR && _userItem != nil {

		//默认硬胡
		GameEnd.HardHu = 1
		GameEnd.ChiHuKind = _userItem.Ctx.ChiHuResult.ChiHuKind
		//胡牌计数
		_userItem.Ctx.ChiHuUserCount++

		huKind := _userItem.Ctx.ChiHuResult.ChiHuKind
		huRight := _userItem.Ctx.ChiHuResult.ChiHuRight
		if huRight&public.CHR_GANG_SHANG_KAI_HUA != 0 {
			GameEnd.BigHuKind = public.GameBigHuKindGSK
		}
		if huKind&public.CHK_7_DUI != 0 {
			GameEnd.BigHuKind = public.GameBigHuKind_7
		}

		keepTag := true
		if huKind&public.CHK_PING_HU_NOMAGIC != 0 &&
			huKind&public.CHK_7_DUI != 0 {
			huDetail.Private(_userItem.Seat, modules.TagHard7Dui, 1, modules.DetailTypeFirst)
			keepTag = false
		}

		if keepTag && huRight&public.CHR_TIAN_HU != 0 {
			huDetail.Private(_userItem.Seat, modules.TagStartHu, 1, modules.DetailTypeFirst)
			keepTag = false
		}
		if keepTag && huKind&public.CHK_PING_HU_MAGIC != 0 &&
			huKind&public.CHK_7_DUI != 0 {
			huDetail.Private(_userItem.Seat, modules.TagSoft7Dui, 1, modules.DetailTypeFirst)
			keepTag = false
		}
		if keepTag && huKind&public.CHK_PING_HU_NOMAGIC != 0 {
			huDetail.Private(_userItem.Seat, modules.TagYingZiMo, 1, modules.DetailTypeFirst)
			keepTag = false
		}

		if keepTag && huRight&public.CHR_JIAN_ZI_HU != 0 {
			huDetail.Private(_userItem.Seat, modules.TagJianZiHu, 1, modules.DetailTypeFirst)
			keepTag = false
		}

		if keepTag && huKind&public.CHK_PING_HU_MAGIC != 0 {
			huDetail.Private(_userItem.Seat, modules.TagRuanZiMo, 1, modules.DetailTypeFirst)
			keepTag = false
		}
	} else {
		//流局
		GameEnd.ChiHuCard = 0
		GameEnd.ChiHuUserCount = 0
		self.HaveHuangZhuang = true
		GameEnd.Winner = public.INVALID_CHAIR
	}

	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		GameEnd.StrEnd[i] = huDetail.GetSeatString(uint16(i))
		if GameEnd.GameScore[i] > _item.Ctx.MaxScoreUserCount {
			_item.Ctx.SetMaxScore(GameEnd.GameScore[i])
		}
	}

	GameEnd.UserScore, GameEnd.UserVitamin = self.OnSettle(GameEnd.GameScore, constant.EventSettleGameOver)
	//发送信息
	self.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	self.SaveGameData()
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		GameEnd.GameScore[i] += _item.Ctx.StorageScore
	}
	self.VecGameEnd = append(self.VecGameEnd, GameEnd) //保存，用于汇总计算
	self.OnWriteGameRecord(public.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	//荒庄
	if self.HaveHuangZhuang {
		//记录荒庄
		self.addReplayOrder(0, info.E_HuangZhuang, 0)

		//记录胡牌类型
		self.ReplayRecord.BigHuKind = 2
		self.ReplayRecord.ProvideUser = 9
	} else {
		//记录胡牌类型
		self.ReplayRecord.BigHuKind = GameEnd.BigHuKind
		self.ReplayRecord.ProvideUser = GameEnd.ProvideUser
	}

	// 数据库写出牌记录
	self.TableWriteOutDate(int(self.CurCompleteCount), self.ReplayRecord)

	// 写完后清除数据
	self.ReplayRecord.Reset()

	//数据库写分
	for _, v := range self.PlayerInfo {
		wintype := public.ScoreKind_Draw
		if GameEnd.GameScore[v.Seat] > 0 {
			wintype = public.ScoreKind_Win
		} else {
			wintype = public.ScoreKind_Lost
		}
		self.TableWriteGameDate(int(self.CurCompleteCount), v, wintype, GameEnd.GameScore[v.Seat])
	}

	//扣房卡
	if self.CurCompleteCount == 1 {
		self.TableDeleteFangKa(self.CurCompleteCount)
	}
	//结束游戏
	if int(self.CurCompleteCount) >= self.Rule.JuShu { //局数够了
		self.CalculateResultTotal(public.GER_NORMAL, wChairID, 0) //计算总发送总结算

		self.UpdateOtherFriendDate(&GameEnd, false)
		//通知框架结束游戏
		//self.SetGameStatus(public.GS_MJ_FREE)
		self.ConcludeGame()

	} else {
	}

	// 谁胡谁当庄，荒庄起到最后一张牌的玩家为下一局的庄
	if nWinner == public.INVALID_CHAIR {
		self.BankerUser = self.LastSendCardUser
	} else {
		self.BankerUser = nWinner
	}

	self.OnGameEnd()
	self.accountManage(self.Rule.Overtime_dismiss, self.Rule.Overtime_trust, 15)
	self.RepositTable() // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	return true
}

// ! 强退，结束游戏
func (self *Game_mj_jlhzlzg) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	//变量定义
	var GameEnd public.Msg_S_GameEnd
	GameEnd.EndStatus = cbReason

	GameEnd.MagicCard = self.MagicCard

	//设置变量
	GameEnd.ProvideUser = public.INVALID_CHAIR
	GameEnd.Winner = public.INVALID_CHAIR
	GameEnd.Contractor = public.INVALID_CHAIR
	GameEnd.IsQuit = true
	for i := 0; i < int(self.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = self.RepertoryCard[i]
	}
	for i := 0; i < 4; i++ {
		GameEnd.JingDingUser[i] = public.DING_NULL
	}

	//抢杠分数，解散了也要结算
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = self.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])

		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		GameEnd.GameScore[i] += _item.Ctx.StorageScore

		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount

		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
	}

	//税收计算,按正常情况下进行计算,如果有调整,在后面进行调整
	var ScoreInfo [meta.MAX_PLAYER]rule.TagScoreInfo

	for i := 0; i < self.GetPlayerCount(); i++ {
		//记录各家的分数
		ScoreInfo[i].Score = GameEnd.GameScore[i]

		//如果分数为0，设置状态和局状态
		if GameEnd.GameScore[i] == 0 {
			ScoreInfo[i].ScoreKind = public.ScoreKind_Draw
		} else { //否则如果分数为正,设置为赢,为负设置为输,即使后面有人承包,也是输
			if GameEnd.GameScore[i] > 0 {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Win
			} else {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Lost
			}
		}
	}

	//游戏记录
	self.OnWriteGameRecord(wChairID, "强退游戏结束")

	//记录异常结束数据
	self.addReplayOrder(wChairID, info.E_Li_Xian, 0)

	//记录胡牌类型
	self.ReplayRecord.BigHuKind = GameEnd.BigHuKind
	self.ReplayRecord.ProvideUser = GameEnd.ProvideUser
	//荒庄
	if self.HaveHuangZhuang {
		//记录荒庄
		self.addReplayOrder(0, info.E_HuangZhuang, 0)

		self.ReplayRecord.BigHuKind = 2
		self.ReplayRecord.ProvideUser = 9
	}

	nextCardLen := len(GameEnd.NextCard)
	for i := 0; i < nextCardLen; i++ {
		GameEnd.NextCard[i] = 0x00
	}
	//发送几张牌出去
	if int(self.LeftCardCount) >= nextCardLen {
		for i := 0; i < nextCardLen; i++ {
			self.LeftCardCount--
			GameEnd.NextCard[i] = self.RepertoryCard[self.LeftCardCount]
		}
	}

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, GameEnd) //保存，用于汇总计算
	self.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	self.SaveGameData()

	if self.GetGameStatus() != public.GS_MJ_FREE {
		//数据库写出牌记录
		self.TableWriteOutDate(int(self.CurCompleteCount), self.ReplayRecord)
		// 写完后清除数据
		self.ReplayRecord.Reset()

		//数据库写分
		for _, v := range self.PlayerInfo {
			if v.Seat != public.INVALID_CHAIR {
				if wChairID != public.INVALID_CHAIR && v.Seat == wChairID {
					self.TableWriteGameDate(int(self.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
				} else {
					self.TableWriteGameDate(int(self.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
				}
			}
		}
	}

	self.UpdateOtherFriendDate(&GameEnd, true)
	//结束游戏
	self.CalculateResultTotal(public.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	//self.SetGameStatus(public.GS_MJ_FREE)
	self.ConcludeGame()

	return true
}

// ! 解散，结束游戏
func (self *Game_mj_jlhzlzg) OnGameEndDissmiss(wChairID uint16, cbReason byte, cbSubReason byte) bool {
	//变量定义
	var GameEnd public.Msg_S_GameEnd
	GameEnd.LastSendCardUser = self.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.ProvideUser = public.INVALID_CHAIR
	GameEnd.Winner = public.INVALID_CHAIR
	GameEnd.Contractor = public.INVALID_CHAIR
	GameEnd.MagicCard = self.MagicCard
	for i := 0; i < int(self.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = self.RepertoryCard[i]
	}
	for i := 0; i < 4; i++ {
		GameEnd.JingDingUser[i] = public.DING_NULL
	}
	//记录异常结束数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == public.US_OFFLINE && i != int(wChairID) {
			self.addReplayOrder(uint16(i), info.E_Li_Xian, 0)
		}
	}

	self.addReplayOrder(wChairID, info.E_Jie_san, 0)

	//抢杠分数，解散了也要结算
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.GameScore[i] += _item.Ctx.StorageScore

		if self.Rule.HasPao {
			GameEnd.PaoCount[i] = uint16(_item.Ctx.VecXiaPao.Num)
		} else {
			GameEnd.PaoCount[i] = 0xFF
		}

		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = self.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
	}

	//税收计算,按正常情况下进行计算,如果有调整,在后面进行调整
	var ScoreInfo [meta.MAX_PLAYER]rule.TagScoreInfo

	for i := 0; i < self.GetPlayerCount(); i++ {
		//记录各家的分数
		ScoreInfo[i].Score = GameEnd.GameScore[i]

		//如果分数为0，设置状态和局状态
		if GameEnd.GameScore[i] == 0 {
			ScoreInfo[i].ScoreKind = public.ScoreKind_Draw
		} else { //否则如果分数为正,设置为赢,为负设置为输,即使后面有人承包,也是输
			if GameEnd.GameScore[i] > 0 {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Win
			} else {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Lost
			}
		}
	}
	GameEnd.IsQuit = true

	nextCardLen := len(GameEnd.NextCard)
	for i := 0; i < nextCardLen; i++ {
		GameEnd.NextCard[i] = 0x00
	}
	//发送几张牌出去
	if int(self.LeftCardCount) >= nextCardLen {
		for i := 0; i < nextCardLen; i++ {
			self.LeftCardCount--
			GameEnd.NextCard[i] = self.RepertoryCard[self.LeftCardCount]
		}
	}

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, GameEnd) //保存，用于汇总计算
	self.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	self.SaveGameData()
	switch cbSubReason {
	case 0:
		//游戏记录
		self.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

		if self.GetGameStatus() != public.GS_MJ_FREE {
			//数据库写出牌记录
			self.TableWriteOutDate(int(self.CurCompleteCount), self.ReplayRecord)
			// 写完后清除数据
			self.ReplayRecord.Reset()

			//数据库写分
			for _, v := range self.PlayerInfo {
				if v.Seat != public.INVALID_CHAIR {
					if wChairID != public.INVALID_CHAIR && v.Seat == wChairID {
						self.TableWriteGameDate(int(self.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
					} else {
						self.TableWriteGameDate(int(self.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
					}
				}
			}
		}
	case 1:
		self.OnWriteGameRecord(wChairID, "前面某个时刻程序出错过，需要排查错误日志，无法恢复这局游戏，解散游戏OnGameEndErrorDissmis")
	}

	self.UpdateOtherFriendDate(&GameEnd, true)
	// 写总计算
	self.CalculateResultTotal(public.GER_DISMISS, wChairID, cbSubReason)
	//结束游戏
	//self.SetGameStatus(public.GS_MJ_FREE)
	self.ConcludeGame()

	return true
}

// ! 解散牌桌
func (self *Game_mj_jlhzlzg) OnEnd() {
	if self.IsGameStarted() {
		self.OnGameOver(public.INVALID_CHAIR, public.GER_DISMISS)
	}
}

// ! 计算总发送总结算
func (self *Game_mj_jlhzlzg) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	// 给客户端发送总结算数据
	var balanceGame public.Msg_S_BALANCE_GAME
	balanceGame.Userid = self.Rule.FangZhuID
	balanceGame.CurTotalCount = self.CurCompleteCount //总盘数
	self.TimeEnd = time.Now().Unix()                  //大局结束时间
	balanceGame.TimeStart = self.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = self.TimeEnd
	for i := 0; i < len(self.VecGameEnd); i++ {
		for j := 0; j < self.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += self.VecGameEnd[i].GameScore[j] //总分
		}
	}

	//打印日志
	recordStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）", cbReason)
	self.OnWriteGameRecord(wChairID, recordStr)

	for i := 0; i < self.GetPlayerCount(); i++ {
		self.OnWriteGameRecord(uint16(i), "当前该玩家离线")
	}

	//非正常结束，记录总结算时玩家状态：0正常，1解散，2离线
	if public.GER_USER_LEFT == cbReason {
		for i := 0; i < self.GetPlayerCount(); i++ {
			if i == int(wChairID) {
				balanceGame.UserEndState[i] = 2
			}

		}
	} else {
		if public.GER_DISMISS == cbReason {
			for i := 0; i < self.GetPlayerCount(); i++ {
				if i == int(wChairID) {
					balanceGame.UserEndState[i] = 1
				} else {
					if _item := self.GetUserItemByChair(uint16(i)); _item != nil && _item.UserStatus == public.US_OFFLINE {
						balanceGame.UserEndState[i] = 2
					}
				}
			}
		}
	}

	// 如果是正常结束
	//if (GER_NORMAL == cbReason)
	{
		// 有打赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		//		iChairID := 0
		for i := 0; i < self.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < self.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				//				iChairID = j
				iMaxScoreCount++
			}
		}
		if iMaxScoreCount == 1 && self.Rule.CreateType == 3 { // 大赢家支付
			//IServerUserItem * pIServerUserItem = m_pITableFrame->GetServerUserItem(iChairID);
			//DWORD userid = pIServerUserItem->GetUserID();
			//				m_pITableFrame->TableDeleteDaYingJiaFangKa(userid);
		}
	}

	//! 计算大赢家的座位号
	bigWinScore := 0
	for i := 0; i < self.GetPlayerCount(); i++ {
		if balanceGame.GameScore[i] >= bigWinScore {
			bigWinScore = balanceGame.GameScore[i]
		}
	}

	//开始就解散时不加，每局结束后游戏状态会复位为GS_MJ_FREE
	wintype := public.ScoreKind_Draw
	if self.CurCompleteCount == 1 && self.GetGameStatus() != public.GS_MJ_END {
		wintype = public.ScoreKind_pass
	} else {
		if self.CurCompleteCount == 0 { //有可能第一局还没有开始，就解散了（比如在吓跑的过程中解散）
			wintype = public.ScoreKind_pass
		}
	}
	if cbSubReason == 0 {
		for i := 0; i < self.GetPlayerCount(); i++ {
			_userItem := self.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			balanceGame.FXMaxUserCount[i] = _userItem.Ctx.MaxFanUserCount
			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
			balanceGame.FXScoreUserCount[i] = _userItem.Ctx.MaxScoreUserCount
			if balanceGame.FXScoreUserCount[i] < 0 {
				balanceGame.FXScoreUserCount[i] = 0
			}

			if wintype != public.ScoreKind_pass {
				if balanceGame.GameScore[i] > 0 {
					wintype = public.ScoreKind_Win
				} else {
					wintype = public.ScoreKind_Lost
				}
			}

			isBigWin := 0
			if bigWinScore == balanceGame.GameScore[i] {
				isBigWin = 1
			}
			//写记录
			self.TableWriteGameDateTotal(int(self.CurCompleteCount), i, wintype, balanceGame.GameScore[i], isBigWin)
		}
	} else {
		self.UpdateErrGameTotal(self.GetTableInfo().GameNum)
	}

	// 记录用户好友房历史战绩
	if wintype != public.ScoreKind_pass {
		self.TableWriteHistoryRecord(&balanceGame)
		self.TableWriteHistoryRecordDetail(&balanceGame)
	}

	balanceGame.End = 0

	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}

		gameendStr := ""
		if len(self.VecGameEnd) > 0 {
			gameendStr = public.HF_JtoA(self.VecGameEnd[len(self.VecGameEnd)-1])
		}
		gamedataStr := ""
		if len(self.VecGameDataAllP[i]) > 0 {
			gamedataStr = public.HF_JtoA(self.VecGameDataAllP[i][len(self.VecGameDataAllP[i])-1])
		}
		self.SaveLastGameinfo(_userItem.Uid, gameendStr, public.HF_JtoA(balanceGame), gamedataStr)
	}

	//发消息
	self.SendTableMsg(constant.MsgTypeGameBalanceGame, balanceGame)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameBalanceGame, balanceGame)

	self.resetEndDate()
}

// ! 重置优秀结束数据
func (self *Game_mj_jlhzlzg) resetEndDate() {
	self.CurCompleteCount = 0
	self.VecGameEnd = []public.Msg_S_GameEnd{}

	for _, v := range self.PlayerInfo {
		v.OnEnd()
	}
}

func (self *Game_mj_jlhzlzg) UpdateOtherFriendDate(GameEnd *public.Msg_S_GameEnd, bEnd bool) {

}

// ! 给客户端发送总结算
func (self *Game_mj_jlhzlzg) CalculateResultTotal_Rep(msg *public.Msg_C_BalanceGameEeq) {
	// 给客户端发送总结算数据
	var balanceGame public.Msg_S_BALANCE_GAME
	balanceGame.Userid = self.Rule.FangZhuID
	balanceGame.CurTotalCount = self.CurCompleteCount //总盘数

	/*为什么要特意判断下0？之前没有这个协议，是新需求要加的之前很多数据都在RepositTableFrameSink里面重置了，但是由于新功能的开发
	部分数据放到游戏开始的时候重置了，但是存在的问题是，如果一个桌子解散了，这个桌子在服务器内部还是存在的，他的部分数据也是存在的
	当某个玩家进来，还没有点开始的时候点战绩，这时候存在把本茶楼桌子上次数据发送过来的情况，但是改动m_MaxFanUserCount之类变量的值又
	有很大风险，因此此处做个特出处理，如果是第0局，没有开始，那就无条件全部返回0*/
	if 0 == balanceGame.CurTotalCount {
		for i := 0; i < len(self.VecGameEnd); i++ {
			for j := 0; j < self.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += 0 //总分
			}
		}

		for i := 0; i < self.GetPlayerCount(); i++ {
			balanceGame.ChiHuUserCount[i] = 0
			balanceGame.ProvideUserCount[i] = 0
			balanceGame.FXMaxUserCount[i] = 0
			balanceGame.HHuUserCount[i] = 0
			balanceGame.UserEndState[i] = 0
		}
	} else {
		for i := 0; i < len(self.VecGameEnd); i++ {
			for j := 0; j < self.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += self.VecGameEnd[i].GameScore[j] //总分
			}
		}

		for i := 0; i < self.GetPlayerCount(); i++ {
			balanceGame.UserEndState[i] = 0
		}

		// 有打赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		for i := 0; i < self.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < self.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				iMaxScoreCount++
			}
		}

		for i := 0; i < self.GetPlayerCount(); i++ {
			_userItem := self.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			balanceGame.FXMaxUserCount[i] = _userItem.Ctx.MaxFanUserCount
			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
		}
	}
	balanceGame.End = 1
	//发消息
	self.SendPersonMsg(constant.MsgTypeGameBalanceGame, balanceGame, self.GetChairByUid(msg.Id))
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameBalanceGame, balanceGame)
}

// ! 发送游戏开始场景数据  旁观玩家
func (self *Game_mj_jlhzlzg) sendGameSceneStatusPlayLookon(player *modules.Player) bool {

	if player.LookonTableId == 0 {
		return false
	}
	wChiarID := player.GetChairID()
	if int(wChiarID) >= self.GetPlayerCount() {
		wChiarID = 0
	}
	//是否要获取wChiarID位置真正玩家的信息 ？
	playerOnChair := self.GetUserItemByChair(wChiarID)

	//变量定义
	var StatusPlay public.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = self.SiceCount
	StatusPlay.BankerUser = self.BankerUser
	StatusPlay.CurrentUser = self.CurrentUser
	StatusPlay.CellScore = self.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = self.PiZiCard
	StatusPlay.KaiKou = self.Rule.KouKou
	StatusPlay.RenWuAble = self.RenWuAble
	StatusPlay.TheOrder = self.CurCompleteCount
	if self.CurrentUser == wChiarID {
		StatusPlay.SendCardData = self.SendCardData
	} else {
		StatusPlay.SendCardData = public.INVALID_BYTE
	}

	StatusPlay.CardLeft.CardArray = make([]int, self.RepertoryCardArray.MaxCount, self.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, self.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = self.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(self.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = self.RepertoryCardArray.Kaikou

	for i := 0; i < self.GetPlayerCount(); i++ {

		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang + _item.Ctx.PiZiGangCount
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady

		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = self.ProvideCard
	StatusPlay.LeftCardCount = self.LeftCardCount
	StatusPlay.ActionMask = 0 //player.Ctx.UserAction

	if playerOnChair != nil && playerOnChair.Ctx.Response {
		StatusPlay.ActionMask = public.WIK_NULL
	}

	if playerOnChair != nil && playerOnChair.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = playerOnChair.Ctx.CheckTimeOut
	} else {
		CurUserItem := self.GetUserItemByChair(self.CurrentUser)
		if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
		}

		for _, v := range self.PlayerInfo {
			if v.GetChairID() == wChiarID {
				continue
			}
			if v.GetChairID() == self.CurrentUser {
				continue
			}
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if self.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = self.OutCardUser
	StatusPlay.OutCardData = self.OutCardData
	StatusPlay.LastOutCardUser = self.LastOutCardUser
	//扑克数据
	if playerOnChair != nil {
		StatusPlay.CardCount, StatusPlay.CardData = self.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
	}

	//发送旁观数据
	self.SendPersonLookonMsg(constant.MsgTypeGameStatusPlay, StatusPlay, player.Uid)

	//发小结消息
	if byte(len(self.VecGameEnd)) == self.CurCompleteCount && self.CurCompleteCount != 0 {
		//发消息
		gamend := self.VecGameEnd[self.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		//发送旁观数据
		self.SendPersonLookonMsg(constant.MsgTypeGameEnd, gamend, player.Uid)
	}

	//发送解散房间所有玩家的反应
	self.SendAllPlayerDissmissInfo(player)

	return true
}

// ! 发送游戏开始场景数据
func (self *Game_mj_jlhzlzg) sendGameSceneStatusPlay(player *modules.Player) bool {

	if player.LookonTableId > 0 {
		self.sendGameSceneStatusPlayLookon(player)
		return true
	}

	chairID := player.GetChairID()

	if chairID == public.INVALID_CHAIR {
		syslog.Logger().Debug("sendGameSceneStatusPlay invalid chair")
		return false
	}
	//取消托管
	player.Ctx.SetTrustee(false)
	var Trustee public.Msg_S_Trustee
	Trustee.Trustee = false
	Trustee.ChairID = chairID
	self.SendTableMsg(constant.MsgTypeGameTrustee, Trustee)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameTrustee, Trustee)

	//变量定义
	var StatusPlay public.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = self.SiceCount
	StatusPlay.BankerUser = self.BankerUser
	StatusPlay.CurrentUser = self.CurrentUser
	StatusPlay.CellScore = self.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = self.PiZiCard
	StatusPlay.KaiKou = self.Rule.KouKou
	StatusPlay.RenWuAble = self.RenWuAble
	StatusPlay.TheOrder = self.CurCompleteCount
	if self.CurrentUser == player.Seat {
		StatusPlay.SendCardData = self.SendCardData
	} else {
		StatusPlay.SendCardData = public.INVALID_BYTE
	}

	StatusPlay.CardLeft.CardArray = make([]int, self.RepertoryCardArray.MaxCount, self.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, self.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = self.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(self.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = self.RepertoryCardArray.Kaikou

	for i := 0; i < self.GetPlayerCount(); i++ {

		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang + _item.Ctx.PiZiGangCount
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady

		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = self.ProvideCard
	StatusPlay.LeftCardCount = self.LeftCardCount
	StatusPlay.ActionMask = player.Ctx.UserAction

	if player.Ctx.Response {
		StatusPlay.ActionMask = public.WIK_NULL
	}

	if player.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = player.Ctx.CheckTimeOut
	} else {
		CurUserItem := self.GetUserItemByChair(self.CurrentUser)
		if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
		}

		for _, v := range self.PlayerInfo {
			if v.GetChairID() == player.GetChairID() {
				continue
			}
			if v.GetChairID() == self.CurrentUser {
				continue
			}
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if self.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = self.OutCardUser
	StatusPlay.OutCardData = self.OutCardData
	StatusPlay.LastOutCardUser = self.LastOutCardUser
	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardData = self.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)

	//发送场景
	self.SendPersonMsg(constant.MsgTypeGameStatusPlay, StatusPlay, chairID)
	//发小结消息
	if byte(len(self.VecGameEnd)) == self.CurCompleteCount && self.CurCompleteCount != 0 && int(chairID) < self.GetPlayerCount() && chairID >= 0 {
		//发消息
		gamend := self.VecGameEnd[self.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		self.SendPersonMsg(constant.MsgTypeGameEnd, gamend, chairID)
	}

	//发送解散房间所有玩家的反应
	self.SendAllPlayerDissmissInfo(player)

	return true
}

func (self *Game_mj_jlhzlzg) SaveGameData() {
	//变量定义
	var StatusPlay public.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = self.SiceCount
	StatusPlay.BankerUser = self.BankerUser
	StatusPlay.CurrentUser = self.CurrentUser
	StatusPlay.CellScore = self.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = self.PiZiCard
	StatusPlay.KaiKou = self.Rule.KouKou
	StatusPlay.RenWuAble = self.RenWuAble
	StatusPlay.TheOrder = self.CurCompleteCount

	StatusPlay.CardLeft.CardArray = make([]int, self.RepertoryCardArray.MaxCount, self.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, self.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = self.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(self.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = self.RepertoryCardArray.Kaikou

	for i := 0; i < self.GetPlayerCount(); i++ {

		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang + _item.Ctx.PiZiGangCount
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady

		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = self.ProvideCard
	StatusPlay.LeftCardCount = self.LeftCardCount

	if self.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = self.OutCardUser
	StatusPlay.OutCardData = self.OutCardData
	StatusPlay.LastOutCardUser = self.LastOutCardUser

	//玩家的个人数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		player := self.GetUserItemByChair(uint16(i))
		if player == nil {
			continue
		}
		if self.CurrentUser == player.Seat {
			StatusPlay.SendCardData = self.SendCardData
		} else {
			StatusPlay.SendCardData = public.INVALID_BYTE
		}
		StatusPlay.ActionMask = player.Ctx.UserAction
		if player.Ctx.Response {
			StatusPlay.ActionMask = public.WIK_NULL
		}
		if player.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = player.Ctx.CheckTimeOut
		} else {
			CurUserItem := self.GetUserItemByChair(self.CurrentUser)
			if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
				StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
			}
			for _, v := range self.PlayerInfo {
				if v.GetChairID() == player.GetChairID() {
					continue
				}
				if v.GetChairID() == self.CurrentUser {
					continue
				}
				if v.Ctx.UserAction > 0 {
					StatusPlay.Overtime = 0
					break
				}
			}
		}
		//扑克数据
		StatusPlay.CardCount, StatusPlay.CardData = self.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
		self.VecGameDataAllP[i] = append(self.VecGameDataAllP[i], StatusPlay) //保存，用于汇总计算
	}
}

// 游戏场景消息发送
func (self *Game_mj_jlhzlzg) SendGameScene(uid int64, status byte, secret bool) {
	player := self.GetUserItemByUid(uid)
	if player == nil {
		//不是游戏玩家就是旁观玩家
		player = self.GetLookonUserItemByUid(uid)
		if player == nil {
			self.OnWriteGameRecord(public.INVALID_CHAIR, "SendGameScene 发送游戏场景，玩家空指针")
			return
		}
	}
	switch status {
	case public.GS_MJ_FREE:
		self.SendGameSceneStatusFree(player)
	case public.GS_MJ_PLAY:
		self.sendGameSceneStatusPlay(player)
	case public.GS_MJ_END:
		self.sendGameSceneStatusPlay(player)
	}
}

// ! 游戏退出
func (self *Game_mj_jlhzlzg) OnExit(uid int64) {
	self.GameCommon.OnExit(uid)
}

// ! 定时器
func (self *Game_mj_jlhzlzg) OnTime() {
	self.GameCommon.OnTime()
}

// ! 计时器事件
func (self *Game_mj_jlhzlzg) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	if dwTimerID == modules.GameTime_Nine {
		if TablePerson := self.GetUserItemByUid(wBindParam); TablePerson != nil {
			//到时间了
			self.OnAutoOperate(TablePerson.Seat, true)
		}
	}

	return true
}

// ! 玩家开启超时
func (self *Game_mj_jlhzlzg) LockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(self.GetPlayerCount()) {
		return
	}

	_userItem := self.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + public.GAME_OPERATION_TIME
	self.LimitTime = _userItem.Ctx.CheckTimeOut

	if self.Rule.NineSecondRoom {
		_userItem.Ctx.Timer.SetTimer(modules.GameTime_Nine, public.GAME_OPERATION_TIME)
	}

	if self.Rule.Overtime_trust > 0 {
		checkTime := self.Rule.Overtime_trust
		self.LimitTime = time.Now().Unix() + int64(checkTime)
		if _userItem.CheckTRUST() {
			// 托管状态
			checkTime = 1
		}
		_userItem.Ctx.CheckTimeOut = time.Now().Unix() + int64(checkTime)
		_userItem.Ctx.Timer.SetTimer(modules.GameTime_Nine, checkTime)
	}
}

// ! 玩家关闭超时
func (self *Game_mj_jlhzlzg) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(self.GetPlayerCount()) {
		return
	}

	_userItem := self.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0
	_userItem.Ctx.Timer.KillTimer(modules.GameTime_Nine)
}

// ! 写日志记录
func (self *Game_mj_jlhzlzg) WriteGameRecord() {
	//写日志记录
	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始监利红中赖子杠  发牌......")

	// 玩家手牌
	for i := 0; i < len(self.PlayerInfo); i++ {
		v := self.GetUserItemByChair(uint16(i))
		if v != nil {
			handCardStr := fmt.Sprintf("发牌后手牌:%s", self.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
			self.OnWriteGameRecord(uint16(v.Seat), handCardStr)
		}
	}

	// 牌堆牌
	leftCardStr := fmt.Sprintf("牌堆牌:%s", self.m_GameLogic.SwitchToCardNameByDatas(self.RepertoryCard[0:self.LeftCardCount+2], 0))
	self.OnWriteGameRecord(public.INVALID_CHAIR, leftCardStr)

	//赖子牌
	magicCardStr := fmt.Sprintf("癞子牌:%s", self.m_GameLogic.SwitchToCardNameByData(self.MagicCard, 1))
	self.OnWriteGameRecord(public.INVALID_CHAIR, magicCardStr)
}

// ! 场景保存
func (self *Game_mj_jlhzlzg) Tojson() string {
	var _json modules.GameJsonSerializer
	_json.ToJson(&self.GameMeta)

	_json.GameCommonToJson(&self.GameCommon)

	return public.HF_JtoA(&_json)
}

// ! 场景恢复
func (self *Game_mj_jlhzlzg) Unmarsha(data string) {
	var _json modules.GameJsonSerializer
	if data != "" {
		json.Unmarshal([]byte(data), &_json)

		_json.Unmarsha(&self.GameMeta)
		_json.JsonToStruct(&self.GameCommon)

		self.ParseRule(self.GetTableInfo().Config.GameConfig)
		self.m_GameLogic.Rule = self.Rule
		self.m_GameLogic.HuType = self.HuType
		self.m_GameLogic.SetMagicCard(self.MagicCard)
		self.m_GameLogic.SetPiZiCard(self.PiZiCard)
	}
}

func (self *Game_mj_jlhzlzg) OnUserTustee(msg *public.Msg_S_DG_Trustee) bool {
	if self.Rule.Overtime_trust > 0 {
		item := self.GetUserItemByChair(msg.ChairID)
		if item == nil {
			return false
		}

		if item.CheckTRUST() == msg.Trustee {
			return true
		}
		var tuoguan public.Msg_S_DG_Trustee
		tuoguan.ChairID = msg.ChairID
		tuoguan.Trustee = msg.Trustee
		//校验规则
		if tuoguan.ChairID < public.MAX_PLAYER_4P {
			if tuoguan.Trustee == true /*&& (self.GameState != gameserver.GsNull)*/ {
				self.TrustCounts[tuoguan.ChairID]++
				//进入托管啥都不用做
				item.ChangeTRUST(true)
				self.SendTableMsg(constant.MsgTypeGameTrustee, tuoguan)
				//发送旁观数据
				self.SendTableLookonMsg(constant.MsgTypeGameTrustee, tuoguan)
				//fmt.Println(fmt.Sprintf("(%d)进入托管（%v）状态（%t）",tuoguan.ChairID,self.TuoGuanPlayer,item.CheckTRUST()))
				return true
			} else if tuoguan.Trustee == false {
				//如果是当前的玩家，那么重新设置一下开始时间
				item.ChangeTRUST(false)
				if tuoguan.ChairID == self.CurrentUser {
					//self.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
					_item := self.GetUserItemByChair(self.CurrentUser)
					if _item != nil {
						//if time.Now().Unix() < _item.Ctx.CheckTimeOut { //如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < self.LimitTime
						self.LockTimeOut(self.CurrentUser)
						//self.setLimitedTime(int64(self.PlayTime) + self.PowerStartTime - time.Now().Unix() + 1)
						tuoguan.Overtime = _item.Ctx.CheckTimeOut
					}
				}
				self.SendTableMsg(constant.MsgTypeGameTrustee, tuoguan)
				//发送旁观数据
				self.SendTableLookonMsg(constant.MsgTypeGameTrustee, tuoguan)
				//fmt.Println(fmt.Sprintf("(%d)取消托管（%v）状态（%t）",tuoguan.ChairID,self.TuoGuanPlayer,item.CheckTRUST()))
				return false
			}
		} else {
			//详细日志
			LogStr := string("托管动作:游戏状态不正确 ")
			self.OnWriteGameRecord(tuoguan.ChairID, LogStr)
			return false
		}
		return false
	}
	return true
}
func (self *Game_mj_jlhzlzg) Greate_OutCardRecord(handcards string, msg *public.Msg_C_OutCard, _userItem *modules.Player) string {
	tempstr := "客户端的操作"
	if !msg.ByClient {
		tempstr = "服务器自动操作"
		if _userItem.UserOfflineTag^-1 != 0 {
			tempstr += "(离线)"
		}
		if _userItem.CheckTRUST() {
			tempstr += "(托管)"
		}
	}
	return fmt.Sprintf("%s，打出：%s，来源：%s", handcards, self.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1), tempstr)
}

// ! 创建弃操作消息
func (self *Game_mj_jlhzlzg) Greate_OperateRecord(msg *public.Msg_C_OperateCard, _userItem *modules.Player) {
	if msg.Code == public.WIK_NULL {
		if msg.ByClient {
			self.OnWriteGameRecord(_userItem.Seat, "客户端点击弃！")
		} else {
			tempstr := "服务器点击弃！"
			if _userItem.CheckTRUST() {
				tempstr += "(托管)"
			}
			if _userItem.UserOfflineTag^-1 != 0 {
				tempstr += "（离线）"
			}
			self.OnWriteGameRecord(_userItem.Seat, tempstr)
		}
	}
}

func (self *Game_mj_jlhzlzg) Greate_ContendRecord(code byte, operate bool, card byte, byclient bool, _userItem *modules.Player) {
	tempstr := fmt.Sprintf("迟到的消息来源于(服务端)")
	if byclient {
		tempstr = fmt.Sprintf("迟到的消息来源于(客户端)")
	}
	if operate {
		tempstr += fmt.Sprintf(",消息体:操作(%s)", public.GetPaiQuanStr(uint64(code)))
		if card != public.INVALID_BYTE {
			tempstr += fmt.Sprintf(",牌(%s)", self.m_GameLogic.SwitchToCardNameByData(card, 1))
		}
	} else {
		tempstr += fmt.Sprintf(",消息体:打出(%s)", self.m_GameLogic.SwitchToCardNameByData(card, 1))
	}
	if byclient {
		tempstr += "，将发送刷新消息给客户端"
	}
	self.OnWriteGameRecord(_userItem.Seat, tempstr)
}
func (self *Game_mj_jlhzlzg) entryTrust(byclient bool, _userItem *modules.Player) {
	if !byclient {
		if !_userItem.CheckTRUST() {
			//var msg = &public.Msg_C_Trustee{
			//	Id:      _userItem.Uid,
			//	Trustee: true,
			//}
			var msg = &public.Msg_S_DG_Trustee{
				ChairID: _userItem.Seat,
				Trustee: true,
			}
			if self.OnUserTustee(msg) {
				self.OnWriteGameRecord(_userItem.Seat, "超时进入托管")
			}
		}
	}
}

func (self *Game_mj_jlhzlzg) accountManage(dismisstime int, trusttime int, autonexttimer int) int {
	if !(int(self.CurCompleteCount) >= self.Rule.JuShu) && self.CurCompleteCount != 0 {
		check := false
		if dismisstime != -1 {
			for i := 0; i < public.MAX_PLAYER_4P; i++ {
				if item := self.GetUserItemByChair(uint16(i)); item != nil {
					//fmt.Println(fmt.Sprintf("玩家（%d）托管（%t）", item.Seat, item.CheckTRUST()))
					if item.CheckTRUST() {
						if check {
							var _msg = &public.Msg_C_DismissFriendResult{
								Id:   item.Uid,
								Flag: true,
							}
							self.OnDismissResult(item.Uid, _msg)
							//fmt.Println(fmt.Sprintf("玩家（%d）托管（%t）同意解散",item.Seat,item.CheckTRUST() ))
							//发准备
							//_msgauto:= &public.Msg_C_GoOnNextGame{
							//	Id:   item.Uid,
							//}
							//self.OnUserClientNextGame(_msgauto)
						} else {
							check = true
							var msg = &public.Msg_C_DismissFriendReq{
								Id: item.Uid,
							}
							self.SetDismissRoomTime(dismisstime)
							self.OnDismissFriendMsg(item.Uid, msg)
							//fmt.Println(fmt.Sprintf("玩家（%d）托管（%t）申请解散",item.Seat,item.CheckTRUST() ))
							//发准备
							//_msgauto:= &public.Msg_C_GoOnNextGame{
							//	Id:   item.Uid,
							//}
							//self.OnUserClientNextGame(_msgauto)
						}
					}
				}
			}
		}
		if !check && trusttime > 0 && autonexttimer > 0 {
			//fmt.Println("自动下一局")
			self.SetAutoNextTimer(autonexttimer) //自动开始下一局
			return autonexttimer
		}
	}
	return 0
}
