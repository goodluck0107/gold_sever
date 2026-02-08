package lib_win

import (
	//"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

/*
特殊坎 大三元
*/
type SanKanObj_dsy struct {
	handCards  []byte //去掉三坎后的数据
	index      byte   //三坎的下标
	lostCards  byte   //姐妹铺去牌数据 如果还原的话，用这个参考加回去，这个和3坎不一样，有可能是赖子补的
	useGodNum  byte
	lastGodNum byte //剩下鬼牌
}

/*
最多2组
*/
type SanKanResult_dsy_dsy struct {
	handCards  []byte  //去掉三坎的数据
	index      [3]byte //三坎下标
	lostCards  [3]byte //去掉的三坎数据
	lastGodNum byte    //剩下鬼牌
	//isDouble bool  //是不是  3坎一定是去掉3个
	huKind byte //硬胡 258标志
}

//-----------------------------------------------

//因为倒牌里面是死的，所以创建一下就好了
func (this *CheckHu) createWeaveItemForSiJiFengAndDaShanYuan(WeaveItem []static.TagWeaveItem, checkJFnum byte) (sjfMask byte, dsyMask byte) {
	isjifeng := true
	isSanDaJiang := true
	var chiNum byte = 0
	for _, v := range WeaveItem {
		if chiNum > 2 {
			//不可能三大将
			dsyMask = 0
			isSanDaJiang = false
		}
		if chiNum > checkJFnum-1 {
			//不可能季风
			sjfMask = 0
			isjifeng = false
		}
		//目前倒牌里面只有碰杠
		switch v.CenterCard {
		case 0x31:
			//东
			if isjifeng {
				sjfMask |= 0x8
			}
		case 0x32:
			//南
			if isjifeng {
				sjfMask |= 0x4
			}
		case 0x33:
			//西
			if isjifeng {
				sjfMask |= 0x2
			}
		case 0x34:
			//北
			if isjifeng {
				sjfMask |= 0x1
			}
		case 0x35:
			//红中
			if isSanDaJiang {
				dsyMask |= 0x4
			}
		case 0x36:
			//发财
			if isSanDaJiang {
				dsyMask |= 0x2
			}
		case 0x37:
			//白板
			if isSanDaJiang {
				dsyMask |= 0x1
			}
		default:
			if v.CenterCard != 0 {
				chiNum += 1
			}
		}
	}
	return
}

/*
//针对成刻的情况 独立一个单元 kenum可以是3也可以是2,
注意这里会修改手牌
*/
func RepairKeCards(checkcards []byte, index byte, kenum byte, godNum *byte) (resutl bool, losecard byte) {
	resutl = false
	losecard = 0
	if (checkcards[index] + *godNum) < kenum {
		//这个不能成刻
		return
	}
	if checkcards[index]/kenum > 0 {
		checkcards[index] -= kenum
		losecard = kenum
	} else {
		//不够求余，因为有为0的情况
		needgod := kenum - checkcards[index]
		losecard = checkcards[index]
		*godNum -= needgod
	}
	//-----------------
	return true, losecard
}

/*
三季风有个恶心的情况  2个东风 1个南风 2个西风 2个北风 手上3个赖子，只有优先处理缺少最少的牌，就是东西北才行，如果顺序就挂了
4季风就简单了，直接配完，所以要修改一下，逻辑上优先处理缺少最少的牌
要记录空位 和缺少位？
checkJFnum 带表是3季还是4季
level 是优先处理 第一次是优先处理2张的，然后是1张的，最后是空的
*/
func (this *CheckHu) AnalyseJiFeng(jfMask *byte, cbCardIndex []byte, guiNum *byte, checkJFnum byte, level int) (result bool) {
	//循环记录
	if level < 0 {
		return false
	}
	var record byte = 0
	var i byte
	for i = 0; i < 4; i++ {
		if *jfMask&(1<<i) > 0 {
			//已经够了
			record += 1
			continue
		}
		//处理检查
		cardIndex, _ := mahlib2.CardToIndex(0x34 - i)
		if cbCardIndex[cardIndex] < byte(level) {
			//达不到处理等级
			continue
		}
		if ok, _ := RepairKeCards(cbCardIndex, cardIndex, 3, guiNum); ok {
			*jfMask |= 1 << i
			record += 1
		}
	}
	if record >= checkJFnum {
		return true
	} else {
		result = this.AnalyseJiFeng(jfMask, cbCardIndex, guiNum, checkJFnum, level-1)
	}
	return
}

//20190702 苏大强 大三元和风 这个没有倒牌也能检查 cbCardIndex是全牌
func (this *CheckHu) AnalyseSiJiFengAndDaShanYuan(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbCurrentCard byte, isNormalCard bool, godCards []byte, checkJFnum byte, ChiHuResult *static.TagChiHuResult, usegod bool, seaf bool) byte {
	//不能超过2个非风的碰牌，吃也是，超过2个吃也就不行了
	checkGodCards := []byte{}
	if len(godCards) != 0 {
		static.HF_DeepCopy(&checkGodCards, &godCards)
	}
	if !usegod {
		checkGodCards = []byte{}
	}
	checkcards, handNum, guiNum, err := ReSetHandCards_Nomal(cbCardIndex, cbCurrentCard, isNormalCard, checkGodCards, seaf)
	if err != nil {
		fmt.Println(fmt.Sprintf("检查大三元和风出问题（%v）", err))
		return 0
	}
	//大三元和风不检查手牌数
	_ = handNum
	var i byte
	jfMask, dsyMask := this.createWeaveItemForSiJiFengAndDaShanYuan(WeaveItem, checkJFnum)
	//在倒牌里面就能断定能不能有2种胡，如果没有那就不用判断手牌了
	var cardIndex byte
	//大三元和jf只能检查一个  如果有小三元。。。。又要改
	var dsycheckcards []byte
	guiNum1 := guiNum
	static.HF_DeepCopy(&dsycheckcards, &checkcards)
	//检查大三元
	for i = 0; i < 3; i++ {
		if dsyMask&(1<<i) == 0 {
			cardIndex, _ = mahlib2.CardToIndex(0x37 - i)
			if ok, _ := RepairKeCards(dsycheckcards, cardIndex, 3, &guiNum1); ok {
				//可以补
				dsyMask |= (0x1 << i)
			}
		}
	}
	if dsyMask^0x7 == 0 {
		//去牌检查，因为有可能胡的碰碰胡，3个白板中的一个当赖子用了
		//dsycheckcards[31]=0
		//dsycheckcards[32]=0
		//dsycheckcards[33]=0
		hu, _ := this.CheckHU.Split_Byte(dsycheckcards, guiNum1, 1, true)
		if hu {
			ChiHuResult.ChiHuKind2 |= static.CHR_SAN_DA_JIANG
			return dsyMask
		} else {
			return 0
		}
	}
	//大三元和季风不可能共存
	//不是大三元，就是季风，这里有个问题，老三番没有4季风，那么优先检查有牌的那个，因为赖子最多4章，最多也就只能补充一个位置，4季风需要至少3个成型，3季风至少需要2个成型
	result := this.AnalyseJiFeng(&jfMask, checkcards, &guiNum, checkJFnum, 2)
	if !result {
		return 0
	}
	fmt.Println(fmt.Sprintf("季风mask（%b）", jfMask))
	//去风查胡
	for i = 0; i < 4; i++ {
		if jfMask&(1<<i) != 0 {
			checkcards[30-i] = 0
		}
	}
	hu, _ := this.CheckHU.Split_Byte(checkcards, guiNum, 1, true)
	if hu {
		ChiHuResult.ChiHuKind2 |= static.CHR_SI_JI_FENG
		return jfMask
	} else {
		return 0
	}

}
