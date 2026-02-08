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
	_SCORING_83_ID     = 83
	_SCORING_83_NAME   = "将一色"
	_SCORING_83_FANSHU = 6
)

var _SCORING_83_CHECKID_ = []int{49} //
//自己注册
func init() {
	// fmt.Println("83")
	G_ScoringManager.RegisterBaseHander(&scoring_83{
		id:           _SCORING_83_ID,
		name:         _SCORING_83_NAME,
		fanShu:       _SCORING_83_FANSHU,
		setDiscardID: _SCORING_83_CHECKID_,
		huKind:       scoringlib2.SCORING_SPECIAL,
		humask:       scoringlib2.CANBE_CHIHU,
	})
}

type scoring_83 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int
	humask       byte
}

func (this *scoring_83) GetID() int {
	return this.id
}

func (this *scoring_83) Name() string {
	return this.name
}

func (this *scoring_83) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_83) GethuKind() int {
	return this.huKind
}

//不能直接用这个就结束
func (this *scoring_83) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	color, is258 := handInfoCtx.GetAllCardItembase()
	//将将胡不能有风箭
	// if color&(common.CARDS_WITHOUT_BAMBOO|common.CARDS_WITHOUT_WAN|common.CARDS_WITHOUT_DOT) != 0 {
	// 	return false
	// }
	if color&mahlib2.CARDS_WITHOUT_DRAGON != 0 || color&mahlib2.CARDS_WITHOUT_WIND != 0 {
		return false
	}
	return is258
}

//将将胡能胡硬的，优先硬的
func (this *scoring_83) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) (uint64, byte, []int) {
	var result uint64 = static.CHK_NULL
	// if handInfoCtx.ChiHuKind != 0 {
	// if handInfoCtx.ChiHuKind&public.CHK_DA_HU_NOMAGIC > 0 {
	if this.CheckSatisfySelf(handInfoCtx) {
		//不计3n+2情况下，如果赖子也是258，那么就是硬胡
		result |= static.CHK_DA_HU_NOMAGIC
		handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_NOMAGIC)
	} else {
		if handInfoCtx.CheckGodOrg != nil {
			handInfoCtx.DecreaseGui()
			if this.CheckSatisfySelf(handInfoCtx) {
				result |= static.CHK_DA_HU_MAGIC
				handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_MAGIC)
			}
			handInfoCtx.RecoverGui()
		}
	}
	// return result | handInfoCtx.ChiHuKind, this.setDiscardID
	// }
	if result != 0 {
		var guiNum byte = 0
		if handInfoCtx.CheckGodOrg != nil {
			guiNum = handInfoCtx.CheckGodOrg.GodNum
		}
		//20190418 如果是硬的将一色，那么才去掉相抵触的，软清一色只要赖子大于1那么就不能干掉
		if result&static.CHK_DA_HU_MAGIC == 0 || guiNum == 1 {
			return result | handInfoCtx.ChiHuKind, guiNum, this.setDiscardID
		}
		return result | handInfoCtx.ChiHuKind, guiNum, nil
	}
	return result, 0, this.setDiscardID
}
func (this *scoring_83) Check_base(cbCardIndex []byte, weaveItem []static.TagWeaveItem) (result bool) {
	color, is258 := cardmgr2.GetAllCardItembase(cbCardIndex, weaveItem)
	if color&mahlib2.CARDS_WITHOUT_DRAGON != 0 || color&mahlib2.CARDS_WITHOUT_WIND != 0 {
		return false
	}
	return is258
}

func (this *scoring_83) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
	result = static.CHK_NULL
	needgodNum = 0
	//_,err=common.CheckHandCardsSafe(cbCardIndex)
	_, err = mahlib2.CheckHandCardsSafe_ex(cbCardIndex, checkCard)
	if err != nil {
		return result, needgodNum, err
	}
	if checkCard > 0x37 && checkCard != static.INVALID_BYTE {
		return result, needgodNum, errors.New(fmt.Sprintf("将一色 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//生成，手牌数据（去赖子的），赖子计数 这里需要判断软硬胡的情况
	var checkCards []byte
	checkCards, _ = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, true)
	//先算硬的
	if this.Check_base(checkCards, weaveItem) {
		result |= static.CHK_DA_HU_NOMAGIC
	}
	//软的
	checkCards, needgodNum = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	if this.Check_base(checkCards, weaveItem) {
		result |= static.CHK_DA_HU_MAGIC
	}
	return result, needgodNum, nil
}
