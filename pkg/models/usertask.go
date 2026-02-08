package models

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"
)

// 任务状态
type UserTask struct {
	Id        int64     `gorm:"primary_key;column:id"`              // 主键id
	TaskId    int       `gorm:"column:taskid;index:idx_uid2taskid"` //! 任务配置表id
	Uid       int64     `gorm:"column:uid;index:idx_uid2taskid"`    //! 用户id
	Num       int       `gorm:"column:num"`                         //! 任务当前计数
	Step      int       `gorm:"column:step"`                        //! 任务当前进度（分阶段）
	Sta       int       `gorm:"column:sta"`                         //! 任务状态
	Time      time.Time `gorm:"column:time"`                        //! 最后操作时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"`    //! 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"`    //! 更新时间
}

func (UserTask) TableName() string {
	return "user_tasks"
}

func (self *UserTask) ConvertModel() *static.Task {
	task := new(static.Task)
	task.Id = self.Id
	task.TcId = self.TaskId
	task.Uid = self.Uid
	task.Num = self.Num
	task.Sta = self.Sta
	task.Step = self.Step
	task.Time = self.Time.Unix()
	return task
}

func initUserTask(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserTask{}) {
		err = db.AutoMigrate(&UserTask{}).Error
	} else {
		err = db.CreateTable(&UserTask{}).Error
	}
	return err
}

// 任务日志
type UserTaskRewardLog struct {
	Id         int       `gorm:"column:id"`     //! id
	TaskId     int       `gorm:"column:taskid"` //! id
	Type       int       `gorm:"column:type"`   //! 任务类型
	Kind       int       `gorm:"column:kind"`
	KindId     int       `gorm:"column:kind_id"`
	Uid        int64     `gorm:"column:uid"`  //! 用户id
	Step       int       `gorm:"column:step"` //!
	RewardType int       `gorm:"column:rewardtype"`
	RewardNum  int       `gorm:"column:rewardnum"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (UserTaskRewardLog) TableName() string {
	return "user_task_reward_log"
}

func initUserTaskRewardLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserTaskRewardLog{}) {
		err = db.AutoMigrate(&UserTaskRewardLog{}).Error
	} else {
		err = db.CreateTable(&UserTaskRewardLog{}).Error
	}
	return err
}

//游戏内显示的任务
// 任务状态
type UserTaskGame struct {
	Id        int64     `gorm:"primary_key;column:id"`
	TaskId    int       `gorm:"column:taskid;index:idx_uid2taskid"` //! id
	Num       int       `gorm:"column:num"`                         //! 任务完成数量
	Uid       int64     `gorm:"column:uid;index:idx_uid2taskid"`    //! 用户id
	Step      int       `gorm:"column:step"`                        //! 进度
	Time      time.Time `gorm:"column:time"`                        //! 最后操作时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"`    //! 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"`    //! 更新时间
}

func (UserTaskGame) TableName() string {
	return "user_tasks_game"
}

func (self *UserTaskGame) ConvertModel() *static.Task {
	task := new(static.Task)
	task.Id = self.Id
	task.TcId = self.TaskId
	task.Uid = self.Uid
	task.Num = self.Num
	task.Step = self.Step
	task.Time = self.Time.Unix()
	return task
}

func initUserTaskGame(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserTaskGame{}) {
		err = db.AutoMigrate(&UserTaskGame{}).Error
	} else {
		err = db.CreateTable(&UserTaskGame{}).Error
	}
	return err
}

// 任务日志
type UserTaskGameRewardLog struct {
	Id         int       `gorm:"column:id"`     //! id
	TaskId     int       `gorm:"column:taskid"` //! id
	Type       int       `gorm:"column:type"`   //! 任务类型
	Kind       int       `gorm:"column:kind"`
	KindId     int       `gorm:"column:kind_id"`
	SiteType   int       `gorm:"column:sitetype;default:0"` //! 场次类型 0表示所有场次
	Uid        int64     `gorm:"column:uid"`                //! 用户id
	Step       int       `gorm:"column:step"`               //!
	RewardType int       `gorm:"column:rewardtype"`
	RewardNum  int       `gorm:"column:rewardnum"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (UserTaskGameRewardLog) TableName() string {
	return "user_task_game_reward_log"
}

func initUserTaskGameRewardLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserTaskGameRewardLog{}) {
		err = db.AutoMigrate(&UserTaskGameRewardLog{}).Error
	} else {
		err = db.CreateTable(&UserTaskGameRewardLog{}).Error
	}
	return err
}
