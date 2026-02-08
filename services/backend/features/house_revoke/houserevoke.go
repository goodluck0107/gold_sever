package house_revoke

import (
	"fmt"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

func HouseRevoke(header string, data interface{}) (interface{}, *xerrors.XError) { // 接触合并包厢
	fmt.Println("11111")
	_, ok := data.(*static.GmHouseRevoke)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgHouseRevokeGm, data)
	if err != nil {
		xlog.Logger().Errorln("house revoke failed:", err.Error())
		return nil, xerrors.ServerMaintainError
	}
	strVal := fmt.Sprintf("%s", value)
	fmt.Println(strVal)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}
