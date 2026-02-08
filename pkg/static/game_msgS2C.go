//! 服务器之间的消息
package static

// uid
type Msg_S_Uid struct {
	Uid int64 `json:"uid"`
}

// uid + chair
type Msg_S_UidChair struct {
	Uid   int64  `json:"uid"`
	Chair uint16 `json:"chair"`
}

// 可少人开局人数
type Msg_S_FewerNum struct {
	Uid      int64 `json:"uid"`
	FewerNum int   `json:"num"`
}

// 少人申请关闭
type Msg_S_FewerClose struct {
	FewerNum int `json:"num"`
	State    int `json:"state"` // 1申请前  2申请中 3申请通过
}

//超级防作弊，相关数据
type Msg_S_SuperInfo struct {
	Num        int `json:"current_ai_num"`   //排队人数
	Percent    int `json:"ai_super_percent"` //百分比，分子
	AiSuperNum int `json:"ai_super_num"`
}

//################
//游戏配置
type Msg_S_Option struct {
	GameStatus  byte   `json:"gamestatus"`  //游戏状态 0:空闲状态；101 游戏状态； 102：海底捞状态 103 一局结束状态
	AllowLookon int    `json:"allowlookon"` //允许旁观
	GameConfig  string `json:"gameconfig"`  //游戏会在
	UserReady   []bool `json:"userready"`   //玩家准备状态
	UserStatus  []int  `json:"userstatus"`  //玩家状态
}

//坐下失败
type Msg_S_SitFailed struct {
	Describe string `json:"describe"` //错误描述
}

//用户基本信息结构
type Msg_S_UserInfoHead struct {
	//用户属性
	//FaceID      int    `json:"describe"` //头像索引
	Uid    int64 `json:"uid"`    //用户 I D
	GameID int   `json:"gameid"` //游戏 I D
	//GroupID     int    `json:"describe"` //社团索引
	//UserRight   int    `json:"describe"` //用户等级
	//Loveliness  int    `json:"describe"` //用户魅力
	//MasterRight int    `json:"describe"` //管理权限
	Name string `json:"name"` //用户名字

	//用户属性
	Gender      int  `json:"gender"`      //用户性别
	MemberOrder byte `json:"memberorder"` //会员等级
	MasterOrder byte `json:"masterorder"` //管理等级

	//用户状态
	TableID uint16 `json:"tableid"` //桌子号码
	ChairID uint16 `json:"chairid"` //椅子位置
	//UserStatus int    `json:"userstatus"` //用户状态
	//用户积分
	UserScoreInfo TagUserScoreV2 `json:"scoreinfo"` //积分信息
	//CustomFaceVer    int                  `json:"describe"` //上传头像
	//PropResidualTime [PROP_MEMBER_MAX]int `json:"describe"` //道具时间
	FaceUrl string `json:"faceurl"` //头像
	//ExtInfo          string               `json:"extinfo"`  //扩张信息
}

//游戏状态		// 手机兼容
type Msg_S_StatusFree struct {
	CellScore  int    `json:"cellscore"`  //基础金币
	BankerUser uint16 `json:"bankeruser"` //庄家用户
	GameType   byte   `json:"gametype"`   //游戏类型 0:7pi4lai, 1:11pi4lai
	KaiKou     bool   `json:"kaikou"`     //确定开口口口
	TaskAble   bool   `json:"taskable"`   //任务是否开启
}

//用户托管
type Msg_S_Trustee struct {
	Trustee bool   `json:"trustee"` //是否托管
	ChairID uint16 `json:"chairid"` //托管用户
}

//游戏状态
type CMD_S_StatusPlay struct {
	//游戏变量
	CellScore   int    `json:"cellscore"`   //单元积分
	SiceCount   uint16 `json:"sicecount"`   //骰子点数
	BankerUser  uint16 `json:"bankeruser"`  //庄家用户
	CurrentUser uint16 `json:"currentuser"` //当前用户

	//状态变量
	ActionCard    byte   `json:"actioncard"`    //动作扑克
	ActionMask    byte   `json:"actionmask"`    //动作掩码
	ChiHuKindMask uint64 `json:"chihukindmask"` //吃胡类型(当ActionMask有胡的牌权时此字段表示发送胡牌类型牌权）

	LeftCardCount byte  `json:"leftcardcount"` //剩余数目
	VecGangCard   []int `json:"vecgangcard"`   //本局弃杠的牌

	//出牌信息
	OutCardUser        uint16      `json:"outcarduser"`        //出牌用户
	OutCardData        byte        `json:"outcarddata"`        //出牌扑克
	DiscardCount       [4]byte     `json:"discardcount"`       //丢弃数目
	DiscardCard        [4][55]byte `json:"discardcard"`        //丢弃记录
	DiscardCardClass   [4][55]byte `json:"discardcardclass"`   //丢弃记录
	Pilaicardcard      [4][12]byte `json:"pilaicardcard"`      //皮籁杠数
	PilaicardcardClass [4][12]byte `json:"pilaicardcardclass"` //皮籁杠来源
	//扑克数据
	CardCount byte     `json:"cardcount"` //扑克数目
	CardData  [14]byte `json:"carddata"`  //扑克列表
	// //20181208 苏大强 填加发送牌信息
	SendCardData byte `json:"sendcarddata"` //发牌信息
	//组合扑克
	WeaveCount     [4]byte            `json:"weavecount"`     //组合数目
	WeaveItemArray [4][4]TagWeaveItem `json:"weaveitemarray"` //组合扑克
	PlayerFan      [4]int             `json:"playerfan"`      //玩家当前番数

	PiZiCard     byte    `json:"pzicard"`      //皮子
	GameType     byte    `json:"gametype"`     //游戏类型
	KaiKou       bool    `json:"kaikou"`       //确定开口口口
	RenWuAble    bool    `json:"renwuable"`    //任务是否开启
	PiGangCount  [4]byte `json:"pigangcount"`  //皮子癞子个数
	LaiGangCount [4]byte `json:"laigangcount"` //癞子癞子个数
	TheOrder     byte    `json:"theorder"`     //当前是第几局
	PaoNum       [4]byte `json:"paonum"`
	PaoStatus    [4]bool `json:"paostatus"`
	PayPaostatus bool    `json:"paypaostatus"`
	//20181210 苏大强 添加超时信息
	Overtime           int64                        `json:"overtime"`           //超时时间
	LastOutCardUser    uint16                       `json:"lastoutcarduser"`    //最近一次出牌人
	GangHouBuPai       [4]bool                      `json:"ganghoubupai"`       //是否杠后补牌
	BaoQingType        [4]int                       `json:"baoqingtype"`        //用户报请状态
	IsBaoTing          [4]bool                      `json:"isbaoting"`          //用户报听状态
	TingStart          [4]bool                      `json:"tingStart"`          //用户报听状态原始
	DaoChePai          [4][2]byte                   `json:"daochepai"`          //用户倒车的牌,不能立马出
	FengQuan           byte                         `json:"fengquan"`           //当前风圈，滁州玩法
	CanAction          bool                         `json:"canaction"`          //滁州老三番 庄家能不能动
	IsJinHu            [4]bool                      `json:"isjinhu"`            //20191010 沈强恩施麻将 玩家禁胡
	CardLeft           TagCardLeftItem              `json:"cardleft"`           //牌堆
	Whotrust           [4]bool                      `json:"whotrust"`           //20191115 苏大强 崇阳 托管
	ExchangeThreeState [4]TagExchangeThreeStateItem `json:"exchangethreestate"` //换三张的状态 add by zw for汉川搓虾子
	GuoHuCount         [4]int                       `json:"guohucount"`         //过户次数
	//add by zwj for黄州晃晃
	ShowKaikouTip      bool                      `json:"showkaikoutip"`      //开口提示
	OnlyBigHu          bool                      `json:"onlybighu"`          //是否只能胡大胡
	ExchangeCardStatus bool                      `json:"exchangecardstatus"` //换牌状态
	ExchangeCardInfo   [4]TagExchangeCardInfo    `json:"exchangecardinfo"`   //换牌信息
	ChihuCardsInfo     [4][]*ChihuCardsInfo_xlch `json:"chihucardsinfo"`     //玩家胡牌信息
	GameHuCards        []int                     `json:"gamehucards"`        //当前这局牌胡了哪些牌
	CurPunishSeat      uint16                    `json:"curpunishseat"`      // 超时罚分的当前玩家
	PunishStartTime    int64                     `json:"punishstarttime"`    // 超时罚分的开始时间
	AlarmStatus        [4]int                    `json:"alarmstatus"`        //玩家报警标识
	ExOperateInfo      *ExOperateInfo            `json:"exOperateInfo"`      //玩家牌权扩展信息
	IsOutCard          bool                      `json:"is_out_card"`        //是否可以出牌
	MaxPiao            [4]int                    `json:"maxpiao"`            //最大漂分
	ShangLou           bool                      `json:"shanglou"`
	TotalTime          [4]int64                  `json:"totalTime"`      //累计时间列表
	TotalTimeruser     uint16                    `json:"totaltimeruser"` //正在记录累计时间的玩家
	UserReady          [4]bool                   `json:"userready"`      //下一局准备好的用户（点击准备好的用户）
}
type CMD_S_StatusPlay32 struct {
	CMD_S_StatusPlay
	ActionMask int `json:"actionmask"` //动作掩码
}

//海底状态
type Msg_S_StatusHiDi struct {
	//游戏变量
	CellScore   int    `json:"cellscore"`   //单元积分
	BankerUser  uint16 `json:"bankeruser"`  //庄家用户
	CurrentUser uint16 `json:"currentuser"` //当前用户

	LeftCardCount byte `json:"leftcardcount"` //剩余数目

	//出牌信息
	DiscardCount [4]byte     `json:"discardcount"` //丢弃数目
	DiscardCard  [4][55]byte `json:"discardcard"`  //丢弃记录

	//扑克数据
	CardCount byte     `json:"cardcount"` //扑克数目
	CardData  [14]byte `json:"carddata"`  //扑克列表

	//组合扑克
	WeaveCount     [4]byte            `json:"weavecount"`     //组合数目
	WeaveItemArray [4][4]TagWeaveItem `json:"weaveitemarray"` //组合扑克

	PiZiCard     byte    `json:"pzicard"`      //皮子
	GameType     byte    `json:"gametype"`     //游戏类型
	KaiKou       bool    `json:"kaikou"`       //确定开口口口
	RenWuAble    bool    `json:"renwuable"`    //任务是否开启
	PiGangCount  [4]byte `json:"pigangcount"`  //皮子癞子个数
	LaiGangCount [4]byte `json:"laigangcount"` //癞子癞子个数

	HiDiCardData [14]byte   `json:"hidicarddata"` //海底扑克
	HDCard       [4]TagCard `json:"hdcard"`       //海底牌记录
	RemainTime   int        `json:"remaintime"`   //操作剩余时间 （海底捞
	FengQuan     byte       `json:"fengquan"`     //当前风圈，滁州玩法
}

//!下跑
type Msg_S_Xiapao struct {
	//	Code string `json:"code"`
	Num    [4]int  `json:"num"`    //下跑数据
	Status [4]bool `json:"status"` //是否下跑
	Always [4]bool `json:"always"` //是否以后每局自动下跑
}

type Msg_S_PaoSetting struct {
	//	enum {XY_ID = SUB_S_GAME_SETTING};
	PaoStatus    bool   `json:"paostatus"` //是否已经选过
	PaoCount     int    `json:"paocount"`  //跑数
	Always       bool   `json:"always"`
	BankerUser   uint16 `json:"bankeruser"`   //庄家用户
	Overtime     int64  `json:"overtime"`     //超时时间
	Currentcount int    `json:"currentcount"` //当前局数
}

//!选漂
type Msg_S_XuanPiao struct {
	//	Code string `json:"code"`
	ChairID uint16  `json:"chairid"` //当前选漂椅子位置
	Num     [4]int  `json:"num"`     //下跑数据
	Status  [4]bool `json:"status"`  //是否以后每局自动选漂
}

type Msg_S_PiaoSetting struct {
	PiaoStatus   bool   `json:"piaostatus"`   //是否已经选过
	PiaoCount    int    `json:"piaocount"`    //漂数
	Always       bool   `json:"always"`       //是否一直选这个漂
	BankerUser   uint16 `json:"bankeruser"`   //庄家用户
	CurrentCount byte   `json:"currentcount"` //20190807 通城麻将使用
	Overtime     int64  `json:"overtime"`     //超时时间
	MaxPiao      [4]int `json:"maxpiao"`      //历史上最大的漂
}

//换三张设置
type Msg_S_ExchangeThreeSet struct {
	ExchangeStatus bool    `json:"exchangestatus"` //是否已经选过
	CardThree      [3]byte `json:"cardthree"`      //换的哪三张牌
}

//换完牌以后每个玩家手上的新牌
type Msg_S_AfterExchangeThreeCard struct {
	CardData   [14]byte `json:"carddata"`   //扑克列表
	TimeSecond int64    `json:"timesecond"` //下一个定时器的时间戳
	MagicCard  byte     `json:"magiccard"`  //赖子牌
}

//换三张状态推送
type TagExchangeThreeStateItem struct {
	ExchangeState     bool    `json:"exchangestate"`     //ture 选好了 false 没选好
	ThreeExchagneCard [3]byte `json:"threeexchangecard"` //玩家选好的三张牌（选好后才解析这个三个字段）
}

//换三张的状态
type Msg_S_ExchangeThreeState struct {
	AllExchangeState [4]TagExchangeThreeStateItem `json:"allexchangestate"` //玩家的选中状态
}

//桌子状态信息
type Msg_S_TableStatus struct {
	TableID    uint16 `json:"tableid"`    //桌子号码
	TableLock  bool   `json:"tablelock"`  //锁定状态
	PlayStatus bool   `json:"playstatus"` //游戏状态
}

//换牌状态信息
type TagExchangeCardInfo struct {
	Status          bool  `json:"status"`          //ture 选好了 false 没选好
	ExchangeType    int   `json:"exchangetype"`    //换牌方式:顺时针,逆时针,对换
	ExchangeCards   []int `json:"exchangecards"`   //玩家选好的三张牌（玩家换出的牌）
	ExChangeInCards []int `json:"exchangeincards"` //玩家换来的牌
}

//游戏开始
type Msg_S_GameStart struct {
	SiceCount     uint16          `json:"sicecount"`   //骰子点数
	BankerUser    uint16          `json:"bankeruser"`  //庄家用户
	CurrentUser   uint16          `json:"currentuser"` //当前用户
	UserAction    byte            `json:"useraction"`  //用户动作
	CardData      [14]byte        `json:"carddata"`    //扑克列表
	MagicCard     byte            `json:"magiccard"`
	Overtime      int64           `json:"overtime"`      //超时时间
	SendCardData  byte            `json:"sendcarddata"`  //发牌信息
	LeftCardCount byte            `json:"leftcardcount"` //剩余数目
	PlayerFan     [4]int          `json:"playerfan"`     //玩家当前番数
	GangHouBuPai  [4]bool         `json:"ganghoubupai"`  //是否杠后补牌
	FengQuan      byte            `json:"fengquan"`      //当前风圈，滁州玩法
	TingStart     [4]bool         `json:"tingStart"`     //用户报听状态
	CanAction     bool            `json:"canaction"`     //庄家能不能动
	CurrentCount  byte            `json:"currentcount"`  //当前局数
	CardLeft      TagCardLeftItem `json:"cardleft"`      //牌堆
	Whotrust      [4]bool         `json:"whotrust"`      //20191115 苏大强 崇阳 托管
}

type Msg_S_GameStartWhHH_WhMJ struct { //武汉晃晃武汉麻将
	Msg_S_GameStart
	MagicAnimate byte `json:"magicanimate"` //0没有动画  1显示原赖  2显示风见原赖
}

type Msg_S_GameStart32 struct {
	Msg_S_GameStart
	UserAction int `json:"useraction"` //用户动作32
}

//潜江胡牌分详情
type WinScoreDetailQJ struct {
	WinScore     float64 `json:"winscore"`     //胡牌分
	GangScore    float64 `json:"gangscore"`    //杠牌分
	BeiGangScore float64 `json:"beigangscore"` //被杠分
	OtherScore   float64 `json:"otherscore"`   //其他分
}

//游戏结束
type Msg_S_GameEnd struct {
	//	enum {XY_ID = SUB_S_GAME_END};
	HuangZhuang     bool        `json:"hz"`              // 是否荒庄
	ChiHuCard       byte        `json:"chihucard"`       // 吃胡扑克
	KaiKou          bool        `json:"kaikou"`          // 标记是开口还是口口
	ProvideUser     uint16      `json:"provideuser"`     // 点炮用户
	TaskFinished    bool        `json:"taskfinished"`    // 第三期任务完成
	GameScore       [4]int      `json:"gamescore"`       // 游戏积分
	WinOrLose       [4]byte     `json:"winorlose"`       // 输赢，0表示输，1表示赢，2表示平
	GameScorefloat  [4]float64  `json:"gamescorefloat"`  // 游戏积分浮点
	GameAdjustScore [4]int      `json:"gameadjustscore"` // 修订游戏积分
	GameTotal       [4]int      `json:"gametotal"`       // 总成绩
	UserScore       [4]int      `json:"user_score"`      // 玩家剩余积分
	UserVitamin     [4]float64  `json:"user_vitamin"`    // 玩家剩余疲劳值
	CardCount       [4]byte     `json:"cardcount"`       // 扑克数目
	CardData        [4][14]byte `json:"carddata"`        // 扑克数据
	StrEnd          [4]string   `json:"strend"`          // 胡牌信息
	BirdCard        byte        `json:"birdcard"`        // 鸟牌，复用此字段表示断线重连
	BirdUser        uint16      `json:"birduser"`        // 中鸟玩家
	Tax             int         `json:"tax"`             // 游戏税收
	ChiHuUserCount  uint16      `json:"chihuusercount"`  // 胡牌总数
	IsQuit          bool        `json:"isquit"`          // 是否强退
	Winner          uint16      `json:"winner"`          // 胡牌玩家

	BaseScore          int         `json:"basescore"`          //基础分
	OperateCount       [4]uint16   `json:"operatecount"`       //开口次数
	ShowGangCount      [4]uint16   `json:"showgangcount"`      //明杠次数
	XuGangCount        [4]uint16   `json:"xugangcount"`        //续杠次数
	DianGangCount      [4]uint16   `json:"diangangcount"`      //点杠次数
	HideGangCount      [4]uint16   `json:"hidegangcount"`      //暗杠次数
	HongzhongGangCount [4]uint16   `json:"hongzhonggangcount"` //红中杠次数
	FaCaiGangCount     [4]uint16   `json:"facaigangcount"`     //发财杠次数
	PiZiGangCount      [4]uint16   `json:"pizigangcount"`      //皮子杠次数
	MagicGangCount     [4]uint16   `json:"magicgangcount"`     //赖子杠次数
	MagicOutCount      [4]uint16   `json:"magicoutcount"`      //丢赖子的个数
	MaxFSCount         [4]uint16   `json:"maxfscount"`         //用户翻数
	QiangFailCount     [4]byte     `json:"qiangfialcount"`     //抢错次数
	Contractor         uint16      `json:"contractor"`         //承包用户
	WContractor        [4]uint16   `json:"wcontractor"`        //承包用户（武汉麻将（斑马汉麻）存在多人承包的情况）
	WWinner            [4]bool     `json:"wWinner"`            //胡牌玩家 放放 一炮多响  20181108 苏大强
	CbBirdData         [4][8]byte  `json:"cbBirdData"`         //每个人抓中的鸟牌  每人最多8张  20181108 苏大强
	CbBirdData_ex      [4][10]byte `json:"cbBirdDataex"`       //每个人抓中的鸟牌  荆州红中杠最大10鸟
	BingoBirdCount     [4]byte     `json:"bingobirdcount"`     //每个人的中鸟个数
	WChiHuKind         [4]uint64   `json:"wchihukind"`         //吃胡类型(二进制，客户端解析多少种胡牌类型) 放放有一炮多响
	WHardHu            [4]byte     `json:"whardhu"`            //是否硬胡 1硬胡 0软胡 放放一炮可能多响

	//	HardChiHu bool `json:"yhu"` //标记是否是硬胡

	BigHuKind        byte       `json:"bighukind"`        //标记大胡类型
	WBigHuKind       [4]byte    `json:"wbighukind"`       //标记大胡类型
	FengDingKind     [4]byte    `json:"fengdingkind"`     //封顶类型
	TheOrder         byte       `json:"theorder"`         //第几局结算
	EndStatus        byte       `json:"endstatus"`        //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
	PaoCount         [4]uint16  `json:"paocount"`         //跑数
	PaoCountfloat    [4]float64 `json:"paocountfloat"`    //跑数 浮点
	NextCard         [4]byte    `json:"nextcard"`         //4人玩法发4张牌 3人玩法三张牌
	LastSendCardUser uint16     `json:"lastsendcarduser"` //最近一次发牌的接收者
	ChiHuKind        uint64     `json:"chihukind"`        //吃胡类型(二进制，客户端解析多少种胡牌类型)
	HardHu           byte       `json:"hardhu"`           //是否硬胡 1硬胡 0软胡

	WeaveItemCount [4]byte            `json:"weaveitemcount"` //组合数目
	WeaveItemArray [4][4]TagWeaveItem `json:"weaveitemarray"` //组合扑克
	MagicCard      byte               `json:"magiccard"`      //赖子
	LeftCardCount  byte               `json:"leftcardcount"`  //剩余数目

	JingDingUser  [4]int              `json:"jingdinguser"`  //金顶用户
	RepertoryCard [MAX_REPERTORY]byte `json:"repertorycard"` //牌堆牌

	LianJinUser      uint16     `json:"lianjinuser"`      //连金用户
	FanJinUser       uint16     `json:"fanjinuser"`       //反金用户
	JianFengYuanUser uint16     `json:"jianfengyuanuser"` //连金用户
	UserDianCount    [4]float32 `json:"userdiancount"`    //用户点数
	ContractorType   int        `json:"contractortype"`   //承包类型
	WContractorType  [4]int     `json:"wcontractortype"`  //承包类型(多人承包)
	IsReChong        bool       `json:"isrechong"`        //是否热冲
	// GangScore  []int `json:"gangscore"`  // 本局杠分
	// MagicScore []int `json:"magicscore"` // 本局飘癞子分
	IsQuitUser [4]bool  `json:"isquituser"` //是否是强退玩家
	DisCard    [4][]int `json:"discard"`    //摊牌
	//20191115 苏大强 阳新，最后一张牌发
	LastCard     byte                 `json:"lastcard"`     //最后一张
	PeiScore     [4]int               `json:"peiscore"`     //卡五星查听赔分
	TingFlag     [4]int               `json:"tingflag"`     //卡五星赔庄标识
	PaiXingScore [4]int               `json:"paixingscore"` //牌型分
	PiaoScore    [4]int               `json:"piaoscore"`    //漂分
	StorageScore [4]int               `json:"storagescore"` //游戏过程中的分数变化
	HuDetails    [4][]*HuDetails_xlch `json:"hudetails"`    //胡牌详情(血流成河)

	MaiMaInfo        [12]byte `json:"maimainfo"`        //买马
	HitMaInfo        [12]byte `json:"hitmainfo"`        //中马
	HitMaNumber      byte     `json:"hitmanumber"`      //中马数
	AutoNextGameTime int      `json:"autoNextGameTime"` //自动下一局时间
	RecordTime       int64    `json:"recordTime"`       //20200520 苏大强 结算记录时间
	CbBirdData_ex2   [4][]int `json:"cbBirdDataex2"`    //每个人抓中的鸟牌  荆州红中杠最大12鸟

	ExtEndType int `json:"extendtype"` // 拓展结束类型，比如 1 托管3局结束游戏

	WinnerScoreDetail  [4]WinScoreDetailQJ `json:"winnerscoredetail"` //潜江胡分详情
	GangPaoProvideUser uint16              `json:"gangpaoprovideuser"`
	FanFeng            [4]int              `json:"fanfeng"` //番分(卡五星小结算面板显示用)
	MaScore            [4]int              `json:"mascore"`
	TrustPunishScore   [4]int              `json:"trustpunishscore"` //托管惩罚分数
	GangFen            [4]int              `json:"gangfen"`          //杠分
	HuFen              [4]int              `json:"hufen"`            //胡分
	TrustFaScore       [4]int              `json:"trustfascore"`     //托管承担罚分
}

//操作提示
type Msg_S_IsOutCard struct {
	Seat      uint16 `json:"seat"`        //座位号
	IsOutCard bool   `json:"is_out_card"` // 是否可以出牌 true 可以出
}

//操作提示
type Msg_S_OperateNotify struct {
	//	enum {XY_ID = SUB_S_OPERATE_NOTIFY};

	ResumeUser    uint16 `json:"resumeuser"`    //还原用户
	ActionMask    byte   `json:"actionmask"`    //动作掩码
	ChiHuKindMask uint64 `json:"chihukindmask"` //吃胡类型(当ActionMask有胡的牌权时此字段表示发送胡牌类型牌权）
	ActionCard    byte   `json:"actioncard"`    //动作扑克
	EnjoinHu      bool   `json:"enjoinhu"`      //是否4番下禁止胡
	//20181205 需要限时操作的 苏大强
	Overtime int64 `json:"overtime"` //超时时间
	//玩家不可杠的牌列表
	VecGangCard []int `json:"vecgangcard"` //本局弃杠的牌
	//add by zwj for 黄州晃晃（需要给客户端一个提示）
	ShowKaikouTip bool `json:"showkaikoutip"`
	//20200429 苏大强 添加
	ExOperateInfo *ExOperateInfo `json:"exOperateInfo"` //玩家牌权扩展信息
}

//操作提示
type Msg_S_OperateNotify32 struct {
	//	enum {XY_ID = SUB_S_OPERATE_NOTIFY};
	Msg_S_OperateNotify

	ActionMask int `json:"actionmask"` //动作掩码
}

//操作命令
type Msg_S_OperateResult struct {
	//enum {XY_ID = SUB_S_OPERATE_RESULT};
	OperateUser uint16  `json:"operateuser"` //操作用户
	ProvideUser uint16  `json:"provideuser"` //供应用户
	OperateCode byte    `json:"operatecode"` //操作代码
	OperateCard byte    `json:"operatecard"` //操作扑克
	HaveGang    [4]bool `json:"havegang"`    //是否杠过
	//20181205 需要限时操作的 苏大强
	Overtime     int64      `json:"overtime"`     //超时时间
	GameScore    [4]int     `json:"gamescore"`    //最新总分
	GameVitamin  [4]float64 `json:"game_vitamin"` //最新疲劳值信息
	ScoreOffset  [4]int     `json:"scoreoffset"`  //分数变化量
	LaiGangCount [4]byte    `json:"laigangcount"` //玩家飘癞子的个数
	DaoChePai    [2]byte    `json:"dochepai"`     //倒车牌
	PlayerFan    [4]int     `json:"playerfan"`    //玩家当前番数
}

//操作命令
type Msg_S_OperateResult32 struct {
	Msg_S_OperateResult
	OperateCode int `json:"operatecode"` //操作代码

}

//用户状态
type Msg_S_UserStatus struct {
	UserID     int64 `json:"userid"`     //数据库 ID
	TableID    int   `json:"tableid"`    //桌子位置
	ChairID    int   `json:"chairid"`    //椅子位置
	UserStatus uint8 `json:"userstatus"` //用户状态
	UserReady  bool  `json:"userready"`  //玩家准备状态
}

// 服务器发给客户端 总结算
type Msg_S_BALANCE_GAME struct {
	Userid           int64  `json:"userid"`           // userid房主
	GameScore        [4]int `json:"gamescore"`        //游戏积分
	ChiHuUserCount   [4]int `json:"chihuusercount"`   //胡牌总数
	ProvideUserCount [4]int `json:"provideusercount"` //点炮次数
	FXMaxUserCount   [4]int `json:"fxmaxusercount"`   //最大翻数
	HHuUserCount     [4]int `json:"hhuusercount"`     //大胡次数
	ZimoCount        [4]int `json:"zimocount"`        //自摸次数
	ShowGangCount    [4]int `json:"showgangcount"`    //明杠次数
	HidGangCount     [4]int `json:"hidgangcount"`     //暗杠次数
	CurTotalCount    byte   `json:"curtotalcount"`    //当前总局数
	UserEndState     [4]int `json:"userendstate"`     //总结算时玩家状态：0正常，1解散，2离线
	End              byte   `json:"end"`              //0表示结束了的总结算  1表示没有结束，客户端点击查看
	FXScoreUserCount [4]int `json:"fxscoreusercount"` //最大分数
	LaiYouHuCount    [4]int `json:"laiyouhucount"`    //赖油胡牌次数
	PunishCount      [4]int `json:"punishcount"`      //超时罚分的次数计数
	ZhuangCount      [4]int `json:"zhuangcount"`      //做庄次数
	//20200916 苏大强 大结算追加 队长信息
	PartnerName      [4]string `json:"partnerName"`      //队长姓名
	PartnerId        [4]int64  `json:"partnerId"`        //队长id
	TrustPunishScore [4]int    `json:"trustpunishscore"` //托管惩罚分数
	TimeStart        int64     `json:"timestart"`        //游戏开始时间
	TimeEnd          int64     `json:"timeend"`          //游戏结束时间
	EndStatus        byte      `json:"endstatus"`        //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
}

type Msg_Card struct {
	Card byte `json:"card"` //牌数据
}

//出牌命令
type Msg_S_OutCard struct {
	User int  `json:"user"` //出牌用户
	Data byte `json:"data"` //出牌扑克
	//20181123 苏大强 添加 从哪里来的 自动打牌
	ByClient bool `json:"byClient"` //
	//20181211 苏大强 打出一张牌也给超时时间
	Overtime    int64 `json:"overtime"`    //超时时间
	OutCardType int   `json:"outcardtype"` //出牌方式(玩家托管服务器自动、玩家主动、服务器自动)
}

//出牌命令（麻城麻将）
type Msg_S_OutCard_Mcmj struct {
	Msg_S_OutCard
	OutOrOpreate byte `json:"code"` //0出牌 1杠牌
}

//出牌命令 黄梅麻将
type Msg_S_OutCard_HMMJ struct {
	Msg_S_OutCard
	CardDataArray [4]byte `json:"dataarray"` //批量数据 add by zwj
}

//出牌命令
type Msg_S_Notify_BaoTing struct {
	IsBaoTing [4]bool `json:"isbaoting"` //出牌报听信息
	TingStart [4]bool `json:"tingStart"` //用户报听状态
	CanAction bool    `json:"canaction"` //庄家能不能行动
}

//发送扑克
type Msg_S_SendCard struct {
	CardData      byte   `json:"carddata"`      //扑克数据
	ActionMask    byte   `json:"actionmask"`    //动作掩码
	ChiHuKindMask uint64 `json:"chihukindmask"` //吃胡类型(当ActionMask有胡的牌权时此字段表示发送胡牌类型牌权）
	CurrentUser   uint16 `json:"currentuser"`   //当前用户
	IsHD          bool   `json:"ishd"`          //是否海底
	IsGang        bool   `json:"isgang"`        //是否是杠牌后发牌
	EnjoinHu      bool   `json:"enjoinhu"`      //是否4番下禁止胡
	//20181205 限时操作
	Overtime    int64 `json:"overtime"`    //超时时间
	VecGangCard []int `json:"vecgangcard"` //本局弃杠的牌
	//20191018 苏大强
	NewPiZiCard byte `json:"newpizicard"` //颍州玩法，活杠修改杠后的牌
	//20191021 苏大强 追加剩牌数
	LeftCardCount byte `json:"leftcardcount"` //剩余数目
	//这张牌是否亮 zwj
	ShowThisCard bool `json:"showthiscard"`
	//20200902 苏大强
	NoAction bool `json:"noaction"` //恩施禁止养痞，如果网卡，可能导致客户端可以打牌出去
	//20200426 苏大强 附加操作信息
	ExOperateInfo *ExOperateInfo `json:"exOperateInfo"`
}

//恩施麻将附加操作信息
type ExOperateInfo struct {
	ExInfo string `json:"exInfo"` //扩展显示信息
	FenShu [4]int `json:"fenShu"` //分数结构
}

//发送扑克 黄梅麻将
type Msg_S_SendCardHMMJ struct {
	Msg_S_SendCard
	CardDataArray [4]byte `json:"carddataarray"` //扑克数据 批量发牌
}

//发送扑克
type Msg_S_SendCard32 struct {
	Msg_S_SendCard
	ActionMask int `json:"actionmask"` //动作掩码

}

//海底提示
type Msg_S_HD struct {
	User   uint16  `json:"user"`   //下一个赌海底用户
	HDCard TagCard `json:"hdcard"` //被选过的海底牌
	//20181205 限时操作
	Overtime int64 `json:"overtime"` //超时时间
}

type Msg_S_UserGPSReq struct {
	Type      int     `json:"type"`      //类型
	UserID    int64   `json:"userid"`    //玩家UID
	Longitude float32 `json:"longitude"` //经纬度
	Latitude  float32 `json:"latitude"`  //经纬度
	//AddrLen   uint8   `json:"len"` //地址长度
	Addr string `json:"addr"` //详细地址
}

//断线重连服务器推送过来的玩家准备信息
type Msg_S_PLAYER_REDEAY_STATE struct {
	Situation [MAX_CHAIR_NORMAL]bool `json:"situation"` //玩家的准备状态
}

//断线重连服务器推送过来的解散房间的信息（如果有）
type Msg_S_DisMissRoom struct {
	Situation [MAX_CHAIR_NORMAL]byte `json:"situation"` //其他玩家对解散房间的反应 (其中下标 0 1 2 3表示椅子编号)
	Timer     int64                  `json:"timer"`     //自动结算倒计时时间
	FewerNum  int                    `json:"num"`
	Reason    string                 `json:"reason"`
}

//游戏中请求提前结束好友房
type Msg_S_DismissFriendRep struct {
	Id     int64  `json:"id"`    //玩家 ID
	Timer  int    `json:"timer"` //服务器计算出来的解散计时时间的剩余秒数
	Reason string `json:"reason"`
}

//旁观控制
type Msg_S_LookonControl struct {
	Id          uint16 `json:"id"`          //用户标识
	AllowLookon byte   `json:"allowlookon"` //允许旁观
}

//消息数据包
type Msg_S_Message struct {
	Type    int    `json:"type"`    //消息类型
	Content string `json:"content"` //消息内容
}

//玩家断线的剩余时间
type Msg_S_UserOfflineTime struct {
	Time             [8]int `json:"time"`             //最多8个玩家，每个人的剩余时间。单位： 秒钟
	Code             [8]int `json:"code"`             //玩家是否离线；
	ShortOfflineTime [8]int `json:"shortOfflineTime"` //20191127 苏大强 离线申请解散时间
	Msg_S_UserTotalTime
}
type Msg_S_UserTotalTime struct {
	TotalTimerUser uint16   `json:"totaltimeruser"` //20210121 苏大强 正在累计时的玩家
	TotalTime      [4]int64 `json:"totaltime"`      //20210121 苏大强 累计时玩法中需要的，刷时间用

}

//玩家比赛分过低暂停游戏的剩余时间
type Msg_S_UserVitaminLowTime struct {
	Time [8]int `json:"time"` //最多8个玩家，每个人的剩余时间。单位： 秒钟
	Code [8]int `json:"code"` //玩家是否离线；
}

//海底捞
type Msg_S_SendHDV struct {
	Cards       [8]byte `json:"cards"`       //海底扑克数据
	CurrentUser uint16  `json:"currentuser"` //当前用户
	WinUser     uint16  `json:"winuser"`     //海底捞胡牌用户
	SendNum     int     `json:"sendnum"`     //发牌数
}

//玩家飘癞子
type Msg_S_OutMagic struct {
	User        int    `json:"user"`        //出牌用户
	Data        byte   `json:"data"`        //出牌扑克
	Show        bool   `json:"show"`        //是否出动画
	GameScore   [4]int `json:"gamescore"`   //用户当前积分
	ScoreOffset [4]int `json:"scoreoffset"` //本次积分偏移量
}

//通知消息
type Msg_S_NotificationMessage struct {
	Content string `json:"content"` //通知消息内容
}

//20191218 苏大强 恩施麻将 客户端想知道是那种情况出现的禁胡
type Msg_S_Notification_ex struct {
	OperateAction_ex uint64 `json:"operateaction_ex"` //通知操作类型扩展
	Card             byte   `json:"card"`             //牌
	CardClass        byte   `json:"cardclass"`        //牌类型（普通牌，痞子牌，癞子牌）
	Seat             uint16 `json:"seat"`             //玩家座位号
	Ctx              byte   `json:"ctx"`              //游戏场景（正常，杠开，热冲，海底）暂定，未用
}

//通知游戏操作状态
type Msg_S_NotificationOperateStatus struct {
	OperateUser   uint16                `json:"operateuser"`           //通知操作类型
	OperateAction int                   `json:"operateaction"`         //通知操作类型
	OperateResult string                `json:"operateresult"`         //通知操作结果
	ExInfo        Msg_S_Notification_ex `json:"msg_s_notification_ex"` //拓展消息
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//打拱游戏专用协议
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//游戏开始
type Msg_S_DG_GameStart struct {
	//MsgTypeGameStart
	BankerUser       uint16   `json:"bankeruser"`       //庄家用户
	MySeat           uint16   `json:"myseat"`           //发给谁的消息
	CurrentUser      uint16   `json:"currentuser"`      //当前牌权用户
	MsgType          byte     `json:"msgtype"`          //类型(开始游戏或是重连）
	CardData         [40]byte `json:"carddata"`         //扑克列表
	CardCount        byte     `json:"cardcount"`        //扑克列表长度
	CardCountSZ      [4]byte  `json:"cardcountsz"`      //所有人的扑克列表长度,每个人的牌数目不一致时需要这个字段
	CurCompleteCount byte     `json:"curcompletecount"` //当前局数
	Whotrust         [4]bool  `json:"whotrust"`         //20191115 苏大强 崇阳 托管
	BankerCard       byte     `json:"bankercard"`       //随机定庄牌
	BirdPlayer       [4]int   `json:"birdplayer"`       //红桃10玩家
	PaiQuanIndex     byte     `json:"paiquanindex"`     //手牌分值排名
}

//玩家手牌数目
type Msg_S_DG_SendCount struct {
	//	MsgTypeSendPaiCount

	CardCount  [4]byte `json:"cardcount"`  //玩家手牌数目
	ActionMask byte    `json:"actionmask"` //动作掩码 暂没使用
}

//权限
type Msg_S_DG_Power struct {
	//	MsgTypeSendPower

	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	Overtime    int64  `json:"overtime"`    //定时器时间 单位秒
	Power       int    `json:"power"`       //什么权限(1吼牌，2打牌，3抢庄)
	Must        byte   `json:"must"`        //3人拱吼牌权限附加信息(0可叫可不叫，1必须叫，2必须不叫，3一王叫或不叫)
}

//出牌命令
type Msg_S_DG_OutCard struct {
	//MsgTypeGameOutCard
	CurrentUser uint16   `json:"currentuser"` //当前牌权用户
	CardCount   byte     `json:"cardcount"`   //出牌扑克数目
	CardData    [40]byte `json:"carddata"`    //扑克列表
	CardType    int      `json:"cardtype"`    //出牌类型
	ByClient    bool     `json:"byClient"`    //
	Overtime    int64    `json:"overtime"`    //超时时间
	OutScorePai [24]byte `json:"outscorepai"` //分牌列表
	CardIndex   [40]int  `json:"cardindex"`   //扑克位置索引
}

//吼牌命令
type Msg_S_DG_Roar struct {
	//MsgTypeRoar
	CurrentUser     uint16 `json:"currentuser"`   //当前牌权用户
	RoarFlag        byte   `json:"roarflag"`      //1表示吼，0表示不吼
	JiaoOrQiangFlag byte   `json:"jiaoqiangflag"` //2表示抢庄，1表示叫庄，0表示忽略或默认
}

//结束吼牌命令
type Msg_S_DG_EndRoar struct {
	//MsgTypeEndRoar
	RoarUser    uint16  `json:"roaruser"`    //吼牌用户 0-3有效，大于等于4无效，表示无人吼牌为2vs2模式
	BankUser    uint16  `json:"bankuser"`    //庄家
	JiaoPai     byte    `json:"jiaopai"`     //叫牌,没有人吼牌时这个变量没有用
	JiaoPaiMate uint16  `json:"jiaopaimate"` //20200525 苏大强 武穴5010k 有鸡牌就告诉大家谁是牌队友
	DownCards3P [6]byte `json:"downcards3p"` //3人拱庄家获得的底牌
	JiaBeiType  int     `json:"jiabeitype"`  // 有抢庄的3人拱的翻倍类型 。定义参考 enum JiabeiType
}

//结束一轮
type Msg_S_DG_EndOut struct {
	//MsgTypeEndOut
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	DataExt     uint16 `json:"dataext"`     //保留字段
}

const (
	TY_KING_BEEP = 1 //拆开天炸的提示
	TY_8Xi_BEEP  = 2 //拆开8喜的提示
	TY_7Xi_BEEP  = 3 //拆开7喜的提示
	TY_JieFeng   = 4 //接风
	TY_ChaDi     = 5 //插底
	TY_AllOut    = 6 //牌出完了
	TY_ChaDi2    = 7 //两张牌报警
)

//播放声音
type Msg_S_DG_PlaySound struct {
	//MsgTypePlaySound
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	SoundType   byte   `json:"soundtype"`   //声音类型 1拆开天炸，2拆开8喜，3拆开7喜,4接风，5插底
}

//玩家分数
type Msg_S_DG_PlayScore struct {
	//MsgTypePlayScore
	PlayScore [4]int `json:"playscore"` //用户分数
	ChairID   uint16 `json:"chairid"`   //本轮分归谁所有
	GetScore  int    `json:"getscore"`  //本轮分数
}

//玩家积分
type Msg_S_DG_PlayJifen struct {
	PlayJiFen [4]int `json:"playjifen"` //用户积分
	ChairID   uint16 `json:"chairid"`   //本轮积分归谁
	GetJifen  [4]int `json:"getscore"`  //本轮积分变化
}

//玩家打出赖子个数
type Msg_S_DG_PlayerOutMagicNum struct {
	OutMaigcNum [4]byte `json:"outmaigicnum"` //打出赖子个数
}

//本轮分数
type Msg_S_DG_TurnScore struct {
	//MsgTypeTurnScore
	TurnScore int `json:"turnscore"` //本轮分数
}

//几游
type Msg_S_DG_SendTurn struct {
	//MsgTypeSendTurn
	Turn [4]byte `json:"turn"` //几游
}

//发送牌给队友
type Msg_S_DG_TeamerPai struct {
	//MsgTypeTeamerPai
	WhoPai    uint16   `json:"whopai"`    //这是谁的牌
	CardCount byte     `json:"cardcount"` //牌数目
	CardData  [40]byte `json:"carddata"`  //牌数据
}

const (
	TY_NULL   = 1 //
	TY_SCORE  = 2 //打分模式，找朋友
	TY_ROAR   = 3 //吼牌模式
	TY_ESCAPE = 4 //有人逃跑
	TY_ACTI   = 5 //反牌
)
const (
	G_Null      = 0  //无效拱型
	G_Bangong   = 1  //半拱
	G_Gong      = 2  //拱
	G_Ganbin    = 3  //干边
	G_Qinggong  = 4  //清拱
	G_Baopai    = 5  //包牌或硬牌
	G_PingJu    = 6  //平局
	G_ChunTian  = 7  //春天
	G_XiaoGuang = 8  //小光
	G_TongZi    = 9  //通子
	G_KuGuang   = 10 //枯光
)

//游戏结束
type Msg_S_DG_GameEnd struct {
	//	MsgTypeGameEnd
	EndType      byte        `json:"endtype"`      //游戏模式 2打分模式，3吼牌，4有人逃跑
	WhoKingBomb  uint16      `json:"whokingbomb"`  //谁有天炸
	FanShu       int         `json:"fanshu"`       //多少番, 吼牌时庄家要X3
	GongType     byte        `json:"gongtype"`     //什么类型结束的，比如清拱、干边,
	LeftPai      [4][40]byte `json:"leftpai"`      //扑克数据
	WinLose      [4]int      `json:"winLose"`      //输还是赢？0输，1赢，2平局
	Have7Xi      [4]int      `json:"have7xi"`      //7喜炸弹个数
	Have8Xi      [4]int      `json:"have8xi"`      //8喜炸弹个数
	Have510K     [4]int      `json:"have510k"`     //510k炸弹个数
	Score        [4]int      `json:"score"`        //得分
	PlayerScore  [4]int      `json:"playerscore"`  //玩家分数
	EndStatus    byte        `json:"endstatus"`    //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
	TotalScore   [4]int      `json:"totalscore"`   //之前的所有局的总积分，比如现在第3局，这个就是1、2局的积分和
	UserVitamin  [4]float64  `json:"user_vitamin"` //玩家剩余疲劳值
	HouPaiChair  uint16      `json:"huopaichair"`  //吼牌玩家
	TheBank      uint16      `json:"thebank"`      //庄家
	TheParter    uint16      `json:"theparter"`    //庄家朋友
	GetScore     [4]int      `json:"getscore"`     //当局抓分
	GongFen      [4]int      `json:"gongfen"`      //当局贡献分
	HandScore    [4]int      `json:"handscore"`    //手分
	TheTurn      [4]byte     `json:"theturn"`      //游次
	XiScore      [4]int      `json:"xiscore"`      //喜钱
	ZongZhaScore [4]int      `json:"zongzhascore"` //总炸
	FaScore      [4]int      `json:"fascore"`      //罚分
	HuapaiScore  [4]int      `json:"huapaiscore"`  //花牌分
	PiaoScore    [4]int      `json:"piaoscore"`    //漂
	ChunTian     [4]int      `json:"chuntian"`     //谁春天或反春了 1表示春天或反春
	ZhuaNiao     [4]int      `json:"zhuaniao"`     //谁被抓了 1表示被抓鸟了
	ExtAddNum    [4][2]int   `json:"extaddnum"`    // 额外加分数据，下标0表示王的数目，下标1表示花牌数目
	ExtAddScore  [4]int      `json:"extaddscore"`  // 额外加分,2王+2花牌或者3王+1花，其他每人给1倍底分,等等
	Have6Xi      [4]int      `json:"have6xi"`      //6喜炸弹个数

	Longzha [4]int `json:"longzha"` //笼炸,分数
	Longshu [4]int `json:"longshu"` //笼数

	TheOrder      byte   `json:"theorder"`     //第几局
	Relink        byte   `json:"relink"`       //是否是断线重连发送的gameend，1表示是断线重连
	TheJiaBeiType int    `json:"jiabeitype"`   //有抢庄的3人拱的加倍类型，一王叫庄加倍，无王叫庄加倍等
	TrustFaScore  [4]int `json:"trustfascore"` //托管承担罚分
	TrustTime     int    `json:"trusttime"`    //托管时间

	//20200521 苏大强 武穴510k反牌玩家
	AntiChair uint16 `json:"antichair"` //反牌玩家
	//20200526 苏大强 武穴510k开局就要知道谁是庄家队友
	JiaoPaiMate uint16 `json:"jiaopaimate"`
	//20200527 苏大强 蕲春打拱，发给客户端
	AutoNextGameTime int `json:"autoNextGameTime"` //自动下一局时间

	//石首510k 总得分达到1000时结束游戏
	TotalPlayerScore [4]int `json:"totalplayerscore"` //总得分

	DataExt         byte   `json:"dataext"`         //保留字段
	MatchState      int    `json:"matchstate"`      //排位赛状态，0没有排位赛，1进行中，2排位赛已经结束
	MatchWinHonors  [4]int `json:"matchwinhonors"`  //玩家在排位赛中，本局获胜额外获得的勋章数
	MatchJoinHonors [4]int `json:"matchjoinhonors"` //玩家在排位赛中，本局参入比赛活动的勋章数

	PaoNum       [4]byte `json:"paonum"`
	PaoStatus    [4]bool `json:"paostatus"`
	WhoJiaoQiang [4]int  `json:"whojiaoqiang"` //叫牌和抢庄过程。定义参考 enum HouState

	//斗地主公共倍数
	BsInit     int `json:"bsinit"`     //初始倍数
	BsMingPai  int `json:"bsmingpai"`  //明牌倍数
	BsQiangDZ  int `json:"bsqiangdz"`  //抢地主倍数
	BsDiPai    int `json:"bsdipai"`    //底牌倍数
	BsBomb     int `json:"bsbomb"`     //炸弹倍数
	BsChunTian int `json:"bschuntian"` //春天倍数
	BsTotal    int `json:"bstotal"`    //公共倍数

	BsDiZhu  int `json:"bsdizhu"`  //地主倍数
	BsFarmer int `json:"bsfarmer"` //农民倍数

	Beishu       [4]int  `json:"beishu"`        // 玩家倍数
	UserBankrupt [4]bool `json:"user_bankrupt"` // 玩家破产标识

	//仙桃跑得快
	RobSpringFlag bool   `json:"robspringflag"` // 抢春是否成功
	WhoRob        uint16 `json:"whorob"`        // 抢春玩家位号

	//监利开机
	GongScore int `json:"gongscore"` //贡献分

	//决战跑得快
	IsFengDing  bool        `json:"isfengding"`  //是否封顶
	BombScore   [4]int      `json:"bombscore"`   //炸弹分
	BirdScore   [4]int      `json:"birdscore"`   //抓鸟分
	LastOutCard [4][40]byte `json:"lastoutcard"` //最后一手牌

	IsEndGame bool    `json:"isendgame"` //是否结束游戏
	UserReady [4]bool `json:"userready"` //玩家准备状态，是否点击了 “继续游戏”按钮
}

type Msg_S_DG_SendGLongCount struct {
	Wchair        int16 `json:"wchair"`        //椅子编号
	TheGLongCount int16 `json:"theglongcount"` //椅子编号对应的人的笼数
}

//游戏状态
type Msg_S_DG_StatusFree struct {
	//MsgTypeGameStatusFree
	CellScore    int    `json:"cellscore"`    //基础金币，底
	CellMinScore int    `json:"cellminscore"` //进入房间最小金币
	CellMaxScore int    `json:"cellmaxscore"` //进入房间最大金币
	FaOfTao      int    `json:"faoftao"`      //逃跑处罚倍数
	BankerUser   uint16 `json:"bankeruser"`   //庄家用户 暂没使用
}

//基础规则
type Msg_S_DG_GameRule struct {
	//MsgTypeGameRule
	CellScore    int    `json:"cellscore"`    //基础金币，底
	CellMinScore int    `json:"cellminscore"` //进入房间最小金币
	CellMaxScore int    `json:"cellmaxscore"` //进入房间最大金币
	FaOfTao      int    `json:"faoftao"`      //逃跑处罚倍数
	BankerUser   uint16 `json:"bankeruser"`   //庄家用户 暂没使用
	GameConfig   string `json:"gameconfig"`   //数据库配置
}

//提示其他玩家，我又回来了
type Msg_S_DG_ReLinkTip struct {
	//MsgTypeReLinkTip
	ReLinkUser uint16 `json:"relinkuser"` //谁断线回来了
	ReLinkTip  uint16 `json:"relinktip"`  //重连标记，目前没有使用
}

//游戏状态
type CMD_S_DG_StatusPlay struct {
	//MsgTypeGameStatusPlay
	//游戏变量
	CellScore   int    `json:"cellscore"`   //单元积分，底
	GameState   byte   `json:"gamestate"`   //游戏状态
	WhoReLink   uint16 `json:"whorelink"`   //谁重连了
	BankerUser  uint16 `json:"bankeruser"`  //庄家用户
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	WhoLastOut  uint16 `json:"wholastout"`  //上一个出牌玩家

	//状态变量
	Overtime     int64  `json:"overtime"`     //超时时间
	TrustCounts  byte   `json:"trustcounts"`  //托管次数
	RoarPai      byte   `json:"roarpai"`      //叫的牌
	WhoRoar      uint16 `json:"whoroar"`      //谁吼了牌
	WhoMJ        uint16 `json:"whomj"`        //谁是明鸡
	MyCardsCount byte   `json:"mycardscount"` //重连用户的手牌长度

	WhoPass         [4]bool     `json:"whopass"`      //谁放弃了
	WhoBreak        [4]bool     `json:"whobreak"`     //谁断线了  //暂时没有使用，客户端通过框架发的消息来确定谁离线了
	TuoGuanPlayer   [4]bool     `json:"whotrust"`     //谁托管了
	WhoReady        [4]bool     `json:"whoready"`     //谁已经完成了叫牌过程
	LastScore       [4]int      `json:"lastscore"`    //上一局输赢分数
	Total           [4]int      `json:"total"`        //总输赢，各小局总计
	Score           [4]int      `json:"score"`        //每个玩家的得分
	LastOutCard     [4][40]byte `json:"lastoutcard"`  //上一次出的牌
	OutCard         [4][40]byte `json:"outcard"`      //出的牌
	MyCards         [40]byte    `json:"mycards"`      //重连用户的牌
	DownCards3P     [6]byte     `json:"downcards3p"`  //3人拱的底牌，6张
	TurnScore       int         `json:"turnscore"`    //本轮分
	LastPaiType     int         `json:"lastpaitype"`  //最后出牌的类型
	TheOrder        byte        `json:"theorder"`     //当前是第几局
	TheJiaBeiType   int         `json:"jiabeitype"`   //有抢庄的3人拱的加倍类型
	WhoJiaoQiang    [4]int      `json:"whojiaoqiang"` //叫牌和抢庄过程。定义参考 enum HouState
	PaoNum          [4]byte     `json:"paonum"`
	PaoStatus       [4]bool     `json:"paostatus"`
	OutScorePai     [24]byte    `json:"outscorepai"`
	CurPunishSeat   int         `json:"curpunishseat"`   // 超时罚分的当前玩家
	PunishStartTime int64       `json:"punishstarttime"` // 超时罚分的开始时间
	//20200521 苏大强 武穴510k 谁反了牌
	WhoAnti     uint16 `json:"whoAnti"`     //
	JiaoPaiMate uint16 `json:"jiaopaimate"` //有鸡牌就告诉大家谁是牌队友
	WhoAntic    [4]int `json:"whoAntic"`    //反牌选择信息（0：放弃；1：选择；-1：没选择）
	//报警状态
	AlermState [4]byte `json:"alermstate"` //报警状态（0没有报警 1单报 2双报）
	//通城打滚
	OutMagicCardNum [4]byte `json:"outmagiccardnum"` //打出赖子牌个数
	//仙桃跑得快
	WhoRob       uint16 `json:"whorob"`       //抢春玩家位号
	WhoRobSpring [4]int `json:"whorobspring"` //玩家抢春操作选择(-1 未操作  0 不抢  1抢春)
	//决战跑得快
	BirdPlayer [4]int `json:"birdplayer"` //红桃10玩家
	BombScore  [4]int `json:"bombscore"`  //炸弹分
	MaxPiao    [4]int `json:"maxpiao"`    //最大漂分
}

//用户托管
type Msg_S_DG_Trustee struct {
	Trustee  bool   `json:"trustee"`  //是否托管
	ChairID  uint16 `json:"chairid"`  //托管用户
	Overtime int64  `json:"overtime"` //超时时间
	Active   bool   `json:"active"`   //主动还是超时
}

const (
	F_NULL    = 0 //无效
	F_JIAOPAI = 1 //叫牌
	F_MINGJI  = 2 //明鸡
)

//明鸡
type Msg_S_DG_Show struct {
	//MsgTypeShow
	MingJiUser uint16 `json:"mingjiuser"` //谁明鸡
	MJCard     byte   `json:"mjcard"`     //叫的哪张牌
	MJFlag     byte   `json:"mjflag"`     //标记
}

// 服务器发给客户端 总结算
type Msg_S_DG_BALANCE_GAME struct {
	Userid         int64  `json:"userid"`         // userid房主
	GameScore      [4]int `json:"gamescore"`      //游戏积分
	MaxGetScore    [4]int `json:"maxgetscore"`    //最高抓分
	FirstTurnCount [4]int `json:"firstturncount"` //一游次数
	RoarCount      [4]int `json:"roarcount"`      //独牌次数
	SixXiCount     [4]int `json:"sixxicount"`     //6喜次数
	SevenXiCount   [4]int `json:"sevenxicount"`   //7喜次数
	EightXiCount   [4]int `json:"eightxicount"`   //8喜次数
	LongZhaCount   [4]int `json:"longzhacount"`   //笼炸次数
	FourKingCount  [4]int `json:"fourkingcount"`  //起手4个以上王的次数
	Same510KCount  [4]int `json:"same510kcount"`  //同色510k的次数
	CurTotalCount  byte   `json:"curtotalcount"`  //当前总局数
	UserEndState   [4]int `json:"userendstate"`   //总结算时玩家状态：0正常，1解散，2离线
	WinCount       [4]int `json:"wincount"`       //胜局次数
	Base           int    `json:"infrastructure"` //底分
	End            byte   `json:"end"`            //0表示结束了的总结算  1表示没有结束，客户端点击查看
	MaxScore       [4]int `json:"maxscore"`       //单局最高得分
	TimeStart      int64  `json:"timestart"`      //游戏开始时间
	TimeEnd        int64  `json:"timeend"`        //游戏结束时间
	MaxLose        [4]int `json:"maxlose"`        //单局最高失分
	PunishCount    [4]int `json:"punishcount"`    //超时罚分的次数计数
	//20200521 苏大强 武穴510k 反牌次数
	AntiCount   [4]int `json:"anticount"`   //反牌次数
	GiveUpCount [4]int `json:"giveupcount"` //投降次数
	//决战跑得快
	BirdCount         [4]int  `json:"birdcount"`         //抓鸟次数
	SpringCount       [4]int  `json:"springcount"`       //被关次数
	PlayerPunishScore [4]int  `json:"playerpunishscore"` //超时罚分
	DiFen             float64 `json:"difen"`
}

//本局选中的游戏任务
type Msg_S_DG_TaskID struct {
	//RandTaskID
	TaskID    int `json:"taskid"`    //选中的任务id
	TaskAward int `json:"taskaward"` //奖励数
}

//游戏任务完成状态
type Msg_S_DG_FinishedTaskID struct {
	//FinishedTaskID
	Player         uint16 `json:"player"`    //谁的任务
	TaskID         int    `json:"taskid"`    //完成了哪个任务id
	TotalTaskAward int    `json:"taskaward"` //奖励数
	Count          int    `json:"count"`     //完成了几次
}

//强退
type Msg_S_ForceExit struct {
	//Id  	uint16 `json:"id"` 	//玩家位置
}

type Msg_S_PlayScore struct {
	PlayScore []int `json:"playscore"` //用户分数
}

//拼三（大吉大利）
//游戏开始
type Msg_S_GameStart_P3 struct {
	BankerUser  uint16 `json:"bankuser"`    //庄家
	CurrentUser uint16 `json:"currentuser"` //当前用户
	Yadi        int    `json:"yadi"`        //压底
}

//下注
type Msg_S_XiaZhu struct {
	Seat      uint16 `json:"seat"`      //座位号
	Score     int    `json:"score"`     //下注数,没有看牌情况下的下注数
	RealScore int    `json:"realscore"` //实际的下注数
	Flag      int    `json:"flag"`      //跟注消息(用于区别跟注和加注或全压)，0标志跟注，1表示加注，2表示全压
	IsOver    bool   `json:"isover"`    //游戏是否结束
}

//弃牌
type Msg_S_QiPai struct {
	Seat   uint16 `json:"seat"`
	IsOver bool   `json:"isover"`
}

//看牌
type Msg_S_KanPai struct {
	Seat     uint16 `json:"seat"`
	IsSystem bool   `json:"issystem"` //系统亮牌之后给未看牌玩家发送的看牌消息
}

//玩家请求亮牌
type Msg_S_LiangPai struct {
	Seat  uint16 `json:"seat"`
	Type  int    `json:"type"`
	Cards []int  `json:"cards"`
}

//轮数更新
type Msg_S_UpdateLunShu struct {
	LunShu byte `json:"lunshu"` //轮数，从1开始
}

//开牌
type Msg_S_OpenCard struct {
	Winner uint16 `json:"winner"` //赢家
	//Loser  []uint16 `json:"loser"`  //输家
}

//系统亮牌
type Msg_S_ShowCard struct {
	//	CT_ERROR = -1,
	//	CT_SINGLE = 0,
	//	CT_DOUBLE = 1,
	//	CT_SHUN_ZI = 2,
	//	CT_TONG_HUA = 3,
	//	GT_TONG_HUA_SUN = 4,
	//	CT_TRIPLE = 5,
	//	CT_SPECIAL = 6,
	//	CT_COUNT = 7
	Seat      uint16            `json:"seat"`
	WinSeat   uint16            `json:"win_seat"`
	CardType  int               `json:"cardtype"`
	CardCount uint16            `json:"cardcount"`
	Cards     [MAX_CARD_P3]byte `json:"cards"` // 牌数据
}

//游戏结束
type Msg_S_P3_GameEnd struct {
	//	MsgTypeGameEnd
	GameType    byte      `json:"gametype"`
	Order       byte      `json:"order"`
	State       []byte    `json:"state"`
	Winner      uint16    `json:"winner"`
	WinnerID    uint16    `json:"winnerid"`
	Score       []int     `json:"score"`
	UserVitamin []float64 `json:"user_vitamin"`
	CardType    byte      `json:"cardtype"`
	Cards       [][]byte  `json:"cards"`
	LoseScore   []int     `json:"losescore"` //输家
}

//发送扑克
type Msg_S_SendPoker struct {
	Lunshu    byte              `json:"lunshu"`    // 发牌次数，一共三轮，1/2/3
	Seat      uint16            `json:"seat"`      // 谁的牌
	Type      int               `json:"type"`      // 牌型
	CardCount byte              `json:"cardcount"` // 牌数目
	Xianshou  byte              `json:"xianshou"`  // 先手
	Bkanpai   bool              `json:"bkanpai"`   // 是否是看牌阶段发牌，1表示是看牌，0表示普通发牌
	Cards     [MAX_CARD_P3]byte `json:"cards"`     // 牌数据
}

//发送牌权
type Msg_S_SendPower struct {
	Seat            uint16 `json:"seat"`            // 谁的权限
	DownTime        int    `json:"downtime"`        // 定时器时间（单位秒）
	Time            int    `json:"time"`            // 整个牌权时间
	Kanpai          bool   `json:"kanpai"`          //是否能看牌
	Qipai           bool   `json:"qipai"`           //是否能弃牌
	Bipai           bool   `json:"bipai"`           //是否能比牌
	Xiazhu          bool   `json:"xianzhu"`         //是否能下注
	Jiazhu          bool   `json:"jiazhu"`          //是否能加注（总体）
	Quanya          bool   `json:"quanya"`          //是否能全压
	NeedScore       []int  `json:"needscore"`       //跟注需要多少分，由于是群发的，非牌权玩家也需要显示自己的跟注分数
	Fending         int    `json:"fending"`         //全压封顶分数，封顶20倍底，但不能超过玩家最小携带分数。
	Itemnum         byte   `json:"itemnum"`         // 加注项实际数目
	Jiabei          []bool `json:"jiabei"`          //是否可以加注（单项）
	Jiabeiscore     []int  `json:"jiabeiscore"`     // 加注需要的分数
	Bycanautogenzhu []bool `json:"bycanautogenzhu"` //是否可以自动跟注
}

//游戏规则
type Msg_S_P3_GameRule struct {
	Yadi     int   `json:"yadi"`           //压底，开始时每个玩家必须投的底
	Base     int   `json:"infrastructure"` // 底
	Charge   int   `json:"charge"`         // 茶水
	Mast     byte  `json:"mast"`           // 游戏规定的闭蒙次数
	MaxRound byte  `json:"maxround"`       // 游戏轮数
	MinScore int   `json:"minscore"`       // 进入房间最小的金币数
	MaxScore int   `json:"maxscore"`       // 进入房间最大的金币数
	Fending  int   `json:"fending"`        //全压封顶分数
	MaxItem  byte  `json:"maxitem"`        // 最大加注项数目
	AddItem  []int `json:"additem"`        // 加注项
	AddItem3 []int `json:"additem3"`       // 第3轮加注项
	ChouMa   []int `json:"chouma"`         // 筹码数值项
	MaxAdd   int   `json:"maxadd"`         //最大单注，
	B235     bool  `json:"b235"`           //是否有235特殊牌型
}

//游戏状态
type CMD_S_P3_StatusPlay struct {
	//MsgTypeGameStatusPlay
	//游戏变量
	GameState   byte   `json:"gamestate"`   //游戏状态
	BankerUser  uint16 `json:"bankeruser"`  //庄家用户
	CurrentUser uint16 `json:"currentuser"` //当前牌权用户
	WholeTime   uint16 `json:"wholetime"`   //牌权时间
	Lunshu      byte   `json:"lunshu"`      //当前轮数
	//	GenzhuScore   int     `json:"genscore"`      //当前跟注分数
	TotalzhuScore int    `json:"totalzhuscore"` //总注
	Xianshou      uint16 `json:"xianshou"`      //先手
	//	Bipai         bool    `json:"bipai"`         //是否在比牌阶段
	CardType  int               `json:"cardtype"`  //牌型
	Bkanpai   []bool            `json:"bkanpai"`   //是否看了牌
	Cardscout []int             `json:"cardscout"` // 每人牌数目
	Cards     [MAX_CARD_P3]byte `json:"cards"`     // 牌数据
	Playstate []int             `json:"playstate"` //玩家状态
	Bipaishu  []bool            `json:"bipaishu"`  //比牌输状态
	Xiazhu    []int             `json:"xiazhu"`    //每个玩家下注的分数
	//	CurrentXiazhu [][]int `json:"currentxiazhu"` //每个玩家当前轮下注的分数
	Totalscore []int  `json:"totalscore"` //玩家总分
	Bqipai     []bool `json:"bqipai"`     //是否弃牌
	Bplaying   []bool `json:"bplaying"`   //是否游戏状态
	Overtime   int64  `json:"overtime"`
	//状态变量
}

//游戏状态
type Msg_S_P3_StatusFree struct {
	//MsgTypeGameStatusFree
	CellScore    int    `json:"cellscore"`    //基础金币，底
	CellMinScore int    `json:"cellminscore"` //进入房间最小金币
	CellMaxScore int    `json:"cellmaxscore"` //进入房间最大金币
	FaOfTao      int    `json:"faoftao"`      //逃跑处罚倍数
	BankerUser   uint16 `json:"bankeruser"`   //庄家用户 暂没使用
}

//玩家分数更新
type Msg_S_UserScore struct {
	Score []int `json:"score"` //分数
}

//拼三 end

//金币麻将begin
//金币麻将游戏状态
type Msg_S_Gold_MJ_StatusFree struct {
	Msg_S_StatusFree
	CellMinScore int `json:"cellminscore"` //进入房间最小金币
	CellMaxScore int `json:"cellmaxscore"` //进入房间最大金币
	FaOfTao      int `json:"faoftao"`      //逃跑处罚倍数
}

//金币麻将游戏状态
type CMD_S_Gold_MJ_StatusPlay struct {
	CMD_S_StatusPlay
	//托管和离线数据
	WhoReLink      uint16  `json:"whorelink"`      //谁重连了
	AutoCardCounts [4]byte `json:"autocardcounts"` //自动出牌的次数
	BreakCounts    [4]byte `json:"breakcounts"`    // 断线次数
	WhoBreak       [4]bool `json:"whobreak"`       //谁断线了  //暂时没有使用，客户端通过框架发的消息来确定谁离线了
	TuoGuanPlayer  [4]bool `json:"whotrust"`       //谁托管了
	TrustCounts    byte    `json:"trustcounts"`    //托管次数
}

//金币麻将end

//字牌游戏，游戏开始
type Msg_S_ZP_GameStart struct {
	//MsgTypeGameStart
	BankerUser       uint16   `json:"bankeruser"`       //庄家用户
	MySeat           uint16   `json:"myseat"`           //发给谁的消息
	CurrentUser      uint16   `json:"currentuser"`      //当前牌权用户
	MsgType          byte     `json:"msgtype"`          //类型(开始游戏或是重连）
	CardData         [40]byte `json:"carddata"`         //扑克列表
	CardCount        byte     `json:"cardcount"`        //扑克列表长度
	JiangCard        byte     `json:"jiangcard"`        //将牌
	SendCard         byte     `json:"sendcard"`         //发给庄家的最后一张牌
	LeftCardCount    byte     `json:"leftcardcount"`    //剩余牌数目
	CurCompleteCount byte     `json:"curcompletecount"` //当前局数
}

//发送牌权
type Msg_S_ZP_SendPower struct {
	CurrentUser uint16 `json:"currentuser"` // 当前出牌的牌权用户
	Status      byte   `json:"status"`      // 保留字段
	LeftTime    int    `json:"lefttime"`    // 倒计时
}

//发送天胡牌权
type Msg_S_ZP_TianHu struct {
	TheSeat     uint16 `json:"theseat"`     // 谁的天胡
	BankerUser  uint16 `json:"bankeruser"`  // 庄家用户
	CurrentUser uint16 `json:"currentuser"` // 当前用户
	UserAction  int    `json:"useraction"`  // 用户动作
	Card        byte   `json:"card"`        // 天胡的哪张牌
	Status      byte   `json:"status"`      // 保留字段
	LeftTime    int    `json:"lefttime"`    //剩余时间
}

//发送天胡牌权
type Msg_S_ZP_TianLong struct {
	TheSeat     uint16     `json:"theseat"`     // 谁的天胡
	BankerUser  uint16     `json:"bankeruser"`  // 庄家用户
	CurrentUser uint16     `json:"currentuser"` // 当前用户
	UserAction  int        `json:"useraction"`  // 用户动作
	CardData    [3][6]byte `json:"carddata"`    // 天拢的哪张牌
	CardCount   [3]byte    `json:"cardcount"`   // 有几张天拢的牌
	Status      byte       `json:"status"`      // 保留字段
}

//发送胡息
type Msg_S_ZP_HuXi struct {
	TheSeat uint16 `json:"theseat"` // 当前出牌的牌权用户
	HuXi    [4]int `json:"huxi"`    // 每个人的胡息,通城个子是4人的
	Tong    [4]int `json:"tong"`    // 每个人的统的次数,荆州花牌
}

//发送胡息
type Msg_S_ZP_HandHuXi struct {
	TheSeat uint16 `json:"theseat"` // 用户
	HuXi    int    `json:"huxi"`    // 用户的胡息
}

//发送下抓数
type Msg_S_ZP_Zhua struct {
	TheSeat uint16 `json:"theseat"` // 用户
	ZhuaNum int    `json:"zhuanum"` // 用户的下抓次数
}

//发送听牌信息
type Msg_S_ZP_TingInfo struct {
	TheSeat  uint16            `json:"theseat"`  // 用户
	TingInfo TagTingCardResult `json:"tinginfo"` // 用户的听牌
}

//发送听牌标签
type Msg_S_ZP_TingTag struct {
	TheSeat     uint16        `json:"theseat"`     // 用户
	TingTagList []TagTingCard `json:"tingtaglist"` // 用户的听牌
}

//发送扑克
type Msg_S_ZP_SendCard struct {
	CurrentUser  uint16 `json:"currentuser"`  // 当前出牌的牌权用户
	CardData     byte   `json:"carddata"`     // 扑克数据
	ActionMask   int    `json:"actionmask"`   // 动作掩码
	ResponseFlag bool   `json:"responseflag"` // 字牌里面，发牌就是出牌，也需要一个字段表明这张牌是否没有人要，1表示有人要
	LeftTime     int    `json:"lefttime"`     //剩余时间
}

//出牌命令
type Msg_S_ZP_OutCard struct {
	OutCardUser  uint16 `json:"outcarduser"`  // 出牌用户
	OutCardData  byte   `json:"outcarddata"`  // 出牌扑克
	ResponseFlag bool   `json:"responseflag"` // 这张牌是否有人要，0表示没有人要，1表示有人要 ，这个字段是为了客户端方便做缩牌动画
	IsAutoOut    byte   `json:"isautoout"`    // 字牌里面，发牌就是出牌，也需要一个字段表明这张牌是否没有人要，1表示有人要
	Jian         bool   `json:"jian"`         // 这张牌是捡的吗?，true是捡的
	OutType      int    `json:"outtype"`      // 1表示出这张牌后要拆句
}

//选择不出牌命令
type Msg_S_ZP_NoOut struct {
	//MsgTypeNoOut
	ChairID uint16 `json:"chairid"` //玩家ID
	Status  int    `json:"status"`  //1表示不出牌
}

//拖牌
type Msg_S_ZP_TuoPai struct {
	//MsgTypeTuoCard
	ChairId    uint16   `json:"chairid"`    //椅子id
	Card       byte     `json:"card"`       //操作扑克
	WeaveCount byte     `json:"weavecount"` //组合数目，不包含牌眼
	WeaveInfo  [10]byte `json:"weaveinfo"`  //组合牌数据，类型对应//摆牌时必须有这两个字段
	Index      int      `json:"index"`
}

//隐藏派遣按钮消息
type Msg_S_ZP_Hide struct {
	//MsgTypeHide
	HideType int `json:"hidetype"` //1表示不出牌
}

//操作提示
type Msg_S_ZP_OperateNotify struct {
	//	enum {XY_ID = SUB_S_OPERATE_NOTIFY};

	ResumeUser uint16 `json:"resumeuser"`  //还原用户
	ActionMask int    `json:"actionmask"`  //动作掩码
	ActionCard byte   `json:"actioncard"`  //动作扑克
	Type       byte   `json:"operatetype"` //操作类型，0不用理会，1表示暗杠，2其它
	LeftTime   int    `json:"lefttime"`    //剩余时间
	MessageID  int    `json:"messageid"`   //消息id，用于服务器校验
	ClockSeat  uint16 `json:"clockseat"`   //闹钟显示位置
}

//操作命令
type Msg_S_ZP_OperateResult struct {
	//enum {XY_ID = SUB_S_OPERATE_RESULT};
	OperateUser    uint16       `json:"operateuser"`    //操作用户
	ProvideUser    uint16       `json:"provideuser"`    //供应用户
	OperateCode    int          `json:"operatecode"`    //操作代码
	OperateCard    byte         `json:"operatecard"`    //操作扑克
	HaveGang       [4]bool      `json:"havegang"`       //是否杠过
	WeaveCount     byte         `json:"weavecount"`     //组合数目，不包含牌眼
	WeaveInfo      [10][10]byte `json:"weaveinfo"`      //组合牌数据，类型对应//摆牌时必须有这两个字段
	WeaveKind      [10]int      `json:"weavekind"`      //组合牌类型，//摆牌时必须有这两个字段
	HasOutPower    byte         `json:"hasoutpower"`    //操作完后是否还有出牌权，0表示没有牌权，1表示有出牌权，正常情况下吃碰后必有出牌权
	Type           byte         `json:"operatetype"`    //操作类型，0不用理会，1表示暗杠，2其它
	LeftTime       int          `json:"lefttime"`       //剩余时间
	Level          int          `json:"level"`          //1 表示后绍，2表示必下抓（播下抓语音
	RemoveCards    []int        `json:"removecards"`    //操作扑克后客户端需要删除的牌，
	LeftCardsCount int          `json:"leftcardscount"` //当前玩家剩余牌数目，给客户端校验用的
}

//游戏状态
type Msg_S_ZP_StatusFree struct {
	//MsgTypeGameStatusFree
	CellScore    int    `json:"cellscore"`    //基础金币，底
	CellMinScore int    `json:"cellminscore"` //进入房间最小金币
	CellMaxScore int    `json:"cellmaxscore"` //进入房间最大金币
	FaOfTao      int    `json:"faoftao"`      //逃跑处罚倍数
	BankerUser   uint16 `json:"bankeruser"`   //庄家用户 暂没使用
}

//游戏状态
type CMD_S_ZP_StatusPlay struct {
	//游戏变量
	CellScore      int    `json:"cellscore"`      //单元积分
	SiceCount      uint16 `json:"sicecount"`      //骰子点数
	BankerUser     uint16 `json:"bankeruser"`     //庄家用户
	CurrentUser    uint16 `json:"currentuser"`    //当前用户，在等待玩家响应时会被置为无效值，
	CurrentOutUser uint16 `json:"currentoutuser"` //当前牌权用户，跟CurrentUser相比不会被置为无效值

	//状态变量
	ActionCard    byte `json:"actioncard"`    //动作扑克
	ActionMask    int  `json:"actionmask"`    //动作掩码
	LeftCardCount byte `json:"leftcardcount"` //剩余数目

	//出牌信息
	OutCardUser   uint16      `json:"outcarduser"`   //出牌用户
	OutCardData   byte        `json:"outcarddata"`   //出牌扑克
	IsOutFlag     byte        `json:"isoutflag"`     //出牌扑克是出牌还是发牌，1表示出牌
	DiscardCount  [3]byte     `json:"discardcount"`  //丢弃数目
	DiscardCard   [3][40]byte `json:"discardcard"`   //丢弃记录
	ToTalOutCount byte        `json:"totaloutcount"` //所有人的总出牌次数
	TuoGuanPlayer [3]bool     `json:"whotrust"`      //谁托管了
	DispatchCard  [3][]int    `json:"dispatchcard"`  //发牌记录，变长数组类型不要用byte，byte类型切片json解析会出问题

	//扑克数据
	CardCount  byte     `json:"cardcount"`  //扑克数目
	CardData   [21]byte `json:"carddata"`   //扑克列表
	CardDataES [40]byte `json:"carddataes"` //扑克列表

	//组合扑克
	WeaveCount        [3]byte               `json:"weavecount"`        //组合数目
	WeaveItemArray    [3][10]TagWeaveItemZP `json:"weaveitemarray"`    //组合扑克
	WeaveItemArrayNew [3][20]TagWeaveItemZP `json:"weaveitemarraynew"` //组合扑克

	JiangCard      byte     `json:"jiangard"`      //将牌
	GameType       byte     `json:"gametype"`      //游戏类型
	GameStatus     int      `json:"gamestatus"`    //游戏状态
	GamEndStatus   byte     `json:"gameendstatus"` //小局游戏结束或进行中状态
	DianPaoPei     byte     `json:"dianpaopei"`    //点炮包赔，其实客户端自己可以解析，
	DianPaoBiHu    byte     `json:"dianpaobihu"`   //点炮必胡，其实客户端自己可以解析，
	TheOrder       byte     `json:"theorder"`      //当前是第几局
	KaiChaoCount   byte     `json:"kaichaocount"`  //玩家开朝次数
	HasOutCard     [10]byte `json:"hasoutcard"`    //对应的开朝是否出过牌,1为出牌
	TotalScore     [4]int   `json:"totalscore"`    //小局积分总计
	ChiOutCards    []int    `json:"chioutcards"`   //吃的牌不能在出了
	PaoNum         [4]byte  `json:"paonum"`
	PaoStatus      [4]bool  `json:"paostatus"`
	TotalScoreInfo [][3]int `json:"totalscoreinfo"` //每小局的结果
	AnFlag         int      `json:"anflag"`         //发牌以外的玩家不知道发的牌是应该明着还是暗着。 1表示要暗着
	MaxPiao        [4]int   `json:"maxpiao"`
	ResponseFlag   bool     `json:"responseflag"` //发的牌或打的牌有没有人响应

	//20181210 苏大强 添加超时信息
	LeftTime  int64   `json:"lefttime"`  //超时时间剩余时间, 预留字段
	TotalTime int64   `json:"totaltime"` //总倒计时时间, 预留字段
	UserReady [4]bool `json:"userready"` //下一局准备好的用户（点击准备好的用户）
}

//游戏结束
type Msg_S_ZP_GameEnd struct {
	//	enum {XY_ID = SUB_S_GAME_END};
	HuangZhuang     bool         `json:"hz"`              // 是否荒庄
	ChiHuCard       byte         `json:"chihucard"`       // 吃胡扑克
	ProvideUser     uint16       `json:"provideuser"`     // 点炮用户
	BankUser        uint16       `json:"bankuser"`        // 庄家用户
	Winner          uint16       `json:"winner"`          // 胡牌玩家
	Score           [3]int       `json:"score"`           // 游戏小局积分
	GameAdjustScore [3]int       `json:"gameadjustscore"` // 修订游戏积分
	TotalScore      [3]int       `json:"totalscore"`      // 总成绩,游戏每小局积分总计
	UserScore       [3]int       `json:"user_score"`      // 玩家剩余积分
	UserVitamin     [4]float64   `json:"user_vitamin"`    // 玩家剩余疲劳值
	WinOrLose       [3]byte      `json:"winorlose"`       // 输赢，0表示输，1表示赢，2表示平
	CardCount       [3]byte      `json:"cardcount"`       // 扑克数目
	CardData        [3][21]byte  `json:"carddata"`        // 扑克数据
	CardDataES      [3][40]byte  `json:"carddataes"`      // 扑克数据
	LeftCardCount   byte         `json:"leftcardcount"`   //剩余数目
	LeftCardData    [80]byte     `json:"leftcarddata"`    // 底牌数据
	HuTypeCount     byte         `json:"hutypecount"`     //剩余数目
	HuType          [20]byte     `json:"hutype"`          // 底牌数据
	WeaveCount      byte         `json:"weavecount"`      //赢家组合数目
	WeaveInfo       [10][10]byte `json:"weaveinfo"`       //赢家组合扑克，胡息对应
	HuXi            [10]int      `json:"huxi"`            //胡息
	WeaveKind       [10]int      `json:"weavekind"`       //赢家组合牌类型
	IsWeave         [10]int      `json:"isweave"`         //每梯的牌是否是碰牌区(组合区)的牌，1表示在碰牌(组合)区
	PublicCard      [10]byte     `json:"publiccard"`      //公开标志，遮罩要用到这张牌
	EyeCard         byte         `json:"eyecard"`         //组合牌数据中的牌眼，可能没有//如果为0表示没有牌眼，不为0时要添加到最后一列
	JiangCard       byte         `json:"jiangcard"`       //翻的将牌
	TotalHuxi       int          `json:"totalhuxi"`       //总胡息
	BeiShu          byte         `json:"beishu"`          //胡牌倍数，所有人都是一样的
	BeiShuFen       int          `json:"beishufen"`       //胡牌倍数对应的分数，所有人都是一样的，新需求要*底分，底分最大100，byte不够用了

	StrEnd         string `json:"strend"`         //胡牌信息
	ChiHuUserCount uint16 `json:"chihuusercount"` // 胡牌总数
	IsQuit         bool   `json:"isquit"`         // 是否强退
	WhoQuit        uint16 `json:"whoquit"`        // 谁强退

	BaseScore    int    `json:"basescore"`    //基础分
	HardChiHu    byte   `json:"hardchihu"`    //标记是否是硬胡
	TheOrder     byte   `json:"theorder"`     //第几局结算
	EndStatus    byte   `json:"endstatus"`    //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
	EndSubType   byte   `json:"endsubtype"`   //结束子类型：0正常(不是自摸也不是放炮)，1自摸，2荒庄，3 放炮
	ExtEndType   int    `json:"extendtype"`   // 拓展结束类型，比如 1 表示鹤峰百胡里面的托管3局结束游戏
	ReLinkFlag   byte   `json:"relinkflag"`   //为1表示断线重连
	ProvideIndex byte   `json:"provideindex"` //供牌的位置
	PaiXingFan   byte   `json:"paixingfan"`   //牌型番数
	HunJiangFan  [3]int `json:"hunjiangfan"`  // 混江番数
	//组合扑克
	TWeaveCount    [3]byte               `json:"tweavecount"`    //组合数目
	WeaveItemArray [3][10]TagWeaveItemZP `json:"weaveitemarray"` //组合扑克
	WeaveHuxi      [3][10]int            `json:"weavehuxi"`      //组合扑克胡息

	WeaveInfoNew [3]TagAnalyseItemZP `json:"weaveinfonew"` //玩家的牌，胜利者胡牌牌型和其它玩家的组合牌区牌型，上面的WeaveInfo和WeaveItemArray合并
	UserReady    [3]bool             `json:"userready"`    //玩家准备状态，是否点击了 “继续游戏”按钮
	PlayerScore  [3]int              `json:"playerscore"`  //玩家分数
}

// 字牌用户托管
type Msg_S_ZP_Trustee struct {
	Trustee  bool   `json:"trustee"`  // 是否托管
	ChairID  uint16 `json:"chairid"`  // 托管用户
	Overtime int64  `json:"overtime"` // 超时时间
}

// 服务器发给客户端 总结算
type Msg_S_ZP_BALANCE_GAME struct {
	Userid           int64  `json:"userid"`           // userid房主
	GameScore        [4]int `json:"gamescore"`        //游戏积分
	ChiHuUserCount   [4]int `json:"chihuusercount"`   //胡牌总数
	ProvideUserCount [4]int `json:"provideusercount"` //点炮次数
	JiePaoUserCount  [4]int `json:"jiepaousercount"`  //接炮次数
	ZiMoUserCount    [4]int `json:"zimousercount"`    //自摸次数
	CurTotalCount    byte   `json:"curtotalcount"`    //当前总局数
	UserEndState     [4]int `json:"userendstate"`     //总结算时玩家状态：0正常，1解散，2离线
	MaxScore         [4]int `json:"maxscore"`         //单局最高得分
	TimeStart        int64  `json:"timestart"`        //游戏开始时间
	TimeEnd          int64  `json:"timeend"`          //游戏结束时间
	MaxHuxi          [4]int `json:"maxhuxi"`          //单局最高胡息
}

//游戏状态
type CMD_S_ZP_StatusPlay_4P struct {
	//游戏变量
	CellScore      int    `json:"cellscore"`      //单元积分
	SiceCount      uint16 `json:"sicecount"`      //骰子点数
	BankerUser     uint16 `json:"bankeruser"`     //庄家用户
	CurrentUser    uint16 `json:"currentuser"`    //当前用户，在等待玩家响应时会被置为无效值，
	CurrentOutUser uint16 `json:"currentoutuser"` //当前牌权用户，跟CurrentUser相比不会被置为无效值

	//状态变量
	ActionCard    byte `json:"actioncard"`    //动作扑克
	ActionMask    int  `json:"actionmask"`    //动作掩码
	LeftCardCount byte `json:"leftcardcount"` //剩余数目

	//出牌信息
	OutCardUser   uint16      `json:"outcarduser"`   //出牌用户
	OutCardData   byte        `json:"outcarddata"`   //出牌扑克
	IsOutFlag     byte        `json:"isoutflag"`     //出牌扑克是出牌还是发牌，1表示出牌
	DiscardCount  [4]byte     `json:"discardcount"`  //丢弃数目
	DiscardCard   [4][40]byte `json:"discardcard"`   //丢弃记录
	ToTalOutCount byte        `json:"totaloutcount"` //所有人的总出牌次数
	TuoGuanPlayer [4]bool     `json:"whotrust"`      //谁托管了
	DispatchCard  [4][]int    `json:"dispatchcard"`  //发牌记录，变长数组类型不要用byte，byte类型切片json解析会出问题

	//扑克数据
	CardCount  byte     `json:"cardcount"`  //扑克数目
	CardData   [21]byte `json:"carddata"`   //扑克列表
	CardDataES [40]byte `json:"carddataes"` //扑克列表

	//组合扑克
	WeaveCount        [4]byte               `json:"weavecount"`        //组合数目
	WeaveItemArray    [4][10]TagWeaveItemZP `json:"weaveitemarray"`    //组合扑克
	WeaveItemArrayNew [4][20]TagWeaveItemZP `json:"weaveitemarraynew"` //组合扑克

	JiangCard      byte     `json:"jiangard"`      //将牌
	GameType       byte     `json:"gametype"`      //游戏类型
	GameStatus     int      `json:"gamestatus"`    //游戏状态
	GamEndStatus   byte     `json:"gameendstatus"` //小局游戏结束或进行中状态
	DianPaoPei     byte     `json:"dianpaopei"`    //点炮包赔，其实客户端自己可以解析，
	DianPaoBiHu    byte     `json:"dianpaobihu"`   //点炮必胡，其实客户端自己可以解析，
	TheOrder       byte     `json:"theorder"`      //当前是第几局
	KaiChaoCount   byte     `json:"kaichaocount"`  //玩家开朝次数
	HasOutCard     [10]byte `json:"hasoutcard"`    //对应的开朝是否出过牌,1为出牌
	TotalScore     [4]int   `json:"totalscore"`    //小局积分总计
	ChiOutCards    []int    `json:"chioutcards"`   //吃的牌不能在出了
	PaoNum         [4]byte  `json:"paonum"`
	PaoStatus      [4]bool  `json:"paostatus"`
	TotalScoreInfo [][4]int `json:"totalscoreinfo"` //每小局的结果
	AnFlag         int      `json:"anflag"`         //发牌以外的玩家不知道发的牌是应该明着还是暗着。 1表示要暗着
	MaxPiao        [4]int   `json:"maxpiao"`
	UserReady      [4]bool  `json:"userready"` //下一局准备好的用户（点击准备好的用户）

	//20181210 苏大强 添加超时信息
	LeftTime  int64 `json:"lefttime"`  //超时时间剩余时间, 预留字段
	TotalTime int64 `json:"totaltime"` //总倒计时时间, 预留字段
}

//游戏结束
type Msg_S_ZP_GameEnd_4P struct {
	//	enum {XY_ID = SUB_S_GAME_END};
	HuangZhuang     bool         `json:"hz"`              // 是否荒庄
	ChiHuCard       byte         `json:"chihucard"`       // 吃胡扑克
	ProvideUser     uint16       `json:"provideuser"`     // 点炮用户
	BankUser        uint16       `json:"bankuser"`        // 庄家用户
	Winner          uint16       `json:"winner"`          // 胡牌玩家
	Score           [4]int       `json:"score"`           // 游戏小局积分
	GameAdjustScore [4]int       `json:"gameadjustscore"` // 修订游戏积分
	TotalScore      [4]int       `json:"totalscore"`      // 总成绩,游戏每小局积分总计
	UserScore       [4]int       `json:"user_score"`      // 玩家剩余积分
	UserVitamin     [4]float64   `json:"user_vitamin"`    // 玩家剩余疲劳值
	WinOrLose       [4]byte      `json:"winorlose"`       // 输赢，0表示输，1表示赢，2表示平
	CardCount       [4]byte      `json:"cardcount"`       // 扑克数目
	CardData        [4][21]byte  `json:"carddata"`        // 扑克数据
	CardDataES      [4][40]byte  `json:"carddataes"`      // 扑克数据
	LeftCardCount   byte         `json:"leftcardcount"`   //剩余数目
	LeftCardData    [80]byte     `json:"leftcarddata"`    // 底牌数据
	HuTypeCount     byte         `json:"hutypecount"`     //剩余数目
	HuType          [20]byte     `json:"hutype"`          // 底牌数据
	WeaveCount      byte         `json:"weavecount"`      //赢家组合数目
	WeaveInfo       [10][10]byte `json:"weaveinfo"`       //赢家组合扑克，胡息对应
	HuXi            [10]int      `json:"huxi"`            //胡息
	WeaveKind       [10]int      `json:"weavekind"`       //赢家组合牌类型
	IsWeave         [10]int      `json:"isweave"`         //每梯的牌是否是碰牌区(组合区)的牌，1表示在碰牌(组合)区
	PublicCard      [10]byte     `json:"publiccard"`      //公开标志，遮罩要用到这张牌
	EyeCard         byte         `json:"eyecard"`         //组合牌数据中的牌眼，可能没有//如果为0表示没有牌眼，不为0时要添加到最后一列
	JiangCard       byte         `json:"jiangcard"`       //翻的将牌
	TotalHuxi       int          `json:"totalhuxi"`       //总胡息
	BeiShu          byte         `json:"beishu"`          //胡牌倍数，所有人都是一样的
	BeiShuFen       int          `json:"beishufen"`       //胡牌倍数对应的分数，所有人都是一样的，新需求要*底分，底分最大100，byte不够用了

	StrEnd         string `json:"strend"`         //胡牌信息
	ChiHuUserCount uint16 `json:"chihuusercount"` // 胡牌总数
	IsQuit         bool   `json:"isquit"`         // 是否强退
	WhoQuit        uint16 `json:"whoquit"`        // 谁强退

	BaseScore    int    `json:"basescore"`    //基础分
	HardChiHu    byte   `json:"hardchihu"`    //标记是否是硬胡
	TheOrder     byte   `json:"theorder"`     //第几局结算
	EndStatus    byte   `json:"endstatus"`    //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
	EndSubType   byte   `json:"endsubtype"`   //结束子类型：0正常(不是自摸也不是放炮)，1自摸，2荒庄，3 放炮
	ExtEndType   int    `json:"extendtype"`   // 拓展结束类型，比如 1 表示鹤峰百胡里面的托管3局结束游戏
	ReLinkFlag   byte   `json:"relinkflag"`   //为1表示断线重连
	ProvideIndex byte   `json:"provideindex"` //供牌的位置
	PaiXingFan   byte   `json:"paixingfan"`   //牌型番数
	HunJiangFan  [4]int `json:"hunjiangfan"`  // 混江番数
	//组合扑克
	TWeaveCount    [4]byte               `json:"tweavecount"`    //组合数目
	WeaveItemArray [4][10]TagWeaveItemZP `json:"weaveitemarray"` //组合扑克
	WeaveHuxi      [4][10]int            `json:"weavehuxi"`      //组合扑克胡息

	WeaveInfoNew [4]TagAnalyseItemZP `json:"weaveinfonew"` //玩家的牌，胜利者胡牌牌型和其它玩家的组合牌区牌型，上面的WeaveInfo和WeaveItemArray合并
	PlayerScore  [4]int              `json:"playerscore"`  //玩家分数
}

type Msg_S_DG_SplitCardsStart struct {
	CardCount int    `json:"cardcount"` //总牌数
	SplitUser uint16 `json:"splituser"` //切牌用户
	LeftTime  int    `json:"lefttime"`  //切牌倒计时时间
}
type Msg_S_DG_SplitCards struct {
	CardCount   int    `json:"cardcount"`   //总牌数
	SplitUser   uint16 `json:"splituser"`   //切牌用户
	Cardindex   int    `json:"cardindex"`   //切牌位置
	SplitStatus bool   `json:"splitstatus"` //是否完成了切牌
	SplitType   int    `json:"splitstype"`  //切牌类型，0手动，1超时自动
}

//////////////begin:升级////////////////
//游戏开始
type Msg_S_SJ_GameStart struct {
	//MsgTypeGameStart
	BankerUser  uint16  `json:"wbankeruser"`  //庄家用户
	PackCount   byte    `json:"cbpackcount"`  //副数数目
	MainValue   byte    `json:"cbmainvalue"`  //主牌数值
	ValueOrder  [2]byte `json:"cbvalueorder"` //双方级数，座位号0/2玩家的下标为0，座位号1/3玩家的下标为1
	ZhuangSpace [4]int  `json:"zhuangspace"`  //下局必打提示,3时要提示下局必打
}

//发送级数和庄家
type Msg_S_SJ_SendOrder struct {
	//MsgTypeSendOrder
	BankerUser uint16  `json:"wbankeruser"`  //庄家用户
	PackCount  byte    `json:"cbpackcount"`  //副数数目
	MainValue  byte    `json:"cbmainvalue"`  //主牌数值
	ValueOrder [2]byte `json:"cbvalueorder"` //双方级数，座位号0/2玩家的下标为0，座位号1/3玩家的下标为1
}

//发送扑克
type Msg_S_SJ_SendCard struct {
	//MsgTypeGameSendCard
	CardCount byte     `json:"cbcardcount"` //扑克数目
	CardData  [33]byte `json:"cbcarddata"`  //扑克列表
}

//用户叫牌
type Msg_S_SJ_CallCard struct {
	//MsgTypeCallCard
	BankerUser   uint16 `json:"wbankeruser"`   //庄家用户
	CallCardUser uint16 `json:"wcallcarduser"` //叫牌用户
	CardCount    byte   `json:"cbcardcount"`   //叫牌数目
	CallCard     byte   `json:"cbcallcard"`    //叫牌扑克
	LeftTime     int64  `json:"thelefttime"`   //剩余时间
}

//底牌扑克
type Msg_S_SJ_SendConceal struct {
	//MsgTypeConcealCard
	BankerUser       uint16   `json:"wbankeruser"`    //庄家用户
	CallCardUser     uint16   `json:"wcallcarduser"`  //叫牌用户
	CurrentUser      uint16   `json:"wcurrentuser"`   //当前用户
	MainValue        byte     `json:"cbmainvalue"`    //主牌数值
	MainColor        byte     `json:"cbmaincolor"`    //主牌花色数值
	ConcealCount     byte     `json:"cbconcealcount"` //底牌数目
	ConcealCard      [8]byte  `json:"cbconcealcard"`  //底牌扑克
	LeftTime         int64    `json:"thelefttime"`    //剩余时间
	ZhuangSpace      [4]int   `json:"zhuangspace"`    //每个玩家3下局必打或4当前罚分
	ConcealMainCards [4][]int `json:"ccmaincard"`     //每个玩家手中对应底牌的主牌
}

//游戏开始
type Msg_S_SJ_GamePlay struct {
	//MsgTypeGamePlay
	CurrentUser  uint16  `json:"wcurrentuser"`   //当前用户
	LeftTime     int64   `json:"thelefttime"`    //剩余时间
	ConcealCount byte    `json:"cbconcealcount"` //底牌数目
	ConcealCard  [8]byte `json:"cbconcealcard"`  //底牌扑克
	CardsCount   [4]byte `json:"cardcount"`      //玩家当前手牌数
}

//用户出牌
type Msg_S_SJ_OutCard struct {
	//MsgTypeGameOutCard
	OutCardUser uint16   `json:"woutcarduser"` //出牌玩家,就是上家
	BiggerUser  uint16   `json:"wbiggeruer"`   //当前最大出牌玩家
	CurrentUser uint16   `json:"wcurrentuser"` //当前用户，新牌权玩家
	LeftTime    int64    `json:"thelefttime"`  //剩余时间
	CardCount   byte     `json:"cbcardcount"`  //扑克数目
	CardData    [33]byte `json:"cbcarddata"`   //扑克列表
	SoundLevel  byte     `json:"cbsoundlevel"` //声音,0客户端决定，1甩牌，2大你，3再大你，4我再大你，5拖拉机，6超长拖拉机，7毙了，8在杀，9我在杀，10垫了,11偏三轮
	OutResult   byte     `json:"theoutresult"` //出牌结果，0表示出牌成功，1表示甩牌失败，
	LastTurn    byte     `json:"thelastturn"`  //是否最后一轮，最后一轮不需要语音
	CardsCount  [4]byte  `json:"cardcount"`    //玩家当前手牌数
}

//甩牌结果
type Msg_S_SJ_ThrowResult struct {
	//MsgTypeThrowResult
	OutCardUser     uint16   `json:"woutcarduser"`      //出牌玩家,就是上家
	CurrentUser     uint16   `json:"wcurrentuser"`      //当前用户，新牌权玩家
	LeftTime        int64    `json:"thelefttime"`       //剩余时间
	ThrowCardCount  byte     `json:"cbthrowcardcount"`  //扑克数目
	ResultCardCount byte     `json:"cbresultcardcount"` //扑克数目
	CardDataArray   [33]byte `json:"cbcarddataarray"`   //扑克数组
}

//一轮统计
type Msg_S_SJ_TurnBalance struct {
	//MsgTypeTurnBalance
	TurnWinner     uint16   `json:"wTurnWinner"`      //一轮胜者
	CurrentUser    uint16   `json:"wcurrentuser"`     //当前用户，新牌权玩家
	LeftTime       int64    `json:"thelefttime"`      //剩余时间
	TotalScore     int      `json:"cbtotalscore"`     //总得分
	Score          int      `json:"cbwcore"`          //得分
	ScoreCardCount byte     `json:"cbscorecardcount"` //扑克数目
	ScoreCardData  [33]byte `json:"cbscorecarddata"`  //得分扑克
	LastTurn       byte     `json:"thelastturn"`      //是否最后一轮，最后一轮可能需要增加延时
	ShowAllScore   bool     `json:"showallsocre"`     //是否显示闲家总抓分
}

//发送上一轮的数据，目前没有使用，客户端使用自己保存的数据了
type Msg_S_SJ_LastTurn struct {
	//MsgTypeLastTurn
	LastFirstOutUser uint16   `json:"wLastFirstoutuser"`  //上一轮的首出牌玩家
	LastOutCardCount byte     `json:"cblastoutcardcount"` //上一轮出牌数目
	LastOutCardData  [33]byte `json:"cbLastoutcarddata"`  //上一轮出牌列表
}

//额外的叫主等待时间
type Msg_S_SJ_HaveExtCallTime struct {
	//MsgTypeHaveExtCallTime
	ExtCallTime byte `json:"byextcalltime"` //额外的叫主等待时间
}

//游戏状态
type Msg_S_SJ_StatusPlay struct {
	//MsgTypeGameStatusPlay
	//游戏变量
	CellScore        int    `json:"cellscore"`         //单元积分，底
	GameState        byte   `json:"thegamestate"`      //游戏状态
	WhoRelink        uint16 `json:"wwhorelink"`        //谁重连了
	BankerUser       uint16 `json:"wbankeruser"`       //庄家用户
	CurrentUser      uint16 `json:"wcurrentuser"`      //当前牌权用户
	BiggerUser       uint16 `json:"wbiggeruser"`       //当前最大的出牌用户，0-3,0xff表示这轮还没有人出牌，没有最大用户
	FirstOutUser     uint16 `json:"wfirstoutuser"`     //首出牌用户
	LastFirstOutUser uint16 `json:"wlastfirstoutuser"` //上一轮首出牌用户
	CallCardUser     uint16 `json:"wcallcarduser"`     //叫牌用户

	//状态变量
	LeftTime     int64 `json:"thelefttime"`    //超时时间
	TrustCounts  byte  `json:"thetgtime"`      //托管次数
	CallCard     byte  `json:"cbcallcard"`     //叫牌扑克
	CallCount    byte  `json:"cbcallcount"`    //叫牌数目
	MainValue    byte  `json:"cbmainvalue"`    //当前级数
	MainColor    byte  `json:"cbmaincolor"`    //当前主级牌花色
	MyCardsCount byte  `json:"cbcardcount"`    //重连用户的牌数目
	ConcealCount byte  `json:"cbconcealcount"` //底牌数目

	WhoBreak          [4]bool     `json:"whobreak"`            //谁断线了  //暂时没有使用，客户端通过框架发的消息来确定谁离线了
	TuoGuanPlayer     [4]bool     `json:"thetuoguan"`          //谁托管了
	LastScore         [4]int      `json:"lastscore"`           //上一局输赢分数
	Total             [4]int      `json:"total"`               //总输赢，各小局总计
	WhoKill           [4]byte     `json:"cbwhokill"`           //谁杀牌了，暂时没有使用
	TotalOutCardCount [4]byte     `json:"cbtotaloutcardcount"` //已出扑克总计,用来统计是否可以甩牌
	TotalOutCardData  [108]byte   `json:"cbtotaloutcarddata"`  //已出扑克列表,用来统计是否可以甩牌
	LastOutCardCount  [4]byte     `json:"cblastoutcardcount"`  //当前轮出牌数目
	LastOutCardData   [4][33]byte `json:"cblastoutcarddata"`   //上一轮出牌列表
	MyCards           [36]byte    `json:"cbcarddata"`          //重连用户的牌
	ConcealCard       [8]byte     `json:"cbconcealcard"`       //底牌
	OutCardCount      [4]byte     `json:"cboutcardcount"`      //当前轮出牌数目
	OutCardData       [4][33]byte `json:"cboutcarddata"`       //当前轮出牌列表
	ScoreCardData     [36]byte    `json:"cbscorecarddata"`     //得分扑克
	ScoreCardCount    byte        `json:"cbscorecardcount"`    //得分扑克数目
	Score             int         `json:"score"`               //得分
	TheOrder          byte        `json:"theorder"`            //当前是第几局
	NTValue           byte        `json:"ntvalue"`             //常主
	BankerCallScore   int         `json:"bankercallscore"`     //叫分结束后庄家叫的分数
	CallScore         [4]int      `json:"callscore"`           //叫的分数
	PaiOutAll         [4][33]byte `json:"paioutall"`           //之前出的所有牌
	ZhuangSpace       [4]int      `json:"zhuangspace"`         //每个玩家3下局必打或4当前罚分
	ConcealMainCards  [4][]int    `json:"ccmaincard"`          //每个玩家手中对应底牌的主牌
	MainNoSameFlag    bool        `json:"mainnosameflag"`      //是否有3个玩家主牌没有对子
}

//游戏结束
type Msg_S_SJ_GameEnd struct {
	//	MsgTypeGameEnd
	EndType      byte    `json:"endtype"`        //
	BankerUser   uint16  `json:"wbankeruser"`    //庄家用户
	ConcealCount byte    `json:"cbconcealcount"` //底牌数目
	ConcealCard  [8]byte `json:"cbconcealcard"`  //底牌扑克

	TotalScore   [4]int  `json:"totalscore"`    // 总成绩,游戏每小局积分总计
	Score        [4]int  `json:"lscore"`        //用户得分
	GameScore    int     `json:"wgamescore"`    //游戏得分 满分200
	PaiMianScore int     `json:"wpaimianscore"` //牌面分，加扣底分之前的分数
	ConcealTime  int     `json:"wconcealtime"`  //扣底倍数
	ConcealScore int     `json:"wconcealscore"` //底牌积分
	KouDiScore   int     `json:"wkoudiscore"`   //扣底分，底牌积分乘以扣底倍数
	PlayerScore  [4]int  `json:"playerscore"`   //玩家分数
	ShengJiShu   int     `json:"cshengjishu"`   //正数是庄家方升，负数是闲家方升
	MainValue    byte    `json:"cbmainvalue"`   //主牌数值
	ValueOrder   [2]byte `json:"cbvalueorder"`  //主牌数值，座位号0/2玩家的下标为0，座位号1/3玩家的下标为1
	WinLoseType  [4]byte `json:"cbwinlosetype"` //双方输赢类型，参考 GameOverType
	WinLose      [4]byte `json:"winlose"`       //输赢，0表示输，1表示赢,2表示平
	DataExt      byte    `json:"dataext"`       //保留字段

	MatchState      int    `json:"matchstate"`      //排位赛状态，0没有排位赛，1进行中，2排位赛已经结束
	MatchWinHonors  [4]int `json:"matchwinhonors"`  //玩家在排位赛中，本局获胜额外获得的勋章数
	MatchJoinHonors [4]int `json:"matchjoinhonors"` //玩家在排位赛中，本局参入比赛活动的勋章数

	TheOrder       byte        `json:"theorder"`      //第几局结算
	EndStatus      byte        `json:"endstatus"`     //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
	ReLink         byte        `json:"relink"`        //为1表示断线重连
	UserVitamin    [4]float64  `json:"user_vitamin"`  //玩家剩余疲劳值
	ZhuangSpace    [4]int      `json:"zhuangspace"`   //每个玩家3下局必打或4当前罚分
	LeftCards      [4][36]byte `json:"leftcards"`     //用户的牌
	LeftCardsCount [4]byte     `json:"leftcarscount"` //用户的牌数目
	TrustFaScore   [4]int      `json:"trustfascore"`  //托管承担罚分
	BankerStatus   int         `json:"bankerstatus"`  //庄家状态，是否巴锅或买七
	BankerPartner  uint16      `json:"bankerpartner"` //庄家队友
}

//基础规则
type Msg_S_SJ_GameRule struct {
	//MsgTypeGameRule
	CellScore    int     `json:"cellscore"`    //基础金币，底
	SerPay       int     `json:"serpay"`       //茶水
	CellMinScore int     `json:"cellminscore"` //进入房间最小金币
	CellMaxScore int     `json:"cellmaxscore"` //进入房间最大金币
	FaOfTao      int     `json:"faoftao"`      //逃跑处罚倍数
	BankerUser   uint16  `json:"bankeruser"`   //庄家用户 暂没使用
	TrustCounts  byte    `json:"ntrustcounts"` //允许托管的次数
	MainValue    byte    `json:"cbmainvalue"`  //当前级数
	MainColor    byte    `json:"cbmaincolor"`  //当前级数
	ValueOrder   [2]byte `json:"cbvalueorder"` //主牌数值，座位号0/2玩家的下标为0，座位号1/3玩家的下标为1
}

//////////////end:升级////////////////

//用户叫分
type Msg_S_SJ_CallScore struct {
	//MsgTypeCallCard
	BankerUser   uint16 `json:"wbankeruser"`   //庄家用户
	CallCardUser uint16 `json:"wcallcarduser"` //叫牌用户
	CallScore    int    `json:"callscore"`     //叫分
}

//展示主牌和副牌数量
type Msg_S_SJ_ShowMainCards struct {
	//MsgTypeShowMainCards
	MainFlag      [4]bool     `json:"mainflag"`      //是否展示主牌,true展示，false不展示
	MainCards     [4][33]byte `json:"maincards"`     //主牌牌值 ，
	NMainFlag     [4]bool     `json:"nmainflag"`     //是否展示副牌数目,true展示，false不展示
	NCardsCnt     [4]byte     `json:"ncardscnt"`     //副牌数量
	SMainCardsCnt [4]byte     `json:"smaincardscnt"` //每个玩家主牌数量
}

//游戏状态
type CMD_S_ZP_TC_StatusPlay struct {
	//游戏变量
	CellScore      int     `json:"cellscore"`      //单元积分
	SiceCount      uint16  `json:"sicecount"`      //骰子点数
	BankerUser     uint16  `json:"bankeruser"`     //庄家用户
	CurrentUser    uint16  `json:"currentuser"`    //当前用户，在等待玩家响应时会被置为无效值，
	CurrentOutUser uint16  `json:"currentoutuser"` //当前牌权用户，跟CurrentUser相比不会被置为无效值
	NoOutUser      uint16  `json:"nooutuser"`      //选择不出牌的用户，在等待玩家响应时会被置为无效值，
	TuoGuanPlayer  [4]bool `json:"thetuoguan"`     //谁托管了

	//状态变量
	ActionCard    byte `json:"actioncard"`    //动作扑克
	ActionMask    int  `json:"actionmask"`    //动作掩码
	LeftCardCount byte `json:"leftcardcount"` //剩余数目

	//出牌信息
	OutCardUser   uint16      `json:"outcarduser"`   //出牌用户
	OutCardData   byte        `json:"outcarddata"`   //出牌扑克
	IsOutFlag     byte        `json:"isoutflag"`     //出牌扑克是出牌还是发牌，1表示出牌
	DiscardCount  [4]byte     `json:"discardcount"`  //丢弃数目
	DiscardCard   [4][40]byte `json:"discardcard"`   //丢弃记录
	ToTalOutCount byte        `json:"totaloutcount"` //所有人的总出牌次数

	//扑克数据
	CardCount          byte      `json:"cardcount"`       //扑克数目
	CardData           [20]byte  `json:"carddata"`        //扑克列表
	CardDataJZ         [60]byte  `json:"carddatajz"`      //扑克列表,荆州花牌
	UserGuanCards      [10][]int `json:"guancarddata"`    // 玩家观生的牌
	UserJianCards      []int     `json:"jiancarddata"`    // 玩家捡的牌
	UserGuanCardsCount byte      `json:"guancardcount"`   // 玩家观生的牌数目
	UserJianCardsCur   []int     `json:"jiancarddatacur"` // 玩家这次捡的牌

	//组合扑克
	WeaveCount     [4]byte               `json:"weavecount"`     //组合数目
	WeaveItemArray [4][10]TagWeaveItemZP `json:"weaveitemarray"` //组合扑克

	JiangCard    byte     `json:"jiangard"`      //将牌
	GameType     byte     `json:"gametype"`      //游戏类型
	GameStatus   int      `json:"gamestatus"`    //游戏状态
	GamEndStatus byte     `json:"gameendstatus"` //小局游戏结束或进行中状态
	DianPaoPei   byte     `json:"dianpaopei"`    //点炮包赔，其实客户端自己可以解析，
	DianPaoBiHu  byte     `json:"dianpaobihu"`   //点炮必胡，其实客户端自己可以解析，
	TheOrder     byte     `json:"theorder"`      //当前是第几局
	KaiChaoCount byte     `json:"kaichaocount"`  //玩家开朝次数
	HasOutCard   [10]byte `json:"hasoutcard"`    //对应的开朝是否出过牌,1为出牌
	TotalScore   [4]int   `json:"totalscore"`    //小局积分总计
	PaoNum       [4]byte  `json:"paonum"`
	PaoStatus    [4]bool  `json:"paostatus"`
	DispTongCnt  [4]int   `json:"disptongcnt"`
	MyTongCnt    [40]int  `json:"mytongcnt"`
	SendGangCard byte     `json:"sendgangcard"` //发的牌是刚刚招的牌，ActionMask有滑时客户端需要在组合牌区找滑，0表示不用找
	MaxPiao      [4]int   `json:"maxpiao"`
	MessageID    int      `json:"messageid"`
	UserReady    [4]bool  `json:"userready"` //下一局准备好的用户（点击准备好的用户）

	//20181210 苏大强 添加超时信息
	LeftTime  int64 `json:"lefttime"`  //超时时间剩余时间, 预留字段
	TotalTime int64 `json:"totaltime"` //总倒计时时间, 预留字段
}

//游戏结束
type Msg_S_ZP_TC_GameEnd struct {
	//	enum {XY_ID = SUB_S_GAME_END};
	HuangZhuang     bool        `json:"hz"`              // 是否荒庄
	ChiHuCard       byte        `json:"chihucard"`       // 吃胡扑克
	ProvideUser     uint16      `json:"provideuser"`     // 点炮用户
	BankUser        uint16      `json:"bankuser"`        // 庄家用户
	Score           [4]int      `json:"score"`           // 游戏小局积分
	GameAdjustScore [4]int      `json:"gameadjustscore"` // 修订游戏积分
	TotalScore      [4]int      `json:"totalscore"`      // 总成绩,游戏每小局积分总计
	UserScore       [4]int      `json:"user_score"`      // 玩家剩余积分
	UserVitamin     [4]float64  `json:"user_vitamin"`    // 玩家剩余疲劳值
	WinOrLose       [4]byte     `json:"winorlose"`       // 输赢，0表示输，1表示赢，2表示平
	ChiHuKind       [4]int      `json:"chihukind"`       //吃胡类型
	CardCount       [4]byte     `json:"cardcount"`       // 扑克数目
	CardData        [4][20]byte `json:"carddata"`        // 扑克数据
	CardDataJZ      [4][60]byte `json:"carddatajz"`      // 扑克数据,荆州花牌
	LeftCardCount   byte        `json:"leftcardcount"`   //剩余数目
	LeftCardData    [110]byte   `json:"leftcarddata"`    // 底牌数据

	//赢家信息，存在一炮多响，所有会有多个赢家
	WeaveCount [4]byte         `json:"weavecount"` //组合数目
	WeaveInfo  [4][10][10]byte `json:"weaveinfo"`  //组合扑克，胡息对应
	HuXi       [4][10]int      `json:"huxi"`       //胡息
	WeaveKind  [4][10]int      `json:"weavekind"`  //组合牌类型
	IsWeave    [4][10]int      `json:"isweave"`    //每梯的牌是否是碰牌区(组合区)的牌，1表示在碰牌(组合)区
	EyeCard    [4]byte         `json:"eyecard"`    //组合牌数据中的牌眼，可能没有//如果为0表示没有牌眼，不为0时要添加到最后一列

	TotalHuxi      [4]int    `json:"totalhuxi"`      //总胡息
	HuaPaiShu      [4]int    `json:"huapaishu"`      //花牌数
	HuaPaiInfo     [4][5]int `json:"huapaiinfo"`     //花牌详情，花乙到花九共5张牌
	StrEnd         string    `json:"strend"`         //胡牌信息
	ChiHuUserCount uint16    `json:"chihuusercount"` // 胡牌总数
	IsQuit         bool      `json:"isquit"`         // 是否强退
	WhoQuit        uint16    `json:"whoquit"`        // 谁强退

	BaseScore    byte `json:"basescore"`    //基础分
	HardChiHu    byte `json:"hardchihu"`    //标记是否是硬胡
	TheOrder     byte `json:"theorder"`     //第几局结算
	EndStatus    byte `json:"endstatus"`    //表示是哪种结束方式：0正常 1游戏解散 2玩家强退
	EndSubType   byte `json:"endsubtype"`   //结束子类型：0正常(不是自摸也不是放炮)，1自摸，2荒庄，3 放炮
	ReLinkFlag   byte `json:"relinkflag"`   //为1表示断线重连
	ProvideIndex byte `json:"provideindex"` //供牌的位置
	//组合扑克，输家信息,或流局
	TWeaveCount    [4]byte               `json:"tweavecount"`    //组合数目
	WeaveItemArray [4][10]TagWeaveItemZP `json:"weaveitemarray"` //组合扑克
	//观和捡的牌，输家信息,或流局
	UserGuanCards      [4][10][5]byte `json:"guancarddata"`  // // 玩家观生的牌
	UserJianCards      [4][]int       `json:"jiancarddata"`  // 玩家捡的牌
	UserGuanCardsCount [4]byte        `json:"guancardcount"` // 玩家观生的牌数目
	JingPai            byte           `json:"jingpai"`       //显示精牌，0表示没有精牌
	PlayerScore        [4]int         `json:"playerscore"`   //玩家分数
}

//游戏报警
type Msg_S_GameAlerm struct {
	BaoJiangType [4]byte `json:"baojingype"` //报警类型
}

//罚时提醒
type Msg_S_StartPunish struct {
	PunishPlayer uint16 `json:"punishplayer"` //罚时玩家
	StartTime    int64  `json:"starttime"`    //罚时开始时间
	StartOrStop  int32  `json:"startorstop"`  //开始or结束
}

type Msg_S_OperateScore struct {
	OperateUser uint16     `json:"operateuser"`  //操作用户
	OperateType uint16     `json:"operatetype"`  //操作类型
	GameScore   [4]int     `json:"gamescore"`    //最新总分
	GameVitamin [4]float64 `json:"game_vitamin"` //最新疲劳值信息
	ScoreOffset [4]int     `json:"scoreoffset"`  //分数变化量
}

type Msg_S_FanLaiZi struct {
	PiziCard      int    `json:"pizicard"`      //皮子
	LaiZiCard     int    `json:"laizicard"`     //癞子
	Seat          uint16 `json:"seat"`          //翻癞子玩家
	LeftCardCount byte   `json:"leftcardcount"` //剩余数目
	PlayerFan     [4]int `json:"playerfan"`     //玩家当前番数
}

type WatcherListItem struct {
	Name     string `json:"name"`     //姓名
	Uid      int64  `json:"uid"`      //id
	ImageUrl string `json:"imageurl"` //头像
	Sex      int    `json:"sex"`      //性别
}

type Msg_S_WatcherList struct {
	Items []WatcherListItem `json:"items"` //玩家状态
}

type Msg_S_WatcherQuit struct {
	TableId int `json:"tableid"` //桌号
	Uid     int `json:"uid"`     //观战id
}

//切换观战的位置
type Msg_S_WatcherSwitch struct {
	TableId int `json:"tableid"` //桌号
	Uid     int `json:"uid"`     //观战id
	Seat    int `json:"seat"`    //观战的座位号
}

//用户可以兑换的记牌器列表
type ToolListItem struct {
	Price int `json:"price"` //价格
	Num   int `json:"num"`   //数量
}

//用户可以兑换的记牌器列表
type Msg_S_Tool_ToolList struct {
	//MsgTypeToolList
	ListItemItem []ToolListItem `json:"toollistitem"` //道具列表
	DeadAt       int64          `json:"deadat"`       //失效时间
}

//用户请求兑换记牌器
type Msg_S_Tool_ToolExchange struct {
	//MsgTypeToolExchange
	Uid    int64 `json:"uid"`
	Type   int8  `json:"type"`   //道具类型，0表示记牌器
	Price  int   `json:"price"`  //价格
	Num    int   `json:"num"`    //数量
	DeadAt int64 `json:"deadat"` //失效时间
}

//记牌器数据
type CardRecorderItem struct {
	Point int `json:"point"` //牌值
	Num   int `json:"num"`   //数量
}

//发送用户的记牌器数据
type Msg_S_Tool_SendCardRecorder struct {
	//MsgTypeSendCardRecorder
	CardRecorderItem []CardRecorderItem `json:"critem"`   //记牌器列表
	MustOpen         int                `json:"mustopen"` //强制打开记牌器列表
}

// 20210303 累计时间数据
type Msg_S_TotalTimer struct {
	Id        int64  `json:"id"`         //发送玩家ID
	Seat      uint16 `json:"seat"`       //玩家座位号
	Open      bool   `json:"open"`       //开始计时
	LimitTime int64  `json:"limittimer"` //累计剩余时间
	StartTime int64  `json:"starttimer"` //累计开始时间
}

type Msg_S_HandCardsData struct {
	CardCount byte  `json:"cardcount"` //发送玩家手牌数量
	CardData  []int `json:"carddata"`  //发送玩家手牌
}
