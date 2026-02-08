package center

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

type httpService struct {
	pool map[string]*router.AppWorker
}

func NewHttpService() *httpService {
	return &httpService{}
}

// 初始化HTTP服务
func (hs *httpService) OnInit() {
	hs.pool = make(map[string]*router.AppWorker)
	R := hs.Register
	// 包厢游戏再来一局
	R(consts.MsgTypeHouseAnotherGame, static.Msg_C2Http_HouseAnotherGame{}, HouseAnotherGame)
	R(consts.MsgTypeHouseChangeTable, static.Msg_C2Http_HouseChangeTable{}, HouseTableChange)
	R(consts.MsgTypeHouseMemberAgree, static.Msg_C2Http_HouseMemberAgree{}, HouseMemAgree)
	R(consts.MsgTypeHouseMemberRefused, static.Msg_C2Http_HouseMemberRefused{}, HouseMemRefuse)
	R(consts.MsgTypeTableDel, static.Msg_C2Http_HouseTableDel{}, DeleteTable)
	R(consts.MsgTypeShareCfg, static.Msg_C2Http_GetShareCfg{}, GetShareCfg)
	R(consts.MsgTypeShareSuc, static.Msg_C2Http_ShareSuc{}, ShareSuc)
	R(consts.MsgTypeGetAllowanceInfo, static.Msg_C2Http_GetAllowanceInfo{}, GetAllowanceInfo)
	R(consts.MsgTypeCheckBuyBankruptcyGift, static.Msg_C2Http_TokenUid{}, CheckUserPurchasedBankruptGiftToday)
	R(consts.MsgTypePlayTime, static.Msg_C2Http_PlayTime{}, SaveUserPlayTime)
	// ...
}

func (hs *httpService) Register(header string, proto interface{}, appHandle router.AppHandlerFunc) {
	hs.pool[header] = &router.AppWorker{DataType: reflect.TypeOf(proto), Handle: appHandle}
}

func (hs *httpService) AppHandler(handler string) *router.AppWorker {
	return hs.pool[handler]
}

func (hs *httpService) EncodeInfo() (encode int, key string) {
	return GetServer().Con.Encode, GetServer().Con.EncodeClientKey
}

func TokenCheckout(uid int64, token string) (*static.Person, *xerrors.XError) {
	person, err := GetDBMgr().GetDBrControl().GetPerson(uid)
	if err != nil {
		return nil, xerrors.DBExecError
	}
	if person == nil {
		return nil, xerrors.UserNotExistError
	}
	if person.Token != token {
		return person, xerrors.TokenError
	}
	return person, nil
}

// 包厢牌桌再来一局
func HouseAnotherGame(request *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_HouseAnotherGame)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	// 正在执行退出流程
	onlyExit := fmt.Sprintf("userstatus_doing_exit_%d", req.Uid)
	count := 0
	for {
		if GetDBMgr().Redis.Exists(onlyExit).Val() == 1 {
			if count > 10 {
				return nil, xerrors.UserStatusExitError
			}
			time.Sleep(time.Millisecond * 200)
			count++
			continue
		}
		break
	}

	// 正在执行加入牌桌流程
	cli := GetDBMgr().Redis
	onlyjoin := fmt.Sprintf("userstatus_doing_join_%d", req.Uid)
	if !cli.SetNX(onlyjoin, req.Uid, time.Second*3).Val() {
		return nil, xerrors.UserStatusJoinError
	}
	defer cli.Del(onlyjoin)

	person, err := TokenCheckout(req.Uid, req.Token)
	if err != nil {
		return nil, err
	}
	// if person.TableId > 0 {
	// 	return nil, xerrors.TableInLockError
	// }
	// 再来一局处理
	record, rdsErr := GetDBMgr().GetDBrControl().SelectHRecordPlayers(person.Uid)
	if rdsErr != nil {
		xlog.Logger().Errorln("redis error:", rdsErr)
		return nil, xerrors.DBExecError
	}
	xlog.Logger().Infof("玩家(%d)上一局的对战记录为:%+v", person.Uid, record)

	longitude, e := strconv.ParseFloat(person.Longitude, 64)
	if e != nil {
		xlog.Logger().Errorf("user longitude parse error:%v", e)
	}

	latitude, e := strconv.ParseFloat(person.Latitude, 64)
	if e != nil {
		xlog.Logger().Errorf("user latitude  parse error:%v", e)
	}

	xlog.ELogger(xlog.ElkGpsInfo).WithFields(map[string]interface{}{
		"uid":       person.Uid,
		"tableid":   0,
		"gameid":    0,
		"kindid":    record.KId,
		"ip":        person.Ip,
		"longitude": longitude,
		"latitude":  latitude,
	}).Info("reqhouseanthergame")

	code, resp := HouseTableIn(&static.GpsInfo{
		Ip:        static.HF_GetHttpIP(request),
		Longitude: longitude,
		Latitude:  latitude,
		Address:   person.Address,
	}, person, &static.Msg_CH_HouseTableIn{
		HId:        record.HId,
		FId:        record.FId,
		NTId:       consts.HOUSETABLEINAGAIN,
		Gps:        true,
		Voice:      true,
		GVoiceOk:   true,
		IgnoreRule: true,
		RestartId:  record.TId,
		KindID:     record.KId,
	},
		record.Users...,
	)
	if code == xerrors.SuccessCode {
		return resp, nil
	}
	return nil, &xerrors.XError{Code: code, Msg: resp.(string)}
}

func HouseTableChange(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_HouseChangeTable)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, cusErr := TokenCheckout(req.Uid, req.Token)
	if cusErr != nil {
		return nil, cusErr
	}
	if p.TableId <= 0 {
		return nil, xerrors.InvalidParamError
	}
	cli := GetDBMgr().Redis
	user_lock_key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_LOCK, p.Uid)
	if !cli.SetNX(user_lock_key, p.Uid, 3*time.Second).Val() {
		return nil, xerrors.HouseFloorTableJoiningError
	}
	defer func() {
		cli.Del(user_lock_key)
	}()
	house, floor, _, cusErr := inspectClubFloorMemberWithRight(req.HId, req.FId, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return nil, cusErr
	}

	table := GetTableMgr().GetTable(p.TableId)
	if table == nil {
		return nil, xerrors.TableIdError
	}
	if table.Begin {
		return nil, xerrors.TableInLockError
	}
	if house.DBClub.OnlyQuickJoin {
		return nil, xerrors.OnlyQucikJoinError
	}

	gameserver := GetServer().GetGame(table.GameId)
	if gameserver == nil {
		return nil, xerrors.GetGameServerError
	}
	var Ip string
	if GetServer().Con.UseSafeIp == 0 {
		Ip = gameserver.ExIp
	} else {
		Ip = gameserver.SafeIp
	}

	tableinMsg := new(static.Msg_S2C_TableIn)
	tableinMsg.Id = table.Id
	tableinMsg.GameId = table.GameId
	tableinMsg.KindId = table.KindId
	tableinMsg.Ip = Ip
	tableinMsg.Version = table.KindVersion
	game := GetAreaGameByKid(floor.Rule.KindId)
	if game != nil {
		tableinMsg.PkgName = game.PackageKey
	} else {
		tableinMsg.PkgName = "uk"
	}
	if table.NTId == req.NTId {
		return &tableinMsg, nil
	}
	frule := new(static.Msg_CreateTable)
	frule.KindId = floor.Rule.KindId
	frule.PlayerNum = floor.Rule.PlayerNum
	frule.RoundNum = floor.Rule.RoundNum
	frule.CostType = floor.Rule.CostType
	frule.Restrict = floor.Rule.Restrict
	frule.GameConfig = floor.Rule.GameConfig
	frule.FewerStart = floor.Rule.FewerStart
	frule.GVoice = floor.Rule.GVoice
	frule.Gps = req.Gps
	frule.Voice = req.Voice
	frule.GVoiceOk = req.GVoiceOk
	_, cuserror := validateCreateTableParam(frule, true)
	if cuserror != nil {
		return &tableinMsg, &xerrors.XError{Code: cuserror.Code, Msg: cuserror.Msg}
	}
	// if floor.Rule.Version != req.Version {
	// 	return &tableinMsg, &xerrors.XError{Code: xerrors.ChangeTableError, Msg: "当前包厢规则已修改，暂时未应用于此桌，确定加入游戏吗"}
	// }
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_TABLE_LOCK, table.FId, table.NTId, p.SiteId)
	if !cli.SetNX(key, p.Uid, time.Second*3).Val() {
		return nil, xerrors.HouseFloorTableJoiningError
	}
	defer func() {
		cli.Del(key)
	}()

	msg := static.MsgHouseUserExitTable{Uid: req.Uid, Tid: table.Id}
	//通知游戏服退出牌桌
	reply, err := GetServer().CallGame(table.GameId, req.Uid, "NewServerMsg", consts.MsgTypeTableExit, xerrors.SuccessCode, &msg)
	if string(reply) != "SUC" || err != nil {
		// xerrors.TableInLockError
		return &tableinMsg, nil
	}
	// 删除内存数据
	table.UserLeaveTable(p.Uid)

	p.TableId = 0
	// 如果是最后一人解散牌桌

	msgExit := new(static.GH_TableExit_Ntf)
	msgExit.Uid = p.Uid
	msgExit.GameId = table.GameId
	msgExit.TableId = table.Id
	msgExit.KindId = table.KindId
	oldFloor := house.GetFloorByFId(table.FId)
	ChOptTableOut_Ntf(oldFloor, nil, nil, msgExit)

	if req.KindID == 0 {
		req.KindID = floor.Rule.KindId
	}

	code, resp := HouseTableIn(&static.GpsInfo{Ip: static.HF_GetHttpIP(r), Longitude: req.Longitude, Latitude: req.Latitude, Address: req.Address}, p, &req.Msg_CH_HouseTableIn)
	if code == xerrors.SuccessCode {
		return resp, nil
	} else if code == 145 {
		return nil, &xerrors.XError{Code: code, Msg: resp.(string)}
	} else {
		req.Msg_CH_HouseTableIn.NTId = table.NTId
		req.Msg_CH_HouseTableIn.KindID = table.KindId
		reInCode, reInresp := HouseTableIn(&static.GpsInfo{static.HF_GetHttpIP(r), req.Longitude, req.Latitude, req.Address}, p, &req.Msg_CH_HouseTableIn)
		if reInCode == xerrors.SuccessCode {
			return reInresp, &xerrors.XError{Code: xerrors.ChangeTableError, Msg: resp.(string)}
		}
		return nil, &xerrors.XError{Code: reInCode, Msg: reInresp.(string)}
	}
}

func HouseMemAgree(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_HouseMemberAgree)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, cusErr := TokenCheckout(req.Opuid, req.Token)
	if cusErr != nil {
		return nil, cusErr
	}
	cusErr = HouseMemAgreeHandle(&req.Msg_CH_HouseMemberAgree, p.Uid, p.Nickname)
	if cusErr != nil {
		return nil, cusErr
	}
	return nil, xerrors.RespOk
}

func HouseMemRefuse(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_HouseMemberRefused)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, cusErr := TokenCheckout(req.Opuid, req.Token)
	if cusErr != nil {
		return nil, cusErr
	}
	cusErr = HouseMemRefuseHandle(&req.Msg_CH_HouseMemberRefused, p.Uid, p.Nickname)
	if cusErr != nil {
		return nil, cusErr
	}
	return nil, xerrors.RespOk
}

func DeleteTable(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_HouseTableDel)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, cusErr := TokenCheckout(req.Opuid, req.Token)
	if cusErr != nil {
		return nil, cusErr
	}
	code, v := Proto_DeleteTable(nil, p, &req.DelTable)
	if code == xerrors.SuccessCode {
		return nil, xerrors.RespOk
	}
	return nil, &xerrors.XError{Code: code, Msg: v.(string)}
}

func ResetHouse(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_Http_HouseTableReset)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	if req.PassCode != "!@#wqsae#%^*" {
		return nil, xerrors.ArgumentError
	}
	// 获取包厢
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		return nil, xerrors.InValidHouseError
	}
	nhouse := new(Club)
	nhouse.DBClub = house.DBClub

	nhouse.OptLock = new(lock2.RWMutex)
	nhouse.Floors = make(map[int64]*HouseFloor)
	nhouse.FloorLock = new(lock2.RWMutex)
	nhouse.AddTableLimitLock = new(lock2.RWMutex)
	nhouse.TableLimitUserLock = new(lock2.RWMutex)
	nhouse.AddUserGroupLock = new(lock2.RWMutex)
	nhouse.ClubMemberSwitchLock = new(lock2.RWMutex)
	nhouse.initPrize()
	nhouse.IsAlive = true
	if nhouse.DBClub.MixActive {
		go nhouse.StartSync()
		go nhouse.StartSyncAiSuperNum()
	}
	// 初始化总人数及在线人数
	nhouse.SyncLiveData()

	for _, f := range house.Floors {
		f.IsAlive = false
		bf := new(HouseFloor)
		bf.Id = f.Id
		bf.Rule = f.Rule
		bf.IsAlive = true
		bf.MemAct = make(map[int64]*HouseMember)
		bf.Tables = make(map[int]*HouseFloorTable)
		bf.MemLock = new(lock2.RWMutex)
		bf.DataLock = new(lock2.RWMutex)
		bf.Name = f.Name
		bf.AiSuperNum = f.AiSuperNum

		bf.HId = house.DBClub.HId
		bf.DHId = house.DBClub.Id
		bf.IsMix = f.IsMix
		bf.IsVip = f.IsVip
		bf.IsCapSetVip = f.IsCapSetVip
		bf.IsDefJoinVip = f.IsDefJoinVip
		bf.FloorVitaminOptions = f.FloorVitaminOptions
		bf.IsHide = f.IsHide
		bf.MinTable = f.MinTable
		bf.MaxTable = f.MaxTable
		nhouse.Floors[bf.Id] = bf

		if bf.IsMix && nhouse.DBClub.MixActive {
			bf.ReloadHft()
		} else {
			for i := 0; i < GetServer().ConHouse.TableNum; i++ {
				hft := new(HouseFloorTable)
				hft.NTId = i
				hft.TId = 0
				hft.UserWithOnline = make([]FTUsers, bf.Rule.PlayerNum)
				hft.DataLock = new(lock2.RWMutex)
				hft.CreateStamp = time.Now().UnixNano()
				bf.Tables[i] = hft
			}
		}
		// go nhfloor.RunOpt()
		if !bf.IsMix || !nhouse.DBClub.MixActive {
			go bf.StartSync()
		}
	}
	syncFloorTable(nhouse)
	hm := GetClubMgr()
	hm.lock.CustomLock()
	hm.hidLock.CustomLock()
	hm.ClubMap[nhouse.DBClub.HId] = nhouse
	hm.ClubIdToHidMap[nhouse.DBClub.Id] = nhouse.DBClub.HId
	hm.hidLock.CustomUnLock()
	hm.lock.CustomUnLock()
	nhouse.AddTableLimitLock.Lock()
	initHouseTableLimit(nhouse)
	nhouse.AddTableLimitLock.Unlock()
	house.IsAlive = false
	nhouse.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorRenameNTF, nil) //让客户端重新进楼
	return nil, xerrors.RespOk
}

func GetShareCfg(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_GetShareCfg)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	// 获取用户
	person, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
	if person == nil {
		return nil, xerrors.UserNotExistError
	}
	// 平台标识 区分app与小程序
	if person.Platform == consts.PlatformWechatApplet {
		req.Platform = 2
	} else {
		req.Platform = 1
	}
	// 获取分享配置
	db := GetDBMgr().GetDBmControl()
	shareCfg, err := models.GetConfigShare(db, req.SceneId, req.Platform, req.KindId, req.SiteType)
	if err != nil {
		xlog.Logger().Errorf("[获取分享配置失败]错误原因:%v, 请求参数：scene_id = %d, platform = %d kind_id = %d site_type = %d", err, req.SceneId, req.Platform, req.KindId, req.SiteType)
		return nil, xerrors.DBExecError
	}

	// 获取已分享的次数
	shareCnt, err := models.GetShareHistoryCnt(db, req.Uid, shareCfg.Id, time.Now())
	if err != nil {
		xlog.Logger().Errorf("[获取已经分享次数失败]错误原因:%v", err)
		return nil, xerrors.DBExecError
	}

	result := &static.Msg_Http2C_ShareCfg{
		ShareId:           shareCfg.Id,
		SceneId:           shareCfg.SceneId,
		ShareTo:           shareCfg.ShareTo,
		ShareType:         shareCfg.ShareType,
		ShareTimes:        shareCfg.ShareTimes,
		Title:             shareCfg.Title,
		Content:           shareCfg.Content,
		ImgDownload:       shareCfg.ImgDownload,
		AlreadyShareTimes: shareCnt,
	}

	// 随机一个跳转域名地址
	linkIdx := static.HF_GetRandom(100) % 3
	if linkIdx == 0 {
		result.Link = shareCfg.Link1
	} else if linkIdx == 1 {
		result.Link = shareCfg.Link2
	} else if linkIdx == 2 {
		result.Link = shareCfg.Link3
	}

	// 解析分享奖励
	if len(shareCfg.Reward) > 0 {
		if err = json.Unmarshal([]byte(shareCfg.Reward), &result.Reward); err != nil {
			xlog.Logger().Errorf("[解析分享奖励id = %d失败]错误原因:%v", shareCfg.Id, err)
		}
	}

	return result, xerrors.RespOk
}

func ShareSuc(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_ShareSuc)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	// 获取分享配置
	db := GetDBMgr().GetDBmControl()
	shareCfg, err := models.GetConfigShareById(db, req.ShareId)
	if err != nil {
		xlog.Logger().Errorf("[获取分享配置失败]错误原因:%v, 请求参数：share_id = ", err, req.ShareId)
		return nil, xerrors.DBExecError
	}

	// 获取已分享的次数
	shareCnt, err := models.GetShareHistoryCnt(db, req.Uid, shareCfg.Id, time.Now())
	if err != nil {
		xlog.Logger().Errorf("[获取已经分享次数失败]错误原因:%v", err)
		return nil, xerrors.DBExecError
	}

	user, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
	if err != nil {
		xlog.Logger().Errorf("[分享领奖][查询用户失败]用户ID:%d, 错误原因err:%v", req.Uid, err)
		return nil, xerrors.DBExecError
	}

	// 解析分享奖励
	var shareReward static.ShareRewardList
	if err = json.Unmarshal([]byte(shareCfg.Reward), &shareReward); err != nil {
		xlog.Logger().Errorf("[解析分享奖励失败]错误原因:%v", err)
	}

	// 返回数据
	result := &static.Msg_Http2C_ShareSuc{
		Gold:     user.Gold,
		Diamond:  user.Diamond,
		Card:     user.Card,
		GoldBean: user.GoldBean,
	}

	// 此分享没有奖励
	if len(shareReward) == 0 {
		result.Reward = make(static.ShareRewardList, 0)
		return result, xerrors.RespOk
	}

	// 分享次数已达上限 不再奖励
	if shareCfg.ShareTimes > 0 && shareCnt >= shareCfg.ShareTimes {
		result.Reward = make(static.ShareRewardList, 0)
		return result, xerrors.RespOk
	}

	// 发放奖励
	var (
		afterNum int64
	)
	for _, reward := range shareReward {
		afterNum, err = updateWealth(req.Uid, int8(reward.WealthType), reward.Num, models.CostTaskReward)
		if err != nil {
			return nil, xerrors.DBExecError
		}
	}

	// 记录分享
	shareHistory := new(models.ShareHistory)
	shareHistory.Uid = req.Uid
	shareHistory.ShareId = req.ShareId
	shareHistory.CreateAt = time.Now()
	if err = db.Create(&shareHistory).Error; err != nil {
		xlog.Logger().Errorf("[分享领奖][记录分享记录异常]err:%v", err)
	}

	// 更新分享返回信息
	result.Reward = shareReward
	result.Gold = user.Gold
	result.Diamond = user.Diamond
	result.Card = user.Card
	result.GoldBean = user.GoldBean
	result.Vitamin = static.SwitchVitaminToF64(afterNum)

	//// 通知游戏服
	//if user.GameId > 0 {
	//	GetServer().CallGame(user.GameId, user.Uid, "NewServerMsg", consts.MsgTypeUserScoreUpdate, xerrors.SuccessCode, &static.Msg_Update_Table_user_score{Uid: user.Uid, TableId: user.TableId})
	//}

	return result, xerrors.RespOk
}

func syncFloorTable(house *Club) error {
	// 楼层牌桌 DB数据初始化
	datas, err := GetDBMgr().GetDBrControl().FloorTableBaseReLoad()
	if err != nil {
		return err
	}
	for _, data := range datas {
		nftable := data.(*static.FloorTable)
		if nftable.DHId != house.DBClub.Id {
			continue
		}

		floor := house.GetFloorByFId(nftable.FId)
		if floor == nil {
			continue
		}
		table := GetTableMgr().GetTable(nftable.TId)
		if table == nil {
			continue
		}
		hft := new(HouseFloorTable)
		hft.TId = table.Id
		hft.NTId = nftable.NTId
		hft.DataLock = new(lock2.RWMutex)
		hft.Begin = table.Begin
		hft.Step = table.Step
		hft.UserWithOnline = make([]FTUsers, table.Config.MaxPlayerNum)
		hft.CreateStamp = table.Table.CreateStamp
		for i := 0; i < len(table.Users); i++ {
			if table.Users[i] != nil {
				p, err := GetDBMgr().GetDBrControl().GetPerson(table.Users[i].Uid)
				if err != nil || p == nil {
					continue
				}
				hft.UserWithOnline[i] = FTUsers{Uid: table.Users[i].Uid, OnLine: p.Online, Ready: table.Users[i].Ready}
			}
		}
		floor.DataLock.Lock()
		floor.Tables[nftable.NTId] = hft
		floor.DataLock.Unlock()
	}
	return nil
}

func initHouseTableLimit(house *Club) {
	sql := `select hid,group_id,uid,updated_at from house_table_limit_user where status = 0 and hid = ?`
	dest := []HTableLimitDb{}
	err := GetDBMgr().GetDBmControl().Raw(sql, house.DBClub.Id).Scan(&dest).Error
	if err != nil {
		panic(err)
	}
	for _, item := range dest {
		if house.LimitGroupsUpdateAt == nil {
			house.LimitGroupsUpdateAt = make(map[int]int64)
		}
		if house.LimitGroups == nil {
			house.LimitGroups = make(map[int]map[int64]bool)
		}
		if house.LimitGroups[item.GroupId] == nil {
			house.LimitGroups[item.GroupId] = make(map[int64]bool)
		}
		if item.Uid == 0 {
			house.LimitGroupsUpdateAt[item.GroupId] = item.UpdateAt.Unix()
			continue
		}
		house.LimitGroups[item.GroupId][item.Uid] = true
		house.TableLimitGropuCount = len(house.LimitGroups)
	}
}

// 获取低保  包括低保补助及低保礼包
func GetAllowanceInfo(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_GetAllowanceInfo)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, xe := TokenCheckout(req.Uid, req.Token)
	if xe != nil {
		return nil, xe
	}
	var (
		ret interface{}
		err error
	)
	nowStr := time.Now().Format("2006-01-02")
	// 屏蔽破产礼包
	if true || req.IgnoreGift {
		// 得到用户的低保
		ret, err = wealthtalk.GetUserAllowanceWithoutGift(
			GetDBMgr().GetDBmControl(),
			GetDBMgr().GetDBrControl(),
			p.Uid,
			nowStr,
			GetServer().ConServers,
		)
	} else {
		// 得到用户的低保
		ret, err = wealthtalk.GetUserAllowance(
			GetDBMgr().GetDBmControl(),
			GetDBMgr().GetDBrControl(),
			p.Uid,
			nowStr,
			GetServer().ConServers,
		)
	}
	if err != nil {
		xlog.Logger().Error(err)
		return nil, xerrors.DBExecError
	}
	var resp static.Msg_S2C_AllowancesInfo
	switch x := ret.(type) {
	case *static.AllowanceGift: // 低保礼包
		x.UserGold = p.Gold
		x.UserDiamond = p.Diamond
		resp.Gift = x
	case *static.Msg_S2C_Allowances: // 包括低保
		resp.Allowance = x
	}

	// 如果没有破产礼包
	if resp.Gift == nil {
		// 如果既没有破产礼包又没有破产补助
		if resp.Allowance == nil {
			return nil, xerrors.GoldNotEnoughError
		} else {
			// 如果大厅在线，则通知客户端更新金币
			person := GetPlayerMgr().GetPlayer(p.Uid)
			if person != nil {
				person.Info.Gold = resp.Allowance.AfterGold
				person.UpdGold(models.CostTypeAllowances, resp.Allowance.Gold)
			}
			xlog.Logger().Infof("玩家领取低保%d, 通知其到账成功", p.Uid)
			// 如果玩家在游戏服，则通知游戏服更新金币
			if p.GameId > 0 {
				xlog.Logger().Warningf("玩家%d在游戏服%d, 通知其到账成功", p.Uid, p.GameId)
				reply, err := GetServer().CallGame(p.GameId, p.Uid, "NewServerMsg", consts.MsgTypeUserGoldUpdate, xerrors.SuccessCode, static.Msg_UpdateGold{Uid: p.Uid, CostType: models.CostTypeAllowances, Offset: resp.Allowance.Gold})
				xlog.AnyErrors(err, "玩家在游戏服金币不足，通过http领取低保后，通知游戏服更新玩家金币信息失败")
				if r := string(reply); r != "SUC" {
					xlog.Logger().Error("通知游戏服更新玩家金币信息失败", r)
				}
			}
		}
	}
	return &resp, nil
}

// 检查玩家 今日是否已购买破产礼包
func CheckUserPurchasedBankruptGiftToday(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_TokenUid)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	p, xe := TokenCheckout(req.Uid, req.Token)
	if xe != nil {
		return nil, xe
	}
	nowStr := time.Now().Format("2006-01-02")
	buy, err := models.IsAllowanceGiftPurchasedToday(GetDBMgr().GetDBmControl(), p.Uid, nowStr)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, xerrors.DBExecError
	}
	if buy {
		return "今日已购买破产礼包，不可多次购买。", xerrors.UserPurchasedBankruptGiftError
	} else {
		return &static.Msg_Null{}, nil
	}
}

// 重置玩家对应包厢的权限
func ResetUserHmRight(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_Http_ResetUserRight)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	key := strconv.Itoa(int(req.Dhid)) + "," + strconv.Itoa(int(req.Uid))
	GetMRightMgr().DeleteRight(key)                            // 删除map权限数据
	GetMRightMgr().deleteRightByHidUid(int(req.Dhid), req.Uid) //删除数据库权限数据
	return nil, xerrors.RespOk
}

const HotVersionKey = "HotVersion"

// 记录热更新版本号 并推送
func ForceHotter(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_Http_ForceHotter)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	GetDBMgr().GetDBrControl().RedisV2.Set(HotVersionKey, req.Version, 0)
	GetPlayerMgr().Broadcast(consts.MsgHotUpdate_Ntf, req)
	return nil, xerrors.RespOk
}

// 注销用户
func WriteOffUser(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_Http_WriteOffUser)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	uid := req.Uid
	items, err := GetDBMgr().ListHouseIdMemberJoin(uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError
	}
	if items != nil {
		cuserror := xerrors.NewXError("还存在未退出的包厢，不允许注销账号")
		return xerrors.ResultErrorCode, cuserror
	}
	//数据库状态 设置为 注销状态
	db := GetDBMgr().GetDBmControl()
	dbR := GetDBMgr().GetDBrControl()
	accType := 1 // 1为注销账号的状态值
	var dbErr error
	tx := db.Begin()
	//设置账号为注销类型
	if dbErr = tx.Model(models.User{Id: uid}).Update("account_type", accType).Error; err != nil {
		xlog.Logger().Errorln(dbErr)
		cuserror := xerrors.NewXError("账号注销失败")
		tx.Rollback()
		return xerrors.ResultErrorCode, cuserror
	}

	//查找house_member表 是否有urole 为3的 也就是申请状态 如果有则需要删除  并且根据 hid 去redis 中的 house_member 数据删除
	var hmModel []models.HouseMember
	isHmDelete := true
	if dbErr = GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("uid = ? and urole = 3", uid).Find(&hmModel).Error; err != nil {
		if err != gorm.ErrRecordNotFound { //这里没有数据存一条默认数据
			xlog.Logger().Errorln(dbErr)
			cuserror := xerrors.NewXError("账号注销失败")
			tx.Rollback()
			return xerrors.ResultErrorCode, cuserror
		} else {
			isHmDelete = false
		}
	}

	if isHmDelete {
		if dbErr = tx.Where("uid = ? and urole = 3", uid).Delete(models.HouseMember{}).Error; err != nil {
			xlog.Logger().Errorln(dbErr)
			cuserror := xerrors.NewXError("账号注销失败")
			tx.Rollback()
			return xerrors.ResultErrorCode, cuserror
		}
		for i := 0; i < len(hmModel); i++ {
			dbErr = dbR.HouseMemberDelete(hmModel[i].ConvertModel())
			if dbErr != nil {
				xlog.Logger().Errorln("delet user housemember to redis error: ", err.Error())
				cuserror := xerrors.NewXError("账号注销失败")
				tx.Rollback()
				return xerrors.ResultErrorCode, cuserror
			}
		}
	}

	// 更新redis
	dbErr = dbR.UpdatePersonAttrs(uid, "AccountType", accType)
	if dbErr != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		cuserror := xerrors.NewXError("账号注销失败")
		tx.Rollback()
		return xerrors.ResultErrorCode, cuserror
	}
	tx.Commit()
	return nil, xerrors.RespOk
}

func SaveUserPlayTime(r *http.Request, data interface{}) (interface{}, *xerrors.XError) {
	req, ok := data.(*static.Msg_C2Http_PlayTime)
	if !ok {
		return nil, xerrors.ArgumentError
	}

	// 未实名 或 已实名未成年 记录累计时长
	personRedis, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
	if personRedis == nil {
		return nil, xerrors.UserNotExistError
	}
	// 已实名且成年的认证用户不需要记录时长
	if len(personRedis.IdcardAuthPI) > 0 && static.HF_GetAgeFromIdcard(personRedis.Idcard) >= 18 {
		return nil, nil
	}

	// 更新时长
	db := GetDBMgr().GetDBmControl()
	playTime, err := models.UpdateUserPlayTime(db, req.Uid, time.Now(), req.TimeSec)
	if err != nil {
		return nil, xerrors.DBExecError
	}

	// 更新缓存数据
	if err := GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "PlayTime", playTime); err != nil {
		return nil, xerrors.DBExecError
	}

	// 返回数据
	result := &static.Msg_Http2C_PlayTime{
		TimeSec: playTime,
	}

	return result, xerrors.RespOk
}
