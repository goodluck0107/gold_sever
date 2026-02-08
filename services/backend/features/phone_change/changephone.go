package phone_change

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
)

func ChangePhone(header string, data interface{}) (interface{}, *xerrors.XError) { // 用户付款成功后下发商品
	_, ok := data.(*static.GmChangeUserPhone)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgUserPhoneChange, data)
	if err != nil {
		xlog.Logger().Errorln("change user phone failed:", err.Error())
		return nil, xerrors.ServerMaintainError
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.ArgumentError
	}
	return value, xerrors.RespOk
}
