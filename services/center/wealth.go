package center

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

func updcard(uid int64, kacost int, frozenkacost int, costType int8, tx *gorm.DB) (befka int, aftka int, beffka int, aftfka int, err error) {
	if kacost == 0 && frozenkacost == 0 {
		return 0, 0, 0, 0, errors.New("kacost & frozenkacost is zero")
	}

	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Error(err)
		return 0, 0, 0, 0, err
	}
	// 操作前的房卡
	befka = user.Card
	beffka = user.FrozenCard

	// 更新房卡
	user.Card = user.Card + kacost
	if user.Card < 0 {
		// 余卡不足, 减少至0
		xlog.Logger().Error(fmt.Sprintf("用户[%d]余卡不足, 当前房卡[%d], 扣卡[%d]", user.Id, user.Card-kacost, -1*kacost))
		user.Card = 0
	}
	// 更新冻结房卡
	user.FrozenCard = user.FrozenCard + frozenkacost
	if user.FrozenCard < 0 {
		// 冻结房卡有误
		xlog.Logger().Error(fmt.Sprintf("用户[%d]冻结房卡有误, 当前已冻结房卡[%d], 加卡[%d]", user.Id, user.FrozenCard-frozenkacost, frozenkacost))
		user.FrozenCard = 0
	}
	// 更新数据库信息
	if err = tx.Model(&user).Update(map[string]interface{}{
		"card":       user.Card,
		"frozencard": user.FrozenCard}).Error; err != nil {
		xlog.Logger().Error("update account failed: ", err.Error())
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
			xlog.Logger().Error(err.Error())
			return 0, 0, 0, 0, err
		}
	}

	return befka, aftka, beffka, aftfka, nil
}

func updCoupon(uid int64, cost int, costType int8, tx *gorm.DB) (befnum int, aftnum int, err error) {
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

func updgold(uid int64, cost int, costType int8, tx *gorm.DB) (befgold int, aftgold int, err error) {
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
		xlog.Logger().Error(fmt.Sprintf("用户[%d]欢乐豆不足, 当前欢乐豆[%d], 扣欢乐豆[%d]", user.Id, user.Gold-cost, -1*cost))
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

func updinsuregold(uid int64, savegold int, tx *gorm.DB) (befgold int, aftgold int, aftinsuregold int, cuserror *xerrors.XError) {
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

	// 如果是存钱的话, 需判断自身携带的金币是否低于标准
	if user.Gold < befgold && user.Gold < GetServer().ConServers.InsureendLine {
		cuserror = xerrors.NewXError(fmt.Sprintf("您携带的欢乐豆不足%d", GetServer().ConServers.InsureendLine))
		return 0, 0, 0, cuserror
	}

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
		insureRecord.Num = -1 * int(record.Cost)
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
