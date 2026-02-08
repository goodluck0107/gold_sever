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
	_SCORING_52_ID     = 52
	_SCORING_52_NAME   = "全求人"
	_SCORING_52_FANSHU = 6
)

// var _SCORING_52_CHECKID_ = []int{79} //
var _SCORING_52_CHECKID_ = []int{} //
//自己注册
func init() {
	// fmt.Println("52")
	G_ScoringManager.RegisterBaseHander(&scoring_52{
		id:           _SCORING_52_ID,
		name:         _SCORING_52_NAME,
		fanShu:       _SCORING_52_FANSHU,
		setDiscardID: _SCORING_52_CHECKID_,
		huKind:       scoringlib2.SCORING_NORMAL,
		humask:       scoringlib2.CANBE_CHIHU,
	})
}

type scoring_52 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int
	humask       byte
}

func (this *scoring_52) GetID() int {
	return this.id
}

func (this *scoring_52) Name() string {
	return this.name
}

func (this *scoring_52) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_52) GethuKind() int {
	return this.huKind
}

//不能直接用这个就结束
func (this *scoring_52) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	//倒牌里面不能有暗杠
	result := false
	//WarnCTXinfo := handInfoCtx.GetWarninfo()
	if handInfoCtx.ReWarnCTXinfo == nil {
		return result
	}
	if len(handInfoCtx.ReWarnCTXinfo.HidTriplet) != 0 || len(handInfoCtx.ReWarnCTXinfo.ChipaiInfo)+len(handInfoCtx.ReWarnCTXinfo.Triplet)+len(handInfoCtx.ReWarnCTXinfo.PengInfo) != 4 {
		return result
	}
	//国标上不能自摸，但是各地有可能自摸所有，能不能靠自摸胡到，需要检查一下，江铃上可以
	return true
}

//20190130 若是全求人不要258，那么这里要修改
/*
CheckSatisfySelf里面已将把全球人限死了状态必须是手上只有一个单张了
*/

func (this *scoring_52) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) (uint64, byte, []int) {
	var result uint64 = static.CHK_NULL
	// if handInfoCtx.ChiHuKind&(public.CHK_PING_HU_NOMAGIC|public.CHK_PING_HU_MAGIC) != 0 {
	// if handInfoCtx.ChiHuKind != 0 {
	if handInfoCtx.Mask258 != 0 && handInfoCtx.ChiHuKind&0xf == 0 {
		return result, 0, nil
	}
	var godnum byte = 0
	if this.CheckSatisfySelf(handInfoCtx) {
		handInfoCtx.DecreaseGui()
		groupinfo := cardmgr2.ClassifyCards(handInfoCtx.CheckCardItem)
		switch len(groupinfo.OneCards) + len(groupinfo.TwoCards) {
		case 0:
			//赖对，这个可以算硬胡的，也能算软胡 20190305 区分与3n+2的判断，并且要算最少需要几个赖子
			godnum = 2
			result |= static.CHK_DA_HU_MAGIC | static.CHK_DA_HU_NOMAGIC
			// handInfoCtx.SetchiHuKind(public.CHK_PING_HU_MAGIC|public.CHK_PING_HU_NOMAGIC)
		case 1:
			if len(groupinfo.OneCards) == 1 {
				godnum = 1
				result |= static.CHK_DA_HU_MAGIC
				// handInfoCtx.SetchiHuKind(public.CHK_PING_HU_MAGIC)
			} else {
				result |= static.CHK_DA_HU_NOMAGIC
				// handInfoCtx.SetchiHuKind(public.CHK_PING_HU_NOMAGIC)
			}
		}
		handInfoCtx.RecoverGui()
		//---------------------------------------
		// if handInfoCtx.ChiHuKind&public.CHK_PING_HU_NOMAGIC > 0 {
		// 	result |= public.CHK_DA_HU_NOMAGIC
		// }
		// if handInfoCtx.ChiHuKind&public.CHK_PING_HU_MAGIC > 0 {
		// 	result |= public.CHK_DA_HU_MAGIC
		// }
		//------------------------------------
		//20190305 这个比较特殊 单吊将要么2个赖子成硬胡，要么1个赖子
		//20190418
		if result&static.CHK_DA_HU_MAGIC == 0 || godnum == 1 {
			return result | handInfoCtx.ChiHuKind, godnum, this.setDiscardID
		}
		//if result^(public.CHK_DA_HU_MAGIC | public.CHK_DA_HU_NOMAGIC)==0|| result&public.CHK_DA_HU_MAGIC==0{
		//	return result | handInfoCtx.ChiHuKind,0, this.setDiscardID
		//}
		return result | handInfoCtx.ChiHuKind, godnum, nil
	}
	// } else {
	// }
	return static.CHK_NULL, 0, nil
}
func (this *scoring_52) Check_base(weaveItem []static.TagWeaveItem) (result bool) {
	result = false
	weaveinfo := cardmgr2.GetWarninfo(weaveItem)
	if weaveinfo == nil {
		return result
	}
	if len(weaveinfo.HidTriplet) != 0 || len(weaveinfo.ChipaiInfo)+len(weaveinfo.Triplet)+len(weaveinfo.PengInfo) != 4 {
		return result
	}
	return true
}
func (this *scoring_52) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
	result = static.CHK_NULL
	needgodNum = 0
	//num,err:=common.CheckHandCardsSafe(cbCardIndex)
	num := 0
	num, err = mahlib2.CheckHandCardsSafe_ex(cbCardIndex, checkCard)
	if err != nil || num > 2 {
		//全球人就1章了
		return result, needgodNum, err
	}
	if checkCard > 0x37 && checkCard != static.INVALID_BYTE {
		return result, needgodNum, errors.New(fmt.Sprintf("AnalyseHu_Normal 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//生成，手牌数据（去赖子的），赖子计数 这里需要判断软硬胡的情况
	var checkCards []byte
	checkCards, needgodNum = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	if this.Check_base(weaveItem) {
		//去掉赖子 这里只会出现2张非赖子，1张赖子，2张赖子，2张孤章（如果优先判断了3n+2，就不会了）
		groupinfo := cardmgr2.ClassifyCards(checkCards)
		switch len(groupinfo.OneCards) + len(groupinfo.TwoCards) {
		case 0:
			//赖对，这个可以算硬胡的，也能算软胡 20190305 区分与3n+2的判断，并且要算最少需要几个赖子
			result |= static.CHK_DA_HU_MAGIC | static.CHK_DA_HU_NOMAGIC
			// handInfoCtx.SetchiHuKind(public.CHK_PING_HU_MAGIC|public.CHK_PING_HU_NOMAGIC)
		case 1:
			if len(groupinfo.OneCards) == 1 {
				result |= static.CHK_DA_HU_MAGIC
				// handInfoCtx.SetchiHuKind(public.CHK_PING_HU_MAGIC)
			} else {
				result |= static.CHK_DA_HU_NOMAGIC
				// handInfoCtx.SetchiHuKind(public.CHK_PING_HU_NOMAGIC)
			}
		}
	}
	return result, needgodNum, nil
}
