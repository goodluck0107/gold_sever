// 用户财富管理
package wealthtalk

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

// 更新房卡数据(正数增加, 负数减少)
func UpdateCard(uid int64, kacost int, frozenkacost int, costType int8, tx *gorm.DB) (befka int, aftka int, beffka int, aftfka int, err error) {
	//if kacost == 0 && frozenkacost == 0 {
	//	return 0, 0, 0, 0, errors.New("kacost & frozenkacost is zero")
	//}

	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Errorln(err)
		return 0, 0, 0, 0, err
	}
	// 操作前的房卡
	befka = user.Card
	beffka = user.FrozenCard

	// 更新房卡
	user.Card = user.Card + kacost
	if user.Card < 0 {
		// 余卡不足, 减少至0
		xlog.Logger().Errorln(fmt.Sprintf("用户[%d]余卡不足, 当前房卡[%d], 扣卡[%d]", user.Id, user.Card-kacost, -1*kacost))
		user.Card = 0
	}
	// 更新冻结房卡
	user.FrozenCard = user.FrozenCard + frozenkacost
	if user.FrozenCard < 0 {
		// 冻结房卡有误
		xlog.Logger().Errorln(fmt.Sprintf("用户[%d]冻结房卡有误, 当前已冻结房卡[%d], 加卡[%d]", user.Id, user.FrozenCard-frozenkacost, frozenkacost))
		user.FrozenCard = 0
	}
	// 更新数据库信息
	if err = tx.Model(&user).Update(map[string]interface{}{
		"card":       user.Card,
		"frozencard": user.FrozenCard}).Error; err != nil {
		xlog.Logger().Errorln("update account failed: ", err.Error())
		return 0, 0, 0, 0, err
	}
	// 操作后的房卡
	aftka = user.Card
	aftfka = user.FrozenCard

	// 记录用户财富消耗流水
	if kacost != 0 {
		record := new(models.UserWealthCost)
		record.Uid = uid
		record.WealthType = consts.WealthTypeCard
		record.CostType = costType
		record.Cost = int64(kacost)
		record.BeforeNum = int64(befka)
		record.AfterNum = int64(aftka)
		if err = tx.Create(&record).Error; err != nil {
			xlog.Logger().Errorln(err.Error())
			return 0, 0, 0, 0, err
		}
	}

	return befka, aftka, beffka, aftfka, nil
}

/*
	更新金币数据(正数增加, 负数减少)
	befgold: 操作前的金币, aftgold: 操作后的金币
*/
func UpdateGold(uid int64, cost int, costType int8, tx *gorm.DB) (befgold int, aftgold int, err error) {
	if cost == 0 {
		return 0, 0, errors.New("cost zero")
	}
	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Error(err)
		return 0, 0, err
	}
	// 操作前的金币
	befgold = user.Gold

	// 更新金币
	user.Gold = user.Gold + cost
	if user.Gold < 0 {
		// 金币不足, 减少至0
		xlog.Logger().Error(fmt.Sprintf("用户[%d]金币不足, 当前金币[%d], 扣金币[%d]", user.Id, user.Gold-cost, -1*cost))
		user.Gold = 0
	}
	// 更新数据库信息
	if err = tx.Model(&user).Update("gold", user.Gold).Error; err != nil {
		xlog.Logger().Error("update account failed: ", err.Error())
		return 0, 0, err
	}
	// 操作后的房卡
	aftgold = user.Gold

	// 记录用户财富消耗流水
	record := new(models.UserWealthCost)
	record.Uid = uid
	record.WealthType = consts.WealthTypeGold
	record.CostType = costType
	record.Cost = int64(cost)
	record.BeforeNum = int64(befgold)
	record.AfterNum = int64(aftgold)
	if err = tx.Create(&record).Error; err != nil {
		xlog.Logger().Error(err.Error())
		return 0, 0, err
	}
	return befgold, aftgold, nil
}

/*
	更新用户保险箱金币数据
	savegold: 操作后身上的金币, aftinsuregold: 操作后保险箱的金币, goldchange: 身上金币的变化量
*/
func UpdateInsureGold(uid int64, savegold int, tx *gorm.DB) (befgold int, aftgold int, aftinsuregold int, cuserror *xerrors.XError) {
	var err error
	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Error(err)
		return 0, 0, 0, xerrors.DBExecError
	}
	// 操作前的总金币
	total := user.Gold + user.InsureGold
	befgold = user.Gold
	user.Gold = total - savegold
	user.InsureGold = savegold

	// 更新金币
	updateMap := make(map[string]interface{})
	updateMap["gold"] = user.Gold
	updateMap["insure_gold"] = user.InsureGold

	// 更新数据库信息
	if err = tx.Model(&user).Updates(updateMap).Error; err != nil {
		xlog.Logger().Error("update account failed: ", err.Error())
		return 0, 0, 0, xerrors.DBExecError
	}
	aftgold = user.Gold

	// 记录用户财富消耗流水
	record := new(models.UserWealthCost)
	record.Uid = uid
	record.WealthType = consts.WealthTypeGold
	record.CostType = models.CostTypeInsureGold
	record.Cost = int64(aftgold - befgold)
	record.BeforeNum = int64(befgold)
	record.AfterNum = int64(aftgold)
	if err = tx.Create(&record).Error; err != nil {
		xlog.Logger().Error(err.Error())
		return 0, 0, 0, xerrors.DBExecError
	}

	// 记录保险箱操作记录
	insureRecord := new(models.InsureGoldRecord)
	insureRecord.Uid = uid
	if record.Cost < 0 {
		// 存
		insureRecord.Type = models.InsureGoldSave
		insureRecord.Num = int(-1 * record.Cost)
		insureRecord.BeforeInsureGold = user.InsureGold - insureRecord.Num
	} else {
		// 取
		insureRecord.Type = models.InsureGoldFetch
		insureRecord.Num = int(record.Cost)
		insureRecord.BeforeInsureGold = user.InsureGold + insureRecord.Num
	}
	insureRecord.BeforeGold = befgold
	insureRecord.AfterGold = aftgold
	insureRecord.AfterInsureGold = user.InsureGold
	if err = tx.Create(&insureRecord).Error; err != nil {
		xlog.Logger().Error(err.Error())
		return 0, 0, 0, xerrors.DBExecError
	}

	return befgold, user.Gold, user.InsureGold, nil
}

/*
	更新礼券数据(正数增加, 负数减少)
	befnum: 操作前的礼券, aftnum: 操作后的礼券
*/
func UpdateCoupon(uid int64, cost int, costType int8, tx *gorm.DB) (befnum int, aftnum int, err error) {
	if cost == 0 {
		return 0, 0, errors.New("cost zero")
	}
	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Error(err)
		return 0, 0, err
	}
	// 操作前的礼券
	befnum = user.GoldBean

	// 更新礼券
	user.GoldBean = user.GoldBean + cost
	if user.GoldBean < 0 {
		// 礼券不足, 报错
		xlog.Logger().Error(fmt.Sprintf("用户[%d]礼券不足, 当前礼券[%d], 扣礼券[%d]", user.Id, user.GoldBean-cost, -1*cost))
		user.GoldBean = 0
	}
	// 更新数据库信息
	if err = tx.Model(&user).Update("gold_bean", user.GoldBean).Error; err != nil {
		xlog.Logger().Error("update account failed: ", err.Error())
		return 0, 0, err
	}
	// 操作后的礼券
	aftnum = user.GoldBean

	// 记录用户财富消耗流水
	record := new(models.UserWealthCost)
	record.Uid = uid
	record.WealthType = consts.WealthTypeCoupon
	record.CostType = costType
	record.Cost = int64(cost)
	record.BeforeNum = int64(befnum)
	record.AfterNum = int64(aftnum)
	if err = tx.Create(&record).Error; err != nil {
		xlog.Logger().Error(err.Error())
		return 0, 0, err
	}

	return befnum, aftnum, nil
}

// 更新加盟商房卡数据(正数增加, 负数减少)
func UpdateLeagueCard(leagueID, uid int64, kacost int, frozenkacost int, costType int8, tx *gorm.DB) error {
	//if kacost == 0 && frozenkacost == 0 {
	//	return errors.New("kacost & frozenkacost is zero")
	//}
	var league models.League
	league.LeagueID = leagueID
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("league_id = ?", league.LeagueID).First(&league).Error; err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	// 更新房卡
	league.Card = league.Card + int64(kacost)
	if league.Card < 0 {
		// 余卡不足, 减少至0
		xlog.Logger().Errorln(fmt.Sprintf("加盟商[%d]余卡不足, 当前房卡[%d], 扣卡[%d]", league.LeagueID, league.Card-int64(kacost), -1*kacost))
		league.Card = 0
	}
	// 更新冻结房卡
	league.FreezeCard = league.FreezeCard + int64(frozenkacost)
	if league.FreezeCard < 0 {
		// 冻结房卡有误
		xlog.Logger().Errorln(fmt.Sprintf("加盟商[%d]冻结房卡有误, 当前已冻结房卡[%d], 加卡[%d]", league.LeagueID, league.FreezeCard-int64(frozenkacost), frozenkacost))
		league.FreezeCard = 0
	}
	if err := tx.Model(&league).Update(map[string]interface{}{
		"card":        league.Card,
		"freeze_card": league.FreezeCard}).Error; err != nil {
		xlog.Logger().Errorln("update account failed: ", err.Error())
		return err
	}
	// 记录用户财富消耗流水
	if kacost != 0 {
		record := new(models.LeagueCardRecord)
		record.Uid = uid
		record.LeagueID = leagueID
		record.Cost = int64(kacost)
		if err := tx.Create(&record).Error; err != nil {
			xlog.Logger().Errorln(err.Error())
			return err
		}
	}

	return nil
}

/*
	更新金币数据(正数增加, 负数减少)
	befgold: 操作前的金币, aftgold: 操作后的金币
*/
func UpdateDiamond(uid int64, cost int, costType int8, tx *gorm.DB) (befDiamond int, aftDiamond int, err error) {
	if cost == 0 {
		return 0, 0, errors.New("cost zero")
	}
	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Error(err)
		return 0, 0, err
	}
	// 操作前的金币
	befDiamond = user.Diamond

	// 更新金币
	user.Diamond = user.Diamond + cost
	if user.Diamond < 0 {
		// 金币不足, 减少至0
		xlog.Logger().Error(fmt.Sprintf("用户[%d]钻石不足, 当前钻石[%d], 扣钻石[%d]", user.Id, befDiamond, -1*cost))
		user.Gold = 0
	}
	// 更新数据库信息
	if err = tx.Model(&user).Update("diamond", user.Diamond).Error; err != nil {
		xlog.Logger().Error("update diamond failed: ", err.Error())
		return 0, 0, err
	}
	// 操作后的房卡
	aftDiamond = user.Diamond

	// 记录用户财富消耗流水
	record := new(models.UserWealthCost)
	record.Uid = uid
	record.WealthType = consts.WealthTypeDiamond
	record.CostType = costType
	record.Cost = int64(cost)
	record.BeforeNum = int64(befDiamond)
	record.AfterNum = int64(aftDiamond)
	if err = tx.Create(&record).Error; err != nil {
		xlog.Logger().Error(err.Error())
		return 0, 0, err
	}
	return befDiamond, aftDiamond, nil
}

type WealthItem struct {
	Wt int8
	Wn int
}

type WealthResultItem struct {
	WealthItem
	BeforeNum int
	AfterNum  int
	TimeAt    int64
}

// 更新财富 chess只有三种货币 房卡 金币 钻石, 后面可以增加
func UpdateWealth(uid int64, wealthItems []WealthItem, costType int8, tx *gorm.DB) (ret []WealthResultItem, err error) {
	if len(wealthItems) == 0 {
		return nil, errors.New("cost zero")
	}
	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}

	wealthMap := make(map[int8]int)
	l := len(wealthItems)
	// 记录
	records := make([]WealthResultItem, 0)
	var notAdd bool
	for i := 0; i < l; i++ {
		wt, wn := wealthItems[i].Wt, wealthItems[i].Wn
		var wr WealthResultItem
		wr.WealthItem = wealthItems[i]
		switch wt {
		case consts.WealthTypeCard:
			wr.BeforeNum = user.Card
			user.Card += wn
			if user.Card < 0 {
				// 金币不足, 减少至0
				xlog.Logger().Error(fmt.Sprintf("用户[%d]%s不足, 当前[%d], 扣金币[%d]", user.Id, consts.WealthTypeString(wr.Wt), wr.BeforeNum, wn))
				user.Card = 0
			}
			wr.AfterNum = user.Card
		case consts.WealthTypeGold:
			wr.BeforeNum = user.Gold
			user.Gold += wn
			if user.Gold < 0 {
				// 金币不足, 减少至0
				xlog.Logger().Error(fmt.Sprintf("用户[%d]%s不足, 当前[%d], 扣金币[%d]", user.Id, consts.WealthTypeString(wr.Wt), wr.BeforeNum, wn))
				user.Gold = 0
			}
			wr.AfterNum = user.Gold
		case consts.WealthTypeDiamond:
			wr.BeforeNum = user.Diamond
			user.Diamond += wn
			if user.Diamond < 0 {
				// 金币不足, 减少至0
				xlog.Logger().Error(fmt.Sprintf("用户[%d]%s不足, 当前[%d], 扣金币[%d]", user.Id, consts.WealthTypeString(wr.Wt), wr.BeforeNum, wn))
				user.Diamond = 0
			}
			wr.AfterNum = user.Diamond
		case consts.WealthTypeCardRcd:
			wr.BeforeNum = 0
			wr.AfterNum = wn
		default:
			notAdd = true
			xlog.Logger().Errorf("uid%d,不支持的财富类型:%d", uid, wt)
		}
		if !notAdd {
			records = append(records, wr)
			wealthMap[wt] += wn
		}
	}

	// 制作结果
	for wt, wn := range wealthMap {
		var wr WealthResultItem
		wr.WealthItem.Wt = wt
		wr.WealthItem.Wn = wn
		switch wt {
		case consts.WealthTypeCard:
			wr.AfterNum = user.Card
		case consts.WealthTypeGold:
			wr.AfterNum = user.Gold
		case consts.WealthTypeDiamond:
			wr.AfterNum = user.Diamond
		case consts.WealthTypeCardRcd:
			wr.AfterNum = wn
			//写数据库并返回截止时间
			_, wr.TimeAt = SaveUserToolInfo(uid, 0, wr.AfterNum, tx)
		}
		ret = append(ret, wr)
	}

	// 更新字段
	updates := make(map[string]interface{})
	for i := 0; i < len(ret); i++ {
		switch ret[i].Wt {
		case consts.WealthTypeCard:
			updates["card"] = ret[i].AfterNum
		case consts.WealthTypeGold:
			updates["gold"] = ret[i].AfterNum
		case consts.WealthTypeDiamond:
			updates["diamond"] = ret[i].AfterNum
		}
	}

	// 更新数据库信息
	if len(updates) > 0 {
		if err = tx.Model(&user).Update(updates).Error; err != nil {
			xlog.Logger().Error("update account failed: ", err.Error())
			return nil, err
		}
	}

	for i, l := 0, len(records); i < l; i++ {
		// 记录用户财富消耗流水
		record := new(models.UserWealthCost)
		record.Uid = uid
		record.WealthType = records[i].Wt
		record.CostType = costType
		record.Cost = int64(records[i].Wn)
		record.BeforeNum = int64(records[i].BeforeNum)
		record.AfterNum = int64(records[i].AfterNum)
		if err = tx.Create(&record).Error; err != nil {
			xlog.Logger().Error(err.Error())
			return nil, err
		}
	}
	return ret, nil
}

// 保存道具信息
func SaveUserToolInfo(uid int64, toolType int16, num int, tx *gorm.DB) (retB bool, retTimeAt int64) {
	var records []models.UserTools
	err := tx.Model(models.UserTools{}).Where("uid = ? and tool_type = ?", uid, toolType).Find(&records).Error
	if err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("插入UserTools出错"))
	} else {
		// 写数据库记录
		if len(records) == 0 {
			if num > 0 {
				newRecord := new(models.UserTools)
				timeNow := time.Now()
				deadAt := timeNow.Unix() + int64(num)*24*3600
				timeStrAt := time.Unix(deadAt, 0).Format("2006-01-02 15:04:05")
				timeAt, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStrAt, time.Local)
				newRecord.Uid = uid
				newRecord.ToolType = toolType
				newRecord.CreatedAt = timeNow
				newRecord.DeadAt = timeAt
				if errCreate := tx.Create(&newRecord).Error; errCreate != nil {
					xlog.Logger().Errorln(fmt.Sprintf("插入UserTools出错：（%v）, uid = %s, tool_type = %d", errCreate, uid, toolType))
					retB = false
				} else {
					retB = true
				}
				retTimeAt = deadAt
			}
		} else {
			deadAt := int64(0)
			for _, record := range records {
				timeNow := time.Now()
				if num > 0 {
					if record.DeadAt.Unix() >= timeNow.Unix() {
						deadAt = record.DeadAt.Unix() + int64(num)*24*3600
					} else {
						deadAt = timeNow.Unix() + int64(num)*24*3600
					}
				} else {
					deadAt = record.DeadAt.Unix()
				}
				timeStrAt := time.Unix(deadAt, 0).Format("2006-01-02 15:04:05")
				timeAt, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStrAt, time.Local)
				//写数据库
				updateMap := make(map[string]interface{})
				updateMap["created_at"] = timeNow //更新时间
				updateMap["dead_at"] = timeAt     //截止时间
				errUp := tx.Model(models.UserTools{}).Where("uid = ? and tool_type = ?", uid, toolType).Update(updateMap).Error
				if errUp != nil {
					xlog.Logger().Errorln(fmt.Sprintf("更新UserTools失败, uid = %s, tool_type = %d", uid, toolType))
					retB = false
					break
				}
				retB = true
				break
			}
			retTimeAt = deadAt
		}
	}
	return retB, retTimeAt
}
