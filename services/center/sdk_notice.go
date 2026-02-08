package center

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	common "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

// func BroadcastMaintain() {
// 	defer func() {
// 		if x := recover(); x != nil {
// 			syslog.Logger().Errorln(x, string(debug.Stack()))
// 		}
// 	}()
//
// 	ticker := time.NewTicker(time.Second * 30)
// 	key := public.NoticePTypeRedisKey(public.NoticePositionTypeMaintain)
//
// 	pushed := make([]int, 0)
// 	hasPushed := func(nid int) bool {
// 		for _, id := range pushed {
// 			if id == nid {
// 				return true
// 			}
// 		}
// 		return false
// 	}
//
// 	for {
// 		if GetServer().ShutDown {
// 			break
// 		}
// 		<-ticker.C
// 		maintainNotices := make(public.NoticeList, 0)
//
// 		err := GetDBMgr().GetDBrControl().RedisV2.LRange(key, 0, -1).ScanSlice(&maintainNotices)
//
// 		if err != nil {
// 			if err == redis.Nil {
// 				syslog.Logger().Debug("redis nil:", err)
// 				continue
// 			}
// 			syslog.Logger().Errorln(err)
// 			continue
// 		}
//
// 		maintainNotices.GameTimer()
// 		now := time.Now()
//
// 		for _, n := range maintainNotices {
// 			if n == nil {
// 				continue
// 			}
//
// 			if n.Start.IsZero() {
// 				continue
// 			}
//
// 			if hasPushed(n.Id) {
// 				continue
// 			}
//
// 			if now.After(n.Start) {
// 				if n.End.IsZero() || n.End.After(now) {
// 					pushed = append(pushed, n.Id)
// 					if n.KindId != 0 {
// 						// 子游戏维护, 通知对应的服务器
// 						servers := GetServer().GetGamesByKindId(n.KindId)
// 						for _, s := range servers {
// 							_, _ = GetServer().CallGame(s.Id, 0, "NewServerMsg", "servermaintain", common.SuccessCode, &public.Msg_HG_UpdateGameServer{KindId: n.KindId})
// 						}
// 					} else {
// 						// 全服维护广播
// 						GetPlayerMgr().Broadcast(constant.MsgTypeMaintainNotice, n)
// 						GetServer().BroadcastGame(0, "NewServerMsg", "servermaintain", common.SuccessCode, &public.Msg_HG_UpdateGameServer{KindId: n.KindId})
// 					}
// 					break
// 				}
// 			}
// 		}
// 	}
// 	ticker.Stop()
// }

func checkMaintainNotice(t time.Duration) {
	if t <= 0 {
		t = time.Second * 10
	}

	// syslog.Logger().Infof("Check the countdown for maintenance notices:%s", t)
	timer := time.NewTimer(t)
	defer timer.Stop()

	for !GetServer().ShutDown {
		<-timer.C
		// syslog.Logger().Info("Time's up, check maintenance notices!")
		tt := BroadcastMaintain()
		// syslog.Logger().Infof("The start time of the next maintenance notices:%s", tt)
		if tt > 0 {
			timer.Reset(tt)
		} else {
			timer.Reset(time.Second * 10)
		}
	}
}

func BroadcastMaintain() time.Duration {
	key := static.NoticePTypeRedisKey(static.NoticePositionTypeMaintain)
	maintainNotices := make(static.NoticeList, 0)

	err := GetDBMgr().GetDBrControl().RedisV2.LRange(key, 0, -1).ScanSlice(&maintainNotices)

	if err != nil {
		if err == redis.Nil {
			// syslog.Logger().Debug("redis nil:", err)
			return 0
		}
		xlog.Logger().Errorln(err)
		return time.Second * 10 // 下次继续
	}

	maintainNotices.Timer()

	nextTime := maintainNotices.WaitTime()

	maintainNotices.RmExpired()

	for index, n := range maintainNotices {
		if n == nil {
			continue
		}
		if n.IsWorked() {
			continue
		}
		n.SetWorked()
		err = GetDBMgr().GetDBrControl().RedisV2.LSet(key, int64(index), n).Err()
		if err != nil {
			xlog.Logger().Error(err)
			continue
		}
		switch n.GameServerId {
		case static.NoticeMaintainServerAllServer:
			GetPlayerMgr().Broadcast(consts.MsgTypeMaintainNotice, n)
			GetServer().BroadcastGame(0, "NewServerMsg", "servermaintain", common.SuccessCode, &static.Msg_HG_UpdateGameServer{GameId: n.GameServerId})
		case static.NoticeMaintainServerAllGame:
			GetServer().BroadcastGame(0, "NewServerMsg", "servermaintain", common.SuccessCode, &static.Msg_HG_UpdateGameServer{GameId: n.GameServerId})
		default:
			_, _ = GetServer().CallGame(int(n.GameServerId), 0, "NewServerMsg", "servermaintain", common.SuccessCode, &static.Msg_HG_UpdateGameServer{GameId: n.GameServerId})
		}
	}
	return nextTime
}

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
