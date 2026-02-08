package wuhan

import (
	"encoding/json"
	"fmt"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"

	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/pkg/errors"
)

var areamgrsingleton *AreaMgr = nil

type AreaMgr struct {
	mu *lock2.RWMutex
}

func GetAreaMgr() *AreaMgr {
	if areamgrsingleton == nil {
		areamgrsingleton = new(AreaMgr)
		areamgrsingleton.mu = new(lock2.RWMutex)
	}
	return areamgrsingleton
}

func (am *AreaMgr) Update() {
	// 获取函数执行时间
	defer static.HF_FuncElapsedTime()()
	am.clean()
	am.UpdateArea()
	am.UpdateExplainById(0)
}

func (am *AreaMgr) UpdateArea() {
	// 防止并发写入
	am.mu.Lock()
	defer am.mu.Unlock()
	err := am.StoreData(am.GetData())
	if err != nil {
		xlog.Logger().Panic("area redis store data error:", err)
	}
	xlog.Logger().Info("load area data succeed.")
}

// 新增接口 根据游戏id获取玩法说明介绍
func (am *AreaMgr) UpdateExplainById(kid int) {
	// 防止并发写入
	am.mu.Lock()
	defer am.mu.Unlock()
	url := am.explainUrlWhitId(kid)
	xlog.Logger().Info("explainUrl:", url)
	data, err := util.HttpGet(url, nil)
	if err != nil {
		xlog.Logger().Error("http get explain error:", err)
		return
	}
	res := new(static.MsgAreaExplain)
	err = json.Unmarshal(data, res)
	if err != nil {
		xlog.Logger().Error("json unmarshal explain error:", err)
		return
	}

	if res.Code != 0 {
		xlog.Logger().Error("php explain response error:", res.Msg)
		return
	}

	// 如果是复写整个key
	if kid == 0 {
		if GetDBMgr().GetDBrControl().RedisV2.Exists(static.AreaExplainRedisKey()).Val() == 1 {
			if err := GetDBMgr().GetDBrControl().RedisV2.Del(static.AreaExplainRedisKey()).Err(); err != nil {
				xlog.Logger().Error("redis del explain data error", err)
			} else {
				xlog.Logger().Info("del explain data from redis succeed")
			}
		}
	}

	err = GetDBMgr().GetDBrControl().AreaKidToExplainSave(res.ToMap())
	if err != nil {
		xlog.Logger().Error("redis save explain data error:", err)
		return
	}
	xlog.Logger().Info("redis save explain data succeed. kind id:", kid)
}

func (am *AreaMgr) url() (pkgUrl, goldPkgUrl, wxUrl, appletUrl string) {
	pkgUrl = fmt.Sprintf("%s/api/game/groupCity?game_type=%d", GetServer().Con.AdminHost, static.AreaPackageKindCard)
	goldPkgUrl = fmt.Sprintf("%s/api/game/groupCity?game_type=%d", GetServer().Con.AdminHost, static.AreaPackageKindGold)
	appletUrl = fmt.Sprintf("%s/api/game/xcx/groupCity", GetServer().Con.AdminHost)
	wxUrl = fmt.Sprintf("%s/api/kefu/wx", GetServer().Con.AdminHost)
	return
}

func (am *AreaMgr) explainUrlWhitId(id int) string {
	return fmt.Sprintf("%s/api/game/explain?kind_id=%d", GetServer().Con.AdminHost, id)
}

// 得到数据
func (am *AreaMgr) GetData() (pkgData *static.AreaData /*房卡游戏*/, goldPkgData *static.AreaData /*金币游戏*/, gamesData *static.AppletAreaData /*小程序游戏*/, wx map[string]*static.AreaGmWeChatInfo) {
	areaUrl, goldUrl, wxUrl, appletUrl := am.url()

	xlog.Logger().Info("areaUrl:", areaUrl)
	xlog.Logger().Info("goldUrl:", goldUrl)
	xlog.Logger().Info("wxUrl:", wxUrl)
	xlog.Logger().Info("appletUrl:", appletUrl)

	var err error
	pkgData, err = am.getAreaData(areaUrl)
	if err != nil {
		xlog.Logger().Panic("get area pkg data error:", err)
	}

	goldPkgData, err = am.getAreaData(goldUrl)
	if err != nil {
		xlog.Logger().Panic("get gold area pkg data error:", err)
	}

	//gamesData, err = am.getAppletData(appletUrl)
	//if err != nil {
	//	xlog.Logger().Panic("get applet area pkg data error:", err)
	//}

	wx, err = am.getWeChatData(wxUrl)
	if err != nil {
		// 这里取不到微信就取不到  后面用默认值处理
		xlog.Logger().Error("get area WeChat data error:", err)
	}

	return pkgData, goldPkgData, gamesData, wx
}

// 储存数据
func (am *AreaMgr) StoreData(pkgData, goldPkgData *static.AreaData, gamesData *static.AppletAreaData, wx map[string]*static.AreaGmWeChatInfo) error {

	err := am.storeCardArea(pkgData)
	if err != nil {
		return err
	}

	err = am.storeGoldArea(goldPkgData)
	if err != nil {
		return err
	}

	//err = am.storeAppletArea(gamesData)
	//if err != nil {
	//	return err
	//}

	am.storeWeChat(wx)

	return nil
}

func (am *AreaMgr) getAreaData(url string) (*static.AreaData, error) {
	data, err := util.HttpGet(url, nil)
	if err != nil {
		return nil, err
	}
	result := new(static.AreaData)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}
	return result, nil
}

func (am *AreaMgr) getAppletData(url string) (*static.AppletAreaData, error) {
	data, err := util.HttpGet(url, nil)
	if err != nil {
		return nil, err
	}
	result := new(static.AppletAreaData)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}
	return result, nil
}

func (am *AreaMgr) getWeChatData(url string) (map[string]*static.AreaGmWeChatInfo, error) {
	data, err := util.HttpGet(url, nil)
	if err != nil {
		return nil, err
	}
	result := new(static.AreaGmWeChat)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}
	return result.Convert(), nil
}

func (am *AreaMgr) storeCardArea(pkgData *static.AreaData) error {
	// 新版本取消对CocosCreator的支持
	pkgInfo, rule, areaMap, keyMap, rs, us, codes := pkgData.Convert()
	function := func(errs ...error) error {
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
		return nil
	}
	return function(
		// hash package key to package eve
		GetDBMgr().GetDBrControl().AreaPackageSave(static.AreaPackageRedisKey(static.AreaPackageKindCard), pkgInfo),
		// hash game kind id to game rule
		GetDBMgr().GetDBrControl().AreaKidToRuleSave(static.AreaGameRuleRedisKey(), rule),
		// hash game kind id to package key
		GetDBMgr().GetDBrControl().AreaKidToPackageKeySave(static.AreaKindIdRedisKey(static.AreaPackageKindCard), keyMap),
		// set code contains package key
		GetDBMgr().GetDBrControl().AreaCodeToPkgsSave(static.AreaPackageKindCard, areaMap),
		GetDBMgr().GetDBrControl().AreaPackageRecommendSave(static.AreaPackageKindCard, rs.Convert()),
		GetDBMgr().GetDBrControl().AreaPackageUniversalSave(static.AreaPackageKindCard, us.Convert()),
		GetDBMgr().GetDBrControl().AreaPackageCodeListSave(static.AreaCodeListRedisKey(static.AreaPackageKindCard), codes.Convert()),
	)
}

func (am *AreaMgr) storeGoldArea(pkgData *static.AreaData) error {
	// 新版本取消对CocosCreator的支持
	pkgInfo, _, areaMap, keyMap, rs, us, codes := pkgData.Convert()
	function := func(errs ...error) error {
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
		return nil
	}
	return function(
		// hash package key to package eve
		GetDBMgr().GetDBrControl().AreaPackageSave(static.AreaPackageRedisKey(static.AreaPackageKindGold), pkgInfo),
		// // hash game kind id to game rule
		// GetDBMgr().GetDBrControl().AreaKidToRuleSave(rule), // 金币场不用存rule
		// hash game kind id to package key
		GetDBMgr().GetDBrControl().AreaKidToPackageKeySave(static.AreaKindIdRedisKey(static.AreaPackageKindGold), keyMap),
		// set code contains package key
		GetDBMgr().GetDBrControl().AreaCodeToPkgsSave(static.AreaPackageKindGold, areaMap),
		GetDBMgr().GetDBrControl().AreaPackageRecommendSave(static.AreaPackageKindGold, rs.Convert()),
		GetDBMgr().GetDBrControl().AreaPackageUniversalSave(static.AreaPackageKindGold, us.Convert()),
		GetDBMgr().GetDBrControl().AreaPackageCodeListSave(static.AreaCodeListRedisKey(static.AreaPackageKindGold), codes.Convert()), // 金币场暂时不用有哪些区域
	)
}

func (am *AreaMgr) storeAppletArea(gamesData *static.AppletAreaData) error {
	// 新版本取消对CocosCreator的支持
	gamesInfo, rule, codes := gamesData.Convert()
	function := func(errs ...error) error {
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
		return nil
	}
	return function(
		// hash code to game eve
		GetDBMgr().GetDBrControl().AreaCodeToGamesSave(static.AppletAreaGamesRedisKey(static.AreaPackageKindApplet), gamesInfo),
		// hash game kind id to game rule
		GetDBMgr().GetDBrControl().AreaKidToRuleSave(static.AppletAreaGameRuleRedisKey(), rule),
		// set codes
		GetDBMgr().GetDBrControl().AreaPackageCodeListSave(static.AreaCodeListRedisKey(static.AreaPackageKindApplet), codes.Convert()),
	)
}

func (am *AreaMgr) storeWeChat(wx map[string]*static.AreaGmWeChatInfo) {
	if wx == nil {
		return
	}
	mp := make(map[string]interface{})
	for k, info := range wx {
		if info == nil {
			continue
		}
		mp[k] = info
	}
	if len(mp) > 0 {
		if err := GetDBMgr().GetDBrControl().RedisV2.HMSet(static.AreaWeChatRedisKey(), mp).Err(); err != nil {
			xlog.Logger().Error("hmset area wechat eve to redis error:", err)
		}
	}
	return
}

func (am *AreaMgr) clean() {
	keys, _ := GetDBMgr().GetDBrControl().RedisV2.Keys(static.AreaDataKeys()).Result()
	if len(keys) > 0 {
		if err := GetDBMgr().GetDBrControl().RedisV2.Del(keys...).Err(); err != nil {
			xlog.Logger().Error("clean area data error:", err)
		}
	}
}
