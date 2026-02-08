package center

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// 包厢成员审核设置
func Proto_ClubHouseOptMemCheck(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsMemCheck)

	house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorJoinReviewed)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionMemCheck(req.IsChecked, p.Uid)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsChecked {
		info = "开启"
	} else {
		info = "关闭"
	}
	if mem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了包厢审核", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了包厢审核", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, HouseCheckChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢成员退圈开关设置
func Proto_ClubHouseOptMemExit(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsMemExit)

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorOutReviewed)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionIsMemberExit(req.IsMemExit, p.Uid)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	var info string
	if req.IsMemExit {
		info = "开启"
	} else {
		info = "关闭"
	}

	msg := fmt.Sprintf("盟主%s了退出包厢开关", info)
	CreateClubMassage(house.DBClub.Id, p.Uid, HouseCheckChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢成员审核设置
func Proto_ClubHouseOptParnterMemCheck(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsMemCheck)

	house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionParnterMemCheck(req.IsChecked, p.Uid)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsChecked {
		info = "开启"
	} else {
		info = "关闭"
	}
	if mem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了包厢队长审核", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了包厢队长审核", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, HouseCheckChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢设置冻结
func Proto_ClubHouseOptFrozen(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsFrozen)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorBanTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if mem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}

	if req.IsFrozen {
		// for _, floor := range house.Floors {
		// 	for _, t := range floor.Tables {
		// 		if t.TId != 0 {
		// 			return xerrors.ExistHouseTablePlayingError.Code, xerrors.ExistHouseTablePlayingError.Msg
		// 		}
		// 	}
		// }
		house.DBClub.IsFrozen = true
	} else {
		// 如果是解冻，则判断是否存在合并包厢的请求
		hmls := make([]*models.HouseMergeLog, 0)
		err := GetDBMgr().GetDBmControl().Where("swallowed = ?", house.DBClub.Id).Where("merge_state >= ?", models.HouseMergeStateWaiting).Find(&hmls).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
		} else {
			if len(hmls) > 0 {
				return xerrors.HouseUnFrozenError.Code, xerrors.HouseUnFrozenError.Msg
			}
		}
	}
	custerr := house.OptionFrozen(req.IsFrozen)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsFrozen {
		info = "冻结"
	} else {
		info = "解冻"
	}
	if mem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了隐私设置", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了隐私设置", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, FrozenChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢成员列表
func Proto_ClubHouseOptMemHide(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsMemHide)

	house, _, hmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetPrivacy)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionMemHide(req.IsHide)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsHide {
		info = "开启"
	} else {
		info = "关闭"
	}
	if hmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了隐私设置", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了隐私设置", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, PrivacyChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢包厢号隐藏
func Proto_ClubHouseOptHidHide(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsHidHide)

	house, _, hmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetPrivacy)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	custerr := house.OptionHidHide(req.IsHide)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsHide {
		info = "开启"
	} else {
		info = "关闭"
	}
	if hmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了隐藏包厢号", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了隐藏包厢号", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, PrivacyChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢包厢号隐藏
func Proto_ClubHouseOptHeadHide(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsHeadHide)

	house, _, hmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetPrivacy)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionHeadHide(req.IsHide)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsHide {
		info = "开启"
	} else {
		info = "关闭"
	}
	if hmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了隐藏大厅玩家头像", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了隐藏大厅玩家头像", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, PrivacyChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢包厢号隐藏
func Proto_ClubHouseOptMemUidHide(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsMemUidHide)

	house, _, hmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetPrivacy)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionUidHide(req.IsHide)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsHide {
		info = "开启"
	} else {
		info = "关闭"
	}
	if hmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了隐藏大厅玩家ID", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了隐藏大厅玩家ID", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, PrivacyChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢在线人数隐藏
func Proto_ClubHouseOptOnlineHide(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptIsOnlineHide)

	house, _, hmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetPrivacy)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionOnlineHide(req.IsHide)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.IsHide {
		info = "开启"
	} else {
		info = "关闭"
	}
	if hmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了隐藏在线人数/开局桌数", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了隐藏在线人数/开局桌数", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, PrivacyChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢在线人数隐藏
func Proto_ClubHouseOptPartnerKick(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseOptPartnerKick)

	house, _, hmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.OptionPartnerKick(req.PartnerKick)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	var info string
	if req.PartnerKick {
		info = "开启"
	} else {
		info = "关闭"
	}
	if hmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主%s了队长踢人", info)
	} else {
		msg = fmt.Sprintf("%sID:%d%s了队长踢人", p.Nickname, p.Uid, info)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, PrivacyChange, msg)
	return xerrors.SuccessCode, nil
}

// 根据uid查询包厢牌桌详情
func Proto_ClubHouseTableInfoByUid(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.MsgC2SHidUid)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	houseTable := house.GetTableByUid(req.Uid)
	if houseTable == nil {
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}

	// 获取房间信息
	table := houseTable.GetTableInstance()
	if table == nil {
		// 牌桌不存在
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}

	// 直接从大厅获取牌桌信息
	result := new(static.Msg_S2C_TableInfo)
	result.Hid = table.HId
	result.Fid = table.FId
	result.TId = table.Id
	result.NTId = table.NTId
	result.RoundNum = table.Config.RoundNum
	result.MaxPlayerNum = table.Config.MaxPlayerNum
	result.CurrentRound = table.Step
	result.Person = make([]*static.Msg_S2C_TablePerson, 0)
	for _, u := range table.Users {
		if u != nil {
			// 从内存获取在线用户信息
			p, err := GetDBMgr().GetDBrControl().GetPerson(u.Uid)
			if err != nil {
				xlog.Logger().Errorln("user does not exist: ", u.Uid)
			} else {
				mem := house.GetMemByUId(p.Uid)
				tablePerson := &static.Msg_S2C_TablePerson{
					Id:       p.Uid,
					Imgurl:   p.Imgurl,
					Nickname: p.Nickname,
					Ip:       p.Ip,
					Online:   p.Online,
				}
				//
				if mem != nil {
					// tablePerson.Vitamin = public.SwitchVitaminToF64(mem.UVitamin)
					if mem.Partner > 0 && mem.Partner != 1 {
						tablePerson.Partner = fmt.Sprint(mem.Partner)
					}
				}
				result.Person = append(result.Person, tablePerson)
			}
		}
	}
	return xerrors.SuccessCode, &result
}

// 创建包厢
func Proto_ClubHouseCreate(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseCreate)
	//if !GetServer().ConHouse.IsPerControl {
	//	// 权限判定
	//	league := GetAllianceMgr().GetUserLeagueInfo(p.Uid)
	//	if league == nil || league.Freeze {
	//		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//	}
	//}
	_, xerr := GetDBMgr().GetUserAgentConfig(p.Uid)
	if xerr != nil {
		return xerr.Code, xerr.Msg
	}
	// 敏感词判断
	if static.IsContainSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.HName) {
		custerr := xerrors.NewXError("包厢名称违规，请重新输入")
		return custerr.Code, custerr.Msg
	}

	// 创建数量判定
	count, err := GetDBMgr().GetDBrControl().HouseMemberCreateCounts(p.Uid)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if count >= GetServer().ConHouse.CreateMax {
		return xerrors.HouseMaxOverFlowError.Code, xerrors.HouseMaxOverFlowError.Msg
	}

	// 房卡判定
	if p.Card-p.FrozenCard < GetServer().ConHouse.CardCost {
		custerr := xerrors.NewXError(
			fmt.Sprintf("很抱歉，您的房卡不足%v张，无法创建包厢",
				GetServer().ConHouse.CardCost))
		return custerr.Code, custerr.Msg
	}
	// 创建包厢
	house, custerr := GetClubMgr().HouseCreate(req, p.Uid)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	// 结果返回
	var ack static.Msg_S2C_HouseCreate
	ack.Id = house.DBClub.Id
	ack.HId = house.DBClub.HId

	return xerrors.SuccessCode, &ack
}

// 包厢删除
func Proto_ClubHouseDelete(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	// 数据结构
	req := data.(*static.Msg_CH_HouseDelete)

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if !house.IsNormalNoMerge() {
		if house.IsBusyMerging() {
			return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
		}

		if house.IsMerged() {
			xlog.Logger().Error("包厢合并了其他包厢不能解散")
			return xerrors.HouseDissolveError.Code, xerrors.HouseDissolveError.Msg
		}

		if house.IsBeenMerged() {
			xlog.Logger().Error("包厢被其他包厢合并了不能解散")
			return xerrors.HouseDissolveError.Code, xerrors.HouseDissolveError.Msg
		}
	}

	hmls := make([]*models.HouseMergeLog, 0)
	if err := GetDBMgr().GetDBmControl().Where("swallowed = ?", house.DBClub.Id).Where("merge_state >= ?", models.HouseMergeStateWaiting).Find(&hmls).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else {
		if len(hmls) > 0 {
			xerr := xerrors.NewXError("合并包厢中不能解散。")
			return xerr.Code, xerr.Msg
		}
	}

	if house.ExistActivity() {
		return xerrors.HouseActivityExistError.Code, xerrors.HouseActivityExistError.Msg
	}

	custerr := GetClubMgr().HouseDelete(house)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	return xerrors.SuccessCode, nil
}

// 包厢楼层创建
func Proto_ClubHouseFloorCreate(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseFloorCreate)

	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	// 获取玩法配置
	config := GetServer().GetGameConfig(req.FRule.KindId)
	if config == nil {
		// 没有找到玩法对应的配置
		cuserror := xerrors.NewXError("不支持的游戏玩法")
		return cuserror.Code, cuserror.Msg
	}

	// 判断是否有包厢是否有该玩法的维护权及版本效验
	if cuserror := house.IsMaintainableGame(req.FRule.KindId, req.FRule.Version); cuserror != nil {
		return cuserror.Code, cuserror.Msg
	}

	area_game := GetAreaGameByKid(req.FRule.KindId)
	if area_game == nil {
		// 没有找到玩法对应的区域游戏信息
		cuserror := xerrors.NewXError("该玩法已下线")
		return cuserror.Code, cuserror.Msg
	}

	// 如果客户端发送的规则版本于当前规则版本不一样 则提示玩法已变更
	if req.FRule.Version != area_game.GameRuleVersion {
		cuserror := xerrors.NewXError("规则版本不匹配，请重试。")
		return cuserror.Code, cuserror.Msg
	}

	// 保存强更和弱更版本
	req.FRule.Version = area_game.ForcedVersion
	req.FRule.RecommendVersion = area_game.RecommVersion

	// 创建楼层
	housefloor, custerr := house.FloorCreate(p.Uid, req.FRule)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	// 返回结果
	var ack static.Msg_HC_HouseFloorCreate
	ack.Id = housefloor.Id
	var pkgName string
	var kindName string
	area := GetAreaGameByKid(req.FRule.KindId)
	if area != nil {
		pkgName = area.PackageName
		kindName = area.Name
	}
	msg := fmt.Sprintf("盟主创建了包厢%s-%s(%d)玩法", kindName, pkgName, housefloor.Id)
	CreateClubMassage(house.DBClub.Id, p.Uid, CreateFloor, msg)

	// 增添加删除记录
	go AddPartnerRoyaltyModifyHistory(house.DBClub.Id, mem, optTypeAddFloor, housefloor.GetWanFaName(), housefloor.Id, house.GetFloorIndexByFid(housefloor.Id), 0, 0, 0)

	return xerrors.SuccessCode, &ack
}

// 包厢楼层删除
func Proto_ClubHouseFloorDelete(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	// 数据结构
	req := data.(*static.Msg_CH_HouseFloorDelete)

	house, floor, optMem, cusErr := inspectClubFloorMemberWithRight(req.HId, req.FId, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if floor.ExistActivityNoFinish() {
		return xerrors.HouseActivityExistError.Code, xerrors.HouseActivityExistError.Msg
	}
	floor.DataLock.Lock()
	defer floor.DataLock.Unlock()
	// 判定是否存在正在游戏牌桌
	for _, t := range floor.Tables {
		if t.TId != 0 {
			xlog.Logger().Errorf("删除楼层失败：包厢id:%d,楼层id:%d,delete 存在正在游戏的牌桌: %+v",
				house.DBClub.HId,
				floor.Id,
				t,
			)
			table := t.GetTableInstance()
			if table != nil {
				data, err := GetDBMgr().GetDBrControl().Get(table.GetRedisKey())
				redisTable := new(static.Table)
				if err == nil {
					err = json.Unmarshal(data, redisTable)
				}
				xlog.Logger().Errorf("存在的桌子信息:%+v, Redis桌子:%+v, redis err:%s", table.Table, redisTable, err)
			}
			return xerrors.ExistHouseTablePlayingError.Code, xerrors.ExistHouseTablePlayingError.Msg
		}
	}
	floor.IsAlive = false

	var pkgName string
	var kindName string
	area := GetAreaGameByKid(floor.Rule.KindId)
	if area != nil {
		pkgName = area.PackageName
		kindName = area.Name
	}
	msg := fmt.Sprintf("盟主删除了包厢%s-%s(%d)玩法", pkgName, kindName, floor.Id)
	go CreateClubMassage(house.DBClub.Id, p.Uid, DeleteFloor, msg)

	// 增加删除记录
	go CreateClubFloorDelMsg(house.DBClub.Id, floor.Id, house.GetFloorIndexByFid(floor.Id))

	// 增添加删除记录
	go AddPartnerRoyaltyModifyHistory(house.DBClub.Id, optMem, optTypeDelFloor, floor.GetWanFaName(), floor.Id, house.GetFloorIndexByFid(floor.Id), 0, 0, 0)

	floor.Delete(nil)

	return xerrors.SuccessCode, nil
}

// 包厢楼层列表
func Proto_ClubHouseFloorList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseFloorList)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	var ack static.Msg_HC_HouseFloorList
	ack.Items = make([]static.Msg_HouseFloorMiniInfo, 0)

	house.FloorLock.RLockWithLog()

	var arr []int64
	for _, f := range house.Floors {
		arr = append(arr, f.Id)
	}
	sort.Sort(util.Int64Slice(arr))
	for _, fid := range arr {
		f := house.Floors[fid]
		var item static.Msg_HouseFloorMiniInfo
		item.FId = f.Id
		item.FRule = f.Rule
		item.Name = f.Name
		item.TableNum = len(f.Tables)
		item.TableDefault = house.DBClub.MixTableNum
		item.HideImg = f.IsHide
		item.MinTable = f.MinTable
		item.MaxTable = f.MaxTable
		item.VitaminLowLimit = static.SwitchVitaminToF64(f.VitaminLowLimit)
		item.VitaminHighLimit = static.SwitchVitaminToF64(f.VitaminHighLimit)
		if house.DBClub.MixActive {
			item.IsMix = f.IsMix
		} else {
			item.IsMix = false
		}
		item.IsVip = f.IsVip

		areaGame := GetAreaGameByKid(f.Rule.KindId)
		if areaGame != nil {
			item.ImageUrl = areaGame.Icon
			item.KindName = areaGame.Name
			item.PackageName = areaGame.PackageName
		}
		ack.Items = append(ack.Items, item)
	}

	house.FloorLock.RUnlock()

	return xerrors.SuccessCode, &ack
}

// 包厢成员列表
func Proto_ClubHouseMemberList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMemList)
	// V2.22 由于申请列表逻辑更改，后面将与housememberlist接口脱离关系，这里重定向接口至包厢申请列表接口
	// 以便后面更好的维护各业务逻辑，同时不需要修改housememberlist接口老代码（冗余）。
	if req.Role == consts.ROLE_APLLY {
		return Proto_ClubHouseApplyInfo(s, p, &static.Msg_CH_HouseApplyInfo{
			HId:      req.HId,
			Join:     true,
			Exit:     true,
			Param:    req.Param,
			PBegin:   req.PBegin,
			PEnd:     req.PEnd,
			Role:     req.Role,
			SortType: req.SortType,
		})
	}
	reqBegin := req.PBegin
	reqEnd := req.PEnd
	if req.PBegin < 0 {
		req.PBegin = 0
	}
	// 不包含end位，需要多取一位
	req.PEnd++
	if req.PEnd < req.PBegin {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 1")
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 2")
	if house.DBClub.IsMemHide && optMem.Lower(consts.ROLE_ADMIN) && !optMem.IsPartner() && !optMem.IsVicePartner() {
		return xerrors.HouseMemHideError.Code, xerrors.HouseMemHideError.Msg
	}

	// 条件切片
	var hmListArr []HouseMember
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 3")
	houseMemberMap, CreaterArr, OnLineArr, OffLineArr, OnLineArrParnter, OffLineArrParnter, OnLineArrAdmin, OffLineArrAdmin, _, BlackArr := house.GetAllMemberWithClassify()
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 4")
	partnerAttrMap, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id)
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 5")
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if req.Role == consts.ROLE_MEMBER { //成员列表
		if req.SortType == static.SORT_VITAMIN_DES || req.SortType == static.SORT_VITAMIN_AES {
			//xlog.Logger().Warn("Proto_ClubHouseMemberList 6")
			mems := house.GetAllMem()
			//xlog.Logger().Warn("Proto_ClubHouseMemberList 7")
			hmListArr = mems
		} else {
			// 创建者-> 管理员-> 成员 -- 时间顺序
			hmListArr = append(hmListArr, CreaterArr...)

			hmListArr = append(hmListArr, OnLineArrAdmin...)
			hmListArr = append(hmListArr, OffLineArrAdmin...)

			hmListArr = append(hmListArr, OnLineArrParnter...)
			hmListArr = append(hmListArr, OffLineArrParnter...)

			hmListArr = append(hmListArr, OnLineArr...)
			hmListArr = append(hmListArr, OffLineArr...)
		}
	} else if req.Role == consts.ROLE_BLACK { //黑名单
		hmListArr = BlackArr
	} else {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	limitUserId := house.GetLimitUsers()
	//根据listType获取页签对应数据 	//var conditionArr []HouseMember
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 8")
	conditionArr, allNum, onlineNum, jyNum := getHmListDataByTypeAndRole(hmListArr, limitUserId, req.ListType, optMem)
	if req.Param != "" {
		xlog.Logger().Warn("Proto_ClubHouseMemberList 9")
		conditionArr, allNum, onlineNum, jyNum = searchHmListData(hmListArr, limitUserId, req.Param, req.ListType, optMem)
	}
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 10")
	var ack static.Msg_HC_HouseMemList
	ack.PBegin = reqBegin
	ack.PEnd = reqEnd

	ack.Totalnum = allNum

	ack.HMemNum = allNum
	ack.HMemOnLineNum = onlineNum

	ack.PartnerMemsNum = allNum
	ack.PartnerMemsOnlineNum = onlineNum

	ack.LimitUserNum = jyNum

	if optMem.IsPartner() || optMem.IsVicePartner() {
		//这里先只做 team 的排除
	} else {
		if house.IsHideOnlineNum(p.Uid) {
			ack.HMemNum = -1
			ack.HMemOnLineNum = -1
		}
	}

	// 分页超出范围
	if len(conditionArr) == 0 || len(conditionArr) < req.PBegin {
		return xerrors.SuccessCode, &ack
	}

	//count := len(conditionArr)
	//for i := 0; i < count; i++ {
	//	hmem := conditionArr[i]
	//	_, err := GetDBMgr().GetDBrControl().GetPlayer(hmem.UId)
	//	if err != nil {
	//		syslog.Logger().Errorf("user not exists in house:%d,uid:%d", house.DBClub.Id, hmem.UId)
	//		house.MemDelete(hmem.UId, true, nil)
	//		continue
	//	}
	//	if cli.SIsMember(fmt.Sprintf("user_limit_game_%d", house.DBClub.Id), hmem.UId).Val() {
	//		conditionArr = append(conditionArr, hmem)
	//		conditionArr = append(conditionArr[:i], conditionArr[i+1:]...)
	//		i = i - 1
	//		count = count - 1
	//	}
	//}

	// 开始索引
	var idxBeg int
	var idxEnd int
	idxBeg = req.PBegin
	if req.PEnd > len(conditionArr) {
		idxEnd = len(conditionArr)
	} else {
		idxEnd = req.PEnd
	}
	conditionArr = conditionArr[idxBeg:idxEnd]
	xlog.Logger().Warn("Proto_ClubHouseMemberList 11")
	ack.FMems = clubHouseMemberList(conditionArr, house, houseMemberMap, partnerAttrMap, req.Role, req.SortType, optMem)
	//xlog.Logger().Warn("Proto_ClubHouseMemberList 12")
	return xerrors.SuccessCode, &ack
}

// 加入包厢
func Proto_ClubHouseMemberJoin(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemberJoin)

	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	// 合并包厢后通过老圈号加入
	if house.IsBeenMerged() {
		mergeHouse := GetClubMgr().GetClubHouseById(house.DBClub.MergeHId)
		if mergeHouse == nil {
			return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
		}
		if mergeHouse.IsBusyMerging() {
			return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
		}

		var existOld, existNew bool
		// 已加入直接进入
		hmem := mergeHouse.GetMemByUId(p.Uid)
		if hmem != nil {
			if hmem.Upper(consts.ROLE_APLLY) {
				existNew = true
			}
		}

		omem := house.GetMemByUId(p.Uid)
		if omem != nil {
			if omem.Upper(consts.ROLE_APLLY) {
				existOld = true
			}
		}

		if existNew {
			if !existOld {
				custerr := house.MemJoin(p.Uid, consts.ROLE_MEMBER, 0, true, nil)
				if custerr != nil {
					return custerr.Code, custerr.Msg
				}
				if req.InviteUid > 0 && house.IsPartner(req.InviteUid) {
					newHm := house.GetMemByUId(p.Uid)
					if newHm != nil {
						newHm.Partner = req.InviteUid
						if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
							Where("hid = ? and uid = ?", house.DBClub.Id, newHm.UId).Update("partner", req.InviteUid).Error; err != nil {
							xlog.Logger().Error(err)
							return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
						}
						newHm.Flush()
					}
				}
			}
			return xerrors.SuccessCode, nil
		} else {
			// 人数上限
			if mergeHouse.GetMemCounts() >= GetServer().ConHouse.MemMax {
				return xerrors.HouseMemJoinMaxError.Code, xerrors.HouseMemJoinMaxError.Msg
			}
			if !mergeHouse.DBClub.IsChecked {
				tx := GetDBMgr().GetDBmControl().Begin()
				custerr := house.MemJoin(p.Uid, consts.ROLE_MEMBER, 0, true, tx)
				if custerr != nil {
					tx.Rollback()
					return custerr.Code, custerr.Msg
				}
				if req.InviteUid > 0 && house.IsPartner(req.InviteUid) {
					newHm := house.GetMemByUId(p.Uid)
					if newHm != nil {
						newHm.Partner = req.InviteUid
						if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
							Where("hid = ? and uid = ?", house.DBClub.Id, newHm.UId).Update("partner", req.InviteUid).Error; err != nil {
							xlog.Logger().Error(err)
							return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
						}
						newHm.Flush()
					}
				}
				custerr = mergeHouse.MemJoin(p.Uid, consts.ROLE_MEMBER, house.DBClub.Id, false, tx)
				if custerr != nil {
					tx.Rollback()
					return custerr.Code, custerr.Msg
				}
				err := tx.Commit().Error
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
				msg1 := fmt.Sprintf("<color=#00A70C>%sID:%d</color>无审核加入包厢", p.Nickname, p.Uid)
				msg2 := fmt.Sprintf("<color=#00A70C>%sID:%d</color>无审核加入包厢(合并包厢<color=#00A70C>ID:%d</color>)", p.Nickname, p.Uid, mergeHouse.DBClub.HId)
				CreateClubMassage(mergeHouse.DBClub.Id, p.Uid, JoinHouse, msg1)
				CreateClubMassage(house.DBClub.Id, p.Uid, JoinHouse, msg2)
			} else {
				// 审核
				tx := GetDBMgr().GetDBmControl().Begin()
				if !existOld {
					custerr := house.MemJoin(p.Uid, consts.ROLE_APLLY, 0, true, tx)
					if custerr != nil {
						tx.Rollback()
						return custerr.Code, custerr.Msg
					}
					if req.InviteUid > 0 && house.IsPartner(req.InviteUid) {
						newHm := house.GetMemByUId(p.Uid)
						if newHm != nil {
							newHm.Partner = req.InviteUid
							if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
								Where("hid = ? and uid = ?", house.DBClub.Id, newHm.UId).Update("partner", req.InviteUid).Error; err != nil {
								xlog.Logger().Error(err)
								return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
							}
							newHm.Flush()
						}
					}
				}
				// 审核
				custerr := mergeHouse.MemJoin(p.Uid, consts.ROLE_APLLY, house.DBClub.Id, false, tx)
				if custerr != nil {
					tx.Rollback()
					return custerr.Code, custerr.Msg
				}

				err := tx.Commit().Error
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}

				mergeHouse.CustomBroadcast(consts.ROLE_ADMIN, true, false, false, consts.MsgTypeHouseMemberApply_Ntf, static.Msg_HouseMenberApplyNTF{HID: house.DBClub.HId})
				return xerrors.InReviewError.Code, xerrors.InReviewError.Msg
			}
		}
	} else {
		// 已经存在 则直接加入
		hmem := house.GetMemByUId(p.Uid)
		if hmem != nil {
			if hmem.Upper(consts.ROLE_APLLY) {
				return xerrors.SuccessCode, nil
			}
		}
		if house.DBClub.ApplySwitch && req.InviteUid <= 0 {
			return xerrors.NewXError("本包厢拒绝加入包厢申请").Code, xerrors.NewXError("本圈拒绝入圈申请").Msg
		}
		// 入楼审核
		ischeck := house.DBClub.IsChecked
		if ischeck {
			// 审核
			custerr := house.MemJoin(p.Uid, consts.ROLE_APLLY, 0, true, nil)
			if custerr != nil {
				return custerr.Code, custerr.Msg
			}
			// 更新邀请用户
			newInviteUid := int64(0)
			if req.InviteUid > 0 && house.IsPartner(req.InviteUid) {
				newInviteUid = req.InviteUid
			}
			newHm := house.GetMemByUId(p.Uid)
			if newHm != nil {
				newHm.Partner = newInviteUid
				if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
					Where("hid = ? and uid = ?", house.DBClub.Id, newHm.UId).Update("partner", newInviteUid).Error; err != nil {
					xlog.Logger().Error(err)
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
				newHm.Flush()
				house.Broadcast2Parnter(req.InviteUid, true, consts.MsgTypeHouseMemberApply_Ntf, static.Msg_HouseMenberApplyNTF{HID: house.DBClub.HId})
			}

			house.CustomBroadcast(consts.ROLE_ADMIN, false, false, false, consts.MsgTypeHouseMemberApply_Ntf, static.Msg_HouseMenberApplyNTF{HID: house.DBClub.HId})

			return xerrors.InReviewError.Code, xerrors.InReviewError.Msg
		} else {
			// 加入
			custerr := house.MemJoin(p.Uid, consts.ROLE_MEMBER, 0, true, nil)
			if custerr != nil {
				return custerr.Code, custerr.Msg
			}
			// 更新邀请用户
			newInviteUid := int64(0)
			if req.InviteUid > 0 && house.IsPartner(req.InviteUid) {
				newInviteUid = req.InviteUid
			}
			newHm := house.GetMemByUId(p.Uid)
			if newHm != nil {
				newHm.Partner = newInviteUid
				if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
					Where("hid = ? and uid = ?", house.DBClub.Id, newHm.UId).Update("partner", newInviteUid).Error; err != nil {
					xlog.Logger().Error(err)
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
				newHm.Flush()
			}
			msg := fmt.Sprintf("<color=#00A70C>%sID:%d</color>无审核加入包厢", p.Nickname, p.Uid)
			CreateClubMassage(house.DBClub.Id, p.Uid, JoinHouse, msg)
		}
	}
	return xerrors.SuccessCode, nil
}

// 包厢成员过审
func Proto_ClubHouseMemberAgree(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemberAgree)

	cer := HouseMemAgreeHandle(req, p.Uid, p.Nickname)
	if cer != nil {
		return cer.Code, cer.Msg
	}
	return xerrors.SuccessCode, nil
}

// 包厢成员拒审
func Proto_ClubHouseMemberRefused(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemberRefused)
	cer := HouseMemRefuseHandle(req, p.Uid, p.Nickname)
	if cer != nil {
		return cer.Code, cer.Msg
	}
	return xerrors.SuccessCode, nil
}

// 包厢成员退出
func Proto_ClubHouseMemberExit(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemberExit)
	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	// 成员退出
	custerr := house.MemExit(p.Uid)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	// 清空禁止娱乐状态
	err := house.LimitUserGame(p.Uid, p.Uid, true)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	_, e := GetMRightMgr().deleteRightByHidUid(int(house.DBClub.Id), p.Uid)
	if e != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, nil
}

// 包厢成员剔除
func Proto_ClubHouseMemberkick(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemberKick)
	if req.UId == p.Uid {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	custerr := house.MemKick(p.Uid, req.UId)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	// 清空禁止娱乐状态
	err := house.LimitUserGame(p.Uid, req.UId, true)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, nil
}

// ! 玩家黑名单加入
func Proto_ClubHouseMemberBlacklistInsert(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseMemberBlacklistInsert)

	if req.UId == p.Uid {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	hp, _ := GetDBMgr().GetDBrControl().GetPerson(req.UId)
	if hp == nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	optmem := house.GetMemByUId(p.Uid)
	if optmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(optmem, MinorMovBlacklist)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if optmem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}

	var custerr *xerrors.XError
	hmem := house.GetMemByUId(req.UId)
	if hmem != nil {
		if _, ok := hmem.CheckRef(); ok {
			xerr := xerrors.NewXError("该玩家为合并包厢盟主，无法加入黑名单。")
			return xerr.Code, xerr.Msg
		}
		custerr = house.ChangeRole(p.Uid, req.UId, consts.ROLE_BLACK)
	} else {
		custerr = house.MemJoin(req.UId, consts.ROLE_BLACK, 0, true, nil)
	}

	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	if optmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主将%sID:%d加入黑名单", hp.Nickname, hp.Uid)
	} else {
		msg = fmt.Sprintf("%sID:%d将%sID:%d加入黑名单", p.Nickname, p.Uid, hp.Nickname, hp.Uid)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, AddBlackList, msg)
	return xerrors.SuccessCode, nil
}

// ! 玩家黑名单删除
func Proto_ClubHouseMemberBlacklistDelete(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseMemberBlacklistDelete)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hp, _ := GetDBMgr().GetDBrControl().GetPerson(req.UId)
	if hp == nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	optmem := house.GetMemByUId(p.Uid)
	if optmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(optmem, MinorMovBlacklist)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if optmem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}

	custerr := house.MemDelete(req.UId, true, nil)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	var msg string
	if optmem.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主将%sID:%d移出黑名单", hp.Nickname, hp.Uid)
	} else {
		msg = fmt.Sprintf("%sID:%d将%sID:%d移出黑名单", p.Nickname, p.Uid, hp.Nickname, hp.Uid)
	}
	CreateClubMassage(house.DBClub.Id, p.Uid, RemoveBlackList, msg)
	return xerrors.SuccessCode, nil
}

// ! 玩家备注
func Proto_ClubHouseMemberRemark(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseMemberRemark)

	if utf8.RuneCountInString(req.URemark) > 10 {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 敏感词检查
	req.URemark = static.CheckSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.URemark)

	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	mem := house.GetMemByUId(req.UId)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	mem.URemark = req.URemark
	mem.Flush()
	return xerrors.SuccessCode, nil

}

// ! 玩家设置角色
func Proto_ClubHouseMemberRoleGen(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseMemberRoleGen)

	if req.UId == p.Uid {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetAdmin)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	custerr := house.ChangeRole(p.Uid, req.UId, req.URole)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	return xerrors.SuccessCode, nil
}

// ! 进入包厢
func Proto_ClubHouseMemberIn(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMemberIn)

	house, floor, mem, cusErr := inspectClubFloorMemberWithRight(req.HId, req.FId, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	// 玩家
	if mem == nil {
		// 删除玩家历史数据
		if p.HouseId == house.DBClub.HId && p.FloorId == floor.Id {
			// 内存数据
			p.HouseId = 0
			p.FloorId = 0
			// 清理缓存数据
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", 0, "FloorId", 0)
		}
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	// 查找玩家是否进入其余楼层
	// 包厢存在
	oldh := GetClubMgr().GetClubHouseByHId(p.HouseId)
	if oldh != nil {
		// 楼层存在
		oldf := oldh.GetFloorByFId(p.FloorId)
		if oldf != nil {
			if oldf.Id != req.FId {
				oldf.SafeMemOut(p.Uid)
				if p.HouseId == oldf.HId && p.FloorId == oldf.Id {
					// 内存数据
					p.HouseId = 0
					p.FloorId = 0
					// 清理缓存数据
					GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", 0, "FloorId", 0)
				}
			}
		}
	}
	return ChOptMemInV2(floor, s, p, req)
}

// ! 离开包厢
func Proto_ClubHouseMemberOut(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseMemberOut)

	// 包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.SuccessCode, nil
	}

	// 楼层
	if req.FId == 0 {
		// 当前数据
		p.HouseId = 0
		p.FloorId = 0
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", p.HouseId, "FloorId", p.FloorId)
		go GetClubMgr().UserLeaveHouse(p.Uid, req.HId)
		return xerrors.SuccessCode, nil
	}

	// 楼层
	floor := house.GetFloorByFId(req.FId)
	if floor == nil {
		return xerrors.SuccessCode, nil
	}

	floor.SafeMemOut(p.Uid)
	if p.HouseId == floor.HId && p.FloorId == floor.Id {
		// 内存数据
		p.HouseId = 0
		p.FloorId = 0
		// 清理缓存数据
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", 0, "FloorId", 0)
	}
	go GetClubMgr().UserLeaveHouse(p.Uid, req.HId)
	return xerrors.SuccessCode, nil
}

// ! 入驻包厢列表
func Proto_MemberHouseList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	// 数据
	req := data.(*static.Msg_CH_MemberHouseList)

	// 入驻包厢id
	items, err := GetDBMgr().ListHouseIdMemberJoin(p.Uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_MemberHouseList
	ack.Items = make([]static.Msg_HouseItem, 0)

	for _, item := range items {
		house := GetClubMgr().GetClubHouseById(item.Id)
		if house == nil {
			continue
		}

		if house.IsBeenMerged() {
			continue
		}

		// 创建
		if house.DBClub.UId == p.Uid {
			if req.HCreate {
				ack.Items = append(ack.Items, *item)
			}
			continue
		}

		// 入驻
		if req.HJoin {
			ack.Items = append(ack.Items, *item)
		}
	}

	sort.Slice(ack.Items, func(i, j int) bool {
		return ack.Items[i].MergeHId < ack.Items[j].MergeHId
	})

	return xerrors.SuccessCode, &ack
}

// ! 获取包厢信息
func Proto_ClubHouseBaseInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据
	req := data.(*static.Msg_CH_HouseBaseInfo)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	userAgentConfig, _ := GetDBMgr().GetUserAgentConfig(house.DBClub.UId)
	//if xerr != nil {
	//	return xerr.Code, xerr.Msg
	//}

	//if userAgentConfig != nil {
	//	// 联盟配置已关
	//	if userAgentConfig.IsUnion == false {
	//		if house.DBClub.IsVitamin == true {
	//			house.DBClub.IsVitamin = false
	//			house.flush()
	//			if !house.IsBeenMerged() {
	//				house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseAgentUpdateNtf, &static.MsgHouseAgentUpdateNtf{
	//					Hid:                house.DBClub.HId,
	//					VipFloorShowSwitch: userAgentConfig.IsVipFloors,
	//					UnionSwitch:        userAgentConfig.IsUnion,
	//				})
	//			}
	//		}
	//	}
	//}

	info := models.CheckHouseInBlank(house.DBClub.Id, GetDBMgr().GetDBmControl())
	if info != nil {
		return xerrors.BlankHouseErrorCode, info.Reason
	}
	if house.IsBeenMerged() {
		// 重定向
		house = GetClubMgr().GetClubHouseById(house.DBClub.MergeHId)
		if house == nil {
			return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
		}
	}

	hmem := house.GetMemByUId(p.Uid)
	if hmem == nil {
		// 删除玩家历史数据
		if p.HouseId == house.DBClub.HId {
			// 内存数据
			p.HouseId = 0
			p.FloorId = 0
			// 清理缓存数据
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", 0, "FloorId", 0)
		}
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	} else {
		if hmem.Lower(consts.ROLE_MEMBER) {
			return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
		}
	}

	dhouse := house.DBClub
	var ack static.Msg_HC_HouseBaseInfo
	ack.ApplySwitch = house.DBClub.ApplySwitch
	ack.UpdNFIds = make([]int, 0)
	ack.Id = dhouse.Id
	ack.HId = dhouse.HId
	ack.Area = dhouse.Area
	ack.OwnerId = dhouse.UId
	ack.URole = hmem.URole
	ack.UFloor = hmem.FId
	ack.Name = dhouse.Name
	ack.Notify = dhouse.Notify
	ack.IsPartner = hmem.IsPartner()
	ack.SuperiorId = hmem.Superior
	ack.IsChecked = dhouse.IsChecked
	ack.IsFrozen = dhouse.IsFrozen
	ack.IsMemHide = dhouse.IsMemHide
	ack.IsMemExit = dhouse.IsMemExit
	ack.IsVitamin = house.DBClub.IsVitamin
	ack.IsVitaminHide = house.DBClub.IsVitaminHide
	ack.IsGamePause = house.DBClub.IsGamePause
	ack.IsPartnerHide = house.DBClub.IsPartnerHide
	ack.IsVitaminModi = house.DBClub.IsVitaminModi
	ack.IsPartnerModi = house.DBClub.IsPartnerModi
	ack.IsMemberSend = house.DBClub.IsMemberSend
	ack.PrivateGPS = house.DBClub.PrivateGPS
	ack.IsActivity = house.IsActivity()
	ack.FangKaTipsMinNum = house.DBClub.FangKaTipsMinNum
	ack.RecordTimeInterval = house.DBClub.RecordTimeInterval
	var act bool
	if ack.IsActivity {
		arr, err := GetDBMgr().GetDBrControl().HouseActivityList(house.DBClub.Id, true)
		if err != nil {
			xlog.Logger().Errorln(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}

		for _, ar := range arr {
			if ar == nil {
				continue
			}
			if ar.Type == 1 && time.Now().Unix() < ar.EndTime && time.Now().Unix() > ar.BegTime {
				act = true
				ack.LuckTimes = house.GetUserLuckTimes(p.Uid, ar.Id, ar.TicketCount)
			} else {
				if time.Now().Unix() < ar.EndTime && time.Now().Unix() > ar.BegTime {
					act = true
				}
			}
		}
	}
	ack.IsActivity = act
	ack.IsPartnerApply = house.DBClub.IsPartnerApply
	ack.MaxTable = GetServer().ConHouse.TableNum
	ack.DisableSetJuniorVitamin = house.DBClub.DisVitaminJunior
	ack.VitaminAdmin = hmem.VitaminAdmin
	ack.VicePartner = hmem.VicePartner
	if house.IsHideOnlineNum(p.Uid) {
		ack.OnlineCur = -1
		ack.OnlineTotal = -1
		ack.OnlineTable = -1
	} else {
		ack.OnlineCur = house.OnlineMemCount()
		ack.OnlineTotal = house.TotalMemCount()
		ack.OnlineTable = house.GetTabOnlineCounts()
	}
	ack.FloorIDs = make([]int64, 0)
	ack.MixFloorIDs = make([]int64, 0)
	cardPool, _ := GetAllianceMgr().CheckUserLeagueCardPool(house.DBClub.UId)
	ack.CardPool = cardPool.CanPool
	ack.Vitamin = static.SwitchVitaminToF64(hmem.UVitamin)
	ack.VitaminPool = static.SwitchVitaminToF64(house.GetHouseVitaminPool(userAgentConfig))
	ack.TableJoinType = house.DBClub.TableJoinType
	ack.MixActive = house.DBClub.MixActive
	ack.Dialog = house.DBClub.Dialog
	ack.DialogActive = house.DBClub.DialogActive
	ack.AutoPayPartner = house.DBClub.AutoPayPartnrt
	ack.IsHidHide = house.DBClub.IsHidHide
	ack.TableShowCount = house.DBClub.TableShowCount
	ack.MinTableNum = house.DBClub.MinTableNum
	ack.MaxTableNum = house.DBClub.MaxTableNum
	ack.IsAiSuper = house.DBClub.TableJoinType == consts.NoCheat && house.DBClub.AiSuper
	ack.GameOn = house.DBClub.GameOn
	ack.AdminGameOn = house.DBClub.AdminGameOn
	ack.PartnerKick = house.DBClub.PartnerKick

	// 　包厢楼层id
	for k, v := range house.Floors {
		ack.FloorIDs = append(ack.FloorIDs, k)
		if v.IsMix && house.DBClub.MixActive {
			ack.MixFloorIDs = append(ack.MixFloorIDs, k)
		}
	}
	ack.OnlyQucikJoin = house.DBClub.OnlyQuickJoin
	if refHid, ok := hmem.CheckRef(); ok {
		ack.URefHId = refHid
	}

	ack.EmptyTableMax = house.DBClub.EmptyTableMax
	ack.EmptyTableBack = house.DBClub.EmptyTableBack
	ack.TableSortType = house.DBClub.TableSortType
	ack.IsHeadHide = house.DBClub.IsHeadHide
	ack.IsMemUidHide = house.DBClub.IsMemUidHide
	ack.IsOnlineHide = house.DBClub.IsOnlineHide
	ack.IsLimitGame = hmem.IsLimitGame
	ack.CreateTableType = house.DBClub.CreateTableType
	newTblSortType := house.DBClub.NewTableSortType
	if newTblSortType == 0 {
		if house.DBClub.EmptyTableBack {
			newTblSortType = 6
		} else {
			newTblSortType = 3
		}
	}
	ack.NewTableSortType = newTblSortType

	// 当前用户这个包厢的对应权限，如果没有 就根据身份给一个对应的默认权限
	urdata, _ := GetMRightMgr().FindRightByMember(hmem, false)
	ack.Uright = urdata

	shData, _ := house.GetMemberSwitch() //GetClubHouseByHId(req.HId)
	ack.HmSwitch = shData

	limitDistance := house.GetHouseTableLimitDistance()
	ack.Distance = limitDistance
	ack.RankOpen = house.DBClub.RankOpen
	ack.IsCurUserTeamOffWork = false
	tempUid := hmem.Partner
	if hmem.Partner == 0 {
		tempUid = house.DBClub.UId
	} else if hmem.Partner == 1 {
		tempUid = hmem.UId
	}
	tempstr := GetDBMgr().GetDBrControl().RedisV2.HGet(fmt.Sprintf(consts.REDIS_KEY_OFFWORK, house.DBClub.Id), fmt.Sprintf("%d", tempUid)).Val()
	var tempBool bool
	tempBool, _ = strconv.ParseBool(tempstr)
	ack.IsCurUserTeamOffWork = tempBool

	// 查盟主后台数据
	if userAgentConfig != nil {
		ack.VipFloorShowSwitch = userAgentConfig.IsVipFloors
		ack.UnionSwitch = userAgentConfig.IsUnion
	} else {
		ack.VipFloorShowSwitch = true
		ack.UnionSwitch = true
	}
	// ack.VitaminPoolMax = static.SwitchVitaminInt64(userAgentConfig.VitaminPoolMax)
	//syslog.Logger().Println("rdata", ack.Uright)
	// 如果是盟主切有楼层 则检查推荐更新
	if hmem.URole == consts.ROLE_CREATER && len(ack.FloorIDs) > 0 {
		updFIds := house.GetRecommendUpdateFloors()
		for _, updFid := range updFIds {
			ack.UpdNFIds = append(ack.UpdNFIds, house.GetFloorIndexByFid(updFid)+1)
		}
	}
	ack.FloorsColor = house.GetFloorColorArray()
	// 排序
	sort.Sort(util.Int64Slice(ack.FloorIDs))
	go GetClubMgr().UserInToHouse(p, req.HId)
	return xerrors.SuccessCode, &ack
}

// ! 获取包厢信息
func Proto_ClubHouseMemOnline(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据
	req := data.(*static.Msg_CH_HouseMemOnline)

	// 包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	// 成员
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorSetBackground)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	ack := new(static.Msg_HC_HouseMemberOnline)
	if house.IsHideOnlineNum(p.Uid) {
		ack.UNums = -1
	} else {
		ack.UNums = house.OnlineMemCount()
	}
	ack.Vitamin = static.SwitchVitaminToF64(mem.UVitamin)

	return xerrors.SuccessCode, ack
}

// ! 修改包厢公告名称
func Proto_ClubHouseBaseNNModify(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseBaseNNModify)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	// 敏感词检查
	req.HName = static.CheckSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.HName)
	req.HNotify = static.CheckSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.HNotify)

	custerr := house.ModifyNN(p.Uid, req.HName, req.HNotify)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	return xerrors.SuccessCode, nil
}

// 包厢楼层玩法修改
func Proto_ClubHouseFloorRuleModify(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	// 数据结构
	req := data.(*static.Msg_CH_HouseFloorRuleModify)
	house, floor, _, cusErr := inspectClubFloorMemberWithRight(req.HId, req.FId, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 获取玩法配置
	config := GetServer().GetGameConfig(req.FRule.KindId)
	if config == nil {
		// 没有找到玩法对应的配置
		cuserror := xerrors.NewXError("不支持的游戏玩法")
		return cuserror.Code, cuserror.Msg
	}

	if cuserror := house.IsMaintainableGame(req.FRule.KindId, req.FRule.Version); cuserror != nil {
		return cuserror.Code, cuserror.Msg
	}

	area_game := GetAreaGameByKid(req.FRule.KindId)
	if area_game == nil {
		// 没有找到玩法对应的区域游戏信息
		cuserror := xerrors.NewXError("该玩法已下线")
		return cuserror.Code, cuserror.Msg
	}

	// 如果客户端发送的规则版本于当前规则版本不一样 则提示玩法已变更
	if req.FRule.Version != area_game.GameRuleVersion {
		cuserror := xerrors.NewXError("规则版本不匹配，请重试。")
		return cuserror.Code, cuserror.Msg
	}

	// 保存强更和弱更版本
	req.FRule.Version = area_game.ForcedVersion
	req.FRule.RecommendVersion = area_game.RecommVersion

	if floor.ExistActivity() {
		return xerrors.HouseActivityExistError.Code, xerrors.HouseActivityExistError.Msg
	}

	return ChOptFloorRuleModify(floor, s, p, req)
}

// Proto_ClubHouseTableIn 玩家入桌
func Proto_ClubHouseTableIn(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseTableIn)
	if !ok {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}

	// 检查是否能观战,能观战直接进去入游戏
	if ok, ack := HouseTableWatch(req.HId, req.FId, req.NTId, p); ok {
		return xerrors.SuccessCode, ack
	}

	// 已加入牌桌 进入历史牌桌
	key := fmt.Sprintf("userstatus_doing_join_%d", p.Uid)
	cli := GetDBMgr().Redis
	if !cli.SetNX(key, p.Uid, time.Second*30).Val() {
		return xerrors.HouseFloorTableJoiningError.Code, xerrors.HouseFloorTableJoiningError.Msg
	}
	defer cli.Del(key)
	return HouseTableIn(&static.GpsInfo{Ip: s.IP, Longitude: req.Longitude, Latitude: req.Latitude, Address: req.Address}, p, req)
}

// 成员统计查询总计
func Proto_ClubHouseMemStatisticsTotal(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemStatisticsTotal)

	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	houseMemMap := house.GetMemberMap(false)

	hmem, ok := houseMemMap[p.Uid]
	if !ok {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	isOK, _ := GetMRightMgr().CheckRight(&hmem, MinorUserRecord)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	//if hmem.URole > constant.ROLE_ADMIN && hmem.Partner != 1 {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}

	statisticsItems, err := GetDBMgr().SelectHouseMemberStatistics(int(house.DBClub.Id), req.DFid, req.DayType, 0)
	if err != nil {
		xlog.Logger().Errorln("selecthousememberstatistics error :", err)
		return
	}

	var ack static.Msg_HC_HouseMemberStatisticsTotal

	for _, item := range statisticsItems {
		itemMem, ok := houseMemMap[item.UId]
		if !ok {
			continue
		}

		if itemMem.Partner == p.Uid || itemMem.UId == p.Uid {

			if req.GroupId < 0 {
				ack.PlayTimes += item.PlayTimes
				ack.BwTimes += item.BwTimes
				ack.TotalScore += item.TotalScore
				ack.ValidTimes += item.ValidTimes
				ack.BigValidTimes += item.BigValidTimes
			} else {
				inuids := house.GetGroupUser(p.Uid, req.GroupId)
				for _, uid := range inuids {
					if uid == item.UId {
						ack.PlayTimes += item.PlayTimes
						ack.BwTimes += item.BwTimes
						ack.TotalScore += item.TotalScore
						ack.ValidTimes += item.ValidTimes
						ack.BigValidTimes += item.BigValidTimes
					}
				}
			}
		}
	}
	ack.TotalScore = static.HF_DecimalDivide(ack.TotalScore, 1, 2)

	return xerrors.SuccessCode, &ack
}

// 成员统计查询
func Proto_ClubHouseMemStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMemStatistics)

	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	hmem := house.GetMemByUId(p.Uid)
	if hmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	if hmem.URole > consts.ROLE_MEMBER && !hmem.IsPartner() && !hmem.IsVicePartner() && !hmem.IsVitaminAdmin() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 校验时段、区间是否匹配
	timeRangeCnt := 1
	if req.QueryTimeInterval > 0 {
		timeRangeCnt = 24 / req.QueryTimeInterval
	}
	if req.QueryTimeRange > timeRangeCnt || req.QueryTimeRange <= 0 {
		return xerrors.InvalidTimeRangeError.Code, xerrors.InvalidTimeRangeError.Msg
	}

	// 筛选时间的起点和终点
	nowTime := time.Now()
	selectTime1 := nowTime
	selectTime2 := nowTime
	var zeroTime time.Time
	switch req.DayType {
	case static.DAY_RECORD_TODAY:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, 0))
		// 时间区间
		if req.QueryTimeInterval > 0 {
			selectTime1 = zeroTime.Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
			selectTime2 = zeroTime.Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
		} else {
			selectTime1 = zeroTime
			selectTime2 = zeroTime.Add(24 * time.Hour)
		}
	case static.DAY_RECORD_YESTERDAY:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, -1))
		// 时间区间
		if req.QueryTimeInterval > 0 {
			selectTime1 = zeroTime.Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
			selectTime2 = zeroTime.Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
		} else {
			selectTime1 = zeroTime
			selectTime2 = zeroTime.Add(24 * time.Hour)
		}
	case static.DAY_RECORD_3DAYS:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, -3))
		selectTime1 = zeroTime
		selectTime2 = zeroTime.AddDate(0, 0, 3)
	case static.DAY_RECORD_7DAYS:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, -7))
		selectTime1 = zeroTime
		selectTime2 = zeroTime.AddDate(0, 0, 7)
	default:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, req.DayType))
		// 时间区间
		if req.QueryTimeInterval > 0 {
			selectTime1 = zeroTime.Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
			selectTime2 = zeroTime.Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
		} else {
			selectTime1 = zeroTime
			selectTime2 = zeroTime.Add(24 * time.Hour)
		}
	}

	if hmem.IsPartner() {
		req.Partner = hmem.UId
	} else if hmem.IsVicePartner() {
		req.Partner = hmem.Partner
	}

	houseMem := house.GetMemberMap(false)

	leaveMemMap := make(map[int64]bool)
	if hmem.IsPartner() || hmem.IsVicePartner() {
		// leaveTime之前的记录为有效记录
		leaveTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
		leaveMemMap, _ = GetDBMgr().SelectLeaveHousePartnerMember(hmem.DHId, hmem.UId, leaveTime)
	}

	statisticsItemMap := make(map[int64]*static.HouseMemberStatisticsItem)

	var gIdMap map[int64]int64
	if req.GroupId != -1 {
		gIdMap = house.GetGroupUserMap(req.Partner, req.GroupId)
	}

	for _, mem := range houseMem {
		if req.Partner > 1 {
			if mem.UId == req.Partner || mem.Partner == req.Partner {
				if req.GroupId != -1 {
					if _, ok := gIdMap[mem.UId]; !ok {
						continue
					}
				}
			} else {
				continue
			}
		} else if req.Partner == -1 {
			if mem.Partner > 0 {
				continue
			}
		}

		// 数据补齐
		var staticItem static.HouseMemberStatisticsItem

		staticItem.UId = mem.UId
		staticItem.UName = mem.NickName
		staticItem.UUrl = mem.ImgUrl
		staticItem.UGender = mem.Sex
		staticItem.UJoinTime = 0
		staticItem.PlayTimes = 0
		staticItem.BwTimes = 0
		staticItem.ValidTimes = 0
		staticItem.InValidTimes = 0
		staticItem.TotalScore = 0
		staticItem.IsLike = false
		staticItem.IsExit = false
		statisticsItemMap[staticItem.UId] = &staticItem
	}

	// 声明统计结果变量
	var statisticsItems map[int64]models.QueryMemberStatisticsResult
	var selStsErr error
	statisticsItems = make(map[int64]models.QueryMemberStatisticsResult)

	// 合伙人按自己设置的低分局值筛选
	if req.LowScoreFlag > 0 && (hmem.IsPartner() || hmem.IsVicePartner()) {
		// 查询合伙人名下的成员战绩记录
		statisticsItems, selStsErr = GetDBMgr().SelectPartnerMemberStatisticsWithTotal(req.Partner, house.DBClub.Id, req.DFid, selectTime1, selectTime2)
	} else {
		// 默认查询统计信息
		statisticsItems, selStsErr = GetDBMgr().SelectHouseMemberStatisticsWithTotal(house.DBClub.Id, req.DFid, selectTime1, selectTime2)
	}
	if selStsErr != nil {
		xlog.Logger().Error(selStsErr.Error())
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	for _, item := range statisticsItems {
		if staticItem, ok := statisticsItemMap[item.Uid]; ok {
			staticItem.PlayTimes = item.PlayTimes
			staticItem.BwTimes = item.BigWinTimes
			staticItem.TotalScore = static.HF_DecimalDivide(item.TotalScore, 1, 2)
			staticItem.ValidTimes = item.ValidTimes
			staticItem.InValidTimes = item.PlayTimes - item.ValidTimes
		} else {
			// 盟主,管理员,比赛分管理员可以查看已经退圈的成员战绩
			if hmem.URole == consts.ROLE_CREATER || hmem.URole == consts.ROLE_ADMIN || hmem.IsVitaminAdmin() {
				if _, ok := houseMem[item.Uid]; ok {
					continue
				}
			} else if hmem.IsPartner() || hmem.IsVicePartner() {
				if _, ok := leaveMemMap[item.Uid]; !ok {
					continue
				}
			}

			var staticItem static.HouseMemberStatisticsItem
			staticItem.UId = item.Uid
			person, err := GetDBMgr().GetDBrControl().GetPerson(item.Uid)
			if err == nil {
				staticItem.UName = person.Nickname
				staticItem.UUrl = person.Imgurl
				staticItem.UGender = person.Sex
			}
			staticItem.PlayTimes = item.PlayTimes
			staticItem.BwTimes = item.BigWinTimes
			staticItem.TotalScore = static.HF_DecimalDivide(item.TotalScore, 1, 2)
			staticItem.ValidTimes = item.ValidTimes
			staticItem.InValidTimes = item.PlayTimes - item.ValidTimes
			staticItem.IsLike = false
			staticItem.IsExit = true
			statisticsItemMap[staticItem.UId] = &staticItem
		}
	}

	// 查询点赞信息
	curLikeCount := 0
	likeOptUserType := models.OptUserTypeAdmin
	if req.Partner > 0 {
		likeOptUserType = models.OptUserTypePartner
	}

	// 构造查询区间字符串 如01-03
	timeRangeStr := ""
	if req.QueryTimeInterval > 0 {
		startHour := (req.QueryTimeRange - 1) * req.QueryTimeInterval
		endHour := req.QueryTimeRange * req.QueryTimeInterval
		timeRangeStr = fmt.Sprintf("%02d-%02d", startHour, endHour)
	} else {
		timeRangeStr = "00-24"
	}

	likeMap, err := GetDBMgr().SelectHouseRecordUserLike(house.DBClub.Id, likeOptUserType, req.DayType, timeRangeStr)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 更新statisticsItemMap中的点赞信息
	for _, item := range statisticsItemMap {
		if _, ok := likeMap[item.UId]; ok {
			item.IsLike = true
		} else {
			item.IsLike = false
		}
	}
	// 定义一个新的map用于接收筛选数据
	newStatisticsItemMap := make(map[int64]*static.HouseMemberStatisticsItem)

	// 从 statisticsItemMap 筛选点赞和未点赞 数据
	if req.LikeFlag == 1 {
		for key, item := range statisticsItemMap {
			if item.IsLike {
				newStatisticsItemMap[key] = item
			}
		}
	} else if req.LikeFlag == 2 {
		for key, item := range statisticsItemMap {
			if !item.IsLike {
				newStatisticsItemMap[key] = item
			}
		}
	} else {
		// 默认全部数据
		for key, item := range statisticsItemMap {
			newStatisticsItemMap[key] = item
		}
	}

	var pMemsNum, pMemsPlayedNum int

	var searchItems []*static.HouseMemberStatisticsItem
	// 搜索功能
	if req.SearchKey != "" {
		for _, sItem := range newStatisticsItemMap {
			if strings.Contains(sItem.UName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", sItem.UId), req.SearchKey) {
				if sItem.PlayTimes > 0 {
					searchItems = append(searchItems, sItem)
				}
			}
		}
	} else {
		for _, sItem := range newStatisticsItemMap {
			if sItem.PlayTimes > 0 {
				searchItems = append(searchItems, sItem)
			}
		}
	}

	for _, item := range searchItems {
		if item.PlayTimes > 0 {
			pMemsPlayedNum++
		}
		pMemsNum++

		if req.Partner > 1 {
			item.GroupIndex = house.GetUserGroupIndex(req.Partner, item.UId)
		} else {
			item.GroupIndex = -1
		}
	}

	if err == nil {
		if req.SortType == static.SORT_PLAYTIMES_DES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.PlayTimes > item2.PlayTimes {
					return true
				} else if item1.PlayTimes == item2.PlayTimes {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})

		} else if req.SortType == static.SORT_PLAYTIMES_AES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.PlayTimes < item2.PlayTimes {
					return true
				} else if item1.PlayTimes == item2.PlayTimes {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_BWTIMES_DES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.BwTimes > item2.BwTimes {
					return true
				} else if item1.BwTimes == item2.BwTimes {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_BWTIMES_AES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.BwTimes < item2.BwTimes {
					return true
				} else if item1.BwTimes == item2.BwTimes {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_TOTALSCORE_DES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.TotalScore > item2.TotalScore {
					return true
				} else if item1.TotalScore == item2.TotalScore {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_TOTALSCORE_AES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.TotalScore < item2.TotalScore {
					return true
				} else if item1.TotalScore == item2.TotalScore {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_INVALIDROUND_DES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.InValidTimes > item2.InValidTimes {
					return true
				} else if item1.InValidTimes == item2.InValidTimes {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_INVALIDROUND_AES {
			sort.Sort(static.HouseMemberStatisticsItemWrapper{Item: searchItems, By: func(item1, item2 *static.HouseMemberStatisticsItem) bool {
				if item1.InValidTimes < item2.InValidTimes {
					return true
				} else if item1.InValidTimes == item2.InValidTimes {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		}
	}

	var ack static.Msg_HC_HouseMemberStatistics
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd
	ack.PartnerMemsPlayedNum = pMemsPlayedNum
	ack.PartnerMemsNum = pMemsNum

	// 统计人次
	{
		for _, item := range searchItems {
			// 记录总战绩,总人次,低分局人次
			ack.TotalScore += item.TotalScore
			ack.TotalPlayTimes += item.PlayTimes
			ack.TotalInValidPlayTimes += (item.PlayTimes - item.ValidTimes)
		}
	}

	// 统计局数（盟主/管理员）
	if req.Partner <= 0 {
		// 搜索、点赞筛选、低分局筛选 需要指定 SearchUids
		var SearchUids []int64
		if req.SearchKey != "" || req.Partner == -1 || req.LikeFlag > 0 || req.LowScoreFlag > 0 {
			for _, item := range searchItems {
				SearchUids = append(SearchUids, item.UId)
			}
		}
		playRound, validRound, err := GetDBMgr().SelectHouseGameRound(house.DBClub.Id, req.DFid, SearchUids, selectTime1, selectTime2)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		ack.TotalRound = playRound
		ack.InValidRound = playRound - validRound
	}

	// 计算点赞数
	curLikeCount = 0
	for _, item := range searchItems {
		if _, ok := likeMap[item.UId]; ok {
			curLikeCount++
		}
	}
	ack.LikeCount = curLikeCount

	ack.Items = make([]*static.HouseMemberStatisticsItem, 0)

	var idxBeg int = req.PBegin
	var idxEnd int = req.PEnd
	if idxEnd > len(searchItems) {
		idxEnd = len(searchItems)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}
	ack.Items = append(ack.Items, searchItems[idxBeg:idxEnd]...)

	return xerrors.SuccessCode, &ack
}

// 包厢我的战绩查询
func Proto_ClubHouseMyRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseMyRecord)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	houseRecords, _ := GetDBMgr().SelectHouseHallMyRecord(house.DBClub.Id, p.Uid)

	var ack static.Msg_HC_HouseMyRecord
	ack.MyStatlist = make([]static.TotalRecordStatItem, 0)

	// ! 每日统计数据
	for i := 0; i < 3; i++ {
		myStat := new(static.TotalRecordStatItem)
		myStat.PlayTimes = 0
		myStat.BwTimes = 0
		todayRecord, err := GetDBMgr().SelectUserGameDayRecordToDaySet(int(house.DBClub.Id), p.Uid, time.Now().AddDate(0, 0, -i))
		if err == nil && len(todayRecord) >= 1 {
			for _, recorddata := range todayRecord {
				myStat.PlayTimes += recorddata.PlayTimes
				myStat.BwTimes += recorddata.BwTimes
			}
		}

		ack.MyStatlist = append(ack.MyStatlist, *myStat)
	}

	// ! 战绩列表
	ack.Items = make([]static.HouseRecordItem, 0)

	var searchItems []static.HouseRecordItem
	for i := 0; i < len(houseRecords); i++ {
		rItem := new(static.HouseRecordItem)
		rItem.GameNum = houseRecords[i].GameNum
		rItem.PlayCount = houseRecords[i].PlayCount
		for _, recordPlay := range houseRecords[i].Player {
			if recordPlay.Uid == p.Uid {
				rItem.WinScore = recordPlay.Score
			}
		}

		rItem.WriteDate = houseRecords[i].PlayedAt
		searchItems = append(searchItems, *rItem)
	}

	sort.Sort(static.HouseRecordItemWrapper{searchItems, func(item1, item2 *static.HouseRecordItem) bool {
		if item1.WriteDate > item2.WriteDate {
			return true
		} else {
			return false
		}
	}})

	ack.Items = append(ack.Items, searchItems...)

	return xerrors.SuccessCode, &ack
}

// 包厢战绩查询
func Proto_ClubHouseRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseRecord)

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr == xerrors.InvalidPermission {
		if !house.IsPartner(p.Uid) {
			return cusErr.Code, cusErr.Msg
		}
	} else {
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
	}

	houseRecords, _ := GetDBMgr().SelectHouseHallRecord(house.DBClub.Id)

	var ack static.Msg_HC_HouseRecord
	// ! 战绩列表
	ack.Items = make([]static.HouseRecordItem, 0)

	var searchItems []static.HouseRecordItem
	for i := 0; i < len(houseRecords); i++ {
		rItem := new(static.HouseRecordItem)
		rItem.GameNum = houseRecords[i].GameNum
		rItem.PlayCount = houseRecords[i].PlayCount
		rItem.IsJoin = 0
		for _, recordPlay := range houseRecords[i].Player {
			if recordPlay.Uid == p.Uid {
				rItem.IsJoin = 1
				rItem.WinScore = recordPlay.Score
			}
		}
		rItem.IsHeart = houseRecords[i].IsHeart
		rItem.WriteDate = houseRecords[i].PlayedAt
		searchItems = append(searchItems, *rItem)
	}

	sort.Sort(static.HouseRecordItemWrapper{searchItems, func(item1, item2 *static.HouseRecordItem) bool {
		if item1.WriteDate > item2.WriteDate {
			return true
		} else {
			return false
		}
	}})

	var idxBeg int = req.PBegin
	var idxEnd int = req.PEnd
	if idxEnd > len(searchItems) {
		idxEnd = len(searchItems)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}

	ack.Items = append(ack.Items, searchItems[idxBeg:idxEnd]...)

	return xerrors.SuccessCode, ack
}

// 包厢战绩点赞
func Proto_ClubHouseRecordHeart(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseRecordHeart)

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 保存数据库数据
	err := GetDBMgr().UpdateGameRecordHeart(house.DBClub.Id, req.GameNum, 0, req.IsHeart)

	var ack static.Msg_HC_HouseRecordHeart
	if err == nil {
		return xerrors.SuccessCode, &ack
	} else {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
}

// 包厢大赢家统计查询
func Proto_ClubHouseRecordStatus(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseRecordStatus)
	if req.RecordType == static.RECORD_STATUS_PLAYTIMES { //对局统计

	} else if req.RecordType == static.RECORD_STATUS_BWTIMES { //大赢家统计
	}

	//house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, constant.ROLE_ADMIN, MinorWinerStatistical)
	//
	//if cusErr != xerrors.RespOk {
	//	if cusErr == xerrors.InvalidPermission {
	//		if !mem.IsVitaminAdmin() {
	//			return cusErr.Code, cusErr.Msg
	//		}
	//	} else {
	//		return cusErr.Code, cusErr.Msg
	//	}
	//}

	var housemembers []models.HouseMember
	var err error
	if req.RecordType == static.RECORD_STATUS_PLAYTIMES {
		house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorPvPStatistical)
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
		housemembers, err = GetDBMgr().HouseRecordPlayList(house.DBClub.Id)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else if req.RecordType == static.RECORD_STATUS_BWTIMES {
		house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorWinerStatistical)
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
		housemembers, err = GetDBMgr().HouseRecordBWList(house.DBClub.Id)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	var searchItems []static.HouseRecordStatusItem

	for i := 0; i < len(housemembers); i++ {
		dPersion, err := GetDBMgr().GetDBrControl().GetPerson(housemembers[i].UId)
		if err != nil {
			continue
		}

		// 空串
		if req.Param != "" {
			// ID/昵称 包含
			if !strings.Contains(static.HF_I64toa(dPersion.Uid), req.Param) &&
				!strings.Contains(dPersion.Nickname, req.Param) {
				continue
			}
		}

		if err == nil {
			mItem := new(static.HouseRecordStatusItem)
			mItem.UId = housemembers[i].UId

			person, err := GetDBMgr().GetDBrControl().GetPerson(mItem.UId)
			if err != nil {
				xlog.Logger().Errorln(err)
				continue
			}
			mItem.UName = person.Nickname
			mItem.UUrl = person.Imgurl
			if req.RecordType == static.RECORD_STATUS_PLAYTIMES {
				mItem.RecordTimes = housemembers[i].PlayTimes
			} else if req.RecordType == static.RECORD_STATUS_BWTIMES {
				mItem.RecordTimes = housemembers[i].BwTimes
			}

			if mItem.RecordTimes > 0 {
				searchItems = append(searchItems, *mItem)
			}
		}
	}

	var ack static.Msg_HC_HouseRecordStatus
	ack.Items = make([]static.HouseRecordStatusItem, 0)
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd
	for i := 0; i < len(searchItems); i++ {
		if i >= req.PBegin && i <= req.PEnd {
			ack.Items = append(ack.Items, searchItems[i])
		}
	}

	return xerrors.SuccessCode, ack
}

// 包厢大赢家统计清除
func Proto_ClubHouseRecordStatusClean(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseRecordStatusClean)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	houseMember := house.GetMemByUId(req.UId)
	var err error
	if req.RecordType == static.RECORD_STATUS_PLAYTIMES {
		house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorPvPStatistical)
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
		err = GetDBMgr().UpdateHouseRecordPlayTimes(house.DBClub.Id, req.UId, 0)
		if err == nil {
			houseMember.PlayTimes = 0
		}
	} else if req.RecordType == static.RECORD_STATUS_BWTIMES {
		house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorWinerStatistical)
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
		err = GetDBMgr().UpdateHouseRecordBwTimes(house.DBClub.Id, req.UId, 0)
		if err == nil {
			houseMember.BwTimes = 0
		}
	} else {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	if err != nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	var ack static.Msg_HC_HouseRecordStatusClean
	return xerrors.SuccessCode, ack
}

// 包厢大赢家统计清除所有
func Proto_ClubHouseRecordStatusCleanAll(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseRecordStatusCleanAll)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	var err error
	mems := house.GetMemSimple(false)
	for _, member := range mems {
		if req.RecordType == static.RECORD_STATUS_PLAYTIMES {
			house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorPvPStatistical)
			if cusErr != xerrors.RespOk {
				return cusErr.Code, cusErr.Msg
			}
			err = GetDBMgr().UpdateHouseRecordPlayTimes(house.DBClub.Id, member.UId, 0)
			if err == nil {
				member.PlayTimes = 0
			}
		} else if req.RecordType == static.RECORD_STATUS_BWTIMES {
			house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorWinerStatistical)
			if cusErr != xerrors.RespOk {
				return cusErr.Code, cusErr.Msg
			}
			err = GetDBMgr().UpdateHouseRecordBwTimes(house.DBClub.Id, member.UId, 0)
			if err == nil {
				member.BwTimes = 0
			}
		}
	}

	if err != nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	var ack static.Msg_HC_HouseRecordStatusCleanAll
	return xerrors.SuccessCode, ack
}

// 包厢扣卡统计查询
func Proto_ClubHouseRecordKaCost(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	// 数据结构
	req := data.(*static.Msg_CH_HouseRecordKaCost)

	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	hmem := house.GetMemByUId(p.Uid)
	if hmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	recordcostnimis, err := GetDBMgr().HouseRecordCostList(house.DBClub.Id)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseRecordKaCost
	ack.Items = make([]static.RecordGameCostMini, 0)

	for i := 0; i < len(recordcostnimis); i++ {
		precordcostnimi := new(static.RecordGameCostMini)
		precordcostnimi.HId = recordcostnimis[i].HId
		precordcostnimi.KaCost = recordcostnimis[i].KaCost
		precordcostnimi.PlayTime = recordcostnimis[i].PlayTime
		precordcostnimi.Date = recordcostnimis[i].Date.Unix()
		ack.Items = append(ack.Items, *precordcostnimi)
	}

	return xerrors.SuccessCode, &ack
}

// 新版包厢战绩查询
func Proto_ClubHouseGameRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseGameRecord)

	house := GetClubMgr().GetClubHouseByHId(int(req.HId))
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	memMap := house.GetMemberMap(false)
	opMem, ok := memMap[p.Uid]
	if !ok {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	isFaker := GetDBMgr().GetDBrControl().RedisV2.SIsMember("faker_admin", p.Uid).Val()

	if !isFaker {
		if p.Uid == 0 {
			isOK, _ := GetMRightMgr().CheckRight(&opMem, MinorTeaRecord)
			if !isOK {
				return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
			}
		} else {
			isOK, _ := GetMRightMgr().CheckRight(&opMem, MinorMyRecord)
			if !isOK {
				return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
			}
		}
	}

	puid := int64(0)
	if opMem.IsPartner() {
		puid = opMem.UId
	} else if opMem.IsVicePartner() {
		puid = opMem.Partner
	}

	// 校验时段、区间是否匹配
	timeRangeCnt := 1
	if req.QueryTimeInterval > 0 {
		timeRangeCnt = 24 / req.QueryTimeInterval
	}
	if req.QueryTimeRange > timeRangeCnt || req.QueryTimeRange <= 0 {
		return xerrors.InvalidTimeRangeError.Code, xerrors.InvalidTimeRangeError.Msg
	}

	// 筛选时间的起点和终点
	nowTime := time.Now()
	selectTime1 := nowTime
	selectTime2 := nowTime
	zeroTime := static.GetZeroTime(nowTime.AddDate(0, 0, req.SelectTime))
	if req.QueryTimeInterval > 0 {
		selectTime1 = zeroTime.Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
		selectTime2 = zeroTime.Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
	} else {
		selectTime1 = zeroTime
		selectTime2 = zeroTime.Add(24 * time.Hour)
	}

	// 声明变量
	var houseRecords []static.GameRecordDetal
	var curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curBwTimes, curPlayTimes, curInValidTimes int
	var curUserScore float64

	// 统计查询总点赞数
	curTotalLikeRound := 0
	totalLikeMap := make(map[string]bool)
	recordLikeMap := make(map[string]bool)

	optUserType := -1 // 我的战绩页签 optUserType为-1
	if isFaker || opMem.URole == consts.ROLE_CREATER || opMem.URole == consts.ROLE_ADMIN || opMem.IsVitaminAdmin() {
		optUserType = models.OptUserTypeAdmin
	} else if opMem.IsPartner() || opMem.IsVicePartner() {
		optUserType = models.OptUserTypePartner
	}

	// 圈子战绩页签/成员战绩页签--详情按钮
	if optUserType >= models.OptUserTypeAdmin {
		// 构造查询区间字符串 如01-03
		timeRangeStr := ""
		if req.QueryTimeInterval > 0 {
			startHour := (req.QueryTimeRange - 1) * req.QueryTimeInterval
			endHour := req.QueryTimeRange * req.QueryTimeInterval
			timeRangeStr = fmt.Sprintf("%02d-%02d", startHour, endHour)
		} else {
			timeRangeStr = "00-24"
		}

		// 查询点赞记录
		totalLikeMap, _ = GetDBMgr().SelectHouseRecordGameLike(house.DBClub.Id, optUserType, models.OptTypeGameTotal, req.SelectTime, timeRangeStr)
		recordLikeMap = totalLikeMap // 点赞的地方 不再区分 req.RecordType  0圈子战绩 1对局详情 2大赢家详情
		//if req.RecordType == model.OptTypeGameTotal {
		//	recordLikeMap = totalLikeMap
		//} else {
		//	recordLikeMap, _ = GetDBMgr().SelectHouseRecordGameLike(house.DBClub.Id, optUserType, req.RecordType, req.SelectTime, timeRangeStr)
		//}

		// 查询记录信息
		if req.LowScoreFlag > 0 && (opMem.IsPartner() || opMem.IsVicePartner()) && !isFaker {
			houseRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound, _ = GetDBMgr().SelectPartnerGameRecordByDate(house.DBClub.Id, req.DFID, req.UId, req.BwUser, puid, memMap, selectTime1, selectTime2, req.SearchKey, recordLikeMap, req.LikeFlag, req.RoundType)
		} else {
			houseRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound, _ = GetDBMgr().SelectHouseGameRecordByDate(house.DBClub.Id, req.DFID, req.UId, req.BwUser, puid, memMap, selectTime1, selectTime2, req.SearchKey, recordLikeMap, req.LikeFlag, req.RoundType)
		}
	} else {
		// 我的战绩页签
		houseRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound, _ = GetDBMgr().SelectHouseGameRecordByDate(house.DBClub.Id, req.DFID, req.UId, req.BwUser, puid, memMap, selectTime1, selectTime2, req.SearchKey, recordLikeMap, req.LikeFlag, req.RoundType)

	}

	var ack static.Msg_HC_HouseGameRecord
	ack.UId = req.UId

	ack.TotalScore = curUserScore
	ack.TotalRound = curTotalRound
	ack.CompleteRound = curCompleteRound
	ack.DismissRound = curDismissRound
	ack.ValidRound = curValidRound
	ack.InValidRound = curInValidRound
	ack.TotalBWTimes = curBwTimes
	ack.PlayTimes = curPlayTimes
	ack.InvalidTimes = curInValidTimes
	ack.TotalLike = curTotalLikeRound

	// ! 战绩列表
	ack.Items = make([]static.GameRecordDetal, 0)

	sort.Sort(static.GameRecordDetalWrapper{Item: houseRecords, By: func(item1, item2 *static.GameRecordDetal) bool {
		if item1.PlayedAt > item2.PlayedAt {
			return true
		} else {
			return false
		}
		/*
			if item1.FinishType != public.FINISH_STA_PLAYING && item2.FinishType != public.FINISH_STA_PLAYING {
				if item1.PlayedAt > item2.PlayedAt {
					return true
				} else {
					return false
				}
			} else {
				if item1.FinishType == public.FINISH_STA_PLAYING {
					return true
				} else {
					return false
				}
			}
		*/
	}})

	gameNums := make(map[string]struct{})
	for i := 0; i < len(houseRecords); i++ {
		if len(ack.Items) >= 50 {
			break
		}
		if ((houseRecords[i].PlayedAt < req.QueryBeginTime && req.QueryBeginTime != 0) || req.QueryBeginTime == 0) && len(ack.Items) < 50 {
			// 更新头像跟性别
			for index, userinfo := range houseRecords[i].Player {
				p, err := GetDBMgr().GetDBrControl().GetPerson(userinfo.Uid)
				if p != nil && err == nil {
					houseRecords[i].Player[index].HeadUrl = p.Imgurl
					houseRecords[i].Player[index].Sex = p.Sex
					houseRecords[i].Player[index].IsExit = false
					tempMem, ok := memMap[p.Uid]
					if ok {
						//tempMem.Partner == 0 可能是管理员 和 盟主
						//tempMem.Partner  == 1 是队长   副队长的partenr 是队长的
						if opMem.UId == tempMem.UId { // 操作的人 循环到自己  给 1 代表我
							houseRecords[i].PlayerTags[index] = 1 //! 0：不显示  1：我  2：队员
						} else {
							if opMem.URole == consts.ROLE_CREATER || opMem.URole == consts.ROLE_ADMIN || opMem.IsVitaminAdmin() {
								if tempMem.Partner == 0 {
									houseRecords[i].PlayerTags[index] = 2 //! 0：不显示  1：我  2：队员
								}
							} else if opMem.IsPartner() {
								if tempMem.Partner == opMem.UId {
									houseRecords[i].PlayerTags[index] = 2 //! 0：不显示  1：我  2：队员
									if tempMem.IsJunior() && tempMem.Superior == opMem.UId {
										houseRecords[i].PlayerTags[index] = 0 //! 0：不显示  1：我  2：队员
									}
								}
							} else if opMem.IsVicePartner() {
								if tempMem.Partner == opMem.Partner || opMem.Partner == tempMem.UId {
									houseRecords[i].PlayerTags[index] = 2 //! 0：不显示  1：我  2：队员
									if tempMem.IsJunior() && tempMem.Superior == opMem.UId {
										houseRecords[i].PlayerTags[index] = 0 //! 0：不显示  1：我  2：队员
									}
								}
							}
						}
					} else {
						houseRecords[i].Player[index].IsExit = true
					}
				}
			}
			houseRecords[i].GameIndex = curTotalRound - i

			areaGame := GetAreaGameByKid(houseRecords[i].KindId)
			if areaGame != nil {
				houseRecords[i].WanFa = areaGame.Name
			} else {
				houseRecords[i].WanFa = ""
			}

			ack.Items = append(ack.Items, houseRecords[i])
			gameNums[houseRecords[i].GameNum] = struct{}{}
		}
	}
	if l := len(gameNums); l > 0 {
		gameNumList := make([]string, l)
		var i int
		for gn := range gameNums {
			gameNumList[i] = gn
			i++
		}
		var (
			pointMap map[string]int64
			err      error
		)
		if req.RecordType == 0 { // 我的战绩
			pointMap, err = GetDBMgr().SelectHouseMemberPayWithGameNums(house.DBClub.Id, p.Uid, gameNumList)
		} else {
			pointMap, err = GetDBMgr().SelectHousePartnerRewardWithGameNums(house.DBClub.Id, p.Uid, gameNumList)
		}
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		for i, rec := range ack.Items {
			rec.Point = static.SwitchVitaminToF64(pointMap[rec.GameNum])
			ack.Items[i] = rec
		}
	}
	return xerrors.SuccessCode, &ack
}

// 新版包厢经营状况
func Proto_ClubHouseOperationalStatus(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseOperationalStatus)

	house := GetClubMgr().GetClubHouseByHId(int(req.HId))
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	isFaker := GetDBMgr().GetDBrControl().RedisV2.SIsMember("faker_admin", p.Uid).Val()

	if !isFaker {
		isOK, _ := GetMRightMgr().CheckRight(mem, MinorFloorStatistical)
		if !isOK {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}
	}

	houseosmap, err := GetDBMgr().SelectHouseRoundByDate(house.DBClub.Id, 7)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var keys []string
	for k := range houseosmap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var ack static.Msg_HC_HouseOperationalStatus

	for i := len(keys) - 1; i >= 0; i-- {
		ack.Items = append(ack.Items, *(houseosmap[keys[i]]))
	}

	return xerrors.SuccessCode, &ack
}

// 包厢楼层活动创建
func Proto_ClubHouseActCreate(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseActCreate)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	// 敏感词检查
	req.ActName = static.CheckSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.ActName)

	if len(req.FIds) != 0 {
		for _, fid := range req.FIds {
			// 有楼层
			floor := house.GetFloorByFId(fid)
			if floor == nil {
				return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
			}
		}
	} else {
		return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
	}

	hmem := house.GetMemByUId(p.Uid)
	if hmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if hmem.Lower(consts.ROLE_CREATER) {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 条件判定
	if req.ActEndTime-req.ActBegTime < 600 {
		return xerrors.HouseFloorActivityTimeError.Code, xerrors.HouseFloorActivityTimeError.Msg
	}

	actlist, err := GetDBMgr().GetDBrControl().HouseActivityList(house.DBClub.Id, true)
	if len(actlist) > GetServer().ConHouse.ActMax {
		return xerrors.HouseActMaxOverFlowError.Code, xerrors.HouseActMaxOverFlowError.Msg
	}
	if req.Type == 1 { // 抽奖活动
		for _, item := range actlist {
			if item.Type == 1 && item.EndTime > time.Now().Unix() {
				return xerrors.HouseAllowOneLuckAct.Code, xerrors.HouseAllowOneLuckAct.Msg
			}
		}
		if req.TicketCount == 0 {
			return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}
	}

	hfAct := new(static.HouseActivity)
	hfAct.FId = GetClubMgr().HouseFidsToStr(req.FIds)
	hfAct.FIdIndex = GetClubMgr().HouseFidsToStr(house.GetFloorIndex(req.FIds))
	hfAct.DHId = house.DBClub.Id
	hfAct.Kind = req.ActType
	hfAct.Name = req.ActName
	hfAct.Status = 1
	hfAct.HideInfo = req.HideInfo
	hfAct.BegTime = req.ActBegTime
	hfAct.EndTime = req.ActEndTime
	hfAct.Type = req.Type
	hfAct.TicketCount = req.TicketCount
	actID, err := GetDBMgr().HouseActivityCreate(hfAct, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	msg := fmt.Sprintf("盟主开启了活动%s", req.ActName)
	go CreateClubMassage(house.DBClub.Id, p.Uid, ActiveCreate, msg)

	if req.Type == 1 {
		cusErr := house.CreateLuckActive(req.Rewords, actID)
		if cusErr != nil && cusErr != xerrors.RespOk {
			GetDBMgr().HouseActivityDelete(house.DBClub.Id, actID)
			return cusErr.Code, cusErr.Msg
		}
	}

	return xerrors.SuccessCode, nil
}

// 包厢楼层活动删除
func Proto_ClubHouseActDelete(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseActDelete)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	hact := GetDBMgr().HouseActivityExist(house.DBClub.Id, req.ActId)
	if hact == nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	err := GetDBMgr().HouseActivityDelete(house.DBClub.Id, req.ActId)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	msg := fmt.Sprintf("盟主关闭了活动%s", hact.Name)
	CreateClubMassage(house.DBClub.Id, p.Uid, ActiveClose, msg)
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseActivityDeleteNTF, req)
	return xerrors.SuccessCode, nil
}

// 包厢楼层活动列表
func Proto_ClubHouseActList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseActList)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hmem := house.GetMemByUId(p.Uid)
	if hmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	items, err := GetDBMgr().GetDBrControl().HouseActivityList(house.DBClub.Id, false)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	ack := new(static.Msg_HC_HouseActList)
	ack.ActItems = make([]*static.ActListItem, 0)

	actmap := make(map[int64]*static.ActListItem)
	beginingAct := make([]int64, 0)
	unbeginingAct := make([]int64, 0)
	endingAct := make([]int64, 0)

	for _, item := range items {
		actitem := new(static.ActListItem)
		actitem.ActId = item.Id
		actitem.ActName = item.Name
		actitem.ActHideInfo = item.HideInfo
		actitem.Type = item.Type
		if item.Type == 1 {
			actitem.Rewords = house.GetLuckConfig(item.Id)
		} else {
			xlog.Logger().Error("error type")
		}
		curtime := time.Now().Unix()
		if curtime < item.BegTime {
			actitem.ActState = consts.HFACT_UNBEGUN
			unbeginingAct = append(unbeginingAct, actitem.ActId)
		} else if curtime > item.EndTime {
			actitem.ActState = consts.HFACT_ENDING
			endingAct = append(endingAct, actitem.ActId)
		} else {
			actitem.ActState = consts.HFACT_BEGINING
			beginingAct = append(beginingAct, actitem.ActId)
		}

		actmap[actitem.ActId] = actitem
		// ack.ActItems = append(ack.ActItems, actitem)
	}

	// 排序
	sort.Sort(util.Int64Slice(beginingAct))
	sort.Sort(util.Int64Slice(unbeginingAct))
	sort.Sort(util.Int64Slice(endingAct))

	for _, id := range beginingAct {
		ack.ActItems = append(ack.ActItems, actmap[id])
	}
	for _, id := range endingAct {
		ack.ActItems = append(ack.ActItems, actmap[id])
	}
	for _, id := range unbeginingAct {
		ack.ActItems = append(ack.ActItems, actmap[id])
	}

	return xerrors.SuccessCode, ack
}

// 包厢楼层活动列表
func Proto_ClubHouseActInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseActInfo)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hmem := house.GetMemByUId(p.Uid)
	if hmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	// 活动数据
	fact, err := GetDBMgr().GetDBrControl().HouseActivityInfo(house.DBClub.Id, req.ActId)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if fact == nil {
		return xerrors.HouseActivityNotExistError.Code, xerrors.HouseActivityNotExistError.Msg
	}

	// 数据填充
	ack := new(static.Msg_HC_HouseActInfo)
	ack.FIds = GetClubMgr().GetHouseFidsByStr(fact.FId)
	ack.FIdIndexs = GetClubMgr().GetHouseFidsByStr(fact.FIdIndex)
	ack.ActId = req.ActId
	ack.ActName = fact.Name
	ack.ActType = fact.Kind
	ack.ActHideInfo = fact.HideInfo
	curtime := time.Now().Unix()
	if curtime < fact.BegTime {
		ack.ActState = consts.HFACT_UNBEGUN
	} else if curtime > fact.EndTime {
		ack.ActState = consts.HFACT_ENDING
	} else {
		ack.ActState = consts.HFACT_BEGINING
	}
	ack.ActBegTime = fact.BegTime
	ack.ActEndTime = fact.EndTime
	ack.Type = fact.Type
	// 活动详情
	if fact.Type == 0 {
		items, err := GetDBMgr().GetDBrControl().HouseActivityRecordList(req.ActId, 0, 299)
		if err != nil {
			xlog.Logger().Errorln(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		ack.UserItems = make([]*static.ActRecordItem, len(items))
		for n, item := range items {
			ack.UserItems[n] = item
		}
		return xerrors.SuccessCode, ack
	}
	items := house.GetLuckDetail(req.ActId, req.Uid)
	ack.UserItems = make([]*static.ActRecordItem, 0, len(items))
	mems := house.GetMemSimpleToMap(false)
	for _, item := range items {
		mem := mems[item.Uid]
		if mem == nil {
			continue
		}
		ack.UserItems = append(ack.UserItems, &static.ActRecordItem{UId: item.Uid,
			Rank: item.Rank, CreatedTime: item.CreatedAt.Unix(), UName: mem.NickName})
	}
	return xerrors.SuccessCode, ack

}

// 包厢游戏玩法列表
func Proto_ClubHouseAreaGames(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseActList)
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	xlog.Logger().Debugln("得到包厢信息，lid：", house.DBClub.HId, "区域码：", house.DBClub.Area)
	// 结果
	ack := new(static.AreaPackageSeek)
	ack.AreaCode = house.DBClub.Area
	pkgs, xerr := house.GetMaintainablePkg()
	if xerr != nil {
		return xerr.Code, xerr.Msg
	}
	ack.Packages = pkgs
	return xerrors.SuccessCode, ack
}

func Proto_FloorRename(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseFloorRename)
	if len([]rune(req.Name)) > 8 {
		return xerrors.InvalidFloorNameError.Code, xerrors.InvalidFloorNameError.Msg
	}
	house, floor, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, req.FloorID, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	// 敏感词检查
	req.Name = static.CheckSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.Name)

	err := floor.Rename(req.Name)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorRenameNTF, &static.MsgHouseFloorRenameNTF{Fid: req.FloorID, Name: req.Name})
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseMixEditor(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseMixFloor)
	if req.EmptyTableMax <= 0 {
		req.EmptyTableMax = 1
	}
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_ADMIN, MinorSetRoll)

	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	house.FloorLock.CustomLock()
	defer house.FloorLock.CustomUnLock()
	// 校验请求
	// 如果准备开启智能防作弊则校验分数上线是否在合法范围
	if req.MixActive && req.TableJoinType == consts.NoCheat && req.AICheck {
		if req.AITotalScoreLimit < consts.AICHECKTOTALSCOREMIN || req.AITotalScoreLimit > consts.AICHECKTOTALSCOREMAX {
			xerr := xerrors.NewXError(fmt.Sprintf("数值范围为%d-%d，请重新填写。", consts.AICHECKTOTALSCOREMIN, consts.AICHECKTOTALSCOREMAX))
			return xerr.Code, xerr.Msg
		}
		house.DBClub.AITotalScoreLimit = req.AITotalScoreLimit
	}
	// 混排类型：0手动加桌 1自动加桌 2智能防作弊
	house.DBClub.TableJoinType = req.TableJoinType
	// 智能超级防作弊
	house.DBClub.AiSuper = req.AISuper
	// 智能防作弊
	house.DBClub.AICheck = req.AICheck
	// 写缓存
	house.flush()

	if house.DBClub.MixActive == req.MixActive {
		for f, floor := range house.Floors {
			nowMix := false
			for _, fid := range req.FIDs {
				if f == fid {
					nowMix = true
				}
			}
			floor.DataLock.RLock()
			if nowMix != floor.IsMix {
				for _, table := range floor.Tables {
					if table.TId > 0 {
						floor.DataLock.RUnlock()
						return xerrors.UserFloorGameError.Code, xerrors.UserFloorGameError.Msg
					}
				}
			} else if house.DBClub.MixTableNum != req.TableNum {
				for _, table := range floor.Tables {
					if table.TId > 0 {
						floor.DataLock.RUnlock()
						return xerrors.UserFloorGameError.Code, xerrors.UserFloorGameError.Msg
					}
				}
			} else if house.DBClub.TableJoinType == consts.AutoAdd && house.DBClub.EmptyTableMax > req.EmptyTableMax {
				var sum int
				for _, table := range floor.Tables {
					if !table.Begin && table.UserCount() > 0 {
						sum++
						if sum > req.EmptyTableMax {
							floor.DataLock.RUnlock()
							return xerrors.ResultErrorCode, "当前楼层有空桌子被占用,暂无法调整空桌数量"
						}
					}
				}
			}
			floor.DataLock.RUnlock()
		}
	} else if req.MixActive {
		for _, fid := range req.FIDs {
			floor := house.Floors[fid]
			if floor == nil {
				continue // 可能是被删除的楼层
			}
			floor.IsMix = true
			floor.DataLock.RLock()
			for _, table := range floor.Tables {
				if table.TId > 0 {
					floor.DataLock.RUnlock()
					return xerrors.UserFloorGameError.Code, xerrors.UserFloorGameError.Msg
				}
			}
			floor.DataLock.RUnlock()
		}

	} else {
		for _, floor := range house.Floors {
			if floor.IsMix {
				floor.DataLock.RLock()
				for _, table := range floor.Tables {
					if table.TId > 0 {
						floor.DataLock.RUnlock()
						return xerrors.UserFloorGameError.Code, xerrors.UserFloorGameError.Msg
					}
				}
				floor.DataLock.RUnlock()
			}
		}
	}
	err := house.UpdateMixInfo(req.FIDs, req.TableNum, req.MixActive, req.EmptyTableMax, req.CreateTableType, req.NewTableSortType) //req.EmptyTableBack, req.TableSortType
	if err != nil {
		return err.Code, err.Msg
	}
	var mixFloor []*HouseFloor
	if house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				mixFloor = append(mixFloor, floor)
			}
		}
		if sortTableByCreateTime(mixFloor) {
			xlog.Logger().Warnf("sortTableByCreateTime after UpdateMixInfo")
			house.FloorLock.CustomUnLock()
			house.SyncTablesWithSorted()
			house.FloorLock.CustomLock()
		}
	}
	go func() {
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseMixFloorEdirotNTF,
			&static.MsgHouseFloorMixEditor{
				Hid:               house.DBClub.HId,
				FIDs:              req.FIDs,
				TableNum:          req.TableNum,
				IsActie:           req.MixActive,
				AICheck:           house.DBClub.AICheck,
				AISuper:           house.DBClub.AiSuper,
				AITotalScoreLimit: house.DBClub.AITotalScoreLimit,
				TableJoinType:     house.DBClub.TableJoinType,
				EmptyTableBack:    house.DBClub.EmptyTableBack,
				EmptyTableMax:     house.DBClub.EmptyTableMax,
				TableSortType:     house.DBClub.TableSortType,
				CreateTableType:   house.DBClub.CreateTableType,
				NewTableSortType:  house.DBClub.NewTableSortType,
			})
	}()

	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseMixInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseMixInfo)
	house := GetClubMgr().GetClubHouseByHId(req.HID)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	dest := []int64{}
	for k, floor := range house.Floors {
		if floor.IsMix {
			dest = append(dest, k)
		}
	}
	newTblSortType := house.DBClub.NewTableSortType
	if newTblSortType == 0 {
		if house.DBClub.EmptyTableBack {
			newTblSortType = 6
		} else {
			newTblSortType = 3
		}
	}
	return xerrors.SuccessCode, &static.MsgRespHouseMixInfo{
		IsMix:             true,
		MixActive:         house.DBClub.MixActive,
		TableNum:          house.DBClub.MixTableNum,
		FIDs:              dest,
		AICheck:           house.DBClub.AICheck,
		AISuper:           house.DBClub.AiSuper,
		AITotalScoreLimit: house.DBClub.AITotalScoreLimit,
		TableJoinType:     house.DBClub.TableJoinType,
		EmptyTableBack:    house.DBClub.EmptyTableBack,
		EmptyTableMax:     house.DBClub.EmptyTableMax,
		TableSortType:     house.DBClub.TableSortType,
		NewTableSortType:  newTblSortType,
		CreateTableType:   house.DBClub.CreateTableType,
	}
}

func Proto_ClubHouseMixFloorTableCreate(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseMixFloorTableCreate)
	house, floor, _, cusErr := inspectClubFloorMemberWithRight(req.HID, int64(req.FID), p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if !house.DBClub.MixActive {
		return xerrors.SuccessCode, &static.MsgRespHouseMixInfo{MixActive: false}
	}
	if len(floor.Tables) >= GetServer().ConHouse.TableNum {
		return xerrors.TableTooManyError.Code, xerrors.TableTooManyError.Msg
	}

	floor.Tables[999] = NewHFT(floor.Rule.PlayerNum, 999)

	go func() {
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseMixFloorTableChangeNTF,
			&static.Msg_Null{})
		house.FloorLock.RLock()
		for _, f := range house.Floors {
			f.RedisPub(consts.MsgTypeHouseMixFloorTableChangeNTF, &static.Msg_Null{})
		}
		house.FloorLock.RUnlock()
	}()

	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseMixFloorTableDelete(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseMixFloorTableCreate)
	house, floor, _, cusErr := inspectClubFloorMemberWithRight(req.HID, int64(req.FID), p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if !house.DBClub.MixActive {
		return xerrors.SuccessCode, &static.MsgRespHouseMixInfo{MixActive: false}
	}
	for k, hft := range floor.Tables {
		// if hft.IsDefault {
		// 	continue
		// }
		hasUser := false
		if len(hft.UserWithOnline) > 0 {
			for _, v := range hft.UserWithOnline {
				if v.Uid != 0 {
					hasUser = true
					break
				}
			}
		}
		if hasUser {
			continue
		} else {
			delete(floor.Tables, k)

			go func() {
				house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseMixFloorTableChangeNTF,
					&static.Msg_Null{})
				house.FloorLock.RLock()
				for _, f := range house.Floors {
					f.RedisPub(consts.MsgTypeHouseMixFloorTableChangeNTF, &static.Msg_Null{})
				}
				house.FloorLock.RUnlock()
			}()
			return xerrors.SuccessCode, nil
		}
	}
	return xerrors.NoEmptyTableCode, xerrors.NoEmptyTableError.Msg
}

func Proto_ClubHouseTableChange(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseMixFloorTableChange)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HID, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if !house.DBClub.MixActive {
		return xerrors.SuccessCode, &static.MsgRespHouseMixInfo{IsMix: false, MixActive: false}
	}
	// house.OnReload = true
	// defer func() {
	// 	house.OnReload = false
	// }()
	for _, item := range req.Detail {
		if item.TableNum == 0 {
			continue
		}
		floor := house.GetFloorByFId(item.FID)
		if floor == nil {
			return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}
		floor.DataLock.Lock()
		if item.TableNum > 0 {
			if len(floor.Tables)+item.TableNum > GetServer().ConHouse.TableNum {
				floor.DataLock.Unlock()
				return xerrors.TableTooManyError.Code, xerrors.TableTooManyError.Msg
			}
			for i := 0; i < item.TableNum; i++ {
				time.Sleep(10 * time.Nanosecond) // 防止桌子纳秒数相同
				floor.Tables[969+i] = NewHFT(floor.Rule.PlayerNum, 969+i)
			}
		}
		if item.TableNum < 0 {
			var emptyTable int
			for _, hft := range floor.Tables {
				hasUser := false
				for _, v := range hft.UserWithOnline {
					if v.Uid != 0 {
						hasUser = true
						break
					}
				}
				if hasUser {
					continue
				} else {
					emptyTable++
				}
			}
			if emptyTable+item.TableNum < 0 {
				floor.DataLock.Unlock()
				return xerrors.NoEmptyTableCode, xerrors.NoEmptyTableError.Msg
			} else {
				var deleteTable int
				for k, hft := range floor.Tables {
					if deleteTable+item.TableNum == 0 {
						break
					}
					hasUser := false
					for _, v := range hft.UserWithOnline {
						if v.Uid != 0 {
							hasUser = true
							break
						}
					}
					if hasUser {
						continue
					} else {
						deleteTable++
						delete(floor.Tables, k)
					}
				}
			}
		}
		floor.DataLock.Unlock()
	}
	var mixFloor []*HouseFloor
	if house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				mixFloor = append(mixFloor, floor)
			}
		}
		if sortTableByCreateTime(mixFloor) {
			xlog.Logger().Warnf("sortTableByCreateTime after Proto_ClubHouseTableChange")
			house.SyncTablesWithSorted()
		}
	}
	go func() {
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseMixFloorTableChangeNTF,
			&static.Msg_Null{})
		house.FloorLock.RLock()
		for _, f := range house.Floors {
			f.RedisPub(consts.MsgTypeHouseMixFloorTableChangeNTF, &static.Msg_Null{})
		}
		house.FloorLock.RUnlock()
	}()
	return xerrors.SuccessCode, nil

}

func Proto_ClubHouseMsg(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseMsg)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if req.Start >= req.End {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	msg, err := house.GetMsg(req.Start, req.End, req.Flag)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, &msg
}

func Proto_ClubHouseTableInviteAck(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHTableInvitrResp)
	// 暂时只处理屏蔽邀请需求，后面可能会加上通知
	if req.Notips {
		house := GetClubMgr().GetClubHouseByHId(req.Hid)
		if house == nil {
			return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
		}
		mem := house.GetMemByUId(p.Uid)
		if mem == nil {
			return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
		}
		if err := mem.IgnoreInvite(); err != nil {
			xlog.Logger().Errorln("house mem ignore invite error:", err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}
	return xerrors.SuccessCode, nil
}

// 编辑包厢禁止同桌成员列表
func Proto_ClubHouseMemberTableLimitList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.MsgHouseMemberTableLimit)
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorBanTableSitAt)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if req.PBegin < 0 {
		req.PBegin = 0
	}
	// 不包含end位，需要多取一位
	req.PEnd++
	if req.PEnd < req.PBegin {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	//house := GetClubMgr().GetClubHouseByHId(req.Hid)

	if req.GroupID <= 0 {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	arr := house.GetMemSimple(false)
	seMem := []HouseMember{}
	if req.Param == "" {
		seMem = arr
	} else {
		for _, mem := range arr {
			// ID 包含
			suid := static.HF_I64toa(mem.UId)
			if strings.Contains(suid, req.Param) {
				seMem = append(seMem, mem)
				continue
			}
			if strings.Contains(mem.NickName, req.Param) {
				seMem = append(seMem, mem)
				continue
			}
		}
	}
	// 存在
	limit := []HouseMember{}
	// 不存在
	nolimit := []HouseMember{}
	for _, mem := range seMem {
		// 判定是否存在列表内
		s := mem
		if house.CheckUserIsLimitInGroup(req.GroupID, s.UId) {

			limit = append(limit, s)
			continue
		}
		nolimit = append(nolimit, s)
	}
	// 排序
	sort.Sort(HouseMemberItemWrapper{limit, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})
	sort.Sort(HouseMemberItemWrapper{nolimit, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})
	conditionArr := append(limit, nolimit...)

	var ack static.Msg_HC_HouseMemTableLimitList
	ack.Totalnum = len(conditionArr)
	// 分页超出范围
	if len(conditionArr) == 0 || len(conditionArr) < req.PBegin {
		return xerrors.SuccessCode, &ack
	}

	// 开始索引
	var idxBeg int
	var idxEnd int
	idxBeg = req.PBegin
	if req.PEnd > len(conditionArr) {
		idxEnd = len(conditionArr)
	} else {
		idxEnd = req.PEnd
	}
	conditionArr = conditionArr[idxBeg:idxEnd]
	arrLen := len(conditionArr)
	ack.FMems = make([]static.Msg_HouseMemberItem, 0, arrLen)
	for _, hmem := range conditionArr {
		var titem static.Msg_HouseMemberItem
		titem.UId = hmem.UId
		titem.UOnline = hmem.IsOnline
		titem.UName = hmem.NickName
		titem.URole = hmem.URole
		titem.UVitamin = static.SwitchVitaminToF64(hmem.UVitamin)
		titem.UPartner = hmem.Partner
		titem.URemark = hmem.URemark
		titem.UUrl = hmem.ImgUrl
		titem.UGender = hmem.Sex
		titem.UJoinTime = hmem.ApplyTime
		titem.Limit = house.CheckUserIsLimitInGroup(req.GroupID, hmem.UId)
		ack.FMems = append(ack.FMems, titem)
	}

	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseTableLimitGroupAdd(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseTableLimitGroupAdd)
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hmem := house.GetMemByUId(p.Uid)
	isOK, _ := GetMRightMgr().CheckRight(hmem, MinorBanTableSitAt)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if hmem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}
	_, err := house.AddTableLimitGroup()
	if err != xerrors.RespOk {
		return err.Code, err.Msg
	}
	return xerrors.SuccessCode, house.GetTableLimitInfo()
}

func Proto_ClubHouseTableLimitGroupRemove(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseTableLimitGroupRemove)
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hmem := house.GetMemByUId(p.Uid)
	isOK, _ := GetMRightMgr().CheckRight(hmem, MinorBanTableSitAt)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if hmem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}
	err := house.RemoveTableLimitGroup(req.GroupID)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, house.GetTableLimitInfo()
}

func Proto_ClubHouseTableLimitUserAdd(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseTableLimitUserAdd)
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hmem := house.GetMemByUId(p.Uid)
	isOK, _ := GetMRightMgr().CheckRight(hmem, MinorBanTableSitAt)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if hmem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}
	err := house.AddTableLimitUser(req.GroupID, req.Uid)
	if err != xerrors.RespOk {
		return err.Code, err.Msg
	}
	msg := fmt.Sprintf("ID:%d将用户ID:%d 添加至禁止同桌列表%d", p.Uid, req.Uid, req.GroupID)
	CreateClubMassage(house.DBClub.Id, p.Uid, MemLimitTable, msg)
	return xerrors.SuccessCode, house.GetTableLimitInfo()
}

func Proto_ClubHouseTableLimitUserRemove(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseTableLimitUserAdd)
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	hmem := house.GetMemByUId(p.Uid)
	isOK, _ := GetMRightMgr().CheckRight(hmem, MinorBanTableSitAt)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//if hmem.Lower(constant.ROLE_ADMIN) {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}
	err := house.RemoveTableLimitUser(req.GroupID, req.Uid)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	msg := fmt.Sprintf("ID:%d将用户ID:%d 移除禁止同桌列表%d", p.Uid, req.Uid, req.GroupID)
	CreateClubMassage(house.DBClub.Id, p.Uid, MemLimitTable, msg)
	return xerrors.SuccessCode, house.GetTableLimitInfo()
}

func Proto_ClubHouseTableLimitInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseTableLimitGroupAdd)
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorBanTableSitAt)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	return xerrors.SuccessCode, house.GetTableLimitInfo()
}

// 2人桌子禁止同桌不生效 的设置
func Proto_ClubHouse2PTableLimitNotEffect(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouse2PTableLimitNotEffectSet)
	// 暂时使用 MinorJoinReviewed 权限代替
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorJoinReviewed)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	custerr := house.Option2PTableLimitNotEffect(req.Sta)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}
	return xerrors.SuccessCode, nil
}

func Proto_ProtoGetHmemByid(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgHouseUserLimitGame)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	_, _, mem, cuserr := inspectClubFloorMemberWithRight(req.HID, -1, req.UID, consts.ROLE_MEMBER, MinorRightNull)
	if cuserr != xerrors.RespOk {
		return cuserr.Code, cuserr.Msg
	}
	hp, _ := GetDBMgr().GetDBrControl().GetPerson(mem.UId)
	if hp == nil {
		return xerrors.UserExistError.Code, xerrors.UserExistError.Msg
	}
	res := static.GameMember{}
	res.Uid = hp.Uid
	res.Uname = hp.Nickname
	res.URole = mem.URole
	res.Gender = hp.Sex
	res.IsOnline = mem.IsOnline
	res.UUrl = hp.Imgurl
	res.Hid = req.HID
	res.Partner = mem.Partner
	return xerrors.SuccessCode, &res
}

// 包厢合并包厢意向信息输入
func Proto_ClubHouseMergeIntention(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseOwnerInfo)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if house.DBClub.UId != p.Uid {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if house.DBClub.HId == req.THId {
		xerr := xerrors.NewXError("合并包厢申请不能向自己的包厢发送申请")
		return xerr.Code, xerr.Msg
	}

	thouse := GetClubMgr().GetClubHouseByHId(req.THId)
	if thouse == nil {
		xerr := xerrors.NewXError("包厢号输入错误。")
		return xerr.Code, xerr.Msg
	}

	if mem := thouse.GetMemByUId(p.Uid); mem != nil {
		if mem.Upper(consts.ROLE_APLLY) {
			xerr := xerrors.NewXError("您已是该包厢成员，请先退出对方的包厢再申请合并包厢。")
			return xerr.Code, xerr.Msg
		} else if mem.URole == consts.ROLE_APLLY {
			xerr := xerrors.NewXError("当前处于加入包厢申请中，无法合并包厢。")
			return xerr.Code, xerr.Msg
		} else if mem.URole == consts.ROLE_BLACK {
			xerr := xerrors.NewXError("当前处于对方黑名单中，无法合并包厢。")
			return xerr.Code, xerr.Msg
		}
	}

	if thouse.DBClub.UId != req.Owner {
		xerr := xerrors.NewXError("包厢号和盟主ID不匹配，请重新输入。")
		return xerr.Code, xerr.Msg
	}

	if thouse.IsBeenMerged() {
		xerr := xerrors.NewXError("该包厢已经被合并。")
		return xerr.Code, xerr.Msg
	}

	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	if thouse.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	if house.IsBeenMerged() {
		xerr := xerrors.NewXError("当前包厢已被合并，无法申请合并。")
		return xerr.Code, xerr.Msg
	}

	if house.IsMerged() {
		xerr := xerrors.NewXError("您的包厢已合并过其他包厢，不能再申请合并。")
		return xerr.Code, xerr.Msg
	}

	// 查找包厢的于合并包厢相关的记录
	houseMergeLogs := make([]*models.HouseMergeLog, 0)
	err := GetDBMgr().GetDBmControl().
		Where("swallowed = ?", house.DBClub.Id).
		Find(&houseMergeLogs).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			hml := new(models.HouseMergeLog)
			hml.Devourer = thouse.DBClub.Id
			hml.Swallowed = house.DBClub.Id
			hml.Sponsor = house.DBClub.Id
			hml.MergeState = models.HouseMergeStateHavmind
			if err = GetDBMgr().GetDBmControl().Create(hml).Error; err != nil {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			} else {
				return xerrors.SuccessCode, nil
			}
		} else {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	// 分析记录
	for _, hml := range houseMergeLogs {
		if hml == nil {
			continue
		}
		if hml.MergeState > models.HouseMergeStateRevoked {
			if hml.Devourer == thouse.DBClub.Id {
				xerr := xerrors.NewXError("您已向该包厢发送过合并包厢申请，请等待处理。")
				return xerr.Code, xerr.Msg
			} else {
				xerr := xerrors.NewXError("不可同时像多个包厢发送申请。")
				return xerr.Code, xerr.Msg
			}
		}
	}

	flag := false
	// 再次分析记录
	for _, hml := range houseMergeLogs {
		if hml == nil {
			continue
		}
		// 有与目标圈的记录
		if hml.Devourer == thouse.DBClub.Id {
			flag = true
			hml.MergeState = models.HouseMergeStateHavmind
			hml.Sponsor = house.DBClub.Id
			hml.UpdatedAt = time.Now()
			if err := GetDBMgr().GetDBmControl().Save(hml).Error; err != nil {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			} else {
				return xerrors.SuccessCode, nil
			}
		}
	}

	// 没有找到合法的记录 则创建一条记录
	if !flag {
		hml := new(models.HouseMergeLog)
		hml.Devourer = thouse.DBClub.Id
		hml.Swallowed = house.DBClub.Id
		hml.Sponsor = house.DBClub.Id
		hml.MergeState = models.HouseMergeStateHavmind
		if err = GetDBMgr().GetDBmControl().Create(hml).Error; err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		} else {
			return xerrors.SuccessCode, nil
		}
	}

	return xerrors.SuccessCode, nil
}

// 包厢合并包厢信息校验
func Proto_ClubHouseMergeCheck(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	ack := &static.MsgHouseMergeCheck{
		HId:      req.HId,
		IsFrozen: house.DBClub.IsFrozen,
		IsGaming: house.CheckInGameTable(),
	}

	return xerrors.SuccessCode, ack
}

// 包厢合并包厢请求
func Proto_ClubHouseMergeRequest(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)

	house := GetClubMgr().GetClubHouseByHId(req.HId)

	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if house.DBClub.UId != p.Uid {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if !house.DBClub.IsFrozen {
		return xerrors.NotFrozenError.Code, xerrors.NotFrozenError.Msg
	}

	if house.CheckInGameTable() {
		xerr := xerrors.NewXError("当前大厅有玩家已入桌，无法申请合并包厢。")
		return xerr.Code, xerr.Msg
	}

	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	if house.IsBeenMerged() {
		xerr := xerrors.NewXError("当前包厢已被合并，无法申请合并。")
		return xerr.Code, xerr.Msg
	}

	if house.IsMerged() {
		xerr := xerrors.NewXError("您的包厢已合并过其他圈，不能再申请合并。")
		return xerr.Code, xerr.Msg
	}

	// 效验申请
	// 查找包厢合并包厢相关的记录
	houseMergeLogs := make([]*models.HouseMergeLog, 0)
	err := GetDBMgr().GetDBmControl().
		Where("swallowed = ?", house.DBClub.Id).
		Find(&houseMergeLogs).Error

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	for _, hml := range houseMergeLogs {
		if hml == nil {
			continue
		}
		if hml.MergeState > models.HouseMergeStateRevoked {
			xerr := xerrors.NewXError("不可同时向多个包厢发送申请。")
			return xerr.Code, xerr.Msg
		}
	}

	hml := new(models.HouseMergeLog)

	err = GetDBMgr().GetDBmControl().
		Where("swallowed = ?", house.DBClub.Id).
		Where("sponsor = ?", house.DBClub.Id).
		Where("merge_state = ?", models.HouseMergeStateHavmind).
		Order("updated_at desc").First(hml).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			xerr := xerrors.NewXError("未找到相关包厢。")
			return xerr.Code, xerr.Msg
		}
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	thouse := GetClubMgr().GetClubHouseById(hml.Devourer)

	if thouse == nil {
		xerr := xerrors.NewXError("目标包厢不存在。")
		return xerr.Code, xerr.Msg
	}

	if thouse.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	if thouse.IsBeenMerged() {
		xerr := xerrors.NewXError("不可向已经被合并的包厢发送合并请求。")
		return xerr.Code, xerr.Msg
	}

	hml.MergeState = models.HouseMergeStateWaiting
	hml.Sponsor = house.DBClub.Id
	err = GetDBMgr().GetDBmControl().Save(hml).Error

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 向老盟主推送消息
	if thouseOwner := GetPlayerMgr().GetPlayer(thouse.DBClub.UId); thouseOwner != nil {
		thouseOwner.SendMsg(consts.MsgTypeHouseMerge_NTF, &static.MsgHouseMergeRequest{HId: thouse.DBClub.HId, MergeHId: house.DBClub.HId})
	}

	return xerrors.SuccessCode, nil
}

// 包厢合并包厢请求撤销
func Proto_ClubHouseMergeReqRevoke(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMerge)
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	thouse := GetClubMgr().GetClubHouseByHId(req.THId)
	if thouse == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	hml := new(models.HouseMergeLog)
	err := GetDBMgr().GetDBmControl().
		Where("swallowed = ?", house.DBClub.Id).
		Where("sponsor = ?", house.DBClub.Id).
		Where("devourer = ?", thouse.DBClub.Id).
		Where("merge_state = ?", models.HouseMergeStateWaiting).
		First(hml).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			xlog.Logger().Error("撤销合并包厢请求时 记录未找到")
		}
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	hml.MergeState = models.HouseMergeStateRevoked
	err = GetDBMgr().GetDBmControl().Save(hml).Error
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 給对方盟主推送撤销申请的消息
	if towner := GetPlayerMgr().GetPlayer(thouse.DBClub.UId); towner != nil {
		towner.SendMsg(consts.MsgTypeHouseMergeReqRevokeNtf, static.MsgHouseMergeRevokeOk{
			Hid:  thouse.DBClub.HId,
			THid: house.DBClub.HId,
			Msg:  fmt.Sprintf("包厢(%d)盟主刚刚撤销了向您的包厢(%d)发起的合并包厢请求。", house.DBClub.HId, thouse.DBClub.HId),
		})
	}

	if house.DBClub.IsFrozen {
		xerr := house.OptionFrozen(false)
		if xerr != nil {
			return xerr.Code, xerr.Msg
		}
	}

	return xerrors.SuccessCode, nil
}

// 包厢合并包厢响应
func Proto_ClubHouseMergeResponse(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMergeRsp)

	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if p.Uid != house.DBClub.UId {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	thouse := GetClubMgr().GetClubHouseByHId(req.THId)
	if thouse == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	hml := new(models.HouseMergeLog)

	err := GetDBMgr().GetDBmControl().
		Where("swallowed = ?", thouse.DBClub.Id).
		Where("sponsor = ?", thouse.DBClub.Id).
		Where("devourer = ?", house.DBClub.Id).
		Where("merge_state = ?", models.HouseMergeStateWaiting).First(hml).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			xerr := xerrors.NewXError("对方已撤销合并包厢申请。")
			return xerr.Code, xerr.Msg
		}
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	if req.Result {
		if err := GetClubMgr().HouseMerge(house.DBClub.HId, thouse.DBClub.HId); err.Error() == nil {
			thouse.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseMergeOk, &static.MsgHouseMergeRevokeOk{
				Hid:    thouse.DBClub.HId,
				THid:   house.DBClub.HId,
				THName: house.DBClub.Name,
				Msg:    fmt.Sprintf("包厢(%d)已被合并，请在包厢列表进入新的包厢。", thouse.DBClub.HId),
			})
			GetPlayerMgr().SendNotify(thouse.DBClub.UId, true,
				"合并包厢成功",
				fmt.Sprintf("您的包厢(%d)向包厢(%d)\n发起的合并包厢申请于%s被对方盟主确认并同意。\n\n您有%d个玩家同时也是被合包厢成员\n将不会与您成为绑定关系",
					thouse.DBClub.HId, house.DBClub.HId, time.Now().Format(static.TIMEFORMAT), err.NumRepeat()))
		} else {
			xlog.Logger().Error("合并包厢失败", err.Error().Error())
			if _, ok := err.Error().(DBError); ok {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			} else {
				return xerrors.ResultErrorCode, err.Error().Error()
			}
		}
	} else {
		GetPlayerMgr().SendNotify(thouse.DBClub.UId, true,
			"合并包厢失败",
			fmt.Sprintf("您的包厢(%d)向包厢(%d)\n发起的合并包厢申请于%s\n被对方盟主拒绝。", thouse.DBClub.HId, house.DBClub.HId, time.Now().Format(static.TIMEFORMAT)))
		hml.MergeState = models.HouseMergeStateRefused
		if err := GetDBMgr().GetDBmControl().Save(hml).Error; err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		if thouse.DBClub.IsFrozen {
			xerr := thouse.OptionFrozen(false)
			if xerr != nil {
				return xerr.Code, xerr.Msg
			}
		}
	}
	return xerrors.SuccessCode, nil
}

// 包厢合并包厢记录
func Proto_ClubHouseMergeRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.包厢不存在。 req：%+v", req)
		return xerrors.SuccessCode, nil
	}

	member := house.GetMemByUId(p.Uid)
	if member == nil {
		xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.非包厢成员。 req：%+v", req)
		return xerrors.SuccessCode, nil
	}
	isOK, _ := GetMRightMgr().CheckRight(member, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	// 当前请求的dhid标识
	dhid := house.DBClub.Id

	if _, ok := member.CheckRef(); ok {
		dhid = member.Ref
	} else {
		if house.DBClub.UId != p.Uid {
			xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.无效权限。 req：%+v", req)
			return xerrors.SuccessCode, nil
		}
	}

	housereq := GetClubMgr().GetClubHouseById(dhid)
	if housereq == nil {
		xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.请求包厢不存在。 req：%+v", req)
		return xerrors.SuccessCode, nil
	}

	hmls := make([]*models.HouseMergeLog, 0)
	if err := GetDBMgr().GetDBmControl().Where("swallowed = ?", dhid).Or("devourer = ?", dhid).Order("updated_at DESC").Find(&hmls).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerrors.SuccessCode, nil
		}
		xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.数据库操作异常:%v。 req：%+v", err, req)
		return xerrors.SuccessCode, nil
	}

	ack := static.MsgHouseMergeRecords{}
	ack.Items = make([]static.MsgHouseMergeRecordItem, 0)

	for _, hml := range hmls {
		if hml == nil {
			continue
		}
		if hml.MergeState <= models.HouseMergeStateHavmind {
			continue
		}

		var item static.MsgHouseMergeRecordItem
		houseSwa := GetClubMgr().GetClubHouseById(hml.Swallowed)
		if houseSwa == nil {
			continue
		}

		houseDev := GetClubMgr().GetClubHouseById(hml.Devourer)
		if houseDev == nil {
			continue
		}

		item.Hid = housereq.DBClub.HId

		if dhid == hml.Devourer {
			item.THid = houseSwa.DBClub.HId
		} else {
			item.THid = houseDev.DBClub.HId
		}
		// }

		switch hml.MergeState {
		case models.HouseMergeStateRevoked:
			item.State = models.HouseMergeClientStateReqRvk
		case models.HouseMergeStateWaiting:
			if dhid == hml.Devourer {
				item.State = models.HouseMergeClientStateRsp
			} else {
				item.State = models.HouseMergeClientStateReq
			}
		case models.HouseMergeStateRefused:
			if dhid == hml.Devourer {
				item.State = models.HouseMergeClientStateRspRef
			} else {
				item.State = models.HouseMergeClientStateReqRef
			}
		case models.HouseMergeStateAproved:
			if dhid == hml.Devourer {
				item.State = models.HouseMergeClientStateRspApr
			} else {
				item.State = models.HouseMergeClientStateReqApr
			}
		case models.HouseMergeStateRevoking:
			if dhid == hml.Sponsor {
				item.State = models.HouseMergeClientStateRvkReq
			} else {
				item.State = models.HouseMergeClientStateRvkRsp
			}
		case models.HouseMergeStateRevokeRef:
			if dhid == hml.Sponsor {
				item.State = models.HouseMergeClientStateRvkReqRef
			} else {
				item.State = models.HouseMergeClientStateRvkRspRef
			}
		}
		item.At = hml.UpdatedAt.Unix()
		ack.Items = append(ack.Items, item)
	}
	return xerrors.SuccessCode, &ack
}

// 包厢撤销合并包厢请求
func Proto_ClubHouseRevokeRequest(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMerge)

	// 发起者house
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	member := house.GetMemByUId(p.Uid)
	if member == nil {
		xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.非包厢成员。 req：%+v", req)
		return xerrors.SuccessCode, nil
	}
	isOK, _ := GetMRightMgr().CheckRight(member, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if house.DBClub.UId != p.Uid {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 目标/对方house
	thouse := GetClubMgr().GetClubHouseByHId(req.THId)
	if thouse == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	if thouse.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	var devourer, swallowed int64
	// 自己申请撤销合并包厢
	if house.IsDevourer(thouse.DBClub.Id) {
		devourer = thouse.DBClub.Id
		swallowed = house.DBClub.Id

	} else if thouse.IsDevourer(house.DBClub.Id) {
		devourer = house.DBClub.Id
		swallowed = thouse.DBClub.Id
	} else {
		xerr := xerrors.NewXError("对方已撤销合并包厢。")
		return xerr.Code, xerr.Msg
	}

	// 找到合并包厢记录
	hml := new(models.HouseMergeLog)
	err := GetDBMgr().GetDBmControl().
		Where("devourer = ?", devourer).
		Where("swallowed = ?", swallowed).
		Where("merge_state >= ?", models.HouseMergeStateAproved).
		First(hml).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			xerr := xerrors.NewXError("未找到合并包厢记录。")
			return xerr.Code, xerr.Msg
		}
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	if hml.MergeState == models.HouseMergeStateRevoking {
		xerr := xerrors.NewXError("已申请撤销合并包厢，请耐心等待对方同意。")
		return xerr.Code, xerr.Msg
	}
	// 更新合并包厢记录
	hml.MergeState = models.HouseMergeStateRevoking
	hml.Sponsor = house.DBClub.Id
	dberr := GetDBMgr().GetDBmControl().Save(hml).Error
	if dberr != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	if tOwner := GetPlayerMgr().GetPlayer(thouse.DBClub.UId); tOwner != nil {
		tOwner.SendMsg(consts.MsgTypeHouseRevoke_NTF, &static.MsgHouseMergeRequest{HId: thouse.DBClub.HId, MergeHId: house.DBClub.HId})
	}
	return xerrors.SuccessCode, req
}

// 包厢撤销合并包厢响应
func Proto_ClubHouseRevokeResonse(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMergeRsp)
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	member := house.GetMemByUId(p.Uid)
	if member == nil {
		xlog.Logger().Warnf("Proto_ClubHouseMergeRecord.非包厢成员。 req：%+v", req)
		return xerrors.SuccessCode, nil
	}
	isOK, _ := GetMRightMgr().CheckRight(member, MinorMergeTea)
	if !isOK {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if p.Uid != house.DBClub.UId {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	thouse := GetClubMgr().GetClubHouseByHId(req.THId)

	if thouse == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	var houseDev, houseSwa *Club
	// 自己申请撤销合并包厢
	if house.IsDevourer(thouse.DBClub.Id) {
		houseDev = thouse
		houseSwa = house
	} else if thouse.IsDevourer(house.DBClub.Id) {
		houseDev = house
		houseSwa = thouse
	} else {
		xerr := xerrors.NewXError("对方已撤销合并包厢。")
		return xerr.Code, xerr.Msg
	}

	_, e := GetMRightMgr().deleteRightByHidUid(int(houseDev.DBClub.Id), houseSwa.DBClub.UId)
	if e != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	hml := new(models.HouseMergeLog)

	err := GetDBMgr().GetDBmControl().
		Where("swallowed = ?", houseSwa.DBClub.Id).
		Where("devourer = ?", houseDev.DBClub.Id).
		Where("merge_state = ?", models.HouseMergeStateRevoking).First(hml).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			xerr := xerrors.NewXError("未找到撤销合并包厢申请记录。")
			return xerr.Code, xerr.Msg
		}
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	if req.Result {
		if err := GetClubMgr().HouseRevoke(houseDev.DBClub.HId, houseSwa.DBClub.HId); err.Error() == nil {
			GetPlayerMgr().SendNotify(thouse.DBClub.UId, true,
				"撤销合并包厢成功",
				fmt.Sprintf("您的包厢(%d)向包厢(%d)\n发起的撤销合并包厢申请于%s被对方盟主确认并同意",
					thouse.DBClub.HId, house.DBClub.HId, time.Now().Format(static.TIMEFORMAT)))
		} else {
			xlog.Logger().Error("撤销合并包厢失败", err.Error().Error())
			if _, ok := err.Error().(DBError); ok {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			} else {
				return xerrors.ResultErrorCode, err.Error().Error()
			}
		}
	} else {
		GetPlayerMgr().SendNotify(thouse.DBClub.UId, true,
			"撤销合并包厢失败",
			fmt.Sprintf("您的包厢(%d)向包厢(%d)\n发起的撤销合并包厢申请于%s\n被对方盟主拒绝",
				thouse.DBClub.HId, house.DBClub.HId, time.Now().Format(static.TIMEFORMAT)))
		hml.MergeState = models.HouseMergeStateRevokeRef
		if err := GetDBMgr().GetDBmControl().Save(hml).Error; err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	return xerrors.SuccessCode, req
}

func Proto_ClubHouseJoinSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseJoinTableSet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorSetJoinTable)

	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house.DBClub.OnlyQuickJoin != req.OnlyQuick {
		house.DBClub.OnlyQuickJoin = req.OnlyQuick
		house.flush()
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseJoinTableChangeNtf, req)

	}
	return xerrors.SuccessCode, nil
}

// 包厢邀请加入请求
func ProtoHouseInviteJoinReq(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseJoinInvite)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	// return xerrors.SuccessCode, nil

	house, _, _, xerr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xerr != xerrors.RespOk {
		return xerr.Code, xerr.Msg
	}

	if !GetDBMgr().GetDBrControl().Exists(fmt.Sprintf("user_%d", req.TUid)) {
		xerr = xerrors.NewXError("ID输入有误，请重新输入。")
		return xerr.Code, xerr.Msg
	}

	if tmem := house.GetMemByUId(req.TUid); tmem != nil {
		switch tmem.URole {
		case consts.ROLE_APLLY:
			xerr = xerrors.NewXError("请等待管理员审核。")
			return xerr.Code, xerr.Msg
		case consts.ROLE_BLACK:
			xerr = xerrors.NewXError("该用户已被盟主加入黑名单。")
			return xerr.Code, xerr.Msg
		default:
			xerr = xerrors.NewXError("该玩家已经是本包厢成员。")
			return xerr.Code, xerr.Msg
		}
	}

	if IsHouseInviteBlack(req.TUid, p.Uid) {
		xlog.Logger().Warnf("玩家%d设置了不再接受玩家%d的入包厢邀请。", req.TUid, p.Uid)
		return xerrors.SuccessCode, nil
	}

	invitationLetter := &static.Msg_HC_HouseInviteJoin{
		HId:       house.DBClub.HId,
		IsHidHide: house.DBClub.IsHidHide,
	}

	invitationLetter.Inviter.Uid = p.Uid
	invitationLetter.Inviter.Nickname = p.Nickname
	invitationLetter.Inviter.Imgurl = p.Imgurl
	invitationLetter.Inviter.Gender = p.Sex

	if tp := GetPlayerMgr().GetPlayer(req.TUid); tp != nil {
		tp.SendMsg(consts.MsgTypeHouseJoinInviteRecv, invitationLetter)
	} else {
		if err := AddUnreadHouseJoinInvite(req.TUid, invitationLetter); err != nil {
			xlog.Logger().Error("AddUnreadHouseJoinInvite", err)
		}
	}
	return xerrors.SuccessCode, nil
}

// 包厢邀请加入响应
func ProtoHouseInviteJoinRsp(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseJoinInviteRsp)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if req.Notips {
		if err := AddHouseInviteBlack(p.Uid, req.Inviter); err != nil {
			xlog.Logger().Error("AddHouseInviteBlack", err)
		}
	}
	return xerrors.SuccessCode, nil
}

// 包厢弹窗内容编辑及开关设置
func ProtoHouseDialogEdit(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseDialogEdit)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, _, xerr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if xerr != xerrors.RespOk {
		return xerr.Code, xerr.Msg
	}

	// 敏感词检查
	req.Content = static.CheckSensitiveWord(GetDBMgr().GetDBrControl().RedisV2, req.Content)

	house.DBClub.Dialog = req.Content
	house.DBClub.DialogActive = req.Active
	house.flush()

	var ntf static.Ntf_HC_HouseBaseNNmodify
	ntf.HId = house.DBClub.HId
	ntf.HName = house.DBClub.Name
	ntf.HNotify = house.DBClub.Notify
	ntf.HDialog = house.DBClub.Dialog
	ntf.HDialogActive = house.DBClub.DialogActive
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseBaseNNModify_Ntf, &ntf)

	return xerrors.SuccessCode, nil
}

// 通过邀请码加入包厢
func ProtoHouseJoinByInviteCode(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_HC_HouseJoinByCode)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	partnerInviteCode, err := GetDBMgr().SelectPartnerByInviteCode(req.Code)
	if err != nil {
		return xerrors.HouseInviteCodeError.Code, xerrors.HouseInviteCodeError.Msg
	}

	hid := partnerInviteCode.UsedHid
	partnerId := partnerInviteCode.UseUid

	house := GetClubMgr().GetClubHouseById(hid)
	if house == nil {
		return xerrors.HouseInviteCodeError.Code, xerrors.HouseInviteCodeError.Msg
	}

	partner := GetClubMgr().GetHouseMember(hid, partnerId)

	if partner == nil {
		return xerrors.HouseInviteCodeError.Code, xerrors.HouseInviteCodeError.Msg
	} else if partner.Partner != 1 {
		return xerrors.HouseInviteCodeError.Code, xerrors.HouseInviteCodeError.Msg
	}

	if mem := house.GetMemByUId(p.Uid); mem != nil {
		switch mem.URole {
		case consts.ROLE_APLLY:
			if mem.Partner > 0 && partnerId == mem.Partner {
				return xerrors.ResultErrorCode, "已向该包厢发送邀请，请等待管理员审核。"
			}
		case consts.ROLE_BLACK:
			return xerrors.ResultErrorCode, "您已被盟主加入黑名单。"
		default:
			return xerrors.ResultErrorCode, "您已是此包厢成员。"
		}
	}

	invitationLetter := &static.Msg_HC_HouseInviteJoin{
		HId:       house.DBClub.HId,
		IsHidHide: house.DBClub.IsHidHide,
	}
	invitationLetter.Inviter.Uid = partner.UId
	invitationLetter.Inviter.Nickname = partner.NickName
	invitationLetter.Inviter.Imgurl = partner.ImgUrl
	invitationLetter.Inviter.Gender = partner.Sex
	return xerrors.SuccessCode, invitationLetter
}

// 修改防作弊包厢显示桌数
func ProtoHouseTblShowCountEdit(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseTableShowCount)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if req.Count < 0 {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, _, xe := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorSetTableNum)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	house.DBClub.TableShowCount = req.Count
	if req.MaxTableNum < req.MinTableNum {
		req.MaxTableNum, req.MinTableNum = req.MinTableNum, req.MaxTableNum
	}
	xlog.Logger().Warnf("ProtoHouseTblShowCountEdit.修改防作弊包厢显示桌数:%+v", req)
	house.DBClub.MinTableNum = req.MinTableNum
	house.DBClub.MaxTableNum = req.MaxTableNum
	house.flush()
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseTblShowCountChangeNtf, &static.Msg_HC_HouseTableShowCount{
		HId:         house.DBClub.HId,
		Count:       house.DBClub.TableShowCount,
		MinTableNum: house.DBClub.MinTableNum,
		MaxTableNum: house.DBClub.MaxTableNum,
	})
	return xerrors.SuccessCode, nil
}

func ProtoHouseFloorWaitingNumSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseFloorWaitingNum)
	house, _, _, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	house.FloorLock.CustomLock()
	defer house.FloorLock.CustomUnLock()
	for k, floor := range house.Floors {
		if num, ok := req.FloorsMap[k]; ok {
			if num < consts.HOUSEAISUPERMIN || num > consts.HOUSEAISUPERMAX {
				if floor.AiSuperNum != 0 {
					floor.AiSuperNum = 0
					if err := GetDBMgr().HouseFloorUpdate(floor); err != nil {
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
				}
			} else {
				if floor.AiSuperNum != num {
					floor.AiSuperNum = num
					if err := GetDBMgr().HouseFloorUpdate(floor); err != nil {
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
				}
			}
		}
	}
	return xerrors.SuccessCode, nil
}

func ProtoHouseFloorWaitingNumGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house, _, _, xe := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	var ack static.Msg_CH_HouseFloorWaitingNum
	ack.Hid = house.DBClub.HId
	ack.FloorsMap = make(map[int64]int)
	house.FloorLock.RLock()
	defer house.FloorLock.RUnlock()
	for _, floor := range house.Floors {
		ack.FloorsMap[floor.Id] = floor.AiSuperNum
	}
	return xerrors.SuccessCode, &ack
}

func ProtoHouseGroupAdd(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupAdd)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !mem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	gid, xe := house.AddUserGroup(p.Uid)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	msg := fmt.Sprintf("用户ID:%d添加用户分组%d", p.Uid, gid)
	CreateClubMassage(house.DBClub.Id, p.Uid, AddUserGroup, msg)
	return xerrors.SuccessCode, house.GetUserGroupInfo(p.Uid)
}

func ProtoHouseGroupDel(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupDel)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if mem.Partner != 1 {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	err := house.RemoveUserGroup(p.Uid, req.GroupId)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	msg := fmt.Sprintf("用户ID:%d删除用户分组%d", p.Uid, req.GroupId)
	CreateClubMassage(house.DBClub.Id, p.Uid, AddUserGroup, msg)
	return xerrors.SuccessCode, house.GetUserGroupInfo(p.Uid)
}

func ProtoHouseUserGroupAddUser(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupUserAdd)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !mem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	addMem := house.GetMemByUId(req.Uid)
	if addMem == nil || addMem.Partner != p.Uid {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	xe = house.AddUserGroupUser(p.Uid, req.GroupId, req.Uid)
	if xe != xerrors.RespOk && xe != nil {
		return xe.Code, xe.Msg
	}
	msg := fmt.Sprintf("用户ID:%d将用户ID:%d添加至分组%d", p.Uid, req.Uid, req.GroupId)
	CreateClubMassage(house.DBClub.Id, p.Uid, AddGroupUser, msg)
	return xerrors.SuccessCode, house.GetUserGroupInfo(p.Uid)
}

func ProtoHouseUserGroupDelUser(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupUserAdd)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !mem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	err := house.RemoveUserGroupUser(p.Uid, req.GroupId, req.Uid)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	msg := fmt.Sprintf("用户ID:%d将用户ID:%d添加至分组%d", p.Uid, req.Uid, req.GroupId)
	CreateClubMassage(house.DBClub.Id, p.Uid, AddGroupUser, msg)
	return xerrors.SuccessCode, house.GetUserGroupInfo(p.Uid)
}

func ProtoHouseUserGroupInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupAdd)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !mem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	return xerrors.SuccessCode, house.GetUserGroupInfo(p.Uid)
}

func ProtoHouseGroupUserList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupUserList)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !mem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	house.AddUserGroupLock.CustomLock()
	defer house.AddUserGroupLock.CustomUnLock()
	uidmap := house.UserGroup[p.Uid]
	if uidmap == nil {
		return xerrors.SuccessCode, nil
	}
	uids := uidmap[req.GroupId]
	if uids == nil || len(uids) == 0 {
		return xerrors.SuccessCode, nil
	}
	ack := static.GroupUserListInfo{}
	ack.Start = req.Start
	mems := house.SearchUserToMap(req.SearchKey)
	if len(mems) < req.Start {
		return xerrors.SuccessCode, ack
	}
	dest := []*static.LimitUserInfo{}

	for _, u := range uids {
		hmem := mems[u]
		if hmem == nil {
			continue
		}
		var titem static.LimitUserInfo
		titem.UId = u
		titem.UName = hmem.NickName
		titem.UUrl = hmem.ImgUrl
		titem.UGender = hmem.Sex
		titem.Limit = true
		dest = append(dest, &titem)
	}
	sort.Sort(static.LimitUserSlie(dest))
	if len(dest) < req.Start {
		return xerrors.SuccessCode, ack
	}
	dest = dest[req.Start-1 : len(dest)]
	if len(dest) <= req.Count {
		ack.UserInfo = dest
	} else {
		ack.UserInfo = dest[0:req.Count]
	}
	return xerrors.SuccessCode, ack
}

func ProtoHouseGroupUserAddList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.HouseGroupUserList)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !mem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	totalUids := house.GetUIDsByPartner(nil, p.Uid)
	var leftUids []int64
	var inUids []int64
	house.AddUserGroupLock.CustomLock()
	defer house.AddUserGroupLock.CustomUnLock()
	uidmap := house.UserGroup[p.Uid]
	if uidmap == nil {
		leftUids = totalUids
	} else {
		for _, uids := range uidmap {
			inUids = append(inUids, uids...)
		}
	}
	if len(inUids) == 0 {
		leftUids = totalUids
	} else {
		for _, u := range totalUids {
			var in bool
			for _, lu := range inUids {
				if u == lu {
					in = true
				}
			}
			if !in {
				leftUids = append(leftUids, u)
			}
		}
	}
	ack := static.GroupUserListInfo{}
	ack.Start = req.Start
	ack.GroupId = req.GroupId
	mems := house.SearchUserToMap(req.SearchKey)
	if len(mems) < req.Start {
		return xerrors.SuccessCode, ack
	}
	dest := []*static.LimitUserInfo{}
	for _, u := range leftUids {
		hmem, ok := mems[u]
		if !ok {
			continue
		}
		var titem static.LimitUserInfo
		titem.UId = u
		titem.UName = hmem.NickName
		titem.UUrl = hmem.ImgUrl
		titem.UGender = hmem.Sex
		titem.Limit = true
		dest = append(dest, &titem)
	}
	sort.Sort(static.LimitUserSlie(dest))
	if len(dest) < req.Start {
		return xerrors.SuccessCode, ack
	}
	dest = dest[req.Start:len(dest)]
	if len(dest) <= req.Count {
		ack.UserInfo = dest
	} else {
		ack.UserInfo = dest[0:req.Count]
	}
	return xerrors.SuccessCode, ack
}

func Proto_ClubHouseGameSwitch(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseUserGameSwitch)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	err := house.GameSwitch(req.On, false)
	if err != nil {
		return xerrors.NewXError(err.Error()).Code, xerrors.NewXError(err.Error()).Msg
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgGameSwitchNtf, static.MsgHouseGameSwitchNtf{req.Hid, house.DBClub.GameOn, house.DBClub.AdminGameOn, house.DBClub.IsVitamin})
	return xerrors.SuccessCode, nil
}

func Proto_ClubHousePrizeInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHousePrizeInfo)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	var val []int64
	if req.Type == 1 {
		val = house.PrizeVal
	} else if req.Type == 2 {
		val = house.GroupPrizeVal
	}
	info := static.MsgHousePrizeSetS{Hid: req.Hid, Value: val, Type: req.Type}
	return xerrors.SuccessCode, &info
}

func Proto_ClubHousePrizeSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHousePrizeSetS)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if req.Type == 1 {
		house.PrizeVal = req.Value
		house.savePrize()
	} else if req.Type == 2 {
		house.GroupPrizeVal = req.Value
		house.saveGroupPrize()

	} else {
		return xerrors.ArgumentErrorCode, xerrors.ArgumentError.Msg
	}

	house.Broadcast(consts.ROLE_MEMBER, consts.MsgHousePrizeSetNtf, req)
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseLuckSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.HouseLuckSet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	for _, item := range req.Rewords {
		if item.Rank < 0 || item.Rank > 10 || item.Count < 0 {
			return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}
	}
	tx := GetDBMgr().GetDBmControl().Begin()
	var fail bool = true
	defer func() {
		if fail {
			tx.Rollback()
		}
	}()
	sql := `insert into luck_config(hid,actid,opuid,rank,count) values(?,?,?,?,?) on DUPLICATE KEY UPDATE count = ?,opuid = ?`
	for _, item := range req.Rewords {
		err := tx.Exec(sql, house.DBClub.Id, req.ActId, p.Uid, item.Rank, item.Count, item.Count, p.Uid).Error
		if err != nil {
			xlog.Logger().Errorf("db insert error：%v", err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}
	tx.Commit()
	fail = false
	GetDBMgr().Redis.Del(fmt.Sprintf(consts.REDIS_KEY_LUCKCONFIG, house.DBClub.Id, req.ActId))
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseLuckInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.HouseLuckInfo)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	return xerrors.SuccessCode, house.GetLuckConfig(req.ActId)

}

func Proto_ClubHouseMemLuckCheck(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.HouseMemLuck)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	ack := static.MsgMemLuckCount{Hid: req.Hid}
	if req.ActId != 0 {
		act, err := GetDBMgr().GetDBrControl().HouseActivityInfo(house.DBClub.Id, req.ActId)
		if err != nil || act == nil {
			return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}
		if act.Type != 1 {
			return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("无效的活动id").Msg
		}
		if act.BegTime > time.Now().Unix() {
			return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("活动未开始").Msg
		}
		if act.EndTime < time.Now().Unix() {
			return xerrors.NewXError("活动已结束").Code, xerrors.NewXError("活动已结束").Msg
		}
		ack.Count = house.GetUserLuckTimes(p.Uid, req.ActId, act.TicketCount)
		ack.ActId = req.ActId
		return xerrors.SuccessCode, ack
	}
	acts, err := GetDBMgr().GetDBrControl().HouseActivityList(house.DBClub.Id, true)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if len(acts) == 0 {
		return xerrors.NewXError("暂无活动").Code, xerrors.NewXError("暂无活动").Msg
	}
	for _, item := range acts {
		if item.Type != 1 {
			continue
		}
		if item.BegTime > time.Now().Unix() || item.EndTime < time.Now().Unix() {
			return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("暂无活动").Msg
		}
		ack.Count = house.GetUserLuckTimes(p.Uid, item.Id, item.TicketCount)
		ack.ActId = item.Id
		return xerrors.SuccessCode, ack
	}
	return xerrors.SuccessCode, ack
}

func Proto_ClubHouseMemGetLuck(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.HouseMemLuck)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	act, err := GetDBMgr().GetDBrControl().HouseActivityInfo(house.DBClub.Id, req.ActId)
	if err != nil || act == nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	if act.Type != 1 {
		return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("无效的活动id").Msg
	}
	if act.BegTime > time.Now().Unix() {
		return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("活动未开始").Msg
	}
	if act.EndTime < time.Now().Unix() {
		return xerrors.NewXError("活动已结束").Code, xerrors.NewXError("活动已结束").Msg
	}

	count := house.GetUserLuckTimes(p.Uid, req.ActId, act.TicketCount)
	if count <= 0 {
		return xerrors.NoLuckTimes.Code, xerrors.NoLuckTimes.Msg
	}
	cli := GetDBMgr().Redis
	ok := cli.SetNX(fmt.Sprintf(consts.REDIS_KEY_USER_LUCK, p.Uid), 1, 30*time.Second).Val()
	if !ok {
		return xerrors.NoLuckTimes.Code, xerrors.NoLuckTimes.Msg
	}
	defer cli.Del(fmt.Sprintf(consts.REDIS_KEY_USER_LUCK, p.Uid))
	config := house.GetLuckConfig(act.Id)
	used := house.GetUsedLuckRank(act.Id)
	lefts := []static.RewordInfo{}
	for _, set := range *config {
		var use bool
		for _, item := range used {
			if item.Rank == set.Rank {
				use = true
				lefts = append(lefts, static.RewordInfo{Rank: item.Rank, Count: set.Count - item.Used})
			}
		}
		if !use {
			lefts = append(lefts, static.RewordInfo{Rank: set.Rank, Count: set.Count})
		}
	}
	var totalLeft int64
	for _, left := range lefts {
		totalLeft += left.Count
	}
	var reword int64
	if totalLeft <= 0 {
		reword = 9 // 下次好运
	} else {
		rand.Seed(time.Now().UnixNano())
		res := rand.Int63() % totalLeft
		for _, left := range lefts {
			if res < left.Count {
				reword = left.Rank
				break
			} else {
				res = res - left.Count
			}
		}
	}
	if reword == 0 {
		reword = 9
	}
	ack := static.MsgMemGetLuck{}
	ack.Rank = reword
	ack.CountLeft = count - 1
	ack.ActId = req.ActId
	ack.Hid = req.Hid
	tx := GetDBMgr().GetDBmControl().Begin()
	var suc bool
	defer func() {
		if !suc {
			tx.Rollback()
		}
	}()
	sql := `insert into luck_record(hid,actid,uid,rank) values(?,?,?,?)`
	err = tx.Exec(sql, house.DBClub.Id, req.ActId, p.Uid, reword).Error
	if err != nil {
		xlog.Logger().Errorf("insert error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	sql2 := `update luck_ticket set used = used + ? where hid = ? and actid = ? and uid = ?`
	err = tx.Exec(sql2, act.TicketCount, house.DBClub.Id, req.ActId, p.Uid).Error
	if err != nil {
		xlog.Logger().Errorf("update error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Errorf("commit error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	suc = true
	return xerrors.SuccessCode, ack
}

func Proto_ClubHouseActDetail(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.HouseMemLuck)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	ack := static.MsgLuckDetail{Hid: req.Hid}
	if req.ActId != 0 {
		act, err := GetDBMgr().GetDBrControl().HouseActivityInfo(house.DBClub.Id, req.ActId)
		if err != nil || act == nil {
			return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}
		if act.Type != 1 {
			return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}
		if act.BegTime > time.Now().Unix() {
			return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("活动未开始").Msg
		}
		if act.EndTime < time.Now().Unix() {
			return xerrors.NewXError("活动已结束").Code, xerrors.NewXError("活动已结束").Msg
		}
		ack.ActId = req.ActId
		ack.UserLeft, ack.LeftTicket = house.GetUserLuckTimesWithTicket(p.Uid, ack.ActId, act.TicketCount)
		ack.ActTicket = act.TicketCount
		ack.ActStart = act.BegTime
		ack.ActEnd = act.EndTime
		return xerrors.SuccessCode, ack
	}
	acts, err := GetDBMgr().GetDBrControl().HouseActivityList(house.DBClub.Id, true)
	if err != nil {
		return xerrors.SuccessCode, ack
	}
	for _, item := range acts {
		if item.Type != 1 {
			continue
		}
		if item.BegTime > time.Now().Unix() {
			return xerrors.NewXError("活动未开始").Code, xerrors.NewXError("活动未开始").Msg
		}
		if item.EndTime < time.Now().Unix() {
			return xerrors.NewXError("活动已结束").Code, xerrors.NewXError("活动已结束").Msg
		}
		ack.UserLeft, ack.LeftTicket = house.GetUserLuckTimesWithTicket(p.Uid, item.Id, item.TicketCount)
		ack.ActId = item.Id
		ack.ActTicket = item.TicketCount
		ack.ActStart = item.BegTime
		ack.ActEnd = item.EndTime
		return xerrors.SuccessCode, ack
	}
	return xerrors.SuccessCode, ack
}

// 获取包厢楼层为vip楼层
func Proto_ClubHouseVipFloorGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorSetVipFloor)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	_, kindIds := house.GetFloors()
	areaGameMap := make(map[int]bool)
	for _, kindId := range kindIds {
		if _, ok := areaGameMap[kindId]; !ok {
			areaGame := GetAreaGameByKid(kindId)
			if areaGame != nil {
				areaGameMap[kindId] = areaGame.CanVipFloor
			}
		}
	}

	var ack static.MsgHouseFloorVip
	ack.Hid = house.DBClub.HId

	house.FloorLock.RLockWithLog()
	defer house.FloorLock.RUnlock()
	for _, floor := range house.Floors {
		var item static.MsgHouseFloorVipItem
		item.Fid = floor.Id
		item.IsVip = floor.IsVip
		item.IsCapSetVip = floor.IsCapSetVip   // 2.28新增队长可设数据
		item.IsDefJoinVip = floor.IsDefJoinVip // 2.30新增新入圈的玩家是否默认加入VIP楼层
		item.Disable = areaGameMap[floor.Rule.KindId]
		if optMem.IsPartner() {
			memMap := house.GetMemberMap(false)
			vipNum := floor.GetVipUsersNumByCap(optMem.UId, memMap)
			item.NumViper = vipNum
			if floor.IsCapSetVip && floor.IsVip {
				ack.Items = append(ack.Items, &item)
			}
		} else {
			item.NumViper = floor.NumVipUsers()
			ack.Items = append(ack.Items, &item)
		}
	}

	return xerrors.SuccessCode, &ack
}

// 设置包厢楼层为vip楼层
func Proto_ClubHouseVipFloorSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseFloorVip)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_ADMIN, MinorSetVipFloor)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	_, kindIds := house.GetFloors()
	areaGameMap := make(map[int]bool)
	for _, kindId := range kindIds {
		if _, ok := areaGameMap[kindId]; !ok {
			areaGame := GetAreaGameByKid(kindId)
			if areaGame != nil {
				areaGameMap[kindId] = areaGame.CanVipFloor
			}
		}
	}
	for i := 0; i < len(req.Items); i++ {
		if item := req.Items[i]; item != nil && item.IsVip {
			floor := house.GetFloorByFId(item.Fid)
			if floor == nil {
				return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
			}
			// 楼层本来就是vip
			if floor.IsVip {
				continue
			}
			if areaGameMap[floor.Rule.KindId] == false {
				return xerrors.ResultErrorCode, fmt.Sprintf("%d楼玩法不能被设置为VIP楼层。", house.GetFloorIndexByFid(floor.Id)+1)
			}
		}
	}
	var do bool
	now := time.Now()
	for i := 0; i < len(req.Items); i++ {
		if item := req.Items[i]; item != nil {
			floor := house.GetFloorByFId(item.Fid)
			if floor == nil {
				return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
			}
			item.NumViper = floor.NumVipUsers()
			if floor.IsVip != item.IsVip || floor.IsCapSetVip != item.IsCapSetVip || floor.IsDefJoinVip != item.IsDefJoinVip {
				do = true
				beforeVip := floor.IsVip
				floor.IsVip = item.IsVip
				beforeCapSetVip := floor.IsCapSetVip
				floor.IsCapSetVip = item.IsCapSetVip

				beforeDefJoinVip := floor.IsDefJoinVip
				floor.IsDefJoinVip = item.IsDefJoinVip

				err := GetDBMgr().HouseFloorUpdate(floor)
				if err != nil {
					floor.IsVip = beforeVip
					floor.IsCapSetVip = beforeCapSetVip
					floor.IsDefJoinVip = beforeDefJoinVip
					xlog.Logger().Error(err)
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
				err = GetDBMgr().GetDBmControl().Create(&models.HouseVipFloorLog{
					Fid:         floor.Id,
					OptId:       p.Uid,
					Uid:         0,
					UVip:        false,
					FVip:        floor.IsVip,
					FCapSetVip:  floor.IsCapSetVip,
					FDefJoinVip: floor.IsDefJoinVip,
					CreatedAt:   now,
				}).Error
				if err != nil {
					xlog.Logger().Error(err)
				}
			}
		}
	}
	if do {
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseSetVipFloorNtf, req)
	}
	return xerrors.SuccessCode, nil
}

// 获取vip楼层加入vip玩家或非vip玩家
func Proto_ClubHouseFloorVipUsersListGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseFloorVipUser)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	if req.PBegin < 0 {
		req.PBegin = 0
	}
	// 不包含end位，需要多取一位
	req.PEnd++
	if req.PEnd < req.PBegin {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	house, floor, optMem, cusErr := inspectClubFloorMemberWithRight(req.HId, req.Fid, p.Uid, consts.ROLE_ADMIN, MinorSetVipFloor)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	vipUser := floor.GetVipUsersSet()
	memMap := house.GetMemberMap(false)
	arrMembers := make([]HouseMember, 0)

	for _, mem := range memMap {
		_, ok = vipUser[mem.UId]
		if req.IsVip == ok {
			if optMem.IsPartner() {
				if mem.Partner == optMem.UId || mem.UId == optMem.UId {
					arrMembers = append(arrMembers, mem)
				}
			} else {
				arrMembers = append(arrMembers, mem)
			}
		}
	}

	var ack static.Msg_HC_HouseFloorVipUsers
	ack.Hid = house.DBClub.HId
	ack.Fid = floor.Id
	// 条件切片
	var conditionArr []HouseMember
	if req.Param == "" {
		conditionArr = arrMembers
	} else {
		for _, mem := range arrMembers {
			// ID 包含
			if strings.Contains(fmt.Sprintf("%d", mem.UId), req.Param) {
				conditionArr = append(conditionArr, mem)
				continue
			}
			if strings.Contains(mem.NickName, req.Param) {
				conditionArr = append(conditionArr, mem)
				continue
			}
		}
	}

	ack.Totalnum = len(conditionArr)
	ack.FMems = make([]static.Msg_HouseMemberLiteItem, 0)

	// 分页超出范围
	if len(conditionArr) == 0 || len(conditionArr) < req.PBegin {
		return xerrors.SuccessCode, &ack
	}
	// 开始索引
	var idxBeg int
	var idxEnd int
	idxBeg = req.PBegin
	if idxBeg < 0 {
		idxBeg = 0
	}
	if req.PEnd > len(conditionArr) {
		idxEnd = len(conditionArr)
	} else {
		idxEnd = req.PEnd
	}
	conditionArr = conditionArr[idxBeg:idxEnd]

	for _, mem := range conditionArr {
		var item static.Msg_HouseMemberLiteItem
		item.UId = mem.UId
		item.UName = mem.NickName
		item.UUrl = mem.ImgUrl
		item.UGender = mem.Sex
		// titem.UId = mem.UId
		// titem.UOnline = GetPlayerMgr().IsUserOnline(mem.UId)
		// titem.UName = mem.NickName
		// titem.URole = mem.URole
		// titem.UVitamin = public.SwitchVitaminToF64(mem.UVitamin)
		// titem.URemark = mem.URemark
		// titem.UUrl = mem.ImgUrl
		// titem.UGender = mem.Sex
		// titem.UJoinTime = mem.ApplyTime
		// titem.UPartner = mem.Partner
		ack.FMems = append(ack.FMems, item)
	}
	return xerrors.SuccessCode, &ack
}

// 包厢楼层设置vip玩家
// 获取vip楼层加入vip玩家或非vip玩家
func Proto_ClubHouseFloorVipUserSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseFloorVipUserSet)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	house, floor, optMem, cusErr := inspectClubFloorMemberWithRight(req.HId, req.Fid, p.Uid, consts.ROLE_ADMIN, MinorSetVipFloor)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	mem := house.GetMemByUId(req.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	vipUser := floor.GetVipUsersSet()
	_, ok = vipUser[mem.UId]
	if ok != req.IsVip {
		var err error
		if req.IsVip {
			err = floor.AddVipUsers(mem.UId)
		} else {
			err = floor.RemVipUsers(mem.UId)
		}
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		err = GetDBMgr().GetDBmControl().Create(&models.HouseVipFloorLog{
			Fid:        floor.Id,
			OptId:      p.Uid,
			Uid:        mem.UId,
			UVip:       req.IsVip,
			FVip:       floor.IsVip,
			FCapSetVip: floor.IsCapSetVip,
			CreatedAt:  time.Now(),
		}).Error
		if err != nil {
			xlog.Logger().Error(err)
		}
	}

	f_index := house.GetFloorIndexByFid(req.Fid) + 1
	var msg string
	if req.IsVip {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主将%s ID:%d加入%d楼VIP", mem.URemark, mem.UId, f_index)
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%s ID:%d将玩家%s ID:%d加入%d楼VIP", optMem.URemark, optMem.UId, mem.URemark, mem.UId, f_index)
		} else if optMem.Partner == 1 {
			msg = fmt.Sprintf("队长%s ID:%d将玩家%s ID:%d加入%d楼VIP", optMem.URemark, optMem.UId, mem.URemark, mem.UId, f_index)
		}
	} else {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主将%s ID:%d移除%d楼VIP", mem.URemark, mem.UId, f_index)
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%s ID:%d将玩家%s ID:%d移出%d楼VIP", optMem.URemark, optMem.UId, mem.URemark, mem.UId, f_index)
		} else if optMem.Partner == 1 {
			msg = fmt.Sprintf("队长%s ID:%d将玩家%s ID:%d移出%d楼VIP", optMem.URemark, optMem.UId, mem.URemark, mem.UId, f_index)
		}
	}
	go CreateClubMassage(house.DBClub.Id, optMem.UId, FloorVipChange, msg)

	house.Broadcast(consts.ROLE_ADMIN, consts.MsgTypeHouseFloorVipUserSetNtf, &static.Msg_CH_HouseFloorVipUserSetNtf{
		Msg_CH_HouseFloorVipUserSet: *req,
		UName:                       mem.NickName,
		UUrl:                        mem.ImgUrl,
		UGender:                     mem.Sex,
		NumViper:                    floor.NumVipUsers(),
	})

	if mem.Partner >= 1 {
		capId := mem.Partner
		if mem.Partner == 1 {
			capId = mem.UId
		}
		memMap := house.GetMemberMap(false)
		vipNum := floor.GetVipUsersNumByCap(capId, memMap)
		house.Broadcast2Parnter(capId, false, consts.MsgTypeHouseFloorVipUserSetNtf, &static.Msg_CH_HouseFloorVipUserSetNtf{
			Msg_CH_HouseFloorVipUserSet: *req,
			UName:                       mem.NickName,
			UUrl:                        mem.ImgUrl,
			UGender:                     mem.Sex,
			NumViper:                    vipNum,
		})
	}

	return xerrors.SuccessCode, nil
}

// 获取vip楼层一键加入所以  vip玩家或非vip玩家
func Proto_ClubHouseFloorVipUserAllSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseFloorVipUserAllSet)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}
	house, floor, optMem, cusErr := inspectClubFloorMemberWithRight(req.HId, req.Fid, p.Uid, consts.ROLE_ADMIN, MinorSetVipFloor)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	vipUser := floor.GetVipUsersSet()
	memMap := house.GetMemberMap(false)
	arrMembers := make([]HouseMember, 0)
	for _, mem := range memMap {
		_, ok = vipUser[mem.UId]
		if req.IsVip != ok {
			if optMem.IsPartner() {
				if mem.Partner == optMem.UId || mem.UId == optMem.UId {
					arrMembers = append(arrMembers, mem)
				}
			} else {
				arrMembers = append(arrMembers, mem)
			}
		}
	}

	var ids []int64
	sql := "insert into `house_vip_floor_log` (`fid`,`optid`,`uid`,`uvip`,`fvip`,`f_cap_set_vip`,`created_at`) values "
	for _, itemMem := range arrMembers {
		ids = append(ids, itemMem.UId)
		sql += fmt.Sprintf("(%d, %d, %d, %v, %v, %v, now()),", floor.Id, p.Uid, itemMem.UId, req.IsVip, floor.IsVip, floor.IsCapSetVip)
	}
	sql = strings.TrimRight(sql, ",")
	sql += ";"
	var err error
	if req.IsVip {
		err = floor.AddVipUsers(ids...)
	} else {
		err = floor.RemVipUsers(ids...)
	}
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	//syslog.Logger().Error("idsppppp:  ", ids)
	go func() {
		e := GetDBMgr().GetDBmControl().Exec(sql).Error
		if e != nil {
			xlog.Logger().Error(e)
		}
	}()

	f_index := house.GetFloorIndexByFid(req.Fid) + 1
	var msg string
	if req.IsVip {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主在%d楼执行了一键导入VIP", f_index)
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%s ID:%d在%d楼执行了一键导入VIP", optMem.URemark, optMem.UId, f_index)
		} else if optMem.Partner == 1 {
			msg = fmt.Sprintf("队长%s ID:%d在%d楼执行了一键导入VIP", optMem.URemark, optMem.UId, f_index)
		}
	} else {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主在%d楼执行了一键导出VIP", f_index)
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%s ID:%d在%d楼执行了一键导出VIP", optMem.URemark, optMem.UId, f_index)
		} else if optMem.Partner == 1 {
			msg = fmt.Sprintf("队长%s ID:%d在%d楼执行了一键导出VIP", optMem.URemark, optMem.UId, f_index)
		}
	}
	go CreateClubMassage(house.DBClub.Id, optMem.UId, FloorVipChange, msg)

	house.Broadcast(consts.ROLE_ADMIN, consts.MsgTypeHouseFloorSetAllVipUser_Ntf, nil)
	house.ParnterBroadcast(consts.MsgTypeHouseFloorSetAllVipUser_Ntf, nil)

	return xerrors.SuccessCode, nil
}

// ! 获取包厢入桌距离限制
func Proto_ClubHouseTableDistanceLimitGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_TableDistanceLimitGet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	limitDistance := house.GetHouseTableLimitDistance()
	//if limitDistance == -1 {
	//	return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	//}
	var ack static.Msg_HC_TableDistanceLimitGet

	ack.Distance = limitDistance

	return xerrors.SuccessCode, &ack
}

// ! 设置包厢入桌距离限制
func Proto_ClubHouseTableDistanceLimitSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_TableDistanceLimitSet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorSetDistance)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if req.Distance < 0 {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}
	var modeLog models.HouseTableDistanceLimitLog
	var model models.HouseTableDistanceLimit

	err := GetDBMgr().GetDBmControl().Where("dhid = ?", house.DBClub.Id).First(&model).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		model.DHId = house.DBClub.Id
		model.TableDistanceLimit = req.Distance
		model.CreatedAt = time.Now()
		model.UpdatedAt = time.Now()
		err = GetDBMgr().GetDBmControl().Create(&model).Error
	} else {
		model.TableDistanceLimit = req.Distance

		updateAttr := make(map[string]interface{})
		updateAttr["table_distance_limit"] = req.Distance
		updateAttr["updated_at"] = time.Now()
		err = GetDBMgr().GetDBmControl().Model(&model).Update(updateAttr).Error
	}

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 插入修改日志
	modeLog.DHId = house.DBClub.Id
	modeLog.TableDistanceLimit = req.Distance
	modeLog.CreatedAt = time.Now()
	GetDBMgr().GetDBmControl().Save(&modeLog)

	// 发布通知
	ntf := new(static.Msg_CH_TableDistanceLimitSet)
	ntf.HId = req.HId
	ntf.Distance = req.Distance
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseTableDistanceLimitSetNtf, ntf)

	return xerrors.SuccessCode, nil
}

// ! 获取奖励均衡均摊方式
func Proto_ClubHouseRewardBalancedTypeGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	return xerrors.SuccessCode, &static.Msg_CH_RewardBalancedTypeSet{
		HId:                house.DBClub.HId,
		RewardBalancedType: house.DBClub.RewardBalancedType,
	}
}

// ! 设置奖励均衡均摊方式租房
func Proto_ClubHouseRewardBalancedTypeSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_RewardBalancedTypeSet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house.DBClub.RewardBalancedType != req.RewardBalancedType {
		cusErr = house.ModifyRewardBalancedType(req.RewardBalancedType)
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
		house.flush()
		opt := ""
		if req.RewardBalancedType == 0 {
			opt = "低分局"
		} else {
			opt = "所有局"
		}
		msg := fmt.Sprintf("盟主修改均摊方式为:%s均摊。", opt)
		CreateClubMassage(house.DBClub.Id, house.DBClub.UId, MemVitaminSend, msg)
	}
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseApplySwitch(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseApplySwitchSet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorJoinReviewed)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house.DBClub.ApplySwitch == req.Switch {
		return xerrors.SuccessCode, nil
	}
	house.DBClub.ApplySwitch = req.Switch
	house.flush()
	if req.Switch {
		uids := house.GetMemUIdsByRole(consts.ROLE_APLLY)
		for _, uid := range uids {
			house.MemRefused(p.Uid, uid)
		}
	}
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseFloorHideImg(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseFloorHideImg)
	house, floor, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, req.Fid, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if floor.IsHide == req.IsHide {
		return xerrors.SuccessCode, nil
	}
	floor.IsHide = req.IsHide
	err := GetDBMgr().HouseFloorUpdate(floor)
	if err != nil {
		return 0, xerrors.DBExecError
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseFloorHideImgNTF, req)
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseFloorFakeTable(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseFloorFakeTable)
	_, floor, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, req.Fid, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if floor.MinTable == req.MinTable && floor.MaxTable == floor.MaxTable {
		return xerrors.SuccessCode, nil
	}
	if req.MinTable > req.MaxTable {
		req.MinTable, req.MaxTable = req.MaxTable, req.MinTable
	}
	floor.MinTable = req.MinTable
	floor.MaxTable = req.MaxTable
	err := GetDBMgr().HouseFloorUpdate(floor)
	if err != nil {
		return 0, xerrors.DBExecError
	}
	// house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseFloorHideImgNTF, req)
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseRecordGameLike(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseRecordGameLike)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 校验时段、区间是否匹配
	timeRangeCnt := 1
	if req.TimeInterval > 0 {
		timeRangeCnt = 24 / req.TimeInterval
	}
	if req.TimeRange > timeRangeCnt || req.TimeRange <= 0 {
		return xerrors.InvalidTimeRangeError.Code, xerrors.InvalidTimeRangeError.Msg
	}

	// 构造点赞的时间区间字符串 如01-03
	timeRangeStr := ""
	if req.TimeInterval > 0 {
		startHour := (req.TimeRange - 1) * req.TimeInterval
		endHour := req.TimeRange * req.TimeInterval
		timeRangeStr = fmt.Sprintf("%02d-%02d", startHour, endHour)
	} else {
		timeRangeStr = "00-24"
	}
	//recordType := req.RecordType
	recordType := 0 //点赞先不分类型 全用 圈子战绩的类型
	var err error
	if optMem.URole == consts.ROLE_CREATER || optMem.URole == consts.ROLE_ADMIN || optMem.IsVitaminAdmin() {
		err = GetDBMgr().UpdateRecordGameLike(house.DBClub.Id, req.DateType, req.GameNum, models.OptUserTypeAdmin, recordType, req.IsLike, timeRangeStr)
	} else if optMem.IsPartner() || optMem.IsVicePartner() {
		err = GetDBMgr().UpdateRecordGameLike(house.DBClub.Id, req.DateType, req.GameNum, models.OptUserTypePartner, recordType, req.IsLike, timeRangeStr)
	} else {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseRecordUserLike(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseRecordUserLike)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 校验时段、区间是否匹配
	timeRangeCnt := 1
	if req.TimeInterval > 0 {
		timeRangeCnt = 24 / req.TimeInterval
	}
	if req.TimeRange > timeRangeCnt || req.TimeRange <= 0 {
		return xerrors.InvalidTimeRangeError.Code, xerrors.InvalidTimeRangeError.Msg
	}

	// 构造点赞的时间区间字符串 如01-03
	timeRangeStr := ""
	if req.TimeInterval > 0 {
		startHour := (req.TimeRange - 1) * req.TimeInterval
		endHour := req.TimeRange * req.TimeInterval
		timeRangeStr = fmt.Sprintf("%02d-%02d", startHour, endHour)
	} else {
		timeRangeStr = "00-24"
	}

	var err error
	if optMem.URole == consts.ROLE_CREATER || optMem.URole == consts.ROLE_ADMIN || optMem.IsVitaminAdmin() {
		err = GetDBMgr().UpdateRecordUserLike(house.DBClub.Id, req.DateType, req.LikeUser, models.OptUserTypeAdmin, req.IsLike, timeRangeStr, req.IsTeamLike)
	} else if optMem.IsPartner() || optMem.IsVicePartner() {
		err = GetDBMgr().UpdateRecordUserLike(house.DBClub.Id, req.DateType, req.LikeUser, models.OptUserTypePartner, req.IsLike, timeRangeStr, req.IsTeamLike)
	} else {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, nil
}

// 查询包厢的申请信息
func Proto_ClubHouseApplyInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseApplyInfo)

	reqBegin := req.PBegin
	reqEnd := req.PEnd
	if req.PBegin < 0 {
		req.PBegin = 0
	}
	// 不包含end位，需要多取一位
	req.PEnd++
	if req.PEnd < req.PBegin {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 1")
	house, _, mem, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 2")
	if mem.Lower(consts.ROLE_ADMIN) && !mem.IsPartner() && !mem.IsVicePartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 3")
	houseMemberMap, _, _, _, _, _, _, _, ApplyArr, _ := house.GetAllMemberWithClassify()
	//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 4")
	// 申请列表
	var conditionArr []*static.Msg_HouseMemberItem
	if req.Join {
		for i, l := 0, len(ApplyArr); i < l; i++ {
			hmem := ApplyArr[i]
			if hmem.UId == mem.UId {
				continue
			}
			if mem.IsPartner() && hmem.Partner != mem.UId {
				continue
			} else if mem.IsVicePartner() && hmem.Partner != mem.Partner {
				continue
			}
			if req.Param != "" && !strings.Contains(fmt.Sprint(hmem.UId), req.Param) && !strings.Contains(hmem.NickName, req.Param) {
				continue
			}
			var titem static.Msg_HouseMemberItem
			titem.UId = hmem.UId
			titem.ApplyAt = hmem.ApplyTime
			titem.ApplyType = consts.HouseMemberApplyJoin
			conditionArr = append(conditionArr, &titem)
		}
	}

	if req.Exit {
		// 退出申请列表
		//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 5")
		exits := house.GetExitApplicants()
		//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 6")
		for i, l := 0, len(exits); i < l; i++ {
			uid := exits[i].Uid
			if uid == mem.UId {
				continue
			}
			hmem, ok := houseMemberMap[uid]
			if !ok {
				continue
			}
			if mem.IsPartner() && hmem.Partner != mem.UId {
				continue
			} else if mem.IsVicePartner() && hmem.Partner != mem.Partner {
				continue
			}
			if req.Param != "" && !strings.Contains(fmt.Sprint(hmem.UId), req.Param) && !strings.Contains(hmem.NickName, req.Param) {
				continue
			}
			var titem static.Msg_HouseMemberItem
			titem.UId = hmem.UId
			titem.ApplyAt = exits[i].UpdatedAt.Unix()
			titem.ApplyType = consts.HouseMemberApplyExit
			conditionArr = append(conditionArr, &titem)
		}
	}

	var ack static.Msg_HC_HouseMemList
	ack.PBegin = reqBegin
	ack.PEnd = reqEnd
	ack.Totalnum = len(conditionArr)
	//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 6")
	if house.IsHideOnlineNum(p.Uid) {
		ack.HMemNum = -1
		ack.HMemOnLineNum = -1
	} else {
		// 为了更加匹配成员列表的数据  这里取一次redis数据
		ack.HMemNum = house.GetMemCounts()
		ack.HMemOnLineNum = house.GetMemOnlineCounts()
	}
	//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 7")
	ack.FMems = make([]*static.Msg_HouseMemberItem, 0)

	// 分页超出范围
	if len(conditionArr) == 0 || len(conditionArr) < req.PBegin {
		return xerrors.SuccessCode, &ack
	}

	sort.Sort(&static.MsgHouseMemberItemWrapper{
		Hms: conditionArr,
		By: func(i, j *static.Msg_HouseMemberItem) bool {
			return i.ApplyAt > j.ApplyAt
		},
	})

	// 开始索引
	var idxBeg int
	var idxEnd int
	idxBeg = req.PBegin
	if req.PEnd > len(conditionArr) {
		idxEnd = len(conditionArr)
	} else {
		idxEnd = req.PEnd
	}
	conditionArr = conditionArr[idxBeg:idxEnd]
	cli := GetDBMgr().Redis
	for _, titem := range conditionArr {
		hmem, ok := houseMemberMap[titem.UId]
		if !ok {
			continue
		}
		dmem, err := GetDBMgr().GetDBrControl().GetPerson(hmem.UId)
		if err != nil {
			xlog.Logger().Errorf("user not exists in house:%d,uid:%d", house.DBClub.Id, hmem.UId)
			//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 8")
			house.MemDelete(hmem.UId, true, nil)
			//xlog.Logger().Warn("Proto_ClubHouseApplyInfo 9")
			continue
		}
		titem.UOnline = hmem.IsOnline
		titem.UPlaying = dmem.TableId > 0
		titem.UName = hmem.NickName
		titem.URole = hmem.URole
		titem.UVitamin = static.SwitchVitaminToF64(hmem.UVitamin)
		titem.UPartner = hmem.Partner
		titem.VitaminAdmin = hmem.VitaminAdmin
		titem.VicePartner = hmem.VicePartner
		if refHid, ok := hmem.CheckRef(); ok {
			titem.URefHId = refHid
		}
		if titem.UPartner > 1 {
			dpmem, ok := houseMemberMap[titem.UPartner]
			if ok {
				titem.UPartnerName = dpmem.NickName
				titem.UPartnerUrl = dpmem.ImgUrl
			} else {
				hmem.Partner = 0 // 错误数据修复
			}
		}
		titem.URemark = hmem.URemark
		titem.UUrl = hmem.ImgUrl
		titem.UGender = hmem.Sex
		titem.UJoinTime = hmem.ApplyTime
		titem.GameLimit = cli.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, house.DBClub.Id), hmem.UId).Val()
		titem.LastLoginAt = dmem.LastOffLineTime
		ack.FMems = append(ack.FMems, titem)
	}

	return xerrors.SuccessCode, &ack
}

// 包厢成员小红点
func Proto_ClubHouseMemberTrackPoint(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)

	house, _, mem, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	var ack static.Msg_HC_HouseMemberTrackPoint
	ack.HId = house.DBClub.HId
	if mem.Lower(consts.ROLE_ADMIN) && !mem.IsPartner() && !mem.IsVicePartner() {
		xlog.Logger().Error("包厢普通成员 请求 小红点。", mem.UId, mem.URole, mem.Partner)
	} else {
		houseMemberMap, _, _, _, _, _, _, _, ApplyArr, _ := house.GetAllMemberWithClassify()
		// 申请列表

		for i, l := 0, len(ApplyArr); i < l; i++ {
			hmem := ApplyArr[i]
			if hmem.UId == mem.UId {
				continue
			}
			if mem.IsPartner() && hmem.Partner != mem.UId {
				continue
			}
			ack.ApplyCount++
		}

		// 退出申请列表
		exits := house.GetExitApplicants()
		for i, l := 0, len(exits); i < l; i++ {
			uid := exits[i].Uid
			if uid == mem.UId {
				continue
			}
			hmem, ok := houseMemberMap[uid]
			if !ok {
				continue
			}
			if mem.IsPartner() && hmem.Partner != mem.UId {
				continue
			}
			ack.ApplyCount++
		}
	}
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseFloorColorSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseFloorColorSet)
	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorSetTableColor)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	str := strings.Join(req.FloorsColor, ",")
	floorColor := &models.HouseFloorColor{
		Id:         house.DBClub.Id,
		FloorColor: str,
	}
	err := GetDBMgr().GetDBmControl().Save(floorColor).Error
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorColorSetNtf, req)

	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseRankSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseRankSet)
	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	rankRound := req.RankRound
	rankWiner := req.RankWiner
	rankRecord := req.RankRecord
	rankOpen := req.RankOpen
	err := house.UpdateRankInfo(rankRound, rankWiner, rankRecord, rankOpen)
	if err != nil {
		return err.Code, err.Msg
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseRankSet_ntf, req)
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseRankGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseRankGet)
	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	var ack static.Msg_S2C_HouseMemberRankGet
	ack.RankRound = house.DBClub.RankRound
	ack.RankWiner = house.DBClub.RankWiner
	ack.RankRecord = house.DBClub.RankRecord
	ack.RankOpen = house.DBClub.RankOpen
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseRankInfoGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseRankInfoGet)
	house, _, _, err := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if err != xerrors.RespOk {
		return err.Code, err.Msg
	}
	// 筛选时间的起点和终点
	nowTime := time.Now()
	selectTime1 := nowTime
	selectTime2 := nowTime
	var zeroTime time.Time
	switch req.TimeType {
	case static.RANK_TIME_TODAY:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, 0))
		selectTime1 = zeroTime
		selectTime2 = zeroTime.Add(24 * time.Hour)
	case static.RANK_TIME_YESTERDAY:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, -1))
		selectTime1 = zeroTime
		selectTime2 = zeroTime.Add(24 * time.Hour)
	case static.RANK_TIME_WEEK:
		offset := int(time.Monday - nowTime.Weekday())
		if offset > 0 {
			offset = -6
		}
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, offset-7)) //offset 计算的是本周第一天的偏移 -7 就是上周的第一天
		selectTime1 = zeroTime
		selectTime2 = zeroTime.AddDate(0, 0, 7)
	case static.RANK_TIME_MONTH:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, -1, -nowTime.Day()+1))
		selectTime1 = zeroTime
		selectTime2 = zeroTime.AddDate(0, 1, 0)
	default:
		zeroTime = static.GetZeroTime(nowTime.AddDate(0, 0, 0))
		selectTime1 = zeroTime
		selectTime2 = zeroTime.Add(24 * time.Hour)
	}

	//syslog.Logger().Println("开始时间：", selectTime1)
	//syslog.Logger().Println("结束时间：", selectTime2)
	houseMem := house.GetMemberMap(false)
	rankUserMap := make(map[int64]*static.HouseMemberRankItem)
	for _, mem := range houseMem {
		// 数据补齐
		var rankItem static.HouseMemberRankItem
		rankItem.UId = mem.UId
		rankItem.UName = mem.NickName
		rankItem.UUrl = mem.ImgUrl
		rankItem.UGender = mem.Sex
		rankItem.URankNum = 0
		rankUserMap[rankItem.UId] = &rankItem
	}

	// 声明统计结果变量
	var selStsErr error
	var statisticsItems map[int64]models.QueryHouseRankResult
	statisticsItems = make(map[int64]models.QueryHouseRankResult)
	if req.TimeType < static.RANK_TIME_CARVE {
		statisticsItems, selStsErr = GetDBMgr().SelectHouseMemberRank(house.DBClub.Id, static.RANK_TIME_CARVE-100, selectTime1, selectTime2, req.RankType, req.PBegin, req.PEnd) // 减100 只是为了函数判断 大小与
	} else {
		statisticsItems, selStsErr = GetDBMgr().SelectHouseMemberRank(house.DBClub.Id, static.RANK_TIME_CARVE+100, selectTime1, selectTime2, req.RankType, req.PBegin, req.PEnd)
	}

	if selStsErr != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	for _, sitem := range statisticsItems {
		if staticItem, ok := rankUserMap[sitem.Uid]; ok {
			if req.RankType == static.RANK_TYPE_ROUND {
				staticItem.URankNum = float64(sitem.RankRound)
			} else if req.RankType == static.RANK_TYPE_WINER {
				staticItem.URankNum = float64(sitem.RankWiner)
			} else if req.RankType == static.RANK_TYPE_RECORD {
				staticItem.URankNum = static.HF_DecimalDivide(sitem.RankRecord, 1, 2)
			}
		}
	}
	var rankUserArr []*static.HouseMemberRankItem
	for _, sItem := range rankUserMap {
		//if sItem.URankNum > 0 {
		//	rankUserArr = append(rankUserArr, sItem)
		//}
		rankUserArr = append(rankUserArr, sItem)
	}
	if selStsErr == nil {
		sort.Sort(static.HouseMemberRankDataWrapper{Item: rankUserArr, By: func(item1, item2 *static.HouseMemberRankItem) bool {
			if item1.URankNum > item2.URankNum {
				return true
			} else if item1.URankNum == item2.URankNum {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	}
	var ack static.Msg_S2C_HouseMemberRankData
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd

	// 查询前50条
	var idxBeg int = req.PBegin
	var idxEnd int = req.PEnd
	if idxEnd > len(rankUserArr) {
		idxEnd = len(rankUserArr)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}
	ack.UserItem = append(ack.UserItem, rankUserArr[idxBeg:idxEnd]...)
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseOffWork(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgOffWork)
	// req.IsOffWork   true 打烊了  false 营业中
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.HID, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr == xerrors.InvalidPermission /*&& !req.AllowGame*/ {
		if !optMem.IsPartner() {
			if !optMem.IsVicePartner() {
				return cusErr.Code, cusErr.Msg
			}
		}
	} else {
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
	}
	tempPartnerUid := optMem.UId
	if optMem.IsVicePartner() {
		tempPartnerUid = optMem.Partner
	}
	if err := GetDBMgr().GetDBrControl().RedisV2.HSet(fmt.Sprintf(consts.REDIS_KEY_OFFWORK, optMem.DHId), fmt.Sprintf("%d", tempPartnerUid), req.IsOffWork).Err(); err != nil {
		xlog.Logger().Error("hmset OffWork eve to redis error:", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var msg string
	if req.IsOffWork {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主禁用小队")
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%sID:%d禁用小队", optMem.URemark, optMem.UId)
		} else if optMem.Partner == 1 {
			msg = fmt.Sprintf("队长%sID:%d禁用小队", optMem.URemark, optMem.UId)
		} else if optMem.IsVicePartner() {
			msg = fmt.Sprintf("副队长%sID:%d禁用小队", optMem.URemark, optMem.UId)
		}
	} else {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主恢复小队")
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%sID:%d恢复小队", optMem.URemark, optMem.UId)
		} else if optMem.Partner == 1 {
			msg = fmt.Sprintf("队长%sID:%d恢复小队", optMem.URemark, optMem.UId)
		} else if optMem.IsVicePartner() {
			msg = fmt.Sprintf("副队长%sID:%d恢复小队", optMem.URemark, optMem.UId)
		}
	}
	if msg != "" {
		go CreateClubMassage(house.DBClub.Id, optMem.UId, TeamOffWork, msg)
	}

	if optMem.URole == 0 {
		tempPartnerUid = 0
	}
	house.BroadcastTeam(tempPartnerUid, consts.MsgTypeOffWork_ntf, req)
	return xerrors.RespOk.Code, nil
}

// 设置房卡低于xx时提示盟主
func Proto_SetFangKaTipsMinNum(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_C2S_SetFangKaTipsMinNum)
	house, _, _, custErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorSetCardRemind)
	if custErr != xerrors.RespOk {
		return custErr.Code, custErr.Msg
	}

	if house != nil {
		custErr = house.SetFangKaTipsMinNum(req.MinNum)
		if custErr != nil {
			return custErr.Code, custErr.Msg
		}
	}

	return xerrors.SuccessCode, nil
}

// 搜索玩家
func Proto_HallSearchUser(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	//(*public.Person, error)
	req, ok := data.(*static.MsgC2SHidUid)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	house, _, opHm, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	hm := house.GetMemByUId(req.Uid)
	if hm != nil {
		return xerrors.ResultErrorCode, "当前玩家已在本包厢"
	}
	if opHm.IsPartner() || opHm.URole == consts.ROLE_CREATER { //&& !opHm.IsVicePartner()
		per, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
		if err != nil {
			return xerrors.ResultErrorCode, "请输入完整的玩家ID"
		}
		return xerrors.SuccessCode, &static.Msg_S2C_SearchUser{per.Uid, per.Nickname, per.Imgurl}
	} else {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
}

// 手动添加玩家
func Proto_ClubHouseMtAddUser(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	//(*public.Person, error)
	req, ok := data.(*static.MsgC2SHidUid)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, opHm, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorMtAdd)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	hm := house.GetMemByUId(req.Uid)
	if hm != nil {
		return xerrors.ResultErrorCode, "当前玩家已在本包厢"
	}
	//
	if opHm.IsPartner() || opHm.URole == consts.ROLE_CREATER { //&& !opHm.IsVicePartner()
		// 验证 req.Uid这个玩家自己加入圈的上限 和 允许拒绝的选项
		per, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		if per.RefuseInvite {
			return xerrors.ResultErrorCode, "玩家拒绝邀请加入包厢，添加失败。"
		}
		// 加入
		custerr := house.MemJoin(per.Uid, consts.ROLE_MEMBER, 0, true, nil)
		if custerr != nil {
			switch custerr {
			case xerrors.HouseMemJoinMaxError:
				return xerrors.ResultErrorCode, "添加失败，包厢人数已满。"
			case xerrors.MemJoinHouseMaxError:
				return xerrors.ResultErrorCode, "添加失败，玩家加入包厢数量达到上限。"
			default:
				return custerr.Code, custerr.Msg
			}
		}
		msg := ""
		if opHm.IsPartner() {
			newHm := house.GetMemByUId(per.Uid)
			if newHm != nil {
				newHm.Partner = opHm.UId
				if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
					Where("hid = ? and uid = ?", house.DBClub.Id, newHm.UId).Update("partner", opHm.UId).Error; err != nil {
					xlog.Logger().Error(err)
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
				newHm.Flush()
			}
			msg = fmt.Sprintf("队长<color=#F93030>%sID:%d</color>添加玩家<color=#00A70C>%sID:%d</color>加入包厢", opHm.NickName, opHm.UId, per.Nickname, per.Uid)
		} else {
			msg = fmt.Sprintf("盟主添加玩家<color=#00A70C>%sID:%d</color>加入包厢", per.Nickname, per.Uid)
		}
		CreateClubMassage(house.DBClub.Id, p.Uid, JoinHouse, msg)
		return xerrors.SuccessCode, nil
	} else {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
}

// 玩家被包厢邀请开关
func Proto_UserRefuseInvite(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgC2SRefuseInvite)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if req.RefuseInvite == p.RefuseInvite {
		return xerrors.SuccessCode, nil
	}
	tx := GetDBMgr().GetDBmControl().Begin()
	var err error
	defer static.TxCommit(tx, err)
	err = tx.Model(&models.User{}).Where("id = ?", p.Uid).Update("refuse_invite", req.RefuseInvite).Error
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "RefuseInvite", req.RefuseInvite)
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if per := GetPlayerMgr().GetPlayer(p.Uid); per != nil {
		per.Info.RefuseInvite = req.RefuseInvite
	}
	p.RefuseInvite = req.RefuseInvite
	return xerrors.SuccessCode, nil
}

// 权限
// 查询当前玩家权限
func Proto_HmLookUserRight(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgC2SHidUid)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, opHm, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	hm := house.GetMemByUId(req.Uid)
	if hm == nil {
		return xerrors.ResultErrorCode, "当前包厢无此玩家"
	}
	_, err := GetMRightMgr().checkRoleSetRight(opHm, hm)
	if err != nil {
		return xerrors.ResultErrorCode, "您没有权限查看成员权限内容"
	}
	roleMap, err := GetMRightMgr().FindRightByMember(hm, true)
	if err != nil {
		return xerrors.ResultErrorCode, "查询玩家权限失败"
	}
	return xerrors.SuccessCode, &static.Msg_S2C_HmUserRight{Right: roleMap}
}

// 修改当前玩家权限
func Proto_HmUpdateUserRight(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgC2SUpdateHmUright)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	if req.UpdateRight == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	house, _, opHm, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	hm := house.GetMemByUId(req.Uid)
	if hm == nil {
		return xerrors.ResultErrorCode, "当前包厢无此玩家"
	}
	_, err := GetMRightMgr().checkRoleSetRight(opHm, hm)
	if err != nil {
		return xerrors.ResultErrorCode, "您没有权限修改成员权限内容"
	}
	// 读取当前被查看人的权限 并且转换成 dataMap 发送
	upStr, err := GetMRightMgr().UpdateRightByMember(hm, req.UpdateRight, house)
	if err != nil {
		return xerrors.ResultErrorCode, "修改玩家权限失败"
	}
	if player := GetPlayerMgr().GetPlayer(hm.UId); player != nil {
		player.SendMsg(consts.MsgTypeHmUpdateUserRight_NTF, static.Msg_S2C_UpdateHmUserRight{
			Hid:         req.Hid,
			Uid:         req.Uid,
			UpdateRight: upStr,
		})
	}
	return xerrors.SuccessCode, &static.Msg_S2C_UpdateHmUserRight{req.Hid, req.Uid, upStr}
}

// 修改包厢功能开关
func Proto_HmSetSwitch(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgC2SHmSwitch)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	//keyName := ""
	var keyName string
	for k, _ := range req.Switch {
		keyName = k
	}
	rightKey := GetRightKey(keyName)
	house, _, _, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, rightKey)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	_, err := house.SetMemberSwitch(req.Switch)
	if err != nil {
		return err.Code, err.Msg
	}

	// 发送通知
	ntf := new(static.Msg_S2C_HmSwitch)
	ntf.HId = house.DBClub.HId
	ntf.Switch = req.Switch
	if rightKey == MinorBanWxTea || rightKey == IsRecShowParent {
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHmSetSwitch_ntf, ntf)
	} else if rightKey == CapSetDep {
		house.ParnterBroadcast(consts.MsgTypeHmSetSwitch_ntf, ntf)
	}
	return xerrors.SuccessCode, ntf
}

// 设置战绩筛选时段
func Proto_SetRecordTimeInterval(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_C2S_SetRecordTimeInterval)
	house, _, _, custErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if custErr != xerrors.RespOk {
		return custErr.Code, custErr.Msg
	}

	if house != nil {
		custErr = house.SetRecordTimeInterval(req.TimeInterval)
		if custErr != nil {
			return custErr.Code, custErr.Msg
		}
	}

	return xerrors.SuccessCode, nil
}

// 队长备注
func Proto_PartnerRemark(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_PartnerRemark)
	if utf8.RuneCountInString(req.Name) > 10 {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, opHm, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	hm := house.GetMemByUId(req.Uid)
	if hm == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if opHm.Partner != 1 && !opHm.IsVicePartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	if hm.UId == opHm.UId || hm.Partner == opHm.UId || hm.Partner == opHm.Partner || (opHm.IsVicePartner() && hm.IsPartner() && opHm.Partner == hm.UId) {
		hm.PRemark = req.Name
		hm.Flush()
		return xerrors.SuccessCode, nil
	}
	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
}

// 获取隐藏地理位置开关
func Proto_ClubHousePrivateGPSGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)

	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	return xerrors.SuccessCode, static.Msg_S2C_HousePrivateGPSGet{PrivateGPS: house.DBClub.PrivateGPS}
}

// 设置隐藏地理位置开关
func Proto_ClubHousePrivateGPSSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HousePrivateGPSSet)
	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorSetPrivacy)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	//house.DBClub.PrivateGPS = req.PrivateGPS
	//
	//house.flush()

	custerr := house.OptionMemGPSHide(req.PrivateGPS)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	return xerrors.SuccessCode, nil
}

// 获取包厢房卡消耗统计包括日活
func Proto_ClubHouseCardStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseCardStatistics)
	house, _, viewMem, cer := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	if viewMem.Lower(consts.ROLE_ADMIN) && !viewMem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	isPartnerView := viewMem.IsPartner()
	housemembers := house.GetMemberMap(false)
	partnerMap := make(map[int64]*static.Msg_HC_HouseCardStatisticsItem)
	memToPartner := make(map[int64]int64)
	for _, mem := range housemembers {
		if isPartnerView {
			if mem.UId == viewMem.UId {
				partnerMap[mem.UId] = &static.Msg_HC_HouseCardStatisticsItem{
					UId:     mem.UId,
					UName:   mem.NickName,
					UUrl:    mem.ImgUrl,
					UGender: mem.Sex,
				}
				memToPartner[mem.UId] = mem.UId
			} else {
				if mem.Partner > 0 {
					memToPartner[mem.UId] = mem.Partner
				} else {
					memToPartner[mem.UId] = house.DBClub.UId
				}
			}
		} else {
			if mem.IsPartner() || mem.UId == house.DBClub.UId {
				partnerMap[mem.UId] = &static.Msg_HC_HouseCardStatisticsItem{
					UId:     mem.UId,
					UName:   mem.NickName,
					UUrl:    mem.ImgUrl,
					UGender: mem.Sex,
				}
				memToPartner[mem.UId] = mem.UId
			} else {
				if mem.Partner > 0 {
					memToPartner[mem.UId] = mem.Partner
				} else {
					memToPartner[mem.UId] = house.DBClub.UId
				}
			}
		}
	}

	partnerIds := make([]int64, 0)

	// 搜索功能
	for uid, sItem := range partnerMap {
		if req.SearchKey != "" {
			if !strings.Contains(sItem.UName, req.SearchKey) && !strings.Contains(fmt.Sprintf("%d", sItem.UId), req.SearchKey) {
				delete(partnerMap, uid)
				continue
			}
		}
		partnerIds = append(partnerIds, sItem.UId)
	}

	zeroTime := static.GetZeroTime(time.Now().AddDate(0, 0, req.SelectTime))
	selectTime1 := zeroTime
	selectTime2 := zeroTime.Add(24 * time.Hour)
	clubId := house.DBClub.Id
	uidMap, gameNumMap, err := GetDBMgr().SelectFloorPlayStatistics(clubId, req.DFid, selectTime1, selectTime2)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	var ack static.Msg_HC_HouseCardStatistics
	ack.TotalTable = len(gameNumMap)
	gameNumList := make([]string, ack.TotalTable)
	var i int
	for gameNum := range gameNumMap {
		gameNumList[i] = gameNum
		i++
	}

	if len(partnerIds) > 0 {
		// 查询每个人的房卡消耗
		costMap, err := GetDBMgr().GetHouseRecordCardCostAndPlayTimes(
			house.DBClub.Id,
			gameNumList,
			partnerIds...,
		)
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		// 查询每个人的剩余房卡
		users := make([]*models.User, 0)
		err = GetDBMgr().GetDBmControl().Find(&users, "id in(?)", partnerIds).Error
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		usersMap := make(map[int64]*models.User)
		for _, user := range users {
			usersMap[user.Id] = user
		}

		for idx, item := range partnerMap {
			user, ok := usersMap[item.UId]
			if ok {
				item.Card = user.Card
			}
			cost, ok := costMap[item.UId]
			if ok {
				item.CardCost = cost.KaCost
				ack.TotalCard -= cost.KaCost
			}
			partnerMap[idx] = item
		}
	}

	if ack.TotalTable > 0 {
		ack.TotalPlayer = len(uidMap)
		for uid := range uidMap {
			partner, ok := memToPartner[uid]
			if ok {
				item, ok2 := partnerMap[partner]
				if ok2 {
					item.Player++
					partnerMap[partner] = item
				}
			}
		}

		i = 0
		for _, playerMap := range gameNumMap {
			refPartner := make(map[int64]struct{})
			for uid := range playerMap {
				partner, ok := memToPartner[uid]
				if ok {
					refPartner[partner] = struct{}{}
				}
			}
			for pid := range refPartner {
				item, ok := partnerMap[pid]
				if ok {
					item.Round++
					partnerMap[pid] = item
				}
			}
			i++
		}
	}

	ack.Items = make([]static.Msg_HC_HouseCardStatisticsItem, len(partnerMap))
	i = 0
	for _, item := range partnerMap {
		ack.Items[i] = *item
		i++
	}

	// 排序
	switch req.SortType {
	case static.SORT_ROUND_DES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.Round == b.Round {
					return a.UId > b.UId
				}
				return a.Round > b.Round
			},
		})
	case static.SORT_ROUND_AES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.Round == b.Round {
					return a.UId < b.UId
				}
				return a.Round < b.Round
			},
		})
	case static.SORT_PLAYS_DES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.Player == b.Player {
					return a.UId > b.UId
				}
				return a.Player > b.Player
			},
		})
	case static.SORT_PLAYS_AES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.Player == b.Player {
					return a.UId < b.UId
				}
				return a.Player < b.Player
			},
		})
	case static.SORT_CARD_DES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.Card == b.Card {
					return a.UId > b.UId
				}
				return a.Card > b.Card
			},
		})
	case static.SORT_CARD_AES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.Card == b.Card {
					return a.UId < b.UId
				}
				return a.Card < b.Card
			},
		})
	case static.SORT_COST_DES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.CardCost == b.CardCost {
					return a.UId > b.UId
				}
				return a.CardCost > b.CardCost
			},
		})
	case static.SORT_COST_AES:
		sort.Sort(&static.Msg_HC_HouseCardStatisticsItem_Wrapper{
			Items: ack.Items,
			LessFn: func(a, b *static.Msg_HC_HouseCardStatisticsItem) bool {
				if a.CardCost == b.CardCost {
					return a.UId < b.UId
				}
				return a.CardCost < b.CardCost
			},
		})
	}

	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd

	var idxBeg int = req.PBegin
	var idxEnd int = req.PEnd
	if idxEnd > len(ack.Items) {
		idxEnd = len(ack.Items)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}
	ack.Items = ack.Items[idxBeg:idxEnd]
	return xerrors.SuccessCode, &ack
}

// 获取隐藏地理位置开关
func Proto_ClubHouseMemberNoFloorsSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMemberNoFloorsSet)

	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	mem := house.GetMemByUId(req.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	mem.NoFloors = static.I64sToA(req.NoFloors)
	mem.Flush()

	house.Broadcast(consts.ROLE_ADMIN, consts.MsgTypeHouseMemberNoFloorsSet_ntf, req)

	return xerrors.SuccessCode, nil
}

// 设置隐藏地理位置开关
func Proto_ClubHouseMemberNoFloorsGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMemberNoFloorsSet)
	house, _, _, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	mem := house.GetMemByUId(req.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	req.NoFloors = static.AtoI64s(mem.NoFloors)
	return xerrors.SuccessCode, req
}
