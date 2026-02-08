package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type HousePartnerRoyaltyHistory struct {
	Id            int64     `gorm:"primary_key;column:id"`           //! id
	DHid          int64     `gorm:"column:dhid"`                     //! 包厢id
	OptUser       int64     `gorm:"column:optuser"`                  //! 操作用户ID
	OptUserType   int       `gorm:"column:optusertype"`              //! 操作用户角色类型
	OptType       int       `gorm:"column:opttype"`                  //! 操作类型(增加,删除楼层,修改队长配置)
	OptFloorName  string    `gorm:"column:optfloorname"`             //! 操作楼层名称
	OptFloorId    int64     `gorm:"column:optfloorid"`               //! 操作楼层ID
	OptFloorIndex int       `gorm:"column:optfloorindex"`            //! 操作楼层索引
	BeOptUser     int64     `gorm:"column:beoptuser"`                //! 被操作用户ID
	Before        int       `gorm:"column:before"`                   //! 被操作用户操作前数据
	After         int       `gorm:"column:after"`                    //! 操作用户操作后数据
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (HousePartnerRoyaltyHistory) TableName() string {
	return "house_partner_royalty_history"
}

func InitHousePartnerRoyaltyHistory(db *gorm.DB) error {
	var err error
	if db.HasTable(&HousePartnerRoyaltyHistory{}) {
		err = db.AutoMigrate(&HousePartnerRoyaltyHistory{}).Error
	} else {
		err = db.CreateTable(&HousePartnerRoyaltyHistory{}).Error
		db.Model(&HousePartnerRoyaltyHistory{}).AddIndex("idx_dhid_beoptuser", "dhid", "beoptuser")
	}
	return err
}

type HousePartnerRoyaltyDetailItem struct {
	Beneficiary int64 `gorm:"column:beneficiary"` //! 收益者玩家id
	SelfProfit  int   `gorm:"column:selfprofit"`  //! 玩家自身收益
	SubProfit   int   `gorm:"column:subprofit"`   //! 下级提供收益
	ValidTimes  int   `gorm:"column:validtimes"`  //! 有效局
	DFid        int64 `gorm:"column:dfid"`        //! 包厢楼层id
}

type HousePartnerRoyaltyDetail struct {
	Id            int64     `gorm:"primary_key;column:id"`           //! id
	DHid          int64     `gorm:"column:dhid"`                     //! 包厢id
	DFid          int64     `gorm:"column:dfid"`                     //! 包厢楼层id
	DFloorIndex   int       `gorm:"column:dfloorindex"`              //! 包厢楼层索引
	GameNum       string    `gorm:"column:gamenum"`                  //! 游戏唯一编码
	ProviderRound int       `gorm:"column:providerround"`            //! 游戏提供收益的对局
	Beneficiary   int64     `gorm:"column:beneficiary"`              //! 收益者玩家id
	SelfProfit    int64     `gorm:"column:selfprofit"`               //! 玩家自身收益
	SubProfit     int64     `gorm:"column:subprofit"`                //! 下级提供收益
	PlayerUser    int64     `gorm:"column:playeruser"`               //! 玩游戏玩家id
	PlayerPartner int64     `gorm:"column:playerpartner"`            //! 游戏玩家合伙人
	Rent          int64     `gorm:"column:rent"`                     //! 房费
	PartnerLink   string    `gorm:"column:partnerlink"`              //! 合伙人关系
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (HousePartnerRoyaltyDetail) TableName() string {
	return "house_partner_royalty_detail"
}

func InitHousePartnerRoyaltyDetail(db *gorm.DB) error {
	var err error
	if db.HasTable(&HousePartnerRoyaltyDetail{}) {
		err = db.AutoMigrate(&HousePartnerRoyaltyDetail{}).Error
	} else {
		err = db.CreateTable(&HousePartnerRoyaltyDetail{}).Error
		db.Model(&HousePartnerRoyaltyDetail{}).AddIndex("idx_dhid_createdat_dfid_beneficiary", "dhid", "created_at", "dfid", "beneficiary")
		db.Model(&HousePartnerRoyaltyDetail{}).AddIndex("idx_dhid_createdat_dfloorindex_beneficiary", "dhid", "created_at", "dfloorindex", "beneficiary")
	}
	return err
}
