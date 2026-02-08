package center

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
)

type HouseFloorProtocolHandler func(*HouseFloor, *Session, *static.Person, interface{}) (int16, interface{})

//////////////////////////////////////////////////////////////
//! 协议管理者

var houseFloorProtocolmgrSingleton *HouseFloorProtocolMgr = nil

// ! 得到包厢管理单例
func GetHouseFloorProtocolMgr() *HouseFloorProtocolMgr {
	if houseFloorProtocolmgrSingleton == nil {

		houseFloorProtocolmgrSingleton = new(HouseFloorProtocolMgr)
		houseFloorProtocolmgrSingleton.funcProtocol = make(map[string]HouseFloorProtocolHandler)

		hfProtocol := houseFloorProtocolmgrSingleton

		// 解散牌桌
		hfProtocol.RegisterMessage(consts.HFChOptTableDel_Ntf, hfProtocol.ChOpt_TableDel_Ntf)
		// 桌子信息变更推送
		hfProtocol.RegisterMessage(consts.HFCHOptTableUpdate, hfProtocol.Chopt_TableUpdate)
		//
		hfProtocol.RegisterMessage(consts.HFCHOptTableStepUpdate, hfProtocol.Chopt_TableStepUpdate)
	}

	return houseFloorProtocolmgrSingleton
}

type HouseFloorProtocolInfo struct {
	Header      string
	ChOptHeader string
	Session     *Session
	Person      *static.Person
	Data        interface{}
}

type HouseFloorProtocolMgr struct {
	funcProtocol map[string]HouseFloorProtocolHandler
}

// 注册协议 对象 处理函数
func (self *HouseFloorProtocolMgr) RegisterMessage(choptheader string, protocol HouseFloorProtocolHandler) {
	self.funcProtocol[choptheader] = protocol
}

func (self *HouseFloorProtocolMgr) GetProtocol(choptheader string) HouseFloorProtocolHandler {
	return self.funcProtocol[choptheader]
}

func (self *HouseFloorProtocolMgr) CallGame(gameid int, uid int64, header string, errcode int16, data interface{}) error {

	bytes, err := GetServer().CallGame(gameid, uid, "NewServerMsg", header, errcode, data)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	bytestr := string(bytes)
	if bytestr != "SUC" {
		xlog.Logger().Errorln(fmt.Sprintf("header:%s %s ", header, " is not accept"))
		return errors.New(xerrors.UnAcceptRpcError.Msg)
	}

	return nil
}

// ! 修改楼层玩法
func ChOptFloorRuleModify(floor *HouseFloor, s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	house := GetClubMgr().GetClubHouseByHId(floor.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
	}

	req := data.(*static.Msg_CH_HouseFloorRuleModify)

	// 规则合法判定
	var msgct static.Msg_CreateTable

	msgct.Gps = true
	msgct.Voice = true
	msgct.GVoiceOk = true

	msgct.KindId = req.FRule.KindId
	msgct.PlayerNum = req.FRule.PlayerNum
	msgct.RoundNum = req.FRule.RoundNum
	msgct.CostType = req.FRule.CostType
	msgct.Restrict = req.FRule.Restrict
	msgct.GameConfig = req.FRule.GameConfig
	msgct.GVoice = req.FRule.GVoice

	config, custerr := validateCreateTableParam(&msgct, true)
	if custerr != nil {
		return custerr.Code, custerr.Msg
	}

	req.FRule.KindId = config.KindId
	req.FRule.PlayerNum = config.MaxPlayerNum
	req.FRule.RoundNum = config.RoundNum
	req.FRule.CostType = config.CostType
	req.FRule.Restrict = "false"
	if config.Restrict {
		req.FRule.Restrict = "true"
	}
	req.FRule.GameConfig = config.GameConfig
	floor.Rule = req.FRule

	err := GetDBMgr().HouseFloorUpdate(floor)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	//修改每桌人数
	floor.ModifUserNum()

	ntf := new(static.Ntf_HC_HouseFloorKIdModify)
	ntf.HId = floor.HId
	ntf.FId = floor.Id
	ntf.FRule = req.FRule
	ntf.VitaminLowLimit = static.SwitchVitaminToF64(floor.VitaminLowLimit)
	ntf.VitaminHighLimit = static.SwitchVitaminToF64(floor.VitaminHighLimit)
	area := GetAreaGameByKid(floor.Rule.KindId)
	if area != nil {
		ntf.ImageUrl = area.Icon
		ntf.KindName = area.Name
		ntf.PackageName = area.PackageName
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorRuleModify_Ntf, ntf)
	msg := fmt.Sprintf("盟主修改了包厢%s-%s(%d)玩法", ntf.KindName, ntf.PackageName, floor.Id)
	CreateClubMassage(house.DBClub.Id, p.Uid, FloorRuleChange, msg)
	return xerrors.SuccessCode, nil
}

// ! 进入包厢楼层
func ChOptMemInV2(floor *HouseFloor, s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 包厢
	house := GetClubMgr().GetClubHouseByHId(floor.HId)
	if house == nil {
		return xerrors.InValidHouseError.Code, xerrors.InValidHouseFloorError.Msg
	}

	// 成员
	mem := house.GetMemByUId(p.Uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	// 活跃数据
	floor.MemIn(mem)
	// 玩家数据
	p.HouseId = floor.HId
	p.FloorId = floor.Id
	GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", p.HouseId, "FloorId", p.FloorId)
	// 历史登陆数据
	mem.HId = floor.HId
	mem.DHId = floor.DHId
	mem.FId = floor.Id
	GetDBMgr().HouseMemberUpdate(mem.DHId, mem)

	acks := make([]*static.Msg_HC_HouseMemberIn, 0)

	var mixFloor []*HouseFloor
	if house.DBClub.MixActive && floor.IsMix {
		for _, floor := range house.Floors {
			if floor.IsMix {
				mixFloor = append(mixFloor, floor)
			}
		}
		if sortTableByCreateTime(mixFloor) {
			xlog.Logger().Warnf("sortTableByCreateTime after ChOptMemInV2")
			if time.Now().Unix()-house.LastSyncSorted > 5 {
				house.SyncTablesWithSorted()
			}
		}
		for _, floor := range mixFloor {
			ack := floor.BuildAck(floor.GetAllTables(), false)
			if ack == nil {
				ack = GetFloorDetail(floor)
			}
			acks = append(acks, ack)
			floor.SaveHft()
		}
	} else {
		ack := floor.BuildAck(floor.GetAllTables(), false)
		if ack == nil {
			ack = GetFloorDetail(floor)
		}
		acks = append(acks, ack)
	}
	go GetClubMgr().UserInToHouse(p, floor.HId)
	return xerrors.SuccessCode, &acks
}

// ! 离开包厢楼层
func ChOptMemOut(floor *HouseFloor, s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 包厢
	house := GetClubMgr().GetClubHouseByHId(floor.HId)
	if house == nil {
		return xerrors.SuccessCode, nil
	}

	// 删除活跃数据
	floor.SafeMemOut(p.Uid)

	// 当前玩家数据
	person := GetPlayerMgr().GetPlayer(p.Uid)
	if person == nil {
		return xerrors.SuccessCode, nil
	}

	if person.Info.HouseId == floor.HId && person.Info.FloorId == floor.Id {
		// 内存数据
		person.Info.HouseId = 0
		person.Info.FloorId = 0
		// 清理缓存数据
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", 0, "FloorId", 0)
	}

	return xerrors.SuccessCode, nil
}

func HouseTableWatch(hid int, fid int64, ntid int, p *static.Person) (bool, *static.Msg_S2C_TableIn) {
	house := GetClubMgr().GetClubHouseByHId(hid)
	if house == nil {
		return false, nil
	}
	floor := house.GetFloorByFId(fid)
	if floor == nil {
		return false, nil
	}

	hft := floor.GetTableByNTId(ntid)
	if hft == nil {
		return false, nil
	}

	table := GetTableMgr().GetTable(hft.TId)
	if table == nil {
		return false, nil
	}

	con := GetServer().GetGameConfig(table.KindId)
	if con == nil {
		return false, nil
	}

	if !con.IsSupportWatch {
		return false, nil
	}

	//游戏桌子不支持旁观时大厅服务器需要拦截，不然大厅客户端界面会显示有玩家在旁观
	if strings.Contains(table.Config.GameConfig, "LookonSupport") {
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(table.Config.GameConfig), &dat); err == nil {
			if dat["LookonSupport"] == "false" {
				return false, nil
			}
		}
	}

	if seatId, _ := table.GetSeat(p.Uid); seatId != -1 {
		return false, nil
	}

	if (table.Begin || table.GetEmptySeat() == -1) && len(hft.Watchers) < 20 {
		gameserver := GetServer().GetGame(table.GameId)
		if gameserver == nil {
			return false, nil
		}

		// 桌子添加观战人员
		hft.AddWatcher(p.Uid)
		// 桌子添加观战人员
		GetDBMgr().GetDBrControl().AddWatchPlayerToTable(table.Id, p.Uid)

		var ack static.Msg_S2C_TableIn
		ack.Id = table.Id
		ack.GameId = table.GameId
		ack.KindId = table.KindId
		ack.Ip = gameserver.ExIp
		ack.PkgName = GetAreaPackageKeyByKid(table.KindId)
		return true, &ack
	}

	return false, nil
}

func HouseTableIn(gpsInfo *static.GpsInfo, p *static.Person, req *static.Msg_CH_HouseTableIn, priorityUsers ...int64 /*指定匹配优先级比较高的人*/) (code int16, v interface{}) {
	// 数据结构
	if !atomic.CompareAndSwapInt64(&p.IsJoin, 0, 1) {
		return xerrors.HouseFloorTableJoiningError.Code, xerrors.HouseFloorTableJoiningError.Msg
	}
	defer func() {
		p.IsJoin = 0
	}()
	house, floor, mem, cusErr := inspectClubFloorMemberWithRight(req.HId, req.FId, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}
	if house.DBClub.IsFrozen {
		return xerrors.IsFrozenError.Code, xerrors.IsFrozenError.Msg
	}
	if p.TableId != 0 {
		table := GetTableMgr().GetTable(p.TableId)
		if table == nil {
			return xerrors.InValidHouseTableError.Code, xerrors.InValidHouseTableError.Msg
		}
		//获取游戏服务器
		gameserver := GetServer().GetGame(table.GameId)
		if gameserver == nil {
			// 获取游戏服失败
			return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
		}

		var ack static.Msg_S2C_TableIn
		ack.Id = table.Id
		ack.GameId = table.GameId
		ack.KindId = table.KindId
		ack.Ip = gameserver.ExIp
		ack.PkgName = GetAreaPackageKeyByKid(table.KindId)
		return xerrors.SuccessCode, &ack
	}

	// if floor.Rule.KindId != req.KindID {
	// 	xerr := xerrors.NewXError("楼层玩法已变更，请稍后重试。")
	// 	return xerr.Code, xerr.Msg
	// }

	// 如果玩家是重新加入，这里清掉玩家的所有旧数据
	ResetUserCacheData(p)

	// 大厅是否维护
	if notice := CheckServerMaintainWithWhite(p.Uid, static.NoticeMaintainServerAllServer); notice != nil {
		// return xerrors.ServerMaintainError.Code, xerrors.ServerMaintainError.Msg
		return xerrors.ServerMaintainError.Code, static.HF_JtoA(notice)
	}

	if floor.IsVipFloor() {
		if !floor.IsVipUser(mem.UId) {
			return xerrors.HouseFloorJoinVipError.Code, xerrors.HouseFloorJoinVipError.Msg
		}
	}

	// 疲劳值下限判定
	if house.DBClub.IsVitamin && floor.IsVitamin {
		if floor.VitaminLowLimit != consts.VitaminInvalidValueSrv && mem.UVitamin < floor.VitaminLowLimit {
			return xerrors.VitaminNotEnoughError.Code, xerrors.VitaminNotEnoughError.Msg
		}
		if floor.VitaminHighLimit > 0 && floor.VitaminHighLimit != consts.VitaminInvalidValueSrv && mem.UVitamin >= floor.VitaminHighLimit {
			return xerrors.VitaminTooEnoughError.Code, xerrors.VitaminTooEnoughError.Msg
		}
	}

	if house.DBClub.OnlyQuickJoin && req.NTId != -1 {
		return xerrors.OnlyQucikJoinError.Code, xerrors.OnlyQucikJoinError.Msg
	}

	if mem.Ref > 0 {
		hml := new(models.HouseMergeLog)
		if err := GetDBMgr().GetDBmControl().
			Where("devourer = ?", house.DBClub.Id).
			Where("swallowed = ?", mem.Ref).First(hml).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				xlog.Logger().Errorf("该成员通过合并包厢%d加入，但未找到合并包厢记录", mem.Ref)
			}
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		if hml.MergeState == models.HouseMergeStateRevoking {
			xerr := xerrors.NewXError("解除合并包厢中，您暂时不能游戏，请联系盟主或者管理员。")
			return xerr.Code, xerr.Msg
		}
	}

	msg := new(static.Msg_HouseTableInOpt)
	param := static.Msg_CH_HouseTableIn{}
	param.FId = floor.Id
	param.HId = floor.HId
	param.NTId = req.NTId
	param.Gps = req.Gps
	param.Voice = req.Voice
	param.GVoiceOk = req.GVoiceOk
	param.IgnoreRule = req.IgnoreRule
	param.RestartId = req.RestartId
	param.KindID = req.KindID
	msg.Param = param
	msg.Header = consts.MsgTypeHouseTableIn
	return ChOptClubHouseTableIn(house, floor, mem, gpsInfo, p, msg, priorityUsers...)
}

// 玩家进入牌桌
func ChOptClubHouseTableIn(clubHouse *Club, floor *HouseFloor, mem *HouseMember,
	gpsInfo *static.GpsInfo,
	p *static.Person,
	msg *static.Msg_HouseTableInOpt, priorityUsers ...int64) (code int16, v interface{}) {
	req := msg.Param
	if len(mem.NoFloors) > 0 {
		if static.In64(static.AtoI64s(mem.NoFloors), floor.Id) {
			return xerrors.ResultErrorCode, "您暂无法进入该楼层玩法，请联系管理员。"
		}
	}

	if clubHouse.DBClub.OnlyQuickJoin && req.NTId != -1 {
		return xerrors.OnlyQucikJoinError.Code, xerrors.OnlyQucikJoinError.Msg
	}
	// 获取玩法配置
	config := GetServer().GetGameConfig(floor.Rule.KindId)
	if config == nil {
		// 没有找到玩法对应的配置
		return xerrors.ResultErrorCode, "不支持的游戏玩法"
	}
	if floor.Rule.KindId != req.KindID {
		xlog.Logger().Errorf("floor:", floor.Rule.KindId, req.KindID)
		return xerrors.ResultErrorCode, "楼层玩法已变更，请重试1。"
	}
	areaGame := GetAreaGameByKid(floor.Rule.KindId)
	if areaGame == nil {
		// 没有找到玩法对应的区域游戏信息
		return xerrors.ResultErrorCode, "该玩法已下线"
	}

	// 玩法版本更新
	if floor.Rule.Version != areaGame.ForcedVersion {
		xlog.Logger().Warnf("houseTablein error: floor game rule need update. curVer:%d, latest:%d.",
			floor.Rule.Version, areaGame.ForcedVersion)
		return xerrors.HouseFloorRuleChangeStrongError.Code, xerrors.HouseFloorRuleChangeStrongError.Msg
	}

	//小队打烊了
	if mem.URole != 0 && !mem.IsPartner() {
		tempUid := mem.Partner
		if mem.Partner == 0 {
			tempUid = clubHouse.DBClub.UId
		}
		tempstr := GetDBMgr().GetDBrControl().RedisV2.HGet(fmt.Sprintf(consts.REDIS_KEY_OFFWORK, clubHouse.DBClub.Id), fmt.Sprintf("%d", tempUid)).Val()
		var tempBool bool
		tempBool, _ = strconv.ParseBool(tempstr)
		if tempBool {
			xerr := xerrors.NewXError("您暂时不能游戏，详情请联系盟主或管理员")
			return xerr.Code, xerr.Msg
		}
	}

	//这里判断是否被禁止娱乐
	if GetDBMgr().Redis.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clubHouse.DBClub.Id), p.Uid).Val() {
		return xerrors.LimitGameError.Code, xerrors.LimitGameError.Msg
	}

	var (
		partnerLink []int64                // 合伙人链表
		memMap      map[int64]HouseMember  // 成员集合
		payer       = clubHouse.DBClub.UId // 房卡谁买单
	)
	if mem.Lower(consts.ROLE_ADMIN) {
		memMap = clubHouse.GetMemberMap(false)
		partnerLink = clubHouse.GetPartnerLinkByUid(memMap, mem.UId)
	} else {
		mem.Logger().Info("管理员(圈主)入座不检查整队禁止/警戒值/AA支付")
	}
	mem.Logger().Infof("table in, partner link = %v", partnerLink)
	if len(partnerLink) > 0 {
		partnerAttrs, err := GetDBMgr().SelectHouseAllPartnerAttr(clubHouse.DBClub.Id, partnerLink...)
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		// 全队禁止
		for pid, attr := range partnerAttrs {
			if attr.TeamBan {
				mem.Logger().Warnf("partner=%d, ban, can't table in", pid)
				return xerrors.PartnerLinkBanError.Code, xerrors.PartnerLinkBanError.Msg
			}
		}
		// 警戒值检查
		if clubHouse.DBClub.IsVitamin && floor.IsVitamin {
			for _, pid := range partnerLink {
				if attr, ok := partnerAttrs[pid]; ok && attr.AlarmValue >= 0 {
					sumVitamin := clubHouse.GetPartnerMembersSumVitamin(memMap, pid)
					alarm := sumVitamin <= attr.AlarmValue
					mem.Logger().Warnf("队长ID:%d, 警戒值:%d, 名下玩家疲劳值总和:%d, 触发?[%t]",
						pid, attr.AlarmValue, sumVitamin, alarm)
					if alarm {
						return xerrors.AlarmValueError.Code, xerrors.AlarmValueError.Msg
					}
				} else {
					mem.Logger().Warnf("队长ID:%d, 未配置警戒值，不检查", pid)
				}
			}
			// AA支付一条线
			if consts.ClubHouseOwnerPay == false {
				for _, pid := range partnerLink {
					if attr, ok := partnerAttrs[pid]; ok && attr.AA == true {
						payer = pid
						mem.Logger().Warnf("AA支付，队长ID:%d", payer)
						break
					} else {
						mem.Logger().Warnf("队长ID:%d, 未配置AA支付", pid)
					}
				}
			} else {
				mem.Logger().Warnf("AA支付，BUT圈主支付")
			}
			//// AA支付直属
			//if directPartner := mem.GetDirectPartner(); directPartner > 0 {
			//	if attr, ok := partnerAttrs[directPartner]; ok {
			//		if consts.ClubHouseOwnerPay == false && attr.AA {
			//			payer = directPartner
			//			mem.Logger().Warnf("AA支付，队长ID:%d", payer)
			//		}
			//	}
			//}
		}
		//if consts.ClubHouseOwnerPay == false {
		//	for _, pid := range partnerLink {
		//		if attr, ok := partnerAttrs[pid]; ok && attr.AA {
		//			payer = pid
		//		}
		//	}
		//}
	}
	/*
		////////////////////////
			检查完毕，正式入桌
		///////////////////////
	*/
	hft, seat, cusErr := floor.GetQuickTable(p, gpsInfo, req.NTId, req.RestartId, priorityUsers...)
	if cusErr != nil && cusErr != xerrors.RespOk {
		xlog.Logger().Error(cusErr.Msg)
		return cusErr.Code, cusErr.Msg
	}
	if hft == nil || seat == -1 {
		return xerrors.HouseNoConditionsError.Code, xerrors.HouseNoConditionsError.Msg
	}
	if gpsInfo != nil && p != nil {
		if gpsInfo.Longitude != 0 && gpsInfo.Latitude != 0 && gpsInfo.Address == "" {
			gpsInfo.Address = "未知区域"
		}
		if gpsInfo.Latitude != -1 && gpsInfo.Longitude != -1 {
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "Latitude", gpsInfo.Latitude, "Longitude", gpsInfo.Longitude, "GameIp", gpsInfo.Ip, "Address", gpsInfo.Address)
		}
	}

	// 是否需要检测禁止同桌
	isCheckLimitTable := true
	curTable := GetTableMgr().GetTable(hft.TId)
	if curTable != nil {
		if curTable.Config.MaxPlayerNum == 2 && clubHouse.DBClub.IsNotEft2PTale == true {
			isCheckLimitTable = false
		}
	}

	if isCheckLimitTable {
		limitCode, respMsg := floor.TableLimitMsg(p.Uid, hft)
		if limitCode != xerrors.SuccessCode {
			defer func() {
				hft.DataLock.Unlock()
				hft.LockClose <- struct{}{}
				hft.IsOccupy = 0
			}()

			hft.UserWithOnline[seat].Uid = 0
			table := GetTableMgr().GetTable(hft.TId)
			if table != nil {
				table.UserLeaveTable(p.Uid)
			}
			return xerrors.HouseTableLimitJoin.Code, respMsg
		}
	}

	//开始入桌
	var seatOk bool
	// 牌桌是否创建
	if hft.TId == 0 {
		defer func() {
			if !seatOk {
				tb := GetTableMgr().GetTable(hft.TId)
				if tb != nil {
					GetTableMgr().DelTable(tb)
				}
				hft.Clear(floor.DHId)
			} else {
				if clubHouse.DBClub.TableJoinType != consts.SelfAdd && floor.IsMix && clubHouse.DBClub.MixActive {
					floor.CheckEmptyAndAdd()
				}
			}
		}()
		defer func() {
			hft.DataLock.Unlock()
			hft.LockClose <- struct{}{}
			hft.IsOccupy = 0
		}()
		//获取游戏服务器
		gameserver, notice := GetServer().GetGameByKindId(p.Uid, static.GAME_TYPE_FRIEND, floor.Rule.KindId)
		if gameserver == nil {
			// 如果有维护公告 则弹维护公告
			if notice != nil {
				if notice.GameServerId == static.NoticeMaintainServerAllServer {
					return xerrors.ServerMaintainError.Code, static.HF_JtoA(notice)
				} else {
					return xerrors.GameMaintainError.Code, notice
				}
			}
			return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
		}
		// 游戏房卡消耗
		cardcost := config.GetCardCost(floor.Rule.RoundNum, floor.Rule.PlayerNum, consts.ClubHouseOwnerPay)
		// 房卡消耗校验
		tablePayer, err := GetDBMgr().GetDBrControl().GetPerson(payer)
		if err != nil {
			xlog.Logger().Errorln(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		if tablePayer == nil {
			return xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg
		}
		useLeague, league := GetAllianceMgr().CheckUserLeagueCardPool(clubHouse.DBClub.UId)
		if useLeague.CanPool {
			if league.Card-league.FreezeCard < int64(cardcost) {
				useLeague.CanPool = false
			} else {
				var ok bool
				for {
					if ok {
						defer league.DelLockWithLog()
						break
					} else {
						time.Sleep(5 * time.Millisecond)
						ok = league.SetNx()
					}
				}
			}
		}
		if !useLeague.CanPool {
			if tablePayer.Card-tablePayer.FrozenCard < cardcost {
				if tablePayer.Uid != clubHouse.DBClub.UId {
					return xerrors.HousePartnerNotEnoughCardError.Code, xerrors.HousePartnerNotEnoughCardError.Msg
				} else {
					return xerrors.HouseNotEnoughCardError.Code, xerrors.HouseNotEnoughCardError.Msg
				}
			}
		}

		// 规则合法判定
		var msgct static.Msg_CreateTable
		msgct.KindId = floor.Rule.KindId
		msgct.PlayerNum = floor.Rule.PlayerNum
		msgct.RoundNum = floor.Rule.RoundNum
		msgct.CostType = floor.Rule.CostType
		msgct.Restrict = floor.Rule.Restrict
		msgct.GameConfig = floor.Rule.GameConfig
		msgct.GVoice = floor.Rule.GVoice
		msgct.Gps = req.Gps
		msgct.Voice = req.Voice
		msgct.GVoiceOk = req.GVoiceOk
		msgct.FewerStart = floor.Rule.FewerStart
		newConfig, cusErr := validateCreateTableParam(&msgct, true)
		if cusErr != nil {
			xlog.Logger().Error(cusErr.Msg)
			return cusErr.Code, cusErr.Msg
		}

		// 包厢开关
		houseSwitch, _ := clubHouse.GetMemberSwitch()

		// 创建内存牌桌
		table := new(static.Table)
		nowTime := time.Now()
		table.Id = GetTableMgr().GetRandomTableId()
		if table.Id <= 0 {
			xlog.Logger().Errorln("GetRandomTableId error")
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		}
		table.NTId = hft.NTId
		table.HId = floor.HId
		table.DHId = floor.DHId
		table.FId = floor.Id
		table.NFId = clubHouse.GetFloorIndexByFid(floor.Id)
		table.IsCost = false
		table.Creator = clubHouse.DBClub.UId
		table.CreateType = consts.CreateTypeHouse
		table.GameId = gameserver.Id
		table.KindId = newConfig.KindId
		table.KindVersion = floor.Rule.Version
		table.Users = make([]*static.TableUser, newConfig.MaxPlayerNum)
		table.GameNum = fmt.Sprintf("%s_%d_%d", nowTime.Format("20060102150405"), gameserver.Id, table.Id)
		table.IsFloorHideImg = floor.IsHide
		table.IsMemUidHide = clubHouse.DBClub.IsMemUidHide
		if houseSwitch != nil {
			table.IsForbidWX = houseSwitch["BanWeChat"] == 0
		}
		table.Config = &static.TableConfig{
			MaxPlayerNum: newConfig.MaxPlayerNum,
			MinPlayerNum: newConfig.MinPlayerNum,
			RoundNum:     newConfig.RoundNum,
			CardCost:     cardcost,
			CostType:     newConfig.CostType,
			View:         config.DefaultView,
			Restrict:     newConfig.Restrict,
			GameConfig:   newConfig.GameConfig,
			FewerStart:   newConfig.FewerStart,
			GVoice:       newConfig.GVoice,
			GameType:     2, // 默认好友房
			Difen:        newConfig.Difen,
		}
		table.IsVitamin = clubHouse.DBClub.IsVitamin && floor.IsVitamin
		table.IsHidHide = clubHouse.DBClub.IsHidHide
		table.IsAiSuper = clubHouse.DBClub.MixActive &&
			// 楼层是否为混排楼层
			floor.IsMix &&
			// 防作弊模式
			clubHouse.DBClub.TableJoinType == consts.NoCheat &&
			// 超级防作弊模式
			clubHouse.DBClub.AiSuper && floor.AiSuperNum > 0

		table.CreateStamp = hft.CreateStamp
		// 得到匹配信息
		if table.IsAiSuper {
			table.CurrentMappingNum, _, table.TotalMappingNum = floor.MappingInfo(false)
		}
		//table.CurrentMappingNum = floor.CurrentMappingInfo()
		//table.TotalMappingNum = floor.NumTotalMapping()
		if table.CurrentMappingNum >= table.TotalMappingNum {
			table.CurrentMappingNum = table.TotalMappingNum - 1
		}

		if clubHouse.DBClub.MixActive && floor.IsMix {
			table.JoinType = clubHouse.DBClub.TableJoinType
		}
		ntable := GetTableMgr().CreateTable(table)
		if ntable == nil {
			xlog.Logger().Errorln("protocol_housefloor 408 table创建失败", table.Id)
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		}

		nowPackageKey := GetAreaPackageKeyByKid(table.KindId)
		if nowPackageKey != areaGame.PackageKey {
			return xerrors.ResultErrorCode, "楼层玩法已变更，请重试2。"
		}

		req := new(static.HG_HTableCreate_Req)
		req.Table = *table
		req.AutoSeat = seat
		req.AutoUid = p.Uid
		req.Payer = payer
		if GetServer().Con.UseSafeIp == 0 {
			req.Ip = gameserver.ExIp
		} else {
			req.Ip = gameserver.SafeIp
		}
		if useLeague.CanPool {
			req.Table.LeagueID = league.AllianceBizID
			req.Table.NotPool = useLeague.NotPool
		} else {
			req.Table.LeagueID = 0
		}
		// 关联牌桌数据
		hft.TId = ntable.Id
		// 发送创建协议
		reply, err := GetServer().CallGame(table.GameId, p.Uid, "NewServerMsg", consts.MsgTypeHTableCreate_Req, xerrors.SuccessCode, req)
		if string(reply) != "SUC" || err != nil {
			xlog.Logger().Errorln("CallGame error1")
			// 删除内存数据
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		} else {
			if cardcost > 0 {
				// 房卡冻结变更
				tx := GetDBMgr().GetDBmControl().Begin()
				// defer tx.Rollback()
				if useLeague.CanPool {
					err := wealthtalk.UpdateLeagueCard(league.AllianceBizID, p.Uid, 0, cardcost, models.CostFrozen, tx) // 增加加盟商房卡消耗记录
					if err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
					err = GetAllianceMgr().UpdateLeagueFreezeCard(league.AllianceBizID, cardcost, true)
					if err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
					if err := tx.Commit().Error; err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						GetAllianceMgr().UpdateLeagueFreezeCard(league.AllianceBizID, cardcost, false)
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
					if useLeague.NotPool {
						GetAllianceMgr().UpdateUserCard(clubHouse.DBClub.UId, 0, int64(cardcost))
					}

				} else {
					_, aftka, _, aftfka, err := wealthtalk.UpdateCard(tablePayer.Uid, 0, cardcost, models.CostFrozen, tx)
					if err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
					} else {
						// 更新db
						tx.Commit()
						// 更新内存
						p := GetPlayerMgr().GetPlayer(tablePayer.Uid)
						if p != nil {
							p.Info.Card = aftka
							p.Info.FrozenCard = aftfka
						}
						// 更新redis
						GetDBMgr().GetDBrControl().UpdatePersonAttrs(tablePayer.Uid, "Card", aftka, "FrozenCard", aftfka)
					}
				}
			}
			p.TableId = table.Id
			p.GameId = table.GameId
			GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "TableId", p.TableId, "GameId", p.GameId)

			// 修改牌桌玩家座位信息
			ntable.lock.CustomLock()
			ntable.Users[seat] = &static.TableUser{Uid: p.Uid, JoinAt: time.Now().Unix(), Payer: payer}
			ntable.lock.CustomUnLock()

			// 桌子信息 redis
			floortable := new(static.FloorTable)
			floortable.NTId = hft.NTId
			floortable.TId = hft.TId
			floortable.DHId = ntable.DHId
			floortable.FId = ntable.FId
			GetDBMgr().GetDBrControl().FloorTableInsert(floortable)

			tableinMsg := new(static.Msg_S2C_TableIn)
			tableinMsg.Id = table.Id
			tableinMsg.GameId = table.GameId
			tableinMsg.KindId = table.KindId
			tableinMsg.Ip = req.Ip
			seatOk = true

			go GetClubMgr().UserInGame(p.Uid)
			//if table.IsAiSuper {
			//	floor.PubRedisMappingNumUpdate()
			//}
			tableinMsg.PkgName = nowPackageKey
			return xerrors.SuccessCode, &tableinMsg
		}

	} else {
		defer func() {
			if !seatOk {
				hft.UserWithOnline[seat].Uid = 0
				table := GetTableMgr().GetTable(hft.TId)
				if table != nil {
					table.UserLeaveTable(p.Uid)
				}
			}
		}()
		defer func() {
			hft.DataLock.Unlock()
			hft.LockClose <- struct{}{}
			hft.IsOccupy = 0
		}()
		// 加入牌桌
		table := GetTableMgr().GetTable(hft.TId)
		if table == nil {
			return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
		}

		if notice := CheckServerMaintainWithWhite(p.Uid, static.NoticeMaintainServerType(table.GameId)); notice != nil {
			return xerrors.GameMaintainError.Code, notice
		}

		if table.Config.GVoice == "true" {
			if !req.Voice {
				// 未开启录音权限
				// return xerrors.OpenGVoiceError.Code, xerrors.OpenGVoiceError.Msg
			}
			if !req.GVoiceOk {
				// 未开启录音权限
				return xerrors.InitGVoiceError.Code, xerrors.InitGVoiceError.Msg
			}
		}

		// 判断gps限制

		//str := fmt.Sprintf("这里检查入桌限制 ChOptTableIn 桌子ID：%d , 最大开始人数：%d , 是否需要验证IP：%v", table.Id, table.Config.MaxPlayerNum, table.Config.Restrict)
		//syslog.Logger().Errorf(str)

		if table.Config.Restrict && table.Config.MaxPlayerNum != 2 { //2人玩 不做IP和GPS限制
			// 如果不是再来一局且未开启gps则提示开启
			if !req.Gps {
				// 未开启gps服务
				cuserror := xerrors.OpenGpsError
				return cuserror.Code, cuserror.Msg
			}
			if ok := CheckUserIp(gpsInfo.Ip, table.Users); !ok {
				// 未开启gps服务
				cuserror := xerrors.NewXError("相同ip无法加入该桌")
				return cuserror.Code, cuserror.Msg
			}
			uids := make([]int64, 0, 4)
			for _, u := range table.Users {
				if u != nil {
					uids = append(uids, u.Uid)
				}
			}
			if e := CheckUserGps(clubHouse, gpsInfo.Longitude, gpsInfo.Latitude, uids); e != nil {
				xlog.Logger().Error(e)
				cuserror := xerrors.NewXError("本桌有距离过近玩家，无法进入")
				return cuserror.Code, cuserror.Msg
			}
		}

		//获取游戏服务器
		gameserver := GetServer().GetGame(table.GameId)
		if gameserver == nil {
			return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
		}
		// 非空桌 判定规则是否相同
		frule := new(static.Msg_CreateTable)
		frule.KindId = floor.Rule.KindId
		frule.PlayerNum = floor.Rule.PlayerNum
		frule.RoundNum = floor.Rule.RoundNum
		frule.CostType = floor.Rule.CostType
		frule.Restrict = floor.Rule.Restrict
		frule.GameConfig = floor.Rule.GameConfig
		frule.FewerStart = floor.Rule.FewerStart
		frule.Gps = req.Gps
		frule.GVoice = floor.Rule.GVoice
		frule.Voice = req.Voice
		frule.GVoiceOk = req.GVoiceOk
		config, cuserror := validateCreateTableParam(frule, true)
		if cuserror != nil {
			return cuserror.Code, cuserror.Msg
		}

		if !req.IgnoreRule {
			if config.KindId != table.KindId ||
				config.MaxPlayerNum != table.Config.MaxPlayerNum ||
				config.RoundNum != table.Config.RoundNum ||
				config.CardCost != table.Config.CardCost ||
				config.CostType != table.Config.CostType ||
				config.View != table.Config.View ||
				config.Restrict != table.Config.Restrict ||
				config.FewerStart != table.Config.FewerStart ||
				config.GameConfig != table.Config.GameConfig {
				return xerrors.HouseFloorRuleChangeError.Code, xerrors.HouseFloorRuleChangeError.Msg
			}
		}

		// 再来一局校验
		if msg.Param.NTId == consts.HOUSETABLEINAGAIN {
			// 校验加入桌子参数
			custerr := table.CheckRestartJoin(p, gpsInfo.Ip, seat)
			if custerr != nil {
				return custerr.Code, custerr.Msg
			}
		} else {
			// 校验加入桌子参数
			custerr := table.CheckUserJoinTable(p, seat)
			if custerr != nil {
				return custerr.Code, custerr.Msg
			}
		}

		nowPackageKey := GetAreaPackageKeyByKid(table.KindId)
		if nowPackageKey != areaGame.PackageKey {
			return xerrors.ResultErrorCode, "楼层玩法已变更，请重试3。"
		}
		var cardcost int
		// 游戏房卡消耗
		if consts.ClubHouseOwnerPay == false {
			cardcost = table.Table.Config.CardCost
		}
		// 房卡消耗校验
		tablePayer, err := GetDBMgr().GetDBrControl().GetPerson(payer)
		if err != nil {
			xlog.Logger().Errorln(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		if tablePayer == nil {
			return xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg
		}
		useLeague, league := GetAllianceMgr().CheckUserLeagueCardPool(clubHouse.DBClub.UId)
		if useLeague.CanPool {
			if league.Card-league.FreezeCard < int64(cardcost) {
				useLeague.CanPool = false
			} else {
				var ok bool
				for {
					if ok {
						defer league.DelLockWithLog()
						break
					} else {
						time.Sleep(5 * time.Millisecond)
						ok = league.SetNx()
					}
				}
			}
		}
		if !useLeague.CanPool {
			if tablePayer.Card-tablePayer.FrozenCard < cardcost {
				if tablePayer.Uid != clubHouse.DBClub.UId {
					return xerrors.HousePartnerNotEnoughCardError.Code, xerrors.HousePartnerNotEnoughCardError.Msg
				} else {
					return xerrors.HouseNotEnoughCardError.Code, xerrors.HouseNotEnoughCardError.Msg
				}
			}
		}
		// 通知游戏服加入牌桌	table.UserLeaveTable(msg.Uid)
		var resp static.HG_HTableIn_Req
		resp.TId = table.Id
		resp.NTid = table.NTId
		resp.GameId = table.GameId
		resp.KindId = table.KindId
		if GetServer().Con.UseSafeIp == 0 {
			resp.Ip = gameserver.ExIp
		} else {
			resp.Ip = gameserver.SafeIp
		}
		resp.Uid = p.Uid
		resp.Seat = seat
		resp.Payer = payer
		reply, err := GetServer().CallGame(table.GameId, p.Uid, "NewServerMsg", msg.Header, xerrors.SuccessCode, &resp)
		if string(reply) != "SUC" || err != nil {
			// 删除内存数据
			xlog.Logger().Error(err, string(reply))
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		} else {
			if cardcost > 0 {
				// 房卡冻结变更
				tx := GetDBMgr().GetDBmControl().Begin()
				// defer tx.Rollback()
				if useLeague.CanPool {
					err := wealthtalk.UpdateLeagueCard(league.AllianceBizID, p.Uid, 0, cardcost, models.CostFrozen, tx) // 增加加盟商房卡消耗记录
					if err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
					err = GetAllianceMgr().UpdateLeagueFreezeCard(league.AllianceBizID, cardcost, true)
					if err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
					if err := tx.Commit().Error; err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						GetAllianceMgr().UpdateLeagueFreezeCard(league.AllianceBizID, cardcost, false)
						return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
					}
					if useLeague.NotPool {
						GetAllianceMgr().UpdateUserCard(clubHouse.DBClub.UId, 0, int64(cardcost))
					}

				} else {
					_, aftka, _, aftfka, err := wealthtalk.UpdateCard(tablePayer.Uid, 0, cardcost, models.CostFrozen, tx)
					if err != nil {
						xlog.Logger().Errorln(err)
						tx.Rollback()
						return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
					} else {
						// 更新db
						tx.Commit()
						// 更新内存
						p := GetPlayerMgr().GetPlayer(tablePayer.Uid)
						if p != nil {
							p.Info.Card = aftka
							p.Info.FrozenCard = aftfka
						}
						// 更新redis
						GetDBMgr().GetDBrControl().UpdatePersonAttrs(tablePayer.Uid, "Card", aftka, "FrozenCard", aftfka)
					}
				}
			}
			// 占有桌子座位
			p.TableId = table.Id
			p.GameId = table.GameId

			// 修改牌桌玩家座位信息
			ntable := GetTableMgr().GetTable(table.Id)
			if ntable == nil {
				return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
			}
			ntable.lock.CustomLock()
			ntable.Users[seat] = &static.TableUser{Uid: p.Uid, JoinAt: time.Now().Unix(), Payer: payer}
			ntable.lock.CustomUnLock()
			tableinMsg := new(static.Msg_S2C_TableIn)
			tableinMsg.Id = table.Id
			tableinMsg.GameId = table.GameId
			tableinMsg.KindId = table.KindId
			tableinMsg.Ip = resp.Ip

			// 如果指定了优先匹配的玩家，而且这里匹配到入桌，则置该桌子的再来一局标识为true
			seatOk = true
			// go floor.RedisPub(constant.MsgTypeHouseTableIn_Ntf, ntf)
			go GetClubMgr().UserInGame(p.Uid)
			//if table.IsAiSuper {
			//	floor.PubRedisMappingNumUpdate()
			//}
			tableinMsg.PkgName = nowPackageKey
			return xerrors.SuccessCode, &tableinMsg
		}
	}
}

// 玩家离开牌桌
func ChOptTableOut_Ntf(floor *HouseFloor, s *Session, p *static.Person, msg *static.GH_TableExit_Ntf) (code int16, v interface{}) {

	hft := floor.GetHftByTid(msg.TableId)
	if hft == nil {
		xlog.Logger().Errorln("header:", consts.HFChOptTableOut_Ntf, " tid:", msg.TableId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	//桌子加锁，直到退出桌子
	hft.DataLock.Lock()
	defer hft.DataLock.Unlock()
	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Errorf("header:%s ,tid:%d not exist", consts.MsgTypeTableExit_Ntf, msg.TableId)
		return xerrors.AsyncRespErrorCode, nil
	}
	// 离开广播
	house := GetClubMgr().GetClubHouseByHId(floor.HId)
	if house == nil {
		xlog.Logger().Errorln("header:", consts.HFChOptTableOut_Ntf, " hid:", floor.HId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	seat := hft.GetUserSeat(msg.Uid)
	if seat < 0 {
		xlog.Logger().Errorln("header:", consts.HFChOptTableOut_Ntf, " user:", msg.Uid, " not in table")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 解锁座位
	hft.UnlockSeat(seat, msg.Uid)

	// 最后一个人离开解散牌桌
	if hft.UserCount() == 0 {
		// 通知游戏服解散牌桌
		var req static.Msg_HG_TableDel_Req
		req.TableId = hft.TId

		go GetServer().CallGame(msg.GameId, 0, "ServerMethod.ServerMsg", consts.MsgTypeHTableDel_Req, xerrors.SuccessCode, &req)

		// redis
		floortable := new(static.FloorTable)
		floortable.NTId = hft.NTId
		floortable.TId = hft.TId
		floortable.DHId = floor.DHId
		floortable.FId = floor.Id
		GetDBMgr().GetDBrControl().FloorTableDelete(floortable)
		// memory
		hft.Clear(floor.DHId)
		if len(hft.UserWithOnline) != floor.Rule.PlayerNum {
			hft.UserWithOnline = make([]FTUsers, floor.Rule.PlayerNum)
		}
		// 删除内存数据
		for _, u := range table.Users {
			if u != nil {
				table.UserLeaveTable(u.Uid)
			}
		}
		GetTableMgr().DelTable(table)
	} else {
		table.UserLeaveTable(msg.Uid)
		go GetClubMgr().UserOutGame(msg.Uid)
	}

	return xerrors.SuccessCode, nil
}

// 包厢牌桌删除
func ChOptTableDel_Ntf(floor *HouseFloor, s *Session, p *static.Person, ntf *static.FloorTable) (code int16, v interface{}) {
	hft := floor.GetHftByTid(ntf.TId)
	if hft == nil {
		xlog.Logger().Warnf("header:%s,table id :%d is not exist.", consts.HFChOptTableDel_Ntf, ntf.TId)
		return xerrors.InValidHouseTableError.Code, xerrors.InValidHouseTableError.Msg
	}
	hft.DataLock.Lock()
	defer hft.DataLock.Unlock()

	table := GetTableMgr().GetTable(ntf.TId)
	if table == nil {
		xlog.Logger().Errorln("ChOptTableDel_Ntf error: table is nil")
		return xerrors.InValidHouseTableError.Code, xerrors.InValidHouseTableError.Msg
	}

	// redis
	GetDBMgr().GetDBrControl().FloorTableDelete(ntf)
	// 清理数据
	hft.Clear(floor.DHId)
	if len(hft.UserWithOnline) != floor.Rule.PlayerNum {
		hft.UserWithOnline = make([]FTUsers, floor.Rule.PlayerNum)
	}
	// floor.ReviseUsableTbl()
	return xerrors.SuccessCode, nil
}

// 解散牌桌
func (self *HouseFloorProtocolMgr) ChOpt_TableDel_Ntf(floor *HouseFloor, s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	ntf := data.(*static.FloorTable)

	hft := floor.GetTableByNTId(ntf.NTId)
	if hft == nil {
		xlog.Logger().Errorln("header:", consts.HFChOptTableDel_Ntf, " hft:", ntf.TId, " is not exist.")
		return xerrors.AsyncRespErrorCode, nil
	}

	if ntf.TId != hft.TId {
		return xerrors.AsyncRespErrorCode, nil
	}
	// redis
	GetDBMgr().GetDBrControl().FloorTableDelete(ntf)

	hft.DataLock.Lock()
	defer hft.DataLock.Unlock()

	// msg := new(public.Ntf_HC_HouseTableDissovel)
	// msg.HId = floor.HId
	// msg.FId = ntf.FId
	// msg.NTId = ntf.NTId
	// go floor.BroadCastMix(constant.ROLE_MEMBER, constant.MsgTypeHouseTableDissovle_Ntf, msg)
	// go floor.RedisPub(constant.MsgTypeHouseTableDissovle_Ntf, ntf)
	// 清理数据
	hft.Clear(floor.DHId)
	if len(hft.UserWithOnline) != floor.Rule.PlayerNum {
		hft.UserWithOnline = make([]FTUsers, floor.Rule.PlayerNum)
	}

	return xerrors.AsyncRespErrorCode, nil
}

// 桌子信息更新
func (self *HouseFloorProtocolMgr) Chopt_TableUpdate(floor *HouseFloor, s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	msg := data.(*static.FloorTableId)
	go floor.BroadCastMix(consts.ROLE_MEMBER, consts.MsgTypeHouseTableUpdate_Ntf, floor.MemTableBaseInfo(GetTableMgr().GetTable(msg.TId)))
	return xerrors.AsyncRespErrorCode, nil
}

// 桌子信息更新
func (self *HouseFloorProtocolMgr) Chopt_TableStepUpdate(floor *HouseFloor, s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	go floor.BroadCastMix(consts.ROLE_MEMBER, consts.MsgTtypeTableSetBegin_NTF, data)
	return xerrors.AsyncRespErrorCode, nil
}

func GetFloorDetail(floor *HouseFloor) (ack *static.Msg_HC_HouseMemberIn) {
	ack = floor.BuildAck(floor.GetAllTables(), false)
	if ack == nil {
		ack = &static.Msg_HC_HouseMemberIn{}
		ack.FId = floor.Id
		ack.FTableItems = make([]static.Msg_HouseTableItem, 0)
	}
	return
}

type CreateTimeInfo struct {
	IDUsed bool
	Stamp  int64
}

type StampSlice []CreateTimeInfo

func (p StampSlice) Len() int           { return len(p) }
func (p StampSlice) Less(i, j int) bool { return p[i].Stamp < p[j].Stamp }
func (p StampSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func sortTableByCreateTime(mixFloor []*HouseFloor) bool {
	xlog.Logger().Warnf("sortTableByCreateTime")
	var length int
	for _, ack := range mixFloor {
		length += len(ack.Tables)
	}
	stampSlice := make([]CreateTimeInfo, 0, length)
	for _, floor := range mixFloor {
		for _, v := range floor.Tables {
			stampSlice = append(stampSlice, CreateTimeInfo{false, v.CreateStamp})
		}
	}
	sort.Sort(StampSlice(stampSlice))
	var flag bool
	for _, floor := range mixFloor {
		tmpTableMap := make(map[int]*HouseFloorTable)
		for _, v := range floor.Tables {
			for i := 0; i < length; i++ {
				if v.CreateStamp == stampSlice[i].Stamp && !stampSlice[i].IDUsed {
					if v.NTId != i {
						if !flag {
							xlog.Logger().Warnf("ntid change do sync")
						}
						flag = true
					}
					v.NTId = i
					stampSlice[i].IDUsed = true
					tmpTableMap[i] = v
					break
				}
			}
		}
		floor.Tables = tmpTableMap
	}
	return flag
}
