package notice_update

import (
	"fmt"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

func NoticeUpdate(header string, data interface{}) (interface{}, *xerrors.XError) { // 公告更新
	msgdata, ok := data.(*static.Msg_NoticeUpdate)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	// 校验公告类型
	if msgdata.PositionType != consts.NoticePositionTypeDialog && msgdata.PositionType != consts.NoticePositionTypeMaintain && msgdata.PositionType != consts.NoticePositionTypeMarquee &&
		msgdata.PositionType != consts.NoticePositionTypeOption {
		xlog.Logger().Errorf("公告类型不合法：%+v", msgdata)
		return nil, xerrors.NewXError("公告类型不合法")
	}

	server2.GetNoticeMgr().Update()

	if msgdata.PositionType == consts.NoticePositionTypeOption {
		switch msgdata.Kind {
		case consts.NoticeOptionKind_UpdDeliveryInfo:
			result, err := server2.GetServer().CallHall("ServerMethod.ServerMsg", consts.MsgTypeDeliveryInfoUpd, &msgdata)
			if err != nil {
				xlog.Logger().Infoln("玩家头像更新:", string(result), "err:", err)
				return nil, xerrors.NewXError(fmt.Sprintf("update userImg error:%v", err))
			}
			return "true", xerrors.RespOk
		case consts.NoticeOptionKind_UpdKindPanel:
			server2.GetAreaMgr().UpdateArea()
		}

	}

	// 通知大厅服更新
	_, _ = server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeNoticeUpdate, &msgdata)
	return "true", xerrors.RespOk
}

func ForceHotter(header string, data interface{}) (interface{}, *xerrors.XError) {
	// 通知客户端
	value, err := server2.GetServer().CallHall("NewServerMsg", header, data)
	if err != nil {
		xlog.Logger().Errorln("ForceHotter reset failed:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	strVal := fmt.Sprintf("%s", value)
	if strVal != "true" && strVal != "SUC" {
		return value, xerrors.NewXError(strVal)
	}
	return value, xerrors.RespOk
}
