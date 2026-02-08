package center

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////////
// 兑换商店管理器

type StoreMgr struct {
	lock *sync.RWMutex
}

var storeMgrSingleton *StoreMgr = nil

func GetShopMgr() *StoreMgr {
	if storeMgrSingleton == nil {
		storeMgrSingleton = new(StoreMgr)
		storeMgrSingleton.lock = new(sync.RWMutex)
	}
	return storeMgrSingleton
}

// 任务管理器初始化
func (self *StoreMgr) Init() bool {
	self.InitConfig()
	return true
}

// 加载任务列表
func (self *StoreMgr) InitConfig() bool {
	return true
}

// 兑换商品
func (self *StoreMgr) exchange(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	self.lock.Lock()
	defer self.lock.Unlock()
	req := data.(*static.Msg_Shop_Exchange)

	// 获取最新的商品列表
	var shopItems []*models.ConfigShop
	if err := GetDBMgr().db_M.Where("deleted = 0 ").Order("config_shop.order asc").Find(&shopItems).Error; err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	// 寻找兑换商品
	var tgtProduct *models.ConfigShop
	for _, item := range shopItems {
		if item.Id == req.Id {
			tgtProduct = item
			break
		}
	}

	if tgtProduct != nil {
		tgtProduct.Num += tgtProduct.Gift
	}

	// 返回信息
	var ack = &static.Msg_S2C_Shop_Exchange{
		Code:     0,
		Id:       req.Id,
		Card:     p.Card,
		GoldBean: p.GoldBean,
		Gold:     p.Gold,
	}

	// 最新商品信息
	for _, v := range shopItems {
		item := new(static.Shop_Product)
		item.Id = v.Id
		item.Num = v.Num
		item.Type = v.Type
		item.Name = v.Name
		item.Left = v.Left
		item.Price = v.Price
		item.Order = v.Order
		item.Image = v.Image
		item.Gift = v.Gift
		ack.Product = append(ack.Product, item)
	}

	if tgtProduct == nil {
		ack.Code = 3
		return xerrors.ShopProductOffShelfError.Code, xerrors.ShopProductOffShelfError.Msg
	}

	if tgtProduct.Price > p.GoldBean {
		ack.Code = 1
		return xerrors.CouponNotEnoughError.Code, xerrors.CouponNotEnoughError.Msg
	}

	if tgtProduct.Left <= 0 {
		ack.Code = 3
		return xerrors.ShopProductSoldOutError.Code, xerrors.ShopProductSoldOutError.Msg
	}

	// 实物和话费需要绑定手机
	var shopPhone models.ShopPhone
	if tgtProduct.Type == consts.SHOP_PRODUCT_GOODS || tgtProduct.Type == consts.SHOP_PRODUCT_BILL {
		if err := GetDBMgr().GetDBmControl().Where("uid = ?", req.Uid).Find(&shopPhone).Error; err != nil {
			ack.Code = 2
			return xerrors.ShopPhoneNotBindError.Code, xerrors.ShopPhoneNotBindError.Msg
		}
	}

	// 更新领取记录
	record := new(models.ShopRecord)
	record.Num = tgtProduct.Num
	record.Price = tgtProduct.Price
	record.Name = tgtProduct.Name
	record.Status = consts.SHOP_PRODUCT_WAITING
	record.CreatedAt = time.Now()
	record.ProductId = tgtProduct.Id
	record.UId = req.Uid
	record.Type = tgtProduct.Type
	record.Passtime = time.Now()
	if tgtProduct.Type == consts.SHOP_PRODUCT_GOODS || tgtProduct.Type == consts.SHOP_PRODUCT_BILL {
		record.Tel = shopPhone.Tel
	} else {
		record.Tel = ""
	}

	// 发放奖励
	tx := GetDBMgr().GetDBmControl().Begin()
	if tgtProduct.Type == consts.SHOP_PRODUCT_CARD {
		_, afterCard, _, _, err := updcard(p.Uid, tgtProduct.Num, 0, models.ConstExchange, tx)
		if err != nil {
			xlog.Logger().Error(err)
			tx.Rollback()
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}

		p.Card = afterCard
		record.Status = consts.SHOP_PRODUCT_SUC
	} else if tgtProduct.Type == consts.SHOP_PRODUCT_GOLD {
		_, afterGold, err := updgold(p.Uid, tgtProduct.Num, models.ConstExchange, tx)
		if err != nil {
			xlog.Logger().Error(err)
			tx.Rollback()
			return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
		}
		p.Gold = afterGold

		record.Status = consts.SHOP_PRODUCT_SUC
	}

	_, afterCoupon, err := updCoupon(p.Uid, -tgtProduct.Price, models.ConstExchange, tx)
	p.GoldBean = afterCoupon
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	if err := tx.Model(models.ConfigShop{}).Where("deleted = 0 and id = ? ", req.Id).Update("left", tgtProduct.Left-1).Error; err != nil {
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	// 更新redis
	if err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "Card", p.Card, "Gold", p.Gold, "GoldBean", p.GoldBean); err != nil {
		tx.Rollback()
		xlog.Logger().Error(err)
		return xerrors.DBExecError.Code, xerrors.DBExecError.Msg
	}

	tx.Commit()

	// 更新返回信息
	ack.Code = 0
	ack.Card = p.Card
	ack.GoldBean = p.GoldBean
	ack.Gold = p.Gold
	// 更新购买商品的数量
	for i := 0; i < len(ack.Product); i++ {
		if ack.Product[i].Id == tgtProduct.Id {
			ack.Product[i].Left -= 1
		}
	}

	return xerrors.SuccessCode, ack
}

// 获取兑换商品列表
func (self *StoreMgr) GetShopProduct(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	//req := data.(*public.Msg_Shop_Product)

	// 获取兑换商品
	var shopItems []*models.ConfigShop
	if err := GetDBMgr().db_M.Where("deleted =0 ").Order("config_shop.order asc").Find(&shopItems).Error; err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	var msg static.Msg_S2C_Shop_Product

	for _, v := range shopItems {
		item := new(static.Shop_Product)
		item.Id = v.Id
		item.Num = v.Num
		item.Type = v.Type
		item.Name = v.Name
		item.Left = v.Left
		item.Price = v.Price
		item.Order = v.Order
		item.Image = v.Image
		item.Gift = v.Gift
		msg.Product = append(msg.Product, item)
	}

	return xerrors.SuccessCode, msg
}

// 获取兑换记录列表
func (self *StoreMgr) GetShopRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Shop_Record)
	num := 7
	// 获取兑换商品
	var record []*models.ShopRecord

	//db:= GetDBMgr().GetDBmControl().Where("UId =? ", req.Uid).Order("created_at asc").Limit(num).Limit(num).Offset(req.Page)
	if err := GetDBMgr().GetDBmControl().Where("UId =? ", req.Uid).Order("created_at desc").Find(&record).Error; err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}

	var msg static.Msg_S2C_Shop_Record
	msg.Page = req.Page
	_len := len(record)
	msg.Total = _len / num
	if _len%num != 0 {
		msg.Total = msg.Total + 1
	}

	_minIndex := num * req.Page
	_maxIndex := num * (req.Page + 1)
	for i := 0; i < len(record); i++ { //_, v := range record {
		if i < _minIndex {
			continue
		}
		if i >= _maxIndex {
			break
		}
		v := record[i]
		_date := new(static.Shop_Record)
		_date.Id = v.Id
		_date.Num = v.Num
		_date.Name = v.Name
		_date.Price = v.Price
		_date.Status = v.Status
		_date.Time = v.CreatedAt.Format(static.TIMEFORMAT)
		msg.Record = append(msg.Record, _date)
	}

	return xerrors.SuccessCode, msg
}

// 获取礼卷奖励列表
func (self *StoreMgr) GetGoldRecord(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Shop_GoldRecord)
	num := 7
	// 获取礼卷奖励
	var record []*models.GameMatchCouponRecord
	//GetDBMgr().GetDBmControl().Where("uid =? ", req.Uid).Order("created_at asc").Limit(num).Offset(req.Page)
	if err := GetDBMgr().GetDBmControl().Where("uid =? ", req.Uid).Order("created_at desc").Find(&record).Error; err != nil {
		return xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg
	}
	_len := len(record)

	var msg static.Msg_S2C_Shop_GoldRecord
	msg.Page = req.Page
	msg.Total = _len / num
	if _len%num != 0 {
		msg.Total = msg.Total + 1
	}

	_minIndex := num * req.Page
	_maxIndex := num * (req.Page + 1)
	for i := 0; i < len(record); i++ { //_, v := range record {
		if i < _minIndex {
			continue
		}
		if i >= _maxIndex {
			break
		}
		v := record[i]
		_date := new(static.Gold_Record)
		_date.Id = v.Id
		_date.Num = v.Awards
		_date.Type = v.AwardType
		_date.MatchKey = v.MatchKey
		_date.Time = v.CreatedAt.Format(static.TIMEFORMAT)
		msg.Record = append(msg.Record, _date)
	}

	return xerrors.SuccessCode, msg
}

// 获取绑定兑换手机
func (self *StoreMgr) GetBindPhone(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Shop_GetPhoneBind)
	// 获取兑换手机绑定信息
	var record models.ShopPhone

	if err := GetDBMgr().GetDBmControl().Where("uid=?", req.Uid).Find(&record).Error; err != nil {
		return xerrors.SuccessCode, static.Msg_S2C_Shop_GetPhoneBind{Code: 1, Tel: ""}
	} else {
		return xerrors.SuccessCode, static.Msg_S2C_Shop_GetPhoneBind{Code: 0, Tel: record.Tel}
	}

	return xerrors.SuccessCode, nil
}

// 绑定兑换手机
func (self *StoreMgr) PhoneBind(s *Session, p *static.Person, data interface{}) (code int16, v interface{}) {
	req := data.(*static.Msg_Shop_PhoneBind)

	// 获取兑换手机绑定信息
	var record models.ShopPhone
	reqCode := 0

	if err := GetDBMgr().GetDBmControl().Where("uid=?", req.Uid).Find(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			record.Tel = req.Tel
			record.UId = req.Uid
			GetDBMgr().GetDBmControl().Create(&record)
		} else {
			reqCode = 1
		}
	} else {
		record.UId = req.Uid
		if err := GetDBMgr().GetDBmControl().Model(&record).Update("tel", req.Tel).Error; err != nil {
			reqCode = 1
		}
	}

	return xerrors.SuccessCode, static.Msg_S2C_Shop_PhoneBind{Code: reqCode, Msg: ""}
}
