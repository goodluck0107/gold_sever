package fanlib

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	scoringlib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_scoring"
	cardmgr2 "github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*说明：
清一色：由一种花色的序数牌组成的和牌。不计无字。
*/

const (
	_SCORING_22_ID     = 22
	_SCORING_22_NAME   = "清一色"
	_SCORING_22_FANSHU = 24
)

// var _SCORING_22_DISCARDID_ = []int{75,76} //不计算"无字"、“缺一门”，
//20190121 现实中，硬清一色和硬混一色不能共存
var _SCORING_22_DISCARDID_ = []int{}

//自己注册
func init() {
	// fmt.Println("22")
	G_ScoringManager.RegisterBaseHander(&scoring_22{
		id:           _SCORING_22_ID,
		name:         _SCORING_22_NAME,
		fanShu:       _SCORING_22_FANSHU,
		setDiscardID: _SCORING_22_DISCARDID_,
		huKind:       scoringlib2.SCORING_NORMAL,
		humask:       scoringlib2.CANBE_ZIMO | scoringlib2.CANBE_CHIHU,
	})
}

type scoring_22 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int //
	humask       byte
}

func (this *scoring_22) GetID() int {
	return this.id
}

func (this *scoring_22) Name() string {
	return this.name
}

func (this *scoring_22) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_22) GethuKind() int {
	return this.huKind
}

/*
//具体检查是不是符合要求，清一色的话就是所有牌都是一种花色
只有万条筒（标准）
*/
func (this *scoring_22) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	color, _ := handInfoCtx.GetAllCardItembase()
	//标准清一色，只能有万条筒，不能有风刻
	//首先过风刻
	if color&mahlib2.CARDS_WITHOUT_DRAGON != 0 || color&mahlib2.CARDS_WITHOUT_WIND != 0 {
		return false
	}
	//剩下的只能有一种颜色
	if color^mahlib2.CARDS_WITHOUT_BAMBOO == 0 || color^mahlib2.CARDS_WITHOUT_WAN == 0 || color^mahlib2.CARDS_WITHOUT_DOT == 0 {
		return true
	}
	return false
}

//这个用做特殊处理，如果有赖子，要根据实际情况重做手牌数据
func (this *scoring_22) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) (uint64, byte, []int) {
	var result uint64 = static.CHK_NULL
	// if handInfoCtx.ChiHuKind&(public.CHK_DA_HU_NOMAGIC|public.CHK_DA_HU_MAGIC) > 0 {
	if handInfoCtx.ChiHuKind != 0 {
		if handInfoCtx.ChiHuKind&(static.CHK_PING_HU_NOMAGIC|scoringlib2.MASK_SPECIAL_NOMAGIC) > 0 {
			if this.CheckSatisfySelf(handInfoCtx) {
				result |= static.CHK_DA_HU_NOMAGIC
			}
		}
		if handInfoCtx.ChiHuKind&(static.CHK_PING_HU_MAGIC|scoringlib2.MASK_SPECIAL_MAGIC) > 0 {
			handInfoCtx.DecreaseGui()
			if this.CheckSatisfySelf(handInfoCtx) {
				result |= static.CHK_DA_HU_MAGIC
			}
			handInfoCtx.RecoverGui()
		}
	}
	if result != 0 {
		var guiNum byte = 0
		if handInfoCtx.CheckGodOrg != nil {
			guiNum = handInfoCtx.CheckGodOrg.GodNum
		}
		if result&static.CHK_DA_HU_MAGIC == 0 || guiNum == 1 {
			return result | handInfoCtx.ChiHuKind, guiNum, this.setDiscardID
		}
		return result | handInfoCtx.ChiHuKind, guiNum, nil
	}
	return result, 0, nil
}

func (this *scoring_22) Check_base(cbCardIndex []byte, weaveItem []static.TagWeaveItem) (result bool) {
	color, _ := cardmgr2.GetAllCardItembase(cbCardIndex, weaveItem)
	if color&mahlib2.CARDS_WITHOUT_DRAGON != 0 || color&mahlib2.CARDS_WITHOUT_WIND != 0 {
		return false
	}
	//剩下的只能有一种颜色
	if color^mahlib2.CARDS_WITHOUT_BAMBOO == 0 || color^mahlib2.CARDS_WITHOUT_WAN == 0 || color^mahlib2.CARDS_WITHOUT_DOT == 0 {
		return true
	}
	return false
}

func (this *scoring_22) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
	result = static.CHK_NULL
	needgodNum = 0
	//_,err=common.CheckHandCardsSafe(cbCardIndex)
	_, err = mahlib2.CheckHandCardsSafe_ex(cbCardIndex, checkCard)
	if err != nil {
		return result, needgodNum, err
	}
	if checkCard > 0x37 && checkCard != static.INVALID_BYTE {
		return result, needgodNum, errors.New(fmt.Sprintf("AnalyseHu_Normal 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//生成，手牌数据（去赖子的），赖子计数 这里需要判断软硬胡的情况
	var checkCards []byte
	checkCards, _ = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, true)
	//先算硬的
	if this.Check_base(checkCards, weaveItem) {
		result |= static.CHK_DA_HU_NOMAGIC
	}
	checkCards, needgodNum = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	//软的
	if this.Check_base(checkCards, weaveItem) {
		result |= static.CHK_DA_HU_MAGIC
	}
	return result, needgodNum, nil
}
