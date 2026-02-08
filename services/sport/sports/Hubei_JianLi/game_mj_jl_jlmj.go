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
斗棋麻将-监利麻将
*/

//好友房规则相关属性
/*
三人面板：底分、局数、去万、胡牌类型、坑庄、围增、出增
四人面板：底分、局数、补花、胡牌类型、坑庄、围增、出增
*/
type FriendRule_jlmj struct {
	Difen            int    `json:"difen"`            //底分
	Radix            int    `json:"scoreradix"`       //底分基数
	Nowan            string `json:"quwan"`            //没有万字牌
	BuHuaType        int    `json:"buhua"`            //补花类型(1.红中(默认) 2.红发白)
	HuType           int    `json:"hutype"`           //胡牌类型(1.自摸胡(默认) 2.点炮胡)
	IsKengZhuang     string `json:"kengzhuang"`       //是否坑庄
	WeiZeng          int    `json:"kengzeng"`         //围增(坑增)(1分、2分、3分，默认1分)
	IsChuZeng        string `json:"jiazeng"`          //是否出增(加增)(买炮或者下跑类似)
	ChuZengNum       int    `json:"jiazengnum"`       //兼容ChuZengNum
	NineSecondRoom   string `json:"sc9"`              //九秒场
	IpLimit          string `json:"ip"`               //ipgps限制
	Overtime_trust   int    `json:"overtime_trust"`   // 超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` // 超时解散
	Kechi            string `json:"kechi"`            //是否可吃
	LookonSupport    string `json:"LookonSupport"`    //本局游戏是否支持旁观
}

// 操作命令
type Msg_S_OperateResult_JLMJ struct {
	public.Msg_S_OperateResult
	OperateCode int `json:"operatecode"` //操作代码
}

// 操作命令
type Msg_S_OperateBuHuaNotify struct {
	OperateUser uint16 `json:"operateuser"` //补花用户
}

// 操作命令
type Msg_C_OperateCard_JLMJ struct {
	public.Msg_C_OperateCard
	Code int `json:"code"` //操作类型
}

type Game_mj_jl_jlmj struct {
	// 游戏共用部分
	modules.GameCommon
	// 游戏流程数据
	meta.GameMeta
	//游戏逻辑
	m_GameLogic       GameLogic_jlmj
	m_FlowerCardFan   [meta.MAX_PLAYER]int //花牌番
	M_outcard         byte                 //记录一下前一打出的是个什么牌
	M_broadCastFlower bool                 //是否可以播放补花动画
}

// ! 设置游戏可胡牌类型
func (self *Game_mj_jl_jlmj) HuTypeInit(_type *public.TagHuType) {
	_type.HAVE_SIXI_HU = false
	_type.HAVE_QUE_YISE_HU = false
	_type.HAVE_BANBAN_HU = false
	_type.HAVE_LIULIU_HU = false
	_type.HAVE_QING_YISE_HU = false
	_type.HAVE_FENG_YI_SE = false
	_type.HAVE_HAO_HUA_DUI_HU = false
	_type.HAVE_JIANG_JIANG_HU = false
	_type.HAVE_FENG_YISE_HU = false
	_type.HAVE_QUAN_QIU_REN = false
	_type.HAVE_PENG_PENG_HU = false
	_type.HAVE_HAI_DI_HU = false
	_type.HAVE_QIANG_GANG_HU = false
	_type.HAVE_GANG_SHANG_KAI_HUA = true
	_type.HAVE_MENG_QING = false
	_type.HAVE_DI_HU = false
	_type.HAVE_TIAN_HU = false
	_type.HAVE_ZIMO_JIAO_1 = false
}

// ! 获取游戏配置
func (self *Game_mj_jl_jlmj) GetGameConfig() *public.GameConfig { //获取游戏相关配置
	return &self.Config
}

// ! 重置桌子数据
func (self *Game_mj_jl_jlmj) RepositTable() {
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
	self.m_FlowerCardFan = [meta.MAX_PLAYER]int{0, 0, 0, 0}

	//状态变量
	self.GangFlower = false
	self.SendStatus = false
	self.HaveHuangZhuang = false
	for k, _ := range self.RepertoryCard {
		self.RepertoryCard[k] = 0
	}

	self.FanScore = [4]meta.Game_mj_fan_score{}

	for _, v := range self.PlayerInfo {
		v.Reset()
	}
	//结束信息
	self.ChiHuCard = 0
	//self.ReplayRecord.Reset()
}

// ! 解析配置的任务
func (self *Game_mj_jl_jlmj) ParseRule(strRule string) {

	syslog.Logger().Debug("parserRule :" + strRule)

	self.Rule.CreateType = 0
	self.Rule.NineSecondRoom = false

	self.Rule.FangZhuID = self.GetTableInfo().Creator
	self.Rule.JuShu = self.GetTableInfo().Config.RoundNum
	self.Rule.CreateType = self.FriendInfo.CreateType

	self.Rule.DiFen = 1                   //底分
	self.Rule.NoWan = false               //3人玩法独有
	self.Rule.BuHuaType = BuHua_Type_NULL //4人玩法独有的
	self.Rule.HuType = HuFa_Type_Zimo     //默认自摸胡
	self.Rule.WeiZeng = 0
	self.Rule.KengZhuang = false
	self.Rule.ChuZeng = false
	self.Config.PlayerCount = uint16(self.GetTableInfo().Config.MaxPlayerNum)
	self.Config.ChairCount = uint16(self.GetTableInfo().Config.MaxPlayerNum)

	if len(strRule) == 0 {
		return
	}

	/*
		三人面板：底分、局数、去万、胡牌类型、坑庄、围增、出增
		四人面板：底分、局数、补花、胡牌类型、坑庄、围增、出增
	*/

	var _msg FriendRule_jlmj
	if err := json.Unmarshal(public.HF_Atobytes(strRule), &_msg); err == nil {
		self.Rule.DiFen = _msg.Difen
		if _msg.Radix == 0 {
			self.Rule.Radix = 1
		} else {
			self.Rule.Radix = _msg.Radix
		}
		self.Rule.NineSecondRoom = _msg.NineSecondRoom == "true" //9秒场
		if _msg.Overtime_trust != 0 {
			self.Rule.NineSecondRoom = false
		}

		self.Rule.HuType = _msg.HuType                     //胡牌类型
		self.Rule.KengZhuang = _msg.IsKengZhuang == "true" //是否坑庄
		self.Rule.WeiZeng = _msg.WeiZeng                   //围增
		self.Rule.ChuZeng = _msg.IsChuZeng == "true"       //是否出增(类似下跑、漂)
		self.Rule.KeChi = _msg.Kechi == "true"
		if _msg.ChuZengNum != 0 {
			if _msg.ChuZengNum > 0 {
				self.Rule.ChuZeng = true
			} else {
				self.Rule.ChuZeng = false
			}
		}
		//20200313 沈强 1.2 补花可选，去万也改可选
		self.Rule.BuHuaType = _msg.BuHuaType   //4人玩法解析补花类型
		self.Rule.NoWan = _msg.Nowan == "true" //3人玩法解析去万
		//if self.GetProperPNum() == 3 {
		//	if len(_msg.IsChuZeng)!=0{
		//		fmt.Println(fmt.Sprint("改（%s）",_msg.IsChuZeng))
		//		self.Rule.BuHuaType=0
		//	}
		//}
		//if self.GetProperPNum() == 4{
		//	if len(_msg.IsChuZeng)!=0 {
		//		self.Rule.NoWan = false
		//	}
		//}
		self.Rule.Overtime_trust = _msg.Overtime_trust
		self.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		if _msg.LookonSupport == "" {
			self.Config.LookonSupport = true
		} else {
			self.Config.LookonSupport = _msg.LookonSupport == "true"
		}
	}
	cardclass := 0
	if self.Rule.NoWan {
		cardclass |= common.CARDS_WITHOUT_WAN
	}
	//根据补花类型再次修改
	switch self.Rule.BuHuaType {
	case 1:
		//只有红中
		cardclass |= (common.CARDS_WITHOUT_FA | common.CARDS_WITHOUT_BAI)
	case 2:
	//中发白
	//什么都不做
	default:
		cardclass |= common.CARDS_WITHOUT_DRAGON
	}
	self.m_GameLogic.SetCardClass(cardclass)
}

// ! 开局
func (self *Game_mj_jl_jlmj) OnBegin() {
	syslog.Logger().Debug("onbegin")
	self.RepositTable()

	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range self.PlayerInfo {
		v.OnBegin()
	}

	// 第一局开放玩家为庄家
	self.BankerUser = 0
	self.ParseRule(self.GetTableInfo().Config.GameConfig)
	self.m_GameLogic.Rule = self.Rule
	self.m_GameLogic.HuType = self.HuType
	self.CurCompleteCount = 0
	self.VecGameEnd = []public.Msg_S_GameEnd{}
	self.VecGameDataAllP = [4][]public.CMD_S_StatusPlay{}

	self.SetOfflineRoomTime(30 * 60)

	// 记录游戏开始时间
	self.GameCommon.GameBeginTime = time.Now()

	//self.GetTable().SetBegin(true)
	//self.GetTable().OnBegin()

	self.OnGameStart()
}

func (self *Game_mj_jl_jlmj) OnGameStart() {
	if self.CanContinue() {
		//发送最新状态
		for i := 0; i < self.GetPlayerCount(); i++ {
			self.SendUserStatus(i, public.US_PLAY) //把状态发给其他人
		}
		self.ReplayRecord.Reset()

		if self.Rule.ChuZeng {
			self.SendPaoSetting()
		} else {
			self.StartNextGame()
		}
	}
}

// ! 发送下跑对话框
func (self *Game_mj_jl_jlmj) SendPaoSetting() {
	self.GetTable().SetXiaPaoIng(true)
	self.GameEndStatus = public.GS_MJ_PLAY
	//设置状态
	self.SetGameStatus(public.GS_MJ_PLAY)

	for _, v := range self.PlayerInfo {
		v.Ctx.CleanWeaveItemArray()
		v.Ctx.InitCardIndex()
	}

	self.PayPaoStatus = true //设置玩家选漂的状态
	var PaoSetting public.Msg_S_PaoSetting
	//向每个玩家发送数据
	for _, v := range self.PlayerInfo {
		self.LockTimeOut(v.Seat)
		if v.Ctx.UserPaoReady == true {
			PaoSetting.PaoStatus = true
			PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
			PaoSetting.Always = v.Ctx.VecXiaPao.Status
			PaoSetting.BankerUser = self.BankerUser
			PaoSetting.Overtime = v.Ctx.CheckTimeOut
		} else {
			PaoSetting.PaoStatus = false
			PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
			PaoSetting.Always = v.Ctx.VecXiaPao.Status
			PaoSetting.BankerUser = self.BankerUser
			PaoSetting.Overtime = v.Ctx.CheckTimeOut
		}
		self.SendPersonMsg(constant.MsgTypeGamePaoSetting, PaoSetting, v.Seat)
	}
}

func (self *Game_mj_jl_jlmj) SendPaoSettingOffline() {
	var PaoSetting public.Msg_S_PaoSetting
	//向每个玩家发送数据
	for _, v := range self.PlayerInfo {
		if v.Ctx.UserPaoReady {
			continue
		}
		PaoSetting.PaoStatus = false
		PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
		PaoSetting.Always = v.Ctx.VecXiaPao.Status
		PaoSetting.BankerUser = self.BankerUser
		PaoSetting.Overtime = v.Ctx.CheckTimeOut
		self.SendPersonMsg(constant.MsgTypeGamePaoSetting, PaoSetting, v.Seat)
	}
}

// ! 玩家选择跑
func (self *Game_mj_jl_jlmj) OnUserClientXiaPao(msg *public.Msg_C_Xiapao) bool {
	nChiarID := self.GetChairByUid(msg.Id)
	_userItem := self.GetUserItemByChair(nChiarID)
	if _userItem == nil {
		return false
	}
	if nChiarID >= 0 && nChiarID < meta.MAX_PLAYER {
		_userItem.Ctx.XiaPao(msg)

		self.NotifyXiaPao(nChiarID)
		//fmt.Println(fmt.Sprintf("玩家%d,选跑%d", nChiarID, msg.Num))
	}

	// 如果4个玩家都准备好了，自动开启下一局
	_beginCount := 0
	for _, v := range self.PlayerInfo {
		if !v.Ctx.UserPaoReady {
			recordStr := fmt.Sprintf("椅子编号[%d] 还没有完成选跑", v.Seat)
			self.OnWriteGameRecord(uint16(v.Seat), recordStr)
			break
		}
		_beginCount++
	}

	if _beginCount >= self.GetPlayerCount() {
		self.OnWriteGameRecord(uint16(nChiarID), "所有人都完成选跑了，开始游戏")
		self.PayPaoStatus = false
		//游戏没有开始发牌
		if !self.GameStartForXiapao {
			self.StartNextGame()
		}
	}

	return true
}

// ! 广播玩家的状态和选漂的数目
func (self *Game_mj_jl_jlmj) NotifyXiaPao(wChairID uint16) bool {
	var sXiaPiao public.Msg_S_Xiapao

	for _, v := range self.PlayerInfo {
		if v.Seat != public.INVALID_CHAIR {
			sXiaPiao.Num[v.Seat] = v.Ctx.VecXiaPao.Num
			sXiaPiao.Always[v.Seat] = v.Ctx.VecXiaPao.Status
			sXiaPiao.Status[v.Seat] = v.Ctx.UserPaoReady
		}
	}

	//发送数据
	self.SendTableMsg(constant.MsgTypeGameXiapao, sXiaPiao)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameXiapao, sXiaPiao)

	//游戏记录
	if wChairID == wChairID {
		recordStr := fmt.Sprintf("发送跑数：%d， 是否默认 %t", sXiaPiao.Num[wChairID], sXiaPiao.Status[wChairID])
		self.OnWriteGameRecord(wChairID, recordStr)

		self.addReplayOrder(wChairID, info.E_Pao, sXiaPiao.Num[wChairID])
	}
	return true
}

// ! 开始下一局游戏
func (self *Game_mj_jl_jlmj) StartNextGame() {
	self.OnStartNextGame()

	self.LastSendCardUser = uint16(public.INVALID_CHAIR)

	//组合扑克
	self.MagicCard = 0x00

	self.LeftCardCount = 0
	self.RepertoryCard = []byte{}

	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始StartNextGame......")
	self.OnWriteGameRecord(public.INVALID_CHAIR, self.GetTableInfo().Config.GameConfig)

	for _, v := range self.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	self.ParseRule(self.GetTableInfo().Config.GameConfig)
	self.m_GameLogic.Rule = self.Rule

	//设置状态
	self.SetGameStatus(public.GS_MJ_PLAY)
	self.GameEndStatus = public.GS_MJ_PLAY

	self.CurCompleteCount++
	self.GetTable().SetBegin(true)

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(self.GetTableId()+self.KIND_ID*100+self.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	self.SiceCount = modules.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	self.LeftCardCount = byte(len(self.RepertoryCard))
	self.LeftBu = 10 //剩下的补牌数

	//这里在没有调用混乱扑克的函数时m_cbRepertoryCard中是空的，当它调用了这个函数之后
	//在这个函数中把固定的牌打乱后放到这个数组中，在放的同时不断增加数组m_cbRepertoryCard
	//的长度
	self.LeftCardCount, self.RepertoryCard = self.m_GameLogic.RandCardData()

	self.CreateLeftCardArray(self.GetPlayerCount(), int(self.LeftCardCount), false)
	//--------------------------------
	cardsIndex := common.CardsToCardIndex(self.RepertoryCard)
	//common.Print_cards(cardsIndex[:])
	recordstr := fmt.Sprintf("补花选择（%d）去万（%t）混牌数（%d）其中红中（%d）发财（%d）白板（%d）", self.Rule.BuHuaType, self.Rule.NoWan, self.LeftCardCount, cardsIndex[31], cardsIndex[32], cardsIndex[33])
	self.OnWriteGameRecord(public.INVALID_CHAIR, recordstr)
	//---------------------------------------
	//分发扑克--即每一个人解析他的14张牌结果存放在m_cbCardIndex[i]中

	for _, v := range self.PlayerInfo {
		if v.Seat != public.INVALID_CHAIR {
			self.LeftCardCount -= (public.MAX_COUNT - 1)
			v.Ctx.SetCardIndex(&self.Rule, self.RepertoryCard[self.LeftCardCount:], public.MAX_COUNT-1)
		}
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	err := self.InitDebugCards("mahjongjlmj_test", &self.RepertoryCard, &self.BankerUser)
	if err != nil {
		self.OnWriteGameRecord(public.INVALID_CHAIR, err.Error())
	}
	//////////////读取配置文件设置牌型end////////////////////////////////////
	//发送扑克---这是发送给庄家的第十四张牌
	self.SendCardCount++
	self.LeftCardCount--
	self.SendCardData = self.RepertoryCard[self.LeftCardCount]
	self.M_outcard = public.INVALID_BYTE
	//监利麻将没有皮籁
	//self.LeftCardCount--
	//self.PiZiCard = self.RepertoryCard[self.LeftCardCount]
	//
	////转换赖子值
	//cbValue := byte(self.PiZiCard & public.MASK_VALUE)
	//cbColor := byte(self.PiZiCard & public.MASK_COLOR)
	//
	//if cbValue == 9 && cbColor <= 0x20 {
	//	//牌值等于9,但是牌花色是万 同 条(九万九筒九条)
	//	cbValue = 0
	//	self.MagicCard = (cbValue + 1) | cbColor
	//} else {
	//	self.MagicCard = (cbValue + 1) | cbColor
	//}
	//
	//self.m_GameLogic.SetMagicCard(self.MagicCard)
	//self.m_GameLogic.SetPiZiCard(self.PiZiCard)

	//写游戏日志
	self.WriteGameRecord()

	_userItem := self.GetUserItemByChair(self.BankerUser)
	_userItem.Ctx.DispatchCard(self.SendCardData)

	//设置变量
	self.ProvideCard = 0
	self.ProvideUser = public.INVALID_CHAIR
	self.CurrentUser = self.BankerUser //供应用户
	self.LastSendCardUser = self.BankerUser

	//动作分析,只分析庄家的牌，只有庄家初始才可以杠
	//杠牌判断
	var GangCardResult public.TagGangCardResult
	_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, nil, 0, &GangCardResult)

	//胡牌判断
	wChiHuRight := uint16(0)
	_userItem.Ctx.DeleteCard(self.SendCardData)
	_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:], _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, self.SendCardData, wChiHuRight, &_userItem.Ctx.ChiHuResult)
	_userItem.Ctx.DispatchCard(self.SendCardData)

	//构造数据,发送开始信息
	var GameStart public.Msg_S_GameStart
	GameStart.SiceCount = self.SiceCount
	GameStart.BankerUser = self.BankerUser
	GameStart.CurrentUser = self.CurrentUser
	GameStart.MagicCard = self.PiZiCard
	self.LockTimeOut(self.BankerUser)
	GameStart.Overtime = self.LimitTime
	GameStart.LeftCardCount = self.LeftCardCount

	GameStart.CardLeft.MaxCount = self.RepertoryCardArray.MaxCount
	GameStart.CardLeft.Seat = int(self.RepertoryCardArray.Seat)
	GameStart.CardLeft.Kaikou = self.RepertoryCardArray.Kaikou

	//记录癞子牌
	self.ReplayRecord.PiziCard = self.PiZiCard
	//记录发完牌后剩牌数量
	self.ReplayRecord.LeftCardCount = self.LeftCardCount

	//玩家番数
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		GameStart.PlayerFan[i] = self.GetUserFlowerCardFan(uint16(i))
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
		_, self.ReplayRecord.RecordHandCard[i] = self.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, self.ReplayRecord.RecordHandCard[i])
		//记录玩家初始分
		UserItem := self.GetUserItem(i)
		if UserItem != nil {
			//TODO 玩家分数设置
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

	for self.BrushCardByWind() {
	}
}

// ! 初始化游戏
func (self *Game_mj_jl_jlmj) OnInit(table base.TableBase) {
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
	self.Unmarsha(table.GetTableInfo().GameInfo)
	self.SetOfflineRoomTime(30 * 60)
	if table.GetTableInfo().GameInfo == "" && self.GameTable.GetTableInfo().JoinType == constant.NoCheat {
		self.SetOfflineRoomTime(60)
	}
	table.GetTableInfo().GameInfo = ""

	self.ParseRule(self.GetTableInfo().Config.GameConfig)
	self.m_GameLogic.Rule = self.Rule
	self.m_GameLogic.HuType = self.HuType
}

// ! 发送消息
func (self *Game_mj_jl_jlmj) OnMsg(msg *base.TableMsg) bool {

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
		var _msg Msg_C_OperateCard_JLMJ
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return self.OnUserOperateCard(&_msg)
		}
	case constant.MsgTypeGameGoOnNextGame: //下一局
		var _msg public.Msg_C_GoOnNextGame
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.OnUserClientNextGame(&_msg)
		}
	case constant.MsgTypeGameXiapao: //选漂
		var _msg public.Msg_C_Xiapao
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.OnUserClientXiaPao(&_msg)
		}
	case constant.MsgCommonToGameContinue:
		opt, ok := msg.V.(*public.TagSendCardInfo)
		if ok {
			self.DispatchCardData(opt.CurrentUser, opt.GangFlower)
		} else {
			self.OnWriteGameRecord(public.INVALID_CHAIR, "common to game 断言失败。")
		}
	case constant.MsgTypeGameTrustee: // 托管
		var _msg public.Msg_S_DG_Trustee
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			self.OnUserTustee(&_msg)
		}
	default:
		//self.GameCommon.OnMsg(msg)
	}
	return true
}

// ! 下一局
func (self *Game_mj_jl_jlmj) OnUserClientNextGame(msg *public.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(self.CurCompleteCount) >= self.Rule.JuShu || self.GetGameStatus() != public.GS_MJ_END {
		return true
	}

	// 记录游戏开始时间
	self.GameCommon.GameBeginTime = time.Now()

	nChiarID := self.GetChairByUid(msg.Id)

	self.SendTableMsg(constant.MsgTypeGameGoOnNextGame, *msg)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameGoOnNextGame, msg)
	self.SendUserStatus(int(nChiarID), public.US_READY)

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
			if !item.CheckTRUST() {
				break
			} else {
				item.UserReady = true
			}
		}
		if i == self.GetPlayerCount()-1 {
			// 复位桌子
			self.RepositTable()
			// 有跑就发送跑的设置
			//self.CurCompleteCount++
			//self.GetTable().SetBegin(true)
			self.OnGameStart()
		}
	}
	return true
}

// ! 清除吃胡记录
func (self *Game_mj_jl_jlmj) initChiHuResult() {
	for _, v := range self.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// ! 清除单个玩家记录
func (self *Game_mj_jl_jlmj) ClearChiHuResultByUser(wCurrUser uint16) {
	for _, v := range self.PlayerInfo {
		if v.GetChairID() == wCurrUser {
			v.Ctx.InitChiHuResult()
			break
		}
	}
}

// ! 反向清除单个玩家记录
func (self *Game_mj_jl_jlmj) ClearChiHuResultByUserReverse(wCurrUser uint16) {
	for _, v := range self.PlayerInfo {
		if v.GetChairID() != wCurrUser {
			v.Ctx.InitChiHuResult()
		}
	}
}

// ! 用户操作牌
func (self *Game_mj_jl_jlmj) OnUserOperateCard(msg *Msg_C_OperateCard_JLMJ) bool {

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

	// 能胡牌没有胡需要过庄
	if (_userItem.Ctx.UserAction&public.WIK_CHI_HU) != 0 && msg.Code != public.WIK_CHI_HU {
		//别人打牌自己弃胡才算过庄
		if self.CurrentUser == public.INVALID_CHAIR {
			_userItem.Ctx.NeedGuoZhuang = true
		}
	}

	if msg.Code != public.WIK_NULL ||
		(msg.Code == public.WIK_NULL && self.CurrentUser != wChairID) {
		// 解锁用户超时操作
		self.UnLockTimeOut(wChairID)
	}

	//游戏记录
	if msg.Code == public.WIK_NULL {
		self.entryTrust(msg.ByClient, _userItem)
		self.Greate_OperateRecord(&(msg.Msg_C_OperateCard), _userItem)
	}

	// 回放中记录牌权操作
	self.addReplayOrder(wChairID, info.E_HandleCardRight, msg.Code)

	//被动动作,被动操作没有红中杠，赖子杠,不分析抢杠
	if self.CurrentUser == public.INVALID_CHAIR {
		self.OnUserOperateInvalidChair(msg, _userItem)
		if msg.Code != public.WIK_NULL {
			for self.BrushCardByWind() {
			}
		}
		return true
	}

	//主动动作，杠的是赖子，和暗杠
	if self.CurrentUser == wChairID {
		self.OnUserOperateByChair(msg, _userItem)
		return true
	}

	return false
}

// ! 被动动作，别人打的牌,我吃碰杠胡牌
func (self *Game_mj_jl_jlmj) OnUserOperateInvalidChair(msg *Msg_C_OperateCard_JLMJ, _userItem *modules.Player) bool {

	wTargetUser := self.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card

	//效验状态
	if _userItem.Ctx.Response {
		return false
	}
	if (cbOperateCode != public.WIK_NULL) && ((_userItem.Ctx.UserAction & byte(cbOperateCode)) == 0) {
		return false
	}
	if cbOperateCard != self.ProvideCard {
		return false
	}

	//变量定义
	cbTargetAction := byte(cbOperateCode)
	//构造结果
	var OperateResult public.Msg_S_OperateResult

	//设置变量
	_userItem.Ctx.SetOperate(cbOperateCard, byte(cbOperateCode))
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

	//可能有多人是最高等级，多人胡牌，先把已经选择胡的人装起来
	var wTargetCHS []uint16
	iTargetResponse := 0
	for _, v := range self.PlayerInfo {
		//获取动作 只有多胡才会出现多人能够请求一个这里修改,因为到这里吃碰不可能有多个玩家请求,所以去掉获取玩家的选择
		cbUserAction := v.Ctx.UserAction //self.m_cbUserAction[i]
		//动作判断 把请求是最高权限的用户放在一起(一炮多响)
		if self.m_GameLogic.GetUserActionRank(cbUserAction) == self.m_GameLogic.GetUserActionRank(cbTargetAction) {
			wTargetCHS = append(wTargetCHS, v.Seat)
			if v.Ctx.Response {
				iTargetResponse++
			}
		}
	}
	if len(wTargetCHS) > 1 {
		syslog.Logger().Debug(fmt.Sprintf("最高权限人：%v, cbTargetAction:%v", wTargetCHS, cbTargetAction))
	}
	//如果有2个以上的人要胡，其中一个人点了胡，另外的人也就胡了
	if cbTargetAction == public.WIK_CHI_HU {
		if iTargetResponse != len(wTargetCHS) && (_userItem.Ctx.PerformAction != public.WIK_CHI_HU) {
			syslog.Logger().Debug("都还未响应,最大玩家选的不是胡牌")
			return true
		}
	} else {
		if userItem := self.GetUserItemByChair(wTargetUser); userItem != nil {
			if userItem.Ctx.Response == false {
				syslog.Logger().Debug(fmt.Sprintf("玩家（%d）座位号（%d）请求（%d）是否响应（%t）", userItem.Uid, userItem.Seat, userItem.Ctx.PerformAction, userItem.Ctx.Response))
				return true
			}
		}
	}

	//变量定义
	cbTargetCard := _userItem.Ctx.OperateCard
	//出牌变量
	self.SendStatus = true
	if cbTargetAction != public.WIK_NULL {
		self.OutCardData = 0
		self.OutCardUser = public.INVALID_CHAIR

		if providItem := self.GetUserItemByChair(self.ProvideUser); providItem != nil {
			providItem.Ctx.Requiredcard(cbTargetCard)
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
				self.DispatchCardData(self.ResumeUser, false)
			} else {
				self.ChiHuCard = 0
				self.ProvideUser = public.INVALID_CHAIR
				self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
			}
			return true
		}
	} else if cbTargetAction == public.WIK_CHI_HU {
		////胡牌操作
		//for tempIndex := 0; tempIndex < self.GetPlayerCount(); tempIndex++ {
		//	wUser := uint16(self.GetNextSeat(self.ProvideUser + uint16(tempIndex)))
		//
		//	if _item := self.GetUserItemByChair(wUser); _item != nil {
		//		//找到的第一个离放炮的用户最近并且有胡牌操作的用户
		//		if _item.Ctx.UserAction&public.WIK_CHI_HU != 0 {
		//			wTargetUser = wUser
		//			_userItem = _item
		//			if _userItem.Ctx.OperateCard == 0 {
		//				_userItem.Ctx.SetOperateCard(self.ProvideCard)
		//			}
		//			break
		//		}
		//	}
		//}
		//
		////结束信息
		//self.ChiHuCard = cbTargetCard
		//self.ProvideUser = self.ProvideUser
		//
		////插入扑克
		//if _userItem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
		//	_userItem.Ctx.DispatchCard(self.ChiHuCard)
		//}
		//
		////清除别人胡牌的牌权
		//self.ClearChiHuResultByUserReverse(_userItem.GetChairID())
		//
		////游戏记录
		//recordStr := fmt.Sprintf("%s，胡牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		//self.OnWriteGameRecord(wTargetUser, recordStr)
		//
		////记录胡牌
		//self.addReplayOrder(wTargetUser, E_Hu, cbTargetCard)
		//
		//self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
		//
		//return true
		//结束信息
		self.ChiHuCard = cbTargetCard
		self.ProvideUser = self.ProvideUser
		self.OutCardData = 0

		for _, v := range wTargetCHS {
			wTargetUser = v
			if userItem := self.GetUserItemByChair(wTargetUser); userItem != nil {
				//普通胡牌，有人点炮
				if self.ChiHuCard != 0 {
					if userItem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
						userItem.Ctx.CardIndex[self.m_GameLogic.SwitchToCardIndex(self.ChiHuCard)]++
					}
				} else { //自摸的 自摸只有一个玩家，这里用大海的 正常的自摸没到这里，估计是杠开花的才会来 测试了一下没有来，估计这块代码可以去掉了
					// syslog.Logger().Debug(fmt.Sprintf("自摸玩家（%d）座位号（%d）记录座位号（%d），结算结果（%v）", userItem.Uid, userItem.Seat, wTargetUser, userItem.Ctx.ChiHuResult))
					if userItem.Ctx.UserAction != public.WIK_NULL {
						self.ProvideUser = uint16(wTargetUser)
					}
					if userItem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
						userItem.Ctx.DeleteCard(self.ChiHuCard)
					}
				}
				//游戏记录
				recordStr := fmt.Sprintf("%s，胡牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
				self.OnWriteGameRecord(wTargetUser, recordStr)
				//  syslog.Logger().Debug(fmt.Sprintf("发送给客户端的数据：座位号（%d）,(%s)", wTargetUser, recordStr))
				//记录胡牌
				self.addReplayOrder(wTargetUser, info.E_Hu, int(cbTargetCard))
				//syslog.Logger().Debug(fmt.Sprintf("玩家（%d）座位号（%d）结算前的胡牌状态：%v", userItem.Uid, userItem.Seat, userItem.Ctx.ChiHuResult))
			}
		}
		//结束游戏
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
		_userItem.Ctx.AddWeaveItemArray_Modify(wIndex, 1, _provideUser, cbTargetAction, cbTargetCard)

		//删除扑克
		switch cbTargetAction {
		case public.WIK_LEFT: //左吃操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_RIGHT: //中吃操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_CENTER: //右吃操作
			self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_PENG: //碰牌操作
			{
				var GangCardResult public.TagGangCardResult
				var cbHighAction byte

				//判断该玩家是否可以杠这张牌
				_userItem.Ctx.DispatchCard(cbTargetCard)
				cbHighAction |= self.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex,
					_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult)
				_userItem.Ctx.DeleteCard(cbTargetCard)
				// 玩家在碰的时候有杠的牌权但是没有杠,加入弃杠记录
				var isCurGangCard bool
				for _, card := range GangCardResult.CardData {
					if card == cbTargetCard {
						isCurGangCard = true
					}
				}
				if cbHighAction&public.WIK_GANG != 0 && isCurGangCard {
					self.m_GameLogic.AppendGiveUpGang(_userItem, cbTargetCard)
				}
				self.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
			}

		case public.WIK_GANG: //杠牌操作
			{
				//删除扑克
				_userItem.Ctx.ShowGangAction()
				if self.ProvideUser != public.INVALID_CHAIR {
					if _item := self.GetUserItemByChair(self.ProvideUser); _item != nil {
						_item.Ctx.DianGangAction()
					}
				}
				//mingGangScore := self.GetScoreOnGang(info.E_Gang)
				//payscore := mingGangScore * self.GetUserAddItemScore(_provideUser) * self.GetUserAddItemScore(_userItem.GetChairID())
				//OperateResult.ScoreOffset[_provideUser] -= payscore
				//OperateResult.ScoreOffset[_userItem.GetChairID()] += payscore
				cbRemoveCard := []byte{cbTargetCard, cbTargetCard, cbTargetCard}
				_userItem.Ctx.RemoveCards(&self.Rule, cbRemoveCard)

				//游戏记录
				recordStr := fmt.Sprintf("%s，杠牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
				self.OnWriteGameRecord(wTargetUser, recordStr)

				//self.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("被点杠加分%d\n", payscore))
				//self.OnWriteGameRecord(_provideUser, fmt.Sprintf("点杠减分%d\n", payscore))
				//记录杠牌
				self.addReplayOrder(wTargetUser, info.E_Gang, int(cbTargetCard))
			}
		}

		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = cbTargetCard
		OperateResult.OperateCode = cbTargetAction
		OperateResult.ProvideUser = self.ProvideUser
		OperateResult.Overtime = self.LimitTime
		if self.ProvideUser == public.INVALID_CHAIR {
			OperateResult.ProvideUser = wTargetUser
		}
		for i := 0; i < self.GetPlayerCount(); i++ {
			OperateResult.PlayerFan[i] = self.GetUserFlowerCardFan(uint16(i))

			// 最新数据发送给客户端
			_user := self.GetUserItemByChair(uint16(i))
			if _user == nil {
				syslog.Logger().Debug("空指针...杠牌分数同步")
				continue
			}
			OperateResult.LaiGangCount[i] = _user.Ctx.CurMagicOut
			// 记录分数
			if OperateResult.OperateCard != self.MagicCard { //飘癞子
				// 杠
				_user.Ctx.GangScore += OperateResult.ScoreOffset[i]
			}
			// 改变数据
			//self.OnUserScoreOffset(_user, OperateResult.ScoreOffset[i])
			// OperateResult.GameScore[i] = _user.UserScoreInfo.Score + _user.Ctx.StorageScore
			//self.m_GameLogic.lastGangScore[i] = OperateResult.ScoreOffset[i]
		}
		//操作次数记录
		if self.ProvideUser != public.INVALID_CHAIR {
			//有人点炮的情况下,增加操作用户的操作次数,并保存第三次供牌的用户
			_userItem.Ctx.AddThirdOperate(self.ProvideUser)
		}

		OperateResult.HaveGang[wTargetUser] = _userItem.Ctx.HaveGang

		//OperateResult.GameScore, OperateResult.GameVitamin = self.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)

		//发送消息
		self.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		self.SendTableLookonMsg(constant.MsgTypeGameOperateResult, OperateResult)

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

		//如果是吃碰操作，再判断目标用户是否还有杠牌动作动作判断
		if self.LeftCardCount > 0 {
			//杠牌判断
			var GangCardResult public.TagGangCardResult

			_item := self.GetUserItemByChair(self.CurrentUser)

			_item.Ctx.UserAction |= self.m_GameLogic.AnalyseGangCard(_item, _item.Ctx.CardIndex,
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

// ! 主动动作，自己摸牌然后暗杠、痞子杠、赖子杠、续杠、胡牌
func (self *Game_mj_jl_jlmj) OnUserOperateByChair(msg *Msg_C_OperateCard_JLMJ, _userItem *modules.Player) bool {
	wChairID := self.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card
	//效验操作
	if cbOperateCode == public.WIK_NULL {
		////如果是杠权选择弃，加入弃杠记录
		//if _userItem.Ctx.UserAction&public.WIK_GANG != 0 {
		//	self.m_GameLogic.AppendGiveUpGang(_userItem, self.SendCardData)
		//}
		return true //放弃
	}

	//扑克效验
	if (cbOperateCode != public.WIK_NULL) && (cbOperateCode != public.WIK_CHI_HU) && (self.m_GameLogic.IsValidCard(cbOperateCard) == false) {
		return false
	}

	//设置变量
	self.SendStatus = true
	_userItem.Ctx.UserAction = public.WIK_NULL
	_userItem.Ctx.PerformAction = public.WIK_NULL

	//构造结果,向客户端发送操作结果
	var OperateResult Msg_S_OperateResult_JLMJ
	//执行动作
	switch cbOperateCode {
	case public.WIK_GANG: //杠牌操作
		{
			//弃杠牌处理
			if self.m_GameLogic.IsGiveUpGang(_userItem, cbOperateCard) {
				return false
			}

			bAnGang := false

			//变量定义
			cbWeaveIndex := 0xFF

			cbCardIndex := self.m_GameLogic.SwitchToCardIndex(cbOperateCard)

			if cbOperateCard == self.MagicCard {
				//赖子杠
				if _userItem.Ctx.CardIndex[cbCardIndex] == 0 {
					return false
				}
				//构造数据
				var OutCard public.Msg_S_OutCard
				OutCard.User = int(wChairID)
				OutCard.Data = cbOperateCard

				//发送消息
				self.SendTableMsg(constant.MsgTypeGameOutCard, OutCard)
				//发送旁观数据
				self.SendTableLookonMsg(constant.MsgTypeGameOutCard, OutCard)

				self.OutCardUser = wChairID
				self.OutCardData = cbOperateCard

				//删除扑克
				if !_userItem.Ctx.OutCard(&self.Rule, cbOperateCard) {
					syslog.Logger().Debug("removecard failed")
					return false
				}

				//游戏记录
				recordStr := fmt.Sprintf("%s，杠：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				self.OnWriteGameRecord(wChairID, recordStr)

				//记录蓄杠牌
				if cbOperateCard == self.MagicCard {
					self.addReplayOrder(wChairID, info.E_Gang_LaiziGand, int(cbOperateCard))
					_userItem.Ctx.GangMagic()
				}
			} else if _userItem.Ctx.CardIndex[cbCardIndex] == 1 {
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
				recordStr := fmt.Sprintf("%s，蓄杠牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				self.OnWriteGameRecord(wChairID, recordStr)

				//记录蓄杠牌
				self.addReplayOrder(wChairID, info.E_Gang_XuGand, int(cbOperateCard))

				_userItem.Ctx.XuGangAction()
				//xuGangScore := self.GetScoreOnGang(info.E_Gang_XuGand)
				//
				//for i := 0; i < self.GetPlayerCount(); i++ {
				//	_item := self.GetUserItemByChair(uint16(i))
				//	if wChairID != _item.GetChairID() {
				//		payscore := xuGangScore * self.GetUserAddItemScore(_item.GetChairID()) * self.GetUserAddItemScore(wChairID)
				//		OperateResult.ScoreOffset[_item.GetChairID()] -= payscore
				//		OperateResult.ScoreOffset[wChairID] += payscore
				//
				//		self.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("被续杠减分%d\n", payscore))
				//		self.OnWriteGameRecord(wChairID, fmt.Sprintf("续杠加分%d\n", payscore))
				//	}
				//
				//}
				bAnGang = false
				//组合扑克
				//_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 1, wChairID, byte(cbOperateCode), cbOperateCard)
				_userItem.Ctx.AddWeaveItemArray_Modify(cbWeaveIndex, 1, wChairID, byte(cbOperateCode), cbOperateCard)

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
				_userItem.Ctx.AddWeaveItemArray_Modify(cbWeaveIndex, 0, wChairID, byte(cbOperateCode), cbOperateCard)

				_userItem.Ctx.HidGangAction()
				bAnGang = true

				//anGangScore := self.GetScoreOnGang(info.E_Gang_AnGang)
				//				//for i := 0; i < self.GetPlayerCount(); i++ {
				//				//	_item := self.GetUserItemByChair(uint16(i))
				//				//	if wChairID != _item.GetChairID() {
				//				//		payscore := anGangScore * self.GetUserAddItemScore(_item.GetChairID()) * self.GetUserAddItemScore(_userItem.GetChairID())
				//				//		OperateResult.ScoreOffset[_item.GetChairID()] -= payscore
				//				//		OperateResult.ScoreOffset[wChairID] += payscore
				//				//
				//				//		self.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("被暗杠减分%d\n", payscore))
				//				//		self.OnWriteGameRecord(wChairID, fmt.Sprintf("暗杠加分%d\n", payscore))
				//				//	}
				//				//}

				//游戏记录
				recordStr := fmt.Sprintf("%s，暗杠牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				self.OnWriteGameRecord(wChairID, recordStr)

				//记录暗杠牌
				self.addReplayOrder(wChairID, info.E_Gang_AnGang, int(cbOperateCard))

				//删除扑克
				_userItem.Ctx.CleanCard(cbCardIndex)
			}

			// 最新数据发送给客户端
			for i, l := 0, self.GetPlayerCount(); i < l; i++ {
				_user := self.GetUserItemByChair(uint16(i))
				if _user == nil {
					syslog.Logger().Debug("空指针...杠牌分数同步")
					continue
				}
				OperateResult.LaiGangCount[i] = _user.Ctx.CurMagicOut
				// 记录分数
				if OperateResult.OperateCard != self.MagicCard {
					//杠
					_user.Ctx.GangScore += OperateResult.ScoreOffset[i]
				}
				// 改变数据
				//self.OnUserScoreOffset(_user, OperateResult.ScoreOffset[i])
				// OperateResult.GameScore[i] = _user.UserScoreInfo.Score + _user.Ctx.StorageScore
				//self.m_GameLogic.lastGangScore[i] = OperateResult.ScoreOffset[i]
			}

			OperateResult.OperateUser = wChairID
			OperateResult.ProvideUser = wChairID
			OperateResult.HaveGang[wChairID] = _userItem.Ctx.HaveGang //是否杠过
			OperateResult.OperateCode = cbOperateCode
			OperateResult.OperateCard = cbOperateCard
			for i := 0; i < self.GetPlayerCount(); i++ {
				OperateResult.PlayerFan[i] = self.GetUserFlowerCardFan(uint16(i))
			}

			//OperateResult.GameScore, OperateResult.GameVitamin = self.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)

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
			return true
		}
	case public.WIK_CHI_HU: //吃胡操作,主动状态下没有抢杠的说法，有自摸胡牌，杠上开花胡牌
		{
			//普通胡牌
			self.ClearChiHuResultByUserReverse(_userItem.GetChairID())
			self.ProvideCard = self.SendCardData

			if self.ProvideCard != 0 {
				self.ProvideUser = wChairID
			}

			//游戏记录
			recordStr := fmt.Sprintf("%s，胡牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(self.ProvideCard, 1))
			self.OnWriteGameRecord(wChairID, recordStr)

			//记录胡牌
			self.addReplayOrder(wChairID, info.E_Hu, int(self.ProvideCard))

			//结束信息
			self.ChiHuCard = self.ProvideCard

			//结束游戏
			self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)

			return true
		}
	case public.WIK_BUHUA: //补花操作
		{
			cbCardIndex := self.m_GameLogic.SwitchToCardIndex(cbOperateCard)
			if self.m_GameLogic.IsFlowerCard(cbOperateCard) {

				if _userItem.Ctx.CardIndex[cbCardIndex] == 0 {
					return false
				}
				//花牌加番
				self.m_FlowerCardFan[wChairID]++

				var OutCard public.Msg_S_OutCard
				OutCard.User = int(wChairID)
				OutCard.Data = cbOperateCard
				OutCard.ByClient = msg.ByClient

				//发送消息
				self.SendTableMsg(constant.MsgTypeGameOutCard, OutCard)
				//发送旁观数据
				self.SendTableLookonMsg(constant.MsgTypeGameOutCard, OutCard)

				//构造数据
				OperateResult.OperateUser = wChairID
				OperateResult.ProvideUser = wChairID
				OperateResult.HaveGang[wChairID] = _userItem.Ctx.HaveGang //是否杠过
				OperateResult.OperateCode = cbOperateCode
				OperateResult.OperateCard = cbOperateCard
				for i := 0; i < self.GetPlayerCount(); i++ {
					OperateResult.PlayerFan[i] = self.GetUserFlowerCardFan(uint16(i))
				}
				OperateResult.GameScore, OperateResult.GameVitamin = self.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)
				//发送消息
				self.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
				//发送旁观数据
				self.SendTableLookonMsg(constant.MsgTypeGameOperateResult, OperateResult)

				self.OutCardUser = wChairID
				self.OutCardData = cbOperateCard
				////删除扑克
				_userItem.Ctx.DeleteCard(cbOperateCard)

				////删除扑克
				//if !_userItem.Ctx.OutCard(&self.Rule, cbOperateCard) {
				//	syslog.Logger().Debug("removecard failed")
				//	return false
				//}

				//记录补花牌
				if _userItem.CheckTRUST() {
					self.addReplayOrder(wChairID, info.E_BuHua_TG, int(cbOperateCard))
				} else {
					self.addReplayOrder(wChairID, info.E_BuHua, int(cbOperateCard))
				}

				//游戏记录
				recordStr := fmt.Sprintf("%s，补花牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				self.OnWriteGameRecord(wChairID, recordStr)

				if self.M_broadCastFlower {
					buhuaMsg := Msg_S_OperateBuHuaNotify{OperateUser: _userItem.GetChairID()}
					//发送消息
					self.SendTableMsg(constant.MsgTypeGameOperateBuHuaNotify, buhuaMsg)
					//发送旁观数据
					self.SendTableLookonMsg(constant.MsgTypeGameOperateBuHuaNotify, buhuaMsg)
				}

				//发送扑克
				if self.LeftCardCount > 0 {
					self.DispatchCardData(wChairID, false)
				} else {
					self.ChiHuCard = 0
					self.ProvideUser = public.INVALID_CHAIR
					self.OnEventGameEnd(self.ProvideUser, public.GER_NORMAL)
				}
			}
		}
	}
	return true
}

// ! 操作牌
func (self *Game_mj_jl_jlmj) operateCard(cbTargetAction byte, cbTargetCard byte, _userItem *modules.Player) {
	var cbRemoveCard []byte
	var wik_kind int

	//变量定义
	switch cbTargetAction {
	case public.WIK_LEFT: //上牌操作
		cbRemoveCard = []byte{cbTargetCard + 1, cbTargetCard + 2}
		wik_kind = info.E_Wik_Left

		//游戏记录
		recordStr := fmt.Sprintf("%s，左吃牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)

	case public.WIK_RIGHT:
		cbRemoveCard = []byte{cbTargetCard - 2, cbTargetCard - 1}
		wik_kind = info.E_Wik_Right

		//游戏记录
		recordStr := fmt.Sprintf("%s，右吃牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_CENTER:
		cbRemoveCard = []byte{cbTargetCard - 1, cbTargetCard + 1}
		wik_kind = info.E_Wik_Center

		//游戏记录
		recordStr := fmt.Sprintf("%s，中吃牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_PENG: //碰牌操作
		cbRemoveCard = []byte{cbTargetCard, cbTargetCard}
		wik_kind = info.E_Peng

		//游戏记录
		recordStr := fmt.Sprintf("%s，碰牌：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		self.OnWriteGameRecord(_userItem.Seat, recordStr)
	default:
	}

	//吃碰过手可以胡牌,不需要过庄
	_userItem.Ctx.NeedGuoZhuang = false

	//记录左吃
	self.addReplayOrder(_userItem.Seat, wik_kind, int(cbTargetCard))
	//删除扑克
	_userItem.Ctx.RemoveCards(&self.Rule, cbRemoveCard)

	self.LockTimeOut(_userItem.Seat)
}

// ! 用户出牌
func (self *Game_mj_jl_jlmj) OnUserOutCard(msg *public.Msg_C_OutCard) bool {
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

	//手上有花牌,必须优先打出花牌
	if self.m_GameLogic.GetFlowerCardCount(_userItem.Ctx.CardIndex[:]) > 0 && !self.m_GameLogic.IsFlowerCard(msg.CardData) {
		syslog.Logger().Errorln("玩家手上有花牌必须优先出花牌")
		return false
	}

	//玩家出牌后清掉所有玩家的胡牌信息
	self.initChiHuResult()

	if _userItem.CheckTRUST() {
		msg.ByClient = false
	}
	self.entryTrust(msg.ByClient, _userItem)
	var class byte = 0
	if _userItem.CheckTRUST() {
		class = 1
	}
	//产品需求 当有多个风牌时 由于是自动打出的，会在很短的时间内连续播放补花动画和音效，决定只第一个牌风牌才播放
	self.M_broadCastFlower = true
	if self.M_outcard == logic.CARD_HONGZHONG || self.M_outcard == logic.CARD_FACAI || self.M_outcard == logic.CARD_BAIBAN {
		self.M_broadCastFlower = false
	}
	self.M_outcard = msg.CardData
	//花牌处理
	if self.m_GameLogic.IsFlowerCard(msg.CardData) {
		// 放到皮籁杠区
		_userItem.Ctx.DiscardPiLai_ex(msg.CardData, class)
		_msg := self.Greate_Operatemsg(_userItem.Uid, msg.ByClient, public.WIK_BUHUA, msg.CardData)
		return self.OnUserOperateCard(_msg)
	} else {
		// 出牌丢进弃牌区
		_userItem.Ctx.Discard_ex(msg.CardData, class)
	}

	// 解锁用户超时操作
	self.UnLockTimeOut(wChairID)

	//删除扑克
	if !_userItem.Ctx.OutCard(&self.Rule, msg.CardData) {
		syslog.Logger().Debug("removecard failed")
		return false
	}

	//游戏记录
	//recordStr := fmt.Sprintf("%s，打出：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1))
	//self.OnWriteGameRecord(wChairID, recordStr)
	//

	//游戏记录
	handCards := self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1)
	recordStr := self.Greate_OutCardRecord(handCards, msg, _userItem)
	// 游戏记录
	self.OnWriteGameRecord(wChairID, recordStr)

	//设置变量
	self.SendStatus = true

	//出牌记录
	self.OutCardCount++
	////////////////出的是一般的牌或赖子牌////////////////
	self.OutCardUser = wChairID
	self.OutCardData = msg.CardData

	//构造数据
	var OutCard public.Msg_S_OutCard
	OutCard.User = int(wChairID)
	OutCard.Data = msg.CardData
	OutCard.ByClient = msg.ByClient
	OutCard.Overtime = time.Now().Unix() + public.GAME_OPERATION_TIME
	if self.Rule.Overtime_trust > 0 {
		OutCard.Overtime = time.Now().Unix() + int64(self.Rule.Overtime_trust)
	}
	//记录出牌
	if _userItem.CheckTRUST() {
		// 记录下是不是托管出的牌
		self.addReplayOrder(wChairID, info.E_OutCard_TG, int(msg.CardData))
	} else {
		self.addReplayOrder(wChairID, info.E_OutCard, int(msg.CardData))
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
	//如果当前用户自己 出了牌，不能自己对自己进行分析吃，碰杠
	bAroseAction := false
	if !self.m_GameLogic.IsFlowerCard(msg.CardData) {
		bAroseAction = self.EstimateUserRespond(wChairID, msg.CardData, public.EstimatKind_OutCard)
	}

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

// 托管下的自动补花
func (self *Game_mj_jl_jlmj) BrushCardByWind() bool {
	curUser := self.GetUserItemByChair(self.CurrentUser)
	if curUser == nil {
		return false
	}
	if !curUser.CheckTRUST() {
		return false
	}
	for k, v := range curUser.Ctx.CardIndex {
		cardData := self.m_GameLogic.SwitchToCardData(byte(k))
		if self.m_GameLogic.IsFlowerCardM(cardData) && v > 0 {
			self.OnUserOutCard(&public.Msg_C_OutCard{
				Id:       curUser.Uid,
				CardData: cardData,
				ByClient: true, // 伪装成客户端操作，方便托管处理
			})
			return true
		}
	}
	return false
}

// ! 超时自动出牌
func (self *Game_mj_jl_jlmj) OnAutoOperate(wChairID uint16, bBreakin bool) {

	if bBreakin == false {
		return
	}

	//if !self.Rule.NineSecondRoom {
	//self.UnLockTimeOut(wChairID)
	//return
	//}

	if self.GetGameStatus() == public.GS_MJ_FREE {
		self.UnLockTimeOut(wChairID)
		return
	}

	if self.GetGameStatus() != public.GS_MJ_PLAY {
		self.UnLockTimeOut(wChairID)
		return
	}
	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	// 处理下跑
	if self.PayPaoStatus {
		if !_userItem.Ctx.UserPaoReady {
			_msg := self.Greate_XiaPaomsg(_userItem.Uid, false,
				_userItem.Ctx.VecXiaPao.Status, _userItem.Seat == self.BankerUser)
			if !self.OnUserClientXiaPao(_msg) {
				self.OnWriteGameRecord(_userItem.Seat, "服务器自动选飘时，可能被客户端抢先了")
			}
		}
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
		if index >= 0 && index < public.MAX_INDEX {
			if 0 != _userItem.Ctx.CardIndex[index] {
				_msg := self.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
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

// 创建消息
func (self *Game_mj_jl_jlmj) Greate_XiaPaomsg(Id int64, byClient bool, status bool, isBanker bool) *public.Msg_C_Xiapao {
	_msg := new(public.Msg_C_Xiapao)
	_msg.Num = 0
	_msg.Status = status
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建操作牌消息
func (self *Game_mj_jl_jlmj) Greate_Operatemsg(Id int64, byClient bool, Code int, Card byte) *Msg_C_OperateCard_JLMJ {
	_msg := new(Msg_C_OperateCard_JLMJ)
	_msg.Card = Card
	_msg.Code = Code
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建出牌消息
func (self *Game_mj_jl_jlmj) Greate_OutCardmsg(Id int64, byClient bool, Card byte) *public.Msg_C_OutCard {
	_msg := new(public.Msg_C_OutCard)
	_msg.CardData = Card
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 派发扑克
func (self *Game_mj_jl_jlmj) DispatchCardData(wCurrentUser uint16, bGangFlower bool) bool {

	if self.IsPausing() {
		self.CurrentUser = public.INVALID_CHAIR
		self.SetSendCardOpt(public.TagSendCardInfo{
			CurrentUser: wCurrentUser,
			GangFlower:  bGangFlower,
		})
		return true
	}

	//状态效验
	if wCurrentUser == public.INVALID_CHAIR {
		return false
	}

	_userItem := self.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return false
	}

	//吃碰过手可以胡牌,不需要过庄
	_userItem.Ctx.NeedGuoZhuang = false

	//剩余牌校验
	if self.LeftCardCount <= 0 {
		return false
	}

	bEnjoinHu := true
	self.GangFlower = bGangFlower
	//发牌处理
	if self.SendStatus == true {
		//发送扑克
		self.LockTimeOut(wCurrentUser)
		self.SendCardCount++
		self.LeftCardCount--
		self.SendCardData = self.RepertoryCard[self.LeftCardCount]

		_userItem.Ctx.DispatchCard(self.SendCardData)

		self.SetLeftCardArray()

		//游戏记录
		recordStr := fmt.Sprintf("牌型%s，发来：%s", self.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), self.m_GameLogic.SwitchToCardNameByData(self.SendCardData, 1))
		self.OnWriteGameRecord(wCurrentUser, recordStr)

		//记录发牌
		self.addReplayOrder(wCurrentUser, info.E_SendCard, int(self.SendCardData))

		//设置变量
		self.ProvideUser = wCurrentUser
		self.ProvideCard = self.SendCardData
		//给用户发牌后，判断用户是否可以杠牌
		if self.LeftCardCount > 0 {
			var GangCardResult public.TagGangCardResult
			_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex,
				_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult)
		}

		// 判断是否胡牌
		self.initChiHuResult()
		self.CheckHu(wCurrentUser, wCurrentUser, 0, bGangFlower, false)

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
	SendCard.EnjoinHu = bEnjoinHu
	//SendCard.Overtime = time.Now().Unix() + public.GAME_OPERATION_TIME
	self.LastSendCardUser = wCurrentUser
	// 设置开始超时操作

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
	recordStr := fmt.Sprintf("发送牌权：%x", _userItem.Ctx.UserAction)
	self.OnWriteGameRecord(wCurrentUser, recordStr)

	// 回放记录中记录牌权显示
	if _userItem.Ctx.UserAction > 0 {
		self.addReplayOrder(wCurrentUser, info.E_SendCardRight, int(_userItem.Ctx.UserAction))
	}

	for self.BrushCardByWind() {
	}

	return true
}

// ! 响应判断
func (self *Game_mj_jl_jlmj) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int) bool {
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
			//吃碰判断
			if self.LeftCardCount > 0 {
				//碰牌判断
				_item.Ctx.UserAction |= self.m_GameLogic.EstimatePengCard(_item.Ctx.CardIndex, cbCenterCard)

				if self.Rule.KeChi {
					//吃牌判断--吃牌就判断出牌的人的下一个人
					wEatUser := self.GetNextSeat(wCenterUser)
					if wEatUser == i {
						_item.Ctx.UserAction |= self.m_GameLogic.EstimateEatCard(_item.Ctx.CardIndex, cbCenterCard)
					}
				}

				//杠牌判断
				_item.Ctx.UserAction |= self.m_GameLogic.EstimateGangCard(_item.Ctx.CardIndex, cbCenterCard)
			}
		}

		if self.Rule.HuType != HuFa_Type_Zimo {
			self.CheckHu(uint16(i), wCenterUser, cbCenterCard, false, false)
		}

		//if (_item.Ctx.UserAction & public.WIK_CHI_HU) > 0 {
		//	tingCount := self.m_GameLogic.AnalyseTingCardCount(_item.Ctx.CardIndex[:], _item.Ctx.WeaveItemArray[:], _item.Ctx.WeaveItemCount, 0)
		//	if tingCount <= 20 {
		//		// 非见字胡可以接炮
		//	} else {
		//		// 见字胡不能接炮
		//		_item.Ctx.UserAction ^= public.WIK_CHI_HU
		//
		//		self.SendGameNotificationMessage(uint16(i), "见字胡不能炮胡")
		//	}
		//}

		//结果判断
		if _item.Ctx.UserAction != public.WIK_NULL {
			bAroseAction = true
		}
	}

	//结果处理，标志为真说明上一个用户所出的牌有人可以所以要发送一个提示信息的框
	if bAroseAction == true {
		//设置变量
		self.ProvideUser = uint16(wCenterUser)
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
func (self *Game_mj_jl_jlmj) SendOperateNotify() bool {
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
			OperateNotify.VecGangCard = public.HF_BytesToInts(v.Ctx.VecGangCard)

			//发送数据
			//抢的牌权需要发送给所有玩家，因为其他玩家需要知道轮到谁抢暗杠了
			if v.Ctx.UserAction == public.WIK_QIANG {
				OperateNotify.ActionCard = byte(v.Seat)
				self.SendTableMsg(constant.MsgTypeGameOperateNotify, OperateNotify)
			} else {
				self.SendPersonMsg(constant.MsgTypeGameOperateNotify, OperateNotify, v.Seat)
			}

			// 游戏记录
			recrodStr := fmt.Sprintf("发送牌权：%x", v.Ctx.UserAction)
			self.OnWriteGameRecord(v.Seat, recrodStr)

			// 回放记录中记录牌权显示
			self.addReplayOrder(v.Seat, info.E_SendCardRight, int(v.Ctx.UserAction))
		}
	}

	return true
}

// ! 增加回放操作记录
func (self *Game_mj_jl_jlmj) addReplayOrder(chairId uint16, operation int, card int) {
	var order meta.Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	self.ReplayRecord.VecOrder = append(self.ReplayRecord.VecOrder, order)
}

// ! 检查是否能胡
func (self *Game_mj_jl_jlmj) CheckHu(wCurrentUser uint16, wProvideUser uint16, cbCurrentCard byte, bGangFlower bool, bQiangGang bool) byte {
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

	//给用户发牌后，胡牌判断
	_userItem.Ctx.UserAction |= self.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:],
		_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, cbCurrentCard, wChiHuRight, &_userItem.Ctx.ChiHuResult)

	//if (_userItem.Ctx.UserAction & public.WIK_CHI_HU) != 0 {
	//	self.CheckFan(wCurrentUser, wProvideUser, &_userItem.Ctx.ChiHuResult)
	//}

	//需要过庄才能胡
	if _userItem.Ctx.NeedGuoZhuang {
		if (_userItem.Ctx.UserAction & public.WIK_CHI_HU) != 0 {
			_userItem.Ctx.UserAction ^= public.WIK_CHI_HU
			self.ClearChiHuResultByUser(wCurrentUser)
			self.SendGameNotificationMessage(_userItem.GetChairID(), "过庄不能胡")
		}
		return public.WIK_NULL
	}

	return _userItem.Ctx.UserAction
}

func (self *Game_mj_jl_jlmj) JlmjSuanFen(GameEnd *public.Msg_S_GameEnd) {
	nWinnerCnt := 0                           //胡牌的人数
	var nWinner uint16 = public.INVALID_CHAIR //换庄者

	//找出胡牌玩家
	for i := 0; i < self.GetPlayerCount(); i++ {
		checkitem := self.GetUserItemByChair(uint16(i))
		if checkitem == nil {
			continue
		}

		if checkitem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
			nWinner = uint16(i)
			nWinnerCnt++
			GameEnd.WWinner[nWinner] = true
			GameEnd.WChiHuKind[nWinner] = checkitem.Ctx.ChiHuResult.ChiHuKind
			fmt.Printf("算番 玩家(%d), ChiHuKind:(%x), UserAction:(%x),ChiHuRight:(%x)\n", i, checkitem.Ctx.ChiHuResult.ChiHuKind, checkitem.Ctx.UserAction, checkitem.Ctx.ChiHuResult.ChiHuRight)
		}
	}

	if nWinnerCnt > 0 {

		self.GetPlayerFan()
		if nWinnerCnt == 1 {
			//一人胡牌,分自摸 点炮
			if self.ProvideUser == nWinner {
				//自摸
				checkitem := self.GetUserItemByChair(nWinner)
				if checkitem == nil {
					return
				}
				checkitem.Ctx.HuBySelf()
				for i := 0; i < self.GetPlayerCount(); i++ {
					if i == int(self.ProvideUser) {
						continue
					}
					//自摸2分
					payScore := 2 * self.FanScore[i].FanNum[i] * self.FanScore[i].FanNum[nWinner] * self.Rule.DiFen
					GameEnd.GameScore[i] -= payScore
					GameEnd.GameScore[nWinner] += payScore
				}
			} else {
				//点炮
				checkitem := self.GetUserItemByChair(nWinner)
				if checkitem == nil {
					return
				}
				//放炮1分
				payScore := 1 * self.FanScore[nWinner].FanNum[self.ProvideUser] * self.FanScore[nWinner].FanNum[nWinner] * self.Rule.DiFen
				GameEnd.GameScore[self.ProvideUser] -= payScore
				GameEnd.GameScore[nWinner] += payScore
				//吃胡统计
				checkitem.Ctx.ChiHuUserCount++
				//点炮统计
				if provideitem := self.GetUserItemByChair(self.ProvideUser); provideitem != nil {
					provideitem.Ctx.ProvideCard()
				}
			}
		} else {
			//多人胡牌,只能是一炮多响
			for winner := 0; winner < self.GetPlayerCount(); winner++ {
				checkitem := self.GetUserItemByChair(uint16(winner))
				if checkitem == nil {
					return
				}
				if checkitem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
					for j := 0; j < self.GetPlayerCount(); j++ {
						if j == winner {
							//放炮1分
							payScore := 1 * self.FanScore[j].FanNum[j] * self.FanScore[j].FanNum[self.ProvideUser] * self.Rule.DiFen
							GameEnd.GameScore[self.ProvideUser] -= payScore
							GameEnd.GameScore[j] += payScore
						}
					}
					//吃胡统计
					checkitem.Ctx.ChiHuUserCount++
				}
			}
			//点炮统计
			if provideitem := self.GetUserItemByChair(self.ProvideUser); provideitem != nil {
				provideitem.Ctx.ProvideCard()
			}
		}
	} else {
		//流局
		syslog.Logger().Debug("荒庄")
		// 慌庄所有人都是0分
		for i := 0; i < self.GetPlayerCount(); i++ {
			GameEnd.GameScore[i] = 0
		}
		//流局处理
		GameEnd.ChiHuCard = 0
		GameEnd.ChiHuUserCount = 0
		GameEnd.Winner = public.INVALID_CHAIR
		self.HaveHuangZhuang = true
	}
	//记录下胡分
	for i := 0; i < self.GetPlayerCount(); i++ {
		GameEnd.HuFen[i] = GameEnd.GameScore[i]
	}
}

// ! 得到玩家番数
func (self *Game_mj_jl_jlmj) GetPlayerFan() {
	for i := 0; i < self.GetProperPNum(); i++ {
		for j := 0; j < self.GetProperPNum(); j++ {
			self.FanScore[i].FanNum[j] = self.GetUserAddItemScore(uint16(j))
		}
		fmt.Println(fmt.Sprintf("玩家%d, 加分番为%d", i, self.FanScore[i].FanNum[i]))
	}
}

// ! 单局结算
func (self *Game_mj_jl_jlmj) OnGameOver(wChairID uint16, cbReason byte) bool {
	self.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 游戏结束, 流局结束，统计积分
func (self *Game_mj_jl_jlmj) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
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

// 计算杠分并写入对应数据结构
func (self *Game_mj_jl_jlmj) CalcGangScore() []int {
	scores := make([]int, self.GetPlayerCount())
	for _, player := range self.PlayerInfo {
		seat := player.Seat
		for i := byte(0); i < player.Ctx.WeaveItemCount; i++ {
			weaveItem := player.Ctx.WeaveItemArray[i]
			if weaveItem.WeaveKind == public.WIK_GANG {
				if seat == weaveItem.ProvideUser {
					var gangScore int
					if weaveItem.ProvideUser1 == public.INVALID_CHAIR {
						// 暗杠
						gangScore = self.GetScoreOnGang(info.E_Gang_AnGang)
					} else {
						// 补杠
						gangScore = self.GetScoreOnGang(info.E_Gang_XuGand)
					}
					for tmpSeat := 0; tmpSeat < self.GetPlayerCount(); tmpSeat++ {
						if uint16(tmpSeat) != seat {
							payScore := gangScore * self.GetUserAddItemScore(seat) *
								self.GetUserAddItemScore(uint16(tmpSeat))
							scores[seat] += payScore
							scores[tmpSeat] -= payScore
						}
					}
				} else {
					// 明杠
					mingGangScore := self.GetScoreOnGang(info.E_Gang)
					payScore := mingGangScore * self.GetUserAddItemScore(seat) *
						self.GetUserAddItemScore(weaveItem.ProvideUser)
					scores[seat] += payScore
					scores[weaveItem.ProvideUser] -= payScore
				}
			}
		}
	}

	// 补花4个红中等于暗杠
	if self.Rule.BuHuaType != BuHua_Type_NULL {
		var checkCardIdx []byte
		if self.Rule.BuHuaType == BuHua_Type_HongZhong {
			checkCardIdx = append(checkCardIdx, logic.CARD_HONGZHONG)
		} else if self.Rule.BuHuaType == BuHua_Type_ZhongFaBai {
			checkCardIdx = append(checkCardIdx, logic.CARD_HONGZHONG)
			checkCardIdx = append(checkCardIdx, logic.CARD_FACAI)
			checkCardIdx = append(checkCardIdx, logic.CARD_BAIBAN)
		}

		for _, player := range self.PlayerInfo {
			seat := player.Seat
			handCardsIdx := common.CardsToCardIndex(player.Ctx.Pilaicardcard)
			for _, cardIdx := range checkCardIdx {
				if handCardsIdx[self.m_GameLogic.SwitchToCardIndex(cardIdx)] == 4 {
					for tmpSeat := 0; tmpSeat < self.GetPlayerCount(); tmpSeat++ {
						if uint16(tmpSeat) != seat {
							payScore := self.GetScoreOnGang(info.E_Gang_AnGang) * self.GetUserAddItemScore(seat) *
								self.GetUserAddItemScore(uint16(tmpSeat))
							scores[seat] += payScore
							scores[tmpSeat] -= payScore
						}
					}
				}
			}
		}
	}

	for _, player := range self.PlayerInfo {
		self.OnUserScoreOffset(player, scores[player.Seat])
	}

	return scores
}

// ! 结束，结束游戏
func (self *Game_mj_jl_jlmj) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	//设置游戏结束状态
	self.SetGameStatus(public.GS_MJ_END)

	//定义变量
	var GameEnd public.Msg_S_GameEnd
	var huDetail modules.TagHuCostDetail

	GameEnd.LastSendCardUser = self.LastSendCardUser
	GameEnd.EndStatus = cbReason

	GameEnd.MagicCard = self.MagicCard

	GameEnd.ProvideUser = wChairID
	GameEnd.ChiHuCard = self.ChiHuCard
	GameEnd.ChiHuUserCount = 1
	for i := 0; i < int(self.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = self.RepertoryCard[i]
	}

	self.JlmjSuanFen(&GameEnd)
	self.CalcGangScore()

	//可以有多个赢家，循环判断找出赢家
	nWinnerCount := 0
	for _, v := range self.PlayerInfo {
		if GameEnd.WWinner[v.Seat] {
			nWinnerCount++
			if v.Seat == self.ProvideUser {
				huDetail.Private(v.Seat, modules.TagHuSelf, 1, modules.DetailTypeFirst)
			} else {
				huDetail.Private(v.Seat, modules.TagJiePao, 1, modules.DetailTypeFirst)
			}
		} else {
			if v.Seat == self.ProvideUser {
				huDetail.Private(v.Seat, modules.TagDianPao, 1, modules.DetailTypeFirst)
			}
		}
	}
	//胡牌玩家
	GameEnd.IsQuit = false
	GameEnd.TheOrder = self.CurCompleteCount
	GameEnd.HuangZhuang = self.HaveHuangZhuang

	//计算各玩家明杠，暗杠，赖子
	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		//保存四家明杠的次数
		GameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang)
		//保存四家暗杠的次数
		GameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		//保存四家续杠的次数
		GameEnd.XuGangCount[i] = uint16(_userItem.Ctx.XuGang)
		//保存四家点杠的次数
		GameEnd.DianGangCount[i] = uint16(_userItem.Ctx.DianGang)
		//保存四家补花的次数

		//huDetail.Private(_userItem.Seat, modules.TagBuHua, self.m_FlowerCardFan[i], modules.DetailTypeCost) //补花
		//if self.Rule.ChuZeng {
		//	huDetail.Private(_userItem.Seat, modules.TagChuZeng, _userItem.Ctx.VecXiaPao.Num, modules.DetailTypeCost) //出增
		//}
		//huDetail.Private(_userItem.Seat, modules.TagWeiZeng, self.Rule.WeiZeng, modules.DetailTypeCost) //围增
		//if self.Rule.KengZhuang && _userItem.Seat == self.BankerUser {
		//	huDetail.Private(_userItem.Seat, modules.TagKengZhuang, 1, modules.DetailTypeFirst) //庄家显示坑庄
		//}

		//huDetail.Private(_userItem.Seat, modules.TagShowGan, int(_userItem.Ctx.ShowGang), modules.DetailTypeCost) //明杠
		//huDetail.Private(_userItem.Seat, modules.TagXuGan, int(_userItem.Ctx.XuGang), modules.DetailTypeCost)     //蓄杠
		//huDetail.Private(_userItem.Seat, modules.TagAnGan, int(_userItem.Ctx.HidGang), modules.DetailTypeCost)    //暗杠
		//huDetail.Private(_userItem.Seat, modules.TagDianGan, int(_userItem.Ctx.DianGang), modules.DetailTypeCost) //点杠
	}

	//if GameEnd.Winner != public.INVALID_CHAIR {
	//	if winnerItem := self.GetUserItemByChair(GameEnd.Winner); winnerItem != nil {
	//		if winnerItem.Ctx.ChiHuResult.ChiHuKind&public.CHK_GANG_SHANG_KAI_HUA != 0 {
	//			GameEnd.BigHuKind = public.GameBigHuKindGSK
	//			huDetail.Private(winnerItem.Seat, TagHuGangKai, 1, DetailTypeFirst)
	//		}
	//
	//		if winnerItem.Ctx.ChiHuResult.ChiHuKind&public.CHK_PING_HU_NOMAGIC != 0 {
	//			huDetail.Private(winnerItem.Seat, TagHuHard, 1, DetailTypeFirst)
	//		}
	//
	//		if GameEnd.Winner == self.ProvideUser {
	//			huDetail.Private(winnerItem.Seat, TagHuSelf, 1, DetailTypeFirst)
	//		}
	//	}
	//}

	//判断调整分
	//结算分
	Result := [meta.MAX_PLAYER]int{}
	for i := 0; i < self.GetPlayerCount(); i++ {
		_item := self.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		if nWinnerCount == 0 {
			GameEnd.MaxFSCount[i] = 0
		}

		if GameEnd.WWinner[i] {
			GameEnd.MaxFSCount[i] = uint16(self.FanScore[i].FanNum[i])
		}
		GameEnd.GameScore[i] += _item.Ctx.StorageScore //游戏过程中的杠分
		GameEnd.GangFen[i] += _item.Ctx.StorageScore   //记录杠分
		Result[i] = GameEnd.GameScore[i]
		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore //用这个结构来保存游戏过程中的杠分,传递给客户端
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount

		GameEnd.StrEnd[i] = huDetail.GetSeatString(uint16(i))
		if GameEnd.MaxFSCount[i] > uint16(_item.Ctx.MaxFanUserCount) {
			_item.Ctx.SetMaxFan(int(GameEnd.MaxFSCount[i])) //统计最大番数
		}
		if int(GameEnd.GameScore[i]) > _item.Ctx.MaxScoreUserCount {
			_item.Ctx.SetMaxScore(int(GameEnd.GameScore[i])) //统计最大分数
		}

		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = self.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
		fmt.Println(fmt.Sprintf("%v", GameEnd.CardData[i]))
	}

	GameEnd.UserScore, GameEnd.UserVitamin = self.OnSettle(Result, constant.EventSettleGameOver)

	//发送信息
	self.VecGameEnd = append(self.VecGameEnd, GameEnd) //保存，用于汇总计算
	self.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	self.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	self.SaveGameData()

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

	//第一局房主为庄，之后胡牌玩家为庄,流局这当局庄家连庄,一炮多响的情况下，点炮玩家坐庄
	var nextBanker int = -1
	if nWinnerCount == 1 {
		//一个赢家,自摸或者点炮
		for i := 0; i < self.GetPlayerCount(); i++ {
			if GameEnd.WWinner[i] {
				nextBanker = i
				break
			}
		}

		self.BankerUser = uint16(nextBanker)
	} else if nWinnerCount > 1 {
		//一炮多响,点炮者坐庄
		self.BankerUser = self.ProvideUser
	} else if nWinnerCount == 0 {
		//荒庄庄家不变
	}

	self.OnGameEnd()
	self.accountManage(self.Rule.Overtime_dismiss, self.Rule.Overtime_trust, 15)
	self.RepositTable() // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	return true

}

// ! 强退，结束游戏
func (self *Game_mj_jl_jlmj) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	//变量定义
	var GameEnd public.Msg_S_GameEnd
	var huDetail modules.TagHuCostDetail

	//设置变量
	GameEnd.EndStatus = cbReason
	GameEnd.MagicCard = self.MagicCard
	GameEnd.ProvideUser = public.INVALID_CHAIR
	GameEnd.Winner = public.INVALID_CHAIR
	GameEnd.Contractor = public.INVALID_CHAIR
	GameEnd.IsQuit = true
	GameEnd.HuangZhuang = self.HaveHuangZhuang

	for i := 0; i < int(self.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = self.RepertoryCard[i]
	}

	self.CalcGangScore()

	//计算各玩家明杠，暗杠，赖子
	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = self.m_GameLogic.SwitchToCardData2(_userItem.Ctx.CardIndex, GameEnd.CardData[i])

		GameEnd.GameScore[i] += _userItem.Ctx.QiangScore
		//GameEnd.GameScore[i] += _userItem.Ctx.StorageScore      //游戏过程中的杠分
		GameEnd.GameAdjustScore[i] = _userItem.Ctx.StorageScore //用这个结构来保存游戏过程中的杠分,传递给客户端
		//玩家番数
		GameEnd.MaxFSCount[i] = uint16(self.GetUserFlowerCardFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _userItem.Ctx.QiangFailCount

		GameEnd.WeaveItemArray[i] = _userItem.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _userItem.Ctx.WeaveItemCount

		//保存四家明杠的次数
		GameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang)
		//保存四家暗杠的次数
		GameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		//保存四家续杠的次数
		GameEnd.XuGangCount[i] = uint16(_userItem.Ctx.XuGang)
		//保存四家赖子杠的次数
		GameEnd.MagicGangCount[i] = uint16(_userItem.Ctx.MagicCardGang)
		//保存四家点杠的次数
		GameEnd.DianGangCount[i] = uint16(_userItem.Ctx.DianGang)

		//huDetail.Private(_userItem.Seat, modules.TagShowGan, int(_userItem.Ctx.ShowGang), modules.DetailTypeADD)  //明杠
		//huDetail.Private(_userItem.Seat, modules.TagXuGan, int(_userItem.Ctx.XuGang), modules.DetailTypeADD)      //蓄杠
		//huDetail.Private(_userItem.Seat, modules.TagAnGan, int(_userItem.Ctx.HidGang), modules.DetailTypeADD)     //暗杠
		//huDetail.Private(_userItem.Seat, modules.TagMagic, int(_userItem.Ctx.MagicCardGang), modules.DetailTypeF) //癞子杠
		//huDetail.Private(_userItem.Seat, modules.TagDianGan, int(_userItem.Ctx.DianGang), modules.DetailTypeADD)  //点杠

		GameEnd.StrEnd[i] = huDetail.GetSeatString(uint16(i))
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
		// 数据库写出牌记录
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
func (self *Game_mj_jl_jlmj) OnGameEndDissmiss(wChairID uint16, cbReason byte, cbSubReason byte) bool {
	//变量定义
	var huDetail modules.TagHuCostDetail
	var GameEnd public.Msg_S_GameEnd
	GameEnd.LastSendCardUser = self.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.ProvideUser = public.INVALID_CHAIR
	GameEnd.Winner = public.INVALID_CHAIR
	GameEnd.Contractor = public.INVALID_CHAIR
	GameEnd.HuangZhuang = self.HaveHuangZhuang
	for i := 0; i < int(self.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = self.RepertoryCard[i]
	}

	GameEnd.MagicCard = self.MagicCard

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
	self.CalcGangScore()

	//抢杠分数，解散了也要结算
	for i := 0; i < self.GetPlayerCount(); i++ {
		_userItem := self.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = self.m_GameLogic.SwitchToCardData2(_userItem.Ctx.CardIndex, GameEnd.CardData[i])

		GameEnd.GameScore[i] += _userItem.Ctx.QiangScore
		//GameEnd.GameScore[i] += _userItem.Ctx.StorageScore      //游戏过程中的杠分
		GameEnd.GameAdjustScore[i] = _userItem.Ctx.StorageScore //用这个结构来保存游戏过程中的杠分,传递给客户端
		//玩家番数
		GameEnd.MaxFSCount[i] = uint16(self.GetUserFlowerCardFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _userItem.Ctx.QiangFailCount

		GameEnd.WeaveItemArray[i] = _userItem.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _userItem.Ctx.WeaveItemCount

		//保存四家明杠的次数
		GameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang)
		//保存四家暗杠的次数
		GameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		//保存四家续杠的次数
		GameEnd.XuGangCount[i] = uint16(_userItem.Ctx.XuGang)
		//保存四家点杠的次数
		GameEnd.DianGangCount[i] = uint16(_userItem.Ctx.DianGang)
		//保存四家补花的次数

		//	huDetail.Private(_userItem.Seat, modules.TagBuHua, self.m_FlowerCardFan[i], modules.DetailTypeCost) //补花
		//	if self.Rule.ChuZeng {
		//		huDetail.Private(_userItem.Seat, modules.TagChuZeng, _userItem.Ctx.VecXiaPao.Num, modules.DetailTypeCost) //出增
		//	}
		//	huDetail.Private(_userItem.Seat, modules.TagWeiZeng, self.Rule.WeiZeng, modules.DetailTypeCost) //围增
		//	if self.Rule.KengZhuang && _userItem.Seat == self.BankerUser {
		//		huDetail.Private(_userItem.Seat, modules.TagKengZhuang, 1, modules.DetailTypeFirst) //庄家显示坑庄
		//	}

		//	huDetail.Private(_userItem.Seat, modules.TagShowGan, int(_userItem.Ctx.ShowGang), modules.DetailTypeCost) //明杠
		//	huDetail.Private(_userItem.Seat, modules.TagXuGan, int(_userItem.Ctx.XuGang), modules.DetailTypeCost)     //蓄杠
		//	huDetail.Private(_userItem.Seat, modules.TagAnGan, int(_userItem.Ctx.HidGang), modules.DetailTypeCost)    //暗杠
		//	huDetail.Private(_userItem.Seat, modules.TagDianGan, int(_userItem.Ctx.DianGang), modules.DetailTypeCost) //点杠

		GameEnd.StrEnd[i] = huDetail.GetSeatString(uint16(i))
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

		if self.GetGameStatus() != public.GS_MJ_FREE && !self.PayPaoStatus {
			//数据库写出牌记录
			self.TableWriteOutDate(int(self.CurCompleteCount), self.ReplayRecord)
			// 写完后清除数据
			self.ReplayRecord.Reset()

			//数据库写入单局结算
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
func (self *Game_mj_jl_jlmj) OnEnd() {
	if self.IsGameStarted() {
		self.OnGameOver(public.INVALID_CHAIR, public.GER_DISMISS)
	}
}

// ! 计算总发送总结算
func (self *Game_mj_jl_jlmj) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	// 给客户端发送总结算数据
	var balanceGame public.Msg_S_BALANCE_GAME
	balanceGame.Userid = self.Rule.FangZhuID
	balanceGame.CurTotalCount = self.CurCompleteCount //总盘数
	self.TimeEnd = time.Now().Unix()                  //大局结束时间
	balanceGame.TimeStart = self.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = self.TimeEnd
	for i := 0; i < len(self.VecGameEnd); i++ {
		for j := 0; j < self.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += self.VecGameEnd[i].GameScore[j]              //总分
			balanceGame.ShowGangCount[j] += int(self.VecGameEnd[i].ShowGangCount[j]) //明杠次数
			balanceGame.HidGangCount[j] += int(self.VecGameEnd[i].HideGangCount[j])  //暗杠次数
		}
	}

	//打印日志
	recrodStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）", cbReason)
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
		if self.CurCompleteCount == 0 {
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
			balanceGame.ZimoCount[i] = _userItem.Ctx.HuBySelfCount

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

			////记录每日统计数据
			//if wintype != public.ScoreKind_pass {
			//	self.TableWriteGameDayRecord(i, 1, isBigWin, balanceGame.GameScore[i])
			//}
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
func (self *Game_mj_jl_jlmj) resetEndDate() {
	self.CurCompleteCount = 0
	self.VecGameEnd = []public.Msg_S_GameEnd{}

	for _, v := range self.PlayerInfo {
		v.OnEnd()
	}
}

func (self *Game_mj_jl_jlmj) UpdateOtherFriendDate(GameEnd *public.Msg_S_GameEnd, bEnd bool) {

}

// ! 给客户端发送总结算
func (self *Game_mj_jl_jlmj) CalculateResultTotal_Rep(msg *public.Msg_C_BalanceGameEeq) {
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

// ! 发送游戏开始场景数据 旁观玩家
func (self *Game_mj_jl_jlmj) sendGameSceneStatusPlayLookon(player *modules.Player) bool {

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

		//玩家番数
		StatusPlay.PlayerFan[i] = int(self.GetUserFlowerCardFan(uint16(i)))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}

		for j := 0; j < len(_item.Ctx.Pilaicardcard) && j < len(StatusPlay.Pilaicardcard[i]); j++ {
			StatusPlay.Pilaicardcard[i][j] = _item.Ctx.Pilaicardcard[j]
		}
		for j := 0; j < len(_item.Ctx.PilaicardcardClass) && j < len(StatusPlay.PilaicardcardClass[i]); j++ {
			StatusPlay.PilaicardcardClass[i][j] = _item.Ctx.PilaicardcardClass[j]
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
	StatusPlay.PayPaostatus = self.PayPaoStatus

	if playerOnChair != nil && playerOnChair.Ctx.Response {
		StatusPlay.ActionMask = public.WIK_NULL
	}
	if playerOnChair != nil {
		StatusPlay.VecGangCard = public.HF_BytesToInts(playerOnChair.Ctx.VecGangCard)
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

	//扑克数据
	if playerOnChair != nil {
		StatusPlay.CardCount, StatusPlay.CardData = self.m_GameLogic.SwitchToCardData2(playerOnChair.Ctx.CardIndex, StatusPlay.CardData)
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

	if self.PayPaoStatus {
		self.SendPaoSettingOffline()
	}

	return true
}

// ! 发送游戏开始场景数据
func (self *Game_mj_jl_jlmj) sendGameSceneStatusPlay(player *modules.Player) bool {

	if player.LookonTableId > 0 {
		self.sendGameSceneStatusPlayLookon(player)
		return true
	}

	wChiarID := player.GetChairID()

	if wChiarID == public.INVALID_CHAIR {
		syslog.Logger().Debug("sendGameSceneStatusPlay invalid chair")
		return false
	}

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

		//玩家番数
		StatusPlay.PlayerFan[i] = int(self.GetUserFlowerCardFan(uint16(i)))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}

		for j := 0; j < len(_item.Ctx.Pilaicardcard) && j < len(StatusPlay.Pilaicardcard[i]); j++ {
			StatusPlay.Pilaicardcard[i][j] = _item.Ctx.Pilaicardcard[j]
		}
		for j := 0; j < len(_item.Ctx.PilaicardcardClass) && j < len(StatusPlay.PilaicardcardClass[i]); j++ {
			StatusPlay.PilaicardcardClass[i][j] = _item.Ctx.PilaicardcardClass[j]
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
	StatusPlay.PayPaostatus = self.PayPaoStatus

	if player.Ctx.Response {
		StatusPlay.ActionMask = public.WIK_NULL
	}

	StatusPlay.VecGangCard = public.HF_BytesToInts(player.Ctx.VecGangCard)

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

	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardData = self.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)

	//发送场景
	self.SendPersonMsg(constant.MsgTypeGameStatusPlay, StatusPlay, wChiarID)
	//发小结消息
	if byte(len(self.VecGameEnd)) == self.CurCompleteCount && self.CurCompleteCount != 0 && int(wChiarID) < self.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := self.VecGameEnd[self.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		self.SendPersonMsg(constant.MsgTypeGameEnd, gamend, wChiarID)
	}

	//发送解散房间所有玩家的反应
	self.SendAllPlayerDissmissInfo(player)

	if self.PayPaoStatus {
		self.SendPaoSettingOffline()
	}

	return true
}

func (self *Game_mj_jl_jlmj) SaveGameData() {
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

		//玩家番数
		StatusPlay.PlayerFan[i] = int(self.GetUserFlowerCardFan(uint16(i)))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}

		for j := 0; j < len(_item.Ctx.Pilaicardcard) && j < len(StatusPlay.Pilaicardcard[i]); j++ {
			StatusPlay.Pilaicardcard[i][j] = _item.Ctx.Pilaicardcard[j]
		}
		for j := 0; j < len(_item.Ctx.PilaicardcardClass) && j < len(StatusPlay.PilaicardcardClass[i]); j++ {
			StatusPlay.PilaicardcardClass[i][j] = _item.Ctx.PilaicardcardClass[j]
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = self.ProvideCard
	StatusPlay.LeftCardCount = self.LeftCardCount
	StatusPlay.PayPaostatus = self.PayPaoStatus

	if self.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = self.OutCardUser
	StatusPlay.OutCardData = self.OutCardData

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

		StatusPlay.VecGangCard = public.HF_BytesToInts(player.Ctx.VecGangCard)

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
func (self *Game_mj_jl_jlmj) SendGameScene(uid int64, status byte, secret bool) {
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
func (self *Game_mj_jl_jlmj) OnExit(uid int64) {
	self.GameCommon.OnExit(uid)
}

// ! 定时器
func (self *Game_mj_jl_jlmj) OnTime() {
	self.GameCommon.OnTime()
}

// ! 计时器事件
func (self *Game_mj_jl_jlmj) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	if dwTimerID == modules.GameTime_Nine {
		if TablePerson := self.GetUserItemByUid(wBindParam); TablePerson != nil {
			//到时间了
			self.OnAutoOperate(TablePerson.Seat, true)
		}
	}
	return true
}

// ! 玩家开启超时
func (self *Game_mj_jl_jlmj) LockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(self.GetPlayerCount()) {
		return
	}

	_userItem := self.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + public.GAME_OPERATION_TIME

	if self.Rule.NineSecondRoom {
		_userItem.Ctx.Timer.SetTimer(modules.GameTime_Nine, public.GAME_OPERATION_TIME)
	}

	self.LimitTime = _userItem.Ctx.CheckTimeOut
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
func (self *Game_mj_jl_jlmj) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(self.GetPlayerCount()) {
		return
	}

	_userItem := self.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0

	//if self.Rule.NineSecondRoom {
	_userItem.Ctx.Timer.KillTimer(modules.GameTime_Nine)
	//}
}

// ! 写日志记录
func (self *Game_mj_jl_jlmj) WriteGameRecord() {
	//写日志记录
	self.OnWriteGameRecord(public.INVALID_CHAIR, "开始监利麻将  发牌......")

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

// 杠分
func (self *Game_mj_jl_jlmj) GetScoreOnGang(gangType uint16) int {
	var Score int = 0
	switch gangType {
	case info.E_Gang_XuGand:
		Score = 1
	case info.E_Gang_AnGang:
		Score = 2
	case info.E_Gang:
		Score = 2
	default:
		syslog.Logger().Debug("杠牌类型找不到")
	}
	return Score * self.Rule.DiFen
}

//获得玩家的加分项
/*

1+加增分数+坑增分数+坑庄分数+补花分数
*/
func (self *Game_mj_jl_jlmj) GetUserAddItemScore(wChairID uint16) int {

	if wChairID < 0 || wChairID > uint16(self.GetPlayerCount()) {
		return 0
	}

	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return 0
	}

	num := 1
	//加增分数
	if self.Rule.ChuZeng {
		num += _userItem.Ctx.VecXiaPao.Num
		self.OnWriteGameRecord(wChairID, fmt.Sprintf("玩家%d出增%d", wChairID, _userItem.Ctx.VecXiaPao.Num))
	}
	//坑增分数
	num += self.Rule.WeiZeng
	self.OnWriteGameRecord(wChairID, fmt.Sprintf("坑增+%d", self.Rule.WeiZeng))
	//坑庄分数
	if self.Rule.KengZhuang && wChairID == self.BankerUser {
		num += 1
		self.OnWriteGameRecord(wChairID, fmt.Sprintf("坑庄+1", wChairID))
	}
	//补花分数
	if self.Rule.BuHuaType != BuHua_Type_NULL {
		num += self.m_FlowerCardFan[wChairID]
		self.OnWriteGameRecord(wChairID, fmt.Sprintf("玩家%d打出花牌:%d个", wChairID, self.m_FlowerCardFan[wChairID]))
	}

	self.OnWriteGameRecord(wChairID, fmt.Sprintf("玩家%d加分项为:%d", wChairID, num))

	return num
}

// 获得玩家补花分数
func (self *Game_mj_jl_jlmj) GetUserFlowerCardFan(wChairID uint16) int {

	if wChairID < 0 || wChairID > uint16(self.GetPlayerCount()) {
		return 0
	}

	_userItem := self.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return 0
	}

	num := 0
	//补花分数
	if self.Rule.BuHuaType != BuHua_Type_NULL {
		num += self.m_FlowerCardFan[wChairID]
	}

	return num
}

// 游戏过程中，玩家的分数发生变化事件
func (self *Game_mj_jl_jlmj) OnUserScoreOffset(_userItem *modules.Player, offset int) bool {
	//if !self.IsGaming() {
	//	syslog.Logger().Debug("非法偏移积分:不在游戏中")
	//	return false
	//}
	//if !_userItem.Acitve {
	//	syslog.Logger().Debug("非法偏移积分:玩家未激活")
	//	return false
	//}
	_userItem.Ctx.StorageScore += offset
	return true
}

// ! 场景保存
func (self *Game_mj_jl_jlmj) Tojson() string {
	var _json modules.GameJsonSerializer
	_json.ToJson(&self.GameMeta)

	_json.GameCommonToJson(&self.GameCommon)

	return public.HF_JtoA(&_json)
}

// ! 场景恢复
func (self *Game_mj_jl_jlmj) Unmarsha(data string) {
	var _json modules.GameJsonSerializer
	if data != "" {
		json.Unmarshal([]byte(data), &_json)

		_json.Unmarsha(&self.GameMeta)
		_json.JsonToStruct(&self.GameCommon)
		self.ParseRule(self.GetTableInfo().Config.GameConfig)
		self.m_GameLogic.Rule = self.Rule
		self.m_GameLogic.HuType = self.HuType
	}
}

func (self *Game_mj_jl_jlmj) OnUserTustee(msg *public.Msg_S_DG_Trustee) bool {
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
						self.LockTimeOut(_item.Seat)
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

func (self *Game_mj_jl_jlmj) entryTrust(byclient bool, _userItem *modules.Player) {
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

func (self *Game_mj_jl_jlmj) accountManage(dismisstime int, trusttime int, autonexttimer int) int {
	if !(int(self.CurCompleteCount) >= self.Rule.JuShu) && self.CurCompleteCount != 0 {
		check := false
		if dismisstime != -1 {
			for i := 0; i < public.MAX_PLAYER_4P; i++ {
				if item := self.GetUserItemByChair(uint16(i)); item != nil {
					if item.CheckTRUST() {
						if check {
							var _msg = &public.Msg_C_DismissFriendResult{
								Id:   item.Uid,
								Flag: true,
							}
							self.OnDismissResult(item.Uid, _msg)
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
			self.SetAutoNextTimer(autonexttimer) //自动开始下一局
			return autonexttimer
		}
	}
	return 0
}

func (self *Game_mj_jl_jlmj) Greate_OutCardRecord(handcards string, msg *public.Msg_C_OutCard, _userItem *modules.Player) string {
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
func (self *Game_mj_jl_jlmj) Greate_OperateRecord(msg *public.Msg_C_OperateCard, _userItem *modules.Player) {
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
