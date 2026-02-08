package fanlib

import (
	"fmt"
	scoringlib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_scoring"
)

/*
游戏番表管理单元
20190122 特殊番型除了7对、将一色、风一色。。。可能还有别的不用3n+2
//20190123 检查序列 修改成两个，一个是3n+2 一个是非3n+2
//20190131 武汉晃晃里面见字胡能改变后续检查，不能捉铳； 斑马里面全球人必须258
20190201 设定固定的8层查询0~7

*/

//单独的单元
//这个单元用来创建每个游戏需要的东西
type BalanceRule struct {
	GamebaseRule *scoringlib2.GameKindRule //游戏的基本信息，规则
	//20190131
	SpecialOrg *BaseClass //类似7对这种不用符合3n+2的胡型
	BaseOrg    *BaseClass //非SpecialOrg所有的（3n+2 或带上海底ctx此类的）
	//20190103 为了方便查询和提番，还是生成一个以番号ID为map的记录
	CustomerScoringRecordMap map[int]*scoringlib2.ScoringInfo
}

//这个单元的初始化在每个游戏里面
func NewBalanceRule(gamebaserule *scoringlib2.GameKindRule, baseScoring *ScoringManager) (*BalanceRule, error) {
	newBalanceRule := &BalanceRule{
		GamebaseRule:             gamebaserule,
		BaseOrg:                  NewBaseClass(),
		SpecialOrg:               NewBaseClass(),
		CustomerScoringRecordMap: make(map[int]*scoringlib2.ScoringInfo),
	}
	//20190306 现在已经支持 风一色 将将胡而非7对了
	//if gamebaserule.SpecialMask != public.CHK_NULL {
	//	newBalanceRule.SpecialOrg = NewBaseClass()
	//}
	//创建所有的map
	err := newBalanceRule.CreatGameScoringMap(baseScoring.BaseScoringMap)
	if err != nil {
		return nil, err
	}
	return newBalanceRule, nil
}

/*
//创建一个游戏的私有判番库
*/
func (this *BalanceRule) CreatGameScoringMap(baseScoringMap map[int]BaseSCRIO) (err error) {
	for _, v := range this.GamebaseRule.ScoringMask {
		FanID := v.FanInfo.FanID
		if _, ok := this.CustomerScoringRecordMap[FanID]; ok {
			//2090103 如果有了就不加了
			fmt.Println(fmt.Sprintf("游戏ID（%d）中番型（%d）的规则已经存在，不重复创建", this.GamebaseRule.GameID, FanID))
			continue
		}
		item, ok := baseScoringMap[FanID]
		if !ok {
			fmt.Println(fmt.Sprintf("国标库中没有游戏ID（%d）的判番表(%d)(%s)", this.GamebaseRule.GameID, v.FanInfo.FanID, v.FanInfo.Name))
			continue
		}
		//第一步把规则都创建完成
		customIO := scoringlib2.NewScoringInfo(baseScoringMap[FanID].Name(), v)
		//检查下有没有要提权的
		if customIO.CustomInfo.ModifyInfo != nil {
			//注意提权项的创建必须放在最后，因为我这里提权项的番数是依赖私有的，其实也可以不依赖，多写一个参数但是可能会出现纰漏导致番数不同
			if item, ok := this.CustomerScoringRecordMap[customIO.CustomInfo.ModifyInfo.ModifyID]; ok {
				customIO.CustomInfo.ModifyInfo.ModifyFanInfo = item.CustomInfo.FanInfo
			} else {
				fmt.Println(fmt.Sprintf("未发现特定游戏（%d）要提权（%d）的信息", this.GamebaseRule.GameID, FanID))
				customIO.CustomInfo.ModifyInfo = nil
			}
		}
		this.CustomerScoringRecordMap[FanID] = customIO
		this.Reordering(customIO, item.GethuKind())
	}
	return nil
}

//以每个游戏自己的番数重新排序，江陵里面单大胡都是3番，全求人是6番   第二步
//根据不同的属性写到不同的表里
func (this *BalanceRule) Reordering(customIO *scoringlib2.ScoringInfo, huKind int) {
	reorderingFanId := customIO.CustomInfo.FanInfo.FanID
	//20190213 修改为如果有权重就按权重来 不然按照分来
	index := customIO.CustomInfo.FanInfo.Weight
	if index == 0 {
		index = customIO.CustomInfo.FanInfo.FanShu
	}
	if huKind == scoringlib2.SCORING_SPECIAL {

		this.SpecialOrg.recordering(reorderingFanId, index)
	} else {
		this.BaseOrg.recordering(reorderingFanId, index)
	}
}
