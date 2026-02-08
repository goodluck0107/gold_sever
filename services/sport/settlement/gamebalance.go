package settlement

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

//! 游戏结算接口
func WriteGameBalance(gameCommon *components2.Common, gameMeta *meta2.Metadata, gameEnd static.Msg_S_GameEnd, dismiss bool) {
	houseApi := gameCommon.GetHouseApi()
	// 写入回放数据
	WriteGameOutData(gameCommon, gameMeta)
	// 写入小结算数据
	WriteGameRound(gameCommon, gameEnd, dismiss)

	// 检查游戏是否结束
	if CheckGameFinish(gameCommon, dismiss) {
		// 游戏结束需要写入总结算
		gameScore, nowTime, floorGearPay := WriteGameTotal(houseApi, gameCommon, gameEnd, dismiss, true, -1)

		// 判断是否是有效对局
		if CheckGameValid(gameCommon) {
			// 记录包厢活动数据
			WriteGameHouseActivity(gameCommon, gameScore, nowTime.Unix())

			// 写入每日战绩统计
			WriteGameDay(houseApi, gameCommon, gameScore, floorGearPay)
		} else {
			// 处理无效局处理
			WriteInvalidRound(gameCommon)
		}

		// 通知游戏结束
		WriteGameNotify(gameCommon)
	} else {
		//  游戏小局结束需要更新大结算数据
		WriteGameTotal(houseApi, gameCommon, gameEnd, dismiss, false, -1)
	}
	// 写入游戏扣房卡
	WriteGameDeleteFangKa(gameCommon)
}

//20210401 苏大强 选飘、下跑的时候申请解散
func WriteGameBalance_ex(gameCommon *components2.Common, gameMeta *meta2.Metadata, gameEnd static.Msg_S_GameEnd, dismiss bool, xuanpiaostatus bool) {
	houseApi := gameCommon.GetHouseApi()
	jushu := int(gameCommon.CurCompleteCount)
	if !xuanpiaostatus {
		// 写入回放数据
		WriteGameOutData(gameCommon, gameMeta)
		// 写入小结算数据
		WriteGameRound(gameCommon, gameEnd, dismiss)
	} else {
		if jushu > 1 {
			jushu--
		}
	}
	// 检查游戏是否结束
	if CheckGameFinish(gameCommon, dismiss) {
		// 游戏结束需要写入总结算
		gameScore, nowTime, floorGearPay := WriteGameTotal(houseApi, gameCommon, gameEnd, dismiss, true, jushu)

		// 判断是否是有效对局
		if CheckGameValid(gameCommon) {
			// 记录包厢活动数据
			WriteGameHouseActivity(gameCommon, gameScore, nowTime.Unix())

			// 写入每日战绩统计
			WriteGameDay(houseApi, gameCommon, gameScore, floorGearPay)
		} else {
			// 处理无效局处理
			WriteInvalidRound(gameCommon)
		}

		// 通知游戏结束
		WriteGameNotify(gameCommon)
	} else {
		//  游戏小局结束需要更新大结算数据
		WriteGameTotal(houseApi, gameCommon, gameEnd, dismiss, false, jushu)
	}
	// 写入游戏扣房卡
	WriteGameDeleteFangKa(gameCommon)
}

//! 游戏出牌记录统计
func WriteGameOutData(gameCommon *components2.Common, gameMeta *meta2.Metadata) {
	recordReplay := new(models.RecordGameReplay)
	recordReplay.GameNum = gameCommon.GetTableInfo().GameNum
	recordReplay.RoomNum = gameCommon.GetTableInfo().Id
	recordReplay.PlayNum = int(gameCommon.CurCompleteCount)
	recordReplay.ServerId = server2.GetServer().Con.Id
	recordReplay.HandCard = GetWriteHandReplayRecordCString(gameCommon, gameMeta.ReplayRecord)
	recordReplay.OutCard = GetWriteOutReplayRecordCString(gameCommon, gameMeta.ReplayRecord)
	recordReplay.KindID = gameCommon.GetTableInfo().KindId
	recordReplay.CardsNum = int(gameMeta.ReplayRecord.LeftCardCount)
	recordReplay.UVitaminMap = gameMeta.ReplayRecord.UVitamin
	recordReplay.CreatedAt = time.Now()
	recordReplay.EndInfo = static.HF_JtoA(gameMeta.ReplayRecord.EndInfo)

	server2.GetDBMgr().InsertGameRecordReplay(recordReplay)

	gameCommon.RoundReplayId = recordReplay.Id

	gameMeta.ReplayRecord.Reset()
}

//! 游戏小结算统计
func WriteGameRound(gameCommon *components2.Common, gameEnd static.Msg_S_GameEnd, dismiss bool) {
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		user := gameCommon.GetPlayerByChair(uint16(i))
		if user == nil {
			xlog.Logger().Errorln(fmt.Sprintf("写战绩获取玩家失败，gamenum = %s, seatid = %d", gameCommon.GetTableInfo().GameNum, i))
			continue
		}
		record := new(models.RecordGameRound)
		record.UId = user.Uid
		record.UName = user.Name
		record.Ip = user.Ip
		record.UUrl = user.ImgUrl
		record.UGenber = user.Sex
		record.GameNum = gameCommon.GetTableInfo().GameNum
		record.RoomNum = gameCommon.GetTableInfo().Id
		record.PlayNum = int(gameCommon.CurCompleteCount)
		record.ServerId = server2.GetServer().Con.Id
		record.SeatId = i
		record.ReplayId = gameCommon.RoundReplayId
		record.WinScore = gameEnd.GameScore[i]
		record.BeginDate = gameCommon.GameBeginTime
		record.CreatedAt = time.Now()
		record.EndDate = time.Now()
		record.Radix = gameCommon.Rule.Radix
		record.ScoreKind = GetScoreKindByDismiss(record.WinScore, dismiss)

		//写入小结算
		server2.GetDBMgr().InsertGameRecord(record)
	}
}

//! 游戏扣房卡
func WriteGameDeleteFangKa(gameCommon *components2.Common) {
	if gameCommon.CurCompleteCount == 1 {
		gameCommon.TableDeleteFangKa(gameCommon.CurCompleteCount)
	}
}

//! 游戏小结算更新总结算
//20210401 苏大强 修改一下，下跑下漂的地方要退一个数，现在改一下
func WriteGameTotal(houseApi *components2.HouseApi, gameCommon *components2.Common, gameEnd static.Msg_S_GameEnd, dismiss bool, finish bool, jushu int) ([meta2.MAX_PLAYER]int, time.Time, *models.HouseFloorGearPay) {
	nowTime := time.Now()

	validGame := CheckGameValid(gameCommon)

	gameScore := [meta2.MAX_PLAYER]int{}

	userRecords := [meta2.MAX_PLAYER]*models.RecordGameTotal{}
	PlayCount := jushu
	if PlayCount == -1 {
		PlayCount = int(gameCommon.CurCompleteCount)
	}
	// 先得到每个玩家的record_game_total记录，用于计算有效局和大赢家
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		user := gameCommon.GetPlayerByChair(uint16(i))

		if user == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，第（%d）局，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）",
				gameCommon.CurCompleteCount, i, gameCommon.GetTableInfo().KindId, gameCommon.GetTableInfo().HId, gameCommon.GetTableInfo().Id))
			continue
		}

		recordTotal, err := server2.GetDBMgr().SelectGameRecordTotal(gameCommon.GetTableInfo().GameNum, user.Uid)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				xlog.Logger().Errorln(fmt.Sprintf("查询总战绩失败，gamenum = %s, seatid = %d, error = %v", gameCommon.GetTableInfo().GameNum, i, err))
			}
		} else {
			if recordTotal.Id > 0 {
				userRecords[i] = &recordTotal
			}
		}
	}

	// 得到每个玩家的总分
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		if record := userRecords[i]; record != nil {
			gameScore[i] = record.WinScore + gameEnd.GameScore[i]
		} else {
			gameScore[i] = gameEnd.GameScore[i]
		}
	}

	var (
		bigWinScore  int
		floorGearPay *models.HouseFloorGearPay
	)
	for _, score := range gameScore {
		if score > bigWinScore {
			bigWinScore = score
		}
	}
	if finish && houseApi != nil {
		floorGearPay = houseApi.GetFloorPayInfo()
		//gameCommon.OnWriteGameRecord(static.INVALID_CHAIR,
		//	fmt.Sprintf("大局结束=> 楼层支付信息为：%#v,",floorGearPay))
	}

	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		user := gameCommon.GetPlayerByChair(uint16(i))

		if user == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，第（%d）局，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）",
				gameCommon.CurCompleteCount, i, gameCommon.GetTableInfo().KindId, gameCommon.GetTableInfo().HId, gameCommon.GetTableInfo().Id))
			continue
		}

		recordTotal := userRecords[i]
		if recordTotal == nil {
			recordTotal = new(models.RecordGameTotal)
			recordTotal.KindId = gameCommon.GetTableInfo().KindId
			recordTotal.GameNum = gameCommon.GetTableInfo().GameNum
			recordTotal.RoomNum = gameCommon.GetTableInfo().Id
			recordTotal.PlayCount = PlayCount
			recordTotal.Round = gameCommon.GetTableInfo().Config.RoundNum
			recordTotal.ServerId = server2.GetServer().Con.Id
			recordTotal.SeatId = i
			recordTotal.ScoreKind = static.ScoreKind_pass
			recordTotal.WinScore = gameEnd.GameScore[i]
			// gameScore[i] = recordTotal.WinScore
			recordTotal.HId = gameCommon.GetTableInfo().DHId
			recordTotal.IsHeart = 0
			recordTotal.FId = int(gameCommon.GetTableInfo().FId)
			recordTotal.DFId = gameCommon.GetTableInfo().NFId

			recordTotal.Uid = user.Uid
			recordTotal.UName = user.Name
			recordTotal.Ip = user.Ip
			recordTotal.HalfWayDismiss = false
			recordTotal.CreatedAt = nowTime
			recordTotal.Radix = gameCommon.Rule.Radix
			recordTotal.IsBigWinner = false
			recordTotal.IsValidRound = false

			if finish {
				if !validGame {
					recordTotal.ScoreKind = static.ScoreKind_pass
				} else {
					recordTotal.ScoreKind = GetScoreKindByFinish(recordTotal.WinScore, finish)
				}

				var partnerId, superiorId int64
				if houseApi != nil {
					partnerId, superiorId = houseApi.GetHouseMemberPartnerAndSuperiorId(user.Uid)
				} else {
					partnerId = 0
					superiorId = 0
				}

				if validGame {
					if houseApi != nil && floorGearPay != nil {
						if floorGearPay.IsValidRound(static.SwitchF64ToVitamin(gameCommon.GetRealScore(bigWinScore))) {
							recordTotal.IsValidRound = true
						} else {
							recordTotal.IsValidRound = false
						}
					} else {
						recordTotal.IsValidRound = true
					}

					if recordTotal.IsValidRound && bigWinScore > 0 && recordTotal.WinScore >= bigWinScore {
						recordTotal.IsBigWinner = true
					} else {
						recordTotal.IsBigWinner = false
					}
				} else {
					recordTotal.IsValidRound = false
					recordTotal.IsBigWinner = false
				}

				recordTotal.Partner = partnerId
				recordTotal.SuperiorId = superiorId
			}

			//! 插入数据库
			_, err := server2.GetDBMgr().InsertGameRecordTotal(recordTotal)
			if err != nil {
				xlog.Logger().Errorln(fmt.Sprintf("插入总战绩失败，gamenum = %s, seatid = %d, error = %v", gameCommon.GetTableInfo().GameNum, i, err))
			}
			continue
		}

		recordTotal.HalfWayDismiss = dismiss
		recordTotal.WinScore += gameEnd.GameScore[i]
		// gameScore[i] = recordTotal.WinScore
		recordTotal.PlayCount = PlayCount
		recordTotal.CreatedAt = nowTime
		if finish {
			recordTotal.ScoreKind = GetScoreKindByFinish(recordTotal.WinScore, finish)
			if houseApi != nil {
				recordTotal.Partner, recordTotal.SuperiorId = houseApi.GetHouseMemberPartnerAndSuperiorId(user.Uid)
			} else {
				recordTotal.Partner = 0
				recordTotal.SuperiorId = 0
			}
			if validGame {
				if houseApi != nil && floorGearPay != nil {
					if floorGearPay.IsValidRound(static.SwitchF64ToVitamin(gameCommon.GetRealScore(bigWinScore))) {
						recordTotal.IsValidRound = true
					} else {
						recordTotal.IsValidRound = false
					}
				} else {
					recordTotal.IsValidRound = true
				}

				if recordTotal.IsValidRound && bigWinScore > 0 && recordTotal.WinScore >= bigWinScore {
					recordTotal.IsBigWinner = true
				} else {
					recordTotal.IsBigWinner = false
				}
			} else {
				recordTotal.IsBigWinner = false
				recordTotal.IsValidRound = false
			}
		} else {
			recordTotal.IsBigWinner = false
			recordTotal.IsValidRound = false
		}
		updateMap := make(map[string]interface{})
		updateMap["win_score"] = recordTotal.WinScore
		updateMap["score_kind"] = recordTotal.ScoreKind
		updateMap["play_count"] = recordTotal.PlayCount
		updateMap["created_at"] = recordTotal.CreatedAt
		updateMap["halfwaydismiss"] = recordTotal.HalfWayDismiss
		updateMap["partner"] = recordTotal.Partner
		updateMap["superiorid"] = recordTotal.SuperiorId
		updateMap["is_big_winner"] = recordTotal.IsBigWinner
		updateMap["is_valid_round"] = recordTotal.IsValidRound

		//mysql更新总结算
		server2.GetDBMgr().UpdataGameRecordTotal(recordTotal.Id, updateMap)

		//mysql更新玩家胜率
		if finish {
			server2.GetDBMgr().UpdataUserRateOfWinning(user.Uid, recordTotal.WinScore)
		}
	}

	return gameScore, nowTime, floorGearPay
}

//! 游戏每日结算统计
func WriteGameDay(houseApi *components2.HouseApi, gameCommon *components2.Common, gameScore [4]int, floorGearPay *models.HouseFloorGearPay) {
	var realScore [meta2.MAX_PLAYER]float64
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		realScore[i] = gameCommon.GetRealScore(gameScore[i])
	}

	if houseApi == nil {
		return
	}

	if floorGearPay == nil {
		return
	}

	//! 计算大赢家的座位号
	bigWinScore := float64(0)
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		if realScore[i] >= bigWinScore {
			bigWinScore = realScore[i]
		}
	}

	IsValidRound := false
	IsBigValidRound := false

	if components2.FloorDeductCompatible {
		ValidScore, ValidBigScore, _ := server2.GetDBMgr().SelectHouseValidRound(gameCommon.GetTableInfo().DHId, gameCommon.GetTableInfo().FId)
		if bigWinScore >= float64(ValidScore) {
			IsValidRound = true

			if bigWinScore >= float64(ValidBigScore) && ValidBigScore > ValidScore {
				IsBigValidRound = true
			}
		}
	} else {
		IsValidRound = floorGearPay.IsValidRound(static.SwitchF64ToVitamin(bigWinScore))
	}

	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		user := gameCommon.GetPlayerByChair(uint16(i))

		if user == nil {
			xlog.Logger().Errorln(fmt.Sprintf("TableWriteGameDateTotal，第（%d）局，玩家（%d）信息为空，游戏类型（%d）包厢id（%d）牌桌号（%d）",
				gameCommon.CurCompleteCount, i, gameCommon.GetTableInfo().KindId, gameCommon.GetTableInfo().HId, gameCommon.GetTableInfo().Id))
			continue
		}

		ValidRound := 0
		ValidBigRound := 0
		if IsValidRound {
			ValidRound = 1
		}
		if IsBigValidRound {
			ValidBigRound = 1
		}

		var partnerId, superiorId int64
		if houseApi := gameCommon.GetHouseApi(); houseApi != nil {
			partnerId, superiorId = houseApi.GetHouseMemberPartnerAndSuperiorId(user.Uid)
		} else {
			partnerId = 0
			superiorId = 0
		}

		bwTimes := 0
		if IsValidRound && bigWinScore > 0 && realScore[i] >= bigWinScore {
			bwTimes++
		}

		server2.GetDBMgr().HouseUpdataGameDayRecord(gameCommon.GetTableInfo().DHId, gameCommon.GetTableInfo().FId,
			gameCommon.GetTableInfo().NFId, user.Uid, 1, bwTimes, gameScore[i], ValidRound,
			ValidBigRound, partnerId, superiorId, gameCommon.Rule.Radix)
		server2.GetDBMgr().UpdataGameHouseMemberRecord(gameCommon.GetTableInfo().DHId, user.Uid, 1, bwTimes)
	}

	// 跳转至总结算事件
	if gameCommon.CurCompleteCount > 0 {
		gameCommon.OnBalance(houseApi, int(gameCommon.CurCompleteCount), bigWinScore, realScore[:], floorGearPay)
	} else {
		gameCommon.OnInvalid()
	}
}

//! 游戏结算通知
func WriteGameNotify(gameCommon *components2.Common) {
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		user := gameCommon.GetPlayerByChair(uint16(i))
		// 结算通知
		if gameCommon.IsHouse() {
			// 更新玩家包厢上局对战记录
			recordPlayer := new(static.HTableRecordPlayers)
			recordPlayer.HId = gameCommon.GetTableInfo().HId
			recordPlayer.FId = gameCommon.GetTableInfo().FId
			recordPlayer.TId = gameCommon.GetTableId()
			recordPlayer.KId = gameCommon.KIND_ID
			recordPlayer.Users = gameCommon.GetOtherUids(user.Uid)
			if err := server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(user.Uid, recordPlayer); err != nil {
				xlog.Logger().Errorf("更新玩家(%d)的包厢桌子最新的对战玩家记录失败:%v", user.Uid, err)
			}
		} else {
			// 如果不是包厢玩法，则删掉包厢上局对战记录，以免影响再来一局功能
			_ = server2.GetDBMgr().GetDBrControl().UpdateHRecordPlayers(user.Uid, nil)
		}
	}
}

//! 游戏无效局处理
func WriteInvalidRound(gameCommon *components2.Common) {
	if !CheckGameValid(gameCommon) {
		gameCommon.OnInvalid()
		//未打完解散房间,扣0房卡,统计一次扣卡
		gameCommon.TableDeleteFangKa(0)
	}
}

//! 更新包厢活动
func WriteGameHouseActivity(gameCommon *components2.Common, gameScore [4]int, finishTime int64) {

	//! 计算大赢家的座位号
	bigWinScore := 0
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		if gameScore[i] >= bigWinScore {
			bigWinScore = gameScore[i]
		}
	}

	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		// 活动数据
		dhid := gameCommon.GetTableInfo().DHId
		if dhid != 0 {
			hact, err := server2.GetDBMgr().GetDBrControl().HouseActivityList(dhid, true)
			if err != nil {
				xlog.Logger().Errorln(err)
				return
			}
			for _, act := range hact {
				// 楼层活动
				InActFloor := false
				fidstrs := strings.Split(act.FId, ",")
				for i := 0; i < len(fidstrs); i++ {
					if gameCommon.GetTableInfo().FId == static.HF_Atoi64(fidstrs[i]) {
						InActFloor = true
						break
					}
				}
				if !InActFloor {
					continue
				}
				// 活动时间
				if finishTime < act.BegTime || finishTime > act.EndTime {
					continue
				}

				// 局数累计 活动
				if act.Kind == consts.HFACT_ROUNDS {
					err := server2.GetDBMgr().GetDBrControl().HouseActivityRecordInsert(act.Id, gameCommon.GetTableInfo().Users[i].Uid, 1)
					if err != nil {
						xlog.Logger().Errorln(err)
					}
					if act.Type == 1 { //抽奖活动
						err := server2.AddUserTicket(dhid, gameCommon.GetTableInfo().Users[i].Uid, act.Id, 1)
						if err != nil {
							xlog.Logger().Errorln(err)
						}
					}

				}
				// 活动 大赢家统计
				if act.Kind == consts.HFACT_BW {
					if gameScore[i] == bigWinScore && bigWinScore > 0 {
						err := server2.GetDBMgr().GetDBrControl().HouseActivityRecordInsert(act.Id, gameCommon.GetTableInfo().Users[i].Uid, 1)
						if err != nil {
							xlog.Logger().Errorln(err)
							continue
						}
					}
				}
				// 活动 积分统计
				if act.Kind == consts.HFACT_SCORE {
					err := server2.GetDBMgr().GetDBrControl().HouseActivityRecordInsert(act.Id, gameCommon.GetTableInfo().Users[i].Uid, gameCommon.GetRealScore(gameScore[i]))
					if err != nil {
						xlog.Logger().Errorln(err)
						continue
					}
				}
			}

		}
	}
}

func GetWriteHandReplayRecordCString(gameCommon *components2.Common, replayRecord meta2.Replay_Record) string {
	handCardStr := ""
	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < static.MAX_COUNT; j++ {
			handCardStr += fmt.Sprintf("%02x,", replayRecord.RecordHandCard[i][j])
		}
	}

	//写入分数
	handCardStr += "S:"

	for i := 0; i < gameCommon.GetPlayerCount(); i++ {
		handCardStr += fmt.Sprintf("%d,", replayRecord.Score[i])
	}

	return handCardStr
}

func GetWriteOutReplayRecordCString(gameCommon *components2.Common, replayRecord meta2.Replay_Record) string {
	ourCardStr := ""
	ourCardStr += fmt.Sprintf("P:%02x,", replayRecord.PiziCard)
	if replayRecord.Fengquan > 0 {
		ourCardStr += fmt.Sprintf("f:%02d,", replayRecord.Fengquan)
	}
	var hasHu bool
	// 把胡牌的U拿出来
	endMsgUpdateScore := [meta2.MAX_PLAYER]float64{}
	for i := 0; i < len(replayRecord.VecOrder); i++ {
		recordI := replayRecord.VecOrder[i]
		if recordI.Operation == info2.E_Hu {
			for j, count := i, 0; j < len(replayRecord.VecOrder); j++ {
				recordJ := replayRecord.VecOrder[j]
				if recordJ.Operation == info2.E_GameScore {
					count++
					if count > gameCommon.GetPlayerCount() {
						break
					}
					recordJ.Operation = -1 // 置为无效
					endMsgUpdateScore[recordJ.Chair_id] = recordJ.UserScore
					replayRecord.VecOrder[j] = recordJ
				}
			}
			hasHu = true
			break
		}
	}

	for _, record := range replayRecord.VecOrder {
		if record.Operation < 0 {
			continue
		}
		if len(record.Value) == 0 && record.Operation != info2.E_GameScore {
			xlog.Logger().Errorf("记录数据牌对象异常（空牌值）:玩家ID(%d)发动操作（%d）\n", record.Chair_id, record.Operation)
			continue
		}
		ourCardStr += fmt.Sprintf("%d:", record.Chair_id)
		switch record.Operation {
		case info2.E_SendCard:
			ourCardStr += fmt.Sprintf("S%02x,", record.Value[0])
			break
		case info2.E_OutCard:
			ourCardStr += fmt.Sprintf("O%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG:
			ourCardStr += fmt.Sprintf("o%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG_Magic:
			ourCardStr += fmt.Sprintf("m%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG_PiZi:
			ourCardStr += fmt.Sprintf("z%02x,", record.Value[0])
			break
		case info2.E_Wik_Left:
			ourCardStr += fmt.Sprintf("L%02x,", record.Value[0])
			break
		case info2.E_Wik_Center:
			ourCardStr += fmt.Sprintf("C%02x,", record.Value[0])
			break
		case info2.E_Wik_Right:
			ourCardStr += fmt.Sprintf("R%02x,", record.Value[0])
			break
		case info2.E_Peng:
			ourCardStr += fmt.Sprintf("P%02x,", record.Value[0])
			break
		case info2.E_Gang:
			ourCardStr += fmt.Sprintf("G%02x,", record.Value[0])
			break
		case info2.E_Gang_HongZhongGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_FaCaiGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_PiziGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_LaiziGand:
			ourCardStr += fmt.Sprintf("M%02x,", record.Value[0])
			break
		case info2.E_Gang_XuGand:
			ourCardStr += fmt.Sprintf("X%02x,", record.Value[0])
			break
		case info2.E_Gang_AnGang:
			ourCardStr += fmt.Sprintf("A%02x,", record.Value[0])
			break
		case info2.E_Qiang:
			ourCardStr += fmt.Sprintf("Q%02x,", record.Value[0])
			break
		case info2.E_Hu:
			ourCardStr += fmt.Sprintf("H%02x,", record.Value[0])
			break
		case info2.E_HuangZhuang:
			ourCardStr += fmt.Sprintf("N%02x,", record.Value[0])
			break
		case info2.E_Bird:
			ourCardStr += fmt.Sprintf("B%02x,", record.Value[0])
			break
		case info2.E_Li_Xian:
			ourCardStr += fmt.Sprintf("l%02x,", record.Value[0])
			break
		case info2.E_Jie_san:
			ourCardStr += fmt.Sprintf("j%02x,", record.Value[0])
			break
		case info2.E_Pao:
			ourCardStr += fmt.Sprintf("K%02x,", record.Value[0])
			break
		case info2.E_SendCardRight:
			ourCardStr += fmt.Sprintf("s%02x,", record.Value[0])
			break
		case info2.E_HandleCardRight:
			ourCardStr += fmt.Sprintf("h%02x,", record.Value[0])
			break
		case info2.E_Gang_ChaoTianGand:
			ourCardStr += fmt.Sprintf("T%02x,", record.Value[0])
			break
		case info2.E_Gang_SmallChaoTianGand:
			ourCardStr += fmt.Sprintf("a%02x,", record.Value[0]) //小朝天
			break
		case info2.E_Baoqing:
			ourCardStr += fmt.Sprintf("q%02x,", record.Value[0])
			break
		case info2.E_Baojing: //报警
			ourCardStr += fmt.Sprintf("c%02x,", record.Value[0])
			break
		case info2.E_Baojiang:
			ourCardStr += fmt.Sprintf("J%02x,", record.Value[0])
			break
		case info2.E_Baofeng:
			ourCardStr += fmt.Sprintf("F%02x,", record.Value[0])
			break
		case info2.E_Baoqi:
			ourCardStr += fmt.Sprintf("D%02x,", record.Value[0])
			break
		case info2.E_BaoTing:
			ourCardStr += fmt.Sprintf("t%02x,", record.Value[0])
			break
		case info2.E_ExchageThreeCard: //换三张换的牌
			ourCardStr += fmt.Sprintf("E%02x,", record.Value[0])
			break
		case info2.E_ExchageThreeEnd: //换三张结束
			ourCardStr += fmt.Sprintf("e%02x,", record.Value[0])
			break
		case info2.E_LastCard: //牌堆最后一张牌
			ourCardStr += fmt.Sprintf("r%02x,", record.Value[0])
			break
		case info2.E_GameScore:
			if fs := strings.Split(fmt.Sprintf("%v", record.UserScore), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s,", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f,", record.UserScore)
			}
		case info2.E_BuHua:
			ourCardStr += fmt.Sprintf("BH%02x,", record.Value[0])
			break
		case info2.E_Liang:
			ourCardStr += fmt.Sprintf("b%02x,", record.Value[0])
			break
		case info2.E_Change_Pizhi:
			ourCardStr += fmt.Sprintf("p%02x,", record.Value[0])
		case info2.E_HaiDiLao:
			ourCardStr += fmt.Sprintf("d%02x,", record.Value[0])
			break
		default:
			break
		}
	}

	//一炮多响的话有多个人胡牌，回放有几个人胡就应该播放几个胡
	if 0 != len(replayRecord.BigHuKindArray) {
		for _, huKind := range replayRecord.BigHuKindArray {
			ourCardStr += getWriteOutReplayBigHuCString(huKind, replayRecord.ProvideUser)
		}
	} else {
		ourCardStr += getWriteOutReplayBigHuCString(replayRecord.BigHuKind, replayRecord.ProvideUser)
	}

	//汉川搓虾子 回放需要显示最后一张牌(其他规则这个未赋值的字段编译器会自动初始化为0，不会走到这个逻辑)
	if replayRecord.LeftCard != 0x00 && replayRecord.LeftCard != 0xff {
		ourCardStr += fmt.Sprintf("%d:", 0)
		ourCardStr += fmt.Sprintf("r%02x,", replayRecord.LeftCard)
	}

	if hasHu {
		// 最后补上胡牌U
		for i, s := range endMsgUpdateScore {
			if i >= gameCommon.GetPlayerCount() {
				break
			}
			ourCardStr += fmt.Sprintf("%d:", i)
			if fs := strings.Split(fmt.Sprintf("%v", s), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s,", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f,", s)
			}
		}
	}
	return ourCardStr
}

func getWriteOutReplayBigHuCString(huKind byte, ProvideUser uint16) string {
	var ourCardStr string
	switch huKind {
	case 11:
		ourCardStr += fmt.Sprintf("JYS%d,", ProvideUser)
		break
	case 4:
		ourCardStr += fmt.Sprintf("QGH%d,", ProvideUser)
		break
	case 13:
		ourCardStr += fmt.Sprintf("QSB%d,", ProvideUser)
		break
	case 5:
		ourCardStr += fmt.Sprintf("HDL%d,", ProvideUser)
		break
	case 6:
		ourCardStr += fmt.Sprintf("FYS%d,", ProvideUser)
		break
	case 7:
		ourCardStr += fmt.Sprintf("GSK%d,", ProvideUser)
		break
	case 8:
		ourCardStr += fmt.Sprintf("QQR%d,", ProvideUser)
		break
	case 9:
		ourCardStr += fmt.Sprintf("PPH%d,", ProvideUser)
		break
	case 10:
		ourCardStr += fmt.Sprintf("QYS%d,", ProvideUser)
		break
	case 12:
		ourCardStr += fmt.Sprintf("MQQ%d,", ProvideUser)
		break
	case 14:
		ourCardStr += fmt.Sprintf("QID%d,", ProvideUser)
		break
	case static.GameNoMagicHu:
		ourCardStr += fmt.Sprintf("YHM%d,", ProvideUser)
		break
	case static.GameYgk:
		ourCardStr += fmt.Sprintf("YGK%d,", ProvideUser)
		break
	case static.GameRgk:
		ourCardStr += fmt.Sprintf("RGK%d,", ProvideUser)
		break
	case static.GameMagicHu:
		ourCardStr += fmt.Sprintf("RHM%d,", ProvideUser)
		break
	case static.GameReChongHu:
		ourCardStr += fmt.Sprintf("RCH%d,", ProvideUser)
		break
	default:
		ourCardStr += fmt.Sprintf("NIL%d,", ProvideUser)
	}
	return ourCardStr
}

func CheckGameFinish(gameCommon *components2.Common, dismiss bool) bool {
	gameFinish := false
	if int(gameCommon.CurCompleteCount) >= gameCommon.Rule.JuShu || dismiss {
		gameFinish = true
	}
	return gameFinish
}

func CheckGameValid(gameCommon *components2.Common) bool {
	if gameCommon.CurCompleteCount < 1 {
		// 第一局没开始不算有效局
		return false
	} else if gameCommon.CurCompleteCount == 1 && gameCommon.GetGameRoundStatus() == static.GS_MJ_PLAY {
		// 第一句没打完不算有效局
		return false
	}
	return true
}

func GetScoreKindByFinish(score int, finish bool) int {
	if finish {
		if score > 0 {
			return static.ScoreKind_Win
		} else {
			return static.ScoreKind_Lost
		}
	} else {
		return static.ScoreKind_pass
	}
}

func GetScoreKindByDismiss(score int, dismiss bool) int {
	if dismiss {
		return static.ScoreKind_pass
	} else {
		if score > 0 {
			return static.ScoreKind_Win
		} else {
			return static.ScoreKind_Lost
		}
	}
}
