package center

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ProtocolWorkers struct {
	router.ProtocolWorker
}

func (pw *ProtocolWorkers) Init() {
	pw.FuncProtocol = make(map[string]*router.ProtocolInfo)

	// 创建牌桌
	pw.RegisterMessage(consts.MsgTypeTableCreate_Ack, static.GH_HTableCreate_Ack{}, pw.Protocol_TableCreate_Ack)
	pw.RegisterMessage(consts.MsgTypeHTableCreate_Ack, static.GH_HTableCreate_Ack{}, pw.Protocol_HTableCreate_Ack)
	// 加入牌桌
	pw.RegisterMessage(consts.MsgTypeTableIn_Ack, static.GH_HTableIn_Ack{}, pw.Protocol_TableIn_Ack)
	pw.RegisterMessage(consts.MsgTypeTableIn_Ntf, static.GH_HTableIn_Ntf{}, pw.Protocol_TableIn_Ntf)
	pw.RegisterMessage(consts.MsgTypeHTableIn_Ack, static.GH_HTableIn_Ack{}, pw.Protocol_HTableIn_Ack)
	// 玩家离开
	pw.RegisterMessage(consts.MsgTypeTableExit_Ntf, static.GH_TableExit_Ntf{}, pw.Protocol_TableOut_Ntf)
	// 解散牌桌
	pw.RegisterMessage(consts.MsgTypeHTableDel_Ack, static.GH_TableDel_Ntf{}, pw.Protocol_TableDel_Ack) // 主动或离开解散
	pw.RegisterMessage(consts.MsgTypeTableDel_Ntf, static.GH_TableDel_Ntf{}, pw.Protocol_TableDel_Ntf)  // 游戏结束解散
	// 牌桌大结算通知
	pw.RegisterMessage(consts.MsgTypeTableRes_Ntf, static.GH_TableRes_Ntf{}, pw.Protocol_TableRes_Ntf) // 游戏结束解散

	// 加盟商卡池增加通知
	pw.RegisterMessage(consts.MsgTypeLeagueCardAdd, static.MsgLeagueCardAdd{}, pw.ProtoLeagueCardAddNTF)
	// 包厢发送玩家入桌邀请
	pw.RegisterMessage(consts.MsgTypeHouseTableInviteSend, static.MsgHouseTableInvite{}, pw.ProtoHouseTableInvite)

	// 定时统计疲劳值仓库数据
	// pw.RegisterMessage(constant.MsgHouseVitaminPoolSum, public.Msg_Null{}, pw.Protocol_HouseVitaminPoolSum)
	// pw.RegisterMessage(constant.MsgHouseMemVitaminSum, public.Msg_Null{}, pw.Protocol_HouseMemVitaminSum)
	pw.RegisterMessage(consts.MsgHouseMemLeftStatistic, static.Msg_Null{}, pw.Protocol_HouseMemLeftStatistic)
	pw.RegisterMessage(consts.MsgHouseAutoPayPartnerMsg, static.Msg_Null{}, pw.Protocol_HousePartnerAutoPay)
	pw.RegisterMessage(consts.MsgHouseSingleAutoPayPartnerMsg, static.MsgSingleHousePartnerAutoPay{}, pw.Protocol_SingleHousePartnerAutoPay)

}

func (pw *ProtocolWorkers) ProtocolParams(params []interface{}) (int64, int16, interface{}) {
	return params[0].(int64), params[1].(int16), params[2]
}

func (pw *ProtocolWorkers) CallGame(gameid int, uid int64, header string, errcode int16, data interface{}) error {

	bytes, err := GetServer().CallGame(gameid, uid, "NewServerMsg", header, errcode, data)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	bytestr := string(bytes)
	if bytestr != "SUC" {
		xlog.Logger().Errorln(fmt.Sprintf("header: %s %s", header, " is not accept"))
		return errors.New(xerrors.UnAcceptRpcError.Msg)
	}

	return nil
}

func (pw *ProtocolWorkers) GetPersonSession(uid int64) (*static.Person, *Session) {
	ph := GetPlayerMgr().GetPlayer(uid)
	var pPerson *static.Person
	var pSession *Session
	if ph != nil {
		pPerson = &ph.Info
		pSession = ph.session
	} else {
		p, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if err != nil {
			pPerson = p
		} else {
			pPerson = nil
		}
		pSession = nil
	}

	return pPerson, pSession
}

// 消息映射
var protocolworkers *ProtocolWorkers = nil

type ServerMethod struct{}

func (self *ServerMethod) ServerMsg(ctx context.Context, args *static.Rpc_Args, reply *[]byte) error {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()
	if args == nil || reply == nil {
		return errors.New("nil paramters !")
	}

	head, _, code, data, ok, uid, _ := static.HF_DecodeMsg(args.MsgData)
	if !ok {
		return errors.New("args err !")
	}
	if head != "getonline" && head != "housememberonline" {
		xlog.Logger().WithFields(logrus.Fields{
			"head": head,
			"data": string(data),
		}).Infoln("【RECEIVED RPC】")
	}

	// 未正常受理
	*reply = []byte("ERR")
	protocolInfo := protocolworkers.GetProtocol(head)
	if protocolInfo != nil {
		protodata := reflect.New(protocolInfo.ProtoType).Interface()
		err := json.Unmarshal(data, protodata)
		if err != nil {
			return errors.New(fmt.Sprintf("rpc %s,%s", head, " no reflect protocoldata."))
		}

		go protocolInfo.ProtoHandler(uid, code, protodata)

		// 已正常受理
		*reply = []byte("SUC")
		return nil
	}
	switch head {
	case "regameserver": //! rg from gameserver
		gamserver := new(static.Msg_GameServer)
		err := json.Unmarshal(data, &gamserver)
		if err != nil {
			xlog.Logger().Errorln("rpc from  regameserver err :", err.Error())
			*reply = []byte("false")
		} else {
			xlog.Logger().Infoln("registe from gameserver ExIp:", gamserver.ExIp)
			if gamserver.Status == consts.ServerStatusOnline {
				GetServer().AddOneGameServer(gamserver.Id, gamserver)
				*reply = []byte("true")
				xlog.Logger().Infoln("servermsg,servermsg, suc head:regameserver:gameinip==", gamserver.InIp)
			} else if gamserver.Status == consts.ServerStatusMaintain {
				GetServer().UpdateServerStatus(gamserver.Id, consts.ServerStatusMaintain)

				// 如果是维护消息 则告知游戏服(停止该服下的所有玩法)
				GetServer().CallGame(gamserver.Id, 0, "NewServerMsg", "servermaintain", xerrors.SuccessCode, &static.Msg_HG_UpdateGameServer{KindId: 0})
			} else if gamserver.Status == consts.ServerStatusOffline {
				GetServer().DelOneGameServer(gamserver.Id)
				*reply = []byte("true")
				xlog.Logger().Infoln("servermsg, fail head:regameserver:gameinip==", gamserver.InIp)
			}
		}

	case "loginserverclose": //! rg from loginserver
		xlog.Logger().Infoln("servermsg, fhead:loginserverclose:")
		*reply = []byte("true")
	case consts.MsgTtypeTableSetBegin: // 更新牌桌局数
		_msg := new(static.Msg_GH_SetBegin)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTtypeTableSetBegin err :", err.Error())
			*reply = []byte("false")
		} else {
			table := GetTableMgr().GetTable(_msg.TableId)
			if table != nil {
				table.Begin = _msg.Begin
				table.Step = _msg.Step
				house := GetClubMgr().GetClubHouseById(table.DHId)
				if house == nil {
					return nil
				}
				floor := house.GetFloorByFId(table.FId)
				if floor == nil {
					return nil
				}
				hft := floor.GetTableByTId(_msg.TableId)
				hft.DataLock.Lock()
				hft.Begin = _msg.Begin
				hft.Step = _msg.Step
				hft.Changed = true
				hft.DataLock.Unlock()
				table.flush()
				//if table.IsAiSuper {
				//	floor.PubRedisMappingNumUpdate()
				//}
				// floor.ReviseUsableTbl()
			}
		}
	case consts.MsgTypeUpdWealth: // 房卡内存数据更新
		_msg := new(static.Msg_UpdWealth)
		err1 := json.Unmarshal(data, &_msg)
		if err1 != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeUpdWealth err :", err1.Error())
			*reply = []byte("false")
		} else {
			// 更新牌桌数据
			if _msg.TableId > 0 {
				table := GetTableMgr().GetTable(_msg.TableId)
				if table == nil {
					*reply = []byte("false")
					return nil
				}
				// 该桌已支付
				table.IsCost = true
				if table.LeagueID > 0 {
					*reply = []byte("true")
					return nil
				}
			}

			// 更新用户数量
			person := GetPlayerMgr().GetPlayer(_msg.Uid)
			if person != nil {
				if _msg.WealthType == consts.WealthTypeCard {
					person.Info.Card = _msg.WealthNum
					person.UpdCard(_msg.CostType)
				} else if _msg.WealthType == consts.WealthTypeGold {
					person.Info.Gold = _msg.WealthNum
					person.UpdGold(_msg.CostType, _msg.WealthNum)
				} else if _msg.WealthType == consts.WealthTypeCoupon {
					person.Info.GoldBean = _msg.WealthNum
					person.UpdCoupon(_msg.CostType)
				} else if _msg.WealthType == consts.WealthTypeDiamond {
					person.Info.Diamond = _msg.WealthNum
					person.UpdDiamond(_msg.CostType)
				}
			}
			if _msg.WealthType == consts.WealthTypeGold {
				// 金币实时通知游戏服更新
				user, err := GetDBMgr().GetDBrControl().GetPerson(_msg.Uid)
				if err == nil && user.GameId > 0 && user.SiteId > 0 {
					if user.TableId > 0 {
						xlog.Logger().Warningf("划扣金币时，玩家%d在金币游戏服%d 牌桌%d中，通知更新...", user.Uid, user.GameId, user.TableId)
						_, _ = GetServer().CallGame(user.GameId, user.Uid, "NewServerMsg", consts.MsgTypeUserScoreUpdate, xerrors.SuccessCode, &static.Msg_Update_Table_user_score{Uid: user.Uid, TableId: user.TableId})
					} else {
						xlog.Logger().Warningf("划扣金币时，玩家%d在游戏服%d 假房间中，通知更新...", user.Uid, user.GameId)
						_, _ = GetServer().CallGame(user.GameId, user.Uid, "NewServerMsg", consts.MsgTypeUserGoldUpdate, xerrors.SuccessCode, &static.Msg_UpdateGold{Uid: user.Uid, CostType: _msg.CostType, Offset: _msg.WealthNum})
					}
				}
			}
			*reply = []byte("true")
		}
	case consts.MsgTypeNoticeUpdate: // 更新公告
		_msg := new(static.Msg_NoticeUpdate)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeNoticeUpdate err :", err.Error())
			*reply = []byte("false")
		} else {
			switch _msg.PositionType {
			case consts.NoticePositionTypeMarquee:
				GetPlayerMgr().Broadcast("marqueenotice", GetNoticeMarqueeNotice())
			case consts.NoticePositionTypeMaintain:
				if t := BroadcastMaintain(); t > 0 {
					go checkMaintainNotice(t)
				}
			case consts.NoticePositionTypeOption:
				ntlist := GetNoticeOptions()
				for _, item := range ntlist {
					// 删除
					err = service2.GetAdminClient().DeleteNotices(item.Id)
					if err != nil {
						xlog.Logger().Errorln(err)
						continue
					}
					var optNotice static.OptNotice
					err := json.Unmarshal([]byte(item.Content), &optNotice)
					if err != nil {
						xlog.Logger().Errorln(err)
						continue
					}
					switch optNotice.Kind {
					// 变更面板玩法
					case consts.NoticeOptionKind_UpdKindPanel:
						{
							// if err := GetAreaMgr().LoadData(); err != nil {
							// 	syslog.Logger().Errorln("AreaMgr LoadData err:", err)
							// }
							var data static.OptGKRule
							err := json.Unmarshal([]byte(optNotice.Data), &data)
							if err != nil {
								xlog.Logger().Errorln(err)
								continue
							}
							// 修改对应玩法版本
							config := GetServer().GetGameConfig(item.KindId)
							if config == nil {
								xlog.Logger().Errorln("option notify not active game. ", item.KindId)
							} else {
								// Memory
								config.Version = data.Grversion

								// // DB
								// if err = GetDBMgr().GetDBmControl().Model(config).Update("version", data.Grversion).Error; err != nil {
								// 	syslog.Logger().Errorln("option notify failed: ", err)
								// }
							}
							// 推送
							GetPlayerMgr().Broadcast("optionnotice", item)
							*reply = []byte("true")
						}
					case consts.NoticeOptionKind_UpdLeague:
						{
							data := static.LeagueUpdate{}
							err := json.Unmarshal([]byte(optNotice.Data), &data)
							if err != nil {
								xlog.Logger().Errorln("rpc from  NoticeOptionKind_UpdLeague err :", err.Error())
								*reply = []byte("false")
							}
							if len(data.IDs) == 0 {
								xlog.Logger().Errorln("参数错误")
								continue
							}
							for _, id := range data.IDs {
								err := GetAllianceMgr().UpdateLeagueInfo(id)
								if err != nil {
									xlog.Logger().Errorln("错误id", id)
									continue
								}
							}
							*reply = []byte("true")
						}
					case consts.NoticeOptionKind_UpdLeagueUser:
						data := static.LeagueUserUpdate{}
						err := json.Unmarshal([]byte(optNotice.Data), &data)
						if err != nil {
							xlog.Logger().Errorln(err)
							continue
						}
						if len(data.IDs) == 0 {
							xlog.Logger().Errorln("id 为空")
							*reply = []byte("false")
						}
						for _, id := range data.IDs {
							err := GetAllianceMgr().UpdateLeagueUser(id)
							if err != nil {
								xlog.Logger().Errorln("id 为空")
							}
						}
					}
				}
			}

			*reply = []byte("true")
		}
	case consts.MsgTypeUserOnline:
		_msg := new(static.Msg_Uid)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeUserOffline err :", err.Error())
			*reply = []byte("false")
		}
		p := GetPlayerMgr().GetPlayer(_msg.Uid)
		if p != nil {
			p.Info.Online = true
		}
		if p != nil && p.Info.TableId > 0 {
			_, floor, _, cuserr := inspectClubFloorMemberWithRight(p.Info.HouseId, _msg.Fid, p.Info.Uid, consts.ROLE_MEMBER, MinorRightNull)
			if cuserr == xerrors.RespOk {
				hft := floor.GetHftByTid(p.Info.TableId)
				if hft != nil {
					hft.UserOnlineChange(p.Info.Uid, true)
				}
			}
		}
		go GetClubMgr().UserInGame(_msg.Uid)
		go GetClubMgr().UserInToHouse(&p.Info, _msg.Hid)
	case consts.MsgTypeUserOffline:
		_msg := new(static.Msg_Uid)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeUserOffline err :", err.Error())
			*reply = []byte("false")
		} else {
			// 更改用户在线状态
			//GetPlayerMgr().userOffline(_msg.Uid)
			//GetPlayerMgr().DelPerson(_msg.Uid)
			p := GetPlayerMgr().GetPlayer(_msg.Uid)
			if p != nil {
				p.Info.Online = false
				p.Info.LastOffLineTime = time.Now().Unix()
			}
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(_msg.Uid, "LastOffLineTime", time.Now().Unix(), "Online", false)
			if p != nil && p.Info.TableId > 0 {
				table := GetTableMgr().GetTable(p.Info.TableId)
				if table == nil {
					return nil
				}
				_, floor, _, cuserr := inspectClubFloorMemberWithRight(p.Info.HouseId, table.FId, p.Info.Uid, consts.ROLE_MEMBER, MinorRightNull)
				if cuserr == xerrors.RespOk {
					hft := floor.GetHftByTid(p.Info.TableId)
					if hft != nil {
						hft.UserOnlineChange(p.Info.Uid, false)
					}
				}
			}
			go GetClubMgr().UserOffLine(uid)

			// 正在执行用户状态操作不允许其余操作
			//GetDBMgr().Redis.Del(fmt.Sprintf("userstatus_doing_exit_%d", _msg.Uid))

			*reply = []byte("true")
		}
	case consts.MsgTypeTableWatchOut_Ntf:
		_msg := new(static.Msg_GH_TableWatchOut)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeTableWatchOut_Ntf err :", err.Error())
			*reply = []byte("false")
		} else {
			GetClubMgr().RemoveHouseTableWatch(_msg.TableId, _msg.Uid)
			GetDBMgr().GetDBrControl().RemoveWatchPlayerToTable(_msg.TableId, _msg.Uid)
			*reply = []byte("true")
		}
	case consts.MsgTypeReloadConfig:
		_msg := new(static.Msg_ReloadConfig)
		json.Unmarshal(data, &_msg)
		if err := reloadConfig(_msg); err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeReloadConfig err :", err.Error())
			*reply = []byte(err.Error())
		} else {
			*reply = []byte("succeed")
		}
	case consts.MsgTypeGetValidKindid:
		{
			var _msg static.Msg_KindIdList
			server := GetServer()
			for _, tmp := range server.ConGame {
				_msg.Kindids = append(_msg.Kindids, tmp.KindId)
			}
			*reply = static.HF_JtoB(&_msg)
		}
	case consts.MsgTypeCompulsoryDiss:
		{
			_msg := new(static.Msg_UserSeat)
			if err := json.Unmarshal(data, &_msg); err != nil {
				return err
			}
			// 直接在redis里面操作该玩家
			if person, err := GetDBMgr().GetDBrControl().GetPerson(_msg.Uid); err != nil {
				return err
			} else if person == nil {
				return fmt.Errorf("redis person[uid:%d] is nil", _msg.Uid)
			} else {
				if person.GameId <= 0 {
					return fmt.Errorf("玩家不在游戏中：%d", _msg.Uid)
				}
				if person.TableId <= 0 {
					return fmt.Errorf("玩家不在游戏中：%d", _msg.Uid)
				}
				_msg.TableId = person.TableId
				if *reply, err = GetServer().CallGame(person.GameId, person.Uid, "NewServerMsg", consts.MsgTypeCompulsoryDiss, xerrors.SuccessCode, _msg); err != nil {
					return err
				}
				if table := GetTableMgr().GetTable(person.TableId); table != nil {
					if table.HId > 0 {
						if house := GetClubMgr().GetClubHouseByHId(table.HId); house != nil {
							if floor := house.GetFloorByFId(table.FId); floor != nil {
								if fTable := floor.GetTableByTId(table.Id); fTable != nil {
									fTable.Clear(floor.DHId)
								}
							}
						}
					}
					GetTableMgr().DelTable(table)
				}
			}
		}
	case consts.MsgTypeGetOnlineNumber: // 获取在线人数
		GetPlayerMgr().GetOnlineNumber()
		numMap := GetPlayerMgr().GetAllKindIdOnlineNumber()
		data, _ := json.Marshal(numMap)
		*reply = data
	case consts.MsgTypeAreaUpdate: // 更新区域列表
		// _msg := new(public.Msg_AreaUpdate)
		// json.Unmarshal(data, &_msg)
		//
		// // 重新获取区域列表数据
		// // err := GetAreaMgr().LoadData()
		// if err != nil {
		// 	syslog.Logger().Errorln("update area failed: ", err.Error())
		// 	*reply = []byte("false")
		// } else {
		// 	// 通知对应区域下的用户
		// 	// GetPlayerMgr().NotifyGameServerChange(_msg.Area)
		// 	*reply = []byte("true")
		// }
	case consts.MsgTypeSiteIn_Ntf: // 用户加入场次成功
		_msg := new(static.Msg_SiteIn_Result)
		json.Unmarshal(data, &_msg)

		hp := GetPlayerMgr().GetPlayer(_msg.Uid)
		if hp == nil {
			rp, err := GetDBMgr().GetDBrControl().GetPerson(_msg.Uid)
			if err != nil {
				xlog.Logger().Errorln("get person eve failed: ", err.Error())
				*reply = []byte("false")
			} else {
				hp := new(PlayerCenterMemory)
				hp.Info = *rp
				hp.Ip = _msg.Ip
				GetPlayerMgr().AddPerson(hp)
				GetPlayerMgr().userOnline(_msg.Uid)
				go GetClubMgr().UserInToHouse(rp, rp.HouseId)
			}
		} else {
			// 更新内存数据
			hp.Info.SiteId = _msg.SiteId
			hp.Info.GameId = _msg.GameId
			GetPlayerMgr().userOnline(_msg.Uid)
			go GetClubMgr().UserInToHouse(&hp.Info, hp.Info.HouseId)
		}
	case consts.MsgGameTaskComplete: // 游戏任务状态更新
		_msg := new(static.Msg_Task_Complete)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Error("rpc from  MsgTtypeTaskComplete err :", err.Error())
			*reply = []byte("false")
		} else {
			GetTasksMgr().UpdateGameTaskSta(_msg.Uid, _msg.Kind, _msg.KindId, _msg.Num, _msg.CardType)
			*reply = []byte("false")
		}
	case consts.MsgTypeHosueIDChange: // 修改包厢hid
		_msg := new(static.Msg_HosueIDChange)
		err := json.Unmarshal(data, &_msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeHosueIDChange err :", err.Error())
			*reply = []byte("false")
		} else {
			err := GetClubMgr().HouseIDChange(_msg)
			if err != nil {
				*reply = []byte(err.Msg)
			} else {
				*reply = []byte("true")
			}
		}

	case consts.MsgTypeOnFewerStart:
		{
			msg := new(static.Msg_GH_OnFewer)
			err := json.Unmarshal(data, &msg)
			if err != nil {
				xlog.Logger().Errorln("rpc from  MsgTypeOnFewerStart err :", err.Error())
				*reply = []byte("false")
				return err
			}
			table := GetTableMgr().GetTable(msg.TableId)
			if table == nil {
				*reply = []byte("false")
				return fmt.Errorf("table is not exist")
			}
			table.OnFewerStart()
			isInArray := func(uid int64) bool {
				for _, v := range msg.ActiveUsers {
					if v == uid {
						return true
					}
				}
				return false
			}
			// 清掉异常玩家
			for _, u := range table.Users {
				if u == nil {
					continue
				}
				if !isInArray(u.Uid) {
					_ = GetDBMgr().GetDBrControl().UpdatePersonAttrs(u.Uid, "TableId", 0, "GameId", 0)
					if p := GetPlayerMgr().GetPlayer(u.Uid); p != nil {
						p.Info.TableId = 0
						p.Info.GameId = 0
					}
				}
			}
			table.InvalidUserFree(isInArray)
			if table.IsTeaHouse() {
				floor := GetClubMgr().GetHouseFloorById(table.DHId, table.FId)
				if floor == nil {
					*reply = []byte("false")
					return fmt.Errorf("house floor is not exist")
				}
				if hft := floor.GetTableByNTId(table.NTId); hft != nil {
					hft.InvalidUserFree(msg.ActiveUsers...)
				}

				// !!异步队列改为同步处理
				GetHouseFloorProtocolMgr().Chopt_TableUpdate(floor, nil, nil, &static.FloorTableId{msg.TableId})
			}
			*reply = []byte("true")
		}
	case consts.MsgTypeDeliveryInfoUpd:
		msg := new(static.Msg_NoticeUpdate)
		err := json.Unmarshal(data, msg)
		if err != nil {
			xlog.Logger().Errorln(err)
			return err
		}
		uid := static.HF_Atoi64(msg.Id)
		img, err := service2.GetDeliveryInfo(GetServer().ConServers.GetImgUrl, uid)
		if err != nil {
			xlog.Logger().Errorln("get user DeliveryInfo error:", err)
			return err
		}

		// 更新数据库
		tx := GetDBMgr().GetDBmControl().Begin()
		// mysql
		if err = tx.Model(&models.User{Id: uid}).Update("delivery_img", img).Error; err != nil {
			xlog.Logger().Errorln("update mysql user DeliveryInfo error:", err)
			tx.Rollback()
			return err
		}

		if p, err := GetDBMgr().GetDBrControl().GetPerson(uid); p == nil || err != nil {
			xlog.Logger().Errorln("get person from redis err:", err, "person eve:", p)
			tx.Rollback()
			return err
		}

		// redis
		if err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "DeliveryImg", img); err != nil {
			xlog.Logger().Errorln("update redis user DeliveryInfo error:", err)
			tx.Rollback()
			return err
		}
		// 提交
		tx.Commit()
		// 如果内存有，更新内存
		if p := GetPlayerMgr().GetPlayer(uid); p != nil {
			p.Info.DeliveryImg = img
			var msg static.Msg_HC_DeliveryInfoUpd
			msg.DeliveryImg = img
			p.SendMsg(consts.MsgTypeDeliveryInfoUpd, &msg)
		}
	case consts.MsgTypeHFInfo:
		{
			msg := new(static.MsgHFloorInfo)
			err := json.Unmarshal(data, msg)
			if err != nil {
				xlog.Logger().Errorln(err)
				return err
			}
			house := GetClubMgr().GetClubHouseById(msg.Hid)
			if house == nil {
				return nil
			}
			floor := house.GetFloorByFId(msg.Fid)
			if floor == nil {
				return nil
			}
			var acks static.GameHfInfo
			acks.IsPartnerApply = house.DBClub.IsPartnerApply
			var mixFloor []*HouseFloor
			if house.DBClub.MixActive && floor.IsMix {
				acks.IsMix = true
				// 混排类型：0手动加桌 1自动加桌 2智能防作弊
				acks.TableJoinType = int(house.DBClub.TableJoinType)
				for _, floor := range house.Floors {
					if floor.IsMix {
						mixFloor = append(mixFloor, floor)
					}
				}
				if sortTableByCreateTime(mixFloor) {
					xlog.Logger().Warnf("sortTableByCreateTime after get hfinfo")
					if time.Now().Unix()-house.LastSyncSorted > 5 {
						house.SyncTablesWithSorted()
					}
				}
				ack := floor.BuildAck(floor.GetAllTables(), false)
				if ack == nil {
					ack = GetFloorDetail(floor)
				}
				game := GetAreaGameByKid(floor.Rule.KindId)
				if game != nil {
					ack.KindName = game.Name
					ack.PackageKey = game.PackageKey
				} else {
					ack.KindName = "未知"
					ack.PackageKey = "uk"
				}
				if len(floor.Name) > 0 {
					ack.KindName = floor.Name
				}
				acks.Infos = append(acks.Infos, ack)

				for _, f := range mixFloor {
					if f.Id == floor.Id {
						continue
					}
					ack := f.BuildAck(f.GetAllTables(), false)
					if ack == nil {
						ack = GetFloorDetail(f)
					}
					game := GetAreaGameByKid(f.Rule.KindId)
					if game != nil {
						ack.KindName = game.Name
						ack.PackageKey = game.PackageKey
					} else {
						ack.KindName = "未知"
						ack.PackageKey = "uk"
					}
					if len(f.Name) > 0 {
						ack.KindName = f.Name
					}
					acks.Infos = append(acks.Infos, ack)
				}
				floor.SaveHft()
			} else {
				ack := floor.BuildAck(floor.GetAllTables(), false)
				if ack == nil {
					ack = GetFloorDetail(floor)
				}
				game := GetAreaGameByKid(floor.Rule.KindId)
				if game != nil {
					ack.KindName = game.Name
				} else {
					ack.KindName = "未知"
				}
				if len(floor.Name) > 0 {
					ack.KindName = floor.Name
				}
				acks.Infos = append(acks.Infos, ack)
			}
			buf, _ := json.Marshal(acks)
			*reply = buf
			return nil
		}
	case consts.MsgUserPhoneChange:
		{
			// 数据解析
			req := &static.GmChangeUserPhone{}
			err := json.Unmarshal(data, req)
			if err != nil {
				xlog.Logger().Errorln(err)
				return err
			}
			if len(req.Phone) != 11 {
				return errors.New("phone error")
			}
			// 手机号码加密处理
			req.Phone, _ = static.HF_EncodeStr(static.HF_Atobytes(req.Phone), static.UserEncodeKey)
			user := GetPlayerMgr().GetPlayer(req.Uid)
			if user != nil {
				user.Info.Tel = req.Phone
			}
			p, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
			if err != nil || p == nil {
				return errors.New("user not exists")
			}
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "Tel", req.Phone)
			sql := `update user set tel = ? where id = ? `
			GetDBMgr().GetDBmControl().Exec(sql, req.Phone, req.Uid)
			*reply = []byte("SUC")
			return nil
		}
	case consts.MsgHouseRevokeGm:
		{
			req := &static.GmHouseRevoke{}
			err := json.Unmarshal(data, req)
			if err != nil {
				xlog.Logger().Errorln(err)
				return err
			}
			hpr := GetClubMgr().HouseRevoke(req.ParentId, req.SonId)
			if hpr.Error() != nil {
				*reply = []byte(hpr.Error().Error())
			} else {
				*reply = []byte("SUC")
			}
			return nil
		}
	case consts.MsgHouseMemKick:
		{
			req := new(static.AdminKickMem)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc from  MsgTypeOnFewerStart err :", err.Error())
				*reply = []byte("false")
				return err
			}
			house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, req.Uid, consts.ROLE_MEMBER, MinorRightNull)
			if cusErr != xerrors.RespOk {
				*reply = []byte("false")
				return fmt.Errorf(cusErr.Msg)
			}

			if mem.UVitamin < 0 {
				*reply = []byte("false")
				return fmt.Errorf(xerrors.VitaminLessZero.Msg)
			}
			xe := house.MemKick(house.DBClub.UId, req.Uid)
			if xe != xerrors.RespOk && xe != nil {
				*reply = []byte("false")
				return fmt.Errorf(xe.Msg)
			}
			*reply = []byte("SUC")
			return nil
		}
	case consts.MsgHouseOwnerChange:
		req := new(static.MsgHouseOwnerChange)
		err := json.Unmarshal(data, &req)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeOnFewerStart err :", err.Error())
			*reply = []byte("false")
			return err
		}
		err = ChangeHouseOwner(req)
		if err != nil {
			*reply = []byte("false")
			return err
		}
		*reply = []byte("SUC")
		return nil
	case consts.MsgGameSwitch:
		req := new(static.MsgHouseGameSwitch)
		err := json.Unmarshal(data, &req)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeOnFewerStart err :", err.Error())
			*reply = []byte("false")
			return err
		}
		p, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)

		if p == nil {
			*reply = []byte("false")
			xlog.Logger().Errorln("person not found :", req)
			return fmt.Errorf("person not exists")
		}
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "admin_game_on", req.On)

		// 入驻包厢id
		items, err := GetDBMgr().ListHouseIdMemberJoin(p.Uid)
		if err != nil {
			xlog.Logger().Errorln(err)
			*reply = []byte("false")
			return err
		}
		for _, item := range items {
			house := GetClubMgr().GetClubHouseById(item.Id)
			if house == nil {
				continue
			}

			// 创建
			if house.DBClub.UId == p.Uid {
				if !req.On {
					house.DBClub.AdminGameOn = false
					house.DBClub.GameOn = false
					house.DBClub.IsVitamin = false //
				} else {
					house.DBClub.AdminGameOn = true
				}
				house.flush()
				if !house.IsBeenMerged() {
					house.Broadcast(consts.ROLE_MEMBER, consts.MsgGameSwitchNtf, static.MsgHouseGameSwitchNtf{house.DBClub.HId, house.DBClub.GameOn, house.DBClub.AdminGameOn, house.DBClub.IsVitamin})
				}
				continue
			}
		}
		*reply = []byte("SUC")
		return nil
	case consts.MsgUserAgentUpdate:
		req := new(static.Msg_GetUserInfo)
		err := json.Unmarshal(data, &req)
		if err != nil {
			xlog.Logger().Errorln("rpc from  MsgUserAgentUpdate err :", err.Error())
			*reply = []byte("false")
			return err
		}

		p, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)
		if p == nil {
			*reply = []byte("false")
			xlog.Logger().Errorln("person not found :", req)
			return fmt.Errorf("person not exists")
		}

		config, xerr := GetDBMgr().GetUserAgentConfig(p.Uid)
		var (
			isVitamin   bool
			isVipFloors bool
		)
		xlog.Logger().Infof("rpc got useragentupdate: select db: res=%#v, err=%#v", config, err)
		if xerr != nil {
			if xerr == xerrors.UserAgentNotConfigError {
				isVitamin = false
				isVipFloors = false
			} else {
				*reply = []byte("false")
				xlog.Logger().Errorln("GetUserAgentConfig error:", xerr.Error())
				return fmt.Errorf("get user agent config from db error: %w", xerr)
			}
		} else {
			isVitamin = config.IsUnion
			isVipFloors = config.IsVipFloors
		}

		isVitamin = true

		// 入驻包厢id
		items, err := GetDBMgr().ListHouseIdMemberJoin(p.Uid)
		if err != nil {
			xlog.Logger().Errorln(err)
			*reply = []byte("false")
			return err
		}
		for _, item := range items {
			house := GetClubMgr().GetClubHouseById(item.Id)
			if house == nil {
				continue
			}

			// 创建
			if house.DBClub.UId == p.Uid {
				if !isVitamin {
					house.DBClub.IsVitamin = false //
					house.flush()
				}
				if !house.IsBeenMerged() {
					house.Broadcast(consts.ROLE_MEMBER, consts.MsgHouseAgentUpdateNtf, &static.MsgHouseAgentUpdateNtf{
						Hid:                house.DBClub.HId,
						VipFloorShowSwitch: isVipFloors,
						UnionSwitch:        isVitamin,
					})
				}
				continue
			}
		}
		*reply = []byte("SUC")
		return nil
	case consts.MsgChangeBlankUser:
		{
			req := new(static.MsgChangeBlankUser)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			p, err := GetDBMgr().GetDBrControl().GetPerson(req.Uid)

			if p == nil {
				*reply = []byte("false")
				xlog.Logger().Errorln("person not found :", req)
				return fmt.Errorf("person not exists")
			}
			db := GetDBMgr().GetDBmControl()
			err = models.AddBlankUser(req.Uid, req.EndTime, req.Reason, req.Bind, db)
			if err != nil {
				xlog.Logger().Errorf("db error:%v", err)
				*reply = []byte("false")
				return err
			}
			blackInfo := models.CheckUserInBlank(req.Uid, db)
			if blackInfo != nil {
				dest := static.MsgBlackUserChangeNtf{Uid: req.Uid, Bind: blackInfo.Status, EndTime: blackInfo.End.Unix(), Reason: blackInfo.Reason}
				SendPersonMsg(req.Uid, consts.MsgBlankUserChangeNTF, dest)
			}

			*reply = []byte("SUC")
			return nil
		}
	case consts.MsgChangeBlankHouse:
		{
			req := new(static.MsgChangeBlankUser)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			house := GetClubMgr().GetClubHouseByHId(req.Hid)
			if house == nil {
				*reply = []byte("false")
				xlog.Logger().Errorln("house not found :", req)
				return fmt.Errorf("house not exists")
			}
			db := GetDBMgr().GetDBmControl()
			err = models.AddBlankHouse(house.DBClub.Id, req.EndTime, req.Reason, req.Bind, db)
			if err != nil {
				xlog.Logger().Errorf("db error:%v", err)
				*reply = []byte("false")
				return err
			}
			*reply = []byte("SUC")
			blackInfo := models.CheckHouseInBlank(house.DBClub.Id, db)
			if blackInfo != nil {
				dest := static.MsgBlackHouseChangeNtf{Hid: req.Hid, Bind: blackInfo.Status, EndTime: blackInfo.End.Unix(), Reason: blackInfo.Reason}
				house.Broadcast(consts.ROLE_MEMBER, consts.MsgBlankHouseChangeNTF, dest)
			}
			return nil

		}
	case consts.MsgTypeSetLogFileLevel:
		{
			req := new(static.MsgSetLogFileLevel)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			if strings.Contains(req.Servers, "hall") {
				xlog.SetFileLevel(req.Level)
			}
			if req.Game != 0 {
				if req.Game > 0 {
					_, er := GetServer().CallGame(req.Game, 0, "NewServerMsg", consts.MsgTypeSetLogFileLevel, 0, req)
					if er != nil {
						*reply = []byte("false")
						return err
					}
				} else if req.Game == -1 {
					GetServer().BroadcastGame(0, "NewServerMsg", consts.MsgTypeSetLogFileLevel, 0, req)
				}
			}
		}
	case consts.MsgHouseTableReset:
		{
			req := new(static.Msg_Http_HouseTableReset)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			_, xErr := ResetHouse(nil, req)
			if xErr == xerrors.RespOk {
				*reply = []byte("true")
			} else {
				if xErr == nil {
					xlog.Logger().Error("nil error")
					*reply = []byte("false")
				} else {
					xlog.Logger().Error(xErr.Msg, req.PassCode)
					*reply = []byte(fmt.Sprintf("%s:pc:%s", xErr.Msg, req.PassCode))
				}
			}
		}
	case consts.MsgResetHmUserRight:
		{
			req := new(static.Msg_Http_ResetUserRight)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			_, xErr := ResetUserHmRight(nil, req)
			if xErr == xerrors.RespOk {
				*reply = []byte("true")
			} else {
				if xErr == nil {
					xlog.Logger().Error("nil error")
					*reply = []byte("false")
				} else {
					xlog.Logger().Error(xErr.Msg, -1)
					*reply = []byte(fmt.Sprintf("%s:pc:%s", xErr.Msg, "-1"))
				}
			}
		}
	case consts.MsgForceHotter:
		{
			req := new(static.Msg_Http_ForceHotter)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			_, xErr := ForceHotter(nil, req)
			if xErr == xerrors.RespOk {
				*reply = []byte("true")
			} else {
				if xErr == nil {
					xlog.Logger().Error("nil error")
					*reply = []byte("false")
				} else {
					xlog.Logger().Error(xErr.Msg, -1)
					*reply = []byte(fmt.Sprintf("%s:pc:%s", xErr.Msg, "-1"))
				}
			}
		}
	case consts.MsgWriteOffUser:
		{
			req := new(static.Msg_Http_WriteOffUser)
			err := json.Unmarshal(data, &req)
			if err != nil {
				xlog.Logger().Errorln("rpc argue err :", err.Error())
				*reply = []byte("false")
				return err
			}
			_, xErr := WriteOffUser(nil, req)
			if xErr == xerrors.RespOk {
				*reply = []byte("true")
			} else {
				if xErr == nil {
					xlog.Logger().Error("nil error")
					*reply = []byte("false")
				} else {
					xlog.Logger().Error(xErr.Msg, -1)
					*reply = []byte(fmt.Sprintf("%s:pc:%s", xErr.Msg, "-1"))
				}
			}
		}
	case consts.MsgTypeSiteExit_Ntf:
		// 内存同步
		if p := GetPlayerMgr().GetPlayer(uid); p != nil {
			p.Info.GameId = 0
			p.Info.SiteId = 0
		}
		*reply = []byte("true")
	case consts.MsgTypeOneStepBuyGift:
		_msg := new(static.MsgGmBankruptcyGift)
		err1 := json.Unmarshal(data, &_msg)
		if err1 != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeBankruptcyGift err :", err1.Error())
			*reply = []byte("false")
		} else {
			user, err2 := GetDBMgr().GetDBrControl().GetPerson(_msg.Uid)
			if err2 != nil {
				xlog.Logger().Errorln("rpc from  MsgTypeBankruptcyGift  GetPlayer err :", err1.Error())
				*reply = []byte("false")
			} else {
				p := GetPlayerMgr().GetPlayer(_msg.Uid)
				for _, detail := range _msg.Wts {
					switch detail.Wt {
					case consts.WealthTypeGold:
						if user.GameId > 0 && user.SiteId > 0 {
							if user.TableId > 0 {
								xlog.Logger().Warningf("领取礼包礼包时，玩家%d在金币游戏服%d 牌桌%d中，通知更新...", user.Uid, user.GameId, user.TableId)
								_, err2 = GetServer().CallGame(user.GameId, user.Uid, "NewServerMsg", consts.MsgTypeUserScoreUpdate, xerrors.SuccessCode, &static.Msg_Update_Table_user_score{Uid: user.Uid, TableId: user.TableId})
								if err2 != nil {
									xlog.Logger().Errorf("通知更新游戏中玩家%d,错误1 %s", user.Uid, err2)
								}
							} else {
								xlog.Logger().Warningf("领取礼包礼包时，玩家%d在游戏服%d 假房间中，通知更新...", user.Uid, user.GameId)
								_, err2 = GetServer().CallGame(user.GameId, user.Uid, "NewServerMsg", consts.MsgTypeUserGoldUpdate, xerrors.SuccessCode, &static.Msg_UpdateGold{Uid: user.Uid, CostType: models.CostTypeBankruptcyGift, Offset: detail.Offset})
								xlog.Logger().Errorf("通知更新游戏中玩家%d,错误2 %s", user.Uid, err2)
							}
						}
						if p != nil {
							p.Info.Gold = user.Gold
							p.UpdGold(models.CostTypeBankruptcyGift, detail.Offset)
						}
					case consts.WealthTypeCard:
						if p != nil {
							p.Info.Card = user.Card
							p.UpdCard(models.CostTypeBankruptcyGift)
						}
					case consts.WealthTypeDiamond:
						if p != nil {
							p.Info.Diamond = user.Diamond
							p.UpdDiamond(models.CostTypeBankruptcyGift)
						}
					}
				}
			}
		}
	case consts.MsgTypeToolExchangeHall:
		_msg := new(static.Msg_S_Tool_ToolExchange)
		err1 := json.Unmarshal(data, &_msg)
		if err1 != nil {
			xlog.Logger().Errorln("rpc from  MsgTypeToolExchange err :", err1.Error())
			*reply = []byte("false")
		} else {
			user, err2 := GetDBMgr().GetDBrControl().GetPerson(_msg.Uid)
			if err2 != nil {
				xlog.Logger().Errorln("rpc from  MsgTypeToolExchange  GetPlayer err :", err1.Error())
				*reply = []byte("false")
			} else {
				p := GetPlayerMgr().GetPlayer(_msg.Uid)
				if user.GameId > 0 {
					xlog.Logger().Warningf("兑换记牌器时，玩家%d在游戏服%d 中，通知更新...", user.Uid, user.GameId)
					_, err2 = GetServer().CallGame(user.GameId, user.Uid, "NewServerMsg", consts.MsgTypeToolExchangeHall, xerrors.SuccessCode, _msg)
					xlog.Logger().Errorf("通知更新游戏中玩家%d,错误2 %s", user.Uid, err2)
				}
				if p != nil {
					p.Info.Diamond = user.Diamond
					p.UpdDiamond(models.CostTypeDiamondExchangeRcd)
				}
			}
		}
	}

	return nil
}

func (self *ServerMethod) NewServerMsg(ctx context.Context, args *[]byte, reply *[]byte) error {
	return self.ServerMsg(ctx, &static.Rpc_Args{MsgData: *args}, reply)
}

// ！执行重读配置文件操作
func reloadConfig(msg *static.Msg_ReloadConfig) error {
	GetDBMgr().ReadAllConfig()
	GetTasksMgr().InitConfig()
	GetServer().BroadcastGame(0, "NewServerMsg", consts.MsgTypeReloadConfig, xerrors.SuccessCode, msg)
	return nil
}

func (pw *ProtocolWorkers) Protocol_TableCreate_Ack(params ...interface{}) (code int16, v interface{}) {
	uid, code, data := pw.ProtocolParams(params)

	// 数据解析
	ack := data.(*static.GH_HTableCreate_Ack)
	// 消息接收者
	p := GetPlayerMgr().GetPlayer(uid)
	var person *static.Person
	var session *Session
	if p != nil {
		person = &p.Info
		if p.session != nil {
			session = p.session
		} else {
			session = nil
		}
	} else {
		p, _ := GetDBMgr().GetDBrControl().GetPerson(uid)
		person = p
		session = nil
	}

	// 数据解析
	table := GetTableMgr().GetTable(ack.TId)
	if table == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeTableCreate_Ack, " tid:", ack.TId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	// 处理正常
	if code == xerrors.SuccessCode {

		p.Info.CreateTable = table.Id

		if session != nil {
			var _msg static.Msg_S2C_TableCreate
			_msg.Id = ack.TId
			session.SendMsg(consts.MsgTypeTableCreate, code, _msg, uid)
		}

		return xerrors.AsyncRespErrorCode, nil
	}

	// 错误处理
	xlog.Logger().Debug("header: ", consts.MsgTypeTableCreate_Ack, "errcode: ", code)
	if session != nil {
		session.SendMsg(consts.MsgTypeTableCreate, xerrors.TableInError.Code, xerrors.TableInError.Msg, uid)
	}

	// 回退冻结数据
	tx := GetDBMgr().GetDBmControl().Begin()
	_, aftka, _, aftfka, err := wealthtalk.UpdateCard(person.Uid, 0, -table.Config.CardCost, consts.WealthTypeCard, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		tx.Rollback()
		return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
	} else {
		// 更新db
		tx.Commit()
		// 更新内存
		p := GetPlayerMgr().GetPlayer(person.Uid)
		if p != nil {
			p.Info.Card = aftka
			p.Info.FrozenCard = aftfka
		}
		// 更新redis
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Card", aftka, "FrozenCard", aftfka)
	}

	// 删除内存数据
	GetTableMgr().DelTable(table)

	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) Protocol_HTableCreate_Ack(params ...interface{}) (code int16, v interface{}) {
	uid, code, data := pw.ProtocolParams(params)

	// 数据解析
	ack := data.(*static.GH_HTableCreate_Ack)
	// 消息接收者
	p := GetPlayerMgr().GetPlayer(uid)
	var persion *static.Person
	var session *Session
	if p != nil {
		persion = &p.Info
		if p.session != nil {
			session = p.session
		} else {
			session = nil
		}
	} else {
		p, _ := GetDBMgr().GetDBrControl().GetPerson(uid)
		persion = p
		session = nil
	}

	// 数据解析
	table := GetTableMgr().GetTable(ack.TId)
	if table == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableCreate_Ack, " tid:", ack.TId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 包厢
	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " hid:", table.HId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 楼层
	floor := house.GetFloorByFId(table.FId)
	if floor == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " fid:", table.FId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	// 处理正常
	if code == xerrors.SuccessCode {

		//protocolInfo := &HouseFloorProtocolInfo{constant.MsgTypeHouseTableIn, constant.HFChOptTableCreate_Ack, session, persion, ack}
		//floor.OptPush(protocolInfo)

		p.Info.TableId = table.Id
		p.Info.GameId = table.GameId

		table.lock.CustomLock()
		table.Users[ack.Seat] = &static.TableUser{
			Uid:    ack.Uid,
			JoinAt: ack.JoinAt,
			Payer:  ack.Payer,
		}
		table.lock.CustomUnLock()

		var _msg static.Msg_S2C_TableIn
		_msg.Id = ack.TId
		_msg.GameId = ack.GameId
		_msg.KindId = ack.KindId
		_msg.Ip = ack.Ip
		if session != nil {
			session.SendMsg(consts.MsgTypeHouseTableIn, code, _msg, uid)
		}

		return xerrors.AsyncRespErrorCode, nil
	}

	// 错误处理
	xlog.Logger().Debug("header: ", consts.MsgTypeHTableCreate_Ack, "errcode: ", code)
	if session != nil {
		session.SendMsg(consts.MsgTypeHouseTableIn, xerrors.TableInError.Code, xerrors.TableInError.Msg, uid)
	}

	_ = persion
	//protocolInfo := &HouseFloorProtocolInfo{constant.MsgTypeHouseTableIn, constant.HFChOptTableCreate_Ack, session, persion, data}
	//floor.OptPush(protocolInfo)

	// 回退冻结数据
	player, err := GetDBMgr().GetDBrControl().GetPerson(ack.Payer)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	tx := GetDBMgr().GetDBmControl().Begin()
	_, aftka, _, aftfka, err := wealthtalk.UpdateCard(player.Uid, 0, -table.Config.CardCost, consts.WealthTypeCard, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		tx.Rollback()
		return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
	} else {
		// 更新db
		tx.Commit()
		// 更新内存
		p := GetPlayerMgr().GetPlayer(player.Uid)
		if p != nil {
			p.Info.Card = aftka
			p.Info.FrozenCard = aftfka
		}
		// 更新redis
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(player.Uid, "Card", aftka, "FrozenCard", aftfka)
	}

	// 删除内存数据
	GetTableMgr().DelTable(table)

	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) Protocol_TableIn_Ack(params ...interface{}) (code int16, v interface{}) {

	uid, code, data := pw.ProtocolParams(params)

	// 数据解析
	ack := data.(*static.GH_HTableIn_Ack)
	// 消息接收者
	ph := GetPlayerMgr().GetPlayer(uid)

	var pPerson *static.Person
	var pSession *Session
	if ph != nil {
		pPerson = &ph.Info
		pSession = ph.session
	} else {
		p, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if err != nil {
			pPerson = p
		} else {
			pPerson = nil
		}
		pSession = nil
	}

	// 数据解析
	table := GetTableMgr().GetTable(ack.TId)
	if table == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeTableIn_Ack, " tid:", ack.TId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	if table.IsTeaHouse() {
		// 包厢
		house := GetClubMgr().GetClubHouseByHId(table.HId)
		if house == nil {
			xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " hid:", table.HId, " not exist")
			return xerrors.AsyncRespErrorCode, nil
		}
		// 楼层
		floor := house.GetFloorByFId(table.FId)
		if floor == nil {
			xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " fid:", table.FId, " not exist")
			return xerrors.AsyncRespErrorCode, nil
		}

		if code != xerrors.SuccessCode {
			ack.Uid = -1

		}

		//protocolInfo := &HouseFloorProtocolInfo{constant.MsgTypeHouseTableIn, constant.HFChOptTableIn_Ack, pSession, pPerson, ack}
		//floor.OptPush(protocolInfo)
	}

	if code == xerrors.SuccessCode {

		pPerson.TableId = table.Id
		pPerson.GameId = table.GameId

		table.lock.CustomLock()
		table.Users[ack.Seat] = &static.TableUser{
			Uid:    ack.Uid,
			JoinAt: ack.JoinAt,
			Payer:  ack.Payer,
		}
		table.lock.CustomUnLock()

		if pSession != nil {
			_msg := new(static.Msg_S2C_TableIn)
			_msg.Id = ack.TId
			_msg.GameId = ack.GameId
			_msg.KindId = ack.KindId
			_msg.Ip = ack.Ip
			pSession.SendMsg(consts.MsgTypeTableIn, xerrors.SuccessCode, _msg, uid)
		}
		return xerrors.AsyncRespErrorCode, nil
	} else {
		if pSession != nil {
			pSession.SendMsg(consts.MsgTypeTableIn, code, xerrors.TableInError.Msg, uid)
		}
	}

	if table.IsTeaHouse() && consts.ClubHouseOwnerPay == false && ack.Payer > 0 {
		// 回退冻结数据
		player, err := GetDBMgr().GetDBrControl().GetPerson(ack.Payer)
		if err != nil {
			xlog.Logger().Errorln(err)
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		tx := GetDBMgr().GetDBmControl().Begin()
		_, aftka, _, aftfka, err := wealthtalk.UpdateCard(player.Uid, 0, -table.Config.CardCost, consts.WealthTypeCard, tx)
		if err != nil {
			xlog.Logger().Errorln(err)
			tx.Rollback()
			return xerrors.CreateTableError.Code, xerrors.CreateTableError.Msg
		} else {
			// 更新db
			tx.Commit()
			// 更新内存
			p := GetPlayerMgr().GetPlayer(player.Uid)
			if p != nil {
				p.Info.Card = aftka
				p.Info.FrozenCard = aftfka
			}
			// 更新redis
			GetDBMgr().GetDBrControl().UpdatePersonAttrs(player.Uid, "Card", aftka, "FrozenCard", aftfka)
		}
	}

	// 修改牌桌玩家座位信息
	// person
	pPerson.TableId = 0
	pPerson.GameId = 0
	// table
	table.lock.CustomLock()
	table.Users[ack.Seat] = nil
	table.lock.CustomUnLock()

	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) Protocol_HTableIn_Ack(params ...interface{}) (code int16, v interface{}) {
	uid, code, data := pw.ProtocolParams(params)

	// 数据解析
	ack := data.(*static.GH_HTableIn_Ack)
	// 消息接收者
	ph := GetPlayerMgr().GetPlayer(uid)
	var pPerson *static.Person
	var pSession *Session
	if ph != nil {
		pPerson = &ph.Info
		pSession = ph.session
	} else {
		p, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if err != nil {
			pPerson = p
		} else {
			pPerson = nil
		}
		pSession = nil
	}

	// 数据解析
	table := GetTableMgr().GetTable(ack.TId)
	if table == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " tid:", ack.TId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 包厢
	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " hid:", table.HId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 楼层
	floor := house.GetFloorByFId(table.FId)
	if floor == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableIn_Ack, " fid:", table.FId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	if code != xerrors.SuccessCode {
		ack.Uid = -1
	}

	//protocolInfo := &HouseFloorProtocolInfo{constant.MsgTypeHouseTableIn, constant.HFChOptTableIn_Ack, pSession, pPerson, ack}
	//floor.OptPush(protocolInfo)

	// 添加玩家数据
	table.lock.CustomLock()
	table.Users[ack.Seat] = &static.TableUser{
		Uid:    ack.Uid,
		JoinAt: ack.JoinAt,
		Payer:  ack.Payer,
	}
	table.lock.CustomUnLock()

	if pPerson != nil && pSession != nil {
		if code == xerrors.SuccessCode {
			_msg := new(static.Msg_S2C_TableIn)
			_msg.Id = ack.TId
			_msg.GameId = ack.GameId
			_msg.KindId = ack.KindId
			_msg.Ip = ack.Ip
			if pPerson.HouseId != 0 {
				pSession.SendMsg(consts.MsgTypeHouseTableIn, xerrors.SuccessCode, _msg, pPerson.Uid)
			} else {
				pSession.SendMsg(consts.MsgTypeTableIn, xerrors.SuccessCode, _msg, pPerson.Uid)
			}
		} else {
			if pPerson.HouseId != 0 {
				pSession.SendMsg(consts.MsgTypeHouseTableIn, xerrors.TableInError.Code, xerrors.TableInError.Msg, pPerson.Uid)
			} else {
				pSession.SendMsg(consts.MsgTypeTableIn, xerrors.TableInError.Code, xerrors.TableInError.Msg, pPerson.Uid)
			}
		}
	}
	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) Protocol_TableIn_Ntf(params ...interface{}) (code int16, v interface{}) {
	_, code, data := pw.ProtocolParams(params)

	// 数据解析
	ntf := data.(*static.GH_HTableIn_Ntf)
	// 获取房间信息
	table := GetTableMgr().GetTable(ntf.TableId)
	if table == nil {
		// 牌桌不存在
		xlog.Logger().Errorln("header:", consts.MsgTypeTableIn_Ntf, " tid:", ntf.TableId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	// 入桌成功
	if code == xerrors.SuccessCode {
		hp := GetPlayerMgr().GetPlayer(ntf.Uid)

		if hp != nil {
			// 更新内存
			hp.CloseSession()
			GetPlayerMgr().AddPerson(hp)
			go GetClubMgr().UserInToHouse(&hp.Info, hp.Info.HouseId)
			if hp.Info.TableId > 0 {
				_, floor, _, cuserr := inspectClubFloorMemberWithRight(hp.Info.HouseId, table.FId, hp.Info.Uid, consts.ROLE_MEMBER, MinorRightNull)
				if cuserr == xerrors.RespOk {
					hft := floor.GetHftByTid(ntf.TableId)
					if hft != nil {
						hft.UserOnlineChange(hp.Info.Uid, true)
					}
				}
			}
		} else {
			p, err := GetDBMgr().GetDBrControl().GetPerson(ntf.Uid)
			if p == nil || err != nil {
				return xerrors.AsyncRespErrorCode, nil
			}
			GetPlayerMgr().AddPerson(&PlayerCenterMemory{Info: *p})
			go GetClubMgr().UserInToHouse(p, p.HouseId)
			if p.TableId > 0 {
				_, floor, _, cuserr := inspectClubFloorMemberWithRight(p.HouseId, table.FId, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
				if cuserr == xerrors.RespOk {
					hft := floor.GetHftByTid(p.TableId)
					if hft != nil {
						hft.UserOnlineChange(p.Uid, true)
					}
				}
			}
		}

	} else {
		// 入桌失败
		p, err := GetDBMgr().GetDBrControl().GetPerson(ntf.Uid)
		if p == nil || err != nil {
			return xerrors.AsyncRespErrorCode, nil
		}
		if p.TableId > 0 {
			_, floor, _, cuserr := inspectClubFloorMemberWithRight(p.HouseId, table.FId, p.Uid, consts.ROLE_MEMBER, MinorRightNull)
			if cuserr == xerrors.RespOk {
				hft := floor.GetHftByTid(p.TableId)
				if hft != nil {
					hft.UserStandUp(p.Uid)
				}
			}
		}
		table.UserLeaveTable(ntf.Uid)
	}

	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) Protocol_TableOut_Ntf(params ...interface{}) (code int16, v interface{}) {
	_, code, data := pw.ProtocolParams(params)
	// 数据解析
	msg := data.(*static.GH_TableExit_Ntf)
	defer func() {
		GetDBMgr().Redis.Del(fmt.Sprintf("userstatus_doing_exit_%d", msg.Uid))
		GetPlayerMgr().DelPerson(msg.Uid)
	}()

	// 消息接收者
	ph := GetPlayerMgr().GetPlayer(msg.Uid)
	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Errorln(fmt.Sprintf("header:%s tid:%d not exist", consts.MsgTypeTableExit_Ntf, msg.TableId))
		return xerrors.AsyncRespErrorCode, nil
	}

	user_lock_key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_LOCK, msg.Uid)
	if GetDBMgr().Redis.Exists(user_lock_key).Val() == 1 {
		return xerrors.AsyncRespErrorCode, nil
	}

	if !table.IsTeaHouse() {
		table.UserLeaveTable(msg.Uid)
		return xerrors.AsyncRespErrorCode, nil
	}

	// 包厢
	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeTableExit_Ntf, " hid:", msg.TableId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 包厢楼层
	floor := house.GetFloorByFId(table.FId)
	if floor == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeTableExit_Ntf, " fid:", msg.TableId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 包厢牌桌事件
	var pPerson *static.Person
	var pSession *Session
	if ph != nil {
		pPerson = &ph.Info
		pSession = ph.session
	} else {
		p, err := GetDBMgr().GetDBrControl().GetPerson(msg.Uid)
		if err != nil {
			pPerson = p
		} else {
			pPerson = nil
		}
		pSession = nil
	}
	return ChOptTableOut_Ntf(floor, pSession, pPerson, msg)
}

func (pw *ProtocolWorkers) Protocol_TableDel_Ack(params ...interface{}) (code int16, v interface{}) {
	_, code, data := pw.ProtocolParams(params)

	// 数据解析
	msg := data.(*static.GH_TableDel_Ntf)

	// 牌桌信息
	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableDel_Ack, " hid:", msg.TableId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableDel_Ack, " hid:", table.HId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}
	// 包厢楼层
	floor := house.GetFloorByFId(table.FId)
	if floor == nil {
		xlog.Logger().Errorln("header:", consts.MsgTypeHTableDel_Ack, " fid:", table.FId, " not exist")
		return xerrors.AsyncRespErrorCode, nil
	}

	// 更新玩家数据 用户退出牌桌
	for _, u := range table.Users {
		if u != nil {
			// 更新内存
			hp := GetPlayerMgr().GetPlayer(u.Uid)
			if hp != nil {
				hp.Info.TableId = 0
				hp.Info.GameId = 0
				// 不是包厢牌桌 && 房间类型为自己创建自己玩 && 房主是自己, 则删除创建房间记录
				if !table.IsTeaHouse() && table.CreateType == consts.CreateTypeSelf && table.Creator == hp.Info.Uid {
					hp.Info.CreateTable = 0
				}
			}
		}
	}

	// 销毁牌桌
	GetTableMgr().DelTable(table)
	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) Protocol_TableDel_Ntf(params ...interface{}) (code int16, v interface{}) {
	_, _, data := pw.ProtocolParams(params)

	// 数据解析
	msg := data.(*static.GH_TableDel_Ntf)

	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Warnf("header:%s  tid:%d not exists", consts.MsgTypeTableDel_Ntf, msg.TableId) //用户离开桌子已经清理掉桌子数据，游戏服30分钟之后再次通知解散导致
		return xerrors.AsyncRespErrorCode, nil
	}

	if table.IsTeaHouse() {

		house := GetClubMgr().GetClubHouseByHId(table.HId)
		if house == nil {
			xlog.Logger().Errorln("header:", consts.MsgTypeTableDel_Ntf, " hid:", table.HId, " not exist")
			return xerrors.AsyncRespErrorCode, nil
		}
		floor := house.GetFloorByFId(table.FId)
		if floor == nil {
			xlog.Logger().Errorln("header:", consts.MsgTypeTableDel_Ntf, " fid:", table.FId, " not exist")
			return xerrors.AsyncRespErrorCode, nil
		}
		// 队列数据
		floortable := new(static.FloorTable)
		floortable.NTId = table.NTId
		floortable.TId = table.Id
		floortable.DHId = table.DHId
		floortable.FId = table.FId
		code, _ = ChOptTableDel_Ntf(floor, nil, nil, floortable)
		if code == xerrors.SuccessCode {
			// 更新玩家数据 用户退出牌桌
			for _, u := range table.Users {
				if u != nil {
					table.UserLeaveTable(u.Uid)
					GetPlayerMgr().DelPerson(u.Uid)
				}
			}
			// 删除内存数据
			GetTableMgr().DelTable(table)
		}
		return
	}

	// 更新玩家数据 用户退出牌桌
	for _, u := range table.Users {
		if u != nil {
			table.UserLeaveTable(u.Uid)
			GetPlayerMgr().DelPerson(u.Uid)
		}
	}
	// 删除内存数据
	GetTableMgr().DelTable(table)
	return
}

// TODO 新版疲劳值扣除结算记录等 挪到游戏服务器处理，这个接口游戏服不再调用
func (pw *ProtocolWorkers) Protocol_TableRes_Ntf(params ...interface{}) (code int16, v interface{}) {
	// _, _, data := pw.ProtocolParams(params)
	//
	// // 数据解析
	// msg := data.(*public.GH_TableRes_Ntf)
	//
	// house := GetClubMgr().GetClubHouseByHId(msg.HId)
	// if house == nil {
	// 	syslog.Logger().Errorln("header:", constant.MsgTypeTableRes_Ntf, " hid: ", msg.HId, " not exist")
	// 	return xerrors.AsyncRespErrorCode, nil
	// }
	// mem := house.GetMemByUId(msg.UId)
	// if mem == nil {
	// 	syslog.Logger().Errorln("header:", constant.MsgTypeTableRes_Ntf, " uid: ", msg.UId, " not exist")
	// 	return xerrors.AsyncRespErrorCode, nil
	// }
	// cli := GetDBMgr().RedisLock
	// mem.Lock(cli)
	//
	// // 疲劳值结算
	// VitaminPlayDeduct := int64(msg.WinScore)
	// // 对局结算
	// // VitaminPlayDeduct -= house.VitaminDeductCount
	// // 大赢家结算
	// // if msg.IsBigWin == 1 {
	// // 	VitaminPlayDeduct -= house.VitaminDeductCount
	// // }
	//
	// befVitamin, aftVitamin, _ := mem.VitaminIncrement(msg.UId, VitaminPlayDeduct, model.GameCost, nil) // 对局扣除
	// mem.Unlock(cli)
	//
	// // db_log
	// // 记录用户财富消耗流水
	// record := new(model.UserWealthCost)
	// record.Uid = mem.UId
	// record.WealthType = constant.WealthTypeVitamin
	// record.CostType = model.CostTypeGame
	// record.Cost = VitaminPlayDeduct
	// record.BeforeNum = befVitamin
	// record.AfterNum = aftVitamin
	// tx := GetDBMgr().GetDBmControl().Begin()
	// if err := tx.Create(&record).Error; err != nil {
	// 	tx.Rollback()
	// 	syslog.Logger().Errorln(constant.MsgTypeTableRes_Ntf, err.Error())
	// }
	// tx.Commit()
	// 记录玩家包厢牌桌的历史玩家
	return xerrors.AsyncRespErrorCode, nil
}

func (pw *ProtocolWorkers) ProtoLeagueCardAddNTF(params ...interface{}) (code int16, v interface{}) {
	_, _, data := pw.ProtocolParams(params)

	// 数据解析
	msg := data.(*static.MsgLeagueCardAdd)

	if msg.Uid > 0 && msg.NotPool {
		ul := UserLeague{Uid: msg.Uid}
		ul.NotifyCardPool(msg.UpdCount)
	} else {
		league := AllianceBiz{AllianceBizID: msg.LeagueID}
		league.NotifyCardPool(msg.UpdCount)
	}
	return xerrors.SuccessCode, nil
}

func (pw *ProtocolWorkers) ProtoHouseTableInvite(params ...interface{}) (code int16, v interface{}) {
	_, _, data := pw.ProtocolParams(params)

	// 数据解析
	msg := data.(*static.MsgHouseTableInvite)

	table := GetTableMgr().GetTable(msg.TId)
	if table == nil {
		xlog.Logger().Errorln("发起邀请的桌子为空")
		return xerrors.InValidHouseTableError.Code, xerrors.InValidHouseTableError.Msg
	}

	house, floor, _, xerr := inspectClubFloorMemberWithRight(table.HId, table.FId, msg.Inviter.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if xerr != xerrors.RespOk {
		return xerr.Code, xerr.Msg
	}

	area_game := GetAreaGameByKid(table.KindId)
	if area_game == nil {
		xlog.Logger().Errorln("发起邀请的桌子找不到区域信息", table.KindId)
		return xerrors.ResultErrorCode, xerrors.NewXError(fmt.Sprintf("区域游戏包不存在:%d", table.KindId))
	}

	if static.IsAnonymous(table.Config.GameConfig) {
		msg.Inviter.Nickname = "匿名昵称"
		msg.Inviter.Imgurl = ""
	}

	// 邀请函
	invitationLetter := &static.Msg_HC_HTInvitationLetter{
		TId:       table.Id,
		NTId:      table.NTId,
		HId:       table.HId,
		DHId:      table.DHId,
		FId:       table.FId,
		NFId:      table.NFId,
		KindId:    table.KindId,
		GameName:  area_game.Name,
		FName:     floor.Name,
		Inviter:   msg.Inviter,
		IsHidHide: house.DBClub.IsHidHide,
	}

	// 受邀人为0 则认为邀请包厢全部成员
	if msg.Invitee == 0 {
		// 广播邀请函
		house.BroadcastMsg(consts.MsgTypeHouseTableInviteRecv, invitationLetter, func(member *HouseMember) bool {
			// 只给正式成员发送消息
			if member.Lower(consts.ROLE_MEMBER) {
				return false
			}
			if !member.IsOnline {
				return false
			}
			// 不能邀请自己
			if member.UId == msg.Inviter.Uid {
				return false
			}
			// 校验能否发起邀请
			if !member.CanInvite(table.DHId) {
				return false
			}
			return true
		})
	} else {
		mem := house.GetMemByUId(msg.Invitee)
		if mem == nil {
			xlog.Logger().Errorln("受邀包厢成员不存在")
			return xerrors.UserExistError.Code, xerrors.UserExistError.Msg
		}
		if mem.CanInvite(house.DBClub.Id) {
			mem.SendMsg(consts.MsgTypeHouseTableInviteRecv, invitationLetter)
		} else {
			xlog.Logger().Warnf("对方不满足邀请条件。")
			return xerrors.ResultErrorCode, xerrors.NewXError("对方不满足邀请条件。")
		}
	}
	return xerrors.SuccessCode, nil
}

// func (self *ProtocolWorkers) Protocol_HouseVitaminPoolSum(params ...interface{}) (code int16, v interface{}) {
// 	if time.Now().Hour() > 1 { //0点，1点通知更新
// 		return xerrors.SuccessCode, nil
// 	}
// 	start, end := public.GetTimeLastSixHourStartAndEnd(time.Now())
// 	sql := `select distinct hid from house_vitamin_pool_tax_log where status = 0 and created_at > ? and created_at <= ? `
// 	db := GetDBMgr().GetDBmControl()
// 	dest := []int64{}
// 	db.Raw(sql, start, end).Scan(&dest)
// 	syslog.Logger().Infof("update house lenth:%d", len(dest))
// 	for _, hid := range dest {
// 		h := GetClubMgr().GetClubHouseById(hid)
// 		if h == nil {
// 			continue
// 		}
// 		h.BuildHouseTaxSum(time.Now())
// 		h.BuildHouseSendSum(time.Now())
// 	}
// 	return xerrors.SuccessCode, nil
// }

type HouseMemGorm struct {
	Value int64 `gorm:"value"`
	Dhid  int64 `gorm:"dhid"`
	Uid   int64 `gorm:"uid"`
}

func (pw *ProtocolWorkers) Protocol_HouseMemVitaminSum(params ...interface{}) (code int16, v interface{}) {
	// if time.Now().Hour()!= 0 {
	// 	return xerrors.SuccessCode, nil
	// }
	start, end := static.GetTimeLastDayStartAndEnd(time.Now())

	var gn []int64
	db := GetDBMgr().GetDBmControl()
	dest := []HouseMemGorm{}
	err := db.Table("house_member_vitamin_log").Select("SUM(value) as value,dhid,uid").
		Where("status = 0 and type = ? and created_at > ? and created_at <= ?",
			models.GameCost, start, end).Group("dhid").Group("uid").Scan(&dest).Error

	if err != nil || len(gn) != 0 {
		return
	}
	xlog.Logger().Infof("update user lenth:%d", len(dest))
	for _, item := range dest {
		if item.Value == 0 {
			continue //暂时不统计
		}
		h := GetClubMgr().GetClubHouseById(item.Dhid)
		if h == nil {
			continue
		}
		tx := db.Begin()
		sql := `insert into house_member_vitamin_log(dhid,uid,optuid,type,value,created_at) values(?,?,?,?,?,?)`
		err := tx.Exec(sql, item.Dhid, item.Uid, item.Uid, models.GameTotal, item.Value, end).Error
		if err != nil {
			xlog.Logger().Errorf("更新用户疲劳统计失败:%v", err)
			tx.Rollback()
			continue
		}
		tx.Exec(`update house_member_vitamin_log set status = 1 where dhid = ? and uid = ? and status = 0 and 
		type = ? and created_at > ? and created_at <= ? `, item.Dhid, item.Uid, models.GameCost, start, end)
		tx.Commit()

	}
	return xerrors.SuccessCode, nil
}

func (pw *ProtocolWorkers) Protocol_HouseMemLeftStatistic(params ...interface{}) (code int16, v interface{}) {
	//GetClubMgr().StatisticHouseMemVitaminLeftDay()
	return xerrors.SuccessCode, nil
}

func (pw *ProtocolWorkers) Protocol_HousePartnerAutoPay(params ...interface{}) (code int16, v interface{}) {
	hmgr := GetClubMgr()
	hmgr.lock.RLock()
	defer hmgr.lock.RUnlock()
	for _, h := range GetClubMgr().ClubMap {
		if !h.DBClub.IsVitamin {
			continue
		}
		go h.AutoPayPartner()
	}
	return xerrors.SuccessCode, nil
}

func (pw *ProtocolWorkers) Protocol_SingleHousePartnerAutoPay(params ...interface{}) (code int16, v interface{}) {
	_, _, data := pw.ProtocolParams(params)
	req := data.(*static.MsgSingleHousePartnerAutoPay)

	h := GetClubMgr().GetClubHouseByHId(int(req.HId))

	if h == nil {
		xlog.Logger().Errorf("house err:%d", req.HId)
		return xerrors.SuccessCode, nil
	}
	if !h.DBClub.AutoPayPartnrt || !h.DBClub.IsVitamin {
		xlog.Logger().Errorf("house attr error :%v, %v", h.DBClub.AutoPayPartnrt, h.DBClub.IsVitamin)
		return xerrors.SuccessCode, nil
	}

	nowDay := time.Now().Format("20060102150505")
	if h.LastAutoPay == nowDay { //防止admin服务期间重启导致重复付款
		xlog.Logger().Errorf("house autp paied:%d", h.DBClub.Id)
		return xerrors.SuccessCode, nil
	}
	h.AutoPayPartner()
	h.LastAutoPay = nowDay

	return xerrors.SuccessCode, nil
}

func (pw *ProtocolWorkers) AdminKickMem(params ...interface{}) (code int16, v interface{}) {
	_, _, data := pw.ProtocolParams(params)
	req := data.(*static.AdminKickMem)
	house, _, mem, cusErr := inspectClubFloorMemberWithRight(req.Hid, -1, req.Uid, consts.ROLE_MEMBER, MinorRightNull)
	if cusErr != xerrors.RespOk {
		return cusErr.Code, cusErr.Msg
	}

	if mem.UVitamin < 0 {
		return xerrors.VitaminLessZero.Code, xerrors.VitaminLessZero.Msg
	}
	xe := house.MemKick(house.DBClub.UId, req.Uid)
	if xe != xerrors.RespOk {
		return xe.Code, xe.Msg
	}
	return xerrors.SuccessCode, nil
}

func ChangeHouseOwner(req *static.MsgHouseOwnerChange) error {
	house := GetClubMgr().GetClubHouseByHId(req.Hid)
	if house == nil {
		xlog.Logger().Errorln("house nil")
		return fmt.Errorf("%s", xerrors.InValidHouseError.Msg)
	}

	if house.IsBeenMerged() {
		xlog.Logger().Error("包厢已被合并,无法更换盟主,请在撤销合并包厢后重试。")
		return fmt.Errorf("包厢:%d已被合并,当前无法更换盟主,请在撤销合并包厢后重试。", house.DBClub.HId)
	}

	oldUid := house.DBClub.UId
	oldOwner := house.GetMemByUId(oldUid)
	if oldOwner != nil {
		oldOwner.URole = 2
	}
	house.DBClub.UId = req.Uid
	mem := house.GetMemByUId(req.Uid)
	if mem == nil {
		cusErr := house.MemJoin(req.Uid, 2, 0, true, nil)
		if cusErr != nil {
			xlog.Logger().Errorln("error:", cusErr.Msg)
			return fmt.Errorf("%s", cusErr.Msg)
		}
		mem = house.GetMemByUId(req.Uid)
		if mem == nil {
			xlog.Logger().Errorln("join failed:")
			return fmt.Errorf("%s", xerrors.DBExecError.Msg)
		}
		mem.URole = 0
	} else {
		mem.URole = 0
	}

	if oldOwner != nil {
		oldOwner.Flush()
		GetMRightMgr().deleteRightByHidUid(int(house.DBClub.Id), oldUid)
	}
	mem.Flush()
	GetMRightMgr().deleteRightByHidUid(int(house.DBClub.Id), req.Uid)
	house.flush()
	xlog.Logger().Errorf("suc change house owner:%d to %d", oldUid, req.Uid)
	return nil
}
