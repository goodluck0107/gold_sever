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
碰碰和：由4副刻子（或杠）、将牌组成的和牌。又称“对对和(胡)”
*/

const (
	_SCORING_48_ID     = 48
	_SCORING_48_NAME   = "碰碰和"
	_SCORING_48_FANSHU = 6
)

var _SCORING_48_DISCARDID_ = []int{}

// //自己注册
func init() {
	// fmt.Println("48")
	G_ScoringManager.RegisterBaseHander(&scoring_48{
		id:           _SCORING_48_ID,
		name:         _SCORING_48_NAME,
		fanShu:       _SCORING_48_FANSHU,
		setDiscardID: _SCORING_48_DISCARDID_,
		huKind:       scoringlib2.SCORING_NORMAL,
		humask:       scoringlib2.CANBE_ZIMO | scoringlib2.CANBE_CHIHU,
	})
}

type scoring_48 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int
	humask       byte
}

func (this *scoring_48) GetID() int {
	return this.id
}

func (this *scoring_48) Name() string {
	return this.name
}

func (this *scoring_48) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_48) GethuKind() int {
	return this.huKind
}

/*
//具体检查是不是符合要求
20190123
*/
func (this *scoring_48) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	result := false
	//WarnCTXinfo := handInfoCtx.GetWarninfo()
	tripletNum := 0
	if handInfoCtx.ReWarnCTXinfo != nil {
		if len(handInfoCtx.ReWarnCTXinfo.ChipaiInfo) > 0 {
			return result
		}
		tripletNum += len(handInfoCtx.ReWarnCTXinfo.Triplet) + len(handInfoCtx.ReWarnCTXinfo.HidTriplet) + len(handInfoCtx.ReWarnCTXinfo.PengInfo)
	}
	var godNum byte = 0
	if handInfoCtx.CheckGodOrg != nil {
		godNum = handInfoCtx.CheckGodOrg.GodNum
	}
	handInfoCtx.DecreaseGui()
	var need byte = 0
	for i := 0; i < 34; i++ {
		remnant := handInfoCtx.CheckCardItem[i] % 3
		if remnant != 0 {
			need = need + 3 - remnant
		}
	}
	//这里可能有问题了
	if need <= godNum+1 {
		groupinfo := cardmgr2.ClassifyCards(handInfoCtx.CheckCardItem)
		//20190130 苏大强 考虑到可能赖子当刻的情况
		fill := 0
		if (godNum+1-need)%3 == 0 {
			fill = int((godNum + 1 - need) / 3)
		}
		if tripletNum+len(groupinfo.OneCards)+len(groupinfo.FourCards)*2+len(groupinfo.TwoCards)+len(groupinfo.ThreeCards)+fill == 5 {
			result = true
		}
	}
	handInfoCtx.RecoverGui()
	return result
}

//碰碰胡一定要符合3n+2
func (this *scoring_48) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) (uint64, byte, []int) {
	//---------如果是258的，要优先判断，手牌一定要有258才行----------
	//这样看来，只有两种方式了，判断mask258
	//-----------------------------------
	var result uint64 = static.CHK_NULL
	//20190130 苏大强 修改碰碰胡不用查胡表的结果，自行判断
	//强制258需要这样
	// if handInfoCtx.ChiHuKind&(public.CHK_PING_HU_NOMAGIC|public.CHK_PING_HU_MAGIC) != 0 {
	// if handInfoCtx.ChiHuKind != 0 {
	if handInfoCtx.Mask258 != 0 && handInfoCtx.ChiHuKind&0xf == 0 {
		return result, 0, nil
	}
	var needGui byte = 0
	if this.CheckSatisfySelf(handInfoCtx) {
		handInfoCtx.DecreaseGui()
		groupinfo := cardmgr2.ClassifyCards(handInfoCtx.CheckCardItem)
		handInfoCtx.RecoverGui()
		// needGui := len(groupinfo.OneCards)*2 + len(groupinfo.FourCards)*2 + len(groupinfo.TwoCards) - 1
		//鬼牌剩下的值0/2/3
		//20190130 苏大强 碰碰胡的检查规则不用依赖查胡表，但是有肯能要求258，反正都是
		//20190305 加一个计算鬼牌的
		needGui = byte((len(groupinfo.OneCards)+len(groupinfo.FourCards))*2+len(groupinfo.TwoCards)) - 1
		if needGui > 0 {
			result |= static.CHK_DA_HU_MAGIC
			//  handInfoCtx.SetchiHuKind(public.CHK_PING_HU_MAGIC)
		} else {
			//20190418 有可能刚好就3个赖子的情况
			if handInfoCtx.CheckGodOrg != nil {
				if handInfoCtx.CheckGodOrg.GodNum > 2 {
					result |= static.CHK_DA_HU_MAGIC
				}
			}
			result |= static.CHK_DA_HU_NOMAGIC
			//20190221 如果赖子牌是3的倍数。那么可以软胡,这个暂时失去意义了先去掉
			//if handInfoCtx.CheckGodOrg!=nil&&handInfoCtx.CheckGodOrg.GodNum%3==0{
			//	result |= public.CHK_DA_HU_MAGIC
			//}
			//  handInfoCtx.SetchiHuKind(public.CHK_PING_HU_NOMAGIC)
		}
	}
	if result != 0 {
		if result&static.CHK_DA_HU_MAGIC == 0 || needGui == 1 {
			return result | handInfoCtx.ChiHuKind, needGui, this.setDiscardID
		}
		return result | handInfoCtx.ChiHuKind, needGui, nil
	}

	return result, 0, nil
}
func (this *scoring_48) Check_base(cbCardIndex []byte, godNum byte, weaveItem []static.TagWeaveItem) (result bool, needGui byte) {
	tripletNum := 0
	result = false
	weaveinfo := cardmgr2.GetWarninfo(weaveItem)
	if weaveinfo != nil {
		if len(weaveinfo.ChipaiInfo) > 0 {
			return result, 0
		}
		tripletNum += len(weaveinfo.Triplet) + len(weaveinfo.HidTriplet) + len(weaveinfo.PengInfo)
	}
	var need byte = 0
	for i := 0; i < 34; i++ {
		remnant := cbCardIndex[i] % 3
		if remnant != 0 {
			need = need + 3 - remnant
		}
	}
	if need <= godNum+1 {
		groupinfo := cardmgr2.ClassifyCards(cbCardIndex)
		record := len(groupinfo.OneCards) + len(groupinfo.FourCards)*2 + len(groupinfo.TwoCards)
		needGui = byte(record+len(groupinfo.OneCards)) - 1

		if need == needGui+1 {
			result = true
		}
		//fill := 0
		//if (godNum+1-need)%3 == 0 {
		//	fill = int((godNum + 1 - need) / 3)
		//}
		//if tripletNum+record+len(groupinfo.ThreeCards)+fill == 5 {
		//	result = true
		//}
	}
	return result, needGui
}

//碰碰胡这个normal检查不检查258.258通过checkhu来
func (this *scoring_48) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
	result = static.CHK_NULL
	needgodNum = 0
	//num,err=common.CheckHandCardsSafe(cbCardIndex)
	_, err = mahlib2.CheckHandCardsSafe_ex(cbCardIndex, checkCard)
	if err != nil {
		//目前7对只能门清
		return result, needgodNum, err
	}
	//if num>2{
	//	return result,needgodNum,nil
	//}
	if checkCard > 0x37 && checkCard != static.INVALID_BYTE {
		return result, needgodNum, errors.New(fmt.Sprintf("碰碰胡 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//生成，手牌数据（去赖子的），赖子计数 这里需要判断软硬胡的情况
	var checkCards []byte
	checkCards, needgodNum = cardmgr2.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	if isok, guiNum := this.Check_base(checkCards, needgodNum, weaveItem); isok {
		if guiNum > 0 && guiNum < 255 {
			result |= static.CHK_DA_HU_MAGIC
			//  handInfoCtx.SetchiHuKind(public.CHK_PING_HU_MAGIC)
		} else {
			result |= static.CHK_DA_HU_NOMAGIC
			//20190221 如果赖子牌是3的倍数。那么可以软胡,这个暂时失去意义了先去掉
			//if handInfoCtx.CheckGodOrg!=nil&&handInfoCtx.CheckGodOrg.GodNum%3==0{
			//	result |= public.CHK_DA_HU_MAGIC
			//}
			//  handInfoCtx.SetchiHuKind(public.CHK_PING_HU_NOMAGIC)
		}
		//needgodNum=godNum-guiNum
	}
	return result, needgodNum, nil
}
