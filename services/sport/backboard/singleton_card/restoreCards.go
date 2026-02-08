package card_mgr

import (
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

//20190228  苏大强 这个单元做还原处理的
/*
20190116 考虑从这里生成几种可能的情况
无番型判断的，赖将就不管了
有翻型的还要判断赖将的可能去向（屁胡的清一色，比硬胡）
*/

//查胡，返回所有结构 这个方法还未完成，目前放放不需要，先放下
//func (this *CheckHu) ReWeaveItem() [][]TagWeaveItem {
//	//手牌+倒牌 无判断的那张
//	checkcard, err := this.userHandInfo.ReadyForCheckHu(true)
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	//判胡
//	if this.AnalyseOneHu(checkcard, nil, 0, nil) == public.CHK_NULL {
//		return nil
//	}
//	checkcard, err = this.userHandInfo.ReadyForZygosity()
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	//去赖子
//	err = subCardBatch(checkcard, this.guicards, true)
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//
//	//配型 第二个参数已经没用了 ，还是要从头玩到尾
//	//遇到带赖子的，就要用赖子来配型
//	guiNum := this.getHandguinum(this.userHandInfo.cbCardIndex,)
//	result := this.reAllItem(checkcard, guiNum)
//	if len(result) == 0 {
//		errors.New("创建出错")
//		return nil
//	}
//
//	//-------------------------------
//	// array := this.createItemArray(result)
//	// if len(array) == 0 {
//	// 	errors.New("创建返回出错")
//	// 	return nil
//
//	// }
//	// fmt.Println(array)
//	return result
//}

// 去除顺子,目前就来一遍，把数据拿回来
/*
从AnalyseHuRestore和AnalyseCard返回的hu_struct *[4][2]byte中可以知道每种花色牌序中有多少对（不是258的也是）和最低需要多少赖子，
但是不知道是什么牌值，当一种花色需要4~8章的时候，那基本上就构成了很多可能性，那要死人了
//参数
*/

func (this *CheckHuCTX) removeThreeLink(cards []byte, index int) []*static.TagWeaveItem {
	// marks := make([]int, 0, 4)
	var tempItem []*static.TagWeaveItem
	for i := 0; i < len(cards)-2; i++ {
		// if i == 1 {
		// 	fmt.Println(i)
		// }
		if cards[i] > 0 && cards[i+1] > 0 && cards[i+2] > 0 {
			cards[i] -= 1
			cards[i+1] -= 1
			cards[i+2] -= 1
			//默认返回都是左吃，因为下标是在左,
			item := new(static.TagWeaveItem)
			item.WeaveKind = static.WIK_LEFT
			item.CenterCard = mahlib2.CombinCard(byte(index), byte(i+1))
			tempItem = append(tempItem, item)
			i--
		}
	}
	return tempItem
}

//去刻
func (this *CheckHuCTX) removeSameThree(cards []byte, index int) []*static.TagWeaveItem {
	var tempItem []*static.TagWeaveItem
	for i := 0; i < len(cards); i++ {
		if cards[i] >= 3 {
			cards[i] -= 3
			item := new(static.TagWeaveItem)
			item.WeaveKind = static.WIK_PENG
			item.CenterCard = mahlib2.CombinCard(byte(index), byte(i+1))
			tempItem = append(tempItem, item)
		}
	}
	return tempItem
}

//只去一个眼 乱将
func (this *CheckHuCTX) removEye(cards []byte, index int) byte {
	// var marks []byte
	for i := 0; i < len(cards); i++ {
		if cards[i] >= 2 {
			cards[i] -= 2
			CenterCard := mahlib2.CombinCard(byte(index), byte(i+1))
			return CenterCard
			// marks = append(marks, CenterCard)
		}
	}
	return 0
}

//
////检查判型,这是单一色的 要把index 传进来
//func (this *CheckHu) checkGuiRuleAll(cards []byte, index int, GuiCount byte, chi bool, eye1 byte) (tar [][]TagWeaveItem, eye byte, lguicount byte) {
//	checkcards := make([]byte, len(cards))
//	copy(checkcards, cards)
//	// var tar [][]TagWeaveItem
//	var markitem []TagWeaveItem
//	//先扔掉可用的克
//	lguicount = GuiCount
//	TagSameThreeItem := this.removeSameThree(checkcards, index)
//	if len(TagSameThreeItem) != 0 {
//		markitem = append(markitem, TagSameThreeItem...)
//	}
//	//再扔掉顺子
//	TagThreeLinkItem := this.removeThreeLink(checkcards, index)
//	if len(TagThreeLinkItem) != 0 {
//		markitem = append(markitem, TagThreeLinkItem...)
//	}
//	//去眼
//	if eye1 == 0 {
//		eye1 = this.removeeye(checkcards, index)
//	}
//	eye = eye1
//	if GuiCount != 0 {
//		//剩下的牌就要配型了.
//
//		for i := 0; i < len(checkcards); i++ {
//			if checkcards[i] > 0 {
//				//凑对 位上只有1或2
//				tempMahjong := make([]byte, len(checkcards))
//				copy(tempMahjong, checkcards)
//				// tempMahjong[i] = 0
//				needGuiCount, marks2 := this.getNeedGuiCount(tempMahjong, index, eye, chi)
//				if needGuiCount <= GuiCount {
//					if len(marks2) != 0 {
//						markitem = append(markitem, marks2...)
//						tar = append(tar, markitem)
//						continue
//					}
//					lguicount = GuiCount - needGuiCount
//				}
//				// needGuiCount, marks3, eye := this.getNeedGuiCount(tempMahjong, index, eye, chi)
//				// if needGuiCount <= GuiCount {
//				// 	if len(marks3) != 0 {
//				// 		markitem = append(markitem, marks3...)
//				// 		marks = append(marks, markitem)
//				// 	}
//
//				// }
//			}
//		}
//		//check位没有数据 填两个当对子
//		// needGuiCount, marks3 := this.getNeedGuiCount(checkcards, index, 2, chi)
//		// if needGuiCount <= GuiCount {
//		// 	if len(marks3) != 0 {
//		// 		markitem = append(markitem, marks3...)
//		// 		marks = append(marks, markitem)
//		// 	}
//
//		// }
//
//	} else {
//		if len(markitem) != 0 {
//			tar = append(tar, markitem)
//		}
//
//	}
//	return
//}
//
////返回可能的情况
////chi 吃，代表如果是风字牌不检查吃
//func (this *CheckHu) getNeedGuiCount(cards []byte, index int, eye byte, chi bool) (byte, []TagWeaveItem) {
//	var minGuiCount byte = 0
//	var marks []TagWeaveItem
//	if !this.checkZero(cards) {
//		num := len(cards)
//		for j := 0; j < num; j++ {
//			if cards[j] <= 0 {
//				continue
//			}
//			if chi {
//				//这有问题，左吃还是右吃呢
//				if j < 8 {
//					switch j {
//					case 0:
//						if cards[j+1] > 0 {
//							cards[j]--
//							cards[j+1]--
//							item := new(TagWeaveItem)
//							item.CenterCard = CombinCard(byte(index), byte(j+3))
//							item.WeaveKind = public.WIK_RIGHT
//							marks = append(marks, *item)
//							j--
//							minGuiCount++
//							continue
//						}
//					case 7:
//						if cards[j+1] > 0 {
//							cards[j]--
//							cards[j+1]--
//
//							item := new(TagWeaveItem)
//							item.CenterCard = CombinCard(byte(index), byte(j-1))
//							item.WeaveKind = public.WIK_LEFT
//							marks = append(marks, *item)
//							j--
//							minGuiCount++
//							continue
//						}
//					default:
//						if cards[j+1] > 0 {
//							cards[j]--
//							cards[j+1]--
//							item := new(TagWeaveItem)
//							//就以它为主
//							item.CenterCard = CombinCard(byte(index), byte(j+1))
//							item.WeaveKind = public.WIK_LEFT | public.WIK_RIGHT
//							marks = append(marks, *item)
//							j--
//							minGuiCount++
//							continue
//						}
//					}
//				}
//				if j < 7 {
//					if cards[j+2] > 0 {
//						cards[j]--
//						cards[j+2]--
//						j--
//						item := new(TagWeaveItem)
//						item.CenterCard = CombinCard(byte(index), byte(j+1))
//						item.WeaveKind = public.WIK_CENTER
//						marks = append(marks, *item)
//						minGuiCount++
//						continue
//					}
//				}
//
//			}
//			if cards[j] == 1 {
//				cards[j]--
//				minGuiCount += 2
//				item := new(TagWeaveItem)
//				item.WeaveKind = public.WIK_PENG
//				item.CenterCard = CombinCard(byte(index), byte(j+1))
//				marks = append(marks, *item)
//
//				continue
//			}
//			if cards[j] == 2 {
//				cards[j] -= 2
//				minGuiCount += 1
//				item := new(TagWeaveItem)
//				item.WeaveKind = public.WIK_PENG
//				item.CenterCard = CombinCard(byte(index), byte(j+1))
//				marks = append(marks, *item)
//			}
//
//		}
//	}
//	return minGuiCount, marks
//}
//
////局限，将牌必须是万条筒
////去对，最多是7对。最少一对
////将所有可能都挖出来
//func (this *CheckHu) RemoveTwoCards(cards []byte, specialcards []byte, max_eye int) ([][]byte, error) {
//	var eyemap [][]byte
//	tempcard := make([]byte, len(cards))
//	copy(tempcard, cards)
//	for i := 0; i < len(cards); i++ {
//		if cards[i] >= 2 {
//			if len(this.eyecards) == 0 {
//				tempcard[i] -= 2
//				max_eye--
//				eyemap = append(eyemap, tempcard)
//			} else {
//				if this.eyeisNotFeng && i < 27 {
//					for _, v := range this.eyecards {
//						if v == byte(i+1) {
//							if v >= byte(len(cards)) {
//								return nil, errors.New(fmt.Sprintf("将牌（%x）越界（%v）长度（%d）", v, cards, len(cards)))
//							}
//							tempcard[i] -= 2
//							max_eye--
//							eyemap = append(eyemap, tempcard)
//							break
//						}
//					}
//				}
//			}
//		}
//		if max_eye == 0 {
//			break
//		}
//	}
//	return eyemap, nil
//}
//
////局限，将牌必须是万条筒
//// 通过去除麻将矩阵中一个将之后的所有型 （将是万饼筒 专用）
////原因，将牌是哪一行的，目前我们俗称258，代表万饼筒的，这里直接通过开关屏蔽掉风牌的去258
//func (this *CheckHu) createListByRemoveEye(cards []byte, maxeyeNum int) ([][]byte, error) {
//	if maxeyeNum > MAX_COUNT/2 {
//		return nil, errors.New(fmt.Sprintf("手牌总数（%d）最多有（%d）对", MAX_INDEX, MAX_COUNT/2))
//	}
//	// var newList [][]byte
//	//最多不会超过7
//	checkcards, err := this.RemoveTwoCards(cards[:], this.eyecards, maxeyeNum)
//	if err != nil {
//		return nil, err
//	}
//	// newList = append(newList, checkcards...)
//	return checkcards, nil
//}
//
///*
//经过测试参数2实际意义已经完蛋了
//*/
//func (this *CheckHu) reAllItem(cards []byte, guiNum byte) [][]TagWeaveItem {
//	//先去将，把去将后的牌型都拿出来
//	//普通胡牌牌型3n+2 胡了必然有一对将，考虑赖子补将的可能（单吊的情况）
//	var tagWeaveItem [][]TagWeaveItem
//	var TagItem [][]TagWeaveItem
//	//扔掉指定数量的对子，胡牌的话必然是一对将
//	newList, err := this.createListByRemoveEye(cards, 1)
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	for _, newcard := range newList {
//		fmt.Println(newcard)
//	}
//	min := 0
//	max := 0
//	var eye byte = 0
//	nextgui := guiNum
//	for i := 0; i < 4; i++ {
//		if i < 3 {
//			min = i * 9
//			max = min + 9
//		} else {
//			min = i * 9
//			max = min + 7
//		}
//		TagItem, eye, nextgui = this.checkGuiRuleAll(cards[min:max], i, nextgui, i < 3, eye)
//		tagWeaveItem = append(tagWeaveItem, TagItem...)
//		// }
//	}
//	if eye != 0 {
//		for _, v := range tagWeaveItem {
//			if len(v) != 4-len(this.userHandInfo.weaveItem) {
//				fmt.Println("配型失败")
//				continue
//				// } else {
//				// 	fmt.Println(tagWeaveItem[i])
//			}
//		}
//		return tagWeaveItem
//	}
//	//必须是3n+2=14 所以n必须是4了
//	//创建
//	return nil
//}
