package gameserver

import (
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
	"strings"
)

//UpdateGameServer 更新游戏服务器
func UpdateGameServer(header string, data interface{}) (interface{}, *xerrors.XError) { // 游戏服状态更新
	msgdata, ok := data.(*static.Msg_UpdateGameServer)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 校验参数
	if msgdata.GameId <= 0 {
		return nil, xerrors.ArgumentError
	}
	if msgdata.Status != consts.ServerStatusMaintain {
		return nil, xerrors.ArgumentError
	}
	// 通知大厅服维护游戏服
	server2.GetServer().CallHall("NewServerMsg", "regameserver", &static.Msg_GameServer{Id: msgdata.GameId, Status: msgdata.Status})
	return "", xerrors.RespOk
}

//GetValidKindid 获取游戏id
func GetValidKindid(header string, data interface{}) (interface{}, *xerrors.XError) {
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeGetValidKindid, data)
	result := string(value)
	if result != "false" && err == nil {
		return result, xerrors.RespOk
	}
	return nil, xerrors.NewXError(fmt.Sprintf("err:%s;result:%s...", err.Error(), result))

}

//HouseConUpdate 更新房间
func HouseConUpdate(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_UpdHouseCon)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	if err := server2.GetDBMgr().GetDBmControl().Model(models.ConfigHouse{}).Update(msgdata.Column, msgdata.Value).Error; err != nil {
		xlog.Logger().Errorln("update ConfigHouse failed: ", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	//！让所有服务器重读config
	server2.GetServer().CallLogin("NewServerMsg", consts.MsgTypeReloadConfig, &static.Msg_AssIgnReLoad{
		Server: 2,
		Games:  []int{},
	})
	return msgdata.Column, xerrors.RespOk
}

//ClearUserInfo 清除用户信息
func ClearUserInfo(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_ClearUserInfo)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	tx := server2.GetDBMgr().GetDBmControl().Begin()
	defer tx.Rollback()
	var err error
	if msgdata.ClearColumn == "Id" { //！如果要清空的字段为id则删除一个
		if err = tx.Where("id = ?", msgdata.ClearUid).Delete(models.User{}).Error; err != nil {
			xlog.Logger().Errorln("mysql del user failed: ", err.Error())
			return nil, xerrors.NewXError(err.Error())
		}
		if ok, err := server2.GetDBMgr().GetDBrControl().DelPerson(msgdata.ClearUid); err != nil || !ok {
			xlog.Logger().Errorln("redis del user failed: ")
			return nil, xerrors.NewXError("redis del user failed: ")
		}
	} else {
		// 参数校验
		if msgdata.ClearColumn != "Tel" && msgdata.ClearColumn != "ReName" && msgdata.ClearColumn != "Idcard" && msgdata.ClearColumn != "MachineCode" {
			return nil, xerrors.ArgumentError
		}

		column := strings.ToLower(msgdata.ClearColumn)
		if err = tx.Model(models.User{}).Where("id = ?", msgdata.ClearUid).Update(column, "").Error; err != nil {
			xlog.Logger().Errorln("mysql clear user failed: ", err.Error())
			return nil, xerrors.NewXError(err.Error())
		}
		person := new(static.Person)
		person, err = server2.GetDBMgr().GetDBrControl().GetPerson(msgdata.ClearUid)
		if person == nil || err != nil {
			return nil, xerrors.NewXError("redis中找不到该玩家")
		}
		if err = static.Setter(person, msgdata.ClearColumn, nil); err != nil {
			xlog.Logger().Errorln("set user zore fail", err.Error())
			return nil, xerrors.NewXError(err.Error())
		}
		if err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, msgdata.ClearColumn, ""); err != nil {
			xlog.Logger().Errorln("redis clear user failed: ", err.Error())
			return nil, xerrors.NewXError(err.Error())
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
	}
	return msgdata.ClearColumn, xerrors.RespOk
}

// CompulsoryDiss 强制解散桌子
func CompulsoryDiss(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_UserSeat)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeCompulsoryDiss, &msgdata)
	if err != nil {
		//syslog.Logger().Errorln("hall return from websocket err:", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	//syslog.Logger().Errorln("hall return from websocket :", string(value))
	result := string(value)
	if result == "true" || result == "OK" {
		return result, xerrors.RespOk
	}
	return result, xerrors.NewXError(result)
}

//SaveToDatabase 同步数据到数据库
func SaveToDatabase(header string, data interface{}) (interface{}, *xerrors.XError) {
	if server2.GetServer().DataSynchronism {
		return nil, xerrors.NewXError("正在执行，请勿重复操作")
	}
	msgdata, ok := data.(*static.MsgDataID)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	err := server2.GetServer().SaveRedisToDBData()
	if err != nil {
		fmt.Println(err)
		err = service2.GetAdminClient().WriteBackResult(msgdata.Id, -1)
		if err != nil {
			xlog.Logger().Errorln("redistodb_writeback error: ", err.Error())
		}
		return nil, xerrors.NewXError(fmt.Sprintf("redistodb_writeback error:%s ", err.Error()))
	}
	err = service2.GetAdminClient().WriteBackResult(msgdata.Id, 1)
	if err != nil {
		xlog.Logger().Errorln("redistodb_writeback error: ", err.Error())
		return nil, xerrors.NewXError(fmt.Sprintf("redistodb_writeback error:%s ", err.Error()))
	}
	return "", xerrors.RespOk
}

func SaveToRedis(header string, data interface{}) (interface{}, *xerrors.XError) {
	if server2.GetServer().DataSynchronism {
		return nil, xerrors.NewXError("正在执行，请勿重复操作")
	}
	msgdata, ok := data.(*static.MsgDataID)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	err := server2.GetServer().SaveDBDataToRedis()
	if err != nil {
		fmt.Println(err)
		err = service2.GetAdminClient().WriteBackResult(msgdata.Id, -1)
		if err != nil {
			xlog.Logger().Errorln("dbtoredis_writeback error: ", err.Error())
		}
		return nil, xerrors.NewXError(fmt.Sprintf("dbtoredis_writeback error:%s ", err.Error()))
	}
	err = service2.GetAdminClient().WriteBackResult(msgdata.Id, 1)
	if err != nil {
		xlog.Logger().Errorln("dbtoredis_writeback error: ", err.Error())
		return nil, xerrors.NewXError(fmt.Sprintf("dbtoredis_writeback error:%s ", err.Error()))
	}
	return "", xerrors.RespOk
}

//SetBlack 设置黑名单
func SetBlack(header string, data interface{}) (interface{}, *xerrors.XError) {
	//! 设置用户黑名单
	msgdata, ok := data.(*static.Msg_SetBlack)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 校验参数
	if msgdata.Uid <= 0 {
		return nil, xerrors.ArgumentError
	}
	var user models.User
	user.Id = msgdata.Uid
	if err := server2.GetDBMgr().GetDBmControl().First(&user).Error; err != nil {
		return nil, xerrors.UserNotExistError
	}
	// 更新用户数据
	if err := server2.GetDBMgr().GetDBmControl().Model(&user).Update("is_black", msgdata.IsBlack).Error; err != nil {
		xlog.Logger().Errorln("database update failed: ", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	if err := server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(user.Id, "IsBlack", msgdata.IsBlack); err != nil {
		xlog.Logger().Errorln("redis update failed: ", err.Error())
		return nil, xerrors.NewXError(err.Error())
	}
	return "", xerrors.RespOk
}

// GetOnlineNumber 获取在线人数
func GetOnlineNumber(header string, data interface{}) (interface{}, *xerrors.XError) {
	// 获取当前在线人数
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeGetOnlineNumber, nil)
	if err != nil {
		xlog.Logger().Errorln("get online number from hall failed:", err.Error())
		return nil, xerrors.NewXError("无法连接到大厅服")
	}

	return value, xerrors.RespOk
}

// GetUserInfo 获取用户详情
func GetUserInfo(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_GetUserInfo)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 校验参数
	if msgdata.Uid <= 0 {
		return nil, xerrors.ArgumentError
	}
	// 获取用户详情
	p, err := server2.GetDBMgr().GetDBrControl().GetPerson(msgdata.Uid)
	if err != nil {
		xlog.Logger().Errorln("get person from redis error: ", err.Error())
		return nil, xerrors.UserNotExistError
	}

	return getPersonInfo(p), xerrors.RespOk
}

//AreaUpdate 区域列表下的子游戏更新, 需要大厅重新同步区域数据, 并下发通知到客户端
func AreaUpdate(header string, data interface{}) (interface{}, *xerrors.XError) {
	// msgdata, ok := data.(*public.Msg_AreaUpdate)
	// if !ok {
	// 	return nil, xerrors.ArgumentError
	// }
	// // 通知大厅
	// _, err := wuhan.GetServer().CallHall("NewServerMsg", constant.MsgTypeAreaUpdate, msgdata)
	// if err != nil {
	// 	syslog.Logger().Errorln("area update failed:", err.Error())
	// 	return nil, xerrors.ServerMaintainError
	// }
	server2.GetAreaMgr().UpdateArea()
	return "", xerrors.RespOk
}

func ExplainUpdate(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_ExplainUpdate)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	server2.GetAreaMgr().UpdateExplainById(msgdata.KindId)
	return "", xerrors.RespOk
}

//ExecStatisticsTask 执行统计数据任务
func ExecStatisticsTask(header string, data interface{}) (interface{}, *xerrors.XError) {
	server2.GetServer().WriteStatisticsData()
	return "", xerrors.RespOk
}

//ExecStatisticsTask 执行统计数据任务
func ExecGameStatisticsTask(header string, data interface{}) (interface{}, *xerrors.XError) {
	server2.GetServer().WriteGameStatisticsData()
	return "", xerrors.RespOk
}

//HosueIDChange  修改包厢id
func HosueIDChange(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_HosueIDChange)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeHosueIDChange, msgdata)
	if err != nil {
		xlog.Logger().Errorln("get online number from hall failed:", err.Error())
		return nil, xerrors.ServerMaintainError
	}
	if fmt.Sprintf("%s", value) != "true" {
		return value, xerrors.ArgumentError
	}
	return value, xerrors.RespOk
}

// 个人名片更新
func DeliveryInfoUpdate(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_GetUserInfo)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	value, err := server2.GetServer().CallHall("ServerMethod.ServerMsg", consts.MsgTypeDeliveryInfoUpd, msgdata)
	if err != nil {
		xlog.Logger().Errorln("get online number from hall failed:", err.Error())
		return nil, xerrors.ServerMaintainError
	}

	return value, xerrors.RespOk
}

func AddHotVersion(header string, data interface{}) (interface{}, *xerrors.XError) {
	msgdata, ok := data.(*static.Msg_Userid)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, err := server2.GetDBMgr().GetDBrControl().GetPerson(msgdata.Uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, xerrors.NewXError(err.Error())
	}
	if p == nil {
		xlog.Logger().Errorln("person is nil")
		return nil, xerrors.NewXError("person is nil")
	}
	err = server2.GetDBMgr().GetDBrControl().HotVersionRecord(fmt.Sprint(p.Uid))
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, xerrors.NewXError(err.Error())
	}
	return "", xerrors.RespOk
}
