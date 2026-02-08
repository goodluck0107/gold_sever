package eve

import "github.com/open-source/game/chess.git/pkg/static"

const (
	ZiYouPiao = -1 //自由漂
)

const (
	HuType_ZiMo    = iota //自摸
	HuType_FangPao        //放炮
	HuType_Null           //不胡
)

const (
	Ts_YinDing_Score       = 6
	Ts_JinDing_Score       = 8
	Ts_ZuanShi_Score       = 10
	Ts_YangGuangDing_Score = 12
)

const (
	Calculate_Score_Zhengshu  = 1
	Calculate_Score_ChuanTong = 2
)

const (
	CYDG_CARDS    = 110 //崇阳打拱扑克总数
	CYDG_MAX_CARD = 28  //4人玩法时每人的最大扑克数
)

const (
	MAX_ROUND        = 40 //最大下注轮数
	MAX_XIAZHU_COUNT = 50 //玩家每轮最大下注次数
	MAX_DANZHU       = 50 //最大单注
	ADDSCOREITEM     = 5  //每次显示的加注项的最大数目
	CHIPSITEM        = 5  //筹码数值个数

	MIN_PLAYER = 2 //最少人数

	//牌型
	CT_ERROR        = -1
	CT_SINGLE       = 0
	CT_DOUBLE       = 1
	CT_SHUN_ZI      = 2
	CT_TONG_HUA     = 3
	GT_TONG_HUA_SUN = 4
	CT_TRIPLE       = 5
	CT_SPECIAL      = 6
)

//操作命令
type Msg_S_OperateResult_JY struct {
	static.Msg_S_OperateResult
	OperateCode int  `json:"operatecode"` //操作代码
	ActionType  byte `json:"actiontype"`  //报清类型
}

//操作提示
type Msg_S_OperateNotify_JY struct {
	static.Msg_S_OperateNotify
	ActionMask int  `json:"actionmask"` //动作掩码
	ActionType byte `json:"actiontype"` //报清类型
}

//出牌
type Msg_S_SendCard_JY struct {
	static.Msg_S_SendCard
	ActionMask int `json:"actionmask"` //动作掩码
}

//操作命令
type Msg_C_OperateCard_JY struct {
	static.Msg_C_OperateCard
	Code int `json:"code"` //操作类型
}

//游戏状态
type CMD_S_StatusPlay_JY struct {
	static.CMD_S_StatusPlay
	ActionMask int `json:"actionmask"` //操作类型
}

//游戏状态
type Msg_S_GameStart_JY struct {
	static.Msg_S_GameStart
	UserAction int `json:"useraction"` //用户动作
}
