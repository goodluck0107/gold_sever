package fanlib

import (
	cardmgr2 "github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*说明：
海底捞月：和打出的最后一张牌。
*/

const (
	_SCORING_45_ID     = 45
	_SCORING_45_NAME   = "海底捞月"
	_SCORING_45_FANSHU = 8
)

var _SCORING_45_DISCARDID_ = []int{}

// //自己注册
// func init() {
// 	G_ScoringManager.RegisterBaseHander(&Scoring_45{
// 		id:           _SCORING_45_ID,
// 		name:         _SCORING_45_NAME,
// 		fanShu:       _SCORING_45_FANSHU,
// 		setDiscardID: _SCORING_45_DISCARDID_,
// 		huKind:       lib_scoring.SCORING_NORMAL,
// 		humask:       lib_scoring.CANBE_ZIMO | lib_scoring.CANBE_CHIHU,
// 	})
// }

type scoring_45 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int
	humask       byte
}

func (this *scoring_45) GetID() int {
	return this.id
}

func (this *scoring_45) Name() string {
	return this.name
}

func (this *scoring_45) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_45) GethuKind() int {
	return this.huKind
}

/*
海底捞其实可以不进这里
*/
func (this *scoring_45) CheckSatisfySelf(handInfoCtx *cardmgr2.CheckHuCTX) bool {

	return false
}

func (this *scoring_45) SpircalProcess(handInfoCtx *cardmgr2.CheckHuCTX) bool {
	return true
}
