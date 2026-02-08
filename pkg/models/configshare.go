package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

// 分享配置表
type ConfigShare struct {
	Id          int    `gorm:"primary_key;column:id"`          // 表id（分享id）
	SceneId     int    `gorm:"column:scene_id"`                // 分享场景id(0 大厅主界面分享 1 大厅任务界面分享 2 小结算界面分享 3 礼券界面分享)
	Platform    int    `gorm:"column:platform;default:0"`      // 平台标识 0 不区分平台 1 app端 2 小程序
	ShareTo     int    `gorm:"column:share_to"`                // 分享到哪儿 0 朋友圈 1 微信好友 2 朋友圈和微信好友
	ShareType   int    `gorm:"column:share_type"`              // 分享类型 0 文字 1 图片 2 图文 3 链接
	ShareTimes  int    `gorm:"column:share_times;default:0"`   // 是否限制分享次数 0 不限制次数
	Reward      string `gorm:"column:reward;default:''"`       // 分享奖励（支持多个奖励）[{"wealth_type":2,"num":100},{"wealth_type":1,"num":100}] 默认没有奖励
	KindId      int    `gorm:"column:kind_id;default:0"`       // 指定的游戏分享 0 通用数据
	SiteType    int    `gorm:"column:site_type;default:0"`     // 场次标识（config_site） 0 通用数据
	Title       string `gorm:"column:title"`                   // 分享的标题
	Content     string `gorm:"column:content"`                 // 分享的内容
	ImgDownload string `gorm:"column:img_download;default:''"` // 图片下载地址 默认不需要下载图片
	Link1       string `gorm:"column:link1"`                   // 跳转域名地址1
	Link2       string `gorm:"column:link2"`                   // 跳转域名地址2
	Link3       string `gorm:"column:link3"`                   // 跳转域名地址3
}

func (ConfigShare) TableName() string {
	return "config_share"
}

func initConfigShare(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigShare{}) {
		err = db.AutoMigrate(&ConfigShare{}).Error
	} else {
		err = db.CreateTable(&ConfigShare{}).Error
	}
	return err
}

// 获取分享配置
func GetConfigShare(db *gorm.DB, sceneId int, platform int, kindId int, siteType int) (*ConfigShare, error) {
	var config ConfigShare
	err := db.Model(ConfigShare{}).Where("scene_id = ? and platform = ? and kind_id = ? and site_type = ?", sceneId, platform, kindId, siteType).First(&config).Error
	// 如果未查询到记录 则在通用数据里找一遍
	if err == gorm.ErrRecordNotFound {
		err = db.Model(ConfigShare{}).Where("scene_id = ? and platform = ? and kind_id = ? and site_type = ?", sceneId, platform, kindId, 0).First(&config).Error
	}
	if err == gorm.ErrRecordNotFound {
		err = db.Model(ConfigShare{}).Where("scene_id = ? and platform = ? and kind_id = ? and site_type = ?", sceneId, platform, 0, 0).First(&config).Error
	}
	if err == gorm.ErrRecordNotFound {
		err = db.Model(ConfigShare{}).Where("scene_id = ? and platform = ? and kind_id = ? and site_type = ?", sceneId, 0, 0, 0).First(&config).Error
	}
	return &config, err
}

// 获取分享配置
func GetConfigShareById(db *gorm.DB, shareId int) (*ConfigShare, error) {
	var config ConfigShare
	err := db.Model(ConfigShare{}).Where("id = ?", shareId).First(&config).Error
	return &config, err
}

// 分享记录表
type ShareHistory struct {
	Id       int64     `gorm:"primary_key;column:id"`          // 表id
	Uid      int64     `gorm:"column:uid"`                     // 分享玩家uid
	ShareId  int       `gorm:"column:share_id"`                // 分享id
	CreateAt time.Time `gorm:"column:create_at;type:datetime"` // 创建时间
}

func (ShareHistory) TableName() string {
	return "share_history"
}

func initShareHistory(db *gorm.DB) error {
	var err error
	if db.HasTable(&ShareHistory{}) {
		err = db.AutoMigrate(&ShareHistory{}).Error
	} else {
		err = db.CreateTable(&ShareHistory{}).Error
		db.Model(ShareHistory{}).AddIndex("idx_uid_shareId_createAt", "uid", "share_id", "create_at")
	}
	return err
}

// 获取分享次数
func GetShareHistoryCnt(db *gorm.DB, uid int64, shareId int, time time.Time) (int, error) {
	var count = 0
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(ShareHistory{}).Where(`uid = ? and share_id = ? and date_format(create_at, "%Y-%m-%d") = ?`, uid, shareId, selectStr).Count(&count).Error
	return count, err
}
