package center

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/cast"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/authentication"
	"github.com/open-source/game/chess.git/pkg/static/util"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	"github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"math/rand/v2"
	"sort"
	"strconv"
	"time"
)

// 更新用户签名
func Proto_UpdateUserDescribe(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_UpdateDescribeInfo)

	// 参数校验
	if len(req.Describe) > 72 {
		cuserror := xerrors.NewXError("个性签名不能超过72个英文或者24个汉字")
		return cuserror.Code, cuserror.Msg
	}
	// 更新内存
	p.DescribeInfo = req.Describe
	// 更新redis
	GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "DescribeInfo", p.DescribeInfo)

	return xerrors.SuccessCode, nil
}

// 更新用户信息
func Proto_UpdateUserInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_UpdateUserInfo)

	// 参数校验
	if len(req.Nickname) > 21 {
		cuserror := xerrors.NewXError("昵称不能超过14个英文或者7个汉字")
		return cuserror.Code, cuserror.Msg
	}
	if req.Sex != consts.SexMale && req.Sex != consts.SexFemale {
		cuserror := xerrors.NewXError("性别参数不合法")
		return cuserror.Code, cuserror.Msg
	}
	// 校验用户权限
	if p.UserType != consts.UserTypeYk && p.UserType != consts.UserTypeMobile && p.UserType != consts.UserTypeMobile2 {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	// 更新内存
	p.Sex = req.Sex
	p.Nickname = req.Nickname
	// 更新redis
	GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Sex", p.Sex, "Nickname", p.Nickname)

	return xerrors.SuccessCode, nil
}

// 获取区域列表
func Proto_GetAreas(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// areas := GetAreaMgr().GetAreas()
	// result := make([]*public.Msg_S2C_Area, 0)
	// for _, area := range areas {
	// 	result = append(result, &public.Msg_S2C_Area{
	// 		Code: area.Code,
	// 		Name: area.Name,
	// 	})
	// }
	return xerrors.SuccessCode, nil
}

// 加入区域
func Proto_AreaIn(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// // 数据结构
	// code, v = pm.Proto_AreaEnter(s, p, data)
	// if code != xerrors.SuccessCode {
	// 	return code, v
	// }
	//
	// person := GetPlayerMgr().GetPlayer(p.Uid)
	// if person != nil {
	// 	person.SendMsg(constant.MsgTypeGameCollections_Ntf, getGameCollections())
	// 	person.SendMsg(constant.MsgTypeGameList_Ntf, getGameList(p.Area))
	// }
	// // TODO: 推送公告

	return xerrors.SuccessCode, nil
}

// 加入房间
func Proto_SiteIn(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 判断是否已经加入过场次, 是则重连, 直接返回上个游戏服的信息
	if p.GameId != 0 && p.SiteId != 0 {
		gameserver := GetServer().GetGame(p.GameId)
		if gameserver == nil { // 没有对应的游戏服
			// 获取游戏服失败
			p.GameId = 0
			p.SiteId = 0
			GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "SiteId", p.SiteId, "GameId", p.GameId)
			return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
		}

		var _msg static.Msg_S2C_SiteIn
		_msg.GameId = gameserver.Id
		_msg.KindId = p.SiteId / 100
		_msg.SiteType = p.SiteId % 100
		_msg.Ip = gameserver.ExIp
		_msg.PackageKey = GetAreaPackageKeyByKid(_msg.KindId)
		return xerrors.SuccessCode, _msg
	}

	// 数据结构
	req := data.(*static.Msg_SiteIn)

	// 参数校验
	if req.KindId <= 0 {
		return xerrors.ArgumentErrorCode, xerrors.ArgumentError.Msg
	}

	// 入场门槛校验
	c := GetServer().GetRoomConfig(req.KindId, req.SiteType)
	if c != nil {
		if c.MinScore != 0 && p.Gold < c.MinScore {
			if p.Platform == consts.PlatformWechatApplet {
				return xerrors.GoldNotEnoughError3.Code, xerrors.GoldNotEnoughError3.Msg
			} else {
				return xerrors.GoldNotEnoughError2.Code, xerrors.GoldNotEnoughError2.Msg
			}
		}
		if c.MaxScore != 0 && p.Gold > c.MaxScore {
			return xerrors.GoldExceedingError.Code, xerrors.GoldExceedingError.Msg
		}
	} else {
		return xerrors.SiteNotExistError.Code, xerrors.SiteNotExistError.Msg
	}

	gameserver, notice := GetServer().GetGameBySiteType(p.Uid, req.KindId, req.SiteType)
	if gameserver == nil {
		if notice != nil {
			if notice.GameServerId == static.NoticeMaintainServerAllServer {
				return xerrors.ServerMaintainError.Code, static.HF_JtoA(notice)
			} else {
				return xerrors.GameMaintainError.Code, notice
			}
		}
		// 获取游戏服失败
		return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
	}

	// 如果玩家是重新加入，这里清掉玩家的所有旧数据
	ResetUserCacheData(p)

	// 只分配服务器, 进入服务器后再更新相应的用户信息
	var _msg static.Msg_S2C_SiteIn
	_msg.GameId = gameserver.Id
	_msg.KindId = req.KindId
	_msg.SiteType = req.SiteType
	_msg.Ip = gameserver.ExIp
	_msg.PackageKey = GetAreaPackageKeyByKid(_msg.KindId)
	return xerrors.SuccessCode, _msg
}

// 检查回放码
func Proto_CheckReplayId(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CheckReplayId)

	if req.ReplayId <= 0 {
		// return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		cuserror := xerrors.NewXError("回放记录不存在")
		return cuserror.Code, cuserror.Msg
	}

	var gameReplay models.RecordGameReplay
	if err := GetDBMgr().GetDBmControl().Model(&gameReplay).Where("id = ?", req.ReplayId).First(&gameReplay).Error; err != nil {
		cuserror := xerrors.NewXError("回放记录不存在")
		return cuserror.Code, cuserror.Msg
	}

	areaPkg := GetAreaPackageByKid(gameReplay.KindID)
	if areaPkg == nil {
		cuserror := xerrors.NewXError("回放记录所对应的区域包不存在")
		return cuserror.Code, cuserror.Msg
	}

	return xerrors.SuccessCode, &static.Msg_S2C_CheckReplayId{KindId: gameReplay.KindID, PkgKey: areaPkg.PackageKey}
}

// 大厅历史战绩
func Proto_GetHallGameRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_HallGameRecord)

	if req.End > consts.MaxRecordDays {
		return xerrors.ResultErrorCode, xerrors.NewXError(fmt.Sprintf("最多只能查看%d天之内的战绩", consts.MaxRecordDays))
	}

	// 小程序展示展示限定游戏玩法
	kindidRange := ""
	if p.Platform == consts.PlatformWechatApplet {
		for _, kid := range AppletCardShowGame {
			if len(kindidRange) == 0 {
				kindidRange = strconv.Itoa(kid)
			} else {
				kindidRange += fmt.Sprintf(",%d", kid)
			}
		}
	}

	var list []static.GameRecordHistory
	for i := 0; i < 8; i++ {
		if i >= req.Start && i <= req.End {
			daylist, err := GetDBMgr().SelectNormalHallRecord(p.Uid, time.Now().AddDate(0, 0, -i), kindidRange)
			if err != nil {
				continue
			}
			list = append(list, daylist...)
		}
	}

	sort.Sort(static.GameRecordHistoryWrapper{list, func(item1, item2 *static.GameRecordHistory) bool {
		if item1.PlayedAt > item2.PlayedAt {
			return true
		} else {
			return false
		}
	}})

	result := make([]*static.Msg_S2C_GameRecordHistory, 0)
	for i := 0; i < len(list); i++ {
		item := list[i]
		areaGame := GetAreaGameByKid(item.KindId)
		if areaGame == nil {
			xlog.Logger().Errorln("Proto_GetHallGameRecord:不存在的区域游戏", item.KindId)
			continue
		}
		areaPkg := GetAreaPackageByPKey(areaGame.PackageKey)
		if areaPkg == nil {
			xlog.Logger().Errorln("Proto_GetHallGameRecord:不存在的区域包", areaGame.PackageKey)
			continue
		}
		result = append(result, &static.Msg_S2C_GameRecordHistory{
			GameNum:     item.GameNum,
			RoomNum:     item.RoomNum,
			KindId:      item.KindId,
			Wf:          areaGame.Name,
			Icon:        areaPkg.Icon,
			PkgKey:      areaPkg.PackageKey,
			Time:        item.PlayedAt,
			Player:      item.Player,
			RoundPlayed: item.PlayCount,
			RoundSum:    item.Round,
			PlayerCount: len(item.Player),
		})
	}

	ack := make([]*static.Msg_S2C_GameRecordHistory, 0)
	var idxBeg int = req.PBegin
	var idxEnd int = req.PEnd
	if idxEnd == 0 {
		idxEnd = 10
	}
	idxEnd++
	if idxEnd > len(result) {
		idxEnd = len(result)
	}
	if idxBeg > idxEnd {
		idxBeg = idxEnd
	}
	ack = append(ack, result[idxBeg:idxEnd]...)

	return xerrors.SuccessCode, ack
}

// 战绩详情
func Proto_GetGameRecordInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_GameRecordInfo)

	// 参数校验
	if req.GameNum == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	info, err := GetDBMgr().SelectGameRecordInfo(req.GameNum)
	if err != nil {
		xlog.Logger().Errorln(err)
		cuserror := xerrors.NewXError("获取战绩详情失败")
		return cuserror.Code, cuserror.Msg
	}

	// 更新头像跟性别
	for index, userinfo := range info.UserArr {
		p, err := GetDBMgr().GetDBrControl().GetPerson(userinfo.Uid)
		if p != nil && err == nil {
			info.UserArr[index].Imgurl = p.Imgurl
			info.UserArr[index].Sex = p.Sex
		}
	}

	return xerrors.SuccessCode, info
}

// 保存游戏列表
func Proto_SaveGames(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_SaveGames)

	p.Games = req.Games
	var err error
	// 更新redis数据
	err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Games", p.Games)
	if err != nil {
		xlog.Logger().Errorln("update user to redis failed: ", err.Error())
		cuserror := xerrors.NewXError("更新用户信息失败")
		return cuserror.Code, cuserror.Msg
	}

	return xerrors.SuccessCode, nil
}

// 实名认证
func Proto_Certification(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_Certification)

	if len(req.Name) < 6 || len(req.Name) > 15 {
		cuserror := xerrors.NewXError("姓名格式错误，\n请重新输入2-5位中文字符！")
		return cuserror.Code, cuserror.Msg
	}

	// 身份证校验
	if !static.CheckIdcard(req.Idcard) {
		// 身份证号错误
		cuserror := xerrors.NewXError("请输入有效的身份证号码")
		return cuserror.Code, cuserror.Msg
	}

	// 绑定身份证号码及姓名加密处理
	encodeName, err := static.HF_EncodeStr(static.HF_Atobytes(req.Name), static.UserEncodeKey)
	if err != nil {
		xlog.Logger().Errorf("实名认证-加密姓名:%s失败, err:%s", req.Name, err)
		cuserror := xerrors.NewXError("实名认证失败")
		return cuserror.Code, cuserror.Msg
	}
	encodeIdCard, err := static.HF_EncodeStr(static.HF_Atobytes(req.Idcard), static.UserEncodeKey)
	if err != nil {
		xlog.Logger().Errorf("实名认证-加密身份证号:%s失败, err:%s", req.Idcard, err)
		cuserror := xerrors.NewXError("身份证号认证失败")
		return cuserror.Code, cuserror.Msg
	}

	// 判断用户是否已经实名认证过
	if p.Idcard != "" && p.IdcardAuthPI != "" {
		cuserror := xerrors.NewXError("您已经实名认证过了, 无需重复认证！")
		return cuserror.Code, cuserror.Msg
	} else if p.Idcard != "" && p.IdcardAuthPI == "" {
		// 官方认证尚未完成 向官方查询认证结果
		authSta, authPI, _ := authentication.AuthenticationQuery(strconv.FormatInt(p.Uid, 10))
		if authSta == authentication.AuthenticationCheckStatusSuccess {
			// 更新数据库
			updateMap := make(map[string]interface{})
			updateMap["idcard"] = encodeIdCard
			updateMap["rename"] = encodeName
			updateMap["idcard_auth_pi"] = authPI
			if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
				xlog.Logger().Errorln(err)
				cuserror := xerrors.NewXError("实名认证失败")
				return cuserror.Code, cuserror.Msg
			}

			// 更新redis
			p.Idcard = encodeIdCard
			p.ReName = encodeName
			p.IdcardAuthPI = authPI
			if err := GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Idcard", p.Idcard, "ReName", p.ReName, "IdcardAuthPI", p.IdcardAuthPI); err != nil {
				xlog.Logger().Errorln("set user data to redis error: ", err.Error())
				cuserror := xerrors.NewXError("实名认证失败")
				return cuserror.Code, cuserror.Msg
			}

			return xerrors.SuccessCode, &static.WealthAward{}
		} else if authSta == authentication.AuthenticationCheckStatusIng {
			cuserror := xerrors.NewXError("认证中，请耐心等候！")
			return cuserror.Code, cuserror.Msg
		} else if authSta == authentication.AuthenticationCheckStatusFail {
			// 重新发起认证（兼容老数据）
			newAuthSta, newAuthPI, newAuthErrCode := authentication.AuthenticationCheck(strconv.FormatInt(p.Uid, 10), req.Name, req.Idcard)
			if newAuthSta == authentication.AuthenticationCheckStatusSuccess {
				// 更新数据库
				updateMap := make(map[string]interface{})
				updateMap["idcard"] = encodeIdCard
				updateMap["rename"] = encodeName
				updateMap["idcard_auth_pi"] = newAuthPI
				if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
					xlog.Logger().Errorln(err)
					cuserror := xerrors.NewXError("实名认证失败")
					return cuserror.Code, cuserror.Msg
				}

				// 更新redis
				p.Idcard = encodeIdCard
				p.ReName = encodeName
				p.IdcardAuthPI = newAuthPI
				if err := GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Idcard", p.Idcard, "ReName", p.ReName, "IdcardAuthPI", p.IdcardAuthPI); err != nil {
					xlog.Logger().Errorln("set user data to redis error: ", err.Error())
					cuserror := xerrors.NewXError("实名认证失败")
					return cuserror.Code, cuserror.Msg
				}

				return xerrors.SuccessCode, &static.WealthAward{}
			} else if newAuthSta == authentication.AuthenticationCheckStatusIng {
				cuserror := xerrors.NewXError("认证中，请耐心等候！")
				return cuserror.Code, cuserror.Msg
			} else {
				// 更新数据库
				updateMap := make(map[string]interface{})
				updateMap["idcard"] = ""
				updateMap["rename"] = ""
				updateMap["idcard_auth_pi"] = ""
				if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
					xlog.Logger().Errorln(err)
					cuserror := xerrors.NewXError("实名认证失败")
					return cuserror.Code, cuserror.Msg
				}

				// 更新redis
				p.Idcard = ""
				p.ReName = ""
				p.IdcardAuthPI = ""
				if err := GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Idcard", p.Idcard, "ReName", p.ReName, "IdcardAuthPI", p.IdcardAuthPI); err != nil {
					xlog.Logger().Errorln("set user data to redis error: ", err.Error())
					cuserror := xerrors.NewXError("实名认证失败")
					return cuserror.Code, cuserror.Msg
				}
				errAuthMsg := ""
				if newAuthErrCode == 0 {
					errAuthMsg = "认证失败，请核对信息！"
				} else if newAuthErrCode == authentication.AuthErrCode_IllegalIDCard {
					errAuthMsg = "认证失败，身份证格式有误，请输入正确的身份证号。"
				} else if newAuthErrCode == authentication.AuthErrCode_ResourceLimit {
					errAuthMsg = "认证失败，身份证已实名过，请更换其他身份证实名。"
				} else if newAuthErrCode == authentication.AuthErrCode_NoAuthRecord {
					errAuthMsg = "认证失败，实名认证已失效，请重新认证。"
				} else if newAuthErrCode == authentication.AuthErrCode_RepeatAuth {
					errAuthMsg = "认证失败，实名认证频繁，请稍后再试。"
				} else {
					errAuthMsg = fmt.Sprintf("认证失败，错误码：%d", newAuthErrCode)
				}
				cuserror := xerrors.NewXError(errAuthMsg)
				return cuserror.Code, cuserror.Msg
			}
		}
	}

	// 判断身份证号有无被使用过
	var count int
	if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("idcard = ? and account_type = 0", encodeIdCard).Count(&count).Error; err != nil {
		xlog.Logger().Errorln(err)
		cuserror := xerrors.NewXError("实名认证失败")
		return cuserror.Code, cuserror.Msg
	}
	if count > 0 {
		cuserror := xerrors.NewXError("认证失败，此身份证已经被其他账户认证")
		return cuserror.Code, cuserror.Msg
	}

	// 向官方发起认证
	authSta, authPI, authErrCode := authentication.AuthenticationCheck(strconv.FormatInt(p.Uid, 10), req.Name, req.Idcard)
	if authSta == authentication.AuthenticationCheckStatusSuccess {
		// 更新数据库
		updateMap := make(map[string]interface{})
		updateMap["idcard"] = encodeIdCard
		updateMap["rename"] = encodeName
		updateMap["idcard_auth_pi"] = authPI
		if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
			xlog.Logger().Errorln(err)
			cuserror := xerrors.NewXError("实名认证失败")
			return cuserror.Code, cuserror.Msg
		}

		// 更新person信息
		p.Idcard = encodeIdCard
		p.ReName = encodeName
		p.IdcardAuthPI = authPI
		err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Idcard", p.Idcard, "ReName", p.ReName, "IdcardAuthPI", p.IdcardAuthPI)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			cuserror := xerrors.NewXError("实名认证失败")
			return cuserror.Code, cuserror.Msg
		}

		return xerrors.SuccessCode, &static.WealthAward{}
	} else if authSta == authentication.AuthenticationCheckStatusIng {
		cuserror := xerrors.NewXError("认证中，请耐心等候！")
		// 更新数据库
		updateMap := make(map[string]interface{})
		updateMap["idcard"] = encodeIdCard
		updateMap["rename"] = encodeName
		if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
			xlog.Logger().Errorln(err)
			cuserror := xerrors.NewXError("实名认证失败")
			return cuserror.Code, cuserror.Msg
		}

		// 更新person信息
		p.Idcard = encodeIdCard
		p.ReName = encodeName

		err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Idcard", p.Idcard, "ReName", p.ReName)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			cuserror := xerrors.NewXError("实名认证失败")
			return cuserror.Code, cuserror.Msg
		}

		return cuserror.Code, cuserror.Msg
	} else if authSta == authentication.AuthenticationCheckStatusFail {
		errAuthMsg := ""
		if authErrCode == 0 {
			errAuthMsg = "认证失败，请核对信息！"
		} else if authErrCode == authentication.AuthErrCode_IllegalIDCard {
			errAuthMsg = "认证失败，身份证格式有误，请输入正确的身份证号。"
		} else if authErrCode == authentication.AuthErrCode_ResourceLimit {
			errAuthMsg = "认证失败，身份证已实名过，请更换其他身份证实名。"
		} else if authErrCode == authentication.AuthErrCode_NoAuthRecord {
			errAuthMsg = "认证失败，实名认证已失效，请重新认证。"
		} else if authErrCode == authentication.AuthErrCode_RepeatAuth {
			errAuthMsg = "认证失败，实名认证频繁，请稍后再试。"
		} else {
			errAuthMsg = fmt.Sprintf("认证失败，错误码：%d", authErrCode)
		}
		cuserror := xerrors.NewXError(errAuthMsg)
		return cuserror.Code, cuserror.Msg
	}

	return xerrors.SuccessCode, &static.WealthAward{}
}

// 解绑手机
func Proto_UnbindMobile(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_UnbindMobile)

	if req.Mobile == "" || req.Code == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 校验验证码
	err := service2.GetAdminClient().CheckSmsCode(req.Mobile, req.Code, service2.SmsTypeUnbind)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.SmsCodeError.Code, xerrors.SmsCodeError.Msg
	}

	req.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(req.Mobile), static.UserEncodeKey)
	if err != nil {
		return xerrors.MobileInvalidError.Code, xerrors.MobileInvalidError.Msg
	}

	// 验证用户提交的手机号码是否是之前绑定的手机号码
	if p.Tel != req.Mobile {
		cuserror := xerrors.NewXError("请输入原绑定手机号进行解绑操作")
		return cuserror.Code, cuserror.Msg
	}

	return xerrors.SuccessCode, nil
}

// 绑定微信
func Proto_BindWechat(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_BindWechat)

	// 校验参数
	if req.Code == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 默认app
	if req.AppId == "" {
		req.AppId = consts.AppDefault
	}

	config := GetServer().GetAppConfig(req.AppId)
	if config == nil {
		return xerrors.ThirdpartyError.Code, xerrors.ThirdpartyError.Msg
	}

	info, err := service2.NewWeixinClient(config.WxAppId, config.WxAppSecret).GetWeixinUserInfo(req.Code)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.ThirdpartyError.Code, xerrors.ThirdpartyError.Msg
	}

	if info.Sex == consts.SexUnknown {
		info.Sex = consts.SexFemale
	}

	if p.UserType == consts.UserTypeYk { // 游客账号
		// 先判断是否已经存在微信账号
		var user models.User
		if err = GetDBMgr().GetDBmControl().Where("union_id = ?", info.UnionId).First(&user).Error; err == nil {
			// 存在用户则报错提示
			cuserror := xerrors.NewXError("该微信已经注册过了，请使用新的微信号进行授权，或者在设置页面点击账号切换按钮切换成微信登录。")
			return cuserror.Code, cuserror.Msg
		}
		// 不存在用户则升级成微信账号
		// 更新用户信息
		updateMap := make(map[string]interface{})
		updateMap["nickname"] = info.Nickname
		updateMap["open_id"] = info.OpenId
		updateMap["union_id"] = info.UnionId
		updateMap["imgurl"] = info.Headimgurl
		updateMap["sex"] = info.Sex
		updateMap["user_type"] = consts.UserTypeWechat
		if err = GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
			xlog.Logger().Errorln(err)
			cuserror := xerrors.NewXError("绑定微信失败")
			return cuserror.Code, cuserror.Msg
		}

		// 更新内存
		p.Nickname = info.Nickname
		p.Openid = info.OpenId
		p.Imgurl = info.Headimgurl
		p.Sex = info.Sex
		p.UserType = consts.UserTypeWechat

		// 更新redis
		err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Nickname", p.Nickname, "Openid", p.Openid, "Imgurl", p.Imgurl, "Sex", p.Sex, "UserType", p.UserType)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			cuserror := xerrors.NewXError("绑定微信失败")
			return cuserror.Code, cuserror.Msg
		}
	} else if p.UserType == consts.UserTypeMobile { // 手机账号
		// 更新用户信息(不更新open_id和union_id, 使用微信登录后是新账号)
		updateMap := make(map[string]interface{})
		updateMap["nickname"] = info.Nickname
		// updateMap["open_id"] = eve.OpenId
		// updateMap["union_id"] = eve.UnionId
		updateMap["imgurl"] = info.Headimgurl
		updateMap["sex"] = info.Sex
		updateMap["user_type"] = consts.UserTypeMobile2
		if err = GetDBMgr().GetDBmControl().Model(models.User{Id: p.Uid}).Updates(updateMap).Error; err != nil {
			xlog.Logger().Errorln(err)
			cuserror := xerrors.NewXError("绑定微信失败")
			return cuserror.Code, cuserror.Msg
		}

		// 更新内存
		p.Nickname = info.Nickname
		// p.Openid = eve.OpenId
		p.Imgurl = info.Headimgurl
		p.Sex = info.Sex
		p.UserType = consts.UserTypeMobile2

		// 更新redis
		err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Nickname", p.Nickname, "Imgurl", p.Imgurl, "Sex", p.Sex, "UserType", p.UserType)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			cuserror := xerrors.NewXError("绑定微信失败")
			return cuserror.Code, cuserror.Msg
		}
	} else {
		// 其他类型账号只更新用户信息
		updateMap := make(map[string]interface{})
		updateMap["nickname"] = info.Nickname
		updateMap["imgurl"] = info.Headimgurl
		updateMap["sex"] = info.Sex
		if err = GetDBMgr().GetDBmControl().Model(models.User{Id: p.Uid}).Updates(updateMap).Error; err != nil {
			xlog.Logger().Errorln(err)
			cuserror := xerrors.NewXError("绑定微信失败")
			return cuserror.Code, cuserror.Msg
		}

		// 更新内存
		p.Nickname = info.Nickname
		p.Imgurl = info.Headimgurl
		p.Sex = info.Sex

		// 更新redis
		err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Nickname", p.Nickname, "Imgurl", p.Imgurl, "Sex", p.Sex)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			cuserror := xerrors.NewXError("更新微信信息失败")
			return cuserror.Code, cuserror.Msg
		}
	}

	return xerrors.SuccessCode, &static.Msg_S2C_BindWechatResult{Nickname: p.Nickname, Imgurl: p.Imgurl, Sex: p.Sex}
}

// 授权微信
func Proto_AuthWechat(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_AuthWechat)

	if req.RawData == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	userInfo := new(service2.WeiXinAppletUserInfo)
	err := json.Unmarshal([]byte(req.RawData), userInfo)
	if err != nil {
		cuserror := xerrors.NewXError("解析用户信息失败")
		return cuserror.Code, cuserror.Msg
	}

	// 更新用户信息
	updateMap := make(map[string]interface{})
	updateMap["nickname"] = userInfo.Nickname
	updateMap["imgurl"] = userInfo.AvatarUrl
	updateMap["sex"] = userInfo.Gender
	if err = GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", p.Uid).Updates(updateMap).Error; err != nil {
		xlog.Logger().Errorln(err)
		cuserror := xerrors.NewXError("授权微信失败")
		return cuserror.Code, cuserror.Msg
	}

	// 更新内存
	p.Nickname = userInfo.Nickname
	p.Imgurl = userInfo.AvatarUrl
	p.Sex = userInfo.Gender

	// 更新redis
	err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Nickname", p.Nickname, "Imgurl", p.Imgurl, "Sex", p.Sex)
	if err != nil {
		xlog.Logger().Errorln(err)
		cuserror := xerrors.NewXError("授权微信失败")
		return cuserror.Code, cuserror.Msg
	}

	return xerrors.SuccessCode, nil
}

// 绑定手机
func Proto_BindMobile(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	// 数据结构
	req := data.(*static.Msg_BindMobile)

	// 校验手机号
	if !static.CheckMobile(req.Mobile) {
		return xerrors.MobileInvalidError.Code, xerrors.MobileInvalidError.Msg
	}
	if req.Code == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 校验验证码
	err := service2.GetAdminClient().CheckSmsCode(req.Mobile, req.Code, service2.SmsTypeBind)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.SmsCodeError.Code, xerrors.SmsCodeError.Msg
	}

	req.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(req.Mobile), static.UserEncodeKey)
	if err != nil {
		return xerrors.MobileInvalidError.Code, xerrors.MobileInvalidError.Msg
	}

	// 如果新手机号和旧手机号相同, 则直接返回成功
	if p.Tel != "" && req.Mobile == p.Tel {
		return xerrors.SuccessCode, nil
	}

	// 校验手机号是否已经绑定过账号

	tx := GetDBMgr().GetDBmControl().Begin()

	var user models.User
	if err = tx.Where("tel = ? and account_type = 0", req.Mobile).First(&user).Error; err == nil {
		cuserror := xerrors.NewXError("该手机号已经绑定过账号")
		tx.Rollback()
		return cuserror.Code, cuserror.Msg
	}
	// 队长更换手机号
	if league := GetAllianceMgr().GetUserLeagueID(p.Uid); league != 0 {
		xlog.Logger().Debugln("盟主更换手机号")
		if err := service2.GetAdminClient().UpdateUserInfo(p.Uid, "tel", req.Mobile, "mobile", req.Mobile); err != nil {
			xlog.Logger().Errorf("盟主更换手机号失败:error:%v", err)
			cuserror := xerrors.NewXError("绑定手机失败")
			tx.Rollback()
			return cuserror.Code, cuserror.Msg
		}
	} else {
		// 更新database
		if err = tx.Model(models.User{Id: p.Uid}).Update("tel", req.Mobile).Error; err != nil {
			xlog.Logger().Errorln(err)
			cuserror := xerrors.NewXError("绑定手机失败")
			tx.Rollback()
			return cuserror.Code, cuserror.Msg
		}
	}
	// 更新redis
	err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Tel", req.Mobile)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		cuserror := xerrors.NewXError("绑定手机失败")
		tx.Rollback()
		return cuserror.Code, cuserror.Msg
	}

	tx.Commit()
	// 更新内存
	p.Tel = req.Mobile
	return xerrors.SuccessCode, nil
}

// 绑定手机v2(金币场)
func Proto_BindMobileV2(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_BindMobile)

	// 校验手机号
	if !static.CheckMobile(req.Mobile) {
		return xerrors.MobileInvalidError.Code, xerrors.MobileInvalidError.Msg
	}
	if req.Code == "" {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	// 校验密码
	if len(req.Password) < 6 || len(req.Password) > 16 {
		cuserror := xerrors.NewXError("密码长度应在6-16个字符之间")
		return cuserror.Code, cuserror.Msg
	}

	// 判断用户权限
	if p.UserType != consts.UserTypeYk {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 效验验证码
	err := service2.GetAdminClient().CheckSmsCode(req.Mobile, req.Code, service2.SmsTypeBind)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.SmsCodeError.Code, xerrors.SmsCodeError.Msg
	}

	req.Mobile, err = static.HF_EncodeStr(static.HF_Atobytes(req.Mobile), static.UserEncodeKey)
	if err != nil {
		return xerrors.MobileInvalidError.Code, xerrors.MobileInvalidError.Msg
	}

	// 校验手机号是否已经绑定过账号
	var user models.User
	if err = GetDBMgr().GetDBmControl().Where("tel = ? and account_type = 0", req.Mobile).First(&user).Error; err == nil {
		cuserror := xerrors.NewXError("该手机号已被注册，请更换新的手机号进行注册~")
		return cuserror.Code, cuserror.Msg
	}

	// 更新database
	updateMap := make(map[string]interface{})
	updateMap["tel"] = req.Mobile
	updateMap["user_type"] = consts.UserTypeMobile
	updateMap["password"] = util.MD5(req.Password)
	if err = GetDBMgr().GetDBmControl().Model(models.User{Id: p.Uid}).Updates(updateMap).Error; err != nil {
		xlog.Logger().Errorln(err)
		cuserror := xerrors.NewXError("绑定手机失败")
		return cuserror.Code, cuserror.Msg
	}

	// 更新内存
	p.Tel = req.Mobile
	p.UserType = consts.UserTypeMobile // 升级成手机账号

	// 更新redis
	err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Tel", p.Tel, "UserType", p.UserType)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		cuserror := xerrors.NewXError("绑定手机失败")
		return cuserror.Code, cuserror.Msg
	}

	return xerrors.SuccessCode, nil
}

// 创建牌桌
func Proto_CreateTable(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CreateTable)

	// 参数校验
	config, cuserror := validateCreateTableParam(req, false)
	if cuserror != nil {
		return cuserror.Code, cuserror.Msg
	}

	// 判断大厅是否维护
	if notice := CheckServerMaintainWithWhite(p.Uid, static.NoticeMaintainServerAllServer); notice != nil {
		// return xerrors.ServerMaintainError.Code, xerrors.ServerMaintainError.Msg
		return xerrors.ServerMaintainError.Code, static.HF_JtoA(notice)
	}

	// 判断子游戏是否维护
	// if notice := CheckMaintainWithWhite(p.Uid, config.KindId); notice != nil {
	// 	return xerrors.GameMaintainError.Code, notice
	// }

	// 从redis获取一下最新的数据
	rPerson, err := GetDBMgr().GetDBrControl().GetPerson(p.Uid)
	if err != nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	// 使用redis数据更新一下缓存中的冻结房卡
	p.FrozenCard = rPerson.FrozenCard

	// 平台标识
	createChannel := consts.ChannelApp
	if rPerson.UserType == consts.UserTypeWechat && rPerson.Platform == consts.PlatformWechatApplet {
		createChannel = consts.ChannelApplet
	} else if rPerson.UserType == consts.UserTypeHW {
		createChannel = consts.ChannelHW
	}

	// 判断用户是否创建过房间且未被销毁
	if config.CreateType == consts.CreateTypeSelf {
		// 若创建过牌桌, 则无法继续创建
		if p.CreateTable != 0 {
			// 检查桌子是否还存在, 不存在则删除
			table := GetTableMgr().GetTable(p.CreateTable)
			if table != nil {
				// 如果桌子存在且归属正确则报错
				if table.Creator == p.Uid {
					// 判断子游戏是否维护
					if notice := CheckServerMaintainWithWhite(p.Uid, static.NoticeMaintainServerType(table.GameId)); notice != nil {
						return xerrors.GameMaintainError.Code, notice
					}
					// 获取游戏服务器
					gameserver := GetServer().GetGame(table.GameId)
					if gameserver == nil {
						// 获取游戏服失败
						return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
					}
					var ack static.Msg_S2C_TableIn
					ack.Id = table.Id
					ack.GameId = table.GameId
					ack.KindId = table.KindId
					ack.Ip = gameserver.ExIp
					return xerrors.SuccessCode, &ack
				}
			}
			// 修复用户错误数据
			p.CreateTable = 0
			err := GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "CreateTable", p.CreateTable)
			if err != nil {
				xlog.Logger().Errorln("update user redis failed: ", err.Error())
			}
		}
	}

	config.Creator = p.Uid

	// 获取游戏服务器
	gameserver, notice := GetServer().GetGameByKindId(p.Uid, static.GAME_TYPE_FRIEND, config.KindId)
	if gameserver == nil {
		if notice != nil {
			if notice.GameServerId == static.NoticeMaintainServerAllServer {
				return xerrors.ServerMaintainError.Code, static.HF_JtoA(notice)
			} else {
				return xerrors.GameMaintainError.Code, notice
			}
		}
		// 获取游戏服失败
		return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
	}

	// 剩余房卡判断
	if createChannel == consts.ChannelApplet {
		// 小程序创建好友房免费
	} else {
		if p.Card-p.FrozenCard < config.CardCost {
			return xerrors.NotEnoughCardCode, xerrors.NotEnoughCardError.Msg
		}
	}

	// 创建内存牌桌
	table := new(static.Table)
	nowTime := time.Now()
	table.Id = GetTableMgr().GetRandomTableId()
	if table.Id <= 0 {
		xlog.Logger().Errorln("GetRandomTableId error1")
		return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
	}
	table.Channel = createChannel
	table.IsCost = false
	table.Creator = p.Uid
	table.CreateType = consts.CreateTypeSelf
	table.GameId = gameserver.Id
	table.KindId = config.KindId
	table.Users = make([]*static.TableUser, config.MaxPlayerNum)
	table.GameNum = fmt.Sprintf("%s_%d_%d", nowTime.Format("20060102150405"), gameserver.Id, table.Id)
	table.Config = &static.TableConfig{
		MaxPlayerNum:   config.MaxPlayerNum,
		MinPlayerNum:   config.MinPlayerNum,
		RoundNum:       config.RoundNum,
		CardCost:       config.CardCost,
		CostType:       config.CostType,
		View:           config.View,
		Restrict:       config.Restrict,
		FewerStart:     config.FewerStart,
		GVoice:         config.GVoice,
		GameConfig:     config.GameConfig,
		GameType:       static.GAME_TYPE_FRIEND, // 默认好友房
		IsLimitChannel: config.IsLimitChannel,
	}

	// 创建房间
	ntb := GetTableMgr().CreateTable(table)
	if ntb == nil {
		// 还原链表
		GetTableMgr().TableIds.Prepend(static.NewINode(static.ElementType(table.Id), nil))
		// 创建牌桌失败
		xlog.Logger().Errorln("CreateTable error")
		return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
	}

	msg := new(static.HG_HTableCreate_Req)
	msg.Table = *table
	if GetServer().Con.UseSafeIp == 0 {
		msg.Ip = gameserver.ExIp
	} else {
		msg.Ip = gameserver.SafeIp
	}
	msg.Payer = 0 // 非包厢桌子 没有payer
	msg.AutoUid = p.Uid
	msg.AutoSeat = 0
	msg.Table.LeagueID = 0
	// 发送创建协议
	tableinMsg := new(static.Msg_S2C_TableIn)
	reply, err := GetServer().CallGame(gameserver.Id, p.Uid, "NewServerMsg", consts.MsgTypeTableCreate_Req, xerrors.SuccessCode, msg)
	if string(reply) != "SUC" || err != nil {
		// 协议未正常受理
		// 删除内存数据
		tb := GetTableMgr().GetTable(table.Id)
		if tb != nil {
			GetTableMgr().DelTable(tb)
			xlog.Logger().Errorln("CallGame error2", err)
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		}
	} else {
		tableinMsg.Id = table.Id
		tableinMsg.GameId = table.GameId
		tableinMsg.KindId = table.KindId
		tableinMsg.Ip = msg.Ip
	}

	// 更新person
	if config.CreateType == consts.CreateTypeSelf {
		p.CreateTable = table.Id
	} else if config.CreateType == consts.CreateTypeOther {
		// CreateTypeOther
	}

	// 冻结房卡
	if createChannel == consts.ChannelApplet {
		// 小程序创建好友房免费
	} else {
		tx := GetDBMgr().GetDBmControl().Begin()
		defer tx.Rollback()
		if p.Card-p.FrozenCard < config.CardCost {
			return xerrors.NotEnoughCardCode, xerrors.NotEnoughCardError.Msg
		}
		_, aftka, _, aftfka, err := wealthtalk.UpdateCard(p.Uid, 0, config.CardCost, models.CostFrozen, tx)
		if err != nil {
			xlog.Logger().Errorln(err)
		} else {
			// 更新内存
			p.Card = aftka
			p.FrozenCard = aftfka
			// 更新redis
			err := GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Card", p.Card, "FrozenCard", p.FrozenCard)
			if err != nil {
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
			err = tx.Commit().Error
			if err != nil {
				xlog.Logger().Errorln(err)
			}
		}
	}

	return xerrors.SuccessCode, &tableinMsg
}

// 加入牌桌
func Proto_JoinTable(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_JoinTable)

	// 已加入牌桌 进入历史牌桌
	key := fmt.Sprintf("userstatus_doing_join_%d", p.Uid)
	cli := GetDBMgr().Redis
	if !cli.SetNX(key, p.Uid, time.Second*3).Val() {
		return xerrors.HouseFloorTableJoiningError.Code, xerrors.HouseFloorTableJoiningError.Msg
	}

	// 小程序没有这个玩法的话 温馨提示 暂不支持此玩法
	if p.Platform == consts.PlatformWechatApplet {
		var table *Table
		if p.TableId != 0 {
			table = GetTableMgr().GetTable(p.TableId)
		} else {
			table = GetTableMgr().GetTable(req.Id)
		}

		if table == nil {
			return xerrors.TableInError.Code, xerrors.TableInError.Msg
		}

		// 玩法检查
		isFind := false
		for _, kid := range AppletCardShowGame {
			if kid == table.KindId {
				isFind = true
				break
			}
		}
		if !isFind {
			cuserror := xerrors.NewXError("小程序暂不支持此玩法，请进入游戏APP体验")
			return cuserror.Code, cuserror.Msg
		}

		// 定位检查
		if table.Config.Restrict {
			cuserror := xerrors.NewXError("小程序暂不支持定位功能，请进入游戏APP体验")
			return cuserror.Code, cuserror.Msg
		}
	}

	// 判断是否加入新桌子
	if p.TableId != 0 {

		table := GetTableMgr().GetTable(p.TableId)
		if table == nil {
			return xerrors.TableInError.Code, xerrors.TableInError.Msg
		}

		// 获取游戏服务器
		gameserver := GetServer().GetGame(table.GameId)
		if gameserver == nil {
			// 获取游戏服失败
			return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
		}
		var _msg static.Msg_S2C_TableIn
		_msg.Id = table.Id
		_msg.GameId = table.GameId
		_msg.KindId = table.KindId
		_msg.Ip = gameserver.ExIp
		return xerrors.SuccessCode, _msg
	}

	// 如果玩家是重新加入，这里清掉玩家的所有旧数据
	ResetUserCacheData(p)

	// 获取房间信息
	table := GetTableMgr().GetTable(req.Id)
	if table == nil {
		// 牌桌不存在
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}

	// 房间是否开启渠道隔离
	if table.Config.IsLimitChannel {
		if table.Channel == consts.ChannelApplet && p.Platform != consts.PlatformWechatApplet {
			cuserror := xerrors.NewXError("此房间为小程序专享，请进入小程序体验")
			return cuserror.Code, cuserror.Msg
		} else if (table.Channel == consts.ChannelApp || table.Channel == consts.ChannelHW) && p.Platform == consts.PlatformWechatApplet {
			cuserror := xerrors.NewXError("此房间为游戏APP专享，请进入游戏APP体验")
			return cuserror.Code, cuserror.Msg
		}
	}

	if table.Config.GVoice == "true" {
		if !req.Voice {
			// return xerrors.OpenGVoiceError.Code, xerrors.OpenGVoiceError.Msg
		}
		if !req.GVoiceOk {
			return xerrors.InitGVoiceError.Code, xerrors.InitGVoiceError.Msg
		}
	}

	// 判断gps限制
	if table.Config.Restrict {
		if p.Platform == consts.PlatformWechatApplet {
			cuserror := xerrors.NewXError("小程序暂不支持定位功能，请进入游戏APP体验")
			return cuserror.Code, cuserror.Msg
		}
		//2人玩法不做 IP和GPS 限制
		if table.Config.MaxPlayerNum != 2 {
			if !req.Gps {
				// 未开启gps服务
				cuserror := xerrors.OpenGpsError
				return cuserror.Code, cuserror.Msg
			}
			if ok := CheckUserIp(s.IP, table.Users); !ok {
				// 未开启gps服务
				cuserror := xerrors.NewXError("相同ip无法加入该桌")
				return cuserror.Code, cuserror.Msg
			}
		}
	}

	// 校验加入桌子参数
	cuserror := table.CheckUserJoinTable(p, -1)
	if cuserror != nil {
		return cuserror.Code, cuserror.Msg
	}
	// 判断子游戏是否维护
	if notice := CheckServerMaintainWithWhite(p.Uid, static.NoticeMaintainServerType(table.GameId)); notice != nil {
		return xerrors.GameMaintainError.Code, notice
	}
	// 获取游戏服务器
	gameserver := GetServer().GetGame(table.GameId)
	if gameserver == nil {
		// 获取游戏服失败
		return xerrors.GetGameServerError.Code, xerrors.GetGameServerError.Msg
	}

	// 如果是包厢房间则校验是否是该包厢的用户
	if table.CreateType == consts.CreateTypeHouse {

		house := GetClubMgr().GetClubHouseByHId(table.HId)
		if house == nil {
			return xerrors.InValidHouseError.Code, xerrors.InValidHouseError.Msg
		}
		if house.DBClub.IsFrozen {
			return xerrors.IsFrozenError.Code, xerrors.IsFrozenError.Msg
		}
		if house.DBClub.OnlyQuickJoin {
			return xerrors.OnlyQucikJoinError.Code, xerrors.OnlyQucikJoinError.Msg
		}
		hmem := house.GetMemByUId(p.Uid)
		if hmem == nil {
			return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
		} else {
			if hmem.Lower(consts.ROLE_MEMBER) {
				return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
			}
		}
		if len(table.Users) > 0 {
			uids := make([]int64, 0, 4)
			uids = append(uids, p.Uid)
			for _, v := range table.Users {
				if v != nil {
					uid := v.Uid
					uids = append(uids, uid)
				}
			}
			if e, notAllowUids := house.CheckUsersAllowSameTable(uids...); e != nil {
				var msg []string
				for _, uid := range notAllowUids {
					dmem, err := GetDBMgr().GetDBrControl().GetPerson(uid)
					if err != nil {
						xlog.Logger().Errorln(err)
						continue
					}
					msg = append(msg, dmem.Nickname)
				}
				var respMsg string
				switch len(msg) {
				case 0: // 假如没有获取到用户昵称
					respMsg = "您暂时无法与该桌用户同桌游戏，如有疑问请联系盟主"
				case 1:
					respMsg = fmt.Sprintf("您暂时无法与%s同桌游戏，如有疑问请联系盟主", msg[0])
				case 2:
					respMsg = fmt.Sprintf("您暂时无法与%s,%s同桌游戏，如有疑问请联系盟主", msg[0], msg[1])
				case 3:
					respMsg = fmt.Sprintf("您暂时无法与%s,%s,%s同桌游戏，如有疑问请联系盟主", msg[0], msg[1], msg[2])
				default: // 超过三人
					respMsg = fmt.Sprintf("您暂时无法与%s,%s,%s等用户同桌游戏，如有疑问请联系盟主", msg[0], msg[1], msg[2])
				}
				return xerrors.HouseTableLimitJoin.Code, respMsg

			}
		}
		hfloor := house.GetFloorByFId(table.FId)
		if hfloor == nil {
			return xerrors.InValidHouseFloorError.Code, xerrors.InValidHouseFloorError.Msg
		}

		if hfloor.IsVipFloor() {
			if !hfloor.IsVipUser(p.Uid) {
				return xerrors.HouseFloorJoinVipError.Code, xerrors.HouseFloorJoinVipError.Msg
			}
		}

		// 疲劳值下限判定
		if house.DBClub.IsVitamin && hfloor.IsVitamin {
			if hfloor.VitaminLowLimit != consts.VitaminInvalidValueSrv && hmem.UVitamin < hfloor.VitaminLowLimit {
				return xerrors.VitaminNotEnoughError.Code, xerrors.VitaminNotEnoughError.Msg
			}
			if hfloor.VitaminHighLimit > 0 && hfloor.VitaminHighLimit != consts.VitaminInvalidValueSrv && hmem.UVitamin >= hfloor.VitaminLowLimit {
				return xerrors.VitaminTooEnoughError.Code, xerrors.VitaminTooEnoughError.Msg
			}
		}

		_msg := new(static.Msg_HouseTableInOpt)
		param := static.Msg_CH_HouseTableIn{}
		param.FId = table.FId
		param.HId = table.HId
		param.NTId = table.NTId
		param.Gps = req.Gps
		param.Voice = req.Voice
		param.GVoiceOk = req.GVoiceOk
		param.KindID = table.KindId
		param.IgnoreRule = true
		_msg.Param = param
		_msg.Header = consts.MsgTypeHouseTableIn
		code, v = ChOptClubHouseTableIn(
			house,
			hfloor,
			hmem,
			&static.GpsInfo{
				Ip:        s.IP,
				Longitude: req.Longitude,
				Latitude:  req.Latitude,
				Address:   req.Address,
			},
			p,
			_msg,
		)
		return code, v
	} else {
		// 非包厢玩法
		// 分配牌桌座位
		seat := table.GetEmptySeat()
		if seat == -1 {
			return xerrors.TableIsFullError.Code, xerrors.TableIsFullError.Msg
		}

		// 通知游戏服加入牌桌
		var msg static.HG_HTableIn_Req
		msg.TId = table.Id
		msg.NTid = table.NTId
		msg.GameId = table.GameId
		msg.KindId = table.KindId
		if GetServer().Con.UseSafeIp == 0 {
			msg.Ip = gameserver.ExIp
		} else {
			msg.Ip = gameserver.SafeIp
		}
		msg.Uid = p.Uid
		msg.Seat = seat
		msg.Payer = 0 // 非包厢没这个东西
		reply, err := GetServer().CallGame(table.GameId, p.Uid, "NewServerMsg", consts.MsgTypeTableIn, xerrors.SuccessCode, &msg)
		if string(reply) != "SUC" || err != nil {
			xlog.Logger().Errorln("call game error3", err)
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		} else {

			gpsInfo := &static.GpsInfo{Ip: s.IP, Longitude: req.Longitude, Latitude: req.Latitude, Address: req.Address}
			if gpsInfo != nil && p != nil {
				if gpsInfo.Longitude != 0 && gpsInfo.Latitude != 0 && gpsInfo.Address == "" {
					gpsInfo.Address = "未知区域"
				}
				if gpsInfo.Latitude != -1 && gpsInfo.Longitude != -1 {
					GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "Latitude", gpsInfo.Latitude, "Longitude", gpsInfo.Longitude, "GameIp", gpsInfo.Ip, "Address", gpsInfo.Address)
				}
			}

			tableinMsg := new(static.Msg_S2C_TableIn)
			tableinMsg.Id = table.Id
			tableinMsg.GameId = table.GameId
			tableinMsg.KindId = table.KindId
			tableinMsg.Ip = msg.Ip

			p.TableId = table.Id
			p.GameId = table.GameId
			// 修改牌桌玩家座位信息
			table.lock.CustomLock()
			// person
			table.Users[seat] = &static.TableUser{Uid: p.Uid, JoinAt: 0, Payer: 0}
			table.lock.CustomUnLock()
			return xerrors.SuccessCode, &tableinMsg
		}
	}
}

// 解散牌桌
func Proto_DeleteTable(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_TableDel)
	table := GetTableMgr().GetTable(req.Id)
	if table == nil {
		xlog.Logger().Errorf("table not in tablemgr:%d", req.Id)
		// return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}
	var hid int
	var fid int64
	if table == nil {
		hid = req.Hid
		fid = req.Fid
	} else {
		hid = table.HId
		fid = table.FId
	}
	house, floor, mem, cusErr := inspectClubFloorMemberWithRight(hid, fid, p.Uid, consts.ROLE_ADMIN, MinorDisRoom)
	if cusErr == xerrors.InvalidPermission {
		if !house.IsPartner(p.Uid) {
			return cusErr.Code, cusErr.Msg
		}
	} else {
		if cusErr != xerrors.RespOk {
			return cusErr.Code, cusErr.Msg
		}
	}

	// 获取桌子
	hft := floor.GetTableByTId(req.Id)
	if hft == nil {
		return xerrors.InValidHouseTableError.Code, xerrors.InValidHouseTableError.Msg
	}

	// redis
	floortable := new(static.FloorTable)
	floortable.NTId = hft.NTId
	floortable.FId = floor.Id
	GetDBMgr().GetDBrControl().FloorTableDelete(floortable)
	if len(hft.UserWithOnline) != floor.Rule.PlayerNum {
		hft.UserWithOnline = make([]FTUsers, floor.Rule.PlayerNum)
	}
	// 删除内存数据
	if table != nil {
		for _, u := range table.Users {
			if u != nil {
				table.UserLeaveTable(u.Uid)
			}
		}
		GetTableMgr().DelTable(table)
	} else {
		for _, u := range hft.UserWithOnline {
			if u.Uid > 0 {
				GetDBMgr().GetDBrControl().UpdatePersonAttrs(u.Uid, "TableId", 0, "GameId", 0)
			}
		}
	}

	if table != nil {
		t := time.Now()
		disRoomTimeStr := fmt.Sprintf("%d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
		tempDismissType := 0 // 0：无,  1：盟主解散，2：管理员解散 3：队长解散 4：申请解散 5：超时解散  6：托管解散
		tempDismissDet := ""
		// 通知游戏服解散牌桌
		var info string
		switch mem.URole {
		case consts.ROLE_MEMBER:
			info = fmt.Sprintf("房间被队长:%d解散", mem.UId)
			tempDismissDet = fmt.Sprintf("队长:%s解散 %s", mem.NickName, disRoomTimeStr)
			tempDismissType = 3
		case consts.ROLE_ADMIN:
			info = fmt.Sprintf("房间被管理员:%d解散", mem.UId)
			tempDismissDet = fmt.Sprintf("管理员:%s解散 %s", mem.NickName, disRoomTimeStr)
			tempDismissType = 2
		case consts.ROLE_CREATER:
			info = fmt.Sprintf("房间被盟主:%d解散", mem.UId)
			tempDismissDet = fmt.Sprintf("盟主:%s解散 %s", mem.NickName, disRoomTimeStr)
			tempDismissType = 1
		}
		hMsg := fmt.Sprintf("ID %d解散了牌桌 房间号%d", mem.UId, table.Id)
		CreateClubMassage(house.DBClub.Id, mem.UId, PartnerDelTable, hMsg)
		var msg static.Msg_HG_TableDel_Req
		msg.TableId = req.Id
		msg.Info = info
		//在中途解散的游戏中 记录一下 解散的详情
		recorddet := new(models.RecordDismiss)
		recorddet.GameNum = table.GameNum
		recorddet.DismissTime = disRoomTimeStr
		recorddet.DismissType = tempDismissType // 0：无,  1：盟主解散，2：管理员解散 3：队长解散 4：申请解散 5：超时解散  6：托管解散
		recorddet.DismissDet = tempDismissDet
		if err := GetDBMgr().GetDBmControl().Model(models.RecordDismiss{}).Save(&recorddet).Error; err != nil {
			xlog.Logger().Errorf("大厅解散房间写入详情失败:%v", err)
		}
		go GetServer().CallGame(table.GameId, 0, "ServerMethod.ServerMsg", consts.MsgTypeHTableDel_Req, xerrors.SuccessCode, &msg)
	}

	// memory
	hft.Clear(floor.DHId)
	return xerrors.SuccessCode, nil
}

// 牌桌详情
func Proto_TableInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_TableInfo)

	// 获取房间信息
	table := GetTableMgr().GetTable(req.Id)
	if table == nil {
		// 牌桌不存在
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}
	// 只允许包厢牌桌查看房间信息
	if !table.IsTeaHouse() {
		// 牌桌不存在
		xlog.Logger().Errorf("get table not thouse:%d", table.Id)
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}

	house, floor, _, cusErr := inspectClubFloorMemberWithRight(table.HId, table.FId, 0, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		xlog.Logger().Errorf("check table error:%s,tableid:%d", cusErr.Msg, table.Id)
		return cusErr.Code, cusErr.Msg
	}

	// 直接从大厅获取牌桌信息
	result := new(static.Msg_S2C_TableInfo)
	result.Begin = table.Begin
	result.Hid = table.HId
	result.Fid = table.FId
	result.TId = table.Id
	result.NTId = table.NTId
	result.RoundNum = table.Config.RoundNum
	result.MaxPlayerNum = table.Config.MaxPlayerNum
	result.CurrentRound = table.Step
	result.Person = make([]*static.Msg_S2C_TablePerson, 0)
	hft := floor.GetHftByTid(table.Id)
	if hft == nil {
		xlog.Logger().Errorf("get table not thouse:%d", table.Id)
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}
	if hft.FakeEndAt > 0 {
		result.Begin = true
	}
	for _, u := range hft.UserWithOnline {
		if hft.FakeEndAt > 0 {
			tablePerson := &static.Msg_S2C_TablePerson{
				Id:       u.Uid,
				Imgurl:   u.UUrl,
				Nickname: u.UName,
				Ip:       u.Ip,
				Online:   true,
			}
			result.Person = append(result.Person, tablePerson)
		} else {
			if u.Uid != 0 {
				// 从内存获取在线用户信息
				p, err := GetDBMgr().GetDBrControl().GetPerson(u.Uid)
				if err != nil {
					xlog.Logger().Errorln("user does not exist: ", u.Uid)
					continue
				} else {
					mem := house.GetMemByUId(p.Uid)
					tablePerson := &static.Msg_S2C_TablePerson{
						Id:       p.Uid,
						Imgurl:   p.Imgurl,
						Nickname: p.Nickname,
						Ip:       p.Ip,
						Online:   u.OnLine,
					}
					if static.IsAnonymous(table.Config.GameConfig) {
						tablePerson = &static.Msg_S2C_TablePerson{
							Id:       p.Uid,
							Imgurl:   "",
							Nickname: "匿名昵称",
							Ip:       p.Ip,
							Online:   u.OnLine,
						}
					}
					//
					if mem != nil {
						// tablePerson.Vitamin = public.SwitchVitaminToF64(mem.UVitamin)
						if mem.Partner > 0 && mem.Partner != 1 {
							tablePerson.Partner = fmt.Sprint(mem.Partner)
						}
					}
					result.Person = append(result.Person, tablePerson)
				}
			}
		}

	}
	if IsGameSupportWatch(table.KindId) {
		result.CanWatch = true
		b, cer := static.IsLookOnSupport(table.Config.GameConfig)
		if cer == nil {
			result.CanWatch = b
		}
	}

	return xerrors.SuccessCode, &result
}

// 向客户端发送所有公告内容
func Proto_SendDialogNoticeList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	ack := make(static.NoticeList, 0)
	ns := GetNoticeDialogs()

	for _, n := range ns {
		if n == nil {
			continue
		}
		if n.ShowType == static.ShowTypeDaily {
			if !IsNoticeShowToday(n.Id, p.Uid) {
				SetNoticeShowToday(n.Id, p.Uid)
				ack = append(ack, n)
			}
		} else {
			ack = append(ack, n)
		}
	}
	return xerrors.SuccessCode, &ack
}

func Proto_UserLimitGame(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseUserLimitGame)
	key := MinorBanPlay
	if req.AllowGame {
		key = MinorAllowPlay
	}
	house, _, optMem, cusErr := inspectClubFloorMemberWithRight(req.HID, -1, p.Uid, consts.ROLE_ADMIN, key)
	if cusErr != xerrors.RespOk {
		if cusErr == xerrors.InvalidParamError {
			if !optMem.IsPartner() {
				return cusErr.Code, cusErr.Msg
			}
		} else {
			return cusErr.Code, cusErr.Msg
		}
	}
	mem := house.GetMemByUId(req.UID)
	if mem == nil {
		return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
	}
	// 自身被禁止娱乐下 不能 恢复其他人娱乐
	if optMem.IsLimitGame && req.AllowGame {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}
	// 自己不能操作自己
	if p.Uid == req.UID {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 队长/副队长的权限处理
	if optMem.IsPartner() || optMem.IsVicePartner() {
		// 确定队长uid
		tgtUId := optMem.UId
		if optMem.IsVicePartner() {
			tgtUId = optMem.Partner
		}
		// 确定队长信息
		tgtPartner := house.GetMemByUId(tgtUId)
		if tgtPartner == nil {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}
		// 队长/副队长 可管理 队长2 下面的成员
		isOK, _ := GetMRightMgr().CheckRight(optMem, MinorManageSuperior)
		if isOK {
			// 队长/副队长既能管理自己名下成员 又可以 管理队长2名下成员
			bIsPartnerMem, nLv := mem.IsHaveTgtSuperior(tgtPartner.UId)
			if bIsPartnerMem == false {
				return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
			}

			// 队长/副队长 不能 操作 二级队长/副队长
			if (optMem.IsPartner() || optMem.IsVicePartner()) && nLv > 1 && (mem.IsPartner() || mem.IsVicePartner()) {
				return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
			}
		} else {
			// 队长/副队长只能管理自己名下成员
			if mem.Partner != tgtPartner.UId {
				return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
			}
		}

		// 上级被禁娱 下级的恢复权限也被暂停（队长禁娱，副队长也不能恢复权限）
		if tgtPartner.IsLimitGame && req.AllowGame {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}

		// 副队长之间 不能 相互管理
		if optMem.IsVicePartner() && mem.IsVicePartner() {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}
	}

	err := house.LimitUserGame(p.Uid, req.UID, req.AllowGame)
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	var msg string
	if req.AllowGame {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主恢复%sID:%d娱乐权限", mem.URemark, mem.UId)
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%sID:%d恢复玩家%sID:%d娱乐权限", optMem.URemark, optMem.UId, mem.URemark, mem.UId)
		} else if optMem.IsPartner() {
			msg = fmt.Sprintf("队长%sID:%d恢复其名下玩家%sID:%d娱乐权限", optMem.URemark, optMem.UId, mem.URemark, mem.UId)
		} else if optMem.IsVicePartner() {
			msg = fmt.Sprintf("副队长%sID:%d恢复其名下玩家%sID:%d娱乐权限", optMem.URemark, optMem.UId, mem.URemark, mem.UId)
		}
	} else {
		if house.DBClub.UId == optMem.UId {
			msg = fmt.Sprintf("盟主暂停%sID:%d娱乐权限", mem.URemark, mem.UId)
		} else if optMem.URole == consts.ROLE_ADMIN {
			msg = fmt.Sprintf("管理员%sID:%d暂停玩家%sID:%d娱乐权限", optMem.URemark, optMem.UId, mem.URemark, mem.UId)
		} else if optMem.IsPartner() {
			msg = fmt.Sprintf("队长%sID:%d暂停其名下玩家%sID:%d娱乐权限", optMem.URemark, optMem.UId, mem.URemark, mem.UId)
		} else if optMem.IsVicePartner() {
			msg = fmt.Sprintf("副队长%sID:%d暂停其名下玩家%sID:%d娱乐权限", optMem.URemark, optMem.UId, mem.URemark, mem.UId)
		}
	}
	go CreateClubMassage(house.DBClub.Id, optMem.UId, GameStatueChange, msg)

	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeLimitUserPlay_NTF, req)

	return xerrors.RespOk.Code, nil
}

func Proto_HouseCaptainLimitGame(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.MsgHouseCaptainLimitGame)
	key := MinorBanPlay
	if req.AllowGame {
		key = MinorAllowPlay
	}
	house, _, partner, cusErr := inspectClubFloorMemberWithRight(req.HID, -1, p.Uid, consts.ROLE_MEMBER, key)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if partner.Lower(consts.ROLE_ADMIN) && !partner.IsPartner() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if partner.IsLimitGame && req.AllowGame {
		return xerrors.ResultErrorCode, "您已被禁止娱乐，暂无操作权限，请联系盟主/管理员处理。"
	}
	mem := house.GetMemByUId(req.UID)
	if mem == nil {
		return xerrors.InValidHouseMemberError.Code, xerrors.InValidHouseMemberError.Msg
	}

	if !mem.IsPartner() {
		return xerrors.InValidHousePartnerError.Code, xerrors.InValidHousePartnerError.Msg
	}

	if partner.IsPartner() && (mem.Partner != p.Uid && mem.Superior != p.Uid) {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	if partner.IsPartner() && partner.IsLimitGame {
		return xerrors.ResultErrorCode, "您已被禁止娱乐，暂无操作权限，请联系盟主/管理员处理。"
	}

	if req.IsTeamMember {
		ids := house.GetUIDsByPartner(nil, mem.UId)
		ids = append(ids, mem.UId)
		err := house.LimitUsersGame(p.Uid, req.AllowGame, ids...)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		var msg string
		if req.AllowGame {
			if house.DBClub.UId == partner.UId {
				msg = fmt.Sprintf("盟主恢复了队长%sID:%d及其名下所有队员的娱乐权限", mem.URemark, mem.UId)
			} else if partner.URole == consts.ROLE_ADMIN {
				msg = fmt.Sprintf("管理员%sID:%d恢复玩家%sID:%d及其名下所有队员的娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.Partner == 1 {
				msg = fmt.Sprintf("队长%sID:%d恢复其名下玩家%sID:%d及其名下所有队员的娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.IsVicePartner() {
				msg = fmt.Sprintf("副队长%sID:%d恢复其名下玩家%sID:%d及其名下所有队员的娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			}
		} else {
			if house.DBClub.UId == partner.UId {
				msg = fmt.Sprintf("盟主暂停%sID:%d及其名下所有队员娱乐权限", mem.URemark, mem.UId)
			} else if partner.URole == consts.ROLE_ADMIN {
				msg = fmt.Sprintf("管理员%sID:%d暂停玩家%sID:%d及其名下所有队员娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.Partner == 1 {
				msg = fmt.Sprintf("队长%sID:%d暂停其名下玩家%sID:%d及其名下所有队员娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.IsVicePartner() {
				msg = fmt.Sprintf("副队长%sID:%d暂停其名下玩家%sID:%d及其名下所有队员娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			}
		}
		go CreateClubMassage(house.DBClub.Id, partner.UId, GameStatueChange, msg)
	} else {
		err := house.LimitUserGame(p.Uid, req.UID, req.AllowGame)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		var msg string
		if req.AllowGame {
			if house.DBClub.UId == partner.UId {
				msg = fmt.Sprintf("盟主恢复%sID:%d娱乐权限", mem.URemark, mem.UId)
			} else if partner.URole == consts.ROLE_ADMIN {
				msg = fmt.Sprintf("管理员%sID:%d恢复玩家%sID:%d娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.Partner == 1 {
				msg = fmt.Sprintf("队长%sID:%d恢复其名下玩家%sID:%d娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.IsVicePartner() {
				msg = fmt.Sprintf("副队长%sID:%d恢复其名下玩家%sID:%d娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			}
		} else {
			if house.DBClub.UId == partner.UId {
				msg = fmt.Sprintf("盟主暂停%sID:%d娱乐权限", mem.URemark, mem.UId)
			} else if partner.URole == consts.ROLE_ADMIN {
				msg = fmt.Sprintf("管理员%sID:%d暂停玩家%sID:%d娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.Partner == 1 {
				msg = fmt.Sprintf("队长%sID:%d暂停其名下玩家%sID:%d娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			} else if partner.IsVicePartner() {
				msg = fmt.Sprintf("副队长%sID:%d暂停其名下玩家%sID:%d娱乐权限", partner.URemark, partner.UId, mem.URemark, mem.UId)
			}
		}
		go CreateClubMassage(house.DBClub.Id, partner.UId, GameStatueChange, msg)
	}
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeLimitCaptainPlay_NTF, req)
	return xerrors.RespOk.Code, nil
}

func Proto_TableUserKick(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_TableUserKick)
	table := GetTableMgr().GetTable(req.TId)
	if table == nil {
		// 牌桌不存在
		return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
	}
	// 鉴权
	if table.IsTeaHouse() {
		_, floor, _, cer := inspectClubFloorMemberWithRight(table.HId, table.FId, p.Uid, consts.ROLE_ADMIN, MinorRoomKickout)
		if cer != xerrors.RespOk {
			return cer.Code, cer.Msg
		}
		hft := floor.GetHftByTid(table.Id)
		if hft == nil {
			xlog.Logger().Errorf("get table not in house on kick:%d", table.Id)
			return xerrors.TableNotExistError.Code, xerrors.TableNotExistError.Msg
		}
	} else {
		if table.Creator != p.Uid {
			return xerrors.InvalidParamError.Code, xerrors.InvalidParamError.Msg
		}
	}
	err := table.Kick(req.Uid, p.Uid)
	if err != nil {
		return xerrors.ResultErrorCode, fmt.Sprintf("踢出失败，%s", err.Error())
	}
	return xerrors.SuccessCode, nil
}

// 任务系统

// 任务列表
func Proto_TaskList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_CH_TaskList)

	ack := new(static.Msg_HC_TaskList)
	ack.Items = make([]*static.TaskItem, 0)

	tl := GetTasksMgr().GetUserTaskList(p.Uid)
	if tl == nil {
		return xerrors.TaskNotExistError.Code, xerrors.TaskNotExistError.Msg
	}

	// 遍历任务
	for _, task := range tl {
		taskCfg := GetTasksMgr().GetTaskConfig(task.TcId)
		if taskCfg == nil {
			continue
		}

		// 任务主类型
		if taskCfg.MainType != req.TaskType {
			continue
		}

		// 任务是否和区域绑定
		if len(taskCfg.Area) > 0 && p.Area != taskCfg.Area {
			continue
		}

		// 小程序不显示分享任务
		if p.Platform == consts.PlatformWechatApplet && taskCfg.SubType == consts.TASK_KIND_SHARE {
			continue
		}

		// 小程序限制展示游戏个数
		if p.Platform == consts.PlatformWechatApplet {
			if taskCfg.SubType == consts.TASK_KIND_GAME_ROUND || taskCfg.SubType == consts.TASK_KIND_GAME_WIN {
				bFind := false
				for _, kid := range AppletGoldShowGame {
					if kid == taskCfg.GameKindId {
						bFind = true
						break
					}
				}
				if !bFind {
					continue
				}
			}
		}

		item := new(static.TaskItem)
		item.Id = task.TcId
		item.MainType = taskCfg.MainType
		item.SubType = taskCfg.SubType
		item.TgtNum = taskCfg.TgtCompleteNum
		item.GameKindId = taskCfg.GameKindId
		item.Desc = taskCfg.Desc
		item.RewardDesc = taskCfg.RewardDesc
		item.Order = taskCfg.Sort
		item.Num = task.Num
		item.Sta = task.Sta

		ack.Items = append(ack.Items, item)
	}

	return xerrors.SuccessCode, ack
}

// 任务提交
func Proto_TaskCommit(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_TaskCommit)

	var tgtTask *static.Task
	tl := GetTasksMgr().GetUserTaskList(p.Uid)
	if tl == nil {
		return xerrors.TaskNotExistError.Code, xerrors.TaskNotExistError.Msg
	}

	if req.TaskId > 0 {
		tgtTask = GetTasksMgr().GetUserTask(p.Uid, req.TaskId)
	} else {
		for _, task := range tl {
			tc := GetTasksMgr().GetTaskConfig(task.TcId)
			if tc == nil {
				continue
			}
			if tc.SubType == req.TaskType {
				tgtTask = task
				break
			}
		}
	}

	if tgtTask == nil {
		return xerrors.TaskNotExistError.Code, xerrors.TaskNotExistError.Msg
	}

	if tgtTask.Sta != consts.TASK_STA_DOING {
		return xerrors.TaskStaExceptionError.Code, xerrors.TaskStaExceptionError.Msg
	}

	GetTasksMgr().UpdateTaskSta(p.Uid, tgtTask.TcId, req.Num, nil, true)

	return xerrors.SuccessCode, nil
}

// 任务奖励领取
func Proto_TaskReward(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_CH_TaskReward)

	task := GetTasksMgr().GetUserTask(p.Uid, req.TaskId)
	if task == nil {
		return xerrors.TaskNotExistError.Code, xerrors.TaskNotExistError.Msg
	}

	if task.Sta != consts.TASK_STA_COMPLETED {
		return xerrors.TaskStaExceptionError.Code, xerrors.TaskStaExceptionError.Msg
	}

	// 读取奖励配置
	taskCfg := GetTasksMgr().GetTaskConfig(task.TcId)
	if taskCfg == nil {
		return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
	}

	// 发放奖励
	for _, reward := range taskCfg.Reward {
		_, err := updateWealth(p.Uid, int8(reward.WealthType), reward.Num, models.CostTaskReward)
		if err != nil {
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}

		// 领取日志
		record := new(models.UserTaskRewardLog)
		record.TaskId = task.TcId
		record.Type = taskCfg.MainType
		record.Kind = taskCfg.SubType
		record.KindId = taskCfg.GameKindId
		record.Uid = p.Uid
		record.Step = task.Step
		record.RewardType = reward.WealthType
		record.RewardNum = reward.Num
		if err := GetDBMgr().GetDBmControl().Create(record).Error; err != nil {
			xlog.Logger().Error("taskrewardlog err :", err)
		}
	}

	// 更新任务状态
	task.Sta = consts.TASK_STA_RECEIVED
	// 更新后台任务数据
	if err := GetDBMgr().UpdateUserTask(task); err != nil {
		xlog.Logger().Errorln("UpdateUserTask err :", err.Error())
	}

	var ack static.Msg_HC_TaskAward
	item := new(static.TaskItem)
	item.Id = task.TcId
	item.MainType = taskCfg.MainType
	item.SubType = taskCfg.SubType
	item.TgtNum = taskCfg.TgtCompleteNum
	item.Desc = taskCfg.Desc
	item.RewardDesc = taskCfg.RewardDesc
	item.Order = taskCfg.Sort
	item.GameKindId = taskCfg.GameKindId
	item.Num = task.Num
	item.Sta = task.Sta
	ack.Task = item

	return xerrors.SuccessCode, ack
}

// 获取兑换商品列表
func Proto_ShopProduct(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return GetShopMgr().GetShopProduct(s, p, data)
}

func Proto_PayOrderId(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Shop_Exchange)
	var shopItem models.ConfigShop
	if err := GetDBMgr().GetDBmControl().First(&shopItem, "id = ? and deleted = 0 ", req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerrors.ShopProductOffShelfError.Code, xerrors.ShopProductOffShelfError.Msg
		}
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	paymentOrder := &models.PaymentOrder{
		TradeNo:    fmt.Sprintf("%d_%d", p.Uid, time.Now().UnixMilli()),
		UserId:     p.Uid,
		WealthType: consts.WealthTypeGold,
		ShopId:     shopItem.Id,
		Price:      shopItem.Price,
		Num:        shopItem.Num,
		Gift:       shopItem.Gift,
		Currency:   "CNY",
		Remark:     shopItem.Name,
		Status:     models.OrderStatusInit,
	}
	if err := GetDBMgr().GetDBmControl().Create(paymentOrder).Error; err != nil {
		xlog.Logger().Error("Create paymentOrder error: ", err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	resp := &static.Msg_Payment_OrderId{
		TradeNo: paymentOrder.TradeNo,
	}
	return xerrors.SuccessCode, &resp
}

func Proto_PayOrder(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Payment_OrderId)
	app, err := CreateOrder(context.Background(), req.TradeNo)
	if err != nil {
		xlog.Logger().Error("CreateOrder error: ", err)
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	return xerrors.SuccessCode, app
}

// 兑换商品
func Proto_ShopExchange(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return GetShopMgr().exchange(s, p, data)
}

// 获取兑换记录列表
func Proto_ShopRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return GetShopMgr().GetShopRecord(s, p, data)
}

// 获取礼卷奖励列表
func Proto_GoldRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return GetShopMgr().GetGoldRecord(s, p, data)
}

// 绑定兑换手机
func Proto_ShopPhoneBind(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return GetShopMgr().PhoneBind(s, p, data)
}

// 获取绑定兑换手机
func Proto_GetShopPhoneBind(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	return GetShopMgr().GetBindPhone(s, p, data)
}

// func ShopExchange(s *Session, p *public.Person, data interface{}) (code int16, v interface{}) {
//
// 	return GetShopMgr().exchange(s, p, data)
// }

/*
// 每日签到
func Checkin(s *Session, p *public.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*public.Msg_Checkin)

	// 返回结果
	ack := new(public.Msg_CheckinAward)
	ack.Checkin = req.Checkin

	if req.Checkin {
		// 参数校验
		if req.Type != 1 && req.Type != 2 {
			req.Type = 1
			// return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
		}

		// 判断双倍领取
		double := 1
		costType := int8(model.CostTypeCheckin)
		if req.Type == 2 {
			double = 2
			costType = model.CostTypeCheckinShare
		}

		// 获取任务
		tasks := GetTasksMgr().GetTasksBySubType(p.Uid, constant.TASK_KIND_SIGN)
		if len(tasks) <= 0 {
			return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
		}

		for _, task := range tasks {

			// 如果用户断签, 初始化任务阶段
			lastDate := time.Unix(task.Time, 0)
			todayDate := time.Now()
			t1 := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 0, 0, 0, 0, time.Local)
			t2 := time.Date(todayDate.Year(), todayDate.Month(), todayDate.Day(), 0, 0, 0, 0, time.Local)
			subDay := int(t2.Sub(t1).Hours() / 24)
			if subDay > 1 {
				// 断签情况
				// task.Step = 0
				// task.Num = 0
				// task.flush()
			} else if subDay == 1 {
				// 连续签到中, 今日未签到
			} else {
				// 当日已签到
				custerr := xerrors.NewXError("本日已经签到过, 无法重复签到")
				return custerr.Code, custerr.Msg
			}

			taskCon := GetTasksMgr().GetTaskConfig(task.TaskId)
			if taskCon == nil {
				return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
			}

			// 签到发放奖励
			tsConV, ok := taskCon.V.(*TaskWeekly)
			if !ok {
				return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
			}

			if task.Step > len(tsConV.ReWards)-1 {
				return xerrors.InvalidIdError.Code, xerrors.InvalidIdError.Msg
			}

			rewards := tsConV.ReWards[task.Step]
			// 发放奖励
			var buf bytes.Buffer
			for i := 0; i < len(rewards); i++ {
				reward := rewards[i]
				if err := deleteWealth(p.Uid, reward.WealthType, reward.Num*double, costType); err != nil {
					syslog.Logger().Error(err)
					return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
				}
				ack.Double = double
				if i > 0 {
					buf.WriteString("+")
				}
				buf.WriteString(fmt.Sprintf("%sx%d", constant.WealthTypeString(reward.WealthType), reward.Num*double))
			}
			ack.AwardName = buf.String()

			task.Time = time.Now().Unix()
			task.Num = task.Num + 1
			if task.AddStep() {
				// 如果一轮任务完成, 则重新开始
				length := len(taskCon.V.(*TaskWeekly).ReWards)
				if task.Step == length {
					task.Step = 0
					task.Num = 0
				}
				task.flush()
			}

			ts := GetTasksMgr().GetUserTaskList(p.Uid)
			ts.lock.Lock()
			ts.MapTasks[task.TaskId] = task
			ts.lock.Unlock()

			// 更新签到任务
			// GetTasksMgr().ChangeTaskState(p.Uid, task.TaskId, 0, nil)
		}
	}
	// if person := GetPlayerMgr().GetPlayer(p.Uid); person != nil {
	// 	person.checkLowIncome()
	// }
	return xerrors.SuccessCode, ack
}


// 获取见面礼
func GetWelcomeGift(s *Session, p *public.Person, data interface{}) (code int16, v interface{}) {
	// 获取任务配置
	taskConn := GetTasksMgr().GetTaskCfgByType(constant.TASK_TYPE_NORMAL, constant.TASK_KIND_WELCOMEGIFT)
	if taskConn == nil {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 判断用户是否符合日期规范
	tm, _ := time.ParseInLocation("2006-01-02 15:04:05", taskConn.V.(*WelcomeGift).Time, time.Local)
	// 在此时间之前的都认为是老用户
	if p.CreateTime >= tm.Unix() {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 判断用户有没有领取过见面礼
	task := GetTasksMgr().GetUserTask(p.Uid, taskConn.Id)
	if task == nil || task.Step == 1 {
		return xerrors.InvalidPermission.Code, xerrors.InvalidPermission.Msg
	}

	// 更改任务状态
	GetTasksMgr().UpdateTaskSta(p.Uid, taskConn.Id, 0, nil, false)

	return xerrors.SuccessCode, nil
}
*/

// 领取双倍低保
func Proto_DisposeAllowancesDouble(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	// 数据结构
	req := data.(*static.Msg_GetAllowancesDouble)

	// 参数校验
	if req.Current < 0 || req.Current > GetServer().ConServers.Allowances.Num {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 获取低保记录
	record := new(models.UserAllowances)
	if err := GetDBMgr().db_M.Where("uid = ? and date = ?", p.Uid, time.Now().Format("2006-01-02")).Limit(1).Offset(req.Current - 1).First(&record).Error; err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 判断是否领取过奖励
	if record.Double {
		cuserror := xerrors.NewXError("您已经领取过双倍低保奖励, 无法重复领取")
		return cuserror.Code, cuserror.Msg
	}

	// 发放奖励
	tx := GetDBMgr().db_M.Begin()
	_, aftgold, err := wealthtalk.UpdateGold(p.Uid, record.AwardGold, models.CostTypeAllowances, tx)
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	// 更新领取记录
	if err = tx.Model(&record).Update("double", true).Error; err != nil {
		tx.Rollback()
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	// 更新内存
	p.Gold = aftgold
	// 更新redis
	if err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Gold", p.Gold); err != nil {
		tx.Rollback()
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	tx.Commit()

	return xerrors.SuccessCode, static.Msg_S2C_Allowances{Current: req.Current, Gold: record.AwardGold, Remain: GetServer().ConServers.Allowances.Num - req.Current}
}

// 每日礼包领取
func Proto_GetDailyRewards(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {

	var count int
	if err := GetDBMgr().GetDBmControl().Table(models.UserDailyRewards{}.TableName()).Where("uid = ? and date = ?", p.Uid, time.Now().Format("2006-01-02")).Count(&count).Error; err != nil {
		xlog.Logger().Error(err)
	} else {
		if count <= 0 {
			err = disposeUserDailyRewards(p)
			if err != nil {
				xlog.Logger().Error(err)
			} else {
				return xerrors.SuccessCode, nil
			}
		} else {
			return xerrors.DailyRewardError.Code, xerrors.DailyRewardError.Msg
		}
	}
	return xerrors.DailyRewardError.Code, xerrors.DailyRewardError.Msg
}

// 进入金币场
func Proto_EnterGoldGame(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	db := GetDBMgr().GetDBmControl()

	// 金币场用户增加
	gUser, err := models.GoldUserGet(db, p.Uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			models.GoldUserAdd(db, p.Uid, time.Now())
		} else {
			xlog.Logger().Errorf("[查询金币场用户失败]uid = %d:error:%v", p.Uid, err)
			return xerrors.SuccessCode, nil
		}
	} else {
		gUser.LastLoginAt = time.Now()
		//更新用户登录时间
		models.GoldUserSave(db, gUser)
	}

	// 增加登录金币场记录
	models.GoldUserLoginRecordAdd(db, p.Uid, p.Platform, p.Ip, "", time.Now())

	return xerrors.SuccessCode, nil
}

// 设置隐藏地理位置开关
func Proto_RecommendGoldGameGetmeGet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	areaCode := GetDBMgr().SelectPlayLastGameHouseAreaCode(p.Uid)

	gameName := GetDBMgr().SelectRecommentGameByAreaCode(areaCode)

	return xerrors.SuccessCode, static.Msg_S2C_RecommendGoldGame{RecommendGame: gameName}
}

// 场次信息
func Proto_GameSiteList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Game_SiteList)
	ack := getGameSiteList(req.KindId)
	if ack == nil {
		return xerrors.AreaGoldGameNullError.Code, xerrors.AreaGoldGameNullError.Msg
	}
	return xerrors.SuccessCode, ack
}

// 金币场游戏列表
func Proto_GoldGameList(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	ack := getGameList(p.Area)
	if len(ack) <= 0 {
		return xerrors.AreaGoldGameNullError.Code, xerrors.AreaGoldGameNullError.Msg
	}
	return xerrors.SuccessCode, ack
}

/*
// 是否可以签到
func IsCheckin(s *Session, p *public.Person, data interface{}) (code int16, v interface{}) {
	_, isCheckin := GetTasksMgr().ReInitTaskWeekly(p.Uid)
	return xerrors.SuccessCode, &public.Msg_S2C_IsCheckIn{Checkin: isCheckin}
}

// 签到信息
func CheckinInfo(s *Session, p *public.Person, data interface{}) (code int16, v interface{}) {
	step, checkIn := GetTasksMgr().ReInitTaskWeekly(p.Uid)
	taskCon := GetTasksMgr().GetTaskCfgByType(constant.TASK_TYPE_NORMAL, constant.TASK_KIND_SIGN)
	if taskCon != nil {
		// 当日未签到, 返回签到任务信息
		res := new(public.Msg_S2C_TaskCheckin)
		res.Step = step
		res.Checkin = checkIn
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

		s.SendMsg(constant.MsgTypeTaskCheckin, xerrors.SuccessCode, res, p.Uid)

		return xerrors.SuccessCode, nil
	}
	return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
}
*/

// 玩家修改性别
func Proto_UserSetSex(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgC2SUserSetSex)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	if req.Sex == p.Sex {
		return xerrors.SuccessCode, nil
	}
	tx := GetDBMgr().GetDBmControl().Begin()
	var err error
	defer static.TxCommit(tx, err)
	err = tx.Model(&models.User{}).Where("id = ?", p.Uid).Update("sex", req.Sex).Error
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "Sex", req.Sex)
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if per := GetPlayerMgr().GetPlayer(p.Uid); per != nil {
		per.Info.Sex = req.Sex
	}
	p.Sex = req.Sex
	return xerrors.SuccessCode, nil
}

func Proto_GetTHDismissRoomDet(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.MsgC2STHDismissRoomDet)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	recordDismissDet := models.RecordDismiss{}
	err := GetDBMgr().GetDBmControl().Where("game_num = ?", req.GameNum).Find(&recordDismissDet).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	var ack static.Msg_S2C_HouseDismissRoomDet
	ack.DismissTime = recordDismissDet.DismissTime
	ack.DismissType = recordDismissDet.DismissType
	ack.DismissDet = recordDismissDet.DismissDet
	return xerrors.SuccessCode, &ack
}

// 保存玩家的gps信息
func Proto_SaveUserGps(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req, ok := data.(*static.Msg_C2S_GpsInfo)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 校正区域代码的个数
	if len(req.Area) != 6 {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 定位区域 如果没有游戏 则向上一级查找 还是没有的话 就不保存
	// 武汉市-江夏区 420115 没有玩法 校正为 420100
	// https://baike.baidu.com/item/%E8%A1%8C%E6%94%BF%E5%8C%BA%E5%88%92%E4%BB%A3%E7%A0%81/5650987?fr=aladdin
	if len(GetAreaPackagesByCode(static.AreaPackageKindGold, req.Area, false)) == 0 {
		chCode := []byte(req.Area)
		chCode[4] = '0'
		chCode[5] = '0'
		// 更新区域码
		req.Area = string(chCode)
		if len(GetAreaPackagesByCode(static.AreaPackageKindGold, req.Area, false)) == 0 {
			return xerrors.SuccessCode, nil
		}
	}

	// 是否已获取过定位区域码
	if len(p.Area2nd) > 0 {
		return xerrors.SuccessCode, nil
	}

	// 仅保存一次

	// 更新mysql
	if err := GetDBMgr().GetDBmControl().Model(&models.User{}).Where("id = ?", p.Uid).Update("area2nd", req.Area).Error; err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 更新redis
	if err := GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "Area2nd", req.Area); err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 更新缓存信息
	p.Area2nd = req.Area
	if per := GetPlayerMgr().GetPlayer(p.Uid); per != nil {
		per.Info.Area2nd = req.Area
	}

	return xerrors.SuccessCode, nil
}

func Proto_BattleLevel(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	var resp static.Msg_S2C_BattleLevel
	resp.Configs = make([]static.Msg_S2C_BattleLevelConfig, 0, len(GetServer().ConBattleLevel))
	for _, v := range GetServer().ConBattleLevel {
		resp.Configs = append(resp.Configs, static.Msg_S2C_BattleLevelConfig{
			Id:    v.Id,
			Level: v.Level,
			Desc:  v.Desc,
			Limit: v.Limit,
		})
	}
	resp.BattleRound = p.TotalCount
	return xerrors.SuccessCode, &resp
}

func Proto_ActivitySpinInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	var resp static.Msg_S2C_SpinInfo
	resp.Awards = make([]static.Msg_S2C_SpinAward, 0, len(GetServer().ConSpinAward))
	for _, v := range GetServer().ConSpinAward {
		resp.Awards = append(resp.Awards, static.Msg_S2C_SpinAward{
			Id:    v.Id,
			Desc:  v.Desc,
			Type:  v.Type,
			Count: v.Count,
			Icon:  v.Icon,
		})
	}
	chanceKey := fmt.Sprintf("user_activity:spin:times:%d", p.Uid)
	resp.SpinTimes, _ = GetDBMgr().GetDBrControl().RedisV2.Get(chanceKey).Int()
	if resp.SpinTimes <= 0 {
		resp.SpinTimes = 0
	}
	return xerrors.SuccessCode, &resp
}

func Proto_ActivitySpinDo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	chanceKey := fmt.Sprintf("user_activity:spin:times:%d", p.Uid)
	spinTimes, _ := GetDBMgr().GetDBrControl().RedisV2.Get(chanceKey).Int()
	if spinTimes <= 0 {
		return xerrors.NoChanceToDoError.Code, xerrors.NoChanceToDoError.Msg
	}
	sumWeight := 0
	for _, v := range GetServer().ConSpinAward {
		sumWeight += v.Weight
	}
	if sumWeight <= 0 {
		return xerrors.InternalError.Code, xerrors.InternalError.Msg
	}
	rdm := rand.IntN(sumWeight)
	var minWeight int
	var award *models.ConfigSpinAward
	for _, item := range GetServer().ConSpinAward {
		if rdm >= minWeight && rdm < minWeight+item.Weight {
			award = item
			break
		} else {
			minWeight += item.Weight
		}
	}
	if award == nil {
		return xerrors.InternalError.Code, xerrors.InternalError.Msg
	}
	afterTimes, err := GetDBMgr().GetDBrControl().RedisV2.IncrBy(chanceKey, -1).Result()
	if err != nil {
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if afterTimes < 0 {
		afterTimes = 0
	}
	record := &models.RecordSpinAward{
		Uid:       p.Uid,
		CreatedAt: time.Now(),
		AwardId:   award.Id,
		Seq:       award.Seq,
		Desc:      award.Desc,
		Type:      award.Type,
		Count:     award.Count,
		Weight:    award.Weight,
		Icon:      award.Icon,
	}
	GetDBMgr().GetDBmControl().Create(record)
	var resp static.Msg_S2C_SpinResult
	resp.SpinTimes = int(afterTimes)
	resp.AwardId = award.Id
	if award.Type != 0 {
		return xerrors.SuccessCode, &resp
	}
	afterNum, err := updateWealth(p.Uid, consts.WealthTypeGold, int(award.Count), models.CostTypeActivitySpin)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	resp.VitaminChanged = true
	resp.Vitamin = afterNum
	return xerrors.SuccessCode, &resp
}

func Proto_ActivityCheckinInfo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	res, err := getCheckin(p.Uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	var resp static.Msg_S2C_CheckinInfo
	resp.GoldAwards = make([]int64, 0, len(GetServer().ConCheckIn))
	for _, _item := range GetServer().ConCheckIn {
		resp.GoldAwards = append(resp.GoldAwards, _item.Gold)
	}
	resp.MsgCheckIn = res
	return xerrors.SuccessCode, &resp
}

func Proto_ActivityCheckinDo(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	res, err := getCheckin(p.Uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if res.Has {
		return xerrors.CheckinHasError.Code, xerrors.CheckinHasError.Msg
	}
	curDay := res.CurDay + 1
	if curDay > 6 {
		curDay = 1
	}
	var gold int64
	for _, v := range GetServer().ConCheckIn {
		if curDay == v.Id {
			gold = v.Gold
			break
		}
	}
	err = doCheckin(p.Uid, curDay-1)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	record := &models.RecordCheckin{
		Uid:       p.Uid,
		CreatedAt: time.Now(),
		Day:       curDay,
		Gold:      gold,
	}
	GetDBMgr().GetDBmControl().Create(record)
	var resp static.Msg_S2C_CheckinResult
	resp.CurDay = res.CurDay
	resp.GoldAward = gold
	if resp.GoldAward > 0 {
		afterNum, err := updateWealth(p.Uid, consts.WealthTypeGold, int(resp.GoldAward), models.CostTypeActivityCheckin)
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		resp.VitaminChanged = true
		resp.Vitamin = afterNum
	}
	return xerrors.SuccessCode, &resp
}

func getCheckin(uid int64) (*static.MsgCheckIn, error) {
	key := fmt.Sprintf("user_activity:checkin:info:%d", uid)
	var fields = []string{"lastAt", "curDay"}
	res, err := GetDBMgr().GetDBrControl().RedisV2.HMGet(key, fields...).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	var (
		date string
		day  int
	)
	if len(res) > 0 {
		if res[0] != nil {
			date, _ = res[0].(string)
		}
	}
	if len(res) > 1 {
		if res[1] != nil {
			dayStr, _ := res[1].(string)
			day = static.HF_Atoi(dayStr)
		}
	}
	var ret static.MsgCheckIn
	if date == time.Now().Format(time.DateOnly) {
		ret.CurDay = day
		ret.Has = true
		if ret.CurDay > 6 {
			ret.CurDay = 6
		} else if ret.CurDay < 0 {
			ret.CurDay = 6
		}
	} else if date == time.Now().AddDate(0, 0, -1).Format(time.DateOnly) {
		ret.CurDay = day + 1
		if ret.CurDay > 6 {
			ret.CurDay = 0
		} else if ret.CurDay < 0 {
			ret.CurDay = 6
		}
	} else {
		ret.CurDay = 0
	}
	// ret.CurDay++
	return &ret, nil
}

func doCheckin(uid int64, day int) error {
	key := fmt.Sprintf("user_activity:checkin:info:%d", uid)
	pipe := GetDBMgr().GetDBrControl().RedisV2.Pipeline()
	defer pipe.Close()
	pipe.HSet(key, "lastAt", time.Now().Format(time.DateOnly))
	pipe.HSet(key, "curDay", day)
	_, err := pipe.Exec()
	if err != nil {
		return err
	}
	return nil
}

func Proto_BattleRank(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	limit := int64(10)
	req, ok := data.(*static.Msg_C2S_BattleRank)
	if !ok {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	myUid := p.Uid
	now := time.Now()
	if req.DayType != 0 {
		now = now.AddDate(0, 0, -1)
	}
	rankName := "totalround"
	if req.RankType != 0 {
		rankName = "winround"
	}
	rankKey := fmt.Sprintf("rank:%s:%s", rankName, now.Format(time.DateOnly))
	xlog.Logger().Infof("rank key: %s", rankKey)
	cli := GetDBMgr().GetDBrControl().RedisV2
	result, err := cli.ZRevRangeWithScores(rankKey, 0, limit-1).Result()
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	resp := &static.Msg_S2C_BattleRank{
		MyRank:  0,
		MyRound: 0,
		List:    make([]*static.Msg_S2C_BattleRankItem, 0, len(result)),
	}
	userIds := make([]int64, 0, len(result))
	for i, z := range result {
		score := int64(z.Score)
		uid := cast.ToInt64(z.Member)
		userIds = append(userIds, uid)
		resp.List = append(resp.List, &static.Msg_S2C_BattleRankItem{
			Uid:   uid,
			Round: score,
			Rank:  int64(i + 1),
		})
	}
	if len(userIds) > 0 {
		var headInfos []models.UserHeadInfo
		err := GetDBMgr().GetDBmControl().Model(models.User{}).Select("id,nickname,imgurl,sex").Where("id in(?)", userIds).Scan(&headInfos).Error
		if err != nil {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		for _, item := range resp.List {
			for _, headInfo := range headInfos {
				if item.Uid == headInfo.Id {
					item.Nickname = headInfo.Nickname
					item.Imgurl = headInfo.Imgurl
					item.Sex = headInfo.Sex
					break
				}
			}
		}
	}

	score, err := cli.ZScore(rankKey, cast.ToString(myUid)).Result()
	if err != nil {
		if errors.Is(redis.Nil, err) {
		} else {
			xlog.Logger().Error(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
	} else {
		resp.MyRound = int64(score)
		rank, err := cli.ZRevRank(rankKey, cast.ToString(myUid)).Result()
		if err != nil {
			if errors.Is(redis.Nil, err) {
			} else {
				xlog.Logger().Error(err)
				return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
			}
		}
		resp.MyRank = rank + 1
	}
	return xerrors.SuccessCode, resp
}
