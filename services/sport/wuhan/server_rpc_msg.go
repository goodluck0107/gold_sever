package wuhan

import (
	"context"
	"github.com/open-source/game/chess.git/pkg/xlog"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	"log"
	"reflect"

	"github.com/sirupsen/logrus"

	// "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/protocolworker"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"runtime/debug"

	"github.com/open-source/game/chess.git/pkg/models"

	"github.com/jinzhu/gorm"
)

type ProtocolWorkers struct {
	protocolworker.ProtocolWorker
}

func (self *ProtocolWorkers) Init() {
	self.FuncProtocol = make(map[string]*protocolworker.ProtocolInfo)

	// Req
	// 重复挤号
	self.RegisterMessage(consts.MsgTypeRelogin, static.Msg_HG_Relogin{}, self.Protocol_Relogin)

	// 创建牌桌
	self.RegisterMessage(consts.MsgTypeTableCreate_Req, static.HG_HTableCreate_Req{}, self.Protocol_TableCreate_Req)
	// 创建牌桌
	self.RegisterMessage(consts.MsgTypeHTableCreate_Req, static.HG_HTableCreate_Req{}, self.Protocol_HTableCreate_Req)
	// 加入牌桌
	self.RegisterMessage(consts.MsgTypeHouseTableIn, static.HG_HTableIn_Req{}, self.Protocol_HTableIn)
	// 加入牌桌
	// self.RegisterMessage(constant.MsgTypeHTableIn_Req, public.HG_HTableIn_Req{}, self.Protocol_HTableIn_Req)
	// 加入牌桌
	self.RegisterMessage(consts.MsgTypeTableIn, static.HG_HTableIn_Req{}, self.Protocol_TableIn_Req)
	// 解散牌桌 强制
	self.RegisterMessage(consts.MsgTypeHTableDel_Req, static.Msg_HG_TableDel_Req{}, self.Protocol_TableDel_Req)

}

func (self *ProtocolWorkers) ProtocolParams(params []interface{}) (int64, int16, interface{}) {
	// uid, errcode, data
	return params[0].(int64), params[1].(int16), params[2]
}

func (self *ProtocolWorkers) CallHall(header string, code int16, data interface{}, uid int64) error {
	bytes, err := GetServer().CallHall("NewServerMsg", header, code, data, uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	bytestr := string(bytes)
	if bytestr != "SUC" {
		xlog.Logger().Errorln(fmt.Sprintf("header: %s,%s", header, " is not accept"))
		return errors.New(xerrors.UnAcceptRpcError.Msg)
	}
	return nil
}

var Protocolworkers *ProtocolWorkers = nil

type ServerMethod struct{}

func (self *ServerMethod) ServerMsg(ctx context.Context, args *static.Rpc_Args, reply *[]byte) error {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	// 未正常受理
	*reply = []byte("ERR")

	if args == nil || reply == nil {
		return errors.New("nil paramters !")
	}

	head, _, errcode, data, ok, uid, _ := static.HF_DecodeMsg(args.MsgData)
	if !ok {
		return errors.New("args err !")
	}

	xlog.Logger().WithFields(logrus.Fields{
		"head": head,
		"data": string(data),
	}).Infoln("【RECEIVED RPC】")

	protocolInfo := Protocolworkers.GetProtocol(head)
	if protocolInfo != nil {
		protodata := reflect.New(protocolInfo.ProtoType).Interface()
		err := json.Unmarshal(data, protodata)
		if err != nil {
			return errors.New(fmt.Sprintf("rpc %s,%s", head, " no reflect protocoldata."))
		}
		if head == "housetablein" || head == "htablecreate_req" ||
			head == "tablein" || head == "tablecreate_req" ||
			head == consts.MsgTypeHTableDel_Req || head == consts.MsgTypeTableExit {
			code, _ := protocolInfo.ProtoHandler(uid, errcode, protodata)
			if code == xerrors.SuccessCode {
				// 已正常受理
				*reply = []byte("SUC")
			} else {
				*reply = []byte("ERR")
			}
			return nil
		} else {
			go protocolInfo.ProtoHandler(uid, errcode, protodata)
			// 已正常受理
			*reply = []byte("SUC")
			return nil
		}
	}
	//return errors.New(fmt.Sprintf("rpc ", head, " no reflect protocol."))
	switch head {
	case "closehallserver": //! 大厅服务器
		//go GetServer().ConnectHall()
		xlog.Logger().Debug("rpc form hall closehallserver")
	case "servermaintain": // 更新进程状态
		var msg static.Msg_HG_UpdateGameServer
		err := json.Unmarshal(data, &msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from hall servermaintain:err:", err.Error())
			return err
		}
		// 如果是维护所有服务器 或者 维护当前服务器 则执行
		if msg.GameId <= 0 || int(msg.GameId) == GetServer().Con.Id {
			// 将尚未开始的游戏桌子解散掉
			GetTableMgr().ForceCloseAllTable()
		}
	case consts.MsgTypeTableIn: //! 加入牌桌
		msg := new(static.Msg_HG_TableIn)
		err := json.Unmarshal(data, &msg)
		if err != nil {
			log.Println("rpc from hall MsgTypeTableIn:err:", err.Error())
			*reply = static.HF_JtoB(static.Msg_GH_TableIn{Seat: -1})
			return err
		}
		table := GetTableMgr().GetTable(msg.TableId)
		if table == nil {
			log.Println("can't find table: ", msg.TableId)
			*reply = static.HF_JtoB(static.Msg_GH_TableIn{Seat: -1})
		} else {
			p, err := GetDBMgr().GetDBrControl().GetPerson(msg.Uid)
			if err != nil {
				log.Println(err)
				*reply = static.HF_JtoB(static.Msg_GH_TableIn{Seat: -1})
			} else {
				seat, joinAt, err := table.UserJoinTable(p.Uid, msg.Seat, msg.Payer)
				if err != nil {
					log.Println(err)
					*reply = static.HF_JtoB(static.Msg_GH_TableIn{Seat: -1, JoinAt: joinAt})
				} else {
					*reply = static.HF_JtoB(static.Msg_GH_TableIn{Seat: seat, JoinAt: joinAt})
				}
			}
		}

	case consts.MsgTypeRelogin: // 用户重复登录挤号

		msg := new(static.Msg_HG_Relogin)
		err := json.Unmarshal(data, &msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from hall MsgTypeRelogin:err:", err.Error())
			*reply = []byte("false")
			return err
		}

		//onlyExitGame := fmt.Sprintf("userstatus_doing_exitgame_%d", msg.Uid)
		//defer GetDBMgr().Redis.Del(onlyExitGame)

		// 校验桌子是否存在
		table := GetTableMgr().GetTable(msg.TableId)
		if table == nil {
			xlog.Logger().Errorln("servermsg can't find table: ", msg.TableId)
			*reply = []byte("false")
		} else {
			// 强制用户下线
			p := GetPersonMgr().GetPerson(msg.Uid)
			if p != nil {
				// 通知客户端
				p.SendMsg(consts.MsgTypeRelogin, xerrors.SuccessCode, nil)
				// 断开连接
				p.CloseSession(consts.SESSION_CLOED_FORCE)
			}
			*reply = []byte("true")
		}
	case consts.MsgTypeReloadConfig: //！重新读取配置
		msg := new(static.Msg_ReloadConfig)
		err := json.Unmarshal(data, &msg)
		if err != nil {
			xlog.Logger().Errorln("rpc from hall MsgTypeReloadConfig:err:", err.Error())
			*reply = []byte("false")
			return err
		}
		servermgr := GetServer()
		flag := len(msg.Games) == 0
		if !flag {
			for _, id := range msg.Games {
				if id == servermgr.Con.Id {
					flag = true
					break
				}
			}
		}
		*reply = []byte("true")
		if flag {
			go func() {
				_ = GetDBMgr().ReadAllConfig()
				GetSiteMgr().InitSites(GetServer().GameTypes)
				svr := GetServer()
				var msg static.Msg_GameServer
				msg.Id = svr.Con.Id
				msg.InIp = svr.Con.InHost
				msg.ExIp = svr.Con.Host
				msg.SafeIp = svr.Con.SafeHost
				msg.GameTypes = svr.GameTypes
				msg.Status = consts.ServerStatusOnline
				if _, err := svr.CallHall("NewServerMsg", "regameserver", xerrors.SuccessCode, &msg, 0); err != nil {
					xlog.Logger().Debug("重读数据库配置后，通知大厅失败：", err)
				}
			}()
		}
	case consts.MsgTypeCompulsoryDiss:
		{
			msg := new(static.Msg_UserSeat)
			err := json.Unmarshal(data, &msg)
			if err != nil {
				return err
			}
			if err = GetTableMgr().CompulsoryDissmiss(msg); err == nil {
				*reply = []byte("OK")
			} else {
				*reply = []byte(err.Error())
			}
		}
	case consts.MsgTypeUserAreaBroadcast: // 用户区域广播
		{
			msg := new(static.Msg_HG_UserAreaBroadcast)
			err := json.Unmarshal(data, &msg)
			if err != nil {
				xlog.Logger().Errorln("rpc from hall MsgTypeUserAreaBroadcast:err:", err.Error())
				*reply = []byte("false")
				return err
			}
			GetSiteMgr().SendUserBroadcast(msg.KindId, consts.MsgTypeUserAreaBroadcast, msg.Broadcast)
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
			xlog.SetFileLevel(req.Level)
		}
	case consts.MsgTypeTableExit: // 大厅通知游戏服用户要离开桌子
		{
			msg := new(static.MsgHouseUserExitTable)
			err := json.Unmarshal(data, &msg)
			if err != nil {
				xlog.Logger().Error("参数错误")
				*reply = []byte("ERR")
				return err
			}
			table := GetTableMgr().GetTable(msg.Tid)
			if table != nil {
				if !table.Begin {
					ok := table.game.OnStandup(msg.Uid, msg.Uid)
					if ok {
						*reply = []byte("SUC")
					} else {
						*reply = []byte("ERR")
					}

				} else {
					xlog.Logger().Error("游戏开始不能推出房间")
					*reply = []byte("ERR")
				}
			}
		}
	case consts.MsgTypeTableUserKick: // 大厅通知游戏服用户要离开桌子
		{
			msg := new(static.MsgHouseTableUserKick)
			err := json.Unmarshal(data, &msg)
			if err != nil {
				xlog.Logger().Error("参数错误")
				*reply = []byte("参数错误。")
				return err
			}
			table := GetTableMgr().GetTable(msg.Tid)
			if table != nil {
				if !table.Begin {
					ok := table.game.OnStandup(msg.Uid, msg.Opt)
					if ok {
						*reply = []byte("SUC")
					} else {
						*reply = []byte("操作异常。")
					}
				} else {
					xlog.Logger().Error("游戏开始不能踢出房间")
					*reply = []byte("游戏已开始。")
				}
			}
		}
	case consts.MsgTypeToolExchangeHall:
		{
			msg := new(static.Msg_S_Tool_ToolExchange)
			err := json.Unmarshal(data, &msg)
			if err != nil {
				xlog.Logger().Error("参数错误")
				*reply = []byte("参数错误。")
				return err
			}
			p := GetPersonMgr().GetPerson(msg.Uid)
			if p != nil {
				person, err := GetDBMgr().GetDBrControl().GetPerson(p.Info.Uid)
				if err != nil {
					xlog.Logger().Error(err)
					*reply = []byte(fmt.Sprintf("从redis获取玩家信息失败。%s", err))
				}

				p.SendMsg(consts.MsgTypeToolExchange, xerrors.SuccessCode, msg)

				if person != nil {
					table := GetTableMgr().GetTable(person.TableId)
					if table != nil {
						table.Operator(base2.NewTableMsg(consts.MsgTypeToolExchangeHall, static.HF_JtoA(*msg), msg.Uid, msg))
					}
				}
			} else {
				*reply = []byte("玩家不在游戏服。")
			}
		}
	}

	return nil
}

func (self *ServerMethod) NewServerMsg(ctx context.Context, args *[]byte, reply *[]byte) error {
	return self.ServerMsg(ctx, &static.Rpc_Args{MsgData: *args}, reply)
}

/*
	更新房卡数据(正数增加, 负数减少)
	befka: 操作前的房卡, aftka: 操作后的房卡, beffka: 操作前冻结的房卡, aftfka: 操作后冻结的房卡
*/
func updcard(uid int64, kacost int, frozenkacost int, tx *gorm.DB, costType int8) (befka int, aftka int, beffka int, aftfka int, err error) {
	//if kacost == 0 && frozenkacost == 0 {
	//	return 0, 0, 0, 0, errors.New("cost zero")
	//}

	var user models.User
	user.Id = uid
	if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&user).Error; err != nil {
		xlog.Logger().Errorln(err)
		return 0, 0, 0, 0, err
	}
	// 操作前的房卡
	befka = user.Card
	beffka = user.FrozenCard

	// 更新房卡
	user.Card = user.Card + kacost
	if user.Card < 0 {
		// 余卡不足, 减少至0
		xlog.Logger().Errorln(fmt.Sprintf("用户[%d]余卡不足, 当前房卡[%d], 扣卡[%d]", user.Id, user.Card-kacost, -1*kacost))
		user.Card = 0
	}
	// 更新冻结房卡
	user.FrozenCard = user.FrozenCard + frozenkacost
	if user.FrozenCard < 0 {
		// 冻结房卡有误
		xlog.Logger().Errorln(fmt.Sprintf("用户[%d]冻结房卡有误, 当前已冻结房卡[%d], 加卡[%d]", user.Id, user.FrozenCard-frozenkacost, frozenkacost))
		user.FrozenCard = 0
	}
	// 更新数据库信息
	if err = tx.Model(&user).Update(map[string]interface{}{
		"card":       user.Card,
		"frozencard": user.FrozenCard}).Error; err != nil {
		xlog.Logger().Errorln("update account failed: ", err.Error())
		return 0, 0, 0, 0, err
	}
	// 操作后的房卡
	aftka = user.Card
	aftfka = user.FrozenCard

	if kacost != 0 {
		// 记录用户房卡消耗流水
		record := new(models.UserWealthCost)
		record.Uid = uid
		record.Cost = int64(kacost)
		record.AfterNum = int64(aftka)
		record.BeforeNum = int64(befka)
		record.WealthType = consts.WealthTypeCard
		record.CostType = costType
		if err = tx.Create(&record).Error; err != nil {
			xlog.Logger().Errorln(err)
			return 0, 0, 0, 0, err
		}
	}

	return befka, aftka, beffka, aftfka, nil
}

// 更新加盟商房卡数据(正数增加, 负数减少)
func updLeagueCard(leagueID, uid int64, kacost int, frozenkacost int, costType int8, tx *gorm.DB) (int, int, int, int, error) {
	//if kacost == 0 && frozenkacost == 0 {
	//	return 0, 0, 0, 0, errors.New("kacost & frozenkacost is zero")
	//}

	var league models.League
	league.LeagueID = leagueID
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("league_id = ?", league.LeagueID).First(&league).Error; err != nil {
		xlog.Logger().Errorln(err)
		return 0, 0, 0, 0, err
	}
	befka := league.Card
	beffka := league.FreezeCard
	// 更新房卡
	league.Card = league.Card + int64(kacost)
	if league.Card < 0 {
		// 余卡不足, 减少至0
		xlog.Logger().Errorln(fmt.Sprintf("加盟商[%d]余卡不足, 当前房卡[%d], 扣卡[%d]", league.LeagueID, league.Card-int64(kacost), -1*kacost))
		league.Card = 0
	}
	// 更新冻结房卡
	league.FreezeCard = league.FreezeCard + int64(frozenkacost)
	if league.FreezeCard < 0 {
		// 冻结房卡有误
		xlog.Logger().Errorln(fmt.Sprintf("加盟商[%d]冻结房卡有误, 当前已冻结房卡[%d], 加卡[%d]", league.LeagueID, league.FreezeCard-int64(frozenkacost), frozenkacost))
		league.FreezeCard = 0
	}
	if err := tx.Model(&league).Update(map[string]interface{}{
		"card":        league.Card,
		"freeze_card": league.FreezeCard}).Error; err != nil {
		xlog.Logger().Errorln("update account failed: ", err.Error())
		return 0, 0, 0, 0, err
	}

	// 操作后的房卡
	aftka := league.Card
	aftfka := league.FreezeCard
	// 记录用户财富消耗流水
	if kacost != 0 {
		sql := `insert into league_card_record(league_id,uid,cost) values(?,?,?)`
		if err := tx.Exec(sql, league.LeagueID, uid, kacost).Error; err != nil {
			xlog.Logger().Errorln(err.Error())
			return 0, 0, 0, 0, err
		}
	}
	return int(befka), int(aftka), int(beffka), int(aftfka), nil
}

func (self *ProtocolWorkers) Protocol_Relogin(params ...interface{}) (code int16, v interface{}) {
	_, _, data := self.ProtocolParams(params)
	// 数据解析
	msg := data.(*static.Msg_HG_Relogin)

	//defer GetDBMgr().Redis.Del(fmt.Sprintf("userstatus_doing_exitgame_%d", msg.Uid))

	// 校验桌子是否存在
	table := GetTableMgr().GetTable(msg.TableId)
	if table == nil {
		xlog.Logger().Errorln("can't find table: ", msg.TableId)
		return xerrors.AsyncRespErrorCode, nil
	} else {
		// 强制用户下线
		p := GetPersonMgr().GetPerson(msg.Uid)
		if p != nil {
			// 通知客户端
			p.SendMsg(consts.MsgTypeRelogin, xerrors.SuccessCode, nil)
			// 断开连接
			p.CloseSession(consts.SESSION_CLOED_FORCE)
		}
	}

	return xerrors.AsyncRespErrorCode, nil
}

func (self *ProtocolWorkers) Protocol_TableCreate_Req(params ...interface{}) (code int16, v interface{}) {
	//fmt.Println(fmt.Sprintf("%+v", params[2]))
	_, _, data := self.ProtocolParams(params)
	// 数据解析
	req := data.(*static.HG_HTableCreate_Req)
	xlog.Logger().Debugf("%+v", *req)
	tb := &req.Table
	// 创建牌桌
	table := GetTableMgr().CreateTable(tb)
	if table == nil {
		return xerrors.CreateTableErrorCode, nil
	}

	return xerrors.SuccessCode, nil
}

func (self *ProtocolWorkers) Protocol_TableIn_Req(params ...interface{}) (code int16, v interface{}) {
	_, _, data := self.ProtocolParams(params)
	req := data.(*static.HG_HTableIn_Req)

	table := GetTableMgr().GetTable(req.TId)
	if table == nil {
		return xerrors.TableNotExistErrorCode, nil
	}
	//modify by liujing 20200924 换桌的时候如果频繁换桌可能大厅会传个已经有人的椅子编号过来，这个不认了，无条件改成-1 服务器自己选椅子编号
	req.Seat = -1
	_, _, custerr := table.UserJoinTable(req.Uid, req.Seat, req.Payer)
	if custerr != nil {
		return custerr.Code, nil
	}
	return xerrors.SuccessCode, nil
}

func (self *ProtocolWorkers) Protocol_HTableCreate_Req(params ...interface{}) (code int16, v interface{}) {
	_, _, data := self.ProtocolParams(params)
	// 数据解析
	req := data.(*static.HG_HTableCreate_Req)
	tb := &req.Table

	// 创建牌桌
	table := GetTableMgr().CreateTable(tb)
	if table == nil {
		return xerrors.CreateTableErrorCode, nil
	}
	// 自动加入
	_, _, custerr := GetTableMgr().GetTable(table.Id).UserJoinTable(req.AutoUid, req.AutoSeat, req.Payer)
	if custerr != nil {
		// 接收异常销毁牌桌
		ntable := GetTableMgr().GetTable(table.Id)
		if ntable != nil {
			ntable.WriteTableLog(uint16(req.AutoSeat), "Protocol_HTableCreate_Req 接收异常销毁牌桌")
			xlog.Logger().Warnln("Protocol_HTableCreate_Req 接收异常销毁牌桌", table.Id)
			GetTableMgr().DelTable(ntable)
		}
		return xerrors.CreateTableErrorCode, nil
	}
	return xerrors.SuccessCode, nil
}

func (self *ProtocolWorkers) Protocol_HTableIn(params ...interface{}) (code int16, v interface{}) {
	_, _, data := self.ProtocolParams(params)
	req := data.(*static.HG_HTableIn_Req)

	// 加入牌桌
	table := GetTableMgr().GetTable(req.TId)
	if table == nil {
		return xerrors.TableNotExistError.Code, nil
	}

	_, _, custerr := table.UserJoinTable(req.Uid, -1, req.Payer)
	if custerr != nil {
		return custerr.Code, nil
	}
	return xerrors.SuccessCode, nil
}

func (self *ProtocolWorkers) Protocol_HTableIn_Req(params ...interface{}) (code int16, v interface{}) {
	_, _, data := self.ProtocolParams(params)
	req := data.(*static.HG_HTableIn_Req)

	// 加入牌桌
	table := GetTableMgr().GetTable(req.TId)
	if table == nil {
		return xerrors.TableNotExistError.Code, nil
	}

	_, _, custerr := table.UserJoinTable(req.Uid, req.Seat, req.Payer)
	if custerr != nil {
		return custerr.Code, nil
	}

	return xerrors.SuccessCode, nil
}

func (self *ProtocolWorkers) Protocol_TableDel_Req(params ...interface{}) (code int16, v interface{}) {
	_, _, data := self.ProtocolParams(params)
	req := data.(*static.Msg_HG_TableDel_Req)

	table := GetTableMgr().GetTable(req.TableId)
	if table == nil {
		log.Println("can't find table: ", req.TableId)
	} else {
		// 添加日志
		table.WriteTableLog(static.INVALID_CHAIR, req.Info)

		msg := new(static.Msg_S2C_TableDel)
		msg.Type = forceCloseByHouseManager
		msg.Msg = req.Info
		table.Operator(base2.NewTableMsg(consts.MsgTypeTableDel, "now", 0, msg))
	}
	return xerrors.SuccessCode, nil
}
