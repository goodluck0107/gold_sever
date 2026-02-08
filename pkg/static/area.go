package static

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-source/game/chess.git/pkg/consts"
)

// 客服手机号
const (
	GMMobile           = "暂无"
	AreaRedisKeyPrefix = "{areadata}"
)

type AreaPackageKind uint

const (
	AreaPackageKindCard   = 1
	AreaPackageKindGold   = 2
	AreaPackageKindApplet = 3
)

func (apt AreaPackageKind) String() string {
	switch apt {
	case AreaPackageKindCard:
		return "gamecard"
	case AreaPackageKindGold:
		return "gamegold"
	case AreaPackageKindApplet:
		return "gameapplet"
	default:
		return "gamecard"
	}
}

type AreaPackageType uint

const (
	AreaPackageTypeUnknown AreaPackageType = iota // 未知玩法类型
	AreaPackageTypeFlower  AreaPackageType = 1    // 花牌
	AreaPackageTypePlate   AreaPackageType = 2    // 字牌
	AreaPackageTypePoke    AreaPackageType = 3    // 扑克
	AreaPackageTypeMahJng  AreaPackageType = 4    // 麻将
)

func (apt AreaPackageType) String() string {
	switch apt {
	case AreaPackageTypeFlower:
		return "花牌"
	case AreaPackageTypePlate:
		return "字牌"
	case AreaPackageTypePoke:
		return "扑克"
	case AreaPackageTypeMahJng:
		return "麻将"
	default:
		return "未知"
	}
}

func GetAreaPackageType(packageType string) AreaPackageType {
	switch {
	case strings.Contains(packageType, "麻"):
		return AreaPackageTypeMahJng
	case strings.Contains(packageType, "扑"):
		return AreaPackageTypePoke
	case strings.Contains(packageType, "字"):
		return AreaPackageTypePlate
	case strings.Contains(packageType, "花"):
		return AreaPackageTypeFlower
	}
	return AreaPackageTypeUnknown
}

type AreaSeekType uint

const (
	AreaSeekTypeAllPackage AreaSeekType = iota // 未知玩法类型
	AreaSeekTypeRecommend  AreaSeekType = 1    // 推荐玩法
	AreaSeekTypeNearby     AreaSeekType = 2    // 周边玩法
)

func (apt AreaSeekType) String() string {
	switch apt {
	case AreaSeekTypeRecommend:
		return "推荐玩法"
	case AreaSeekTypeNearby:
		return "周边玩法"
	default:
		return "所有玩法"
	}
}

// keys
func AreaDataKeys() string {
	return fmt.Sprintf("%s:*", AreaRedisKeyPrefix)
}

// hash
func AreaPackageRedisKey(kind AreaPackageKind) string {
	return fmt.Sprintf("%s:%s:pkg", AreaRedisKeyPrefix, kind)
}

// hash
func AreaKindIdRedisKey(kind AreaPackageKind) string {
	return fmt.Sprintf("%s:%s:kindid2pkg", AreaRedisKeyPrefix, kind)
}

// set
func AreaCodeRedisKey(kind AreaPackageKind, code string) string {
	return fmt.Sprintf("%s:%s:code2pkg:%s", AreaRedisKeyPrefix, kind, code)
}

// set
func AreaCodeListRedisKey(kind AreaPackageKind) string {
	return fmt.Sprintf("%s:%s:codes", AreaRedisKeyPrefix, kind)
}

// hash
func AreaGameRuleRedisKey() string {
	return fmt.Sprintf("%s:gamerule", AreaRedisKeyPrefix)
}

// hash
func AppletAreaGameRuleRedisKey() string {
	return fmt.Sprintf("%s:gamerule_applet", AreaRedisKeyPrefix)
}

// hash
func AreaExplainRedisKey() string {
	return fmt.Sprintf("%s:explain", AreaRedisKeyPrefix)
}

// set
func AreaUniversalRedisKey(kind AreaPackageKind) string {
	return fmt.Sprintf("%s:%s:universal", AreaRedisKeyPrefix, kind)
}

// set
func AreaRecommendRedisKey(kind AreaPackageKind) string {
	return fmt.Sprintf("%s:%s:recommend", AreaRedisKeyPrefix, kind)
}

// hash
func AreaWeChatRedisKey() string {
	return fmt.Sprintf("%s:wechat", AreaRedisKeyPrefix)
}

// hash
func AppletAreaGamesRedisKey(kind AreaPackageKind) string {
	return fmt.Sprintf("%s:%s:games", AreaRedisKeyPrefix, kind)
}

// PHP 区域数据结构
// http://dqgm.facai.cn/api/game/groupCity
type AreaData struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data []*AreaDataItem `json:"data"`
}

type AreaDataItem struct {
	Code    int            `json:"code"`
	Region  string         `json:"region"`
	Package []*AreaPackage `json:"package"`
}

type AreaPackage struct {
	Id          int         `json:"id"`
	PackageKey  string      `json:"package_key"`
	PackageName string      `json:"package_name"`
	Icon        string      `json:"icon"`
	Version     string      `json:"version"`
	Recommend   int         `json:"reco"`
	IsPublic    int         `json:"is_public"`
	Country     int         `json:"country"`
	Province    int         `json:"province"`
	City        int         `json:"city"`
	PackageType string      `json:"package_type"`
	Engine      int         `json:"game_engine"`
	Games       []*AreaGame `json:"games"`
	Sort        int         `json:"sort"`
}

type AreaGame struct {
	PackageId        int    `json:"package_id"`
	KindId           int    `json:"kind_id"`
	Name             string `json:"name"`
	GameIcon         string `json:"icon"`
	GameRule         string `json:"-"`
	GameRuleVersion  int    `json:"version"`
	ClientVersion    int    `json:"client_version"`
	RecommendVersion int    `json:"recommend_version"`
	GameRuleJs       string `json:"rule_js"`
	Sort             int    `json:"sort"`
	IsVip            int    `json:"is_vip"`
}

// 小程序玩法包信息
// http://dqgm.facai.cn/api/game/xcx/groupcity
type AppletAreaData struct {
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
	Data []*AppletAreaDataItem `json:"data"`
}

type AppletAreaDataItem struct {
	Code   int         `json:"code"`
	Region string      `json:"region"`
	Games  []*AreaGame `json:"games"`
}

// 区域客户微信
// http://dqgm.facai.cn/api/kefu/wx

type AreaGmWeChat struct {
	Code int                 `json:"code"`
	Msg  string              `json:"msg"`
	Data []*AreaGmWeChatItem `json:"data"`
}

type AreaGmWeChatItem struct {
	Code           int    `json:"city"`
	CustomerWX1    string `json:"kefu1"`
	CustomerWX2    string `json:"kefu2"`
	CustomerWX3    string `json:"kefu3"`
	CustomerName1  string `json:"kefu1_name"`
	CustomerName2  string `json:"kefu2_name"`
	CustomerName3  string `json:"kefu3_name"`
	CustomerMobile string `json:"kefu_mobile"`
}

type AreaGmWeChatInfo struct {
	Area   string                  `json:"area"`
	WeChat []*AreaGmWeChatInfoItem `json:"we_chat"`
	Mobile string                  `json:"mobile"`
}

func (awx *AreaGmWeChatInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, awx)
}

func (awx *AreaGmWeChatInfo) MarshalBinary() (data []byte, err error) {
	return json.Marshal(awx)
}

type AreaGmWeChatInfoItem struct {
	Wx   string `json:"wx"`
	Name string `json:"name"`
}

type AreaGameRuleCompiled struct {
	GameRuleVersion int    `json:"game_rule_version"`
	ClientVersion   int    `json:"client_version"`
	KindId          int    `json:"kind_id"`
	Name            string `json:"name"`
	PackageKey      string `json:"package_key"`
	GameRule        string `json:"rule"`
	Engine          int    `json:"engine"`
	PackageVersion  string `json:"package_version"`
	TimeLimitFree   bool   `json:"timelimit_free"` // 显示免费
}

type AreaGameRuleCompiledList []*AreaGameRuleCompiled

func (agrcs *AreaGameRuleCompiledList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, agrcs)
}

func (agrcs *AreaGameRuleCompiledList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(agrcs)
}

func (agrc *AreaGameRuleCompiled) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, agrc)
}

func (agrc *AreaGameRuleCompiled) MarshalBinary() (data []byte, err error) {
	return json.Marshal(agrc)
}

// 区域包的集合
type AreaPkgCompiledList []*AreaPackageCompiled

// 解码
func (apcs *AreaPkgCompiledList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, apcs)
}

// 编码
func (apcs *AreaPkgCompiledList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(apcs)
}

// 长度
func (apcs *AreaPkgCompiledList) Len() int {
	return len(*apcs)
}

// 排序
func (apcs *AreaPkgCompiledList) Less(i, j int) bool {
	if (*apcs)[i].Sort == (*apcs)[j].Sort {
		if (*apcs)[i].GameType == (*apcs)[j].GameType {
			return (*apcs)[i].ID < (*apcs)[j].ID
		} else {
			return (*apcs)[i].GameType > (*apcs)[j].GameType
		}
	} else {
		return (*apcs)[i].Sort < (*apcs)[j].Sort
	}
}

// 交换
func (apcs *AreaPkgCompiledList) Swap(i, j int) {
	(*apcs)[i], (*apcs)[j] = (*apcs)[j], (*apcs)[i]
}

// 包含
func (apcs *AreaPkgCompiledList) Contains(key string) bool {
	for _, apc := range *apcs {
		if apc.PackageKey == key {
			return true
		}
	}
	return false
}

// 包含
func (apcs *AreaPkgCompiledList) GameByKid(kid int) *AreaGameCompiled {
	for _, apc := range *apcs {
		for _, game := range apc.Games {
			if game.KindId == kid {
				return game
			}
		}
	}
	return nil
}

// 搜索子包
func (apcs *AreaPkgCompiledList) Search(keyword string) AreaPkgCompiledList {
	if keyword == "" {
		return *apcs
	} else {
		result := make(AreaPkgCompiledList, 0)
		for i := 0; i < len(*apcs); i++ {
			if (*apcs)[i] == nil {
				continue
			}
			if keyword == (*apcs)[i].Code {
				result = append(result, (*apcs)[i])
			} else if keyword == (*apcs)[i].GameType.String() {
				result = append(result, (*apcs)[i])
			} else if keyString := (*apcs)[i].KeyString(); strings.Contains(keyString, keyword) {
				result = append(result, (*apcs)[i])
			} else {
				for _, game := range (*apcs)[i].Games {
					if strings.Contains(game.Name, keyword) {
						result = append(result, (*apcs)[i])
						break
					}
				}
			}
		}
		return result
	}
}

// 搜索子包
func (apcs *AreaPkgCompiledList) SearchByPType(pType AreaPackageType) AreaPkgCompiledList {
	switch pType {
	case AreaPackageTypeMahJng, AreaPackageTypePoke, AreaPackageTypePlate, AreaPackageTypeFlower:
		result := make(AreaPkgCompiledList, 0)
		for i := 0; i < len(*apcs); i++ {
			if (*apcs)[i] == nil {
				continue
			}
			if (*apcs)[i].GameType == pType {
				result = append(result, (*apcs)[i])
			}
		}
		return result
	default:
		return *apcs
	}
}

// 去重
func (apcs *AreaPkgCompiledList) UnDuplicate() {
	tempMap := make(map[string]struct{})
	for i := 0; i < len(*apcs); {
		apc := (*apcs)[i]
		if apc == nil {
			i++
			continue
		}
		_, ok := tempMap[apc.PackageKey]
		if ok {
			copy((*apcs)[i:], (*apcs)[i+1:])
			*apcs = (*apcs)[:len(*apcs)-1]
		} else {
			tempMap[apc.PackageKey] = struct{}{}
			i++
		}
	}
}

// 去重
func (apcs *AreaPkgCompiledList) ToMap() map[string]*AreaPackageCompiled {
	result := make(map[string]*AreaPackageCompiled)
	for _, apc := range *apcs {
		if apc == nil {
			continue
		}
		if _, ok := result[apc.PackageKey]; ok {
			continue
		}
		result[apc.PackageKey] = apc
	}
	return result
}

type AreaPackageCompiled struct {
	ID             int                  `json:"id"`              // id
	Region         string               `json:"region"`          // 地区名字
	Code           string               `json:"code"`            // 区域码
	City           string               `json:"city"`            // 城市代码
	Province       string               `json:"province"`        // 省级代码
	Country        string               `json:"country"`         // 县级代码
	PackageKey     string               `json:"package_key"`     // 游戏包key
	PackageName    string               `json:"package_name"`    // 游戏包名
	PackageVersion string               `json:"package_version"` // 游戏包版本
	GameType       AreaPackageType      `json:"gametype"`        // 游戏类型 // 包类型 0所有 1花牌 2字牌 3扑克 4麻将
	Engine         int                  `json:"game_engine"`     // 游戏引擎 0未知 1CocosCreator  2CocosJS
	Icon           string               `json:"icon"`            // 游戏图标
	Recommend      int                  `json:"reco"`            // 推荐开关
	Universal      int                  `json:"is_public"`       // 通用开关
	Sort           int                  `json:"sort"`            // 排序标识
	Games          AreaGameCompiledList `json:"games"`
}

// 小程序游戏包列表
type AppletAreaGamesCompiled struct {
	Games AreaGameCompiledList `json:"games"`
}

func (apc *AreaPackageCompiled) KeyString() string {
	return fmt.Sprintf("[%s][%s][%s]",
		apc.Region,
		apc.PackageKey,
		apc.PackageName,
	)
}

// 筛选出
func (apc *AreaPackageCompiled) ScreenGame(kIds ...int) {
	in := func(kid int) bool {
		for _, id := range kIds {
			if id == kid {
				return true
			}
		}
		return false
	}

	for i := 0; i < len(apc.Games); {
		game := apc.Games[i]
		if game == nil || !in(game.KindId) {
			copy(apc.Games[i:], apc.Games[i+1:])
			apc.Games = apc.Games[:len(apc.Games)-1]
		} else {
			i++
		}
	}
}

func (apc *AreaPackageCompiled) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, apc)
}

func (apc *AreaPackageCompiled) MarshalBinary() (data []byte, err error) {
	return json.Marshal(apc)
}

func (apc *AppletAreaGamesCompiled) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, apc)
}

func (apc *AppletAreaGamesCompiled) MarshalBinary() (data []byte, err error) {
	return json.Marshal(apc)
}

type AreaGameCompiledList []*AreaGameCompiled

func (agcs *AreaGameCompiledList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, agcs)
}

func (agcs *AreaGameCompiledList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(agcs)
}

// 去重
func (agcs *AreaGameCompiledList) UnDuplicate() {
	tempMap := make(map[int]struct{})
	for i := 0; i < len(*agcs); {
		apc := (*agcs)[i]
		if apc == nil {
			i++
			continue
		}
		_, ok := tempMap[apc.KindId]
		if ok {
			copy((*agcs)[i:], (*agcs)[i+1:])
			*agcs = (*agcs)[:len(*agcs)-1]
		} else {
			tempMap[apc.KindId] = struct{}{}
			i++
		}
	}
}

// 去重
func (agcs AreaGameCompiledList) ToMap() map[int]*AreaGameCompiled {
	result := make(map[int]*AreaGameCompiled)
	for _, agc := range agcs {
		if agc == nil {
			continue
		}
		if _, ok := result[agc.KindId]; ok {
			continue
		}
		result[agc.KindId] = agc
	}
	return result
}

type AreaGameCompiled struct {
	PackageKey      string `json:"package_key"`       // 游戏包key
	PackageName     string `json:"package_name"`      // 游戏包名
	PackageVersion  string `json:"package_version"`   // 游戏包版本
	Name            string `json:"name"`              // 游戏名
	Icon            string `json:"icon"`              // 游戏图标
	KindId          int    `json:"kind_id"`           // 子游戏id
	GameRuleVersion int    `json:"game_rule_version"` // 游戏规则版本
	ClientVersion   int    `json:"client_version"`    // 客户端需要的规则版本
	RecommVersion   int    `json:"recomm_version"`    // 推荐更新版本
	ForcedVersion   int    `json:"forced_version"`    // 强制更新版本
	Engine          int    `json:"game_engine"`       // 游戏引擎 0未知 1CocosCreator  2CocosJS
	TimeLimitFree   bool   `json:"timelimit_free"`    // 显示免费
	Sort            int    `json:"sort"`              // 排序标识（同一区域）
	CanVipFloor     bool   `json:"can_vip_floor"`     // 能否设置为vip玩法
}

func GenAreaDefaultGMWeChat(area string) *AreaGmWeChatInfo {
	return &AreaGmWeChatInfo{
		Area: area,
		WeChat: []*AreaGmWeChatInfoItem{
			&AreaGmWeChatInfoItem{
				Wx:   "111",
				Name: "客服1",
			},
			&AreaGmWeChatInfoItem{
				Wx:   "111",
				Name: "客服2",
			},
			&AreaGmWeChatInfoItem{
				Wx:   "111",
				Name: "客服3",
			},
		},
		Mobile: GMMobile,
	}
}

func (aw *AreaGmWeChat) Convert() map[string]*AreaGmWeChatInfo {
	result := make(map[string]*AreaGmWeChatInfo)
	for _, data := range aw.Data {
		if data == nil {
			continue
		}
		item := new(AreaGmWeChatInfo)
		item.WeChat = make([]*AreaGmWeChatInfoItem, 0)
		if data.CustomerMobile == "" {
			item.Mobile = GMMobile
		} else {
			item.Mobile = data.CustomerMobile
		}
		item.Area = fmt.Sprint(data.Code)
		item.WeChat = append(
			item.WeChat,
			&AreaGmWeChatInfoItem{Wx: data.CustomerWX1, Name: data.CustomerName1},
			&AreaGmWeChatInfoItem{Wx: data.CustomerWX2, Name: data.CustomerName2},
			&AreaGmWeChatInfoItem{Wx: data.CustomerWX3, Name: data.CustomerName3},
		)
		result[item.Area] = item
	}
	return result
}

type PackKeys []string

func (pks *PackKeys) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, pks)
}

func (pks *PackKeys) MarshalBinary() (data []byte, err error) {
	return json.Marshal(pks)
}

func (pks PackKeys) UnDuplicate() PackKeys {
	tempMap := make(map[string]struct{})
	result := make(PackKeys, 0)
	for i := 0; i < len(pks); i++ {
		p := pks[i]
		_, ok := tempMap[p]
		if ok {
			continue
		} else {
			tempMap[p] = struct{}{}
			result = append(result, p)
		}
	}
	return result
}

// 筛选掉
func (pks PackKeys) WeedOut(codes ...string) PackKeys {
	result := make(PackKeys, 0)
	in := func(code string) bool {
		for _, c := range codes {
			if fmt.Sprint(c) == code {
				return true
			}
		}
		return false
	}

	for _, p := range pks {
		if !in(p) {
			result = append(result, p)
		}
	}
	return result
}

func (pks *PackKeys) Convert() []interface{} {
	result := make([]interface{}, 0)
	for _, k := range *pks {
		result = append(result, k)
	}
	return result
}

// 新版本取消对CocosCreator的支持
func (ad *AreaData) Convert() (
	map[string]interface{}, // pkgKey - pkgInfo
	map[string]interface{}, // kindId - gameRule
	map[string]*PackKeys, // area - pkgKeys
	map[string]interface{}, // kindId - pkgKey
	*PackKeys, // recommend pkg keys
	*PackKeys, // public pkg keys
	*PackKeys, // codes
) {
	key2info := make(map[string]interface{})
	kid2rule := make(map[string]interface{})
	area2key := make(map[string]*PackKeys)
	kid2key := make(map[string]interface{})
	recommend := make(PackKeys, 0)
	universal := make(PackKeys, 0)
	codes := make(PackKeys, 0)
	for _, data := range ad.Data {
		if data == nil {
			continue
		}
		areaCode := fmt.Sprint(data.Code)
		codes = append(codes, areaCode)
		pkgKeys := make(PackKeys, 0)
		for _, pkg := range data.Package {
			if pkg == nil {
				continue
			}
			engine := HF_CvtGameEngine(pkg.Engine)
			if engine != consts.EngineCocosJs {
				continue
			}

			// 区域包
			pkgKeys = append(pkgKeys, pkg.PackageKey)

			// 通用包
			if pkg.IsPublic != 0 {
				universal = append(universal, pkg.PackageKey)
			}

			// 推荐包
			if pkg.Recommend != 0 {
				recommend = append(recommend, pkg.PackageKey)
			}

			apc := &AreaPackageCompiled{
				ID:             pkg.Id,
				Code:           areaCode,
				Region:         data.Region,
				PackageKey:     pkg.PackageKey,
				PackageName:    pkg.PackageName,
				PackageVersion: pkg.Version,
				GameType:       GetAreaPackageType(pkg.PackageType),
				Icon:           pkg.Icon,
				Recommend:      pkg.Recommend,
				Universal:      pkg.IsPublic,
				Engine:         engine,
				Province:       fmt.Sprint(pkg.Province),
				City:           fmt.Sprint(pkg.City),
				Country:        fmt.Sprint(pkg.Country),
				Sort:           pkg.Sort,
			}

			for _, game := range pkg.Games {
				if game == nil {
					continue
				}
				kid2key[fmt.Sprint(game.KindId)] = pkg.PackageKey

				kid2rule[fmt.Sprint(game.KindId)] = &AreaGameRuleCompiled{
					GameRuleVersion: HF_MaxInt(game.GameRuleVersion, game.RecommendVersion),
					KindId:          game.KindId,
					Name:            game.Name,
					PackageKey:      pkg.PackageKey,
					GameRule:        game.GameRuleJs,
					Engine:          engine,
					ClientVersion:   game.ClientVersion,
				}

				agc := &AreaGameCompiled{
					PackageKey:      pkg.PackageKey,
					PackageName:     pkg.PackageName,
					PackageVersion:  pkg.Version,
					Name:            game.Name,
					Icon:            game.GameIcon,
					KindId:          game.KindId,
					ClientVersion:   game.ClientVersion,
					RecommVersion:   game.RecommendVersion,
					ForcedVersion:   game.GameRuleVersion,
					Engine:          engine,
					GameRuleVersion: HF_MaxInt(game.GameRuleVersion, game.RecommendVersion),
					Sort:            game.Sort,
					CanVipFloor:     game.IsVip > 0,
				}
				apc.Games = append(apc.Games, agc)
			}
			key2info[apc.PackageKey] = apc
		}
		area2key[fmt.Sprint(data.Code)] = &pkgKeys
	}
	return key2info, kid2rule, area2key, kid2key, &recommend, &universal, &codes
}

// 小程序区域数据
func (ad *AppletAreaData) Convert() (
	map[string]interface{}, // gamesInfo
	map[string]interface{}, // kindId - gameRule
	*PackKeys, // codes
) {
	key2info := make(map[string]interface{})
	kid2rule := make(map[string]interface{})
	codes := make(PackKeys, 0)
	for _, data := range ad.Data {
		if data == nil {
			continue
		}

		// 区域码
		codes = append(codes, fmt.Sprint(data.Code))

		// 区域对应的游戏列表
		agc := new(AppletAreaGamesCompiled)

		// 游戏包
		for _, game := range data.Games {
			if game == nil {
				continue
			}

			// 规则
			kid2rule[fmt.Sprint(game.KindId)] = &AreaGameRuleCompiled{
				GameRuleVersion: HF_MaxInt(game.GameRuleVersion, game.RecommendVersion),
				KindId:          game.KindId,
				Name:            game.Name,
				GameRule:        game.GameRuleJs,
				Engine:          1,
				ClientVersion:   game.ClientVersion,
			}

			val := &AreaGameCompiled{
				Name:            game.Name,
				Icon:            game.GameIcon,
				KindId:          game.KindId,
				ClientVersion:   game.ClientVersion,
				RecommVersion:   game.RecommendVersion,
				ForcedVersion:   game.GameRuleVersion,
				Engine:          1,
				GameRuleVersion: HF_MaxInt(game.GameRuleVersion, game.RecommendVersion),
				Sort:            game.Sort,
				CanVipFloor:     game.IsVip > 0,
			}
			agc.Games = append(agc.Games, val)
		}

		// 保存此区域下的游戏包
		key2info[fmt.Sprint(data.Code)] = agc
	}

	return key2info, kid2rule, &codes
}

type AreaPackageSeek struct {
	Keyword     string              `json:"keyword"`
	Type        AreaSeekType        `json:"type"`
	AreaCode    string              `json:"code"`
	PackageType AreaPackageType     `json:"package_type"` // 包类型 0所有 1花牌 2字牌 3扑克 4麻将
	Packages    AreaPkgCompiledList `json:"packages"`
}

type MsgAreaExplain struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data []*AreaExplain `json:"data"`
}

type AreaExplain struct {
	KindId  int    `json:"kind_id"`
	Explain string `json:"explain"`
}

func (ae *AreaExplain) MarshalBinary() (data []byte, err error) {
	return json.Marshal(ae)
}

func (ae *AreaExplain) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, ae)
}

func (mse *MsgAreaExplain) ToMap() map[string]interface{} {
	res := make(map[string]interface{})
	for _, es := range mse.Data {
		res[fmt.Sprint(es.KindId)] = es
	}
	return res
}
