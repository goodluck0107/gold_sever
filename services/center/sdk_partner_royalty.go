package center

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	dbModels "github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sort"
	"time"
)

const (
	optUserTypeCreate       = 0 //圈住修改
	optUserTypeVitaminAdmin = 1 //比赛分管理员修改
	optUserTypePartner      = 2 //队长修改
)

const (
	optTypeAddFloor      = 0 //增加楼层
	optTypeDelFloor      = 1 //删除楼层
	optTypeModifyRoyalty = 2 //修改队长分层配置
)

const (
	// 默认的队长收益值
	DefaultPartnerProfit = -1
	// 无效的房费扣除
	InvalidVitaminCost = -1
	// 队长收益值合法值下限
	PartnerProfitLowerLimit = 0
	// 队长收益值百分比上限
	PartnerPercentHigherLimit = 100
)

type ClubFloorVitaminCost map[int64]int64

func (hf ClubFloorVitaminCost) ToList() util.Int64Slice {
	keys := make(util.Int64Slice, len(hf))
	for fid := range hf {
		keys = append(keys, fid)
	}
	sort.Sort(keys)
	values := make(util.Int64Slice, len(hf))
	for i := 0; i < len(keys); i++ {
		values = append(values, hf[keys[i]])
	}
	return values
}

func GetStatisticsMap(memUid int64, sMap *map[int64]map[int64]*static.ClubPartnerFloorStatisticsItem, floorId int64) *static.ClubPartnerFloorStatisticsItem {
	_, ok := (*sMap)[memUid]
	if !ok {
		(*sMap)[memUid] = make(map[int64]*static.ClubPartnerFloorStatisticsItem)
	}

	if floorId == -1 {
		return nil
	}

	_, ok = (*sMap)[memUid][floorId]
	if !ok {
		statisticsItem := new(static.ClubPartnerFloorStatisticsItem)
		statisticsItem.UId = memUid
		statisticsItem.ValidTimes = 0
		statisticsItem.BigValidTimes = 0
		statisticsItem.RoundProfit = 0
		statisticsItem.SubordinateProfit = 0
		statisticsItem.TotalProfit = 0

		(*sMap)[memUid][floorId] = statisticsItem
	}

	return (*sMap)[memUid][floorId]
}

func GetDelFloorInfo(dhid int64, dfindex int) (map[int64]dbModels.HouseFloorDelMsg, error) {
	delMsgMap := make(map[int64]dbModels.HouseFloorDelMsg)

	beginTime := static.GetZeroTime(time.Now().AddDate(0, 0, -7)).Unix()

	var delMsg []dbModels.HouseFloorDelMsg
	if dfindex == -1 {
		err := GetDBMgr().GetDBmControl().Model(dbModels.HouseFloorDelMsg{}).
			Where("dhid = ? and create_stamp >= ?", dhid, beginTime).Find(&delMsg).Error
		if err != nil {
			return delMsgMap, err
		}
	} else {
		err := GetDBMgr().GetDBmControl().Model(dbModels.HouseFloorDelMsg{}).
			Where("dhid = ? and dfindex = ? and create_stamp >= ?", dhid, dfindex, beginTime).Find(&delMsg).Error
		if err != nil {
			return delMsgMap, err
		}
	}

	for _, item := range delMsg {
		delMsgMap[item.DFId] = item
	}
	return delMsgMap, nil
}

func GetDelFloorInfoWithTime(dhid int64, beginTime int64) (map[int64]dbModels.HouseFloorDelMsg, map[int64]dbModels.HouseFloorDelMsg, error) {
	oldDelMsgMap := make(map[int64]dbModels.HouseFloorDelMsg)
	newDelMsgMap := make(map[int64]dbModels.HouseFloorDelMsg)
	var delMsg []dbModels.HouseFloorDelMsg
	err := GetDBMgr().GetDBmControl().Model(dbModels.HouseFloorDelMsg{}).
		Where("dhid = ? and create_stamp >= ?", dhid, beginTime).Find(&delMsg).Error
	if err != nil {
		return oldDelMsgMap, newDelMsgMap, err
	}

	for _, item := range delMsg {
		if item.FloorRoyalty != "" {
			oldDelMsgMap[item.DFId] = item
		} else {
			newDelMsgMap[item.DFId] = item
		}
	}
	return oldDelMsgMap, newDelMsgMap, nil
}

func GetDelFloorInfoWithFid(dhid int64, dfid int64) (map[int64]dbModels.HouseFloorDelMsg, map[int64]dbModels.HouseFloorDelMsg, error) {
	oldDelMsgMap := make(map[int64]dbModels.HouseFloorDelMsg)
	newDelMsgMap := make(map[int64]dbModels.HouseFloorDelMsg)
	var delMsg []dbModels.HouseFloorDelMsg
	err := GetDBMgr().GetDBmControl().Model(dbModels.HouseFloorDelMsg{}).
		Where("dhid = ? and dfid = ?", dhid, dfid).Find(&delMsg).Error
	if err != nil {
		return oldDelMsgMap, newDelMsgMap, err
	}

	for _, item := range delMsg {
		if item.FloorRoyalty != "" {
			oldDelMsgMap[item.DFId] = item
		} else {
			newDelMsgMap[item.DFId] = item
		}
	}
	return oldDelMsgMap, newDelMsgMap, nil
}

func NewPartnerRoyaltyModifyHistory(dhid int64, optUser *HouseMember, optType int,
	optFloorName string, optFloorId int64, optFloorIndex int, beOptUser int64, before int, after int) *dbModels.HousePartnerRoyaltyHistory {
	history := new(dbModels.HousePartnerRoyaltyHistory)
	history.DHid = dhid
	history.OptUser = optUser.UId
	if optUser.URole == consts.ROLE_CREATER {
		history.OptUserType = optUserTypeCreate
	} else if optUser.IsVitaminAdmin() {
		history.OptUserType = optUserTypeVitaminAdmin
	} else if optUser.IsPartner() {
		history.OptUserType = optUserTypePartner
	}

	history.OptType = optType
	history.OptFloorName = optFloorName
	history.OptFloorId = optFloorId
	history.OptFloorIndex = optFloorIndex
	history.BeOptUser = beOptUser
	history.Before = before
	history.After = after
	history.CreatedAt = time.Now()
	return history
}

func AddPartnerRoyaltyModifyHistory(dhid int64, optUser *HouseMember, optType int,
	optFloorName string, optFloorId int64, optFloorIndex int, beOptUser int64, before int, after int) {
	history := NewPartnerRoyaltyModifyHistory(dhid, optUser, optType, optFloorName, optFloorId, optFloorIndex, beOptUser, before, after)
	GetDBMgr().GetDBmControl().Save(history)
}

func AddPartnerRoyaltyModifyHistorys(historys []*dbModels.HousePartnerRoyaltyHistory) {
	tx := GetDBMgr().GetDBmControl().Begin()
	for _, history := range historys {
		tx.Save(history)
	}
	tx.Commit()
}

func GetRoyaltyModifyHistoryForPartner(dhid int64, userId int64) ([]dbModels.HousePartnerRoyaltyHistory, error) {
	var history []dbModels.HousePartnerRoyaltyHistory
	err := GetDBMgr().GetDBmControl().Model(dbModels.HousePartnerRoyaltyHistory{}).
		Where("dhid = ? and ((opttype = 2 and beoptuser = ?) or opttype = 0 or opttype = 1) and created_at >= date_add(now(),interval -1 month)", dhid, userId).
		Order("created_at desc").
		Limit(50).
		Find(&history).Error
	return history, err
}

func GetRoyaltyModifyHistoryOptInfo(optUserType, optType, before, after int) (string, string) {
	userTypeStr := ""
	if optUserType == optUserTypeCreate {
		userTypeStr = "盟主"
	} else if optUserType == optUserTypeVitaminAdmin {
		userTypeStr = "比赛分管理员"
	} else if optUserType == optUserTypePartner {
		userTypeStr = "上级队长"
	} else {
		userTypeStr = "未知角色"
	}

	optInfo := ""
	if optType == optTypeAddFloor {
		optInfo = "增加楼层"
	} else if optType == optTypeDelFloor {
		optInfo = "删除楼层"
	} else if optType == optTypeModifyRoyalty {
		if before < PartnerProfitLowerLimit {
			before = 0
		}
		if after < PartnerProfitLowerLimit {
			after = 0
		}
		optInfo = fmt.Sprintf("调整比例 %d%%->%d%%", before, after)
	} else {
		optInfo = "未知操作"
	}
	return userTypeStr, optInfo
}

// 更新队长单局收益百分比
func UpdateRoyaltyPercent(tx *gorm.DB, memMap static.HouseMemberMap, cur *dbModels.HousePartnerPyramid, nowRp int) (err error) {
	// 百分比计算及容错
	if nowRp > PartnerPercentHigherLimit {
		nowRp = PartnerPercentHigherLimit
	}

	// 更新百分比
	if err = tx.Model(dbModels.HousePartnerPyramid{}).
		Where("dhid = ? and dfid = ? and uid = ?", cur.DHid, cur.DFid, cur.Uid).
		Update("royalty_percent", nowRp).Error; err != nil {
		return
	}

	// 如果选择重置收益
	if nowRp < 0 {
		// 如果玩家重置百分比 则其所有下级奖没有可配置总额
		return updateJuniorsTotalByReset(tx, memMap, cur.DHid, cur.DFid, cur.Uid)
	} else {
		// 如果之前的百分比大于0 则可以按照等比变化来更新所有下级的百分比
		if cur.RoyaltyPercent > 0 {
			return updateJuniorsTotalByRpOffset(tx, memMap, cur.DHid, cur.DFid, cur.Uid, cur.RoyaltyPercent, nowRp)
		} else {
			// 如果之前是0 则无法按照等比变化来更新，只能逐级重新计算
			cur.RoyaltyPercent = nowRp
			return updateJuniorsTotalByRpDevelop(tx, memMap, cur.DHid, cur.DFid, cur.Uid, cur)
		}
	}
}

func updateJuniorsTotalByReset(tx *gorm.DB, memMap static.HouseMemberMap, hid, fid, uid int64) (err error) {
	// 下一级节点组
	juniors := make([]int64, 0)
	// 遍历关系链
	for next := memMap.JuniorsBySuperiors(uid); len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		juniors = append(juniors, next...)
	}
	// 等比更新
	sql := "update house_partner_pyramid set total = ? where dhid = ? and dfid = ? and uid in(?)"
	return tx.Exec(sql, DefaultPartnerProfit, hid, fid, juniors).Error
}

// 通过单局收益百分比的变化量来更新其所有下级链的可分配总额
func updateJuniorsTotalByRpOffset(tx *gorm.DB, memMap static.HouseMemberMap, hid, fid, uid int64, before, now int) (err error) {
	// 下一级节点组
	juniors := make([]int64, 0)
	// 遍历关系链
	for next := memMap.JuniorsBySuperiors(uid); len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		juniors = append(juniors, next...)
	}
	// 等比更新
	if len(juniors) > 0 {
		sql := "update house_partner_pyramid set " +
			"total = (case when total >= 0 then total * ?/ ? end) where dhid = ? and dfid = ? and uid in(?)"
		return tx.Exec(sql, now, before, hid, fid, juniors).Error
	}
	return nil
}

// 通过单局收益的从无到有（从0到非0）更新其所有下级链的可分配总额
func updateJuniorsTotalByRpDevelop(tx *gorm.DB, memMap static.HouseMemberMap, hid, fid, uid int64, top *dbModels.HousePartnerPyramid) (err error) {
	resultSet := make(map[int64]*dbModels.HousePartnerPyramid)
	resultSet[uid] = top
	for next := memMap.JuniorsBySuperiors(uid); len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		nextConfig := make([]*dbModels.HousePartnerPyramid, 0)
		if err = tx.Where("dhid = ? and dfid = ? and uid in(?)", hid, fid, next).Find(&nextConfig).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return
		}
		for i := 0; i < len(nextConfig); i++ {
			cur := nextConfig[i]
			mem, ok := memMap[cur.Uid]
			if !ok {
				continue
			}
			if mem == nil {
				continue
			}
			sup, ok := resultSet[mem.Superior]
			if !ok {
				continue
			}
			if sup != nil {
				cur.Total = sup.Total * sup.RealRoyaltyPercent() / PartnerPercentHigherLimit
			}
			resultSet[cur.Uid] = cur
			if err = tx.Model(dbModels.HousePartnerPyramid{}).
				Where("dhid = ? and dfid = ? and uid = ?", hid, fid, cur.Uid).
				Update("total", cur.Total).Error; err != nil {
				return
			}
		}
	}
	return
}

// 得到包厢指定楼层（包含已删除）的所有合伙人的单局收益
func GetHouseFloorPartnersRoyalty(hid, fid int64) map[int64]int64 /*uid 对应 royalty*/ {
	floorPyramid := GetHouseFloorWholePartnersPyramid(hid, fid)
	res := make(map[int64]int64)
	for uid, pyramid := range floorPyramid {
		var royalty int64 = DefaultPartnerProfit
		if pyramid.Configurable() && pyramid.ConfiguredRoyaltyPercent() {
			royalty, _ = pyramid.EarningsInfo()
		}
		res[uid] = royalty
	}
	return res
}

// 得到包厢所有合伙人在所有楼层（包含已删除的）的单局收益
func GetClubPartnersRoyalty(hid int64) map[int64]map[int64]int64 /*uid  对应 fids 对应 royalty*/ {
	housePyramid := GetHouseWholePartnersPyramid(hid)
	res := make(map[int64]map[int64]int64)
	for uid, floorPyramid := range housePyramid {
		res[uid] = make(map[int64]int64)
		for i := 0; i < len(floorPyramid); i++ {
			pyramid := floorPyramid[i]
			var royalty int64 = DefaultPartnerProfit
			if pyramid.Configurable() && pyramid.ConfiguredRoyaltyPercent() {
				royalty, _ = pyramid.EarningsInfo()
			}
			res[uid][pyramid.DFid] = royalty
		}
	}
	return res
}

// 得到包厢某个楼层的所有的队长配置
func GetHouseFloorWholePartnersPyramid(hid, fid int64) map[int64]*dbModels.HousePartnerPyramid {
	models := make(dbModels.HousePartnerPyramidFloors, 0)
	err := GetDBMgr().GetDBmControl().Where("dhid = ? and dfid = ?", hid, fid).Find(&models).Error
	if err != nil {
		xlog.Logger().Error(err)
	}
	return models.ToMapFloor(fid)
}

// 得到整个包厢队长的分成配置
func GetHouseWholePartnersPyramid(hid int64) map[int64]dbModels.HousePartnerPyramidFloors {
	models := make(dbModels.HousePartnerPyramidFloors, 0)
	err := GetDBMgr().GetDBmControl().Where("dhid = ?", hid).Find(&models).Error
	if err != nil {
		xlog.Logger().Error(err)
	}
	return models.ToMapFloors()
}

// 得到包厢部分队长的分成配置
func GetHousePartPartnersPyramid(hid int64, ids ...int64) map[int64]dbModels.HousePartnerPyramidFloors {
	models := make(dbModels.HousePartnerPyramidFloors, 0)
	err := GetDBMgr().GetDBmControl().Where("dhid = ? and uid in(?)", hid, ids).Find(&models).Error
	if err != nil {
		xlog.Logger().Error(err)
	}
	return models.ToMapFloors()
}

// 得到包厢单个队长的分成配置
func GetHousePartnerPyramid(hid, uid int64) dbModels.HousePartnerPyramidFloors {
	models := make(dbModels.HousePartnerPyramidFloors, 0)
	err := GetDBMgr().GetDBmControl().Where("dhid = ? and uid = ?", hid, uid).Find(&models).Error
	if err != nil {
		xlog.Logger().Error(err)
	}
	return models
}

// 得到包厢成员 在指定楼层 打一把他的每级队长该拿多少钱
func GetFloorRoyaltyByProvider(hid, fid, provider int64) map[int64] /*uid*/ int64 /*收益*/ /*每个上级对应楼层的收益*/ {
	res := make(map[int64]int64)
	memMap := GetDBMgr().GetHouseMemMap(hid)
	configMap := GetHouseFloorWholePartnersPyramid(hid, fid)
	member, ok := memMap[provider]
	for ok && member != nil {
		if member.IsPartner() { // 如果这个提供者是队长
			var royalty, superRoyalty int64
			config, ok1 := configMap[member.UId]
			if ok1 {
				royalty, superRoyalty = config.EarningsInfo()
			}
			if member.UId == provider {
				res[member.UId] = royalty
			}
			if member.Superior > 0 {
				res[member.Superior] = superRoyalty
			}
			member, ok = memMap[member.Superior]
		} else if member.Partner > 0 { // 如果提供者是某个队长名下成员
			provider = member.Partner
			member, ok = memMap[member.Partner]
		} else { // 既不是队长 也不是队长名下玩家，则无法计算收益
			break
		}
	}
	return res
}

// 得到包厢成员 在所有楼层 打一把他的每级队长该拿多少钱
func GetClubRoyaltyByProvider(hid int64, fidList []int64, provider int64) map[int64] /*uid*/ map[int64] /*fid*/ int64 /*收益*/ /*每个上级每个楼层的收益*/ {
	res := make(map[int64]map[int64]int64)
	memMap := GetDBMgr().GetHouseMemMap(hid)
	configMap := GetHouseWholePartnersPyramid(hid)

	res = GetClubMemberRoyalty(provider, memMap, fidList, configMap)

	return res
}

// 得到所有队长所在楼层打一把他的每级合人该那多少钱
func GetClubRoyaltyByPartner(hid int64, fidList []int64) (map[int64] /*partner*/ map[int64] /*uid*/ map[int64] /*fid*/ int64 /*收益*/ /*每个上级每个楼层的收益*/, map[int64]int64) {
	pRes := make(map[int64]map[int64]map[int64]int64)
	pPartner := make(map[int64]int64)
	memMap := GetDBMgr().GetHouseMemMap(hid)
	configMap := GetHouseWholePartnersPyramid(hid)

	for _, member := range memMap {
		if member.IsPartner() {
			pRes[member.UId] = GetClubMemberRoyalty(member.UId, memMap, fidList, configMap)
			pPartner[member.UId] = member.UId
		} else if member.Partner > 1 {
			pPartner[member.UId] = member.Partner
		}
	}
	return pRes, pPartner
}

// 得到玩家分层数据
func GetClubMemberRoyalty(provider int64, memMap static.HouseMemberMap, fidList []int64, configMap map[int64]dbModels.HousePartnerPyramidFloors) map[int64]map[int64]int64 {
	res := make(map[int64]map[int64]int64)
	member, ok := memMap[provider]
	for ok && member != nil {
		if member.IsPartner() { // 如果这个提供者是队长
			for i := 0; i < len(fidList); i++ {
				fid := fidList[i]
				var royalty, superRoyalty int64
				allConfig, ok1 := configMap[member.UId]
				if ok1 {
					config := allConfig.GetPyramidByFid(fid)
					if config != nil {
						royalty, superRoyalty = config.EarningsInfo()
					}
				}
				if member.UId == provider {
					fidRoyalty, ok2 := res[member.UId]
					if !ok2 {
						fidRoyalty = make(map[int64]int64)
					}
					fidRoyalty[fid] = royalty
					res[member.UId] = fidRoyalty
				}
				if member.Superior > 0 {
					fidRoyalty, ok2 := res[member.Superior]
					if !ok2 {
						fidRoyalty = make(map[int64]int64)
					}
					fidRoyalty[fid] = superRoyalty
					res[member.Superior] = fidRoyalty
				}
			}
			member, ok = memMap[member.Superior]
		} else if member.Partner > 0 { // 如果提供者是某个队长名下成员
			provider = member.Partner
			member, ok = memMap[member.Partner]
		} else { // 既不是队长 也不是队长名下玩家，则无法计算收益
			break
		}
	}
	return res
}

func FixClubPartnersPyramidForTop(configs *dbModels.HousePartnerPyramidFloors, hid, uid int64, houseFloorIds []int64, costs map[int64]int64) error {
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		// 得到原有的配置
		curConfig := configs.GetPyramidByFid(fid)
		if curConfig == nil { // 如果原来没配置 那么此刻则生成配置
			// 调用生成配置
			curConfig = dbModels.HousePartnerPyramidDefault(hid, fid, uid, i)
			if err := GetDBMgr().GetDBmControl().Create(curConfig).Error; err != nil {
				return err
			}
			if cost, ok := costs[fid]; ok {
				curConfig.Total = cost
			}
			*configs = append(*configs, curConfig)
			if err := GetDBMgr().GetDBmControl().Save(curConfig).Error; err != nil {
				return err
			}
		} else {
			var flag bool
			if curConfig.RoyaltyPercent >= 0 {
				if PartnerPercentHigherLimit < curConfig.RoyaltyPercent {
					curConfig.RoyaltyPercent = PartnerPercentHigherLimit
					flag = true
				}
			} else {
				if curConfig.RoyaltyPercent != DefaultPartnerProfit {
					curConfig.RoyaltyPercent = DefaultPartnerProfit
					flag = true
				}
			}
			if flag {
				if err := GetDBMgr().GetDBmControl().Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", curConfig.DHid, curConfig.DFid, curConfig.Uid).
					Update("royalty_percent", curConfig.RoyaltyPercent).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}


func FixClubPartnersPyramidBySuperSuper(cur, sup, ssup *dbModels.HousePartnerPyramidFloors, hid, uid int64, houseFloorIds []int64) error {
	// 修复
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		ssupConfig := ssup.GetPyramidByFid(fid)
		supConfig := sup.GetPyramidByFid(fid)
		// 得到原有的配置
		curConfig := cur.GetPyramidByFid(fid)
		if supConfig == nil {
			supConfig = dbModels.HousePartnerPyramidDefault(hid, fid, uid, i)
			if ssupConfig != nil && ssupConfig.Configurable() && ssupConfig.ConfiguredRoyaltyPercent() {
				supConfig.Total = ssupConfig.Total * ssupConfig.RealSuperiorPercent() / PartnerPercentHigherLimit
			}
			*sup = append(*sup, supConfig)
			if err := GetDBMgr().GetDBmControl().Create(supConfig).Error; err != nil {
				return err
			}
		} else {
			var flag bool
			if supConfig.RoyaltyPercent >= 0 {
				if PartnerPercentHigherLimit < curConfig.RoyaltyPercent {
					supConfig.RoyaltyPercent = PartnerPercentHigherLimit
					flag = true
				}
			} else {
				if supConfig.RoyaltyPercent != DefaultPartnerProfit {
					supConfig.RoyaltyPercent = DefaultPartnerProfit
					flag = true
				}
			}
			if flag {
				if err := GetDBMgr().GetDBmControl().Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", supConfig.DHid, supConfig.DFid, supConfig.Uid).
					Update("royalty_percent", supConfig.RoyaltyPercent).Error; err != nil {
					return err
				}
			}
		}
		if curConfig == nil { // 如果原来没配置 那么此刻则生成配置
			// 调用生成配置
			curConfig = dbModels.HousePartnerPyramidDefault(hid, fid, uid, i)
			if supConfig != nil && supConfig.Configurable() && supConfig.ConfiguredRoyaltyPercent() { // 如果可配置有问题而上面校验是可以取到可配置额度的则需要生成可配置额度
				// 更改内存数据cur
				curConfig.Total = supConfig.Total * supConfig.RealRoyaltyPercent() / PartnerPercentHigherLimit
			}
			*cur = append(*cur, curConfig)
			if err := GetDBMgr().GetDBmControl().Create(curConfig).Error; err != nil {
				return err
			}
		} else {
			var flag bool
			if curConfig.RoyaltyPercent >= 0 {
				if PartnerPercentHigherLimit < curConfig.RoyaltyPercent {
					curConfig.RoyaltyPercent = PartnerPercentHigherLimit
					flag = true
				}
			} else {
				if curConfig.RoyaltyPercent != DefaultPartnerProfit {
					curConfig.RoyaltyPercent = DefaultPartnerProfit
					flag = true
				}
			}
			if flag {
				if err := GetDBMgr().GetDBmControl().Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", curConfig.DHid, curConfig.DFid, curConfig.Uid).
					Update("royalty_percent", curConfig.RoyaltyPercent).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}


func FixClubPartnersPyramidBySuper(cur, sup *dbModels.HousePartnerPyramidFloors, hid, uid int64, houseFloorIds []int64) error {
	// 修复
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		supConfig := sup.GetPyramidByFid(fid)
		// 得到原有的配置
		curConfig := cur.GetPyramidByFid(fid)
		if curConfig == nil { // 如果原来没配置 那么此刻则生成配置
			// 调用生成配置
			curConfig = dbModels.HousePartnerPyramidDefault(hid, fid, uid, i)
			if supConfig != nil && supConfig.Configurable() && supConfig.ConfiguredRoyaltyPercent() { // 如果可配置有问题而上面校验是可以取到可配置额度的则需要生成可配置额度
				// 更改内存数据cur
				curConfig.Total = supConfig.Total * supConfig.RealRoyaltyPercent() / PartnerPercentHigherLimit
			}
			*cur = append(*cur, curConfig)
			if err := GetDBMgr().GetDBmControl().Create(curConfig).Error; err != nil {
				return err
			}
		} else {
			var flag bool
			if curConfig.RoyaltyPercent >= 0 {
				if PartnerPercentHigherLimit < curConfig.RoyaltyPercent {
					curConfig.RoyaltyPercent = PartnerPercentHigherLimit
					flag = true
				}
			} else {
				if curConfig.RoyaltyPercent != DefaultPartnerProfit {
					curConfig.RoyaltyPercent = DefaultPartnerProfit
					flag = true
				}
			}
			if flag {
				if err := GetDBMgr().GetDBmControl().Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", curConfig.DHid, curConfig.DFid, curConfig.Uid).
					Update("royalty_percent", curConfig.RoyaltyPercent).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func UpdateTopPartnersTotal(tx *gorm.DB, memMap static.HouseMemberMap, hid, fid, cost int64) (err error) {
	resultSet := make(map[int64]*dbModels.HousePartnerPyramid)
	for next := memMap.JuniorsBySuperiors(0); len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		nextConfig := make([]*dbModels.HousePartnerPyramid, 0)
		if err = tx.Where("dhid = ? and dfid = ? and uid in(?)", hid, fid, next).Find(&nextConfig).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return
		}
		for i := 0; i < len(nextConfig); i++ {
			cur := nextConfig[i]
			mem, ok := memMap[cur.Uid]
			if !ok {
				continue
			}
			if mem == nil {
				continue
			}
			before := cur.Total
			if mem.Superior > 0 {
				sup, ok := resultSet[mem.Superior]
				if !ok {
					continue
				}
				if sup != nil {
					if sup.Total < 0 {
						cur.Total = sup.Total
					} else if sup.ConfiguredRoyaltyPercent() {
						cur.Total = sup.Total * sup.RealRoyaltyPercent() / PartnerPercentHigherLimit
					} else {
						cur.Total = DefaultPartnerProfit
					}
				}
			} else {
				cur.Total = cost
			}
			resultSet[cur.Uid] = cur
			if cur.Total != before {
				if err = tx.Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", hid, fid, cur.Uid).
					Update("total", cur.Total).Error; err != nil {
					return
				}
			}
		}
	}
	return
}

func UpdateFloorPartnerTotal(tx *gorm.DB, memMap static.HouseMemberMap, hid, fid, uid, total int64) (err error) {
	resultSet := make(map[int64]*dbModels.HousePartnerPyramid)

	for next := []int64{uid}; len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		nextConfig := make([]*dbModels.HousePartnerPyramid, 0)
		if err = tx.Where("dhid = ? and dfid = ? and uid in(?)", hid, fid, next).Find(&nextConfig).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return
		}
		for i := 0; i < len(nextConfig); i++ {
			cur := nextConfig[i]
			mem, ok := memMap[cur.Uid]
			if !ok {
				continue
			}
			if mem == nil {
				continue
			}

			before := cur.Total

			if mem.UId == uid {
				cur.Total = total
			} else if mem.Superior > 0 {
				sup, ok := resultSet[mem.Superior]
				if !ok {
					continue
				}
				if sup != nil {
					if sup.Total < 0 {
						cur.Total = sup.Total
					} else if sup.ConfiguredRoyaltyPercent() {
						cur.Total = sup.Total * sup.RealRoyaltyPercent() / PartnerPercentHigherLimit
					} else {
						cur.Total = DefaultPartnerProfit
					}
				}
			} else {
				cur.Total = total
			}

			resultSet[cur.Uid] = cur
			if cur.Total != before {
				if err = tx.Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", hid, fid, cur.Uid).
					Update("total", cur.Total).Error; err != nil {
					return
				}
			}
		}
	}
	return
}

func UpdateHousePartnerTotal(tx *gorm.DB, memMap static.HouseMemberMap, hid, uid int64, costs map[int64]int64) (err error) {
	for fid, cost := range costs {
		if err = UpdateFloorPartnerTotal(tx, memMap, hid, fid, uid, cost); err != nil {
			return err
		}
	}
	return
}

func SyncFloorAllPartnerTotal(tx *gorm.DB, memMap dbModels.HouseMembersMap, hid, fid, total int64) (err error) {
	resultSet := make(map[int64]*dbModels.HousePartnerPyramid)

	for next := memMap.JuniorsBySuperiors(0); len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		nextConfig := make([]*dbModels.HousePartnerPyramid, 0)
		if err = tx.Where("dhid = ? and dfid = ? and uid in(?)", hid, fid, next).Find(&nextConfig).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return
		}
		for i := 0; i < len(nextConfig); i++ {
			cur := nextConfig[i]
			mem, ok := memMap[cur.Uid]
			if !ok {
				continue
			}
			if mem == nil {
				continue
			}

			before := cur.Total
			if mem.Partner == 1 {
				if mem.Superior == 0 {
					cur.Total = total
				} else if mem.Superior > 0 {
					sup, ok := resultSet[mem.Superior]
					if !ok {
						continue
					}
					if sup != nil {
						if sup.Total < 0 {
							cur.Total = sup.Total
						} else if sup.ConfiguredRoyaltyPercent() {
							cur.Total = sup.Total * sup.RealRoyaltyPercent() / 100
						} else {
							cur.Total = -1
						}
					}
				} else {
					cur.Total = total
				}
			}

			resultSet[cur.Uid] = cur
			if cur.Total != before {
				if err = tx.Model(dbModels.HousePartnerPyramid{}).
					Where("dhid = ? and dfid = ? and uid = ?", hid, fid, cur.Uid).
					Update("total", cur.Total).Error; err != nil {
					return
				}
			}
		}
	}
	return
}
