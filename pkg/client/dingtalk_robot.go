// 钉钉群聊机器人服务
package client

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/sirupsen/logrus"
	"strings"
)

// 钉钉群聊机器人
type DingTalkRobot_Client struct {
	DTRobotToken string
}

var dingtalkrobotClient *DingTalkRobot_Client

func InitDingTalkClient(c *CommonConfig) {
	dingtalkrobotClient = new(DingTalkRobot_Client)
	dingtalkrobotClient.DTRobotToken = c.DTRobotToken
}

func (self *DingTalkRobot_Client) SendText(group, content string, atAll bool, phone ...string) error {
	var msg static.Msg_DingTalk_Text
	msg.MsgType = "text"
	msg.Text.Content = content
	msg.At.AtMobiles = phone
	msg.At.IsAtAll = atAll
	go self.CallDingTalkRobot(group, msg)
	return nil
}

func (self *DingTalkRobot_Client) SendLink(group, title, text, msgUrl, picUrl string) error {
	var msg static.Msg_DingTalk_Link
	msg.MsgType = "link"
	msg.Link.Title = title
	msg.Link.Text = text
	if len(msgUrl) == 0 {
		// 默认路径
		msgUrl = "www.facai.com"
	}
	if len(picUrl) == 0 {
		// 默认图片
		picUrl = "www.facai.com/img/logo.png"
	}
	prefix := "https://"
	if !strings.HasPrefix(msgUrl, prefix) {
		msgUrl = fmt.Sprintf("%s%s", prefix, msgUrl)
	}
	if !strings.HasPrefix(picUrl, prefix) {
		picUrl = fmt.Sprintf("%s%s", prefix, picUrl)
	}
	msg.Link.MessageUrl = msgUrl
	msg.Link.PicUrl = picUrl
	go self.CallDingTalkRobot(group, msg)
	return nil
}

// 调用钉钉群聊机器人api
func (self *DingTalkRobot_Client) CallDingTalkRobot(group string /*群名称*/, msg interface{} /*协议消息内容*/) error {
	// access_token, ok := self.DTRobotToken[group]
	// if !ok {
	// 	return errors.New(fmt.Sprintf("Group:[%s] chat robot url does not exist", group))
	// }
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["encoding"] = "utf-8"
	body, err := util.HttpPost(fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", self.DTRobotToken), header, static.HF_JtoA(&msg))
	if err != nil {
		return err
	}
	type response struct {
		ErrMag  string `json:"errmsg"`
		ErrCode int    `json:"errcode"`
	}
	result := new(response)
	err = json.Unmarshal(body, result)
	xlog.Logger().WithFields(logrus.Fields{
		"code": result.ErrCode,
		"msg":  result.ErrMag,
		"data": string(body),
	}).Infoln("dingTalk response..")
	if err != nil {
		return err
	}
	return nil
}
