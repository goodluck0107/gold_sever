package lib_gar

import (
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

/*
20181129 苏大强
杠牌单元设计：
目前了解的几种杠：
	暗杠：判断的时候手牌有4张（原本就有或者3+1）
	回头杠：已经倒了3张，又起了一张
	蓄杠：已经倒了3张，手上还有一张
	明杠：手上3张，别人打了一张（必须操作权最大，别人胡牌的情况不可能）

	//几种杠的判断需要知道判牌的来源，结构在cardutils中
*/
//空掉吃位
const (
	NOOPERATOR = 0x00 //无操作
	PENG       = 0x8  //碰
	CHIGONG    = 0x10 //明杠
	XUGONG     = 0x20 //蓄杠
	BACKGONG   = 0x40 //回头杠
	HIDGONG    = 0x80 //暗杠
)

var (
	GongInfoMap = map[int]string{
		NOOPERATOR: "无操作",
		PENG:       "碰",
		CHIGONG:    "明杠",
		XUGONG:     "蓄杠",
		BACKGONG:   "回头杠",
		HIDGONG:    "暗杠",
	}
)

//--------------这段数据将来放到handInfo里面
type GongPengGropu struct {
	fourcards []byte
}

//--------------这段数据将来放到handInfo里面
type GongPengResult struct {
	Chigongcards  []byte
	Xugongcards   []byte
	BackgongCards []byte
	HidgongCards  []byte
	PengCards     []byte
}

//生成手牌中4张牌的记录
//20181203 其实只要4张牌的记录，其他的就找对应的牌的数量即可
func ClassifyGPCards(checkCards []byte) *GongPengGropu {
	result := &GongPengGropu{
		fourcards: []byte{},
		// threecards: []byte{},
		// TwoCards:   []byte{},
	}
	maxinde := len(checkCards)
	if maxinde > static.MAX_INDEX {
		return nil
	}
	for i := 0; i < maxinde; i++ {
		switch checkCards[i] {
		case 4: //最大3组
			result.fourcards = append(result.fourcards, mahlib2.IndexToCard(byte(i)))
			// case 3: //最多4组
			// 	result.threecards = append(result.fourcards, common.IndexToCard(byte(i)))
			// case 2: //最多7组
			// 	result.TwoCards = append(result.fourcards, common.IndexToCard(byte(i)))
		}
	}
	return result
}

//获得倒牌里碰牌的数据.
func GetWeavePengCards(WeaveItem []static.TagWeaveItem) (PengCards []byte) {
	if len(WeaveItem) == 0 {
		return
	}
	for _, v := range WeaveItem {
		if v.WeaveKind == static.WIK_PENG {
			//碰牌类型
			PengCards = append(PengCards, v.CenterCard)
		}
	}
	return
}
