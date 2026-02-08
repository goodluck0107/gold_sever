package api

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

type httpApi struct {
	pool map[string]*router.AppWorker
}

func NewHttpApi() *httpApi {
	return &httpApi{}
}

// 初始化HTTP服务
func (hs *httpApi) OnInit() {
	hs.pool = make(map[string]*router.AppWorker)
	// 获取区域游戏规则面板数据
	hs.Register(consts.MsgTypeAreaGameRules, static.Msg_CH_GameRules{}, AreaGameRules)
	hs.Register(consts.MsgTypeAreaGameExplain, static.Msg_ExplainUpdate{}, AreaGameExplain)
	// ...
}

func (hs *httpApi) Register(header string, proto interface{}, appHandle router.AppHandlerFunc) {
	hs.pool[header] = &router.AppWorker{DataType: reflect.TypeOf(proto), Handle: appHandle}
}

func (hs *httpApi) AppHandler(handler string) *router.AppWorker {
	return hs.pool[handler]
}

func (hs *httpApi) EncodeInfo() (encode int, key string) {
	return GetServer().Con.Encode, GetServer().Con.EncodeClientKey
}

// 根据kindId获取规则
func AreaGameRules(request *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_CH_GameRules)

	if !ok {
		return nil, xerrors.ArgumentError
	}

	if req.KindIds == nil || len(req.KindIds) == 0 {
		return nil, xerrors.ArgumentError
	}

	fields := make([]string, 0)

	for _, kindId := range req.KindIds {
		fields = append(fields, static.HF_Itoa(kindId))
	}

	ack := new(static.Msg_HC_GameRules)

	cmd := new(redis.SliceCmd)
	if req.Channel == consts.ChannelApplet {
		cmd = GetDBMgr().GetDBrControl().RedisV2.HMGet(static.AppletAreaGameRuleRedisKey(), fields...)
	} else {
		cmd = GetDBMgr().GetDBrControl().RedisV2.HMGet(static.AreaGameRuleRedisKey(), fields...)
	}

	err := redis.NewStringSliceResult(static.RedisSwitchObjectsToStrings(cmd.Result())).ScanSlice(&ack.Games)

	if err != nil {
		xlog.Logger().Errorln("get game rule form redis error:", err)
	}

	// 获取限免游戏
	freeGameMap := GetServer().GetLimitFreeGameKindIds()

	for _, game := range ack.Games {
		if free, ok := freeGameMap[game.KindId]; ok {
			game.TimeLimitFree = free
		} else {
			game.TimeLimitFree = false
		}
	}

	return &ack, nil
}

func AreaGameExplain(request *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_ExplainUpdate)

	if !ok {
		return nil, xerrors.ArgumentError
	}

	ack := new(static.AreaExplain)
	err := GetDBMgr().GetDBrControl().RedisV2.HGet(static.AreaExplainRedisKey(), fmt.Sprint(req.KindId)).Scan(ack)

	if err != nil {
		xlog.Logger().Error("get area game explain error:", err)
		// return nil, xerrors.DBExecError
	}

	ack.KindId = req.KindId

	return &ack, nil
}
