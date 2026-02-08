package config_reload

import (
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
	"strings"
)

func ReloadConfig(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_AssIgnReLoad)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知登录服
	value, err := server2.GetServer().CallLogin("NewServerMsg", consts.MsgTypeReloadConfig, &msgdata)
	if err != nil {
		return value, xerrors.ServerMaintainError
	}
	server2.GetAreaMgr().Update()
	return value, xerrors.RespOk
}

func SetLogFileLevel(header string, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.MsgSetLogFileLevel)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	var (
		value interface{}
		err   error
	)

	if strings.Contains(req.Servers, "admin") {
		xlog.SetFileLevel(req.Level)
	}

	if strings.Contains(req.Servers, "login") {
		value, err = server2.GetServer().CallLogin("NewServerMsg", consts.MsgTypeSetLogFileLevel, req)
		if err != nil {
			return value, xerrors.ServerMaintainError
		}
	}

	if strings.Contains(req.Servers, "hall") || req.Game != 0 {
		value, err = server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeSetLogFileLevel, req)
		if err != nil {
			return value, xerrors.ServerMaintainError
		}
	}

	value = "SUC"
	return value, xerrors.RespOk
}
