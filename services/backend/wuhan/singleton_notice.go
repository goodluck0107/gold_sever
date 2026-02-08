package wuhan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

var noticemgrsingleton *NoticeMgr = nil

type NoticeMgr struct {
	mu *lock2.RWMutex
}

func GetNoticeMgr() *NoticeMgr {
	if noticemgrsingleton == nil {
		noticemgrsingleton = new(NoticeMgr)
		noticemgrsingleton.mu = new(lock2.RWMutex)
	}
	return noticemgrsingleton
}

func (nm *NoticeMgr) Update() {
	// 获取函数执行时间
	defer static.HF_FuncElapsedTime()()
	// 防止并发写入
	nm.mu.Lock()
	defer nm.mu.Unlock()
	err := nm.StoreData(nm.GetData())
	if err != nil {
		xlog.Logger().Panic("notice redis store data error:", err)
	}
	xlog.Logger().Info("load notice data succeed.")
}

func (nm *NoticeMgr) url() string {
	return fmt.Sprintf("%s/api/notice/byType", GetServer().Con.AdminHost)
}

func (nm *NoticeMgr) GetData() map[static.NoticePositionType]static.NoticeList {
	url := nm.url()
	xlog.Logger().Info("noticeUrl:", url)
	data, err := util.HttpGet(url, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil
	}

	// 去掉bom头
	data = bytes.TrimPrefix(data, []byte{239, 187, 191})

	var result = struct {
		Code int               `json:"code"`
		Msg  string            `json:"msg"`
		Data static.NoticeList `json:"data"`
	}{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil
	}

	if result.Code != 0 {
		err = errors.New(result.Msg)
		xlog.Logger().Errorln(err)
		return nil
	}

	// 处理数据
	// for _, notice := range result.Data {
	// 	notice.Start, _ = time.ParseInLocation("2006-01-02 15:04:05", notice.StartAt, time.Local)
	// 	notice.End, _ = time.ParseInLocation("2006-01-02 15:04:05", notice.EndAt, time.Local)
	// }

	return result.Data.ToMap()
}

func (nm *NoticeMgr) StoreData(data map[static.NoticePositionType]static.NoticeList) error {
	if data == nil {
		return errors.New("notice nil data to store")
	}

	// 维护公告特俗处理
	maintainNotices := make(static.NoticeList, 0)
	err := GetDBMgr().GetDBrControl().RedisV2.LRange(static.NoticePTypeRedisKey(static.NoticePositionTypeMarquee), 0, -1).ScanSlice(&maintainNotices)
	if err != nil && err != redis.Nil {
		return err
	}

	// 清楚数据
	err = nm.CleanData()
	if err != nil {
		return err
	}

	for npt, notices := range data {

		if npt == static.NoticePositionTypeMaintain {
			for i := 0; i < len(notices); i++ {
				for j := 0; j < len(maintainNotices); j++ {
					if notices[i].Id == maintainNotices[j].Id {
						new, _ := time.ParseInLocation("2006-01-02 15:04:05", notices[i].UpdateAt, time.Local)
						old, _ := time.ParseInLocation("2006-01-02 15:04:05", maintainNotices[j].UpdateAt, time.Local)
						if new.Equal(old) {
							notices[i].Flag = maintainNotices[j].Flag
						}
					}
				}
			}
		}

		if err := GetDBMgr().GetDBrControl().RedisV2.LPush(static.NoticePTypeRedisKey(npt), notices.ToObjects()...).Err(); err != nil {
			return err
		}

	}
	return nil
}

func (nm *NoticeMgr) CleanData() error {
	// return public.RedisBatchDeleteScript.Run(GetDBMgr().GetDBrControl().RedisV2, []string{public.NoticeDataRedisKey()}).Err()
	keys := GetDBMgr().GetDBrControl().RedisV2.Keys(static.NoticeDataRedisKey()).Val()

	if len(keys) > 0 {
		return GetDBMgr().GetDBrControl().RedisV2.Del(keys...).Err()
	}

	return nil
}
