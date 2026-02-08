package lib_scoring

import "errors"

const (
	MAX_SCORING_INDEX = 81 //可能用不了这么多
	MAX_SOCRING_NUM   = 88 //国标
	SCORING_NORMAL    = 1  //普通胡法
	SCORING_SPECIAL   = 2  //特殊 7对 将一色 风一色 不用3n+2 判断的

	CANBE_ZIMO  = 1      //可以自摸
	CANBE_CHIHU = 1 << 1 //可以吃胡
	//20190122 新加一些东西 因为有些特殊的玩意 7对（豪华 2豪华 3豪华）将一色、风一色
	MASK_7DUI              = 1
	MASK_7DUI_HAOHUA       = 1 << 1
	MASK_7DUI_HAOHUA_TWO   = 1 << 2
	MASK_7DUI_HAOHUA_THREE = 1 << 3
	//MASK_7DUI_NOCONCEALED = 1 << 4   //20190218 7对不用门清
	//MASK_7DUI_TRIPLET = 1 <<5   //20190219 7对支持明杠
	//MASK_7DUI_WARNNOHAO = 1 <<6   //20190219 7对杠牌不能拆开
	MASK_YISE_FENG = 1 << 7 //風一色
	MASK_YISE_EYE  = 1 << 8 //將將胡

	MASK_OFFSET      = 4
	MASK_NOM_NOMAGIC = 0x0001 //混将硬胡
	MASK_NOML_MAGIC  = 0x0002 //混将软胡
	MASK_258_NOMAGIC = 0x0004 //258硬胡
	MASK_258_MAGIC   = 0x0008 //258软胡
	//下面两个是特质的，比如将将胡和清一色，检查将将胡的时候如果赖子也是将牌，这个时候如果不用成牌型，那么这个位就代表其是硬胡，具体看将将胡的代码
	MASK_SPECIAL_NOMAGIC = 0x0010 //非3n+2硬胡
	MASK_SPECIAL_MAGIC   = 0x0020 //非3n+2软胡
	MASK_SPECIAL_GODHU   = 0x0040 //20190219 见字胡
)

var (
	SevenPairsMap = map[int]string{
		MASK_7DUI:              "7小对",
		MASK_7DUI_HAOHUA:       "豪华七对",
		MASK_7DUI_HAOHUA_TWO:   "双豪华七对",
		MASK_7DUI_HAOHUA_THREE: "超豪华七对",
		//MASK_7DUI_NOCONCEALED:"7对不用门清",
		//MASK_7DUI_TRIPLET:"7对支持明杠",
		//MASK_7DUI_WARNNOHAO:"7对杠牌可拆"
	}
)
var (
	CheckPatternMap = map[int]string{
		MASK_7DUI:              "7小对",
		MASK_7DUI_HAOHUA:       "豪华七对",
		MASK_7DUI_HAOHUA_TWO:   "双豪华七对",
		MASK_7DUI_HAOHUA_THREE: "超豪华七对",
		//MASK_7DUI_NOCONCEALED:"7对不用门清",
		//MASK_7DUI_TRIPLET:"7对支持明杠",
		//MASK_7DUI_WARNNOHAO:"7对杠牌可拆"
	}
)

/*
20181228 每个玩法对番型有自己的需求，重新定义番数，
这个可能会扩充的比较庞大，先放这里 目前江陵的需求可以放在公共单元所有都写这里
特殊需求
全求人可以自摸 （能不能自摸可以统一，所有判番只管型，在判胡后再判断是不是自摸，捉铳，抢杠）阶段考虑
混一色可提升为清一色（这个也能吧，判型成功后，往哪个番型上提，再设定一个开关来确定返番的时候要不要改番数）
20190122 非3n+2的番型 由7对，增加到风一色、将一色。这两个好像是武汉玩法的，国标里面还发现了13幺，组合龙，这种比较奇葩的玩意
*/
type GameKindRule struct {
	GameID        int                   `json:"GameID"`        //玩法id 599 江陵
	GetMaxScoring bool                  `json:"GetMaxScoring"` //不是获取最大番数就是叠加番数
	MinHuFan      int                   `json:"MinHuFan"`      //胡牌最小番数
	SpecialMask   uint64                `json:"SpecialMask"`   //20190122 目前只用在7对检查上面 只用到了低4位
	ScoringMask   []*SpecialScoringNeed `json:"ScoringMask"`   //每个游戏的番型的特殊需求 生成专有结构  切片 比如江陵玩法要生成清一色（混一色） 碰碰胡 全求人
	GodHuMask     uint64                `json:"GodHuMask"`     //20190214 見字胡mask 只是对见字胡的来源做了设定
}

//20190105 统一一下，不管是自定义的结构还是modify的都是这个型的
type FanInfo struct {
	FanID      int    `json:"FanID"`      //目标番型编号
	Name       string `json:"Name"`       //20190103 考虑到可能需要返给客户端
	FanShu     int    `json:"FanShu"`     //修改为番值 在创建的时候直接拿对应ID的番数过来
	ZiMoFanShu int    `json:"ZiMoFanShu"` //20190129 自摸番数
	Weight     int    `json:"Weight"`     //20190214 追加一个权重，构建判番的层次 江陵揪马以分为主
}
type FanInfoEX struct {
	ID          int //20190121 源ID
	HuClass     uint64
	BaseFanInfo *FanInfo
}

//变更翻型结构
type ModifyFanInfo struct {
	ModifyID      int      `json:"ModifyID"` //目标番型编号
	ModifyFanInfo *FanInfo `json:"ModifyFanInfo"`
}

//具体的番型结构，主要的
type SpecialScoringNeed struct {
	FanInfo    *FanInfo       `json:"FanInfo"`
	HuMask     byte           `json:"HuMask"`     //不够再改准备用来确定能不能胡（自摸，捉铳) 20181229
	ModifyInfo *ModifyFanInfo `json:"ModifyInfo"` //这个是需要修改的，比如江陵里面混一色当清一色算，生成这个结构会分成2步骤，因为必须要先生成转换番型的数据，才能完整
	//20190121 追加私有不记番 例如江陵里面，全求人就不用记碰碰胡了
	DiscardID []int `json:"DiscardID"`
	//20190322 通城麻将（258）里面，将一色必须成型（3n+2或者是7对？） 设定如果为0 就是检查一下低8位大于0，如果有值 目前就是
	//目前设定
	PatternMask byte `json:"PatternMask"`
}

func NewFanInfoEX(checkId int, ChiHuKind uint64, baseinfo FanInfo) *FanInfoEX {
	copyInfo := NewFanInfo(baseinfo.FanID, baseinfo.Name, baseinfo.FanShu)
	return &FanInfoEX{
		ID:          checkId,
		HuClass:     ChiHuKind,
		BaseFanInfo: copyInfo,
	}
}
func NewgameKindRule(gameID int, getMaxScoring bool, MinHuFan int, specialMask uint64, scoringMask []*SpecialScoringNeed) (*GameKindRule, error) {
	if gameID == 0 || scoringMask == nil {
		return nil, errors.New("NewGameRule参数有问题gameID==0，没有scoringMask")
	}
	newFan := MinHuFan
	if newFan < 0 {
		newFan = 0
	}
	newGameKindRule := &GameKindRule{
		GameID:        gameID,
		GetMaxScoring: getMaxScoring, //不是获取最大番数就是叠加番数
		MinHuFan:      newFan,        //胡牌最小番数
		SpecialMask:   specialMask,   //低8位
		ScoringMask:   scoringMask,
	}
	return newGameKindRule, nil
}

func NewspecialScoringNeed(baseInfo *FanInfo, ctxMask byte, modifyID int, discardID []int) *SpecialScoringNeed {
	if baseInfo == nil {
		return nil
	}
	//20190105 这个地方自摸和放炮的场景准备从checkcard里面拿，这个地方准备设置杠后和海底
	checkCtx := ctxMask
	if ctxMask == 0 || ctxMask > 0x3 {
		checkCtx = 0x3
	}
	var modifyInfo *ModifyFanInfo = nil
	if modifyID != 0 {
		modifyInfo = &ModifyFanInfo{
			ModifyID: modifyID,
		}
	}
	newspecial := &SpecialScoringNeed{
		FanInfo:    baseInfo,
		HuMask:     checkCtx,
		ModifyInfo: modifyInfo,
		DiscardID:  discardID,
	}
	return newspecial
}

//这个要放在最后，因为不知道modify的那个ID的私有型什么时候做好
func NewFanInfo(modifyID int, name string, modifyfanShu int) *FanInfo {
	if modifyID == 0 {
		return nil
	}
	return &FanInfo{
		FanID:  modifyID,
		Name:   name,
		FanShu: modifyfanShu,
	}
}

//公共

func MergeSlace(TarSlace *[]int, souSlace []int) {
	for _, v := range souSlace {
		AddIdToSlace(TarSlace, v)
	}
}

func AddIdToSlace(slace *[]int, value int) {
	if !CheckIdInSlace(*slace, value) {
		*slace = append(*slace, value)
	}
}

//20190121 检查ID是不是在切片里
func CheckIdInSlace(checkSlace []int, value int) bool {
	for _, data := range checkSlace {
		if data == value {
			return true
		}
	}
	return false
}
