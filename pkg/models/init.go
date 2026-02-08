package models

import (
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/jinzhu/gorm"
)

func InitModel(db *gorm.DB) error {
	var err error

	err = initUser(db)
	if err != nil {
		xlog.Logger().Errorln("init db user failed: ", err.Error())
		return err
	}

	err = initUserLoginRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_login_record failed: ", err.Error())
		return err
	}

	err = initGameConfig(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_config failed: ", err.Error())
		return err
	}

	err = initHouse(db)
	if err != nil {
		xlog.Logger().Errorln("init db house failed: ", err.Error())
		return err
	}
	err = initHouseLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_log failed: ", err.Error())
		return err
	}

	err = initHouseFloor(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_floor failed: ", err.Error())
		return err
	}

	err = initHouseMember(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_member failed: ", err.Error())
		return err
	}

	err = initHouseMemberLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_member_log failed: ", err.Error())
		return err
	}

	err = initHouseMemberVitaminLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_member_vitamin_log failed: ", err.Error())
		return err
	}

	err = InitHouseActivity(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_activity failed: ", err.Error())
		return err
	}

	err = InitHouseActivityLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_activity_log failed: ", err.Error())
		return err
	}

	err = InitHouseActRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_activity_record failed: ", err.Error())
		return err
	}

	err = InitHouseActRecordLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_activity_record_log failed: ", err.Error())
		return err
	}

	err = initRecordGameTotal(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_total failed: ", err.Error())
		return err
	}

	err = initRecordGameCost(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_cost failed: ", err.Error())
		return err
	}

	err = initConfig(db)
	if err != nil {
		xlog.Logger().Errorln("init db config failed: ", err.Error())
		return err
	}

	err = initConfigHouse(db)
	if err != nil {
		xlog.Logger().Errorln("init db config_house failed: ", err.Error())
		return err
	}

	err = initRecordGame(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_round failed: ", err)
		return err
	}

	err = initRecordGameReplay(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_outdata failed: ", err)
		return err
	}

	err = initRecordGameDay(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_day_record failed: ", err)
		return err
	}

	err = initGameResultDetail(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_result_detail failed: ", err)
		return err
	}

	err = initGameMatchDetail(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_match_detail failed: ", err)
		return err
	}

	err = initGameMatchTotal(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_match_Total failed: ", err)
		return err
	}

	err = initGameMatchCoupon(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_match_coupon failed: ", err)
		return err
	}

	err = initGameMatchCouponRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_match_coupon_record failed: ", err)
		return err
	}

	err = initUserWealthCost(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_wealth_cost failed: ", err)
		return err
	}

	err = initUserTools(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_tools failed: ", err)
		return err
	}

	//err = initConfigTool(db)
	//if err != nil {
	//	syslog.Logger().Errorln("init db config_tool failed: ", err)
	//	return err
	//}

	err = initGameOnline(db)
	if err != nil {
		xlog.Logger().Errorln("init db game_online failed: ", err)
		return err
	}

	err = initUserBroadcastRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_broadcast_record failed: ", err)
		return err
	}

	err = initUserAllowances(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_allowances failed: ", err)
		return err
	}

	err = initConfigTask(db)
	if err != nil {
		xlog.Logger().Errorln("init db config_task failed: ", err)
		return err
	}

	err = initConfigTaskGame(db)
	if err != nil {
		xlog.Logger().Errorln("init db config_task_game failed: ", err)
		return err
	}

	err = initConfigMatch(db)
	if err != nil {
		xlog.Logger().Errorln("init db config_match failed: ", err)
		return err
	}

	err = initUserTask(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_task failed: ", err)
		return err
	}

	err = initUserTaskRewardLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_task_reward_log failed: ", err)
		return err
	}

	err = initUserTaskGame(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_task_game failed: ", err)
		return err
	}

	err = initUserTaskGameRewardLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_task_game_reward_log failed: ", err)
		return err
	}

	err = initInsureGoldRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db insure_gold_record failed: ", err)
		return err
	}

	err = initStatisticsUser(db)
	if err != nil {
		xlog.Logger().Errorln("init db statistics_user failed: ", err)
		return err
	}

	err = initStatisticsUser(db)
	if err != nil {
		xlog.Logger().Errorln("init db statistics_user failed: ", err)
		return err
	}

	err = initStatisticsGame(db)
	if err != nil {
		xlog.Logger().Error("init db statistics_game failed: ", err)
		return err
	}

	err = initStatisticsUserGameHistory(db)
	if err != nil {
		xlog.Logger().Errorln("init db statistics_user_game_history failed: ", err)
		return err
	}

	err = initConfigApp(db)
	if err != nil {
		xlog.Logger().Errorln("init db config_channel failed: ", err)
		return err
	}

	err = initLeague(db)
	if err != nil {
		xlog.Logger().Errorln("init db league failed: ", err)
		return err
	}

	err = initLeagueUser(db)
	if err != nil {
		xlog.Logger().Errorln("init db league_user failed: ", err)
		return err
	}

	err = initLeagueCardRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db league_card_record failed: ", err)
		return err
	}

	err = initConfigSqlserver(db)
	if err != nil {
		xlog.Logger().Errorln("init db config sqlserver failed: ", err)
		return err
	}

	err = initHouseTableLimitUsers(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_table_limit_user failed: ", err)
		return err
	}

	err = initUserWhite(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_white failed: ", err)
		return err
	}
	err = initHouseUserLimit(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_user_limit failed: ", err)
	}
	err = initHousemixfloorTable(db)
	if err != nil {
		xlog.Logger().Errorln("init db House mix floor Table failed: ", err)
		return err
	}
	err = initConfigGameControl(db)
	if err != nil {
		xlog.Logger().Errorln("init db config game control failed: ", err)
		return err
	}
	err = initHouseValidRound(db)
	if err != nil {
		xlog.Logger().Errorln("init db house valid round failed: ", err)
		return err
	}
	err = initHouseFloorValidRound(db)
	if err != nil {
		xlog.Logger().Errorln("init db house floor valid round failed: ", err)
		return err
	}
	err = initHouseValidRoundLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house valid round log failed: ", err)
		return err
	}
	err = initHouseMsg(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_msg failed: ", err)
		return err
	}

	err = initRecordVitaminDay(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_vitamin_day_record failed: ", err)
		return err
	}

	err = initRecordVitaminDayClear(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_vitamin_day_record_clear failed: ", err)
		return err
	}

	err = initRecordVitaminMgr(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_vitamin_day_mgr failed: ", err)
		return err
	}

	err = initHouseMemberVitaminDay(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_vitamin_day_member failed: ", err)
		return err
	}

	err = initRecordGameRoundBak(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_round_bak failed: ", err)
		return err
	}

	err = initRecordGameCostBak(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_cost_bak failed: ", err)
		return err
	}

	err = initRecordGameDayBak(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_day_record_bak failed: ", err)
		return err
	}

	err = initRecordGameReplayBak(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_outdata_bak failed: ", err)
		return err
	}

	err = initRecordGameTotalBak(db)
	if err != nil {
		xlog.Logger().Errorln("init db record_game_total_bak failed: ", err)
		return err
	}

	// err = initHouseMergeRepeatUsers(db)
	// if err != nil {
	// 	syslog.Logger().Errorln("init db house_merge_repeat_users failed: ", err)
	// 	return err
	// }

	err = initHouseMergeLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_merge_log failed: ", err)
		return err
	}

	err = initHousePartnerInviteCode(db)
	if err != nil {
		xlog.Logger().Errorln("init db house_partner_invitecode failed: ", err)
		return err
	}

	err = InitHousePartnerRoyaltyHistory(db)
	if err != nil {
		xlog.Logger().Errorln("init db InitHousePartnerRoyaltyHistory failed: ", err)
		return err
	}

	err = InitHousePartnerRoyaltyDetail(db)
	if err != nil {
		xlog.Logger().Errorln("init db InitHousePartnerRoyaltyDetail failed: ", err)
		return err
	}

	err = initHouseFloorDelMsg(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseFloorDelMsg failed: ", err)
		return err
	}

	err = initHouseGroupUser(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseGroupUser failed: ", err)
		return err
	}

	err = initHouseFloorVitaminDeduct(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseFloorVitaminDeduct failed: ", err)
		return err
	}

	err = initHousePartnerPyramid(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHousePartnerPyramid failed: ", err)
		return err
	}

	err = initHouseFloorGearPay(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseFloorGearPay failed: ", err)
		return err
	}

	err = initHouseTableDistanceLimit(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseTableDistanceLimit failed: ", err)
		return err
	}

	err = initHouseTableDistanceLimitLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseTableDistanceLimitLog failed: ", err)
		return err
	}

	err = initHouseVipFloorLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseVipFloorLog failed: ", err)
		return err
	}

	err = initHouseRevokeLog(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseRevokeLog failed: ", err)
	}

	err = initHousePartnerAttr(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHousePartnerAttr failed: ", err)
		return err
	}

	err = initHouseRecordLike(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseRecordLike failed: ", err)
		return err
	}

	err = initUserLocation(db)
	if err != nil {
		xlog.Logger().Errorln("init db initUserLocation failed: ", err)
		return err
	}

	err = initHouseExit(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseExit failed: ", err)
		return err
	}

	err = initHouseFloorColor(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseFloorColor failed: ", err)
		return err
	}

	err = initHouseMemberLeaveLink(db)
	if err != nil {
		xlog.Logger().Errorln("init db initHouseMemberLeaveLink failed: ", err)
		return err
	}

	err = initStatisticsUserGameOpt(db)
	if err != nil {
		xlog.Logger().Errorln("init db initStatisticsUserGameOpt failed: ", err)
		return err
	}

	err = initConfigShop(db)
	if err != nil {
		xlog.Logger().Error("init db config_shop failed: ", err)
		return err
	}

	err = initShopRecord(db)
	if err != nil {
		xlog.Logger().Error("init db config_channel failed: ", err)
		return err
	}
	err = initShopPhone(db)
	if err != nil {
		xlog.Logger().Error("init db config_channel failed: ", err)
		return err
	}

	err = initConfigContributionSystem(db)
	if err != nil {
		xlog.Logger().Error("init db config_cntribution_system failed: ", err)
		return err
	}

	err = initDeliverProductRecord(db)
	if err != nil {
		xlog.Logger().Error("init db record_deliver_product failed: ", err)
		return err
	}

	err = initDayAvgScore(db)
	if err != nil {
		xlog.Logger().Error("init db user_day_avgscore failed: ", err)
	}

	err = initCurrencyStatistics(db)
	if err != nil {
		xlog.Logger().Errorln("init db initCurrencyStatistics failed: ", err)
		return err
	}

	err = initUserStatisticsGlod(db)
	if err != nil {
		xlog.Logger().Errorln("init db initUserStatisticsGlod failed: ", err)
		return err
	}

	err = initGoldUser(db)
	if err != nil {
		xlog.Logger().Errorln("init db initUserStatisticsGlod failed: ", err)
		return err
	}

	err = initGoldGameRetentionStatistics(db)
	if err != nil {
		xlog.Logger().Errorln("init db initUserStatisticsGlod failed: ", err)
		return err
	}

	err = initGoldUserLoginRecord(db)
	if err != nil {
		xlog.Logger().Errorln("init db initUserStatisticsGlod failed: ", err)
		return err
	}

	err = initConfigSignIn(db)
	if err != nil {
		xlog.Logger().Errorln("init db initConfigSignIn failed: ", err)
		return err
	}

	err = initUserDailyRewards(db)
	if err != nil {
		xlog.Logger().Errorln("init db initUserDailyRewards failed: ", err)
		return err
	}

	err = initConfigRecommendGame(db)
	if err != nil {
		xlog.Logger().Errorln("init db initConfigGameRecommendGame failed: ", err)
		return err
	}

	err = initConfigShare(db)
	if err != nil {
		xlog.Logger().Errorln("init db initConfigShare failed: ", err)
		return err
	}

	err = initShareHistory(db)
	if err != nil {
		xlog.Logger().Errorln("init db initShareHistory failed: ", err)
		return err
	}

	err = initGameWinRate(db)
	if err != nil {
		xlog.Logger().Errorln("init db initGameWinRate failed: ", err)
		return err
	}

	err = initArea(db)
	if err != nil {
		xlog.Logger().Errorln("init db initArea failed: ", err)
		return err
	}

	err = initHouseMemberRight(db)
	if err != nil {
		xlog.Logger().Errorln("init db HouseMemberRight failed: ", err.Error())
		return err
	}

	err = initHouseMemberUserRight(db)
	if err != nil {
		xlog.Logger().Errorln("init db HouseMemberUserRight failed: ", err.Error())
		return err
	}

	err = initHouseMemberSwitch(db)
	if err != nil {
		xlog.Logger().Errorln("init db HouseMemberSwitch failed: ", err.Error())
		return err
	}

	err = initHousePartnerSetting(db)
	if err != nil {
		xlog.Logger().Errorln("init house partner setting failed: ", err.Error())
		return err
	}

	err = initHouseRank(db)
	if err != nil {
		xlog.Logger().Errorln("init initHouseRank failed: ", err.Error())
		return err
	}

	err = initRecordDismiss(db)
	if err != nil {
		xlog.Logger().Errorln("init initRecordDismiss failed: ", err.Error())
		return err
	}

	err = initRobot(db)
	if err != nil {
		xlog.Logger().Error("init db robot failed: ", err)
		return err
	}

	err = initConfigGovAuth(db)
	if err != nil {
		xlog.Logger().Errorln("init db config_government_auth failed: ", err)
		return err
	}

	err = initDailyPlayTime(db)
	if err != nil {
		xlog.Logger().Errorln("init db user_daily_play_time failed: ", err)
		return err
	}

	err = initPartnerRewardT(db)
	if err != nil {
		xlog.Logger().Errorln("init db partner_reward_t failed: ", err)
		return err
	}

	err = initConfigBattleLevel(db)
	if err != nil {
		xlog.Logger().Errorln("initConfigBattleLevel failed: ", err)
		return err
	}

	err = initConfigSpinBase(db)
	if err != nil {
		xlog.Logger().Errorln("initConfigSpinBase failed: ", err)
		return err
	}

	err = initConfigSpinAward(db)
	if err != nil {
		xlog.Logger().Errorln("initConfigSpinAward failed: ", err)
		return err
	}

	err = initConfigCheckin(db)
	if err != nil {
		xlog.Logger().Errorln("initConfigCheckin failed: ", err)
		return err
	}

	err = initRecordSpinAward(db)
	if err != nil {
		xlog.Logger().Errorln("initRecordSpinAward failed: ", err)
		return err
	}

	err = initRecordCheckin(db)
	if err != nil {
		xlog.Logger().Errorln("initRecordCheckin failed: ", err)
		return err
	}

	err = initPaymentOrder(db)
	if err != nil {
		xlog.Logger().Errorln("initPaymentOrder failed: ", err)
		return err
	}

	return nil
}
