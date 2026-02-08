//! 服务器之间的消息
package static

import (
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/consts"
)

type Msg_Sign struct {
	Time   int64 `json:"time"`
	Encode int   `json:"encode"`
}

type Msg_Header struct {
	Header  string   `json:"head"`
	Data    string   `json:"data"`
	ErrCode int16    `json:"errcode"`
	ErrMsg  string   `json:"errmsg"`
	Sign    Msg_Sign `json:"msgsign"`
}

//! 消息结构
type Msg_MsgBase struct {
	Uid     int64    `json:"uid"`
	IP      string   `json:"ip"`
	Head    string   `json:"head"`
	ErrCode int16    `json:"errcode"`
	Sign    Msg_Sign `json:"msgsign"`
	Data    string   `json:"data"`
}

type Rpc_Args struct {
	MsgData []byte
}

type Msg_Null struct {
}

type Msg_Userid struct {
	Uid int64 `json:"uid"`
}

type Msg_Ping struct {
	GodTime int64 `json:"godtime"`
}

//! 保险箱金币存取
type Msg_SaveInsureGold struct {
	Gold int `json:"gold"`
}

//! 更新用户信息
type Msg_UpdateUserInfo struct {
	Sex      int    `json:"sex"`
	Nickname string `json:"nickname"`
}

//! 更新用户信息
type Msg_UpdateDescribeInfo struct {
	Describe string `json:"describe"`
}

//! 用户广播消息
type Msg_UserAreaBroadcast struct {
	Content string `json:"content"`
}

//! 领取双倍低保
type Msg_GetAllowancesDouble struct {
	Current int `json:"current"`
}

//! 加减房卡
type Msg_ChangeCard struct {
	Uid int64 `json:"uid"`
	Num int   `json:"num"`
}

//! 玩家
type Msg_Uid struct {
	Uid   int64  `json:"uid"`
	Token string `json:"token"`
	Fid   int64  `json:"fid"`
	Hid   int    `json:"hid"`
}

//！玩家牌桌信息
type Msg_UserTableId struct {
	Uid     int64 `json:"uid"`
	TableId int   `json:"table_id"`
	Vitamin int64 `json:"vitamin"`
}

//! 玩家集合
type Msg_Uids struct {
	Uids []int64 `json:"uids"`
}

//! 房号
type Msg_TableId struct {
	Id int `json:"roomid"`
}

//! 玩家微信id
type Msg_Openid struct {
	Openid string `json:"openid"`
}

//令牌
type Msg_Token struct {
	Token string `json:"token"`
}

//！game服务器桌子数量变化
type Msg_TableNumChange struct {
	Id       int `json:"id"`
	TableNum int `json:"table_num"`
}

//! gameserver开关
type Msg_GameServer struct {
	Id        int                       `json:"id"`
	InIp      string                    `json:"inip"`
	ExIp      string                    `json:"exip"`
	SafeIp    string                    `json:"safeip"`
	GameTypes map[int][]*ServerGameType `json:"gametypes"` // 支持的子游戏类型及房间场次类型
	Status    int                       `json:"status"`    // 服务器状态
}

type ServerGameType struct {
	GameType     int `json:"game_type"`      // 金币场 or 好友场 or 比赛场
	KindId       int `json:"kind_id"`        // 游戏id
	SiteType     int `json:"site_type"`      // 场次类型
	TableNum     int `json:"table_num"`      // 桌子数量
	MaxPeopleNum int `json:"max_people_num"` // 最大限制进入人数
}

//! hallserver开关
type Msg_HallServer struct {
	InIp string `json:"inip"`
	ExIp string `json:"exip"`
}

//################

//! 删除房间
type Msg_DelTable struct {
	Id      int     `json:"id"`
	TableId int     `json:"tableid"`
	Uid     []int64 `json:"uid"`
	Host    int64   `json:"host"`
}

//! 开始房间
type Msg_BeginTable struct {
	Id      int `json:"id"`
	TableId int `json:"tableid"`
}

//! 删除房间
type Msg_TableIn_Result struct {
	Uid     int64  `json:"uid"`
	Result  bool   `json:"result"`
	Ip      string `json:"ip"`
	TableId int    `json:"tableid"`
}

//! 离开房间
type GH_TableExit_Ntf struct {
	Uid     int64 `json:"uid"`
	TableId int   `json:"tableid"`
	GameId  int   `json:"gameid"` //! 游戏服id
	KindId  int   `json:"kindid"` //! 游戏类型
	Hid     int   `json:"hid"`
	Fid     int64 `json:"fid"`
}

//! 房间大结算数据
type GH_TableRes_Ntf struct {
	UId      int64 `json:"uid"`
	TId      int   `json:"tid"`
	HId      int   `json:"hid"`
	GameId   int   `json:"gameid"`   //! 游戏服id
	KindId   int   `json:"kindid"`   //! 游戏类型
	WinScore int   `json:"winscore"` //! 赢取积分
	IsBigWin int   `json:"isbigwin"` //! 是否大赢家
}

type GH_TableDel_Ntf struct {
	TableId int   `json:"tableid"`
	Hid     int   `json:"hid"`
	Fid     int64 `json:"fid"`
}

//! 用户落座
type Msg_UserSeat struct {
	Uid     int64 `json:"uid"`
	TableId int   `json:"tableid"`
	Seat    int   `json:"seat"`
}

// 包厢相关逻辑

// 创建牌桌
type HG_HTableCreate_Req struct {
	Table    Table  `json:"table"`
	AutoUid  int64  `json:"autouid"`  // 自动加入玩家id
	AutoSeat int    `json:"autoseat"` // 自动加入座位
	Ip       string `json:"ip"`       // 游戏服ip
	Payer    int64  `json:"payer"`
}
type GH_HTableCreate_Ack struct {
	TId    int    `json:"tid"`     //! 游戏服id
	NTid   int    `json:"ntid"`    //! 游戏服id
	GameId int    `json:"gameid"`  //! 游戏服id
	KindId int    `json:"kindid"`  //! 游戏类型
	Ip     string `json:"ip"`      //! 游戏服ip
	Uid    int64  `json:"uid"`     //! 玩家id
	Seat   int    `json:"seat"`    //! 座位
	JoinAt int64  `json:"join_at"` // 加入时间
	Fid    int64  `json:"fid"`
	Hid    int    `json:"hid"`
	Payer  int64  `json:"payer"`
}

// 加入牌桌
type HG_HTableIn_Req struct {
	TId    int    `json:"tid"`    //! 游戏服id
	NTid   int    `json:"ntid"`   //! 游戏服id
	GameId int    `json:"gameid"` //! 游戏服id
	KindId int    `json:"kindid"` //! 游戏类型
	Ip     string `json:"ip"`     //! 游戏服ip
	Uid    int64  `json:"uid"`    //! 玩家id
	Seat   int    `json:"seat"`   //! 座位
	Payer  int64  `json:"payer"`  //! 谁支付
}
type GH_HTableIn_Ack struct {
	TId    int    `json:"tid"`     //! 游戏服id
	NTid   int    `json:"ntid"`    //! 游戏服id
	GameId int    `json:"gameid"`  //! 游戏服id
	KindId int    `json:"kindid"`  //! 游戏类型
	Ip     string `json:"ip"`      //! 游戏服ip
	Uid    int64  `json:"uid"`     //! 玩家id
	Seat   int    `json:"seat"`    //! 座位
	JoinAt int64  `json:"join_at"` // 加入时间
	Fid    int64  `json:"fid"`
	Hid    int    `json:"hid"`
	Payer  int64  `json:"payer"`
}
type GH_HTableIn_Ntf struct {
	Uid     int64  `json:"uid"`
	Result  bool   `json:"result"`
	Ip      string `json:"ip"`
	TableId int    `json:"tableid"`
}

//######################
//! 设置黑名单
type Msg_SetBlack struct {
	Uid     int64 `json:"uid"`
	IsBlack int8  `json:"is_black"`
}

type Msg_GetUserInfo struct {
	Uid int64 `json:"uid"`
}

type Msg_AreaUpdate struct {
	Area []string `json:"area"`
}

type Msg_CheckUserConnection struct {
	Uid  int64 `json:"uid"`
	Seat int   `json:"seat"`
}

type Msg_TimeEventMsg struct {
	Uid int64  `json:"uid"`
	Id  uint16 `json:"id"`
}

// 钉钉机器人推送
type Msg_DingTalk_Text struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		IsAtAll   bool     `json:"isAtAll"`
	} `json:"at"`
}

// 钉钉机器人推送
type Msg_DingTalk_Link struct {
	MsgType string `json:"msgtype"`
	Link    struct {
		Text       string `json:"text"`
		Title      string `json:"title"`
		PicUrl     string `json:"picUrl"`
		MessageUrl string `json:"messageUrl"`
	} `json:"text"`
}
type Msg_SiteIn_Result struct {
	Uid    int64  `json:"uid"`
	Result bool   `json:"result"`
	Ip     string `json:"ip"`
	SiteId int    `json:"siteid"`
	GameId int    `json:"gameid"`
}

type Msg_HosueIDChange struct {
	OldHId int `json:"oldhId"`
	NewHId int `json:"newhid"`
}

//! 任务
type Msg_Task_Complete struct {
	Kind     int   `json:"kind"`     //! 任务id
	KindId   int   `json:"kindid"`   //! 游戏kindid
	SitType  int   `json:"sittype"`  //! 游戏场次
	Num      int   `json:"num"`      //! 任务完成数量
	Uid      int64 `json:"uid"`      //! 用户名
	CardType int   `json:"cardtype"` //! 牌型
}

type MsgDataID struct {
	Id int `json:"id"`
}

type Msg_G2P_UpdateUser struct {
	CirclerId  string                 `json:"circler_id"`
	GoUser     map[string]interface{} `json:"go_user"`
	PhpCircler map[string]interface{} `json:"php_circler"`
}

type Msg_HouseMenberApplyNTF struct {
	HID       int    `json:"hid"`
	Uid       int64  `json:"uid"`
	UUrl      string `json:"uurl"`
	NickName  string `json:"nick_name"`
	ApplyTime int64  `json:"apply_time"`
	IsOnline  bool   `json:"is_online"`
	ApplyType int    `json:"apply_type"`
}

type MsgLeagueCardAdd struct {
	LeagueID int64 `json:"league_id"`
	Uid      int64 `json:"uid"`
	NotPool  bool  `json:"not_pool"`
	UpdCount int64 `json:"upd_count"`
}

type MsgHouseTableInvite struct {
	TId     int `json:"tid"`
	Inviter struct {
		Uid      int64  `json:"uid"`
		Imgurl   string `json:"imgurl"`
		Nickname string `json:"nickname"`
		Gender   int    `json:"gender"`
	} `json:"inviter"` // 邀请人信息
	Invitee int64 `json:"invitee"` // 被邀请人id
}

type MsgHTableInvitrResp struct {
	TId     int   `json:"tid"`     // 桌子id
	Hid     int   `json:"hid"`     // 包厢数据库dhid
	Inviter int64 `json:"inviter"` // 邀请人uid
	Agree   bool  `json:"agree"`   // 同意与否
	Notips  bool  `json:"notips"`  // 是否今日不再提示
}

type MsgHouseFloorInfo struct {
	TMemItems []Msg_HouseMemberItem `json:"tmemitems"`
}
type MsgRedisSub struct {
	Head string `json:"head"`
	Data string `json:"data"`
}

type MsgHouseMergeRequest struct {
	HId      int `json:"hid"`
	MergeHId int `json:"thid"`
}

type MsgSingleHousePartnerAutoPay struct {
	HId int64 `json:"hid"`
}

type UserReadyState struct {
	Ready bool  `json:"ready"`
	Uid   int64 `json:"uid"`
	Hid   int64 `json:"hid"`
	Fid   int64 `json:"fid"`
	Tid   int   `json:"tid"`
}

//! 玩家
type MsgTableExit struct {
	Tid int    `json:"tid"`
	Uid int64  `json:"uid"`
	Fid int64  `json:"fid"`
	Hid int    `json:"hid"`
	Msg string `json:"msg"`
}

// 签到
type Msg_Checkin struct {
	Type    int  `json:"type"`    // 签到类型：1普通签到(默认类型) 2签到并且分享
	Checkin bool `json:"checkin"` // 是否签到
}

// 签到奖励结果
type Msg_CheckinAward struct {
	Checkin   bool   `json:"checkin"`    // 是否签到
	Double    int    `json:"double"`     // 签到奖励倍数
	AwardName string `json:"award_name"` // 奖品名称
}

// 机器人信息
type Msg_Robots_Task struct {
	KindId      int    `json:"id"`
	TableId     int    `json:"tableId"`     //! 桌子id
	TableNTId   int    `json:"tableNTId"`   //! 桌子NTid
	ServerId    int    `json:"wuhan"`       //! 游戏服id
	Ip          string `json:"ip"`          //! 游戏服ip
	Site        int    `json:"site"`        //! site_type,评估玩家如何入桌
	Seat        int    `json:"seat"`        //! 分配桌位，默认设置为-1
	Time        int64  `json:"time"`        //! 游戏时间
	Task        int    `json:"task"`        //! 任务类型，生成机器人进入游戏（0），退出游戏（1）
	MachineCode string `json:"machinecode"` //! 机器人生成密钥
	Token       string `json:"token"`       //! token
	Uid         int64  `json:"uid"`
	Hall        string `json:"hall"` //大厅ip
}

type Msg_Update_Table_user_score struct {
	Uid     int64 `json:"uid"`
	TableId int   `json:"tableid"`
}

type MsgUpdateUserGold struct {
	Type     int8  `json:"type"`
	Offset   int   `json:"offset"`
	Uid      int64 `json:"uid"`      // 玩家id
	Gold     int   `json:"gold"`     // 最新金币信息
	Bankrupt bool  `json:"bankrupt"` // 是否还是破产状态
}

type MsgAllowanceGift struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data *AllowanceGift `json:"data"`
}

// 拉取低保礼包数据
type AllowanceGift struct {
	// Id              int    `json:"id"`                // 商品id
	// ImgUrl          string `json:"img_url"`           // 图片路径
	// ExchangeGoldNum int    `json:"exchange_gold_num"` // 金币数量
	// NeedDiamondNum  int    `json:"need_diamond_num"`  // 钻石数量
	// BaseCount       int64  `json:"base_count"`        // 基础已购买玩家数量
	PurchasedCount int64 `json:"purchased_count"` // 已购买玩家数量
	UserGold       int   `json:"user_gold"`       // 玩家金币数量
	UserDiamond    int   `json:"user_diamond"`    // 用户钻石数量
}

func (a *AllowanceGift) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *AllowanceGift) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

type MsgFakeTableIn struct {
	Uid      int64 `json:"uid"`
	Bankrupt bool  `json:"bankrupt"` // 是否破产/金币不足
}

type UserStandUp struct {
	Uid int64                  // 站起玩家
	By  int64                  // 被谁站起
	Typ consts.GameStandUpType // 站起类型
}

type Msg_UpdateGold struct {
	Uid      int64 `json:"uid"`
	CostType int8  `json:"costtype"`
	Offset   int   `json:"offset"`
}

type MsgGmBankruptcyGift struct {
	Uid      int64                `json:"uid"`
	CostType int8                 `json:"cost_type"`
	Wts      []MsgGmGotGiftDetail `json:"wts"`
}

type MsgGmGotGiftDetail struct {
	Wt     int8 `json:"wt"`
	Offset int  `json:"offset"`
}

type MsgHotVersion struct {
	Version string `json:"version"`
}

type G2HTableStatsChangeNtf struct {
	Head string `json:"head"`
	Hid  int    `json:"hid"`
	Data string `json:"data"`
}
