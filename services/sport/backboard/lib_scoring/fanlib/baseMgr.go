package fanlib

import (
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*
管理单元 国标
BaseScoringMap 是当前支持的所有基础番型处理结构（国标）
*/
var (
	G_ScoringManager *ScoringManager
)

func init() {
	// fmt.Println("22222222222222222222222222222222222222222222222222222222222222222")
	G_ScoringManager = newScoringManager()
}

//是否满足，并且返回番数
type BaseSCRIO interface {
	GetID() int                                             //番数ID
	GetFanShu() int                                         //获取番数
	GethuKind() int                                         //获取胡方式（3n+2或者不用）
	Name() string                                           //番数名称
	CheckSatisfySelf(handInfoCtx *card_mgr.CheckHuCTX) bool //判番
	//20190109 增加升级设置，如果有赖子，每种番型的处理不太一样
	//20190305 有限定鬼牌数的情况
	SpircalProcess(handInfoCtx *card_mgr.CheckHuCTX) (uint64, byte, []int)
	//20190403 写个普通检查接口的 手牌、检查牌（做什么牌用）、鬼牌的切片
	//返回mask，需要鬼牌的数量
	Check_Normal(cbCardIndex []byte, warnItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needGuiNum byte, err error)
}

type ScoringManager struct {
	BasehanderLen  int               //目前支持的国标番数记录
	BaseScoringMap map[int]BaseSCRIO //以番数为key的，具体番型记录 国标的 目前写了几个，目标是覆盖所有国标和拓展的番型
}

func newScoringManager() *ScoringManager {
	return &ScoringManager{
		BasehanderLen:  0,
		BaseScoringMap: make(map[int]BaseSCRIO), //目前有多少番型，这里就有多少，以id为key的国标番型表
	}
}

//初始化的时候，目前服务器所有番型自动注册上来 国标番
func (self *ScoringManager) RegisterBaseHander(hander BaseSCRIO) {
	id := hander.GetID()
	if _, ok := self.BaseScoringMap[id]; ok {
		return
	}
	self.BaseScoringMap[id] = hander
	self.BasehanderLen += 1
}
