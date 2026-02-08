package center

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/jinzhu/gorm"
)

type HouseMemberItemWrapper struct {
	Item []HouseMember
	By   func(p, q *HouseMember) bool
}

func (pw HouseMemberItemWrapper) Len() int { // 重写 Len() 方法
	return len(pw.Item)
}

func (pw HouseMemberItemWrapper) Swap(i, j int) { // 重写 Swap() 方法
	pw.Item[i], pw.Item[j] = pw.Item[j], pw.Item[i]
}

func (pw HouseMemberItemWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return pw.By(&pw.Item[i], &pw.Item[j])
}

// ! 包厢成员
type HouseMember struct {
	static.HouseMember
}

func (hm *HouseMember) Logger() logrus.FieldLogger {
	return xlog.ELogger(ElkHouseTbl).WithFields(map[string]interface{}{
		"hid": hm.HId,
		"fid": hm.FId,
		"mid": hm.UId,
	})
}

func (hm *HouseMember) Upper(role int) bool {
	if hm.URole < role {
		return true
	}
	return false
}

func (hm *HouseMember) Lower(role int) bool {
	if hm.URole > role {
		return true
	}
	return false
}

func (hm *HouseMember) SetRemark(dhid int64, remark string) {
	hm.URemark = remark
	GetDBMgr().HouseMemberUpdate(dhid, hm)
}

func (hm *HouseMember) VitaminChangeNoNtf(opUid int64, value int64, utype models.VitaminChangeType, tx *gorm.DB) (int64, int64, error) {
	befVitamin := hm.UVitamin
	hm.UVitamin += value
	aftVitamin := hm.UVitamin
	tx = GetDBMgr().GetDBmControl().Begin()
	err := models.AddVitaminLog(hm.DHId, opUid, hm.UId, befVitamin, aftVitamin, utype, "", tx)
	if err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return 0, 0, err
	}
	hm.Flush()
	hm.OnMemVitaminOffset()
	//通知目标用户
	if opUid != hm.UId {
		ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
		ntf.HId = hm.HId
		ntf.OptId = opUid
		ntf.OptRole = 2
		ntf.UId = hm.UId
		ntf.Value = static.SwitchVitaminToF64(hm.UVitamin)
		hp := GetPlayerMgr().GetPlayer(hm.UId)
		if hp != nil {
			hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
		}
	}

	hm.SendMsg(consts.MsgTypeHouseVitaminChange_Ntf, &static.Msg_HC_HouseVitaminChange_Ntf{UVitamin: static.SwitchVitaminToF64(hm.UVitamin)})

	return befVitamin, aftVitamin, nil
}

func (hm *HouseMember) AddPoolLog(opUid int64, value int64, tx *gorm.DB) error {
	// mem
	sql := `insert into house_member_vitamin_log(dhid,optuid,uid,aftvitamin,value,type) values(?,?,?,?,?,?)`

	err := tx.Exec(sql, hm.DHId, opUid, hm.UId, hm.UVitamin, value, models.BackPool).Error
	return err
}

// VitaminIncrement 外部传入事务时写入redis应该在提交成功之后
func (hm *HouseMember) VitaminIncrement(opUid int64, value int64, utype models.VitaminChangeType, tx *gorm.DB) (int64, int64, error) {
	// mem
	befVitamin := hm.UVitamin
	hm.UVitamin += value
	aftVitamin := hm.UVitamin
	var flush bool
	if tx == nil {
		flush = true
		tx = GetDBMgr().GetDBmControl().Begin()
		err := models.AddVitaminLog(hm.DHId, opUid, hm.UId, befVitamin, aftVitamin, utype, "", tx)
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}

		err = tx.Commit().Error
		if err != nil {
			tx.Rollback()
			return 0, 0, err
		}
	} else {
		err := models.AddVitaminLog(hm.DHId, opUid, hm.UId, befVitamin, aftVitamin, utype, "", tx)
		if err != nil {
			return 0, 0, err
		}
	}
	// redis
	if flush {
		hm.Flush()
		hm.OnMemVitaminOffset()
	}

	house := GetClubMgr().GetClubHouseById(hm.DHId)
	if house == nil {
		return befVitamin, aftVitamin, nil
	}
	ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
	ntf.HId = house.DBClub.HId
	ntf.OptId = opUid
	ntf.OptRole = 2
	ntf.UId = hm.UId
	ntf.Value = static.SwitchVitaminToF64(hm.UVitamin)
	//通知操作用户
	hp := GetPlayerMgr().GetPlayer(opUid)
	if hp != nil {
		hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
	}

	hm.SendMsg(consts.MsgTypeHouseVitaminChange_Ntf, &static.Msg_HC_HouseVitaminChange_Ntf{UVitamin: static.SwitchVitaminToF64(hm.UVitamin)})

	return befVitamin, aftVitamin, nil
}

func (hm *HouseMember) Flush() {
	cli := GetDBMgr().Redis
	buf, err := json.Marshal(hm)
	if err != nil {
		return
	}
	cli.HSet(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, hm.DHId), fmt.Sprintf("%d", hm.UId), fmt.Sprintf("%s", buf))
}

func (hm *HouseMember) CheckRef() (int /*ref hid*/, bool /*是否为老盟主*/) {
	if hm.Ref > 0 {
		house := GetClubMgr().GetClubHouseById(hm.Ref)
		if house == nil {
			xlog.Logger().Error("CheckRef old house not exist")
			return 0, true
		}
		return house.DBClub.HId, hm.UId == house.DBClub.UId
	}
	return 0, false
}

// 忽略邀请的key
func (hm *HouseMember) IgnoreInviteKey() string {
	return fmt.Sprintf("ignoreinvite_%d_%d", hm.HId, hm.UId)
}

// 今日忽略包厢牌桌邀请
func (hm *HouseMember) IgnoreInvite() error {
	cli := GetDBMgr().GetDBrControl()
	err := cli.Set(hm.IgnoreInviteKey(), []byte(time.Now().Format(static.TIMEFORMAT)))
	if err != nil {
		return err
	}
	// return cli.Expire(hm.IgnoreInviteKey(), 60)
	return cli.Expire(hm.IgnoreInviteKey(), static.HF_GetTodayRemainSecond())
}

// 取消今日忽略包厢牌桌邀请
func (hm *HouseMember) UnIgnoreInvite() (bool, error) {
	return GetDBMgr().GetDBrControl().Remove(hm.IgnoreInviteKey())
}

// 是否忽略了邀请信息
func (hm *HouseMember) IsIgnoreInvite() bool {
	return GetDBMgr().GetDBrControl().Exists(hm.IgnoreInviteKey())
}

// 能否被邀请
func (hm *HouseMember) CanInvite(dhid int64) bool {
	person := GetPlayerMgr().GetPlayer(hm.UId)
	if person == nil {
		xlog.Logger().Debug("CanInvite: person not online.", hm.UId)
		return false
	}
	if person.Info.GameId > 0 || person.Info.TableId > 0 {
		xlog.Logger().Debug("CanInvite: person gaming.", hm.UId)
		return false
	}

	if hm.IsIgnoreInvite() {
		xlog.Logger().Debug("CanInvite: person ignore invite")
		return false
	}

	return true
}

// 发送消息
func (hm *HouseMember) SendMsg(header string, v interface{}) {
	if person := GetPlayerMgr().GetPlayer(hm.UId); person != nil {
		person.SendMsg(header, v)
	}
}

func (hm *HouseMember) OnMemVitaminOffset() {
	// 暂时先写在这里
	if p, err := GetDBMgr().GetDBrControl().GetPerson(hm.UId); err == nil && p != nil {
		if p.TableId > 0 {
			if table := GetTableMgr().GetTable(p.TableId); table != nil {
				if floor := GetClubMgr().GetHouseFloorById(hm.DHId, table.FId); floor != nil {
					go floor.RedisPub(consts.MsgTypeHouseVitaminSet_Ntf, &static.Msg_UserTableId{Uid: hm.UId, TableId: p.TableId})
				}
			}
		}
	}
}

type HMVitamin struct {
	Vitamin int64
}

func (hm *HouseMember) GetVitaminFromDb() (int64, error) {
	sql := `select vitamin from house_member_vitamin where uid = ? and hid = ? `
	dest := HMVitamin{}
	err := GetDBMgr().GetDBmControl().Raw(sql, hm.UId, hm.DHId).Scan(&dest).Error
	if err != nil {
		if err.Error() == "record not found" {
			return 0, nil
		}
		return 0, err
	}
	return dest.Vitamin, nil
}

// // 记录是否为每天第一次进圈
// func (hm *HouseMember) todayHouseJoinKey() string {
// 	return fmt.Sprintf("housememberjointoday:%d:%d", hm.DHId, hm.UId)
// }
//
// // 成员今天第一次进入包厢标记
// func (hm *HouseMember) RecordFirstJoinToday() error {
// 	return GetDBMgr().GetDBrControl().RedisV2.Set(
// 		hm.todayHouseJoinKey(),
// 		time.Now().String(),
// 		time.Duration(public.HF_GetTodayRemainSecond())*time.Second).Err()
// }

func (hm *HouseMember) GetVitaminFromDbWithLock(tx *gorm.DB) (int64, error) {
	if tx == nil {
		return 0, errors.New("tx nil")
	}
	sql := `select vitamin from house_member_vitamin where uid = ? and hid = ? for update `
	dest := HMVitamin{}
	err := tx.Raw(sql, hm.UId, hm.DHId).Scan(&dest).Error
	if err != nil {
		if err.Error() == "record not found" {
			return hm.UVitamin, nil
		}
		return 0, err
	}
	return dest.Vitamin, nil
}

// 是否为比赛分管理员
func (hm *HouseMember) IsVitaminAdmin() bool {
	return hm.VitaminAdmin
}

// 是否为队长
func (hm *HouseMember) IsPartner() bool {
	return hm.Partner == 1
}

// 是否为队长
func (hm *HouseMember) GetDirectPartner() int64 {
	if hm.IsPartner() {
		return hm.UId
	}
	return hm.Partner
}

// 是否为副队长
func (hm *HouseMember) IsVicePartner() bool {
	return hm.VicePartner
}

// 是否为下级队长
func (hm *HouseMember) IsJunior() bool {
	return hm.Partner == 1 && hm.Superior > 0
}

// 目标用户是否为自身的上级队长
func (hm *HouseMember) IsHaveTgtSuperior(tgtUId int64) (bool, int) {
	house := GetClubMgr().GetClubHouseById(hm.DHId)
	if house == nil || hm.Partner == 0 || tgtUId == 0 {
		return false, 0
	}
	// 查找队长级别
	bFind := false
	nlv := 0
	// 查找起点
	UId := int64(0)
	if hm.Partner == 1 && hm.Superior > 0 {
		UId = hm.Superior
		nlv = 2
	} else {
		UId = hm.Partner
		nlv = 1
	}
	// 避免写死循环 10次循环够用就可以了
	for i := 0; i < 10; i++ {
		if UId == tgtUId {
			bFind = true
			break
		}
		prePartner := house.GetMemByUId(UId)
		if prePartner == nil {
			break
		}
		if prePartner.Partner == 1 && prePartner.Superior > 0 {
			UId = prePartner.Superior
			nlv++
		} else {
			break
		}
	}

	return bFind, nlv
}

// 得到一级队长ID
func (hm *HouseMember) GetPartner1stUId() int64 {
	house := GetClubMgr().GetClubHouseById(hm.DHId)
	if house == nil || hm.Partner == 0 {
		return 0
	}
	UId := int64(0)
	if hm.Partner == 1 {
		UId = hm.Superior
	} else {
		UId = hm.Partner
	}
	// 避免写死循环 10次循环够用就可以了
	for i := 0; i < 10; i++ {
		prePartner := house.GetMemByUId(UId)
		if prePartner == nil {
			break
		}
		if prePartner.Superior > 0 {
			UId = prePartner.Superior
		} else {
			break
		}
	}
	return UId
}

// 得到一级队长ID
func (hm *HouseMember) ClearVitaminOnKick(opt int64, typ models.VitaminChangeType, clb *Club, tx *gorm.DB) (int64, *xerrors.XError) {
	if hm.UVitamin != 0 {
		//cli := GetDBMgr().Redis
		//hm.Lock(cli)
		//defer hm.Unlock(cli)
		_, after, err := hm.VitaminIncrement(opt, -1*hm.UVitamin, typ, tx)
		if err != nil {
			xlog.Logger().Error(err)
			return 0, xerrors.DBExecError
		}
		//修改疲劳值统计管理节点信息
		err = GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, hm.UId, after, tx)
		if err != nil {
			xlog.Logger().Error(err)
			return 0, xerrors.DBExecError
		}
	}
	return hm.UVitamin, nil
}

//// 是否为包厢盟主
//func (hm *HouseMember) IsCreator() bool {
//	return hm.URole == constant.ROLE_CREATER
//}
//
//// 是否为包厢管理员
//func (hm *HouseMember) IsAdmin() bool {
//	return hm.URole == constant.ROLE_ADMIN
//}
//
//// 是否为包厢成员
//func (hm *HouseMember) IsMember() bool {
//	return hm.URole == constant.ROLE_MEMBER
//}
//
//// 是否为包厢黑名单成员
//func (hm *HouseMember) IsBlack() bool {
//	return hm.URole == constant.ROLE_BLACK
//}
//
//// 是否为包厢申请中成员
//func (hm *HouseMember) IsApplying() bool {
//	return hm.URole == constant.ROLE_APLLY
//}
//
//// 权限认证
//func (hm *HouseMember) IsAccessible() bool {
//	return false
//}

//type HouseMemberMap map[int64]*HouseMember
//
//func (hmm *HouseMemberMap) SuperiorsByJuniors(juniors ...int64) []int64 {
//	in := func(jid int64) bool {
//		for i := 0; i < len(juniors); i++ {
//			if jid == juniors[i] {
//				return true
//			}
//		}
//		return false
//	}
//	res := make([]int64, 0)
//	for id, mem := range *hmm {
//		if in(id) && mem.Superior > 0 {
//			res = append(res, mem.Superior)
//		}
//	}
//	return res
//}
//
//func (hmm *HouseMemberMap) JuniorsBySuperiors(superiors ...int64) []int64 {
//	in := func(sid int64) bool {
//		for i := 0; i < len(superiors); i++ {
//			if sid == superiors[i] {
//				return true
//			}
//		}
//		return false
//	}
//	res := make([]int64, 0)
//	for id, mem := range *hmm {
//		if mem.IsPartner() && in(mem.Superior) {
//			res = append(res, id)
//		}
//	}
//	return res
//}
