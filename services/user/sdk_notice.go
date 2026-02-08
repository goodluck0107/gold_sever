package user

import (
	"fmt"
	"time"

	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

// 根据游戏id得到维护公告 0为大厅
func GetNoticeMaintain(id static.NoticeMaintainServerType) *static.Notice {
	key := static.NoticePTypeRedisKey(static.NoticePositionTypeMaintain)

	maintainNotices := make(static.NoticeList, 0)

	err := GetDBMgr().GetDBrControl().RedisV2.LRange(key, 0, -1).ScanSlice(&maintainNotices)

	if err != nil {
		xlog.Logger().Errorln(err)
		return nil
	}

	maintainNotices.Timer()

	maintainNotices.RmExpired()

	for _, mn := range maintainNotices {
		if mn == nil {
			continue
		}

		switch mn.GameServerId {
		case static.NoticeMaintainServerAllServer:
			return mn
		case static.NoticeMaintainServerAllGame:
			if id != static.NoticeMaintainServerAllServer {
				return mn
			}
		default:
			if mn.GameServerId == id {
				return mn
			}
		}
	}
	return nil
}

// 得到大厅普通公告列表
func GetNoticeDialogs() static.NoticeList {
	key := static.NoticePTypeRedisKey(static.NoticePositionTypeDialog)

	dialogNotices := make(static.NoticeList, 0)

	err := GetDBMgr().GetDBrControl().RedisV2.LRange(key, 0, -1).ScanSlice(&dialogNotices)

	if err != nil {
		xlog.Logger().Errorln(err)
		return nil
	}
	dialogNotices.Timer()
	dialogNotices.RmExpired()
	return dialogNotices
}

func GetNoticeMarqueeNotice() static.NoticeList {
	key := static.NoticePTypeRedisKey(static.NoticePositionTypeMarquee)

	marqueeNotices := make(static.NoticeList, 0)

	err := GetDBMgr().GetDBrControl().RedisV2.LRange(key, 0, -1).ScanSlice(&marqueeNotices)

	if err != nil {
		xlog.Logger().Errorln(err)
		return nil
	}
	marqueeNotices.Timer()
	marqueeNotices.RmExpired()
	return marqueeNotices
}

func GetNoticeOptions() static.NoticeList {
	key := static.NoticePTypeRedisKey(static.NoticePositionTypeOption)

	optionsNotices := make(static.NoticeList, 0)

	err := GetDBMgr().GetDBrControl().RedisV2.LRange(key, 0, -1).ScanSlice(&optionsNotices)

	if err != nil {
		xlog.Logger().Errorln(err)
		return nil
	}
	return optionsNotices
}

func CheckServerMaintainWithWhite(uid int64, id static.NoticeMaintainServerType) *static.Notice {
	mn := GetNoticeMaintain(id)
	if mn == nil {
		return nil
	}
	uWhite := new(models.UserWhite)
	if err := GetDBMgr().GetDBmControl().Where("uid = ?", uid).First(uWhite).Error; err == nil {
		xlog.Logger().Warningln("白名单用户,不给维护公告。uid:", uid)
		return nil
	}
	return mn
}

// 判断某个公告是否展示
func IsNoticeShowToday(noticeId int, uid int64) bool {
	key := fmt.Sprintf("notice_%d_%d", uid, noticeId)
	return GetDBMgr().GetDBrControl().RedisV2.Exists(key).Val() == 1
}

// 设置某个公告已经展示
func SetNoticeShowToday(noticeId int, uid int64) {
	key := fmt.Sprintf("notice_%d_%d", uid, noticeId)
	err := GetDBMgr().GetDBrControl().RedisV2.Set(key, time.Now().Format(static.TIMEFORMAT), time.Duration(static.HF_GetTodayRemainSecond())*time.Second).Err()
	if err != nil {
		xlog.Logger().Errorln("SetNoticeShowToday error:", err)
	}
}
