package eve

import (
	"github.com/open-source/game/chess.git/pkg/static"
)

type TagTableFriend struct {
	//	enum { AGREE=1, DISAGREE, DISMISS_CREATOR, WATING, DISSMISS_MAX };  //同意 不同意 解散房间发起者  等待中（尚未决定同意与否）
	CreateTime       int64                    //刷新时间
	CreateUserID     int64                    //创建者
	GameNum          string                   //好友房对局的唯一标识
	Name             string                   //桌子名称
	Password         string                   //桌子密码
	RoomID           int                      //桌子ID
	ServerID         uint16                   //ServerID
	TableID          uint16                   //桌子ID
	Rule             string                   //桌子规则
	AgreeDismiss     byte                     //同意解散的人数
	JuShu            uint32                   //局数
	CreateType       byte                     //房间类型
	Type             uint16                   //房间状态 未开始：0   已经开始：1   已解散：2
	MissItem         [static.MAX_CHAIR]byte   //记录每个玩家的同意拒绝信息(其中下表0 1 2 3依次表示4个椅子)
	DismissOptTime   [static.MAX_CHAIR]string //记录每个玩家的同意拒绝信息时的时间
	MissCause        string                   //20200227 苏大强 申请解算的原因
	Dismiss_end_time int64                    //解散房间倒计时（某个玩家申请解散以后，默认最后解散时间秒数，收到解散请求时间+10分钟）
	DismissType      int                      //不同的type走解散逻辑 0是正常 1：离线玩家默认同意解散
	MissStatus       [static.MAX_CHAIR]uint64 //解散时记录每个玩家的状态
	FewerItem        [static.MAX_CHAIR]byte   //记录每个玩家的同意拒绝少人开局的信息(其中下标0 1 2 3依次表示4个椅子)
	FewerShow        bool                     //是否显示少人开局
	AgreeFewer       byte                     //同意少人开局的人数
	KindID           uint32
	AAPay            byte //是否AA支付
	JuShuPay         byte //是否按局数支付

	/********************************Begin: 包厢模块**************************************************/
	TeaHouseTableIndex uint32 //包厢桌子ID
	TeaHouseID         string //包厢ID
	//teaHouseCaptionId      int64  //楼主ID
	CostMoney uint16 //玩一轮需要扣多少
	/********************************包厢模块***************************************************/
}

func (self *TagTableFriend) InitMissItem() {
	for i := 0; i < len(self.MissItem); i++ {
		self.MissItem[i] = 0
	}
	for i := 0; i < len(self.MissStatus); i++ {
		self.MissStatus[i] = 0
	}
	for i := 0; i < len(self.DismissOptTime); i++ {
		self.DismissOptTime[i] = ""
	}
	self.Dismiss_end_time = 0
	self.MissCause = ""
}

func (self *TagTableFriend) InitFewer() {
	for i := 0; i < len(self.FewerItem); i++ {
		self.FewerItem[i] = 0
	}
	self.FewerShow = false
	self.AgreeFewer = 0
}

func (self *TagTableFriend) FewerIsNil() bool {
	for i := 0; i < len(self.FewerItem); i++ {
		if self.FewerItem[i] != 0 {
			return false
		}
	}
	return true
}
