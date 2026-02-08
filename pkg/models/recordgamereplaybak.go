package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

//! 游戏单局回放记录
type RecordGameReplayBak struct {
	Id          int64             `gorm:"primary_key;column:id"`            //! id
	GameNum     string            `gorm:"column:gamenum"`                   //! 游戏ID,唯一标识
	RoomNum     int               `gorm:"column:roomnum"`                   //! 游戏房间ID
	PlayNum     int               `gorm:"column:playnum"`                   //! 游戏局数
	ServerId    int               `gorm:"column:serverid"`                  //! 游戏服务ID
	HandCard    string            `gorm:"column:handcard;type:text"`        //! 玩家手牌
	OutCard     string            `gorm:"column:outcard;type:text"`         //! 玩家出牌记录
	KindID      int               `gorm:"column:kindid"`                    //! 游戏kindID
	CardsNum    int               `gorm:"column:cardsnum"`                  //! 发完牌之后剩余牌数量
	UserVitamin string            `gorm:"column:user_vitamin;default:'{}'"` //! 玩家起始疲劳值
	UVitaminMap map[int64]float64 `gorm:"-"`                                //! 对应UserVitamin转化map
	CreatedAt   time.Time         `gorm:"column:created_at;type:datetime"`  //! 创建时间
}

func (RecordGameReplayBak) TableName() string {
	return "record_game_outdata_bak"
}

func initRecordGameReplayBak(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameReplayBak{}) {
		err = db.AutoMigrate(&RecordGameReplayBak{}).Error
	} else {
		err = db.CreateTable(&RecordGameReplayBak{}).Error
		if err == nil {
			// 修改递增初始值
			err = db.Exec("alter table record_game_outdata AUTO_INCREMENT=100000").Error
		}
	}
	return err
}
