package house_message

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
)

func HouseMsg(header string, data interface{}) (interface{}, *xerrors.XError) {
	_, ok := data.(*static.AdminKickMem)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgHouseMemKick, data)
	if err != nil {
		xlog.Logger().Errorln("house revoke failed:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}

func HouseChangeOwner(header string, data interface{}) (interface{}, *xerrors.XError) {
	_, ok := data.(*static.MsgHouseOwnerChange)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgHouseOwnerChange, data)
	if err != nil {
		xlog.Logger().Errorln("house revoke failed:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}

func HouseGameSwitch(header string, data interface{}) (interface{}, *xerrors.XError) {
	_, ok := data.(*static.MsgHouseGameSwitch)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgGameSwitch, data)
	if err != nil {
		xlog.Logger().Errorln("house revoke failed:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}

func UserAgentUpdate(header string, data interface{}) (interface{}, *xerrors.XError) {
	_, ok := data.(*static.Msg_GetUserInfo)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgUserAgentUpdate, data)
	if err != nil {
		xlog.Logger().Errorln("UserAgentUpdate RPC failed:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}
