package static

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"time"

	goRedis "github.com/go-redis/redis"
)

type MapInfo struct {
	Longitude string `json:"longitude"` //! 经度
	Latitude  string `json:"latitude"`  //! 纬度
	Country   string `json:"country"`   //! 国家
	Province  string `json:"province"`  //! 省
	City      string `json:"city"`      //! 市
	Citycode  string `json:"citycode"`  //! 城市编码
	District  string `json:"district"`  //! 区
	Adcode    string `json:"adcode"`    //! 区域码
	Address   string `json:"address"`   //! 地址
}

// ! 玩家的数据结构
type Person struct {
	Uid         int64  `json:"uid"`         //! 用户数字id
	Nickname    string `json:"nickname"`    //! 昵称
	Imgurl      string `json:"imgurl"`      //! 头像
	Guestid     string `json:"guestid"`     //! 游客id
	Openid      string `json:"openid"`      //! 微信openid
	Card        int    `json:"card"`        //! 房卡数量
	FrozenCard  int    `json:"frozencard"`  //! 冻结房卡数量
	Gold        int    `json:"gold"`        //! 金币数量
	GoldBean    int    `json:"gold_bean"`   //! 金豆
	InsureGold  int    `json:"insure_gold"` //! 保险箱金币
	GameId      int    `json:"game_id"`     //! 当前处于哪个game中
	TableId     int    `json:"table_id"`    //! 当前处于哪个room中
	HouseId     int    `json:"house_id"`    //! 当前处于哪个house中
	FloorId     int64  `json:"floor_id"`    //! 当前处于哪个floor中
	Area        string `json:"area"`        //! 区域code
	SiteId      int    `json:"site_id"`     //! 当前处于哪个场次中
	Sex         int    `json:"sex"`         //! 性别 1男 其他女
	Tel         string `json:"tel"`         //! 手机号
	ReName      string `json:"rename"`      //! 真实姓名
	Idcard      string `json:"idcard"`      //! 身份证号
	Token       string `json:"token"`       //! 登录凭证
	CreateTable int    `json:"createtable"` //! 用户创建自己玩的牌桌
	//OtherTables []int  `json:"othertables"` //! 用户替他人创建的牌桌id数组
	Games           string `json:"games"`           //! 用户游戏列表
	IsBlack         int8   `json:"isblack"`         //! 是否进入黑名单
	CreateTime      int64  `json:"create_time"`     //! 创建时间
	DescribeInfo    string `json:"describe_info"`   //! 个性签名
	UserType        int    `json:"user_type"`       // 用户账号类型
	Platform        int    `json:"platform"`        // 平台类型
	LastLoginTime   int64  `json:"last_login_at"`   //最近一次登录时间
	LastOffLineTime int64  `json:"last_offline_at"` //上次离线时间
	// 战绩
	WinCount    int    `json:"win_count"`    // 胜利局数
	LostCount   int    `json:"lost_count"`   // 失败局数
	DrawCount   int    `json:"draw_count"`   // 和局局数
	FleeCount   int    `json:"flee_count"`   // 逃跑局数
	TotalCount  int    `json:"total_count"`  // 总局数
	Engine      int    `json:"engine"`       // 游戏引擎信息
	IsJoin      int64  `json:"-"`            // 是否正在入桌
	DeliveryImg string `json:"delivery_img"` // 用户个人名片图片所在地址
	Ip          string `json:"ip"`
	Online      bool   `json:"online"`
	Address     string `json:"address"`   //! 地址
	Longitude   string `json:"longitude"` //! 经度
	Latitude    string `json:"latitude"`  //! 纬度
	GameIp      string `json:"gameip"`    //! 进入游戏的ip

	AdminGameOn bool `json:"admin_game_on"`

	//贡献值系统
	ContributionScore int64 `json:"contributionScore"` //玩家贡献值

	Diamond int `json:"diamond"` //! 钻石

	IsRobot            int    `json:"isrobot"`        //是否是机器人
	UnionCode          string `json:"union_code"`     //联运id
	RefuseInvite       bool   `json:"refuse_invite"`  //拒绝入圈邀请
	WatchTable         int    `json:"watchtable"`     //观战桌子
	HotVersion         string `json:"hot_version"`    //热更新版本
	CardRecorderDeadAt int64  `json:"crdead_at"`      //记牌器失效时间
	AccountType        int    `json:"account_type"`   //账号状态
	Area2nd            string `json:"area2nd"`        //第二区域码
	Area3rd            string `json:"area3rd"`        //第三区域码
	IdcardAuthPI       string `json:"idcard_auth_pi"` //实名认证官方唯一认证码
	PlayTime           int    `json:"play_time"`      //游戏时长
	IsVip              bool   `json:"is_vip"`         //是否开通会员
}

// 玩家游戏历史
type GameHistory struct {
	KindId int   `json:"kind_id"` // 游戏id
	Times  int   `json:"times"`   // 游戏次数
	LastAt int64 `json:"last_at"` // 最后一次游戏时间
}

// 游戏配置
type ConfigGame struct {
	KindId           int    `json:"kindid"`           // 游戏玩法id
	Name             string `json:"name"`             // 游戏玩法名称
	MinPlayerNum     int    `json:"minplayernum"`     // 游戏开始最少人数
	MaxPlayerNum     int    `json:"maxplayernum"`     // 游戏开始最大人数
	DefaultPlayerNum int    `json:"defaultplayernum"` // 游戏开始默认人数
	MaxRoundNum      int    `json:"maxroundnum"`      // 最大游戏局数
	MinRoundNum      int    `json:"minroundnum"`      // 最小游戏局数
	DefaultRoundNum  int    `json:"defaultroundnum"`  // 默认游戏局数
	MinCardCost      int    `json:"mincardcost"`      // 最低房卡消耗
	DefaultCostType  int    `json:"defaultcosttype"`  // 默认支付方式
	DefaultRestrict  bool   `json:"defaultrestrict"`  // 默认ip限制
	DefaultView      bool   `json:"defaultview"`      // 默认是否允许观看
	GameConfig       string `json:"gameconfig"`       // 游戏玩法特色配置
}

// 包厢配置
type ConfigHouse struct {
	Id        int  `json:"id"`
	CreateMax int  `json:"createMax" gorm:"column:createMax"`   // 创建包厢最大数
	JoinMax   int  `json:"joinMax" gorm:"column:joinMax"`       // 加入包厢最大数
	MemMax    int  `json:"memMax" gorm:"column:memMax"`         // 成员最大数
	AdminMax  int  `json:"adminMax" gorm:"column:adminMax"`     // 管理员最大数
	TableNum  int  `json:"tableNum" gorm:"column:tableNum"`     // 显示桌子数
	IsChecked bool `json:"ischecked" gorm:"column:ischecked"`   // 是否开启加入审核
	IsFrozen  bool `json:"isfrozen" gorm:"column:isfrozen"`     // 是否冻结
	IsMemHide bool `json:"ismemhide" gorm:"column:ismemhide"`   // 是否成员隐藏
	CardCost  int  `json:"createCard" gorm:"column:createCard"` // 创建所需房卡数
}

type HouseFloor struct {
	Id           int64  `json:"id"`     // id
	DHId         int64  `json:"dhid"`   // 包厢id
	Rule         string `json:"rule"`   // 楼层规则
	IsMix        bool   `json:"is_mix"` // 是否混排
	Name         string `json:"name"`   // 用户自定义名称
	IsVip        bool   `json:"is_vip"` // 是否vip
	AiSuperNum   int    `json:"ai_super_num"`
	IsHide       bool   `json:"ishide"`
	IsCapSetVip  bool   `json:"is_cap_set_vip"`  // 队长是否可以设置vip
	IsDefJoinVip bool   `json:"is_def_join_vip"` // 新入圈的玩家是否自动加入VIP楼层
	MinTable     int    `json:"min_table"`
	MaxTable     int    `json:"max_table"`
	FloorVitaminOptions
}

func (hf *HouseFloor) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, hf)
}

func (hf *HouseFloor) MarshalBinary() (data []byte, err error) {
	return json.Marshal(hf)
}

type HouseMember struct {
	Id           int64  `json:"id"`            //! id
	DHId         int64  `json:"dhid"`          //! 包厢id
	FId          int64  `json:"fid"`           //! 历史楼层id
	UId          int64  `json:"uid"`           //! 玩家id
	UVitamin     int64  `json:"uvitamin"`      //! 疲劳值
	URole        int    `json:"urole"`         //! 玩家角色
	URemark      string `json:"uremark"`       //! 玩家备注
	ApplyTime    int64  `json:"apply_time"`    //! 申请时间
	AgreeTime    int64  `json:"agree_time"`    //! 进入时间
	BwTimes      int    `json:"bw_times"`      //! 大赢家次数
	PlayTimes    int    `json:"play_times"`    //! 对局次数
	Forbid       int    `json:"forbid"`        //! 0正常娱乐1禁止娱乐
	Partner      int64  `json:"partner"`       //! 队长 0否 非0为uid
	IsLimitGame  bool   `json:"is_limit_game"` //! 是否被限制游戏
	IsOnline     bool   `json:"is_online"`     //! 是否在线
	HId          int    `json:"hid"`           //! 包厢id
	NickName     string `json:"nickname"`      //! 昵称
	ImgUrl       string `json:"imgurl"`        //! 头像
	Sex          int    `json:"sex"`           //! 性别 1男 其他女
	Ref          int64  `json:"ref"`           //! 参考来源Dhid
	Superior     int64  `json:"superior"`      //! 队长的上级
	Agent        int64  `json:"agent"`         //! 隶属于那个区域代理/副盟主/副会长名下
	VitaminAdmin bool   `json:"vitamin_admin"` //! 是否是比赛分管理员
	VicePartner  bool   `json:"vice_partner"`  //! 是否为副队长
	PRemark      string `json:"p_remark"`      //! 队长设置的备注
	NoFloors     string `json:"no_floors"`     //! 不能加入的楼层
}

func (hm *HouseMember) IsPartner() bool {
	return hm.Partner == 1
}

func (hm *HouseMember) RedisWriteKey() string {
	return fmt.Sprintf("hmember_w_%d_%d", hm.DHId, hm.UId)
}

func (hm *HouseMember) RedisReadKey() string {
	return fmt.Sprintf("hmember_r_%d_%d", hm.DHId, hm.UId)
}

func (hm *HouseMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(hm)
}

func (hm *HouseMember) UnmarshalBinary(b []byte) error {
	return json.Unmarshal(b, hm)
}

const (
	SleepDuring = 30 * time.Microsecond
)

// RLock 疲劳值读锁
func (hm *HouseMember) RLock(cli *goRedis.Client) {
	for {
		if cli.Exists(hm.RedisWriteKey()).Val() == 1 {
			time.Sleep(SleepDuring)
		}
		cli.Incr(hm.RedisReadKey())
		if cli.Exists(hm.RedisWriteKey()).Val() == 1 {
			fmt.Println("write key happen when write")
			cli.Del(hm.RedisReadKey())
		} else {
			break
		}
	}
}

// UnRlock
func (hm *HouseMember) RUnlock(cli *goRedis.Client) error {
	for {
		if cli.Exists(hm.RedisWriteKey()).Val() != 1 { //读取期间异常导致写锁被打开
			if cli.Decr(hm.RedisReadKey()).Val() == 0 { //为零，删除key
				cli.Del(hm.RedisReadKey())
			}
			return errors.New("redis lock error")
		} else {
			return errors.New("redis UnRlock error, write lock exists")
		}
	}
}

func (hm *HouseMember) Lock(cli *goRedis.Client) {
	reload := false
	for {
		if cli.SetNX(hm.RedisWriteKey(), 1, 30*time.Second).Val() {
			break
		}
		time.Sleep(SleepDuring)
		reload = true
	}
	if reload {
		buf := cli.HGet(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, hm.DHId), fmt.Sprintf("%d", hm.UId)).Val()
		if len(buf) != 0 {
			json.Unmarshal([]byte(buf), hm)
		}
	}
}

func (hm *HouseMember) Unlock(cli *goRedis.Client) {
	cli.Del(hm.RedisWriteKey())
}

type FloorTable struct {
	NTId int   `json:"ntid"` //! ntid
	TId  int   `json:"tid"`  //! tid
	DHId int64 `json:"dhid"` //! 包厢id
	FId  int64 `json:"fid"`  //! 楼层id
}

type FloorTableId struct {
	TId int `json:"tid"` //! tid
}

// 楼层活动
type HouseActivity struct {
	Id          int64        `json:"id"`           //! id
	DHId        int64        `json:"hid"`          //! 包厢id
	FId         string       `json:"fid"`          //! 楼层id
	FIdIndex    string       `json:"fidindex"`     //! 楼层id
	Kind        int          `json:"kind"`         //! 类型
	Name        string       `json:"name"`         //! 名称
	HideInfo    bool         `json:"hideinfo"`     //! 名称
	BegTime     int64        `json:"begtime"`      // 开始时间
	Status      int          `json:"status"`       //! 是否活跃
	EndTime     int64        `json:"endtime"`      // 结束时间
	Type        int64        `json:"type"`         //是否是幸运星抽奖
	TicketCount int64        `json:"ticket_count"` //多少次抽奖一次
	Reowords    []RewordInfo `json:"reowords"`
}

// ! 游戏总结算记录
type GameRecordTotal struct {
	Id        int64  `json:"id"`        //! id
	GameNum   string `json:"gamenum"`   //! 游戏ID,唯一标识
	RoomNum   int    `json:"roomnum"`   //! 游戏房间ID
	PlayCount int    `json:"playcount"` //! 游戏局数
	Round     int    `json:"round"`     //! 游戏总局数
	ServerId  int    `json:"serverid"`  //! 游戏服务ID
	SeatId    int    `json:"seatid"`    //! 玩家座位ID
	Uid       int64  `json:"uid"`       //! 玩家用户ID
	UName     string `json:"uname"`     //! 玩家名称
	ScoreKind int    `json:"scorekind"` //! 游戏结束类型
	WinScore  int    `json:"winscore"`  //! 玩家积分
	Ip        string `json:"ip"`        //! 玩家IP地址
	WriteDate string `json:"writedate"` //! 写入时间
	HId       int64  `json:"hid"`       //! 包厢ID
	IsHeart   int    `json:"isheart"`   //! 该战绩是否点赞
	FId       int    `json:"fid"`       //! 包厢楼层
}

// ! 游戏单局总结算记录
type GameRecord struct {
	Id        int64  `json:"id"`        //! id
	GameNum   string `json:"gamenum"`   //! 游戏ID,唯一标识
	RoomNum   int    `json:"roomnum"`   //! 游戏房间ID
	PlayNum   int    `json:"playnum"`   //! 游戏局数
	ServerId  int    `json:"serverid"`  //! 游戏服务ID
	SeatId    int    `json:"seatid"`    //! 玩家座位ID
	UId       int64  `json:"uid"`       //! 玩家用户ID
	UName     string `json:"uname"`     //! 玩家名称
	ScoreKind int    `json:"scorekind"` //! 游戏结束类型
	WinScore  int    `json:"winscore"`  //! 玩家积分
	Ip        string `json:"ip"`        //! 玩家IP地址
	ReplayId  int64  `json:"replayid"`  //! 该局游戏回放ID
	UUrl      string `json:"uurl"`      //! 玩家头像
	UGenber   int    `json:"ugender"`   //! 玩家性别
	BeginDate int64  `json:"begindate"` //! 游戏开始时间
	WriteDate int64  `json:"writedate"` //! 写入时间(游戏结束时间)
}

// ! 游戏单局回放记录
type GameRecordReplay struct {
	Id        int64  `json:"id"`        //! id
	GameNum   string `json:"gamenum"`   //! 游戏ID,唯一标识
	RoomNum   int    `json:"roomnum"`   //! 游戏房间ID
	PlayNum   int    `json:"playnum"`   //! 游戏局数
	ServerId  int    `json:"serverid"`  //! 游戏服务ID
	HandCard  string `json:"handcard"`  //! 玩家手牌
	OutCard   string `json:"outcard"`   //! 玩家出牌记录
	WriteData string `json:"writedate"` //! 写入时间
	KindID    int    `json:"kindid"`    //! 游戏kindID
	CardsNum  int    `json:"cardsnum"`  //！发完牌后剩余牌数
}

// ！ 游戏每日统计数据
type GameDayRecord struct {
	Id         int    `json:"id"`          //! id
	DHId       int64  `json:"dhid"`        //! 包厢id
	DFId       int64  `json:"dfid"`        //! 包厢楼层id
	UId        int64  `json:"uid"`         //! 包厢玩家id
	PlayDate   string `json:"playdate"`    //! 当天游戏时间
	PlayTimes  int    `json:"play_times"`  //! 当天总局数
	BwTimes    int    `json:"bw_times"`    //! 当天大赢家次数
	TotalScore int    `json:"total_score"` //! 当天累计积分
}

// ! 游戏扣卡统计数据
type RecordGameCostMini struct {
	HId      int64 `json:"hid"`      //! 包厢id(数据库自增id,客户端不作展示)
	PlayTime int   `json:"playtime"` //! 对局次数
	KaCost   int   `json:"kacost"`   //! 消耗
	Date     int64 `json:"date"`
}
type GameCostRecord struct {
	Id         int64  `json:"id"`         //! id
	UId        int64  `json:"uid"`        //! 玩家id
	TId        int    `json:"tid"`        //! 牌桌id
	HId        int64  `json:"hid"`        //! 包厢id(数据库自增id)
	FId        int64  `json:"fid"`        //! 楼层id
	NTId       int    `json:"ntid"`       //! 牌桌索引
	BefKa      int    `json:"befka"`      //! 消耗前
	AftKa      int    `json:"aftka"`      //! 消耗后
	KaCost     int    `json:"kacost"`     //! 消耗
	Gamenum    string `json:"gamenum"`    //! 牌局编号
	GameConfig string `json:"gameconfig"` //! 游戏配置
	Date       int64  `json:"date"`       //! 时间
}

// 玩家好友房对局历史
type GameRecordHistory struct {
	GameNum   string                     `json:"gamenum"`   //! 游戏ID,唯一标识
	RoomNum   int                        `json:"roomnum"`   //! 游戏房间ID
	KindId    int                        `json:"kindid"`    //! 游戏玩法标识
	PlayedAt  int64                      `json:"playedat"`  //! 对局时间(时间戳)
	HId       int64                      `json:"hid"`       //! 包厢Id(数据库自增id)
	FId       int                        `json:"fid"`       //! 包厢楼层Id
	IsHeart   int                        `json:"isheart"`   //! 是否点赞
	PlayCount int                        `json:"playcount"` //! 游戏局数
	Round     int                        `json:"round"`     //! 房间总局数
	Player    []*GameRecordHistoryPlayer `json:"player"`    //! 玩家数组
}

type GameRecordHistoryWrapper struct {
	Item []GameRecordHistory
	By   func(p, q *GameRecordHistory) bool
}

func (pw GameRecordHistoryWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw GameRecordHistoryWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw GameRecordHistoryWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(&pw.Item[i], &pw.Item[j])
}

// 好友房对局历史记录
type GameRecordHistoryPlayer struct {
	Uid      int64   `json:"uid"`      // 玩家id
	Nickname string  `json:"nickname"` // 玩家昵称
	Score    float64 `json:"score"`    // 玩家得分
}

// 游戏战绩
type GameRecordDetal struct {
	GameNum    string                  `json:"gamenum"`     //! 游戏ID,唯一标识
	GameIndex  int                     `json:"gameindex"`   //! 游戏索引序号
	RoomNum    int                     `json:"roomnum"`     //! 游戏房间ID
	KindId     int                     `json:"kindid"`      //! 游戏玩法标识
	PlayedAt   int64                   `json:"playedat"`    //! 对局时间(时间戳)
	HId        int64                   `json:"hid"`         //! 包厢Id(数据库自增id)
	FId        int                     `json:"fid"`         //! 包厢楼层Id
	DFId       int                     `json:"dfid"`        //! 包厢楼层索引
	IsHeart    int                     `json:"isheart"`     //! 是否点赞
	PlayRound  int                     `json:"playround"`   //! 游戏局数
	TotalRound int                     `json:"totalround"`  //! 房间总局数
	WanFa      string                  `json:"wf"`          //! 游戏玩法
	FinishType int                     `json:"finishtype"`  //! 游戏结束类型
	PartnerIds []int64                 `json:"partnerid"`   //! 队长id
	PlayerTags []int64                 `json:"player_tags"` //! 0：不显示  1：我  2：队员
	Player     []GameRecordDetalPlayer `json:"player"`      //! 玩家数组
	Point      float64                 `json:"point"`       //! 疲劳值房费扣除
}

type GameRecordDetalWrapper struct {
	Item []GameRecordDetal
	By   func(p, q *GameRecordDetal) bool
}

func (pw GameRecordDetalWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}
func (pw GameRecordDetalWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}
func (pw GameRecordDetalWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(&pw.Item[i], &pw.Item[j])
}

// 玩家对局信息
type GameRecordDetalPlayer struct {
	Uid      int64   `json:"uid"`      // 玩家id
	NickName string  `json:"nickname"` // 玩家昵称
	HeadUrl  string  `json:"headurl"`  // 玩家昵称
	Sex      int     `json:"sex"`      // 玩家性别
	Score    float64 `json:"score"`    // 玩家得分
	IsExit   bool    `json:"isexit"`   // 是否退圈
}

// 任务列表
type TaskList []*Task

// 任务
type Task struct {
	Id   int64 `json:"id"`   // id
	TcId int   `json:"tcid"` // 任务配置表id
	Uid  int64 `json:"uid"`  // 用户id
	Num  int   `json:"num"`  // 任务当前计数
	Step int   `json:"step"` // 当前阶段
	Sta  int   `json:"Sta"`  // 任务状态 进行中0、完成未领取1、已领取2
	Time int64 `json:"time"` // 更新时间
}

// ! 奖励
type AwardItem struct {
	Type int `json:"type"`
	Num  int `json:"num"`
}

type HTableRecordPlayers struct {
	HId   int     `json:"hid"`
	FId   int64   `json:"fid"`
	TId   int     `json:"tid"`
	KId   int     `json:"kid"`
	Users []int64 `json:"users"`
}
type LastGameInfo struct {
	HId             int                   `json:"hid"`
	FId             int64                 `json:"fid"`
	TId             int                   `json:"tid"`
	KId             int                   `json:"kid"`
	Users           []int64               `json:"users"`
	PlayersInfo     []LastGamePlayersInfo `json:"playersinfo"`
	GameendInfo     string                `json:"gameendinfo"`     //小结算
	ResultTotalInfo string                `json:"resulttotalinfo"` //大结算
	CardsDataInfo   string                `json:"cardsdatainfo"`   //桌面数据
}
type LastGamePlayersInfo struct {
	Uid           int64        `json:"uid"`           //!
	Name          string       `json:"name"`          //! 名字
	ImgUrl        string       `json:"imgurl"`        //! 头像
	Sex           int          `json:"sex"`           //! 性别
	Ip            string       `json:"ip"`            //IP地址
	FaceID        int          `json:"faceid"`        //玩家默认头像
	FaceUrl       string       `json:"faceurl"`       //玩家头像
	UserRight     int          `json:"userright"`     //用户等级
	Loveliness    int          `json:"loveliness"`    //用户魅力
	UserScoreInfo TagUserScore `json:"userscoreinfo"` //积分信息
}
type HouseMemberMap map[int64]*HouseMember

func (hmm *HouseMemberMap) SuperiorsByJuniors(juniors ...int64) []int64 {
	in := func(jid int64) bool {
		for i := 0; i < len(juniors); i++ {
			if jid == juniors[i] {
				return true
			}
		}
		return false
	}
	res := make([]int64, 0)
	for id, mem := range *hmm {
		if in(id) && mem.Superior > 0 {
			res = append(res, mem.Superior)
		}
	}
	return res
}

func (hmm *HouseMemberMap) JuniorsBySuperiors(superiors ...int64) []int64 {
	in := func(sid int64) bool {
		for i := 0; i < len(superiors); i++ {
			if sid == superiors[i] {
				return true
			}
		}
		return false
	}
	res := make([]int64, 0)
	for id, mem := range *hmm {
		if mem.IsPartner() && in(mem.Superior) {
			res = append(res, id)
		}
	}
	return res
}

func (hmm *HouseMemberMap) MemAndJuniorsBySuperiors(superiors ...int64) []int64 {
	in := func(sid int64) bool {
		for i := 0; i < len(superiors); i++ {
			if sid == superiors[i] {
				return true
			}
		}
		return false
	}
	res := make([]int64, 0)
	for id, mem := range *hmm {
		if mem.IsPartner() {
			if in(mem.Superior) {
				res = append(res, id)
			}
		} else if mem.Partner > 0 {
			if in(mem.Partner) {
				res = append(res, id)
			}
		}
	}
	return res
}

// ! 机器人的数据结构
type Robot struct {
	Mid          int64  `json:"mid"`           //! 机器人数字id
	Nickname     string `json:"nickname"`      //! 昵称
	Imgurl       string `json:"imgurl"`        //! 头像
	Guestid      string `json:"guestid"`       //! 游客id
	Openid       string `json:"openid"`        //! 微信openid
	Card         int    `json:"card"`          //! 房卡数量
	FrozenCard   int    `json:"frozencard"`    //! 冻结房卡数量
	Gold         int    `json:"gold"`          //! 金币数量
	GoldBean     int    `json:"gold_bean"`     //! 金豆
	InsureGold   int    `json:"insure_gold"`   //! 保险箱金币
	GameId       int    `json:"game_id"`       //! 当前处于哪个game中
	TableId      int    `json:"table_id"`      //! 当前处于哪个room中
	HouseId      int    `json:"house_id"`      //! 当前处于哪个house中
	FloorId      int64  `json:"floor_id"`      //! 当前处于哪个floor中
	Area         string `json:"area"`          //! 区域code
	Appid        string `json:"appid"`         //! Appid
	SiteId       int    `json:"site_id"`       //! 当前处于哪个场次中
	Sex          int    `json:"sex"`           //! 性别
	Tel          string `json:"tel"`           //! 手机号
	ReName       string `json:"rename"`        //! 真实姓名
	Idcard       string `json:"idcard"`        //! 身份证号
	Token        string `json:"token"`         //! 登录凭证
	CreateTable  int    `json:"createtable"`   //! 用户创建自己玩的牌桌
	Games        string `json:"games"`         //! 用户游戏列表
	IsBlack      int8   `json:"isblack"`       //! 是否进入黑名单
	CreateTime   int64  `json:"create_time"`   //! 创建时间
	DescribeInfo string `json:"describe_info"` //! 个性签名
	UserType     int    `json:"user_type"`     //! 用户账号类型
	Platform     int    `json:"platform"`      //! 平台类型
	Diamond      int    `json:"diamond"`       //! 钻石

	// 战绩
	WinCount          int    `json:"win_count"`         //! 胜利局数
	LostCount         int    `json:"lost_count"`        //! 失败局数
	DrawCount         int    `json:"draw_count"`        //! 和局局数
	FleeCount         int    `json:"flee_count"`        //! 逃跑局数
	TotalCount        int    `json:"total_count"`       //! 总局数
	UnionCode         string `json:"union_code"`        //! 联运id
	ContributionScore int64  `json:"contributionScore"` //! 玩家贡献值
	MachineCode       string `json:"machine_code"`      //! 机器码
	IsRobot           int    `json:"is_robot"`          //! 是否是机器人
	IsWorking         int    `json:"is_working"`        //! 机器人是否在工作(0否1是)
	WorkingAt         int64  `json:"working_at"`        //! 机器人开始工作的时间
	RestAt            int64  `json:"rest_at"`           //! 机器人开始休息的时间
	KindID            int    `json:"kind_id"`           //! 机器人所在的游戏id
}

func (u *Robot) Convert2Person() *Person {
	p := new(Person)
	p.Uid = u.Mid
	p.UserType = u.UserType
	p.Nickname = u.Nickname
	p.Imgurl = u.Imgurl
	p.Guestid = u.Guestid
	p.Openid = u.Openid
	p.Card = u.Card
	p.Gold = u.Gold
	p.InsureGold = u.InsureGold
	p.GoldBean = u.GoldBean
	p.FrozenCard = u.FrozenCard
	p.DescribeInfo = u.DescribeInfo
	p.Sex = u.Sex
	p.Tel = u.Tel
	p.ReName = u.ReName
	p.Idcard = u.Idcard
	p.Token = u.Token
	p.Games = u.Games
	p.IsBlack = u.IsBlack
	p.CreateTime = u.CreateTime
	p.WinCount = u.WinCount
	p.LostCount = u.LostCount
	p.DrawCount = u.DrawCount
	p.FleeCount = u.FleeCount
	p.TotalCount = u.TotalCount
	p.UnionCode = u.UnionCode
	p.ContributionScore = u.ContributionScore
	p.Diamond = u.Diamond
	p.IsRobot = 1
	return p
}
