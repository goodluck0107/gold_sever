package backend

import (
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/services/backend/features/config_reload"
	"github.com/open-source/game/chess.git/services/backend/features/game_server"
	"github.com/open-source/game/chess.git/services/backend/features/house_message"
	"github.com/open-source/game/chess.git/services/backend/features/house_revoke"
	"github.com/open-source/game/chess.git/services/backend/features/league"
	"github.com/open-source/game/chess.git/services/backend/features/notice_update"
	"github.com/open-source/game/chess.git/services/backend/features/phone_change"
	"github.com/open-source/game/chess.git/services/backend/features/product_deliver"
	"github.com/open-source/game/chess.git/services/backend/features/write_off_user"
)

var httpRouter router.HttpWorker

//InitRouter 初始化路由，需要添加新接口要在这里增加路由
func InitRouter() *router.HttpWorker {
	// 游戏服务相关
	httpRouter.Router(consts.MsgTypeHouseConUpdate, static.Msg_UpdHouseCon{}, gameserver.HouseConUpdate)
	httpRouter.Router(consts.MsgTypeClearUserInfo, static.Msg_ClearUserInfo{}, gameserver.ClearUserInfo)
	httpRouter.Router(consts.MsgTypeCompulsoryDiss, static.Msg_UserSeat{}, gameserver.CompulsoryDiss)
	httpRouter.Router(consts.MsgTypeUpdateGameServer, static.Msg_UpdateGameServer{}, gameserver.UpdateGameServer)
	httpRouter.Router(consts.MsgTypeSaveToDatabase, static.MsgDataID{}, gameserver.SaveToDatabase)
	httpRouter.Router(consts.MsgTypeSaveToRedis, static.MsgDataID{}, gameserver.SaveToRedis)
	httpRouter.Router(consts.MsgTypeSetBlack, static.Msg_SetBlack{}, gameserver.SetBlack)
	httpRouter.Router(consts.MsgTypeGetOnlineNumber, static.Msg_Null{}, gameserver.GetOnlineNumber)
	httpRouter.Router(consts.MsgTypeGetUserInfo, static.Msg_GetUserInfo{}, gameserver.GetUserInfo)
	httpRouter.Router(consts.MsgTypeAreaUpdate, static.Msg_AreaUpdate{}, gameserver.AreaUpdate)
	httpRouter.Router(consts.MsgTypeExecStatisticsTask, static.Msg_Null{}, gameserver.ExecStatisticsTask)
	httpRouter.Router(consts.MsgTypeExecGameStatisticsTask, static.Msg_Null{}, gameserver.ExecGameStatisticsTask)
	httpRouter.Router(consts.MsgTypeHosueIDChange, static.Msg_HosueIDChange{}, gameserver.HosueIDChange)
	httpRouter.Router(consts.MsgTypeDeliveryInfoUpd, static.Msg_GetUserInfo{}, gameserver.DeliveryInfoUpdate)

	// 加盟商相关，待分类
	httpRouter.Router(consts.MsgTypeAddLeague, static.MsgAddLeague{}, league.AddLeague)
	httpRouter.Router(consts.MsgTypeLeagueFreeze, static.MsgFreezeLeague{}, league.LeageFreeze)
	httpRouter.Router(consts.MsgTypeLeagueUnFreeze, static.MsgUnFreezeLeague{}, league.LeagueUnFreeze)
	httpRouter.Router(consts.MsgTypeAddLeagueUser, static.MsgAddLeageUser{}, league.AddLeagueUser)
	httpRouter.Router(consts.MsgTypeLeagueUserFreeze, static.MsgFreezeLeagueUser{}, league.LeagueUserFreeze)
	httpRouter.Router(consts.MsgTypeLeagueUserUnFreeze, static.MsgUnFreezeLeagueUser{}, league.LeagueUserUnFreeze)

	// 用户财富划扣
	httpRouter.Router(consts.MsgTypeDeliverProduct, static.Msg_DeliverProduct{}, product_deliver.DeliverProduct)
	// 待分类
	httpRouter.Router(consts.MsgTypeNoticeUpdate, static.Msg_NoticeUpdate{}, notice_update.NoticeUpdate)
	httpRouter.Router(consts.MsgTypeExplainUpdate, static.Msg_ExplainUpdate{}, gameserver.ExplainUpdate)
	httpRouter.Router(consts.MsgTypeReloadConfig, static.Msg_AssIgnReLoad{}, config_reload.ReloadConfig)
	httpRouter.Router(consts.MsgTypeGetValidKindid, static.Msg_Null{}, gameserver.GetValidKindid)
	httpRouter.Router(consts.MsgUserPhoneChange, static.GmChangeUserPhone{}, phone_change.ChangePhone)
	httpRouter.Router(consts.MsgHouseRevokeGm, static.GmHouseRevoke{}, house_revoke.HouseRevoke)
	httpRouter.Router(consts.MsgHouseMemKick, static.AdminKickMem{}, house_message.HouseMsg)
	httpRouter.Router(consts.MsgTypeSetLogFileLevel, static.MsgSetLogFileLevel{}, config_reload.SetLogFileLevel)

	// house
	httpRouter.Router(consts.MsgHouseOwnerChange, static.MsgHouseOwnerChange{}, house_message.HouseChangeOwner)
	// house
	httpRouter.Router(consts.MsgGameSwitch, static.MsgHouseGameSwitch{}, house_message.HouseGameSwitch)
	httpRouter.Router(consts.MsgUserAgentUpdate, static.Msg_GetUserInfo{}, house_message.UserAgentUpdate)
	httpRouter.Router(consts.MsgChangeBlankUser, static.MsgChangeBlankUser{}, house_message.AddBlankUser)
	httpRouter.Router(consts.MsgChangeBlankHouse, static.MsgChangeBlankUser{}, house_message.AddBlankUser)
	httpRouter.Router(consts.MsgHouseTableReset, static.Msg_Http_HouseTableReset{}, house_message.HouseReset)
	httpRouter.Router(consts.MsgResetHmUserRight, static.Msg_Http_ResetUserRight{}, house_message.ResetUserHmRight)
	httpRouter.Router(consts.MsgForceHotter, static.Msg_Http_ForceHotter{}, notice_update.ForceHotter)

	httpRouter.Router(consts.MsgWriteOffUser, static.Msg_Http_WriteOffUser{}, write_off_user.WriteOffUser)
	httpRouter.Router(consts.MsgAddHotVersion, static.Msg_Userid{}, gameserver.AddHotVersion)

	return &httpRouter

}
