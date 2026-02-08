package models

import (
	"github.com/jinzhu/gorm"
	"strings"
)

const defaultHouseFloorColor = "0,1,2,3,4,0,1,2,3,4,0,1,2,3,4,0,1,2,3,4"

type HouseFloorColor struct {
	Id         int64  `gorm:"primary_key;column:id" json:"id"` // id
	FloorColor string `gorm:"column:floor_color;comment:'每层楼的颜色代码'"`
}

func (HouseFloorColor) TableName() string {
	return "house_floor_color"
}

func (hfc *HouseFloorColor) Color() []string {
	refStr := strings.Split(defaultHouseFloorColor, ",")
	str := strings.Split(hfc.FloorColor, ",")
	l1 := len(refStr)
	for i, l2 := 0, len(str); i < l2; i++ {
		if i < l1 {
			if s := str[i]; s != "" {
				refStr[i] = s
			}
		}
	}
	return refStr
}

func initHouseFloorColor(db *gorm.DB) error {
	var err error
	model := &HouseFloorColor{}
	if db.HasTable(model) {
		err = db.AutoMigrate(model).Error
	} else {
		err = db.CreateTable(model).Error
	}
	return err
}
