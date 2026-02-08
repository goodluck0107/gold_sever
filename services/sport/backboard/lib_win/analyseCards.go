package lib_win

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

//分析扑克
/*
这个目前是将癞子都替换成牌值后的结果
或者直接测试一个硬胡看看结果
*/
func AnalyseCards(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) (bool, []static.TagAnalyseItem) {
	var AnalyseItemArray []static.TagAnalyseItem
	//统计索引数组中的所有牌数
	cbCardCount := byte(0)
	for i := byte(0); i < static.MAX_INDEX; i++ {
		cbCardCount += cbCardIndex[i]
	}

	//效验数目，索引数组中牌的总数-2（一对将）后剩下的数是3的倍数
	if (cbCardCount < 2) || (cbCardCount > static.MAX_COUNT) || ((cbCardCount-2)%3 != 0) {
		return false, AnalyseItemArray
	}

	////////////////////////下面出现的情况只能是：索引数组中牌的总数-2（一对将）后剩下的数是3的倍数，不可能出现4，6。。张的情况///////////////////////////////////
	//类型子项
	//变量定义
	cbKindItemCount := byte(0)
	var KindItem [static.MAX_COUNT - 2]static.TagKindItem

	//需求判断，cbLessKindItem必须是3的倍数
	cbLessKindItem := byte((cbCardCount - 2) / 3)
	//	ASSERT((cbLessKindItem+cbWeaveCount)==4);

	//单吊判断，cbCardCount=2的情况：原来手中牌只有一张牌，加入要分析的牌后正好构成两张，其他的牌都在组合牌中
	/*
		单吊的话，其实就是设置结果的将眼AnalyseItem.CardEye = self.SwitchToCardData(i)
		剩下的组合都在倒牌里
	*/
	if cbLessKindItem == 0 {
		//效验参数
		//此种情况是：原来手中只有一张牌，加入一张待分析的牌正好构成两张，并且有四组组合牌型
		//ASSERT((cbCardCount==2)&&(cbWeaveCount==4));

		//判断加入了待分析的牌后是否构成对子
		for i := byte(0); i < static.MAX_INDEX; i++ {
			//如果牌索引数组中有一个对子，保存分析结果
			if cbCardIndex[i] == 2 {
				//分析子项
				//变量定义
				var AnalyseItem static.TagAnalyseItem

				//分析每一组组合牌，得到组合牌的组合牌型和中间牌，比如说WK_PENG,WK_CHI保存到分析子项中
				for j := byte(0); j < cbWeaveCount; j++ {
					AnalyseItem.WeaveKind[j] = WeaveItem[j].WeaveKind
					if index, err := mahlib2.CardToIndex(WeaveItem[j].CenterCard); index != static.INVALID_BYTE {
						//AnalyseItem.CenterCard[j] = public.SwitchToCardIndex(WeaveItem[j].CenterCard)//centercard统一风格
						AnalyseItem.CenterCard[j] = index
					} else {
						fmt.Println(err)
						return false, nil
					}

				}
				//将待分析的牌索引转换成牌值，作为牌眼保存起来
				if card := mahlib2.IndexToCard(i); card != static.INVALID_BYTE {
					AnalyseItem.CardEye = card
				} else {
					return false, nil
				}

				//将分析结果插入到分析数组中
				//AnalyseItemArray.Add(AnalyseItem);
				AnalyseItemArray = append(AnalyseItemArray, AnalyseItem)

				return true, AnalyseItemArray
			}
		}
		return false, AnalyseItemArray
	}
	//加入待分析的牌后，手中牌>=3的情况，对手中牌索引数组进行分析
	if cbCardCount >= 3 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			//1.同牌判断，胡牌时，杠牌算碰牌
			if cbCardIndex[i] >= 3 {
				KindItem[cbKindItemCount].CenterCard = i
				KindItem[cbKindItemCount].CardIndex[0] = i
				KindItem[cbKindItemCount].CardIndex[1] = i
				KindItem[cbKindItemCount].CardIndex[2] = i
				KindItem[cbKindItemCount].WeaveKind = static.WIK_PENG
				cbKindItemCount++
				//4个赖子的情况，最多可能出现8个同牌
				if cbCardIndex[i] >= 6 {
					KindItem[cbKindItemCount].CenterCard = i
					KindItem[cbKindItemCount].CardIndex[0] = i
					KindItem[cbKindItemCount].CardIndex[1] = i
					KindItem[cbKindItemCount].CardIndex[2] = i
					KindItem[cbKindItemCount].WeaveKind = static.WIK_PENG
					cbKindItemCount++
				}
			}
			//2.连牌判断
			//????(i<(public.MAX_INDEX-2)不清楚为什么会这样写，麻将总点数也才到26，就算能到34，风也不能构成连牌?????????
			/*
				看样子应该是改过的，只有万条筒才能有顺子，万条筒做顺子，下标只能到7，789是最后的
				这里做j是为了一次将i这张牌能做的顺子都做完
			*/
			if (i < (29 - 2)) && (cbCardIndex[i] > 0) && ((i % 9) < 7) {
				for j := byte(1); j <= cbCardIndex[i]; j++ {
					if (cbCardIndex[i+1] >= j) && (cbCardIndex[i+2] >= j) {
						KindItem[cbKindItemCount].CenterCard = i
						KindItem[cbKindItemCount].CardIndex[0] = i
						KindItem[cbKindItemCount].CardIndex[1] = i + 1
						KindItem[cbKindItemCount].CardIndex[2] = i + 2
						KindItem[cbKindItemCount].WeaveKind = static.WIK_LEFT
						cbKindItemCount++
					}
				}
			}
		}
	}
	//组合分析，cbLessKindItem是手中牌总数-2后得到的3的倍数，这个值最大是4，cbKindItemCount是对手中牌进行分析后得出来的最多的组合类型
	if cbKindItemCount >= cbLessKindItem {
		//变量定义
		cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
		//变量定义
		cbIndex := [4]byte{0, 1, 2, 3}
		var pKindItem [4]*static.TagKindItem
		//do循环的作用：
		/*
				cbLessKindItem：一组牌中如果可以胡牌的话，需要的组合数，
				cbKindItemCount：对手中牌的组合进行分析，可以得出的最多的组合数
				1.将待分析的牌保存到临时数组中，取前面分析出来的所有组合，每次分析cbKindItemCount个，
				  在临时数组中减去这cbKindItemCount个组合中的牌，对剩下的牌进行分析，如果还有对牌，可以
				  胡牌，保存这种组合类型到分析数组中，
			   2.设置索引数组，将后面的组合下标放到索引数组中，下一次循环的时候就取新设置的索引数组对应的
			     分析子项。再进行判断。
		*/
		nAccert := 0
		for {
			nAccert++
			if nAccert > 600 {
				//m_mylog.Log("分析扑克时死循环啦");
				break
			}
			//每次循环将传进来的牌索引数组拷贝到临时数组中，进行分析
			copy(cbCardIndexTemp, cbCardIndex[:])
			//每次从上面分析得出的分析子项中取cbLessKindItem个分析子项进行分析，
			//注意：索引数组cbIndex[]在每次循环结束时都重新设置了
			for i := byte(0); i < cbLessKindItem; i++ {
				pKindItem[i] = &KindItem[cbIndex[i]]
			}
			//数量判断
			bEnoughCard := true
			//修改临时数组的值，把临时数组中构成cbLessKindItem个分析子项里的每一张牌，牌数减1，
			//把组合里的牌剪掉
			for i := byte(0); i < cbLessKindItem*3; i++ {
				//存在判断
				cbCardIndex := pKindItem[i/3].CardIndex[i%3]
				if cbCardIndexTemp[cbCardIndex] == 0 {
					bEnoughCard = false
					break
				} else {
					cbCardIndexTemp[cbCardIndex]--
				}
			}
			//胡牌判断，注意下面使用到的cbCardIndexTemp[]数组是经前面修改过后的
			if bEnoughCard == true {
				//牌眼判断
				cbCardEye := byte(0)
				//检查cbCardIndexTemp[]数组中是否还有对牌，如果有就将该对牌设置为牌眼，说明此种组合有可能
				//会胡牌，跳出
				for i := byte(0); i < static.MAX_INDEX; i++ {
					if cbCardIndexTemp[i] == 2 {
						if card := mahlib2.IndexToCard(i); card != static.INVALID_BYTE {
							cbCardEye = card
							break
						} else {
							return false, nil
						}
					}
				}
				//组合类型
				if cbCardEye != 0 {
					//变量定义
					var AnalyseItem static.TagAnalyseItem

					//得到组合牌中的牌型，保存到分析子项中
					for i := byte(0); i < cbWeaveCount; i++ {
						AnalyseItem.WeaveKind[i] = WeaveItem[i].WeaveKind
						if index, err := mahlib2.CardToIndex(WeaveItem[i].CenterCard); index != static.INVALID_BYTE {
							//AnalyseItem.CenterCard[j] = public.SwitchToCardIndex(WeaveItem[j].CenterCard)//centercard统一风格
							AnalyseItem.CenterCard[i] = index
						} else {
							fmt.Println(err)
							return false, nil
						}
						//AnalyseItem.CenterCard[i] = self.SwitchToCardIndex(WeaveItem[i].CenterCard)
					}
					//得到手中牌的牌型，保存到分析子项中
					for i := byte(0); i < cbLessKindItem; i++ {
						AnalyseItem.WeaveKind[i+cbWeaveCount] = pKindItem[i].WeaveKind
						AnalyseItem.CenterCard[i+cbWeaveCount] = pKindItem[i].CenterCard
					}
					//设置牌眼
					AnalyseItem.CardEye = cbCardEye
					//将分析子项插入到分析数组中
					//AnalyseItemArray.Add(AnalyseItem);
					AnalyseItemArray = append(AnalyseItemArray, AnalyseItem)
				}
			}
			//设置索引，索引数组中存放的是分析子项数组的下标，每次取分析子项进行分析时，都是按照索引数组
			//里面存放的下标值进行存取，当cbIndex[cbLessKindItem-1]的最后一位存放的值与得出的分析子项下标相同，
			//重新调整索引数组，下一次取值就会取新的组合
			if cbIndex[cbLessKindItem-1] == (cbKindItemCount - 1) {
				var i byte
				for i = cbLessKindItem - 1; i > 0; i-- {
					if (cbIndex[i-1] + 1) != cbIndex[i] {
						cbNewIndex := cbIndex[i-1]
						for j := (i - 1); j < cbLessKindItem; j++ {
							cbIndex[j] = cbNewIndex + j - i + 2
						}
						break
					}
				}
				//跳出整个while循环
				if i == 0 {
					break
				}
			} else {
				cbIndex[cbLessKindItem-1]++
			}
		}
	}
	return (len(AnalyseItemArray) > 0), AnalyseItemArray
}
