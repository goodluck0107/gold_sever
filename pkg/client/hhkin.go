package client

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
)

// 得到玩家信息
func GetDeliveryInfo(host string, uid int64) (string, error) {
	// self.RobotHost = "http://robot.hhkin.com"
	reqUrl := fmt.Sprintf("%s?id=%d", host, uid)
	data, err := util.HttpGet(reqUrl, nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		Data struct {
			Id      int64  `json:"id"`
			UserId  int64  `json:"userid"`
			UserImg string `json:"userimg"`
		} `json:"data"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return "", err
	}
	if resp.Code != 1 {
		return "", fmt.Errorf("get delivery eve failed. code:%d, msg:%s", resp.Code, resp.Message)
	}
	if resp.Data.UserId != uid {
		return "", fmt.Errorf("user id error. request:%d, response:%d", uid, resp.Data.UserId)
	}
	return resp.Data.UserImg, nil
}

// 推送战绩到线聊
// http://update.hhkin.com/web/#/47?page_id=1241
func PushRecordToXianTalk(host string, msg *static.MsgXianTalkRecord) error {
	// self.RobotHost = "http://robot.hhkin.com"
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	data, err := util.HttpPost(host, header, static.HF_JtoA(msg))
	if err != nil {
		return fmt.Errorf("http post request error:%v", err)
	}

	var resp struct {
		NResult  int    `json:"nresult"`
		VcResult string `json:"vcresult"`
	}

	// fmt.Println(string(data))
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("json unmarshal error.data:%s, error:%v", string(data), err)
	}

	if resp.NResult != 0 {
		return fmt.Errorf("push data error. code:%d, msg:%s", resp.NResult, resp.VcResult)
	}
	return nil
}
