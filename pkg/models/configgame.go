package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
)

//游戏配置
type GameConfig struct {
	Id               int64               `gorm:"primary_key;column:id"`
	KindId           int                 `gorm:"unique;column:kindid"`            // 游戏玩法id
	Name             string              `gorm:"column:name;size:50"`             // 游戏玩法名称
	EnName           string              `gorm:"column:enname;size:50"`           // 游戏玩法名称英文
	PlayerNum        string              `gorm:"column:playernum;size:255"`       // 游戏开始人数,形如3,4
	DefaultPlayerNum int                 `gorm:"column:defaultplayernum"`         // 游戏开始默认人数
	RoundNum         string              `gorm:"column:roundnum;size:255"`        // 局数限制,形如{"8":1,"16":2} 多少局对应扣多少卡
	DefaultRoundNum  int                 `gorm:"column:defaultroundnum"`          // 默认游戏局数
	DefaultCostType  int                 `gorm:"column:defaultcosttype"`          // 默认支付方式
	DefaultRestrict  bool                `gorm:"column:defaultrestrict"`          // 默认ip限制
	DefaultView      bool                `gorm:"column:defaultview"`              // 默认是否允许观看
	GameConfig       string              `gorm:"column:gameconfig;type:text"`     // 游戏玩法特色配置
	Players          []int               `gorm:"-"`                               // 对应PlayerNum配置
	RoundMap         map[int]map[int]int `gorm:"-"`                               // 对应RoundNum配置
	Version          int                 `gorm:"column:version"`                  // 子游戏版本号
	IsSupportWatch   bool                `gorm:"column:issupportwatch;default:0"` // 是否支持观战
	IsLimitChannel   bool                `gorm:"column:islimitchannel;default:0"` // 是否限制渠道同服
}

func (GameConfig) TableName() string {
	return "config_games"
}

func initGameConfig(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameConfig{}) {
		err = db.AutoMigrate(&GameConfig{}).Error
	} else {
		err = db.CreateTable(&GameConfig{}).Error
	}
	return err
}

func (c *GameConfig) AfterFind() error {
	tempArr := strings.Split(c.PlayerNum, ",")
	for _, item := range tempArr {
		k, _ := strconv.Atoi(item)
		c.Players = append(c.Players, k)
	}
	sort.Ints(c.Players)
	return json.Unmarshal([]byte(c.RoundNum), &c.RoundMap)
}

// 获取房卡消耗
func (c *GameConfig) GetCardCost(round int, players int, NoAA bool) int {
	cost := c.RoundMap[players][round]
	if cost <= 0 {
		return 0
	}
	if NoAA {
		return cost
	}
	if players <= 0 {
		return cost
	}
	extra := cost % players
	if extra != 0 {
		return cost/players + 1
	}
	return cost / players
}

// 检测玩家人数
func (c *GameConfig) CheckPlayerNum(num int) bool {
	for _, item := range c.Players {
		if item == num {
			return true
		}
	}
	return false
}

// 获取最小最大玩家人数
func (c *GameConfig) GetPlayerNum() (int, int) {
	if len(c.Players) == 0 {
		return 0, 0
	}
	return c.Players[0], c.Players[len(c.Players)-1]
}

// 检测游戏局数
func (c *GameConfig) CheckRoundNum(num int, playernum int) bool {
	_, ok := c.RoundMap[playernum][num]
	return ok
}

func (c *GameConfig) String() string {
	return fmt.Sprintf("\n---------------%s----------------]:\n[Id]:%d\n[KindId]:%d\n[Name]:%s\n[EnName]:%s\n[PlayerNum]:%s\n[DefaultPlayerNum]:%d\n[RoundNum]:%s\n[DefaultRoundNum]:%d\n[DefaultCostType]:%d\n[DefaultRestrict]:%t\n[DefaultView]:%t\n[GameConfig]:%s", c.TableName(), c.Id, c.KindId, c.Name, c.EnName, c.PlayerNum, c.DefaultPlayerNum, c.RoundNum, c.DefaultRoundNum, c.DefaultCostType, c.DefaultRestrict, c.DefaultView, c.GameConfig)
}
