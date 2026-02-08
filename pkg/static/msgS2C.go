// ! 服务器之间的消息
package static

import (
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/consts"
)

// ! 用户详情(返回给客户端)
type Msg_S2C_Person struct {
	Uid               int64  `json:"uid"`               //! 用户数字id
	Nickname          string `json:"nickname"`          //! 昵称
	RealName          string `json:"realname"`          //! 玩家真实姓名
	Idcard            string `json:"idcard"`            //! 玩家身份证号码
	Imgurl            string `json:"imgurl"`            //! 头像
	Card              int    `json:"card"`              //! 房卡数量
	Diamond           int    `json:"diamond"`           //! 钻石数量
	Gold              int    `json:"gold"`              //! 金币数量
	GoldBean          int    `json:"gold_bean"`         //! 金豆
	InsureGold        int    `json:"insure_gold"`       //! 保险箱金币
	HId               int    `json:"hid"`               //! 历史包厢id
	FId               int64  `json:"fid"`               //! 历史包厢楼层id
	GameId            int    `json:"game_id"`           //! 当前处于哪个game中
	SiteId            int    `json:"site_id"`           //! 场次id
	TableId           int    `json:"table_id"`          //! 当前处于哪个room中
	Area              string `json:"area"`              //! 区域code
	Sex               int    `json:"sex"`               //! 性别
	Tel               string `json:"tel"`               //! 手机号
	Certification     bool   `json:"certification"`     //! 是否实名认证
	Ip                string `json:"ip"`                //! ip地址
	Games             string `json:"games"`             //! 用户游戏列表
	DescribeInfo      string `json:"describe_info"`     //! 个性签名
	DeliveryImg       string `json:"delivery_img"`      //! 个人名片头像
	UserType          int    `json:"user_type"`         //! 用户账号类型
	ContributionScore int64  `json:"contributionScore"` // 玩家贡献值
	RefuseInvite      bool   `json:"refuse_invite"`     // 拒绝入圈邀请
	ChuanQiParam      string `json:"chuanqi_param"`     // 传奇参数信息
	HotVersion        string `json:"hot_version"`       // 热更新版本
}

// ! 用户详情(返回给php后台)
type Msg_S2S_Person struct {
	Uid      int64  `json:"uid"`      //! 用户数字id
	Nickname string `json:"nickname"` //! 昵称
	Imgurl   string `json:"imgurl"`   //! 头像
	Card     int    `json:"card"`     //! 房卡数量
	Sex      int    `json:"sex"`      //! 性别
	Tel      string `json:"tel"`      //! 手机号
}

// 发放低保
type Msg_S2C_Allowances struct {
	Current   int `json:"current"`    // 今天第N次领取低保
	Gold      int `json:"gold"`       // 发放金币数量
	Remain    int `json:"remain"`     // 今天剩余低保领取次数
	AfterGold int `json:"after_gold"` // 发放之后的金币数量
}

// 见面礼
type Msg_S2C_WelcomeGift struct {
	PlayCount    int            `json:"play_count"`    // 玩了多少局
	RegisteredAt int64          `json:"registered_at"` // 注册时间时间戳
	Rewards      []*WealthAward `json:"rewards"`       // 财富奖励
}

// ! 告诉客户端卡变化
type Msg_S2C_UpdCard struct {
	Card int  `json:"card"`
	Type int8 `json:"type"`
}

// ! 告诉客户端卡金币变化
type Msg_S2C_UpdGold struct {
	Gold   int  `json:"gold"`
	Offset int  `json:"offset"`
	Type   int8 `json:"type"`
}

// ! 告诉客户端卡金币变化
type Msg_S2C_UpdGoldBean struct {
	GoldBean int  `json:"gold_bean"`
	Type     int8 `json:"type"`
}

// ! 告诉客户端钻石发生变化
type Msg_S2C_UpdDiamond struct {
	Diamond int  `json:"diamond"`
	Type    int8 `json:"type"`
}

// !登录
type Msg_S2C_login struct {
	Openid string `json:"openid"`
	Name   string `json:"name"`   //! 名字
	Imgurl string `json:"imgurl"` //! 头像
	Sex    int    `json:"sex"`    //! 性别
}

// ! 是否是游客账号
type Msg_S2C_CheckYK struct {
	ShowYK bool `json:"showyk"` // 是否是游客账号
}

// 创建牌桌
type Msg_S2C_TableCreate struct {
	Id int `json:"id"` //! 牌桌id
}

// ! 加入牌桌
type Msg_S2C_TableIn struct {
	Id      int    `json:"id"`     //! 牌桌id
	GameId  int    `json:"gameid"` //! 游戏服id
	KindId  int    `json:"kindid"` //! 游戏类型
	Ip      string `json:"ip"`     //! 游戏服ip
	PkgName string `json:"package_key"`
	Version int    `json:"version"`
}

// ! 加入房间
type Msg_S2C_SiteIn struct {
	PackageKey string `json:"package_key"` //! 子游戏包名
	GameId     int    `json:"gameid"`      //! 游戏服id
	KindId     int    `json:"kindid"`      //! 游戏类型
	SiteType   int    `json:"site_type"`   //! 场次类型
	Ip         string `json:"ip"`          //! 游戏服ip
}

// ! 详情牌桌
type Msg_S2C_TableInfo struct {
	Begin        bool                   `json:"begin"`
	Hid          int                    `json:"hid"`
	Fid          int64                  `json:"fid"`
	TId          int                    `json:"tid"`
	NTId         int                    `json:"ntid"`
	MaxPlayerNum int                    `json:"maxplayernum"` // 游戏开始最大人数
	RoundNum     int                    `json:"roundnum"`     //! 总局数
	CurrentRound int                    `json:"currentround"` //! 当前局数
	Person       []*Msg_S2C_TablePerson `json:"person"`       // 牌桌上的用户
	CanWatch     bool                   `json:"canwatch"`     // 是否可以观战
}

// 区域列表详情
type Msg_S2C_Area struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// 财富奖励
type WealthAward struct {
	WealthType int8   `json:"wealth_type"` // 财富类型
	Num        int    `json:"num"`         // 奖励数量
	Url        string `json:"url"`         // 奖励图片
}

// 签到任务详情
type Msg_S2C_TaskCheckin struct {
	Checkin bool           `json:"checkin"`
	Step    int            `json:"step"`
	ReWards []*SignInAward `json:"rewards"`
}

// 签到奖励详情
type SignInAward struct {
	AwardName string `json:"award_name"`
	AwardUrl  string `json:"award_url"`
}

// 用户金币详情
type Msg_S2C_UserGold struct {
	Gold       int `json:"gold"`        //! 金币数量
	InsureGold int `json:"insure_gold"` //! 保险箱金币
}

// 牌桌上的用户
type Msg_S2C_TablePerson struct {
	Id       int64  `json:"id"`       // 用户id
	Nickname string `json:"nickname"` // 昵称
	Imgurl   string `json:"imgurl"`   // 头像
	Ip       string `json:"ip"`       // ip地址
	Online   bool   `json:"online"`   // 是否在线
	Partner  string `json:"partner"`  // 归属队长
}

// 房间解散
type Msg_S2C_TableDel struct {
	Type int8   `json:"type"`
	Msg  string `json:"msg"`
}

// ! 包厢创建
type Msg_S2C_BuyProduct struct {
	Payurl string `json:"payurl"` // 支付链接
}

// ! 离开牌桌
type Msg_S2C_ExitTable struct {
}

// ! 离开牌桌
type Msg_S2C_JoinTableResult struct {
	Result int `json:"result"`
}

// ! 牌桌坐下
type Msg_S2C_TableSeat struct {
	Uid  int64 `json:"uid"`
	Seat int   `json:"seat"` //! 游戏服ip
}

// 定时推送实时在线人数
type Msg_S2C_GameOnline struct {
	KindId   int                       `json:"kind_id"`          // 子游戏id
	SiteType []*Msg_S2C_SiteTypeOnline `json:"site_type_online"` // 房间场次类型在线人数
	Online   int                       `json:"online"`           // 在线人数
}

// 绑定微信结果
type Msg_S2C_BindWechatResult struct {
	Nickname string `json:"nickname"` //! 昵称
	Imgurl   string `json:"imgurl"`   //! 头像
	Sex      int    `json:"sex"`      //! 性别
}

type Msg_S2C_SiteTypeOnline struct {
	Type   int `json:"type"`   // 场次类型
	Online int `json:"online"` // 在线人数
}

// areain成功后返回, 游戏类型及对应的场次类型
type Msg_S2C_GameList struct {
	PackageKey     string              `json:"package_key"`     // 游戏包key
	PackageName    string              `json:"package_name"`    // 游戏包名
	PackageVersion string              `json:"package_version"` // 游戏包版本
	Name           string              `json:"name"`            // 游戏名
	Icon           string              `json:"icon"`            // 游戏图标
	KindId         int                 `json:"kind_id"`         // 子游戏id
	SiteType       []*Msg_S2C_SiteType `json:"site_type"`       // 房间场次类型
	Online         int                 `json:"online"`          // 在线人数
	MatchFlag      int                 `json:"match_flag"`      // 是否开启排位赛，1表示开启，0表示未开启或没有排位赛
	// OrderId        int                 `json:"order_id"`
}

// 场次列表信息
type Msg_S2C_SiteList struct {
	SiteList []*Msg_S2C_SiteType `json:"site_list"` // 房间场次类型
}

type Msg_S2C_SiteType struct {
	Name      string `json:"name"`       // 场次名
	Type      int    `json:"type"`       // 场次类型
	Online    int    `json:"online"`     // 在线人数
	MinScore  int    `json:"min_score"`  // 最小财富限制
	MaxScore  int    `json:"max_score"`  // 最大财富限制
	Difen     int    `json:"difen"`      // 底分
	SitMode   int    `json:"sit_mode"`   // 坐桌模式
	MatchFlag int    `json:"match_flag"` // 是否开启排位赛，1表示开启，0表示未开启或没有排位赛
	Sta       int    `json:"sta"`        // 房间状态0 空闲 1 普通 2 繁忙 3 火爆
}

// 用户区域广播
type Msg_S2C_UserAreaBroadcast struct {
	Uid      int64  `json:"uid"`
	Imgurl   string `json:"imgurl"`   // 用户头像
	Nickname string `json:"nickname"` // 用户昵称
	Content  string `json:"content"`  // 广播内容
}

// 玩家好友房对局历史
type Msg_S2C_GameRecordHistory struct {
	GameNum     string                     `json:"gamenum"`      //! 游戏ID,唯一标识
	RoomNum     int                        `json:"roomid"`       //! 游戏房间ID
	KindId      int                        `json:"kindid"`       //! 游戏玩法标识
	Wf          string                     `json:"wf"`           //! 玩法名
	Time        int64                      `json:"time"`         //! 对局时间(时间戳)
	Icon        string                     `json:"icon"`         //! 游戏icon
	PkgKey      string                     `json:"pkgkey"`       //! 包key
	Player      []*GameRecordHistoryPlayer `json:"playerArr"`    //! 玩家数组
	RoundSum    int                        `json:"round_sum"`    //! 总局数
	RoundPlayed int                        `json:"round_played"` //! 玩了多少局
	PlayerCount int                        `json:"player_count"` //! 玩家数量
}

// 对局详情
type Msg_S2C_GameRecordInfo struct {
	GameNum    string                         `json:"gamenum"`    //! 游戏ID,唯一标识
	KindId     int                            `json:"kindid"`     // 游戏玩法
	RoomId     int                            `json:"roomid"`     // 游戏房间ID
	Time       int64                          `json:"time"`       // 时间
	TotalRound int                            `json:"totalround"` // 总局数
	FloorIndex int                            `json:"floorindex"` // 楼层信息
	UserArr    []*Msg_S2C_GameRecordInfoUser  `json:"userArr"`    // 玩家数组
	ScoreArr   []*Msg_S2C_GameRecordInfoScore `json:"scoreArr"`   // 积分数组
	DiFen      int                            `json:"difen"`      // 当前底分
}

// 对局详情用户信息
type Msg_S2C_GameRecordInfoUser struct {
	Uid         int64   `json:"uid"`          // 用户id
	Nickname    string  `json:"nickname"`     // 用户昵称
	Imgurl      string  `json:"imgurl"`       // 用户头像
	Sex         int     `json:"sex"`          // 性别
	Score       float64 `json:"score"`        // 总分
	Vitamin     float64 `json:"vitamin"`      // 疲劳值
	CapId       int64   `json:"cap_id"`       // 队长ID
	CapNickname string  `json:"cap_nickname"` //队长的昵称
}

// 回放详情
type Msg_S2C_ReplayInfo struct {
	UserArr  []*Msg_S2C_ReplayInfoUser `json:"userArr"`  // 用户信息
	Table    *Msg_S2C_ReplayInfoTable  `json:"table"`    // 牌桌信息
	HandCard string                    `json:"handcard"` // 初始手牌
	OutCard  string                    `json:"outcard"`  // 出牌记录
	EndInfo  string                    `json:"endinfo"`  //小结算详情
}

// 回放牌桌信息
type Msg_S2C_ReplayInfoTable struct {
	HId        int    `json:"hid"`        //! 包厢id
	IsVitamin  bool   `json:"isvitamin"`  //！疲劳值开关
	IsHidHide  bool   `json:"ishidhide"`  //！圈号隐藏
	TId        int    `json:"tid"`        //! 牌桌id
	FId        int    `json:"fid"`        //! 楼层id
	NTId       int    `json:"ntid"`       //! 牌桌索引
	KindId     int    `json:"kindid"`     //! 游戏玩法
	GameConfig string `json:"gameconfig"` //! 游戏配置
	CardsNum   int    `json:"cardsnum"`   //！游戏发完牌之后剩余牌数
}

// 回放用户信息
type Msg_S2C_ReplayInfoUser struct {
	Uid            int64   `json:"uid"`            // 用户id
	Nickname       string  `json:"nickname"`       // 用户昵称
	Imgurl         string  `json:"imgurl"`         // 用户头像
	Sex            int     `json:"sex"`            // 性别
	LastRoundScore float64 `json:"lastroundscore"` // 上一局积分
	Vitamin        float64 `json:"vitamin"`        // 疲劳值
}

// 对局详情用户积分数据
type Msg_S2C_GameRecordInfoScore struct {
	ReplayId  int64     `json:"iReplayid"` // 回放id
	StartTime int64     `json:"starttime"` // 每一局的开始时间
	EndTime   int64     `json:"endtime"`   // 每一局的结束时间
	Score     []float64 `json:"score"`     // 玩家积分数组
	Uids      []int64   `json:"uids"`      // 玩家id数组
}

// 回放详情
type Msg_S2C_CheckReplayId struct {
	KindId int    `json:"kindid"` // 游戏玩法
	PkgKey string `json:"pkgkey"` // 区域包key
}

// ! 包厢创建
type Msg_S2C_HouseCreate struct {
	Id  int64 `json:"id"`
	HId int   `json:"hid"`
}

// ! 包厢删除
type Msg_HC_HouseDelete struct {
	Type   int `json:"type"`
	Num    int `json:"num"`
	Param1 int `json:"param1"`
	Param2 int `json:"param2"`
}

// ! 包厢楼层创建
type Msg_HC_HouseFloorCreate struct {
	Id int64 `json:"fid"`
}

// ! 包厢楼层删除
type Msg_HC_HouseFloorDelete struct {
}

// ! 包厢楼层列表
type Msg_HouseFloorMiniInfo struct {
	FId              int64   `json:"fid"`
	ImageUrl         string  `json:"imageurl"`    // 玩法图标
	KindName         string  `json:"kindname"`    // 玩法名称
	PackageName      string  `json:"packagename"` // 玩法名称
	FRule            FRule   `json:"frule"`
	Name             string  `json:"name"`   // 用户自定义的名称
	IsMix            bool    `json:"is_mix"` // 是否混排
	IsVip            bool    `json:"is_vip"` // 是否为vip楼层
	TableNum         int     `json:"table_num"`
	TableDefault     int     `json:"table_default"` //默认桌数
	HideImg          bool    `json:"hideimg"`
	MinTable         int     `json:"mintable"`
	MaxTable         int     `json:"maxtable"`
	VitaminLowLimit  float64 `json:"vitaminlowlimit"`  // 入桌門檻
	VitaminHighLimit float64 `json:"vitaminhighlimit"` // 入桌門檻
}

type Msg_HC_HouseFloorList struct {
	Items []Msg_HouseFloorMiniInfo `json:"items"`
}

type Msg_HouseItem struct {
	Id            int64    `json:"id"`
	HId           int      `json:"hid"`
	HName         string   `json:"hname"`
	HMems         int      `json:"hmems"`
	OwnerId       int64    `json:"ownerid"`
	OwnerName     string   `json:"ownername"`
	OwnerUrl      string   `json:"ownerurl"`
	OwnerGender   int      `json:"ownergender"`
	Role          int      `json:"role"`
	JoinTime      int64    `json:"jointime"`
	OnlineTable   int      `json:"onlinetable"`
	OnlineCur     int      `json:"onlinecur"`
	OnlineTotal   int      `json:"onlinetotal"`
	MergeHId      int64    `json:"mergehid"`
	IsHidHide     bool     `json:"ishidhide"`
	FloorIDs      []int64  `json:"hfloorids"`
	FloorGameUrls []string `json:"hfloorgameurl"`
}

// ! 玩家入驻包厢列表
type Msg_HC_MemberHouseList struct {
	Items []Msg_HouseItem `json:"items"`
}

// ! 包厢基本信息
type Msg_HC_HouseBaseInfo struct {
	Id                      int64                     `json:"id"`
	HId                     int                       `json:"hid"`
	Area                    string                    `json:"area"`
	LeagueArea              int64                     `json:"league_area"`
	OwnerId                 int64                     `json:"hownerid"`
	UFloor                  int64                     `json:"ufloor"`
	URole                   int                       `json:"urole"`
	Name                    string                    `json:"hname"`
	Notify                  string                    `json:"hnotify"`
	Dialog                  string                    `json:"dialog"`
	DialogActive            bool                      `json:"dialogactive"`
	URefHId                 int                       `json:"urefhid"`
	IsPartner               bool                      `json:"ispartner"`
	SuperiorId              int64                     `json:"superiorid"`
	IsChecked               bool                      `json:"hischecked"`
	IsFrozen                bool                      `json:"hisfrozen"`
	IsMemHide               bool                      `json:"hismemhide"`
	IsMemExit               bool                      `json:"hismemexit"`
	IsActivity              bool                      `json:"isactivity"`
	MaxTable                int                       `json:"hmaxtable"`
	OnlineTable             int                       `json:"onlinetable"`
	OnlineCur               int                       `json:"onlinecur"`
	OnlineTotal             int                       `json:"onlinetotal"`
	FloorIDs                []int64                   `json:"hfloorids"`
	MixFloorIDs             []int64                   `json:"mixhfloorids"`
	CardPool                bool                      `json:"card_pool"`
	Vitamin                 float64                   `json:"vitamin"`       // 用户自己疲劳值
	VitaminPool             float64                   `json:"vitamin_pool"`  // 疲劳值仓库的疲劳值
	IsVitamin               bool                      `json:"isvitamin"`     // 防沉迷开关
	IsGamePause             bool                      `json:"isgamepause"`   // 游戏中到达下限暂停
	IsVitaminHide           bool                      `json:"isvitaminhide"` // 防沉迷管理员可见
	IsVitaminModi           bool                      `json:"isvitaminmodi"` // 防沉迷管理员可调
	IsPartnerHide           bool                      `json:"ispartnerhide"` // 防沉迷队长可见
	IsPartnerModi           bool                      `json:"ispartnermodi"` // 防沉迷队长可见
	IsMemberSend            bool                      `json:"ismembersend"`  // 防成谜成员之间赠送
	IsPartnerApply          bool                      `json:"ipa"`           // 队长是否可以批准加入
	OnlyQucikJoin           bool                      `json:"only_quick"`
	UpdNFIds                []int                     `json:"updnfids"`
	TableJoinType           consts.HouseTableJoinType `json:"house_table_join_type"`
	MixActive               bool                      `json:"mix_active"`
	AutoPayPartner          bool                      `json:"auto_pay_partner"`
	IsHidHide               bool                      `json:"ishidhide"`    // 是否开启了圈号隐藏
	TableShowCount          int                       `json:"tblshowcount"` // 桌子展示个数
	MinTableNum             int                       `json:"mintablenum"`
	MaxTableNum             int                       `json:"maxtablenum"`
	IsAiSuper               bool                      `json:"isaisuper"`        // 是否为超级防作弊模式
	EmptyTableBack          bool                      `json:"empty_table_back"` // 是否空桌子在后面
	EmptyTableMax           int                       `json:"empty_table_max"`  // 最大空桌数
	TableSortType           int                       `json:"table_sort_type"`  // 桌子排序类型 0 正常 1 极左
	IsHeadHide              bool                      `json:"isheadhide"`       // 是否隐藏头像
	DisableSetJuniorVitamin bool                      `json:"disablejuniorv"`   // 禁用队长调整下级比赛分
	VitaminAdmin            bool                      `json:"vitamin_admin"`
	VicePartner             bool                      `json:"vice_partner"` // 副队长
	IsOnlineHide            bool                      `json:"isonlinehide"` // 是否隐藏在线人数
	GameOn                  bool                      `json:"game_on"`
	AdminGameOn             bool                      `json:"admin_game_on"`
	PartnerKick             bool                      `json:"partnerkick"`
	LuckTimes               int64                     `json:"luck_times"` //是否有抽奖机会
	ApplySwitch             bool                      `json:"apply_switch"`
	FloorsColor             []string                  `json:"floors_color"`
	PrivateGPS              bool                      `json:"private_gps"`          // 隐藏地理位置
	FangKaTipsMinNum        int                       `json:"fangka_tips_min_num"`  // 房卡低于xx时提示盟主
	Uright                  map[string]interface{}    `json:"uright"`               //当前用户对应这个圈的权限
	HmSwitch                map[string]int            `json:"hm_switch"`            // 包厢功能开关
	Distance                int                       `json:"distance"`             //距离限制
	RecordTimeInterval      int                       `json:"record_time_interval"` // 战绩筛选时段
	IsLimitGame             bool                      `json:"is_limit_game"`        // 当前用户 是否禁止娱乐  true 禁止  false 没有禁止
	NewTableSortType        int                       `json:"new_table_sort_type"`
	CreateTableType         int                       `json:"create_table_type"`
	RankOpen                bool                      `json:"rank_open"`                 //排行榜是否显示 true 显示 false 隐藏
	IsCurUserTeamOffWork    bool                      `json:"is_cur_user_team_off_work"` //当前用户所在小队是否打烊了 true 打烊了  false 营业
	VipFloorShowSwitch      bool                      `json:"vip_floor_show_switch"`
	UnionSwitch             bool                      `json:"union_switch"`
	NoSkipVitaminSet        bool                      `json:"no_skip_vitamin_set"` // 是否禁止跨级调整vitamin
	IsMemUidHide            bool                      `json:"is_mem_uid_hide"`     // 是否隐藏UID

}

// ! 玩家基本信息
type Msg_HouseMemberItem struct {
	NUId         int                         `json:"nuid"`
	UId          int64                       `json:"uid"`
	UVitamin     float64                     `json:"uvitamin"` //! 疲劳值
	UOnline      bool                        `json:"uonline"`
	UPlaying     bool                        `json:"uplaying"`
	UName        string                      `json:"uname"`
	URole        int                         `json:"urole"`
	UPartner     int64                       `json:"upartner"`
	URefHId      int                         `json:"urefhid"`
	URemark      string                      `json:"uremark"`
	UUrl         string                      `json:"uurl"`
	UGender      int                         `json:"ugender"`
	UJoinTime    int64                       `json:"ujointime"`
	Limit        bool                        `json:"limit"`
	GameLimit    bool                        `json:"game_limit"`
	LastLoginAt  int64                       `json:"ulasttime"`
	UPartnerName string                      `json:"upartnername"`
	UPartnerUrl  string                      `json:"upartnerurl"`
	VitaminAdmin bool                        `json:"vitamin_admin"` // 是否是比赛分管理员
	VicePartner  bool                        `json:"vice_partner"`
	ApplyType    consts.HouseMemberApplyType `json:"apply_type"` // 申请类型
	ApplyAt      int64                       `json:"apply_at"`   // 申请时间
	Superior     int64                       `json:"superior"`
	SuperiorName string                      `json:"superiorname"`
	TeamBan      bool                        `json:"team_ban"`
	AA           bool                        `json:"aa"`
}

type MsgHouseMemberItemWrapper struct {
	Hms []*Msg_HouseMemberItem
	By  func(i, j *Msg_HouseMemberItem) bool
}

func (m *MsgHouseMemberItemWrapper) Len() int {
	return len(m.Hms)
}

func (m *MsgHouseMemberItemWrapper) Less(i, j int) bool {
	return m.By(m.Hms[i], m.Hms[j])
}

func (m *MsgHouseMemberItemWrapper) Swap(i, j int) {
	m.Hms[i], m.Hms[j] = m.Hms[j], m.Hms[i]
}

type LimitUserInfo struct {
	UId     int64  `json:"uid"`
	UName   string `json:"uname"`
	UUrl    string `json:"uurl"`
	UGender int    `json:"ugender"`
	Limit   bool   `json:"limit"`
}

type LimitUserSlie []*LimitUserInfo

func (p LimitUserSlie) Len() int           { return len(p) }
func (p LimitUserSlie) Less(i, j int) bool { return p[i].UId < p[j].UId }
func (p LimitUserSlie) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ! 包厢创建
type HouseVitaminRecord struct {
	Id            int64   `json:"id"`
	UpdatedTime   int64   `json:"updatedtime"`
	BefVitamin    float64 `json:"befvitamin"`
	OptType       string  `json:"opt_type"`
	ChangeVitamin float64 `json:"change_vitamin"`
	OptUName      string  `json:"opt_name"`
	UId           int64   `json:"uid"`
	AftVitamin    float64 `json:"aftvitamin"`
	OptTypeInt    int64   `json:"opt_typeint"`
}
type Msg_S2C_HouseVitaminRecords struct {
	Items          []*HouseVitaminRecord `json:"items"`
	CurrentVitamin float64               `json:"current_vitamin"`
	UName          string                `json:"uname"`
}

type MSgHouseVitaminPoolRecord struct {
	Items []*HouseVitaminRecord `json:"items"`
	// Total         float64               `json:"total"`       //总疲劳值，固定100万
	// PoolLeft      float64               `json:"pool_left"`   // 疲劳值剩余
	// Used          float64               `json:"pool_used"`   //已使用
	// WaitJoin      float64               `json:"wait_join"`   //待入账，当前为实时入账，一般为0
	TotalCount int `json:"total_count"` //记录总数
	// LastShouldPay float64               `json:"last_should_pay"`
	// LastPaied     float64               `json:"last_paied"`
	// EarnSum       float64               `json:"earn_sum"`
}

type Msg_HC_HouseMemberOnline struct {
	UNums   int     `json:"unums"`
	Vitamin float64 `json:"vitamin"`
}

// ! 包厢牌桌信息
type Msg_HouseTableItem struct {
	NTId         int              `json:"ntid"`
	TId          int              `json:"tid"`
	ATId         int              `json:"atid"`
	TRule        LessFRule        `json:"trule"`
	TMemItems    []FloorTableUser `json:"tmemitems"`
	Begin        bool             `json:"begin"`
	Step         int              `json:"step"`
	Deleted      bool             `json:"deleted"`
	CanWatch     bool             `json:"canwatch"`
	WatcherIcons []string         `json:"watchericons"`
}

type LessFRule struct {
	PlayerNum int   `json:"playernum"` // 游戏开始人数
	RoundNum  int   `json:"roundnum"`  // 游戏局数
	KindId    int   `json:"kindid"`    // 游戏玩法
	Difen     int64 `json:"difen"`     // 底分
}

type FloorTableUser struct {
	UId      int64  `json:"uid"`
	UName    string `json:"uname"`
	UUrl     string `json:"uurl"`
	UGender  int    `json:"ugender"`
	IsOnline bool   `json:"online"`
	Ready    bool   `json:"ready"`
}

// ! 包厢玩家信息
type Msg_HC_HouseMemList struct {
	Totalnum             int                    `json:"totalnum"`
	HMemOnLineNum        int                    `json:"hmemonlinenum"`
	HMemNum              int                    `json:"hmemnum"`
	FMems                []*Msg_HouseMemberItem `json:"hmemitems"`
	PBegin               int                    `json:"pbegin"`
	PEnd                 int                    `json:"pend"`
	PartnerMemsNum       int                    `json:"partnermemsnum"`
	PartnerMemsOnlineNum int                    `json:"partnermemsonlinenum"`
	LimitUserNum         int                    `json:"limit_user_num"`
}

type MsgHouseMemberIn struct {
	NFIds []int                   `json:"nfids"`
	FIds  []int64                 `json:"fids"`
	Acks  []*Msg_HC_HouseMemberIn `json:"acks"`
}

// !玩家进入包厢
type Msg_HC_HouseMemberIn struct {
	FId         int64                `json:"fid"`
	KindName    string               `json:"kname,omitempty"`
	PackageKey  string               `json:"package_key,omitempty"`
	FTableItems []Msg_HouseTableItem `json:"ftableitems"`
	MaxMatchNum int                  `json:"m_num"`
	TotalNum    int                  `json:"t_num"`
}

// !玩家进入包厢
type GameHfInfo struct {
	Infos          []*Msg_HC_HouseMemberIn `json:"infos"`
	IsMix          bool                    `json:"is_mix"`
	TableJoinType  int                     `json:"mix_type"`
	IsPartnerApply bool                    `json:"is_partner_apply"`
}

type MixFloorAcks struct {
	FIDS []int64                 `json:"fids"`
	Acks []*Msg_HC_HouseMemberIn `json:"acks"`
}

// ! 包厢桌子详细信息
type HouseTableUserItem struct {
	NTid     int    `json:"ntid"`
	UId      int64  `json:"uid"`
	NickName string `json:"nickname"`
	Url      string `json:"url"`
	Ip       string `json:"ip"`
}
type Msg_HC_HouseTableBaseInfo struct {
	NTId        int `json:"ntid"`
	TId         int `json:"tid"`
	CurProgress int `json:"curprogress"`
	MaxProgress int `json:"maxprogress"`
}

// ! 包厢玩家信息
type Msg_HC_HousePartnerList struct {
	Totalnum int                   `json:"totalnum"`
	FMems    []Msg_HouseMemberItem `json:"hmemitems"`
}

// ! 包厢玩家信息
type Msg_HC_HouseFloorVipUsers struct {
	Hid      int                       `json:"hid"`
	Fid      int64                     `json:"fid"`
	Totalnum int                       `json:"totalnum"`
	FMems    []Msg_HouseMemberLiteItem `json:"items"`
}

// ! 玩家基本信息
type Msg_HouseMemberLiteItem struct {
	UId     int64  `json:"uid"`
	UName   string `json:"uname"`
	UUrl    string `json:"uurl"`
	UGender int    `json:"ugender"`
}

// ! 包厢任命队长
type Ntf_HC_HousePartnerCreate struct {
	HId int   `json:"hid"`
	UId int64 `json:"uid"`
}

// ! 包厢队长任命下级队长
type Ntf_HC_HousePartnerJunior struct {
	HId int   `json:"hid"` // 包厢id
	Opt int64 `json:"opt"` // 队长id
	Uid int64 `json:"uid"` // 下级
}

type Ntf_HC_HouseParnterBindSuperior struct {
	Hid        int   `json:"hid"`
	Opt        int64 `json:"opt"`
	ParnterId  int64 `json:"parnterid"`
	SuperiorId int64 `json:"superiorid"`
}

// ! 包厢卸职队长
type Ntf_HC_HousePartnerDelete struct {
	HId int   `json:"hid"`
	UId int64 `json:"uid"`
}

// ! 包厢防沉迷开关
type Ntf_HC_HouseVitaminStatus struct {
	HId           int     `json:"hid"`
	Status        bool    `json:"status"`        // 疲劳值开关 false关 true开
	AdminHide     bool    `json:"adminhide"`     // 管理员可见
	AdminModi     bool    `json:"adminmodi"`     // 管理员可调
	PartnerHide   bool    `json:"partnerhide"`   // 队长可见
	GamePause     bool    `json:"gamepause"`     // 到达下限暂定游戏
	LowLimit      float64 `json:"lowlimit"`      // 疲劳值下限
	LowLimitPause float64 `json:"lowlimitpause"` // 游戏暂停下限
	DeductCount   float64 `json:"playdeduct"`    // 对局扣除
	DeductType    int     `json:"bwdeduct"`      // 0大赢家扣 其他AA
}

// ! 包厢防沉迷管理员课件
type Ntf_HC_HouseVitaminOptHide struct {
	OptHide bool `json:"opthide"`
	Hid     int  `json:"hid"`
}

// ! 包厢疲劳值数值修改
type Msg_HC_HouseVitaminSet struct {
	Value float64 `json:"value"`
}

// ! 包厢疲劳值数值修改
type Msg_HC_HouseVitaminSet_Ntf struct {
	HId     int     `json:"hid"`
	OptId   int64   `json:"optid"`
	OptRole int     `json:"optrole"`
	UId     int64   `json:"uid"`
	Value   float64 `json:"value"`
}

type Msg_HC_HouseVitaminChange_Ntf struct {
	UVitamin float64 `json:"uvitamin"`
}

// ! 包厢疲劳值数值一键修改
type Msg_HC_HouseVitaminAdminSet_Ntf struct {
	HId int `json:"hid"`
}

type Msg_HC_HouseCardStatistics struct {
	Hid         int                              `json:"hid"`
	DFid        int                              `json:"dfid"`
	SelectTime  int                              `json:"selecttime"`
	TotalTable  int                              `json:"totaltable"`  // 总局数
	TotalPlayer int                              `json:"totalplayer"` // 总日活
	TotalCard   int                              `json:"totalcard"`   // 总房卡消耗
	Items       []Msg_HC_HouseCardStatisticsItem `json:"items"`       // 队长数据
	PBegin      int                              `json:"pbeg"`
	PEnd        int                              `json:"pend"`
}

type Msg_HC_HouseCardStatisticsItem struct {
	UId      int64  `json:"uid"`
	UName    string `json:"uname"`
	UUrl     string `json:"uurl"`
	UGender  int    `json:"ugender"`
	Card     int    `json:"card"`
	CardCost int    `json:"card_cost"`
	Round    int    `json:"round"`
	Player   int    `json:"player"`
}

type Msg_HC_HouseCardStatisticsItem_Wrapper struct {
	Items  []Msg_HC_HouseCardStatisticsItem `json:"items"` // 队长数据
	LessFn func(a, b *Msg_HC_HouseCardStatisticsItem) bool
}

func (m *Msg_HC_HouseCardStatisticsItem_Wrapper) Len() int {
	return len(m.Items)
}

func (m *Msg_HC_HouseCardStatisticsItem_Wrapper) Less(i, j int) bool {
	return m.LessFn(&m.Items[i], &m.Items[j])
}

func (m *Msg_HC_HouseCardStatisticsItem_Wrapper) Swap(i, j int) {
	m.Items[i], m.Items[j] = m.Items[j], m.Items[i]
}

// ! 包厢疲劳值数值一键修改
type Msg_HC_HouseKick_Ntf struct {
	HId   int   `json:"hid"`    // 圈号
	OptId int64 `json:"opt_id"` // 谁踢的人
	Pid   int64 `json:"pid"`    // 踢的哪个队长
}

type MsgVitaminStatisticItem struct {
	// 玩家基本信息
	UId         int64   `json:"uid"`          // 玩家id
	UName       string  `json:"uname"`        // 昵称
	UUrl        string  `json:"uurl"`         // 头像
	UGender     int     `json:"ugender"`      // 性别
	URole       int     `json:"urole"`        // 角色
	UVitamin    float64 `json:"uvitamin"`     // 个人赛分
	Partner     int64   `json:"partner"`      // 合伙人id，1代表队长身份， 0代表为圈主名下玩家， 其他代表属于哪个队长名下玩家
	Superior    int64   `json:"superior"`     // 上级队长id，当 Partner == 1时 Superior > 0 表示他为小队长
	AlarmValue  float64 `json:"alarm_value"`  // 个人警戒值
	PartnerType int     `json:"partner_type"` // 队长类型，0=非队长(普通玩家 右侧无按钮) 1=队长(显示详情按钮) 2=上级队长(显示下级按钮)

	// 统计信息
	PlayerNum int `json:"player_num"` // 旗下玩家人数(总人数)

	VitaminLeft     float64 `json:"vitaminleft"`     // 总赛分
	VitaminMinus    float64 `json:"vitaminminus"`    // 负数总额（总负分）
	VitaminCost     float64 `json:"vitamincost"`     // 总房费（总赛点）
	VitaminLeftInt  int64   `json:"vitaminleftint"`  // 总赛分
	VitaminMinusInt int64   `json:"vitaminminusint"` // 负数总额（总负分）
	VitaminCostInt  int64   `json:"vitamincostint"`  // 总房费（总赛点）

	VitaminCostRound float64 `json:"vitamincostround"` // AA房费扣除
	VitaminCostBW    float64 `json:"vitamincostbw"`    // 大赢家房费扣除
	VitaminWinLose   float64 `json:"vitaminwinlose"`   // 游戏输赢

	VitaminCostRoundInt int64 `json:"vitamincostroundint"` // AA房费扣除
	VitaminCostBWInt    int64 `json:"vitamincostbwint"`    // 大赢家房费扣除
	VitaminWinLoseInt   int64 `json:"vitaminwinloseint"`   // 游戏输赢

	// 其他字段
	Exp int64 `json:"exp"` // 队长经验值
}

type MsgVitaminStatisticItemWrapper struct {
	Item []MsgVitaminStatisticItem
	By   func(p, q *MsgVitaminStatisticItem) bool
}

func (pw MsgVitaminStatisticItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw MsgVitaminStatisticItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw MsgVitaminStatisticItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(&pw.Item[i], &pw.Item[j])
}

type MsgRewardStatisticItem struct {
	UId        int64   `json:"uid"`
	UName      string  `json:"uname"`
	UUrl       string  `json:"uurl"`
	UGender    int     `json:"ugender"`
	CurVitamin float64 `json:"curvitamin"`
	CurReward  float64 `json:"curreward"`
}

type MsgRewardStatisticItemWrapper struct {
	Item []MsgRewardStatisticItem
	By   func(p, q *MsgRewardStatisticItem) bool
}

func (pw MsgRewardStatisticItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw MsgRewardStatisticItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw MsgRewardStatisticItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(&pw.Item[i], &pw.Item[j])
}

// ! 包厢队长疲劳值统计
type Msg_HC_HouseRewardStatistic struct {
	Items  []MsgRewardStatisticItem `json:"items"`
	At     int64                    `json:"at"`
	PBegin int                      `json:"pbeg"`
	PEnd   int                      `json:"pend"`
}

// ! 包厢队长疲劳值统计
type Msg_HC_HouseVitaminStatistic struct {
	Items                 []MsgVitaminStatisticItem `json:"items"`
	TotalVitaminCost      float64                   `json:"totalvitamincost"`
	TotalVitaminCostRound float64                   `json:"totalvitamincostround"`
	TotalVitaminCostBW    float64                   `json:"totalvitamincostbw"`
	TotalVitaminLeft      float64                   `json:"totalvitaminleft"`
	TotalVitaminMinus     float64                   `json:"totalvitaminminus"`
	TotalVitaminWinLose   float64                   `json:"totalvitaminwinlose"`
	PBegin                int                       `json:"pbeg"`
	PEnd                  int                       `json:"pend"`
}

type MsgVitaminStatisticClearItem struct {
	VitaminCost      float64 `json:"vitamincost"`
	VitaminCostRound float64 `json:"vitamincostround"`
	VitaminCostBW    float64 `json:"vitamincostbw"`
	VitaminLeft      float64 `json:"vitaminleft"`
	VitaminMinus     float64 `json:"vitaminminus"`
	VitaminPayment   float64 `json:"vitaminpayment"`
	BeginAt          int64   `json:"beginat"`
	EndAt            int64   `json:"endat"`
}

// ! 包厢疲劳值统计
type Msg_HC_HouseVitaminStatisticClear struct {
	Items []MsgVitaminStatisticClearItem `json:"items"`
}

type MsgVitaminMgrItem struct {
	UId                   int64   `json:"uid"`
	UName                 string  `json:"uname"`
	UUrl                  string  `json:"uurl"`
	UGender               int     `json:"ugender"`
	URole                 int     `json:"urole"`
	IsPartner             bool    `json:"ispartner"`
	CurVitamin            float64 `json:"curvitamin"`
	PreNodeVitamin        float64 `json:"prenodevitamin"`
	VitaminWinLoseCost    float64 `json:"vitaminwinlosecost"`
	VitaminPlayCost       float64 `json:"vitaminplaycost"`
	VitaminCostRound      float64 `json:"vitamincostround"`
	VitaminCostBW         float64 `json:"vitamincostbw"`
	IsJunior              bool    `json:"isjunior"`
	VitaminAdmin          bool    `json:"vitamin_admin"`
	VicePartner           bool    `json:"vice_partner"`
	UPartner              int64   `json:"upartner"`
	VitaminWinLoseCostInt int64
	VitaminPlayCostInt    int64
	VitaminCostRoundInt   int64
	VitaminCostBWInt      int64
	PlayTimes             int     `json:"playtimes"`
	BwTimes               int     `json:"bwtimes"`
	TotalScore            float64 `json:"totalscore"`
	ValidTimes            int     `json:"validtimes"`
	InValidTimes          int     `json:"invalidtimes"`
	BigValidTimes         int     `json:"bigvalidtimes"`
}

type MsgVitaminMgrItemWrapper struct {
	Item []*MsgVitaminMgrItem
	By   func(p, q *MsgVitaminMgrItem) bool
}

func (pw MsgVitaminMgrItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw MsgVitaminMgrItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw MsgVitaminMgrItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(pw.Item[i], pw.Item[j])
}

// ! 包厢疲劳值管理成员列表
type Msg_HC_HouseVitaminMgrList struct {
	Items    []*MsgVitaminMgrItem `json:"items"`
	UPartner int64                `json:"upartner"`
	PBegin   int                  `json:"pbeg"`
	PEnd     int                  `json:"pend"`
}

////////////////////////////////////////////////////////////////////////////////
// 统计消息

type HouseRecordBBItem struct {
	UId   int64  `json:"uid"`
	UName string `json:"uname"`
	UUrl  string `json:"uurl"`
	Times int    `json:"times"`
}

// ! 包厢大赢家统计
type Msg_CH_HouseRecordBW struct {
	Items []HouseRecordBBItem `json:"items"`
}

// ! 包厢对局统计
type Msg_CH_HouseRecordBattle struct {
	Items []HouseRecordBBItem `json:"items"`
}

type HouseMemberStatisticsItemWrapper struct {
	Item []*HouseMemberStatisticsItem
	By   func(p, q *HouseMemberStatisticsItem) bool
}

func (pw HouseMemberStatisticsItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw HouseMemberStatisticsItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw HouseMemberStatisticsItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(pw.Item[i], pw.Item[j])
}

type HouseMemberStatisticsItem struct {
	UId           int64   `json:"uid"`
	UName         string  `json:"uname"`
	UUrl          string  `json:"uurl"`
	UGender       int     `json:"ugender"`
	UJoinTime     int64   `json:"ujointime"`
	PlayTimes     int     `json:"playtimes"`
	BwTimes       int     `json:"bwtimes"`
	TotalScore    float64 `json:"totalscore"`
	ValidTimes    int     `json:"validtimes"`
	InValidTimes  int     `json:"invalidtimes"`
	BigValidTimes int     `json:"bigvalidtimes"`
	Partner       int64   `json:"partner"`
	GroupIndex    int     `json:"groupindex"`
	IsLike        bool    `json:"islike"`
	IsExit        bool    `json:"isexit"`
}

type HousePartnerStatisticsItem struct {
	Uid           int64
	Playtimes     int
	Bwtimes       int
	Totalscore    float64
	Validtimes    int
	Bigvalidtimes int
}

// ! 包厢成员统计数据
type Msg_HC_HouseMemberStatistics struct {
	Items                []*HouseMemberStatisticsItem `json:"items"`
	PBegin               int                          `json:"pbegin"`
	PEnd                 int                          `json:"pend"`
	PartnerMemsNum       int                          `json:"partner_mems_num"`
	PartnerMemsPlayedNum int                          `json:"partner_mems_played_num"`

	// 盟主,管理员,裁判->牌局数,低分局数
	TotalRound   int `json:"totalround"`
	InValidRound int `json:"invalidround"`

	// 队长,副队长->总战绩,总人次,低分局人次
	TotalScore            float64 `json:"totalscore"`
	TotalPlayTimes        int     `json:"totalplaytimes"`
	TotalInValidPlayTimes int     `json:"totalinvalidplaytimes"`

	LikeCount int `json:"likecount"`
}

type Msg_HC_HouseMemberStatisticsTotal struct {
	PlayTimes     int     `json:"playtimes"`
	BwTimes       int     `json:"bwtimes"`
	TotalScore    float64 `json:"totalscore"`
	ValidTimes    int     `json:"validtimes"`
	BigValidTimes int     `json:"bigvalidtimes"`
}

type TotalRecordStatItem struct {
	PlayTimes int `json:"playtimes"`
	BwTimes   int `json:"bwtimes"`
}

type HouseRecordItemWrapper struct {
	Item []HouseRecordItem
	By   func(p, q *HouseRecordItem) bool
}

func (pw HouseRecordItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw HouseRecordItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw HouseRecordItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(&pw.Item[i], &pw.Item[j])
}

type HouseRecordItem struct {
	GameNum   string  `json:"gamenum"`
	PlayCount int     `json:"playcount"`
	WinScore  float64 `json:"winscore"`
	IsHeart   int     `json:"isheart"`
	IsJoin    int     `json:"isjoin"`
	WriteDate int64   `json:"writedate"`
}

// ! 包厢我的战绩
type Msg_HC_HouseMyRecord struct {
	MyStatlist []TotalRecordStatItem `json:"mystatlist"`
	Items      []HouseRecordItem     `json:"items"`
}

// ! 包厢战绩
type Msg_HC_HouseRecord struct {
	Items []HouseRecordItem `json:"items"`
}

// ! 包厢我的战绩
type Msg_HC_HouseGameRecord struct {
	Items []GameRecordDetal `json:"items"`
	UId   int64             `json:"uid"`

	TotalRound    int `json:"totalround"`    //总局数
	ValidRound    int `json:"validround"`    //有效局
	InValidRound  int `json:"invalidround"`  //低分局
	CompleteRound int `json:"completeround"` //完整局数
	DismissRound  int `json:"dismissround"`  //解散局数

	TotalScore   float64 `json:"totalscore"`   //总战绩
	PlayTimes    int     `json:"playtimes"`    //总人次
	InvalidTimes int     `json:"invalidtimes"` //低分局人次
	TotalBWTimes int     `json:"totalbwtimes"` //大赢家人次

	TotalLike int `json:"totallike"`
}

type HouseOperationalStatusItem struct {
	QueryTime       int64              `json:"querytime"`
	PlayRounds      [MaxFloorIndex]int `json:"playrounds"`
	TotalRounds     int                `json:"totalrounds"`
	TotalFangkaCost int                `json:"totalfangkacost"`
}

// ! 包厢经营状况
type Msg_HC_HouseOperationalStatus struct {
	Items []HouseOperationalStatusItem `json:"items"`
}

type HouseValidRoundScoreGetItem struct {
	FId   int64   `json:"fid"`
	Score float64 `json:"score"`
}

// ! 包厢有效对局积分查询
type Msg_HC_HouseValidRoundScoreGet struct {
	Items        []HouseValidRoundScoreGetItem `json:"items"`
	InvalidRound int                           `json:"invalidround"`
	HId          int                           `json:"hid"`
}

// ! 包厢有效对局积分设置
type Msg_HC_HouseValidRoundScoreSet struct {
	Score int `json:"score"`
}

type HouseGameRecordItem struct {
	GameNum    string `json:"gamenum"`
	RoomId     string `json:"roomid"`
	GameIndex  int    `json:"gameindex"`
	KindId     int    `json:"kindid"`
	PlayRound  int    `json:"playround"`
	TotalRound int    `json:"totalround"`
	Floor      int    `json:"floor"`
	WanFa      string `json:"wf"`

	WinScore int `json:"winscore"`
	IsHeart  int `json:"isheart"`
	IsJoin   int `json:"isjoin"`
}

// ！大赢家统计项
type HouseRecordStatusItem struct {
	UId         int64  `json:"uid"`
	UName       string `json:"uname"`
	UUrl        string `json:"uurl"`
	RecordTimes int    `json:"recordtimes"`
}

// ! 大赢家统计
type Msg_HC_HouseRecordStatus struct {
	Items  []HouseRecordStatusItem `json:"items"`
	PBegin int                     `json:"pbegin"`
	PEnd   int                     `json:"pend"`
}

// ！战绩点赞功能
type Msg_HC_HouseRecordHeart struct {
}

// ! 房卡统计
type Msg_HC_HouseRecordKaCost struct {
	Items []RecordGameCostMini `json:"items"`
}

// ！大赢家对局清除
type Msg_HC_HouseRecordStatusClean struct {
}

// ！大赢家对局清除所有
type Msg_HC_HouseRecordStatusCleanAll struct {
}

//！ 对局统计
////////////////////////////////////////////////////////////////////////////////
// 通知消息

// ! 玩家过审
type Ntf_HC_HouseMemberAgree struct {
	UId int64 `json:"uid"`
	HId int64 `json:"hid"`
}

// ! 玩家拒审
type Ntf_HC_HouseMemberRused struct {
	UId int64 `json:"uid"`
	HId int   `json:"hid"`
}

// ! 包厢名称公告修改
type Ntf_HC_HouseBaseNNmodify struct {
	HId           int    `json:"hid"`
	HName         string `json:"hname"`
	HNotify       string `json:"hnotify"`
	HDialog       string `json:"hdialog"`
	HDialogActive bool   `json:"hdialogactive"`
}

// ! 包厢设置审核
type Ntf_HC_HouseOptIsMemCheck struct {
	HId       int  `json:"hid"`
	IsChecked bool `json:"ischecked"`
}

// ! 包厢设置审核
type Ntf_HC_HouseOptIsMemExitCheck struct {
	HId       int  `json:"hid"`
	IsChecked bool `json:"ischecked"`
}

// ! 包厢设置冻结
type Ntf_HC_HouseOptIsFrozen struct {
	HId      int  `json:"hid"`
	IsFrozen bool `json:"isfrozen"`
}

// ! 包厢设置成员列表隐藏
type Ntf_HC_HouseOptIsMemHide struct {
	HId       int  `json:"hid"`
	IsMemHide bool `json:"ismemhide"`
}

// ! 包厢设置成员列表隐藏
type Ntf_HC_HouseOptAutoPay struct {
	HId     int  `json:"hid"`
	AutoPay bool `json:"auto_pay"`
}

// ! 包厢设置成员列表隐藏
type Ntf_HC_HouseOptIsHidHide struct {
	HId       int  `json:"hid"`
	IsHidHide bool `json:"ishidhide"`
}

// ! 包厢设置头像隐藏
type Ntf_HC_HouseOptIsHeadHide struct {
	HId        int  `json:"hid"`
	IsHeadHide bool `json:"isheadhide"`
}

// ! 包厢设置Uid隐藏
type Ntf_HC_HouseOptIsMemUidHide struct {
	HId          int  `json:"hid"`
	IsMemUidHide bool `json:"is_mem_uid_hide"`
}

// ! 包厢设置头像隐藏
type Ntf_HC_HouseOptIsOnlineHide struct {
	HId          int  `json:"hid"`
	IsOnlineHide bool `json:"isonlinehide"`
	OnlineTable  int  `json:"onlinetable"`
	OnlineCur    int  `json:"onlinecur"`
	OnlineTotal  int  `json:"onlinetotal"`
}

// ! 包厢设置头像隐藏
type NtfHCHouseOptPartnerKick struct {
	HId         int  `json:"hid"`
	PartnerKick bool `json:"partnerkick"`
}

// ! 包厢设置成员位置信息隐藏
type NtfHCHouseGpsHide struct {
	HId        int  `json:"hid"`
	PrivateGPS bool `json:"privategps"`
}

// ! 包厢设置奖励均衡
type NtfHCHouseOptRewardBalanced struct {
	HId            int  `json:"hid"`
	RewardBalanced bool `json:"reward_balanced"`
}

// ! 包厢成员剔除
type Ntf_HC_HouseMemberKick struct {
	HId int `json:"hid"`
}

// ! 包厢楼层玩法修改
type Ntf_HC_HouseFloorKIdModify struct {
	HId              int     `json:"hid"`
	FId              int64   `json:"fid"`
	ImageUrl         string  `json:"imageurl"`    // 玩法图标
	KindName         string  `json:"kindname"`    // 玩法名称
	PackageName      string  `json:"packagename"` // 玩法名称
	FRule            FRule   `json:"frule"`
	VitaminLowLimit  float64 `json:"vitaminlowlimit"`  // 入桌限制
	VitaminHighLimit float64 `json:"vitaminhighlimit"` // 入桌門檻

}

// ! 包厢成员过审
type Ntf_HC_HouseMemberRoleGen struct {
	HId    int    `json:"hid"`
	UId    int64  `json:"uid"`
	OURole int    `json:"oldurole"`
	URole  int    `json:"urole"`
	HName  string `json:"hname"`
}

// ! 包厢删除
type Ntf_HC_HouseDelete struct {
	HId int `json:"hid"`
}

// ! 包厢楼层创建
type Ntf_HC_HouseFloorCreate struct {
	HId              int     `json:"hid"`
	FId              int64   `json:"fid"`
	ImageUrl         string  `json:"imageurl"`    // 玩法图标
	KindName         string  `json:"kindname"`    // 玩法名称
	PackageName      string  `json:"packagename"` // 玩法名称
	FRule            FRule   `json:"frule"`
	VitaminLowLimit  float64 `json:"vitaminlowlimit"`
	VitaminHighLimit float64 `json:"vitaminhighlimit"` // 入桌門檻
}

// ! 包厢楼层删除
type Ntf_HC_HouseFloorDelete struct {
	HId int   `json:"hid"`
	FId int64 `json:"fid"`
}

// ! 包厢牌桌玩家进入
type Ntf_HC_HouseTableIn struct {
	HId     int    `json:"hid"`
	FId     int64  `json:"fid"`
	NTId    int    `json:"ntid"`
	NUId    int    `json:"nuid"`
	UId     int64  `json:"uid"`
	UName   string `json:"uname"`
	URole   int    `json:"urole"`
	URemark string `json:"uremark"`
	UUrl    string `json:"uurl"`
	UGender int    `json:"ugender"`
}

// ! 包厢牌桌玩家退出
type Ntf_HC_HouseTableOut struct {
	HId  int   `json:"hid"`
	FId  int64 `json:"fid"`
	NTId int   `json:"ntid"`
	UId  int64 `json:"uid"`
	NUId int   `json:"nuid"`
}

// ! 包厢牌桌创建
type Ntf_HC_HouseTableCreate struct {
	HId        int   `json:"hid"`
	FId        int64 `json:"fid"`
	NTId       int   `json:"ntid"`
	TId        int   `json:"tid"`
	Begin      bool  `json:"begin"`
	Step       int   `json:"step"`
	TotalRound int   `json:"total_round"`
}

// ! 包厢牌桌解散
type Ntf_HC_HouseTableDissovel struct {
	HId  int   `json:"hid"`
	FId  int64 `json:"fid"`
	NTId int   `json:"ntid"`
}

// ! 包厢楼层活动列表
type ActListItem struct {
	ActId       int64         `json:"actid"`
	ActName     string        `json:"actname"`
	ActState    int           `json:"actstate"`
	ActHideInfo bool          `json:"hideinfo"`
	Type        int64         `json:"type"`
	Rewords     *[]RewordInfo `json:"rewords"`
}
type Msg_HC_HouseActList struct {
	ActItems []*ActListItem `json:"actitmes"`
}

// ! 包厢楼层活动数据
type ActRecordItem struct {
	ActId       int64   `json:"actid"`
	UId         int64   `json:"uid"`
	UName       string  `json:"uname"`
	Score       float64 `json:"score"`
	Rank        int64   `json:"rank"`
	CreatedTime int64   `json:"created_time"`
}
type Msg_HC_HouseActInfo struct {
	FIds        []int64          `json:"fids"`
	FIdIndexs   []int64          `json:"fidindexs"`
	ActId       int64            `json:"actid"`
	ActName     string           `json:"actname"`
	ActType     int              `json:"acttype"`
	ActBegTime  int64            `json:"begtime"`
	ActEndTime  int64            `json:"endtime"`
	ActState    int              `json:"actstate"`
	ActHideInfo bool             `json:"hideinfo"`
	UserItems   []*ActRecordItem `json:"useritems"`
	Type        int64            `json:"type"`
}

// ! 任务列表
type TaskItem struct {
	Id       int `json:"id"`     // 任务id
	MainType int `json:"type"`   // 任务主类型   // 0每日任务  1系统任务
	SubType  int `json:"kind"`   // 任务子类型   // 0 分享 1 对局次数 2 胜场次数 3 每日签到
	Num      int `json:"num"`    // 任务当前计数
	TgtNum   int `json:"tgtnum"` // 任务目标计数
	//Step      int    `json:"step"`        // 任务当前进度（阶段任务 暂无）
	//TotalStep int    `json:"totalstep"`   // 任务总进度
	Sta        int    `json:"sta"`        // 任务状态 0 进行中  1完成未领取 2 已领取
	Desc       string `json:"desc"`       // 任务描述
	RewardDesc string `json:"rewarddesc"` // 奖励描述
	GameKindId int    `json:"gamekindid"` // 与任务相关的游戏kind id
	Order      int    `json:"order"`      // 排序标识
}

// ! 任务列表
type TaskGameItem struct {
	Id             int    `json:"id"`             //! 任务id
	Type           int    `json:"type"`           //! 任务类型   // 0每日任务  1系统任务
	Kind           int    `json:"kind"`           //! 任务类别   // 0对战 1分享 2胜利 10充值
	Num            int    `json:"num"`            //! 任务当前进度
	ComNum         int    `json:"comp"`           //! 任务上限
	Step           int    `json:"step"`           //! 任务进度
	TotalStep      int    `json:"totalstep"`      //! 任务总进度
	KindId         int    `json:"kindid"`         //! 玩法类型
	SiteType       int    `json:"sitetype"`       // 场次类型 0表示所有场次
	Describe       string `json:"describe"`       //! 描述
	IsRewarded     int    `json:"isrewarded"`     //! 是否已领取
	RewardType     string `json:"rewardtype"`     //! 任务奖励类型
	ReWardCount    int    `json:"rewardcount"`    //! 阶段奖励数量
	ReWardCountMax int    `json:"rewardcountmax"` //! 阶段奖励数量区间大值
}

// ! 任务数据提交
type Msg_S2C_TaskDataUpdate struct {
	Id     int  `json:"id"`     //! 任务id 若为0，表示所有prokind的类型的任务都修改
	Kind   int  `json:"kind"`   //! 任务类别   // 0对战 1分享 2胜利 10充值
	Add    bool `json:"add"`    // true表示在原来的数据上增加pronum，false表示覆盖
	ProNum int  `json:"pronum"` //! 完成局数
	ConNum int  `json:"connum"` //! 局数上限
}

type Msg_HC_TaskList struct {
	Items []*TaskItem `json:"items"`
}

// 游戏房间内任务
type Msg_SC_TaskList struct {
	Items []*TaskGameItem `json:"items"`
}

// ! 任务领取
type Msg_HC_TaskAward struct {
	Task *TaskItem `json:"task"`
}

// ! 游戏房间内任务领取
type Msg_SC_TaskAward struct {
	Count     int           `json:"count"`     //! 还有多少个任务的奖励没有领
	ReWardNum int           `json:"rewardnum"` //! 本次领了多少奖励
	Task      *TaskGameItem `json:"task"`
}

// ！ 金币场
// //////////////////////////////////////////////////////////////////////////////
// ! 场次信息
type Site_SC_Msg struct {
	SiteId     int                    `json:"siteid"`
	SitMode    int                    `json:"sit_mode"`
	KindId     int                    `json:"kindid"`
	SiteType   int                    `json:"sitetype"`
	TableNum   int                    `json:"tablenum"`
	Person     *Person_SC_Msg         `json:"person"`
	GameConfig map[string]interface{} `json:"gameconfig"`
}
type Person_SC_Msg struct {
	Uid      int64  `json:"uid"`
	Ntid     int    `json:"ntid"`     // 桌号
	Imgurl   string `json:"imgurl"`   // 头像
	Nickname string `json:"nickname"` // 昵称
	Sex      int    `json:"sex"`      // 性别
	Gold     int    `json:"gold"`     // 金币数
}

type Site_SC_Table_Info struct {
	Index   int   `json:"index"`
	TableId int64 `json:"tableid"`
}

type Site_SC_UserInfo struct {
	Uid        int64  `json:"uid"`
	Imgurl     string `json:"imgurl"`      // 头像
	Nickname   string `json:"nickname"`    // 昵称
	Sex        int    `json:"sex"`         // 性别
	Gold       int    `json:"gold"`        // 金币数量
	WinCount   int    `json:"win_count"`   // 胜利局数
	LostCount  int    `json:"lost_count"`  // 失败局数
	DrawCount  int    `json:"draw_count"`  // 和局局数
	FleeCount  int    `json:"flee_count"`  // 逃跑局数
	TotalCount int    `json:"total_count"` // 总局数
}

// ! 通过包key取得rule
type Msg_HC_GameRules struct {
	Games AreaGameRuleCompiledList `json:"games"`
}

// ↑
type Son_HC_GameRule struct {
	GameRuleVersion int    `json:"game_rule_version"`
	ClientVersion   int    `json:"client_version"`
	KindId          int    `json:"kind_id"`
	Name            string `json:"name"`
	PackageKey      string `json:"package_key"`
	GameRule        string `json:"rule"`
	Engine          int    `json:"engine"`
	PackageVersion  string `json:"package_version"`
}

// 排位赛相关协议，获取所有排位赛列表
type Msg_S2C_MatchGameList struct {
	Name      string `json:"name"`       // 游戏名
	Icon      string `json:"icon"`       // 游戏图标
	KindId    int    `json:"kind_id"`    // 子游戏id
	SiteType  int    `json:"site_type"`  // 房间场次类型
	SitMode   int    `json:"sit_mode"`   // 坐桌模式
	Online    int    `json:"online"`     // 在线人数
	MatchFlag int    `json:"match_flag"` // 是否开启排位赛，1表示开启，0表示未开启或没有排位赛
	State     int    `json:"state"`      // 排位赛是否开始，0表示未开始，1表示进行中，2表示已经结束
	BeginTime int64  `json:"begintime"`  // 开始时间,只取时间
	EndTime   int64  `json:"endtime"`    // 结束时间
}

// 获取排位赛奖励列表
type Site_SC_MatchAwardList struct {
	KindId        int                      `json:"kind_id"`   // 游戏id
	SiteType      int                      `json:"site_type"` // 场次类型
	WinAward      int                      `json:"winaward"`  // 玩家胜利一次奖励数
	JoinAward     int                      `json:"joinaward"` // 玩家对局一次的奖励数，逃跑不算
	MatchAwardOrg []*Site_SC_MatchAwardOrg `json:"matchawardorg"`
	BeginDate     int64                    `json:"begindate"` // 开始日期
	EndDate       int64                    `json:"enddate"`   // 结束日期
	BeginTime     int64                    `json:"begintime"` // 开始时间
	EndTime       int64                    `json:"endtime"`   // 结束时间
}
type Site_SC_MatchAwardOrg struct {
	AwardType  int    `json:"awardtype"`  // 奖励类别，礼券或话费，0表示礼券，1表示话费
	Awards     int    `json:"awards"`     // 排名对应的奖励数目
	AwardStr   string `json:"awardstr"`   // 奖励的文案提醒
	LowerRange int    `json:"lowerrange"` // 排名范围的起始点  4-10名 奖励100个礼券，11-50名奖励1元话费
	UpperRange int    `json:"upperrange"` // 排名范围的终点
}

// 获取排位赛排位列表
type Site_SC_MatchRankingList struct {
	MatchKey           string                        `json:"match_key"`     //! 排位赛编号
	KindId             int                           `json:"kind_id"`       // 游戏id
	SiteType           int                           `json:"site_type"`     // 场次类型
	State              int                           `json:"state"`         // 排位赛状态，0表示未开始，1表示进行中，2表示已经结束
	MyRanking          int                           `json:"myranking"`     // 我的排名
	MyAward            int                           `json:"myaward"`       // 我已经获得的礼券，
	InRankingFlag      int                           `json:"inrankingflag"` // 我是否入榜
	MatchRankingPerson []*Site_SC_MatchRankingPerson `json:"rankingperson"`
	BeginDate          int64                         `json:"begindate"` // 开始日期
	EndDate            int64                         `json:"enddate"`   // 结束日期
	BeginTime          int64                         `json:"begintime"` // 开始时间
	EndTime            int64                         `json:"endtime"`   // 结束时间
	LeftTime           int64                         `json:"lefttime"`  // 结束时间
}
type Site_SC_MatchRankingPerson struct {
	UId          int64  `json:"uid"` // 玩家id
	Name         string `json:"name"`
	Icon         string `json:"icon"`         // 玩家图标
	Sex          int    `json:"sex"`          // 性别，1为男，2为女
	Score        int    `json:"score"`        // 当前得分,扣除茶水的
	LowerCoupon  int    `json:"lowercoupon"`  // 已经获得的礼券，起始勋章
	UpperCoupon  int    `json:"uppercoupon"`  // 已经获得的礼券，终止勋章
	LowerRanking int    `json:"lowerranking"` // 排名,游戏没有结束时大家都是0，
	UpperRanking int    `json:"upperranking"` // 排名,游戏没有结束时大家都是0，
	Award        int    `json:"award"`        // 奖励礼券数
	ConStr       string `json:"constr"`       // 奖励介绍
	TotalCount   int    `json:"totalcount"`   // 玩家总对局次数，逃跑局不算
	UpdatedTime  int64  `json:"updatedtime"`  // 最后更新时间
}
type MatchRankingPersonItemWrapper struct {
	Item []*Site_SC_MatchRankingPerson
	By   func(p, q *Site_SC_MatchRankingPerson) bool
}

func (pw MatchRankingPersonItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw MatchRankingPersonItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw MatchRankingPersonItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(pw.Item[i], pw.Item[j])
}

type MsgHouseCardPoolChange struct {
	Hid    int64 `json:"hid"`
	Status bool  `json:"status"`
}

type MsgHouseFloorRenameNTF struct {
	Fid  int64  `json:"fid"`
	Name string `json:"name"`
}

type MsgHouseFloorMix struct {
	FIDs     []int64 `json:"fids"`
	TableNum int     `json:"table_num"`
}

type MsgHouseFloorMixEditor struct {
	Hid               int                       `json:"hid"`
	FIDs              []int64                   `json:"fids"`
	TableNum          int                       `json:"table_num"`
	IsActie           bool                      `json:"mix_active"`
	AICheck           bool                      `json:"ai_check"`
	AISuper           bool                      `json:"ai_super"`
	AITotalScoreLimit int                       `json:"ai_total_score_limit"` // 触发智能防作弊的条件之一：总分上限
	TableJoinType     consts.HouseTableJoinType `json:"house_table_join_type"`
	EmptyTableBack    bool                      `json:"empty_table_back"`    // 是否空桌子在后面
	EmptyTableMax     int                       `json:"empty_table_max"`     // 最大空桌数
	TableSortType     int                       `json:"table_sort_type"`     // 桌子排序类型 0 正常 1 极左
	NewTableSortType  int                       `json:"new_table_sort_type"` //  新版本桌子排序
	CreateTableType   int                       `json:"create_table_type"`   // 开桌类型
}

type MsgRespHouseMixInfo struct {
	IsMix             bool                      `json:"is_mix"`
	MixActive         bool                      `json:"mix_active"`
	TableNum          int                       `json:"table_num"`
	FIDs              []int64                   `json:"fids"`
	AICheck           bool                      `json:"ai_check"`
	AISuper           bool                      `json:"ai_super"`
	AITotalScoreLimit int                       `json:"ai_total_score_limit"` // 触发智能防作弊的条件之一：总分上限
	TableJoinType     consts.HouseTableJoinType `json:"house_table_join_type"`
	EmptyTableBack    bool                      `json:"empty_table_back"`    // 是否空桌子在后面
	EmptyTableMax     int                       `json:"empty_table_max"`     // 最大空桌数
	TableSortType     int                       `json:"table_sort_type"`     // 桌子排序类型 0 正常 1 极左
	NewTableSortType  int                       `json:"new_table_sort_type"` //  新版本桌子排序
	CreateTableType   int                       `json:"create_table_type"`   // 开桌类型
}

// ! 包厢禁止同桌玩家信息
type Msg_HC_HouseMemTableLimitList struct {
	Totalnum int                   `json:"totalnum"`
	FMems    []Msg_HouseMemberItem `json:"hmemitems"`
}

type GroupInfo struct {
	GroupID   int              `json:"group_id"`
	UserCount int              `json:"user_count"`
	CreateAt  string           `json:"create_at"`
	Users     []*LimitUserInfo `json:"users"`
}

type LimitGroups []*GroupInfo

func (p LimitGroups) Len() int           { return len(p) }
func (p LimitGroups) Less(i, j int) bool { return p[i].GroupID < p[j].GroupID }
func (p LimitGroups) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type HouseMemUsers []*Msg_HouseMemberItem

func (p HouseMemUsers) Len() int           { return len(p) }
func (p HouseMemUsers) Less(i, j int) bool { return p[i].NUId < p[j].NUId }
func (p HouseMemUsers) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Msg_HC_HouseTableLimitInfo struct {
	Is2PNotEffect bool         `json:"is2pnoteffect"`
	Groups        []*GroupInfo `json:"groups"`
	TotalGroup    int          `json:"total_group"`
}

type Msg_HC_DeliveryInfoUpd struct {
	DeliveryImg string `json:"delivery_img"` // 用户个人名片图片所在地址
}

type Msg_HC_HTInvitationLetter struct {
	TId       int    `json:"tid"`       // 牌桌号
	NTId      int    `json:"ntid"`      // 牌桌索引
	HId       int    `json:"hid"`       // 包厢id(6位)
	IsHidHide bool   `json:"ishidhide"` // 是否隐藏圈号
	DHId      int64  `json:"dhid"`      // 包厢id(数据库自增id)
	FId       int64  `json:"fid"`       // 楼层id(数据库自增id)
	NFId      int    `json:"nfid"`      // 楼层索引
	KindId    int    `json:"kindid"`    // 游戏类型
	GameName  string `json:"game_name"` // 游戏玩法名字
	FName     string `json:"fname"`     // 楼层名字
	Inviter   struct {
		Uid      int64  `json:"uid"`      // id
		Imgurl   string `json:"imgurl"`   // 头像
		Nickname string `json:"nickname"` // 昵称
		Gender   int    `json:"gender"`   // 性别
	} `json:"inviter"` // 邀请人信息
}

type Msg_HC_HouseInviteJoin struct {
	HId       int  `json:"hid"`       // 包厢id(6位)
	IsHidHide bool `json:"ishidhide"` // 是否隐藏圈号
	Inviter   struct {
		Uid      int64  `json:"uid"`      // id
		Imgurl   string `json:"imgurl"`   // 头像
		Nickname string `json:"nickname"` // 昵称
		Gender   int    `json:"gender"`   // 性别
	} `json:"inviter"` // 邀请人信息
}

func (m *Msg_HC_HouseInviteJoin) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m *Msg_HC_HouseInviteJoin) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

type Msg_HouseRound_NTF struct {
	Hid        int   `json:"hid"`
	Fid        int64 `json:"fid"`
	Ntid       int   `json:"ntid"`
	Begin      bool  `json:"begin"`
	Step       int   `json:"step"`
	Tid        int   `json:"tid"`
	TotalRound int   `json:"total_round"`
}

type Msg_HC_HouseMemOnline struct {
	Hid    int   `json:"hid"`
	Uid    int64 `json:"uid"`
	Online bool  `json:"on_line"`
}

type Msg_UserSubFloor struct {
	Hid            int                     `json:"hid"`
	Fid            int64                   `json:"fid"`
	IsMix          bool                    `json:"is_mix"`
	IsAnonymous    bool                    `json:"isanonymous"` // 是否开启了匿名游戏功能(与超级防作弊一样，但是每个游戏自己设置，不是包厢设置的)
	MixType        int                     `json:"mix_type"`
	IsPartnerApply bool                    `json:"is_partner_apply"`
	TableInfo      []*Msg_HC_HouseMemberIn `json:"table_info"`
	OnlineUser     []*GameMember           `json:"online_users"`
	ApplyUser      []*GameMember           `json:"apply_users"`
	UserInfo       *GameMember             `json:"user_info"`
}

// GameMember 游戏服包厢成员结构
type GameMember struct {
	Uid         int64                       `json:"uid"`
	Uname       string                      `json:"uname"`
	URole       int                         `json:"urole"`
	Gender      int                         `json:"gender"`
	IsOnline    bool                        `json:"is_online"`
	ApplyTime   int64                       `json:"apply_time"`
	AgreeTime   int64                       `json:"agree_time"`
	UUrl        string                      `json:"uurl"`
	IsLimitGame bool                        `json:"is_limit_game"`
	Partner     int64                       `json:"partner"`
	IsInGame    bool                        `json:"-"`
	Hid         int                         `json:"hid"`
	ApplyType   consts.HouseMemberApplyType `json:"apply_type"`
}

type MsgHFloorInfo struct {
	Hid int64 `json:"hid"`
	Fid int64 `json:"fid"`
}

type MsgUserInOutGame struct {
	Uid int64 `json:"uid"`
}

type MsgFloorDeductInfo struct {
	Fid                  int64   `json:"fid"`
	IsVitaminLowest      bool    `json:"il"` // 是否勾选了单局低于
	IsVitaminHighest     bool    `json:"ih"` // 是否勾选了单局高于
	VitaminLowest        float64 `json:"l"`  // 单局结算低于
	VitaminHighest       float64 `json:"h"`  // 单局结算高于（或等于）
	VitaminLowestDeduct  float64 `json:"ld"` // 单局结算低于扣除值
	VitaminHighestDeduct float64 `json:"hd"` // 单局结算高于或等于 扣除值
	VitaminDeductCount   float64 `json:"dn"` // 扣费额度
	VitaminDeductType    int     `json:"dt"` // 扣费方式 0大赢家 1AA
}

type MsgFloorEffectInfo struct {
	Fid                  int64   `json:"fid"`
	IsVitamin            bool    `json:"isvitamin"`            // 防沉迷开关
	IsGamePause          bool    `json:"isgamepause"`          // 游戏中到达下限暂停
	VitaminLowLimit      float64 `json:"vitaminlowlimit"`      // 疲劳值入桌下限
	VitaminHighLimit     float64 `json:"vitaminhighlimit"`     // 入桌門檻
	VitaminLowLimitPause float64 `json:"vitaminlowlimitpause"` // 疲劳值暂停下限
}

type MsgHouseFloorDeductInfo struct {
	Hid   int                  `json:"hid"`
	Items []MsgFloorDeductInfo `json:"items"`
}

type MsgHouseFloorEffectInfo struct {
	Hid   int                  `json:"hid"`
	Items []MsgFloorEffectInfo `json:"items"`
}

type MsgHouseMergeCheck struct {
	HId      int  `json:"hid"`
	IsFrozen bool `json:"isfrozen"`
	IsGaming bool `json:"isgaming"`
}

type MsgHouseMergeRecords struct {
	Items []MsgHouseMergeRecordItem `json:"items"`
}

type MsgHouseMergeRecordItem struct {
	At    int64  `json:"at"`
	State int64  `json:"state"`
	THid  int    `json:"thid"`
	Hid   int    `json:"hid"`
	Msg   string `json:"msg"`
}

type MsgHouseMergeResultNtf struct {
	HId  int  `json:"hid"`
	THId int  `json:"thid"`
	Ok   bool `json:"ok"`
	Num  int  `json:"num"`
}

type MsgHouseJoinTableSet struct {
	Hid       int  `json:"hid"`
	OnlyQuick bool `json:"only_quick"`
}

type MsgHouseMergeRevokeOk struct {
	Hid    int    `json:"hid"`
	THid   int    `json:"thid"`
	THName string `json:"thname"`
	Msg    string `json:"msg"`
}

type Msg_HC_NotifyPush struct {
	Msg string `json:"msg"`
}

type MsgHouseOwnerChange struct {
	Hid      int    `json:"hid"`
	Uid      int64  `json:"uid"`
	PassCode string `json:"pass_code"`
}

type Msg_HC_HouseParnterRoyaltyGet struct {
	Hid             int       `json:"hid"`
	ParnterId       int64     `json:"parnterid"`
	NickName        string    `json:"nickname"`
	Royaltys        []float64 `json:"royaltys"`
	SuperiorProfit  []float64 `json:"superiorprofit"`
	Distributable   []float64 `json:"distributable"`
	RoyaltyPercent  []int     `json:"royalty_percent"`
	SuperiorPercent []int     `json:"superior_percent"`
}

type HouseParnterSuperiorListItem struct {
	UId     int64  `json:"uid"`
	UName   string `json:"uname"`
	UUrl    string `json:"uurl"`
	UGender int    `json:"ugender"`
}

type Msg_HC_HouseParnterSuperiorList struct {
	Hid        int                            `json:"hid"`
	SuperiorId int64                          `json:"superiorid"`
	CurPage    int                            `json:"curpage"`
	TotalPage  int                            `json:"totalpage"`
	Items      []HouseParnterSuperiorListItem `json:"items"`
}

type ClubPartnerFloorStatisticsItem struct {
	UId               int64  `json:"uid"`
	UName             string `json:"uname"`
	UUrl              string `json:"uurl"`
	UGender           int    `json:"ugender"`
	ValidTimes        int    `json:"validtimes"`        //有效局
	BigValidTimes     int    `json:"bigvalidtimes"`     //超级有效局
	RoundProfit       int    `json:"roundprofit"`       //单局收益
	SubordinateProfit int    `json:"subordinateprofit"` //下级收益
	TotalProfit       int    `json:"totalprofit"`       //总收益
	Royalty           int    `json:"royalty"`           //单局收益设置
	IsJunior          bool   `json:"isjunior"`          //上级id
	ChangeProfit      int64  `json:"changeprofit"`
	VitaminAdmin      bool   `json:"vitamin_admin"`
	VicePartner       bool   `json:"vice_partner"`
	PartnerDeep       int    `json:"partner_deep"`
	Superior          int64  `json:"superior"`
	Exp               int64  `json:"exp"` //经验值
}

type ClubPartnerFloorStatisticsItemWrapper struct {
	Item []*ClubPartnerFloorStatisticsItem
	By   func(p, q ClubPartnerFloorStatisticsItem) bool
}

func (pw ClubPartnerFloorStatisticsItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw ClubPartnerFloorStatisticsItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw ClubPartnerFloorStatisticsItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(*pw.Item[i], *pw.Item[j])
}

type Msg_HC_HouseParnterFloorStatistics struct {
	Hid   int                               `json:"hid"`
	Items []*ClubPartnerFloorStatisticsItem `json:"items"`
}

type ClubPartnerFloorHistoryStatisticsItem struct {
	DFid        int64 `json:"dfid"`
	Fid         int   `json:"fid"`
	TotalProfit int   `json:"totalprofit"` //总收益
	ValidTimes  int   `json:"validtimes"`  //有效局

	Royalty           int `json:"royalty"`           //单局收益设置
	SubordinateProfit int `json:"subordinateprofit"` //下级收益

	DeteleTime int64
}

type ClubPartnerFloorHistoryStatisticsItemWrapper struct {
	Item []*ClubPartnerFloorHistoryStatisticsItem
	By   func(p, q ClubPartnerFloorHistoryStatisticsItem) bool
}

func (pw ClubPartnerFloorHistoryStatisticsItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw ClubPartnerFloorHistoryStatisticsItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw ClubPartnerFloorHistoryStatisticsItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(*pw.Item[i], *pw.Item[j])
}

type Msg_HC_HouseParnterFloorHistoryStatistics struct {
	Items []*ClubPartnerFloorHistoryStatisticsItem `json:"items"`
}

type Msg_HouseAutoPayPartner struct {
	Hid     int  `json:"hid"`
	AutoPay bool `json:"auto_pay"`
}

type Msg_HC_HousePartnerInviteCode struct {
	Code string `json:"code"`
}

type Msg_HC_HouseTableShowCount struct {
	HId         int `json:"hid"`
	Count       int `json:"count"`
	MinTableNum int `json:"mintablenum"`
	MaxTableNum int `json:"maxtablenum"`
}

type Msg_HC_HouseOwnerRoyaltyGet struct {
	Hid            int       `json:"hid"`
	ParnterId      int64     `json:"parnterid"`
	NickName       string    `json:"nickname"`
	Royaltys       []float64 `json:"royaltys"`
	JuniorProfit   []float64 `json:"junior_profit"`
	SingleCost     []float64 `json:"single_cost"`
	JuniorPercent  []int     `json:"junior_percent"`
	RoyaltyPercent []int     `json:"royalty_percent"`
}

type MsgHouseFloorMappingUpdate struct {
	CurrentMappingNum int `json:"current_mapping_num"` // 当前匹配中人数
	TotalMappingNum   int `json:"total_mapping_num"`   // 匹配总数
}

type MsgHouseGroupInfo struct {
	Groups     []*GroupInfo `json:"groups"`
	TotalGroup int          `json:"total_group"`
	Uid        int64        `json:"uid"`
}

type GroupUserListInfo struct {
	UserInfo []*LimitUserInfo `json:"users"`
	Start    int              `json:"start"`
	GroupId  int              `json:"group_id"`
}
type HousePartnerRoyaltyForMeItem struct {
	FloorIndex int     `json:"floorindex"`
	FloorName  string  `json:"floorname"`
	MyRoyalty  float64 `json:"myroyalty"`
}

type Msg_HC_HousePartnerRoyaltyForMe struct {
	Item []HousePartnerRoyaltyForMeItem `json:"item"`
}

type HousePartnerRoyaltyHistoryItem struct {
	CreatedAt     int64  `json:"createdat"`
	OptUserType   string `json:"optusertype"`
	OptFloorIndex int    `json:"optfloorindex"`
	OptFloorName  string `json:"optfloorname"`
	OptInfo       string `json:"optinfo"`
}

type Msg_HC_HousePartnerRoyaltyHistory struct {
	Uid  int64                            `json:"uid"`
	Name string                           `json:"name"`
	Item []HousePartnerRoyaltyHistoryItem `json:"item"`
}

type HouseNoLeagueStatisticsItem struct {
	UId          int64  `json:"uid"`
	UName        string `json:"uname"`
	UUrl         string `json:"uurl"`
	UGender      int    `json:"ugender"`
	ChangeProfit int64  `json:"changeprofit"`
	Bwtimes      int    `json:"bwtimes"`
	Playtimes    int    `json:"playtimes"`
	Invalidtimes int    `json:"invalidtimes"`
	ParnterLevel int    `json:"parnterlevel"`
	Superior     int64  `json:"superior"`
	SuperiorName string `json:"superiorname"`
	IsLike       bool   `json:"islike"`
	IsLimit      bool   `json:"islimit"`
	IsExit       bool   `json:"isexit"`
}

type HouseNoLeagueStatisticsItemWrapper struct {
	Item []*HouseNoLeagueStatisticsItem
	By   func(p, q *HouseNoLeagueStatisticsItem) bool
}

func (pw HouseNoLeagueStatisticsItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw HouseNoLeagueStatisticsItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw HouseNoLeagueStatisticsItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(pw.Item[i], pw.Item[j])
}

type Msg_HC_NoLeagueStatistics struct {
	Items        []*HouseNoLeagueStatisticsItem `json:"items"`
	Total        int                            `json:"total"`
	PBegin       int                            `json:"pbegin"`
	PEnd         int                            `json:"pend"`
	Allplaytimes int                            `json:"allplaytimes"`
	Capplaytimes int                            `json:"capplaytimes"`
}

type MsgMemLuckCount struct {
	Count int64 `json:"count"`
	Hid   int   `json:"hid"`
	ActId int64 `json:"actid"`
}

type MsgMemGetLuck struct {
	CountLeft int64 `json:"countleft"`
	Hid       int   `json:"hid"`
	ActId     int64 `json:"actid"`
	Rank      int64 `json:"rank"`
}

type MsgLuckDetail struct {
	Hid        int   `json:"hid"`
	Total      int64 `json:"total_count"`
	UserLeft   int64 `json:"user_left_count"`
	ActStart   int64 `json:"act_start"`
	ActEnd     int64 `json:"act_end"`
	ActId      int64 `json:"actid"`
	ActTicket  int64 `json:"act_ticket"`
	LeftTicket int64 `json:"left_ticket"`
}

type MsgHouseFloorVipUsers struct {
	Hid      int               `json:"hid"`
	Fid      int64             `json:"fid"`
	Vipers   []*MsgVipUserItem `json:"vipers"`
	Everyman []*MsgVipUserItem `json:"everyman"`
}

type MsgVipUserItem struct {
	Uid      int64  `json:"uid"`
	NickName string `json:"nickname"` //! 昵称
	ImgUrl   string `json:"imgurl"`   //! 头像
	Sex      int    `json:"sex"`      //! 性别 1男 其他女
}

// ! 包厢经验状况
type Msg_HC_TableDistanceLimitGet struct {
	Distance int `json:"distance"` //限制距离
}

// ! 包厢经验状况
type Msg_HC_TableDistanceLimitSet struct {
	HId      int `json:"hid"`      //包厢id
	Distance int `json:"distance"` //限制距离
}

// ! 包厢名称公告修改
type Ntf_HC_HousePartnerGen struct {
	HId        int   `json:"hid"`
	OptId      int64 `json:"optid"`
	Uid        int64 `json:"uid"`
	OldPartner int64 `json:"oldpartner"`
	NewPartner int64 `json:"newpartner"`
}

type Msg_HC_HouseMemberTrackPoint struct {
	HId        int   `json:"hid"`
	ApplyCount int64 `json:"apply_count"`
}

// 游戏合集
type Msg_S2C_GameCollections struct {
	ID       int                 `json:"id"`       // 游戏类型ID
	Name     string              `json:"name"`     // 游戏类型名(扑克、麻将、娱乐)
	GameList []*Msg_S2C_GameList `json:"gamelist"` // 游戏列表
}

// 商品兑换
type Msg_S2C_Shop_Exchange struct {
	Code     int             `json:"code"`      // 0:成功；1：失败（礼卷不足）；2：没有绑定手机号；3：其他（失败）
	Id       int             `json:"id"`        // 商品ID
	Gold     int             `json:"gold"`      // 金币
	Card     int             `json:"card"`      // 房卡
	GoldBean int             `json:"gold_bean"` // 礼卷
	Product  []*Shop_Product `json:"product"`   // 最新商品列表信息
}

type Msg_S2C_Shop_Product struct {
	Product []*Shop_Product `json:"product"`
}

type Shop_Product struct {
	Id    int    `json:"id"`    // 商品ID
	Left  int    `json:"left"`  // 剩余数量
	Price int    `json:"price"` // 价格
	Name  string `json:"name"`  // 商品名称
	Image string `json:"image"` // 商品图片
	Order int    `json:"order"` // 排序
	Type  int    `json:"type"`  // 商品类型，固定为1=金币
	Num   int    `json:"num"`   // 数量, 客户端显示要/100
	Gift  int    `json:"gift"`  // 赠送, 客户端显示要/100
}

// 获取兑换记录
type Msg_S2C_Shop_Record struct {
	Record []*Shop_Record `json:"record"`
	Page   int            `json:"page"`  // 页签
	Total  int            `json:"total"` // 总页数
}

// ↑
type Shop_Record struct {
	Id     int    `json:"id"`     // 订单ID
	Name   string `json:"name"`   // 商品名称
	Price  int    `json:"price"`  // 花费
	Time   string `json:"time"`   // 创建时间
	Num    int    `json:"num"`    // 数量
	Status int    `json:"status"` // 发放状态：0:待发放；1：已发放
}

// 获取奖励记录
type Msg_S2C_Shop_GoldRecord struct {
	Record []*Gold_Record `json:"record"`
	Page   int            `json:"page"`  // 页签
	Total  int            `json:"total"` // 总页数
}

// ↑
type Gold_Record struct {
	Id       int    `json:"id"`       // 订单号
	Name     string `json:"name"`     // 商品名称
	MatchKey string `json:"matchkey"` // 排位赛编号
	Type     int    `json:"type"`     // 类型
	Time     string `json:"time"`     // 创建时间
	Num      int    `json:"num"`      // 数量
}

// 绑定手机号
type Msg_S2C_Shop_PhoneBind struct {
	Code int    `json:"code"` // 0:成功，1：失败
	Msg  string `json:"msg"`
}

// 获取绑定手机号
type Msg_S2C_Shop_GetPhoneBind struct {
	Code int    `json:"code"` // 0:成功，1：失败
	Tel  string `json:"tel"`  // 手机号码
}

type Item_GVoiceMember struct {
	Id  int   `json:"id"` //member
	Uid int64 `json:"uid"`
}

type Msg_S2C_GVoiceMember struct {
	User []*Item_GVoiceMember `json:"user"`
}

// 根据邀请码加入包厢
type Msg_HC_TablePackageInfo struct {
	Tid            int    `json:"tid"`             // 桌子号
	PackageKey     string `json:"package_key"`     // 游戏包key
	PackageName    string `json:"package_name"`    // 游戏包名
	PackageVersion string `json:"package_version"` // 游戏包版本
	GVoice         bool   `json:"gvoice"`          // 是否开启了实时语音
}

type Msg_S2C_HousePrivateGPSGet struct {
	PrivateGPS bool `json:"privategps"` // 是否开启了实时语音
}

type Msg_S2C_RecommendGoldGame struct {
	RecommendGame string `json:"recommendgame"` // 游戏包名
}

// 分享配置信息
type Msg_Http2C_ShareCfg struct {
	ShareId           int             `json:"share_id"`     // 分享id
	SceneId           int             `json:"scene_id"`     // 分享场景id(0 大厅主界面分享 1 大厅任务界面分享 2 小结算界面分享 3 礼券界面分享)
	ShareTo           int             `json:"share_to"`     // 分享到哪儿 0 朋友圈 1 微信好友 3 朋友圈和微信好友
	ShareType         int             `json:"share_type"`   // 分享类型 0 文字 1 图片 2 图文 3 链接
	ShareTimes        int             `json:"share_times"`  // 是否限制分享次数 0 不限制次数
	Title             string          `json:"title"`        // 分享的标题
	Content           string          `json:"content"`      // 分享的内容
	ImgDownload       string          `json:"img_download"` // 图片下载地址
	Link              string          `json:"link"`         // 跳转地址
	Reward            ShareRewardList // 分享奖励
	AlreadyShareTimes int             `json:"already_share_times"` // 已分享次数
}

// 分享成功
type Msg_Http2C_ShareSuc struct {
	Reward   ShareRewardList // 分享获得奖励
	Gold     int             `json:"gold"`     // 玩家金币数量
	Diamond  int             `json:"diamond"`  // 玩家钻石数量
	Card     int             `json:"card"`     // 玩家房卡数量
	GoldBean int             `json:"goldbean"` // 玩家礼券数据
	Vitamin  float64         `json:"vitamin"`  // 剩余比赛点
}

// 分享奖励列表
type ShareRewardList []*ShareReward

// 分享奖励
type ShareReward struct {
	WealthType int `json:"wealth_type"` // 奖励的财富类型 固定为2=金币
	Num        int `json:"num"`         // 奖励数量
}

type Msg_S2C_TaskCompletedCount struct {
	Count int `json:"count"`
}

type Msg_S2C_IsCheckIn struct {
	Checkin bool `json:"checkin"`
}

type Msg_S2C_AllowancesInfo struct {
	Gift      *AllowanceGift      `json:"gift"`
	Allowance *Msg_S2C_Allowances `json:"allowance"`
}

// 设置房卡低于xx时提示盟主
type Msg_S2C_SetFangKaTipsMinNum struct {
	Hid    int `json:"hid"`
	MinNum int `json:"minnum"`
}

// ! 搜索用户数据(返回给客户端)
type Msg_S2C_SearchUser struct {
	Uid      int64  `json:"uid"`      //! 用户数字id
	Nickname string `json:"nickname"` //! 昵称
	Imgurl   string `json:"imgurl"`   //! 头像
}

// 查询
type Msg_S2C_HmUserRight struct {
	Right map[string]interface{} `json:"right"`
}

type Msg_S2C_UpdateHmUserRight struct {
	Hid         int    `json:"hid"`
	Uid         int64  `json:"uid"`
	UpdateRight string `json:"update_right"`
}

type Msg_S2C_HmSwitch struct {
	HId    int            `json:"hid"`
	Switch map[string]int `json:"switch"`
}

// 设置战绩筛选时段
type Msg_S2C_SetRecordTimeInterval struct {
	Hid          int `json:"hid"`
	TimeInterval int `json:"timeinterval"`
}

// 发送排行榜的设置
type Msg_S2C_HouseMemberRankGet struct {
	RankRound  int  `json:"rank_round"`  // 局数排行榜 开启：0001  隐藏成员：0010  隐藏对应数据：0100
	RankWiner  int  `json:"rank_winer"`  // 赢家排行榜 开启：0001  隐藏成员：0010  隐藏对应数据：0100
	RankRecord int  `json:"rank_record"` // 战绩排行榜 开启：0001  隐藏成员：0010  隐藏对应数据：0100
	RankOpen   bool `json:"rank_open"`   //开启排行榜 按钮显示作用
}

// 排行榜数据
type HouseMemberRankItem struct {
	UId      int64   `json:"uid"`
	UName    string  `json:"uname"`
	UUrl     string  `json:"uurl"`
	UGender  int     `json:"ugender"`
	URankNum float64 `json:"rank_num"` //排行榜数值
}

type Msg_S2C_HouseMemberRankData struct {
	UserItem []*HouseMemberRankItem `json:"user_item"`
	PBegin   int                    `json:"pbegin"`
	PEnd     int                    `json:"pend"`
}

type HouseMemberRankDataWrapper struct {
	Item []*HouseMemberRankItem
	By   func(p, q *HouseMemberRankItem) bool
}

func (rdw HouseMemberRankDataWrapper) Len() int {
	return len(rdw.Item)
}

func (rdw HouseMemberRankDataWrapper) Less(i, j int) bool {
	return rdw.By(rdw.Item[i], rdw.Item[j])
}

func (rdw HouseMemberRankDataWrapper) Swap(i, j int) {
	rdw.Item[i], rdw.Item[j] = rdw.Item[j], rdw.Item[i]
}

// 发送排行榜的设置
type Msg_S2C_HouseDismissRoomDet struct {
	DismissTime string `json:"dismiss_time"` //! 解散时间
	DismissType int    `json:"dismiss_type"` //! 解散类型   0：无,  1：盟主解散，2：管理员解散 3：队长解散 4：申请解散 5：超时解散  6：托管解散
	DismissDet  string `json:"dismiss_det"`  //! 详情
}

// 未实名、已实名未成年的游戏时长 返回
type Msg_Http2C_PlayTime struct {
	TimeSec int `json:"timesec"` // 当天游戏时长 秒为单位（跨天服务器重置前一天的累计时长 将最新的累计时长返回给客户端）
}

// 今日游戏时长已达上限 将限制游戏
type Msg_S2C_LimitPlayTime struct {
	Code int    `json:"code"` // 1 未实名 2 已实名但未成年 3 不在指定时间范围内游戏
	Tips string `json:"tips"` // 弹窗提示
}

type Msg_S2C_LeftCards struct {
	Ok             bool                `json:"ok"`
	RepertoryCards [MAX_REPERTORY]byte `json:"repertory_cards"`
	// ErrInfo        string              `json:"err_info"`
}

type Msg_S2C_PeepCard struct {
	Ok bool `json:"ok"`
	*Msg_S2C_PlayCard
}

type Msg_S2C_WantGood struct {
	Ok bool `json:"ok"`
}

type Msg_S2C_PlayCard struct {
	CardData  [40]byte `json:"carddata"`  //扑克列表
	CardCount byte     `json:"cardcount"` //扑克列表长度
}

type Msg_S2C_BattleLevel struct {
	BattleRound int                         `json:"battle_round"` // 总对局数
	Configs     []Msg_S2C_BattleLevelConfig `json:"configs"`
}

type Msg_S2C_BattleLevelConfig struct {
	Id    int64  `json:"id"`    // 配置ID
	Level int    `json:"level"` // 等级
	Desc  string `json:"desc"`  // 等级描述
	Limit int    `json:"limit"` // 等级门槛，从大到小找，总对局数>=门槛就在此等级
}

type Msg_S2C_SpinInfo struct {
	SpinTimes int                 `json:"spin_times"` // 当前可玩次数
	Awards    []Msg_S2C_SpinAward `json:"awards"`     // 奖励列表
}

type Msg_S2C_SpinAward struct {
	Id    int64  `json:"id"`    // 奖励ID
	Desc  string `json:"desc"`  // 奖励描述
	Type  int    `json:"type"`  // 奖励类型 0金币, 1实物
	Count int64  `json:"count"` // 奖励数量 客户端显示要 / 100
	Icon  string `json:"icon"`  // 奖励图标
}

type Msg_S2C_SpinResult struct {
	SpinTimes      int   `json:"spin_times"`      // 剩余次数
	AwardId        int64 `json:"award_id"`        // 中奖奖励ID
	VitaminChanged bool  `json:"vitamin_changed"` // 比赛点是否发生变化
	Vitamin        int64 `json:"vitamin"`         // 最新的比赛点
}

type Msg_S2C_CheckinInfo struct {
	GoldAwards  []int64 `json:"gold_awards"` // 每天的金币奖励，客户端显示要 / 100，下标0-6依次对应1-7天
	*MsgCheckIn         // 其他信息见下方
}

type MsgCheckIn struct {
	CurDay int  `json:"cur_day"` // 当前第几天，0-6依次对应1-7天
	Has    bool `json:"has"`     // 今天是否已签
}

type Msg_S2C_CheckinResult struct {
	CurDay         int   `json:"cur_day"`         // 第几天签到
	GoldAward      int64 `json:"gold_award"`      // 金币奖励， 客户端显示要 / 100
	VitaminChanged bool  `json:"vitamin_changed"` // 比赛点是否发生变化
	Vitamin        int64 `json:"vitamin"`         // 最新的比赛点
}

type Msg_S2C_BattleRank struct {
	MyRank  int64                     `json:"myrank"`  // 我的排名 1-正无穷， 0为未上榜
	MyRound int64                     `json:"myround"` // 我的局数
	List    []*Msg_S2C_BattleRankItem `json:"list"`    // 榜单数据
}

type Msg_S2C_BattleRankItem struct {
	Rank     int64  `json:"rank"`     // 排名 1-10
	Uid      int64  `json:"uid"`      // 玩家ID
	Imgurl   string `json:"imgurl"`   // 头像
	Nickname string `json:"nickname"` // 昵称
	Sex      int    `json:"sex"`      // 性别
	Round    int64  `json:"round"`    // 局数
}
