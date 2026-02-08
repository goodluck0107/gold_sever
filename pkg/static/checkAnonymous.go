package static

import (
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/pkg/errors"
)

// 通过规则玩法检测 是否 是匿名游戏
func IsAnonymous(rule string) bool {
	var gameConfig map[string]interface{}
	gameConfig = make(map[string]interface{})
	err := json.Unmarshal([]byte(rule), &gameConfig)
	if err != nil {
		xlog.Logger().Errorf("查找匿名面板数据时,游戏玩法解析失败")
		return false
	}
	key := "anonymity"
	b, ok := gameConfig[key]
	if !ok {
		return false
	}
	return b.(string) == "true"
}

// 通过规则玩法检测 是否 开启旁观
func IsLookOnSupport(rule string) (bo bool, cer error) {
	var gameConfig map[string]interface{}
	gameConfig = make(map[string]interface{})
	err := json.Unmarshal([]byte(rule), &gameConfig)
	if err != nil {
		xlog.Logger().Errorf("查找旁观面板数据时,游戏玩法解析失败")
		return false, err
	}
	key := "LookonSupport"
	b, ok := gameConfig[key]
	if !ok {
		return false, errors.New("面板没有当前字段 LookonSupport")
	}
	return b.(string) == "true", nil
}
