package card_mgr

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

//20190116 修改 ExInfo这个信息是后设的
type CardCTX struct {
	BaseInfo *CardBaseInfo
	ExInfo   *CardEx //这个属性会修改的
}

//20190108 先丢个架子
//20190116 这个只是基础信息
func NewCardBaseInfo(id byte, origin byte) (*CardBaseInfo, error) {
	if !mahlib2.IsMahjongCard(id) {
		return nil, errors.New(fmt.Sprintf("牌值（%d）~（%d），传入值（%d）", 1, static.MAX_INDEX, id))
	}
	if origin == 0 || origin > 0x8 {
		return nil, errors.New(fmt.Sprintf("牌值（%d），来源异常（%d）", origin))
	}
	return &CardBaseInfo{
		ID:     id,
		Origin: origin,
	}, nil

}

//当需要设定是什么类型的牌的时候需要做 cardClass是handLinfo里面来的，我认为赖子和皮子是弱关联，所以需要
func NewCardCTX(baseInfo *CardBaseInfo, cardClass int) (*CardCTX, error) {
	if baseInfo == nil {
		return nil, errors.New("没有基础牌值信息")
	}
	//未做安全监测
	newExInfo := &CardEx{
		ID:        baseInfo.ID,
		CardClass: cardClass,
	}
	return &CardCTX{
		BaseInfo: baseInfo,
		ExInfo:   newExInfo,
	}, nil
}
