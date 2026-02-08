package Hubei_JianLi

import (
	"encoding/json"
	"errors"
	"fmt"
	constant "github.com/open-source/game/chess.git/pkg/consts"
	model "github.com/open-source/game/chess.git/pkg/models"
	public "github.com/open-source/game/chess.git/pkg/static"
	syslog "github.com/open-source/game/chess.git/pkg/xlog"
	modules "github.com/open-source/game/chess.git/services/sport/components"
	base "github.com/open-source/game/chess.git/services/sport/infrastructure"
	logic "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	gsvr "github.com/open-source/game/chess.git/services/sport/wuhan"
	"math/rand"
	"strings"
	"time"
)

//constant "github.com/open-source/game/chess.git/pkg/consts"
//model "github.com/open-source/game/chess.git/pkg/models"
//public "github.com/open-source/game/chess.git/pkg/static"
//syslog "github.com/open-source/game/chess.git/pkg/xlog"
//modules "github.com/open-source/game/chess.git/services/sport/components"
//base "github.com/open-source/game/chess.git/services/sport/infrastructure"
//info "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
//logic "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
//meta "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
//common "github.com/open-source/game/chess.git/services/sport/backboard"

/*
监利开机纸牌4人好友房
*/
type FriendRule_jlkj struct {
	Difen            int    `json:"difen"`            //底分(级数分)
	Radix            int    `json:"scoreradix"`       //底分基数
	SerPay           int    `json:"revenue"`          //茶水
	Fa               int    `json:"fa"`               //没逃跑处罚倍数
	Jiang            int    `json:"jiang"`            //逃跑奖励别人的倍数
	ShareScoreType   int    `json:"sharescoretype"`   //贡献分类型 (type 1 : 30-60  type 2 : 40-60)
	NineSecondRoom   string `json:"sc9"`              //九秒场
	IpLimit          string `json:"ip"`               //ipgps限制
	Overtime_trust   int    `json:"overtime_trust"`   //超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //超时解散
	FullScoreAward   int    `json:"fullscoreaward"`   //满分奖励
	FourKingScore    int    `json:"fourkingscore"`    //4王换分
	AddDiFen         int    `json:"adddifen"`         //额外加的底分
	ShowHandCardCnt  string `json:"showhandcardcnt"`  //是否显示手牌数
	GetLastScore     string `json:"getlastscore"`     //捡尾分
	SeeTeamerCard    string `json:"seeteamercard"`    //先跑可看队友牌
	Hard510KMode     string `json:"hard510kmode"`     //纯510K大于4炸
	SeePartnerCards  string `json:"seepartnercards"`  //可看队友手牌
}

// 4王换分 client
type Msg_C_4KingScore struct {
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	Flag        byte   `json:"flag"`        //1换，0不换
}

// 4王换分 server
type Msg_S_4KingScore struct {
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	Flag        byte   `json:"flag"`        //1换，0不换
}

/*
贡献分说明：

	贡献分是单选项;
	选择了30-60，就是12游为胜利一方的时候,34游要给12游60分; 13游为胜利一方的时候,24游需要给13游30分
	选择了40-60，就是12游为胜利一方的时候,34游要给12游60分; 13游为胜利一方的时候,24游需要给13游40分
*/
const (
	GongScoreType_3060 = iota + 1
	GongScoreType_4060
)

const (
	YOUTYPE_12 = iota + 1 //一二游
	YOUTYPE_13            //一三游
	YOUTYPE_14            //一四游
	YOUTYPE_34            //三四游
	YOUTYPE_24            //二四游
	YOUTYPE_23            //二三游
)

type Game_jl_jlkj struct {
	modules.GameCommon
	//游戏变量
	meta.GameMetaDG
	m_GameLogic GameLogic_jl_jlkj //游戏逻辑
}

func (self *Game_jl_jlkj) GetGameConfig() *public.GameConfig { //获取游戏相关配置
	return &self.Config
}

// 重置桌子数据
func (self *Game_jl_jlkj) RepositTable(ResetAllData bool) {
	rand.Seed(time.Now().UnixNano())
	for _, v := range self.PlayerInfo {
		v.Reset()
	}
	//游戏变量
	self.GameEndStatus = public.GS_MJ_FREE
	if ResetAllData {
		//游戏变量
		self.GameMetaDG.Reset()
	} else {
		//游戏变量
		self.GameMetaDG.ResetForNext() //保留部分数据给下一局使用
	}
}

func (self *Game_jl_jlkj) switchCard2Ox(_index int) int {
	_hight := (_index - 1) / 13
	_low := (_index-1)%13 + 1

	return (_low + (_hight << 4))
}

func (self *Game_jl_jlkj) getCardIndexByOx(_card byte) int {
	low_index := int(0x0F)
	hight_index := int(0xF0)

	return ((low_index & int(_card)) + ((hight_index&int(_card))>>4)*13)
}

// 解析配置的任务
func (self *Game_jl_jlkj) ParseRule(strRule string) {

	syslog.Logger().Info("parserRule :" + strRule)

	//表示底分要除以10
	self.Rule.NineSecondRoom = false

	self.Rule.JuShu = self.GetTableInfo().Config.RoundNum
	self.Spay = 0
	self.Rule.FangZhuID = self.GetTableInfo().Creator
	self.Rule.CreateType = self.FriendInfo.CreateType

	self.SerPay = 0 //好友房无茶水

	if len(strRule) == 0 {
		return
	}

	var _msg FriendRule_jlkj
	if err := json.Unmarshal(public.HF_Atobytes(strRule), &_msg); err == nil {
		if _msg.Radix == 0 {
			self.IBase = _msg.Difen // 级数底分
			self.Rule.Radix = 1
		} else {
			self.IBase = _msg.Difen // 级数底分
			self.Rule.Radix = _msg.Radix
		}
		self.SerPay = _msg.SerPay
		self.FaOfTao = _msg.Fa
		self.JiangOfTao = _msg.Jiang
		self.Rule.ShareScoreType = _msg.ShareScoreType //贡分类型
		self.Rule.Overtime_trust = _msg.Overtime_trust
		self.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		// 超时托管
		self.PlayTime = _msg.Overtime_trust
		self.Rule.NineSecondRoom = self.PlayTime > 0
		if _msg.Overtime_trust <= 0 {
			self.PlayTime = 15
		}
		// 4王换分
		self.FourKingScore = _msg.FourKingScore
		// 底分
		self.AddDiFen = _msg.AddDiFen
		// 满分奖励
		self.FullScoreAward = _msg.FullScoreAward
		// 是否显示手牌个数
		if strings.Contains(self.GetTableInfo().Config.GameConfig, "showhandcardcnt") {
			self.ShowHandCardCnt = _msg.ShowHandCardCnt == "true"
		} else {
			self.ShowHandCardCnt = true
		}
		// 是否可以捡尾分
		if strings.Contains(self.GetTableInfo().Config.GameConfig, "getlastscore") {
			self.GetLastScore = _msg.GetLastScore == "true"
		} else {
			self.GetLastScore = true
		}
		// 是否可以看队友牌
		if strings.Contains(self.GetTableInfo().Config.GameConfig, "seeteamercard") {
			self.SeeTeamerCard = _msg.SeeTeamerCard == "true"
		} else {
			self.SeeTeamerCard = true
		}
		//self.SeePartnerCards = _msg.SeePartnerCards == "true"
		self.SeePartnerCards = false
		self.Hard510KMode = _msg.Hard510KMode == "true"
	}
	if self.IBase < 1 {
		self.IBase = 1
	}
	//开关
	if self.Debug > 0 {
		//Rule.JuShu = Debug;
	}
}

// ! 开局
func (self *Game_jl_jlkj) OnBegin() {
	syslog.Logger().Info("onbegin")
	self.RepositTable(true)

	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range self.PlayerInfo {
		v.OnBegin()
	}

	//设置状态
	self.SetGameStatus(public.GS_MJ_PLAY)
	self.ParseRule(self.GetTableInfo().Config.GameConfig)
	self.m_GameLogic.Rule = self.Rule
	self.CurCompleteCount = 0
	self.VecGameEnd = []public.Msg_S_DG_GameEnd{}

	// 记录游戏开始时间
	self.GameCommon.GameBeginTime = time.Now()
	_, self.AllCards = self.m_GameLogic.CreateCards()
	//监利开机大小王不是癞子
	//self.m_GameLogic.InitMagicPoint(CARDINDEX_SMALL)//小王是赖子//王必须用牌数值
	//self.m_GameLogic.AddMagicPoint(CARDINDEX_BIG)//大王是赖子
	self.m_GameLogic.SetBombCount(3)                         //设置炸弹的最小长度
	self.m_GameLogic.SetOnestrCount(254)                     //设置单顺的最小长度，254表示无顺子,监利开机没有顺子
	self.m_GameLogic.SetMaxCardCount(public.MAX_CARD_4P)     //设置手牌最大长度
	self.m_GameLogic.SetMaxPlayerCount(public.MAX_PLAYER_4P) //设置玩家最大数目
	self.m_GameLogic.SetHard510KMode(self.Hard510KMode)      //设置纯510K是否大过四炸

	// 设置离线解散时间30分钟
	self.SetOfflineRoomTime(1800)

	self.OnGameStart()
}

func (self *Game_jl_jlkj) OnGameStart() {
	if !self.CanContinue() {
		return
	}
	self.StartNextGame()
}

// 开始下一局游戏
func (self *Game_jl_jlkj) StartNextGame() {
	self.OnStartNextGame()

	// 恢复自动解散时间2分钟
	self.SetDismissRoomTime(120)

	self.GameState = meta.GsNull           //吼牌阶段或打牌阶段.
	self.GameEndStatus = public.GS_MJ_PLAY //当前小局游戏的状态
	self.ReWriteRec = 0

	//发送最新状态
	for i := 0; i < self.GetPlayerCount(); i++ {
		self.SendUserStatus(i, public.US_PLAY) //把状态发给其他人
	}

	//记录日志
	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始StartNextGame......")
	self.OnWriteGameRecord(public.INVALID_CHAIR, self.GetTableInfo().Config.GameConfig)

	//重置所有玩家的状态
	for _, v := range self.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	//self.ParseRule(self.GameTable.Config.GameConfig)
	//self.m_GameLogic.Rule = self.Rule

	//设置状态
	self.SetGameStatus(public.GS_MJ_PLAY)

	// 框架发送开始游戏后开始计算当前这一轮的局数
	self.CurCompleteCount++

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(self.GetTableId()+self.KIND_ID*100+self.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	self.SiceCount = modules.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))

	//分发扑克
	self.ThePaiCount, self.PlayerCards, self.HasKingNum = self.m_GameLogic.RandCardData(self.AllCards)

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	self.initDebugCards("jianlikaiji_test", &self.PlayerCards, &self.Nextbanker, &self.DownPai)
	//////////////读取配置文件设置牌型end////////////////////////////////////

	//重新计算牌数目
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		self.ThePaiCount[i] = self.m_GameLogic.GetCardNum(self.PlayerCards[i], public.MAX_CARD_4P)
	}

	//发送玩家的牌数目
	if self.ShowHandCardCnt {
		self.SendPaiCount(public.MAX_PLAYER_4P)
	}

	//确定庄家，随机坐庄
	if self.Nextbanker >= public.MAX_PLAYER_4P {
		rand_num := rand.Intn(1000)
		self.Banker = uint16(rand_num % self.GetPlayerCount())
	} else {
		self.Banker = self.Nextbanker
	}
	self.Whoplay = self.Banker

	// 庄家有1王时 且队友没有王时 需要将另一个起到相同王（区分大小王）的玩家 换到庄对面 成为庄队友 使用GetTeamer需确保self.BankParter已经赋值
	tmpNextPlayer := self.GetNextFullSeat(self.Banker)
	bankerTeamer := self.GetNextFullSeat(tmpNextPlayer)
	kingCard := self.GetUserKingCard(self.Banker)
	if self.HasKingNum[self.Banker] == 1 && !self.IsHaveSameCard(kingCard, bankerTeamer) {
		bankerTeamerItem := self.GetUserItemByChair(bankerTeamer)
		// 寻找另一个有王的玩家
		for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
			if i == self.Banker || i == bankerTeamer {
				continue
			}
			if !self.IsHaveSameCard(kingCard, i) {
				continue
			}
			// 换桌
			userItem := self.GetUserItemByChair(i)
			if bankerTeamerItem != nil && userItem != nil {
				// 1 调换座位
				self.ExChangSeat(bankerTeamerItem.GetUserID(), userItem.GetUserID())
				// 2 交换手牌
				tmpCards := self.PlayerCards[bankerTeamer]
				self.PlayerCards[bankerTeamer] = self.PlayerCards[i]
				self.PlayerCards[i] = tmpCards
				// 记录信息
				ExChangStr := fmt.Sprintf("原庄家队友:%s 座位：%d 和 新庄家队友: %s 座位：%d 发生调换 ", bankerTeamerItem.Name, bankerTeamer, userItem.Name, i)
				self.OnWriteGameRecord(public.INVALID_CHAIR, ExChangStr)
				break
			}
		}

		// 更新HasKingNum
		for seat := 0; seat < public.MAX_PLAYER_4P; seat++ {
			self.HasKingNum[seat] = 0
			for i := 0; i < public.MAX_CARD_4P; i++ {
				if self.PlayerCards[seat][i] == logic.CARDINDEX_SMALL || self.PlayerCards[seat][i] == logic.CARDINDEX_BIG {
					self.HasKingNum[seat]++
				}
			}
		}
	}

	// 存在调换座位的情况下 会再次发送constant.MsgTypeTableInfo消息 字段Step当前第几局 需要在SetBegin里self.Step++之前发送 客户端的已有逻辑中会执行++
	self.GetTable().SetBegin(true)

	//这里检查天炸，7，8，喜
	if self.Rule.CardTypeScore {
		self.GetSpecialCardTypeScore()
	}

	for seat := 0; seat < public.MAX_PLAYER_4P; seat++ {
		//详细日志
		handCardStr := string("发牌后手牌:")
		for i := 0; i < public.MAX_CARD_4P; i++ {
			temCardStr := fmt.Sprintf("0x%02x,", self.switchCard2Ox(int(self.PlayerCards[seat][i])))
			handCardStr += temCardStr
		}
		self.OnWriteGameRecord(uint16(seat), handCardStr)

		self.ReplayRecord.R_HandCards[seat] = append(self.ReplayRecord.R_HandCards[seat], public.HF_BytesToInts(self.PlayerCards[seat][:])...)
	}

	//庄家所在的队为红队，反之为蓝队

	//构造数据,发送开始信息
	var GameStart public.Msg_S_DG_GameStart
	GameStart.BankerUser = self.Banker
	GameStart.CurrentUser = self.Whoplay

	//向每个玩家发送数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//设置变量
		GameStart.MySeat = uint16(i)
		//GameStart.Overtime = self.LimitTime
		GameStart.CardCount = self.ThePaiCount[i]
		for c := 0; c < public.MAX_CARD_4P; c++ {
			GameStart.CardData[c] = self.PlayerCards[i][c]
		}
		//记录玩家初始分

		//TODO 玩家分数设置
		self.ReplayRecord.R_Score[i] = self.GetUserTotalScore(uint16(i))

		//发送数据
		self.SendPersonMsg(constant.MsgTypeGameStart, GameStart, uint16(i))

		//发送队友的手牌
		if self.SeePartnerCards {
			teamer := self.GetTeamer(uint16(i))
			self.SendPaiToTeamer(teamer, uint16(i))
		}
	}

	// 检查4王情况
	FourKingUser := -1
	for seat := 0; seat < public.MAX_PLAYER_4P; seat++ {
		if self.HasKingNum[seat] == 4 {
			FourKingUser = seat
			break
		}
	}
	// 检查4王换分
	if self.FourKingScore > 0 && FourKingUser > -1 {
		self.GameState = meta.Gs4KingScore
		self.WhoHas4KingPower = uint16(FourKingUser)
		self.SendPower(uint16(FourKingUser), 4, self.PlayTime)
	} else {
		//self.StartRoar(self.Whoplay)
		//选明鸡牌找对家
		self.EndRoar(false)
	}

	//发送随机任务
	//self.SendTaskID(true,self.Whoplay);
	self.GameTask.SendTaskID(self.GameCommon, true, self.Whoplay)
}

// 发送操作
func (self *Game_jl_jlkj) SendPaiCount(wChairID uint16) {

	//构造数据
	var SendCount public.Msg_S_DG_SendCount
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		SendCount.CardCount[i] = self.ThePaiCount[i]
	}
	//发送数据
	if wChairID >= public.MAX_PLAYER_4P {
		self.SendTableMsg(constant.MsgTypeSendPaiCount, SendCount)
	} else {
		for _, v := range self.PlayerInfo {
			if v.Seat == uint16(wChairID) {
				self.SendPersonMsg(constant.MsgTypeSendPaiCount, SendCount, v.Seat)
			}
		}
	}
}

// 发送本轮分
func (self *Game_jl_jlkj) SendTurnScore(wChairID uint16) {
	var turnScore public.Msg_S_DG_TurnScore
	turnScore.TurnScore = self.CardScore

	if wChairID >= public.MAX_PLAYER_4P {
		self.SendTableMsg(constant.MsgTypeTurnScore, turnScore)
	} else {
		self.SendPersonMsg(constant.MsgTypeTurnScore, turnScore, wChairID)
	}
}

// 发送玩家分
func (self *Game_jl_jlkj) SendPlayerScore(wChairID uint16, wGetChairID uint16, iGetScore int) {
	var playScore public.Msg_S_DG_PlayScore
	playScore.ChairID = wGetChairID
	playScore.GetScore = iGetScore

	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		playScore.PlayScore[i] = self.PlayerCardScore[i]
	}

	if wChairID >= public.MAX_PLAYER_4P {
		self.SendTableMsg(constant.MsgTypePlayScore, playScore)
	} else {
		self.SendPersonMsg(constant.MsgTypePlayScore, playScore, wChairID)
	}
}

// 发送声音
func (self *Game_jl_jlkj) SendPlaySoundMsg(seat uint16, bySoundType byte) {
	var soundmsg public.Msg_S_DG_PlaySound
	soundmsg.CurrentUser = seat
	soundmsg.SoundType = bySoundType

	self.SendTableMsg(constant.MsgTypePlaySound, soundmsg)
}

// 发送玩家是几游
func (self *Game_jl_jlkj) SendPlayerTurn(wChairID uint16) {
	var turnmsg public.Msg_S_DG_SendTurn

	///////////////////////////////////////////////////////////////////////////////////////////////////////
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		turnmsg.Turn[i] = self.PlayerTurn[i]
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	if wChairID >= public.MAX_PLAYER_4P {
		self.SendTableMsg(constant.MsgTypeSendTurn, turnmsg)
	} else {
		self.SendPersonMsg(constant.MsgTypeSendTurn, turnmsg, wChairID)
	}
}

// 把seat1的牌发给seat2
func (self *Game_jl_jlkj) SendPaiToTeamer(seat1 uint16, seat2 uint16) {
	if seat1 < 0 || seat1 >= public.MAX_PLAYER_4P {
		return
	}
	if seat2 < 0 || seat2 >= public.MAX_PLAYER_4P {
		return
	}

	//if(!self.BMingJiFlag){return}//没有明鸡，不发送队友的牌

	var teamerPai public.Msg_S_DG_TeamerPai
	teamerPai.WhoPai = seat1
	for i := 0; i < public.MAX_CARD_4P; i++ {
		if self.m_GameLogic.IsValidCard(self.PlayerCards[seat1][i]) {
			teamerPai.CardData[teamerPai.CardCount] = self.PlayerCards[seat1][i]
			teamerPai.CardCount++
		}
	}
	self.SendPersonMsg(constant.MsgTypeTeamerPai, teamerPai, seat2)
}

// 发送权限
func (self *Game_jl_jlkj) SendPower(whoplay uint16, iPower int, iWaitTime int) {
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	//构造数据
	var power public.Msg_S_DG_Power
	iOvertime := 0
	if iWaitTime > 0 {
		iOvertime = iWaitTime
	}
	power.CurrentUser = whoplay
	power.Power = iPower
	if self.TuoGuanPlayer[whoplay] {
		if iOvertime > self.AutoOutTime {
			iOvertime = self.AutoOutTime
		}
	}
	power.Overtime = int64(iOvertime)

	//详细日志
	LogStr := fmt.Sprintf("SUB_S_POWER seat=%d power=%d,time=%d ", power.CurrentUser, power.Power, power.Overtime)
	self.OnWriteGameRecord(power.CurrentUser, LogStr)

	self.SendTableMsg(constant.MsgTypeSendPower, power)

	self.PowerStartTime = time.Now().Unix() //权限开始时间
	self.setLimitedTime(int64(iOvertime + 1))
	if self.GameState == meta.GsPlay {
		//SetActionStep(AS_PLAY,nTime + 1);//设置等待时间，服务端多等一下
	} else if self.GameState == meta.GsRoarPai {
		//SetActionStep(AS_ROAR,nTime + 1);//设置等待时间，服务端多等一下
	} else if self.GameState == meta.Gs4KingScore {
		//SetActionStep(AS_ROAR,nTime + 1);//设置等待时间，服务端多等一下
	}
}

func (self *Game_jl_jlkj) setLimitedTime(iLimitTime int64) {
	// fmt.Println(fmt.Sprintf("limitetimeOP(%d)", self.Rule.limitetimeOP))
	self.LimitTime = time.Now().Unix() + iLimitTime
	self.GameTimer.SetLimitTimer(int(iLimitTime))
}

func (self *Game_jl_jlkj) freeLimitedTime() {
	self.GameTimer.KillLimitTimer()
}

func (self *Game_jl_jlkj) LockTimeOut(cUser uint16, iTime int64) {
	if cUser < 0 || cUser > uint16(self.GetPlayerCount()) {
		return
	}

	_userItem := self.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = iTime
}

func (self *Game_jl_jlkj) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(self.GetPlayerCount()) {
		return
	}

	_userItem := self.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0
}

// 计时器事件
func (self *Game_jl_jlkj) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	//游戏定时器
	if dwTimerID == modules.GameTime_Nine {
		if self.Rule.NineSecondRoom {
			self.OnAutoOperate(true)
		}
	}
	return true
}

// 暂时空着
func (self *Game_jl_jlkj) OnAutoOperate(bBreakin bool) {
	fmt.Println("自动操作")
	//详细日志
	LogStr := string("OnAutoOperate 自动操作!!! ")
	self.OnWriteGameRecord(self.Whoplay, LogStr)

	if self.GameState == meta.GsRoarPai {
		if self.Whoplay < public.MAX_PLAYER_4P && !self.TuoGuanPlayer[self.Whoplay] {
			self.AutoTuoGuan(self.Whoplay)
		}
		self.OnRoarAction(self.Whoplay, false)
	} else if self.GameState == meta.GsPlay {
		tempCurPlay := self.Whoplay

		if self.Whoplay < public.MAX_PLAYER_4P && !self.TuoGuanPlayer[self.Whoplay] {
			self.AutoTuoGuan(self.Whoplay)
		}

		//构造数据
		var outmsg public.Msg_C_DG_OutCard
		outmsg.CurrentUser = self.Whoplay
		if self.WhoLastOut >= public.MAX_PLAYER_4P {
			if true {
				//c++蕲春的做法
				buf := [public.MAX_CARD]byte{}
				n := 0
				for i := 0; i < public.MAX_CARD_4P; i++ {
					if self.m_GameLogic.IsValidCard(self.PlayerCards[self.Whoplay][i]) {
						buf[n] = self.PlayerCards[self.Whoplay][i]
						n++
					}
				}
				buf = self.m_GameLogic.SortByIndex(buf, public.MAX_CARD_4P, true)
				outmsg.CardCount = 1
				outmsg.CardData[0] = buf[0]
				outmsg.CardType = public.TYPE_ONE
			} else {
				//智能一点的做法
				self.m_GameLogic.GetGroupType(self.PlayerCards[self.Whoplay])
				_, beepOut := self.m_GameLogic.BeepFirstCardOut()
				for k := 0; k < len(beepOut[0].Indexes); k++ {
					outmsg.CardData[k] = beepOut[0].Indexes[k]
				}
				outmsg.CardCount = beepOut[0].Count
				outmsg.CardType = self.LastOutType
			}
		} else {
			outmsg.CardCount = 0
			if false {
				//如果跟出要出牌就需要用下面的牌
				self.m_GameLogic.GetGroupType(self.PlayerCards[self.Whoplay])
				_, beepOut := self.m_GameLogic.BeepCardOut(self.AllPaiOut[self.WhoLastOut], self.LastOutType)
				if len(beepOut) == 0 {
					outmsg.CardCount = 0
				} else {
					for k := 0; k < len(beepOut[0].Indexes); k++ {
						outmsg.CardData[k] = beepOut[0].Indexes[k]
					}
					outmsg.CardCount = beepOut[0].Count
					outmsg.CardType = self.LastOutType
				}
			}
		}
		//详细日志
		LogStr := fmt.Sprintf("托管出牌 OnAutoOperate UserID=%d ,CardCount=%d,牌数据:", outmsg.CurrentUser, outmsg.CardCount)
		for i := 0; i < public.MAX_CARD_4P; i++ {
			if outmsg.CardData[i] > 0 {
				CardStr := fmt.Sprintf("0x%02x,", self.switchCard2Ox(int(outmsg.CardData[i])))
				LogStr += CardStr
			}
		}
		self.OnWriteGameRecord(self.Whoplay, LogStr)
		self.OnUserOutCard(&outmsg)

		self.AutoCardCounts[tempCurPlay]++
		if self.AutoCardCounts[tempCurPlay] >= 5 {
			// 连续5轮自动出牌，结束游戏
			//self.OnGameEndUserLeft(tempCurPlay, meta.GOT_TUOGUAN)
		}
	} else if self.GameState == meta.Gs4KingScore {
		if self.WhoHas4KingPower == public.INVALID_CHAIR {
			return
		}

		// 加入托管
		if self.WhoHas4KingPower < public.MAX_PLAYER_4P && !self.TuoGuanPlayer[self.WhoHas4KingPower] {
			self.AutoTuoGuan(self.WhoHas4KingPower)
		}

		// 不换分
		self.On4KingScore(self.WhoHas4KingPower, 0)
		// 日志
		self.OnWriteGameRecord(self.WhoHas4KingPower, "托管4王换分 弃")
	}
}

func (self *Game_jl_jlkj) AutoTuoGuan(theSeat uint16) int {
	var msgtg public.Msg_S_DG_Trustee
	msgtg.Trustee = true
	msgtg.ChairID = theSeat

	if self.GameState == meta.GsNull || theSeat >= public.MAX_PLAYER_4P {
		//游戏状态 GameState，1表示吼牌阶段，2表示打牌阶段
		return 0
	}
	//详细日志
	LogStr := fmt.Sprintf("超时托管,CMD_S_Tuoguan_CB AutoTuoGuan msgtg.theFlag=%t msgtg.theSeat=%d ", msgtg.Trustee, msgtg.ChairID)
	self.OnWriteGameRecord(theSeat, LogStr)

	if self.GameState == meta.GsPlay || self.GameState == meta.GsRoarPai || self.GameState == meta.Gs4KingScore {
		if true == msgtg.Trustee {
			self.TuoGuanPlayer[theSeat] = true
			self.TrustCounts[theSeat]++
			isTrust := false
			for _, val := range self.TrustPlayer {
				if val == byte(theSeat) {
					isTrust = true
				}
			}
			if !isTrust {
				self.TrustPlayer = append(self.TrustPlayer, byte(theSeat))
			}
			if theSeat == self.Whoplay { //如果是当前的玩家，那么重新设置一下开始时间
				//self.setLimitedTime(int64(self.AutoOutTime))//已经超时了，马上就要切换牌权了，不用在设置他的时间了
			}
			self.SendTableMsg(constant.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			self.addReplayOrder(msgtg.ChairID, meta.DG_REPLAY_OPT_TUOGUAN, 1, []int{})
		} else {
			self.TuoGuanPlayer[theSeat] = false
			self.SendTableMsg(constant.MsgTypeGameTrustee, msgtg)
			index := -1
			for i, val := range self.TrustPlayer {
				if val == byte(theSeat) {
					index = i
				}
			}
			if index > -1 {
				self.TrustPlayer = append(self.TrustPlayer[:index], self.TrustPlayer[index+1:]...)
			}
			//回放托管记录
			self.addReplayOrder(msgtg.ChairID, meta.DG_REPLAY_OPT_TUOGUAN, 0, []int{})
		}
	}
	return 1
}

// 硬牌动作响应
func (self *Game_jl_jlkj) OnRoarAction(seat uint16, bRoar bool) bool {
	if seat < 0 || seat >= public.MAX_PLAYER_4P {
		return false
	}
	if seat != self.Whoplay {
		return false
	}

	if self.WhoReady[seat] {
		return false
	}

	self.WhoReady[seat] = true

	//变量定义
	var roar public.Msg_S_DG_Roar
	roar.CurrentUser = seat
	roar.RoarFlag = 0
	if bRoar {
		roar.RoarFlag = 1
	}

	self.SendTableMsg(constant.MsgTypeRoar, roar)

	//回放增加吼牌记录
	self.addReplayOrder(seat, meta.DG_REPLAY_OPT_HOUPAI, int(roar.RoarFlag), []int{})

	if bRoar { //如果硬牌，那么结束硬牌动作
		self.EndRoar(true)
		return true
	} else {
		z := 0
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if self.WhoReady[i] {
				z++
			}
		}
		if z >= public.MAX_PLAYER_4P { //开始游戏
			self.EndRoar(false)
		} else {
			self.GoNextPlayer()
		}
	}
	return true
}

// 4王换分
func (self *Game_jl_jlkj) On4KingScore(seat uint16, flag byte) bool {
	if seat < 0 || seat >= public.MAX_PLAYER_4P {
		return false
	}

	// 没有换分牌权
	if seat != self.WhoHas4KingPower {
		return false
	}

	// 已经换过了
	if self.WhoHas4KingScore != public.INVALID_CHAIR {
		return false
	}

	// 重置换分牌权
	self.WhoHas4KingPower = public.INVALID_CHAIR

	if flag == 1 {
		// 打出4个王 从手牌里删除
		idx := 0
		kingCardData := [40]byte{}
		for i := 0; i < public.MAX_CARD_4P; i++ {
			if self.PlayerCards[seat][i] == logic.CARDINDEX_SMALL || self.PlayerCards[seat][i] == logic.CARDINDEX_BIG {
				kingCardData[idx] = self.PlayerCards[seat][i]
				self.PlayerCards[seat][i] = 0
				idx++
			}
		}

		// 出牌消息
		var msgout public.Msg_S_DG_OutCard
		msgout.CardCount = byte(idx)
		msgout.CardType = public.TYPE_4KING_SCORE
		msgout.CurrentUser = seat
		msgout.Overtime = 15
		msgout.ByClient = false
		msgout.CardData = kingCardData
		msgout.OutScorePai = self.OutScorePai //所有分牌
		self.SendTableMsg(constant.MsgTypeGameOutCard, msgout)

		// 记录出牌信息  需求 断线重连不显示打出的4王
		/*
			for i := 0; i < public.MAX_CARD_4P; i++ {
				self.AllPaiOut[seat][i] = kingCardData[i]
				self.LastPaiOut[seat][i] = kingCardData[i]
			}
		*/
		//self.WhoLastOut = seat // 会影响首出牌 //self.WhoLastOut >= public.MAX_PLAYER_4P
		self.ThePaiCount[seat] -= byte(idx)
		self.LastOutType = public.TYPE_4KING_SCORE
		self.LastOutTypeClient = public.TYPE_4KING_SCORE

		// 置换分数
		self.PlayerCardScore[seat] += self.FourKingScore
		self.SendPlayerScore(public.MAX_PLAYER_4P, seat, self.FourKingScore)
		self.SendTurnScore(public.MAX_PLAYER_4P)

		// 发送手牌数量
		if self.ShowHandCardCnt {
			self.SendPaiCount(public.MAX_PLAYER_4P)
		}

		// 记录谁换了分
		self.WhoHas4KingScore = seat

		// 回放记录4王换分
		self.addReplayOrder(seat, meta.DG_REPLAY_OPT_4KINGSCORE, self.FourKingScore, public.HF_BytesToInts(kingCardData[:]))
	}

	// 转发4王换分消息
	var msg Msg_S_4KingScore
	msg.CurrentUser = seat
	msg.Flag = flag
	self.SendTableMsg(constant.MsgType4KingScore, msg)

	// 开始游戏
	self.EndRoar(false)

	return true
}

func (self *Game_jl_jlkj) StartPlay(whoplay uint16) {
	self.CardScore = 0
	// 开始
	self.GameState = meta.GsPlay

	iPower := 2
	self.SendPower(whoplay, iPower, self.PlayTime+5) //庄家第一次出牌的时间加5秒
}

func (self *Game_jl_jlkj) StartRoar(theSeat uint16) {
	// 开始进入吼牌状态
	self.GameState = meta.GsRoarPai
	iPower := 1
	self.SendPower(theSeat, iPower, self.RoarTime)
}

func (self *Game_jl_jlkj) EndRoar(bRoar bool) {
	//var endroarmsg public.Msg_S_DG_EndRoar

	////有人吼牌
	//if bRoar {
	//	self.WhoRoar = self.Whoplay
	//	endroarmsg.RoarUser = self.WhoRoar
	//	self.GameType = GT_ROAR
	//} else {
	///////////////////////////////////////////////////////////////////////////////////////////
	//	//没人吼牌
	self.GameType = meta.GT_NORMAL
	self.WhoRoar = public.INVALID_CHAIR

	tmp1 := self.GetNextFullSeat(self.Banker)
	tmp2 := self.GetNextFullSeat(tmp1)
	self.BankParter = tmp2
	//	endroarmsg.RoarUser = self.WhoRoar
	//	self.GetJiaoPai() //得到叫牌
	//////////////////////////////////////////////////////////////////////////////////////////
	//}
	////吼牌的为庄家了
	//if bRoar {
	//	self.Banker = self.WhoRoar
	//	self.BankParter = public.INVALID_CHAIR
	//	//详细日志
	//	LogStr := string("包牌(吼牌)")
	//	self.OnWriteGameRecord(self.WhoRoar, LogStr)
	//}
	//////////////////////////////////////////////////////////////////////////////////////////
	self.Whoplay = self.Banker
	//endroarmsg.BankUser = self.Banker
	//endroarmsg.JiaoPai = self.RoarPai
	//self.SendTableMsg(constant.MsgTypeEndRoar, endroarmsg)
	//////////////////////////////////////////////////////////////////////////////////////////
	//回放增加结束吼牌记录
	//self.addReplayOrder(self.Banker, DG_REPLAY_OPT_END_HOUPAI, int(self.RoarPai), []int{})

	//详细日志
	LogStr := string("为庄家")
	self.OnWriteGameRecord(self.Whoplay, LogStr)

	self.StartPlay(self.Whoplay)
}

func (self *Game_jl_jlkj) GoNextPlayer() {
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	for iPlayer := 0; iPlayer < public.MAX_PLAYER_4P; iPlayer++ {
		if self.Whoplay >= public.MAX_PLAYER_4P-1 {
			self.Whoplay = 0
		} else {
			self.Whoplay++
		}

		//如果当前玩家出完了
		if self.WhoAllOutted[self.Whoplay] {
			if self.WhoLastOut == self.Whoplay { //这个玩家是不是上一次出牌玩家
				break
			} else {
				continue
			}
		} else { //没出完？
			break
		}
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////
	if self.WhoLastOut == self.Whoplay {
		if self.EndTurn() {
			return
		}

	}
	if self.GameState == meta.GsRoarPai {
		self.StartRoar(self.Whoplay)
	} else if self.GameState == meta.GsPlay {
		dwPower := 2
		self.SendPower(self.Whoplay, dwPower, self.PlayTime)
	}

	for i := 0; i < public.MAX_CARD_4P; i++ {
		self.AllPaiOut[self.Whoplay][i] = 0
	}
}

// 得到叫牌
func (self *Game_jl_jlkj) GetJiaoPai() byte {
	/*
		明鸡：拥有与庄家叫牌相同牌的玩家是庄家的对家，在牌局开始时由系统直接明鸡；如果随机牌的花色2张都在庄家手上，则庄家正对面的玩家为庄家的朋友。
	*/
	//从庄家手牌中随机抽出一张当作明鸡牌
	seed := rand.Intn(10000) % public.MAX_CARD_4P
	self.RoarPai = self.PlayerCards[self.Banker][seed]
	//做牌文件处理
	if self.DownPai != 0 {
		self.RoarPai = self.DownPai
	}

	num := 0
	for j := 0; j < public.MAX_CARD_4P; j++ {
		if self.PlayerCards[self.Banker][j] == self.RoarPai {
			num++
		}
	}
	//如果随机牌的花色2张都在庄家手上，则庄家正对面的玩家为庄家的朋友。
	if num == 2 {
		//跳2个玩家
		tmp1 := self.GetNextFullSeat(self.Banker)
		tmp2 := self.GetNextFullSeat(tmp1)
		self.BankParter = tmp2
	} else {
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if uint16(i) == self.Banker {
				continue
			}
			for j := 0; j < public.MAX_CARD_4P; j++ {
				if self.PlayerCards[i][j] == self.RoarPai {
					self.BankParter = uint16(i) //明鸡
					break
				}
			}
			if self.BankParter < public.MAX_PLAYER_4P {
				break
			}
		}

	}

	//详细日志
	LogStr := fmt.Sprintf("明鸡牌为:0x%02x ,庄家队友为%d", self.switchCard2Ox(int(self.RoarPai)), self.BankParter)
	self.OnWriteGameRecord(self.Banker, LogStr)
	return self.RoarPai
}

// 结束一轮
func (self *Game_jl_jlkj) EndTurn() bool {
	for i := 0; i < public.MAX_CARD_4P; i++ {
		self.LastPaiOut[self.WhoLastOut][i] = self.AllPaiOut[self.WhoLastOut][i]
	}
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		for j := 0; j < public.MAX_CARD_4P; j++ {
			self.AllPaiOut[i][j] = 0
		}
		self.WhoPass[i] = false
	}
	//打分的模式
	if self.GameType == meta.GT_NORMAL {
		//回放增加本轮抓分
		if self.CardScore > 0 {
			self.addReplayOrder(self.WhoLastOut, meta.DG_REPLAY_OPT_TURN_OVER, self.CardScore, []int{})
		} else {
			self.addReplayOrder(self.WhoLastOut, meta.DG_REPLAY_OPT_TURN_OVER, 0, []int{})
		}
		self.PlayerCardScore[self.WhoLastOut] += self.CardScore
		self.SendPlayerScore(public.MAX_PLAYER_4P, self.WhoLastOut, self.CardScore)
		self.CardScore = 0 //清零
		self.OutScorePai = [24]byte{}
		self.SendTurnScore(public.MAX_PLAYER_4P)
		////这里可以判断下是否结束游戏
		//if self.JudgeEndGame(self.WhoLastOut) {
		//	return true
		//}
	} else {
		self.addReplayOrder(self.WhoLastOut, meta.DG_REPLAY_OPT_TURN_OVER, 0, []int{})
	}

	self.WhoLastOut = public.INVALID_CHAIR

	if self.WhoAllOutted[self.Whoplay] { //这里说明当前玩家出完了
		teamer := self.GetTeamer(self.Whoplay)

		//if self.BMingJiFlag { //如果明鸡确定，那么是该玩家的队友接风
		self.Whoplay = teamer
		self.SendPlaySoundMsg(self.Whoplay, public.TY_JieFeng)
		//} else { //否则就是下家接风
		//
		//}
	}
	for {
		if self.WhoAllOutted[self.Whoplay] {
			if self.Whoplay >= public.MAX_PLAYER_4P-1 {
				self.Whoplay = 0
			} else {
				self.Whoplay++
			}
		} else {
			break
		}
	}

	self.EndOut() //结束一轮
	return false
}
func (self *Game_jl_jlkj) EndOut() {
	self.LastOutType = public.TYPE_NULL
	self.LastOutTypeClient = public.TYPE_NULL

	var endout public.Msg_S_DG_EndOut
	endout.CurrentUser = self.Whoplay
	self.SendTableMsg(constant.MsgTypeEndOut, endout)
}

func (self *Game_jl_jlkj) GetTeamer(who uint16) uint16 {
	//	re := uint16(0)
	if who >= public.MAX_PLAYER_4P {
		return 0
	}
	return (who + 2) % meta.MAX_PLAYER
	//if who == self.Banker {
	//	re = self.BankParter
	//} else if who == self.BankParter {
	//	re = self.Banker
	//} else { //闲家
	//	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
	//		if i == who {
	//			continue
	//		}
	//		if i == self.BankParter {
	//			continue
	//		}
	//		if i == self.Banker {
	//			continue
	//		}
	//		re = i
	//		break
	//	}
	//}
	//return re
}
func (self *Game_jl_jlkj) GetTeamScore(seat uint16) int {
	if self.GameType != meta.GT_NORMAL {
		return 0
	}
	if seat < 0 || seat >= public.MAX_PLAYER_4P {
		return 0
	}
	teamer := self.GetTeamer(seat)
	if self.PlayerTurn[seat] == 1 || self.PlayerTurn[teamer] == 1 { //一游玩家的分
		return self.PlayerCardScore[seat] + self.PlayerCardScore[teamer]
	} else if self.PlayerTurn[seat] == 2 || self.PlayerTurn[teamer] == 2 {
		who2you := teamer
		if self.PlayerTurn[seat] == 2 {
			who2you = seat
		}
		return self.PlayerCardScore[who2you]
	}
	return 0
}
func (self *Game_jl_jlkj) JudgeEndGame(who uint16) bool {
	//提前结束游戏的规则，
	//20190216 这里不能按照队伍来算，只能算1游和2游自己的分数
	if who >= public.MAX_PLAYER_4P {
		return false
	}

	seat1 := public.INVALID_CHAIR
	seat2 := public.INVALID_CHAIR
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		if 1 == self.PlayerTurn[i] {
			seat1 = i
		}
		if 2 == self.PlayerTurn[i] {
			seat2 = i
		}
	}
	if seat1 >= public.MAX_PLAYER_4P || seat2 >= public.MAX_PLAYER_4P {
		return false
	}

	num := 0
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		if self.WhoAllOutted[i] {
			num++
		}
	}
	if num != 2 {
		return false //只有在其中两个玩家牌出完的情况下，才有可能中途结束游戏
	}
	//seat3和seat4的牌没有出完，否则不会走到这里
	seat3 := public.INVALID_CHAIR
	seat4 := public.INVALID_CHAIR
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		if i == seat1 || i == seat2 {
			continue
		}
		if public.INVALID_CHAIR == seat3 {
			seat3 = i
		} else {
			seat4 = i
			break
		}
	}

	if seat3 == seat1 || seat3 == seat2 {
		return false //seat3不会是1游，否则只有一个人出完牌；seat3不会是seat2，否则就是1游的队友，就是双杀不再这里处理。
	}
	score1 := self.PlayerCardScore[seat1]  //一游分数
	score2 := self.PlayerCardScore[seat2]  //二游分数
	score3 := self.PlayerCardScore[seat3]  //其他人分数
	score4 := self.PlayerCardScore[seat4]  //其他人分数
	score1Team := self.GetTeamScore(seat1) //一游和二游肯定不是一队的
	//score2Team := self.GetTeamScore(seat2)//一游和二游肯定不是一队的

	//一游200分
	if score1 >= 200 { //结束游戏
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if self.PlayerTurn[i] == 1 {
				self.Nextbanker = uint16(i)
				break
			}
		}
		self.OnGameEndNormal(seat1, meta.GOT_ZHONGTU)
		return true
	}
	//二游200分
	if score2 >= 200 { //结束游戏
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if self.PlayerTurn[i] == 1 {
				self.Nextbanker = uint16(i)
				break
			}
		}
		self.OnGameEndNormal(seat2, meta.GOT_ZHONGTU)
		return true
	}
	//其他人200分
	if score3 >= 200 || score4 >= 200 {
		//有200分的人不管是谁的队友，牌没有出完，要等牌出完才能结束
		return false
	}

	//20190216，一游分数大于等于100，算赢
	if score1 >= 100 && score2 >= 5 { //二游的队友没有确定是否是四游，所以不能算二游队友的分数
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if self.PlayerTurn[i] == 1 {
				self.Nextbanker = uint16(i)
				break
			}
		}
		self.OnGameEndNormal(seat1, meta.GOT_ZHONGTU)
		return true
	}
	//20190218，一游如果是0分，需要确定1游队友是否是4游，这个时候不能提前结束
	if score1 > 0 && score1Team >= 100 && score2 >= 5 { //结束游戏
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if self.PlayerTurn[i] == 1 {
				self.Nextbanker = uint16(i)
				break
			}
		}
		self.OnGameEndNormal(seat1, meta.GOT_ZHONGTU)
		return true
	}
	//20190216，二游分数大于100，才算赢，由于二游的队友不知道是否是4游，不能把二游的队友分数算进去
	//由于一游的队友不知道是否是4游，不能把一游的队友分数算进去
	if score2 > 100 && score1 >= 5 { //这里必须是1游有分，如果1游0分，可能1游会输2倍
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if self.PlayerTurn[i] == 1 {
				self.Nextbanker = uint16(i)
				break
			}
		}
		self.OnGameEndNormal(seat2, meta.GOT_ZHONGTU)
		return true
	}

	return false
}

func (self *Game_jl_jlkj) GetFinalWinLoseScore(score *[meta.MAX_PLAYER]int) {
	wintotal := 0
	losetotal := 0
	total := [meta.MAX_PLAYER]int{}
	n := 0
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		total[i] = score[i]
		if score[i] <= 0 {
			losetotal += score[i]
		} else {
			n++
			wintotal += score[i]
		}
	}
	if n <= 0 {
		return
	}
	if wintotal > -losetotal { //扣的分比加的分少，那么只能平分了
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			if score[i] > 0 {
				total[i] = -int(float64(losetotal) * float64(score[i]) / (float64(wintotal)))
			}
		}
	}
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		score[i] = total[i]
	}
}

func (self *Game_jl_jlkj) AddSpecailScore(Score *[meta.MAX_PLAYER]int, seat uint16, base int) {
	if seat < 0 && seat >= public.MAX_PLAYER_4P {
		return
	}

	iAddFan := self.WhoSame510K[seat] + self.Who7Xi[seat] + self.Who8Xi[seat] + self.PlayKingBomb[seat]
	iTempScore := iAddFan * base

	//这里来计算各自该赢的，特殊的分
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		if uint16(i) == seat {
			self.XiScore[i] += 3 * iTempScore
			Score[i] += 3 * iTempScore
		} else {
			self.XiScore[i] -= iTempScore //其他人
			Score[i] -= iTempScore
		}
	}
}

// ! 加载测试麻将数据
func (self *Game_jl_jlkj) initDebugCards(configName string, cbRepertoryCard *[meta.MAX_PLAYER][public.MAX_CARD]byte, wBankerUser *uint16, byDownPai *byte) (err error) {
	defer func() {
		if err != nil {
			self.OnWriteGameRecord(public.INVALID_CHAIR, err.Error())
		}
	}()
	//! 做牌文件配置
	var debugCardConfig *meta.CardConfig = new(meta.CardConfig)

	configName = fmt.Sprintf("./%s%d", configName, self.GetPlayerCount())
	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始读取做牌文件，文件名："+configName)
	if !public.GetJsonMgr().ReadData("./json", configName, debugCardConfig) {
		return errors.New("做牌文件:读取失败")
	}

	// 是否开启做牌
	if debugCardConfig.IsAble == 1 {
		//检查做牌文件是否做牌异常
		for _, handCards := range debugCardConfig.UserCards {
			if len(strings.Split(handCards, ",")) != 27 {
				return errors.New("做牌文件:手牌长度不为27")
			}
		}
		//检查牌堆牌是否正常
		if len(debugCardConfig.RepertoryCard) != debugCardConfig.RepertoryCardCount*5-1 {
			return errors.New(fmt.Sprintf("做牌文件:牌库牌数量不一致:::RepertoryCard:[%d]>>>实际做牌牌库数量:[%d]", debugCardConfig.RepertoryCardCount*5-1, len(debugCardConfig.RepertoryCard)))
		}

		// 重置HasKingNum
		for i := 0; i < public.MAX_PLAYER_4P; i++ {
			self.HasKingNum[i] = 0
		}

		// 设置玩家手牌
		for userIndex, handCards := range debugCardConfig.UserCards {
			byCardsCount := byte(0)
			_item := self.GetUserItemByChair(uint16(userIndex))
			if _item != nil {
				_item.Ctx.InitCardIndex() //清理手牌
				handCardsSlice := strings.Split(handCards, ",")
				for _, cardStr := range handCardsSlice {

					if _, cardValue := self.GetCardDataByStr(cardStr); cardValue == public.INVALID_BYTE {
						//return errors.New(fmt.Sprintf("做牌文件:第%d号玩家做牌异常：%s", userIndex, cardStr))
					} else {
						//_item.Ctx.DispatchCard(cardValue)
						(*cbRepertoryCard)[userIndex][byCardsCount] = cardValue
						byCardsCount++
						if cardValue == logic.CARDINDEX_SMALL || cardValue == logic.CARDINDEX_BIG {
							self.HasKingNum[userIndex]++
						}
						if byCardsCount >= public.MAX_CARD_4P {
							break
						}
					}
				}
			}
		}
		//设置牌堆牌
		repertoryCards := strings.Split(debugCardConfig.RepertoryCard, ",")
		for _, cardStr := range repertoryCards {
			if _, cardValue := self.GetCardDataByStr(cardStr); cardValue == public.INVALID_BYTE {
				//return errors.New(fmt.Sprintf("做牌文件:牌堆第%d个做牌异常：%s", cardIndex, cardStr))
			} else {
				(*byDownPai) = cardValue
				break //底牌里面的叫牌只能有一个
			}
		}
		// 设置庄家
		if debugCardConfig.BankerUserSeatId != -1 {
			(*wBankerUser) = uint16(debugCardConfig.BankerUserSeatId)
		}
	} else {
		return errors.New("做牌文件:开关未开启")
	}
	return err
}

// 写日志记录
func (self *Game_jl_jlkj) writeGameRecord() {
	//写日志记录
	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始蕲春打拱  发牌......")

	// 玩家手牌
	//for _, v := range self.PlayerInfo {
	//	if v.Seat != public.INVALID_CHAIR {
	//		handCardStr := fmt.Sprintf("发牌后手牌:%s", self.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
	//		self.OnWriteGameRecord(uint16(v.Seat), handCardStr)
	//	}
	//}

	// 牌堆牌
	//leftCardStr := fmt.Sprintf("牌堆牌:%s", self.m_GameLogic.SwitchToCardNameByDatas(self.RepertoryCard[0:self.LeftCardCount+2], 0))
	//self.OnWriteGameRecord(public.INVALID_CHAIR, leftCardStr)

	//赖子牌
	//magicCardStr := fmt.Sprintf("癞子牌:%s", self.m_GameLogic.SwitchToCardNameByData(self.MagicCard, 1))
	//self.OnWriteGameRecord(public.INVALID_CHAIR, magicCardStr)
}

// ! 解散
func (self *Game_jl_jlkj) OnEnd() {
	if self.IsGameStarted() {
		self.OnGameOver(public.INVALID_CHAIR, public.GER_DISMISS)
	}
}

// ! 单局结算
func (self *Game_jl_jlkj) OnGameOver(wChairID uint16, cbReason byte) bool {
	self.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 初始化游戏
func (self *Game_jl_jlkj) OnInit(table base.TableBase) {
	self.KIND_ID = table.GetTableInfo().KindId
	self.Config.StartMode = public.StartMode_FullReady
	self.Config.PlayerCount = 4 //玩家人数
	self.Config.ChairCount = 4  //椅子数量
	self.PlayerInfo = make(map[int64]*modules.Player)

	self.RepositTable(true)
	self.SetGameStartMode(public.StartMode_FullReady)
	self.GameTable = table
	self.Init()
	self.Unmarsha(table.GetTableInfo().GameInfo)
	table.GetTableInfo().GameInfo = ""

	// 设置自动解散时间2分钟
	self.SetDismissRoomTime(120)
	// 设置离线解散时间30分钟
	self.SetOfflineRoomTime(1800)
	// 离线60s未准备踢出
	if self.CurCompleteCount == 0 && self.GameTable.GetTableInfo().JoinType == constant.NoCheat {
		self.SetOfflineRoomTime(60)
	}

}

// ! 发送消息
func (self *Game_jl_jlkj) OnMsg(msg *base.TableMsg) bool {
	wChairID := self.GetChairByUid(msg.Uid)
	if wChairID == public.INVALID_CHAIR {
		self.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::找不到玩家的座位号:%d", msg.Uid))
		return true
	}

	switch msg.Head {
	case constant.MsgTypeGameBalanceGameReq: //! 请求总结算信息 //暂时没有

		var _msg public.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.CalculateResultTotal_Rep(&_msg)
		}
	case constant.MsgTypeGameOutCard: //! 出牌消息
		var _msg public.Msg_C_DG_OutCard
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			//详细日志
			LogStr := fmt.Sprintf("OnUserOutCard UserID=%d ,CardCount=%d,牌数据:", _msg.CurrentUser, _msg.CardCount)
			for i := 0; i < public.MAX_CARD_4P; i++ {
				if _msg.CardData[i] > 0 {
					CardStr := fmt.Sprintf("0x%02x,", self.switchCard2Ox(int(_msg.CardData[i])))
					LogStr += CardStr
				}
			}
			self.OnWriteGameRecord(self.Whoplay, LogStr)

			return self.OnUserOutCard(&_msg)
		}
	case constant.MsgTypeGameTrustee: //用户托管
		var _msg public.Msg_C_Trustee
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.onUserTustee(&_msg)
			//详细日志
			LogStr := fmt.Sprintf("主动托管动作(true托管,false取消):TrustFlag=%t ", _msg.Trustee)
			self.OnWriteGameRecord(self.GetChairByUid(_msg.Id), LogStr)
		}
	case constant.MsgTypeRoar: //用户吼牌动作
		var _msg public.Msg_C_DG_Roar
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			bRoarFlag := false
			if _msg.RoarFlag == 1 {
				bRoarFlag = true
			}
			//详细日志
			LogStr := fmt.Sprintf("吼牌动作:theFlag=%d ", _msg.RoarFlag)
			self.OnWriteGameRecord(_msg.CurrentUser, LogStr)
			return self.OnRoarAction(_msg.CurrentUser, bRoarFlag)

		}
	case constant.MsgType4KingScore: // 4王换分
		var _msg Msg_C_4KingScore
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			//详细日志
			LogStr := fmt.Sprintf("4王换分:theFlag=%d ", _msg.Flag)
			self.OnWriteGameRecord(_msg.CurrentUser, LogStr)
			return self.On4KingScore(_msg.CurrentUser, _msg.Flag)
		}
	case constant.MsgTypeGameGoOnNextGame: //下一局 //暂时没有下一局
		//详细日志
		LogStr := string("OnUserClientNextGame!!! ")
		self.OnWriteGameRecord(self.Whoplay, LogStr)

		var _msg public.Msg_C_GoOnNextGame
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.OnUserClientNextGame(&_msg)
		}
	case constant.MsgTypeGameDismissFriendResult: //申请解散玩家选择
		if self.GameEndStatus == byte(public.GS_FREE) {
			var _msg public.Msg_C_DismissFriendResult
			if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
				if _msg.Flag == false {
					//不同意解散,托管玩家自动准备
					//是否托管
					for _, i := range self.TrustPlayer {
						if item := self.GetUserItemByChair(uint16(i)); item != nil {
							self.AutoNextGame(item.Uid)
						}
					}
				}
			}
		}
	default:
		//self.GameCommon.OnMsg(msg)
	}
	return true
}

// 下一局
func (self *Game_jl_jlkj) OnUserClientNextGame(msg *public.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(self.CurCompleteCount) >= self.Rule.JuShu || self.GetGameStatus() != public.GS_MJ_PLAY {
		return true
	}

	// 记录游戏开始时间
	self.GameCommon.GameBeginTime = time.Now()

	nChiarID := self.GetChairByUid(msg.Id)
	//将该消息广播出去。游戏开始后，不用广播
	if self.GameEndStatus != public.GS_MJ_PLAY {
		self.SendTableMsg(constant.MsgTypeGameGoOnNextGame, *msg)
		self.SendUserStatus(int(nChiarID), public.US_READY) //把我的状态发给其他人
	}

	//SEND_TABLE_DATA(INVALID_CHAIR,SUB_C_GOON_NEXT_GAME,pDataBuffer);

	if nChiarID >= 0 && nChiarID < uint16(self.GetPlayerCount()) {
		_item := self.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < self.GetPlayerCount(); i++ {
		item := self.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			break
		}
		if i == self.GetPlayerCount()-1 {
			self.RepositTable(false) // 复位桌子
			self.OnGameStart()
		}
	}
	return true
}

// 托管
func (self *Game_jl_jlkj) onUserTustee(msg *public.Msg_C_Trustee) bool {
	//变量定义
	var tuoguan public.Msg_S_DG_Trustee
	tuoguan.ChairID = self.GetChairByUid(msg.Id)
	tuoguan.Trustee = msg.Trustee
	//校验规则
	if tuoguan.ChairID < public.MAX_PLAYER_4P && ((self.GameState == meta.GsPlay) || (self.GameState == meta.GsRoarPai) || (self.GameState == meta.GsNull)) || (self.GameState == meta.Gs4KingScore) {
		if tuoguan.Trustee == true && (self.GameState != meta.GsNull) {
			self.TuoGuanPlayer[tuoguan.ChairID] = true
			self.TrustCounts[tuoguan.ChairID]++
			isTrust := false
			for _, val := range self.TrustPlayer {
				if val == byte(tuoguan.ChairID) {
					isTrust = true
				}
			}
			if !isTrust {
				self.TrustPlayer = append(self.TrustPlayer, byte(tuoguan.ChairID))
			}
			if tuoguan.ChairID == self.Whoplay {
				//self.DownTime = GetCPUTickCount()+self.AutoOutTime;
				if int64(self.LimitTime) > time.Now().Unix() {
					tuoguan.Overtime = self.LimitTime - time.Now().Unix()
				}
				if time.Now().Unix()+int64(self.AutoOutTime) < self.LimitTime { // 如果只剩下托管出牌的时间了，就不重新算了，否则跟改为托管出牌的时间
					self.setLimitedTime(int64(self.AutoOutTime))
				}
			}
			self.SendTableMsg(constant.MsgTypeGameTrustee, tuoguan)
			// 回放托管记录
			self.addReplayOrder(tuoguan.ChairID, meta.DG_REPLAY_OPT_TUOGUAN, 1, []int{})
		} else if tuoguan.Trustee == false {
			self.TuoGuanPlayer[tuoguan.ChairID] = false
			index := -1
			for i, val := range self.TrustPlayer {
				if val == byte(tuoguan.ChairID) {
					index = i
				}
			}
			if index > -1 {
				self.TrustPlayer = append(self.TrustPlayer[:index], self.TrustPlayer[index+1:]...)
			}
			// 如果是当前的玩家，那么重新设置一下开始时间
			if tuoguan.ChairID == self.Whoplay {
				//self.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
				if time.Now().Unix() < self.LimitTime { // 如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < self.LimitTime
					self.setLimitedTime(int64(self.PlayTime) + self.PowerStartTime - time.Now().Unix() + 1)
					tuoguan.Overtime = int64(self.PlayTime) + self.PowerStartTime - time.Now().Unix()
				}
			}

			//tuoguan.theTime = PlayTime-(now-nowTime);
			self.SendTableMsg(constant.MsgTypeGameTrustee, tuoguan)
			//回放增加托管记录
			self.addReplayOrder(tuoguan.ChairID, meta.DG_REPLAY_OPT_TUOGUAN, 0, []int{})
		} else {
			return false
		}
	} else {
		//详细日志
		LogStr := string("托管动作:游戏状态不正确 ")
		self.OnWriteGameRecord(tuoguan.ChairID, LogStr)
		return false
	}
	return true
}

func (self *Game_jl_jlkj) addReplayOrder(chairId uint16, operation int, value int, values []int) {
	var order meta.DG_Replay_Order
	order.R_ChairId = chairId
	order.R_Opt = operation

	if operation == meta.DG_REPLAY_OPT_HOUPAI {
		var order_ext meta.DG_Replay_Order_Ext
		order_ext.Ext_type = meta.DG_EXT_HOUPAI
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == meta.DG_REPLAY_OPT_END_HOUPAI {
		order.R_Value = append(order.R_Value, value)
	} else if operation == meta.DG_REPLAY_OPT_OUTCARD {
		order.R_Value = append(order.R_Value, values[:]...)
	} else if operation == meta.DG_REPLAY_OPT_END_GAME {

	} else if operation == meta.DG_REPLAY_OPT_DIS_GAME {

	} else if operation == meta.DG_REPLAY_OPT_TURN_OVER {
		var order_ext meta.DG_Replay_Order_Ext
		order_ext.Ext_type = meta.DG_EXT_GETSCORE
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == meta.DG_REPLAY_OPT_TUOGUAN {
		var order_ext meta.DG_Replay_Order_Ext
		order_ext.Ext_type = meta.DG_EXT_TUOGUAN
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == meta.DG_REPLAY_OPT_4KINGSCORE {
		order.R_Value = append(order.R_Value, values[:]...)

		var order_ext meta.DG_Replay_Order_Ext
		order_ext.Ext_type = meta.DG_EXT_4KINGSCORE
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	}

	self.ReplayRecord.R_Orders = append(self.ReplayRecord.R_Orders, order)
}

// 用户出牌
func (self *Game_jl_jlkj) OnUserOutCard(msg *public.Msg_C_DG_OutCard) bool {
	syslog.Logger().Info("OnUserOutCard")
	//效验状态
	if self.GetGameStatus() != public.GS_MJ_PLAY {
		return true
	}
	if self.GameEndStatus != public.GS_MJ_PLAY {
		return true
	}
	if self.GameState != meta.GsPlay {
		return false
	}
	wChairID := msg.CurrentUser
	//效验参数"tablecreate"
	if wChairID != self.Whoplay {
		//详细日志
		LogStr := fmt.Sprintf("座位号 %d, OnUserOutCard 出牌玩家不是当前玩家 ", wChairID)
		self.OnWriteGameRecord(wChairID, LogStr)
		return false
	}
	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	//出牌数目为0，则为放弃的情况
	if msg.CardCount == 0 {
		if self.WhoLastOut < public.MAX_PLAYER_4P && self.Whoplay == wChairID {
			var outmsg public.Msg_S_DG_OutCard
			outmsg.CardCount = 0
			outmsg.CurrentUser = msg.CurrentUser
			for i := 0; i < public.MAX_CARD_4P; i++ {
				outmsg.CardData[i] = 0
				self.AllPaiOut[self.Whoplay][i] = 0
				self.LastPaiOut[self.Whoplay][i] = 0
			}
			outmsg.OutScorePai = self.OutScorePai
			self.SendTableMsg(constant.MsgTypeGameOutCard, outmsg)

			//回放增加出牌日志
			self.addReplayOrder(self.Whoplay, meta.DG_REPLAY_OPT_OUTCARD, int(public.INVALID_BYTE), []int{})

			//详细日志
			LogStr := string("OnUserOutCard 玩家pass ")
			self.OnWriteGameRecord(self.Whoplay, LogStr)

			self.WhoPass[self.Whoplay] = true

			//通过游数是否确定 判断是否应该结束游戏
			bAllPlayerTurnIsConfirm := true
			nPlayerTurn1st := public.INVALID_CHAIR
			nPlayerTurn2nd := public.INVALID_CHAIR
			nPlayerTurn3rd := public.INVALID_CHAIR
			nPlayerTurn4th := public.INVALID_CHAIR
			for i := 0; i < public.MAX_PLAYER_4P; i++ {
				if self.PlayerTurn[i] == public.INVALID_BYTE {
					bAllPlayerTurnIsConfirm = false
				} else if self.PlayerTurn[i] == 1 {
					nPlayerTurn1st = uint16(i)
				} else if self.PlayerTurn[i] == 2 {
					nPlayerTurn2nd = uint16(i)
				} else if self.PlayerTurn[i] == 3 {
					nPlayerTurn3rd = uint16(i)
				} else if self.PlayerTurn[i] == 4 {
					nPlayerTurn4th = uint16(i)
				}
			}
			//本轮所有人都已结束出牌
			if bAllPlayerTurnIsConfirm && self.Whoplay == nPlayerTurn4th {
				teamer := self.GetTeamer(nPlayerTurn1st)
				if teamer == nPlayerTurn2nd {
					if self.WhoLastOut == nPlayerTurn3rd {
						// 3游圧了2游的最后一手牌
						self.PlayerCardScore[nPlayerTurn3rd] += self.CardScore
						self.SendPlayerScore(public.MAX_PLAYER_4P, nPlayerTurn3rd, self.CardScore)
						self.OnGameEndNormal(nPlayerTurn2nd, meta.GOT_DOUBLEKILL)
					} else {
						// 双杀
						self.PlayerCardScore[nPlayerTurn2nd] += self.CardScore
						self.SendPlayerScore(public.MAX_PLAYER_4P, nPlayerTurn2nd, self.CardScore)
						self.OnGameEndNormal(nPlayerTurn2nd, meta.GOT_DOUBLEKILL)
					}
				} else {
					// 4游不要3游的最后一手牌
					self.PlayerCardScore[nPlayerTurn3rd] += self.CardScore
					self.SendPlayerScore(public.MAX_PLAYER_4P, nPlayerTurn3rd, self.CardScore)
					self.OnGameEndNormal(nPlayerTurn3rd, meta.GOT_NORMAL)
				}
			} else {
				self.GoNextPlayer()
			}
		}

		return true
	}

	buf := [public.MAX_CARD_4P]byte{}
	for i := 0; i < public.MAX_CARD_4P; i++ {
		buf[i] = self.PlayerCards[self.Whoplay][i]
	}

	z := 0 //重新检验牌数目，并临时删除出的牌
	for i := byte(0); i < msg.CardCount; i++ {
		for j := 0; j < public.MAX_CARD_4P; j++ {
			if msg.CardData[i] == buf[j] {
				buf[j] = 0
				z++
				break
			}
		}
	}
	if byte(z) == msg.CardCount {
		var re1, re2 public.TCardType
		iNumOfKing := 0
		for i := 0; i < public.MAX_CARD_4P; i++ {
			self.AllPaiOut[self.Whoplay][i] = msg.CardData[i]
		}

		//TYPE_NULL代表第一个出
		if self.WhoLastOut >= public.MAX_PLAYER_4P {
			re1.Cardtype = public.TYPE_NULL
			self.LastOutType = public.TYPE_NULL
			self.LastOutTypeClient = public.TYPE_NULL
		} else {
			len := self.m_GameLogic.GetCardNum(self.AllPaiOut[self.WhoLastOut], public.MAX_CARD_4P)
			re1 = self.m_GameLogic.GetType(self.AllPaiOut[self.WhoLastOut], int(len), 0, 0, 0)
			//这里为什么要重新设置下呢？因为很可能出现不同的判断
			//比如A走了334455,B走了44王王王5,那么当C走的时候，GetType很可能把B的牌型判断为444555这样的大牌型
			//所以，在这里，一定要把牌型设置回来，根据客户端传过来的为依据
			re1.Cardtype = self.LastOutType
		}
		re2 = self.m_GameLogic.GetType(self.AllPaiOut[self.Whoplay], int(msg.CardCount), 0, 0, self.LastOutType)
		iNumOfKing = self.m_GameLogic.GetKingNum(self.AllPaiOut[self.Whoplay], int(msg.CardCount))

		//详细日志/////////
		//		LogStr := fmt.Sprintf("re1的类型%d ,card=%d ", re1.Cardtype, re1.Card)
		//		LogStr += fmt.Sprintf("re2的类型%d ,card=%d ", re2.Cardtype, re2.Card)
		//		self.OnWriteGameRecord(self.Whoplay, LogStr)

		self.WhoPass[self.Whoplay] = false
		if self.m_GameLogic.Compare(re1, re2) {
			//回放增加出牌日志
			self.addReplayOrder(self.Whoplay, meta.DG_REPLAY_OPT_OUTCARD, int(public.INVALID_BYTE), public.HF_BytesToInts(msg.CardData[:]))
			// 详细日志 by sam  打出510K或者7喜或者8喜,管得起才算
			if re2.Cardtype == public.TYPE_BOMB_510K {
				self.Play510K[self.Whoplay]++
			} else if re2.Cardtype == public.TYPE_BOMB_7XI {
				self.Play7Xi[self.Whoplay]++
			} else if re2.Cardtype == public.TYPE_BOMB_8XI {
				self.Play8Xi[self.Whoplay]++
			} else if re2.Cardtype == public.TYPE_BOMB_FOUR_KING {

				self.PlayKingBomb[self.Whoplay]++
				//详细日志
				LogStr := fmt.Sprintf("打出了第%d个天炸(4王) ", self.PlayKingBomb[self.Whoplay])
				self.OnWriteGameRecord(wChairID, LogStr)
			}

			//管得起才能算分
			if self.GameType == meta.GT_NORMAL {
				self.CardScore += self.m_GameLogic.GetScore(self.AllPaiOut[self.Whoplay], int(msg.CardCount))
				self.SendTurnScore(public.MAX_PLAYER_4P)
				self.ReplayRecord.R_Orders[len(self.ReplayRecord.R_Orders)-1].AddReplayExtData(meta.DG_EXT_TURNSCORE, self.CardScore)
			} else {
				self.CardScore = 0
			}

			//回放分牌
			recordScorePai := []int{}

			//加入分牌
			index := 24
			for i := 0; i < 24; i++ {
				if self.OutScorePai[i] == 0 {
					index = i
					break
				}
			}
			for i := byte(0); i < msg.CardCount; i++ {
				outcard := self.AllPaiOut[self.Whoplay][i]
				if self.m_GameLogic.isScorePai(outcard) {
					if index >= 24 {
						self.OnWriteGameRecord(wChairID, "分数牌最多24张")
						break
					}
					recordScorePai = append(recordScorePai, int(outcard))
					self.OutScorePai[index] = outcard
					index++
				}
			}

			if len(recordScorePai) > 0 {
				self.ReplayRecord.R_Orders[len(self.ReplayRecord.R_Orders)-1].R_ScoreCard = recordScorePai
			}

			self.LastOutType = re2.Cardtype
			self.LastOutTypeClient = msg.CardType

			self.ThePaiCount[self.Whoplay] -= msg.CardCount
			self.WhoLastOut = self.Whoplay

			z := 0 //计算剩余多少张牌 m_thePaiCount可能不准
			for i := 0; i < public.MAX_CARD_4P; i++ {
				self.PlayerCards[self.Whoplay][i] = buf[i]
				if buf[i] > 0 {
					z++
				}
			}

			// 发送手牌数量
			if self.ShowHandCardCnt {
				self.SendPaiCount(public.MAX_PLAYER_4P)
			}

			var msgout public.Msg_S_DG_OutCard
			msgout.CardCount = msg.CardCount
			msgout.CardType = msg.CardType
			msgout.CurrentUser = msg.CurrentUser
			msgout.Overtime = 15
			msgout.ByClient = false
			msgout.CardData = msg.CardData
			msgout.OutScorePai = self.OutScorePai
			self.SendTableMsg(constant.MsgTypeGameOutCard, msgout)

			//插底提示
			if z == 1 {
				self.SendPlaySoundMsg(self.Whoplay, public.TY_ChaDi)
			}

			//判断任务是否完成
			self.GameTask.IsTaskFinished(self.GameCommon, re2, self.Whoplay, iNumOfKing)
			//end

			//通过游数是否确定 判断是否应该结束游戏 4游可以要3游最后一手牌
			bAllPlayerTurnIsConfirm := true
			nPlayerTurn1st := public.INVALID_CHAIR
			nPlayerTurn2nd := public.INVALID_CHAIR
			nPlayerTurn3rd := public.INVALID_CHAIR
			nPlayerTurn4th := public.INVALID_CHAIR
			for i := 0; i < public.MAX_PLAYER_4P; i++ {
				if self.PlayerTurn[i] == public.INVALID_BYTE {
					bAllPlayerTurnIsConfirm = false
				} else if self.PlayerTurn[i] == 1 {
					nPlayerTurn1st = uint16(i)
				} else if self.PlayerTurn[i] == 2 {
					nPlayerTurn2nd = uint16(i)
				} else if self.PlayerTurn[i] == 3 {
					nPlayerTurn3rd = uint16(i)
				} else if self.PlayerTurn[i] == 4 {
					nPlayerTurn4th = uint16(i)
				}
			}
			//本轮所有人都已结束出牌
			if bAllPlayerTurnIsConfirm && self.Whoplay == nPlayerTurn4th {
				teamer := self.GetTeamer(nPlayerTurn1st)
				if teamer == nPlayerTurn2nd {
					self.PlayerCardScore[self.Whoplay] += self.CardScore
					self.SendPlayerScore(public.MAX_PLAYER_4P, self.Whoplay, self.CardScore)
					self.OnGameEndNormal(nPlayerTurn2nd, meta.GOT_DOUBLEKILL)
				} else {
					self.PlayerCardScore[self.Whoplay] += self.CardScore
					self.SendPlayerScore(public.MAX_PLAYER_4P, self.Whoplay, self.CardScore)
					self.OnGameEndNormal(nPlayerTurn3rd, meta.GOT_NORMAL)
				}
			} else {
				self.GameOverOrNextPlayer(z, re2, iNumOfKing)
			}

		} else //压不住或牌型不正确
		{
			self.WhoPass[self.Whoplay] = true

			//通过游数是否确定 判断是否应该结束游戏
			bAllPlayerTurnIsConfirm := true
			nPlayerTurn1st := public.INVALID_CHAIR
			nPlayerTurn2nd := public.INVALID_CHAIR
			nPlayerTurn3rd := public.INVALID_CHAIR
			nPlayerTurn4th := public.INVALID_CHAIR
			for i := 0; i < public.MAX_PLAYER_4P; i++ {
				if self.PlayerTurn[i] == public.INVALID_BYTE {
					bAllPlayerTurnIsConfirm = false
				} else if self.PlayerTurn[i] == 1 {
					nPlayerTurn1st = uint16(i)
				} else if self.PlayerTurn[i] == 2 {
					nPlayerTurn2nd = uint16(i)
				} else if self.PlayerTurn[i] == 3 {
					nPlayerTurn3rd = uint16(i)
				} else if self.PlayerTurn[i] == 4 {
					nPlayerTurn4th = uint16(i)
				}
			}
			//本轮所有人都已结束出牌
			if bAllPlayerTurnIsConfirm && self.Whoplay == nPlayerTurn4th {
				teamer := self.GetTeamer(nPlayerTurn1st)
				if teamer == nPlayerTurn2nd {
					if self.WhoLastOut == nPlayerTurn3rd {
						// 3游圧了2游的最后一手牌
						self.PlayerCardScore[nPlayerTurn3rd] += self.CardScore
						self.SendPlayerScore(public.MAX_PLAYER_4P, nPlayerTurn3rd, self.CardScore)
						self.OnGameEndNormal(nPlayerTurn2nd, meta.GOT_DOUBLEKILL)
					} else {
						// 双杀
						self.PlayerCardScore[nPlayerTurn2nd] += self.CardScore
						self.SendPlayerScore(public.MAX_PLAYER_4P, nPlayerTurn2nd, self.CardScore)
						self.OnGameEndNormal(nPlayerTurn2nd, meta.GOT_DOUBLEKILL)
					}
				} else {
					// 4游不要3游的最后一手牌
					self.PlayerCardScore[nPlayerTurn3rd] += self.CardScore
					self.SendPlayerScore(public.MAX_PLAYER_4P, nPlayerTurn3rd, self.CardScore)
					self.OnGameEndNormal(nPlayerTurn3rd, meta.GOT_NORMAL)
				}
			}

			return false
		}
	}

	return true
}

// 游戏是否可以结束或者下一个玩家出牌
func (self *Game_jl_jlkj) GameOverOrNextPlayer(byLeftCardNum int, re public.TCardType, iNumOfKing int) int {
	//硬牌，有一个没有牌了就结束游戏
	if self.GameType == meta.GT_ROAR {
		//结束游戏
		if byLeftCardNum == 0 {
			//判断最后一手的任务是否完成
			self.GameTask.IsTaskFinishedOfLastHand(self.GameCommon, re, self.Whoplay, iNumOfKing)
			//end

			self.Nextbanker = self.Whoplay
			self.OnGameEndNormal(self.Whoplay, meta.GOT_NORMAL)
			return 1
		} else {
			self.GoNextPlayer()
		}
	} else if self.GameType == meta.GT_NORMAL {
		//不硬牌，找朋友的模式
		//m_whoplay的牌没有了，那么检查下我的对家结束没有
		if byLeftCardNum == 0 {
			//判断最后一手的任务是否完成
			self.GameTask.IsTaskFinishedOfLastHand(self.GameCommon, re, self.Whoplay, iNumOfKing)
			//end

			teamer := self.GetTeamer(self.Whoplay)
			self.BTeamOut[teamer] = true

			self.WhoAllOutted[self.Whoplay] = true
			self.AllOutCnt++
			self.PlayerTurn[self.Whoplay] = self.AllOutCnt

			self.SendPlayerTurn(public.MAX_PLAYER_4P)
			self.SendPlaySoundMsg(self.Whoplay, public.TY_AllOut) //牌出完了需要客户端加特效

			if teamer < public.MAX_PLAYER_4P {
				//我的对家走完，那么结束游戏
				if self.WhoAllOutted[teamer] {
					//接着检测是不是双扣
					self.WhoAllOutted[self.Whoplay] = true
					ncount := 0
					for i := 0; i < public.MAX_PLAYER_4P; i++ {
						if !self.WhoAllOutted[i] {
							ncount++
						}
					}
					//双扣
					if ncount == 2 {
						for i := 0; i < public.MAX_PLAYER_4P; i++ {
							if self.PlayerTurn[i] == 1 {
								self.Nextbanker = uint16(i)
								break
							}
						}

						//给34游赋值 从2游的下一位玩家找三游
						iStart := self.Whoplay
						if iStart >= public.MAX_PLAYER_4P-1 {
							iStart = 0
						}
						for i := iStart; i < public.MAX_PLAYER_4P; i++ {
							if self.PlayerTurn[i] == public.INVALID_BYTE {
								self.PlayerTurn[i] = 3
								break
							}
						}
						for j := 0; j < public.MAX_PLAYER_4P; j++ {
							if self.PlayerTurn[j] == public.INVALID_BYTE {
								self.PlayerTurn[j] = 4
								break
							}
						}

						// 是否可以捡尾分
						if self.GetLastScore {
							//4游可以要3游最后一手牌
							self.GoNextPlayer()
						} else {
							// 双杀 游戏结束
							self.PlayerCardScore[self.Whoplay] += self.CardScore
							self.SendPlayerScore(public.MAX_PLAYER_4P, self.Whoplay, self.CardScore)
							self.OnGameEndNormal(self.Whoplay, meta.GOT_DOUBLEKILL)
						}

						return 1
					} else if ncount == 1 {
						//3家结束
						for i := 0; i < public.MAX_PLAYER_4P; i++ {
							if self.PlayerTurn[i] == 1 {
								self.Nextbanker = uint16(i)
								break
							}
						}
						//给4游赋值
						for i := 0; i < public.MAX_PLAYER_4P; i++ {
							if !self.WhoAllOutted[i] {
								self.PlayerTurn[i] = 4
							}
						}

						// 是否可以捡尾分
						if self.GetLastScore {
							//4游可以要3游最后一手牌
							self.GoNextPlayer()
						} else {
							// 游戏结束
							self.PlayerCardScore[self.Whoplay] += self.CardScore
							self.SendPlayerScore(public.MAX_PLAYER_4P, self.Whoplay, self.CardScore)
							self.OnGameEndNormal(self.Whoplay, meta.GOT_NORMAL)
						}

						return 1
					} else { //可能吗？如果是其他情况，那么逻辑就错了！！

					}
					////这里检查下，是否结束游戏
					//if self.JudgeEndGame(self.Whoplay) {
					//	return 1
					//}
				} else //接风
				{
					////这里检查下，是否结束游戏
					//if self.JudgeEndGame(self.Whoplay) {
					//	return 1
					//}
					//SendPlaySoundMsg(Whoplay,msgPlaySound::TY_JieFeng);
					if self.SeeTeamerCard {
						self.SendPaiToTeamer(teamer, self.Whoplay)
					}

					//如果勾选了可看队友手牌，自己打完最后一手牌，也要把手牌数据发给队友，用于刷新客户端显示
					if self.SeePartnerCards {
						self.SendPaiToTeamer(self.Whoplay, teamer)
					}

					self.GoNextPlayer()
				}
			}

		} else {
			if (self.BTeamOut[self.Whoplay] && self.SeeTeamerCard) || self.SeePartnerCards {
				teamer := self.GetTeamer(self.Whoplay)
				self.SendPaiToTeamer(self.Whoplay, teamer)
			}
			self.GoNextPlayer()
		}
	}
	return 1
}

func (self *Game_jl_jlkj) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	if wChairID >= public.MAX_PLAYER_4P {
		return false
	}
	if cbReason != meta.GOT_NORMAL && cbReason != meta.GOT_DOUBLEKILL && cbReason != meta.GOT_ZHONGTU {
		return false
	}
	if self.GameType == meta.GT_NORMAL {
		self.CardScore = 0
		self.SendTurnScore(public.MAX_PLAYER_4P)
	}
	self.GameEndStatus = public.GS_MJ_END

	//定义变量
	var endgameMsg public.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = public.GER_NORMAL
	endgameMsg.TheOrder = self.CurCompleteCount
	endgameMsg.WhoKingBomb = public.INVALID_CHAIR //4王一起出了才算
	if self.WhoHasKingBomb >= 0 && self.WhoHasKingBomb < public.MAX_PLAYER_4P {
		if self.PlayKingBomb[self.WhoHasKingBomb] > 0 {
			endgameMsg.WhoKingBomb = self.WhoHasKingBomb
		}
	}

	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.Have7Xi[i] = self.Who7Xi[i]
		endgameMsg.Have8Xi[i] = self.Who8Xi[i]
		//endgameMsg.Have510K[i] = self.WhoSame510K[i]
	}

	fullScoreAward := 0                    //满分奖励
	fourKingScore := 0                     //4王换的分
	addDiFen := 0                          //保底分数
	gongfen := 0                           //贡分
	payfan := 0                            //算番
	isBankerWin := false                   //是否庄家队伍赢
	isScoreEqual := false                  //是否分数相等
	bankeryou := self.GetTeamYouType(true) //庄家队伍几游
	// 庄家、闲家队伍抓分统计(含4王换分)
	bankscore, xianscore := self.GetJLKJParterScore(true)
	// 庄家、闲家队伍抓分统计(不含4王换分)
	bankscore_noking, xianscore_noking := self.GetJLKJParterScore(false)
	// 4王换的分
	if self.FourKingScore > 0 && self.WhoHas4KingScore != public.INVALID_CHAIR {
		fourKingScore = self.FourKingScore
	}
	// 保底分数 跟级数有关
	if self.AddDiFen > 0 {
		addDiFen = self.IBase * self.AddDiFen
	}
	syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)共抓分%d,闲家抓分%d,4王换分%d,额外加分%d", self.Banker, self.BankParter, bankscore, xianscore, fourKingScore, addDiFen))
	if bankeryou == YOUTYPE_12 || bankeryou == YOUTYPE_34 {
		//12对34情况,34游需要给12游贡60分
		if self.Rule.ShareScoreType == GongScoreType_3060 || self.Rule.ShareScoreType == GongScoreType_4060 {
			gongfen = 60
			syslog.Logger().Info(fmt.Sprintf("一二游对三四游,需要贡分%d", gongfen))
		}

		//34游手上没跑掉的分需要给12游
		otherscore := 0
		otherscore += self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(3)], public.MAX_CARD_4P)
		otherscore += self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(4)], public.MAX_CARD_4P)
		self.PlayerCardScore[self.GetYouSeat(1)] += self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(3)], public.MAX_CARD_4P)
		self.PlayerCardScore[self.GetYouSeat(1)] += self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(4)], public.MAX_CARD_4P)
		//算贡分归属
		if bankeryou == YOUTYPE_12 {
			bankscore += otherscore
			bankscore += gongfen
			xianscore -= gongfen
			syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)一二游,闲家三四游,闲家队伍手上未出分%d归庄家队伍,闲家贡分%d给庄家队伍", self.Banker, self.BankParter, otherscore, gongfen))
		} else {
			xianscore += otherscore
			xianscore += gongfen
			bankscore -= gongfen
			syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)三四游,闲家一二游,庄家队伍手上未出分%d归闲家队伍,庄家队伍贡分%d给闲家队伍", self.Banker, self.BankParter, otherscore, gongfen))
		}
		// 满分奖励 跟级数有关
		if self.FullScoreAward > 0 && (bankscore_noking+otherscore >= 200 || xianscore_noking+otherscore >= 200) {
			fullScoreAward = self.IBase * self.FullScoreAward
			syslog.Logger().Info(fmt.Sprintf("满分奖励：%d", fullScoreAward))
		}
		//检验双方是否分数相等
		if bankscore == xianscore {
			isScoreEqual = true
		}
		//算番
		if isScoreEqual {
			//分数一样的话,跑一游一方赢1倍底分
			payfan = 1
			if bankeryou == YOUTYPE_12 {
				isBankerWin = true
			}
		} else {
			payfan, isBankerWin = self.SuanFen(bankscore, xianscore)
		}
		syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)积分%d,闲家积分%d", self.Banker, self.BankParter, bankscore, xianscore))
	} else if bankeryou == YOUTYPE_13 || bankeryou == YOUTYPE_24 {
		//13对24情况
		if self.Rule.ShareScoreType == GongScoreType_3060 {
			gongfen = 30
			syslog.Logger().Info(fmt.Sprintf("贡分选的30-60,一三游对二四游,需要贡分%d", gongfen))
		} else if self.Rule.ShareScoreType == GongScoreType_4060 {
			gongfen = 40
			syslog.Logger().Info(fmt.Sprintf("贡分选的40-60,一三游对二四游,需要贡分%d", gongfen))
		}

		//4游手上没跑掉的分需要给3游
		otherscore := self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(4)], public.MAX_CARD_4P)
		self.PlayerCardScore[self.GetYouSeat(3)] += self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(4)], public.MAX_CARD_4P)
		if bankeryou == YOUTYPE_13 {
			//庄家队伍13游,24游需要贡分给12游
			bankscore += otherscore
			bankscore += gongfen
			xianscore -= gongfen
			syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)一三游,闲家二四游,闲家队伍四游玩家手上未出分%d归庄家队伍,闲家队伍需要贡分%d给庄家队伍", self.Banker, self.BankParter, otherscore, gongfen))
		} else {
			xianscore += otherscore
			xianscore += gongfen
			bankscore -= gongfen
			syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)二四游,闲家一三游,庄家队伍四游玩家手上未出分%d归闲家队伍,庄家队伍需要贡分%d给闲家队伍", self.Banker, self.BankParter, otherscore, gongfen))
		}
		//检验双方是否分数相等
		if bankscore == xianscore {
			isScoreEqual = true
		}
		//算番
		if isScoreEqual {
			//分数一样的话,跑一游一方赢1倍底分
			payfan = 1
			if bankeryou == YOUTYPE_13 {
				isBankerWin = true
			}
		} else {
			payfan, isBankerWin = self.SuanFen(bankscore, xianscore)
		}

		syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)积分%d,闲家积分%d", self.Banker, self.BankParter, bankscore, xianscore))
	} else if bankeryou == YOUTYPE_14 || bankeryou == YOUTYPE_23 {
		//14对23情况
		//4游手上没跑掉的分需要给3游
		otherscore := self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(4)], public.MAX_CARD_4P)
		self.PlayerCardScore[self.GetYouSeat(3)] += self.m_GameLogic.GetScore(self.PlayerCards[self.GetYouSeat(4)], public.MAX_CARD_4P)
		if bankeryou == YOUTYPE_23 {
			//庄家队伍23游,4游手上分要给3游
			bankscore += otherscore
			syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)二三游,闲家一四游,闲家队伍手上未出分%d归庄家队伍", self.Banker, self.BankParter, otherscore))
		} else {
			xianscore += otherscore
			syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)一四游,闲家二三游,庄家队伍四游玩家手上未出分%d归闲家队伍", self.Banker, self.BankParter, otherscore))
		}
		//检验双方是否分数相等
		if bankscore == xianscore {
			isScoreEqual = true
		}
		//算番
		if isScoreEqual {
			//分数一样的话,跑一游一方赢1倍底分
			payfan = 1
			if bankeryou == YOUTYPE_14 {
				isBankerWin = true
			}
		} else {
			payfan, isBankerWin = self.SuanFen(bankscore, xianscore)
		}
		syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)积分%d,闲家积分%d", self.Banker, self.BankParter, bankscore, xianscore))
	}
	//赢番除以10,目前写死
	if !isScoreEqual && !(payfan == 1) {
		payfan /= 10
	}

	//算分
	if isBankerWin {
		for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
			if self.IsBankerTeam(i) {
				endgameMsg.WinLose[i] = 1
				endgameMsg.Score[i] += payfan*self.IBase + fullScoreAward + addDiFen
				syslog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)一四游,赢得比赛,赢分公式: [(庄家队伍积分%d) - (闲家队伍积分%d)] x 底分(%d) + 满分奖励(%d) + 保底得分(%d) = %d分",
					self.Banker, self.BankParter, bankscore, xianscore, self.IBase, fullScoreAward, addDiFen, endgameMsg.Score[i]))
			} else {
				endgameMsg.WinLose[i] = 0
				endgameMsg.Score[i] -= payfan*self.IBase + fullScoreAward + addDiFen
				syslog.Logger().Info(fmt.Sprintf("闲家队伍(%d)二三游,输了比赛,输分公式: [(闲家队伍积分%d) - (庄家队伍积分%d)] x 底分(%d) + 满分奖励(%d) + 保底得分(%d) = %d分",
					i, xianscore, bankscore, self.IBase, fullScoreAward, addDiFen, endgameMsg.Score[i]))
			}
		}
	} else {
		for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
			if !self.IsBankerTeam(i) {
				endgameMsg.WinLose[i] = 1
				endgameMsg.Score[i] += payfan*self.IBase + fullScoreAward + addDiFen
				syslog.Logger().Info(fmt.Sprintf("闲家队伍(%d)一四游,赢得比赛,赢分公式: [(闲家队伍积分%d) - (庄家队伍积分%d)] x 底分(%d) + 满分奖励(%d) + 保底得分(%d) = %d分",
					i, xianscore, bankscore, self.IBase, fullScoreAward, addDiFen, endgameMsg.Score[i]))
			} else {
				endgameMsg.WinLose[i] = 0
				endgameMsg.Score[i] -= payfan*self.IBase + fullScoreAward + addDiFen
				syslog.Logger().Info(fmt.Sprintf("庄家队伍(%d)二三游,输了比赛,输分公式: [(庄家队伍积分%d) - (闲家队伍积分%d)] x 底分(%d) + 满分奖励(%d) + 保底得分(%d) = %d分",
					i, bankscore, xianscore, self.IBase, fullScoreAward, addDiFen, endgameMsg.Score[i]))
			}
		}
	}

	//贡献分
	endgameMsg.GongScore = gongfen

	//胜局次数统计、计算保底得分
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if endgameMsg.WinLose[i] == 1 && _userItem != nil {
			//self.WinCount[i]++
			// 记录在玩家身上
			_userItem.Ctx.WinCount++
		}
	}

	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		cn := 0
		for j := 0; j < public.MAX_CARD_4P; j++ {
			if self.m_GameLogic.IsValidCard(self.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = self.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = self.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)
		//self.LastScore[i] = endgameMsg.Score[i]
		//self.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = self.GetUserTotalScore(i) + endgameMsg.Score[i]

		if self.MaxScore[i] < self.PlayerCardScore[i] {
			self.MaxScore[i] = self.PlayerCardScore[i]
		}
		//一游次数
		if self.PlayerTurn[i] == 1 {
			//self.TotalFirstTurn[i]++
			// 记录在玩家身上
			_userItem := self.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				_userItem.Ctx.PlayTurn1st++
			}
		}

		endgameMsg.TheTurn[i] = self.PlayerTurn[i]
		endgameMsg.GetScore[i] = self.PlayerCardScore[i]
	}
	if self.WhoRoar < meta.MAX_PLAYER {
		self.TotalDuPai[self.WhoRoar]++
	}
	endgameMsg.HouPaiChair = self.WhoRoar
	endgameMsg.TheBank = self.Banker
	endgameMsg.TheParter = self.BankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", endgameMsg.Score[i], endgameMsg.TotalScore[i], self.PlayerCardScore[i])
		self.OnWriteGameRecord(i, recrodStr)
	}

	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.XiScore[i] = self.XiScore[i] //喜钱
	}
	// 调用结算接口
	_, endgameMsg.UserVitamin = self.OnSettle(endgameMsg.Score, constant.EventSettleGameOver)

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, endgameMsg) //保存，用于汇总计算
	self.SaveGameData()
	self.SendTableMsg(constant.MsgTypeGameEnd, endgameMsg)

	self.OnWriteGameRecord(public.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	//回放增加结束数据
	self.addReplayOrder(wChairID, meta.DG_REPLAY_OPT_END_GAME, 0, []int{})

	// 数据库写出牌记录	// 写完后清除数据
	self.TableWriteOutDate()

	for _, v := range self.PlayerInfo { //i := 0; i < self.GetPlayerCount(); i++ {
		wintype := public.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinLose[v.Seat] == 1 {
			wintype = public.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = public.ScoreKind_Lost //enScoreKind_Lost;
		}
		//iAward := self.GetTaskAward(v.Seat)//金豆任务，先留着备用
		// 记录分数用于总结算汇总
		v.Ctx.GameScoreFen = append(v.Ctx.GameScoreFen, endgameMsg.Score[v.Seat])
		self.TableWriteGameDate(int(self.CurCompleteCount), v, wintype, endgameMsg.Score[v.Seat])
	}

	//扣房卡
	if self.CurCompleteCount == 1 {
		self.TableDeleteFangKa(self.CurCompleteCount)
	}

	// 检查托管标识
	checkTuoGuanFlag := false

	//结束游戏
	if int(self.CurCompleteCount) >= self.Rule.JuShu { //局数够了
		self.CalculateResultTotal(public.GER_NORMAL, wChairID, 0) //计算总发送总结算

		self.UpdateOtherFriendDate(&endgameMsg, false)
		//通知框架结束游戏
		//self.SetGameStatus(public.GS_MJ_FREE)
		self.ConcludeGame()

	} else {
		if self.Rule.Overtime_dismiss != -1 {
			for _, val := range self.TrustPlayer {
				if item := self.GetUserItemByChair(uint16(val)); item != nil {
					if checkTuoGuanFlag {
						var _msg = &public.Msg_C_DismissFriendResult{
							Id:   item.Uid,
							Flag: true,
						}
						self.OnDismissResult(item.Uid, _msg)
					} else {
						checkTuoGuanFlag = true
						var msg = &public.Msg_C_DismissFriendReq{
							Id: item.Uid,
						}
						self.SetDismissRoomTime(self.Rule.Overtime_dismiss)
						self.OnDismissFriendMsg(item.Uid, msg)
					}
				}
			}
		}
	}

	// 1、第一局随机坐庄；2、胡牌或流局连庄；3、第一局无人胡牌则下家坐庄；
	if self.BankerUser != public.INVALID_CHAIR {
		self.BankerUser = self.Nextbanker
		self.Banker = self.Nextbanker
	}

	self.OnGameEnd()
	self.RepositTable(false) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	// OnGameEnd 会清除计时器
	if int(self.CurCompleteCount) < self.Rule.JuShu && !checkTuoGuanFlag && self.Rule.Overtime_trust > 0 {
		self.SetAutoNextTimer(15) // 自动开始下一局
	}

	return true
}

func (self *Game_jl_jlkj) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	if wChairID >= public.MAX_PLAYER_4P {
		return false
	}
	if cbReason != meta.GOT_ESCAPE && cbReason != meta.GOT_TUOGUAN {
		return false
	}
	if self.GameType == meta.GT_NORMAL {
		self.CardScore = 0
		self.SendTurnScore(public.MAX_PLAYER_4P)
	}
	self.GameEndStatus = public.GS_MJ_END

	//定义变量
	iScore := [meta.MAX_PLAYER]int{}
	self.XiScore = [meta.MAX_PLAYER]int{}

	byGongType := byte(0)
	byGongType = public.G_Bangong

	var endgameMsg public.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = public.GER_USER_LEFT
	endgameMsg.TheOrder = self.CurCompleteCount
	endgameMsg.WhoKingBomb = public.INVALID_CHAIR //4王一起出了才算
	if self.WhoHasKingBomb >= 0 && self.WhoHasKingBomb < public.MAX_PLAYER_4P {
		if self.PlayKingBomb[self.WhoHasKingBomb] > 0 {
			endgameMsg.WhoKingBomb = self.WhoHasKingBomb
		}
	}

	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.Have7Xi[i] = self.Who7Xi[i]
		endgameMsg.Have8Xi[i] = self.Who8Xi[i]
		endgameMsg.Have510K[i] = self.WhoSame510K[i]
	}

	endgameMsg.EndType = public.TY_ESCAPE

	nFan := 0
	if self.GameState == meta.GsRoarPai {
		nFan = self.FaOfTao * 2
	} else {
		nFan = self.FaOfTao
	}
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		if i == wChairID {
			iScore[i] = -self.IBase * nFan
			endgameMsg.WinLose[i] = 0
		} else {
			if self.GameState == meta.GsRoarPai {
				iScore[i] = self.IBase * (self.JiangOfTao) * 2
			} else {
				iScore[i] = self.IBase * (self.JiangOfTao)
			}
			endgameMsg.WinLose[i] = 1
		}
	}
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		self.AddSpecailScore(&iScore, i, self.IBase)
	}

	endgameMsg.FanShu = 1
	endgameMsg.GongType = byGongType
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.Score[i] = iScore[i] - self.Spay
		cn := 0
		for j := 0; j < public.MAX_CARD_4P; j++ {
			if self.m_GameLogic.IsValidCard(self.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = self.PlayerCards[i][j]
				cn++
			}
		}
		//self.LastScore[i] = endgameMsg.Score[i]
		//self.Total[i] += endgameMsg.Score[i]
	}

	//游戏记录
	self.OnWriteGameRecord(wChairID, "强退游戏结束")

	//记录异常结束数据
	//self.addReplayOrder(wChairID, E_Li_Xian, 0)

	//回放增加结束数据
	self.addReplayOrder(wChairID, meta.DG_REPLAY_OPT_END_GAME, 0, []int{})

	//写入游戏回放数据,写完重置当前回放数据
	self.TableWriteOutDate()

	//将玩家分数发送给玩家
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d", endgameMsg.Score[i], self.GetUserTotalScore(uint16(i))+endgameMsg.Score[i])
		self.OnWriteGameRecord(i, recrodStr)
	}

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, endgameMsg) //保存，用于汇总计算
	self.SaveGameData()
	self.SendTableMsg(constant.MsgTypeGameEnd, endgameMsg)
	//数据库写分
	for _, v := range self.PlayerInfo { //i := 0; i < self.GetPlayerCount(); i++ {
		wintype := public.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinLose[v.Seat] == 1 {
			wintype = public.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = public.ScoreKind_Lost //enScoreKind_Lost;
		}
		if endgameMsg.EndType == public.TY_ESCAPE {
			if v.Seat == wChairID {
				wintype = public.ScoreKind_Flee
			} else {
				wintype = public.ScoreKind_pass //逃跑活动分数在对战统计中忽略
			}
		}
		//iAward := self.GetTaskAward(v.Seat)//金豆任务，先留着备用
		self.TableWriteGameDate(int(self.CurCompleteCount), v, wintype, endgameMsg.Score[v.Seat])
	}

	self.UpdateOtherFriendDate(&endgameMsg, true)
	//结束游戏
	self.CalculateResultTotal(public.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	self.CurCompleteCount = 0
	//self.SetGameStatus(public.GS_MJ_FREE)
	self.ConcludeGame()
	self.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 解散，结束游戏
func (self *Game_jl_jlkj) OnGameEndDismiss(wChairID uint16, cbReason byte) bool {
	//if self.Rule.HasPao && self.PayPaoStatus {
	//	self.CurCompleteCount--
	//}
	//变量定义
	var endgameMsg public.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = public.GER_DISMISS
	endgameMsg.TheOrder = self.CurCompleteCount

	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.Score[i] = endgameMsg.Score[i] - self.Spay
		cn := 0
		for j := 0; j < meta.TS_MAXHANDCARD; j++ {
			if self.m_GameLogic.IsValidCard(self.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = self.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = self.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)
		//self.LastScore[i] = endgameMsg.Score[i]
		//self.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = self.GetUserTotalScore(i) + endgameMsg.Score[i]

		if self.MaxScore[i] < self.PlayerCardScore[i] {
			self.MaxScore[i] = self.PlayerCardScore[i]
		}

		if self.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			//self.TotalFirstTurn[i]++
			// 记录在玩家身上
			_userItem := self.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				_userItem.Ctx.PlayTurn1st++
			}
		}

		endgameMsg.TheTurn[i] = self.PlayerTurn[i]
		endgameMsg.GetScore[i] = self.PlayerCardScore[i]
	}
	if self.WhoRoar < meta.MAX_PLAYER {
		self.TotalDuPai[self.WhoRoar]++
	}
	endgameMsg.HouPaiChair = self.WhoRoar
	endgameMsg.TheBank = self.Banker
	endgameMsg.TheParter = self.BankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", endgameMsg.Score[i], endgameMsg.TotalScore[i], self.PlayerCardScore[i])
		self.OnWriteGameRecord(i, recrodStr)
	}
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.XiScore[i] = self.XiScore[i] //喜钱
	}

	//记录异常结束数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == public.US_OFFLINE && i != int(wChairID) {
			//self.addReplayOrder(uint16(i), E_Li_Xian, 0)
		}
	}

	//回放增加结束数据
	self.addReplayOrder(wChairID, meta.DG_REPLAY_OPT_DIS_GAME, 0, []int{})

	//写入游戏回放数据,写完重置当前回放数据
	self.TableWriteOutDate()

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, endgameMsg) //保存，用于汇总计算
	self.SaveGameData()
	self.SendTableMsg(constant.MsgTypeGameEnd, endgameMsg)

	//游戏记录
	self.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDismiss")

	//数据库写分
	for _, v := range self.PlayerInfo {
		if v.Seat != public.INVALID_CHAIR {
			if wChairID != public.INVALID_CHAIR && v.Seat == wChairID {
				self.TableWriteGameDate(int(self.CurCompleteCount), v, public.ScoreKind_pass, endgameMsg.Score[v.Seat])
			} else {
				self.TableWriteGameDate(int(self.CurCompleteCount), v, public.ScoreKind_pass, endgameMsg.Score[v.Seat])
			}
		}

	}

	self.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	self.CalculateResultTotal(public.GER_DISMISS, wChairID, 0)
	self.GameEndStatus = public.GS_MJ_END
	//结束游戏
	//self.SetGameStatus(public.GS_MJ_FREE)
	self.ConcludeGame()

	self.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 程序异常，解散游戏
func (self *Game_jl_jlkj) OnGameEndErrorDismiss(wChairID uint16, cbReason byte) bool {
	//if self.Rule.HasPao && self.PayPaoStatus {
	//	self.CurCompleteCount--
	//}
	//变量定义
	var endgameMsg public.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = public.GER_DISMISS
	endgameMsg.TheOrder = self.CurCompleteCount

	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.Score[i] = endgameMsg.Score[i] - self.Spay
		cn := 0
		for j := 0; j < meta.TS_MAXHANDCARD; j++ {
			if self.m_GameLogic.IsValidCard(self.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = self.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = self.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)
		//self.LastScore[i] = endgameMsg.Score[i]
		//self.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = self.GetUserTotalScore(i) + endgameMsg.Score[i]

		if self.MaxScore[i] < self.PlayerCardScore[i] {
			self.MaxScore[i] = self.PlayerCardScore[i]
		}

		if self.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			//self.TotalFirstTurn[i]++
			// 记录在玩家身上
			_userItem := self.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				_userItem.Ctx.PlayTurn1st++
			}
		}

		endgameMsg.TheTurn[i] = self.PlayerTurn[i]
		endgameMsg.GetScore[i] = self.PlayerCardScore[i]
	}
	if self.WhoRoar < meta.MAX_PLAYER {
		self.TotalDuPai[self.WhoRoar]++
	}
	endgameMsg.HouPaiChair = self.WhoRoar
	endgameMsg.TheBank = self.Banker
	endgameMsg.TheParter = self.BankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", endgameMsg.Score[i], endgameMsg.TotalScore[i], self.PlayerCardScore[i])
		self.OnWriteGameRecord(i, recrodStr)
	}
	for i := 0; i < public.MAX_PLAYER_4P; i++ {
		endgameMsg.XiScore[i] = self.XiScore[i] //喜钱
	}

	//记录异常结束数据
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == public.US_OFFLINE && i != int(wChairID) {
			//self.addReplayOrder(uint16(i), E_Li_Xian, 0)
		}
	}

	//回放增加结束数据
	self.addReplayOrder(wChairID, meta.DG_REPLAY_OPT_DIS_GAME, 0, []int{})

	//写入游戏回放数据,写完重置当前回放数据
	self.TableWriteOutDate()

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, endgameMsg) //保存，用于汇总计算
	self.SaveGameData()
	self.SendTableMsg(constant.MsgTypeGameEnd, endgameMsg)

	//游戏记录
	self.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDismiss")

	self.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	self.CalculateResultTotal(public.GER_DISMISS, wChairID, 1)
	self.GameEndStatus = public.GS_MJ_END
	//结束游戏
	//self.SetGameStatus(public.GS_MJ_FREE)
	self.ConcludeGame()

	self.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 游戏结束,流局结束，统计积分
func (self *Game_jl_jlkj) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if self.GameEndStatus == public.GS_MJ_END && cbReason == public.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	//by leon总结算时才能
	//m_pITableFrame->KillGameTimer(IDI_OUT_TIME);
	// 清除超时检测
	for _, v := range self.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}
	//1.如果提供的用户为空，不可能，直接返回false
	switch cbReason {
	case public.GER_NORMAL: //常规结束

		return self.OnGameEndNormal(wChairID, meta.GOT_NORMAL)
	case public.GER_USER_LEFT: //用户强退

		return self.OnGameEndUserLeft(wChairID, meta.GOT_ESCAPE)

	case public.GER_DISMISS: //解散游戏

		return self.OnGameEndDismiss(wChairID, meta.GOT_DISMISS)

	case public.GER_GAME_ERROR: //程序异常，解散游戏

		return self.OnGameEndErrorDismiss(wChairID, meta.GOT_DISMISS)

	}
	return false
}

// 计算总发送总结算
func (self *Game_jl_jlkj) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	// 给客户端发送总结算数据
	var balanceGame public.Msg_S_DG_BALANCE_GAME
	balanceGame.Userid = self.Rule.FangZhuID
	balanceGame.CurTotalCount = self.CurCompleteCount //总盘数
	self.TimeEnd = time.Now().Unix()                  //大局结束时间
	balanceGame.TimeStart = self.GameCommon.TimeStart //游戏大局开始时间
	balanceGame.TimeEnd = self.TimeEnd
	// 存在换桌的情况下 总结算根据座位相加得到的数据是错误的  需要根据记录在玩家身上的积分计算
	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		for _, v := range _userItem.Ctx.GameScoreFen {
			balanceGame.GameScore[i] += v
		}
	}

	//打印日志
	recrodStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）总分: %d, %d, %d, %d", cbReason, balanceGame.GameScore[0], balanceGame.GameScore[1], balanceGame.GameScore[2], balanceGame.GameScore[3])
	self.OnWriteGameRecord(wChairID, recrodStr)

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
		// 有大赢家
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
		if self.ReWriteRec <= 1 {
			wintype = public.ScoreKind_pass
		}
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
			balanceGame.FirstTurnCount[i] = _userItem.Ctx.PlayTurn1st
			balanceGame.WinCount[i] = _userItem.Ctx.WinCount

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

	//小结算保留十分钟
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
		if len(self.VecGameData[i]) > 0 {
			gamedataStr = public.HF_JtoA(self.VecGameData[i][len(self.VecGameData[i])-1])
		}

		self.SaveLastGameinfo(_userItem.Uid, gameendStr, public.HF_JtoA(balanceGame), gamedataStr)
	}

	// 记录用户好友房历史战绩
	if wintype != public.ScoreKind_pass {
		self.TableWriteHistoryRecordWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
		self.TableWriteHistoryRecordDetailWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
	}

	balanceGame.End = 0

	//发消息
	self.SendTableMsg(constant.MsgTypeGameBalanceGame, balanceGame)

	self.resetEndDate()
}

func (self *Game_jl_jlkj) resetEndDate() {
	self.CurCompleteCount = 0
	self.VecGameEnd = []public.Msg_S_DG_GameEnd{}
	self.VecGameData = [meta.MAX_PLAYER][]public.CMD_S_DG_StatusPlay{}

	for _, v := range self.PlayerInfo {
		v.OnEnd()
	}
}

func (self *Game_jl_jlkj) UpdateOtherFriendDate(GameEnd *public.Msg_S_DG_GameEnd, bEnd bool) {

}

// ! 给客户端发送总结算
func (self *Game_jl_jlkj) CalculateResultTotal_Rep(msg *public.Msg_C_BalanceGameEeq) {
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
				balanceGame.GameScore[j] += self.VecGameEnd[i].Score[j] //总分
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
}

// 保存桌面数据
func (self *Game_jl_jlkj) SaveGameData() {
	//变量定义
	var StatusPlay public.CMD_S_DG_StatusPlay
	//游戏变量
	StatusPlay.Overtime = 0
	if time.Now().Unix()+1 < self.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.Overtime = self.LimitTime - time.Now().Unix()
	}

	StatusPlay.GameState = byte(self.GameState) //游戏状态，当前处在哪个阶段，1吼牌阶段，2打牌阶段
	StatusPlay.BankerUser = self.Banker         //庄家
	StatusPlay.CurrentUser = self.Whoplay       //当前牌权玩家
	StatusPlay.CellScore = self.GetCellScore()  //m_pGameServiceOption->lCellScore;
	StatusPlay.WhoLastOut = self.WhoLastOut     //上一个出牌玩家
	StatusPlay.RoarPai = self.RoarPai           //叫的什么牌
	StatusPlay.WhoRoar = self.WhoRoar           //谁叫了牌
	StatusPlay.WhoMJ = public.INVALID_CHAIR     //初始化谁鸣鸡为无效值
	if self.BMingJiFlag {
		StatusPlay.WhoMJ = self.BankParter //谁鸣鸡
	}
	StatusPlay.TurnScore = self.CardScore           //本轮分
	StatusPlay.LastPaiType = self.LastOutTypeClient //上一次出牌的类型，可能没用？修改成客户端传过来的
	StatusPlay.TheOrder = self.CurCompleteCount
	StatusPlay.OutScorePai = self.OutScorePai //所有分牌

	for i := 0; i < self.GetPlayerCount(); i++ {

		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == public.US_OFFLINE {
			StatusPlay.WhoBreak[i] = true //掉线的有那些人
		}
		StatusPlay.TuoGuanPlayer[i] = self.TuoGuanPlayer[i] //托管的有那些人
		//StatusPlay.LastScore[i] = int(self.LastScore[i])    		//上一轮输赢
		StatusPlay.Total[i] = self.GetUserTotalScore(uint16(i)) //总输赢
		StatusPlay.WhoPass[i] = self.WhoPass[i]                 //谁放弃
		StatusPlay.WhoReady[i] = self.WhoReady[i]               //谁已经完成叫牌过程
		StatusPlay.Score[i] = self.PlayerCardScore[i]           //每个人的分

		for j := 0; j < public.MAX_CARD_4P; j++ {
			StatusPlay.OutCard[i][j] = self.AllPaiOut[i][j]      //刚才出的牌
			StatusPlay.LastOutCard[i][j] = self.LastPaiOut[i][j] //上一轮出的牌，（这个其实可以不要）
		}
	}

	//玩家的个人数据
	for i := 0; i < self.GetPlayerCount() && i < meta.MAX_PLAYER; i++ {
		StatusPlay.WhoReLink = uint16(i)             //谁断线重连的
		StatusPlay.TrustCounts = self.TrustCounts[i] //叫我托管了几次了
		StatusPlay.MyCards = self.PlayerCards[i]
		StatusPlay.MyCardsCount = self.ThePaiCount[i]
		self.VecGameData[i] = append(self.VecGameData[i], StatusPlay) //保存，用于汇总计算
	}
}

func (self *Game_jl_jlkj) sendGameSceneStatusPlay(player *modules.Player) bool {

	wChiarID := player.GetChairID()

	if wChiarID >= public.MAX_PLAYER_4P {
		syslog.Logger().Info("sendGameSceneStatusPlay invalid chair")
		return false
	}

	//发送底分
	var msgRule public.Msg_S_DG_GameRule
	msgRule.CellScore = self.GetCellScore()
	msgRule.FaOfTao = self.FaOfTao
	self.SendTableMsg(constant.MsgTypeGameRule, msgRule)

	//self.WhoBreak[wChiarID] = false;//重连嘛，所以取消断线

	//提示其他玩家，我又回来了！
	var msgTip public.Msg_S_DG_ReLinkTip
	msgTip.ReLinkUser = wChiarID
	msgTip.ReLinkTip = 0
	self.SendTableMsg(constant.MsgTypeReLinkTip, msgTip)

	//取消托管

	//变量定义
	var StatusPlay public.CMD_S_DG_StatusPlay
	//游戏变量
	StatusPlay.Overtime = 0
	if time.Now().Unix()+1 < self.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.Overtime = self.LimitTime - time.Now().Unix()
	}

	StatusPlay.GameState = byte(self.GameState)         //游戏状态，当前处在哪个阶段，1吼牌阶段，2打牌阶段
	StatusPlay.BankerUser = self.Banker                 //庄家
	StatusPlay.CurrentUser = self.Whoplay               //当前牌权玩家
	StatusPlay.CellScore = self.GetCellScore()          //m_pGameServiceOption->lCellScore;
	StatusPlay.WhoReLink = wChiarID                     //谁断线重连的
	StatusPlay.WhoLastOut = self.WhoLastOut             //上一个出牌玩家
	StatusPlay.TrustCounts = self.TrustCounts[wChiarID] //叫我托管了几次了
	StatusPlay.RoarPai = self.RoarPai                   //叫的什么牌
	StatusPlay.WhoRoar = self.WhoRoar                   //谁叫了牌
	StatusPlay.WhoMJ = public.INVALID_CHAIR             //初始化谁鸣鸡为无效值
	if self.BMingJiFlag {
		StatusPlay.WhoMJ = self.BankParter //谁鸣鸡
	}
	StatusPlay.TurnScore = self.CardScore           //本轮分
	StatusPlay.LastPaiType = self.LastOutTypeClient //上一次出牌的类型，可能没用？修改成客户端传过来的
	StatusPlay.TheOrder = self.CurCompleteCount
	StatusPlay.OutScorePai = self.OutScorePai //所有分牌

	for j := 0; j < public.MAX_CARD_4P; j++ {
		if self.m_GameLogic.IsValidCard(self.PlayerCards[wChiarID][j]) {
			StatusPlay.MyCards[j] = self.PlayerCards[wChiarID][j] //刚才出的牌
			StatusPlay.MyCardsCount++
		}
	}
	for i := 0; i < self.GetPlayerCount(); i++ {

		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == public.US_OFFLINE {
			StatusPlay.WhoBreak[i] = true //掉线的有那些人
		}
		StatusPlay.TuoGuanPlayer[i] = self.TuoGuanPlayer[i] //托管的有那些人
		//StatusPlay.LastScore[i] = int(self.LastScore[i])    		//上一轮输赢
		StatusPlay.Total[i] = self.GetUserTotalScore(uint16(i)) //总输赢
		StatusPlay.WhoPass[i] = self.WhoPass[i]                 //谁放弃
		StatusPlay.WhoReady[i] = self.WhoReady[i]               //谁已经完成叫牌过程
		StatusPlay.Score[i] = self.PlayerCardScore[i]           //每个人的分

		for j := 0; j < public.MAX_CARD_4P; j++ {
			StatusPlay.OutCard[i][j] = self.AllPaiOut[i][j]      //刚才出的牌
			StatusPlay.LastOutCard[i][j] = self.LastPaiOut[i][j] //上一轮出的牌，（这个其实可以不要）
		}
	}

	//发送场景
	self.SendPersonMsg(constant.MsgTypeGameStatusPlay, StatusPlay, wChiarID)

	//出完牌的顺序，游次
	self.SendPlayerTurn(wChiarID)

	//剩余牌数量
	if self.ShowHandCardCnt {
		self.SendPaiCount(wChiarID)
	}

	if self.GameType == meta.GT_NORMAL {
		self.SendPlayerScore(wChiarID, public.INVALID_CHAIR, 0) //每个人的分,这个找朋友模式才有
	}

	//self.SendTaskID(false, wChiarID);
	self.GameTask.SendTaskID(self.GameCommon, false, wChiarID)

	//如果断线之前我的牌出完了,把队友的牌发给我
	if (self.WhoAllOutted[wChiarID] && self.SeeTeamerCard) || self.SeePartnerCards {
		teamer := self.GetTeamer(wChiarID)
		self.SendPaiToTeamer(teamer, wChiarID)
	}

	//发送权限（什么权限、该谁出牌、倒计时等）
	if self.GameState == meta.GsRoarPai {
		self.SendPower(self.Whoplay, 1, int(StatusPlay.Overtime))
	} else if self.GameState == meta.GsPlay {
		self.SendPower(self.Whoplay, 2, int(StatusPlay.Overtime))
	} else if self.GameState == meta.Gs4KingScore {
		self.SendPower(self.WhoHas4KingPower, 4, int(StatusPlay.Overtime))
	}

	//发小结消息
	if byte(len(self.VecGameEnd)) == self.CurCompleteCount && self.CurCompleteCount != 0 && int(wChiarID) < self.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := self.VecGameEnd[self.CurCompleteCount-1]
		gamend.Relink = 1 //表示为断线重连

		self.SendPersonMsg(constant.MsgTypeGameEnd, gamend, wChiarID)
	}

	//发送解散房间所有玩家的反应
	self.SendAllPlayerDissmissInfo(player)

	return true
}

// 游戏场景消息发送
func (self *Game_jl_jlkj) SendGameScene(uid int64, status byte, secret bool) {
	player := self.GetUserItemByUid(uid)
	if player == nil {
		return
	}
	switch status {
	case public.GS_MJ_FREE:
		self.sendGameSceneStatusFree(player)
	case public.GS_MJ_PLAY:
		self.sendGameSceneStatusPlay(player)
	}
}
func (self *Game_jl_jlkj) sendGameSceneStatusFree(player *modules.Player) bool {

	//变量定义
	var StatusFree public.Msg_S_DG_StatusFree
	//构造数据
	StatusFree.BankerUser = self.Banker
	StatusFree.CellScore = self.GetCellScore() //self.m_pGameServiceOption->lCellScore;
	StatusFree.FaOfTao = self.FaOfTao
	StatusFree.CellMinScore = self.GetCellScore() //最低分
	StatusFree.CellMaxScore = self.GetCellScore() //最低分

	//发送场景
	//	self.SendPersonMsg(constant.MsgTypeGameStatusFree, StatusFree, PlayerInfo.GetChairID())
	self.SendUserMsg(constant.MsgTypeGameStatusFree, StatusFree, player.Uid)

	return true
}

// ! 游戏退出
func (self *Game_jl_jlkj) OnExit(uid int64) {
	self.GameCommon.OnExit(uid)
}

func (self *Game_jl_jlkj) OnTime() {
	self.GameCommon.OnTime()
}

// ! 写游戏日志
func (self *Game_jl_jlkj) OnWriteGameRecord(seatId uint16, recordStr string) {
	self.GameTable.WriteTableLog(seatId, recordStr)
}

// ! 写入游戏回放数据
func (self *Game_jl_jlkj) TableWriteOutDate() {
	if self.ReWriteRec != 0 {
		self.ReWriteRec++ //这种情况会>1，表示是在结算时申请解散的。
		// 写完后清除数据
		self.ReplayRecord.ReSet()
		return
	}

	recordReplay := new(model.RecordGameReplay)
	recordReplay.GameNum = self.GetTableInfo().GameNum
	recordReplay.RoomNum = self.GetTableInfo().Id
	recordReplay.PlayNum = int(self.CurCompleteCount)
	recordReplay.ServerId = gsvr.GetServer().Con.Id
	recordReplay.HandCard = self.m_GameLogic.GetWriteHandReplayRecordString(self.ReplayRecord)
	recordReplay.OutCard = self.m_GameLogic.GetWriteOutReplayRecordString(self.ReplayRecord)
	recordReplay.KindID = self.GetTableInfo().KindId
	recordReplay.CardsNum = 0

	if id, err := gsvr.GetDBMgr().InsertGameRecordReplay(recordReplay); err != nil {
		syslog.Logger().Debug(fmt.Sprintf("%d,写游戏出牌记录：（%v）出错（%v）", id, recordReplay, err))
	}

	self.RoundReplayId = recordReplay.Id

	// 写完后清除数据
	self.ReplayRecord.ReSet()

	self.ReWriteRec++ //在小结算过程中解散不在写回放记录了
}

// 场景保存
func (self *Game_jl_jlkj) Tojson() string {
	var _json modules.DGGameJsonSerializer

	_json.ToJsonDG(&self.GameMetaDG)

	_json.GameCommonToJson(&self.GameCommon)

	return public.HF_JtoA(&_json)
}

// 场景恢复
func (self *Game_jl_jlkj) Unmarsha(data string) {
	var _json modules.DGGameJsonSerializer

	if data != "" {
		json.Unmarshal([]byte(data), &_json)

		_json.UnmarshaDG(&self.GameMetaDG)
		_json.JsonToStruct(&self.GameCommon)

		self.ParseRule(self.GetTableInfo().Config.GameConfig)
		self.m_GameLogic.Rule = self.Rule

		//_game.m_GameLogic.InitMagicPoint(CARDINDEX_SMALL) //大小王都是赖子
		//_game.m_GameLogic.AddMagicPoint(CARDINDEX_BIG)
		self.m_GameLogic.SetBombCount(3)                         //设置炸弹的最小长度(普通炸弹：四个或者四个以上相同的牌 5 10 K也是炸弹3张牌)
		self.m_GameLogic.SetOnestrCount(254)                     //设置单顺的最小长度
		self.m_GameLogic.SetMaxCardCount(public.MAX_CARD_4P)     //设置手牌最大长度
		self.m_GameLogic.SetMaxPlayerCount(public.MAX_PLAYER_4P) //设置玩家最大数目
		self.m_GameLogic.SetHard510KMode(self.Hard510KMode)      //设置纯510K是否大过四炸
	}
}

// 解析配置的任务,格式： "1@5/2@5/33@10"
func (self *Game_jl_jlkj) ParseTaskConfig(data string) {
	self.GameTask.Init()
	//.....
	//self.GameTask.AppendTaskMapAndVec(id,award)
}

// 获取特殊牌型分,玩家手中的8喜7喜等数据记录
func (self *Game_jl_jlkj) GetSpecialCardTypeScore() {
	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		//8喜
		self.Who8Xi[i] = self.m_GameLogic.Get8XibombNum(self.PlayerCards[i])
		//7喜
		self.Who7Xi[i] = self.m_GameLogic.Get7XibombNum(self.PlayerCards[i])
		//正510K
		self.WhoSame510K[i] = self.m_GameLogic.GetSame510kNum(self.PlayerCards[i])
		//天炸
		if self.WhoHasKingBomb >= public.MAX_PLAYER_4P && self.m_GameLogic.HasKingBomb(self.PlayerCards[i]) > 0 {
			self.WhoHasKingBomb = i
		}
		if self.Who8Xi[i] > 0 || self.Who7Xi[i] > 0 || self.WhoHasKingBomb == i {
			self.WhoTotal8Xi[i] += self.Who8Xi[i]
			self.WhoTotal7Xi[i] += self.Who7Xi[i]
			//详细日志
			byKingBombNum := 0
			if self.WhoHasKingBomb == i {
				byKingBombNum = 1
				self.WhoToTalMore4KingCount[i] += 1
			}
			LogStr := fmt.Sprintf("获得了%d个7喜,%d个8喜,%d个同色510k,%d个天炸", self.Who7Xi[i], self.Who8Xi[i], self.WhoSame510K[i], byKingBombNum)
			self.OnWriteGameRecord(uint16(i), LogStr)
		}
	}
}

// 参数:是否是庄家一个队伍的
func (self *Game_jl_jlkj) GetTeamYouType(isBankerTeam bool) int {
	youb := self.PlayerTurn[self.Banker]
	youp := self.PlayerTurn[self.BankParter]

	youtype := 0
	if (youb == 1 && youp == 2) || (youb == 2 && youp == 1) {
		//庄家及队友12游
		if isBankerTeam {
			youtype = YOUTYPE_12
		} else {
			youtype = YOUTYPE_34
		}
	} else if (youb == 1 && youp == 3) || (youb == 3 && youp == 1) {
		if isBankerTeam {
			youtype = YOUTYPE_13
		} else {
			youtype = YOUTYPE_24
		}
	} else if (youb == 1 && youp == 4) || (youb == 4 && youp == 1) {
		if isBankerTeam {
			youtype = YOUTYPE_14
		} else {
			youtype = YOUTYPE_23
		}
	} else if (youb == 2 && youp == 3) || (youb == 3 && youp == 2) {
		if isBankerTeam {
			youtype = YOUTYPE_23
		} else {
			youtype = YOUTYPE_14
		}
	} else if (youb == 2 && youp == 4) || (youb == 4 && youp == 2) {
		if isBankerTeam {
			youtype = YOUTYPE_24
		} else {
			youtype = YOUTYPE_13
		}
	} else if (youb == 3 && youp == 4) || (youb == 4 && youp == 3) {
		if isBankerTeam {
			youtype = YOUTYPE_34
		} else {
			youtype = YOUTYPE_12
		}
	}

	return youtype
}

// 参数:是否是庄家一个队伍的
func (self *Game_jl_jlkj) IsBankerTeam(seat uint16) bool {
	if seat == self.Banker || seat == self.BankParter {
		return true
	}
	return false
}

// 获取双方队伍抓分
// 庄家队伍,闲家队伍
func (self *Game_jl_jlkj) GetJLKJParterScore(bCalcKingScore bool) (int, int) {
	bScore := 0 //庄家队伍抓的分
	xScore := 0 //非庄家队伍抓的分

	if bCalcKingScore {
		bScore = self.PlayerCardScore[self.Banker] + self.PlayerCardScore[self.BankParter]
		for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
			if !self.IsBankerTeam(i) {
				xScore += self.PlayerCardScore[i]
			}
		}
	} else {
		// 先减去4王换分
		if self.FourKingScore > 0 && self.WhoHas4KingScore != public.INVALID_CHAIR {
			self.PlayerCardScore[self.WhoHas4KingScore] -= self.FourKingScore
		}

		bScore = self.PlayerCardScore[self.Banker] + self.PlayerCardScore[self.BankParter]
		for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
			if !self.IsBankerTeam(i) {
				xScore += self.PlayerCardScore[i]
			}
		}

		// 再恢复回来
		if self.FourKingScore > 0 && self.WhoHas4KingScore != public.INVALID_CHAIR {
			self.PlayerCardScore[self.WhoHas4KingScore] += self.FourKingScore
		}
	}

	return bScore, xScore
}

func (self *Game_jl_jlkj) SuanFen(bankscore int, xianscore int) (int, bool) {
	if bankscore > xianscore {
		return bankscore - xianscore, true
	} else if bankscore < xianscore {
		return xianscore - bankscore, false
	}

	return 0, false
}

// 获取几游的座位号
func (self *Game_jl_jlkj) GetYouSeat(you byte) uint16 {
	if you > 4 || you < 1 {
		return public.INVALID_CHAIR
	}

	for i := uint16(0); i < public.MAX_PLAYER_4P; i++ {
		if self.PlayerTurn[i] == you {
			return i
		}
	}
	return public.INVALID_CHAIR
}

// 获取玩家总得分
func (self *Game_jl_jlkj) GetUserTotalScore(seat uint16) int {
	score := 0
	_userItem := self.GetUserItemByChair(seat)
	if _userItem != nil {
		for _, v := range _userItem.Ctx.GameScoreFen {
			score += v
		}
	}
	return score
}

// 得到一张王
func (self *Game_jl_jlkj) GetUserKingCard(seat uint16) byte {
	kingCard := byte(0)
	for i := 0; i < len(self.PlayerCards[seat]); i++ {
		if self.PlayerCards[seat][i] == logic.CARDINDEX_SMALL || self.PlayerCards[seat][i] == logic.CARDINDEX_BIG {
			kingCard = self.PlayerCards[seat][i]
			break
		}
	}
	return kingCard
}

// 用户是否拥有相同的牌
func (self *Game_jl_jlkj) IsHaveSameCard(card byte, seat uint16) bool {
	bFind := false
	for i := 0; i < len(self.PlayerCards[seat]); i++ {
		if self.PlayerCards[seat][i] == card {
			bFind = true
			break
		}
	}
	return bFind
}
