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
混一色：由一种花色序数牌及字牌组成的和牌。
*/

const (
	_SCORING_49_ID     = 49
	_SCORING_49_NAME   = "混一色"
	_SCORING_49_FANSHU = 6
)

var _SCORING_49_DISCARDID_ = []int{22, 82}

// //自己注册
func init() {
	// fmt.Println("49")
	G_ScoringManager.RegisterBaseHander(&scoring_49{
		id:           _SCORING_49_ID,
		name:         _SCORING_49_NAME,
		fanShu:       _SCORING_49_FANSHU,
		setDiscardID: _SCORING_49_DISCARDID_,
		huKind:       scoringlib2.SCORING_NORMAL,
		humask:       scoringlib2.CANBE_ZIMO | scoringlib2.CANBE_CHIHU,
	})
}

type scoring_49 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int
	humask       byte
}

func (this *scoring_49) GetID() int {
	return this.id
}

func (this *scoring_49) Name() string {
	return this.name
}

func (this *scoring_49) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_49) GethuKind() int {
	return this.huKind
}

/*
两部分信息：
倒牌的和手牌：分别判断两个部分的花色只能是2种，然后异或一下就可以了
*/
func (this *scoring_49) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	color, _ := handInfoCtx.GetAllCardItembase()
	//标准清一色，只能有万条筒，不能有风刻
	//首先过风刻
	if color&mahlib2.CARDS_WITHOUT_DRAGON == 0 && color&mahlib2.CARDS_WITHOUT_WIND == 0 {
		return false
	}
	//屏蔽掉风 粗暴点
	color &= mahlib2.CARDS_WITHOUT_BAMBOO | mahlib2.CARDS_WITHOUT_WAN | mahlib2.CARDS_WITHOUT_DOT
	//剩下的只能有一种颜色 如果有赖子。。。。
	haveGod := false
	if handInfoCtx.CheckGodOrg != nil {
		haveGod = handInfoCtx.CheckGodOrg.GodNum > 0
	}
	if !haveGod {
		return color^mahlib2.CARDS_WITHOUT_BAMBOO == 0 || color^mahlib2.CARDS_WITHOUT_WAN == 0 || color^mahlib2.CARDS_WITHOUT_DOT == 0
	} else {
		return color == 0
	}
	//20190418 下面的检查有纰漏，没考虑有赖子的情况
	//if color^common.CARDS_WITHOUT_BAMBOO == 0 || color^common.CARDS_WITHOUT_WAN == 0 || color^common.CARDS_WITHOUT_DOT == 0 {
	//	return true
	//}
	return false
}

//混一色和清一色的重组应该一样的，试试
func (this *scoring_49) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) (uint64, byte, []int) {
	var result uint64 = static.CHK_NULL
	// handInfoCtx.CreatNewCheckCards()
	// guiindex, _ := common.CardToIndex(handInfoCtx.CheckCardEX.BaseInfo.ID)
	// handInfoCtx.CheckCardIndex[guiindex] += 1
	// if handInfoCtx.ChiHuKind&(public.CHK_PING_HU_NOMAGIC|public.CHK_PING_HU_MAGIC) != 0 {
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

//发现一种情况，如果除了赖子，其他的牌都是风箭，这种也能算是混一色
func (this *scoring_49) Check_base(cbCardIndex []byte, weaveItem []static.TagWeaveItem, godNUm byte) (result bool) {
	color, _ := cardmgr2.GetAllCardItembase(cbCardIndex, weaveItem)
	if color&mahlib2.CARDS_WITHOUT_DRAGON == 0 && color&mahlib2.CARDS_WITHOUT_WIND == 0 {
		return false
	}
	//屏蔽掉风 粗暴点
	color &= mahlib2.CARDS_WITHOUT_BAMBOO | mahlib2.CARDS_WITHOUT_WAN | mahlib2.CARDS_WITHOUT_DOT
	//剩下的只能有一种颜色 如果没赖子
	if godNUm == 0 {
		return color^mahlib2.CARDS_WITHOUT_BAMBOO == 0 || color^mahlib2.CARDS_WITHOUT_WAN == 0 || color^mahlib2.CARDS_WITHOUT_DOT == 0
	} else {
		return color == 0
	}
	//
	//if color^common.CARDS_WITHOUT_BAMBOO == 0 || color^common.CARDS_WITHOUT_WAN == 0 || color^common.CARDS_WITHOUT_DOT == 0 {
	//if godNUm==0{
	//	return true
	//}else{
	//	return false
	//}
	//}
	return false
}

func (this *scoring_49) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
	result = static.CHK_NULL
	needgodNum = 0
	//_,err=common.CheckHandCardsSafe(cbCardIndex)
	_, err = mahlib2.CheckHandCardsSafe_ex(cbCardIndex, checkCard)
	if err != nil {
		//目前7对只能门清
		return result, needgodNum, err
	}
	if checkCard > 0x37 && checkCard != static.INVALID_BYTE {
		return result, needgodNum, errors.New(fmt.Sprintf("AnalyseHu_Normal 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//生成，手牌数据（去赖子的），赖子计数 这里需要判断软硬胡的情况
	var checkCards []byte
	checkCards, _ = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, true)

	//先算硬的
	if this.Check_base(checkCards, weaveItem, 0) {
		result |= static.CHK_DA_HU_NOMAGIC
	}
	checkCards, needgodNum = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	//软的
	if this.Check_base(checkCards, weaveItem, needgodNum) {
		result |= static.CHK_DA_HU_MAGIC
	}
	return result, needgodNum, nil
}
