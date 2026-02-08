package product_deliver

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	server2 "github.com/open-source/game/chess.git/services/backend/wuhan"
	"time"
)

var supportCostTypes = []interface{}{
	models.CostTypeBuy,
	models.CostTypeGm,
	models.CostTypePartner,
	models.ConstExchange,
	models.CostTypeDiamondExchangeGift,
	models.CostTypeBankruptcyGift,
	models.CostTypeDiamondExchangeRcd,
}

func isSupportCostType(ct int8) bool {
	ict := int(ct)
	for _, t := range supportCostTypes {
		if t.(int) == ict {
			return true
		}
	}
	return false
}

type BankruptcyGiftInfo struct {
	ExpendNum  int              `json:"expend_num"`
	ExpendType int8             `json:"expend_type"`
	Got        []BankruptcyGift `json:"got"`
}

type BankruptcyGift struct {
	WealthType int8 `json:"wt"` // 财富类型
	WealthNum  int  `json:"wn"` // 数量
}

func DeliverProduct(header string, data interface{}) (interface{}, *xerrors.XError) { // 用户付款成功后下发商品
	req, ok := data.(*static.Msg_DeliverProduct)
	if !ok {
		return nil, xerrors.ArgumentError
	}
	// 校验类型
	if !isSupportCostType(req.Type) {
		detail := fmt.Sprintf("当前支持的下发类型有:%d(购买) %d(后台下发) %d(队长发放) %d(商城兑换) %d(礼包兑换) %d(破产礼包) %d(道具-记牌器)",
			supportCostTypes...)
		// syslog.Logger().Errorf("不支持的财富下发类型: type=%d >%s", req.Type, detail)
		return nil, xerrors.NewXError(fmt.Sprintf("不支持的财富下发类型:type==%d >%s", req.Type, detail))
	}

	// php说: 金币钻石只能用3和4传 因为支付服务器那边就是这样定，改不了
	// 所以这里特殊处理下
	// 也就是php认为3是金币 4是钻石，和当前有差异 所以这个转换下
	if req.WealthType == 3 {
		req.WealthType = consts.WealthTypeGold
	} else if req.WealthType == 4 {
		req.WealthType = consts.WealthTypeDiamond
	}

	xlog.Logger().Warnf("Deliver Product Uid%d, Way:%d, %sx%d， extra:%s",
		req.Uid,
		req.Type,
		consts.WealthTypeString(req.WealthType),
		req.Num,
		req.Extra,
	)

	// 如果是破产礼包则解析
	// 礼包有n个财富 这里重写逻辑 以保证事务一致性  else 里面时之前的逻辑
	if req.Type == models.CostTypeBankruptcyGift /*破产礼包*/ || req.Type == models.CostTypeDiamondExchangeGift /*钻石兑换礼包*/ {
		var gift BankruptcyGiftInfo
		err := json.Unmarshal([]byte(req.Extra), &gift)
		if err != nil {
			xlog.Logger().Error(err)
			return nil, xerrors.ArgumentError
		}
		// php说: 金币钻石只能用3和4传 因为支付服务器那边就是这样定，改不了
		// 所以这里特殊处理下
		// 也就是php认为3是金币 4是钻石，和当前有差异 所以这个转换下
		if gift.ExpendType == 3 {
			gift.ExpendType = consts.WealthTypeGold
		} else if gift.ExpendType == 4 {
			gift.ExpendType = consts.WealthTypeDiamond
		}
		// 统计要扣除/增加的财富
		wealthCost := append([]wealthtalk.WealthItem{},
			wealthtalk.WealthItem{
				Wt: req.WealthType,
				Wn: req.Num,
			},
			wealthtalk.WealthItem{
				Wt: gift.ExpendType,
				Wn: -gift.ExpendNum,
			},
		)
		for i := 0; i < len(gift.Got); i++ {
			// php说: 金币钻石只能用3和4传 因为支付服务器那边就是这样定，改不了
			// 所以这里特殊处理下
			// 也就是php认为3是金币 4是钻石，和当前有差异 所以这个转换下
			if gift.Got[i].WealthType == 3 {
				gift.Got[i].WealthType = consts.WealthTypeGold
			} else if gift.Got[i].WealthType == 4 {
				gift.Got[i].WealthType = consts.WealthTypeDiamond
			}
			wealthCost = append(wealthCost, wealthtalk.WealthItem{
				Wt: gift.Got[i].WealthType,
				Wn: gift.Got[i].WealthNum,
			})
		}
		tx := server2.GetDBMgr().GetDBmControl().Begin()
		if req.Type == models.CostTypeBankruptcyGift /*破产礼包*/ {
			nowStr := time.Now().Format("2006-01-02")
			if buy, err := models.IsAllowanceGiftPurchasedToday(tx, req.Uid, nowStr); err != nil {
				xlog.Logger().Error(err)
				return nil, xerrors.DBExecError
			} else {
				if buy {
					return nil, xerrors.NewXError(fmt.Sprintf("破产礼包购买限制:玩家%d在今日(%s)已经购买过破产礼包了", req.Uid, nowStr))
				}
			}
		}
		ret, err := wealthtalk.UpdateWealth(req.Uid, wealthCost, req.Type, tx)
		if err != nil {
			xlog.Logger().Error(err)
			return nil, xerrors.NewXError(err.Error())
		}
		updates := make(map[string]interface{})
		var msg static.MsgGmBankruptcyGift
		l := len(ret)
		msg.CostType = req.Type
		msg.Wts = make([]static.MsgGmGotGiftDetail, l)
		for i := 0; i < l; i++ {
			switch ret[i].Wt {
			case consts.WealthTypeCard:
				updates["Card"] = ret[i].AfterNum
			case consts.WealthTypeGold:
				updates["Gold"] = ret[i].AfterNum
			case consts.WealthTypeDiamond:
				updates["Diamond"] = ret[i].AfterNum
			}
			msg.Wts[i] = static.MsgGmGotGiftDetail{
				Wt:     ret[i].Wt,
				Offset: ret[i].Wn,
			}
		}
		err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrsV2(req.Uid, updates)
		if err != nil {
			tx.Rollback()
			xlog.Logger().Errorln(err)
			return nil, xerrors.NewXError("发放商品信息失败1")
		}
		err = tx.Commit().Error
		if err != nil {
			xlog.Logger().Error(err)
			return nil, xerrors.NewXError("发放商品信息失败2")
		}
		msg.Uid = req.Uid
		res, _ := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeOneStepBuyGift, &msg)
		xlog.Logger().Info(string(res))
		return "true", xerrors.RespOk
	} else if req.Type == models.CostTypeDiamondExchangeRcd {
		return DeliverTools(req)
	} else {
		afterNum := 0
		tx := server2.GetDBMgr().GetDBmControl().Begin()
		var err error
		switch req.WealthType {
		case consts.WealthTypeCard: // 房卡
			_, afterNum, _, _, err = wealthtalk.UpdateCard(req.Uid, req.Num, 0, req.Type, tx)
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return nil, xerrors.NewXError("发放商品信息失败3")
			}

			// 更新redis
			err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "Card", afterNum)
			if err != nil {
				tx.Rollback()
				xlog.Logger().Errorln(err.Error())
				return nil, xerrors.NewXError("发放商品信息失败4")
			}
		case consts.WealthTypeGold: // 房卡
			_, afterNum, err = wealthtalk.UpdateGold(req.Uid, req.Num, req.Type, tx)
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return nil, xerrors.NewXError("发放商品信息失败5")
			}

			// 更新redis
			err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "Gold", afterNum)
			if err != nil {
				tx.Rollback()
				xlog.Logger().Errorln(err.Error())
				return nil, xerrors.NewXError("发放商品信息失败6")
			}
		case consts.WealthTypeDiamond: // 钻石
			_, afterNum, err = wealthtalk.UpdateDiamond(req.Uid, req.Num, req.Type, tx)
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return nil, xerrors.NewXError("发放商品信息失败7")
			}

			// 更新redis
			err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "Diamond", afterNum)
			if err != nil {
				tx.Rollback()
				xlog.Logger().Errorln(err.Error())
				return nil, xerrors.NewXError("发放商品信息失败8")
			}
		case consts.WealthTypeCoupon: // 礼券
			_, afterNum, err = wealthtalk.UpdateCoupon(req.Uid, req.Num, req.Type, tx)
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return nil, xerrors.NewXError("发放商品信息失败9")
			}

			// 更新redis
			err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrs(req.Uid, "GoldBean", afterNum)
			if err != nil {
				tx.Rollback()
				xlog.Logger().Errorln(err.Error())
				return nil, xerrors.NewXError("发放商品信息失败9")
			}
		default:
			return nil, xerrors.NewXError("不支持的财富类型")
		}

		err = tx.Commit().Error
		if err != nil {
			xlog.Logger().Error(err)
			return nil, xerrors.NewXError("发放商品信息失败10")
		}

		if req.Num != 0 {
			msg := &static.Msg_UpdWealth{
				Uid:        req.Uid,
				WealthType: req.WealthType,
				WealthNum:  afterNum,
				CostType:   req.Type,
				Num:        req.Num,
			}
			xlog.Logger().Infof("send hall :%+v", msg)
			// 通知客户端
			res, _ := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeUpdWealth, msg)
			return res, xerrors.RespOk
		}
		return "true", xerrors.RespOk
	}
}
func DeliverTools(req *static.Msg_DeliverProduct) (interface{}, *xerrors.XError) { // 用户付款成功后下发商品
	var gift BankruptcyGiftInfo
	err := json.Unmarshal([]byte(req.Extra), &gift)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, xerrors.ArgumentError
	}
	// php说: 金币钻石只能用3和4传 因为支付服务器那边就是这样定，改不了
	// 所以这里特殊处理下
	// 也就是php认为3是金币 4是钻石，和当前有差异 所以这个转换下
	if gift.ExpendType == 3 {
		gift.ExpendType = consts.WealthTypeGold
	} else if gift.ExpendType == 4 {
		gift.ExpendType = consts.WealthTypeDiamond
	}
	// 统计要扣除/增加的财富
	wealthCost := append([]wealthtalk.WealthItem{},
		//wealthtalk.WealthItem{
		//	Wt: req.WealthType,
		//	Wn: req.Num,
		//},
		wealthtalk.WealthItem{
			Wt: gift.ExpendType,
			Wn: -gift.ExpendNum,
		},
	)
	for i := 0; i < len(gift.Got); i++ {
		// php说: 金币钻石只能用3和4传 因为支付服务器那边就是这样定，改不了
		// 所以这里特殊处理下
		// 也就是php认为3是金币 4是钻石，和当前有差异 所以这个转换下
		if gift.Got[i].WealthType == 3 {
			gift.Got[i].WealthType = consts.WealthTypeGold
		} else if gift.Got[i].WealthType == 4 {
			gift.Got[i].WealthType = consts.WealthTypeDiamond
		}
		wealthCost = append(wealthCost, wealthtalk.WealthItem{
			Wt: gift.Got[i].WealthType,
			Wn: gift.Got[i].WealthNum,
		})
	}
	tx := server2.GetDBMgr().GetDBmControl().Begin()

	ret, err := wealthtalk.UpdateWealth(req.Uid, wealthCost, req.Type, tx)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, xerrors.NewXError(err.Error())
	}
	updates := make(map[string]interface{})
	var msg static.Msg_S_Tool_ToolExchange
	l := len(ret)
	msg.Type = req.Type
	msg.Price = gift.ExpendNum
	msg.Num = req.Num
	for i := 0; i < l; i++ {
		switch ret[i].Wt {
		case consts.WealthTypeCard:
			updates["Card"] = ret[i].AfterNum
		case consts.WealthTypeGold:
			updates["Gold"] = ret[i].AfterNum
		case consts.WealthTypeDiamond:
			updates["Diamond"] = ret[i].AfterNum
		case consts.WealthTypeCoupon:
			updates["GoldBean"] = ret[i].AfterNum
		case consts.WealthTypeCardRcd:
			if ret[i].TimeAt > 0 {
				updates["CardRecorderDeadAt"] = ret[i].TimeAt
				msg.DeadAt = ret[i].TimeAt
			}
		}
	}
	err = server2.GetDBMgr().GetDBrControl().UpdatePersonAttrsV2(req.Uid, updates)
	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorln(err)
		return nil, xerrors.NewXError("发放商品信息失败1")
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Error(err)
		return nil, xerrors.NewXError("发放商品信息失败2")
	}
	msg.Uid = req.Uid
	res, _ := server2.GetServer().CallHall("NewServerMsg", consts.MsgTypeToolExchangeHall, &msg)
	xlog.Logger().Info(string(res))
	return "true", xerrors.RespOk
}
