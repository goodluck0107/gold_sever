package components

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	"math/rand"
)

/**
游戏任务
*/

type Task struct {
	//金豆任务相关变量
	MapAllTask      map[int]int //配置的金豆任务id和对应的金豆奖励数
	VecAllTaskID    []int       //配置了哪些金豆任务
	VecFinishedTask [meta2.MAX_PLAYER][]int
	RandomTask      int  //随机任务,下标
	FirstFinished   bool //是否首次完成任务
}

func (t *Task) Init() {
	t.MapAllTask = map[int]int{}
	t.VecAllTaskID = []int{}
	t.VecFinishedTask = [meta2.MAX_PLAYER][]int{}
	t.RandomTask = -1       //随机任务,下标
	t.FirstFinished = false //是否首次完成任务
}

//设置具体任务
func (t *Task) AppendTaskMapAndVec(id int, award int) {
	t.MapAllTask[id] = award
	t.VecAllTaskID = append(t.VecAllTaskID, id)
}

//发送随机任务  ,bFirst = false 表示断线重连,nSeat在断线重连时才有用
func (t *Task) SendTaskID(gamecommon Common, bFirst bool, nSeat uint16) {
	iSize := len(t.VecAllTaskID)
	if iSize > 0 {
		//生成随机数
		if bFirst {

			t.RandomTask = rand.Intn(1000) % iSize
		}

		if t.RandomTask < iSize && t.RandomTask >= 0 {
			if _, ok := t.MapAllTask[t.VecAllTaskID[t.RandomTask]]; ok {
				//发送给客户端
				var msgTask static.Msg_S_DG_TaskID

				msgTask.TaskID = t.VecAllTaskID[t.RandomTask]
				msgTask.TaskAward = t.MapAllTask[msgTask.TaskID]

				if bFirst {
					gamecommon.SendTableMsg(consts.RandTaskID, msgTask)
				} else {
					gamecommon.SendPersonMsg(consts.RandTaskID, msgTask, nSeat)
				}
			} else {
				//:-( 应该永远不会执行这句
				t.RandomTask = -1
			}
		}
	}
}

func (t *Task) IsTaskFinished(gamecommon Common, re static.TCardType, iSeat uint16, iNumOfKing int) {
	if re.Cardtype == static.TYPE_BOMB_510K && iNumOfKing == 0 {
		switch int(re.Color) {
		case logic2.CC_SPADE_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_HEI_510K)
			}
			break
		case logic2.CC_HEART_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_HONG_510K)
			}
			break
		case logic2.CC_CLUB_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_MEI_510K)
			}
			break
		case logic2.CC_DIAMOND_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FANG_510K)
			}
			break
		default:
			break
		}
	} else if re.Cardtype == static.TYPE_BOMB_NOMORL && re.Len == 4 && iNumOfKing == 0 {
		switch int(t.GetCardPoint(re.Card)) {
		case logic2.CP_2_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_4_2)
			}
			break
		case logic2.CP_5_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_4_5)
			}
			break
		case logic2.CP_10_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_4_10)
			}
			break
		case logic2.CP_K_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_4_K)
			}
			break
		default:
			break
		}
	} else if re.Cardtype == static.TYPE_BOMB_NOMORL && re.Len == 7 {
		t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_7_ZHA)
	} else if re.Cardtype == static.TYPE_BOMB_STR {
		t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_YAO_BAI)
	} else if t.FirstFinished == false && re.Cardtype == static.TYPE_TWOSTR && iNumOfKing == 0 {
		//首次打出连对
		if re.Len > 7 {
			t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_4_STR)
		} else if re.Len == 6 {
			switch int(t.GetCardPoint(re.Card)) {
			case logic2.CP_8_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_STR_678)
				}
				break
			case logic2.CP_9_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_STR_789)
				}
				break
			default:
				break
			}
		} else if re.Len == 4 {
			switch int(t.GetCardPoint(re.Card)) {
			case logic2.CP_A_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_STR_KA)
				}
				break
			case logic2.CP_K_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_STR_QK)
				}
				break
			case logic2.CP_J_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_STR_10J)
				}
				break
			case logic2.CP_10_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_STR_910)
				}
				break
			default:
				break
			}
		}
	} else if t.FirstFinished == false && re.Cardtype == static.TYPE_THREESTR && re.Len > 5 && iNumOfKing == 0 {
		//首次打出飞机
		t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_FIRST_PLANE)
	}
}

func (t *Task) IsTaskFinishedOfLastHand(gamecommon Common, re static.TCardType, iSeat uint16, iNumOfKing int) {
	if re.Cardtype == static.TYPE_BOMB_510K && iNumOfKing == 0 {
		switch int(re.Color) {
		case logic2.CC_SPADE_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_HEI_510K)
			}
			break
		case logic2.CC_HEART_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_HONG_510K)
			}
			break
		case logic2.CC_CLUB_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_MEI_510K)
			}
			break
		default:
			break
		}
	} else if re.Cardtype == static.TYPE_THREE && iNumOfKing == 0 {
		switch int(t.GetCardPoint(re.Card)) {
		case logic2.CP_5_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_THREE_5)
			}
			break
		case logic2.CP_10_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_THREE_10)
			}
			break
		case logic2.CP_K_S:
			{
				t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_THREE_K)
			}
			break
		default:
			break
		}
	} else if re.Cardtype == static.TYPE_THREESTR && re.Len >= 6 && iNumOfKing == 0 {
		//打出飞机
		t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_PLANE)
	} else if re.Cardtype == static.TYPE_TWOSTR && iNumOfKing == 0 {
		//打出连对
		if re.Len == 8 {
			t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_4_STR)
		} else if re.Len == 6 {
			switch int(t.GetCardPoint(re.Card)) {
			case logic2.CP_8_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_STR_678)
				}
				break
			case logic2.CP_9_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_STR_789)
				}
				break
			default:
				break
			}
		} else if re.Len == 4 {
			switch int(t.GetCardPoint(re.Card)) {
			case logic2.CP_A_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_STR_KA)
				}
				break
			case logic2.CP_K_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_STR_QK)
				}
				break
			case logic2.CP_J_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_STR_10J)
				}
				break
			case logic2.CP_10_S:
				{
					t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_STR_910)
				}
				break
			default:
				break
			}
		}
	} else if re.Cardtype == static.TYPE_ONE && int(t.GetCardPoint(re.Card)) == logic2.CP_3_S && iNumOfKing == 0 {
		t.TaskFinish(gamecommon, iSeat, consts.TASK_KIND_BEAN_LAST_3)
	}
}

//完成了一个任务
func (t *Task) TaskFinish(gamecommon Common, iSeat uint16, id int) {
	if iSeat < meta2.MAX_PLAYER && t.RandomTask >= 0 && t.RandomTask < len(t.VecAllTaskID) && t.VecAllTaskID[t.RandomTask] == id {
		t.VecFinishedTask[iSeat] = append(t.VecFinishedTask[iSeat], id)
		t.SendFinishTask(gamecommon, iSeat, id)
		t.FirstFinished = true
	}
}

//发送完成了的任务
func (t *Task) SendFinishTask(gamecommon Common, iSeat uint16, id int) {
	if _, ok := t.MapAllTask[id]; ok {
		//发送给客户端
		var msgFinishedTask static.Msg_S_DG_FinishedTaskID
		msgFinishedTask.TaskID = id
		msgFinishedTask.Player = iSeat
		msgFinishedTask.Count = 1
		msgFinishedTask.TotalTaskAward = t.MapAllTask[id] * msgFinishedTask.Count

		gamecommon.SendTableMsg(consts.FinishedTaskID, msgFinishedTask)

		LogStr := fmt.Sprintf("UserId= %d ,完成了任务 TaskID = %d,获得金豆数量 TotalAward = %d", iSeat, id, msgFinishedTask.TotalTaskAward)
		gamecommon.OnWriteGameRecord(iSeat, LogStr)
	}
}

//获取奖励数
func (t *Task) GetTaskAward(iSeat uint16) int {
	if iSeat >= meta2.MAX_PLAYER {
		return 0
	}
	iTotal := 0
	for i := 0; i < len(t.VecFinishedTask[iSeat]); i++ {
		id := t.VecFinishedTask[iSeat][i]
		if _, ok := t.MapAllTask[id]; ok {
			iTotal += t.MapAllTask[id]
		}
	}

	return iTotal
}

//获取奖励详细.必须小于255
func (t *Task) GetTaskAwardDetails(iSeat int) string {
	if iSeat >= meta2.MAX_PLAYER {
		return ""
	}
	var details string
	for i := 0; i < len(t.VecFinishedTask[iSeat]); i++ {
		id := string("")
		fmt.Sprintf(id, "AwardCount=%d|", t.VecFinishedTask[iSeat][i])

		if len(details)+len(id) > 255 {
			return details
		}
		details += id
	}
	return details
}

func (t *Task) clean() {
	t.MapAllTask = map[int]int{}
	t.VecAllTaskID = []int{}
	t.VecFinishedTask = [meta2.MAX_PLAYER][]int{}
	t.RandomTask = -1       //随机任务,下标
	t.FirstFinished = false //是否首次完成任务
}

//清理
func (t *Task) clear() {
	t.MapAllTask = map[int]int{}
	t.VecAllTaskID = []int{}
	t.VecFinishedTask = [meta2.MAX_PLAYER][]int{}
	t.RandomTask = -1       //随机任务,下标
	t.FirstFinished = false //是否首次完成任务
}

//基础函数，通过牌索引获取点数
func (t *Task) GetCardPoint(byCard byte) byte {
	//处理获取普通牌点数
	if logic2.CARDINDEX_NULL < byCard && byCard < logic2.CARDINDEX_SMALL {
		return ((byCard - 1) % 13) + 1
	} else if byCard == logic2.CARDINDEX_SMALL { //处理王牌的点数
		return byte(logic2.CP_BJ_S) //小王
	} else if byCard == logic2.CARDINDEX_BIG {
		return byte(logic2.CP_RJ_S) //大王
	}
	return 0
}
