package components

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"strings"
	"time"
)

// 是否开启楼层扣除房费兼容模式
const FloorDeductCompatible = false

// 包厢牌桌
type HouseApi struct {
	GameNum         string
	TId             int
	HId, NFId, NTId int
	DHId, FId       int64
	Creator         int64
	FloorPlayerNum  int
	RealPlayerNum   int
}

func (h *HouseApi) HouseKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, h.DHId)
}

func (h *HouseApi) FloorKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, h.DHId)
}

func (h *HouseApi) MemberKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, h.DHId)
}

// 得到桌子包厢成员信息 读不加锁
func (h *HouseApi) GetMember(uid int64) (*static.HouseMember, error) {
	return server2.GetDBMgr().GetDBrControl().HouseMemberQueryById(h.DHId, uid)
}

// 更新桌子包厢成员信息
func (h *HouseApi) SetMember(mem *static.HouseMember) error {
	buf, err := json.Marshal(mem)
	if err != nil {
		return err
	}
	return server2.GetDBMgr().Redis.HSet(h.MemberKey(), fmt.Sprintf("%d", mem.UId), fmt.Sprintf("%s", buf)).Err()
}

func (h *HouseApi) GetHouseMemMap() static.HouseMemberMap {
	meme := make([]*static.HouseMember, 0)
	err := server2.GetDBMgr().GetDBrControl().RedisV2.HVals(h.MemberKey()).ScanSlice(&meme)
	if err != nil {
		xlog.Logger().Error(err)
	}
	res := make(static.HouseMemberMap)
	for i := 0; i < len(meme); i++ {
		res[meme[i].UId] = meme[i]
	}
	return res
}

func (h *HouseApi) GetHouseMemMapOfTable(tableUsers ...int64) (static.HouseMemberMap, []int64) {
	memMap := h.GetHouseMemMap()
	res := make(static.HouseMemberMap)
	partnerIds := make([]int64, 0)
	for i := 0; i < len(tableUsers); i++ {
		next := tableUsers[i]
		_, exist := res[next]
		// 此链路已存在
		if exist {
			continue
		}
		for next > 0 {
			mem, ok := memMap[next]
			if !ok {
				break
			}
			if mem == nil {
				break
			}
			res[next] = mem
			if mem.IsPartner() {
				next = mem.Superior
				partnerIds = append(partnerIds, mem.UId)
			} else {
				next = mem.Partner
			}
		}
	}
	return res, partnerIds
}

// 得到桌子包厢信息 读不加锁
func (h *HouseApi) GetFloor() (*static.HouseFloor, error) {
	floor, err := server2.GetDBMgr().GetDBrControl().HouseFloorSelect(h.DHId, h.FId)
	if err != nil {
		xlog.Logger().Errorln("get house floor from redis error:", err)
		return nil, err
	}
	return floor, nil

}

// 得到桌子包厢信息 读不加锁
func (h *HouseApi) GetHouse() (*models.House, error) {
	house, err := server2.GetDBMgr().GetDBrControl().GetHouseInfoById(h.DHId)
	if err != nil {
		xlog.Logger().Errorln("get house from redis error:", err)
		return nil, err
	}
	return house, nil
}

// 得到桌子包厢信息 读不加锁
func (h *HouseApi) GetFloorVitaminOption() (*static.FloorVitaminOptions, error) {
	house, err := h.GetHouse()
	if err != nil {
		return nil, err
	}
	floor, err := h.GetFloor()
	if err != nil {
		xlog.Logger().Errorln("get house floor from redis error:", err)
		return nil, err
	}
	floor.FloorVitaminOptions.IsVitamin = floor.IsVitamin && house.IsVitamin
	floor.FloorVitaminOptions.IsGamePause = floor.IsGamePause && house.IsGamePause
	return &floor.FloorVitaminOptions, nil
}

// 得到包厢某个楼层的所有的队长配置
func (h *HouseApi) GetHouseFloorWholePartnersPyramid() map[int64]*models.HousePartnerPyramid {
	models := make(models.HousePartnerPyramidFloors, 0)
	err := server2.GetDBMgr().GetDBmControl().Where("dhid = ? and dfid = ?", h.DHId, h.FId).Find(&models).Error
	if err != nil {
		xlog.Logger().Error(err)
	}
	return models.ToMapFloor(h.FId)
}

/**
GetFloorRoyaltyByProvider: 根据桌子所在的包厢楼层，通过玩家的输赢分 得到他提供给每级队长的收益
@param memMap: 包厢所有成员哈希表 // 可通过 GetDBMgr().GetHouseMemMap()接口获取
@param hid: 包厢DHid
@param fid: 楼层Fid
@param provider: 玩家Id
@param cost: 本局房费扣除数量
*/
// TODO: 实现按分数档位计算收益
func (h *HouseApi) GetFloorRoyaltyByProvider(memMap static.HouseMemberMap, configMap map[int64]*models.HousePartnerPyramid, provider int64) (map[int64]int64, []int64) {
	res := make(map[int64]int64)
	partnerLink := make([]int64, 0)
	member, ok := memMap[provider]

	for ok && member != nil {
		if member.IsPartner() { // 如果这个提供者是队长
			partnerLink = append(partnerLink, member.UId)
			var royalty, superRoyalty int64
			config, ok1 := configMap[member.UId]
			if ok1 {
				royalty, superRoyalty = config.EarningsInfo()
			}
			if member.UId == provider {
				res[member.UId] = royalty
			}
			if member.Superior > 0 {
				res[member.Superior] = superRoyalty
			}

			member, ok = memMap[member.Superior]
		} else if member.Partner > 0 { // 如果提供者是某个队长名下成员
			provider = member.Partner
			member, ok = memMap[member.Partner]
		} else { // 既不是队长 也不是队长名下玩家，则无法计算收益
			break
		}
	}

	return res, partnerLink
}

// ! 获取玩家队长信息
func (h *HouseApi) GetHouseMemberPartnerInfo(uid int64) (PartnerInfo *static.HouseMember, err error) {
	PartnerInfo, err = h.GetMember(uid)
	if PartnerInfo != nil && PartnerInfo.Partner > 0 {
		if PartnerInfo.Partner == 1 {
			return
		} else {
			return h.GetMember(PartnerInfo.Partner)
		}
	} else {
		PartnerInfo = nil
	}
	return
}

// ! 获取玩家合伙人id以及合伙人上级id
func (h *HouseApi) GetHouseMemberPartnerAndSuperiorId(uid int64) (int64, int64) {
	var partnerId, superiorId int64
	mem, _ := h.GetMember(uid)
	if mem != nil && mem.Partner > 0 {
		if mem.Partner == 1 {
			partnerId = mem.UId
		} else {
			partnerId = mem.Partner
		}

		pMem, _ := h.GetMember(partnerId)
		if pMem != nil {
			superiorId = pMem.Superior
		} else {
			superiorId = 0
		}
	} else {
		partnerId = 0
		superiorId = 0
	}
	return partnerId, superiorId
}

func (h *HouseApi) GetRealRoyaltyMap(memMap static.HouseMemberMap, realRoyalty int64) (map[int64]*models.HousePartnerPyramid, error) {
	resultSet := make(map[int64]*models.HousePartnerPyramid)
	db := server2.GetDBMgr().GetDBmControl()
	var err error
	for next := memMap.JuniorsBySuperiors(0); len(next) > 0; next = memMap.JuniorsBySuperiors(next...) {
		nextConfig := make([]*models.HousePartnerPyramid, 0)
		if err = db.Where("dhid = ? and dfid = ? and uid in(?)", h.DHId, h.FId, next).Find(&nextConfig).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return resultSet, err
		}
		for i := 0; i < len(nextConfig); i++ {
			cur := nextConfig[i]
			mem, ok := memMap[cur.Uid]
			if !ok {
				continue
			}
			if mem == nil {
				continue
			}
			if mem.IsPartner() {
				if mem.Superior > 0 {
					sup, ok1 := resultSet[mem.Superior]
					if !ok1 {
						continue
					}
					if sup != nil {
						if sup.Total < 0 {
							cur.Total = sup.Total
						} else if sup.ConfiguredRoyaltyPercent() {
							cur.Total = sup.Total * sup.RealRoyaltyPercent() / 100
						} else {
							cur.Total = 0
						}
					}
				} else {
					cur.Total = realRoyalty
				}
			}
			resultSet[cur.Uid] = cur
		}
	}

	return resultSet, nil
}

// ! 统计玩家收益
func (h *HouseApi) StatisticsPartnerProfit(realRoyalty int64, /*桌上每个玩家给上级提供的真实收益*/
	usersCost map[int64]int64 /*每个玩家的真实房费扣除*/, payType int /*是否未托管玩家支付*/, gameCommon *Common) error {
	playerUids := gameCommon.GetUids()
	memMap, partnerIds := h.GetHouseMemMapOfTable(playerUids...)
	/////////////////////////////////////////////////////////////////////////////////////////
	// 上面是老逻辑，不动，这里写新逻辑
	timeNow := time.Now().Unix()
	partnerAttr, err := server2.GetDBMgr().SelectHouseAllPartnerAttr(h.DHId, partnerIds...)
	partnerRewards := make([]*models.PartnerRewardT, 0)

	if err != nil {
		xlog.Logger().Error(err)
		for _, uid := range playerUids {
			partnerRewards = append(partnerRewards, &models.PartnerRewardT{
				DHid:        h.DHId,
				DFid:        h.FId,
				GameNum:     h.GameNum,
				PlayerId:    uid,
				Partner:     h.Creator,
				PartnerType: int(PartnerLevelOwner),
				Reward:      realRoyalty,
				CreatedTime: timeNow,
			})
		}
		gameCommon.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("获取队长收益配置失败:%v", err))
	} else {
		for i := 0; i < gameCommon.GetPlayerCount(); i++ {
			userItem := gameCommon.GetUserItemByChair(uint16(i))
			if userItem != nil {
				// 本局这个人提供的真实收益
				playerMem, ok := memMap[userItem.Uid]
				if !ok {
					partnerRewards = append(partnerRewards, &models.PartnerRewardT{
						DHid:        h.DHId,
						DFid:        h.FId,
						GameNum:     h.GameNum,
						PlayerId:    userItem.Uid,
						Partner:     h.Creator,
						PartnerType: int(PartnerLevelOwner),
						Reward:      realRoyalty,
						CreatedTime: timeNow,
					})
					gameCommon.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("mem uid = %d, 不存在", userItem.Uid))
					continue
				}
				// 直接队长id，0代表不在任何队长名下
				var partnerId int64
				if playerMem.IsPartner() {
					partnerId = playerMem.UId
				} else {
					partnerId = playerMem.Partner
				}

				ps, ok := h.GetRealPartnerAttr(memMap, partnerAttr, h.Creator,
					partnerId, realRoyalty, userItem, gameCommon)

				if ok {
					gameCommon.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("%d -> 本局提供收益详情: %s", userItem.Uid, ps))
					for _, p := range ps {
						partnerRewards = append(partnerRewards, &models.PartnerRewardT{
							DHid:        h.DHId,
							DFid:        h.FId,
							GameNum:     h.GameNum,
							PlayerId:    userItem.Uid,
							Partner:     p.Uid,
							PartnerType: int(p.Level),
							Reward:      p.Reward,
							CreatedTime: timeNow,
						})
					}
				} else {
					partnerRewards = append(partnerRewards, &models.PartnerRewardT{
						DHid:        h.DHId,
						DFid:        h.FId,
						GameNum:     h.GameNum,
						PlayerId:    userItem.Uid,
						Partner:     h.Creator,
						PartnerType: int(PartnerLevelOwner),
						Reward:      realRoyalty,
						CreatedTime: timeNow,
					})
					gameCommon.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("%d -> 计算收益失败。", userItem.Uid))
				}
			} else {
				//partnerRewards = append(partnerRewards, &models.PartnerRewardT{
				//	DHid:        h.DHId,
				//	DFid:        h.FId,
				//	GameNum:     h.GameNum,
				//	PlayerId:    userItem.Uid,
				//	Partner:     h.Creator,
				//	PartnerType: int(PartnerLevelOwner),
				//	Reward:      realRoyalty,
				//	CreatedTime: timeNow,
				//})
				gameCommon.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("座位号%d,获取玩家失败。", i))
			}
		}
	}
	tx := server2.GetDBMgr().GetDBmControl().Begin()
	rewardMap := make(map[int64]int64)
	for _, reward := range partnerRewards {
		err = tx.Create(reward).Error
		if err != nil {
			tx.Rollback()
			xlog.Logger().Error(err)
			gameCommon.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("保存队长收益失败:%v", err))
			return err
		}
		rewardMap[reward.Partner] += reward.Reward
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		xlog.Logger().Error(err)
		gameCommon.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("提交队长收益失败:%v", err))
		return err
	}

	key := fmt.Sprintf("houseOwner:%d:%d", h.HId, h.Creator)
	date := time.Now().Format(time.DateOnly)
	_, err = server2.GetDBMgr().GetDBrControl().RedisV2.HIncrBy(key, date, rewardMap[h.Creator]).Result()
	if err != nil {
		xlog.Logger().Error(err)
		gameCommon.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("保存圈主收益失败:%v", err))
	}

	/////////////////////////////////////////////////////////////////////////////////////////
	configMap, err := h.GetRealRoyaltyMap(memMap, realRoyalty)
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}

	tableInfo := gameCommon.GetTableInfo()
	gameNum := tableInfo.GameNum
	dhid := tableInfo.DHId
	dfid := tableInfo.FId
	dfloorIndex := tableInfo.NFId

	sql := "insert into `house_partner_royalty_detail` (`dhid`,`dfid`,`dfloorindex`,`gamenum`,`playeruser`,`playerpartner`,`beneficiary`,`providerround`, `selfprofit`,`subprofit`,`rent`,`partnerlink`,`created_at`) values "

	haveInsert := false
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		userItem := gameCommon.GetUserItemByChair(uint16(i))
		if userItem != nil {
			royaltyMap, partnerLink := h.GetFloorRoyaltyByProvider(memMap, configMap, userItem.GetUserID())
			if len(royaltyMap) != len(partnerLink) {
				continue
			}

			if len(partnerLink) <= 0 {
				continue
			}

			linkJson, err := json.Marshal(partnerLink)
			linkStr := ""
			if err == nil {
				linkStr = static.HF_Bytestoa(linkJson)
			} else {
				xlog.Logger().Errorf("合伙人关系转json失败:uid:%d, link：%v, err:%v", userItem.Uid, partnerLink, err)
			}

			for j := 0; j < len(partnerLink); j++ {
				pUid := partnerLink[j]
				profit := royaltyMap[pUid]

				if j == 0 {
					sql += fmt.Sprintf("('%d', '%d', '%d', '%s', '%d', '%d', '%d', '%d', '%d', '%d', '%d', '%s', now()),", dhid, dfid, dfloorIndex, gameNum, userItem.Uid, partnerLink[0], pUid, 1, profit, 0, realRoyalty, linkStr)
				} else {
					sql += fmt.Sprintf("('%d', '%d', '%d', '%s', '%d', '%d', '%d', '%d', '%d', '%d', '%d', '%s', now()),", dhid, dfid, dfloorIndex, gameNum, userItem.Uid, partnerLink[0], pUid, 0, 0, profit, realRoyalty, linkStr)
				}

				haveInsert = true
			}
		}
	}

	if !haveInsert {
		err = fmt.Errorf("gamenum : %s  did not have partnerProfit", gameNum)
		xlog.Logger().Error(err)
		return err
	}
	sql = strings.TrimRight(sql, ",")
	sql += ";"

	err = server2.GetDBMgr().GetDBmControl().Exec(sql).Error

	if err != nil {
		err = fmt.Errorf("gamenum : %s StatisticsPartnerProfit error, err : %v", gameNum, err)
		xlog.Logger().Error(err)
		return err
	}
	return nil
}

type PartnerProfitList []PartnerProfit

func (p PartnerProfitList) String() string {
	l := len(p)
	ss := make([]string, l)
	for i := 0; i < l; i++ {
		if i == 0 {
			ss[i] = fmt.Sprintf("圈主:%d[%d]", p[i].Uid, p[i].Reward)
		} else {
			ss[i] = fmt.Sprintf("%d级队长:%d[%d]", i, p[i].Uid, p[i].Reward)
		}
	}
	return strings.Join(ss, "; ")
}

type PartnerLevel int

const (
	PartnerLevelOwner     PartnerLevel = 0
	PartnerLevel1stLeader PartnerLevel = 1
	PartnerLevel2st       PartnerLevel = 2
	PartnerLevel3st       PartnerLevel = 3
	PartnerLevel4st       PartnerLevel = 4
)

type PartnerProfit struct {
	Uid    int64
	Reward int64 // 百分比
	Level  PartnerLevel
}

func (h *HouseApi) GetRealPartnerAttr(memMap static.HouseMemberMap, attrMap map[int64]models.HousePartnerAttr,
	owner, partnerId, rewardTotal int64, userItem *Player, gameCommon *Common) (PartnerProfitList, bool) {
	pfs := make(PartnerProfitList, 0)
	if partnerId > 0 {
		partner, ok := memMap[partnerId]
		if !ok {
			gameCommon.OnWriteGameRecord(userItem.GetChairID(),
				fmt.Sprintf("%d -> 队长不存在:pid=%d", userItem.Uid, partnerId))
			return nil, false
		}
		if s1 := partner.Superior; s1 > 0 {
			p1, ok := memMap[s1]
			if !ok {
				gameCommon.OnWriteGameRecord(userItem.GetChairID(),
					fmt.Sprintf("%d -> 队长的队长不存在:pid=%d=>%d", userItem.Uid, partnerId, s1))
				return nil, false
			}
			if s2 := p1.Superior; s2 > 0 {
				p2, ok := memMap[s2]
				if !ok {
					gameCommon.OnWriteGameRecord(userItem.GetChairID(),
						fmt.Sprintf("%d -> 队长的队长的队长不存在:pid=%d=>%d=>%d", userItem.Uid, partnerId, s1, s2))
					return nil, false
				}
				if s3 := p2.Superior; s3 > 0 {
					gameCommon.OnWriteGameRecord(userItem.GetChairID(),
						fmt.Sprintf("%d -> 四级反利: 圈主(%d), 组长(%d), 大队长(%d), 组长(%d), 小小队长(%d)",
							userItem.Uid, owner, s3, s2, s1, partnerId))
					// s3是组长，s2是大队长，s1是小队长，partnerId是小小队长
					// 取组长收益
					// 取组长收益
					attr1st, ok := attrMap[s3]
					if ok && attr1st.RewardSuperior >= 0 {
						// 盟主配置了该组长收益
						gameCommon.OnWriteGameRecord(userItem.GetChairID(),
							fmt.Sprintf("%d -> 四级反利: 圈主(%d)给组长(%d)收益配置(%d)",
								userItem.Uid, owner, s3, attr1st.RewardSuperior))
						// 组长收益，从组长收益得出圈主收益
						reward1st := rewardTotal * int64(attr1st.RewardSuperior) / 100
						pfs = append(pfs,
							PartnerProfit{
								Uid:    owner,
								Reward: rewardTotal - reward1st,
								Level:  PartnerLevelOwner,
							},
						)
						attr2st, ok := attrMap[s2]
						if ok && attr2st.RewardSuperior >= 0 {
							// 组长给大队长配置了收益
							// 大队长总收益=
							gameCommon.OnWriteGameRecord(userItem.GetChairID(),
								fmt.Sprintf("%d -> 四级反利: 组长(%d)给大队长(%d) 收益配置(%d)",
									userItem.Uid, s3, s2, attr2st.RewardSuperior))
							reward2st := reward1st * int64(attr2st.RewardSuperior) / 100
							pfs = append(pfs,
								PartnerProfit{
									Uid:    s2,
									Reward: reward1st - reward2st,
									Level:  PartnerLevel1stLeader,
								},
							)
							attr3st, ok := attrMap[s1]
							if ok && attr3st.RewardSuperior >= 0 {
								// 大队长给小队长配置了收益
								// 小队长的总收益为：
								gameCommon.OnWriteGameRecord(userItem.GetChairID(),
									fmt.Sprintf("%d -> 四级反利: 大队长(%d)给小队长(%d) 收益配置(%d)",
										userItem.Uid, s2, s1, attr3st.RewardSuperior))

								reward3st := reward2st * int64(attr3st.RewardSuperior) / 100
								pfs = append(pfs,
									PartnerProfit{
										Uid:    s1,
										Reward: reward2st - reward3st,
										Level:  PartnerLevel2st,
									},
								)
								attr4st, ok := attrMap[partnerId]
								if ok && attr4st.RewardSuperior >= 0 {
									// 小队长给小小队长配置了收益
									// 小小队长的总收益为：
									gameCommon.OnWriteGameRecord(userItem.GetChairID(),
										fmt.Sprintf("%d -> 四级反利: 小队长(%d)给小小队长(%d) 收益配置(%d)",
											userItem.Uid, s1, partnerId, attr4st.RewardSuperior))

									reward4st := reward3st * int64(attr3st.RewardSuperior) / 100
									pfs = append(pfs,
										PartnerProfit{
											Uid:    s1,
											Reward: reward3st - reward4st,
											Level:  PartnerLevel3st,
										},
										PartnerProfit{
											Uid:    partnerId,
											Reward: reward4st,
											Level:  PartnerLevel4st,
										},
									)
								} else {
									// 大队长没有给小队长配置收益
									gameCommon.OnWriteGameRecord(userItem.GetChairID(),
										fmt.Sprintf("%d -> 四级反利: 小队长(%d)没有给小小队长(%d)配置收益",
											userItem.Uid, s1, partnerId))
									pfs = append(pfs,
										PartnerProfit{
											Uid:    s3,
											Reward: reward3st,
											Level:  PartnerLevel3st,
										},
									)
								}
							} else {
								// 大队长没有给小队长配置收益
								gameCommon.OnWriteGameRecord(userItem.GetChairID(),
									fmt.Sprintf("%d -> 四级反利: 大队长(%d)没有给小队长(%d)配置收益",
										userItem.Uid, s1, partnerId))
								pfs = append(pfs,
									PartnerProfit{
										Uid:    s1,
										Reward: reward2st,
										Level:  PartnerLevel2st,
									},
								)
							}
						} else {
							// 组长没有给大队长配置收益
							gameCommon.OnWriteGameRecord(userItem.GetChairID(),
								fmt.Sprintf("%d -> 四级反利: 组长(%d)没有给大队长(%d)配置收益",
									userItem.Uid, s2, s1))
							pfs = append(pfs,
								PartnerProfit{
									Uid:    s2,
									Reward: reward1st,
									Level:  PartnerLevel1stLeader,
								},
							)
						}
					} else {
						// 盟主没有配置该组长收益
						gameCommon.OnWriteGameRecord(userItem.GetChairID(),
							fmt.Sprintf("%d -> 四级反利: 圈主(%d)没有给组长(%d)配置收益",
								userItem.Uid, owner, s2))
						pfs = append(pfs,
							PartnerProfit{
								Uid:    owner,
								Reward: rewardTotal,
								Level:  PartnerLevelOwner,
							},
						)
					}
				} else {
					gameCommon.OnWriteGameRecord(userItem.GetChairID(),
						fmt.Sprintf("%d -> 三级反利: 圈主(%d), 组长(%d), 大队长(%d), 小队长(%d)",
							userItem.Uid, owner, s2, s1, partnerId))
					// s2是组长，s1是大队长，partnerId是小队长
					// 取组长收益
					attr1st, ok := attrMap[s2]
					if ok && attr1st.RewardSuperior >= 0 {
						// 盟主配置了该组长收益
						gameCommon.OnWriteGameRecord(userItem.GetChairID(),
							fmt.Sprintf("%d -> 三级反利: 圈主(%d)给组长(%d)收益配置(%d)",
								userItem.Uid, owner, s2, attr1st.RewardSuperior))
						// 组长收益
						reward1st := rewardTotal * int64(attr1st.RewardSuperior) / 100
						pfs = append(pfs,
							PartnerProfit{
								Uid:    owner,
								Reward: rewardTotal - reward1st,
								Level:  PartnerLevelOwner,
							},
						)
						attr2st, ok := attrMap[s1]
						if ok && attr2st.RewardSuperior >= 0 {
							// 组长给大队长配置了收益
							// 大队长总收益=
							gameCommon.OnWriteGameRecord(userItem.GetChairID(),
								fmt.Sprintf("%d -> 三级反利: 组长(%d)给大队长(%d) 收益配置(%d)",
									userItem.Uid, s2, s1, attr2st.RewardSuperior))
							reward2st := reward1st * int64(attr2st.RewardSuperior) / 100
							pfs = append(pfs,
								PartnerProfit{
									Uid:    s2,
									Reward: reward1st - reward2st,
									Level:  PartnerLevel1stLeader,
								},
							)
							attr3st, ok := attrMap[partnerId]
							if ok && attr3st.RewardSuperior >= 0 {
								// 大队长给小队长配置了收益
								// 小队长的总收益为：
								gameCommon.OnWriteGameRecord(userItem.GetChairID(),
									fmt.Sprintf("%d -> 三级反利: 大队长(%d)给小队长(%d) 收益配置(%d)",
										userItem.Uid, s1, partnerId, attr3st.RewardSuperior))

								reward3st := reward2st * int64(attr3st.RewardSuperior) / 100
								pfs = append(pfs,
									PartnerProfit{
										Uid:    s1,
										Reward: reward2st - reward3st,
										Level:  PartnerLevel2st,
									},
									PartnerProfit{
										Uid:    partnerId,
										Reward: reward3st,
										Level:  PartnerLevel3st,
									},
								)
							} else {
								// 大队长没有给小队长配置收益
								gameCommon.OnWriteGameRecord(userItem.GetChairID(),
									fmt.Sprintf("%d -> 三级反利: 大队长(%d)没有给小队长(%d)配置收益",
										userItem.Uid, s1, partnerId))
								pfs = append(pfs,
									PartnerProfit{
										Uid:    s1,
										Reward: reward2st,
										Level:  PartnerLevel2st,
									},
								)
							}
						} else {
							// 组长没有给大队长配置收益
							gameCommon.OnWriteGameRecord(userItem.GetChairID(),
								fmt.Sprintf("%d -> 三级反利: 组长(%d)没有给大队长(%d)配置收益",
									userItem.Uid, s2, s1))
							pfs = append(pfs,
								PartnerProfit{
									Uid:    s2,
									Reward: reward1st,
									Level:  PartnerLevel1stLeader,
								},
							)
						}
					} else {
						// 盟主没有配置该组长收益
						gameCommon.OnWriteGameRecord(userItem.GetChairID(),
							fmt.Sprintf("%d -> 三级反利: 圈主(%d)没有给组长(%d)配置收益",
								userItem.Uid, owner, s2))
						pfs = append(pfs,
							PartnerProfit{
								Uid:    owner,
								Reward: rewardTotal,
								Level:  PartnerLevelOwner,
							},
						)
					}
				}
			} else {
				// s1是组长，partnerId是大队长
				// 取组长收益
				gameCommon.OnWriteGameRecord(userItem.GetChairID(),
					fmt.Sprintf("%d -> 二级反利: 圈主(%d), 组长(%d), 大队长(%d)",
						userItem.Uid, owner, s1, partnerId))
				attr1st, ok := attrMap[s1]
				if ok && attr1st.RewardSuperior >= 0 {
					gameCommon.OnWriteGameRecord(userItem.GetChairID(),
						fmt.Sprintf("%d -> 二级反利: 圈主(%d)给组长(%d)配置收益(%d)",
							userItem.Uid, owner, s1, attr1st.RewardSuperior))
					// 盟主配置了该组长收益
					reward1st := rewardTotal * int64(attr1st.RewardSuperior) / 100
					pfs = append(pfs,
						PartnerProfit{
							Uid:    owner,
							Reward: rewardTotal - reward1st,
							Level:  PartnerLevelOwner,
						},
					)
					attr2st, ok := attrMap[partnerId]
					if ok && attr2st.RewardSuperior >= 0 {
						gameCommon.OnWriteGameRecord(userItem.GetChairID(),
							fmt.Sprintf("%d -> 二级反利: 组长(%d)给大队长(%d)配置收益(%d)",
								userItem.Uid, s1, partnerId, attr2st.RewardSuperior))
						// 组长给大队长配置了收益
						reward2st := reward1st * int64(attr2st.RewardSuperior) / 100
						pfs = append(pfs,
							PartnerProfit{
								Uid:    s1,
								Reward: reward1st - reward2st,
								Level:  PartnerLevel1stLeader,
							},
							PartnerProfit{
								Uid:    partnerId,
								Reward: reward2st,
								Level:  PartnerLevel2st,
							},
						)
					} else {
						// 组长没有给大队长配置收益
						gameCommon.OnWriteGameRecord(userItem.GetChairID(),
							fmt.Sprintf("%d -> 二级反利: 组长(%d)没有给大队长(%d)配置收益",
								userItem.Uid, s1, partnerId))
						pfs = append(pfs,
							PartnerProfit{
								Uid:    s1,
								Reward: reward1st,
								Level:  PartnerLevel1stLeader,
							},
						)
					}
				} else {
					// 盟主没有配置该组长收益
					gameCommon.OnWriteGameRecord(userItem.GetChairID(),
						fmt.Sprintf("%d -> 二级反利: 圈主(%d)没有给组长(%d)配置收益",
							userItem.Uid, owner, s1))
					pfs = append(pfs,
						PartnerProfit{
							Uid:    owner,
							Reward: rewardTotal,
							Level:  PartnerLevelOwner,
						},
					)
				}
			}
		} else {
			// partnerId是组长
			gameCommon.OnWriteGameRecord(userItem.GetChairID(),
				fmt.Sprintf("%d -> 一级反利: 圈主(%d), 组长(%d)",
					userItem.Uid, owner, partnerId))
			attr1st, ok := attrMap[partnerId]
			if ok && attr1st.RewardSuperior >= 0 {
				// 盟主配置了该组长收益
				gameCommon.OnWriteGameRecord(userItem.GetChairID(),
					fmt.Sprintf("%d -> 一级反利: 圈主(%d)给组长(%d)配置收益(%d)",
						userItem.Uid, owner, partnerId, attr1st.RewardSuperior))
				reward1st := rewardTotal * int64(attr1st.RewardSuperior) / 100
				pfs = append(pfs,
					PartnerProfit{
						Uid:    owner,
						Reward: rewardTotal - reward1st,
						Level:  PartnerLevelOwner,
					},
					PartnerProfit{
						Uid:    partnerId,
						Reward: reward1st,
						Level:  PartnerLevel1stLeader,
					},
				)
			} else {
				// 盟主没有配置该组长收益
				gameCommon.OnWriteGameRecord(userItem.GetChairID(),
					fmt.Sprintf("%d -> 一级反利: 圈主(%d)没有给组长(%d)配置收益",
						userItem.Uid, owner, partnerId))
				pfs = append(pfs,
					PartnerProfit{
						Uid:    owner,
						Reward: rewardTotal,
						Level:  PartnerLevelOwner,
					},
				)
			}
		}
	} else {
		// 盟主名下玩家
		gameCommon.OnWriteGameRecord(userItem.GetChairID(),
			fmt.Sprintf("%d -> 零级反利: 圈主(%d)",
				userItem.Uid, owner))
		pfs = append(pfs,
			PartnerProfit{
				Uid:    owner,
				Reward: rewardTotal,
				Level:  PartnerLevelOwner,
			},
		)
	}
	return pfs, true
}

// ！得到楼层支付信息
func (h *HouseApi) GetFloorPayInfo() *models.HouseFloorGearPay {
	payInfo := new(models.HouseFloorGearPay)
	err := server2.GetDBMgr().GetDBmControl().Where("id = ?", h.FId).First(payInfo).Error
	if err != nil {
		xlog.Logger().Error(err)
		payInfo = models.NewHouseFloorGearPay(h.DHId, h.FId, h.FloorPlayerNum)
	}
	return payInfo
}

// ！得到楼层支付信息
func (h *HouseApi) GetFloorDeductInfo() (*models.HouseFloorVitaminDeduct, error) {
	deductInfo := new(models.HouseFloorVitaminDeduct)
	err := server2.GetDBMgr().GetDBmControl().Where("id = ?", h.FId).First(deductInfo).Error
	if err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}
	return deductInfo, nil
}

// ! 更新玩家比赛分
func (h *HouseApi) UpdateUserVitamin(uid int64, offset int64, vitaminType models.VitaminChangeType) (int64, error) {
	var err error
	mem := &static.HouseMember{
		DHId: h.DHId,
		UId:  uid,
	}

	cli := server2.GetDBMgr().Redis
	mem.Lock(cli)

	mem, err = h.GetMember(uid)
	if err != nil {
		mem.Unlock(cli)
		return 0, fmt.Errorf("house api get member error %v", err)
	}
	if offset != 0 {
		// 得到前后值
		before := mem.UVitamin
		after := before + offset

		// 定义疲劳值日志函数
		addVitaminLogFunc := func(uid int64, beforeVitamin int64, afterVitamin int64, vt models.VitaminChangeType) error {
			tx := server2.GetDBMgr().GetDBmControl().Begin()
			err = models.AddVitaminLog(h.DHId, uid, uid, beforeVitamin, afterVitamin, vt, h.GameNum, tx)
			if err != nil {
				return err
			}
			err = tx.Commit().Error
			if err != nil {
				return err
			}
			return nil
		}

		// 执行疲劳值日志函数
		tryNum := 5
		ok := false
		for i := 0; i < tryNum; i++ {
			err = addVitaminLogFunc(uid, before, after, vitaminType)
			if err == nil {
				ok = true
				break
			} else {
				xlog.Logger().Error("AddVitaminLog error:", err)
			}
		}

		if ok {
			if after < 0 {
				for i := 0; i < tryNum; i++ {
					err = addVitaminLogFunc(uid, after, 0, models.SysSend)
					if err == nil {
						ok = true
						break
					} else {
						xlog.Logger().Error("fix AddVitaminLog error:", err)
					}
				}
				if ok {
					after = 0
				}
			}
			mem.UVitamin = after
			err = h.SetMember(mem)
			if err != nil {
				mem.Unlock(cli)
				return before, fmt.Errorf("house api set member error %v", err)
			}
			mem.Unlock(cli)
		} else {
			mem.Unlock(cli)
			return before, fmt.Errorf("house api addVitaminLogFunc error %v", err)
		}

		var winLose, Aa, Bw int64

		if vitaminType == models.GameCost {
			winLose = offset
		} else {
			if vitaminType == models.BigWinCost {
				Bw = offset
			} else if vitaminType == models.GamePay {
				Aa = offset
			}
			err = server2.GetDBMgr().InsertVitaminCostClear(h.DHId, Aa, Bw)
			if err != nil {
				xlog.Logger().Error(err)
			}
		}

		err = server2.GetDBMgr().InsertVitaminCostFromLastNode(h.DHId, uid, before, mem.UVitamin, winLose, Aa, Bw)
		if err != nil {
			xlog.Logger().Error(err)
		}

		err = server2.GetDBMgr().InsertVitaminCost(h.DHId, h.FId, int64(h.NFId), uid, Aa, Bw, winLose, mem.Partner)
		if err != nil {
			xlog.Logger().Error(err)
		}
	} else {
		mem.Unlock(cli)
	}
	return mem.UVitamin, nil
}
