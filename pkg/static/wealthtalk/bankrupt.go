package wealthtalk

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/dao"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

// 用户破产
func GetUserAllowance(db *gorm.DB, cli *dao.DB_r, uid int64, nowStr string, conf *models.ConfigServer) (interface{}, error) {
	// 如果今天还没有购买过低保礼包则提示购买
	buy, err := models.IsAllowanceGiftPurchasedToday(db, uid, nowStr)
	if err != nil {
		return nil, err
	}

	if !buy {
		ret := &static.AllowanceGift{}

		ret.PurchasedCount, err = models.GetAllowanceGiftPurchasedCount(db)

		if err != nil {
			return nil, err
		}

		return ret, nil
	} else {
		return GetUserAllowanceWithoutGift(db, cli, uid, nowStr, conf)
	}
}

// 用户在没有礼包的情况下破产
func GetUserAllowanceWithoutGift(db *gorm.DB, cli *dao.DB_r, uid int64, nowStr string, conf *models.ConfigServer) (interface{}, error) {
	// 如果今天还有低保 则发放低保
	var count int
	if err := db.Table(models.UserAllowances{}.TableName()).Where("uid = ? and date = ?", uid, nowStr).Count(&count).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	} else {
		// 低保次数 还未达到上限
		if count < conf.Allowances.Num {
			var aftGold int
			aftGold, err = disposeUserAllowances(db, cli, uid, conf.Allowances.AwardGold)
			if err != nil {
				return nil, err
			} else {
				ret := new(static.Msg_S2C_Allowances)
				ret.Current = count + 1
				ret.Gold = conf.Allowances.AwardGold
				ret.Remain = conf.Allowances.Num - count - 1
				ret.AfterGold = aftGold
				return ret, nil
			}
		} else {
			// 低保次数达到上限
			return nil, nil
		}
	}
}

// 发放低保
func disposeUserAllowances(db *gorm.DB, cli *dao.DB_r, uid int64, allowancesGold int) (int, error) {
	tx := db.Begin()
	_, aftGold, err := UpdateGold(uid, allowancesGold, models.CostTypeAllowances, tx)
	if err != nil {
		tx.Rollback()
		xlog.Logger().Error("dispose user allowance: ", err)
		return 0, err
	}

	// 添加低保记录
	record := new(models.UserAllowances)
	record.Uid = uid
	record.AwardGold = allowancesGold
	record.Date = time.Now()
	if err = tx.Create(&record).Error; err != nil {
		tx.Rollback()
		xlog.Logger().Error("dispose user allowance: ", err)
		return 0, err
	}
	// 更新redis
	if err = cli.UpdatePersonAttrs(uid, "Gold", aftGold); err != nil {
		tx.Rollback()
		xlog.Logger().Error("dispose user allowance: ", err)
		return 0, err
	}
	tx.Commit()
	return aftGold, nil
}
