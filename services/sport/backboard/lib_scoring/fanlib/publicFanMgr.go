package fanlib

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	scoringlib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_scoring"
	cardmgr2 "github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*
20190105 重新规划，这个单元是服务器整个fan型查找的管理单元
两个属性：
1：BaseScoring 国际表
2：服务器目前支持的所有kindID的查番表
下阶段要处理的东西：
考虑到有的番型不能共存，或者存在不计的情况（国番里面很明显）；比如（清一色实际上是不可能和混一色共存的，当发现符合其中任何一种后应该剔除另外一个，减少查询的情况）
*/
type PublicFanManager struct {
	BaseScoring    *ScoringManager      //国标番型表指针
	allgameScoring map[int]*BalanceRule //以玩法的kindID为key的每个玩法自己的算番对象
}

//20190103 修改为把国标对象整体拿进来，不然就要成为map copy了
func NewPublicScoringManager(baseScoring *ScoringManager) (*PublicFanManager, error) {
	if baseScoring == nil {
		if G_ScoringManager == nil {
			return nil, errors.New("没有创建标准库")
		} else {
			baseScoring = G_ScoringManager
		}
	}

	NewPublicScoringManager := &PublicFanManager{
		BaseScoring:    baseScoring,
		allgameScoring: make(map[int]*BalanceRule),
	}
	return NewPublicScoringManager, nil
}

//每个玩法自己的东西需要添加在map中去
func (this *PublicFanManager) SetNewRule(customRule *BalanceRule) {
	if _, ok := this.allgameScoring[customRule.GamebaseRule.GameID]; ok {
		fmt.Println(fmt.Sprintf("玩法（%d）已经存在", customRule.GamebaseRule.GameID))
		return
	}
	this.allgameScoring[customRule.GamebaseRule.GameID] = customRule
	//添加完后创建番表
	err := this.registerPrivatelyScoring(customRule.GamebaseRule)
	if err != nil {
		fmt.Println(err)
	}
}

//以国标为基础创建每个游戏自己的判番，注意这个参数是每个游戏调用NewGameRule创建的
func (this *PublicFanManager) registerPrivatelyScoring(privatelyNeed *scoringlib2.GameKindRule) error {
	//不太可能会用国标。所以
	if privatelyNeed == nil || len(privatelyNeed.ScoringMask) == 0 {
		return errors.New(fmt.Sprintf("游戏新建独立算番失败，PrivatelyNeed为空吗（%t），要么没有独立番的特值", privatelyNeed == nil))
	}
	if this.BaseScoring.BasehanderLen == 0 {
		return errors.New("国标库为空，请查明原因")
	}
	//20190102 调整直接查gameKindID，存在就不管了
	if _, ok := this.allgameScoring[privatelyNeed.GameID]; ok {
		return nil
	}
	//创建单个游戏的番表
	newRuld, err := NewBalanceRule(privatelyNeed, this.BaseScoring)
	if err != nil {
		return err
	}
	this.allgameScoring[newRuld.GamebaseRule.GameID] = newRuld
	//创建完
	return nil
}

//
func (this *PublicFanManager) GetFanShu_single(method *cardmgr2.CheckHuCTX, checkObj *BalanceRule, checkSpecialFan bool, fireID *[]int) (fanShuSclie []*scoringlib2.FanInfoEX) {
	getMaxScoring := checkObj.GamebaseRule.GetMaxScoring
	recordMap := checkObj.CustomerScoringRecordMap
	checkItem := checkObj.BaseOrg
	if checkSpecialFan {
		checkItem = checkObj.SpecialOrg
	}
	var checkHander []*scoringlib2.ScoringInfo
	var accordID []int //已经检查的番型
	for i := len(checkItem.fanShuArray) - 1; i >= 0; i-- {
		key := checkItem.fanShuArray[i]
		//拿出对应的IO
		slHander := checkItem.CustomerScoringMap[key]
		//拿出来测
		checkHander = make([]*scoringlib2.ScoringInfo, len(slHander))
		index := 0
		for _, hander := range slHander {
			checkHander[index] = recordMap[hander]
			index += 1
		}
		//第二阶段再处理的东西 每种番型可能会有冲突的情况，比如清一色，混一色不能共存，那么检查一种成功后把另外一种去掉
		//从高到低，满足最高的就把所有的都遍历了 ，每次都是获取一个番数中所有的可能的胡番
		for _, hander := range checkHander {
			if hander == nil {
				continue
			}
			checkId := hander.CustomInfo.FanInfo.FanID
			//20190216 优先检查状态是不是符合查番 将来扩充到ctx
			if !hander.CheckCanBeHu(method.CheckCardEX.BaseInfo.Origin) {
				continue
			}
			if result, needgui, tempDiscardID := this.CheckAccord(method, checkId, accordID, *fireID); result != 0 {
				//20190221 对见字胡直接处理 不是很完美，要考虑修改
				if checkObj.GamebaseRule.GodHuMask != 0 && method.ChiHuKind&scoringlib2.MASK_SPECIAL_GODHU != 0 {
					if uint64(method.CheckCardEX.BaseInfo.Origin)&checkObj.GamebaseRule.GodHuMask == 0 {
						method.ChiHuKind = static.CHK_NULL
						return nil
					}
				}
				//20190305 可能会出现软清一色，硬混一色的情况，那么这个地方要限定一下,只有硬胡的情况下，再剔除
				if len(tempDiscardID) != 0 && result&static.CHK_PING_HU_NOMAGIC != 0 {
					//20190121 如果国番里面有不可能共存的番型，加到fire里面去
					scoringlib2.MergeSlace(fireID, tempDiscardID)
					// fmt.Println(fmt.Sprintf("国标查番（%d），不计（%v）fireID（%v）", checkId, tempDiscardID, fireID))
					// fmt.Println(fmt.Sprintf("修改后fireID（%v）", fireID))
					// common.Print_cards(method.CheckCardItem[:])
				}
				//20190305 再做一次处理,嘉鱼红中赖子杠限定赖子 大胡不限制
				if method.Expand_Mask != nil && method.Expand_Mask.Restrict_GodItem != nil && method.Expand_Mask.Restrict_GodItem.Restrict_Fan {
					//Restrict_GodNum是硬指标，在查胡之前就干掉的
					if needgui > method.Expand_Mask.Restrict_GodItem.Restrict_NeedGodNum {
						//不符合要求了 不拿分 这个可以测试一下软清，硬将将胡 20190305
						scoringlib2.AddIdToSlace(&accordID, checkId)
						continue
					}
				}
				tmpFanShu, slTmpBanID := hander.GetFanShuSingle(method)
				//20190105 为了以防万一，如果自身没有设定指定番，那么返回国标的番数
				if tmpFanShu != nil {
					resultFanShu := scoringlib2.NewFanInfoEX(checkId, result, *tmpFanShu)
					//胡kind添加，每个判番里面会反自己的判断，这个地方只是判断的钥匙
					// method.SetchiHuKind(result)
					fanShuSclie = append(fanShuSclie, resultFanShu)
					//只要最大番，就出去了
					if getMaxScoring && len(fanShuSclie) > 0 {
						//根据设定这个时候拿的就是最大番里的东西
						return fanShuSclie
					}
					// setMaxFanShuID = tmpSatisfyID
					if len(slTmpBanID) != 0 {
						//20190121 如果国番里面有不可能共存的番型，加到fire里面去
						scoringlib2.MergeSlace(fireID, slTmpBanID)
						//fmt.Println(fmt.Sprintf("查番（%d），不计（%v）", checkId, slTmpBanID))
					}
					// checkIO := this.BaseScoring.BaseScoringMap[checkId]
					// tmpFanShu = lib_scoring.NewFanInfo(checkIO.GetID(), checkIO.Name(), checkIO.GetFanShu())
				}
			}
			//不管是不是ok，这个ID都要记录了
			scoringlib2.AddIdToSlace(&accordID, checkId)

		}
	}
	return fanShuSclie
}
func (this *PublicFanManager) CheckRuleIsRegister(gameKindId int) (result bool) {
	_, result = this.allgameScoring[gameKindId]
	return
}

//20190123
//-------------------------------------
//判番，在判胡后做
//根据gameKind调用不同玩法的判番
//突然发现还有7对清一色此类混合牌型。。。
func (this *PublicFanManager) GetFanShu(huMethod *cardmgr2.CheckHuCTX, gameKindId int) []*scoringlib2.FanInfoEX {
	//检查有没有
	if huMethod == nil {
		fmt.Println("无手牌数据")
		return nil
	}
	var fanShuSclie []*scoringlib2.FanInfoEX = nil
	if checkItem, ok := this.allgameScoring[gameKindId]; ok {
		// huMethod.SetchiHuKind(checkItem.GamebaseRule.SpecialMask)
		//20190222
		if huMethod.ChiHuKind == static.CHK_NULL {
			//fmt.Println(fmt.Sprintf("3n+2查胡失败，游戏规则（%d）特殊胡需求（%b）", gameKindId, checkItem.GamebaseRule.SpecialMask))
			// return nil
		}
		//----------先查不用3n+2的-----------考虑非常规和常规番数可能不一样的情况--------------
		var fireID []int //不计番型
		//****************
		if checkItem.SpecialOrg != nil {
			//目前只有7对在用
			huMethod.SetSpecialMask(checkItem.GamebaseRule.SpecialMask)
			//如果3n+2不能胡，但是可以胡7对，有god就会是见字胡，那么godmask就要传进去
			if checkItem.GamebaseRule.GodHuMask != 0 {
				huMethod.SetFanGodHuMask(checkItem.GamebaseRule.GodHuMask)
			}
			fanShuSclie = this.GetFanShu_single(huMethod, checkItem, true, &fireID)
			//这里没用检查GetMaxScoring 是因为SpecialOrg里的番型都是独立可成胡的
		}
		// common.Print_cards(huMethod.CheckCardItem[:])
		var baseFanShuSclie []*scoringlib2.FanInfoEX = nil
		if huMethod.ChiHuKind != static.CHK_NULL {
			baseFanShuSclie = this.GetFanShu_single(huMethod, checkItem, false, &fireID)
			if baseFanShuSclie != nil {
				//如果是要max，这里要比较，如果不是，就append了
				if checkItem.GamebaseRule.GetMaxScoring {
					//到这里，最多2个，因为每个都取的是最大番中的一个
					if baseFanShuSclie[0].BaseFanInfo.FanShu > fanShuSclie[0].BaseFanInfo.FanShu {
						return baseFanShuSclie
					}
				} else {
					//合并。因为两种番库不会重叠，那么
					fanShuSclie = append(fanShuSclie, baseFanShuSclie...)
				}
			}
		}
	}
	//----------------------------------------------
	return fanShuSclie
}

/*
//20190121 考虑实情，我们需要知道哪些番不用查，比如清一色和混一色不能共存，江陵里面混一色算清一色，而在江陵里面6番的全求人确认后就不用再查碰碰胡了
//递归？
//参数说明：
1、手牌数据结构
2、已经检查过的序列
*/
func (this *PublicFanManager) CheckAccord(handInfoCtx *cardmgr2.CheckHuCTX, checkFanID int, ccordID []int, fireID []int) (uint64, byte, []int) {
	//进来的ID先过滤一下
	if scoringlib2.CheckIdInSlace(ccordID, checkFanID) || scoringlib2.CheckIdInSlace(fireID, checkFanID) {
		return 0, 0, nil
	}
	//
	return this.CheckAccord_Single(handInfoCtx, checkFanID)
}

//20190105  预备在上层处理这个问题，那么就不用在每个游戏玩法里在做接口了
//同一番型下单个番型检查
func (this *PublicFanManager) CheckAccord_Single(handInfoCtx *cardmgr2.CheckHuCTX, checkFanID int) (uint64, byte, []int) {
	// if this.BaseIO != nil {
	itemIO, ok := this.BaseScoring.BaseScoringMap[checkFanID]
	if !ok {
		fmt.Println(fmt.Sprintf("在基础库里没有ID(%d)的检查番型接口", checkFanID))
		return 0, 0, nil
	}
	return itemIO.SpircalProcess(handInfoCtx)
}

// //重新排列要查的
// func (this *PublicFanManager) getHanderExcept(slExceptChkID []int) []BaseSCRIO {
// 	var allHander []BaseSCRIO
// 	var slFanShuKey sort.IntSlice
// 	for key, _ := range this.allgameScoring {
// 		slFanShuKey = append(slFanShuKey, key)
// 	}

// 	slFanShuKey.Sort()

// 	for i := len(slFanShuKey) - 1; i >= 0; i-- {
// 		key := slFanShuKey[i]
// 		slHander := mgr.allHander[key]
// 		for _, hander := range slHander {
// 			if common.InIntSlace(slExceptChkID, hander.GetID()) {
// 				continue
// 			}
// 			allHander = append(allHander, hander)
// 		}
// 	}
// 	return allHander
// }
