//! 服务器之间的消息
package static

//################
//版本信息
type Msg_C_Info struct {
	AllowLookon int `json:"allowlookon"` //旁观标志
}

//准备消息
type Msg_C_Ready = Msg_C_GoOnNextGame

//取消准备消息
type Msg_C_CancelReady struct {
	Id int64 `json:"id"` //玩家ID
}

//玩家请求总结算消息
type Msg_C_BalanceGameEeq struct {
	Id int64 `json:"id"` //玩家ID
}

//换三张换哪三张牌
type Msg_C_ExchangeThree struct {
	Id           int64   `json:"id"`           //玩家ID
	ExchangeCard [3]byte `json:"exchangecard"` //换三张牌数据
}

type Msg_C_Xiapao struct {
	//	Code string `json:"code"`
	Id       int64 `json:"id"`       //玩家ID
	Num      int   `json:"num"`      //下跑数据
	Status   bool  `json:"status"`   //是否以后每局自动下跑
	ByClient bool  `json:"byclient"` //来源
}

type Msg_C_Piao struct {
	//	Code string `json:"code"`
	Id       int64 `json:"id"`       //玩家ID
	Num      int   `json:"num"`      //选漂数据
	Status   bool  `json:"status"`   //是否以后每局自动选漂
	ByClient bool  `json:"byclient"` //来源
}

//出牌命令
type Msg_C_OutCard struct {
	Id       int64 `json:"id"`   //玩家ID
	CardData byte  `json:"data"` //扑克数据
	//20181123 苏大强 自动出牌
	ByClient bool `json:"byClient"` //
}

//操作命令
type Msg_C_OperateCard struct {
	Id   int64 `json:"id"`   //玩家ID
	Code byte  `json:"code"` //操作类型
	Card byte  `json:"card"` //操作扑克
	//20181207 苏大强 自动出牌
	ByClient bool `json:"byClient"` //
}

//操作命令
type Msg_C_OperateCard32 struct {
	Msg_C_OperateCard
	Code int `json:"code"` //操作类型
}

//用户托管
type Msg_C_Trustee struct {
	Id      int64 `json:"id"`      //玩家ID
	Trustee bool  `json:"trustee"` //是否托管
}

//海底操作
type Msg_C_HD struct {
	Id     int64   `json:"id"`     //玩家ID
	HDCard TagCard `json:"hdcard"` //选择的海底牌
}
type Msg_C_GoOnNextGame struct {
	Id       int64  `json:"id"`       //发送玩家ID 目前参考吧
	Userseat uint16 `json:"userseat"` //20210224 苏大强 托管等待的情况下，旁观玩家不会随旁观对象准备
	ByClient bool   `json:"byClient"` //
}

// 20210108 苏大强 累计时间数据
type Msg_C_TotalTimer struct {
	Id         int64    `json:"id"`         //发送玩家ID 目前参考吧
	Seat       uint16   `json:"seat"`       //玩家座位号
	Open       bool     `json:"open"`       //开始计时
	TotalTimer [4]int64 `json:"totaltimer"` //累计时间，目前估计没有
}

// 即时语音设置
type Msg_C_GVoiceMember struct {
	Id  int   `json:"id"` //member
	Uid int64 `json:"uid"`
}

//游戏中请求提前结束好友房
type Msg_C_DismissFriendReq struct {
	Id     int64  `json:"id"`     //玩家 ID
	Reason string `json:"reason"` //
	By     int    `json:"by"`     //20210714 苏大强 来源服务器（1） 0保留
}

//游戏中请求提前结束好友房的用户选择
type Msg_C_DismissFriendResult struct {
	Id       int64 `json:"id"`   //数据库 ID
	Flag     bool  `json:"flag"` //1:同意  0：拒绝
	FewerNum int   `json:"num"`
	By       int   `json:"by"` //20210714 苏大强 来源服务器（1） 0保留
}

//聊天结构
type Msg_C_UserChat struct {
	Color        int    `json:"color"`        //信息类型
	UserID       int64  `json:"userid"`       //发送用户
	TargetUserID int    `json:"targetuserid"` //目标用户
	Index        int    `json:"index"`        //聊天索引
	Message      string `json:"message"`      //聊天信息
	Rank         byte   `json:"rank"`         //是不是记录详细点的信息
}

//玩家发送语音
type Msg_C_UserYYInfo struct {
	UserID int64  `json:"userid"` // ID
	Addr   string `json:"addr"`   //语音地址
	Rank   byte   `json:"rank"`   //是不是记录详细点的信息
}

type ChatLogs struct {
	UserChat Msg_C_UserChat  `json:"userChat"`
	ExtInfo  ChatLogs_exInfo `json:"extInfo"`
}

type ChatLogs_exInfo struct {
	CurCompleteCount byte  `json:"curCompleteCount"` //局数
	GameEndStatus    byte  `json:"gameEndStatus"`    //发送消息的游戏场景（未开局，游戏中）
	Sendtime         int64 `json:"sendtime"`         //发送时间
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//打拱游戏专用协议
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//出牌命令
type Msg_C_DG_OutCard struct {
	//MsgTypeGameOutCard
	CurrentUser uint16   `json:"currentuser"` //当前牌权用户
	CardCount   byte     `json:"cardcount"`   //出牌扑克数目
	CardData    [40]byte `json:"carddata"`    //扑克列表
	CardType    int      `json:"cardtype"`    //出牌类型
	CardIndex   [40]int  `json:"cardindex"`   //扑克位置索引
}

//吼牌命令
type Msg_C_DG_Roar struct {
	//MsgTypeRoar
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	RoarFlag    byte   `json:"roarflag"`    //1表示吼，0表示不吼
}

//是否杠后需要补牌
type Msg_C_GangHouBuPai struct {
	Id           int64 `json:"id"`           //玩家ID
	GangHouBuPai bool  `json:"ganghoubupai"` //是否杠后补牌
}

//拼三（大吉大利）
//游戏开始

//下注
type Msg_C_XiaZhu struct {
	Seat      uint16 `json:"seat"`      //座位号
	Score     int    `json:"score"`     //下注数,没有看牌情况下的下注数
	RealScore int    `json:"realscore"` //实际的下注数
	Flag      int    `json:"flag"`      //跟注消息(用于区别跟注和加注或全压)，0标志跟注，1表示加注，2表示全压
	IsOver    bool   `json:"isover"`    //游戏是否结束
}

//弃牌
type Msg_C_QiPai struct {
	Seat   uint16 `json:"seat"`
	IsOver bool   `json:"isover"`
}

//看牌
type Msg_C_KanPai struct {
	Seat     uint16 `json:"seat"`
	IsSystem int    `json:"issystem"` //系统亮牌之后给未看牌玩家发送的看牌消息
}

//玩家请求亮牌
type Msg_C_LiangPai struct {
	Seat uint16 `json:"seat"`
}

//消息记录
type Msg_C_ChatLog struct {
	Seat uint16 `json:"seat"`
}

//比牌
type Msg_C_BiPai struct {
	Seat     uint16 `json:"seat"`     //发起者
	Receiver uint16 `json:"receiver"` //接收者
	Score    int    `json:"score"`    //比牌付出的分数
	Winner   uint16 `json:"winner"`   //赢家
	Loser    uint16 `json:"loser"`    //输家
	IsOver   bool   `json:"isover"`
}

//孤注一掷
type Msg_C_GuZhuYiZi struct {
	Seat      uint16   `json:"seat"`      //发起者
	Rich      int      `json:"rich"`      //发起者财富
	Bkanpai   bool     `json:"bkanpai"`   //发起者是否看牌
	Score     int      `json:"score"`     //分数
	RealScore int      `json:"realscore"` //实际分数
	IsOver    bool     `json:"isover"`
	Winner    uint16   `json:"winner"`   //赢家
	Loser     []uint16 `json:"loser"`    //输家
	Receiver  []uint16 `json:"receiver"` //接受者
}

type Msg_C_P3_Auto struct {
	Seat uint16 `json:"seat"`
}
type Msg_C_P3_EndAuto struct {
	Seat uint16 `json:"seat"`
}

//拼三 end

//换座消息
type Msg_C_CHANGESEAT struct {
	ChairID byte `json:"chairid"` //换座位号
}

//操作命令
type Msg_C_ZP_OperateCard struct {
	//MsgTypeGameOperateCard
	Id         int64        `json:"id"`         //玩家ID
	Code       int          `json:"code"`       //操作类型
	Card       byte         `json:"card"`       //操作扑克
	WeaveCount byte         `json:"weavecount"` //组合数目，不包含牌眼
	WeaveInfo  [10][10]byte `json:"weaveinfo"`  //组合牌数据，类型对应//摆牌时必须有这两个字段
	WeaveKind  [10]int      `json:"weavekind"`  //组合牌类型，//摆牌时必须有这两个字段
	ExtType    [10]int      `json:"exttype"`    //扩展类型
	Flag       int          `json:"flag"`       //其它标记  1表示弃时不需要过张
	QZOperate  bool         `json:"qzoperate"`  //true  表示强制动作，false 表示是用户操作
	MessageID  int          `json:"messageid"`  //消息id，用于服务器校验
}

func (self *Msg_C_ZP_OperateCard) Reset() {
	self.Id = 0
	self.Code = 0
	self.Card = 0
	self.WeaveCount = 0
	self.WeaveInfo = [10][10]byte{}
	self.WeaveKind = [10]int{}
	self.ExtType = [10]int{}
	self.Flag = 0
	self.QZOperate = false
}

//出牌命令
type Msg_C_ZP_OutCard struct {
	ChairID  uint16 `json:"chairid"`  //玩家ID
	CardData byte   `json:"carddata"` //扑克数据
	//20181123 苏大强 自动出牌
	ByClient bool `json:"byClient"` //
	Jian     bool `json:"jian"`     // 这张牌是捡的吗?，true是捡的
}

//选择不出牌命令
type Msg_C_ZP_NoOut struct {
	//MsgTypeNoOut
	ChairID uint16 `json:"chairid"` //玩家ID
	Status  int    `json:"status"`  //1表示不出牌
}

//拖牌
type Msg_C_ZP_TuoPai struct {
	//MsgTypeTuoCard
	Id         int64    `json:"id"`         //玩家ID
	Card       byte     `json:"card"`       //操作扑克
	WeaveCount byte     `json:"weavecount"` //组合数目，不包含牌眼
	WeaveInfo  [10]byte `json:"weaveinfo"`  //组合牌数据，类型对应//摆牌时必须有这两个字段
	Index      int      `json:"index"`
}
type Msg_C_DG_SplitCards struct {
	Id        int64 `json:"id"`         //玩家ID
	Cardindex int   `json:"cardindex"`  //切牌位置
	SplitType int   `json:"splitstype"` //切牌类型，0手动，1超时自动
}
type Msg_C_DG_NeedSplitCard struct {
	Id            int64 `json:"id"`   //玩家ID
	NeedSplitFlag bool  `json:"flag"` //是否需要切牌
}

//用户叫分
type Msg_C_SJ_CallScore struct {
	//MsgTypeCallScore
	CallScore int `json:"callscore"` //叫分
}

//用户叫牌
type Msg_C_SJ_CallCard struct {
	//MsgTypeCallCard
	CallCount byte `json:"cbcallcount"` //出牌扑克数目
	CallCard  byte `json:"cbcallcard"`  //扑克列表
}

//底牌扑克
type Msg_C_SJ_ConcealCard struct {
	//MsgTypeConcealCard
	ConcealCount byte    `json:"cbconcealcount"` //底牌数目
	ConcealCard  [8]byte `json:"cbconcealcard"`  //底牌扑克列表
}

//上一轮出牌请求，现在客户端没有使用，客户端使用的是自己保存的数据
type Msg_C_SJ_LastTurn struct {
	//MsgTypeLastTurn
	ChairID uint16 `json:"wchairid"` //
}

//出牌命令
type Msg_C_SJ_OutCard struct {
	//MsgTypeGameOutCard
	CardCount byte     `json:"cbcardcount"` //出牌扑克数目
	CardData  [40]byte `json:"cbcarddata"`  //扑克列表

}

//叫主/买主
type Msg_C_TCBG_CallCard struct {
	CardColor byte `json:"cbcallcolor"` //花色
}

//用户请求兑换记牌器
type Msg_C_Tool_ToolExchange struct {
	//MsgTypeToolExchange
	Price int `json:"price"` //价格
	Num   int `json:"num"`   //数量
}
