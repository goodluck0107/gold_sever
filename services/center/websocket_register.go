package center

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"io"
	"os"
	"reflect"
	"sync"
	"time"
)

type ProtocolHandler func(*Session, *static.Person, interface{}) (code int16, v interface{})

// 声明全局变量
var (
	// 小程序限定展示游戏的个数
	AppletCardShowGame = []int{460, 408, 479, 597}
	AppletGoldShowGame = []int{1012, 1010, 1019, 1005, 1009, 382, 387}
	AppletCardShowArea = []string{"422801", "422828", "429004"}
)

// ////////////////////////////////////////////////////////////
// ! 协议管理者
type ProtocolInfo struct {
	protoType    reflect.Type
	protoHandler ProtocolHandler
}

// 协议计数map
type ProtocolCntMap struct {
	cntIdx sync.Map
	cntArr []int64
}

type ProtocolMgr struct {
	funcProtocol   map[string]*ProtocolInfo
	protocolCntMap *ProtocolCntMap
}

var protocolMgrSingleton *ProtocolMgr = nil

func init() {
	protocolMgrSingleton = new(ProtocolMgr)
	protocolMgrSingleton.funcProtocol = make(map[string]*ProtocolInfo)
	protocolMgrSingleton.protocolCntMap = new(ProtocolCntMap)
	protocolMgrSingleton.protocolCntMap.cntArr = make([]int64, 0)
	R := protocolMgrSingleton.RegisterMessage
	// 更新个性签名
	R(consts.MsgTypeUpdateDescribeInfo, static.Msg_UpdateDescribeInfo{}, Proto_UpdateUserDescribe)
	// 更新用户信息
	R(consts.MsgTypeUpdateUserInfo, static.Msg_UpdateUserInfo{}, Proto_UpdateUserInfo)
	// 获取区域列表
	R(consts.MsgTypeGetAreas, static.Msg_Null{}, Proto_GetAreas)
	// 加入区域
	R(consts.MsgTypeAreaIn, static.Msg_AreaIn{}, Proto_AreaIn)
	// 选择游戏场次, 进入房间(分配游戏服)
	R(consts.MsgTypeSiteIn, static.Msg_SiteIn{}, Proto_SiteIn)
	// 检查回放码
	R(consts.MsgTypeCheckReplayId, static.Msg_CheckReplayId{}, Proto_CheckReplayId)
	// 历史战绩
	R(consts.MsgTypeHallGameRecord, static.Msg_HallGameRecord{}, Proto_GetHallGameRecord)
	// 战绩详情
	R(consts.MsgTypeGameRecordInfo, static.Msg_GameRecordInfo{}, Proto_GetGameRecordInfo)
	// 保存用户游戏列表
	R(consts.MsgTypeHallSaveGames, static.Msg_SaveGames{}, Proto_SaveGames)
	// 实名认证
	R(consts.MsgTypeHallCertification, static.Msg_Certification{}, Proto_Certification)
	// 解绑手机
	R(consts.MsgTypeHallUnbindMobile, static.Msg_UnbindMobile{}, Proto_UnbindMobile)
	// 绑定微信
	R(consts.MsgTypeHallBindWechat, static.Msg_BindWechat{}, Proto_BindWechat)
	// 授权微信用户信息
	R(consts.MsgTypeHallAuthWechat, static.Msg_AuthWechat{}, Proto_AuthWechat)
	// 绑定手机
	R(consts.MsgTypeHallBindMobile, static.Msg_BindMobile{}, Proto_BindMobile)
	// 绑定手机v2
	R(consts.MsgTypeHallBindMobileV2, static.Msg_BindMobile{}, Proto_BindMobileV2)
	// 创建牌桌
	R(consts.MsgTypeTableCreate, static.Msg_CreateTable{}, Proto_CreateTable)
	// 加入牌桌
	R(consts.MsgTypeTableIn, static.Msg_JoinTable{}, Proto_JoinTable)
	// 解散牌桌
	R(consts.MsgTypeTableDel, static.Msg_TableDel{}, Proto_DeleteTable)
	// 牌桌详情
	R(consts.MsgTypeTableInfo, static.Msg_TableInfo{}, Proto_TableInfo)
	// 根据uid查找包厢桌子信息
	R(consts.MsgTypeHouseTableInfoByUid, static.MsgC2SHidUid{}, Proto_ClubHouseTableInfoByUid)
	// 包厢设置审核
	R(consts.MsgTypeHouseOptIsMemCheck, static.Msg_CH_HouseOptIsMemCheck{}, Proto_ClubHouseOptMemCheck)
	// 包厢设置退圈开关
	R(consts.MsgTypeHouseOptIsMemExit, static.Msg_CH_HouseOptIsMemExit{}, Proto_ClubHouseOptMemExit)
	// 包厢队长审核
	R(consts.MsgTypeHouseOptIsParnterMemCheck, static.Msg_CH_HouseOptIsMemCheck{}, Proto_ClubHouseOptParnterMemCheck)
	// 包厢设置冻结
	R(consts.MsgTypeHouseOptIsFrozen, static.Msg_CH_HouseOptIsFrozen{}, Proto_ClubHouseOptFrozen)
	// 包厢设置成员列表隐藏
	R(consts.MsgTypeHouseOptIsMemHide, static.Msg_CH_HouseOptIsMemHide{}, Proto_ClubHouseOptMemHide)
	// 包厢设置圈号隐藏
	R(consts.MsgTypeHouseOptIsHidHide, static.Msg_CH_HouseOptIsHidHide{}, Proto_ClubHouseOptHidHide)
	// 头像隐藏
	R(consts.MsgTypeHouseOptIsHeadHide, static.Msg_CH_HouseOptIsHeadHide{}, Proto_ClubHouseOptHeadHide)
	// ID隐藏
	R(consts.MsgTypeHouseOptIsMemUidHide, static.Msg_CH_HouseOptIsMemUidHide{}, Proto_ClubHouseOptMemUidHide)
	// 在线人数隐藏
	R(consts.MsgTypeHouseOptIsOnlineHide, static.Msg_CH_HouseOptIsOnlineHide{}, Proto_ClubHouseOptOnlineHide)
	// 队长踢人
	R(consts.MsgTypeHouseOptPartnerKick, static.Msg_CH_HouseOptPartnerKick{}, Proto_ClubHouseOptPartnerKick)
	// 包厢队长权限
	R(consts.MsgTypeHousePartner, static.Msg_Null{}, Proto_ClubHousePartner)
	// 包厢队长列表
	R(consts.MsgTypeHousePartnerList, static.Msg_CH_HouseMemStatistics{}, Proto_ClubHousePartnerList)
	// 包厢队长成员列表
	R(consts.MsgTypeHousePartnerMemCustom, static.Msg_CH_HousePartnerMemCustom{}, Proto_ClubHousePartnerMemCustom)
	// 包厢队长成员列表
	R(consts.MsgTypeHousePartnerBindUser, static.Msg_CH_HousePartnerMemCustom{}, Proto_ClubHousePartnerMemCustom)
	// 包厢队长创建
	R(consts.MsgTypeHousePartnerGreate, static.Msg_CH_HousePartnerCreate{}, Proto_ClubHousePartnerCreate)
	// 包厢队长删除
	R(consts.MsgTypeHousePartnerDelete, static.Msg_CH_HousePartnerDelete{}, Proto_ClubHousePartnerDelete)
	// 包厢队长调配
	R(consts.MsgTypeHousePartnerGen, static.Msg_CH_HousePartnerGen{}, Proto_ClubHousePartnerGen)
	// 包厢疲劳值清空
	R(consts.MsgTypeHouseVitaminClear, static.Msg_CH_HouseVitaminClear{}, Proto_ClubHouseVitaminClear)
	// 包厢比赛统计
	R(consts.MsgTypeHousePartnerVitaminStatistic, static.Msg_CH_HousePartnerVitaminStatistic{}, Proto_ClubHousePartnerVitaminStatistic)
	// 包厢疲劳值统计
	R(consts.MsgTypeHouseVitaminStatistic, static.Msg_CH_HouseVitaminStatistic{}, Proto_ClubHouseVitaminStatistic)
	// 包厢疲劳值统计清零
	R(consts.MsgTypeHouseVitaminStatisticClear, static.Msg_CH_HouseVitaminStatisticClear{}, Proto_ClubHouseVitaminStatisticClear)
	// 包厢疲劳值管理列表
	R(consts.MsgTypeHouseVitaminMgr, static.Msg_CH_HouseVitaminMgrList{}, Proto_ClubHouseVitaminMgrList)
	// 包厢创建
	R(consts.MsgTypeHouseCreate, static.Msg_CH_HouseCreate{}, Proto_ClubHouseCreate)
	// 包厢删除
	R(consts.MsgTypeHouseDelete, static.Msg_CH_HouseDelete{}, Proto_ClubHouseDelete)
	// 包厢楼层创建
	R(consts.MsgTypeHouseFloorCreate, static.Msg_CH_HouseFloorCreate{}, Proto_ClubHouseFloorCreate)
	// 包厢楼层删除
	R(consts.MsgTypeHouseFloorDelete, static.Msg_CH_HouseFloorDelete{}, Proto_ClubHouseFloorDelete)
	// 包厢楼层列表
	R(consts.MsgTypeHouseFloorList, static.Msg_CH_HouseFloorList{}, Proto_ClubHouseFloorList)
	// 包厢成员搜索列表
	R(consts.MsgTypeHouseMemberList, static.Msg_CH_HouseMemList{}, Proto_ClubHouseMemberList)

	// 包厢添加队长列表
	R(consts.MsgTypeHousepartnerAddList, static.Msg_CH_HouseMemList{}, Proto_ClubHousePartnerAddList)
	// 包厢玩家加入
	R(consts.MsgTypeHouseMemberJoin, static.Msg_CH_HouseMemberJoin{}, Proto_ClubHouseMemberJoin)
	// 包厢玩家过审
	R(consts.MsgTypeHouseMemberAgree, static.Msg_CH_HouseMemberAgree{}, Proto_ClubHouseMemberAgree)
	// 包厢玩家拒审
	R(consts.MsgTypeHouseMemberRefused, static.Msg_CH_HouseMemberRefused{}, Proto_ClubHouseMemberRefused)
	// 包厢退出
	R(consts.MsgTypeHouseMemberExit, static.Msg_CH_HouseMemberExit{}, Proto_ClubHouseMemberExit)
	// 包厢玩家剔除
	R(consts.MsgTypeHouseMemberKick, static.Msg_CH_HouseMemberKick{}, Proto_ClubHouseMemberkick)
	// 加入黑名单
	R(consts.MsgTypeHouseMemberblacklistInsert, static.Msg_CH_HouseMemberBlacklistInsert{}, Proto_ClubHouseMemberBlacklistInsert)
	// 删除黑名单
	R(consts.MsgTypeHouseMemberblacklistDelete, static.Msg_CH_HouseMemberBlacklistDelete{}, Proto_ClubHouseMemberBlacklistDelete)
	// 包厢玩家备注
	R(consts.MsgTypeHouseMemberRemark, static.Msg_CH_HouseMemberRemark{}, Proto_ClubHouseMemberRemark)
	// 包厢玩家设置角色
	R(consts.MsgTypeHouseMemberRoleGen, static.Msg_CH_HouseMemberRoleGen{}, Proto_ClubHouseMemberRoleGen)
	// 玩家进入楼层
	R(consts.MsgTypeHouseMemberIn, static.Msg_CH_HouseMemberIn{}, Proto_ClubHouseMemberIn)
	// 玩家离开楼层
	R(consts.MsgTypeHouseMemberOut, static.Msg_CH_HouseMemberOut{}, Proto_ClubHouseMemberOut)
	// 玩家包厢列表
	R(consts.MsgTypeMemberHouseList, static.Msg_CH_MemberHouseList{}, Proto_MemberHouseList)
	// 包厢基础信息
	R(consts.MsgTypeHouseBaseInfo, static.Msg_CH_HouseBaseInfo{}, Proto_ClubHouseBaseInfo)
	// 包厢在线人数
	R(consts.MsgTypeHouseMemberOnline, static.Msg_CH_HouseMemOnline{}, Proto_ClubHouseMemOnline)
	// 包厢常玩
	R(consts.MsgTypeHousePlayOften, static.Msg_CH_HouseBaseInfo{}, Proto_ClubHouseBaseInfo)
	// 包厢修改名称公告
	R(consts.MsgTypeHouseBaseNNModify, static.Msg_CH_HouseBaseNNModify{}, Proto_ClubHouseBaseNNModify)
	// 修改包厢楼层基本信息
	R(consts.MsgTypeHouseFloorRuleModify, static.Msg_CH_HouseFloorRuleModify{}, Proto_ClubHouseFloorRuleModify)
	// 包厢牌桌加入
	R(consts.MsgTypeHouseTableIn, static.Msg_CH_HouseTableIn{}, Proto_ClubHouseTableIn)
	// 包厢成员统计
	R(consts.MsgTypeMemberStatistics, static.Msg_CH_HouseMemStatistics{}, Proto_ClubHouseMemStatistics)
	// 包厢成员统计总计
	R(consts.MsgTypeMemberStatisticsTotal, static.Msg_CH_HouseMemStatisticsTotal{}, Proto_ClubHouseMemStatisticsTotal)
	// 包厢我的战绩查询
	R(consts.MsgTypeHouseMyRecord, static.Msg_CH_HouseMyRecord{}, Proto_ClubHouseMyRecord)
	// 包厢战绩查询
	R(consts.MsgTypeHouseRecord, static.Msg_CH_HouseRecord{}, Proto_ClubHouseRecord)
	// 包厢战绩点赞
	R(consts.MsgTypeHouseRecordHeart, static.Msg_CH_HouseRecordHeart{}, Proto_ClubHouseRecordHeart)
	// 包厢房卡统计
	R(consts.MsgTypeHouseRecordKaCost, static.Msg_CH_HouseRecordKaCost{}, Proto_ClubHouseRecordKaCost)
	// 包厢大赢家统计查询
	R(consts.MsgTypeHouseRecordStatus, static.Msg_CH_HouseRecordStatus{}, Proto_ClubHouseRecordStatus)
	// 包厢大赢家统计清除
	R(consts.MsgTypeHouseRecordStatusClean, static.Msg_CH_HouseRecordStatusClean{}, Proto_ClubHouseRecordStatusClean)
	// 包厢大赢家统计清除
	R(consts.MsgTypeHouseRecordStatusCleanAll, static.Msg_CH_HouseRecordStatusCleanAll{}, Proto_ClubHouseRecordStatusCleanAll)
	// 包厢游戏战绩(新版本)
	R(consts.MsgTypeHouseGameRecord, static.Msg_CH_HouseGameRecord{}, Proto_ClubHouseGameRecord)
	// 包厢游戏战绩(新版本)
	R(consts.MsgTypeHouseOperationalStatus, static.Msg_CH_HouseOperationalStatus{}, Proto_ClubHouseOperationalStatus)
	// 包厢有效对局积分查询
	R(consts.MsgTypeHouseValidRoundScoreGet, static.Msg_CH_HouseValidRoundScoreGet{}, Proto_ClubHouseValidRoundScoreGet)
	// 包厢有效对局积分设置
	R(consts.MsgTypeHouseValidRoundScoreSet, static.Msg_CH_HouseValidRoundScoreSet{}, Proto_ClubHouseValidRoundScoreSet)
	// 包厢楼层活动创建
	R(consts.MsgTypeHouseActivityCreate, static.Msg_CH_HouseActCreate{}, Proto_ClubHouseActCreate)
	// 包厢楼层活动删除
	R(consts.MsgTypeHouseActivityDelete, static.Msg_CH_HouseActDelete{}, Proto_ClubHouseActDelete)
	// 包厢楼层活动列表
	R(consts.MsgTypeHouseActivityList, static.Msg_CH_HouseActList{}, Proto_ClubHouseActList)
	// 包厢楼层活动详情
	R(consts.MsgTypeHouseActivityInfo, static.Msg_CH_HouseActInfo{}, Proto_ClubHouseActInfo)
	// 包厢游戏列表拉取
	R(consts.MsgTypeHouseAreaGameList, static.Msg_CH_HouseActList{}, Proto_ClubHouseAreaGames)
	// // 排位赛列表
	// R(constant.MsgTypeMatchList, public.Msg_Match_MatchList{}, Proto_MatchList)
	// // 排位赛排位列表
	// R(constant.MsgTypeMatchRankingList, public.Msg_Match_RankingList{}, Proto_MatchRankingList)
	// // 排位赛奖励列表
	// R(constant.MsgTypeMatchAwardList, public.Msg_Match_AwardList{}, Proto_MatchAwardList)
	R(consts.MsgTypeDialogNoticeList, static.Msg_Null{}, Proto_SendDialogNoticeList)
	// chess进入区域
	R(consts.MsgTypeAreaEnter, static.Msg_AreaIn{}, Proto_AreaEnter)
	// chess搜索游戏
	R(consts.MsgTypeAreaGameSeek, static.Msg_CH_AreaGameSeek{}, Proto_AreaPackageSeek)
	// 根据包名 得到子游戏
	R(consts.MsgTypeAreaPkgGames, static.Msg_CH_AreaGamesByPkg{}, Proto_AreaGamesByPackageKey)
	// 获取区域内房卡游戏包
	R(consts.MsgTypeAreaGameCardListMain, static.Msg_Null{}, Proto_AreaPackageGameCardListMain)
	// 兼容旧客户协议
	R("areagamemain", static.Msg_Null{}, Proto_AreaPackageGameCardListMain)
	// 获取区域内金币游戏包
	R(consts.MsgTypeAreaGameGoldListMain, static.Msg_Null{}, Proto_AreaPackageGameGoldListMain)
	// 编辑禁止同桌列表
	R(consts.MsgTypeHouseMemberTableLimit, static.MsgHouseMemberTableLimit{}, Proto_ClubHouseMemberTableLimitList)
	// 添加禁止同桌分组
	R(consts.MsgTypeHouseTableLimitGroupAdd, static.MsgHouseTableLimitGroupAdd{}, Proto_ClubHouseTableLimitGroupAdd)
	// 移除禁止同桌分组
	R(consts.MsgTypeHouseTableLimitGroupRemove, static.MsgHouseTableLimitGroupRemove{}, Proto_ClubHouseTableLimitGroupRemove)
	// 添加禁止同桌用户
	R(consts.MsgTypeHouseTableLimitUserAdd, static.MsgHouseTableLimitUserAdd{}, Proto_ClubHouseTableLimitUserAdd)
	// 移除禁止同桌用户
	R(consts.MsgTypeHouseTableLimitUserRemove, static.MsgHouseTableLimitUserAdd{}, Proto_ClubHouseTableLimitUserRemove)
	// 禁止同桌信息
	R(consts.MsgTypeHouseTableLimitInfo, static.MsgHouseTableLimitGroupAdd{}, Proto_ClubHouseTableLimitInfo)
	// 2人桌子禁止同桌不生效 设置是否勾选
	R(consts.MsgHouse2PTableLimitNotEffect, static.MsgHouse2PTableLimitNotEffectSet{}, Proto_ClubHouse2PTableLimitNotEffect)

	// 根据桌子号，得到桌子所在的包信息
	R(consts.MsgTypeAreaPkgByTId, static.Msg_TableDel{}, Proto_AreaPackageByTId)

	R(consts.MsgTypeAreaPkgByKId, static.Msg_HG_UpdateGameServer{}, Proto_AreaPackageByKId)
	// 修改楼层名
	R(consts.MsgTypeHouseFloorRename, static.MsgHouseFloorRename{}, Proto_FloorRename)
	// 编辑包厢混排信息
	R(consts.MsgTypeHouseMixFloorEditor, static.MsgHouseMixFloor{}, Proto_ClubHouseMixEditor)
	// 获取包厢混排信息
	R(consts.MsgTYpeHouseMixFloorInfo, static.MsgHouseMixInfo{}, Proto_ClubHouseMixInfo)
	// 增加包厢桌子
	R(consts.MsgTYpeHouseMixFloorTableCreate, static.MsgHouseMixFloorTableCreate{}, Proto_ClubHouseMixFloorTableCreate)
	// 减少包厢桌子
	R(consts.MsgTYpeHouseMixFloorTableDelete, static.MsgHouseMixFloorTableCreate{}, Proto_ClubHouseMixFloorTableDelete)
	// 批量增删包厢桌子
	R(consts.MsgTYpeHouseMixFloorTableChange, static.MsgHouseMixFloorTableChange{}, Proto_ClubHouseTableChange)
	// 包厢消息
	R(consts.MsgTypeHouseMsg, static.MsgHouseMsg{}, Proto_ClubHouseMsg)
	// 包厢入桌邀请响应
	R(consts.MsgTypeHouseTableInviteResp, static.MsgHTableInvitrResp{}, Proto_ClubHouseTableInviteAck)
	// 禁止玩家游戏
	R(consts.MsgTypeLimitUserPlay, static.MsgHouseUserLimitGame{}, Proto_UserLimitGame)
	// 得到客服微信信息
	R(consts.MsgTypeAreaWX, static.Msg_Null{}, Proto_GetCSWX)

	// 疲劳值
	// 包厢疲劳值修改记录
	R(consts.MsgTypeHouseVitaminSetRecords, static.Msg_CH_HouseVitaminSetRecords{}, Proto_ClubHouseVitaminSetRecords)
	// 玩家间赠送疲劳值
	R(consts.MsgVitaminSend, static.Msg_CH_HouseVitaminSet{}, Proto_ClubHouseVitaminSend)
	R(consts.MsgHouseMemberGetById, static.MsgHouseUserLimitGame{}, Proto_ProtoGetHmemByid)
	// 管理存取疲劳值
	R(consts.MsgHouseVitaminPoolAdd, static.MsgVitaminSend{}, Proto_ClubHouseVitaminPoolAdd)
	// 疲劳值仓库记录
	R(consts.MsgHouseVitaminPoolLog, static.Msg_CH_HouseVitaminSetRecords{}, Proto_ClubHouseVitaminPoolLog)
	// 用户疲劳值修改
	R(consts.MsgTypeHouseVitaminSet, static.Msg_CH_HouseVitaminSet{}, Proto_ClubHouseVitaminSet)

	// 包厢防沉迷获取
	R(consts.MsgTypeHouseVitaminInfo, static.Msg_CH_HouseId{}, Proto_GetHouseVitaminInfo)
	// 包厢防沉迷设置
	R(consts.MsgTypeHouseVitaminValues, static.Msg_CH_HouseVitaminValues{}, Proto_ClubHouseVitaminValues)
	// // 包厢楼层疲劳值扣除相关获取
	// R(constant.MsgTypeHouseFloorVitaminDeductGet, public.Msg_CH_HouseId{}, Proto_GetHFVitaminDeductValues)
	// // 包厢楼层疲劳值扣除相关设置
	// R(constant.MsgTypeHouseFloorVitaminDeductSet, public.MsgHouseFloorDeductInfo{}, Proto_SetHFVitaminDeductValues)
	// 包厢楼层疲劳值生效相关获取
	R(consts.MsgTypeHouseFloorVitaminEffectGet, static.Msg_CH_HouseId{}, Proto_GetHFVitaminEffectValues)
	// 包厢楼层疲劳值生效相关设置
	R(consts.MsgTypeHouseFloorVitaminEffectSet, static.MsgHouseFloorEffectInfo{}, Proto_SetHFVitaminEffectValues)

	// 合并包厢相关
	// 意向
	R(consts.MsgTypeHouseMergeIntention, static.Msg_CH_HouseOwnerInfo{}, Proto_ClubHouseMergeIntention)
	// 检查
	R(consts.MsgTypeHouseMergeCheck, static.Msg_CH_HouseId{}, Proto_ClubHouseMergeCheck)
	// 发起
	R(consts.MsgTypeHouseMergeRequest, static.Msg_CH_HouseId{}, Proto_ClubHouseMergeRequest)
	// 撤销
	R(consts.MsgTypeHouseMergeReqRevoke, static.Msg_CH_HouseMerge{}, Proto_ClubHouseMergeReqRevoke)
	// 响应
	R(consts.MsgTypeHouseMergeResponse, static.Msg_CH_HouseMergeRsp{}, Proto_ClubHouseMergeResponse)
	// 记录
	R(consts.MsgTypeHouseMergeRecord, static.Msg_CH_HouseId{}, Proto_ClubHouseMergeRecord)
	// 撤销合并包厢请求
	R(consts.MsgTypeHouseRevokeRequest, static.Msg_CH_HouseMerge{}, Proto_ClubHouseRevokeRequest)
	// 撤销合并包厢响应
	R(consts.MsgTypeHouseRevokeResponse, static.Msg_CH_HouseMergeRsp{}, Proto_ClubHouseRevokeResonse)

	// 包厢禁止快速入桌设置
	R(consts.MsgHouseJoinTableSet, static.MsgHouseJoinTableSet{}, Proto_ClubHouseJoinSet)

	// 队长分层
	// 分层配置
	R(consts.MsgHouseParnterRoyaltySet, static.Msg_CH_HouseParnterRoyaltySet{}, ProtoHousePartnerRoyaltySet)
	R(consts.MsgHouseParnterRoyaltyGet, static.Msg_CH_HouseParnterRoyaltyGet{}, ProtoHousePartnerRoyaltyGet)
	// 盟主得到给一级队长分成信息
	R(consts.MsgHouseOwnerRoyaltySet, static.Msg_CH_HouseOwnerRoyaltySet{}, ProtoHouseOwnerRoyaltySet)
	R(consts.MsgHouseOwnerRoyaltyGet, static.Msg_CH_HouseOwnerRoyaltyGet{}, ProtoHouseOwnerRoyaltyGet)

	R(consts.MsgHouseParnterSuperiorList, static.Msg_CH_HouseParnterSuperiorList{}, ProtoHousePartnerSuperiorList)
	R(consts.MsgHouseParnterBindSuperior, static.Msg_CH_HouseParnterBindSuperior{}, ProtoHouseBindPartnerSuperior)
	R(consts.MsgHouseParnterBindJunior, static.Msg_CH_HouseParnterBindJunior{}, ProtoHouseBindPartnerJunior)

	// 队长结算统计
	R(consts.MsgHouseParnterFloorStatistics, static.Msg_CH_MsgHouseParnterFloorStatistics{}, ProtoHousePartnerFloorStatistics)
	R(consts.MsgHouseParnterFloorJuniorStatistics, static.Msg_CH_MsgHouseParnterFloorStatistics{}, ProtoHousePartnerFloorJuniorStatistics)
	R(consts.MsgHouseParnterFloorMemStatistics, static.Msg_CH_MsgHouseParnterFloorMemStatistics{}, ProtoHousePartnerFloorMedStatistics)

	R(consts.MsgHousePartnerHistoryFloorStatistics, static.Msg_CH_MsgHousePartnerHistoryFloorStatistics{}, ProtoHousePartnerFloorHistoryStatistics)
	R(consts.MsgHousePartnerHistoryFloorDetailStatistics, static.Msg_CH_MsgHousePartnerHistoryFloorDetailStatistics{}, ProtoHousePartnerFloorHistoryDetailStatistics)

	// 队长疲劳值自动划账
	R(consts.MsgHouseAutoPayPartner, static.Msg_HouseAutoPayPartner{}, ProtoHouseAutoPayPartner)

	// 包厢邀请入圈
	R(consts.MsgTypeHouseJoinInviteSend, static.Msg_CH_HouseJoinInvite{}, ProtoHouseInviteJoinReq)
	R(consts.MsgTypeHouseJoinInviteResp, static.Msg_CH_HouseJoinInviteRsp{}, ProtoHouseInviteJoinRsp)
	// 包厢弹窗
	R(consts.MsgTypeHouseDialogEdit, static.Msg_CH_HouseDialogEdit{}, ProtoHouseDialogEdit)
	// 包厢队长得到邀请码
	R(consts.MsgTypeHousePartnerGetCode, static.Msg_CH_HouseId{}, ProtoGetHousePartnerInviteCode)
	// 根据队长邀请码加入包厢
	R(consts.MsgTypeHouseJoinByPCode, static.Msg_HC_HouseJoinByCode{}, ProtoHouseJoinByInviteCode)
	// 修改防作弊包厢显示桌数
	R(consts.MsgTypeHouseSetTblShowCount, static.Msg_CH_HouseTableShowCount{}, ProtoHouseTblShowCountEdit)

	// 修改包厢楼层超级防作弊等待人数
	R(consts.MsgTypeHouseFloorWaitNumSet, static.Msg_CH_HouseFloorWaitingNum{}, ProtoHouseFloorWaitingNumSet)
	// 获取包厢楼层超级防作弊等待人数
	R(consts.MsgTypeHouseFloorWaitNumGet, static.Msg_CH_HouseId{}, ProtoHouseFloorWaitingNumGet)

	// 包厢设置比赛分管理员
	R(consts.MsgTypeHouseVitaminAdminSet, static.MsgHouseVitaminAdminSet{}, ProtoHouseVitaminSet)
	R(consts.MsgTypeHouseVicePartnerSet, static.MsgHouseVicePartnerSet{}, ProtoHouseVicePartnerSet)

	// 玩家分组
	R(consts.MsgTypeHouseGroupAdd, static.HouseGroupAdd{}, ProtoHouseGroupAdd)
	R(consts.MsgTypeHouseGroupDel, static.HouseGroupDel{}, ProtoHouseGroupDel)
	R(consts.MsgTypeHouseGroupInfo, static.HouseGroupAdd{}, ProtoHouseUserGroupInfo)
	R(consts.MsgTypeHouseGroupUserList, static.HouseGroupUserList{}, ProtoHouseGroupUserList)
	R(consts.MsgTypeHouseGroupUserAddList, static.HouseGroupUserList{}, ProtoHouseGroupUserAddList)
	R(consts.MsgTypeGroupUserAdd, static.HouseGroupUserAdd{}, ProtoHouseUserGroupAddUser)
	R(consts.MsgTypeGroupUserRemove, static.HouseGroupUserAdd{}, ProtoHouseUserGroupDelUser)
	// 队长我的配置
	R(consts.MsgPartnerRoyaltyForMe, static.Msg_CH_HouseParnterRoyaltyForMe{}, Proto_ClubHousePartnerRoyaltyForMe)
	R(consts.MsgPartnerRoyaltyHistory, static.Msg_CH_HouseParnterRoyaltyHistory{}, Proto_ClubHousePartnerRoyaltyHistory)

	// 无盟模式
	R(consts.MsgNoLeagueStatistics, static.Msg_CH_HouseNoLeagueStatistics{}, Proto_ClubHouseNoLeagueStatistics)
	R(consts.MsgNoLeagueDetailStatistics, static.Msg_CH_HouseNoLeagueDetailStatistics{}, Proto_ClubHouseNoLeagueDetailStatistics)

	// 比赛场
	R(consts.MsgGameSwitch, static.MsgHouseUserGameSwitch{}, Proto_ClubHouseGameSwitch)
	R(consts.MsgHousePrizeInfo, static.MsgHousePrizeInfo{}, Proto_ClubHousePrizeInfo)
	R(consts.MsgHousePrizeSet, static.MsgHousePrizeSetS{}, Proto_ClubHousePrizeSet)

	// 房费 挡位 配置
	R(consts.MsgHouseFloorGearPayGet, static.Msg_CH_HouseId{}, Proto_ClubHouseFloorGearPayGet)
	R(consts.MsgHouseFloorGearPaySet, static.MsgHouseFloorPay{}, Proto_ClubHouseFloorGearPaySet)

	// 抽奖
	R(consts.MsgHouseLuckSet, static.HouseLuckSet{}, Proto_ClubHouseLuckSet)
	// R(constant.MsgHouseLuckInfo, public.HouseLuckInfo{}, HouseLuckSet)
	R(consts.MsgHouseMemberCanLuck, static.HouseMemLuck{}, Proto_ClubHouseMemLuckCheck)
	R(consts.MsgHouseMemberGetLuck, static.HouseMemLuck{}, Proto_ClubHouseMemGetLuck)
	R(consts.MsgHouseActDetail, static.HouseMemLuck{}, Proto_ClubHouseActDetail)
	// vip楼层 vip玩家
	R(consts.MsgTypeHouseVipFloorGet, static.Msg_CH_HouseId{}, Proto_ClubHouseVipFloorGet)
	R(consts.MsgTypeHouseVipFloorSet, static.MsgHouseFloorVip{}, Proto_ClubHouseVipFloorSet)
	R(consts.MsgTypeHouseFloorVipUsersGet, static.Msg_CH_HouseFloorVipUser{}, Proto_ClubHouseFloorVipUsersListGet)
	R(consts.MsgTypeHouseFloorEverymanGet, static.Msg_CH_HouseFloorVipUser{}, Proto_ClubHouseFloorVipUsersListGet)
	R(consts.MsgTypeHouseFloorSetVipUser, static.Msg_CH_HouseFloorVipUserSet{}, Proto_ClubHouseFloorVipUserSet)
	R(consts.MsgTypeHouseFloorSetAllVipUser, static.Msg_CH_HouseFloorVipUserAllSet{}, Proto_ClubHouseFloorVipUserAllSet)

	// 包厢入桌距离限制
	R(consts.MsgHouseTableDistanceLimitGet, static.Msg_CH_TableDistanceLimitGet{}, Proto_ClubHouseTableDistanceLimitGet)
	R(consts.MsgHouseTableDistanceLimitSet, static.Msg_CH_TableDistanceLimitSet{}, Proto_ClubHouseTableDistanceLimitSet)

	// 包厢申请开关
	R(consts.MsgHouseApplySwitch, static.MsgHouseApplySwitchSet{}, Proto_ClubHouseApplySwitch)

	R(consts.MsgHouseFloorHideImg, static.MsgHouseFloorHideImg{}, Proto_ClubHouseFloorHideImg)
	R(consts.MsgHouseFloorFakeTable, static.MsgHouseFloorFakeTable{}, Proto_ClubHouseFloorFakeTable)

	// 包厢战绩点赞
	R(consts.MsgHouseRecordGameLike, static.MsgHouseRecordGameLike{}, Proto_ClubHouseRecordGameLike)
	// 包厢成员点赞
	R(consts.MsgHouseRecordUserLike, static.MsgHouseRecordUserLike{}, Proto_ClubHouseRecordUserLike)

	// 包厢申请信息
	R(consts.MsgTypeHouseApplyInfo, static.Msg_CH_HouseApplyInfo{}, Proto_ClubHouseApplyInfo)
	// 包厢成员小红点
	R(consts.MsgTypeHouseMemberTrackPoint, static.Msg_CH_HouseId{}, Proto_ClubHouseMemberTrackPoint)
	// 包厢楼层颜色设置
	R(consts.MsgTypeHouseFloorColorSet, static.Msg_CH_HouseFloorColorSet{}, Proto_ClubHouseFloorColorSet)
	// 牌桌踢人
	R(consts.MsgTypeTableUserKick, static.Msg_TableUserKick{}, Proto_TableUserKick)

	// 任务系统
	// 任务列表
	R(consts.MsgTypeTaskList, static.Msg_CH_TaskList{}, Proto_TaskList)
	// 任务数据提交
	R(consts.MsgTypeTaskCommit, static.Msg_CH_TaskCommit{}, Proto_TaskCommit)
	// 任务奖励领取
	R(consts.MsgTypeTaskReward, static.Msg_CH_TaskReward{}, Proto_TaskReward)
	// 用户每日签到
	//R(constant.MsgTypeCheckin, public.Msg_Checkin{}, p.Checkin)
	// 用户领取见面礼
	//R(constant.MsgTypeGetWelcomeGift, public.Msg_Null{}, p.GetWelcomeGift)

	// 兑换系统
	// 获取兑换商品列表
	R(consts.MsgTypeGetShopLists, static.Msg_Shop_Product{}, Proto_ShopProduct)
	// 兑换商品
	R(consts.MsgTypeShopExchange, static.Msg_Shop_Exchange{}, Proto_ShopExchange)
	// 获取兑换记录列表
	R(consts.MsgTypeGetShopRecord, static.Msg_Shop_Record{}, Proto_ShopRecord)
	// 获取礼卷奖励列表
	//R(constant.MsgTypeGetGoldRecord, public.Msg_Shop_GoldRecord{}, Proto_GoldRecord)
	// 绑定兑换手机
	R(consts.MsgTypePhoneBind, static.Msg_Shop_PhoneBind{}, Proto_ShopPhoneBind)
	// 获取绑定兑换手机
	R(consts.MsgTypeGetPhoneBind, static.Msg_Shop_GetPhoneBind{}, Proto_GetShopPhoneBind)

	// 场次信息
	R(consts.MsgTypeGameSiteList, static.Msg_Game_SiteList{}, Proto_GameSiteList)
	// 金币场列表
	R(consts.MsgTypeGoldGameList, static.Msg_Null{}, Proto_GoldGameList)
	// 进入金币场
	R(consts.MsgEnterGoldGame, static.Msg_Null{}, Proto_EnterGoldGame)

	// 用户领取每日礼包
	R(consts.MsgTypeDailyRewardReq, static.Msg_Null{}, Proto_GetDailyRewards)
	// 用户领取双倍低保
	R(consts.MsgTypeGetAllowancesDouble, static.Msg_GetAllowancesDouble{}, Proto_DisposeAllowancesDouble)

	// 队长修改成员备注
	R(consts.MsgTypePartnerRemark, static.Msg_PartnerRemark{}, Proto_PartnerRemark)

	// 获取隐藏地理位置开关
	R(consts.MsgHousePrivateGPSGet, static.Msg_CH_HouseId{}, Proto_ClubHousePrivateGPSGet)
	// 设置隐藏地理位置开关
	R(consts.MsgHousePrivateGPSSet, static.Msg_CH_HousePrivateGPSSet{}, Proto_ClubHousePrivateGPSSet)
	// 金币场推荐游戏获取
	R(consts.MsgRecommendGoldGameGet, static.Msg_Null{}, Proto_RecommendGoldGameGetmeGet)

	// 获取签到信息
	//R(constant.MsgTypeCheckinInfo, public.Msg_Null{}, p.CheckinInfo)
	// 设置房卡低于xx时提示盟主
	R(consts.MsgTypeSetFangKaTipsMinNumReq, static.Msg_C2S_SetFangKaTipsMinNum{}, Proto_SetFangKaTipsMinNum)
	// 禁止队长游戏 或 队长和成员
	R(consts.MsgTypeLimitCaptainPlay, static.MsgHouseCaptainLimitGame{}, Proto_HouseCaptainLimitGame)
	// 搜索玩家
	R(consts.MsgHallSearchUser, static.MsgC2SHidUid{}, Proto_HallSearchUser)
	// 手动添加玩家进入包厢
	R(consts.MsgHouseMtAddUser, static.MsgC2SHidUid{}, Proto_ClubHouseMtAddUser)
	// 接受或拒绝 被邀请加入包厢 开关
	R(consts.MsgUserRefuseInvite, static.MsgC2SRefuseInvite{}, Proto_UserRefuseInvite)
	//查询当前玩家权限
	R(consts.MsgTypeHmLookUserRight, static.MsgC2SHidUid{}, Proto_HmLookUserRight)
	//修改当前玩家权限
	R(consts.MsgTypeHmUpdateUserRight, static.MsgC2SUpdateHmUright{}, Proto_HmUpdateUserRight)

	//修改包厢功能开关
	R(consts.MsgTypeHmSetSwitch, static.MsgC2SHmSwitch{}, Proto_HmSetSwitch)
	// 设置战绩筛选时段
	R(consts.MsgSetRecordTimeIntervalReq, static.Msg_C2S_SetRecordTimeInterval{}, Proto_SetRecordTimeInterval)
	// 排行榜设置
	R(consts.MsgTypeHouseRankSet, static.Msg_CH_HouseRankSet{}, Proto_ClubHouseRankSet)
	R(consts.MsgTypeHouseRankGet, static.Msg_CH_HouseRankGet{}, Proto_ClubHouseRankGet)
	//获取排行榜数据
	R(consts.MsgTypeHouseRankInfoGet, static.Msg_CH_HouseRankInfoGet{}, Proto_ClubHouseRankInfoGet)
	// 包厢打烊 或者 小队打烊
	R(consts.MsgTypeOffWork, static.MsgOffWork{}, Proto_ClubHouseOffWork)
	// 玩家修改性别
	R(consts.MsgUserSetSex, static.MsgC2SUserSetSex{}, Proto_UserSetSex)
	// 获取包厢中途解散的房间的解散详情
	R(consts.MsgTHDismissRoomDet, static.MsgC2STHDismissRoomDet{}, Proto_GetTHDismissRoomDet)
	// 保存客户端传递过来的定位信息
	R(consts.MsgGpsToSvr, static.Msg_C2S_GpsInfo{}, Proto_SaveUserGps)

	R(consts.MsgTypeHouseTeamBan, static.Msg_CH_HouseParterBan{}, Proto_ClubHouseTeamBan)

	R(consts.MsgTypePartnerAlarmValueSet, static.Msg_CH_HouseParterAlarmValueSet{}, Proto_ClubHousePartnerAlarmValueSet)

	R(consts.MsgTypePartnerAlarmValueGet, static.Msg_CH_HouseParterAlarmValueSet{}, Proto_ClubHousePartnerAlarmValueGet)

	R(consts.MsgTypePartnerAACostSet, static.Msg_CH_HouseParterAACostSet{}, Proto_ClubHousePartnerAACostSet)

	R(consts.MsgTypeHouseTeamKick, static.Msg_CH_HouseMemberKick{}, Proto_ClubHouseTeamKick)

	R(consts.MsgTypeHouseCardStatistics, static.Msg_CH_HouseCardStatistics{}, Proto_ClubHouseCardStatistics)

	R(consts.MsgTypePartnerRewardGet, static.Msg_CH_HouseParterRawardSet{}, Proto_ClubHousePartnerRewardGet)

	R(consts.MsgTypePartnerRewardSet, static.Msg_CH_HouseParterRawardSet{}, Proto_ClubHousePartnerRewardSet)

	R(consts.MsgTypeHouseRewardStatistics, static.Msg_CH_HousePartnerRewardStatistic{}, Proto_ClubHousePartnerRewardStatistic)

	R(consts.MsgTypeHouseClearMyReward, static.Msg_CH_HouseId{}, Proto_ClubPartnerClearReward)

	R(consts.MsgTypeHouseMemberNoFloorsGet, static.Msg_CH_HouseMemberNoFloorsSet{}, Proto_ClubHouseMemberNoFloorsGet)

	R(consts.MsgTypeHouseMemberNoFloorsSet, static.Msg_CH_HouseMemberNoFloorsSet{}, Proto_ClubHouseMemberNoFloorsSet)

	// 新增
	R(consts.MsgTypeUserBattleLevel, static.Msg_Null{}, Proto_BattleLevel)
	R(consts.MsgTypeActivitySpinInfo, static.Msg_Null{}, Proto_ActivitySpinInfo)
	R(consts.MsgTypeActivitySpinDo, static.Msg_Null{}, Proto_ActivitySpinDo)
	R(consts.MsgTypeActivityCheckinInfo, static.Msg_Null{}, Proto_ActivityCheckinInfo)
	R(consts.MsgTypeActivityCheckinDo, static.Msg_Null{}, Proto_ActivityCheckinDo)
	R(consts.MsgTypePaymentCreateOrderId, static.Msg_Shop_Exchange{}, Proto_PayOrderId)
	R(consts.MsgTypePaymentCreateOrder, static.Msg_Payment_OrderId{}, Proto_PayOrder)
	R(consts.MsgTypeGetBattleRank, static.Msg_C2S_BattleRank{}, Proto_BattleRank)
}

// ! 得到牌桌管理单例
func GetProtocolMgr() *ProtocolMgr {
	return protocolMgrSingleton
}

// 注册协议 对象 处理函数
func (pm *ProtocolMgr) RegisterMessage(header string, proto interface{}, handler ProtocolHandler) {
	var info ProtocolInfo
	info.protoType = reflect.TypeOf(proto)
	info.protoHandler = handler
	pm.funcProtocol[header] = &info
}

func (pm *ProtocolMgr) GetProtocol(protocol string) *ProtocolInfo {
	return pm.funcProtocol[protocol]
}

// 协议计数
func (pm *ProtocolMgr) AddProtocolCnt(head string) {
	if idx, ok := pm.protocolCntMap.cntIdx.Load(head); ok {
		if idx.(int) < len(pm.protocolCntMap.cntArr) {
			pm.protocolCntMap.cntArr[idx.(int)] += 1
		}
	} else {
		pm.protocolCntMap.cntIdx.Store(head, len(pm.protocolCntMap.cntArr))
		pm.protocolCntMap.cntArr = append(pm.protocolCntMap.cntArr, 1)
	}
}

// 遍历协议 统计次数 写入文件
func (pm *ProtocolMgr) StataProtocolCnt() {
	path := "./log/" + time.Now().Format("20060102") + "/hall_protocol_cnt.txt"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		xlog.Logger().Errorf("统计协议个数 文件读取失败 path = %s, err = %s", path, err)
		return
	}
	defer file.Close()

	// 文件指针移动到末尾
	file.Seek(0, io.SeekEnd)

	pm.protocolCntMap.cntIdx.Range(func(k, v interface{}) bool {
		if v.(int) < len(pm.protocolCntMap.cntArr) {
			strLine := fmt.Sprintf("%s:%d\n", k, pm.protocolCntMap.cntArr[v.(int)])
			_, err = file.WriteString(strLine)
			pm.protocolCntMap.cntIdx.Delete(k)
		}
		return true
	})

	// 清空arr
	pm.protocolCntMap.cntArr = make([]int64, 0)
}
