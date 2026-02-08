package models

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

type GameWinRate struct {
	Id         int64 `gorm:"primary_key;column:id"`        //! id
	KindId     int   `gorm:"column:kindid"`                //! 玩法id
	UId        int64 `gorm:"column:uid"`                   //! 在线人数
	WinCount   int   `gorm:"column:win_count;default:0"`   // 胜利局数
	LostCount  int   `gorm:"column:lost_count;default:0"`  // 失败局数
	DrawCount  int   `gorm:"column:draw_count;default:0"`  // 和局局数
	FleeCount  int   `gorm:"column:flee_count;default:0"`  // 逃跑局数
	TotalCount int   `gorm:"column:total_count;default:0"` // 总局数
}

func (GameWinRate) TableName() string {
	return "game_win_rate"
}

func initGameWinRate(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameWinRate{}) {
		err = db.AutoMigrate(&GameWinRate{}).Error
	} else {
		err = db.CreateTable(&GameWinRate{}).Error
		db.Model(GameWinRate{}).AddUniqueIndex("uidx_uid_kindid", "uid", "kindid")
	}
	return err
}

func GameWinRateUpData(db *gorm.DB, uid int64, kindid int, winCountAdd, lostCountAdd, drawCountAdd, fleeCountAdd, totalCountAdd int) error {
	var rate GameWinRate
	err := db.Model(GameWinRate{}).Where("uid = ? and kindid = ?", uid, kindid).First(&rate).Error
	if err != nil {
		xlog.Logger().Errorf("GameWinRateGet Err:%v", err)
		if err == gorm.ErrRecordNotFound {
			err = db.Save(&GameWinRate{UId: uid,
				KindId:     kindid,
				WinCount:   winCountAdd,
				LostCount:  lostCountAdd,
				DrawCount:  drawCountAdd,
				FleeCount:  fleeCountAdd,
				TotalCount: totalCountAdd}).Error
		}
	} else {
		updateMap := make(map[string]interface{})
		updateMap["win_count"] = gorm.Expr("win_count + ?", winCountAdd)
		updateMap["lost_count"] = gorm.Expr("lost_count + ?", lostCountAdd)
		updateMap["draw_count"] = gorm.Expr("draw_count + ?", drawCountAdd)
		updateMap["flee_count"] = gorm.Expr("flee_count + ?", fleeCountAdd)
		updateMap["total_count"] = gorm.Expr("total_count + ?", totalCountAdd)
		err = db.Model(GameWinRate{Id: rate.Id}).Update(updateMap).Error
	}
	return err
}

func GameWinRateUpDataWithType(db *gorm.DB, uid int64, kindid int, TypeAdd, CountAdd int) error {
	var rate GameWinRate
	err := db.Model(GameWinRate{}).Where("uid = ? and kindid = ?", uid, kindid).First(&rate).Error
	if err != nil {
		xlog.Logger().Errorf("GameWinRateGet Err:%v", err)
		if err == gorm.ErrRecordNotFound {
			rate := &GameWinRate{UId: uid,
				KindId: kindid}
			switch TypeAdd {
			case static.ScoreKind_Win:
				rate.WinCount = CountAdd
			case static.ScoreKind_Lost:
				rate.LostCount = CountAdd
			case static.ScoreKind_Draw:
				rate.DrawCount = CountAdd
			case static.ScoreKind_Flee:
				rate.FleeCount = CountAdd
			}
			rate.TotalCount = CountAdd
			err = db.Create(rate).Error
		}
	} else {
		updateMap := make(map[string]interface{})
		switch TypeAdd {
		case static.ScoreKind_Win:
			updateMap["win_count"] = gorm.Expr("win_count + ?", CountAdd)
		case static.ScoreKind_Lost:
			updateMap["lost_count"] = gorm.Expr("lost_count + ?", CountAdd)
		case static.ScoreKind_Draw:
			updateMap["draw_count"] = gorm.Expr("draw_count + ?", CountAdd)
		case static.ScoreKind_Flee:
			updateMap["flee_count"] = gorm.Expr("flee_count + ?", CountAdd)
		}
		updateMap["total_count"] = gorm.Expr("total_count + ?", CountAdd)
		err = db.Model(GameWinRate{Id: rate.Id}).Update(updateMap).Error
	}
	return err
}

func GameWinRateGet(db *gorm.DB, uid int64, kindid int) (GameWinRate, error) {
	var rate GameWinRate
	err := db.Model(GameWinRate{}).Where("uid = ? and kindid = ?", uid, kindid).First(&rate).Error
	if err != nil {
		xlog.Logger().Errorf("GameWinRateGet Err:%v", err)
	}
	return rate, err
}
