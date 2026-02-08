package client

type CommonConfig struct {
	AdminHost    string `json:"adminhost"` // 游戏GM后台域名
	DTRobotToken string `json:"drobot"`    // 钉钉群聊名字:群聊机器人token
	RobotHost    string `json:"robothost"` // http://robot.hhkin.com
}

func InitConfig(c CommonConfig) {
	InitAdminClient(&c)
	InitDingTalkClient(&c)
}

func GetAdminClient() *AdminClient {
	return adminClient
}

func GetDingTalkRobotClient() *DingTalkRobot_Client {
	return dingtalkrobotClient
}
