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
全求人：全靠吃牌、碰牌、单钓别人打出的牌和牌。不计单钓。
*/

const (
	_SCORING_82_ID     = 82
	_SCORING_82_NAME   = "风一色"
	_SCORING_82_FANSHU = 6
)

var _SCORING_82_CHECKID_ = []int{22, 49} //
//自己注册
func init() {
	// fmt.Println("82")
	G_ScoringManager.RegisterBaseHander(&scoring_82{
		id:           _SCORING_82_ID,
		name:         _SCORING_82_NAME,
		fanShu:       _SCORING_82_FANSHU,
		setDiscardID: _SCORING_82_CHECKID_,
		huKind:       scoringlib2.SCORING_SPECIAL,
		humask:       scoringlib2.CANBE_CHIHU,
	})
}

type scoring_82 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int
	humask       byte
}

func (this *scoring_82) GetID() int {
	return this.id
}

func (this *scoring_82) Name() string {
	return this.name
}

func (this *scoring_82) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_82) GethuKind() int {
	return this.huKind
}

//不能直接用这个就结束
func (this *scoring_82) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	color, _ := handInfoCtx.GetAllCardItembase()
	//风一色和清一色正好反着来
	if color&(mahlib2.CARDS_WITHOUT_BAMBOO|mahlib2.CARDS_WITHOUT_WAN|mahlib2.CARDS_WITHOUT_DOT) != 0 {
		return false
	}
	// if color&common.CARDS_WITHOUT_DRAGON != 0 || color&common.CARDS_WITHOUT_WIND != 0 {
	// 	return false
	// }
	return true
}

func (this *scoring_82) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) (uint64, byte, []int) {
	var result uint64 = static.CHK_NULL
	// if handInfoCtx.ChiHuKind != 0 {
	// if handInfoCtx.ChiHuKind&public.CHK_DA_HU_NOMAGIC > 0 {
	if this.CheckSatisfySelf(handInfoCtx) {
		result |= static.CHK_DA_HU_NOMAGIC
		handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_NOMAGIC)
	}
	// }
	// if handInfoCtx.ChiHuKind&public.CHK_DA_HU_MAGIC > 0 {
	if handInfoCtx.CheckGodOrg != nil {
		handInfoCtx.DecreaseGui()
		if this.CheckSatisfySelf(handInfoCtx) {
			result |= static.CHK_DA_HU_MAGIC
			handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_MAGIC)
		}
		handInfoCtx.RecoverGui()
	}
	// }
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
	// }
	// return result, this.setDiscardID
}
func (this *scoring_82) Check_base(cbCardIndex []byte, weaveItem []static.TagWeaveItem) (result bool) {
	color, _ := cardmgr2.GetAllCardItembase(cbCardIndex, weaveItem)
	if color&(mahlib2.CARDS_WITHOUT_BAMBOO|mahlib2.CARDS_WITHOUT_WAN|mahlib2.CARDS_WITHOUT_DOT) != 0 {
		return false
	}
	return true

}

func (this *scoring_82) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
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
	checkCards, _ = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	//先算硬的
	if this.Check_base(checkCards, weaveItem) {
		result |= static.CHK_DA_HU_NOMAGIC
	}
	checkCards, needgodNum = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, true)
	//软的
	if this.Check_base(checkCards, weaveItem) {
		result |= static.CHK_DA_HU_MAGIC
	}
	return result, needgodNum, nil
}
