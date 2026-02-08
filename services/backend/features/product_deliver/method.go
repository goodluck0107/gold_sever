package product_deliver

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/jinzhu/gorm"
)

// 更新房卡数据(正数增加, 负数减少)
func updcard(uid int64, kacost int, frozenkacost int, tx *gorm.DB, costType int8) (befka int, aftka int, beffka int, aftfka int, err error) {
	if kacost == 0 && frozenkacost == 0 {
		return 0, 0, 0, 0, nil
	}

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

	if kacost != 0 {
		// 记录用户房卡消耗流水
		record := new(models.UserWealthCost)
		record.Uid = uid
		record.Cost = int64(kacost)
		record.AfterNum = int64(aftka)
		record.BeforeNum = int64(befka)
		record.WealthType = consts.WealthTypeCard
		record.CostType = costType
		if err = tx.Create(&record).Error; err != nil {
			xlog.Logger().Errorln(err)
			return 0, 0, 0, 0, err
		}
	}

	return befka, aftka, beffka, aftfka, nil
}
