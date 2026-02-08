// ! 服务器之间的消息
package static

import (
	"github.com/open-source/game/chess.git/pkg/consts"
)

// ################
// 玩家id
type MsgC2SUid struct {
	Uid int64 `json:"uid"`
}

type MsgC2SHidUid struct {
	Hid int   `json:"hid"`
	Uid int64 `json:"uid"`
}

// !游客登录
type Msg_LoginYK struct {
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
}

// !手机登录
type Msg_LoginMobile struct {
	Mobile      string `json:"mobile"`
	Code        string `json:"code"`
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
}

// !手机登录v2
type Msg_LoginMobilev2 struct {
	Mobile      string `json:"mobile"`
	Password    string `json:"password"`
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
}

// !手机注册
type Msg_MobileResgister struct {
	Mobile      string `json:"mobile"`
	Code        string `json:"code"`
	Password    string `json:"password"`
	Nickname    string `json:"nickname"`
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
}

// !重置密码
type Msg_ResetPassword struct {
	Mobile   string `json:"mobile"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

// !微信登录
type Msg_LoginWechat struct {
	Code        string `json:"code"`   // 微信登录code
	Mobile      string `json:"mobile"` // 手机号
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
	AppId       string `json:"appid"` // 应用id
}

// !微信登录
type Msg_LoginWechatV2 struct {
	Code        string `json:"code"`   // 微信登录code
	Mobile      string `json:"mobile"` // 手机号
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
	AppId       string `json:"appid"` // 应用id
	//Info        string `json:"eve"`  // 用户信息
	//Iv          string `json:"iv"`    // 解密偏移量
	Expand  string `json:"expand"`  // 扩展字段
	RawData string `json:"rawData"` //玩家信息
}

// 苹果登录
type Msg_LoginApple struct {
	UserUnionId   string `json:"userunionid"`
	Nickname      string `json:"nickname"`
	IdentityToken string `json:"identitytoken"`
	Platform      int    `json:"platform"`
	MachineCode   string `json:"machinecode"`
	AppId         string `json:"appid"`
}

// 华为登录
type Msg_LoginHW struct {
	OpenId      string `json:"openid"`      // 普通用户的标识
	UnionId     string `json:"unionid"`     // 用户统一标识
	Nickname    string `json:"nickname"`    // 普通用户昵称
	AvatarUrl   string `json:"avatarurl"`   // 用户头像地址
	AccountFlag int    `json:"accountflag"` // 账号类型 0华为账号 1 AppTouch账号
	AccessToken string `json:"accesstoken"` // 服务器解析凭证AccessToken
	Platform    int    `json:"platform"`    // 安卓平台 1
	MachineCode string `json:"machinecode"` // 机器码
	AppId       string `json:"appid"`       // 自定义应用标识 hw_dq20210526_01
}

// !令牌登录
type Msg_LoginToken struct {
	Uid         int64  `json:"uid"`
	Token       string `json:"token"`
	Platform    int    `json:"platform"`
	MachineCode string `json:"machinecode"`
}

// !令牌登录
type Msg_C2Http_TokenUid struct {
	Uid   int64  `json:"uid"`
	Token string `json:"token"`
}

// !获取短信验证码
type Msg_Smscode struct {
	Mobile string `json:"mobile"`
	Type   uint8  `json:"type"`
}

// !下发商品
type Msg_DeliverProduct struct {
	Type       int8   `json:"type"`         // 下发类型
	WealthType int8   `json:"product_type"` // 财富类型
	Num        int    `json:"num"`          // 数量
	Uid        int64  `json:"uid"`          // 用户id
	Extra      string `json:"extra"`        // 扩展数据
}

// !公告更新
type Msg_NoticeUpdate struct {
	PositionType int    `json:"position_type"` // 公告类型
	Kind         int    `json:"kind"`          // 功能公告类型
	Id           string `json:"id"`            // 玩家id
}

// !公告更新
type Msg_ExplainUpdate struct {
	KindId int `json:"kind_id"`
}

// 服务器状态更新
type Msg_UpdateGameServer struct {
	GameId int `json:"gameid"` // 游戏服id
	Status int `json:"status"` // 服务器状态
}

// 服务器维护
type Msg_HG_UpdateGameServer struct {
	KindId int                      `json:"kindid"`
	GameId NoticeMaintainServerType `json:"gameid"`
	//Context Msg_HG_UpdateContext `json:"context"`
}

type Msg_HG_UpdateContext struct {
	Id           int    `json:"id"`            // id
	KindId       int    `json:"kind_id"`       // 关联子游戏
	ContentType  int    `json:"content_type"`  // 内容类型(文字 or 图片)
	Image        string `json:"image"`         // 图片内容
	Content      string `json:"content"`       // 文字内容
	PositionType int    `json:"position_type"` // 公告类型
	ShowType     int    `json:"show_type"`     // 展示类型(每天一次 or 登录一次)
	StartAt      string `json:"start_at"`      // 开始时间
	EndAt        string `json:"end_at"`        // 结束时间
}

// !制定服务器读取
type Msg_AssIgnReLoad struct {
	Server int   `json:"wuhan"`
	Games  []int `json:"games"`
}

// ！得到玩法列表
type Msg_KindIdList struct {
	Kindids []int `json:"kindids"`
}

// ！更新包厢配置
type Msg_UpdHouseCon struct {
	Column string `json:"column"`
	Value  int    `json:"value"`
}

// ！清除玩家的某些信息
type Msg_ClearUserInfo struct {
	ClearUid    int64  `json:"clear_uid"`
	ClearColumn string `json:"clear_column"`
}

// ! 更新财富数据
type Msg_UpdWealth struct {
	Uid        int64 `json:"uid"`        // 用户id
	TableId    int   `json:"tableid"`    // 牌桌id
	WealthType int8  `json:"wealthtype"` // 财富类型
	WealthNum  int   `json:"wealthnum"`  // 房卡数
	CostType   int8  `json:"costtype"`   // 流水类型
	Num        int   `json:"num"`
}

// ! 牌桌局数变化
type Msg_GH_SetBegin struct {
	TableId int   `json:"tableid"` // 牌桌id
	Begin   bool  `json:"begin"`   // 是否开始
	Step    int   `json:"step"`    // 当前局数
	Fid     int64 `json:"fid"`
	Hid     int   `json:"hid"`
}

// 牌桌椅子数发生变化
type Msg_GH_OnFewer struct {
	TableId     int     `json:"table_id"`     //! 牌桌号
	ActiveUsers []int64 `json:"active_users"` //! 座上的玩家
}

// ！重新读取配置文件
type Msg_ReloadConfig struct {
	Games    []int `json:"games"`    //！要通知的game服务器进程编号
	CallGame bool  `json:"callgame"` //！是否通知game服务器
	CallHall bool  `json:"callhall"`
}

// ! 绑定微信
type Msg_BindWechat struct {
	Code  string `json:"code"`
	AppId string `json:"appid"`
}

// ! 授权微信
type Msg_AuthWechat struct {
	RawData string `json:"rawData"` // 用户信息
}

// ! 绑定手机
type Msg_BindMobile struct {
	Mobile   string `json:"mobile"`   // 手机号
	Password string `json:"password"` // 密码
	Code     string `json:"code"`     // 验证码
}

// ! 解绑手机
type Msg_UnbindMobile struct {
	Mobile string `json:"mobile"`
	Code   string `json:"code"`
}

// ! 获取在线玩家
type Msg_GetOnlineUsers struct {
}

// ! 历史战绩
type Msg_HallGameRecord struct {
	Start  int `json:"start"`
	End    int `json:"end"`
	PBegin int `json:"pbegin"`
	PEnd   int `json:"pend"`
}

// ! 检查回放码id
type Msg_CheckReplayId struct {
	ReplayId int64 `json:"replayid"`
}

// ! 加入区域
type Msg_AreaIn struct {
	Code    string `json:"code"`
	IsValid bool   `json:"isvalid"`
}

// ! 加入房间
type Msg_SiteIn struct {
	KindId   int `json:"kind_id"`
	SiteType int `json:"site_type"`
}

// ! 获取回放码详情
type Msg_GetReplayInfo struct {
	ReplayId int64 `json:"replayid"`
}

// ! 历史战绩详情
type Msg_GameRecordInfo struct {
	GameNum string `json:"gamenum"`
}

// ! 实名认证
type Msg_Certification struct {
	Idcard string `json:"idcard"` // 身份证号
	Name   string `json:"name"`   // 真实姓名
}

// ! 保存用户游戏列表
type Msg_SaveGames struct {
	Games string `json:"games"` // 游戏列表字符串
}

// ! 创建房间
type Msg_CreateTable struct {
	KindId           int    `json:"kindid"`            // 游戏玩法
	PlayerNum        int    `json:"playernum"`         // 游戏开始人数
	RoundNum         int    `json:"roundnum"`          // 游戏局数
	CostType         int    `json:"costtype"`          // 房费支付模式
	Restrict         string `json:"restrict"`          // ip及gprs限制
	GVoice           string `json:"gvoice"`            // 有没有勾选语音字段
	FewerStart       string `json:"fewerstart"`        // 是否可以少人开始
	GameConfig       string `json:"gameconfig"`        // 游戏玩法配置(不同玩法配置不同)
	Gps              bool   `json:"gps"`               // 是否开启gps
	Version          int    `json:"version"`           // 子游戏强更版本
	RecommendVersion int    `json:"recommend_version"` // 子游戏提示/推荐更新版本
	Voice            bool   `json:"voice"`             // 客户端有没有开启语音权限
	GVoiceOk         bool   `json:"gvoiceok"`          // 客户端gvoice有没有初始化成功
	Difen            int64  `json:"difen"`             // 底分
	Adddifen         int64  `json:"adddifen"`          // adddifen
}

// 用户区域广播
type Msg_HG_UserAreaBroadcast struct {
	Broadcast *Msg_S2C_UserAreaBroadcast `json:"broadcast"`
	KindId    []int                      `json:"kindid"`
}

// ! 创建房间
type Msg_HG_CreateTable struct {
	Id             int         `json:"id"`             //! 牌桌id(6位)
	NTId           int         `json:"ntid"`           //! 牌桌索引
	HId            int         `json:"hid"`            //! 包厢id(6位)
	DHId           int64       `json:"dhid"`           //! 包厢id(数据库自增id)
	FId            int64       `json:"fid"`            //! 楼层id(数据库自增id)
	NFId           int         `json:"nfid"`           //! 楼层索引
	CreateType     int         `json:"createtype"`     // 创建牌桌类型
	Creator        int64       `json:"creator"`        // 创建牌桌的用户id
	KindId         int         `json:"kindid"`         // 游戏玩法
	MinPlayerNum   int         `json:"minplayernum"`   // 游戏开始最小人数
	MaxPlayerNum   int         `json:"maxplayernum"`   // 游戏开始最大人数
	RoundNum       int         `json:"roundnum"`       // 游戏局数
	CardCost       int         `json:"cardcost"`       // 消耗房卡
	CostType       int         `json:"costtype"`       // 房费支付模式
	View           bool        `json:"view"`           // 是否允许观战
	Restrict       bool        `json:"restrict"`       // ip及gprs限制
	GVoice         string      `json:"gvoice"`         // 语音限制
	FewerStart     string      `json:"fewerstart"`     // 是否可以少人开局
	GameConfig     string      `json:"gameconfig"`     // 游戏玩法配置(不同玩法配置不同)
	MatchConfig    MatchConfig `json:"matchconfig"`    // 排位赛配置
	IsLimitChannel bool        `json:"islimitchannel"` // 是否限制渠道同服
	Difen          int64       `json:"difen"`          // 底分

	AutoUid  int64 `json:"autouid"`  //自动加入玩家id
	AutoSeat int   `json:"autoseat"` //自动加入座位
}

// 任务配置表
type MatchConfig struct {
	Id        int    `json:"id"`           //! id
	Name      string `json:"32"`           // 场次名
	KindId    int    `json:"kind_id"`      //! 子游戏id
	Type      int    `json:"type"`         //! 类型(初级场 中级场 高级场)
	ConfigStr string `json:"match_config"` //! 游戏参数配置
	State     int    `json:"state"`        //!状态，0未开始，1进行中，2已经结束
	Flag      int    `json:"flag"`         //!是否开启，0未开启，1开启
	BeginDate int64  `json:"begindate"`    //! 开始日期
	EndDate   int64  `json:"enddate"`      //! 结束日期
	BeginTime int64  `json:"begintime"`    //! 开始时间
	EndTime   int64  `json:"endtime"`      //! 结束时间
}

// ! 加入房间
type Msg_HG_TableIn struct {
	TableId int   `json:"tableid"` // 牌桌id
	Uid     int64 `json:"uid"`     // 用户id
	Seat    int   `json:"seat"`    // 座位号
	Payer   int64 `json:"payer"`   // 谁支付
}

// 加入房间结果
type Msg_GH_TableIn struct {
	Seat   int   `json:"seat"`    // 座位号
	JoinAt int64 `json:"join_at"` // 加入时间
}

// ! 用户重复登录
type Msg_HG_Relogin struct {
	TableId int   `json:"tableid"` // 牌桌id
	Uid     int64 `json:"uid"`     // 用户id
}

// ! 加入房间
type Msg_HG_TableInfo struct {
	TableId int `json:"tableid"` // 牌桌id
}

// ! 解散牌桌
type Msg_HG_TableDel_Req struct {
	TableId int    `json:"tableid"`  // 牌桌id
	Info    string `json:"eve"`      //解散信息
	HallDel bool   `json:"hall_del"` //大厅是否解散
}
type Msg_GH_TableDel_Ack struct {
	TableId int `json:"tableid"` // 牌桌id
}
type Msg_GH_TableDel_Ntf struct {
	TableId int `json:"tableid"` // 牌桌id
}

// ! 加入房间
type Msg_JoinTable struct {
	Id        int     `json:"id"`        // 牌桌id
	Seat      int     `json:"seat"`      // 座位号,-1时顺序落座(屏蔽, 不需要选位子)
	Gps       bool    `json:"gps"`       // 是否开启gps
	Voice     bool    `json:"voice"`     // 是否开启语音权限
	GVoiceOk  bool    `json:"gvoiceok"`  // 客户端GVoice是否初始化成功
	Longitude float64 `json:"longitude"` // 经度
	Latitude  float64 `json:"latitude"`  // 纬度
	Address   string  `json:"address"`   // 地址
}

// ! 解散房间
type Msg_TableDel struct {
	Id  int   `json:"id"` // 牌桌id
	Hid int   `json:"hid"`
	Fid int64 `json:"fid"`
}

// ! 房间详情
type Msg_TableInfo struct {
	Id int `json:"id"` // 牌桌id
}

type Msg_Game_TableIn struct {
	Id    int    `json:"id"`
	Uid   int64  `json:"uid"`   // 用户id
	Token string `json:"token"` // 登录令牌
	MInfo string `json:"minfo"` // 地理位置信息
}

// 用户落座
type Msg_Table_Seat struct {
	Seat int `json:"seat"` //!座位号
}

type Msg_Table_Standup struct {
	Seat int `json:"seat"` //!座位号
}

// ! 离开房间
type Msg_LeaveTable struct {
	Id int `json:"id"`
}

// ! 登录
type Msg_CH_SetUid struct {
	Uid    int64  `json:"uid"`    // 用户id
	Token  string `json:"token"`  // 登录令牌
	MInfo  string `json:"minfo"`  // 地理位置信息
	Engine int    `json:"engine"` // 游戏引擎信息 空为 cocos creator  非空为 cocos Js
}

// ! 包厢队长列表
type Msg_CH_HousePartnerList struct {
	HId int `json:"hid"`
}

// ! 包厢队长权限创建
type Msg_CH_HousePartnerCreate struct {
	HId int   `json:"hid"`
	Uid int64 `json:"uid"` // 用户id
}

// ! 包厢队长权限删除
type Msg_CH_HousePartnerDelete struct {
	HId int   `json:"hid"`
	UId int64 `json:"uid"` // 用户id
}

// ! 包厢队长权限调配
type Msg_CH_HousePartnerGen struct {
	HId     int   `json:"hid"`
	UId     int64 `json:"uid"`     // 用户id
	Partner int64 `json:"partner"` // 队长id
}

// ! 包厢防沉迷开关
type Msg_CH_HouseVitaminStatus struct {
	HId    int  `json:"hid"`
	Status bool `json:"status"` // false关 true开
}

// ! 包厢防沉迷数值
type Msg_CH_HouseVitaminValues struct {
	HId                     int  `json:"hid"`
	Status                  bool `json:"status"`              // 疲劳值开关 false关 true开
	AdminHide               bool `json:"adminhide"`           // 管理员可见
	AdminModi               bool `json:"adminmodi"`           // 管理员可调
	PartnerHide             bool `json:"partnerhide"`         // 队长可见
	GamePause               bool `json:"gamepause"`           // 到达下限暂定游戏
	MemberSend              bool `json:"membersend"`          // 到达下限暂定游戏
	PartnerModi             bool `json:"partnermodi"`         // 队长可调
	IsDeductConfig          bool `json:"deductconfig"`        // 有没有配置扣除相关
	IsEffectConfig          bool `json:"effectconfig"`        // 有没有配置生效相关
	DisableSetJuniorVitamin bool `json:"disablejuniorv"`      // 是否禁用上级可调下级，false为不禁用（开关为开启状态） true为禁用 （开关为关闭状态）
	PartnerKick             bool `json:"partner_kick"`        // 队长踢人开关
	RewardBalanced          bool `json:"reward_balanced"`     // 奖励均衡开关
	NoSkipVitaminSet        bool `json:"no_skip_vitamin_set"` // 禁止跨级调整疲劳值
	// LowLimit      float64 `json:"lowlimit"`    // 疲劳值下限
	// LowLimitPause float64 `json:"pause"`       // 游戏暂停下限
	// DeductCount   float64 `json:"deductcount"` // 对局扣除
	// DeductType    int     `json:"deducttype"`  // 0大赢家扣 其他AA
}

// ! 包厢疲劳值数值修改
type Msg_CH_HouseVitaminSet struct {
	HId   int     `json:"hid"`
	UId   int64   `json:"uid"`
	Value float64 `json:"value"`
}

// ! 包厢疲劳值数值修改
type Msg_CH_HouseVitaminSetRecords struct {
	HId   int   `json:"hid"`
	UId   int64 `json:"uid"`
	Start int64 `json:"start"`
	Count int64 `json:"count"`
}

// ! 包厢疲劳值数值修改
type Msg_CH_HouseVitaminClear struct {
	HId int `json:"hid"`
}

// ! 包厢盟主队长收益统计
type Msg_CH_HousePartnerRewardStatistic struct {
	HId        int    `json:"hid"`        // 包厢好号
	SelectTime int    `json:"selecttime"` // 时间
	SortType   int    `json:"sorttype"`   // 排序方式
	PBegin     int    `json:"pbeg"`       // 分页开始
	PEnd       int    `json:"pend"`       // 分页结束
	SearchKey  string `json:"searchkey"`  // 模糊搜索
}

// ! 包厢盟主队长疲劳值数值统计
type Msg_CH_HousePartnerVitaminStatistic struct {
	HId        int    `json:"hid"`        // 包厢好号
	Partner    int64  `json:"partner"`    // 0=代表盟主/管理员视角 其他=该队长视角
	ShowType   int    `json:"show_type"`  // 0=展示 Partner 及 Partner 的下级 // 1=展示 Partner 的名下玩家
	SelectTime int    `json:"selecttime"` // 时间 已废弃
	SortType   int    `json:"sorttype"`   // 排序方式
	PBegin     int    `json:"pbeg"`       // 分页开始
	PEnd       int    `json:"pend"`       // 分页结束
	SearchKey  string `json:"searchkey"`  // 模糊搜索
}

// ! 包厢疲劳值数值统计
type Msg_CH_HouseVitaminStatistic struct {
	HId        int `json:"hid"`
	SelectTime int `json:"selecttime"`
	SortType   int `json:"sorttype"`
}

// ! 包厢疲劳值数值统计
type Msg_CH_HouseVitaminMgrList struct {
	HId       int    `json:"hid"`
	SortType  int    `json:"sorttype"`
	PBegin    int    `json:"pbeg"`
	PEnd      int    `json:"pend"`
	SearchKey string `json:"searchkey"`
}

// ! 包厢疲劳值数值统计
type Msg_CH_HouseVitaminStatisticClear struct {
	HId int `json:"hid"`
}

// ! 包厢疲劳值数值修改
type Msg_CH_HouseVitaminBWThreshold struct {
	HId       int `json:"hid"`
	Threshold int `json:"threshold"`
}

// ! 包厢防沉迷管理员可见
type Msg_CH_HouseVitaminOptHide struct {
	HId    int  `json:"hid"`
	Status bool `json:"status"` // false关 true开
}

// ! 包厢创建
type Msg_CH_HouseCreate struct {
	HName   string `json:"hname"`
	HNotify string `json:"hnotify"`
}
type Msg_HG_HouseTableCreate struct {
	Msg_HG_CreateTable
	AppUid  int64 `json:"uid"`  // 申请者用户id
	AppSeat int   `json:"seat"` // 申请者座位号
}

// ! 加入房间
type Msg_HG_HouseTableIn struct {
	TableId int   `json:"tableid"` // 牌桌id
	Uid     int64 `json:"uid"`     // 用户id
	Seat    int   `json:"seat"`    // 座位号
}

// ! 包厢删除
type Msg_CH_HouseDelete struct {
	HId int `json:"hid"`
}

// ! 包厢楼层创建
type FRule Msg_CreateTable
type Msg_CH_HouseFloorCreate struct {
	HId   int   `json:"hid"`
	FRule FRule `json:"frule"`
}

// ! 包厢楼层删除
type Msg_CH_HouseFloorDelete struct {
	HId int   `json:"hid"`
	FId int64 `json:"fid"`
}

// ! 包厢楼层列表
type Msg_CH_HouseFloorList struct {
	HId int `json:"hid"`
}

// ! 玩家加入
type Msg_CH_HouseMemberJoin struct {
	HId       int   `json:"hid"`
	InviteUid int64 `json:"invite_uid"`
}

// ! 玩家审核同意
type Msg_CH_HouseMemberAgree struct {
	Msg_CH_HouseMemberApply
}

// ! 玩家审核拒绝
type Msg_CH_HouseMemberRefused struct {
	Msg_CH_HouseMemberApply
}

// ! 玩家审核拒绝
type Msg_CH_HouseMemberApply struct {
	HId       int                         `json:"hid"`
	UId       int64                       `json:"uid"`
	ApplyType consts.HouseMemberApplyType `json:"apply_type"`
}

// ! 玩家审核拒绝
type Msg_CH_HouseMemberApplyNtf struct {
	Msg_CH_HouseMemberApply
	Opt int64 `json:"opt"`
}

// ! 玩家黑名单加入
type Msg_CH_HouseMemberBlacklistInsert struct {
	HId int   `json:"hid"`
	UId int64 `json:"uid"`
}

// ! 玩家黑名单删除
type Msg_CH_HouseMemberBlacklistDelete struct {
	HId int   `json:"hid"`
	UId int64 `json:"uid"`
}

// ! 玩家退出
type Msg_CH_HouseMemberExit struct {
	HId int `json:"hid"`
}

// ! 玩家剔除
type Msg_CH_HouseMemberKick struct {
	HId int   `json:"hid"`
	UId int64 `json:"uid"`
}

// ! 玩家备注
type Msg_CH_HouseMemberRemark struct {
	HId     int    `json:"hid"`
	UId     int64  `json:"uid"`
	URemark string `json:"uremark"`
}

// ! 玩家设置角色
type Msg_CH_HouseMemberRoleGen struct {
	HId   int   `json:"hid"`
	UId   int64 `json:"uid"`
	URole int   `json:"urole"`
}

// ! 玩家进入
type Msg_CH_HouseMemberIn struct {
	HId     int   `json:"hid"`
	FId     int64 `json:"fid"`
	NeedMix bool  `json:"need_mix"`
	Start   int   `json:"start"`
	Count   int   `json:"count"`
}

// ! 玩家离开
type Msg_CH_HouseMemberOut struct {
	HId int   `json:"hid"`
	FId int64 `json:"fid"`
}

// ! 玩家包厢列表
type Msg_CH_MemberHouseList struct {
	HCreate bool `json:"hcreate"`
	HJoin   bool `json:"hjoin"`
}

// ! 包厢信息
type Msg_CH_HouseBaseInfo struct {
	HId int `json:"hid"`
}

// ! 包厢信息
type Msg_CH_HouseMemOnline struct {
	HId int `json:"hid"`
}

// ! 包厢常玩信息
type Msg_CH_HousePlayOften struct {
	HId int `json:"hid"`
}

// ! 包厢楼层玩法修改
type Msg_CH_HouseFloorRuleModify struct {
	HId   int   `json:"hid"`
	FId   int64 `json:"fid"`
	FRule FRule `json:"frule"`
}

// ! 包厢楼层信息
type Msg_HouseTableInOpt struct {
	Param  Msg_CH_HouseTableIn
	Header string
}

// ! 包厢楼层信息
type Msg_CH_HouseTableIn struct {
	HId        int     `json:"hid"`
	FId        int64   `json:"fid"`
	NTId       int     `json:"ntid"` // -1快速入桌 -2再来一局
	Gps        bool    `json:"gps"`
	Voice      bool    `json:"voice"`
	GVoiceOk   bool    `json:"gvoiceok"` // 客户端GVoice是否初始化成功
	IgnoreRule bool    `json:"ignorerule"`
	RestartId  int     `json:"restart_id"` // 再来一局旧桌子号
	KindID     int     `json:"kindid"`
	Longitude  float64 `json:"longitude"` // 经度
	Latitude   float64 `json:"latitude"`  // 纬度
	Address    string  `json:"address"`   // 地址
}

// ! 修改公告名称
type Msg_CH_HouseBaseNNModify struct {
	HId     int    `json:"hid"`
	HName   string `json:"hname"`
	HNotify string `json:"hnotify"`
}

// ! 包厢成员列表
type Msg_CH_HouseMemList struct {
	HId      int    `json:"hid"`
	Param    string `json:"param"` // id  昵称  或为空  搜索框
	PBegin   int    `json:"pbeg"`
	PEnd     int    `json:"pend"`
	Role     int    `json:"role"`      // 2成员列表，3申请列表，4黑名单列表
	SortType int    `json:"sorttype"`  // 排序方式 6：vtm升序   7：vtm降序 8：离线时间升序 9：离线时间降序
	ListType int    `json:"list_type"` // 0是全部  1是在线 2是禁止娱乐
}

// ! 包厢队长列表
type Msg_CH_HouseParList struct {
	HId    int `json:"hid"`
	PBegin int `json:"pbeg"`
	PEnd   int `json:"pend"`
}

// ! 包厢设置退圈开关
type Msg_CH_HouseOptIsMemExit struct {
	HId       int  `json:"hid"`
	IsMemExit bool `json:"ismemexit"`
}

// ! 包厢设置审核
type Msg_CH_HouseOptIsMemCheck struct {
	HId       int  `json:"hid"`
	IsChecked bool `json:"ischecked"`
}

// ! 包厢设置圈号隐藏
type Msg_CH_HouseOptIsHidHide struct {
	HId    int  `json:"hid"`
	IsHide bool `json:"ishidhide"`
}

// ! 包厢设置圈号隐藏
type Msg_CH_HouseOptIsHeadHide struct {
	HId    int  `json:"hid"`
	IsHide bool `json:"isheadhide"`
}

// ! 包厢设置圈号隐藏
type Msg_CH_HouseOptIsMemUidHide struct {
	HId    int  `json:"hid"`
	IsHide bool `json:"is_mem_uid_hide"`
}

// ! 包厢设置在线人数隐藏
type Msg_CH_HouseOptIsOnlineHide struct {
	HId    int  `json:"hid"`
	IsHide bool `json:"isonlinehide"`
}

// ! 包厢设置成员列表隐藏
type Msg_CH_HouseOptIsMemHide struct {
	HId    int  `json:"hid"`
	IsHide bool `json:"ismemhide"`
}

// ! 包厢设置冻结
type Msg_CH_HouseOptIsFrozen struct {
	HId      int  `json:"hid"`
	IsFrozen bool `json:"isfrozen"`
}

// ! 包厢设置在线人数隐藏
type Msg_CH_HouseOptPartnerKick struct {
	HId         int  `json:"hid"`
	PartnerKick bool `json:"partnerkick"`
}

// ! 包厢设置在线人数隐藏
type Msg_CH_HouseOptRewardBalanced struct {
	HId            int  `json:"hid"`
	RewardBalanced bool `json:"reward_balance"`
}

// ! 包厢成员审核列表
type Msg_CH_HouseMemApplyList struct {
	HId    int    `json:"hid"`
	Param  string `json:"param"`
	PBegin int    `json:"page"`
	PEnd   int    `json:"page"`
}

// ! 包厢桌子详细信息
type Msg_CH_HouseTableBaseInfo struct {
	HId  int   `json:"hid"`
	FId  int64 `json:"fid"`
	NTId int   `json:"ntid"`
}

// ! 包厢成员统计
type Msg_CH_HouseMemStatistics struct {
	Hid               int    `json:"hid"`
	DFid              int    `json:"dfid"`
	GroupId           int    `json:"group_id"`
	Partner           int64  `json:"partner"`
	DayType           int    `json:"daytype"`
	SortType          int    `json:"sorttype"`
	PBegin            int    `json:"pbegin"`
	PEnd              int    `json:"pend"`
	SearchKey         string `json:"searchkey"`
	LikeFlag          int    `json:"likeflag"`          // 0 全部 1 点赞 2 未点赞
	QueryTimeInterval int    `json:"querytimeinterval"` // 查询时段（效验使用）3,6,12,0
	QueryTimeRange    int    `json:"querytimerange"`    // 查询是第几个时间区间(例如 以3个小时为时段，则共有8个区间，参数可传1-8)
	LowScoreFlag      int    `json:"lowscoreflag"`      // 低分局查询方式 0 盟主/成员视角查询 1 队长视角查询
}

type Msg_CH_HouseCardStatistics struct {
	Hid        int    `json:"hid"`
	DFid       int    `json:"dfid"`
	SelectTime int    `json:"selecttime"`
	SortType   int    `json:"sorttype"`
	PBegin     int    `json:"pbeg"`
	PEnd       int    `json:"pend"`
	SearchKey  string `json:"searchkey"`
}

type Msg_CH_HouseMemStatisticsTotal struct {
	Hid     int `json:"hid"`
	DFid    int `json:"dfid"`
	DayType int `json:"daytype"`
	GroupId int `json:"group_id"`
}

// ! 包厢队长绑定
type Msg_CH_HousePartnerMemCustom struct {
	HId    int    `json:"hid"`
	PId    int64  `json:"pid"`
	Param  string `json:"param"`
	PBegin int    `json:"pbeg"`
	PEnd   int    `json:"pend"`
	IsBind bool   `json:"is_bind"`
}

// ! 包厢队长绑定
type Msg_CH_HouseFloorVipUser struct {
	HId    int    `json:"hid"`
	Fid    int64  `json:"fid"`
	Param  string `json:"param"`
	PBegin int    `json:"pbeg"`
	PEnd   int    `json:"pend"`
	IsVip  bool   `json:"is_vip"`
}

// ! 包厢队长绑定
type Msg_CH_HouseFloorVipUserSet struct {
	HId   int   `json:"hid"`
	Fid   int64 `json:"fid"`
	Uid   int64 `json:"uid"`
	IsVip bool  `json:"is_vip"`
}

// ! 包厢楼层添加VIP玩家或删除 一键操作
type Msg_CH_HouseFloorVipUserAllSet struct {
	HId   int   `json:"hid"`
	Fid   int64 `json:"fid"`
	IsVip bool  `json:"is_vip"`
}

// ! 包厢队长绑定
type Msg_CH_HouseFloorVipUserSetNtf struct {
	Msg_CH_HouseFloorVipUserSet
	UName    string `json:"uname"`
	UUrl     string `json:"uurl"`
	UGender  int    `json:"ugender"`
	NumViper int64  `json:"num_viper"`
}

// ! 包厢战绩查询
type Msg_CH_HouseGameRecord struct {
	HId               int64  `json:"hid"`               // 包厢id
	DFID              int    `json:"dfid"`              // 包厢楼层索引(-1查所有楼层)
	UId               int64  `json:"uid"`               // 用户id(0查所有的)
	SelectTime        int    `json:"selecttime"`        // 查询日期(0今天, -1昨天,依次类推)
	QueryBeginTime    int64  `json:"querybegintime"`    // 查询时间戳(0表示服务器当前时间的10条数据, > 0 就从这个时间段查询)
	SearchKey         string `json:"searchkey"`         // 查询时间戳(0表示服务器当前时间的10条数据, > 0 就从这个时间段查询)
	BwUser            bool   `json:"bwuser"`            // 是否只查询大赢家,针对searchkey为uid时候使用
	RecordType        int    `json:"recordtype"`        // 战绩页签类型.0圈子战绩,1对局详情,2大赢家详情(已废弃)
	QueryTimeInterval int    `json:"querytimeinterval"` // 查询时段（效验使用）3,6,12,0
	QueryTimeRange    int    `json:"querytimerange"`    // 查询是第几个时间区间(例如 以3个小时为时段，则共有8个区间，参数可传1-8)
	LikeFlag          int    `json:"likeflag"`          // 0 全部 1 点赞 2 未点赞
	LowScoreFlag      int    `json:"lowscoreflag"`      // 低分局查询方式 0 盟主/成员视角查询 1 队长视角查询
	RoundType         int    `json:"roundtype"`         // 牌局类型 0 全部 1 完整局数 2 中途解散 3 低分局
}

// ! 包厢经验状况
type Msg_CH_HouseId struct {
	HId int `json:"hid"` //包厢id
}

// ! 包厢经验状况
type Msg_CH_HouseOperationalStatus struct {
	HId int64 `json:"hid"` //包厢id
}

// ! 包厢有效对局积分获取
type Msg_CH_HouseValidRoundScoreGet struct {
	HId int `json:"hid"` //包厢id
}

type HouseValidRoundScoreSetItem struct {
	FId   int64   `json:"fid"`
	Score float64 `json:"score"`
	// BigScore int   `json:"bigscore"`
}

// ! 包厢有效对局积分获取
type Msg_CH_HouseValidRoundScoreSet struct {
	HId   int                           `json:"hid"` //包厢id
	Items []HouseValidRoundScoreSetItem `json:"items"`
}

// ! 包厢我的战绩查询
type Msg_CH_HouseMyRecord struct {
	HId int `json:"hid"`
}

type Msg_CH_HouseMerge struct {
	HId  int `json:"hid"`
	THId int `json:"thid"`
}

// ! 包厢我的战绩查询
type Msg_CH_HouseOwnerInfo struct {
	HId   int   `json:"hid"`
	THId  int   `json:"thid"`
	Owner int64 `json:"owner"`
}

type Msg_CH_HouseMergeRsp struct {
	HId    int  `json:"hid"`
	THId   int  `json:"thid"`
	Result bool `json:"result"`
}

// ! 包厢战绩查询
type Msg_CH_HouseRecord struct {
	Hid    int `json:"hid"`
	PBegin int `json:"pbegin"`
	PEnd   int `json:"pend"`
}

// ! 包厢大赢家对对局统计查询
type Msg_CH_HouseRecordStatus struct {
	HId        int    `json:"hid"`
	RecordType int    `json:"recordtype"`
	Param      string `json:"param"`
	PBegin     int    `json:"pbegin"`
	PEnd       int    `json:"pend"`
}

// ! 包厢战绩点赞
type Msg_CH_HouseRecordHeart struct {
	HId     int    `json:"hid"`
	GameNum string `json:"gamenum"`
	IsHeart int    `json:"isheart"`
}

// ! 大赢家对局统计清楚功能
type Msg_CH_HouseRecordStatusClean struct {
	HId        int   `json:"hid"`
	UId        int64 `json:"uid"`
	RecordType int   `json:"recordtype"`
}

// ! 大赢家对局统计清楚功能
type Msg_CH_HouseRecordStatusCleanAll struct {
	HId        int `json:"hid"`
	RecordType int `json:"recordtype"`
}

// ! 包厢房卡消耗统计查询
type Msg_CH_HouseRecordKaCost struct {
	Hid int `json:"hid"`
}

// ! 包厢楼层活动创建
type Msg_CH_HouseActCreate struct {
	HId         int          `json:"hid"`
	FIds        []int64      `json:"fids"`
	ActType     int          `json:"acttype"`
	ActName     string       `json:"actname"`
	HideInfo    bool         `json:"hideinfo"`
	ActBegTime  int64        `json:"actbegtime"`
	ActEndTime  int64        `json:"actendtime"`
	Type        int64        `json:"type"` //0 老活动，1 抽奖活动
	Rewords     []RewordInfo `json:"rewords"`
	TicketCount int64        `json:"ticket_count"` //多少局抽一次
}

// ! 包厢楼层活动删除
type Msg_CH_HouseActDelete struct {
	HId   int   `json:"hid"`
	ActId int64 `json:"actid"`
}

// ! 包厢楼层活动列表
type Msg_CH_HouseActList struct {
	HId int `json:"hid"`
}

// ! 包厢楼层活动信息
type Msg_CH_HouseActInfo struct {
	HId   int   `json:"hid"`
	ActId int64 `json:"actid"`
	Uid   int64 `json:"uid"`
}

type Msg_CH_HouseTableShowCount struct {
	HId         int `json:"hid"`
	Count       int `json:"count"`
	MinTableNum int `json:"mintablenum"`
	MaxTableNum int `json:"maxtablenum"`
}

// ! 任务数据提交
type Msg_CH_TaskCommit struct {
	TaskId   int `json:"taskid"`   // 任务Id(如果不知道可不填)
	TaskType int `json:"tasktype"` // 任务具体类型(分享任务0、对局任务1、胜场次数2)
	Num      int `json:"num"`      // 任务计数
}

// ! 任务列表
type Msg_CH_TaskList struct {
	TaskType int `json:"tasktype"` // 每日任务、系统任务
}

// ! 任务奖励领取
type Msg_CH_TaskReward struct {
	TaskId int `json:"taskid"`
	Share  int `json:"share"`
}

type Msg_CH_AreaGamesByPkg struct {
	PackageKey string `json:"key"`
}

// 区域游戏包统一搜索接口
type Msg_CH_AreaGameSeek struct {
	Code        string          `json:"code"`         // 指定区域码搜索: 如果是传空没指定则会在所有区域搜索
	Keyword     string          `json:"keyword"`      // 指定关键字搜索: 会模糊匹配包名/包名首字母简写/包所在区域/包类型，传空不走过滤步骤
	Type        AreaSeekType    `json:"type"`         // 指定搜索类型: 0代表不指定类型在所有游戏搜索，1代表在推荐玩法中搜索 2代表在附近玩法中搜索
	PackageType AreaPackageType `json:"package_type"` // 指定包类型: 0所有 1花牌 2字牌 3扑克 4麻将
	// 4个字段 分别代表了4个筛选条件，没有筛选先后顺序，四个筛选条件走完，即可得到对应的搜索结果
	// 如果4个字段都没传，按照逻辑 和得到所有区域包
}

// ###################
// 金币场
type Msg_Game_SiteList struct {
	KindId int `json:"kindid"` // 游戏kindid
}

type Msg_Game_SiteIn struct {
	KindId   int    `json:"kindid"`   // 游戏id
	SiteType int    `json:"sitetype"` // 场次类型
	Uid      int64  `json:"uid"`      // 用户id
	Token    string `json:"token"`    // 登录令牌
	MInfo    string `json:"minfo"`    // 地理位置信息
}

// ! 入桌
type Msg_Game_SiteInTable struct {
	Id        int   `json:"id"`         // 桌号
	Uid       int64 `json:"uid"`        // 用户id
	Seat      int   `json:"seat"`       // 座位号
	TableId   int   `json:"table_id"`   // 桌子编号
	AutoReady bool  `json:"auto_ready"` // 入桌后自动准备
}

type Msg_Game_Addiction struct {
	Status     map[int64]int `json:"status"`
	Content    string        `json:"content"`
	Remaintime [4]int64      `json:"remaintime"` // 防沉迷解散时间
}

// ! 换桌
type Msg_Game_ChangeInTable struct {
	Id        int   `json:"id"`         // 桌号
	Uid       int64 `json:"uid"`        // 用户id
	AutoReady bool  `json:"auto_ready"` // 入桌后自动准备
}

// ! 牌桌疲劳值推送更新
type Msg_Game_UserVitaminUpd struct {
	Uid     int64   `json:"uid"` // 用户id
	Seat    uint16  `json:"seat"`
	Vitamin float64 `json:"vitamin"`
}

// !加入列表
type Msg_Game_SiteListIn struct {
	Start int `json:"start"` // 起始偏移量
	End   int `json:"end"`   // 结束偏移量
}

// !牌桌变化通知
type Msg_Game_TableNotify struct {
	Ntid int `json:"ntid"` // 牌桌桌号
}

// ! 通过包kindid取得rule
type Msg_CH_GameRules struct {
	Engine  int   `json:"engine"`
	KindIds []int `json:"kind_ids"`
	Channel int   `json:"channel"`
}

// ! 排位赛列表查询
// matchlist
type Msg_Match_MatchList struct {
	Uid int64 `json:"uid"`
}

// ! 排位赛排位查询
// matchrankinglist
type Msg_Match_RankingList struct {
	Uid           int64 `json:"uid"`
	RecordType    int   `json:"recordtype"`
	KindId        int   `json:"kind_id"`       //! 子游戏id
	Type          int   `json:"type"`          //! 类型(初级场 中级场 高级场)
	LowerRange    int   `json:"lowerrange"`    // 排名范围的起始点
	UpperRange    int   `json:"upperrange"`    // 排名范围的终点 终点为0表示取所有的记录
	BeginDateTime int64 `json:"begindatetime"` //! 开始日期和时间
	EndDateTime   int64 `json:"enddatetime"`   //! 结束日期和时间
}

// ! 排位赛奖励列表查询
// matchawardlist
type Msg_Match_AwardList struct {
	Uid           int64  `json:"uid"`
	RecordType    int    `json:"recordtype"`
	KindId        int    `json:"kind_id"`       //! 子游戏id
	TypeStr       string `json:"typestr"`       //! 类型(初级场 中级场 高级场)
	BeginDate     int64  `json:"begindate"`     //! 开始日期
	EndDate       int64  `json:"enddate"`       //! 结束日期
	BeginDateTime int64  `json:"begindatetime"` //! 开始日期和时间
	EndDateTime   int64  `json:"enddatetime"`   //! 结束日期和时间
}

// 公告结构
type OptNotice struct {
	Kind int    `json:"kind"`
	Data string `json:"data"`
}

// 子玩法面板更新
type OptGKRule struct {
	Grversion int `json:"game_rule_version"`
}

// 区域信息变更

// 加盟商信息变更

// 加盟商楼主信息变更

// MsgHouseFloorRename 修改包厢楼层名称
type MsgHouseFloorRename struct {
	Hid     int    `json:"hid"`
	FloorID int64  `json:"floor_id"`
	Name    string `json:"name"`
}

// MsgHouseMixFloor 创建、编辑包厢混排信息
type MsgHouseMixFloor struct {
	Hid               int                       `json:"hid"`
	FIDs              []int64                   `json:"fids"`
	TableNum          int                       `json:"table_num"`             // 手动加桌模式-桌子数
	MixActive         bool                      `json:"mix_active"`            // 混排总开关
	AICheck           bool                      `json:"ai_check"`              // 智能防作弊开关
	AITotalScoreLimit int                       `json:"ai_total_score_limit"`  // 触发智能防作弊的条件之一：总分上限
	AISuper           bool                      `json:"ai_super"`              // 超级防作弊开关
	TableJoinType     consts.HouseTableJoinType `json:"house_table_join_type"` // 混排模式
	EmptyTableBack    bool                      `json:"empty_table_back"`      // 是否空桌子在后面
	EmptyTableMax     int                       `json:"empty_table_max"`       // 最大空桌数
	TableSortType     int                       `json:"table_sort_type"`       // 桌子排序类型 0 正常 1 极左
	NewTableSortType  int                       `json:"new_table_sort_type"`   //  新版本桌子排序
	CreateTableType   int                       `json:"create_table_type"`     // 开桌类型
}

// MsgHouseMixInfo 获取包厢当前混排信息
type MsgHouseMixInfo struct {
	HID int `json:"hid"`
}
type MsgHouseMixFloorTableCreate struct {
	HID int `json:"hid"`
	FID int `json:"fid"`
}

type TableChangeDetail struct {
	FID      int64 `json:"fid"`
	TableNum int   `json:"table_num"`
}
type MsgHouseMixFloorTableChange struct {
	HID    int                 `json:"hid"`
	Detail []TableChangeDetail `json:"detail"`
}

// 编辑包厢禁止同桌列表
type MsgHouseMemberTableLimit struct {
	Hid     int    `json:"hid"`
	GroupID int    `json:"group_id"`
	Param   string `json:"param"`
	PBegin  int    `json:"pbeg"`
	PEnd    int    `json:"pend"`
}

// 添加包厢禁止同桌分组
type MsgHouseTableLimitGroupAdd struct {
	Hid int `json:"hid"`
}

// 移除包厢禁止同桌分组
type MsgHouseTableLimitGroupRemove struct {
	Hid     int `json:"hid"`
	GroupID int `json:"group_id"`
}

// 添加禁止同桌用户
type MsgHouseTableLimitUserAdd struct {
	Hid     int   `json:"hid"`
	GroupID int   `json:"group_id"`
	Uid     int64 `json:"uid"`
}

// 2人禁止同桌是否不生效
type MsgHouse2PTableLimitNotEffectSet struct {
	Hid int  `json:"hid"`
	Sta bool `json:"sta"`
}

type MsgHouseMsg struct {
	Hid   int `json:"hid"`
	Flag  int `json:"Flag"` // 0 全部消息（兼容线上） 1 包厢消息 2进圈消息 3退圈消息
	Start int `json:"start"`
	End   int `json:"end"`
}

// 闲聊战绩
// see http://update.hhkin.com/web/#/47?page_id=1241
type MsgXianTalkRecord struct {
	Room       string            `json:"room"`
	GameNumber string            `json:"gamenumber"`
	Ante       string            `json:"ante"`
	Time       string            `json:"time"`
	User       []MsgXianTalkUser `json:"user"`
	HouseId    int               `json:"houseid"`
}

type MsgXianTalkUser struct {
	UserId   int64   `json:"userid"`
	UserName string  `json:"username"`
	ImgUrl   string  `json:"imgurl"`
	Qrcode   string  `json:"qrcode"`
	Win      bool    `json:"win"`
	Number   float64 `json:"number"`
}

// 大厅http请求
// 再来一局
type Msg_C2Http_HouseAnotherGame struct {
	Uid   int64  `json:"uid"`   // 玩家uid
	Token string `json:"token"` // 玩家token
}

type MsgHouseUserLimitGame struct {
	HID       int   `json:"hid"`
	UID       int64 `json:"uid"`
	AllowGame bool  `json:"allow_game"`
}

// 在线换桌
type Msg_C2Http_HouseChangeTable struct {
	Msg_CH_HouseTableIn
	Uid   int64  `json:"uid"`   // 玩家uid
	Token string `json:"token"` // 玩家token
}

type MsgHouseUserExitTable struct {
	Uid int64 `json:"uid"`
	Tid int   `json:"tid"`
}

type MsgHouseTableUserKick struct {
	Uid int64 `json:"uid"`
	Tid int   `json:"tid"`
	Opt int64 `json:"opt"`
}

type MsgUserSubFloor struct {
	SubDetail []string `json:"sub_detail"`
}

type Msg_C2Http_HouseMemberAgree struct {
	Msg_CH_HouseMemberAgree
	Opuid int64  `json:"op_uid"` //操作者uid
	Token string `json:"token"`  // 玩家token
}

type Msg_C2Http_HouseMemberRefused struct {
	Msg_CH_HouseMemberRefused
	Opuid int64  `json:"op_uid"` //操作者uid
	Token string `json:"token"`  // 玩家token
}

type MsgVitaminSend struct {
	ToUid int64   `json:"to_uid"`
	Value float64 `json:"value"`
	Hid   int     `json:"hid"`
}

type MsgVitaminSendNtf struct {
	Uid   int64  `json:"sender"`
	Hid   int    `json:"hid"`
	HName string `json:"h_name"`
	Uname string `json:"u_name"`
}

type MsgHouseMemGetById struct {
	Hid int   `json:"hid"`
	Uid int64 `json:"uid"`
}

type GmChangeUserPhone struct {
	Uid   int64  `json:"uid"`
	Phone string `json:"phone"`
}

type GmHouseRevoke struct {
	ParentId int `json:"parentid"`
	SonId    int `json:"sonid"`
}

type Msg_CH_HouseParnterRoyaltySet struct {
	Hid       int   `json:"hid"`
	ParnterId int64 `json:"parnterid"`
	// Royaltys       []int `json:"royaltys"`
	RoyaltyPercent []int `json:"royalty_percent"`
	// SuperiorProfit []int `json:"superiorprofits"`
}

type Msg_CH_HouseParnterRoyaltyGet struct {
	Hid       int   `json:"hid"`
	ParnterId int64 `json:"parnterid"`
}

type Msg_CH_HouseOwnerRoyaltyGet struct {
	Hid       int   `json:"hid"`
	ParnterId int64 `json:"parnterid"`
}

type Msg_CH_HouseOwnerRoyaltySet struct {
	Hid       int   `json:"hid"`
	ParnterId int64 `json:"parnterid"`
	// Royaltys       []int `json:"royaltys"`
	// JuniorProfit   []int `json:"junior_profit"`
	JuniorPercent  []int `json:"junior_percent"`
	RoyaltyPercent []int `json:"royalty_percent"`
}

type Msg_CH_HouseParnterSuperiorList struct {
	Hid       int    `json:"hid"`
	ParnterId int64  `json:"parnterid"`
	SearchKey string `json:"searchkey"`
	Page      int    `json:"page"`
}

type Msg_CH_HouseParnterBindSuperior struct {
	Hid        int   `json:"hid"`
	ParnterId  int64 `json:"parnterid"`
	SuperiorId int64 `json:"superiorid"`
}

type Msg_CH_HouseParnterBindJunior struct {
	Hid    int   `json:"hid"`
	Junior int64 `json:"junior"`
}

type Msg_CH_MsgHouseParnterFloorStatistics struct {
	Hid       int    `json:"hid"`
	DayType   int    `json:"daytype"`
	FidIndex  int    `json:"fidindex"`
	SearchKey string `json:"searchkey"`
}

type Msg_CH_MsgHouseParnterFloorMemStatistics struct {
	Hid       int   `json:"hid"`
	ParnterId int64 `json:"parnterid"`
	DayType   int   `json:"daytype"`
	FidIndex  int   `json:"fidindex"`
}

type Msg_CH_MsgHousePartnerHistoryFloorStatistics struct {
	Hid int `json:"hid"`
}

type Msg_CH_MsgHousePartnerHistoryFloorDetailStatistics struct {
	Hid   int   `json:"hid"`
	DFid  int64 `json:"dfid"`
	Fid   int   `json:"fid"`
	Start int   `json:"start"`
	Count int   `json:"count"`
}

type GpsInfo struct {
	Ip        string  `json:"ip"`
	Longitude float64 `json:"longitude"` //! 经度
	Latitude  float64 `json:"latitude"`  //! 纬度
	Address   string  `json:"address"`   //! 纬度
}

type Msg_CH_HouseJoinInvite struct {
	Hid  int   `json:"hid"`  // 包厢id
	TUid int64 `json:"tuid"` // 对方id
}

type Msg_CH_HouseJoinInviteRsp struct {
	Hid     int   `json:"hid"`     // 包厢数据库dhid
	Inviter int64 `json:"inviter"` // 邀请人uid
	Agree   bool  `json:"agree"`   // 同意与否
	Notips  bool  `json:"notips"`  // 是否今日不再提示
}

type Msg_CH_HouseDialogEdit struct {
	Hid     int    `json:"hid"`     // 包厢id
	Content string `json:"content"` // 内容
	Active  bool   `json:"active"`  // 是否激活
}

// ! 根据邀请码加入包厢
type Msg_HC_HouseJoinByCode struct {
	Code string `json:"code"` // 邀请码
}

type Msg_CH_HouseFloorWaitingNum struct {
	Hid       int           `json:"hid"`        // 包厢id
	FloorsMap map[int64]int `json:"floors_map"` // 每个楼层的等待人数
}

type MsgHouseVitaminAdminSet struct {
	Hid     int   `json:"hid"`
	Uid     int64 `json:"uid"`
	IsAdmin bool  `json:"is_admin"`
}

type MsgHouseVicePartnerSet struct {
	Hid         int   `json:"hid"`
	OptUid      int64 `json:"optuid"`
	Uid         int64 `json:"uid"`
	VicePartner bool  `json:"vicepartner"`
}

type AdminKickMem struct {
	Hid int   `json:"hid"`
	Uid int64 `json:"uid"`
}

type HouseGroupAdd struct {
	Hid int `json:"hid"`
}

type HouseGroupDel struct {
	Hid     int `json:"hid"`
	GroupId int `json:"group_id"`
}

type HouseGroupUserAdd struct {
	Hid     int   `json:"hid"`
	GroupId int   `json:"group_id"`
	Uid     int64 `json:"uid"`
}

type HouseGroupInfo struct {
	Hid     int `json:"hid"`
	GroupId int `json:"group_id"`
}

type HouseGroupUserList struct {
	Hid       int    `json:"hid"`
	GroupId   int    `json:"group_id"`
	Start     int    `json:"start"`
	Count     int    `json:"count"`
	SearchKey string `json:"searchkey"`
}
type Msg_CH_HouseParnterRoyaltyForMe struct {
	Hid int   `json:"hid"`
	Uid int64 `json:"uid"`
}

type Msg_CH_HouseParnterRoyaltyHistory struct {
	Hid int   `json:"hid"`
	Uid int64 `json:"uid"`
}

type Msg_CH_HouseNoLeagueStatistics struct {
	Hid               int    `json:"hid"`
	Fid               int    `json:"fid"`
	DayType           int    `json:"daytype"`
	SearchKey         string `json:"searchkey"`
	PBegin            int    `json:"pbegin"`
	PEnd              int    `json:"pend"`
	PartnerLevel      int    `json:"partnerlevel"`
	QueryTimeInterval int    `json:"querytimeinterval"` // 查询时段（效验使用）3,6,12,0
	QueryTimeRange    int    `json:"querytimerange"`    // 查询是第几个时间区间(例如 以3个小时为时段，则共有8个区间，参数可传1-8)
	LowScoreFlag      int    `json:"lowscoreflag"`      // 低分局查询方式 0 盟主/成员视角查询 1 队长视角查询
	LikeFlag          int    `json:"likeflag"`          // 0 全部 1 点赞 2 未点赞
}

type Msg_CH_HouseNoLeagueDetailStatistics struct {
	Hid               int    `json:"hid"`
	Partner           int64  `json:"partner"`
	Fid               int    `json:"fid"`
	DayType           int    `json:"daytype"`
	SearchKey         string `json:"searchkey"`
	PBegin            int    `json:"pbegin"`
	PEnd              int    `json:"pend"`
	QueryTimeInterval int    `json:"querytimeinterval"` // 查询时段（效验使用）3,6,12,0
	QueryTimeRange    int    `json:"querytimerange"`    // 查询是第几个时间区间(例如 以3个小时为时段，则共有8个区间，参数可传1-8)
	LowScoreFlag      int    `json:"lowscoreflag"`      // 低分局查询方式 0 盟主/成员视角查询 1 队长视角查询
	SortType          int    `json:"sorttype"`          // 排序类型
}

type MsgHouseGameSwitch struct {
	Uid int64 `json:"uid"`
	On  bool  `json:"on"`
}
type MsgHouseUserGameSwitch struct {
	Hid int  `json:"hid"`
	On  bool `json:"on"`
}

type MsgHouseGameSwitchNtf struct {
	Hid         int  `json:"hid"`
	GameOn      bool `json:"game_on"`
	AdminGameOn bool `json:"admin_game_on"`
	IsVitamin   bool `json:"is_vitamin"`
}

type MsgHouseAgentUpdateNtf struct {
	Hid                int  `json:"hid"`
	VipFloorShowSwitch bool `json:"vip_floor_show_switch"`
	UnionSwitch        bool `json:"union_switch"`
}

type MsgHousePrizeSetS struct {
	Hid   int     `json:"hid"`
	Value []int64 `json:"value"`
	Type  int64   `json:"type"`
}

type MsgHousePrizeInfo struct {
	Hid  int   `json:"hid"`
	Type int64 `json:"type"`
}

type MsgHouseFloorPay struct {
	Hid   int                     `json:"hid"`
	Items []*MsgHouseFloorPayItem `json:"items"`
}

type MsgHouseFloorPayItem struct {
	Fid       int64   `json:"fid"`
	AA        bool    `json:"aa"`
	Gear1Cost float64 `json:"c1"`
	// 2
	Gear2      bool    `json:"g2"` // 是否勾选低分局(二档)
	Gear2Under float64 `json:"u2"` // 低分局低于(二档)
	// 3
	Gear3 bool `json:"g3"` // 是否勾选第三档
	//Gear3Above int64  // 第三档高于
	Gear3Under float64 `json:"u3"` // 第三档低于
	Gear3Cost  float64 `json:"c3"` // 第三档扣除
	// 4
	Gear4 bool `json:"g4"` // 是否勾选第四档
	//Gear4Above int64  // 第四档高于
	Gear4Under float64 `json:"u4"` // 第四档低于
	Gear4Cost  float64 `json:"c4"` // 第四档扣除
	// 5
	Gear5 bool `json:"g5"` // 是否勾选第五档
	//Gear5Above int64  // 第五档高于
	Gear5Under float64 `json:"u5"` // 第五档低于
	Gear5Cost  float64 `json:"c5"` // 第五档扣除
	// 6
	Gear6 bool `json:"g6"` // 是否勾选第六档
	//Gear6Above int64  // 第六档高于
	Gear6Under float64 `json:"u6"` // 第六档低于
	Gear6Cost  float64 `json:"c6"` // 第六档扣除
	// 7
	Gear7 bool `json:"g7"` // 是否勾选第六档
	//Gear7Above int64  // 第六档高于
	Gear7Under float64 `json:"u7"` // 第六档低于
	Gear7Cost  float64 `json:"c7"` // 第六档扣除
	// 8
	Gear8 bool `json:"g8"` // 是否勾选第六档
	//Gear9Above int64  // 第六档高于
	Gear8Under float64 `json:"u8"` // 第六档低于
	Gear8Cost  float64 `json:"c8"` // 第六档扣除
	// 9
	Gear9 bool `json:"g9"` // 是否勾选第六档
	//Gear9Above int64  // 第六档高于
	Gear9Under float64 `json:"u9"` // 第六档低于
	Gear9Cost  float64 `json:"c9"` // 第六档扣除
	// 10
	Gear10 bool `json:"g10"` // 是否勾选第六档
	//Gear10Above int64  // 第六档高于
	Gear10Under float64 `json:"u10"` // 第六档低于
	Gear10Cost  float64 `json:"c10"` // 第六档扣除
}

type HouseLuckSet struct {
	Hid     int          `json:"hid"`
	ActId   int64        `json:"actid"`
	Rewords []RewordInfo `json:"rewords"`
}
type RewordInfo struct {
	Rank  int64 `json:"rank"`
	Count int64 `json:"count"`
}

type HouseLuckInfo struct {
	Hid   int   `json:"hid"`
	ActId int64 `json:"actid"`
}

type HouseMemLuck struct {
	Hid   int   `json:"hid"`
	ActId int64 `json:"actid"`
}
type MsgHouseFloorVip struct {
	Hid   int                     `json:"hid"`
	Items []*MsgHouseFloorVipItem `json:"items"`
}

type MsgHouseFloorVipItem struct {
	Fid          int64 `json:"fid"`
	IsVip        bool  `json:"is_vip"`
	Disable      bool  `json:"disable"`
	NumViper     int64 `json:"num_viper"`
	IsCapSetVip  bool  `json:"is_cap_set_vip"`
	IsDefJoinVip bool  `json:"is_def_join_vip"`
}

// ! 包厢经验状况
type Msg_CH_TableDistanceLimitGet struct {
	HId int `json:"hid"` //包厢id
}

// ! 包厢经验状况
type Msg_CH_TableDistanceLimitSet struct {
	HId      int `json:"hid"`      //包厢id
	Distance int `json:"distance"` //限制距离
}

// ! 包厢经验状况
type Msg_CH_RewardBalancedTypeSet struct {
	HId                int `json:"hid"` //包厢id
	RewardBalancedType int `json:"reward_balanced_type"`
}

// ! 包厢经验状况
type MsgHouseApplySwitchSet struct {
	Hid    int  `json:"hid"`
	Switch bool `json:"switch"`
}

type MsgHouseFloorHideImg struct {
	Hid    int   `json:"hid"`
	Fid    int64 `json:"fid"`
	IsHide bool  `json:"ishide"`
}

type MsgHouseFloorFakeTable struct {
	Hid      int   `json:"hid"`      // 包厢id
	Fid      int64 `json:"fid"`      // 楼层id
	MinTable int   `json:"mintable"` // 最小假桌子数
	MaxTable int   `json:"maxtable"` // 最大假桌子数
}

type MsgHouseRecordGameLike struct {
	Hid          int    `json:"hid"`
	GameNum      string `json:"gamenum"`
	IsLike       bool   `json:"islike"`
	DateType     int    `json:"daytype"`
	RecordType   int    `json:"recordtype"`        // 战绩页签类型.0圈子战绩,1对局详情,2大赢家详情
	TimeInterval int    `json:"querytimeinterval"` // 点赞时段（效验使用）3,6,12,0
	TimeRange    int    `json:"querytimerange"`    // 点赞的第几个时间区间(例如 以3个小时为时段，则共有8个区间，参数可传1-8)
}

type MsgHouseRecordUserLike struct {
	Hid          int   `json:"hid"`
	LikeUser     int64 `json:"likeuser"`
	IsLike       bool  `json:"islike"`
	DateType     int   `json:"daytype"`
	TimeInterval int   `json:"querytimeinterval"` // 点赞时段（效验使用）3,6,12,0
	TimeRange    int   `json:"querytimerange"`    // 点赞的第几个时间区间(例如 以3个小时为时段，则共有8个区间，参数可传1-8)
	IsTeamLike   bool  `json:"isteamlike"`        // 团队统计中的用户点赞
}

type MsgChangeBlankUser struct {
	Uid     int64  `json:"uid"`
	Hid     int    `json:"hid"`
	Bind    bool   `json:"bind"`
	EndTime int64  `json:"end_time"`
	Reason  string `json:"reason"`
}

type MsgBlackUserChangeNtf struct {
	Uid     int64  `json:"uid"`
	Bind    bool   `json:"bind"`
	EndTime int64  `json:"end_time"`
	Reason  string `json:"reason"`
}

type MsgBlackHouseChangeNtf struct {
	Hid     int    `json:"hid"`
	Bind    bool   `json:"bind"`
	EndTime int64  `json:"end_time"`
	Reason  string `json:"reason"`
}

type Msg_C2Http_HouseTableDel struct {
	DelTable Msg_TableDel `json:"tab_del"`
	Opuid    int64        `json:"op_uid"` //操作者uid
	Token    string       `json:"token"`  // 玩家token
}

type Msg_Http_HouseTableReset struct {
	Hid      int    `json:"hid"`
	PassCode string `json:"passcode"`
}

type MsgSetLogFileLevel struct {
	Servers string `json:"servers"`
	Game    int    `json:"game"` // 要修改的游戏服务器id -1代表所有
	Level   string `json:"level"`
}

type Msg_CH_HouseApplyInfo struct {
	HId      int    `json:"hid"` //包厢id
	Join     bool   `json:"join"`
	Exit     bool   `json:"exit"`
	Param    string `json:"param"`
	PBegin   int    `json:"pbeg"`
	PEnd     int    `json:"pend"`
	Role     int    `json:"role"`
	SortType int    `json:"sorttype"`
}

// ! 包厢经验状况
type Msg_CH_HouseFloorColorSet struct {
	HId         int      `json:"hid"` //包厢id
	FloorsColor []string `json:"floors_color"`
}

// ! 房间详情
type Msg_TableUserKick struct {
	TId int   `json:"tid"` // 牌桌id
	Uid int64 `json:"uid"` // 玩家id
}

type Msg_PartnerRemark struct {
	Name string `json:"uremark"` // 牌桌id
	Hid  int    `json:"hid"`
	Uid  int64  `json:"uid"` // 玩家id
}

// 获取商品列表
type Msg_Shop_Product struct {
	Uid  int64 `json:"uid"`  // 用户id
	Type int   `json:"type"` // 平台
}

// 获取兑换记录
type Msg_Shop_Record struct {
	Uid  int64 `json:"uid"`  // 用户id
	Type int   `json:"type"` // 平台
	Page int   `json:"page"` // 翻页(从0开始)
}

// 获取奖励记录
type Msg_Shop_GoldRecord struct {
	Uid  int64 `json:"uid"`  // 用户id
	Type int   `json:"type"` // 平台
	Page int   `json:"page"` // 翻页(从0开始)
}

// 绑定手机号
type Msg_Shop_PhoneBind struct {
	Uid int64  `json:"uid"` // 用户id
	Tel string `json:"tel"` // 手机号码
}

// 绑定手机号
type Msg_Shop_GetPhoneBind struct {
	Uid int64 `json:"uid"` // 用户id
}

// 商品兑换
type Msg_Shop_Exchange struct {
	Uid int64 `json:"uid"` // 用户id
	Id  int   `json:"id"`  // 商品ID
}

// 商品兑换
type Msg_Payment_OrderId struct {
	TradeNo string `json:"tradeNo" gorm:"unique_index"`
}

// 包厢楼层信息
type Msg_CH_GetHouseFloorPackageInfo struct {
	HId int   `json:"hid"`
	FId int64 `json:"fid"`
}

// 包厢楼层信息
type Msg_CH_HousePrivateGPSSet struct {
	HId        int  `json:"hid"`
	PrivateGPS bool `json:"privategps"`
}

// 分享的配置信息
type Msg_C2Http_GetShareCfg struct {
	Uid      int64 `json:"uid"`       // 用户id
	SceneId  int   `json:"scene_id"`  // 分享场景id(0 大厅主界面分享 1 大厅任务界面分享 2 小结算界面分享 3 礼券界面分享)
	Platform int   `json:"platform"`  // 平台类型 0 不区分平台 1 app端 2 小程序
	KindId   int   `json:"kind_id"`   // 指定的游戏分享 默认0
	SiteType int   `json:"site_type"` // 指定的场次 默认0
}

// 分享成功
type Msg_C2Http_ShareSuc struct {
	Uid     int64 `json:"uid"`      // 用户id
	ShareId int   `json:"share_id"` // 分享id
}

// ! 排位赛Tag列表查询
// matchtaglist
type Msg_Match_MatchTagList struct {
	Uid    int64 `json:"uid"`
	KindId int   `json:"kind_id"`
}

type Msg_C2Http_GetAllowanceInfo struct {
	Msg_C2Http_TokenUid
	IgnoreGift bool `json:"ignore_gift"`
}

type MsgTableIn struct {
	AutoReady bool `json:"auto_ready"`
}

// 设置房卡低于xx时提示盟主
type Msg_C2S_SetFangKaTipsMinNum struct {
	Hid    int `json:"hid"`
	MinNum int `json:"minnum"`
}

// 设置房卡低于xx时不再提示盟主
type Msg_C2S_SetFangKaLowerTisNotPop struct {
	Hid int `json:"hid"`
}

// 禁止队长
type MsgHouseCaptainLimitGame struct {
	HID          int   `json:"hid"`
	UID          int64 `json:"uid"`
	AllowGame    bool  `json:"allow_game"`
	IsTeamMember bool  `json:"is_team_member"` // 是否禁止 队长和队员   true  false
}

type MsgC2SRefuseInvite struct {
	RefuseInvite bool `json:"refuse_invite"` //拒绝入圈邀请
}

type MsgC2SUpdateHmUright struct {
	Hid         int    `json:"hid"`
	Uid         int64  `json:"uid"`
	UpdateRight string `json:"update_right"`
}

type MsgC2SHmSwitch struct {
	Hid    int            `json:"hid"`
	Switch map[string]int `json:"switch"`
}

// ! 牌桌局数变化
type Msg_HG_TableWatchIn struct {
	TableId int   `json:"tableid"` // 牌桌id
	Begin   bool  `json:"begin"`   // 是否开始
	Step    int   `json:"step"`    // 当前局数
	Fid     int64 `json:"fid"`
	Hid     int   `json:"hid"`
}

// ! 牌桌局数变化
type Msg_GH_TableWatchOut struct {
	TableId int   `json:"tableid"` // 牌桌id
	Uid     int64 `json:"uid"`
}

// 设置战绩筛选时段
type Msg_C2S_SetRecordTimeInterval struct {
	Hid          int `json:"hid"`          // 包厢id
	TimeInterval int `json:"timeinterval"` // 以多长时间为一个时间段
}

// 重置用户指定包厢权限
type Msg_Http_ResetUserRight struct {
	Dhid int64 `json:"dhid"`
	Uid  int64 `json:"uid"`
}

// 重置用户指定包厢权限
type Msg_Http_ForceHotter struct {
	Version string `json:"version"`
}

// 注销用户
type Msg_Http_WriteOffUser struct {
	Uid int64 `json:"uid"`
}

// ! 排行榜设置
type Msg_CH_HouseRankSet struct {
	HId        int  `json:"hid"`         //包厢id
	RankRound  int  `json:"rank_round"`  // 局数排行榜 开启：0001  隐藏成员：0010  隐藏对应数据：0100
	RankWiner  int  `json:"rank_winer"`  // 赢家排行榜 开启：0001  隐藏成员：0010  隐藏对应数据：0100
	RankRecord int  `json:"rank_record"` // 战绩排行榜 开启：0001  隐藏成员：0010  隐藏对应数据：0100
	RankOpen   bool `json:"rank_open"`   //开启排行榜 按钮显示作用
}

type Msg_CH_HouseParterBan struct {
	HId     int   `json:"hid"`      // 包厢id
	Pid     int64 `json:"pid"`      // 合伙人ID
	TeamBan bool  `json:"team_ban"` // 是否全队禁止
}

type Msg_CH_HouseParterAlarmValueSet struct {
	HId        int     `json:"hid"`         // 包厢id
	Pid        int64   `json:"pid"`         // 合伙人ID
	AlarmValue float64 `json:"alarm_value"` // 警戒值
}

type Msg_CH_HouseParterRawardSet struct {
	HId    int   `json:"hid"`    // 包厢id
	Pid    int64 `json:"pid"`    // 合伙人ID
	Reward int   `json:"reward"` // 奖励分成百分比
}

type Msg_CH_HouseMemberNoFloorsSet struct {
	HId      int     `json:"hid"`       // 包厢id
	Uid      int64   `json:"uid"`       // 合伙人ID
	NoFloors []int64 `json:"no_floors"` // 奖励分成百分比
}

type Msg_CH_HouseParterAACostSet struct {
	HId int   `json:"hid"` // 包厢id
	Pid int64 `json:"pid"` // 合伙人ID
	AA  bool  `json:"aa"`  // 警戒值
}

// ! 获取排行榜设置
type Msg_CH_HouseRankGet struct {
	HId int `json:"hid"` //包厢id
}

// ! 获取排行榜数据
type Msg_CH_HouseRankInfoGet struct {
	HId      int `json:"hid"`       //包厢id
	RankType int `json:"rank_type"` // 局数排行榜：0  赢家排行榜 ：1  战绩排行榜 ： 2
	TimeType int `json:"time_type"` // 今天：0   昨天：-1  周榜： 2  月榜： 3
	PBegin   int `json:"pbegin"`
	PEnd     int `json:"pend"`
}

// 盟主小队或者队长小队 打烊了
type MsgOffWork struct {
	HID       int  `json:"hid"`         // 包厢id
	IsOffWork bool `json:"is_off_work"` // true 打烊了 false 营业中
}

type MsgC2SUserSetSex struct {
	Sex int `json:"sex"` //性别   1：男  2：女   客户端大厅 用1判断了男生  客户端游戏 用 ！= 2 判断了男生  在兼容微信性别
}

type MsgC2STHDismissRoomDet struct {
	GameNum string `json:"game_num"` //游戏的唯一标识码
}

// 用户的定位信息
type Msg_C2S_GpsInfo struct {
	Area string `json:"area"` // 定位所在区域码
}

// 未实名、已实名未成年的游戏时长
type Msg_C2Http_PlayTime struct {
	Uid     int64 `json:"uid"`
	TimeSec int   `json:"timesec"` // 游戏时长 秒为单位
}

type Msg_C2S_CheckinDo struct {
	Day int `json:"day"`
}

type Msg_C2S_BattleRank struct {
	DayType  int `json:"daytype"`  // 0今天 1昨天
	RankType int `json:"ranktype"` // 0玩牌局数  1胜利局数
}
