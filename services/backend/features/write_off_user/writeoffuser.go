package write_off_user

import (
	"fmt"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
)

func WriteOffUser(header string, data interface{}) (interface{}, *xerrors.XError) {
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", header, data)
	if err != nil {
		xlog.Logger().Errorln("WriteOffUser failed:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}
