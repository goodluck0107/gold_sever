package center

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sort"
	"strings"
	"time"
)

// 包厢队长权限
func Proto_ClubHousePartner(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	_, xerr := GetDBMgr().GetUserAgentConfig(p.Uid)
	if xerr != nil {
		if xerr.Code == xerrors.UserAgentNotConfigError.Code {
			cwxs := GetAreaWeChat(p.Area)
			msg := ""
			if len(cwxs.WeChat) > 0 {
				msg = cwxs.WeChat[0].Wx
			} else {
				msg = "111"
			}
			return xerrors.InvalidPermission.Code, msg
		}
		return xerr.Code, xerr.Msg
	}

	return xerrors.SuccessCode, nil
}

// 包厢合伙权限
func Proto_ClubHousePartnerList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseMemStatistics)
	if !ok {
		return xerrors.ResultErrorCode, nil
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
	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		if cusErr == xerrors.InvalidPermission {
			if !optmem.IsVitaminAdmin() {
				return cusErr.Code, cusErr.Msg
			}
		} else {
			return cusErr.Code, cusErr.Msg
		}
	}
	statisticsmap := make(map[int64]*static.HouseMemberStatisticsItem)
	houseMem := house.GetMemberMap(false)

	for _, memp := range houseMem {
		if memp.IsPartner() {
			sitem := new(static.HouseMemberStatisticsItem)
			sitem.UId = memp.UId
			sitem.UName = memp.NickName
			sitem.UUrl = memp.ImgUrl
			sitem.UGender = memp.Sex
			sitem.UJoinTime = 0
			sitem.PlayTimes = 0
			sitem.BwTimes = 0
			sitem.TotalScore = 0
			sitem.ValidTimes = 0
			sitem.BigValidTimes = 0
			statisticsmap[memp.UId] = sitem
		}
	}

	statisticsItems, _ := GetDBMgr().SelectHouseMemberStatistics(int(house.DBClub.Id), req.DFid, req.DayType, -1)

	for _, item := range statisticsItems {
		mem, ok := houseMem[item.UId]
		if !ok {
			continue
		}
		partnerId := mem.Partner
		if mem.IsPartner() {
			partnerId = mem.UId
		}

		if statisticsmap[partnerId] != nil {
			statisticsmap[partnerId].PlayTimes += item.PlayTimes
			statisticsmap[partnerId].BwTimes += item.BwTimes
			statisticsmap[partnerId].TotalScore += item.TotalScore
			statisticsmap[partnerId].ValidTimes += item.ValidTimes
			statisticsmap[partnerId].BigValidTimes += item.BigValidTimes
		}
	}

	var totlalItems []*static.HouseMemberStatisticsItem
	for _, item := range statisticsmap {
		item.TotalScore = static.HF_DecimalDivide(item.TotalScore, 1, 2)
		totlalItems = append(totlalItems, item)
	}

	var searchItems []*static.HouseMemberStatisticsItem
	// 搜索功能
	if req.SearchKey != "" {
		for _, sItem := range totlalItems {
			if strings.Contains(sItem.UName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", sItem.UId), req.SearchKey) {
				searchItems = append(searchItems, sItem)
			}
		}
	} else {
		searchItems = append(searchItems, totlalItems...)
	}

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
	}
	// 开始索引
	var idxBeg int
	var idxEnd int
	idxBeg = req.PBegin
	if req.PEnd > len(searchItems) {
		idxEnd = len(searchItems)
	} else {
		idxEnd = req.PEnd
	}
	if idxBeg >= len(searchItems) {
		searchItems = []*static.HouseMemberStatisticsItem{}
	} else {
		searchItems = searchItems[idxBeg:idxEnd]
	}
	var ack static.Msg_HC_HouseMemberStatistics
	ack.PBegin = reqBegin
	ack.PEnd = reqEnd
	ack.Items = make([]*static.HouseMemberStatisticsItem, 0)
	ack.Items = append(ack.Items, searchItems...)

	return xerrors.SuccessCode, ack
}

// 包厢队长成员调配
func Proto_ClubHousePartnerMemCustom(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HousePartnerMemCustom)
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
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 队长Proto_ClubHouseVitaminStatisticClear
	partner := house.GetMemByUId(req.PId)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	if partner.Partner != 1 {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	var arrPartner []HouseMember
	if req.IsBind {
		arrPartner = house.GetMemByPartnerOrderById(req.PId)
	} else {
		arrPartner = house.GetMemNoPartnerMemOrderById()
	}

	var ack static.Msg_HC_HousePartnerList

	// 条件切片
	var conditionArr []HouseMember
	if req.Param == "" {
		conditionArr = arrPartner
	} else {
		for _, mem := range arrPartner {
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
	ack.FMems = make([]static.Msg_HouseMemberItem, 0)

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
		var titem static.Msg_HouseMemberItem
		titem.UId = mem.UId
		titem.UOnline = GetPlayerMgr().IsUserOnline(mem.UId)
		titem.UName = mem.NickName
		titem.URole = mem.URole
		titem.UVitamin = static.SwitchVitaminToF64(mem.UVitamin)
		titem.URemark = mem.URemark
		titem.UUrl = mem.ImgUrl
		titem.UGender = mem.Sex
		titem.UJoinTime = mem.ApplyTime
		titem.UPartner = mem.Partner
		ack.FMems = append(ack.FMems, titem)
	}

	return xerrors.SuccessCode, &ack
}

// 包厢队长权限
func Proto_ClubHousePartnerCreate(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HousePartnerCreate)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorSetCaptain)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	hp, _ := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
	if hp == nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	partnerMem := house.GetMemByUId(req.Uid)
	if partnerMem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Code
	}
	//_, xerr := GetDBMgr().GetUserAgentConfig(partnerMem.UId)
	//if xerr != nil {
	//	if xerr.Code == xerrors.UserAgentNotConfigErrorCode {
	//		return xerrors.ResultErrorCode, "该玩家未注册盟主后台，无法设置成为队长。"
	//	}
	//	return xerr.Code, xerr.Msg
	//}
	// 修改状态
	custerr := house.ModifyPartner(partnerMem.Id, 1)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	// 初始化队长经验
	house.InitCreatePartnerExp(partnerMem.UId)

	ntf := new(static.Ntf_HC_HousePartnerCreate)
	ntf.HId = house.DBClub.HId
	ntf.UId = req.Uid

	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHousePartnerGreate_Ntf, ntf)
	msg := fmt.Sprintf("盟主将%sID:%d设为队长", hp.Nickname, hp.Uid)
	CreateClubMassage(house.DBClub.Id, p.Uid, PartnerChange, msg)
	return xerrors.SuccessCode, nil
}

// 包厢队长权限
func Proto_ClubHousePartnerDelete(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HousePartnerDelete)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorSetCaptain)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}
	hp, _ := GetDBMgr().GetDBrControl().GetPerson(req.UId)
	if hp == nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}
	partmem := house.GetMemByUId(req.UId)
	if partmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Code
	}
	if !partmem.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	if _, ok := partmem.CheckRef(); ok {
		return xerrors.ResultErrorCode, "该队长为合并包厢盟主，无法解除队长身份。"
	}

	if xe := house.PartnerDelete(partmem, nil); xe != nil {
		return xe.Code, xe.Msg
	}
	house.DeletePartner(req.UId)

	// 删除队长经验
	house.DelCreatePartnerExp(req.UId)

	ntf := new(static.Ntf_HC_HousePartnerDelete)
	ntf.HId = house.DBClub.HId
	ntf.UId = req.UId
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHousePartnerDelete_Ntf, ntf)
	msg := fmt.Sprintf("盟主将%sID:%d取消队长", hp.Nickname, hp.Uid)
	CreateClubMassage(house.DBClub.Id, p.Uid, PartnerChange, msg)
	GetMRightMgr().setRoleUpdateRight(partmem, true)
	return xerrors.SuccessCode, nil
}

// 分配包厢队长成员
func Proto_ClubHousePartnerGen(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HousePartnerGen)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorSetCaptain)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	// 源玩家
	hmem := house.GetMemByUId(req.UId)
	if hmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if hmem.URole <= consts.ROLE_ADMIN {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if hmem.UVitamin != 0 {
		return xerrors.ResultErrorCode, "玩家比赛分必须为0才能调配。"
	}
	oldPartner := hmem.Partner
	// 目标玩家
	if req.Partner == 0 {
		// 解绑
		// 取消队长身份 时 自动取消名下玩家
		if hmem.Partner == 1 { // 取消队长应该通过另外一个接口
			return xerrors.InValidHouseMemberRoleError.Code, xerrors.InValidHouseMemberRoleError.Msg
		}

		custerr := house.ModifyPartner(hmem.Id, 0)
		if custerr != nil {
			return custerr.Code, custerr.Msg
		}
		house.RemoveGroupUser(hmem)
		if oldPartner > 1 {
			hp, err := GetDBMgr().GetDBrControl().GetPerson(oldPartner)
			if hp == nil || err != nil {
				return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
			}
			hpMem, err := GetDBMgr().GetDBrControl().GetPerson(hmem.UId)
			if hpMem == nil || err != nil {
				return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
			}
			msg := fmt.Sprintf("盟主删除了队长%sID:%d名下成员%sID:%d", hp.Nickname, hp.Uid, hpMem.Nickname, hpMem.Uid)
			CreateClubMassage(house.DBClub.Id, p.Uid, PartnerUserChange, msg)
		}

	} else if req.Partner == 1 {
		if hmem.Partner == 1 { // 任命队长应该通过另外一个接口
			return xerrors.InValidHouseMemberRoleError.Code, xerrors.InValidHouseMemberRoleError.Msg
		}
	} else {
		// 绑定
		if hmem.URole == consts.ROLE_CREATER {
			return xerrors.HouseCreatorRefusePartnerError.Code, xerrors.HouseCreatorRefusePartnerError.Msg
		}
		if hmem.Partner > 1 { // 玩家名下有队长应该先解除
			return xerrors.DelPartnerFirst.Code, xerrors.DelPartnerFirst.Msg
		}
		if hmem.VitaminAdmin {
			return xerrors.InValidHouseMemberRoleError.Code, xerrors.InValidHouseMemberRoleError.Msg
		}
		hpmem := house.GetMemByUId(req.Partner)
		if hpmem == nil {
			return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
		}
		// 修改状态
		custerr := house.ModifyPartner(hmem.Id, hpmem.UId)
		if custerr != nil {
			return custerr.Code, custerr.Msg
		}
		// 跟随队长是否被限制游戏
		if hpmem.IsLimitGame {
			house.LimitUserGame(hpmem.UId, hmem.UId, false)
		}
		hp, err := GetDBMgr().GetDBrControl().GetPerson(req.Partner)
		if hp == nil || err != nil {
			return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
		}
		hpMem, err := GetDBMgr().GetDBrControl().GetPerson(hmem.UId)
		if hpMem == nil || err != nil {
			return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
		}
		msg := fmt.Sprintf("盟主调配了%sID:%d给队长%sID:%d", hpMem.Nickname, hpMem.Uid, hp.Nickname, hp.Uid)
		CreateClubMassage(house.DBClub.Id, p.Uid, PartnerUserChange, msg)
	}

	ntfUids := append([]int64{}, p.Uid, hmem.UId, oldPartner, req.Partner)
	var ntf static.Ntf_HC_HousePartnerGen
	ntf.HId = house.DBClub.HId
	ntf.Uid = hmem.UId
	ntf.OptId = p.Uid
	ntf.OldPartner = oldPartner
	ntf.NewPartner = req.Partner
	for i := 0; i < len(ntfUids); i++ {
		p := GetPlayerMgr().GetPlayer(ntfUids[i])
		if p != nil {
			p.SendMsg(consts.MsgTypeHousePartnerGen_Ntf, &ntf)
		}
	}

	return xerrors.SuccessCode, nil
}

// 包厢疲劳值数值
func Proto_ClubHouseVitaminSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminSet)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}
	if req.Value >= 0 {
		return Proto_ClubHouseVitaminSend(s, p, data)
	}
	var modByPartner bool
	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	optMemURole := optmem.URole
	if req.UId == p.Uid {
		xlog.Logger().Warn("opt self vitamin, return")
		return xerrors.SuccessCode, nil
	}
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}
	if optmem.URole == consts.ROLE_CREATER || optmem.IsVitaminAdmin() {

	} else if optmem.URole == consts.ROLE_ADMIN {
		if !house.DBClub.IsVitaminModi {
			return xerrors.HouseAdminSetVitaminDisableError.Code, xerrors.HouseAdminSetVitaminDisableError.Msg
		}
	} else {
		mem := house.GetMemByUId(req.UId)
		if mem == nil {
			return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
		}
		var partnerId int64
		if optmem.IsPartner() {
			if req.UId == p.Uid {
				return xerrors.SuccessCode, nil
			}
			partnerId = optmem.UId
		} else if optmem.IsVicePartner() {
			if req.UId == p.Uid {
				return xerrors.SuccessCode, nil
			}
			if mem.UId == optmem.Partner || mem.IsVicePartner() {
				return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
			}
			partnerId = optmem.Partner
		}
		if partnerId > 0 {
			modByPartner = true
			if mem.Superior == partnerId {
				if house.DBClub.DisVitaminJunior {
					return xerrors.HousePartnerSetVitaminDisableError.Code, xerrors.HousePartnerSetVitaminDisableError.Msg
				}
			} else if mem.Partner == partnerId {
				if !house.DBClub.IsPartnerModi {
					return xerrors.HousePartnerSetVitaminDisableError2.Code, xerrors.HousePartnerSetVitaminDisableError2.Msg
				}
			} else {
				return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
			}
		} else {
			return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
		}
	}

	// 获取锁
	mem := house.GetMemByUId(req.UId)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	pSender, err := GetDBMgr().GetDBrControl().GetPerson(req.UId)
	if pSender == nil || err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if pSender.TableId > 0 {
		return xerrors.UserInGameDescVitaminError.Code, xerrors.UserInGameDescVitaminError.Msg
	}

	cli := GetDBMgr().Redis
	mem.Lock(cli)

	tx := GetDBMgr().GetDBmControl().Begin()
	// defer tx.Rollback()
	var suc bool
	defer func() {
		if !suc {
			tx.Rollback()
		}
	}()
	mvitamin, err := mem.GetVitaminFromDbWithLock(tx)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	mem.UVitamin = mvitamin

	value := static.SwitchF64ToVitamin(req.Value)

	// 如果是扣除
	if req.Value < 0 {
		if mem.UVitamin < 0 {
			xerr := xerrors.NewXError("扣除失败，扣除数值不能大于玩家剩余比赛分数量")
			mem.Unlock(cli)
			return xerr.Code, xerr.Msg
		} else if value+mem.UVitamin < 0 {
			mem.Unlock(cli)
			xerr := xerrors.NewXError("扣除失败，扣除数值不能大于玩家剩余比赛分数量")
			return xerr.Code, xerr.Msg
		}
	}

	var modelType models.VitaminChangeType
	if modByPartner {
		modelType = models.PartnerSend
	} else {
		if optmem.VitaminAdmin {
			modelType = models.ViAdminSet
		} else {
			modelType = models.MemberCost
		}
	}

	var ownerPartner bool
	if mem.Partner > 1 {
		if p.Uid != mem.Partner {
			if house.DBClub.NoSkipVitaminSet {
				return xerrors.ResultErrorCode, "请先打开比赛场-跨级调整开关。"
			} else {
				optmem = house.GetMemByUId(mem.Partner)
				if optmem == nil {
					xlog.Logger().Error("nil partner", mem.Partner)
					return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
				}
				ownerPartner = true
			}
		}
	}

	_, aftVitamin, err := mem.VitaminIncrement(p.Uid, value, modelType, tx)
	if err != nil {
		xlog.Logger().Errorf("change user vitamin error:%v", err)
		mem.Unlock(cli)
		return xerrors.DBExecErrorCode, xerrors.DBExecError.Msg
	}

	optmem.Lock(cli)
	pvitamin, err := optmem.GetVitaminFromDbWithLock(tx)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	optmem.UVitamin = pvitamin
	// 获取疲劳值
	_, _, err = optmem.VitaminIncrement(req.UId, -1*value, modelType, tx)
	if err != nil {
		xlog.Logger().Errorf("change user vitamin error:%v", err)
		optmem.Unlock(cli)
		mem.Unlock(cli)
		return xerrors.DBExecErrorCode, xerrors.DBExecError.Msg
	}
	optmem.Flush()
	optmem.UVitamin = optmem.UVitamin
	optmem.Unlock(cli)

	//if modByPartner {
	//	parMemm := house.GetMemByUId(p.Uid)
	//	parMemm.Lock(cli)
	//	pvitamin, err := parMemm.GetVitaminFromDbWithLock(tx)
	//	if err != nil {
	//		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	//	}
	//	parMemm.UVitamin = pvitamin
	//	// 获取疲劳值
	//	_, _, err = parMemm.VitaminIncrement(req.UId, -1*value, models.PartnerSend, tx)
	//	if err != nil {
	//		xlog.Logger().Errorf("change user vitamin error:%v", err)
	//		parMemm.Unlock(cli)
	//		mem.Unlock(cli)
	//		return xerrors.DBExecErrorCode, xerrors.DBExecError.Msg
	//	}
	//	parMemm.Flush()
	//	optmem.UVitamin = parMemm.UVitamin
	//	parMemm.Unlock(cli)
	//} else {
	//	if optmem.VitaminAdmin {
	//		xerr := house.PoolChange(p.Uid, models.ViAdminSet, -1*value, tx)
	//		if xerr != nil {
	//			xlog.Logger().Errorf("change user vitamin error:%v", err)
	//			mem.Unlock(cli)
	//			return xerr.Code, xerr.Msg
	//		}
	//	} else {
	//		xerr := house.PoolChange(p.Uid, models.PoolAdd, -1*value, tx)
	//		if xerr != nil {
	//			xlog.Logger().Errorf("change user vitamin pool add error:%v", err)
	//			mem.Unlock(cli)
	//			return xerr.Code, xerr.Msg
	//		}
	//	}
	//	//if err != nil {
	//	//	xlog.Logger().Errorf("change user vitamin error:%v", err)
	//	//	mem.Unlock(cli)
	//	//	return xerrors.DBExecErrorCode, xerrors.DBExecError.Msg
	//	//}
	//	err = optmem.AddPoolLog(req.UId, value, tx)
	//	if err != nil {
	//		xlog.Logger().Errorf("change user vitamin error:%v", err)
	//		mem.Unlock(cli)
	//		return xerrors.DBExecErrorCode, xerrors.DBExecError.Msg
	//	}
	//}
	// 修改疲劳值统计管理节点信息
	err = GetDBMgr().UpdateVitaminMgrList(house.DBClub.Id, req.UId, aftVitamin, tx)
	if err != nil {
		xlog.Logger().Errorf("UpdateVitaminMgrList error:%v", err)
		mem.Unlock(cli)
		return xerrors.DBExecErrorCode, xerrors.DBExecError.Msg
	}

	if err = tx.Commit().Error; err != nil {
		xlog.Logger().Errorf("commit error:%+v", err)
	}
	suc = true
	mem.Flush()
	mem.Unlock(cli)
	mem.OnMemVitaminOffset()

	// 收支统计(统计群主管理员队长给普通成员加减比赛分,普通成员给绑定队长的比赛分)
	if (optmem.Upper(consts.ROLE_MEMBER) || optmem.Partner == 1) && (mem.URole == consts.ROLE_MEMBER && mem.Partner != 1) {
		GetDBMgr().UpdatePaymentsStatistic(house.DBClub.Id, value)
	} else if (optmem.URole == consts.ROLE_MEMBER && optmem.Partner != 1) && (mem.Partner == 1 || mem.Upper(consts.ROLE_MEMBER)) {
		GetDBMgr().UpdatePaymentsStatistic(house.DBClub.Id, -value)
	}

	// 发送变更通知
	if house.DBClub.IsVitamin {
		ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
		ntf.HId = house.DBClub.HId
		ntf.OptId = p.Uid
		ntf.OptRole = optMemURole
		ntf.UId = mem.UId
		ntf.Value = static.SwitchVitaminToF64(mem.UVitamin)
		if ownerPartner {
			optmem.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
		}
		if house.DBClub.IsPartnerHide {
			// 队长可见
			house.ParnterBroadcast(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
		}
		if house.DBClub.IsVitaminHide {
			// 管理员可见
			house.Broadcast(consts.ROLE_ADMIN, consts.MsgTypeHouseVitaminSet_Ntf, ntf)
		} else {
			// 管理员不可见
			hp := GetPlayerMgr().GetPlayer(house.DBClub.UId)
			if hp != nil {
				hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
			}
		}

		if mem.URole != consts.ROLE_ADMIN && mem.URole != consts.ROLE_CREATER {
			// 通知目标用户
			hp := GetPlayerMgr().GetPlayer(mem.UId)
			if hp != nil {
				hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
			}
		}
		if modByPartner {
			var hp *PlayerCenterMemory
			if mem.Partner == 1 {
				hp = GetPlayerMgr().GetPlayer(mem.Superior)
			} else {
				hp = GetPlayerMgr().GetPlayer(mem.Partner)
			}
			if hp != nil {
				ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
				ntf.HId = house.DBClub.HId
				ntf.OptId = mem.UId
				ntf.OptRole = optmem.URole
				ntf.UId = optmem.UId
				ntf.Value = static.SwitchVitaminToF64(optmem.UVitamin)
				hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
			}
		}
	}

	ack := new(static.Msg_HC_HouseVitaminSet)
	ack.Value = static.SwitchVitaminToF64(aftVitamin)
	return xerrors.SuccessCode, ack
}

// 包厢疲劳值数值日志
func Proto_ClubHouseVitaminSetRecords(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminSetRecords)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}
	if req.Count == 0 {
		req.Count = 10
	}
	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		if cusErr == xerrors.InvalidPermission {
			if !optmem.IsVitaminAdmin() {
				return cusErr.Code, cusErr.Msg
			}
		} else {
			return cusErr.Code, cusErr.Msg
		}
	}

	// 权限 自己可看自己
	if optmem.UId != req.UId {
		if !house.CheckVitaminPermission(optmem) {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}
	}

	mem := house.GetMemByUId(req.UId)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	selectstarttime := static.GetZeroTime(time.Now().AddDate(0, 0, -4))
	selectstartstr := fmt.Sprintf("%d-%02d-%02d", selectstarttime.Year(), selectstarttime.Month(), selectstarttime.Day())
	records := make([]*models.HouseMemberVitaminLog, 0, req.Count-1)
	sql := `select optuid,created_at,befvitamin,type,value,aftvitamin from house_member_vitamin_log 
	where dhid= ? and uid = ? and type in(?,?,?,?,?,?,?,?,?) and date_format(created_at, '%Y-%m-%d') >= ? order by created_at desc limit ? offset ?`
	db := GetDBMgr().GetDBmControl()
	err := db.Raw(sql, house.DBClub.Id, req.UId, models.MemSend, models.AdminSend, models.PartnerSend,
		models.AutoPayPartner, models.ViAdminSet, models.BackPool, models.MemberCost, models.GameReward, models.GameCost,
		selectstartstr, req.Count, req.Start).Scan(&records).Error
	if err != nil {
		xlog.Logger().Errorf("query error:%v", err)
	}

	// var gn []int64
	// if req.Start == 0 {
	// 	err = tx.Table("house_member_vitamin_log").Select("SUM(value)").
	// 		Where("dhid = ?  and uid = ? and status = 0 and type = ?", house.DBClub.Id, req.UId, model.GameCost).Pluck("SUM(value)", &gn).Error

	// 	if err != nil {
	// 		if !strings.Contains(err.Error(), "Scan error") {
	// 			syslog.Logger().Errorf("query error:%v", err)
	// 			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	// 		}
	// 	}
	// }

	ack := new(static.Msg_S2C_HouseVitaminRecords)
	ack.Items = make([]*static.HouseVitaminRecord, 0, len(records)+1)
	dbVitamin, err := mem.GetVitaminFromDb()
	if err == nil {
		if dbVitamin != mem.UVitamin {
			mem.UVitamin = dbVitamin
			mem.Flush()
		}
	}
	ack.CurrentVitamin = static.SwitchVitaminToF64(mem.UVitamin)
	recMem, _ := GetDBMgr().GetDBrControl().GetPerson(req.UId)
	if recMem == nil {
		return xerrors.UserExistError.Code, xerrors.UserExistError.Msg
	}
	ack.UName = recMem.Nickname
	// if req.Start == 0 && len(gn) == 1 {
	// 	item := new(public.HouseVitaminRecord)
	// 	item.UpdatedTime = time.Now().Unix()
	// 	item.OptType = model.GetTypeName(gn[0], model.GameTotal)
	// 	item.OptUName = ""
	// 	item.ChangeVitamin = public.SwitchVitaminToF64(gn[0])
	// 	item.OptTypeInt = int64(model.GameTotal)
	// 	ack.items = append(ack.items, item)
	// }
	for _, record := range records {
		var uname string
		if record.Type == models.GameTotal {
			uname = ""
		} else {
			if record.OptUid == house.DBClub.UId {
				uname = "盟主"
			} else {
				uname = fmt.Sprintf("%d", record.OptUid)
			}
		}
		item := new(static.HouseVitaminRecord)
		item.Id = record.Id
		item.UpdatedTime = record.CreatedAt.Unix()
		if record.BefVitamin == 0 {
			if record.Type == models.BackPool {
				item.BefVitamin = static.SwitchVitaminToF64(record.AftVitamin)
			} else {
				item.BefVitamin = static.SwitchVitaminToF64(record.AftVitamin - record.Value)
			}
		} else {
			item.BefVitamin = static.SwitchVitaminToF64(record.BefVitamin)
		}
		item.OptType = models.GetTypeName(record.Value, record.Type)
		item.OptUName = uname
		item.AftVitamin = static.SwitchVitaminToF64(record.AftVitamin)
		item.ChangeVitamin = static.SwitchVitaminToF64(record.Value)
		item.OptTypeInt = int64(record.Type)
		ack.Items = append(ack.Items, item)
	}
	return xerrors.SuccessCode, ack
}

// 包厢疲劳值数值日志
func Proto_ClubHouseVitaminClear(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminClear)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if mem.URole == consts.ROLE_CREATER {
		custerr := house.MemVitaminClear()
		if custerr != nil {
			return custerr.Code, custerr.Msg
		}
	} else if mem.Partner == 1 {
		custerr := house.PartnerVitaminClear(mem.UId)
		if custerr != nil {
			return custerr.Code, custerr.Msg
		}
	} else {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	msg := fmt.Sprintf("ID %d 进行了一键清零操作", p.Uid)
	CreateClubMassage(house.DBClub.Id, p.Uid, HouseVitaminClear, msg)
	return xerrors.SuccessCode, nil
}

// 比赛统计
func Proto_ClubHousePartnerVitaminStatistic(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HousePartnerVitaminStatistic)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	// 获取包厢
	clubHouse, _, viewMem, xer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xer != xerrors.RespOk {
		return xer.Code, xer.Msg
	}

	if viewMem.URole != consts.ROLE_CREATER && viewMem.URole != consts.ROLE_ADMIN &&
		!viewMem.IsVitaminAdmin() && !viewMem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 队长想查看圈主视角，默认切换为队长视角
	if req.Partner == 0 && viewMem.IsPartner() {
		req.Partner = viewMem.UId
	}

	zeroTime := static.GetZeroTime(time.Now().AddDate(0, 0, req.SelectTime))
	selectStartTime := zeroTime
	selectEndTime := zeroTime.Add(24 * time.Hour)

	allPartnerAttr, err := GetDBMgr().SelectHouseAllPartnerAttr(clubHouse.DBClub.Id)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 查询疲劳值每日统计
	vitaminDayRecords, err := GetDBMgr().SelectPartnerVitaminStatistic(clubHouse.DBClub.Id, -1, selectStartTime, selectEndTime, 0)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	vitaminRecordsMap := make(map[int64]models.RecordVitaminDay)
	for _, item := range vitaminDayRecords {
		record, ok := vitaminRecordsMap[item.UId]
		if !ok {
			record = models.RecordVitaminDay{}
		}
		record.VitaminCostBW += item.VitaminCostBW
		record.VitaminCost += item.VitaminCost
		record.VitaminCostRound += item.VitaminCostRound
		record.VitaminWinLose += item.VitaminWinLose
		vitaminRecordsMap[item.UId] = record
	}

	memberMap := clubHouse.GetMemberMap(false)
	GenMsgVitaminStatisticItemFn := func(clb *Club, mem *HouseMember, memMap map[int64]HouseMember,
		recordMap map[int64]models.RecordVitaminDay, pt int) *static.MsgVitaminStatisticItem {
		item := &static.MsgVitaminStatisticItem{
			UId:         mem.UId,
			UName:       mem.NickName,
			UUrl:        mem.ImgUrl,
			UGender:     mem.Sex,
			URole:       mem.URole,
			UVitamin:    static.SwitchVitaminToF64(mem.UVitamin),
			Partner:     mem.Partner,
			PartnerType: pt,
			Superior:    mem.Superior,
			AlarmValue:  -1,
			Exp:         0,
			PlayerNum:   1,
		}
		if mem.UVitamin < 0 {
			item.VitaminMinusInt += mem.UVitamin
		} else {
			item.VitaminLeftInt += mem.UVitamin
		}
		item.VitaminCostInt += recordMap[item.UId].VitaminCost
		item.VitaminCostRoundInt += recordMap[item.UId].VitaminCostRound
		item.VitaminCostBWInt += recordMap[item.UId].VitaminCostBW
		item.VitaminWinLoseInt += recordMap[item.UId].VitaminWinLose
		if pt > 0 { // 不是查看明细，查看明细不需要下面这些数据
			if item.Partner == 1 || item.UId == clb.DBClub.UId { // 如果是队长或圈主
				attr, ok := allPartnerAttr[item.UId]
				if ok {
					if attr.AlarmValue >= 0 {
						item.AlarmValue = static.SwitchVitaminToF64(attr.AlarmValue)
					}
					item.Exp = attr.Exp
				}
				var all []int64

				//if pt == 1 {
				//	all = clb.GetUIDsByPartner(memMap, item.UId)
				//	xlog.Logger().Infof("玩家id:%d,直属玩家:%+v", item.UId, all)
				//} else if pt == 2 {
				//	js, ms := clb.GetAllMemByPartner(memMap, item.UId)
				//	all = append(js, ms...)
				//	xlog.Logger().Infof("玩家id:%d,整条线玩家:%+v", item.UId, all)
				//}

				js, ms := clb.GetAllMemByPartner(memMap, item.UId)
				all = append(js, ms...)
				xlog.Logger().Infof("玩家id:%d,整条线玩家:%+v", item.UId, all)

				for _, m := range memMap {
					if static.In64(all, m.UId) {
						item.PlayerNum++
						if m.UVitamin < 0 {
							item.VitaminMinusInt += m.UVitamin
						} else {
							item.VitaminLeftInt += m.UVitamin
						}
						item.VitaminCostInt += recordMap[m.UId].VitaminCost
						item.VitaminCostRoundInt += recordMap[m.UId].VitaminCostRound
						item.VitaminCostBWInt += recordMap[m.UId].VitaminCostBW
						item.VitaminWinLoseInt += recordMap[m.UId].VitaminWinLose
					}
				}
			}
		}
		return item
	}
	superiorList := make([]int64, 0)
	for _, mem := range memberMap {
		if mem.IsJunior() && !static.In64(superiorList, mem.Superior) {
			superiorList = append(superiorList, mem.Superior)
		}
	}
	vitaminStatisticMap := make(map[int64]*static.MsgVitaminStatisticItem)
	if req.Partner == 0 {
		// 圈主视角
		if req.ShowType == 0 {
			// 展示圈主及圈主名下大队长(1级队长，组长)，需要计算名下直属成员数量
			for _, mem := range memberMap {
				if mem.UId == clubHouse.DBClub.UId {
					// 可查看详情
					statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 1)
					vitaminStatisticMap[mem.UId] = statisticItem
				} else if mem.IsPartner() && mem.Superior == 0 {
					// 可查看下级
					statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 2)
					vitaminStatisticMap[mem.UId] = statisticItem
				}
			}
		} else {
			// 展示圈主及圈主名下成员
			for _, mem := range memberMap {
				if mem.UId == clubHouse.DBClub.UId || mem.Partner == 0 {
					// 详情里面的 不能再往下看了
					statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 0)
					vitaminStatisticMap[mem.UId] = statisticItem
				}
			}
		}
	} else {
		partnerOrOwner, ok := memberMap[req.Partner]
		if !ok {
			xlog.Logger().Error("partner not exists", req.Partner)
			return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
		}
		var partnerId int64
		if partnerOrOwner.IsPartner() {
			partnerId = partnerOrOwner.UId
		} else {
			if partnerOrOwner.UId != clubHouse.DBClub.UId {
				xlog.Logger().Error("not a partner", req.Partner)
				return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
			}
		}
		// 队长视角
		if req.ShowType == 0 {
			// 队长/圈主及名下小队长
			for _, mem := range memberMap {
				if mem.UId == partnerOrOwner.UId {
					// 如果是自己的数据，可查看详情
					statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 1)
					vitaminStatisticMap[mem.UId] = statisticItem
				} else if mem.IsPartner() && mem.Superior == partnerId {
					if partnerId == 0 { // 圈主
						// 圈主视角，这里是组长，可查看下级
						statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 2)
						vitaminStatisticMap[mem.UId] = statisticItem
					} else {
						if static.In64(superiorList, mem.UId) {
							// 队长视角，这里都是大队长，可查看下级
							statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 2)
							vitaminStatisticMap[mem.UId] = statisticItem
						} else {
							// 队长视角，这里都是小队长，可查看明细
							statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 1)
							vitaminStatisticMap[mem.UId] = statisticItem
						}
					}
				}
			}
		} else {
			// 队长/圈主 查看明细
			for _, mem := range memberMap {
				// 自己或者自己名下的人都算进来。partnerId=0 是圈主
				if mem.UId == partnerOrOwner.UId || mem.Partner == partnerId {
					statisticItem := GenMsgVitaminStatisticItemFn(clubHouse, &mem, memberMap, vitaminRecordsMap, 0)
					vitaminStatisticMap[mem.UId] = statisticItem
				}
			}
		}
	}

	// 整体，统计最终结果
	tmpTotalVitaminCost := int64(0)
	tmpTotalVitaminCostRound := int64(0)
	tmpTotalVitaminCostBW := int64(0)
	tmpTotalVitaminLeft := int64(0)
	tmpTotalVitaminMinus := int64(0)
	tmpTotalVitaminWinLose := int64(0)

	var conItems []static.MsgVitaminStatisticItem
	for _, item := range vitaminStatisticMap {
		item.VitaminLeft = static.SwitchVitaminToF64(item.VitaminLeftInt)
		item.VitaminMinus = static.SwitchVitaminToF64(item.VitaminMinusInt)
		item.VitaminCost = static.SwitchVitaminToF64(item.VitaminCostInt)
		item.VitaminCostRound = static.SwitchVitaminToF64(item.VitaminCostRoundInt)
		item.VitaminCostBW = static.SwitchVitaminToF64(item.VitaminCostBWInt)
		item.VitaminWinLose = static.SwitchVitaminToF64(item.VitaminWinLoseInt)

		conItems = append(conItems, *item)

		tmpTotalVitaminCost += item.VitaminCostInt
		tmpTotalVitaminCostRound += item.VitaminCostRoundInt
		tmpTotalVitaminCostBW += item.VitaminCostBWInt
		tmpTotalVitaminLeft += item.VitaminLeftInt
		tmpTotalVitaminMinus += item.VitaminMinusInt
		tmpTotalVitaminWinLose += item.VitaminWinLoseInt
	}
	var searchItems []static.MsgVitaminStatisticItem
	// 搜索功能
	if req.SearchKey != "" {
		for _, sItem := range conItems {
			if strings.Contains(sItem.UName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", sItem.UId), req.SearchKey) {
				searchItems = append(searchItems, sItem)
			}
		}
	} else {
		searchItems = append(searchItems, (conItems)[0:]...)
	}

	switch req.SortType {
	case static.SORT_MEMROLE_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.URole == b.URole {
				if a.Partner == 1 && b.Partner == 1 {
					if a.Superior == 0 && b.Superior == 0 {
						if a.PartnerType == b.PartnerType {
							return a.UId > b.UId
						} else {
							return a.PartnerType > b.PartnerType
						}
					} else {
						if a.Superior == 0 {
							return true
						} else if b.Superior == 0 {
							return false
						} else {
							return a.UId > b.UId
						}
					}
				} else {
					if a.Partner == 1 {
						return true
					} else if b.Partner == 1 {
						return false
					} else {
						return a.UId > b.UId
					}
				}
			}
			return a.URole < b.URole
		}})
	case static.SORT_VITAMIN_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.UVitamin == b.UVitamin {
				return a.UId > b.UId
			}
			return a.UVitamin > b.UVitamin
		}})
	case static.SORT_VITAMIN_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.UVitamin == b.UVitamin {
				return a.UId < b.UId
			}
			return a.UVitamin < b.UVitamin
		}})
	case static.SORT_VITAMINMINUS_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminMinusInt == b.VitaminMinusInt {
				return a.UId > b.UId
			}
			return a.VitaminMinusInt > b.VitaminMinusInt
		}})
	case static.SORT_VITAMINMINUS_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminMinusInt == b.VitaminMinusInt {
				return a.UId < b.UId
			}
			return a.VitaminMinusInt < b.VitaminMinusInt
		}})
	case static.SORT_VITAMINLEFT_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminLeftInt == b.VitaminLeftInt {
				return a.UId > b.UId
			}
			return a.VitaminLeftInt > b.VitaminLeftInt
		}})
	case static.SORT_VITAMINLEFT_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminLeftInt == b.VitaminLeftInt {
				return a.UId < b.UId
			}
			return a.VitaminLeftInt < b.VitaminLeftInt
		}})
	case static.SORT_ALARMVALUE_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.AlarmValue == b.AlarmValue {
				return a.UId > b.UId
			}
			return a.AlarmValue > b.AlarmValue
		}})
	case static.SORT_ALARMVALUE_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.AlarmValue == b.AlarmValue {
				return a.UId < b.UId
			}
			return a.AlarmValue < b.AlarmValue
		}})
	case static.SORT_PLAYERNUM_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.PlayerNum == b.PlayerNum {
				return a.UId > b.UId
			}
			return a.PlayerNum > b.PlayerNum
		}})
	case static.SORT_PLAYERNUM_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.PlayerNum == b.PlayerNum {
				return a.UId < b.UId
			}
			return a.PlayerNum < b.PlayerNum
		}})
	case static.SORT_VITAMINCOST_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminCostInt == b.VitaminCostInt {
				return a.UId > b.UId
			}
			return a.VitaminCostInt > b.VitaminCostInt
		}})
	case static.SORT_VITAMINCOST_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminCostInt == b.VitaminCostInt {
				return a.UId < b.UId
			}
			return a.VitaminCostInt < b.VitaminCostInt
		}})
	case static.SORT_VITAMINWINLOSEP_DESC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminWinLoseInt == b.VitaminWinLoseInt {
				return a.UId > b.UId
			}
			return a.VitaminWinLoseInt > b.VitaminWinLoseInt
		}})
	case static.SORT_VITAMINWINLOSEP_ASC:
		sort.Sort(static.MsgVitaminStatisticItemWrapper{Item: searchItems, By: func(a, b *static.MsgVitaminStatisticItem) bool {
			if a.VitaminWinLoseInt == b.VitaminWinLoseInt {
				return a.UId < b.UId
			}
			return a.VitaminWinLoseInt < b.VitaminWinLoseInt
		}})
	}

	var ack static.Msg_HC_HouseVitaminStatistic
	ack.TotalVitaminCost = static.SwitchVitaminToF64(tmpTotalVitaminCost)
	ack.TotalVitaminCostRound = static.SwitchVitaminToF64(tmpTotalVitaminCostRound)
	ack.TotalVitaminCostBW = static.SwitchVitaminToF64(tmpTotalVitaminCostBW)
	ack.TotalVitaminLeft = static.SwitchVitaminToF64(tmpTotalVitaminLeft)
	ack.TotalVitaminMinus = static.SwitchVitaminToF64(tmpTotalVitaminMinus)
	ack.TotalVitaminWinLose = static.SwitchVitaminToF64(tmpTotalVitaminWinLose)

	ack.Items = make([]static.MsgVitaminStatisticItem, 0)
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd

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

// 包厢疲劳值统计
func Proto_ClubHouseVitaminStatistic(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminStatistic)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	selectStartTime := time.Now()
	selectEndTime := time.Now()
	if req.SelectTime <= 0 {
		selectStartTime = selectStartTime.AddDate(0, 0, req.SelectTime)
		selectEndTime = selectEndTime.AddDate(0, 0, req.SelectTime)
	} else if req.SelectTime == 3 {
		// 七天数据
		selectStartTime = selectStartTime.AddDate(0, 0, -7)
		selectEndTime = selectEndTime.AddDate(0, 0, -1)
	}

	var err error
	VitaminLeft := int64(0)
	VitaminMinus := int64(0)

	// 查询正在统计中的数据
	if req.SelectTime == 0 {
		// 计算权力所有人剩余疲劳值
		VitaminLeft, VitaminMinus = house.CalculateLeftVitamin()
		_, err := GetDBMgr().SelectVitaminStatisticClearRecording(house.DBClub.Id, selectStartTime, VitaminLeft, VitaminMinus)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else {
		// VitaminLeft, VitaminMinus, err = GetDBMgr().SelectHouseDayLeftVitamin(house.DBClub.Id, req.SelectTime)
		// if err != nil {
		//	return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		// }
	}

	vitamindayclearrecords, err := GetDBMgr().SelectVitaminStatisticClear(house.DBClub.Id, selectStartTime, selectEndTime)

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseVitaminStatisticClear
	ack.Items = make([]static.MsgVitaminStatisticClearItem, 0)

	for i := 0; i < len(vitamindayclearrecords); i++ {
		record := vitamindayclearrecords[i]

		item := new(static.MsgVitaminStatisticClearItem)
		item.VitaminCost = static.SwitchVitaminToF64(record.VitaminCost)
		item.VitaminCostRound = static.SwitchVitaminToF64(record.VitaminCostRound)
		item.VitaminCostBW = static.SwitchVitaminToF64(record.VitaminCostBW)
		item.VitaminPayment = static.SwitchVitaminToF64(record.VitaminPayment)
		item.BeginAt = record.BeginAt.Unix()
		item.EndAt = record.EndAt.Unix()
		if req.SelectTime == 0 && record.Recording == 1 {
			item.EndAt = 0
		}
		if record.Recording == 1 {
			item.VitaminLeft = static.SwitchVitaminToF64(VitaminLeft)
			item.VitaminMinus = static.SwitchVitaminToF64(VitaminMinus)
		} else {
			item.VitaminLeft = static.SwitchVitaminToF64(record.VitaminLeft)
			item.VitaminMinus = static.SwitchVitaminToF64(record.VitaminMinus)
		}

		ack.Items = append(ack.Items, *item)
	}

	return xerrors.SuccessCode, ack
}

// 包厢疲劳值统计清零
func Proto_ClubHouseVitaminStatisticClear(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminStatisticClear)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	// 计算权力所有人剩余疲劳值
	VitaminLeft, VitaminMinus := house.CalculateLeftVitamin()

	// 查询正在统计中的数据
	housevitaminrecordclear, err := GetDBMgr().SelectVitaminStatisticClearRecording(house.DBClub.Id, time.Now(), VitaminLeft, VitaminMinus)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 结扎以前数据
	err = GetDBMgr().UpdateVitaminStatisticClear(house.DBClub.Id, time.Now(), VitaminLeft, VitaminMinus)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 生成新的结扎数据
	err = GetDBMgr().InsertVitaminStatisticClear(house.DBClub.Id, housevitaminrecordclear)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	vitamindayclearrecords, err := GetDBMgr().SelectVitaminStatisticClear(house.DBClub.Id, time.Now(), time.Now())

	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseVitaminStatisticClear
	ack.Items = make([]static.MsgVitaminStatisticClearItem, 0)

	for i := 0; i < len(vitamindayclearrecords); i++ {
		record := vitamindayclearrecords[i]

		item := new(static.MsgVitaminStatisticClearItem)
		item.VitaminCost = static.SwitchVitaminToF64(record.VitaminCost)
		item.VitaminCostRound = static.SwitchVitaminToF64(record.VitaminCostRound)
		item.VitaminCostBW = static.SwitchVitaminToF64(record.VitaminCostBW)
		item.VitaminPayment = static.SwitchVitaminToF64(record.VitaminPayment)

		item.BeginAt = record.BeginAt.Unix()
		item.EndAt = record.EndAt.Unix()
		if record.Recording == 1 {
			item.EndAt = 0
		}

		if record.Recording == 1 {
			item.VitaminLeft = static.SwitchVitaminToF64(VitaminLeft)
			item.VitaminMinus = static.SwitchVitaminToF64(VitaminMinus)
		} else {
			item.VitaminLeft = static.SwitchVitaminToF64(record.VitaminLeft)
			item.VitaminMinus = static.SwitchVitaminToF64(record.VitaminMinus)
		}

		ack.Items = append(ack.Items, *item)
	}

	return xerrors.SuccessCode, ack
}

// 包厢疲劳值管理统计
func Proto_ClubHouseVitaminMgrList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminMgrList)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if time.Now().Unix() > house.lastUpdatetime+60 {
		vitamindayrecord, err := GetDBMgr().SelectPartnerVitaminStatisticMgr(house.DBClub.Id, -1, time.Now(), time.Now(), "")
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		house.historyData = vitamindayrecord
		house.lastUpdatetime = time.Now().Unix()
	}
	vitamindayrecords := house.historyData

	conItemsMap := make(map[int64]*static.MsgVitaminMgrItem)
	for _, dayRecord := range vitamindayrecords {
		dayRecordItem, ok := conItemsMap[dayRecord.UId]
		if ok {
			dayRecordItem.VitaminWinLoseCostInt += dayRecord.VitaminWinLose
			dayRecordItem.VitaminPlayCostInt += dayRecord.VitaminCost
			dayRecordItem.VitaminCostRoundInt += dayRecord.VitaminCostRound
			dayRecordItem.VitaminCostBWInt += dayRecord.VitaminCostBW
		} else {
			dayRecordItem = new(static.MsgVitaminMgrItem)
			dayRecordItem.UId = dayRecord.UId
			dayRecordItem.VitaminWinLoseCostInt = dayRecord.VitaminWinLose
			dayRecordItem.VitaminPlayCostInt = dayRecord.VitaminCost
			dayRecordItem.VitaminCostRoundInt = dayRecord.VitaminCostRound
			dayRecordItem.VitaminCostBWInt = dayRecord.VitaminCostBW
		}
		conItemsMap[dayRecord.UId] = dayRecordItem
	}

	var partnerId int64
	pmem := house.GetMemByUId(p.Uid)
	if pmem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	if pmem.URole >= consts.ROLE_MEMBER {
		if pmem.IsPartner() {
			partnerId = pmem.UId
		} else if pmem.IsVicePartner() {
			partnerId = pmem.Partner
		} else if pmem.IsVitaminAdmin() {
			partnerId = 0
		} else {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}
	} else {
		partnerId = 0
	}
	// 筛选时间的起点和终点
	// 只查今天
	zeroTime := static.GetZeroTime(time.Now().AddDate(0, 0, 0))
	selectTime1 := zeroTime
	selectTime2 := zeroTime.Add(24 * time.Hour)

	// 新增
	var selStsErr error
	statisticsItems := make(map[int64]models.QueryMemberStatisticsResult)
	if partnerId > 0 {
		// 查询合伙人名下的成员战绩记录
		statisticsItems, selStsErr = GetDBMgr().SelectPartnerMemberStatisticsWithTotal(partnerId, house.DBClub.Id, -1, selectTime1, selectTime2)
	} else {
		// 默认查询统计信息
		statisticsItems, selStsErr = GetDBMgr().SelectHouseMemberStatisticsWithTotal(house.DBClub.Id, -1, selectTime1, selectTime2)
	}
	if selStsErr != nil {
		xlog.Logger().Error(selStsErr.Error())
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	for uid, conItem := range conItemsMap {
		item, ok := statisticsItems[uid]
		if ok {
			conItem.PlayTimes = item.PlayTimes
			conItem.BwTimes = item.BigWinTimes
			conItem.TotalScore = static.HF_DecimalDivide(item.TotalScore, 1, 2)
			conItem.ValidTimes = item.ValidTimes
			conItem.InValidTimes = item.PlayTimes - item.ValidTimes
			conItemsMap[uid] = conItem
		}
	}

	var conItems []*static.MsgVitaminMgrItem

	inhpmems := house.GetMemSimple(false)
	for _, hpmem := range inhpmems {
		if partnerId > 0 && hpmem.UId != p.Uid && hpmem.UId != partnerId && hpmem.Partner != partnerId && hpmem.Superior != partnerId {
			continue
		}

		if req.SearchKey != "" {
			if !(strings.Contains(hpmem.NickName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", hpmem.UId), req.SearchKey)) {
				continue
			}
		}

		staticItem := new(static.MsgVitaminMgrItem)

		staticItem.UId = hpmem.UId
		staticItem.UName = hpmem.NickName
		staticItem.UUrl = hpmem.ImgUrl
		staticItem.UGender = hpmem.Sex
		staticItem.URole = hpmem.URole
		staticItem.IsJunior = hpmem.Superior == p.Uid
		staticItem.IsPartner = hpmem.Partner == 1
		staticItem.VitaminAdmin = hpmem.VitaminAdmin
		staticItem.VicePartner = hpmem.VicePartner

		if hpmem.IsJunior() {
			staticItem.UPartner = hpmem.Superior
		} else {
			// 如果非一级合伙人 则upartner标记为1 表示一级合伙人
			staticItem.UPartner = hpmem.Partner
		}

		staticItem.CurVitamin = static.SwitchVitaminToF64(hpmem.UVitamin)

		record, ok := conItemsMap[hpmem.UId]
		if ok {
			staticItem.PreNodeVitamin = 0
			staticItem.VitaminPlayCost = static.SwitchVitaminToF64(record.VitaminPlayCostInt)
			staticItem.VitaminWinLoseCost = static.SwitchVitaminToF64(record.VitaminWinLoseCostInt)
			staticItem.VitaminCostBW = static.SwitchVitaminToF64(record.VitaminCostBWInt)
			staticItem.VitaminCostRound = static.SwitchVitaminToF64(record.VitaminCostRoundInt)
			staticItem.PlayTimes = record.PlayTimes
			staticItem.BwTimes = record.BwTimes
			staticItem.TotalScore = record.TotalScore
			staticItem.ValidTimes = record.ValidTimes
			staticItem.InValidTimes = record.InValidTimes
		} else {
			staticItem.PreNodeVitamin = 0
			staticItem.VitaminPlayCost = 0
			staticItem.VitaminWinLoseCost = 0
			staticItem.VitaminCostRound = 0
			staticItem.VitaminCostBW = 0
			staticItem.PlayTimes = 0
			staticItem.BwTimes = 0
			staticItem.TotalScore = 0
			staticItem.ValidTimes = 0
			staticItem.InValidTimes = 0
		}

		conItems = append(conItems, staticItem)
	}

	conItems = SortVitaminMgrItems(req.SortType, conItems)

	var ack static.Msg_HC_HouseVitaminMgrList

	ack.Items = make([]*static.MsgVitaminMgrItem, 0)
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd
	ack.UPartner = pmem.Partner

	var idxBeg = req.PBegin
	var idxEnd = req.PEnd

	if idxEnd > len(conItems) {
		idxEnd = len(conItems)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}
	ack.Items = append(ack.Items, conItems[idxBeg:idxEnd]...)
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHousePartnerAddList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req := data.(*static.Msg_CH_HouseMemList)

	if req.PBegin < 0 {
		req.PBegin = 0
	}
	// 不包含end位，需要多取一位
	req.PEnd++
	if req.PEnd < req.PBegin {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	h, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	arr := h.GetMemSimple(false)
	unParterList := make([]HouseMember, 0, len(arr)/3)
	for _, mem := range arr {
		// 空串
		if mem.URole <= 1 || mem.Partner != 0 || mem.IsVitaminAdmin() || mem.IsVicePartner() {
			continue
		}
		if req.Param == "" {
			unParterList = append(unParterList, mem)
			continue
		}
		// ID 包含
		if strings.Contains(fmt.Sprintf("%d", mem.UId), req.Param) {
			unParterList = append(unParterList, mem)
			continue
		}

		if strings.Contains(mem.NickName, req.Param) {
			unParterList = append(unParterList, mem)
			continue
		}
	}

	var ack static.Msg_HC_HouseMemList
	ack.Totalnum = len(unParterList)
	ack.FMems = make([]*static.Msg_HouseMemberItem, 0, len(unParterList))

	// 分页超出范围
	if len(unParterList) == 0 || len(unParterList) < req.PBegin {
		return xerrors.SuccessCode, &ack
	}

	// 开始索引
	var idxBeg int
	var idxEnd int
	idxBeg = req.PBegin
	if req.PEnd > len(unParterList) {
		idxEnd = len(unParterList)
	} else {
		idxEnd = req.PEnd
	}
	unParterList = unParterList[idxBeg:idxEnd]

	for _, mem := range unParterList {

		var titem static.Msg_HouseMemberItem
		titem.UId = mem.UId
		titem.UOnline = mem.IsOnline
		titem.UName = mem.NickName
		titem.URole = mem.URole
		titem.UVitamin = static.SwitchVitaminToF64(mem.UVitamin)
		titem.UPartner = mem.Partner
		titem.URemark = mem.URemark
		titem.UUrl = mem.ImgUrl
		titem.UGender = mem.Sex
		titem.UJoinTime = mem.ApplyTime
		ack.FMems = append(ack.FMems, &titem)
	}

	return xerrors.SuccessCode, &ack
}

// 包厢获取有效对局积分
func Proto_ClubHouseValidRoundScoreGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseValidRoundScoreGet)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 队长身份 查询队长设置的低分局数据
	mem := house.GetMemByUId(p.Uid)
	if mem.IsPartner() {
		return Proto_GetPartnerLowScoreVal(s, p, data)
	}

	// house := GetClubMgr().GetClubHouseByHId(int(req.HId))
	// if house == nil {
	//	return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	// }
	//
	// hmem := house.GetMemByUId(p.Uid)
	// if hmem == nil {
	//	return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	// }
	// if hmem.URole != constant.ROLE_CREATER && !hmem.IsVitaminAdmin() {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	// }

	invalidTime, err := GetDBMgr().SelectInvalidRoundByDate(house.DBClub.Id, time.Now().AddDate(0, 0, -1))
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	paySlice := make(models.HouseFloorGearPaySlice, 0)
	err = GetDBMgr().GetDBmControl().Where("hid = ?", house.DBClub.Id).Find(&paySlice).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	floorIds, _ := house.GetFloors()
	var ack static.Msg_HC_HouseValidRoundScoreGet
	ack.HId = house.DBClub.HId
	ack.Items = make([]static.HouseValidRoundScoreGetItem, 0)
	ack.InvalidRound = invalidTime
	for i := 0; i < len(floorIds); i++ {
		fid := floorIds[i]
		payInfo := paySlice.FindByFid(fid)
		if payInfo == nil {
			floor := house.GetFloorByFId(fid)
			if floor == nil {
				return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
			}
			payInfo = models.NewHouseFloorGearPay(house.DBClub.Id, fid, floor.Rule.PlayerNum)
		}
		var item static.HouseValidRoundScoreGetItem
		// item.HId = house.DBClub.HId
		item.FId = fid
		if !payInfo.ConfiguredBase() {
			item.Score = models.IgnorePay
		} else {
			if payInfo.Gear2Under == models.InvalidPay {
				item.Score = models.InvalidPay
			} else {
				item.Score = static.SwitchVitaminToF64(payInfo.Gear2Under)
			}
		}

		ack.Items = append(ack.Items, item)
	}
	return xerrors.SuccessCode, &ack

	// housefloorscores, err := GetDBMgr().SelectHouseFloorValidRound(house.DBClub.Id)
	// if err != nil {
	//	return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	// }
	//
	// allhousesetscore := 0
	// if len(housefloorscores) == 0 {
	//	// 没有包厢楼层有效对局设置,查询包厢整个的设置
	//	housescore, err := GetDBMgr().SelectHouseValidRound(house.DBClub.Id)
	//	if err != nil {
	//		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	//	}
	//
	//	if len(housescore) > 0 {
	//		allhousesetscore = housescore[0].ValidMinScore
	//	} else {
	//		allhousesetscore = 0
	//	}
	// }
	//
	// var ack public.Msg_HC_HouseValidRoundScoreGet
	// ack.Items = make([]public.HouseValidRoundScoreGetItem, 0)
	// ack.InvalidRound = invalidTime
	// floorids, _ := house.GetFloors()
	// for i := 0; i < len(floorids); i++ {
	//	sitem := new(public.HouseValidRoundScoreGetItem)
	//	sitem.HId = house.DBClub.Id
	//	sitem.FId = floorids[i]
	//	sitem.Score = 0
	//	sitem.BigScore = -1
	//
	//	if len(housefloorscores) == 0 {
	//		sitem.Score = allhousesetscore
	//		sitem.BigScore = -1
	//	} else {
	//		for _, housefloorscore := range housefloorscores {
	//			if housefloorscore.FId == sitem.FId {
	//				sitem.Score = housefloorscore.ValidMinScore
	//				sitem.BigScore = housefloorscore.ValidBigScore
	//				break
	//			}
	//		}
	//	}
	//
	//	ack.Items = append(ack.Items, *sitem)
	// }
	// return xerrors.SuccessCode, &ack
}

// 包厢设置有效对局积分
func Proto_ClubHouseValidRoundScoreSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_HouseValidRoundScoreSet)

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorSetLowScore)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 队长设置低分局 执行Proto_PartnerSetLowScoreVal业务逻辑
	mem := house.GetMemByUId(p.Uid)
	if mem.IsPartner() {
		return Proto_PartnerSetLowScoreVal(s, p, data)
	}

	paySlice := make(models.HouseFloorGearPaySlice, 0)
	err := GetDBMgr().GetDBmControl().Where("hid = ?", house.DBClub.Id).Find(&paySlice).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	tx := GetDBMgr().GetDBmControl().Begin()
	for i := 0; i < len(req.Items); i++ {
		item := req.Items[i]
		if item.Score <= models.InvalidPay {
			continue
		}
		payInfo := paySlice.FindByFid(item.FId)
		var create bool
		if payInfo == nil {
			floor := house.GetFloorByFId(item.FId)
			if floor == nil {
				tx.Rollback()
				return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
			}
			create = true
			payInfo = models.NewHouseFloorGearPay(house.DBClub.Id, item.FId, floor.Rule.PlayerNum)
		}
		log := payInfo.GenLog()
		var flag bool
		if newG2u := static.SwitchF64ToVitamin(item.Score); newG2u != payInfo.Gear2Under {
			flag = true
			g2u := payInfo.Gear2Under
			payInfo.Gear2 = true
			payInfo.Gear2Under = newG2u
			if g2u >= 0 && !payInfo.AAPay { //如果原先配置了,且是大赢家支付，影响挡位数值计算
				if g3u := payInfo.Gear3Under; g3u >= 0 && g3u <= payInfo.Gear2Under {
					// 为保证受影响范围尽量小，超过约束 自加+1 来建立约束
					payInfo.Gear3Under = payInfo.Gear2Under + static.SwitchIntVitamin(1)
					if g4u := payInfo.Gear4Under; g4u >= 0 && g4u <= payInfo.Gear3Under {
						payInfo.Gear4Under = payInfo.Gear3Under + static.SwitchIntVitamin(1)
						if g5u := payInfo.Gear5Under; g5u >= 0 && g5u <= payInfo.Gear4Under {
							payInfo.Gear5Under = payInfo.Gear4Under + static.SwitchIntVitamin(1)
							if g6u := payInfo.Gear6Under; g6u >= 0 && g6u <= payInfo.Gear5Under {
								payInfo.Gear6Under = payInfo.Gear5Under + static.SwitchIntVitamin(1)
							}
						}
					}
				}
			}
		}
		if flag {
			// err = payInfo.Check()
			// if err != nil {
			// 	tx.Rollback()
			// 	return xerrors.ResultErrorCode, fmt.Sprintf("%d楼:%s", house.GetFloorIndexByFid(item.FId)+1, err.Error())
			// }

			// 关闭比赛场
			// payInfo.Gear1Cost = -2

			if create {
				err = tx.Create(payInfo).Error
			} else {
				err = tx.Save(payInfo).Error
			}
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
			if !create {
				err = tx.Create(log).Error
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
			}
		}
	}
	if err = tx.Commit().Error; err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, nil
}

func ClubHouseOptPMemberVitamin() *xerrors.XError {
	return nil
}

func Proto_ClubHouseVitaminSend(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseVitaminSet)
	if !ok || req.Value <= 0 {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if req.UId == p.Uid {
		return xerrors.SendSelfVitaminCode, xerrors.SendSelfVitaminError.Msg
	}
	house, _, sender, cuserr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cuserr != xerrors.RespOk {
		return cuserr.Code, cuserr.Msg
	}
	optRole := sender.URole
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if house.IsBusyMerging() {
		return xerrors.HouseBusyError.Code, xerrors.HouseBusyError.Msg
	}

	toMem := house.GetMemByUId(req.UId)
	if toMem == nil || toMem.URole >= consts.ROLE_APLLY {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	// 调整下级队长比赛分判断
	if sender.IsPartner() {
		if toMem.Superior == sender.UId {
			if house.DBClub.DisVitaminJunior {
				return xerrors.HousePartnerSetVitaminDisableError.Code, xerrors.HousePartnerSetVitaminDisableError.Msg
			}
		}
	}

	if sender.IsVicePartner() {
		if toMem.Superior == sender.Partner {
			if house.DBClub.DisVitaminJunior {
				return xerrors.HousePartnerSetVitaminDisableError.Code, xerrors.HousePartnerSetVitaminDisableError.Msg
			}
		}
	}

	var onwerPartner bool
	// 如果是向一个队长名下的玩家赠送则由该队长来承担
	if toMem.Partner > 1 {
		if p.Uid != toMem.Partner {
			if house.DBClub.NoSkipVitaminSet {
				return xerrors.ResultErrorCode, "请先打开比赛场-跨级可调开关。"
			} else {
				sender = house.GetMemByUId(toMem.Partner)
				if sender == nil {
					xlog.Logger().Error("sender is nil", toMem.Partner)
					return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
				}
				onwerPartner = true
			}
		}
	}

	pSender, err := GetDBMgr().GetDBrControl().GetPerson(sender.UId)
	if pSender == nil || err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if pSender.TableId > 0 {
		return xerrors.UserInGameDescVitaminError.Code, xerrors.UserInGameDescVitaminError.Msg
	}

	cli := GetDBMgr().Redis
	// 发送者
	sender.Lock(cli)
	defer sender.Unlock(cli)
	tx := GetDBMgr().GetDBmControl().Begin()
	var suc bool
	defer func() {
		if !suc {
			tx.Rollback()
		}
	}()
	svitamin, err := sender.GetVitaminFromDbWithLock(tx)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	sender.UVitamin = svitamin
	// 接收者
	toMem.Lock(cli)
	defer toMem.Unlock(cli)
	tvitamin, err := toMem.GetVitaminFromDbWithLock(tx)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	toMem.UVitamin = tvitamin

	value := static.SwitchF64ToVitamin(req.Value)
	if sender.UVitamin < value {
		if onwerPartner {
			return xerrors.VitaminSendNotEnoughError.Code, "玩家队长比赛分不足，操作失败。"
		} else {
			return xerrors.VitaminSendNotEnoughError.Code, xerrors.VitaminSendNotEnoughError.Msg
		}
	}
	_, sendaftVitamin, err := sender.VitaminIncrement(req.UId, -1*value, models.MemSend, tx)
	if err != nil {
		xlog.Logger().Errorf("数据库操作失败:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	_, toaftVitamin, err := toMem.VitaminIncrement(p.Uid, value, models.MemSend, tx)
	if err != nil {
		xlog.Logger().Errorf("数据库操作失败:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	// 修改疲劳值统计管理节点信息
	err = GetDBMgr().UpdateVitaminMgrList(house.DBClub.Id, sender.UId, sendaftVitamin, tx)
	if err != nil {
		xlog.Logger().Errorf("数据库操作失败:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 修改疲劳值统计管理节点信息
	err = GetDBMgr().UpdateVitaminMgrList(house.DBClub.Id, req.UId, toaftVitamin, tx)
	if err != nil {
		xlog.Logger().Errorf("数据库操作失败:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Errorf("commit error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	suc = true
	sender.Flush()
	toMem.Flush()

	toMem.OnMemVitaminOffset()

	// 收支统计(统计群主管理员队长给普通成员加减比赛分,普通成员给绑定队长的比赛分)
	if (sender.Upper(consts.ROLE_MEMBER) || sender.Partner == 1) && (toMem.URole == consts.ROLE_MEMBER && toMem.Partner != 1) {
		go GetDBMgr().UpdatePaymentsStatistic(house.DBClub.Id, value)
	} else if (sender.URole == consts.ROLE_MEMBER && sender.Partner != 1) && (toMem.Partner == 1 || sender.Upper(consts.ROLE_MEMBER)) {
		go GetDBMgr().UpdatePaymentsStatistic(house.DBClub.Id, -value)
	}
	ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
	ntf.HId = req.HId
	ntf.OptId = p.Uid
	ntf.OptRole = optRole
	ntf.UId = req.UId
	ntf.Value = static.SwitchVitaminToF64(toMem.UVitamin)
	hp := GetPlayerMgr().GetPlayer(req.UId)
	if hp != nil {
		hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
	}
	if onwerPartner {
		ntf.Value = -1 * ntf.Value
		sender.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
	}
	return xerrors.SuccessCode, nil

}

func Proto_ClubHouseVitaminPoolAdd(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req, ok := data.(*static.MsgVitaminSend)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if req.Value == 0 || req.Hid <= 0 {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	value := static.SwitchF64ToVitamin(req.Value)
	house, _, mem, cuserr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	// if cuserr != xerrors.RespOk {
	//	return cuserr.Code, cuserr.Msg
	// }
	if cuserr != xerrors.RespOk {
		if cuserr == xerrors.InvalidPermission {
			if !mem.IsVitaminAdmin() {
				return cuserr.Code, cuserr.Msg
			}
		} else {
			return cuserr.Code, cuserr.Msg
		}
	}
	if mem.UVitamin <= 0 && req.Value < 0 {
		return xerrors.VitaminSendNotEnoughError.Code, xerrors.VitaminSendNotEnoughError.Msg
	}
	if req.Value < 0 && mem.UVitamin+value < 0 {
		return xerrors.VitaminSendPoolError.Code, xerrors.VitaminSendPoolError.Msg
	}
	if req.Value <= 0 {
		return Proto_ClubHouseVitaminSet(s, p, &static.Msg_CH_HouseVitaminSet{
			HId:   req.Hid,
			UId:   mem.UId,
			Value: req.Value,
		})
	}
	tx := GetDBMgr().GetDBmControl().Begin()
	_, aftVitamin, err := mem.VitaminIncrement(mem.UId, value, models.PoolAdminAdd, tx)
	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorf("error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	xerr := house.PoolChange(mem.UId, models.PoolAdminAdd, -1*static.SwitchF64ToVitamin(req.Value), tx)
	if xerr != nil {
		tx.Rollback()
		xlog.Logger().Errorf("error:%v", err)
		return xerr.Code, xerr.Msg
	}
	// 修改疲劳值统计管理节点信息
	err = GetDBMgr().UpdateVitaminMgrList(house.DBClub.Id, p.Uid, aftVitamin, tx)
	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorf("error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorf("error:%v", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	mem.Flush()

	go house.Broadcast(consts.ROLE_ADMIN, consts.MsgHouseVitaminPoolAdd_NTF, req)
	return xerrors.SuccessCode, nil

}

func Proto_ClubHouseVitaminPoolLog(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	req, ok := data.(*static.Msg_CH_HouseVitaminSetRecords)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, cuserr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cuserr != xerrors.RespOk {
		if cuserr == xerrors.InvalidPermission {
			if !mem.IsVitaminAdmin() {
				return cuserr.Code, cuserr.Msg
			}
		} else {
			return cuserr.Code, cuserr.Msg
		}
	}
	ack, cusErr := house.GetVitaminPoolLog(req.Start, req.Count)
	if cusErr != nil {
		return cusErr.Code, cusErr.Msg
	}
	return xerrors.SuccessCode, ack
}

// 包厢疲劳值主开关信息
func Proto_GetHouseVitaminInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseId)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if optmem.Lower(consts.ROLE_CREATER) {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	ack := new(static.Msg_CH_HouseVitaminValues)
	ack.HId = house.DBClub.HId
	ack.Status = house.DBClub.IsVitamin
	ack.GamePause = house.DBClub.IsGamePause
	ack.AdminHide = house.DBClub.IsVitaminHide
	ack.AdminModi = house.DBClub.IsVitaminModi
	ack.PartnerHide = house.DBClub.IsPartnerHide
	ack.PartnerModi = house.DBClub.IsPartnerModi
	ack.MemberSend = house.DBClub.IsMemberSend
	ack.IsDeductConfig = house.ConfiguredDeduct()
	ack.IsEffectConfig = house.ConfiguredEffect()
	ack.DisableSetJuniorVitamin = house.DBClub.DisVitaminJunior
	ack.PartnerKick = house.DBClub.PartnerKick
	ack.NoSkipVitaminSet = house.DBClub.NoSkipVitaminSet
	ack.RewardBalanced = house.DBClub.RewardBalanced

	return xerrors.SuccessCode, ack
}

// 包厢疲劳值数值设置
func Proto_ClubHouseVitaminValues(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HouseVitaminValues)
	if !ok {
		return xerrors.ResultErrorCode, nil
	}

	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_ADMIN, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if house.DBClub.IsVitaminHide {
		// 权限
		if optmem.Lower(consts.ROLE_ADMIN) {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}

	} else {
		// 权限
		if optmem.Lower(consts.ROLE_CREATER) {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}
	}

	// 广播消息角色标志位
	role := consts.ROLE_CREATER
	// 给队长广播消息标志位
	bp := false
	// 刷新数据库标志位
	do := false
	// 总开关变化标志位
	switchIsVitamin, switchIsGamePause := false, false

	// 需要通知所有成员
	if house.DBClub.IsVitamin != req.Status {
		if !house.DBClub.IsFrozen {
			return xerrors.NotFrozenError.Code, xerrors.NotFrozenError.Msg
		}

		if house.CheckInGameTable() {
			cuserr := xerrors.NewXError("当前大厅有玩家已入桌，无法开/关防沉迷")
			return cuserr.Code, cuserr.Msg
		}

		switchIsVitamin = true
		do = true
		role = consts.ROLE_MEMBER
		house.ModifyVitaminStatus(req.Status)
	}

	// 需要通知管理员
	if house.DBClub.IsVitaminHide != req.AdminHide {
		do = true
		house.ModifyVitaminAdminHide(req.AdminHide)
		if role < consts.ROLE_ADMIN {
			role = consts.ROLE_ADMIN
		}
	}

	// 需要通知管理员
	if house.DBClub.IsVitaminModi != req.AdminModi {
		do = true
		house.ModifyVitaminAdminModify(req.AdminModi)
		if role < consts.ROLE_ADMIN {
			role = consts.ROLE_ADMIN
		}
	}

	// 需要通知队长
	if house.DBClub.IsPartnerHide != req.PartnerHide {
		do = true
		house.ModifyVitaminPartnerHide(req.PartnerHide)
		bp = true
	}

	if house.DBClub.IsPartnerModi != req.PartnerModi {
		do = true
		house.ModifyVitaminPartnerModi(req.PartnerModi)
		opt := ""
		if req.PartnerModi {
			opt = "打开"
		} else {
			opt = "关闭"
		}
		msg := fmt.Sprintf("盟主%s了队长可调开关", opt)
		CreateClubMassage(house.DBClub.Id, house.DBClub.UId, PartnerVitamin, msg)
		bp = true
	}

	// 只用通知盟主
	if house.DBClub.IsGamePause != req.GamePause {
		if house.CheckInGameTable() {
			cuserr := xerrors.NewXError("当前大厅有玩家已入桌，无法开/关中途暂停")
			return cuserr.Code, cuserr.Msg
		}

		do = true
		switchIsGamePause = true
		house.ModifyVitaminGamePause(req.GamePause)
	}

	if house.DBClub.IsMemberSend != req.MemberSend {
		do = true
		house.ModifyVitaminMemberSend(req.MemberSend)
		if role < consts.ROLE_MEMBER {
			role = consts.ROLE_MEMBER
		}
		opt := ""
		if req.MemberSend {
			opt = "打开"
		} else {
			opt = "关闭"
		}
		msg := fmt.Sprintf("盟主%s了玩家赠送开关", opt)
		CreateClubMassage(house.DBClub.Id, house.DBClub.UId, MemVitaminSend, msg)
	}

	if house.DBClub.DisVitaminJunior != req.DisableSetJuniorVitamin {
		do = true
		house.ModifyDisSetJuniorVitamin(req.DisableSetJuniorVitamin)
		if role < consts.ROLE_MEMBER {
			role = consts.ROLE_MEMBER
		}
		opt := ""
		if req.DisableSetJuniorVitamin {
			opt = "关闭"
		} else {
			opt = "打开"
		}
		msg := fmt.Sprintf("盟主%s了上级可调下级开关", opt)
		CreateClubMassage(house.DBClub.Id, house.DBClub.UId, MemVitaminSend, msg)
	}

	if house.DBClub.RewardBalanced != req.RewardBalanced {
		do = true
		house.ModifyRewardBalanced(req.RewardBalanced)
		// if role < constant.ROLE_MEMBER {
		//	role = constant.ROLE_MEMBER
		// }
		opt := ""
		if req.RewardBalanced {
			opt = "打开"
		} else {
			opt = "关闭"
		}
		msg := fmt.Sprintf("盟主%s了奖励均衡开关", opt)
		CreateClubMassage(house.DBClub.Id, house.DBClub.UId, MemVitaminSend, msg)
	}

	if house.DBClub.NoSkipVitaminSet != req.NoSkipVitaminSet {
		do = true
		house.ModifyVitaminNoSkip(req.NoSkipVitaminSet)
		// if role < constant.ROLE_MEMBER {
		//	role = constant.ROLE_MEMBER
		// }
		opt := ""
		if req.NoSkipVitaminSet {
			opt = "关闭"
		} else {
			opt = "打开"
		}
		msg := fmt.Sprintf("盟主%s了跨级调整开关", opt)
		CreateClubMassage(house.DBClub.Id, house.DBClub.UId, MemVitaminSend, msg)
	}

	if do {
		house.flush()
	}

	ack := new(static.Msg_CH_HouseVitaminValues)
	ack.HId = house.DBClub.HId
	ack.Status = house.DBClub.IsVitamin
	ack.AdminHide = house.DBClub.IsVitaminHide
	ack.AdminModi = house.DBClub.IsVitaminModi
	ack.GamePause = house.DBClub.IsGamePause
	ack.PartnerHide = house.DBClub.IsPartnerHide
	ack.PartnerModi = house.DBClub.IsPartnerModi
	ack.MemberSend = house.DBClub.IsMemberSend
	ack.IsDeductConfig = house.ConfiguredDeduct()
	ack.IsEffectConfig = house.ConfiguredEffect()
	ack.DisableSetJuniorVitamin = house.DBClub.DisVitaminJunior
	ack.PartnerKick = house.DBClub.PartnerKick
	ack.RewardBalanced = house.DBClub.RewardBalanced
	ack.NoSkipVitaminSet = house.DBClub.NoSkipVitaminSet
	// // 广播消息
	// house.BroadcastMsg(
	//	constant.MsgTypeHouseVitaminStatus_ntf,
	//	ack,
	//	func(member *HouseMember) bool {
	//		if house.IsPartner(member.UId) && bp {
	//			return true
	//		}
	//		if member.URole <= role {
	//			return true
	//		}
	//		return false
	//	},
	// )
	house.CustomBroadcast(role, bp, bp, true, consts.MsgTypeHouseVitaminStatus_ntf, ack)

	if switchIsVitamin {
		house.FloorLock.RLockWithLog()
		for _, floor := range house.Floors {
			if floor == nil {
				continue
			}
			if floor.IsVitamin != house.DBClub.IsVitamin {
				floor.IsVitamin = house.DBClub.IsVitamin
				if err := GetDBMgr().HouseFloorUpdate(floor); err != nil {
					xlog.Logger().Errorln("HouseFloorUpdate.error", err)
				}
			}
		}
		house.FloorLock.RUnlock()
	}

	if switchIsGamePause {
		house.FloorLock.RLockWithLog()
		for _, floor := range house.Floors {
			if floor == nil {
				continue
			}
			if floor.IsGamePause != house.DBClub.IsGamePause {
				floor.IsGamePause = house.DBClub.IsGamePause
				if err := GetDBMgr().HouseFloorUpdate(floor); err != nil {
					xlog.Logger().Errorln("HouseFloorUpdate.error", err)
				}
			}
		}
		house.FloorLock.RUnlock()
	}

	return xerrors.SuccessCode, nil
}

func Proto_GetHFVitaminEffectValues(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	var ack static.MsgHouseFloorEffectInfo
	ack.Hid = house.DBClub.HId
	ack.Items = make([]static.MsgFloorEffectInfo, 0)

	house.FloorLock.RLockWithLog()

	var arr []int64
	for _, f := range house.Floors {
		arr = append(arr, f.Id)
	}
	sort.Sort(util.Int64Slice(arr))
	for _, fid := range arr {
		f := house.Floors[fid]
		if f == nil {
			continue
		}
		var item static.MsgFloorEffectInfo
		item.Fid = f.Id
		item.IsVitamin = f.IsVitamin && house.DBClub.IsVitamin
		item.IsGamePause = f.IsGamePause && house.DBClub.IsGamePause
		item.VitaminLowLimit = static.SwitchVitaminToF64(f.VitaminLowLimit)
		item.VitaminHighLimit = static.SwitchVitaminToF64(f.VitaminHighLimit)
		item.VitaminLowLimitPause = static.SwitchVitaminToF64(f.VitaminLowLimitPause)
		ack.Items = append(ack.Items, item)
	}

	house.FloorLock.RUnlock()
	return xerrors.SuccessCode, &ack
}

// 包厢楼层生效设置相关设置
func Proto_SetHFVitaminEffectValues(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseFloorEffectInfo)

	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	// 参数效验
	for _, item := range req.Items {
		floor := house.GetFloorByFId(item.Fid)
		if floor == nil {
			cerr := xerrors.NewXError("不存在的包厢楼层")
			return cerr.Code, cerr.Msg
		}

		if item.IsVitamin != floor.IsVitamin {
			if floor.CheckInGameTable() {
				cuserr := xerrors.NewXError("当前大厅有玩家已入桌，无法开/关防沉迷")
				return cuserr.Code, cuserr.Msg
			}
		}

		if item.IsGamePause != floor.IsGamePause {
			if floor.CheckInGameTable() {
				cuserr := xerrors.NewXError("当前大厅有玩家已入桌，无法开/关中途暂停")
				return cuserr.Code, cuserr.Msg
			}
		}

		if item.VitaminLowLimit > consts.VitaminInvalidValueCli ||
			item.VitaminLowLimitPause > consts.VitaminInvalidValueCli {
			return xerrors.ArgumentTooLargerError.Code, xerrors.ArgumentTooLargerError.Msg
		}

		if item.IsVitamin && !house.DBClub.IsVitamin {
			return xerrors.VitaminMainOnError.Code, xerrors.VitaminMainOnError.Msg
		}

		if item.IsGamePause && !house.DBClub.IsGamePause {
			return xerrors.VitaminGPauseOnError.Code, xerrors.VitaminGPauseOnError.Msg
		}

		if item.IsVitamin && item.IsGamePause && item.VitaminLowLimit == consts.VitaminInvalidValueCli && item.VitaminLowLimitPause != consts.VitaminInvalidValueCli {
			return xerrors.VitaminLowerLimitError.Code, xerrors.VitaminLowerLimitError.Msg
		}

		if item.IsVitamin && item.IsGamePause && item.VitaminLowLimit != consts.VitaminInvalidValueCli && item.VitaminLowLimitPause != consts.VitaminInvalidValueCli {
			if item.VitaminLowLimit < item.VitaminLowLimitPause {
				return xerrors.VitaminLowerLimitError.Code, xerrors.VitaminLowerLimitError.Msg
			}
		}
	}

	for _, item := range req.Items {
		floor := house.GetFloorByFId(item.Fid)
		if floor == nil {
			continue
		}
		VitaminLowLimit,
			VitaminHighLimit,
			VitaminLowLimitPause :=
			static.SwitchF64ToVitamin(item.VitaminLowLimit),
			static.SwitchF64ToVitamin(item.VitaminHighLimit),
			static.SwitchF64ToVitamin(item.VitaminLowLimitPause)
		db := false
		if floor.IsVitamin != item.IsVitamin {
			floor.IsVitamin = item.IsVitamin
			db = true
		}

		if floor.IsGamePause != item.IsGamePause {
			floor.IsGamePause = item.IsGamePause
			db = true
		}

		houseVitaminChanged := false

		if floor.VitaminLowLimit != VitaminLowLimit {
			floor.VitaminLowLimit = VitaminLowLimit
			db = true
			houseVitaminChanged = true
		}

		if floor.VitaminHighLimit != VitaminHighLimit {
			floor.VitaminHighLimit = VitaminHighLimit
			db = true
			houseVitaminChanged = true
		}

		if floor.VitaminLowLimitPause != VitaminLowLimitPause {
			floor.VitaminLowLimitPause = VitaminLowLimitPause
			db = true
		}

		if db {
			if err := GetDBMgr().HouseFloorUpdate(floor); err != nil {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
			if houseVitaminChanged {
				ntf := new(static.Ntf_HC_HouseFloorKIdModify)
				ntf.HId = floor.HId
				ntf.FId = floor.Id
				ntf.FRule = floor.Rule
				ntf.VitaminLowLimit = static.SwitchVitaminToF64(floor.VitaminLowLimit)
				ntf.VitaminHighLimit = static.SwitchVitaminToF64(floor.VitaminHighLimit)
				area := GetAreaGameByKid(floor.Rule.KindId)
				if area != nil {
					ntf.ImageUrl = area.Icon
					ntf.KindName = area.Name
					ntf.PackageName = area.PackageName
				}
				house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorRuleModify_Ntf, ntf)
			}
		}
	}
	return xerrors.SuccessCode, nil
}

func ProtoHousePartnerRoyaltySet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParnterRoyaltySet)
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}

	if mem.URole > consts.ROLE_CREATER && mem.Partner != 1 {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}

	curPartner := house.GetMemByUId(req.ParnterId)
	if curPartner == nil || !curPartner.IsPartner() || !curPartner.IsJunior() { // 2019-12-10 此次更新后该接口仅提供非一级队长设置（mem.superior != 0）
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	var sssid int64

	if curPartner.Superior > 0 {
		memSup := house.GetMemByUId(curPartner.Superior)
		if memSup == nil {
			return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
		}
		if memSup.IsJunior() {
			sssid = memSup.Superior
		}
	}

	// 得到楼层号
	houseFloorIds, _ := house.GetFloors()
	existFloor := func(fid int64) bool {
		for i := 0; i < len(houseFloorIds); i++ {
			if houseFloorIds[i] == fid {
				return true
			}
		}
		return false
	}

	singleCost, e := house.GetFloorsSingleCostMap()
	if e != nil {
		xlog.Logger().Error(e)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	for fid, cost := range singleCost {
		if cost < 0 && existFloor(fid) {
			return xerrors.HouseVitaminNotAbleError.Code, fmt.Sprintf("%d楼%s", house.GetFloorIndexByFid(fid)+1, xerrors.HouseVitaminNotAbleError.Msg)
		}
	}
	res := GetHousePartPartnersPyramid(house.DBClub.Id, curPartner.UId, curPartner.Superior, sssid)
	sup, ok1 := res[curPartner.Superior]
	if !ok1 {
		sup = make(models.HousePartnerPyramidFloors, 0)
	}
	ssup, ok1 := res[curPartner.Superior]
	if !ok1 {
		ssup = make(models.HousePartnerPyramidFloors, 0)
	}
	// 校验
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		if i < len(req.RoyaltyPercent) {
			if req.RoyaltyPercent[i] >= 0 {
				// 如果设置的值高出100
				if PartnerPercentHigherLimit < req.RoyaltyPercent[i] {
					return xerrors.ResultErrorCode, fmt.Sprintf("%d楼更改配置比例超出上限，请重新配置。", i+1)
				}
				if sssid > 0 {
					ssupConfig := ssup.GetPyramidByFid(fid)
					// 如果上级没取到配置/没有可分配总额/没有配置收益
					if ssupConfig == nil || !ssupConfig.Configurable() || !ssupConfig.ConfiguredRoyaltyPercent() { // 上级未配置
						return xerrors.ResultErrorCode, fmt.Sprintf("您暂未获得%d楼的可分配权限，请联系您的上级/上上级/盟主处理。", i+1)
					}
				}
				// 得到他的上级配置
				supConfig := sup.GetPyramidByFid(fid)
				// 如果上级没取到配置/没有可分配总额/没有配置收益
				if supConfig == nil || !supConfig.Configurable() || !supConfig.ConfiguredRoyaltyPercent() { // 上级未配置
					return xerrors.ResultErrorCode, fmt.Sprintf("您暂未获得%d楼的可分配权限，请联系您的上级/盟主处理。", i+1)
				}
			}
		}
	}
	// 修复
	// if err := FixClubPartnersPyramidForTop(&sup, house.DBClub.Id, curPartner.UId, houseFloorIds, singleCost); err != nil {
	//	return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	// }
	cur, ok2 := res[curPartner.UId]
	if !ok2 {
		cur = make(models.HousePartnerPyramidFloors, 0)
	}

	if sssid > 0 {
		// 修复
		if err := FixClubPartnersPyramidBySuperSuper(&cur, &sup, &ssup, house.DBClub.Id, curPartner.UId, houseFloorIds); err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else {
		// 修复
		if err := FixClubPartnersPyramidBySuper(&cur, &sup, house.DBClub.Id, curPartner.UId, houseFloorIds); err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	// 记录
	histories := make([]*models.HousePartnerRoyaltyHistory, 0)
	// 执行
	var (
		tx     = GetDBMgr().GetDBmControl().Begin()
		err    error
		memMap static.HouseMemberMap
	)
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		if i < len(req.RoyaltyPercent) {
			curConfig := cur.GetPyramidByFid(fid)
			if curConfig == nil {
				continue
			}

			if curConfig.RoyaltyPercent != req.RoyaltyPercent[i] {
				histories = append(histories,
					NewPartnerRoyaltyModifyHistory(house.DBClub.Id, mem, optTypeModifyRoyalty,
						house.GetFloorByFId(fid).GetWanFaName(), fid, i, req.ParnterId, curConfig.RoyaltyPercent, req.RoyaltyPercent[i]))
				if memMap == nil {
					memMap = GetDBMgr().GetHouseMemMap(house.DBClub.Id)
				}
				err = UpdateRoyaltyPercent(tx, memMap, curConfig, req.RoyaltyPercent[i])
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
			}
		}
	}

	if err = tx.Commit().Error; err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
	}

	// 增加修改日志操作
	go AddPartnerRoyaltyModifyHistorys(histories)

	return xerrors.SuccessCode, nil
}

func ProtoHousePartnerRoyaltyGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParnterRoyaltyGet)
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if mem.URole > consts.ROLE_CREATER && mem.Partner != 1 {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}
	curPartner := house.GetMemByUId(req.ParnterId)
	if curPartner == nil || !curPartner.IsPartner() || !curPartner.IsJunior() { // 2019-12-10 此次更新后该接口仅提供非一级队长设置（mem.superior != 0）
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	var sssid int64

	if curPartner.Superior > 0 {
		memSup := house.GetMemByUId(curPartner.Superior)
		if memSup == nil {
			return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
		}
		if memSup.IsJunior() {
			sssid = memSup.Superior
		}
	}

	// 得到楼层号
	houseFloorIds, _ := house.GetFloors()
	res := GetHousePartPartnersPyramid(house.DBClub.Id, curPartner.UId, curPartner.Superior, sssid)
	ssup, ok1 := res[sssid]
	if !ok1 {
		ssup = make(models.HousePartnerPyramidFloors, 0)
	}
	sup, ok1 := res[curPartner.Superior]
	if !ok1 {
		sup = make(models.HousePartnerPyramidFloors, 0)
	}
	cur, ok2 := res[curPartner.UId]
	if !ok2 {
		cur = make(models.HousePartnerPyramidFloors, 0)
	}

	if sssid > 0 {
		// 修复
		if err := FixClubPartnersPyramidBySuperSuper(&cur, &sup, &ssup, house.DBClub.Id, curPartner.UId, houseFloorIds); err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else {
		// 修复
		if err := FixClubPartnersPyramidBySuper(&cur, &sup, house.DBClub.Id, curPartner.UId, houseFloorIds); err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	var ack static.Msg_HC_HouseParnterRoyaltyGet
	ack.Hid = req.Hid
	ack.ParnterId = curPartner.UId
	ack.NickName = curPartner.NickName
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		// 得到原有的配置
		curConfig := cur.GetPyramidByFid(fid)
		if curConfig == nil {
			xlog.Logger().Error("curConfig is nil")
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		ack.RoyaltyPercent = append(ack.RoyaltyPercent, curConfig.RoyaltyPercent)

		if curConfig.ConfiguredRoyaltyPercent() {
			ack.SuperiorPercent = append(ack.SuperiorPercent, int(curConfig.RealSuperiorPercent()))
		} else {
			ack.SuperiorPercent = append(ack.SuperiorPercent, DefaultPartnerProfit)
		}

		if curConfig.ConfiguredRoyaltyPercent() {
			royalty, supRoyalty := curConfig.EarningsInfo()
			ack.Royaltys = append(ack.Royaltys, static.SwitchVitaminToF64(royalty))
			ack.SuperiorProfit = append(ack.SuperiorProfit, static.SwitchVitaminToF64(supRoyalty))
		} else {
			ack.Royaltys = append(ack.Royaltys, DefaultPartnerProfit)
			ack.SuperiorProfit = append(ack.SuperiorProfit, DefaultPartnerProfit)
		}
		if curConfig.Configurable() {
			ack.Distributable = append(ack.Distributable, static.SwitchVitaminToF64(curConfig.Total))
		} else {
			ack.Distributable = append(ack.Distributable, InvalidVitaminCost)
		}
	}
	return xerrors.SuccessCode, &ack
}

func ProtoHouseOwnerRoyaltyGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseOwnerRoyaltyGet)
	house, _, _, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	curPartner := house.GetMemByUId(req.ParnterId)
	if curPartner == nil || !curPartner.IsPartner() || curPartner.IsJunior() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	houseFloorIds, _ := house.GetFloors()
	singleCost, e := house.GetFloorsSingleCostMap()
	if e != nil {
		xlog.Logger().Error(e)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	cur := GetHousePartnerPyramid(house.DBClub.Id, curPartner.UId)
	if err := FixClubPartnersPyramidForTop(&cur, house.DBClub.Id, curPartner.UId, houseFloorIds, singleCost); err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseOwnerRoyaltyGet
	ack.Hid = house.DBClub.HId
	ack.ParnterId = curPartner.UId
	ack.NickName = curPartner.NickName

	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		// 得到原有的配置
		curConfig := cur.GetPyramidByFid(fid)
		if curConfig == nil {
			xlog.Logger().Error("curConfig is nil")
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}

		if curConfig.Configurable() {
			ack.SingleCost = append(ack.SingleCost, static.SwitchVitaminToF64(curConfig.Total))
		} else {
			ack.SingleCost = append(ack.SingleCost, InvalidVitaminCost)
		}

		if curConfig.ConfiguredRoyaltyPercent() {
			ack.RoyaltyPercent = append(ack.RoyaltyPercent, curConfig.RoyaltyPercent)
			royalty, _ := curConfig.EarningsInfo()
			ack.Royaltys = append(ack.Royaltys, static.SwitchVitaminToF64(royalty))
		} else {
			ack.RoyaltyPercent = append(ack.RoyaltyPercent, DefaultPartnerProfit)
			ack.Royaltys = append(ack.Royaltys, DefaultPartnerProfit)
		}

		ack.JuniorPercent = append(ack.JuniorPercent, DefaultPartnerProfit)
		ack.JuniorProfit = append(ack.JuniorProfit, DefaultPartnerProfit)
	}
	return xerrors.SuccessCode, &ack
}

func ProtoHouseOwnerRoyaltySet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseOwnerRoyaltySet)

	house, _, optMem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}

	curPartner := house.GetMemByUId(req.ParnterId)
	if curPartner == nil || !curPartner.IsPartner() || curPartner.IsJunior() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	houseFloorIds, _ := house.GetFloors()
	existFloor := func(fid int64) bool {
		for i := 0; i < len(houseFloorIds); i++ {
			if houseFloorIds[i] == fid {
				return true
			}
		}
		return false
	}

	singleCost, e := house.GetFloorsSingleCostMap()
	if e != nil {
		xlog.Logger().Error(e)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	for fid, cost := range singleCost {
		if cost < 0 && existFloor(fid) {
			return xerrors.HouseVitaminNotAbleError.Code, fmt.Sprintf("%d楼%s", house.GetFloorIndexByFid(fid)+1, xerrors.HouseVitaminNotAbleError.Msg)
		}
	}

	// 校验
	for i := 0; i < len(houseFloorIds); i++ {
		if i < len(req.RoyaltyPercent) {
			if req.RoyaltyPercent[i] >= 0 {
				// 如果设置的值高出100
				if PartnerPercentHigherLimit < req.RoyaltyPercent[i] {
					return xerrors.ResultErrorCode, fmt.Sprintf("%d楼更改配置比例超出上限，请重新配置。", i+1)
				}
			}
		}
	}

	cur := GetHousePartnerPyramid(house.DBClub.Id, curPartner.UId)
	if err := FixClubPartnersPyramidForTop(&cur, house.DBClub.Id, curPartner.UId, houseFloorIds, singleCost); err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	// 记录
	histories := make([]*models.HousePartnerRoyaltyHistory, 0)
	// 执行
	var (
		tx     = GetDBMgr().GetDBmControl().Begin()
		err    error
		memMap static.HouseMemberMap
	)
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		if i < len(req.RoyaltyPercent) {
			curConfig := cur.GetPyramidByFid(fid)
			if curConfig == nil {
				continue
			}

			if curConfig.RoyaltyPercent != req.RoyaltyPercent[i] {
				histories = append(histories,
					NewPartnerRoyaltyModifyHistory(house.DBClub.Id, optMem, optTypeModifyRoyalty,
						house.GetFloorByFId(fid).GetWanFaName(), fid, i, req.ParnterId, curConfig.RoyaltyPercent, req.RoyaltyPercent[i]))
				if memMap == nil {
					memMap = GetDBMgr().GetHouseMemMap(house.DBClub.Id)
				}
				err = UpdateRoyaltyPercent(tx, memMap, curConfig, req.RoyaltyPercent[i])
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
			}
		}
	}

	if err = tx.Commit().Error; err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
	}

	// 增加修改日志操作
	go AddPartnerRoyaltyModifyHistorys(histories)

	return xerrors.SuccessCode, nil
}

func ProtoHousePartnerSuperiorList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParnterSuperiorList)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	mem := house.GetMemByUId(req.ParnterId)
	if mem == nil {
		return xerrors.UserExistError.Code, xerrors.UserExistError.Msg
	}

	if !mem.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	if req.Page < 0 {
		xlog.Logger().Errorf("请求页数异常,page = %d", req.Page)
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	pMembers := house.GetAllPartnerWithoutJunior(mem.Superior)
	var conditionArr []static.HouseParnterSuperiorListItem

	for i := 0; i < len(pMembers); i++ {
		// 设置的人不需要加到列表
		if pMembers[i].UId == mem.UId {
			continue
		}

		if pMembers[i].Superior == mem.UId {
			continue
		}

		if req.SearchKey != "" {
			if !(strings.Contains(pMembers[i].NickName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", pMembers[i].UId), req.SearchKey)) {
				continue
			}
		}

		item := static.HouseParnterSuperiorListItem{}
		item.UName = pMembers[i].NickName
		item.UId = pMembers[i].UId
		item.UUrl = pMembers[i].ImgUrl
		item.UGender = pMembers[i].Sex

		conditionArr = append(conditionArr, item)
	}

	var ack static.Msg_HC_HouseParnterSuperiorList
	ack.Items = []static.HouseParnterSuperiorListItem{}
	ack.Hid = req.Hid
	ack.SuperiorId = mem.Superior
	ack.TotalPage = (len(conditionArr)-1)/10 + 1

	if ack.TotalPage > 0 {
		if req.Page < ack.TotalPage {
			ack.CurPage = req.Page
		} else {
			ack.CurPage = ack.TotalPage - 1
		}
	} else {
		ack.CurPage = 0
	}

	pbegin := ack.CurPage * 10
	pend := pbegin + 9
	for i := 0; i < len(conditionArr); i++ {
		if i >= pbegin && i <= pend {
			ack.Items = append(ack.Items, conditionArr[i])
		}
	}

	return xerrors.SuccessCode, &ack
}

func ProtoHousePartnerFloorStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_MsgHouseParnterFloorStatistics)

	house, _, opmem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if opmem.URole != consts.ROLE_CREATER && !opmem.IsVitaminAdmin() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	expMap, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 获取包厢所有玩家redis数据
	houseMems := house.GetMemberMap(false)
	floorIds, _ := house.GetFloors()

	floorPartnerRoyalty := make(map[int64]int64)
	if req.FidIndex != -1 && req.FidIndex < len(floorIds) {
		floorPartnerRoyalty = GetHouseFloorPartnersRoyalty(house.DBClub.Id, floorIds[req.FidIndex])
	}
	reqFloorId := int64(0)
	if req.FidIndex < len(floorIds) && req.FidIndex >= 0 {
		reqFloorId = floorIds[req.FidIndex]
	}
	profitMap, err := GetDBMgr().SelectPartnerProfitWithAllPartners(house.DBClub.Id, req.FidIndex, reqFloorId, req.DayType)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseParnterFloorStatistics
	ack.Hid = req.Hid

	for _, hMem := range houseMems {
		if !hMem.IsPartner() {
			continue
		}

		if req.SearchKey != "" {
			if !(strings.Contains(hMem.NickName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", hMem.UId), req.SearchKey)) {
				continue
			}
		}

		statisticsItem := static.ClubPartnerFloorStatisticsItem{}
		statisticsItem.UId = hMem.UId
		statisticsItem.UName = hMem.NickName
		statisticsItem.UUrl = hMem.ImgUrl
		statisticsItem.UGender = hMem.Sex
		statisticsItem.ValidTimes = 0
		statisticsItem.BigValidTimes = 0
		statisticsItem.RoundProfit = 0
		statisticsItem.SubordinateProfit = 0
		statisticsItem.TotalProfit = 0
		statisticsItem.IsJunior = hMem.Superior > 0
		statisticsItem.VitaminAdmin = hMem.VitaminAdmin
		statisticsItem.VicePartner = hMem.VicePartner

		expInfo, ok := expMap[hMem.UId]
		if ok {
			statisticsItem.Exp = expInfo.Exp
		} else {
			statisticsItem.Exp = 0
		}

		profit, ok := profitMap[hMem.UId]
		if ok {
			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SelfProfit
			statisticsItem.SubordinateProfit = profit.SubProfit
			statisticsItem.TotalProfit = profit.SelfProfit + profit.SubProfit
		}

		if req.FidIndex != -1 && req.FidIndex < len(floorIds) {
			royalty, ok := floorPartnerRoyalty[hMem.UId]
			if ok {
				statisticsItem.Royalty = int(royalty)
			} else {
				statisticsItem.Royalty = -1
			}
		}
		ack.Items = append(ack.Items, &statisticsItem)
	}

	sort.Sort(static.ClubPartnerFloorStatisticsItemWrapper{Item: ack.Items, By: func(item1, item2 static.ClubPartnerFloorStatisticsItem) bool {
		if item1.UId == opmem.UId {
			return true // 自己永远在最前面
		} else if item2.UId == opmem.UId {
			return false // 自己永远在最前面
		} else if item1.IsJunior && !item2.IsJunior {
			return false // 一级队长在前面
		} else if !item1.IsJunior && item2.IsJunior {
			return true // 一级队长在前面
		} else {
			// 其他的按有效局排序
			if item1.ValidTimes > item2.ValidTimes {
				return true
			} else if item1.ValidTimes == item2.ValidTimes {
				if item1.UId > item2.UId {
					return true
				} else {
					return false
				}
			} else {
				return false
			}
		}
	}})

	return xerrors.SuccessCode, &ack
}

func ProtoHousePartnerFloorMedStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_MsgHouseParnterFloorMemStatistics)

	house, _, opMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if opMem.Partner != 1 && opMem.URole != consts.ROLE_CREATER && !opMem.IsVitaminAdmin() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if opMem.Partner == 1 && opMem.UId != req.ParnterId {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	memMaps := house.GetMemberMap(false)

	floorIds, _ := house.GetFloors()
	reqFloorId := int64(0)
	if req.FidIndex < len(floorIds) && req.FidIndex >= 0 {
		reqFloorId = floorIds[req.FidIndex]
	}

	selfProfit, partnerProfit, err := GetDBMgr().SelectPartnerProfitWithPartnerDetail(house.DBClub.Id, req.FidIndex, reqFloorId, req.DayType, opMem.UId)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseParnterFloorStatistics
	ack.Hid = req.Hid

	juniorMap := house.GetAllJunior(opMem.UId)

	selfMap := make(map[int64]*static.ClubPartnerFloorStatisticsItem)
	parnterMap := make(map[int64]*static.ClubPartnerFloorStatisticsItem)
	// 初始化名下所有玩家数据
	for _, hMem := range memMaps {
		if hMem.Partner == opMem.UId || hMem.UId == opMem.UId {
			statisticsItem := new(static.ClubPartnerFloorStatisticsItem)
			statisticsItem.UId = hMem.UId
			statisticsItem.UName = hMem.NickName
			statisticsItem.UUrl = hMem.ImgUrl
			statisticsItem.UGender = hMem.Sex
			statisticsItem.ValidTimes = 0
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = 0
			statisticsItem.SubordinateProfit = 0
			statisticsItem.TotalProfit = 0
			statisticsItem.IsJunior = hMem.Superior > 0
			statisticsItem.VitaminAdmin = hMem.VitaminAdmin
			statisticsItem.VicePartner = hMem.VicePartner

			selfMap[hMem.UId] = statisticsItem
		} else if _, ok := juniorMap[hMem.UId]; ok {
			statisticsItem := new(static.ClubPartnerFloorStatisticsItem)
			statisticsItem.UId = hMem.UId
			statisticsItem.UName = hMem.NickName
			statisticsItem.UUrl = hMem.ImgUrl
			statisticsItem.UGender = hMem.Sex
			statisticsItem.ValidTimes = 0
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = 0
			statisticsItem.SubordinateProfit = 0
			statisticsItem.TotalProfit = 0
			statisticsItem.IsJunior = hMem.Superior > 0
			statisticsItem.VitaminAdmin = hMem.VitaminAdmin
			statisticsItem.VicePartner = hMem.VicePartner

			parnterMap[hMem.UId] = statisticsItem
		}
	}

	// 根据统计设置名下相关收益数值
	for uid, profit := range selfProfit {
		statisticsItem, ok := selfMap[uid]
		if !ok {
			statisticsItem := new(static.ClubPartnerFloorStatisticsItem)
			hMem, ok := memMaps[uid]
			if ok {
				statisticsItem.UName = hMem.NickName
				statisticsItem.UUrl = hMem.ImgUrl
				statisticsItem.UGender = hMem.Sex
			} else {
				person, err := GetDBMgr().GetDBrControl().GetPerson(uid)
				if err == nil {
					statisticsItem.UName = person.Nickname
					statisticsItem.UUrl = person.Imgurl
					statisticsItem.UGender = person.Sex
				}
			}

			statisticsItem.UId = uid
			statisticsItem.ValidTimes = 0
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = 0
			statisticsItem.SubordinateProfit = 0
			statisticsItem.TotalProfit = 0
			statisticsItem.IsJunior = hMem.Superior > 0
			statisticsItem.VitaminAdmin = hMem.VitaminAdmin
			statisticsItem.VicePartner = hMem.VicePartner

			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SelfProfit
			statisticsItem.TotalProfit = profit.SelfProfit

			selfMap[uid] = statisticsItem
		} else {
			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SelfProfit
			statisticsItem.TotalProfit = profit.SelfProfit
		}
	}

	// 根据下级合伙人设置相关收益值
	for uid, profit := range partnerProfit {
		statisticsItem, ok := parnterMap[uid]
		if !ok {
			statisticsItem := new(static.ClubPartnerFloorStatisticsItem)
			hMem, ok := memMaps[uid]
			if ok {
				statisticsItem.UName = hMem.NickName
				statisticsItem.UUrl = hMem.ImgUrl
				statisticsItem.UGender = hMem.Sex
			} else {
				person, err := GetDBMgr().GetDBrControl().GetPerson(uid)
				if err == nil {
					statisticsItem.UName = person.Nickname
					statisticsItem.UUrl = person.Imgurl
					statisticsItem.UGender = person.Sex
				}
			}

			statisticsItem.UId = uid
			statisticsItem.ValidTimes = 0
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = 0
			statisticsItem.SubordinateProfit = 0
			statisticsItem.TotalProfit = 0
			statisticsItem.IsJunior = hMem.Superior > 0
			statisticsItem.VitaminAdmin = hMem.VitaminAdmin
			statisticsItem.VicePartner = hMem.VicePartner

			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SubProfit
			statisticsItem.TotalProfit = profit.SubProfit

			parnterMap[uid] = statisticsItem
		} else {
			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SubProfit
			statisticsItem.TotalProfit = profit.SubProfit
		}
	}

	// 添加合伙人相关
	for _, item := range selfMap {
		item.PartnerDeep, item.Superior, _ = house.GetPartnerDeep(memMaps, item.UId)
		ack.Items = append(ack.Items, item)
	}

	// 添加名下玩家相关
	for _, item := range parnterMap {
		item.PartnerDeep, item.Superior, _ = house.GetPartnerDeep(memMaps, item.UId)
		ack.Items = append(ack.Items, item)
	}

	sort.Sort(static.ClubPartnerFloorStatisticsItemWrapper{Item: ack.Items, By: func(item1, item2 static.ClubPartnerFloorStatisticsItem) bool {
		if item1.TotalProfit > item2.TotalProfit {
			return true // 自己永远在最前面
		} else {
			if item1.UId < item2.UId {
				return true
			}
			return false // 自己永远在最前面
		}
	}})

	return xerrors.SuccessCode, &ack
}

func ProtoHousePartnerFloorJuniorStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_MsgHouseParnterFloorStatistics)

	house, _, opmem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if !opmem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 获取包厢所有玩家redis数据
	houseMems := house.GetMemberMap(false)
	floorIds, _ := house.GetFloors()
	// floorCostMap, err := house.GetFloorsSingleCostMap()
	// if err != nil {
	//	syslog.Logger().Error(err)
	//	return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	// }
	floorPartnerRoyalty := make(map[int64]int64)
	if req.FidIndex != -1 && req.FidIndex < len(floorIds) {
		floorPartnerRoyalty = GetHouseFloorPartnersRoyalty(house.DBClub.Id, floorIds[req.FidIndex])
	}

	reqFloorId := int64(0)
	if req.FidIndex < len(floorIds) && req.FidIndex >= 0 {
		reqFloorId = floorIds[req.FidIndex]
	}
	profitMap, err := GetDBMgr().SelectPartnerProfitWithAllPartners(house.DBClub.Id, req.FidIndex, reqFloorId, req.DayType)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseParnterFloorStatistics
	ack.Hid = req.Hid

	for _, hMem := range houseMems {
		if hMem.IsPartner() {
			if hMem.UId != opmem.UId && hMem.Superior != opmem.UId {
				continue
			}
		} else {
			continue
		}

		if req.SearchKey != "" {
			if !(strings.Contains(hMem.NickName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", hMem.UId), req.SearchKey)) {
				continue
			}
		}

		statisticsItem := static.ClubPartnerFloorStatisticsItem{}
		statisticsItem.UId = hMem.UId
		statisticsItem.UName = hMem.NickName
		statisticsItem.UUrl = hMem.ImgUrl
		statisticsItem.UGender = hMem.Sex
		statisticsItem.ValidTimes = 0
		statisticsItem.BigValidTimes = 0
		statisticsItem.RoundProfit = 0
		statisticsItem.SubordinateProfit = 0
		statisticsItem.TotalProfit = 0
		statisticsItem.IsJunior = hMem.Superior > 0
		statisticsItem.VitaminAdmin = hMem.VitaminAdmin
		statisticsItem.VicePartner = hMem.VicePartner

		profit, ok := profitMap[hMem.UId]
		if ok {
			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SelfProfit
			statisticsItem.SubordinateProfit = profit.SubProfit
			statisticsItem.TotalProfit = profit.SelfProfit + profit.SubProfit
		}

		if req.FidIndex != -1 && req.FidIndex < len(floorIds) {
			// reqFid := floorIds[req.FidIndex]
			royalty, ok := floorPartnerRoyalty[hMem.UId]
			if ok {
				statisticsItem.Royalty = int(royalty)
			} else {
				statisticsItem.Royalty = -1
			}
		}

		ack.Items = append(ack.Items, &statisticsItem)
	}

	sort.Sort(static.ClubPartnerFloorStatisticsItemWrapper{Item: ack.Items, By: func(item1, item2 static.ClubPartnerFloorStatisticsItem) bool {
		if item1.UId == opmem.UId {
			return true // 自己永远在最前面
		} else if item2.UId == opmem.UId {
			return false // 自己永远在最前面
		} else {
			// 其他的按有效局排序
			if item1.ValidTimes > item2.ValidTimes {
				return true
			} else if item1.ValidTimes == item2.ValidTimes {
				if item1.UId > item2.UId {
					return true
				} else {
					return false
				}
			} else {
				return false
			}
		}
	}})

	return xerrors.SuccessCode, &ack
}

func ProtoHousePartnerFloorHistoryStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_MsgHousePartnerHistoryFloorStatistics)
	house, _, opmem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if !house.IsPartner(opmem.UId) && opmem.URole != consts.ROLE_CREATER && !opmem.IsVitaminAdmin() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	_, newDelFloorMap, err := GetDelFloorInfoWithTime(house.DBClub.Id, static.GetZeroTime(time.Now().AddDate(0, 0, -2)).Unix())
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	hfss := make(map[int64]*static.ClubPartnerFloorHistoryStatisticsItem)

	if len(newDelFloorMap) > 0 {
		var statisticsItems []models.HousePartnerRoyaltyDetailItem
		if opmem.IsPartner() {
			statisticsItems, _ = GetDBMgr().SelectDelFloorPartnerProfitByPartner(house.DBClub.Id, newDelFloorMap, opmem.UId)
		} else if opmem.IsVicePartner() {
			statisticsItems, _ = GetDBMgr().SelectDelFloorPartnerProfitByPartner(house.DBClub.Id, newDelFloorMap, opmem.Partner)
		} else {
			statisticsItems, _ = GetDBMgr().SelectDelFloorPartnerProfit(house.DBClub.Id, newDelFloorMap)
		}
		floorCostMap := GetClubPartnersRoyalty(house.DBClub.Id)

		for _, item := range statisticsItems {
			statisticsItem := new(static.ClubPartnerFloorHistoryStatisticsItem)
			statisticsItem.DFid = item.DFid
			if opmem.IsPartner() || opmem.IsVicePartner() {
				memRoyalty, ok := floorCostMap[opmem.UId]
				if ok {
					royalty, ok := memRoyalty[item.DFid]
					if ok {
						statisticsItem.Royalty = int(royalty)
					}
				}
				statisticsItem.SubordinateProfit = item.SubProfit
			} else {
				statisticsItem.Royalty = -1
				statisticsItem.SubordinateProfit = -1
			}
			statisticsItem.ValidTimes = item.ValidTimes
			statisticsItem.TotalProfit = item.SelfProfit + item.SubProfit

			dTime, ok := newDelFloorMap[item.DFid]
			if ok {
				statisticsItem.DeteleTime = dTime.CreateStamp
				statisticsItem.Fid = dTime.DFIndex
			} else {
				statisticsItem.DeteleTime = 0
				statisticsItem.Fid = 0
			}
			hfss[item.DFid] = statisticsItem
		}
	}
	ack := static.Msg_HC_HouseParnterFloorHistoryStatistics{}
	ack.Items = []*static.ClubPartnerFloorHistoryStatisticsItem{}
	for _, sItem := range hfss {
		ack.Items = append(ack.Items, sItem)
	}

	sort.Sort(static.ClubPartnerFloorHistoryStatisticsItemWrapper{Item: ack.Items, By: func(item1, item2 static.ClubPartnerFloorHistoryStatisticsItem) bool {
		if item1.DeteleTime > item2.DeteleTime {
			return true
		} else {
			return false
		}
	}})

	return xerrors.SuccessCode, &ack
}

func ProtoHousePartnerFloorHistoryDetailStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_MsgHousePartnerHistoryFloorDetailStatistics)
	house, _, opmem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if opmem.URole != consts.ROLE_CREATER && !opmem.IsVitaminAdmin() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	memMaps := house.GetMemberMap(false)

	_, newDelFloorMap, err := GetDelFloorInfoWithFid(house.DBClub.Id, req.DFid)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var ack static.Msg_HC_HouseParnterFloorStatistics
	ack.Hid = req.Hid
	ack.Items = []*static.ClubPartnerFloorStatisticsItem{}
	var sItem []*static.ClubPartnerFloorStatisticsItem

	if len(newDelFloorMap) == 1 {
		var sItems []models.HousePartnerRoyaltyDetailItem
		sItems, _ = GetDBMgr().SelectDelFloorPartnerProfitWithFloorByPartner(house.DBClub.Id, newDelFloorMap)
		floorCostMap := GetClubPartnersRoyalty(house.DBClub.Id)
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		for _, item := range sItems {
			statisticsItem := new(static.ClubPartnerFloorStatisticsItem)
			pMem, ok := memMaps[item.Beneficiary]
			if ok {
				statisticsItem.UName = pMem.NickName
				statisticsItem.UUrl = pMem.ImgUrl
				statisticsItem.UGender = pMem.Sex

				statisticsItem.IsJunior = pMem.Superior > 0
				statisticsItem.VitaminAdmin = pMem.VitaminAdmin
				statisticsItem.VicePartner = pMem.VicePartner
			} else {
				pPerson, _ := GetDBMgr().GetDBrControl().GetPerson(item.Beneficiary)
				statisticsItem.UName = pPerson.Nickname
				statisticsItem.UUrl = pPerson.Imgurl
				statisticsItem.UGender = pPerson.Sex
			}
			statisticsItem.UId = item.Beneficiary
			statisticsItem.ValidTimes = item.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = 0
			statisticsItem.SubordinateProfit = item.SubProfit
			statisticsItem.TotalProfit = item.SelfProfit + item.SubProfit
			memRoyalty, ok := floorCostMap[pMem.UId]
			if ok {
				royalty, ok := memRoyalty[req.DFid]
				if ok {
					statisticsItem.Royalty = int(royalty)
				}
			}
			sItem = append(sItem, statisticsItem)
		}
	}
	sort.Sort(static.ClubPartnerFloorStatisticsItemWrapper{Item: sItem, By: func(item1, item2 static.ClubPartnerFloorStatisticsItem) bool {
		if item1.UId < item2.UId {
			return true
		} else {
			return false
		}
	}})

	for i := 0; i < len(sItem); i++ {
		if i >= req.Start && len(ack.Items) < req.Count {
			ack.Items = append(ack.Items, sItem[i])
		}

		if len(ack.Items) >= req.Count {
			break
		}
	}

	return xerrors.SuccessCode, &ack
}

func ProtoHouseAutoPayPartner(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_HouseAutoPayPartner)
	house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)

	if mem.URole > consts.ROLE_CREATER && !mem.IsVitaminAdmin() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg

	}

	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house.DBClub.AutoPayPartnrt != req.AutoPay {
		if e := house.OptionAutoPay(req.AutoPay); e != nil {
			return e.Code, e.Msg
		}
	}

	return xerrors.SuccessCode, nil
}

// 绑定队长下级 -仅限队长权限
func ProtoHouseBindPartnerJunior(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	if GetServer().ConHouse.NoPartnerJunior {
		return xerrors.ResultErrorCode, "目前不允许队长设置下级。"
	}

	req, ok := data.(*static.Msg_CH_HouseParnterBindJunior)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	//_, xerr := GetDBMgr().GetUserAgentConfig(req.Junior)
	//if xerr != nil {
	//	if xerr.Code == xerrors.UserAgentNotConfigErrorCode {
	//		return xerrors.ResultErrorCode, "该玩家未注册盟主后台，无法设置成为队长。"
	//	}
	//	return xerr.Code, xerr.Msg
	//}
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}
	sw, _ := house.CheckMemberSwitch("CapSetDep")
	if sw == 0 {
		return xerrors.ResultErrorCode, "盟主关闭了此功能"
	}
	xe := house.BindPartnerJunior(p.Uid, p.Uid, req.Junior)
	if xe != nil {
		return xe.Code, xe.Msg
	}

	// 设置下级队长
	house.InitCreatePartnerExp(req.Junior)

	// 广播消息
	ntf := new(static.Ntf_HC_HousePartnerJunior)
	ntf.HId = house.DBClub.HId
	ntf.Opt = p.Uid
	ntf.Uid = req.Junior
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHousePartnerJunior_Ntf, ntf)
	return xerrors.SuccessCode, nil
}

// 绑定队长上级 -仅限盟主权限
func ProtoHouseBindPartnerSuperior(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParnterBindSuperior)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	if req.SuperiorId < 0 {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	} else if req.SuperiorId == 0 { // 取消绑定上级
		xe := house.UnbindPartnerJunior(optMem.UId, req.ParnterId)
		if xe != nil {
			return xe.Code, xe.Msg
		}
	} else {
		xe := house.BindPartnerJunior(optMem.UId, req.SuperiorId, req.ParnterId)
		if xe != nil {
			return xe.Code, xe.Msg
		}
	}

	// 广播消息
	ntf := new(static.Ntf_HC_HouseParnterBindSuperior)
	ntf.Hid = house.DBClub.HId
	ntf.Opt = p.Uid
	ntf.ParnterId = req.ParnterId
	ntf.SuperiorId = req.SuperiorId
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseParnterBindSuperiorNtf, ntf)

	return xerrors.SuccessCode, nil
}

// 队长得到邀请码
func ProtoGetHousePartnerInviteCode(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseId)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, mem, xe := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if mem.Partner != 1 {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	partnerInviteCode, err := GetDBMgr().SelectPartnerInviteCode(house.DBClub.Id, mem.UId)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	ack := &static.Msg_HC_HousePartnerInviteCode{
		Code: fmt.Sprintf("%d", partnerInviteCode.InviteCode),
	}
	return xerrors.SuccessCode, ack
}

func ProtoHouseVitaminSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgHouseVitaminAdminSet)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, _, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_CREATER, MinorSetJudge)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	cli := GetDBMgr().Redis
	mem := house.GetMemByUId(req.Uid)

	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if mem.URole > consts.ROLE_MEMBER || mem.Partner > 1 {
		return xerrors.GroupUserBind.Code, xerrors.GroupUserBind.Msg
	}
	if mem.VitaminAdmin == req.IsAdmin {
		return xerrors.SuccessCode, nil
	}
	if house.GetVitaminCount() >= 2 && req.IsAdmin {
		return xerrors.ViAdminMax.Code, xerrors.ViAdminMax.Msg
	}

	mem.Lock(cli)
	mem.VitaminAdmin = req.IsAdmin
	mem.Flush()
	mem.Unlock(cli)

	var msg string
	if req.IsAdmin {
		msg = fmt.Sprintf("盟主将ID:%d %s 设置为比赛分管理员", mem.UId, mem.NickName)
	} else {
		msg = fmt.Sprintf("盟主将ID:%d %s 取消比赛分管理员", mem.UId, mem.NickName)
	}
	GetMRightMgr().setRoleUpdateRight(mem, false)
	go CreateClubMassage(house.DBClub.Id, p.Uid, VitaminAdmin, msg)
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseVitAdminSet_Ntf, req)
	return xerrors.SuccessCode, nil
}

// 包厢副队长设置
func ProtoHouseVicePartnerSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgHouseVicePartnerSet)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	house, _, optMem, xe := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorSetDeputy)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	if !optMem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	mem := house.GetMemByUId(req.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if mem.URole > consts.ROLE_MEMBER {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}
	if mem.IsPartner() {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}
	if mem.Partner != optMem.UId {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if req.VicePartner == mem.VicePartner {
		return xerrors.SuccessCode, nil
	}

	if req.VicePartner && house.GetVicePartnerCount(mem.Partner) >= GetServer().ConHouse.VicePartnerMax {
		return xerrors.VicePartnerMaxError.Code, xerrors.VicePartnerMaxError.Msg
	}

	if err := GetDBMgr().GetDBmControl().Model(models.HouseMember{}).
		Where("hid = ? and uid = ?", house.DBClub.Id, mem.UId).
		Update("vice_partner", req.VicePartner).Error; err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	cli := GetDBMgr().Redis
	mem.Lock(cli)
	mem.VicePartner = req.VicePartner
	mem.Flush()
	mem.Unlock(cli)
	var msg string
	if req.VicePartner {
		//根据队长的权限 设置副队长权限
		GetMRightMgr().changeDepRight(mem, optMem)
		msg = fmt.Sprintf("队长%d将ID:%d %s 设置为副队长", p.Uid, mem.UId, mem.NickName)
	} else {
		GetMRightMgr().setRoleUpdateRight(mem, true)
		msg = fmt.Sprintf("队长%d将ID:%d %s 取消副队长", p.Uid, mem.UId, mem.NickName)
	}
	go CreateClubMassage(house.DBClub.Id, p.Uid, VicePartner, msg)
	req.OptUid = p.Uid
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseVicePartnerSet_Ntf, req)
	return xerrors.SuccessCode, nil
}

func Proto_ClubHousePartnerRoyaltyForMe(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParnterRoyaltyForMe)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if !optMem.IsPartner() && optMem.URole != consts.ROLE_CREATER {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	ack := static.Msg_HC_HousePartnerRoyaltyForMe{}
	ack.Item = []static.HousePartnerRoyaltyForMeItem{}

	floors, _ := house.GetFloors()
	pRoyalty := GetClubRoyaltyByProvider(house.DBClub.Id, floors, optMem.UId)

	pFloorRoyalty, ok := pRoyalty[optMem.UId]
	if !ok {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	for floorIndex, floorId := range floors {
		item := new(static.HousePartnerRoyaltyForMeItem)
		item.FloorIndex = floorIndex
		item.FloorName = house.GetFloorByFId(floorId).GetWanFaName()
		r, ok := pFloorRoyalty[floorId]
		if ok {
			item.MyRoyalty = static.SwitchVitaminToF64(r)
		} else {
			item.MyRoyalty = -1
		}

		ack.Item = append(ack.Item, *item)
	}
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHousePartnerRoyaltyHistory(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParnterRoyaltyHistory)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if !optMem.IsVitaminAdmin() && optMem.URole != consts.ROLE_CREATER && !optMem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	mem := house.GetMemByUId(req.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	if optMem.IsPartner() && req.Uid != optMem.UId && mem.Superior != optMem.UId {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	historys, err := GetRoyaltyModifyHistoryForPartner(house.DBClub.Id, mem.UId)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	ack := static.Msg_HC_HousePartnerRoyaltyHistory{}
	ack.Item = []static.HousePartnerRoyaltyHistoryItem{}
	ack.Uid = mem.UId
	ack.Name = mem.NickName

	for _, history := range historys {
		item := new(static.HousePartnerRoyaltyHistoryItem)
		item.CreatedAt = history.CreatedAt.Unix()
		item.OptFloorIndex = history.OptFloorIndex
		item.OptFloorName = history.OptFloorName
		item.OptUserType, item.OptInfo = GetRoyaltyModifyHistoryOptInfo(history.OptUserType, history.OptType, history.Before, history.After)

		ack.Item = append(ack.Item, *item)
	}
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseNoLeagueStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseNoLeagueStatistics)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorTeamStatistical)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	//if !optMem.IsVitaminAdmin() && optMem.URole != constant.ROLE_CREATER && !optMem.IsPartner() && !optMem.IsVicePartner() {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}

	// 校验时段、区间是否匹配
	timeRangeCnt := 1
	if req.QueryTimeInterval > 0 {
		timeRangeCnt = 24 / req.QueryTimeInterval
	}
	if req.QueryTimeRange > timeRangeCnt || req.QueryTimeRange <= 0 {
		return xerrors.InvalidTimeRangeError.Code, xerrors.InvalidTimeRangeError.Msg
	}

	ack := static.Msg_HC_NoLeagueStatistics{}
	ack.Items = []*static.HouseNoLeagueStatisticsItem{}
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd
	var sortItem []*static.HouseNoLeagueStatisticsItem

	var begin, end time.Time
	now := time.Now()
	// 时段0代表筛选一天的数据
	if req.QueryTimeInterval == 0 {
		req.QueryTimeRange = 1
		req.QueryTimeInterval = 24
	}
	switch req.DayType {
	case static.DAY_RECORD_TODAY:
		begin = static.GetZeroTime(now).Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
		end = static.GetZeroTime(now).Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
	case static.DAY_RECORD_YESTERDAY:
		begin = static.GetZeroTime(now.AddDate(0, 0, -1)).Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
		end = static.GetZeroTime(now.AddDate(0, 0, -1)).Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
	case static.DAY_RECORD_3DAYS:
		begin = static.GetZeroTime(now.AddDate(0, 0, -3))
		end = begin.AddDate(0, 0, 3)
	case static.DAY_RECORD_7DAYS:
		begin = static.GetZeroTime(now.AddDate(0, 0, -7))
		end = begin.AddDate(0, 0, 7)
	default:
		// 其他的默认只查今天
		begin = static.GetZeroTime(now)
		end = now
	}

	// leaveTime之前的记录为有效记录
	leaveTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", begin.Year(), begin.Month(), begin.Day(), begin.Hour(), begin.Minute(), begin.Second())
	leaveMemMap, err := GetDBMgr().SelectAllLeaveHousePartnerMember(house.DBClub.Id, leaveTime)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 查询点赞信息  不影响 总人次 和 队长总人次  只影响 查询条目
	likeMap := make(map[int64]bool)
	if req.DayType < static.DAY_RECORD_3DAYS {
		// 查询日期 具体到某一天
		likeTimeStr := fmt.Sprintf("%d-%02d-%02d", begin.Year(), begin.Month(), begin.Day())

		// 构造查询区间字符串 如01-03
		timeRangeStr := ""
		if req.QueryTimeInterval > 0 {
			timeRangeStr = fmt.Sprintf("%02d-%02d", (req.QueryTimeRange-1)*req.QueryTimeInterval, req.QueryTimeRange*req.QueryTimeInterval)
		} else {
			timeRangeStr = "00-24"
		}

		// 查询身份
		likeOptUserType := models.OptUserTypeAdmin
		if optMem.IsPartner() || optMem.IsVicePartner() {
			likeOptUserType = models.OptUserTypePartner
		}

		// 查询数据库
		likeMap, err = GetDBMgr().SelectHouseRecordTeamLike(house.DBClub.Id, likeOptUserType, likeTimeStr, timeRangeStr)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	// 获取所有成员
	hMems := house.GetMemSimpleToMap(false)
	// 统计查询结果
	pStatisticsMap := make(map[int64]*UidGameSum)
	// 合伙人按自己设置的低分局值查询
	if req.LowScoreFlag > 0 && (optMem.IsPartner() || optMem.IsVicePartner()) {
		pid := int64(0)
		if optMem.IsPartner() {
			pid = optMem.UId
		} else {
			pid = optMem.Partner
		}
		pStatisticsMap = house.SelectPartnerTeamStatistics(pid, req.Fid, begin, end, hMems, leaveMemMap)
	} else {
		pStatisticsMap = house.SelectHouseTeamStatistics(req.Fid, begin, end, hMems, leaveMemMap)
	}

	for _, mem := range hMems {
		if !mem.IsPartner() {
			continue
		}

		if optMem.IsPartner() {
			if mem.Superior != optMem.UId && mem.UId != optMem.UId {
				continue
			}
		}

		if optMem.IsVicePartner() {
			if mem.Superior != optMem.Partner && mem.UId != optMem.Partner {
				continue
			}
		}

		// 筛选点赞记录 不影响 总人次 和 队伍总人次
		bLike := false
		if req.LikeFlag == 1 {
			if _, ok := likeMap[mem.UId]; ok {
				bLike = true
			} else {
				continue
			}
		} else if req.LikeFlag == 2 {
			if _, ok := likeMap[mem.UId]; ok {
				continue
			} else {
				bLike = false
			}
		} else {
			if val, ok := likeMap[mem.UId]; ok {
				bLike = val
			} else {
				bLike = false
			}
		}

		item := new(static.HouseNoLeagueStatisticsItem)
		item.UId = mem.UId
		item.UGender = mem.Sex
		item.UUrl = mem.ImgUrl
		item.UName = mem.NickName
		item.ParnterLevel = -1
		item.IsLike = bLike
		pStatistics, ok := pStatisticsMap[item.UId]
		if ok {
			item.ChangeProfit = int64(pStatistics.Send * 100)
			item.Bwtimes = pStatistics.Bwtimes
			item.Playtimes = pStatistics.Playtimes
			item.Invalidtimes = pStatistics.Playtimes - pStatistics.Validtimes
		}
		// 上级队长信息
		if mem.Superior > 0 {
			dsmem := house.GetMemByUId(mem.Superior)
			if dsmem != nil {
				item.Superior = mem.Superior
				item.SuperiorName = dsmem.NickName
			} else {
				item.Superior = 0
				item.SuperiorName = ""
			}
		}

		// 搜索功能
		if req.SearchKey != "" {
			if strings.Contains(item.UName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", item.UId), req.SearchKey) {
				sortItem = append(sortItem, item)
			}
		} else {
			sortItem = append(sortItem, item)
		}
	}

	sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
		if item1.UId == optMem.UId {
			return true
		} else if item2.UId == optMem.UId {
			return false
		} else {
			if item1.UId < item2.UId {
				return true // 自己永远在最前面
			} else {
				return false
			}
		}
	}})

	var typeItems []*static.HouseNoLeagueStatisticsItem
	if !optMem.IsPartner() {
		for i := 0; i < len(sortItem); i++ {
			ackItem := sortItem[i]
			// 处理队长等级
			if hm, ok := hMems[ackItem.UId]; ok && !optMem.IsPartner() {
				if hm.IsPartner() {
					if hm.Superior > 0 {
						ackItem.ParnterLevel = 1
					} else {
						ackItem.ParnterLevel = 0
					}
				} else {
					ackItem.ParnterLevel = -1
				}
			}
			if req.PartnerLevel > -1 {
				if req.PartnerLevel == ackItem.ParnterLevel {
					typeItems = append(typeItems, ackItem)
				}
			} else {
				typeItems = append(typeItems, ackItem)
			}
		}
	} else {
		typeItems = append(typeItems, sortItem...)
	}

	for i := 0; i < len(typeItems); i++ {
		if i >= req.PBegin && i <= req.PEnd {
			ackItem := typeItems[i]
			ack.Items = append(ack.Items, ackItem)
			//capPlayTimes += ackItem.Playtimes
		}
	}
	if optMem.URole == 0 {
		num, err := GetDBMgr().SelectHouseMemberStatisticsWithCount(house.DBClub.Id, req.Fid, -1, begin, end)
		if err == nil {
			ack.Allplaytimes = num
		}
	}
	capPlayTimes := 0
	num2, err := GetDBMgr().SelectHouseMemberStatisticsWithCount(house.DBClub.Id, req.Fid, 0, begin, end)
	if err == nil {
		capPlayTimes = ack.Allplaytimes - num2
	}
	ack.Capplaytimes = capPlayTimes
	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseNoLeagueDetailStatistics(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseNoLeagueDetailStatistics)
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorTeamStatistical)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	//if !optMem.IsVitaminAdmin() && optMem.URole != constant.ROLE_CREATER && !optMem.IsPartner() && !optMem.IsVicePartner() {
	//	return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	//}

	// 校验时段、区间是否匹配
	timeRangeCnt := 1
	if req.QueryTimeInterval > 0 {
		timeRangeCnt = 24 / req.QueryTimeInterval
	}
	if req.QueryTimeRange > timeRangeCnt || req.QueryTimeRange <= 0 {
		return xerrors.InvalidTimeRangeError.Code, xerrors.InvalidTimeRangeError.Msg
	}

	// 筛选时间的起点和终点
	var begin, end time.Time
	nowTime := time.Now()
	zeroTime := static.GetZeroTime(nowTime.AddDate(0, 0, req.DayType))
	if req.QueryTimeInterval > 0 {
		begin = zeroTime.Add(time.Duration((req.QueryTimeRange-1)*req.QueryTimeInterval) * time.Hour)
		end = zeroTime.Add(time.Duration(req.QueryTimeRange*req.QueryTimeInterval) * time.Hour)
	} else {
		begin = zeroTime
		end = zeroTime.Add(24 * time.Hour)
	}

	ack := static.Msg_HC_NoLeagueStatistics{}
	ack.Items = []*static.HouseNoLeagueStatisticsItem{}
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd
	var sortItem []*static.HouseNoLeagueStatisticsItem

	// 获取所有成员
	hMems := house.GetMemberMap(false)
	// 统计查询结果
	var pStatisticsMap map[int64]models.QueryMemberStatisticsResult
	var selStsErr error
	pStatisticsMap = make(map[int64]models.QueryMemberStatisticsResult)

	if req.LowScoreFlag > 0 && (optMem.IsPartner() || optMem.IsVicePartner()) {
		pid := int64(0)
		if optMem.IsPartner() {
			pid = optMem.UId
		} else {
			pid = optMem.Partner
		}
		pStatisticsMap, selStsErr = GetDBMgr().SelectPartnerMemberStatisticsWithTotal(pid, house.DBClub.Id, req.Fid, begin, end)
	} else {
		pStatisticsMap, selStsErr = GetDBMgr().SelectHouseMemberStatisticsWithTotal(house.DBClub.Id, req.Fid, begin, end)
	}

	if selStsErr != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// leaveTime之前的记录为有效记录
	leaveTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", begin.Year(), begin.Month(), begin.Day(), begin.Hour(), begin.Minute(), begin.Second())
	leaveMemMap, err := GetDBMgr().SelectLeaveHousePartnerMember(house.DBClub.Id, req.Partner, leaveTime)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	sItemMap := make(map[int64]*static.HouseNoLeagueStatisticsItem)
	for _, mem := range hMems {
		if mem.Partner == req.Partner || mem.UId == req.Partner {
			item := new(static.HouseNoLeagueStatisticsItem)

			item.UId = mem.UId
			item.UGender = mem.Sex
			item.UUrl = mem.ImgUrl
			item.UName = mem.NickName
			item.IsLimit = GetDBMgr().Redis.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, house.DBClub.Id), mem.UId).Val()
			item.IsExit = false

			item.ChangeProfit = 0
			item.Bwtimes = 0
			item.Playtimes = 0
			item.Invalidtimes = 0
			sItemMap[mem.UId] = item
		}
	}

	for uid, _ := range leaveMemMap {
		if _, ok := sItemMap[uid]; ok {
			continue
		}

		item := new(static.HouseNoLeagueStatisticsItem)

		perdon, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if err == nil {
			item.UGender = perdon.Sex
			item.UUrl = perdon.Imgurl
			item.UName = perdon.Nickname
		}

		item.UId = uid
		item.IsLimit = false
		item.IsExit = true
		item.ChangeProfit = 0
		item.Bwtimes = 0
		item.Playtimes = 0
		item.Invalidtimes = 0
		sItemMap[uid] = item
	}

	for _, item := range sItemMap {
		itemS, ok := pStatisticsMap[item.UId]
		if ok {
			item.ChangeProfit = int64(itemS.TotalScore * 100)
			item.Bwtimes = itemS.BigWinTimes
			item.Playtimes = itemS.PlayTimes
			item.Invalidtimes = itemS.PlayTimes - itemS.ValidTimes
		}
	}

	ack.Total = len(sItemMap)
	// 搜索功能
	if req.SearchKey != "" {
		for _, item := range sItemMap {
			if strings.Contains(item.UName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", item.UId), req.SearchKey) {
				sortItem = append(sortItem, item)
			}
		}
	} else {
		for _, item := range sItemMap {
			sortItem = append(sortItem, item)
		}
	}

	sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
		if item1.UId == optMem.UId {
			return true
		} else if item2.UId == optMem.UId {
			return false
		} else {
			if item1.UId < item2.UId {
				return true // 自己永远在最前面
			} else {
				return false
			}
		}
	}})

	if err == nil {
		if req.SortType == static.SORT_PLAYTIMES_DES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.Playtimes > item2.Playtimes {
					return true
				} else if item1.Playtimes == item2.Playtimes {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})

		} else if req.SortType == static.SORT_PLAYTIMES_AES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.Playtimes < item2.Playtimes {
					return true
				} else if item1.Playtimes == item2.Playtimes {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_BWTIMES_DES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.Bwtimes > item2.Bwtimes {
					return true
				} else if item1.Bwtimes == item2.Bwtimes {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_BWTIMES_AES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.Bwtimes < item2.Bwtimes {
					return true
				} else if item1.Bwtimes == item2.Bwtimes {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_TOTALSCORE_DES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.ChangeProfit > item2.ChangeProfit {
					return true
				} else if item1.ChangeProfit == item2.ChangeProfit {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_TOTALSCORE_AES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.ChangeProfit < item2.ChangeProfit {
					return true
				} else if item1.ChangeProfit == item2.ChangeProfit {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_INVALIDROUND_DES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.Invalidtimes > item2.Invalidtimes {
					return true
				} else if item1.Invalidtimes == item2.Invalidtimes {
					return item1.UId > item2.UId
				} else {
					return false
				}
			}})
		} else if req.SortType == static.SORT_INVALIDROUND_AES {
			sort.Sort(static.HouseNoLeagueStatisticsItemWrapper{Item: sortItem, By: func(item1, item2 *static.HouseNoLeagueStatisticsItem) bool {
				if item1.Invalidtimes < item2.Invalidtimes {
					return true
				} else if item1.Invalidtimes == item2.Invalidtimes {
					return item1.UId < item2.UId
				} else {
					return false
				}
			}})
		}
	}

	for i := 0; i < len(sortItem); i++ {
		if i >= req.PBegin && i <= req.PEnd {
			ack.Items = append(ack.Items, sortItem[i])
		}
	}

	return xerrors.SuccessCode, &ack
}

// 获取包厢楼层分挡位支付房费
func Proto_ClubHouseFloorGearPayGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseId)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	paySlice := make(models.HouseFloorGearPaySlice, 0)
	if err := GetDBMgr().GetDBmControl().Where("hid = ?", house.DBClub.Id).Find(&paySlice).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	floorIds, _ := house.GetFloors()
	var ack static.MsgHouseFloorPay
	ack.Hid = house.DBClub.HId
	for i := 0; i < len(floorIds); i++ {
		fid := floorIds[i]
		floorPay := paySlice.FindByFid(fid)
		if floorPay == nil {
			if floor := house.GetFloorByFId(fid); floor != nil {
				floorPay = models.NewHouseFloorGearPay(house.DBClub.Id, fid, floor.Rule.PlayerNum)
			} else {
				return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
			}
		}
		ack.Items = append(ack.Items, floorPay.CovertModel())
	}
	return xerrors.SuccessCode, &ack
}

// 设置包厢楼层分挡位支付房费
func Proto_ClubHouseFloorGearPaySet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseFloorPay)
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	newModels := make(models.HouseFloorGearPaySlice, 0)
	for i := 0; i < len(req.Items); i++ {
		item := req.Items[i]
		floor := house.GetFloorByFId(item.Fid)
		if floor == nil {
			return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
		}
		mod := models.NewHouseFloorGearPay(house.DBClub.Id, item.Fid, floor.Rule.PlayerNum)
		if item.AA {
			mod.AAPay = item.AA
			if item.Gear1Cost == models.InvalidPay {
				mod.Gear1Cost = models.InvalidPay
			} else {
				mod.Gear1Cost = static.SwitchF64ToVitamin(item.Gear1Cost)
			}
		} else {
			mod.AAPay = item.AA
			mod.Gear2 = item.Gear2
			mod.Gear3 = item.Gear3
			mod.Gear4 = item.Gear4
			mod.Gear5 = item.Gear5
			mod.Gear6 = item.Gear6
			mod.Gear7 = item.Gear7
			mod.Gear8 = item.Gear8
			mod.Gear9 = item.Gear9
			mod.Gear10 = item.Gear10

			if item.Gear1Cost == models.InvalidPay {
				mod.Gear1Cost = models.InvalidPay
			} else {
				mod.Gear1Cost = static.SwitchF64ToVitamin(item.Gear1Cost)
			}

			if item.Gear2Under == models.InvalidPay {
				mod.Gear2Under = models.InvalidPay
			} else {
				mod.Gear2Under = static.SwitchF64ToVitamin(item.Gear2Under)
			}

			if item.Gear3Under <= 0 {
				mod.Gear3Under = models.InvalidPay
			} else {
				mod.Gear3Under = static.SwitchF64ToVitamin(item.Gear3Under)
			}
			if item.Gear3Cost == models.InvalidPay {
				mod.Gear3Cost = models.InvalidPay
			} else {
				mod.Gear3Cost = static.SwitchF64ToVitamin(item.Gear3Cost)
			}

			if item.Gear4Under <= 0 {
				mod.Gear4Under = models.InvalidPay
			} else {
				mod.Gear4Under = static.SwitchF64ToVitamin(item.Gear4Under)
			}
			if item.Gear4Cost == models.InvalidPay {
				mod.Gear4Cost = models.InvalidPay
			} else {
				mod.Gear4Cost = static.SwitchF64ToVitamin(item.Gear4Cost)
			}

			if item.Gear5Cost == models.InvalidPay {
				mod.Gear5Cost = models.InvalidPay
			} else {
				mod.Gear5Cost = static.SwitchF64ToVitamin(item.Gear5Cost)
			}
			if item.Gear5Under <= 0 {
				mod.Gear5Under = models.InvalidPay
			} else {
				mod.Gear5Under = static.SwitchF64ToVitamin(item.Gear5Under)
			}

			if item.Gear6Cost == models.InvalidPay {
				mod.Gear6Cost = models.InvalidPay
			} else {
				mod.Gear6Cost = static.SwitchF64ToVitamin(item.Gear6Cost)
			}
			if item.Gear6Under <= 0 {
				mod.Gear6Under = models.InvalidPay
			} else {
				mod.Gear6Under = static.SwitchF64ToVitamin(item.Gear6Under)
			}

			if item.Gear7Cost == models.InvalidPay {
				mod.Gear7Cost = models.InvalidPay
			} else {
				mod.Gear7Cost = static.SwitchF64ToVitamin(item.Gear7Cost)
			}
			if item.Gear7Under <= 0 {
				mod.Gear7Under = models.InvalidPay
			} else {
				mod.Gear7Under = static.SwitchF64ToVitamin(item.Gear7Under)
			}

			if item.Gear8Cost == models.InvalidPay {
				mod.Gear8Cost = models.InvalidPay
			} else {
				mod.Gear8Cost = static.SwitchF64ToVitamin(item.Gear8Cost)
			}
			if item.Gear8Under <= 0 {
				mod.Gear8Under = models.InvalidPay
			} else {
				mod.Gear8Under = static.SwitchF64ToVitamin(item.Gear8Under)
			}

			if item.Gear9Cost == models.InvalidPay {
				mod.Gear9Cost = models.InvalidPay
			} else {
				mod.Gear9Cost = static.SwitchF64ToVitamin(item.Gear9Cost)
			}
			if item.Gear9Under <= 0 {
				mod.Gear9Under = models.InvalidPay
			} else {
				mod.Gear9Under = static.SwitchF64ToVitamin(item.Gear9Under)
			}

			if item.Gear10Cost == models.InvalidPay {
				mod.Gear10Cost = models.InvalidPay
			} else {
				mod.Gear10Cost = static.SwitchF64ToVitamin(item.Gear10Cost)
			}
			if item.Gear10Under <= 0 {
				mod.Gear10Under = models.InvalidPay
			} else {
				mod.Gear10Under = static.SwitchF64ToVitamin(item.Gear10Under)
			}
		}
		if err := mod.Check(); err != nil {
			xlog.Logger().Error("item: %#v", item)
			return xerrors.ResultErrorCode, fmt.Sprintf("%d楼:%s", house.GetFloorIndexByFid(floor.Id)+1, err.Error())
		}
		newModels = append(newModels, mod)
	}

	tx := GetDBMgr().GetDBmControl().Begin()

	oldModels := make(models.HouseFloorGearPaySlice, 0)
	if err := tx.Where("hid = ?", house.DBClub.Id).Find(&oldModels).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Error(err)
			tx.Rollback()
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	}

	var (
		err    error
		memMap static.HouseMemberMap
	)
	for i := 0; i < len(newModels); i++ {
		now := newModels[i]
		before := oldModels.FindByFid(now.FId)
		var log *models.HouseFloorGearPayLog
		var beforeBaseCost int64 = -1
		if before == nil {
			if err = tx.Create(now).Error; err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
		} else {
			beforeBaseCost = before.BaseCost()
			if now.AAPay {
				if before.AAPay != now.AAPay || before.Gear1Cost != now.Gear1Cost {
					log = before.GenLog()
					before.AAPay = now.AAPay
					before.Gear1Cost = now.Gear1Cost
					if err = tx.Save(before).Error; err != nil {
						xlog.Logger().Error(err)
						tx.Rollback()
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
				}
			} else {
				if !now.Identical(&before.HouseFloorGearPayCore) {
					now.CreatedAt = before.CreatedAt
					log = before.GenLog()
					if err = tx.Save(now).Error; err != nil {
						xlog.Logger().Error(err)
						tx.Rollback()
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
				}
			}
		}

		if nowBaseCost := now.BaseCost(); nowBaseCost != beforeBaseCost {
			if memMap == nil {
				memMap = GetDBMgr().GetHouseMemMap(house.DBClub.Id)
			}
			err = UpdateTopPartnersTotal(tx, memMap, house.DBClub.Id, now.FId, nowBaseCost)
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
		}

		if log != nil {
			if err = tx.Create(log).Error; err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
		}
	}
	if err = tx.Commit().Error; err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	return xerrors.SuccessCode, nil
}

// 队长设置低分局数据
func Proto_PartnerSetLowScoreVal(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 解析协议内容
	req := data.(*static.Msg_CH_HouseValidRoundScoreSet)

	// 检查权限
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorSetLowScore)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 是否是队长身份
	mem := house.GetMemByUId(p.Uid)
	if !mem.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	for i := 0; i < len(req.Items); i++ {
		item := req.Items[i]
		if item.Score <= models.IgnorePay {
			continue
		}
		cusErr = house.PartnerSetLowScoreVal(item.FId, p.Uid, static.SwitchF64ToVitamin(item.Score))
		if cusErr != nil {
			xlog.Logger().Error(cusErr)
		}
	}

	return xerrors.SuccessCode, nil
}

// 获取队长低分局数据
func Proto_GetPartnerLowScoreVal(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 解析协议内容
	req := data.(*static.Msg_CH_HouseValidRoundScoreGet)

	// 检查权限
	house, _, _, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorSetLowScore)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	// 是否是队长身份
	mem := house.GetMemByUId(p.Uid)
	if !mem.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	var ack static.Msg_HC_HouseValidRoundScoreGet
	ack.HId = house.DBClub.HId
	ack.Items = make([]static.HouseValidRoundScoreGetItem, 0)
	ack.InvalidRound = 0
	floorLowScoreMap, err := GetDBMgr().SelectPartnerLowScoreVal(house.DBClub.Id, p.Uid)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	floorIds, _ := house.GetFloors()
	for i := 0; i < len(floorIds); i++ {
		var item static.HouseValidRoundScoreGetItem
		item.FId = floorIds[i]
		item.Score = models.IgnorePay
		if lowScoreVal, ok := floorLowScoreMap[item.FId]; ok {
			item.Score = static.SwitchVitaminToF64(lowScoreVal)
		}
		ack.Items = append(ack.Items, item)
	}

	return xerrors.SuccessCode, &ack
}

func Proto_ClubHouseTeamBan(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParterBan)
	house, _, optMember, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	if optMember.URole != consts.ROLE_CREATER && !optMember.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	partner := house.GetMemByUId(req.Pid)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	db := GetDBMgr().GetDBmControl()
	res, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id, partner.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	attr, ok := res[partner.UId]
	if !ok {
		attr = models.HousePartnerAttr{
			Id:   0,
			DHid: house.DBClub.Id,
			Uid:  partner.UId,
		}
	}
	if req.TeamBan != attr.TeamBan {
		if attr.Id > 0 { // 存在
			err = db.Model(&models.HousePartnerAttr{}).Where("id = ?", attr.Id).Update("team_ban", req.TeamBan).Error
		} else {
			attr.TeamBan = req.TeamBan
			err = db.Create(attr).Error
		}
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Code
		}
		var msg string
		if req.TeamBan {
			if optMember.URole == consts.ROLE_CREATER {
				msg = fmt.Sprintf("盟主将队长<color=#00A70C>%sID:%d</color>整队禁止", partner.NickName, partner.UId)
			} else {
				msg = fmt.Sprintf("<color=#F93030>%sID:%d</color>将队长<color=#00A70C>%sID:%d</color>整队禁止", optMember.NickName, optMember.UId, partner.NickName, partner.UId)
			}
		} else {
			if optMember.URole == consts.ROLE_CREATER {
				msg = fmt.Sprintf("盟主将队长<color=#00A70C>%sID:%d</color>取消整队禁止", partner.NickName, partner.UId)
			} else {
				msg = fmt.Sprintf("<color=#F93030>%sID:%d</color>将队长<color=#00A70C>%sID:%d</color>取消整队禁止", optMember.NickName, optMember.UId, partner.NickName, partner.UId)
			}
		}
		CreateClubMassage(house.DBClub.Id, optMember.UId, TeamBan, msg)
	}
	// 在线通知
	house.BroadcastMsg(consts.MsgTypeHouseTeamBan_ntf, req, func(houseMember *HouseMember) bool {
		if houseMember.URole == consts.ROLE_CREATER {
			return true
		}
		if houseMember.UId == partner.UId {
			return true
		}
		if houseMember.UId == partner.Superior {
			return true
		}
		return false
	})
	return xerrors.SuccessCode, nil
}

func Proto_ClubHousePartnerAlarmValueSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParterAlarmValueSet)
	if req.AlarmValue < 0 {
		xlog.Logger().Error("must > 0")
		return xerrors.ArgumentError.Code, "警戒值不能设置为负数。"
	}
	house, _, optMember, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	if optMember.URole != consts.ROLE_CREATER && !optMember.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	partner := house.GetMemByUId(req.Pid)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	db := GetDBMgr().GetDBmControl()
	res, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id, partner.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	attr, ok := res[partner.UId]
	if !ok {
		attr = models.HousePartnerAttr{
			Id:         0,
			DHid:       house.DBClub.Id,
			Uid:        partner.UId,
			AlarmValue: -1,
		}
	}
	if av := static.SwitchF64ToVitamin(req.AlarmValue); av != attr.AlarmValue {
		if attr.Id > 0 { // 存在
			err = db.Model(&models.HousePartnerAttr{}).Where("id = ?", attr.Id).Update("alarm_value", av).Error
		} else {
			attr.AlarmValue = av
			err = db.Create(attr).Error
		}
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Code
		}
		var msg string
		if optMember.URole == consts.ROLE_CREATER {
			msg = fmt.Sprintf("盟主将队长<color=#00A70C>%sID:%d</color>的警戒值设置为%0.2f", partner.NickName, partner.UId, req.AlarmValue)
		} else {
			msg = fmt.Sprintf("<color=#F93030>%sID:%d</color>将队长<color=#00A70C>%sID:%d</color>的警戒值设置为%0.2f", optMember.NickName, optMember.UId, partner.NickName, partner.UId, req.AlarmValue)
		}
		CreateClubMassage(house.DBClub.Id, optMember.UId, KickHouseMem, msg)
	}
	house.BroadcastMsg(consts.MsgTypePartnerAlarmValueSet_ntf, req, func(houseMember *HouseMember) bool {
		if houseMember.URole == consts.ROLE_CREATER {
			return true
		}
		if houseMember.UId == partner.UId {
			return true
		}
		if houseMember.UId == partner.Superior {
			return true
		}
		return false
	})
	return xerrors.SuccessCode, nil
}

func Proto_ClubHousePartnerAlarmValueGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParterAlarmValueSet)
	house, _, optMember, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	if optMember.URole != consts.ROLE_CREATER && !optMember.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	partner := house.GetMemByUId(req.Pid)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	res, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id, partner.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	attr, ok := res[partner.UId]
	if !ok {
		attr = models.HousePartnerAttr{
			Id:         0,
			DHid:       house.DBClub.Id,
			Uid:        partner.UId,
			AlarmValue: -1,
		}
	}
	if attr.AlarmValue >= 0 {
		req.AlarmValue = static.SwitchVitaminToF64(attr.AlarmValue)
	} else {
		req.AlarmValue = -1
	}
	return xerrors.SuccessCode, req
}

func Proto_ClubHousePartnerAACostSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParterAACostSet)
	house, _, optMember, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	if optMember.URole != consts.ROLE_CREATER && !optMember.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	partner := house.GetMemByUId(req.Pid)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	db := GetDBMgr().GetDBmControl()
	res, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id, partner.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	attr, ok := res[partner.UId]
	if !ok {
		attr = models.HousePartnerAttr{
			Id:   0,
			DHid: house.DBClub.Id,
			Uid:  partner.UId,
		}
	}
	if req.AA != attr.AA {
		if attr.Id > 0 { // 存在
			err = db.Model(&models.HousePartnerAttr{}).Where("id = ?", attr.Id).Update("aa", req.AA).Error
		} else {
			attr.AA = req.AA
			err = db.Create(attr).Error
		}
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Code
		}
		var msg string
		if req.AA {
			if optMember.URole == consts.ROLE_CREATER {
				msg = fmt.Sprintf("盟主将队长<color=#00A70C>%sID:%d</color>开启AA扣卡", partner.NickName, partner.UId)
			} else {
				msg = fmt.Sprintf("<color=#F93030>%sID:%d</color>将队长<color=#00A70C>%sID:%d</color>开启AA扣卡", optMember.NickName, optMember.UId, partner.NickName, partner.UId)
			}
		} else {
			if optMember.URole == consts.ROLE_CREATER {
				msg = fmt.Sprintf("盟主将队长<color=#00A70C>%sID:%d</color>关闭AA扣卡", partner.NickName, partner.UId)
			} else {
				msg = fmt.Sprintf("<color=#F93030>%sID:%d</color>将队长<color=#00A70C>%sID:%d</color>关闭AA扣卡", optMember.NickName, optMember.UId, partner.NickName, partner.UId)
			}
		}
		CreateClubMassage(house.DBClub.Id, optMember.UId, AACost, msg)
	}
	house.BroadcastMsg(consts.MsgTypePartnerAACostSet_ntf, req, func(houseMember *HouseMember) bool {
		if houseMember.URole == consts.ROLE_CREATER {
			return true
		}
		if houseMember.UId == partner.UId {
			return true
		}
		if houseMember.UId == partner.Superior {
			return true
		}
		return false
	})
	return xerrors.SuccessCode, nil
}

func Proto_ClubHouseTeamKick(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseMemberKick)
	house, _, owner, xer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_CREATER, MinorRightNull)
	if xer != xerrors.RespOk {
		return xer.Code, xer.Msg
	}

	kickPartner := house.GetMemByUId(req.UId)
	if kickPartner == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseError.Msg
	}

	if owner.Lower(kickPartner.URole) {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if !kickPartner.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHouseMemberError.Msg
	}

	if _, ok := kickPartner.CheckRef(); ok {
		xer = xerrors.NewXError("该队长为合并包厢盟主，无法被踢出包厢。")
		return xer.Code, xer.Msg
	}

	successor := house.GetMemByUId(kickPartner.Superior)
	if successor == nil {
		successor = owner
	}

	clbMembers := house.GetMemSimple(false)
	team := []HouseMember{*kickPartner}
	nextSuperiors := []int64{kickPartner.UId}
	for len(nextSuperiors) > 0 {
		var juniors []int64
		for _, kickMember := range clbMembers {
			if kickMember.IsJunior() && static.In64(nextSuperiors, kickMember.Superior) {
				juniors = append(juniors, kickMember.UId)
				team = append(team, kickMember)
			} else if static.In64(nextSuperiors, kickMember.Partner) {
				team = append(team, kickMember)
			}
		}
		nextSuperiors = juniors
	}
	var (
		vitaminOffset int64
		err           error
	)
	tx := GetDBMgr().GetDBmControl().Begin()
	defer func(clb *Club, team []HouseMember) {
		if static.TxCommit(tx, err) {
			floorsVipSet := make(map[int64]map[int64]struct{})
			for _, floor := range clb.Floors {
				floorsVipSet[floor.Id] = floor.GetVipUsersSet()
			}
			for _, kickMember := range team {
				//踢出玩家之后 要踢出VIP楼层
				for fid, vipUser := range floorsVipSet {
					_, ok := vipUser[kickMember.UId]
					if ok {
						err := RemVipUsers(fid, kickMember.UId)
						if err != nil {
							xlog.Logger().Error(err)
						}
					}
				}
			}
		}
	}(house, team)

	for _, kickMember := range team {
		_, xer = kickMember.ClearVitaminOnKick(owner.UId, models.AdminSend, house, tx)
		if xer != nil {
			xlog.Logger().Error(xer.Error())
			return xer.Code, xer.Msg
		}
		vitaminOffset += kickMember.UVitamin
	}

	if vitaminOffset != 0 {
		var after int64
		_, after, err = successor.VitaminIncrement(owner.UId, vitaminOffset, models.AdminSend, tx)
		if err != nil {
			xlog.Logger().Error(err)
			return 0, xerrors.DBExecError
		}
		err = GetDBMgr().UpdateVitaminMgrList(house.DBClub.Id, successor.UId, after, tx)
		if err != nil {
			xlog.Logger().Error(err)
			return 0, xerrors.DBExecError
		}
	}

	CreateClubMassageWithTx(
		tx,
		house.DBClub.Id,
		owner.UId,
		TeamKick,
		fmt.Sprintf(
			"%s,玩家<color=#00A70C>%sID:%d</color>被整队踢出包厢，名下%0.2f分继承给玩家<color=#F93030>%sID:%d</color>",
			time.Now().Format("2006/01/02 15:04:05"),
			kickPartner.NickName,
			kickPartner.UId,
			static.SwitchVitaminToF64(vitaminOffset),
			successor.NickName,
			successor.UId,
		),
	)

	//xerr := house.PoolChange(owner.UId, models.AdminSend, vitaminOffset, tx)
	//if xerr != nil { 并不是归还给pool，是继承给上级
	//	err = fmt.Errorf(xerr.Error())
	//	xlog.Logger().Errorf("%v", err)
	//	return xerr.Code, xerr.Msg
	//}
	for _, kickMember := range team {
		kickMember.SendMsg(consts.MsgTypeHouseTeamKick_ntf, &static.Msg_HC_HouseKick_Ntf{
			HId:   house.DBClub.HId,
			OptId: owner.UId,
			Pid:   req.UId,
		})
		xer = house.MemDelete(kickMember.UId, false, tx)
		if xer != nil {
			err = fmt.Errorf(xer.Error())
			xlog.Logger().Errorf("%v", err)
			return xer.Code, xer.Msg
		}
	}

	return xerrors.SuccessCode, nil
}

func Proto_ClubHousePartnerRewardSet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParterRawardSet)
	if req.Reward < 0 || req.Reward > 100 {
		xlog.Logger().Error("must 100 >= x >= 0")
		return xerrors.ArgumentError.Code, "必须设置0~100之间的数字"
	}
	house, _, optMember, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}
	if optMember.URole != consts.ROLE_CREATER && !optMember.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	partner := house.GetMemByUId(req.Pid)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	if !partner.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}
	if (optMember.Upper(consts.ROLE_MEMBER) && partner.Superior != 0) ||
		(optMember.IsPartner() && partner.Superior != optMember.UId) {
		return xerrors.InvalidPermission.Code, "只允许直属上级配置收益。"
	}
	db := GetDBMgr().GetDBmControl()
	res, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id, partner.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	attr, ok := res[partner.UId]
	if !ok {
		attr = models.HousePartnerAttr{
			Id:             0,
			DHid:           house.DBClub.Id,
			Uid:            partner.UId,
			RewardSuperior: -1,
		}
	}
	attr.RewardSuperior = req.Reward
	if attr.Id > 0 { // 存在
		err = db.Model(&models.HousePartnerAttr{}).Where("id = ?", attr.Id).Update("reward_superior", attr.RewardSuperior).Error
	} else {
		err = db.Create(attr).Error
	}
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Code
	}
	var msg string
	if optMember.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主将队长<color=#00A70C>%sID:%d</color>的赛点比例设置为%d%%", partner.NickName, partner.UId, req.Reward)
	} else {
		msg = fmt.Sprintf("<color=#F93030>玩家%sID:%d</color>将其下级队长<color=#00A70C>%sID:%d</color>的赛点比例设置为%d%%", optMember.NickName, optMember.UId, partner.NickName, partner.UId, req.Reward)
	}
	CreateClubMassage(house.DBClub.Id, optMember.UId, GameReward, msg)
	house.BroadcastMsg(consts.MsgTypePartnerRewardSet_ntf, req, func(houseMember *HouseMember) bool {
		if houseMember.URole < consts.ROLE_MEMBER {
			return true
		}
		if houseMember.UId == partner.UId {
			return true
		}
		if houseMember.UId == partner.Superior {
			return true
		}
		return false
	})
	return xerrors.SuccessCode, nil
}

func Proto_ClubHousePartnerRewardGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_HouseParterRawardSet)
	house, _, optMem, cer := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cer != xerrors.RespOk {
		return cer.Code, cer.Msg
	}

	if optMem.URole != consts.ROLE_CREATER && !optMem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	partner := house.GetMemByUId(req.Pid)
	if partner == nil {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	res, err := GetDBMgr().SelectHouseAllPartnerAttr(house.DBClub.Id, partner.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Code
	}

	attr, ok := res[partner.UId]
	if !ok {
		attr = models.HousePartnerAttr{
			Id:             0,
			DHid:           house.DBClub.Id,
			Uid:            partner.UId,
			RewardSuperior: -1,
		}
	}
	req.Reward = attr.RewardSuperior
	//if optMem.IsPartner() {
	//	req.Reward = attr.RewardSuperior
	//}
	return xerrors.SuccessCode, req
}

// 队长或盟主查看队长收益统计
func Proto_ClubHousePartnerRewardStatistic(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req, ok := data.(*static.Msg_CH_HousePartnerRewardStatistic)
	if !ok {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}

	// 获取包厢
	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)

	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if optmem.URole != consts.ROLE_CREATER && optmem.URole != consts.ROLE_ADMIN && !optmem.IsVitaminAdmin() && !optmem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	var partnerView bool
	if optmem.IsPartner() {
		partnerView = true
	}

	zeroTime := static.GetZeroTime(time.Now().AddDate(0, 0, req.SelectTime))
	selectTime1 := zeroTime
	selectTime2 := zeroTime.Add(24 * time.Hour)

	rewardMap, _, _, err := GetDBMgr().SelectHouseAllPartnerReward(house.DBClub.Id, -1, &selectTime1, &selectTime2)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	houseMemberMap := house.GetMemberMap(false)
	conItems := make([]static.MsgRewardStatisticItem, 0)

	for _, mem := range houseMemberMap {
		var add bool
		if partnerView {
			if mem.UId == p.Uid || (mem.IsJunior() && mem.Superior == p.Uid) {
				add = true
			}
		} else {
			if mem.UId == house.DBClub.UId || mem.IsPartner() {
				add = true
			}
		}
		if add {
			item := static.MsgRewardStatisticItem{
				UId:        mem.UId,
				UName:      mem.NickName,
				UUrl:       mem.ImgUrl,
				UGender:    mem.Sex,
				CurVitamin: static.SwitchVitaminToF64(mem.UVitamin),
			}
			rewardDetail, ok := rewardMap[item.UId]
			if ok {
				item.CurReward = static.SwitchVitaminToF64(rewardDetail.Current)
			}
			conItems = append(conItems, item)
		}
	}

	var searchItems []static.MsgRewardStatisticItem
	// 搜索功能
	if req.SearchKey != "" {
		for _, sItem := range conItems {
			if strings.Contains(sItem.UName, req.SearchKey) || strings.Contains(fmt.Sprintf("%d", sItem.UId), req.SearchKey) {
				searchItems = append(searchItems, sItem)
			}
		}
	} else {
		searchItems = append(searchItems, (conItems)[0:]...)
	}

	if req.SortType == static.SORT_REWARDVITAMIN_DESC {
		sort.Sort(static.MsgRewardStatisticItemWrapper{Item: searchItems, By: func(item1, item2 *static.MsgRewardStatisticItem) bool {
			if item1.CurVitamin > item2.CurVitamin {
				return true
			} else if item1.CurVitamin == item2.CurVitamin {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if req.SortType == static.SORT_REWARDVITAMIN_ASC {
		sort.Sort(static.MsgRewardStatisticItemWrapper{Item: searchItems, By: func(item1, item2 *static.MsgRewardStatisticItem) bool {
			if item1.CurVitamin < item2.CurVitamin {
				return true
			} else if item1.CurVitamin == item2.CurVitamin {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	} else if req.SortType == static.SORT_REWARD_DESC {
		sort.Sort(static.MsgRewardStatisticItemWrapper{Item: searchItems, By: func(item1, item2 *static.MsgRewardStatisticItem) bool {
			if item1.CurReward > item2.CurReward {
				return true
			} else if item1.CurReward == item2.CurReward {
				return item1.UId > item2.UId
			} else {
				return false
			}
		}})
	} else if req.SortType == static.SORT_REWARD_ASC {
		sort.Sort(static.MsgRewardStatisticItemWrapper{Item: searchItems, By: func(item1, item2 *static.MsgRewardStatisticItem) bool {
			if item1.CurReward < item2.CurReward {
				return true
			} else if item1.CurReward == item2.CurReward {
				return item1.UId < item2.UId
			} else {
				return false
			}
		}})
	}

	var ack static.Msg_HC_HouseRewardStatistic
	ack.At = time.Now().Unix()
	ack.PBegin = req.PBegin
	ack.PEnd = req.PEnd

	var idxBeg int = req.PBegin
	var idxEnd int = req.PEnd
	if idxEnd > len(searchItems) {
		idxEnd = len(searchItems)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}
	ack.Items = searchItems[idxBeg:idxEnd]
	return xerrors.SuccessCode, &ack
}

func Proto_ClubPartnerClearReward(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_CH_HouseId)
	if !ok {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}
	house, _, optmem, cusErr := inspectClubFloorMemberWithRight(req.HId, -1, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if optmem.URole != consts.ROLE_CREATER && !optmem.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	rewardMap, pk, dpk, err := GetDBMgr().SelectHouseAllPartnerReward(house.DBClub.Id, -1, nil, nil, optmem.UId)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	rewardDetail, ok := rewardMap[optmem.UId]
	if ok {
		if rewardDetail.Current > 0 {
			tx := GetDBMgr().GetDBmControl().Begin()
			var mod models.PartnerRewardT
			if len(pk) > 0 {
				err = tx.Model(mod).Where("id in(?)", pk).Update("cleared_time", time.Now().Unix()).Error
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
			}
			if len(dpk) > 0 {
				err = tx.Delete(mod, "id in(?)", dpk).Error
				if err != nil {
					xlog.Logger().Error(err)
					tx.Rollback()
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
			}
			_, after, err := optmem.VitaminIncrement(optmem.UId, rewardDetail.Current, models.GameReward, tx)
			if err != nil {
				//修改疲劳值统计管理节点信息
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
			err = GetDBMgr().UpdateVitaminMgrList(house.DBClub.Id, optmem.UId, after, tx)
			if err != nil {
				//修改疲劳值统计管理节点信息
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
			err = tx.Commit().Error
			if err != nil {
				//修改疲劳值统计管理节点信息
				xlog.Logger().Error(err)
				tx.Rollback()
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
		} else {
			xlog.Logger().Warnf("%d reward detail current < 0 = %#v", optmem.UId, rewardDetail)
		}
	} else {
		xlog.Logger().Warn("no reward detail of", optmem.UId)
	}
	return xerrors.SuccessCode, nil
}
