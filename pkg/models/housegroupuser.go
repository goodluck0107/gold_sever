package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

/*
drop table if EXISTS `house_group_user`;
create table `house_group_user` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
	`hid` int(11) unsigned not null comment '包厢id',
	`puid` int(11) unsigned not null comment '队长id',
    `group_id` int(11) unsigned not null comment '禁止同桌分组id',
    `uid` int(11) unsigned default 0 comment '用户id,为零表示刚创建分组',
    `status` tinyint default 0 comment '状态0为正常，1为移除',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_hid_group_id` (`hid`,`puid`,`group_id`,`uid`),
  KEY `index_hid_puid_id` (`hid`,`puid`)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='包厢成员分组表';
*/

type HouseGroupUser struct {
	Id        int64     `gorm:"column:id;comment:'主键'"`
	Hid       int       `gorm:"column:hid;index:index_hid;comment:'包厢id'"`
	Puid      int64     `gorm:"column:puid;index:index_hid;comment:'队长id'"`
	GroupID   int       `gorm:"column:group_id;comment:'分组id'"`
	Uid       int64     `gorm:"column:uid;comment:'用户id'"`
	Status    int       `gorm:"column:status;comment:'状态，0为正常，1为失效'"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseGroupUser) TableName() string {
	return "house_group_user"
}
func initHouseGroupUser(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseGroupUser{}) {
		err = db.AutoMigrate(&HouseGroupUser{}).Error
	} else {
		err = db.CreateTable(&HouseGroupUser{}).Error
	}
	return err
}
