package models

type UserAgentConfig struct {
	Id             int64  `gorm:"primary_key;column:id"`
	Uid            int64  `gorm:"column:uid"`
	ClerkId        int64  `gorm:"column:clerk_id"`
	IsVipFloors    bool   `gorm:"column:floor_auth"`
	IsUnion        bool   `gorm:"column:union_switch"`
	VitaminPoolMax int64  `gorm:"column:max"`
	State          bool   `gorm:"column:state"`
	Username       string `gorm:"column:username"`
	Password       string `gorm:"column:password"`
}

func (UserAgentConfig) TableName() string {
	return "user_agent_config"
}

type UserAgentConfigKindId struct {
	Id      int64 `gorm:"primary_key;column:id"`
	Uid     int64 `gorm:"column:uid"`
	ClerkId int64 `gorm:"column:clerk_id"`
	KindId  int   `gorm:"column:kindid"`
}

func (UserAgentConfigKindId) TableName() string {
	return "user_agent_config_kindid"
}
