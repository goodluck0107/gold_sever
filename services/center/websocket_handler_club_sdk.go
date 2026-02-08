package center

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/cast"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sort"
	"strings"
)

func inspectClubFloorMemberWithRight(hid int, fid, uid int64, role int, keyName string) (*Club, *HouseFloor, *HouseMember, *xerrors.XError) {
	house := GetClubMgr().GetClubHouseByHId(hid)
	if house == nil {
		return nil, nil, nil, xerrors.InValidHouseError
	}
	var floor *HouseFloor
	if fid >= 0 {
		floor = house.GetFloorByFId(fid)
		if floor == nil {
			return house, nil, nil, xerrors.InValidHouseFloorError
		}
	}
	// 玩家
	if uid == 0 {
		return house, floor, nil, xerrors.RespOk
	}
	mem := house.GetMemByUId(uid)
	if mem == nil {
		return house, floor, nil, xerrors.InValidHouseMemberError
	}

	if mem.URole == consts.ROLE_CREATER {
		return house, floor, mem, xerrors.RespOk
	}

	isFaker := GetDBMgr().GetDBrControl().RedisV2.SIsMember("faker_admin", uid).Val()
	if !isFaker {
		// 权限
		if keyName != "" {
			isOK, _ := GetMRightMgr().CheckRight(mem, keyName)
			if !isOK {
				return house, floor, mem, xerrors.InvalidPermission
			}
		} else if mem.Lower(role) {
			return house, floor, mem, xerrors.InvalidPermission
		}
	}
	return house, floor, mem, xerrors.RespOk
}

func getHmListDataByTypeAndRole(listData []HouseMember, limitUser []int64, listType int, optHm *HouseMember) ([]HouseMember, int, int, int) {
	var resArr []HouseMember
	var allNum, onlineNum, jyNum int
	isCap := false
	isVap := false
	if optHm.IsPartner() {
		isCap = true
	} else if optHm.IsVicePartner() {
		isVap = true
	}
	for _, hm := range listData {
		// 计算 全部 在线 禁娱的人数
		if isCap || isVap {
			if (isCap && (hm.Partner == optHm.UId || hm.UId == optHm.UId)) ||
				(isVap && (hm.Partner == optHm.Partner || hm.UId == optHm.Partner)) ||
				(isCap && hm.IsPartner() && hm.Superior == optHm.UId) ||
				(isVap && hm.IsPartner() && hm.Superior == optHm.Partner) { //队长视角
				if hm.IsOnline {
					onlineNum++
				}
				if static.In64(limitUser, hm.UId) {
					jyNum++
				}
				allNum++
			}
		} else {
			if hm.IsOnline {
				onlineNum++
			}
			if static.In64(limitUser, hm.UId) {
				jyNum++
			}
			allNum++
		}
		//非当前页签数据在计算人数之后 跳出当前循环
		if listType == consts.HmUserListOnline {
			if !hm.IsOnline {
				continue
			}
		} else if listType == consts.HmUserListJY {
			if !static.In64(limitUser, hm.UId) {
				continue
			}
		}
		//添加数据
		if isCap || isVap {
			if (isCap && (hm.Partner == optHm.UId || hm.UId == optHm.UId)) ||
				(isVap && (hm.Partner == optHm.Partner || hm.UId == optHm.Partner)) ||
				(isCap && hm.IsPartner() && hm.Superior == optHm.UId) ||
				(isVap && hm.IsPartner() && hm.Superior == optHm.Partner) {
				resArr = append(resArr, hm)
			}
		} else {
			resArr = append(resArr, hm)
		}
	}
	return resArr, allNum, onlineNum, jyNum
}

func searchHmListData(listData []HouseMember, limitUser []int64, searchStr string, listType int, optHm *HouseMember) ([]HouseMember, int, int, int) {
	var resArr []HouseMember
	var allNum, onlineNum, jyNum int
	for _, hm := range listData {
		if optHm.IsPartner() {
			if !(hm.Partner == optHm.UId || hm.UId == optHm.UId) {
				continue
			}
		} else if optHm.IsVicePartner() {
			if !(hm.Partner == optHm.Partner || hm.UId == optHm.Partner) {
				continue
			}
		}
		if strings.Contains(fmt.Sprintf("%d", hm.UId), searchStr) || strings.Contains(hm.NickName, searchStr) {
			hmArr, allNum1, onlineNum1, jyNum1 := searchHmListLogic(limitUser, hm)
			resArr = append(resArr, hmArr...)
			allNum += allNum1
			onlineNum += onlineNum1
			jyNum += jyNum1
			continue
		}
	}
	isFaker := GetDBMgr().GetDBrControl().RedisV2.SIsMember("faker_admin", optHm.UId).Val()
	//兼容模糊搜索的多条数据  如果一条数据并且是队长  那么就展示当前这个队长的小队数据
	if len(resArr) == 1 && resArr[0].IsPartner() {
		if isFaker || optHm.URole == consts.ROLE_CREATER || optHm.URole == consts.ROLE_ADMIN {
			return getHmListDataByTypeAndRole(listData, limitUser, listType, &resArr[0])
		}
	}

	return resArr, allNum, onlineNum, jyNum
}

func searchHmListLogic(limitUser []int64, hm HouseMember) ([]HouseMember, int, int, int) {
	var resArr []HouseMember
	var allNum, onlineNum, jyNum int
	resArr = append(resArr, hm)
	allNum++
	if hm.IsOnline {
		onlineNum++
	}
	if static.In64(limitUser, hm.UId) {
		jyNum++
	}
	return resArr, allNum, onlineNum, jyNum
}

func clubHouseMemberList(conditionArr []HouseMember, house *Club, houseMemberMap map[int64]HouseMember, partnerAttrMap map[int64]models.HousePartnerAttr, roleType int, sortType int, optMem *HouseMember) []*static.Msg_HouseMemberItem {
	// xlog.Logger().Warn("clubHouseMemberList 1")
	var userItems []*static.Msg_HouseMemberItem
	userItems = make([]*static.Msg_HouseMemberItem, 0)
	cli := GetDBMgr().Redis
	for _, hmem := range conditionArr {
		// xlog.Logger().Warn("clubHouseMemberList.GetPerson 1")
		dmem, err := GetDBMgr().GetDBrControl().GetPerson(hmem.UId)
		// xlog.Logger().Warn("clubHouseMemberList.GetPerson 2")
		if err != nil {
			xlog.Logger().Errorf("user not exists in house:%d,uid:%d", house.DBClub.Id, hmem.UId)
			house.MemDelete(hmem.UId, true, nil)
			continue
		}
		var titem static.Msg_HouseMemberItem
		titem.UId = hmem.UId
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
			} else {
				hmem.Partner = 0 // 错误数据修复
			}
		}
		if optMem.IsPartner() || optMem.IsVicePartner() {
			titem.URemark = hmem.PRemark
		} else {
			titem.URemark = hmem.URemark
		}
		// 上级队长
		if hmem.Superior > 0 {
			dsmem, ok := houseMemberMap[hmem.Superior]
			if ok {
				titem.Superior = hmem.Superior
				titem.SuperiorName = dsmem.NickName
			} else {
				titem.Superior = 0
				titem.SuperiorName = ""
			}
		}
		titem.UUrl = hmem.ImgUrl
		titem.UGender = hmem.Sex
		titem.UJoinTime = hmem.ApplyTime
		titem.GameLimit = cli.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, house.DBClub.Id), hmem.UId).Val()
		titem.LastLoginAt = dmem.LastOffLineTime
		if hmem.IsPartner() {
			attr := partnerAttrMap[hmem.UId]
			titem.TeamBan = attr.TeamBan
			titem.AA = attr.AA
		}
		userItems = append(userItems, &titem)
	}
	// xlog.Logger().Warn("clubHouseMemberList 2")
	if roleType == consts.ROLE_MEMBER {
		switch sortType {
		case static.SORT_VITAMIN_DES:
			sort.Sort(&static.MsgHouseMemberItemWrapper{
				Hms: userItems,
				By: func(i, j *static.Msg_HouseMemberItem) bool {
					return i.UVitamin > j.UVitamin
				},
			})
		case static.SORT_VITAMIN_AES:
			sort.Sort(&static.MsgHouseMemberItemWrapper{
				Hms: userItems,
				By: func(i, j *static.Msg_HouseMemberItem) bool {
					return i.UVitamin < j.UVitamin
				},
			})
		case static.SORT_OFFTIME_DES:
			sort.Sort(&static.MsgHouseMemberItemWrapper{
				Hms: userItems,
				By: func(i, j *static.Msg_HouseMemberItem) bool {
					a := i.LastLoginAt
					if a == 0 {
						a = i.UJoinTime
					}
					b := j.LastLoginAt
					if b == 0 {
						b = j.UJoinTime
					}
					return a > b
				},
			})
		case static.SORT_OFFTIME_AES:
			sort.Sort(&static.MsgHouseMemberItemWrapper{
				Hms: userItems,
				By: func(i, j *static.Msg_HouseMemberItem) bool {
					a := i.LastLoginAt
					if a == 0 {
						a = i.UJoinTime
					}
					b := j.LastLoginAt
					if b == 0 {
						b = j.UJoinTime
					}
					return a < b
				},
			})
		}
	}
	// xlog.Logger().Warn("clubHouseMemberList 3")
	return userItems
}

func HouseMemAgreeHandle(req *static.Msg_CH_HouseMemberAgree, optid int64, opName string) *xerrors.XError {
	if req.UId == optid {
		return xerrors.InvalidParamError
	}

	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError
	}

	optMem := house.GetMemByUId(optid)
	if optMem == nil {
		return xerrors.InValidHouseMemberError
	}
	if req.ApplyType == consts.HouseMemberApplyJoin {
		isOK, _ := GetMRightMgr().CheckRight(optMem, MinorJoinReviewed)
		if !isOK {
			return xerrors.InvalidPermission
		}
	} else if req.ApplyType == consts.HouseMemberApplyExit {
		isOK, _ := GetMRightMgr().CheckRight(optMem, MinorOutReviewed)
		if !isOK {
			return xerrors.InvalidPermission
		}
	}

	// 记录队长ID
	mem := house.GetMemByUId(req.UId)
	partnerId := int64(0)
	if mem != nil && mem.Partner > 0 {
		partnerId = mem.Partner
	}

	if req.ApplyType == consts.HouseMemberApplyJoin {
		// 人数上限
		if house.GetMemCounts() >= GetServer().ConHouse.MemMax {
			return xerrors.HouseMemJoinMaxError
		}

		// 玩家加入上限
		count, err := GetDBMgr().GetDBrControl().HouseMemberJoinCounts(req.UId)
		if err != nil {
			return xerrors.DBExecError
		}
		if count >= GetServer().ConHouse.JoinMax {
			return xerrors.MemJoinHouseMaxError
		}

		// 成员过审
		custerr := house.MemAgree(optid, req.UId, nil)
		if custerr != nil {
			return custerr
		}

		msg := fmt.Sprintf("<color=#F93030>%sID:%d</color>通过<color=#00A70C>ID:%d</color>审核加入包厢", opName, optid, req.UId)
		CreateClubMassage(house.DBClub.Id, optid, JoinHouse, msg)
		// 新增成员如果是通过小圈申请进来的 在小圈消息中增加一条记录
		house2 := GetClubMgr().GetClubHouseById(mem.Ref)
		if mem.Ref > 0 && house2 != nil {
			msg := fmt.Sprintf("<color=#F93030>%sID:%d</color>通过<color=#00A70C>ID:%d</color>审核加入包厢(合并包厢<color=#00A70C>ID:%d</color>)", opName, optid, req.UId, house.DBClub.HId)
			CreateClubMassage(house2.DBClub.Id, optid, JoinHouse, msg)
		}
	} else if req.ApplyType == consts.HouseMemberApplyExit {
		cer := house.MemberExitAgree(optid, req)
		if cer != nil {
			return cer
		}
		//// 清空禁止娱乐状态
		err := house.LimitUserGame(req.UId, req.UId, true)
		if err != nil {
			return xerrors.DBExecError
		}
		msg := fmt.Sprintf("<color=#F93030>%sID:%d</color>通过了<color=#00A70C>ID:%d</color>发起的退出包厢审核", opName, optid, req.UId)
		CreateClubMassage(house.DBClub.Id, optid, ExitHouse, msg)
	}
	house.RedisPub(consts.MsgTypeHouseMemberAgree_NTF, req)

	// 广播给队长
	if partnerId > 0 {
		house.Broadcast2Parnter(partnerId, true, consts.MsgTypeHouseMemberAgree_NTF, &static.Msg_CH_HouseMemberApplyNtf{
			Msg_CH_HouseMemberApply: req.Msg_CH_HouseMemberApply,
			Opt:                     optid,
		})
	}

	house.CustomBroadcast(consts.ROLE_ADMIN, false, false, false, consts.MsgTypeHouseMemberAgree_NTF, &static.Msg_CH_HouseMemberApplyNtf{
		Msg_CH_HouseMemberApply: req.Msg_CH_HouseMemberApply,
		Opt:                     optid,
	}, req.UId)

	return nil
}

func HouseMemRefuseHandle(req *static.Msg_CH_HouseMemberRefused, optid int64, opName string) *xerrors.XError {
	if req.UId == optid {
		return xerrors.InvalidParamError
	}
	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError
	}

	optMem := house.GetMemByUId(optid)
	if optMem == nil {
		return xerrors.InValidHouseMemberError
	}
	if req.ApplyType == consts.HouseMemberApplyJoin {
		isOK, _ := GetMRightMgr().CheckRight(optMem, MinorJoinReviewed)
		if !isOK {
			return xerrors.InvalidPermission
		}
	} else if req.ApplyType == consts.HouseMemberApplyExit {
		isOK, _ := GetMRightMgr().CheckRight(optMem, MinorOutReviewed)
		if !isOK {
			return xerrors.InvalidPermission
		}
	}

	// 记录队长ID
	mem := house.GetMemByUId(req.UId)
	partnerId := int64(0)
	if mem != nil && mem.Partner > 0 {
		partnerId = mem.Partner
	}

	if req.ApplyType == consts.HouseMemberApplyJoin {
		// 成员拒审
		cer := house.MemRefused(optid, req.UId)
		if cer != nil {
			return cer
		}

	} else if req.ApplyType == consts.HouseMemberApplyExit {
		cer := house.MemberExitRefuse(optid, req)
		if cer != nil {
			return cer
		}
		msg := fmt.Sprintf("<color=#F93030>%sID:%d</color>拒绝了<color=#00A70C>ID:%d</color>发起的退出包厢审核", opName, optid, req.UId)
		CreateClubMassage(house.DBClub.Id, optid, ExitHouse, msg)
	}
	house.RedisPub(consts.MsgTypeHouseMemberRefused_NTF, req)

	// 广播给队长
	if partnerId > 0 {
		house.Broadcast2Parnter(partnerId, true, consts.MsgTypeHouseMemberRefused_NTF, &static.Msg_CH_HouseMemberApplyNtf{
			Msg_CH_HouseMemberApply: req.Msg_CH_HouseMemberApply,
			Opt:                     optid,
		})
	}

	house.CustomBroadcast(consts.ROLE_ADMIN, false, false, false, consts.MsgTypeHouseMemberRefused_NTF, &static.Msg_CH_HouseMemberApplyNtf{
		Msg_CH_HouseMemberApply: req.Msg_CH_HouseMemberApply,
		Opt:                     optid,
	}, req.UId)

	return nil
}

// 获取区域下的游戏列表
// func getGameList(area string) []*public.Msg_S2C_GameList {
//
// 	// return make([]*public.Msg_S2C_GameList, 0)
// 	pkgs := GetAreaPackagesByCode(area, false)
// 	result := make([]*public.Msg_S2C_GameList, 0)
// 	for _, pkg := range pkgs {
// 		for _, game := range pkg.Games {
// 			obj := new(public.Msg_S2C_GameList)
// 			obj.PackageVersion = game.PackageVersion
// 			obj.PackageName = game.PackageName
// 			obj.PackageKey = game.PackageKey
// 			obj.Name = game.Name
// 			obj.KindId = game.KindId
// 			obj.Icon = game.Icon
// 			siteTypes := GetServer().GetSiteTypeByKindId(public.GAME_TYPE_GOLD, game.KindId)
// 			totalOnline := 0
// 			for _, item := range siteTypes {
// 				v := &public.Msg_S2C_SiteType{
// 					Type: item,
// 				}
// 				// 获取房间配置
// 				c := GetServer().GetRoomConfig(obj.KindId, item)
// 				if c != nil {
// 					v.Name = c.Name
// 					v.MaxScore = c.MaxScore
// 					v.MinScore = c.MinScore
// 					v.Difen = int(c.Config["difen"].(float64))
// 					// 获取该游戏该场次的在线人数(加上基础人数)
// 					v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type) + c.BaseNum
// 					v.SitMode = c.SitMode
// 					v.MatchFlag = 0
// 					mat := GetServer().GetMatchConfig(game.KindId, v.Type) //是否开启了排位赛
// 					if mat != nil {
// 						v.MatchFlag = mat.Flag
// 						if mat.Flag > 0 {
// 							obj.MatchFlag = mat.Flag
// 						}
// 					}
// 				} else {
// 					v.SitMode = 0
// 					v.MaxScore = 0
// 					v.MinScore = 0
// 					// 获取该游戏该场次的在线人数(加上基础人数)
// 					v.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type)
// 				}
// 				totalOnline = totalOnline + v.Online
// 				obj.SiteType = append(obj.SiteType, v)
// 			}
// 			// chess需要 不屏蔽
// 			//if len(obj.SiteType) > 0 {
// 			obj.Online = totalOnline
// 			result = append(result, obj)
// 			//}
// 		}
//
// 	}
// 	return result
// }

// 获取排位赛列表
func getGameMatchList(area string, engine int) []*static.Msg_S2C_MatchGameList {
	return make([]*static.Msg_S2C_MatchGameList, 0)
	// games := GetAreaMgr().GetGames(area, engine)
	// result := make([]*public.Msg_S2C_MatchGameList, 0)
	// for _, game := range games {
	// 	siteTypes := GetServer().GetSiteTypeByKindId(public.GAME_TYPE_GOLD, game.KindId)
	// 	for _, item := range siteTypes {
	// 		v := &public.Msg_S2C_SiteType{
	// 			Type: item,
	// 		}
	// 		// 获取房间配置
	// 		c := GetServer().GetRoomConfig(game.KindId, item)
	// 		if c != nil {
	// 			obj := new(public.Msg_S2C_MatchGameList)
	// 			obj.Name = game.Name
	// 			obj.KindId = game.KindId
	// 			obj.Icon = game.Icon
	// 			// 获取该游戏该场次的在线人数(加上基础人数)
	// 			obj.Online = GetPlayerMgr().GetOnlineNumberByKindId(game.KindId, v.Type) + c.BaseNum
	// 			obj.SitMode = c.SitMode
	// 			obj.SiteType = v.Type
	// 			obj.MatchFlag = 0
	// 			mat := GetServer().GetMatchConfig(game.KindId, v.Type) //是否开启了排位赛
	// 			if mat != nil {
	// 				if mat.Flag > 0 {
	// 					obj.Name = mat.Name
	// 					obj.MatchFlag = mat.Flag
	// 					obj.State = mat.State
	// 					obj.BeginTime = mat.BeginTime.Unix()
	// 					obj.EndTime = mat.EndTime.Unix()
	// 					//只要开启了话费赛的游戏
	// 					result = append(result, obj)
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// return result
}

// 创建房间参数校验
func validateCreateTableParam(req *static.Msg_CreateTable, club bool) (*static.Msg_HG_CreateTable, *xerrors.XError) {
	result := new(static.Msg_HG_CreateTable)

	// 获取玩法默认配置
	config := GetServer().GetGameConfig(req.KindId)
	if config == nil {
		// 没有找到玩法对应的配置
		cuserror := xerrors.NewXError("不支持的游戏玩法")
		return nil, cuserror
	}
	result.KindId = req.KindId

	// 默认自己开房自己玩
	result.CreateType = consts.CreateTypeSelf
	// 少人开局效验
	min, max := config.GetPlayerNum()
	if !(max-min > 0) {
		req.FewerStart = "false"
	} else if !(req.PlayerNum > min) {
		req.FewerStart = "false"
	}

	gameConfig := make(map[string]interface{})

	if req.GameConfig != "" {
		err := json.Unmarshal([]byte(req.GameConfig), &gameConfig)
		if err != nil {
			xlog.Logger().Error("parse game config:", req.GameConfig, err)
		}
	}

	result.FewerStart = req.FewerStart
	if req.Adddifen > 0 {
		result.Difen = req.Adddifen
	}
	if result.Difen == 0 {
		result.Difen = req.Difen
	}
	if result.Difen == 0 {
		if val, ok := gameConfig["dff"]; ok {
			result.Difen = cast.ToInt64(val) * 100
		}
	}

	// 语音权限
	result.GVoice = req.GVoice

	// 语音限制
	if result.GVoice == "true" {
		if !req.Voice {
			// 未开启语音授权
			// return nil, xerrors.OpenGVoiceError
		}
		if !req.GVoiceOk {
			// 未开启语音授权
			return nil, xerrors.InitGVoiceError
		}
	}

	// ip限制
	if req.Restrict == "false" {
		result.Restrict = false
	} else if req.Restrict == "true" {
		result.Restrict = true
	} else {
		result.Restrict = config.DefaultRestrict
	}
	// 游戏人数判断
	if req.PlayerNum == 0 {
		// 如果客户端没传, 则最大最小人数从配置文件读取
		result.MinPlayerNum, result.MaxPlayerNum = config.GetPlayerNum()
		// 设置为默认人数
		req.PlayerNum = config.DefaultPlayerNum
	} else {
		if !config.CheckPlayerNum(req.PlayerNum) {
			// 玩家人数错误
			cuserror := xerrors.NewXError("玩家人数不合法")
			return nil, cuserror
		}

		// 如果客户端限定的玩家人数, 则最大最小人数都取该值
		result.MinPlayerNum = req.PlayerNum
		result.MaxPlayerNum = req.PlayerNum
	}

	// gps限制
	if result.MaxPlayerNum != 2 {
		if result.Restrict && !req.Gps {
			// 未开启gps服务
			return nil, xerrors.OpenGpsError
		}
	}

	// 游戏局数判断
	if req.RoundNum == 0 {
		result.RoundNum = config.DefaultRoundNum
	} else {
		if !config.CheckRoundNum(req.RoundNum, req.PlayerNum) {
			// 游戏局数错误
			cuserror := xerrors.NewXError("游戏局数不合法")
			return nil, cuserror
		}
		result.RoundNum = req.RoundNum
	}

	// 游戏房卡消耗
	result.CardCost = config.GetCardCost(result.RoundNum, req.PlayerNum, club && consts.ClubHouseOwnerPay)

	// 房卡支付方式判断
	if req.CostType == 0 {
		result.CostType = config.DefaultCostType
	} else {
		if req.CostType != consts.CostTypeCreator && req.CostType != consts.CostTypeWiner {
			// 房卡支付方式错误
			cuserror := xerrors.NewXError("房卡支付方式不合法")
			return nil, cuserror
		}
		result.CostType = req.CostType
	}

	// 默认不允许观看
	result.View = config.DefaultView

	// TODO: 特殊参数校验
	result.GameConfig = req.GameConfig
	result.IsLimitChannel = config.IsLimitChannel
	return result, nil
}

func SortVitaminMgrItems(sortType int, conItems []*static.MsgVitaminMgrItem) []*static.MsgVitaminMgrItem {
	if sortType == static.SORT_VITAMIN_BY_ROLE {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.URole < item2.URole {
				return true
			} else if item1.URole == item2.URole {
				if item1.IsPartner && !item2.IsPartner {
					return true
				} else if item2.IsPartner && !item1.IsPartner {
					return false
				} else {
					return item1.UId > item2.UId
				}
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINCUR_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.CurVitamin > item2.CurVitamin {
				return true
			} else if item1.CurVitamin == item2.CurVitamin {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINCUR_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.CurVitamin < item2.CurVitamin {
				return true
			} else if item1.CurVitamin == item2.CurVitamin {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINPRE_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.PreNodeVitamin > item2.PreNodeVitamin {
				return true
			} else if item1.PreNodeVitamin == item2.PreNodeVitamin {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINPRE_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.PreNodeVitamin < item2.PreNodeVitamin {
				return true
			} else if item1.PreNodeVitamin == item2.PreNodeVitamin {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINWINLOSE_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.VitaminWinLoseCost > item2.VitaminWinLoseCost {
				return true
			} else if item1.VitaminWinLoseCost == item2.VitaminWinLoseCost {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINWINLOSE_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.VitaminWinLoseCost < item2.VitaminWinLoseCost {
				return true
			} else if item1.VitaminWinLoseCost == item2.VitaminWinLoseCost {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINPLAYCOST_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.VitaminPlayCost > item2.VitaminPlayCost {
				return true
			} else if item1.VitaminPlayCost == item2.VitaminPlayCost {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINPLAYCOST_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.VitaminPlayCost < item2.VitaminPlayCost {
				return true
			} else if item1.VitaminPlayCost == item2.VitaminPlayCost {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINPLAYTIMES_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.PlayTimes > item2.PlayTimes {
				return true
			} else if item1.PlayTimes == item2.PlayTimes {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINPLAYTIMES_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.PlayTimes < item2.PlayTimes {
				return true
			} else if item1.PlayTimes == item2.PlayTimes {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINBWTIMES_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.BwTimes > item2.BwTimes {
				return true
			} else if item1.BwTimes == item2.BwTimes {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINBWTIMES_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.BwTimes < item2.BwTimes {
				return true
			} else if item1.BwTimes == item2.BwTimes {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINVALIDROUND_DESC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.ValidTimes > item2.ValidTimes {
				return true
			} else if item1.ValidTimes == item2.ValidTimes {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if sortType == static.SORT_VITAMINVALIDROUND_ASC {
		sort.Sort(static.MsgVitaminMgrItemWrapper{Item: conItems, By: func(item1, item2 *static.MsgVitaminMgrItem) bool {
			if item1.ValidTimes < item2.ValidTimes {
				return true
			} else if item1.ValidTimes == item2.ValidTimes {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	}
	return conItems
}
