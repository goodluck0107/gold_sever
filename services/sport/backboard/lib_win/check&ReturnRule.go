package lib_win

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	"github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

//20181224 苏大强 这个单元还有个地方没问题完全想好，就是将牌的判断
/*
20190116 考虑从这里生成几种可能的情况
无番型判断的，赖将就不管了
有翻型的还要判断赖将的可能去向（屁胡的清一色，比硬胡）
*/
type CheckHu struct {
	guicards []byte //赖子，最多应该只有2，这里还是切片吧
	//20190107 基本上就是258和混将，那么这个地方 258=2 混将=1
	eyecards byte   //将牌
	CheckHU  *HuLib //判断是不是胡牌

}

func NewCheckHu(needtable bool) (*CheckHu, error) {
	var err error = nil
	newcheck := &CheckHu{
		guicards: make([]byte, 0),
		eyecards: 1,
	}
	newcheck.CheckHU, err = NewHU(needtable)
	return newcheck, err
}

func (this *CheckHu) AnalyseHuMask(checkCtx *card_mgr.CheckHuCTX, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult, mask258 byte) (byte, byte) {
	if checkCtx == nil {
		return static.CHK_NULL, 0
	}
	hardcard := make([]byte, 0)
	static.HF_DeepCopy(&hardcard, &checkCtx.CheckCtx.CbCardIndex)
	prepareHu := false
	//常规的带赖子不带赖子判断
	// nomalHu := true
	//var hu_struct *[4][2]byte
	//20190104 查硬胡先
	var check258 byte = 1
	if mask258 != 0 {
		check258 = 2
	}
	if checkCtx.CheckCardEX != nil {
		prepareHu, checkCtx.Hu_struct = this.CheckHU.AnalyseHuRestore(hardcard, checkCtx.CheckCardEX.BaseInfo, check258)
	} else {
		prepareHu, checkCtx.Hu_struct = this.CheckHU.AnalyseHuRestore(hardcard, nil, check258)
	}

	if prepareHu {
		//正真的硬胡
		// nomalHu = true
		checkCtx.ChiHuKind |= static.CHK_PING_HU_NOMAGIC
		if ChiHuResult != nil {
			ChiHuResult.ChiHuRight |= wChiHuRight
			//20190114 赖子对当258的时候，是软的，判断手牌中如果没有258的对，那么就必然是软的
			ChiHuResult.ChiHuKind |= checkCtx.ChiHuKind
		}
		if check258 == 2 {
			//258 同步置位
			checkCtx.ChiHuKind |= static.CHK_PING_HU_NOMAGIC << 2
		}
		//20190221 为了通用去掉以前的检查，返回所有可能的现象，如果没有god对象，那么就只返回硬胡情况
		if checkCtx.CheckGodOrg == nil {
			return static.WIK_CHI_HU, 0
		}
	}
	//赖子胡 有条件的
	if checkCtx.CheckGodOrg != nil && checkCtx.CheckGodOrg.GodNum != 0 {
		if checkCtx.CheckCardEX != nil {
			prepareHu, checkCtx.GuiHu_struct = this.CheckHU.AnalyseCard(hardcard, checkCtx.CheckCardEX.ExInfo, checkCtx.GodInfo.GetGuiInfo(), check258)
		} else {
			prepareHu, checkCtx.GuiHu_struct = this.CheckHU.AnalyseCard(hardcard, nil, checkCtx.GodInfo.GetGuiInfo(), check258)

		}
		if prepareHu {
			//20180301 硬胡没必要知道需要最少需要多少赖子
			//_, Needgui := this.RestoreNeedGui(hu_struct)
			// if checkCtx.CheckGodOrg.GodNum < Needgui {
			// 	return public.WIK_CHI_HU, 0
			// }
			if ChiHuResult != nil {
				ChiHuResult.ChiHuRight |= wChiHuRight
				//20181224 检查gui的数据
				// fmt.Println(fmt.Sprintf("god 数（%d）", checkCtx.CheckGodOrg.GodNum))
				//if Needgui == 0 && !is258 {
				// if Needgui == 0 && !checkCtx.CheckMFan && !is258 {
				// 	ChiHuResult.ChiHuKind |= public.CHK_PING_HU_NOMAGIC
				// } else {
				ChiHuResult.ChiHuKind |= static.CHK_PING_HU_MAGIC
			}
			checkCtx.ChiHuKind |= static.CHK_PING_HU_MAGIC
			//20190301 考虑到有这样的情况，234 4是赖子，也可以当123使，所以这个地方暂时去掉吧
			//if checkCtx.ChiHuKind^(public.CHK_PING_HU_NOMAGIC|public.CHK_PING_HU_MAGIC) == 0 && checkCtx.CheckGodOrg.GodNum == 1 {
			//	checkCtx.ChiHuKind ^= public.CHK_PING_HU_MAGIC
			//}
			//if checkCtx.ChiHuKind&public.CHK_PING_HU_MAGIC!=0&&is258{
			if check258 == 2 {
				//置位258
				checkCtx.ChiHuKind |= static.CHK_PING_HU_MAGIC << 2
			}
			//20190222 需要gui的数量不一定等于手牌中god数
			//if Needgui != checkCtx.CheckGodOrg.GodNum {
			//	Needgui = checkCtx.CheckGodOrg.GodNum
			//}
			return static.WIK_CHI_HU, checkCtx.CheckGodOrg.GodNum
		}
	}
	//---------------------------------------------------------------------------
	return static.CHK_NULL, 0
}

//包含258的全情况的，这个hu可以用了吧
/*
20190115 258mask，如果是0代表混将，剩下的根据mask来判断
20190115 修改，参数就是checkctx
*/
func (this *CheckHu) AnalyseHu(checkCtx *card_mgr.CheckHuCTX, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult) byte {
	//如果是混将，按照以前的调用即可
	result, _ := this.AnalyseHuMask(checkCtx, wChiHuRight, ChiHuResult, checkCtx.Mask258)
	/*
		//20190122 见字胡查询，设定见字胡查询条件 ，判胡的牌只会出现在发牌区、桌面区、倒牌区；胡牌检查的牌来源不会是手牌区
		只有发牌区、倒牌区、桌面区,考虑到兼容 如果设了card_mgr.ORIGIN_TABILE 就说明对见字胡有要求
	*/
	//fmt.Println(fmt.Sprintf("result(%b),(%t)",result,checkCtx.GodHuMask^card_mgr.ORIGIN_HAND!=0))
	//这里如果要兼容的话，如果没设置见字胡的检查，就不进去最好
	//if result!=public.CHK_NULL&&checkCtx.GodHuMask!=0&&(checkCtx.GodHuMask^card_mgr.ORIGIN_HAND!=0){
	//fmt.Println(fmt.Sprintf("普通见字胡mask（%d）",checkCtx.GodHuMask^card_mgr.ORIGIN_HAND))
	if result != static.CHK_NULL && (checkCtx.GodHuMask^card_mgr.ORIGIN_HAND != 0) {
		//GodHuMask的桌面区位==0，那么就要检查见字胡
		if this.CheckGodHu(checkCtx) {
			//检查
			//fmt.Println(fmt.Sprintf("见字胡检查（%b）(%b)",checkCtx.CheckCardEX.BaseInfo.Origin,checkCtx.GodHuMask))
			if checkCtx.GodHuMask != 0 && uint64(checkCtx.CheckCardEX.BaseInfo.Origin)&checkCtx.GodHuMask == 0 {
				ChiHuResult.ChiHuKind = static.CHK_NULL
				checkCtx.ChiHuKind = static.CHK_NULL
				return static.CHK_NULL
			} else {
				//设置见字胡
				//现在项目里面还没有见字胡的标准 我设定0x40为见字胡标志
				checkCtx.ChiHuKind |= 0x40
			}
		}
	}
	return result
}

//所有需要检查3n+2特殊牌型的胡到这里来，见字胡，
/*
20190402 目前没加甩字胡
*/
func (this *CheckHu) AnalyseHuEx(checkCtx *card_mgr.CheckHuCTX, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult) byte {
	//如果是混将，按照以前的调用即可
	result, _ := this.AnalyseHuMask(checkCtx, wChiHuRight, ChiHuResult, checkCtx.Mask258)
	/*
		//20190122 见字胡查询，设定见字胡查询条件 ，判胡的牌只会出现在发牌区、桌面区、倒牌区；胡牌检查的牌来源不会是手牌区
		只有发牌区、倒牌区、桌面区,考虑到兼容 如果设了card_mgr.ORIGIN_TABILE 就说明对见字胡有要求
	*/
	//fmt.Println(fmt.Sprintf("result(%b),(%t)",result,checkCtx.GodHuMask^card_mgr.ORIGIN_HAND!=0))
	//这里如果要兼容的话，如果没设置见字胡的检查，就不进去最好
	//if result!=public.CHK_NULL&&checkCtx.GodHuMask!=0&&(checkCtx.GodHuMask^card_mgr.ORIGIN_HAND!=0){
	if result != static.CHK_NULL && (checkCtx.GodHuMask^card_mgr.ORIGIN_HAND != 0) {
		//GodHuMask的桌面区位==0，那么就要检查见字胡
		if this.CheckGodHu(checkCtx) {
			//检查
			//fmt.Println(fmt.Sprintf("见字胡检查（%b）(%b)",checkCtx.CheckCardEX.BaseInfo.Origin,checkCtx.GodHuMask))
			if checkCtx.GodHuMask != 0 && uint64(checkCtx.CheckCardEX.BaseInfo.Origin)&checkCtx.GodHuMask == 0 {
				ChiHuResult.ChiHuKind = static.CHK_NULL
				checkCtx.ChiHuKind = static.CHK_NULL
				return static.CHK_NULL
			} else {
				//设置见字胡
				//现在项目里面还没有见字胡的标准 我设定0x40为见字胡标志
				checkCtx.ChiHuKind |= 0x40
			}
		}
	}
	return result
}

//-------------------------------------20190402-------------------------

/*
AnalyseHu_normal
用来检查普通的3n+2（258和乱将）
需要的参数，手牌（最多13张）
判断的牌（1张）
判断的牌是不是普通的牌：如果不是普通，那就是赖子（条件是这章牌可以用来查胡，至于见字胡还是甩字胡等等，再说）
god牌值 用切片，目前支持2个赖子，3个赖子不太可能了;如果是nil或者手牌和检查的牌都不是，那就是硬胡检查
258maks：如果是1代表乱将，2的话目前就当是258，以后再说别的情况
注：查表判胡不用倒牌数据
*/
func (this *CheckHu) AnalyseHu_Normal(cbCardIndex []byte, checkCard byte, isNormalCard bool, guiCards []byte, mask258 byte) (result byte, err error) {
	//安全检查 手牌数目符合（3n+1）
	_, err = mahlib2.CheckHandCardsSafe(cbCardIndex)
	if err != nil {
		return static.CHK_NULL, err
	}
	if checkCard > 0x37 {
		return static.CHK_NULL, errors.New(fmt.Sprintf("AnalyseHu_Normal 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//深copy一次，其实我一直没找到这个地方不深copy的原因
	hardcard := make([]byte, 0)
	static.HF_DeepCopy(&hardcard, &cbCardIndex)
	hu := false
	//首先就检查一下能不能还原胡
	hu, _ = this.CheckHU.GetHuInfo_Byte(hardcard, checkCard, true, static.INVALID_BYTE, static.INVALID_BYTE, mask258)
	if hu {
		result |= static.CHK_PING_HU_NOMAGIC
		if mask258 > 1 {
			result |= static.CHK_PING_HU_NOMAGIC << 2
		}
	}
	var hu_struct *[4][2]byte
	switch len(guiCards) {
	case 1:
		hu, hu_struct = this.CheckHU.GetHuInfo_Byte(hardcard, checkCard, isNormalCard, guiCards[0], static.INVALID_BYTE, mask258)
	case 2:
		hu, hu_struct = this.CheckHU.GetHuInfo_Byte(hardcard, checkCard, isNormalCard, guiCards[0], guiCards[1], mask258)
	default:
		return
	}
	if hu {
		_, needgui := card_mgr.RestoreNeedGui(hu_struct)
		if needgui > 0 {
			result |= static.CHK_PING_HU_MAGIC
			if mask258 > 1 {
				result |= static.CHK_PING_HU_MAGIC << 2
			}
		}
		if hu, err = this.CheckGodHu_Normal(hardcard, guiCards, mask258); err != nil {
			return result, err
		}
		if hu {
			fmt.Println("普通见字胡")
			result |= 0x40
		}
	}
	//检查见字胡
	return
}

// //查胡，返回所有结构 这个方法还未完成，目前放放不需要，先放下
// func (this *CheckHu) ReWeaveItem() [][]TagWeaveItem {
// 	//手牌+倒牌 无判断的那张
// 	checkcard, err := this.userHandInfo.ReadyForCheckHu(true)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}
// 	//判胡
// 	if this.AnalyseOneHu(checkcard, nil, 0, nil) == public.CHK_NULL {
// 		return nil
// 	}
// 	checkcard, err = this.userHandInfo.ReadyForZygosity()
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}
// 	//去赖子
// 	err = subCardBatch(checkcard, this.guicards, true)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}

// 	//配型 第二个参数已经没用了 ，还是要从头玩到尾
// 	//遇到带赖子的，就要用赖子来配型
// 	guiNum := this.getHandguinum(this.userHandInfo.cbCardIndex,)
// 	result := this.reAllItem(checkcard, guiNum)
// 	if len(result) == 0 {
// 		errors.New("创建出错")
// 		return nil
// 	}

// 	//-------------------------------
// 	// array := this.createItemArray(result)
// 	// if len(array) == 0 {
// 	// 	errors.New("创建返回出错")
// 	// 	return nil

// 	// }
// 	// fmt.Println(array)
// 	return result
// }

// // 去除顺子,目前就来一遍，把数据拿回来
// func (this *CheckHu) removeThreeLink(cards []byte, index int) []TagWeaveItem {
// 	// marks := make([]int, 0, 4)
// 	var marks []TagWeaveItem
// 	for i := 0; i < len(cards)-2; i++ {
// 		// if i == 1 {
// 		// 	fmt.Println(i)
// 		// }
// 		if cards[i] > 0 && cards[i+1] > 0 && cards[i+2] > 0 {
// 			cards[i] -= 1
// 			cards[i+1] -= 1
// 			cards[i+2] -= 1
// 			//默认返回都是左吃，因为下标是在左,
// 			item := new(TagWeaveItem)
// 			item.WeaveKind = public.WIK_LEFT
// 			item.CenterCard = CombinCard(byte(index), byte(i+1))
// 			marks = append(marks, *item)
// 			i--
// 		}
// 	}
// 	return marks
// }

// //去刻
// func (this *CheckHu) removeSameThree(cards []byte, index int) []TagWeaveItem {
// 	var marks []TagWeaveItem
// 	for i := 0; i < len(cards); i++ {
// 		if cards[i] >= 3 {
// 			cards[i] -= 3
// 			item := new(TagWeaveItem)
// 			item.WeaveKind = public.WIK_PENG
// 			item.CenterCard = CombinCard(byte(index), byte(i+1))
// 			marks = append(marks, *item)
// 		}
// 	}
// 	return marks
// }

// //只去一个眼 乱将
// func (this *CheckHu) removeeye(cards []byte, index int) byte {
// 	// var marks []byte
// 	for i := 0; i < len(cards); i++ {
// 		if cards[i] >= 2 {
// 			cards[i] -= 2
// 			CenterCard := CombinCard(byte(index), byte(i+1))
// 			return CenterCard
// 			// marks = append(marks, CenterCard)
// 		}
// 	}
// 	return 0
// }

// 检测剩下的元素是否全为0
func (this *CheckHu) checkZero(cards []byte) bool {
	for i := 0; i < len(cards); i++ {
		if cards[i] != 0 {
			return false
		}
	}
	return true
}

// //检查判型,这是单一色的 要把index 传进来
// func (this *CheckHu) checkGuiRuleAll(cards []byte, index int, GuiCount byte, chi bool, eye1 byte) (tar [][]TagWeaveItem, eye byte, lguicount byte) {
// 	checkcards := make([]byte, len(cards))
// 	copy(checkcards, cards)
// 	// var tar [][]TagWeaveItem
// 	var markitem []TagWeaveItem
// 	//先扔掉可用的克
// 	lguicount = GuiCount
// 	TagSameThreeItem := this.removeSameThree(checkcards, index)
// 	if len(TagSameThreeItem) != 0 {
// 		markitem = append(markitem, TagSameThreeItem...)
// 	}
// 	//再扔掉顺子
// 	TagThreeLinkItem := this.removeThreeLink(checkcards, index)
// 	if len(TagThreeLinkItem) != 0 {
// 		markitem = append(markitem, TagThreeLinkItem...)
// 	}
// 	//去眼
// 	if eye1 == 0 {
// 		eye1 = this.removeeye(checkcards, index)
// 	}
// 	eye = eye1
// 	if GuiCount != 0 {
// 		//剩下的牌就要配型了.

// 		for i := 0; i < len(checkcards); i++ {
// 			if checkcards[i] > 0 {
// 				//凑对 位上只有1或2
// 				tempMahjong := make([]byte, len(checkcards))
// 				copy(tempMahjong, checkcards)
// 				// tempMahjong[i] = 0
// 				needGuiCount, marks2 := this.getNeedGuiCount(tempMahjong, index, eye, chi)
// 				if needGuiCount <= GuiCount {
// 					if len(marks2) != 0 {
// 						markitem = append(markitem, marks2...)
// 						tar = append(tar, markitem)
// 						continue
// 					}
// 					lguicount = GuiCount - needGuiCount
// 				}
// 				// needGuiCount, marks3, eye := this.getNeedGuiCount(tempMahjong, index, eye, chi)
// 				// if needGuiCount <= GuiCount {
// 				// 	if len(marks3) != 0 {
// 				// 		markitem = append(markitem, marks3...)
// 				// 		marks = append(marks, markitem)
// 				// 	}

// 				// }
// 			}
// 		}
// 		//check位没有数据 填两个当对子
// 		// needGuiCount, marks3 := this.getNeedGuiCount(checkcards, index, 2, chi)
// 		// if needGuiCount <= GuiCount {
// 		// 	if len(marks3) != 0 {
// 		// 		markitem = append(markitem, marks3...)
// 		// 		marks = append(marks, markitem)
// 		// 	}

// 		// }

// 	} else {
// 		if len(markitem) != 0 {
// 			tar = append(tar, markitem)
// 		}

// 	}
// 	return
// }

//返回可能的情况
//chi 吃，代表如果是风字牌不检查吃
// func (this *CheckHu) getNeedGuiCount(cards []byte, index int, eye byte, chi bool) (byte, []TagWeaveItem) {
// 	var minGuiCount byte = 0
// 	var marks []TagWeaveItem
// 	if !this.checkZero(cards) {
// 		num := len(cards)
// 		for j := 0; j < num; j++ {
// 			if cards[j] <= 0 {
// 				continue
// 			}
// 			if chi {
// 				//这有问题，左吃还是右吃呢
// 				if j < 8 {
// 					switch j {
// 					case 0:
// 						if cards[j+1] > 0 {
// 							cards[j]--
// 							cards[j+1]--
// 							item := new(TagWeaveItem)
// 							item.CenterCard = CombinCard(byte(index), byte(j+3))
// 							item.WeaveKind = public.WIK_RIGHT
// 							marks = append(marks, *item)
// 							j--
// 							minGuiCount++
// 							continue
// 						}
// 					case 7:
// 						if cards[j+1] > 0 {
// 							cards[j]--
// 							cards[j+1]--

// 							item := new(TagWeaveItem)
// 							item.CenterCard = CombinCard(byte(index), byte(j-1))
// 							item.WeaveKind = public.WIK_LEFT
// 							marks = append(marks, *item)
// 							j--
// 							minGuiCount++
// 							continue
// 						}
// 					default:
// 						if cards[j+1] > 0 {
// 							cards[j]--
// 							cards[j+1]--
// 							item := new(TagWeaveItem)
// 							//就以它为主
// 							item.CenterCard = CombinCard(byte(index), byte(j+1))
// 							item.WeaveKind = public.WIK_LEFT | public.WIK_RIGHT
// 							marks = append(marks, *item)
// 							j--
// 							minGuiCount++
// 							continue
// 						}
// 					}
// 				}
// 				if j < 7 {
// 					if cards[j+2] > 0 {
// 						cards[j]--
// 						cards[j+2]--
// 						j--
// 						item := new(TagWeaveItem)
// 						item.CenterCard = CombinCard(byte(index), byte(j+1))
// 						item.WeaveKind = public.WIK_CENTER
// 						marks = append(marks, *item)
// 						minGuiCount++
// 						continue
// 					}
// 				}

// 			}
// 			if cards[j] == 1 {
// 				cards[j]--
// 				minGuiCount += 2
// 				item := new(TagWeaveItem)
// 				item.WeaveKind = public.WIK_PENG
// 				item.CenterCard = CombinCard(byte(index), byte(j+1))
// 				marks = append(marks, *item)

// 				continue
// 			}
// 			if cards[j] == 2 {
// 				cards[j] -= 2
// 				minGuiCount += 1
// 				item := new(TagWeaveItem)
// 				item.WeaveKind = public.WIK_PENG
// 				item.CenterCard = CombinCard(byte(index), byte(j+1))
// 				marks = append(marks, *item)
// 			}

// 		}
// 	}
// 	return minGuiCount, marks
// }

//局限，将牌必须是万条筒
//去对，最多是7对。最少一对
//将所有可能都挖出来
// func (this *CheckHu) RemoveTwoCards(cards []byte, specialcards []byte, max_eye int) ([][]byte, error) {
// 	var eyemap [][]byte
// 	tempcard := make([]byte, len(cards))
// 	copy(tempcard, cards)
// 	for i := 0; i < len(cards); i++ {
// 		if cards[i] >= 2 {
// 			if len(this.eyecards) == 0 {
// 				tempcard[i] -= 2
// 				max_eye--
// 				eyemap = append(eyemap, tempcard)
// 			} else {
// 				if this.eyeisNotFeng && i < 27 {
// 					for _, v := range this.eyecards {
// 						if v == byte(i+1) {
// 							if v >= byte(len(cards)) {
// 								return nil, errors.New(fmt.Sprintf("将牌（%x）越界（%v）长度（%d）", v, cards, len(cards)))
// 							}
// 							tempcard[i] -= 2
// 							max_eye--
// 							eyemap = append(eyemap, tempcard)
// 							break
// 						}
// 					}
// 				}
// 			}
// 		}
// 		if max_eye == 0 {
// 			break
// 		}
// 	}
// 	return eyemap, nil
// }

//局限，将牌必须是万条筒
// 通过去除麻将矩阵中一个将之后的所有型 （将是万饼筒 专用）
//原因，将牌是哪一行的，目前我们俗称258，代表万饼筒的，这里直接通过开关屏蔽掉风牌的去258
// func (this *CheckHu) createListByRemoveEye(cards []byte, maxeyeNum int) ([][]byte, error) {
// 	if maxeyeNum > MAX_COUNT/2 {
// 		return nil, errors.New(fmt.Sprintf("手牌总数（%d）最多有（%d）对", MAX_INDEX, MAX_COUNT/2))
// 	}
// 	// var newList [][]byte
// 	//最多不会超过7
// 	checkcards, err := this.RemoveTwoCards(cards[:], this.eyecards, maxeyeNum)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// newList = append(newList, checkcards...)
// 	return checkcards, nil
// }

/*
经过测试参数2实际意义已经完蛋了
*/
// func (this *CheckHu) reAllItem(cards []byte, guiNum byte) [][]TagWeaveItem {
// 	//先去将，把去将后的牌型都拿出来
// 	//普通胡牌牌型3n+2 胡了必然有一对将，考虑赖子补将的可能（单吊的情况）
// 	var tagWeaveItem [][]TagWeaveItem
// 	var TagItem [][]TagWeaveItem
// 	//扔掉指定数量的对子，胡牌的话必然是一对将
// 	newList, err := this.createListByRemoveEye(cards, 1)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}
// 	for _, newcard := range newList {
// 		fmt.Println(newcard)
// 	}
// 	min := 0
// 	max := 0
// 	var eye byte = 0
// 	nextgui := guiNum
// 	for i := 0; i < 4; i++ {
// 		if i < 3 {
// 			min = i * 9
// 			max = min + 9
// 		} else {
// 			min = i * 9
// 			max = min + 7
// 		}
// 		TagItem, eye, nextgui = this.checkGuiRuleAll(cards[min:max], i, nextgui, i < 3, eye)
// 		tagWeaveItem = append(tagWeaveItem, TagItem...)
// 		// }
// 	}
// 	if eye != 0 {
// 		for _, v := range tagWeaveItem {
// 			if len(v) != 4-len(this.userHandInfo.weaveItem) {
// 				fmt.Println("配型失败")
// 				continue
// 				// } else {
// 				// 	fmt.Println(tagWeaveItem[i])
// 			}
// 		}
// 		return tagWeaveItem
// 	}
// 	//必须是3n+2=14 所以n必须是4了
// 	//创建
// 	return nil
// }
