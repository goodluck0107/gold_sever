package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

/**
包厢合伙人金字塔模型收益配置表
*/

func HousePartnerPyramidDefault(hid, fid, uid int64, fIdx int) *HousePartnerPyramid {
	return &HousePartnerPyramid{
		Id:             0,
		Uid:            uid,
		DHid:           hid,
		DFid:           fid,
		FloorIndex:     fIdx,
		RoyaltyPercent: -1,
		// SuperiorPercent: -1,
		Total: -1,
	}
}

type HousePartnerPyramid struct {
	Id             int64     `gorm:"primary_key;column:id"`             //! id
	Uid            int64     `gorm:"column:uid"`                        //! id
	DHid           int64     `gorm:"column:dhid"`                       //! 包厢id
	DFid           int64     `gorm:"column:dfid"`                       //! 楼层id
	FloorIndex     int       `gorm:"column:floorindex"`                 //! 楼层索引
	RoyaltyPercent int       `gorm:"column:royalty_percent;default:-1"` //! 单局收益百分比/上级视角：他的收益
	Total          int64     `gorm:"column:total;default:-1"`           //! 总额
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime"`   //! 创建时间
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime"`   //! 更新时间
}

func (HousePartnerPyramid) TableName() string {
	return "house_partner_pyramid"
}

// 真实的单局收益百分比
func (hp *HousePartnerPyramid) RealRoyaltyPercent() int64 {
	if hp.RoyaltyPercent > 0 {
		return int64(hp.RoyaltyPercent)
	}
	return 0
}

// 真实的上级收益百分比
func (hp *HousePartnerPyramid) RealSuperiorPercent() int64 {
	return 100 - hp.RealRoyaltyPercent()
}

// 收益信息 返回单局收益 和 上级收益
func (hp *HousePartnerPyramid) EarningsInfo() (int64, int64) {
	if hp.ConfiguredRoyaltyPercent() && hp.Configurable() {
		royalty := hp.Total * hp.RealRoyaltyPercent() / 100
		return royalty, hp.Total - royalty
	}
	return 0, 0
}

// 按比例获取真实收益
func (hp *HousePartnerPyramid) RealEarningsInfo(baseRoyalty, realRoyalty int64) (int64, int64) {
	if hp.ConfiguredRoyaltyPercent() && hp.Configurable() {
		hp.Total = hp.Total * realRoyalty / baseRoyalty
		royalty := hp.Total * hp.RealRoyaltyPercent() / 100
		return royalty, hp.Total - royalty
	}
	return 0, 0
}

func (hp *HousePartnerPyramid) Configurable() bool {
	return hp.Total >= 0
}

func (hp *HousePartnerPyramid) ConfiguredRoyaltyPercent() bool {
	return hp.RoyaltyPercent >= 0
}

func initHousePartnerPyramid(db *gorm.DB) error {
	var err error
	if db.HasTable(&HousePartnerPyramid{}) {
		err = db.AutoMigrate(&HousePartnerPyramid{}).Error
	} else {
		err = db.CreateTable(&HousePartnerPyramid{}).Error
		sql := "ALTER TABLE `house_partner_pyramid` ADD UNIQUE INDEX `idx_dfid_uid` (`uid`, `dfid`) USING BTREE;"
		err = db.Exec(sql).Error
	}
	return err
}

type HousePartnerPyramidFloors []*HousePartnerPyramid

func (hpf *HousePartnerPyramidFloors) ToMapFloors() map[int64]HousePartnerPyramidFloors {
	res := make(map[int64]HousePartnerPyramidFloors)
	for i := 0; i < len(*hpf); i++ {
		mod := (*hpf)[i]
		mods, ok := res[mod.Uid]
		if !ok {
			mods = make(HousePartnerPyramidFloors, 0)
		}
		mods = append(mods, mod)
		res[mod.Uid] = mods
	}
	return res
}

func (hpf *HousePartnerPyramidFloors) ToMapFloor(fid int64) map[int64]*HousePartnerPyramid {
	res := make(map[int64]*HousePartnerPyramid)
	for i := 0; i < len(*hpf); i++ {
		mod := (*hpf)[i]
		if mod.DFid == fid {
			res[mod.Uid] = mod
		}
	}
	return res
}

func (hpf *HousePartnerPyramidFloors) GetPyramidByFid(fid int64) *HousePartnerPyramid {
	for i := 0; i < len(*hpf); i++ {
		mod := (*hpf)[i]
		if mod.DFid == fid {
			return mod
		}
	}
	return nil
}

func (hpf *HousePartnerPyramidFloors) GetPyramidByFIdx(fIdx int) *HousePartnerPyramid {
	for i := 0; i < len(*hpf); i++ {
		mod := (*hpf)[i]
		if mod.FloorIndex == fIdx {
			return mod
		}
	}
	return nil
}
