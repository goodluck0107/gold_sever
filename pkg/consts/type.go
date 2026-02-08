package consts

import (
	"time"
)

// 调试模式：1
// **
// 发包时必须为：0
const Debug = 0

const ClubHouseOwnerPay = false

// 疲劳值功能
const (
	VitaminStartDefault    = 100000                                       // 疲劳值初始值
	VitaminExchangeRate    = 100                                          // 疲劳值汇率
	VitaminInvalidValueCli = 0xfffffffe                                   // 客户端疲劳值无效值
	VitaminInvalidValueSrv = VitaminInvalidValueCli * VitaminExchangeRate // 服务器疲劳值无效值
)

const OfficialWorkLeagueAreaCode = 1 // chess官方自营加盟商区域码

const DefaultAreaCode = "420100"

const (
	// 加密方式
	EncodeNone = 0 // 不加密
	EncodeAes  = 1 // aes加密

	// 更新状态
	UpdateNoNeed = 0 // 不需要更新
	UpdatePrompt = 1 // 提示更新
	UpdateNeed   = 2 // 强制更新

	// 平台
	PlatformWeb          = 0 // web测试
	PlatformAndroid      = 1 // 安卓
	PlatformIos          = 2 // IOS
	PlatformWechatApplet = 3 // 微信小程序

	// 渠道
	AppDefault         = "zljj_20210927_main_01"            // android/ios应用
	AppWeChatApplet    = "zljj_20210927_main_01"            // 微信小程序
	APPHuaWei          = "hw_zl20210526_01"                 // 华为
	UnknownMachineCode = "9f89c84a559f573636a47ff8daed0d33" // 未知的机器码

	// 性别
	SexUnknown = 0 // 未知
	SexMale    = 1 // 男性
	SexFemale  = 2 // 女性

	// 牌桌类型
	CreateTypeSelf   = 0 // 自己创建自己玩
	CreateTypeOther  = 1 // 替他人开房
	CreateTypeHouse  = 2 // 包厢开房
	CreateTypeSystem = 3 // 系统开房

	// 渠道标识
	ChannelApp    = 0 // 默认
	ChannelApplet = 1 // 小程序
	ChannelHW     = 2 // 华为

	// 财富类型
	WealthTypeCard = 1 // 房卡
	WealthTypeGold = 2 // 金币(金币)
	// WealthTypeItem = 3 // 道具
	WealthTypeCoupon  = 5  // 礼券
	WealthTypeVitamin = 6  // 疲劳值
	WealthTypeDiamond = 7  // 钻石
	WealthTypeCardRcd = 10 // 道具-记牌器

	// 封号类型
	BlackStatusForbiddenLogin  = 1 // 不允许登录
	BlackStatusAccountAbnormal = 2 // 账号异常

	// 用户来源
	OriginDefault = 0 // 默认用户
	OriginQichun  = 1 // 蕲春用户

	// 购买商品支付方式
	PayTypeWeixin = 1
	PayTypeAlipay = 2
	PayTypeIos    = 3

	// 房卡支付方式
	CostTypeCreator = 1 // 房主支付房卡
	CostTypeWiner   = 2 // 大赢家支付房卡
	CostTypeRevenue = 3 // 人均支付金币

	VitaminCostTypeWin   = 1
	VitaminCostTypeTotal = 2

	// 公告类型
	NoticePositionTypeAll      = 0
	NoticePositionTypeDialog   = 1 // 大厅公告(弹窗)
	NoticePositionTypeMaintain = 2 // 维护公告
	NoticePositionTypeMarquee  = 3 // 跑马灯
	NoticePositionTypeOption   = 4 // 功能
	// 功能公告类型
	NoticeOptionKind_UpdKindPanel    = 0 // 玩法面板
	NoticeOptionKind_UpdLeague       = 1 // 更新加盟商数据
	NoticeOptionKind_UpdLeagueUser   = 2 // 更新加盟商楼主数据
	NoticeOptionKind_UpdDeliveryInfo = 3 // 更新玩家个人信息图像

	// 服务器状态
	ServerStatusOffline  = 0 // 下线
	ServerStatusOnline   = 1 // 上线
	ServerStatusMaintain = 2 // 维护

	// 心跳包超时时间
	HeartBeatTimeOut = 13

	// 用户类型
	UserTypeYk      = 1 // 游客账号
	UserTypeMobile  = 2 // 手机账号(未绑定微信)
	UserTypeWechat  = 3 // 微信账号
	UserTypeMobile2 = 4 // 手机账号(绑定了微信)
	UserTypeApple   = 5 // 苹果账号
	UserTypeHW      = 6 // 华为账号

	MaxRecordDays       = 7 //
	MaxHouseInviteTimes = 3 // 包厢邀请次数
)

const (
	ROLE_CREATER = iota // 创建者
	ROLE_ADMIN          // 管理员
	ROLE_MEMBER         // 成员
	ROLE_APLLY          // 申请
	ROLE_BLACK          // 黑名单
)

const (
	OPTION_INSERT = iota
	OPTION_DELETE
)

// 包厢楼层活动状态
const (
	HFACT_UNBEGUN  = iota // 未开始
	HFACT_BEGINING        // 进行中
	HFACT_ENDING          // 已结束
)
const (
	HFACT_ROUNDS = iota
	HFACT_BW
	HFACT_SCORE
)

const (
	SESSION_OFFLINE = iota
	SESSION_ONLINE
)

// ！服务端websocket断开原因
const (
	SESSION_CLOED_FORCE    = iota //！服务器主动断开, 客户端不重连
	SESSION_CLOED_ASYSTOLE        //！心跳停止
	SESSION_CLOED_BYGAME          //！游戏数据异常, 踢掉用户
	SESSION_CLOED_CONNECT         //! 服务器主动断开
)

// 校验平台
func CheckPlatform(plat int) bool {
	if plat == PlatformWeb || plat == PlatformAndroid || plat == PlatformIos || plat == PlatformWechatApplet {
		return true
	}
	return false
}

// task pro commit
const (
	TaskTypeWXShare = 0 // 微信分享
)

// task type
const (
	TASK_TYPE_DAYILY = 0 // 每日任务
	TASK_TYPE_SYS    = 1 // 系统任务
)

// task kind
const (
	TASK_KIND_SHARE       = 0 // 分享
	TASK_KIND_GAME_ROUND  = 1 // 对局次数
	TASK_KIND_GAME_WIN    = 2 // 胜场次数
	TASK_KIND_DAYILY_SIGN = 3 // 每日签到
	/*
		TASK_KIND_CARD_OUT       = 3  // 打出某种牌型
		TASK_KIND_CARD_OUT_FIRST = 4  // 第一个打出某种牌型
		TASK_KIND_CARD_END       = 5  // 以某种牌型结尾
		TASK_KIND_CARD_IN        = 6  // 摸到某种牌型

		TASK_KIND_CERTIFICATION  = 8  // 实名认证奖励
		TASK_KIND_WELCOMEGIFT    = 9  // 见面礼
		TASK_KIND_Recharge       = 10 // 充值
	*/
)

// 任务状态
const (
	TASK_STA_DOING     = 0 // 正在进行中
	TASK_STA_COMPLETED = 1 // 完成未领奖
	TASK_STA_RECEIVED  = 2 // 已领奖
)

// Task 保留ID
const (
// TASK_USER_SIGN          = 1 // 签到
// TASK_USER_CERTIFICATION = 2 // 实名认证奖励
// TASK_USER_WELCOMEGIFT   = 3 // 见面礼
)

// task cardtype
const (
	//任务的牌型定义
	TASK_KIND_CARDTYPE_NONE = iota + 100 //无牌型
	//打出
	TASK_KIND_BEAN_HEI_510K  //打出黑体510K
	TASK_KIND_BEAN_HONG_510K //打出红桃510k
	TASK_KIND_BEAN_MEI_510K  //打出玫瑰510K
	TASK_KIND_BEAN_FANG_510K //打出方块510K
	TASK_KIND_BEAN_4_2       //打出4个2
	TASK_KIND_BEAN_4_5       //打出4个5
	TASK_KIND_BEAN_4_10      //打出4个10
	TASK_KIND_BEAN_4_K       //打出4个K
	TASK_KIND_BEAN_7_ZHA     //打出7喜
	//第一个打出
	TASK_KIND_BEAN_FIRST_PLANE   //第一个打出任意飞机
	TASK_KIND_BEAN_FIRST_4_STR   //第一个打出4连对
	TASK_KIND_BEAN_FIRST_STR_678 //第一个打出678连对 667788
	TASK_KIND_BEAN_FIRST_STR_789 //第一个打出789连对
	TASK_KIND_BEAN_FIRST_STR_910 //第一个打出910连对
	TASK_KIND_BEAN_FIRST_STR_10J //第一个打出10J连对
	TASK_KIND_BEAN_FIRST_STR_QK  //第一个打出QK连对
	TASK_KIND_BEAN_FIRST_STR_KA  //首次打出KA连对

	//牌型结尾
	TASK_KIND_BEAN_LAST_HEI_510K  //以黑桃五十K结尾
	TASK_KIND_BEAN_LAST_HONG_510K //以红桃五十K结尾
	TASK_KIND_BEAN_LAST_MEI_510K  //以梅花五十K结尾
	TASK_KIND_BEAN_LAST_THREE_5   //以3个5结尾
	TASK_KIND_BEAN_LAST_THREE_10  //以3个10结尾
	TASK_KIND_BEAN_LAST_THREE_K   //以3个K结尾
	TASK_KIND_BEAN_LAST_PLANE     //以任意飞机结尾
	TASK_KIND_BEAN_LAST_4_STR     //以任意4连对结尾
	TASK_KIND_BEAN_LAST_STR_678   //以678连对结尾
	TASK_KIND_BEAN_LAST_STR_789   //以789连对结尾
	TASK_KIND_BEAN_LAST_STR_910   //以910连对结尾
	TASK_KIND_BEAN_LAST_STR_10J   //以10J连对结尾
	TASK_KIND_BEAN_LAST_STR_QK    //以QK连对结尾
	TASK_KIND_BEAN_LAST_STR_KA    //以KA连对结尾
	TASK_KIND_BEAN_LAST_3         //以一个3结尾
	TASK_KIND_BEAN_YAO_BAI        //打出摇摆

	//摸到
	TASK_KIND_BEAN_ZHUA_TIAN //摸到天炸
	TASK_KIND_BEAN_ZHUA_8_XI //摸到八喜
	TASK_KIND_BEAN_ZHUA_7_XI //摸到七喜
)

const (
	TASK_OP_TYPE_TASK_STATE  = 0 // 操作状态
	TASK_OP_TYPE_TASK_REWARD = 1 // 领奖
	TASK_OP_TYPE_TASK_DATA   = 2 // 向客户端更新部分数据
)

const (
	HALL_SHOW_GAME_FIRST  = iota // 大厅展示一号位
	HALL_SHOW_GAME_SECOND        // 大厅展示二号位
	HALL_SHOW_GAME_THIRD         // 大厅展示三号位
	HALL_SHOW_GAME_MAX           // 大厅城市子包最大展示个数
)

const (
	AREAGAME_OFFICIAL_MAX  = 2
	AREAGAME_RECOMMEND_MAX = 9
	AREAGAME_HISTORY_MAX   = 3
)

const (
	REDIS_DB_DEFULT  = iota // 默认redis库
	REDIS_DB_API_SVR        // api读取的redis库
)

const (
	OneSecond = 1 * time.Second
	OneMinute = OneSecond * 60
	OneHour   = OneMinute * 60
	OneDay    = OneHour * 24
	OneWeek   = OneDay * 7
)

const (
	_        = iota
	ApplyBef // 声请前
	ApplyIng // 声请中
	ApplyOk  // 声请通过后
)

const (
	EngineCocosCreator = 0
	EngineCocosJs      = 1
	PHPCocosCreator    = 1
	PHPCocosJs         = 2
)

const (
	LeagueTypeNone = 0  // 不存在的加盟商
	LeagueTypeXXL  = -1 // 官方直营加盟商
)

const (
	PlayerStatusNormal = iota // 正常状态
	PlayerStatusAnti          // 防沉迷状态
)

const (
	EventSettleOnBegin   = -1   // 游戏开始时校验
	EventSettleNotPause  = iota // 无状态
	EventSettleGameOver         // 游戏小结算事件
	EventSettleGaming           // 游戏过程中结算事件
	EventSettleGang             // 游戏中杠算分事件
	EventSettleMagic            // 游戏中癞子算分事件
	EventSettleQiangGang        // 游戏中抢杠算分事件
	EventChat                   // 游戏中抢杠算分事件
)

type HouseTableJoinType int64

const (
	SelfAdd HouseTableJoinType = 0
	AutoAdd HouseTableJoinType = 1
	NoCheat HouseTableJoinType = 2
)

const (
	AICHECKTOTALSCOREMIN = 1
	AICHECKTOTALSCOREMAX = 9999
)

const (
	HOUSETABLEINQUICK = -1
	HOUSETABLEINAGAIN = -2
	HOUSETABLEINMIXIN = -3
)

const (
	HOUSEAISUPERMIN = 4
	HOUSEAISUPERMAX = 99
)

const MAXHOUSEFLOORTABLE = 70

type HouseMemberApplyType int

const (
	HouseMemberApplyJoin HouseMemberApplyType = 0
	HouseMemberApplyExit HouseMemberApplyType = 1
)

// 场次类型
const (
	GAME_SITTYPE_RAND  = 0 // 随机模式
	GAME_SITTYPE_FIXED = 1 // 做桌模式
)

const (
	SHOP_PRODUCT_GOODS = 0 // 兑换商品，实物
	SHOP_PRODUCT_GOLD  = 1 // 兑换商品，豆子
	SHOP_PRODUCT_CARD  = 2 // 兑换商品，房卡
	SHOP_PRODUCT_BILL  = 3 // 兑换商品，话费
)

const (
	SHOP_PRODUCT_WAITING = 0 // 待发放
	SHOP_PRODUCT_SUC     = 1 // 已经发放
	SHOP_PRODUCT_ERROR   = 2 // 违规
	SHOP_PRODUCT_FAILD   = 3 // 兑换失败
)

type GameStandUpType uint

func (g GameStandUpType) String() (str string) {
	switch g {
	case GameStandUpUserActive:
		str = "主动站起"
	case GameStandUpOfflineTimeOut:
		str = "离线超时站起"
	case GameStandUpOfflineNotBegin:
		str = "游戏未开始，金币场离线，踢出桌子"
	case GameStandUpOfflineLookOn:
		str = "旁观离线，踢出桌子"
	case GameStandUpReadyTimeOut:
		str = "准备超时"
	case GameStandUpForceExit:
		str = "强退"
	case GameStandUpKickUser:
		str = "被踢"
	case GameStandUpGoldNotEnough:
		str = "金币低于场次下限"
	case GameStandUpGoldExceed:
		str = "金币高于场次上限"
	case GameStandUpTableDel:
		str = "桌子解散"
	default:
		str = "未知原因"
	}
	return
}

const (
	GameStandUpUserActive      = iota // 用户主动站起
	GameStandUpOfflineTimeOut         // 离线超时站起
	GameStandUpOfflineNotBegin        // 游戏未开始，金币场离线，踢出桌子
	GameStandUpOfflineLookOn          // 旁观玩家离线，直接踢出
	GameStandUpReadyTimeOut           // 准备超时
	GameStandUpForceExit              // 用户强退
	GameStandUpKickUser               // 用户被踢
	GameStandUpGoldNotEnough          // 金币低于场次下限
	GameStandUpGoldExceed             // 金币高于场次上限
	GameStandUpTableDel               // 桌子解散
)

const (
	HmUserListAll = iota
	HmUserListOnline
	HmUserListJY //静止娱乐
)

const (
	TIME_Y_M_D     = "2006-01-02"
	TIME_Y_M_D_24H = "2006-01-02 00:00:00"
)

const (
	SERVER_FREE_PLAYER_NUM   = 10
	SERVER_NORMAL_PLAYER_NUM = 50
	SERVER_BUSY_PLAYER_NUM   = 100
	SERVER_HOT_PLAYER_NUM    = 100 // > 100
)

const (
	SERVER_STA_FREE   = 0
	SERVER_STA_NORMAL = 1
	SERVER_STA_BUSY   = 2
	SERVER_STA_HOT    = 3
)

// 之前的迁移用户定义 在 migrate infrastructure.go 文件中
const (
	MigratorOriginHBMJ = 6 // chess湖北麻将 6
)
