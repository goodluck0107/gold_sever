package center

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

// ! 玩家的数据结构
type PlayerCenterMemory struct {
	Info    static.Person
	session *Session // session
	Ip      string   `json:"ip"` // 客户端地址
}

func (pcm *PlayerCenterMemory) OnLine(on bool) {
	// 用户上线
	if on {
		// 检查任务状态
		pcm.CheckTasks()
		// 推送跑马灯
		pcm.SendMarqueeNotice()
		// 发送未读消息
		pcm.SendUnreadNotify()
		// 发出待办事项
		go pcm.SendBacklog()
	} else {
		// 用户离线
	}
}

func (pcm *PlayerCenterMemory) SendBuf(buf []byte) {
	if pcm.session == nil {
		return
	}
	pcm.session.SendBuf(buf)
}

func (pcm *PlayerCenterMemory) SendMsg(head string, v interface{}) {
	if pcm.session == nil {
		return
	}
	pcm.session.SendMsg(head, xerrors.SuccessCode, v, pcm.Info.Uid)
}

func (pcm *PlayerCenterMemory) CloseSession() {
	if pcm.session == nil {
		return
	}
	pcm.session.SafeClose(consts.SESSION_CLOED_FORCE)
}

func (pcm *PlayerCenterMemory) SendNullMsg() {
	var msg static.Msg_Null
	pcm.SendMsg("nothing", &msg)
}

func (pcm *PlayerCenterMemory) SaveData() {
	GetDBMgr().db_R.AddPerson(&pcm.Info)
}

func (pcm *PlayerCenterMemory) UpdCard(typ int8) {
	var msg static.Msg_S2C_UpdCard
	msg.Card = pcm.Info.Card
	msg.Type = typ
	pcm.SendMsg("updcard", &msg)
}

func (pcm *PlayerCenterMemory) UpdGold(typ int8, offSet int) {
	var msg static.Msg_S2C_UpdGold
	msg.Gold = pcm.Info.Gold
	msg.Type = typ
	msg.Offset = offSet
	pcm.SendMsg("updgold", &msg)
}

func (pcm *PlayerCenterMemory) UpdCoupon(typ int8) {
	var msg static.Msg_S2C_UpdGoldBean
	msg.GoldBean = pcm.Info.GoldBean
	msg.Type = typ
	pcm.SendMsg("updgoldbean", &msg)
}

func (pcm *PlayerCenterMemory) UpdDiamond(typ int8) {
	var msg static.Msg_S2C_UpdDiamond
	msg.Diamond = pcm.Info.Diamond
	msg.Type = typ
	pcm.SendMsg("upddiamond", &msg)
}

func (pcm *PlayerCenterMemory) SendTickOff() {
	var msg static.Msg_Null
	pcm.SendMsg("tickoff", &msg)
}

func (pcm *PlayerCenterMemory) UpdTel(tel string) {
	if pcm.Info.Tel == tel {
		return
	}
	pcm.Info.Tel = tel
}

// 推送未读消息
func (pcm *PlayerCenterMemory) SendUnreadNotify() {
	if pcm.Info.TableId > 0 {
		return
	}
	var paymentOrder models.PaymentOrder
	if GetDBMgr().GetDBrControl().RedisV2.Get(fmt.Sprintf("offline:recharge_suc:%d", pcm.Info.Uid)).Scan(&paymentOrder) == nil && paymentOrder.ID > 0 {
		defer GetDBMgr().GetDBrControl().RedisV2.Del(fmt.Sprintf("offline:recharge_suc:%d", pcm.Info.Uid))
		pcm.SendMsg(consts.MsgTypePaymentResultNtf, &paymentOrder)
	}

	ntfs, err := GetUnreadNotify(pcm.Info.Uid)
	if err != nil {
		xlog.Logger().Errorln("SendUnreadNotify error", err)
		return
	}
	defer DelUnreadNotify(pcm.Info.Uid)
	for _, ntf := range ntfs {
		var msg static.Msg_HC_NotifyPush
		msg.Msg = ntf
		pcm.SendMsg(consts.MsgTypeHallNotifyPush, &msg)
	}
}

func (pcm *PlayerCenterMemory) SendMarqueeNotice() {
	marqueeNotices := GetNoticeMarqueeNotice()
	if len(marqueeNotices) > 0 {
		pcm.SendMsg(consts.MsgTypeMarqueenNotice, marqueeNotices)
	}
}

// 发送用户待办事项
func (pcm *PlayerCenterMemory) SendBacklog() {
	if pcm.Info.TableId > 0 {
		return
	}
	// 发送包厢未处理的合并包厢申请
	pcm.sendHouseMergeBacklog()
	// 发送未处理的包厢邀请
	pcm.sendHouseJoinInvite()
}

func (pcm *PlayerCenterMemory) sendHouseMergeBacklog() {
	houseIds, err := GetDBMgr().GetDBrControl().ListHouseMemberCreate(pcm.Info.Uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return
	}

	xlog.Logger().Debug("house ids = ", houseIds)
	if len(houseIds) == 0 {
		return
	}

	// 合并包厢消息
	houseMergeLogs := make([]*models.HouseMergeLog, 0)
	err = GetDBMgr().GetDBmControl().Where("devourer in(?)", houseIds).Where("merge_state = ?", models.HouseMergeStateWaiting).Order("updated_at desc").Find(&houseMergeLogs).Error
	if err != nil {
		xlog.Logger().Errorf("get houseMergeLogs owned %d error %v", pcm.Info.Uid, err)
		return
	}
	for _, hml := range houseMergeLogs {
		if hml == nil {
			continue
		}
		house := GetClubMgr().GetClubHouseById(hml.Devourer)
		if house == nil {
			xlog.Logger().Debug("Devourer house is nil", hml.Devourer)
			continue
		}
		thouse := GetClubMgr().GetClubHouseById(hml.Swallowed)
		if house == nil {
			xlog.Logger().Debug("Swallowed house is nil", hml.Swallowed)
			continue
		}
		pcm.SendMsg(consts.MsgTypeHouseMerge_NTF, &static.MsgHouseMergeRequest{HId: house.DBClub.HId, MergeHId: thouse.DBClub.HId})
		return
	}

	// 撤销合并包厢消息
	houseRevokeLogs := make([]*models.HouseMergeLog, 0)
	err = GetDBMgr().GetDBmControl().
		Where("(devourer in(?) or swallowed in(?)) and merge_state = ?", houseIds, houseIds, models.HouseMergeStateRevoking).
		Order("updated_at desc").
		Find(&houseRevokeLogs).Error
	if err != nil {
		xlog.Logger().Errorf("get houseRevokeLogs owned %d error %v", pcm.Info.Uid, err)
		return
	}
	for _, hml := range houseRevokeLogs {
		if hml == nil {
			continue
		}
		house := GetClubMgr().GetClubHouseById(hml.Devourer)
		if house == nil {
			xlog.Logger().Debug("Devourer house is nil", hml.Devourer)
			continue
		}
		thouse := GetClubMgr().GetClubHouseById(hml.Swallowed)
		if house == nil {
			xlog.Logger().Debug("Swallowed house is nil", hml.Swallowed)
			continue
		}
		if house.DBClub.UId == pcm.Info.Uid && thouse.DBClub.Id == hml.Sponsor {
			pcm.SendMsg(consts.MsgTypeHouseRevoke_NTF, &static.MsgHouseMergeRequest{HId: house.DBClub.HId, MergeHId: thouse.DBClub.HId})
			return
		}

		if thouse.DBClub.UId == pcm.Info.Uid && house.DBClub.Id == hml.Sponsor {
			pcm.SendMsg(consts.MsgTypeHouseRevoke_NTF, &static.MsgHouseMergeRequest{HId: thouse.DBClub.HId, MergeHId: house.DBClub.HId})
			return
		}
	}
}

func (pcm *PlayerCenterMemory) sendHouseJoinInvite() {
	v, err := GetUnreadHouseJoinInvite(pcm.Info.Uid)
	if err != nil {
		if err != redis.Nil {
			xlog.Logger().Error("sendHouseJoinInvite", err)
		}
		return
	}
	pcm.SendMsg(consts.MsgTypeHouseJoinInviteRecv, v)
	DelUnreadHouseJoinInvite(pcm.Info.Uid)
}

// 检查用户的任务
func (pcm *PlayerCenterMemory) CheckTasks() {
	// 检查个人任务是否完整
	GetTasksMgr().CheckUserTasks(pcm.Info.Uid)
	// 见面礼
	//pcm.sendWelcome()
	// 签到
	//if pcm.sendCheckin() { // 如果今日已签到，则检查低保，如果还未签到，则稍后检查
	// 检查低保
	// pcm.checkLowIncome()
	//}
}

/*
// 发送签到
func (ph *PlayerCenterMemory) sendCheckin() bool {
	// 初始化签到任务
	step, isCheckin := GetTasksMgr().ReInitTaskWeekly(ph.Info.Uid)
	if !isCheckin {
		// 获取签到任务配置
		taskCon := GetTasksMgr().GetTaskCfgByType(constant.TASK_TYPE_NORMAL, constant.TASK_KIND_SIGN)
		if taskCon != nil {
			// 当日未签到, 返回签到任务信息
			res := new(public.Msg_S2C_TaskCheckin)
			res.Step = step
			rewards := taskCon.V.(*TaskWeekly).ReWards
			var buf bytes.Buffer
			for i := 0; i < len(rewards); i++ {
				var award public.SignInAward
				for j := 0; j < len(rewards[i]); j++ {
					rew := rewards[i][j]
					if j > 0 {
						buf.WriteString("+")
					}
					if award.AwardUrl == "" {
						award.AwardUrl = rew.Url
					}
					buf.WriteString(fmt.Sprintf("%sx%d", constant.WealthTypeString(rew.WealthType), rew.Num))
				}
				award.AwardName = buf.String()
				buf.Reset()
				res.ReWards = append(res.ReWards, &award)
			}
			ph.SendMsg(constant.MsgTypeTaskCheckin, res)
		}
	}
	return isCheckin
}

// 发送见面礼
func (ph *PlayerCenterMemory) sendWelcome() {
	taskConn := GetTasksMgr().GetTaskCfgByType(constant.TASK_TYPE_NORMAL, constant.TASK_KIND_WELCOMEGIFT)
	if taskConn != nil {
		// 判断用户是否符合日期规范
		tm, _ := time.ParseInLocation("2006-01-02 15:04:05", taskConn.V.(*WelcomeGift).Time, time.Local)
		// 在此时间之前的都认为是老用户
		if ph.Info.CreateTime < tm.Unix() {
			// 判断用户有没有领取过见面礼
			task := GetTasksMgr().GetUserTask(ph.Info.Uid, taskConn.Id)
			if task != nil {
				if task.Step == 0 {
					// 没领取过奖励则推送见面礼消息
					gift := new(public.Msg_S2C_WelcomeGift)
					gift.RegisteredAt = ph.Info.CreateTime
					gift.PlayCount = ph.Info.TotalCount
					gift.Rewards = taskConn.V.(*WelcomeGift).ReWards
					ph.SendMsg(constant.MsgTypeWelcomeGift, gift)
				}
			}
		}
	}
}

// 发送签到
func (ph *PlayerCenterMemory) checkLowIncome() bool {
	var err error
	// 用户低保自动领取逻辑
	if ph.Info.Gold+ph.Info.InsureGold < GetServer().ConServers.Allowances.LimitGold {
		// 判断用户领取低保次数
		var count int
		if err = GetDBMgr().GetDBmControl().Table(model.UserAllowances{}.TableName()).Where("uid = ? and date = ?", ph.Info.Uid, time.Now().Format("2006-01-02")).Count(&count).Error; err != nil {
			syslog.Logger().Error(err)
		} else {
			if count < GetServer().ConServers.Allowances.Num {
				err = disposeUserAllowances(&ph.Info)
				if err != nil {
					syslog.Logger().Error(err)
				} else {
					// 推送通知(客户端自己去操作金币加减)
					res := new(public.Msg_S2C_Allowances)
					res.Current = count + 1
					res.Gold = GetServer().ConServers.Allowances.AwardGold
					res.Remain = GetServer().ConServers.Allowances.Num - count - 1
					ph.SendMsg(constant.MsgTypeDisposeAllowances, res)
					ph.UpdGold(model.CostTypeAllowances, res.Gold)
				}
				return true
			}
		}
	}
	return false
}
*/
