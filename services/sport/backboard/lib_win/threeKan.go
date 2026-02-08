package lib_win

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*
姐妹铺
预想包含双姐妹铺
去姐妹铺的条件是2,2,2
即是顺子，又要是2个，但是赖子目前就只有4个可用
*/
type SanKanObj struct {
	handCards  []byte //去掉三坎后的数据
	index      byte   //三坎的下标
	lostCards  byte   //姐妹铺去牌数据 如果还原的话，用这个参考加回去，这个和3坎不一样，有可能是赖子补的
	useGodNum  byte
	lastGodNum byte //剩下鬼牌
}

/*
最多2组
*/
type SanKanResult struct {
	handCards  []byte //去掉三坎的数据
	index      []byte //三坎下标
	lostCards  []byte //去掉的三坎数据
	lastGodNum byte   //剩下鬼牌
	//isDouble bool  //是不是  3坎一定是去掉3个
	huKind byte //硬胡 258标志
}

//eyeMask带表是不是258 等于2就是
func (this *CheckHu) CreateSanKan_base_ex1(handCards []byte, godNum *byte, index byte, level int, eyeMask byte, tempObj *[]SanKanObj, resultObj *[]SanKanResult, godCards []byte, seaf bool) (result byte) {
	//先看牌够不够
	if level == 0 {
		//fmt.Println(fmt.Sprintf("%v",*tempObj))
		//这里checkhu一下
		losecheckcards := (*tempObj)[len(*tempObj)-1].handCards
		losegodNum := (*tempObj)[len(*tempObj)-1].lastGodNum
		//重组
		if len(godCards) != 0 {
			var err error
			losecheckcards, _, losegodNum, err = ReSetHandCards_Nomal(losecheckcards, static.INVALID_BYTE, true, godCards, seaf)
			if err != nil {
				fmt.Println(fmt.Sprintf("重组坎牌检查err（%v）", err))
				*tempObj = nil
				return
			}
			(*tempObj)[len(*tempObj)-1].handCards = losecheckcards
			(*tempObj)[len(*tempObj)-1].lastGodNum = losegodNum
		}
		hu, _ := this.CheckHU.Split_Byte(losecheckcards, losegodNum, eyeMask, true)
		if hu {
			//如果胡了，创建一下
			newresule := this.CreateSanKanResult(tempObj, eyeMask)
			*resultObj = append(*resultObj, *newresule)
		} else {
			//fmt.Println(fmt.Sprintf("最后的牌（%v）剩下鬼牌（%d）不能胡",losecheckcards,losegodNum))
			*tempObj = nil
		}
		return
	} else {
		var lostcard byte = 0
		var i byte
		var needgod byte = 0
		for i = index; i < byte(len(handCards)); i++ {
			//fmt.Println(fmt.Sprintf("记录(%d)（%d）",i,handCards[i]+*godNum))
			result = 0
			if (handCards[i] + *godNum) < 3 {
				//不能销毁
				continue
			}
			if handCards[i]/3 > 0 {
				handCards[i] -= 3
				lostcard = 3
				needgod = 0
			} else {
				//不够求余，因为有为0的情况
				needgod = 3 - handCards[i]
				lostcard = handCards[i]
				handCards[i] = 0
			}
			*godNum -= needgod
			checkhardcards := make([]byte, 0)
			static.HF_DeepCopy(&checkhardcards, &handCards)
			newobj := &SanKanObj{
				handCards:  checkhardcards, //去掉坎后的数据
				index:      i,              //坎的下标
				lostCards:  lostcard,       //去掉的坎的数量
				useGodNum:  needgod,
				lastGodNum: *godNum, //剩下鬼牌
			}
			result = 1
			//*resultObj=append(*resultObj,*newobj)
			//如果是3个了，就记录一下，并且还原
			if level > 0 {
				//追加
				*tempObj = append(*tempObj, *newobj)
				level -= 1
				i += 1
				//需要的赖子数就是3-里面的每个个数
				result = this.CreateSanKan_base_ex1(handCards, godNum, i, level, eyeMask, tempObj, resultObj, godCards, seaf)
				level += 1
				i -= 1
			}
			//还原
			//fmt.Println(fmt.Sprintf("level(%d)",level))
			if len(*tempObj) != 0 {
				i = (*tempObj)[len(*tempObj)-1].index
				handCards[i] += (*tempObj)[len(*tempObj)-1].lostCards
				*godNum += (*tempObj)[len(*tempObj)-1].useGodNum
				(*tempObj) = (*tempObj)[:len(*tempObj)-1]
			} else {
				*godNum += needgod
				handCards[i] += lostcard
			}
		}
	}
	return
}

////构建姐妹铺结果
/*
整合所有结果来创建去掉姐妹铺的牌
//所有可能去拍的总和是2，不可能超过
如果单项就已经满足2的话，另外两色就不用处理了
关键是构造结果

*/
func (this *CheckHu) CreateSanKanResult(tempObj *[]SanKanObj, eyeMask byte) (result *SanKanResult) {
	if len(*tempObj) == 0 {
		return nil
	}
	losehandCards := make([]byte, 0)
	static.HF_DeepCopy(&losehandCards, &(*tempObj)[len(*tempObj)-1].handCards)
	//fmt.Println(fmt.Sprintf("copy最后的手牌(%v)",losehandCards))
	loseindex := make([]byte, 0)
	lostCards := make([]byte, 0)
	for _, v := range *tempObj {
		loseindex = append(loseindex, v.index)
		lostCards = append(lostCards, v.lostCards)
	}
	result = &SanKanResult{
		handCards:  losehandCards,                          //去掉三坎的数据
		index:      loseindex,                              //三坎下标
		lostCards:  lostCards,                              //去掉的三坎数据
		lastGodNum: (*tempObj)[len(*tempObj)-1].lastGodNum, //剩下鬼牌
		//isDouble bool  //是不是  3坎一定是去掉3个

	}
	return
}

//一种花色里面可能有多个姐妹铺的牌型 单色 改成2个的情况试试
func (this *CheckHu) CheckSanKan_base_ex(handCards []byte, index byte, godNum *byte, leve int, eyemask byte, recordObj *SanKanObj, result *[]SanKanResult, godCards []byte, seaf bool) (resultObj []SanKanObj) {
	checkgodNum := *godNum
	checkhardcards := make([]byte, 0)
	static.HF_DeepCopy(&checkhardcards, &handCards)
	//进构造单元
	var lostcards []SanKanObj
	this.CreateSanKan_base_ex1(checkhardcards, &checkgodNum, 0, leve, eyemask, &lostcards, result, godCards, seaf)
	return lostcards
}

func (this *CheckHu) CheckSanKan_hu(checkCardObj *JieMeiPuResult, gui_num byte, eyeMask byte) byte {

	hu, _ := this.CheckHU.Split_Byte(checkCardObj.handCards, checkCardObj.lastGodNum, eyeMask, true)
	if hu {
		//不分类了，根据
		if (checkCardObj.lastGodNum == gui_num && gui_num != 0) || gui_num == 0 {
			if eyeMask > 1 {
				checkCardObj.huKind |= static.CHK_PING_HU_NOMAGIC << 2
			} else {
				checkCardObj.huKind |= static.CHK_PING_HU_NOMAGIC
			}
		} else {
			if eyeMask > 1 {
				checkCardObj.huKind |= static.CHK_PING_HU_MAGIC << 2
			} else {
				checkCardObj.huKind |= static.CHK_PING_HU_MAGIC
			}
		}
		return static.WIK_CHI_HU
	}
	return static.CHK_NULL
}
func (this *CheckHu) CheckKan(handCards []byte, weaveItem []static.TagWeaveItem, cbCurrentCard byte, isNormalCard bool, godCards []byte, kanlevel int, eyeMask byte, usegod bool, seaf bool) bool {
	Obj, err := this.CheckSanKan_Normal(handCards, weaveItem, cbCurrentCard, isNormalCard, godCards, kanlevel, eyeMask, usegod, seaf)
	if err != nil {
		fmt.Println(fmt.Sprintf("检查坎出问题（%v）", err))
		return false
	}
	//for _,v:=range Obj{
	//	fmt.Println(fmt.Sprintf("最后的结果（%v）",v))
	//}
	return len(Obj) > 0
}

//普通接口 姐妹铺
/*
想法就是，一色牌至少要有9张以上，然后就是顺序来排下去
*/
func (this *CheckHu) CheckSanKan_Normal(handCards []byte, weaveItem []static.TagWeaveItem, cbCurrentCard byte, isNormalCard bool, godCards []byte, kanlevel int, eyeMask byte, usegod bool, seaf bool) (checkCardObj []SanKanResult, err error) {
	//检查weaveItem
	tripletNum := 0
	weaveinfo := card_mgr.GetWarninfo(weaveItem)
	if weaveinfo != nil {
		if len(weaveinfo.ChipaiInfo) > 0 {
			return nil, nil
		}
		tripletNum += len(weaveinfo.Triplet) + len(weaveinfo.HidTriplet)
	}
	level := kanlevel - tripletNum
	checkCardObj = make([]SanKanResult, 0)
	if level == 0 {
		//已经够了
		newstruct := SanKanResult{
			handCards: handCards,
		}
		checkCardObj = append(checkCardObj, newstruct)
		return checkCardObj, nil
	}
	checkGodCards := []byte{}
	if len(godCards) != 0 {
		static.HF_DeepCopy(&checkGodCards, &godCards)
	}
	if !usegod {
		checkGodCards = []byte{}
	}
	checkcards, _, guiNum, err := ReSetHandCards_Nomal(handCards, cbCurrentCard, isNormalCard, checkGodCards, seaf)
	if !usegod {
		this.CheckSanKan_base_ex(checkcards, 0, &guiNum, level, eyeMask, nil, &checkCardObj, godCards, seaf)
	} else {
		this.CheckSanKan_base_ex(checkcards, 0, &guiNum, level, eyeMask, nil, &checkCardObj, []byte{}, seaf)
	}

	//this.CreateSanKan_base_ex1(checkCards,&gui_num,0,level,&itemObj,result)
	//-----------------------------------------------------------------
	return
}
