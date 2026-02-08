package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"time"
)

type houseMemberExitState int

const (
	HouseMemberExitInvalid  houseMemberExitState = 0
	HouseMemberExitApplying houseMemberExitState = 1
	HouseMemberExitAgreed   houseMemberExitState = 2
	HouseMemberExitRefused  houseMemberExitState = 3
)

func NewHouseExit(hid, uid int64) *HouseExit {
	return &HouseExit{
		HId: hid,
		Uid: uid,
	}
}

type HouseExit struct {
	Id        int64                `gorm:"primary_key;column:id" json:"id"` // id
	HId       int64                `gorm:"column:hid;comment:'包厢hid'"`
	Uid       int64                `gorm:"column:uid;comment:'玩家uid'"`
	Opt       int64                `gorm:"column:opt;comment:'操作人uid'"`
	State     houseMemberExitState `gorm:"column:state;comment:'状态'"`
	CreatedAt time.Time            `gorm:"column:created_at;type:datetime" json:"-"` // 创建时间
	UpdatedAt time.Time            `gorm:"column:updated_at;type:datetime" json:"-"` // 更新时间
}

func (HouseExit) TableName() string {
	return "house_exit"
}

func (he *HouseExit) beforeAlter(tx *gorm.DB) error {
	// lock row
	err := tx.Set("gorm:query_option", "FOR UPDATE").First(he, "hid = ? and uid = ?", he.HId, he.Uid).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = tx.Create(he).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (he *HouseExit) alterState(tx *gorm.DB, opt int64, state houseMemberExitState) error {
	update := make(map[string]interface{}, 2)
	update["state"] = state
	update["opt"] = opt
	return tx.Model(he).Update(update).Error
}

func (he *HouseExit) Apply(tx *gorm.DB) error {
	err := he.beforeAlter(tx)
	if err != nil {
		return err
	}
	return he.alterState(tx, he.Uid, HouseMemberExitApplying)
}

func (he *HouseExit) Agree(tx *gorm.DB, opt int64) error {
	err := he.beforeAlter(tx)
	if err != nil {
		return err
	}
	if he.State != HouseMemberExitApplying {
		return xerrors.HouseMemExitApplyError
	}
	return he.alterState(tx, opt, HouseMemberExitAgreed)
}

func (he *HouseExit) Refuse(tx *gorm.DB, opt int64) error {
	err := he.beforeAlter(tx)
	if err != nil {
		return err
	}
	if he.State != HouseMemberExitApplying {
		return xerrors.HouseMemExitApplyError
	}
	return he.alterState(tx, opt, HouseMemberExitRefused)
}

func initHouseExit(db *gorm.DB) error {
	var err error
	model := &HouseExit{}
	if db.HasTable(model) {
		err = db.AutoMigrate(model).Error
	} else {
		err = db.CreateTable(model).Error
		if err == nil {
			sql := fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE INDEX `idx_hid_uid` (`hid`, `uid`) USING BTREE", model.TableName())
			err = db.Exec(sql).Error
		}
	}
	return err
}
