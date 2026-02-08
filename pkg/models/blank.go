/*
create table if not exists `black_user` (
`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
`uid` int(11) unsigned DEFAULT 0 COMMENT 'uid',
`hid` int(11) unsigned not null default 0 ,
`reason` varchar(200) not null default "",
`start` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
`end` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '结束时间' ,
`status` tinyint(1) not null default 0,
`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
PRIMARY KEY (`id`),
UNIQUE KEY `uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='封号用户';


create table if not exists `black_house` (
`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
`hid` int(11) unsigned not null  ,
`reason` varchar(200) not null default "",
`start` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
`end` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '结束时间' ,
`status` tinyint(1) not null default 0,
`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
PRIMARY KEY (`id`),
UNIQUE KEY `hid` (`hid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='封号圈';

*/
package models

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"

	"github.com/jinzhu/gorm"
)

func AddBlankUser(uid, endTime int64, reason string, status bool, db *gorm.DB) error {
	sql := `insert into black_user(uid,reason,end,status) values(?,?,?,?) on DUPLICATE KEY UPDATE reason = ?,end = ?,status=? `
	endData := time.Unix(endTime, 0)
	return db.Exec(sql, uid, reason, endData, status, reason, endData, status).Error
}

func AddBlankHouse(hid, endTime int64, reason string, status bool, db *gorm.DB) error {
	sql := `insert into black_house(hid,reason,end,status) values(?,?,?,?) on DUPLICATE KEY UPDATE reason = ?,end = ?,status=? `
	endData := time.Unix(endTime, 0)
	return db.Exec(sql, hid, reason, endData, status, reason, endData, status).Error
}

type BlankInfo struct {
	Reason string    `gorm:"reason"`
	End    time.Time `gorm:"end"`
	Status bool      `gorm:"status"`
}

func CheckUserInBlank(uid int64, db *gorm.DB) *BlankInfo {
	sql := `select reason,end,status from black_user where uid = ? limit 1`
	dest := []BlankInfo{}
	db.Raw(sql, uid).Scan(&dest)
	if len(dest) == 0 {
		return nil
	}
	if len(dest) > 1 {
		xlog.Logger().Errorf("error data:%d,uid:%d", len(dest), uid)
		return nil
	}
	if !dest[0].Status {
		return nil
	}
	if dest[0].End.Unix() < time.Now().Unix() {
		return nil
	}
	return &dest[0]
}
func CheckHouseInBlank(hid int64, db *gorm.DB) *BlankInfo {
	sql := `select reason,end,status from black_house where hid = ? limit 1`
	dest := []BlankInfo{}
	db.Raw(sql, hid).Scan(&dest)
	if len(dest) == 0 {
		return nil
	}
	if len(dest) > 1 {
		xlog.Logger().Errorf("error data:%d,uid:%d", len(dest), hid)
		return nil
	}
	if !dest[0].Status {
		return nil
	}
	if dest[0].End.Unix() < time.Now().Unix() {
		return nil
	}
	return &dest[0]
}
