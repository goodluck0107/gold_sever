package models

import "github.com/jinzhu/gorm"

//推荐游戏配置
type ConfigRecommendGame struct {
	Id       int64  `gorm:"primary_key;column:id"`
	KindId   int    `gorm:"column:kindid"`           // 游戏玩法id
	AreaCode int    `gorm:"column:areacode"`         // 游戏玩法id
	GameName string `gorm:"column:gamename;size:50"` // 游戏玩法名称
}

func (ConfigRecommendGame) TableName() string {
	return "config_recommend_game"
}

func initConfigRecommendGame(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigRecommendGame{}) {
		err = db.AutoMigrate(&ConfigRecommendGame{}).Error
	} else {
		err = db.CreateTable(&ConfigRecommendGame{}).Error
		db.Model(ConfigRecommendGame{}).AddIndex("idx_areacode", "areacode")
	}
	return err
}
