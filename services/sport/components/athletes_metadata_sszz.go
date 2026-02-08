package components

import (
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
)

//游戏流程数据
type PlayerMetaSZ struct {
	PlayerCards [static.MAX_CARD_P3]byte // 玩家分到的牌

	BqiPai        bool //是否弃牌
	BiPaiShu      bool //比牌输
	Xiazhu        byte // 下注状态，0表示下注，1表示加注，2表示全压 ,初始值0xFF
	BkanPai       bool // 记录玩家看牌
	TheQipaiIndex byte //

	EveryOneCurRoundXiaZhu [info2.MAX_XIAZHU_COUNT]int //每个玩家当前轮下注的分数，可能当前轮下了两次，所以使用数组
	PlayerXiazhuscore      [info2.MAX_ROUND + 1]int    // 玩家每轮下的分数,包括底注
	PlayerLoseScore        int                         // 玩家输的分

	//ThePlayerState byte // 玩家状态，每局初始化成null，游戏开始后playing，弃牌不变,离开后null,
	ActionState byte // 玩家每轮发牌后是否已经操作,0表示没有操作，非0是操作次数

	IsPlaying bool //是否正在游戏
	//CanSeeOthersCards []bool   //
	BautoGenZhu bool     //自动跟注
	ShowCard    []uint16 //是否显示牌型
}

func (self *PlayerMetaSZ) SetCardIndex(cbCardData []byte) {
	for i := 0; i < static.MAX_CARD_P3; i++ {
		self.PlayerCards[i] = cbCardData[i]
	}
}

func (self *PlayerMetaSZ) OnNextGame() {
	self.IsPlaying = true
	self.BqiPai = false
	self.BiPaiShu = false
	self.BkanPai = false
	self.PlayerXiazhuscore = [info2.MAX_ROUND + 1]int{} // 玩家每轮下的分数,包括底注
	self.ActionState = 0
	self.PlayerLoseScore = 0
	self.PlayerCards = [static.MAX_CARD_P3]byte{} // 玩家分到的牌
	self.BautoGenZhu = false
	self.ShowCard = []uint16{}
}

func (self *PlayerMetaSZ) OnEnd() {
	self.IsPlaying = false
}

//弃牌
func (self *PlayerMetaSZ) QiPai(seat uint16, round byte) {
	self.ActionState++
	self.BqiPai = true
	self.TheQipaiIndex = round //断线重连需要知道玩家在第几轮弃牌，从1开始标号
	self.Xiazhu = 0xFF
	//self.IsPlaying = false
	self.AddShowCard(seat)
}

//比牌
func (self *PlayerMetaSZ) BiPai(seat uint16, _item *Player) {
	//self.ShowCard = append(self.ShowCard, seat)
	//self.ShowCard = append(self.ShowCard, _item.Seat)
	self.AddShowCard(seat)
	self.AddShowCard(_item.Seat)

	_item.Ctx.AddShowCard(seat)
	_item.Ctx.AddShowCard(_item.Seat)

	//_item.Ctx.ShowCard = append(_item.Ctx.ShowCard, seat)
	//_item.Ctx.ShowCard = append(_item.Ctx.ShowCard, _item.Seat)
}

func (self *PlayerMetaSZ) AddShowCard(seat uint16) {
	for _, v := range self.ShowCard {
		if v == seat {
			return
		}
	}
	self.ShowCard = append(self.ShowCard, seat)
}
