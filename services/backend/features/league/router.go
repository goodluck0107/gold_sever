package league

import (
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
)

// 添加加盟商
func AddLeague(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.MsgAddLeague)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	if msgdata.LID <= 0 || msgdata.AreaCode <= 0 {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeAddLeague, msgdata)
	if err != nil {
		return value, xerrors.ServerMaintainError
	}
	return value, xerrors.RespOk
}

func LeageFreeze(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.MsgFreezeLeague)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeLeagueFreeze, msgdata)
	if err != nil {
		return nil, xerrors.ServerMaintainError
	}
	return value, xerrors.RespOk
}

func LeagueUnFreeze(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.MsgUnFreezeLeague)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeLeagueUnFreeze, msgdata)
	if err != nil {
		return nil, xerrors.ServerMaintainError
	}
	return value, xerrors.RespOk
}

func AddLeagueUser(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.MsgAddLeageUser)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	if msgdata.LID <= 0 || msgdata.Uid <= 0 {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeAddLeagueUser, msgdata)
	if err != nil {
		// return nil, fmt.Errorf("add league user error:%s", err.Error())
		return nil, xerrors.ServerMaintainError
	}
	return value, xerrors.RespOk
}

func LeagueUserFreeze(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.MsgFreezeLeagueUser)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeLeagueUserFreeze, msgdata)
	if err != nil {
		xlog.Logger().Errorln("freeze league user error:", err.Error())
		// return nil, fmt.Errorf("freeze league user error:%s", err.Error())
		return nil, xerrors.ServerMaintainError
	}
	return value, xerrors.RespOk
}

func LeagueUserUnFreeze(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.MsgUnFreezeLeagueUser)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeLeagueUserUnFreeze, msgdata)
	if err != nil {
		xlog.Logger().Errorln("un freeze league user error:", err.Error())
		// return nil, fmt.Errorf("unfreeze league user error:%s", err.Error())
		return nil, xerrors.ServerMaintainError
	}
	return value, xerrors.RespOk
}
