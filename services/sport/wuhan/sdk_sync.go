package wuhan

import (
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

//func PushMsg(head string, msg interface{}) error {
//	buf, err := json.Marshal(msg)
//	if err != nil {
//		syslog.Logger().Errorf("json error：%+v", msg)
//		return err
//	}
//	res := GetDBMgr().Redis.LPush(head, buf).Val()
//	if res == 1 {
//		return nil
//	}
//	return fmt.Errorf("redis push error")
//
//}

func PushTableStatusMsg(hid int, head string, msg interface{}) error {
	var err error
	redisMsg := static.G2HTableStatsChangeNtf{}
	redisMsg.Hid = hid
	redisMsg.Head = head
	buf, err := json.Marshal(msg)
	if err != nil {
		xlog.Logger().Errorf("json error：%+v", msg)
		return err
	}
	redisMsg.Data = string(buf)

	redisBuf, err := json.Marshal(redisMsg)

	err = GetDBMgr().Redis.LPush(consts.MsgTypeTableStatus_Ntf, redisBuf).Err()

	return err
}
