package center

import (
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"runtime/debug"
	"time"
)

// 任务奖励列表
type TaskRewardList []*TaskReward

// 任务奖励 "wealth_type":5,"num":100
type TaskReward struct {
	WealthType int `json:"wealth_type"` // 奖励的财富类型
	Num        int `json:"num"`         // 奖励数量
}

// 任务配置相关属性
type TaskCfg struct {
	Id             int            // id
	MainType       int            // 任务主类型 0 每日任务 1 系统任务
	SubType        int            // 任务子类型 分享任务 对局次数 胜场次数
	Area           string         // 任务是否绑定区域 默认为空字符串不绑定区域
	Sort           int            // 任务排序的标记
	TgtCompleteNum int            // 任务完成所需要的数量
	Reward         TaskRewardList // 任务完成的奖励[{"wealth_type":1,"num":100},{"wealth_type":1,"num":100}]
	GameKindId     int            // 任务与指定游戏的kind id
	StepTaskId     int            // 阶段任务表id
	Desc           string         // 任务描述
	RewardDesc     string         // 奖励描述
}

///////////////////////////////////////////////////////////////////////////////////////////////////////
// 任务管理器

// 管理者
type TasksMgr struct {
	receiveChan chan *TaskMsg
	MapTasksCfg map[int]*TaskCfg
}

var tasksMgrSingleton *TasksMgr = nil

func GetTasksMgr() *TasksMgr {
	if tasksMgrSingleton == nil {
		tasksMgrSingleton = new(TasksMgr)
		tasksMgrSingleton.receiveChan = make(chan *TaskMsg, 2000)
		tasksMgrSingleton.MapTasksCfg = make(map[int]*TaskCfg)
	}
	return tasksMgrSingleton
}

// 任务管理器初始化
func (tm *TasksMgr) Init() bool {
	tm.InitConfig()
	go tasksMgrSingleton.Run()
	return true
}

// 任务运行
func (tm *TasksMgr) Run() {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			// 定时器检查
		case msg, ok := <-tm.receiveChan:
			if !ok {
				continue
			}
			// 从管道中读取更新信息
			tm.OnTaskMsg(msg)
		}
	}
}

// 加载任务列表
func (tm *TasksMgr) InitConfig() bool {
	var err error
	// 读取任务配置
	var configs []*models.ConfigTask
	if err = GetDBMgr().db_M.Find(&configs).Error; err != nil {
		xlog.Logger().Error(err)
		return false
	}

	for _, c := range configs {
		t := &TaskCfg{
			Id:             c.Id,
			MainType:       c.MainType,
			SubType:        c.SubType,
			Area:           c.Area,
			Sort:           c.Sort,
			TgtCompleteNum: c.TgtCompleteNum,
			GameKindId:     c.GameKindId,
			StepTaskId:     c.StepTaskId,
			Desc:           c.Desc,
			RewardDesc:     c.RewardDesc,
		}

		// 任务奖励
		if err = json.Unmarshal([]byte(c.Reward), &t.Reward); err != nil {
			xlog.Logger().Errorln(err.Error())
		}

		tm.MapTasksCfg[t.Id] = t
	}

	return true
}

// 创建用户任务
func (tm *TasksMgr) CreateUserTasks(uid int64) static.TaskList {
	tl := make(static.TaskList, 0)
	for _, tc := range tm.MapTasksCfg {
		task := NewTask(tc, uid)
		if err := GetDBMgr().InsertUserTask(task); err != nil {
			xlog.Logger().Error("new task err : ", err)
			continue
		}
		tl = append(tl, task)
	}
	return tl
}

// 检查用户任务是否完整
func (tm *TasksMgr) CheckUserTasks(uid int64) {
	// 用户区域码
	userArea := ""
	if per := GetPlayerMgr().GetPlayer(uid); per != nil {
		userArea = per.Info.Area
	}

	// 获取用户的任务列表
	tl := tm.GetUserTaskList(uid)

	// 是否存在完成未领取的任务
	bIsHaveCompletedTask := false

	// 检查用户每日任务的时间
	for _, task := range tl {
		tc, ok := tm.MapTasksCfg[task.TcId]
		if !ok {
			continue
		}
		if tc.MainType == consts.TASK_TYPE_DAYILY && time.Now().Format(consts.TIME_Y_M_D) != time.Unix(task.Time, 0).Format(consts.TIME_Y_M_D) {
			task.Num = 0
			task.Step = 0
			task.Sta = consts.TASK_STA_DOING
			task.Time = time.Now().Unix()
			// 每日签到功能自动完成
			if tc.SubType == consts.TASK_KIND_DAYILY_SIGN {
				task.Num = 1
				task.Sta = consts.TASK_STA_COMPLETED
			}
			// 更新redis
			if err := GetDBMgr().UpdateUserTask(task); err != nil {
				xlog.Logger().Errorln("check task delete err: ", err.Error())
			}
		}
		if task.Sta == consts.TASK_STA_COMPLETED {
			if len(tc.Area) == 0 || (len(tc.Area) > 0 && tc.Area == userArea) {
				bIsHaveCompletedTask = true
			}
		}
	}

	// 检查是否有完成未领取的任务
	if bIsHaveCompletedTask {
		// 发送任务完成推送
		if p := GetPlayerMgr().GetPlayer(uid); p != nil {
			p.SendMsg(consts.MsgTypeTaskCompletedNtf, &static.Msg_Null{})
		}
	}

	// 清理无效任务数据
	for _, task := range tl {
		if _, ok := tm.MapTasksCfg[task.TcId]; !ok {
			if err := GetDBMgr().DeleteUserTask(task); err != nil {
				xlog.Logger().Error("check task delete err: ", err)
			}
		}
	}

	// 创建新增任务
	for _, tc := range tm.MapTasksCfg {
		isExist := false
		for _, task := range tl {
			if task.TcId == tc.Id {
				isExist = true
				break
			}
		}
		if !isExist {
			task := NewTask(tc, uid)
			if err := GetDBMgr().InsertUserTask(task); err != nil {
				xlog.Logger().Error("new task err : ", err)
			}
		}
	}
}

// 获取用户任务列表
func (tm *TasksMgr) GetUserTaskList(uid int64) static.TaskList {
	// 从redis读取用户任务数据
	tl, err := GetDBMgr().GetDBrControl().UserTaskList(uid)
	if err != nil {
		xlog.Logger().Error("get tasks error : ", err)
		return nil
	}
	// 防止出现空数据
	if len(tl) == 0 {
		return tm.CreateUserTasks(uid)
	}
	return tl
}

// 获取用户具体任务
func (tm *TasksMgr) GetUserTask(uid int64, tcId int) *static.Task {
	tl := tm.GetUserTaskList(uid)
	for _, task := range tl {
		if task.TcId == tcId {
			return task
		}
	}
	return nil
}

// 重置每日任务
func (tm *TasksMgr) ResetDailyTasks(task *static.Task) bool {
	return false
}

// 获取任务配置
func (tm *TasksMgr) GetTaskConfig(id int) *TaskCfg {
	tc, ok := tm.MapTasksCfg[id]
	if !ok {
		xlog.Logger().Error("GetTaskConfig nil err id:", id)
		return nil
	}
	return tc
}

// 更新任务状态
func (tm *TasksMgr) UpdateTaskSta(uid int64, tcId int, num int, v interface{}, isAddNum bool) {
	msg := NewMsg2TaskMgr(tcId, num, uid, v, isAddNum)
	if tm.receiveChan != nil {
		tm.receiveChan <- msg
	}
}

// 更新游戏相关任务信息
func (tm *TasksMgr) UpdateGameTaskSta(uid int64, subType int, kindId int, num int, cardType int) {
	tl := tm.GetUserTaskList(uid)
	for _, task := range tl {
		tc := tm.GetTaskConfig(task.TcId)
		if tc.SubType == subType && tc.GameKindId == kindId {
			msg := NewMsg2TaskMgr(task.TcId, num, uid, nil, true)
			if tm.receiveChan != nil {
				tm.receiveChan <- msg
			}
		}
	}
}

// 更新任务信息
func (tm *TasksMgr) OnTaskMsg(msg *TaskMsg) {
	task := tm.GetUserTask(msg.Uid, msg.Id)
	if task == nil {
		xlog.Logger().Error("OnTaskMsg err task nil id:", msg.Id, "--uid:", msg.Uid)
		return
	}

	taskCfg := GetTasksMgr().GetTaskConfig(msg.Id)
	if taskCfg == nil {
		xlog.Logger().Error("OnTaskMsg err task config nil id:", msg.Id, "--uid:", msg.Uid)
		return
	}

	if taskCfg.MainType == consts.TASK_TYPE_DAYILY {
		// 校验每日任务时间
		if time.Now().Format(consts.TIME_Y_M_D) != time.Unix(task.Time, 0).Format(consts.TIME_Y_M_D) {
			task.Num = 0
			task.Step = 0
			task.Sta = consts.TASK_STA_DOING
			task.Time = time.Now().Unix()
		}

		// 更新任务数据
		if task.Sta != consts.TASK_STA_DOING {
			return
		}

		// 更新任务计数
		if msg.Add {
			task.Num += msg.Num
		} else {
			task.Num = msg.Num
		}

		// 任务更新时间
		task.Time = time.Now().Unix()

		if task.Num >= taskCfg.TgtCompleteNum {
			// 更新任务状态
			task.Sta = consts.TASK_STA_COMPLETED
			// 发送任务完成推送
			if p := GetPlayerMgr().GetPlayer(task.Uid); p != nil && p.Info.TableId == 0 {
				var taskItem static.TaskItem
				taskItem.Id = task.TcId
				taskItem.MainType = taskCfg.MainType
				taskItem.SubType = taskCfg.SubType
				taskItem.TgtNum = taskCfg.TgtCompleteNum
				taskItem.Desc = taskCfg.Desc
				taskItem.RewardDesc = taskCfg.RewardDesc
				taskItem.Order = taskCfg.Sort
				taskItem.GameKindId = taskCfg.GameKindId
				taskItem.Num = task.Num
				taskItem.Sta = task.Sta
				p.SendMsg(consts.MsgTypeTaskCompletedNtf, &taskItem)
			}
		}
	} else {
		// else to do here
	}

	if err := GetDBMgr().UpdateUserTask(task); err != nil {
		xlog.Logger().Errorln("check task delete err: ", err.Error())
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////
// 任务

// 根据任务配置表新建一个任务
func NewTask(tc *TaskCfg, uid int64) *static.Task {
	if tc == nil {
		return nil
	}
	task := new(static.Task)
	task.TcId = tc.Id
	task.Uid = uid
	if tc.SubType == consts.TASK_KIND_DAYILY_SIGN {
		task.Num = 1
		task.Sta = consts.TASK_STA_COMPLETED
	} else {
		task.Num = 0
		task.Sta = consts.TASK_STA_DOING
	}
	task.Step = 0
	task.Time = time.Now().Unix()
	return task
}

///////////////////////////////////////////////////////////////////////////////////////////////////////
// 任务消息

type TaskMsg struct {
	Id  int         // 任务id
	Num int         // 完成任务数量
	Uid int64       // 用户id
	V   interface{} // 特殊任务扩展
	Add bool        // 添加状态 true， 修改任务状态false
}

func NewMsg2TaskMgr(id int, num int, uid int64, v interface{}, add bool) *TaskMsg {
	taskMsg := new(TaskMsg)
	taskMsg.Id = id
	taskMsg.Num = num
	taskMsg.Uid = uid
	taskMsg.V = v
	taskMsg.Add = add
	return taskMsg
}
