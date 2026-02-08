package center

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"strconv"
	"sync"
)

/*
	//角色 : 1盟主 2裁判  3管理员  4队长  5副队长  6普通成员
*/

const (
	MinorRightNull = "" //"",

)

const (
	management_user        = 1 //"成员管理",
	management_record      = 2 //"战绩管理"
	management_statistical = 3 //"统计管理"
	management_room        = 4 //"房间管理"
	management_teahouse    = 5 //"包厢管理"
)

// 成员管理
const (
	MinorJoinReviewed   = "join_reviewed"   //"加入审核"
	MinorOutReviewed    = "out_reviewed"    //"退出审核"
	MinorOutPlayer      = "out_player"      //"踢出玩家"
	MinorBanPlay        = "ban_play"        //"禁止娱乐"
	MinorAllowPlay      = "allow_play"      //"恢复娱乐"
	MinorMovBlacklist   = "mov_blacklist"   //"移入黑名单"
	MinorSetAdmin       = "set_admin"       //"设置管理员"
	MinorSetJudge       = "set_judge"       //"设置裁判"
	MinorSetCaptain     = "set_captain"     //"设置队长"
	MinorSetDeputy      = "set_deputy"      //"设置副队长"
	MinorMtAdd          = "mt_add"          //"手动添加"
	MinorManageSuperior = "manage_superior" //"管理队长2"
)

// 战绩管理
const (
	MinorMyRecord        = "my_record"        //"我的战绩"
	MinorUserRecord      = "user_record"      //"成员战绩"
	MinorTeaRecord       = "tea_record"       //"圈子战绩"
	MinorTeamStatistical = "team_statistical" //"团队统计"
)

// 统计管理
const (
	MinorPvPStatistical   = "pvp_statistical"   //"对局统计"
	MinorWinerStatistical = "winer_statistical" //"大赢家统计"
	MinorFloorStatistical = "floor_statistical" //"楼层统计"
)

// 房间管理
const (
	MinorDisRoom     = "dis_room"     //"解散房间"
	MinorRoomKickout = "room_kickout" //"房间踢人"
)

// 包厢管理
const (
	MinorSetTeaName      = "set_tea_name"      //"圈名称设置"
	MinorSetNotice       = "set_notice"        //"公告设置"
	MinorSetBackground   = "set_background"    //"背景设置"
	MinorSetTableColor   = "set_table_color"   //"桌布颜色设置"
	MinorSetPrivacy      = "set_privacy"       //"隐私设置"
	MinorSetTableNum     = "set_table_num"     //"牌桌数设置"
	MinorSetLowScore     = "set_low_score"     //"低分局设置"
	MinorSetJoinTable    = "set_join_table"    //"入桌设置"
	MinorSetDistance     = "set_distance"      //"距离设置"
	MinorSetJoinTea      = "set_join_tea"      //"加入设置"  客户端设置面板用来显示 无接口   功能接口 MinorJoinReviewed
	MinorSetOutTea       = "set_out_tea"       //"退出设置"  客户端设置面板用来显示 无接口   功能接口 MinorOutReviewed
	MinorSetBlacklistTea = "set_blacklist_tea" //"黑名单设置" 客户端设置面板用来显示 无接口  功能接口 MinorMovBlacklist
	MinorSetCaptainRight = "set_captain_right" //"队长权限设置" 客户端设置面板用来显示 并且有接口
	MinorBanWxTea        = "ban_wx_tea"        //"禁用微信"   客户端设置面板用来显示 无接口
	MinorBanTea          = "ban_tea"           //"圈冻结"
	MinorSetRoll         = "set_roll"          //"混排设置"
	MinorBanTableSitAt   = "ban_table_sit_at"  //"禁止同桌"
	MinorMergeTea        = "merge_tea"         //"合并包厢"
	MinorSetVipFloor     = "set_vip_floor"     //"VIP楼层设置"
	MinorMyTea           = "my_tea"            //"我的包厢"    客户端设置面板用来显示 无接口
	MinorSetCardRemind   = "set_card_remind"   //"设置房卡提醒"   // 接口暂无
)

const (
	RightNull         = 0 // 无权限
	RightGuDin        = 1 // 固定权
	RightKePeiZhiNull = 2 // 可配置并且默认无
	RightKePeiZhiHave = 3 // 可配置并且默认有
	RightNoPeiZhiNull = 4 // 禁止配置并且无权限
)
const (
	//角色 : 1盟主 2裁判  3管理员  4队长  5副队长  6普通成员
	RoleNull    = iota //无角色 错误
	RoleCreator        //盟主
	RoleReferee        //裁判
	RoleAdmin          //管理员
	RoleCaptain        //队长
	RoleDeputy         //副队长
	RoleMember         //普通成员
)

// 包厢开关功能 与 权限相对应
const (
	BanWeChat       = MinorBanWxTea  //禁用微信开关
	CapSetDep       = MinorRightNull //队长设置其他人为队长的权限
	IsRecShowParent = MinorRightNull //队长设置其他人为队长的权限
)

type (
	memberRight struct {
		right map[string]minorRight
	}

	minorRight struct {
		minorRight map[string]minorRightData
	}

	minorRightData struct {
		minor_id     int
		minor_status int
		minor_key    string
		minor_name   string
	}
)

var (
	houseMemberRightSingleton  *memberRight = nil
	rightOnce                  sync.Once
	rightLock                  sync.RWMutex
	defaultHouseMemberRightMap map[int]string
	houseMemberRightMap        map[string]string
)

func GetMRightMgr() *memberRight {
	rightOnce.Do(func() {
		houseMemberRightSingleton = &memberRight{}
		houseMemberRightSingleton.Init()
	})
	return houseMemberRightSingleton
}

func (mr *memberRight) Init() {
	defaultHouseMemberRightMap = make(map[int]string)
	houseMemberRightMap = make(map[string]string)
}

// 获取用户权限
func (mr *memberRight) GetRight(key string) string {
	rightLock.RLock()
	defer rightLock.RUnlock()
	str, ok := houseMemberRightMap[key]
	if !ok {
		return ""
	}
	return str
}

// 修改用户权限
func (mr *memberRight) SetRight(key string, str string) {
	rightLock.Lock()
	defer rightLock.Unlock()
	houseMemberRightMap[key] = str
}

// 删除用户权限
func (mr *memberRight) DeleteRight(key string) {
	rightLock.Lock()
	defer rightLock.Unlock()
	_, ok := houseMemberRightMap[key]
	if ok {
		delete(houseMemberRightMap, key)
	}
}

// 获取 用户对应包厢的权限   //无权限0 固定权限1  默认权限2  可配权限3
func (mr *memberRight) FindRightByMember(hmem *HouseMember, isAdjust bool) (map[string]interface{}, error) {
	//mr.selectDefaultUserRightByRole(1)
	//**********//
	//查询用户没有权限的时候 初始化一个默认权限
	strRight, err := mr.findRight(hmem)
	if err != nil {
		return nil, err
	}
	if isAdjust {
		//校对全部权限
		strRight, err = mr.adjustRightAll(hmem, strRight)
		if err != nil {
			return nil, err
		}
	}
	uRoleMap, _ := mr.changeUroleData(strRight)
	return uRoleMap, nil
}

// 查询用户包厢权限数据
func (mr *memberRight) findRight(hmem *HouseMember) (string, error) {
	var hmur models.HouseMemberUserRight
	key := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	rightStr := mr.GetRight(key)
	if rightStr == "" {
		if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Where("dhid = ? and uid = ?", hmem.DHId, hmem.UId).First(&hmur).Error; err != nil {
			if err == gorm.ErrRecordNotFound { //这里没有数据存一条默认数据
				urole := mr.getHmUserRole(hmem)
				res, err := mr.selectDefaultUserRightByRole(urole)
				if err != nil {
					return "", err
				}
				hmur.Uid = int(hmem.UId)
				hmur.Hid = hmem.HId
				hmur.Dhid = int(hmem.DHId)
				hmur.Uright = res
				hmur.Role = urole
				if urole == RoleCaptain && hmem.Ref != 0 {
					//合并包厢过来的队长 要给把合并包厢的权限 赋值为 1  固定权限
					var roleMap map[string]interface{}
					roleMap = make(map[string]interface{})
					err := json.Unmarshal([]byte(hmur.Uright), &roleMap)
					if err != nil {
						cuserror := xerrors.NewXError("权限数据转换失败")
						return "", cuserror
					}
					for _, value := range roleMap {
						itemValue := value.(map[string]interface{})
						for _, v := range itemValue {
							itemV := v.(map[string]interface{})
							if itemV["minor_key"].(string) == MinorMergeTea {
								itemV["minor_status"] = RightGuDin
							}
						}
					}
					tempJson, _ := json.Marshal(roleMap)
					hmur.Uright = string(tempJson)
				}
				if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Save(&hmur).Error; err != nil {
					cuserror := xerrors.NewXError("给用户添加默认权限失败")
					return "", cuserror
				}
			} else {
				cuserror := xerrors.NewXError("查询默认权限失败")
				return "", cuserror
			}
		}
		mr.SetRight(key, hmur.Uright)
		rightStr = mr.GetRight(key)
	}
	return rightStr, nil
}

// 更新用户对应包厢的权限
func (mr *memberRight) UpdateRightByMember(hmem *HouseMember, data string, house *Club) (string, error) {
	var updateUrightData map[string]int
	updateUrightData = make(map[string]int)
	err := json.Unmarshal([]byte(data), &updateUrightData)
	if err != nil {
		cuserror := xerrors.NewXError("权限数据转换失败")
		return "", cuserror
	}

	_, err = mr.findRight(hmem)
	if err != nil {
		return "", err
	}
	_, err = mr.changeUpdateUroleData(hmem, updateUrightData, house)
	if err != nil {
		return "", err
	}
	newRightStr, err := mr.UpdateRight(hmem)
	if err != nil {
		return "", err
	}

	// 这里变化了队长的数据  同时要变化副队长 的权限
	if hmem.IsPartner() {
		mems := house.GetMemSimple(false)
		for _, vpMem := range mems {
			if vpMem.IsVicePartner() && vpMem.Partner == hmem.UId {
				_, err = mr.findRight(&vpMem)
				if err != nil {
					return "", err
				}
				_, err = mr.changeUpdateUroleDataLink(&vpMem, updateUrightData, house)
				if err != nil {
					return "", err
				}
				newRightStr2, err := mr.UpdateRight(&vpMem)
				if err != nil {
					return "", err
				}
				if player := GetPlayerMgr().GetPlayer(vpMem.UId); player != nil {
					player.SendMsg(consts.MsgTypeHmUpdateUserRight_NTF, static.Msg_S2C_UpdateHmUserRight{
						Hid:         vpMem.HId,
						Uid:         vpMem.UId,
						UpdateRight: newRightStr2,
					})
				}
			}
		}

	}
	return newRightStr, nil
}

// 修改用户包厢权限数据--数据库
func (mr *memberRight) UpdateRight(hmem *HouseMember) (string, error) {
	key := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	updateData := mr.GetRight(key)
	var hmur models.HouseMemberUserRight
	if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Where("dhid = ? and uid = ?", hmem.DHId, hmem.UId).First(&hmur).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hmur.Uid = int(hmem.UId)
			hmur.Hid = hmem.HId
			hmur.Dhid = int(hmem.DHId)
			hmur.Uright = updateData
			hmur.Role = mr.getHmUserRole(hmem)
			if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Save(&hmur).Error; err != nil {
				cuserror := xerrors.NewXError("给用户添加默认权限失败")
				return "", cuserror
			}
			return hmur.Uright, nil
		} else {
			cuserror := xerrors.NewXError("修改默认权限失败，没有查到玩家权限")
			return "", cuserror
		}
	}
	updateMap := make(map[string]interface{})
	updateMap["uright"] = updateData
	updateMap["role"] = mr.getHmUserRole(hmem)
	if err := GetDBMgr().GetDBmControl().Model(&hmur).Update(updateMap).Error; err != nil {
		//这里没有数据存一条默认数据
		cuserror := xerrors.NewXError("修改默认权限失败")
		return "", cuserror
	}
	return hmur.Uright, nil
}

// 转换更新之后的权限数据可是
func (mr *memberRight) changeUpdateUroleData(hmem *HouseMember, updateData map[string]int, house *Club) (string, error) {
	key := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	rightStr := mr.GetRight(key)
	var roleMap map[string]interface{}
	roleMap = make(map[string]interface{})
	err := json.Unmarshal([]byte(rightStr), &roleMap)
	if err != nil {
		cuserror := xerrors.NewXError("权限数据转换失败")
		return "", cuserror
	}
	isUpdate := false
	for _, value := range roleMap {
		itemValue := value.(map[string]interface{})
		for k, v := range itemValue {
			itemV := v.(map[string]interface{})
			oldStatus := int(itemV["minor_status"].(float64))
			for k1, v1 := range updateData {
				if k1 == k { //&& v1 != oldStatus
					if oldStatus != RightNoPeiZhiNull {
						itemV["minor_status"] = v1
						isUpdate = true
					}
				}
			}
		}
	}
	if !isUpdate {
		cuserror := xerrors.NewXError("没有修改当前数据")
		return "", cuserror
	}
	tempJson, _ := json.Marshal(roleMap)
	mr.SetRight(key, string(tempJson))
	rightStr = mr.GetRight(key)
	return rightStr, nil
}

// 转换更新之后的权限数据可是
func (mr *memberRight) changeUpdateUroleDataLink(hmem *HouseMember, updateData map[string]int, house *Club) (string, error) {
	key := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	rightStr := mr.GetRight(key)
	var roleMap map[string]interface{}
	roleMap = make(map[string]interface{})
	err := json.Unmarshal([]byte(rightStr), &roleMap)
	if err != nil {
		cuserror := xerrors.NewXError("权限数据转换失败")
		return "", cuserror
	}
	for _, value := range roleMap {
		itemValue := value.(map[string]interface{})
		for k, v := range itemValue {
			itemV := v.(map[string]interface{})
			oldStatus := int(itemV["minor_status"].(float64))
			for k1, v1 := range updateData {
				if k1 == k && (oldStatus != RightNull && oldStatus != RightGuDin) {
					if v1 == RightKePeiZhiNull {
						itemV["minor_status"] = RightNoPeiZhiNull
					} else if v1 == RightKePeiZhiHave {
						itemV["minor_status"] = RightKePeiZhiNull
					}
				}
			}
		}
	}
	tempJson, _ := json.Marshal(roleMap)
	mr.SetRight(key, string(tempJson))
	rightStr = mr.GetRight(key)
	return rightStr, nil
}

// 获取默认角色权限表  //! 角色 1盟主 2裁判  3管理员  4队长  5副队长  6普通成员
func (mr *memberRight) selectDefaultUserRightByRole(urole int) (string, error) {
	if urole == 0 {
		cuserror := xerrors.NewXError("权限角色有误")
		return "", cuserror
	}
	_, ok := defaultHouseMemberRightMap[urole]
	if !ok {
		var r []models.HouseMemberRight
		if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberRight{}).Where("urole = ?", urole).Find(&r).Error; err != nil {
			cuserror := xerrors.NewXError("获取默认权限失败")
			return "", cuserror
		}
		var roleMap map[string]map[string]interface{}
		roleMap = make(map[string]map[string]interface{})
		for i := 0; i < len(r); i++ {
			if len(roleMap[r[i].BigKey]) == 0 {
				roleMap[r[i].BigKey] = make(map[string]interface{})
			}
			var minorData map[string]interface{}
			minorData = make(map[string]interface{})
			minorData["minor_id"] = r[i].MinorId
			minorData["minor_key"] = r[i].MinorKey
			minorData["minor_name"] = r[i].MinorName
			minorData["minor_status"] = r[i].MinorStatus
			roleMap[r[i].BigKey][r[i].MinorKey] = minorData
		}
		tempJson, _ := json.Marshal(roleMap)
		defaultHouseMemberRightMap[urole] = string(tempJson)
	}
	return defaultHouseMemberRightMap[urole], nil
}

// 计算出当前角色
func (mr *memberRight) getHmUserRole(hmem *HouseMember) int {
	isFaker := GetDBMgr().GetDBrControl().RedisV2.SIsMember("faker_admin", hmem.UId).Val()
	if isFaker {
		return RoleAdmin
	}
	//urole：盟主 0  管理员1  成员 2  队长  2 && Partner 副队长 2&&VicePartner
	if hmem.URole == 0 {
		return RoleCreator
	} else if hmem.URole == 1 {
		return RoleAdmin
	} else if hmem.URole == 2 {
		if hmem.IsPartner() {
			return RoleCaptain
		} else if hmem.IsVicePartner() {
			return RoleDeputy
		} else if hmem.VitaminAdmin {
			return RoleReferee
		}
		return RoleMember
	}
	return RoleNull
}

// 转换权限数据格式 删除一些数据不需要的数据
func (mr *memberRight) changeUroleData(strUrole string) (map[string]interface{}, error) {
	if strUrole == "" {
		cuserror := xerrors.NewXError("获取权限数据失败")
		return nil, cuserror
	}
	var roleMap map[string]interface{}
	roleMap = make(map[string]interface{})
	err := json.Unmarshal([]byte(strUrole), &roleMap)
	if err != nil {
		cuserror := xerrors.NewXError("权限数据转换失败")
		return nil, cuserror
	}
	return roleMap, nil
}

// 查询玩家对应权限
func (mr *memberRight) CheckRight(hmem *HouseMember, key string) (bool, error) {
	if hmem.URole == consts.ROLE_CREATER {
		return true, nil
	}
	isFaker := GetDBMgr().GetDBrControl().RedisV2.SIsMember("faker_admin", hmem.UId).Val()
	if isFaker {
		return true, nil
	}
	//查询用户没有权限的时候 初始化一个默认权限
	strRight, err := mr.findRight(hmem)
	if err != nil {
		return false, err
	}
	uRoleMap, _ := mr.changeUroleKV(strRight)
	val, ok := uRoleMap[key]
	if !ok {
		//todo 如果这里无这个权限值 要从默认权限中去找 赋值默认
		newStrRight, cer := mr.adjustRight(hmem, key, strRight)
		if cer != nil {
			return false, nil
		}
		newURoleMap, _ := mr.changeUroleKV(newStrRight)
		if newURoleMap[key] == RightGuDin || newURoleMap[key] == RightKePeiZhiHave {
			return true, nil
		}
	} else if val == RightGuDin || val == RightKePeiZhiHave {
		return true, nil
	}

	return false, nil
}

// 转换权限数据 为Key Value
func (mr *memberRight) changeUroleKV(strUrole string) (map[string]int, error) {
	if strUrole == "" {
		cuserror := xerrors.NewXError("获取权限数据失败")
		return nil, cuserror
	}
	var roleMap map[string]interface{}
	roleMap = make(map[string]interface{})
	err := json.Unmarshal([]byte(strUrole), &roleMap)
	if err != nil {
		cuserror := xerrors.NewXError("权限数据转换失败")
		return nil, cuserror
	}
	var newMinorRoleMap map[string]int
	newMinorRoleMap = make(map[string]int)
	for _, value := range roleMap {
		itemValue := value.(map[string]interface{})
		for _, v := range itemValue {
			itemV := v.(map[string]interface{})
			ms := int(itemV["minor_status"].(float64))
			// if ms == RightGuDin || ms == RightKePeiZhiHave { //minor_status 为1 和3 的时候 才有权限
			newMinorRoleMap[itemV["minor_key"].(string)] = ms
			// }
		}
	}
	return newMinorRoleMap, nil
}

// 设置角色的时候 更新权限接口
func (mr *memberRight) setRoleUpdateRight(hmem *HouseMember, islow bool) (bool, error) {
	//这里没有数据存一条默认数据
	urole := mr.getHmUserRole(hmem)
	if islow {
		urole = 6
	}
	res, err := mr.selectDefaultUserRightByRole(urole)
	if err != nil {
		cuserror := xerrors.NewXError("查询默认权限失败")
		return false, cuserror
	}
	key := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	mr.SetRight(key, res)
	_, err = mr.UpdateRight(hmem)
	if err != nil {
		return false, err
	}
	rightStr := mr.GetRight(key)
	if player := GetPlayerMgr().GetPlayer(hmem.UId); player != nil {
		player.SendMsg(consts.MsgTypeHmUpdateUserRight_NTF, static.Msg_S2C_UpdateHmUserRight{
			Hid:         hmem.HId,
			Uid:         hmem.UId,
			UpdateRight: rightStr,
		})
	}
	return true, nil
}

// 检查能否设置别人的权限数据
func (mr *memberRight) checkRoleSetRight(opHm *HouseMember, hm *HouseMember) (int, error) {
	//tody 一种树状关系图   rMap[1][0]  中 1位身份标记  下面的切片中的内容 对应可以配置的 身份的标记  身份标记参考 getHmUserRole    "盟主": 1, "裁判": 2, "管理员": 3, "队长": 4, "副队长": 5, "普通成员": 6
	rMap := make(map[int]map[int]int)
	for i := 1; i <= 6; i++ {
		if len(rMap[i]) == 0 {
			rMap[i] = make(map[int]int)
		}
		for j := 1; j <= 6; j++ {
			if i != j {
				rMap[i][j] = 0
				if i == RoleCreator {
					if j == RoleAdmin || j == RoleCaptain {
						rMap[i][j] = j
					}
				} else if i == RoleCaptain {
					if j == RoleDeputy {
						rMap[i][j] = j
					}
				}
			}
		}
	}

	opRole := mr.getHmUserRole(opHm)
	opbRole := mr.getHmUserRole(hm)

	if rMap[opRole][opbRole] != 0 {
		return opbRole, nil
	}
	cuserror := xerrors.NewXError("您没有权限修改成员权限内容")
	return 0, cuserror
}

// 删除权限的数据
func (mr *memberRight) deleteRightByHidUid(dhid int, uid int64) (bool, error) {
	key := strconv.Itoa(dhid) + "," + strconv.Itoa(int(uid))
	mr.DeleteRight(key)
	err := GetDBMgr().GetDBmControl().Where("dhid = ? and uid = ?", dhid, uid).Delete(models.HouseMemberUserRight{}).Error
	if err != nil {
		xlog.Logger().Errorf("DelUserRight db error：%v, hid:%d, uid:%d", err, dhid, uid)
		return false, err
	}
	return true, nil
}

// 转换更新之后的权限数据可是
func (mr *memberRight) changeDepRight(DepHmem *HouseMember, CapHmem *HouseMember) (string, error) {
	depKey := strconv.Itoa(int(DepHmem.DHId)) + "," + strconv.Itoa(int(DepHmem.UId))
	capRoleMap, err := mr.FindRightByMember(CapHmem, false)
	if err != nil {
		return "", err
	}
	strRight, _ := mr.selectDefaultUserRightByRole(RoleDeputy)
	depRightMap, _ := mr.changeUroleData(strRight)
	for _, value := range capRoleMap {
		capItemValue := value.(map[string]interface{})
		for k1, v := range capItemValue {
			capItemV := v.(map[string]interface{})
			capStatus := int(capItemV["minor_status"].(float64))

			for _, value1 := range depRightMap {
				depItemValue := value1.(map[string]interface{})
				for k, v1 := range depItemValue {
					depItemV := v1.(map[string]interface{})
					//depStatus := int(depItemV["minor_status"].(float64))
					if k1 == k {
						if capStatus == RightKePeiZhiNull {
							depItemV["minor_status"] = RightNoPeiZhiNull
						}
						//else if capStatus == RightKePeiZhiHave && depStatus == RightKePeiZhiHave {
						//	depItemV["minor_status"] = RightKePeiZhiHave
						//}
					}
				}
			}
		}
	}
	tempJson, _ := json.Marshal(depRightMap)
	mr.SetRight(depKey, string(tempJson))
	_, err = mr.UpdateRight(DepHmem)
	if err != nil {
		return "", err
	}
	rightStr := mr.GetRight(depKey)
	if player := GetPlayerMgr().GetPlayer(DepHmem.UId); player != nil {
		player.SendMsg(consts.MsgTypeHmUpdateUserRight_NTF, static.Msg_S2C_UpdateHmUserRight{
			Hid:         DepHmem.HId,
			Uid:         DepHmem.UId,
			UpdateRight: rightStr,
		})
	}
	return rightStr, nil
}

// 校正权限函数 比如有新增的权限 但是当前用户并没有新增对应的权限时做处理
func (mr *memberRight) adjustRight(hmem *HouseMember, key string, mRight string) (string, error) {

	var r models.HouseMemberRight
	if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberRight{}).Where("urole = ? and minor_key = ?", mr.getHmUserRole(hmem), key).First(&r).Error; err != nil {
		cuserror := xerrors.NewXError("获取默认权限失败")
		return "", cuserror
	}
	mRightMap, _ := mr.changeUroleData(mRight)
	for key, value := range mRightMap {
		itemValue := value.(map[string]interface{})
		if key == r.BigKey {
			var newItem map[string]interface{}
			newItem = make(map[string]interface{})
			newItem["minor_status"] = r.MinorStatus
			newItem["minor_id"] = r.MinorId
			newItem["minor_key"] = r.MinorKey
			newItem["minor_name"] = r.MinorName
			itemValue[r.MinorKey] = newItem
		}
		mRightMap[key] = itemValue
	}

	hmemKey := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	mRightStr, _ := json.Marshal(mRightMap)
	updateMap := make(map[string]interface{})
	updateMap["uright"] = string(mRightStr)
	if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Where("dhid = ? and uid = ?", hmem.DHId, hmem.UId).Update(updateMap).Error; err != nil {
		//这里没有数据存一条默认数据
		cuserror := xerrors.NewXError("修改默认权限失败")
		return "", cuserror
	}

	mr.SetRight(hmemKey, string(mRightStr))
	rightStr := mr.GetRight(hmemKey)

	return rightStr, nil

}

// 校正权限函数 比如有新增的权限 但是当前用户并没有新增对应的权限时做处理
func (mr *memberRight) adjustRightAll(hmem *HouseMember, mRight string) (string, error) {

	//这个函数会去map数据 如果默认权限有新增 则要重启服务器
	defaultRight, err := mr.selectDefaultUserRightByRole(mr.getHmUserRole(hmem))
	if err != nil {
		return "", err
	}
	mRightMap, _ := mr.changeUroleData(mRight)
	defaultRightMap, _ := mr.changeUroleData(defaultRight)
	// k和v 紧接着的第一位数 1为默认的数据 2为自身的数据
	isAdjust := false
	for k10, v10 := range defaultRightMap {
		itemV10 := v10.(map[string]interface{})
		for k20, v20 := range mRightMap {
			itemV20 := v20.(map[string]interface{})
			if k10 == k20 { // 这里的 k10  k20 为权限大类型的字符串
				for k11, v11 := range itemV10 {
					itemV11 := v11.(map[string]interface{})
					_, ok := itemV20[k11]
					if !ok {
						isAdjust = true
						var newItemV21 map[string]interface{}
						newItemV21 = make(map[string]interface{})
						newItemV21["minor_status"] = itemV11["minor_status"]
						newItemV21["minor_id"] = itemV11["minor_id"]
						newItemV21["minor_key"] = itemV11["minor_key"]
						newItemV21["minor_name"] = itemV11["minor_name"]
						itemV20[k11] = newItemV21
					}
				}
			}
			mRightMap[k20] = itemV20
		}

	}

	if !isAdjust {
		return mRight, nil
	}
	hmemKey := strconv.Itoa(int(hmem.DHId)) + "," + strconv.Itoa(int(hmem.UId))
	mRightStr, _ := json.Marshal(mRightMap)
	updateMap := make(map[string]interface{})
	updateMap["uright"] = string(mRightStr)
	if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Where("dhid = ? and uid = ?", hmem.DHId, hmem.UId).Update(updateMap).Error; err != nil {
		//这里没有数据存一条默认数据
		cuserror := xerrors.NewXError("修改默认权限失败")
		return "", cuserror
	}

	mr.SetRight(hmemKey, string(mRightStr))
	rightStr := mr.GetRight(hmemKey)

	return rightStr, nil

}

func GetRightKey(str string) string {
	var switchMap map[string]string
	switchMap = make(map[string]string)
	switchMap["BanWeChat"] = MinorBanWxTea
	switchMap["CapSetDep"] = MinorRightNull
	switchMap["IsRecShowParent"] = MinorRightNull // 战绩详情中显示父级归属 仅仅限制 盟主才有的权限
	if switchMap[str] != "" {
		return switchMap[str]
	}
	return ""
}
