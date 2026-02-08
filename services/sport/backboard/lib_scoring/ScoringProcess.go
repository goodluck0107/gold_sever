package lib_scoring

import (
	cardmgr2 "github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*
处理单元
20190105，目前因为只看到两个处理，所以只是做了简单的处理
1：每个游戏的番型是不是能胡（国标中的全求人是不能自摸的，但是江陵里可以，所有放这里处理）
2：番型转换（这个应该只会有一种转换：江陵里面混一色转换成清一色）

*/
type ScoringInfo struct {
	CustomInfo *SpecialScoringNeed //每种玩法的记分，比如江铃 清一色是3分而已
}

//每个游戏创建自己的规则单元，需要把国标的接口带进来 第一步
func NewScoringInfo(fanName string, Scoring *SpecialScoringNeed) *ScoringInfo {
	newScoring := &ScoringInfo{
		CustomInfo: Scoring,
	}
	newScoring.CustomInfo.FanInfo.Name = fanName
	return newScoring
}

//20181229 先确定国标番型是不是OK,再过自己专有的
//同一番型下单个番型检查
//20190119 不计番和不可能番其实处理上是一样的，都是往slBanID里面加
// func (self *ScoringInfo) GetFanShuSingle(handInfoCtx *singleton_card.CheckHuCTX, satisfyedID []int, slBanID []int) (*FanInfo, []int, []int) {
func (self *ScoringInfo) GetFanShuSingle(handInfoCtx *cardmgr2.CheckHuCTX) (*FanInfo, []int) {
	//-------------不可能番处理，暂不添加 0119--------
	//--------------------
	if self.CustomInfo != nil {
		//首先判断能不能胡 来源就是自摸或者放炮 20190216 准备优先检查能不能判断
		//if !self.checkCanBeHu(handInfoCtx.CheckCardEX.BaseInfo.Origin) {
		//	return nil, nil
		//}
		//要不要提权，江陵里面混一色会提成清一色，目前没写混一色，先不判断 20190102 苏大强
		// if len(self.CustomInfo.DiscardID) != 0 {
		// 	fmt.Println(fmt.Sprintf("checked （%v）", self.CustomInfo.DiscardID))
		// }
		return self.modifyScoring(), self.CustomInfo.DiscardID
	}
	return nil, nil
}

//20190121 写个不计番的功能看看

//-----------------------------以下是目前见到的需要判断的场景
/*
1、改番权，混一色按照清一色算
2、场景下能不能胡，江陵里面，全球人自摸允许，国标是不允许的
*/
//改番权 例如江陵里面混一色要改成清一色 在外面创建的时候，直接拿对应ID的番值，这个地方就直接获取
//应该只会有一种番型可提，不太可能出现同时两种
func (self *ScoringInfo) modifyScoring() (result *FanInfo) {
	// if self.CustomInfo == nil {
	// 	return lib_scoring.NewModifyScring(self.BaseIO.GetID(), self.BaseIO.Name(), self.BaseIO.GetFanShu())
	// }
	if self.CustomInfo.ModifyInfo != nil {
		return self.CustomInfo.ModifyInfo.ModifyFanInfo
	}
	return self.CustomInfo.FanInfo
}

//20190322 新规中有出现 通山麻将（258），将将胡成牌型，7对或者3n+2（258）
func (self *ScoringInfo) CheckCanBeHuEX(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	//检查要不要成牌型 兼容以前的，如果没有值，检查低八位是不是为零 但是7对绝对是特殊
	if self.CustomInfo.FanInfo.FanID != 22 {
		//只要不是7对的都查一下,兼容以前，如果是0的话，只要ChiHuKind低8位大于0就行
		if self.CustomInfo.PatternMask == 0 {
			if handInfoCtx.ChiHuKind&0xff == 0 {
				//根本就没有任何牌型
				return false
			}
		}

	}

	//检查能不能自摸啊啥的
	return self.CheckCanBeHu(handInfoCtx.CheckCardEX.BaseInfo.Origin)
}

//这里还是有两种情况，风一色和将一色
func (self *ScoringInfo) CheckCanBeHu_Pattern(cardCtxMask byte) bool {
	if self.CustomInfo.PatternMask != 0 {
		//如果要求检查那就必须严格匹配

	}
	return false
}

//场景判断，比如全球人国标不能自摸，但是有的地方可以，那么能不能返番数要看能不能算数
//参数是从handinfo来的场景信息
func (self *ScoringInfo) CheckCanBeHu(cardCtxMask byte) bool {
	//没有特殊设定，番型成立
	if self.CustomInfo == nil {
		return true
	}
	// 这里的mask就是3，自摸和捉铳，但是牌的来源就有4种
	if cardCtxMask&(cardmgr2.ORIGIN_NOM|cardmgr2.ORIGIN_HAND) > 0 {
		//自摸的情况
		return self.CustomInfo.HuMask&0x1 > 0
	}
	return self.CustomInfo.HuMask&0x2 > 0
}
