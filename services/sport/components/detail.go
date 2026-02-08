package components

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"strconv"
	"strings"
)

/**

游戏小结算，分数计算详情解析

**/

const (
	DetailTypeF      = iota //番
	DetailTypeADD           //累加
	DetailTypeOther         //例外
	DetailTypeAll           //所有番
	DetailTypeCost          //减少
	DetailTypeFColor        //彩色
	DetailTypeFirst         //排头标记,不带番数
	DetailTypeFont          //排头标记,带番数
	DetailTypePao           //排头标记,带炮数
	DetailTypeBei           //排头标记,带倍数
	DetailTypeMax
)

const (
	TagHuHard             = iota //硬胡
	TagHuBigKai                  //大胡杠开
	TagHuMinKai                  //小胡杠开
	TagHuMinSelf                 //小胡自摸
	TagHuMin                     //小胡放铳
	TagHuFour                    //4个赖子胡
	TagHuFourNo                  //4个赖子胡不构成牌型
	TagHuwProvider               //放铳
	TagHuJieChong                //接铳
	TagHuFen7Dui                 //风七对
	TagHuJ7Dui                   //将七对
	TagHuQ7Dui                   //清七对
	TagHuFenYiSe                 //风一色
	TagHuQinYiSe                 //清一色
	TagHuJiangYiSe               //将一色
	TagHu7Dui                    //七对
	TagHuQuanQiuRen              //全求人
	TagHuPengPengHu              //碰碰胡
	TagHuGangKai                 //杠上开花
	TagGangKai                   //杠开
	TagHuHaiDi                   //海底捞月
	TagHuQiangGang               //抢杠
	TagXiaPao                    //吓跑
	TagHuSelf                    //自摸
	TagShowGan                   //明杠
	TagXuGan                     //蓄杠
	TagAnGan                     //暗杠
	TagMagic                     //赖子杠
	TagHongzhong                 //红中杠
	TagFaCai                     //发财杠
	TagDianGan                   //回头杠
	TagMagicCount                //癞子个数
	TagMagicScore                //赖分
	TagGangScore                 //杠分
	TagWindCount                 //风牌个数
	TagYinQue                    //硬缺
	TagRuanQue                   //软缺
	TagDuan19                    //断幺
	TagYiTiaoLong                //一条龙
	TagJieMei                    //姐妹铺
	TagJieMei2                   //双姐妹铺
	TagSanKan                    //三坎
	TagSiKan                     //四坎
	TagZiYiSe                    //字一色
	TagHunYiSe                   //混一色
	TagQuan19                    //全幺九
	TagBuXiaJia                  //不下架
	TagSiYunZi                   //四遇子
	TagKanJiang                  //坎将
	TagSiCha                     //四叉
	TagSanDaJiang                //三大将
	TagSanJiFeng                 //三季风
	TagSiJiFeng                  //四季风
	TagPingHu                    //平胡
	TagJiaPeng                   //将碰
	TagFengPeng                  //风碰
	TagFengGang                  //风杠
	TagJiangGang                 //将杠
	TagDuiDui                    //对对胡
	TagPengPeng                  //碰碰胡
	LianZHuang                   //连庄
	TianHu                       //天胡
	BaoTing                      //报听
	TagHuMagic                   //软胡
	TianTing                     //天听 闲家
	DiTing                       //地听 庄家
	YouQianFeng                  //有钱风
	OutMagicCard                 //打赖子  //打赖子吃赖子 Add by zwj for 通山晃晃
	ChiMagicCard                 //吃赖子
	MenQianQing                  //门清
	TagHu7Dui_1                  //豪华七对
	TagHu7Dui_2                  //双豪华七对
	TagHu7Dui_3                  //三豪华七对
	HeiHu                        //黑胡
	TagKaiKou                    //开口
	TagQiangCuo                  //抢错
	TagQiangAnGan                //抢暗杠
	TagZhuangJiaFan              //庄家加番
	TagDuiKaiKou                 //对开口
	TagPizi                      //皮子杠
	TagLiang                     //亮倒
	TagKaiSanKou                 //开三口
	TagKaiSiKou                  //开四口
	TagBuHua                     //补花
	TagChuZeng                   //出增
	TagWeiZeng                   //围增
	TagKengZhuang                //坑庄
	LiangPai                     //亮牌
	ThirOrphans                  //十三幺
	TagYaoYaoHU                  //幺幺胡
	TagSanDaYuan                 //三大元
	TagJia                       //夹
	TagPiao                      //漂分
	TagPiao7DuiLong              //龙七对
	TagRuanZiMo                  //软自摸
	TagYingZiMo                  //硬自摸
	TagRuanLaiYou                //软癞油
	TagYingLaiYou                //硬癞油
	TagRuanQingYiSeLaiYou        //软清一色癞油
	TagYingQingYiSeLaiYou        //硬清一色癞油
	TagRuanQingYiSe              //软清一色
	TagYingQingYiSe              //硬清一色
	TagYouZhongYou               //油中油
	TagK5x                       //卡五星
	TagD3Y                       //大三元
	TagX3Y                       //小三元
	TagSZ1                       //手抓一
	TagM4G                       //明四归一
	TagA4G                       //暗四归一
	TagGuoHu                     //过胡翻番
	TagGangPao                   //杠上炮
	TagShuKan                    //数坎
	TagXiaoChaoTian              //小朝天
	TagDaChaotian                //大朝天
	TagDianXiao                  //点笑
	TagMengXiao                  //闷笑
	TagFangXiao                  //放笑
	TagRuangMo                   //软摸
	TagHeiMo                     //黑摸
	TagHongHu                    //宏胡
	TagHuiXiao                   //回头笑
	TagHard                      //硬
	TagSoft                      //软
	TagSGY                       //四归一
	TagQingYiSe7Dui              //清一色七对
	Tag4H                        //四混
	YtiaoLong
	TagHongZhongPeng     //红中碰
	TagBaiBanPeng        //白板碰
	TagThreeHongZhong    //三个红中
	TagThreeBaiBan       //三个白板
	TagOutFaCai          //打发财
	TagFourBao           //四个宝
	TagMingGangBanBan    //明杠白板
	TagMingGangHongZhong //明杠红中
	TagAnGangBaiBan      //暗杠白板
	TagAnGangHongZhong   //暗杠红中
	TagDanDiao           //单调
	TagHuKa              //胡卡
	TagHuBian            //胡边
	TagTrheeBaiBanHu     //三白板胡牌
	TagTrheeHongZhongHu  //三红中胡牌
	TagHuaDanDiao        //花单钓
	TagHuaDiaoHua        //花钓花
	TagJiHu              //鸡胡
	TagSiGuiHu           //四鬼胡
	TagJianZiHu          //见字胡
	TagHard7Dui          //硬七对
	TagSoft7Dui          //软七对
	TagStartHu           //起手胡
	TagHaiDiPao          //海底炮
	TagHitBird           //中鸟
	TagZhuangJiaTaiDi    //庄家抬底
	TagJieGang           //接杠
	TagFangGang          //放杠
	TagHu4LaiZi          //四癞子
	TagXiaoHuZhuangJia   //小胡庄家
	TagBao               //宝杠
	TagJieZhaoHu         //接招胡(杠开后点炮别人胡，这种胡叫接招胡，通山晃晃)
	TagMaiMa
	TagShangLou
	TagDianPao //点炮
	TagJiePao  //接炮
	TagHuMAX
)

const (
	Colour_Gold = iota // 金色 默认
	Colour_Max
)

const (
	TagStrTypeNomarl = iota // 金色 默认
	TagStrType1
	TagStrMax
)

var TagStrType = [TagStrMax]string{
	" ", "/",
}

var TagColourMsg = [Colour_Max]string{
	"<color=#fcd468>%s</color>",
}

//明杠1番，蓄杠1番，暗杠2番，赖子杠2番，红中杠1番
var TagHuMessage = [TagHuMAX]string{
	"硬胡", "大胡杠开", "小胡杠开", "小胡自摸", "小胡放铳", "4个赖子胡", "4个赖子不构成牌型", "放铳", "接铳", "风七对",
	"将七对", "清七对", "风一色", "清一色", "将一色", "七对", "全求人", "碰碰胡", "杠上开花", "杠开", "海底捞月", "抢杠", "增分", "自摸", "明杠", "蓄杠", "暗杠", "赖子杠", "红中杠", "发财杠", "点杠",
	"赖", "飘分", "杠分", "风", "硬缺", "软缺", "断幺", "一条龙", "姐妹铺", "双姐妹铺", "三坎", "四坎", "字一色", "混一色", "全幺九", "不下架", "四遇子", "坎将", "四叉", "三大将", "三季风", "四季风", "平胡",
	"将碰", "风碰", "风杠", "将杠", "对对胡", "碰碰胡", "连庄", "天胡", "报听", "软胡", "天听", "地听", "值钱风", "打赖子", "吃赖子", "门前清", "豪华七对", "双豪华七对", "三豪华七对", "黑胡", " 开口", "抢错", "抢暗杠",
	"庄家加番", "对开口", "皮子杠", "中发白亮倒", "开三口", "开四口", "补花", "出增", "围增", "坑庄", "亮牌", "十三幺", "幺幺胡", "三大元", "夹", "漂", "龙七对",
	"软自摸", "硬自摸", "软癞油", "硬癞油", "软清一色癞油", "硬清一色癞油", "软清一色", "硬清一色", "油中油", "卡五星", "大三元", "小三元", "手抓一", "明四归", "暗四归", "过胡", "杠上炮", "数坎",
	"小朝天", "大朝天", "点笑", "闷笑", "放笑", "软摸", "黑摸", "宏胡", "回头笑", "硬", "软", "四归一", "清一色七对", "四混", "一条龙", "红中碰", "白板碰", "三个红中", "三个白板", "打发财", "四个宝", "明杠白板", "明杠红中", "暗杠白板",
	"暗杠红中", "单调", "胡卡", "胡边", "胡白板", "胡红中", "花单钓", "花钓花", "鸡胡", "四鬼胡", "见字胡", "硬七对", "软七对", "起手胡", "海底炮", "中鸟", "庄家抬底", "接杠", "放杠", "四癞子", "小胡庄家", "宝杠", "接招胡",
	"买马", "上楼", "点炮", "接炮",
}

var TagAddMessage = [DetailTypeMax]string{
	"X", "+", "+", "X", "", "番", "", "番", "炮", "倍",
}

var TagMessageType = [DetailTypeMax]string{
	"X", "+", "+", "X", "", "番", "", "番", "炮", "倍",
}

type TagHuDetail struct {
	numb     int  //数量
	id       int  //
	costType int  //类型 0：+积分；1 *番
	colour   int  //颜色
	effect   bool //是否生效
	seat     uint16
}

//胡牌计算详情
type TagHuCostDetail struct {
	Detail []*TagHuDetail //
	Type   int            // 分割符号，默认空格
}

func (self *TagHuCostDetail) Init() {
	//self.Detail = make(map[int]*TagHuDetail)
	self.Detail = make([]*TagHuDetail, 0, 2)
	self.Type = TagStrTypeNomarl
}

//公有的计算数据
func (self *TagHuCostDetail) Set(id int, numb int, costtype int, effect bool) {

	if numb == 0 {
		return
	}

	// 当积分类型为加号，但是分数为负数时，去掉加号，显示 -x
	if costtype == DetailTypeADD && numb < 0 {
		costtype = DetailTypeCost
	}

	for _, v := range self.Detail {
		if v.id == id && v.seat == static.INVALID_CHAIR {
			v.effect = effect
			v.numb = numb
			v.costType = costtype

			return
		}
	}

	_detail := new(TagHuDetail)
	_detail.id = id
	_detail.numb = numb
	_detail.effect = effect
	_detail.costType = costtype
	_detail.seat = static.INVALID_CHAIR

	self.Detail = append(self.Detail, _detail)
}

//个人的计算数据
func (self *TagHuCostDetail) Private(seat uint16, id int, numb int, costtype int) {

	if numb == 0 {
		return
	}

	// 当积分类型为加号，但是分数为负数时，去掉加号，显示 -x
	if costtype == DetailTypeADD && numb < 0 {
		costtype = DetailTypeCost
	}

	for _, v := range self.Detail {
		if v.id == id && v.seat == seat {
			v.effect = true
			v.numb = numb
			v.costType = costtype

			return
		}
	}

	_detail := new(TagHuDetail)
	_detail.id = id
	_detail.numb = numb
	_detail.effect = true
	_detail.costType = costtype
	_detail.seat = seat

	self.Detail = append(self.Detail, _detail)
}

func (self *TagHuCostDetail) Colour(seat uint16, colour int, id int, numb int, costtype int) {
	if numb == 0 {
		return
	}

	// 当积分类型为加号，但是分数为负数时，去掉加号，显示 -x
	if costtype == DetailTypeADD && numb < 0 {
		costtype = DetailTypeCost
	}

	for _, v := range self.Detail {
		if v.id == id && v.seat == seat {
			v.effect = true
			v.numb += numb
			v.costType = costtype
			v.colour = colour
			return
		}
	}

	_detail := new(TagHuDetail)
	_detail.id = id
	_detail.numb = numb
	_detail.effect = true
	_detail.costType = costtype
	_detail.seat = seat
	_detail.colour = colour
	self.Detail = append(self.Detail, _detail)
}

func (self *TagHuCostDetail) AddPrivate(_info *TagHuDetail) {
	if _info.numb == 0 {
		return
	}

	for _, v := range self.Detail {
		if v.id == _info.id && v.seat == _info.seat {
			v.effect = _info.effect
			v.numb = _info.numb
			v.costType = _info.costType

			return
		}
	}

	_detail := new(TagHuDetail)
	_detail.id = _info.id
	_detail.numb = _info.numb
	_detail.effect = _info.effect
	_detail.costType = _info.costType
	_detail.seat = _info.seat

	self.Detail = append(self.Detail, _detail)
}

func (self *TagHuCostDetail) Delete(seat uint16, id int) {
	for _, v := range self.Detail {
		if v.id == id && v.seat == seat {
			v.effect = false
			return
		}
	}
}

func (self *TagHuCostDetail) Add(_detail *TagHuCostDetail) {

	for _, v := range _detail.Detail {
		if v.seat == static.INVALID_CHAIR {
			self.Set(v.id, v.numb, v.costType, v.effect)
		} else {
			self.AddPrivate(v)
		}
	}
}

//计算详情解析
func (self *TagHuCostDetail) GetSeatString(seat uint16) string {

	//if len(self.Detail) <= 0 {//没有数据
	//	return ""
	//}
	str := self.getStringFont(seat)
	str += self.getStringFirst(seat)
	str += self.getString(seat)
	str += self.getCostString(seat)
	str += self.getStringFColor(seat)
	str += self.getStringF(seat)
	str += self.getStringAll(seat)

	str += self.getStringOther(seat)
	str += self.getStringPao(seat)
	str += self.getStringBei(seat)

	return str
}

//计算详情解析
func (self *TagHuCostDetail) GetSeatString1(seat uint16) string {

	//if len(self.Detail) <= 0 {//没有数据
	//	return ""
	//}
	str := self.getStringFont(seat)
	str += self.getStringFirst(seat)
	str += self.getString(seat)
	str += self.getCostString(seat)
	str += self.getStringFColor1(seat)
	str += self.getStringF(seat)
	str += self.getStringAll(seat)

	str += self.getStringOther(seat)
	str += self.getStringPao(seat)

	return str
}

// 使用 '/' 分割
func (self *TagHuCostDetail) GetSeatString2(seat uint16) string {
	self.Type = TagStrType1
	str := self.GetSeatString(seat)
	self.Type = TagStrTypeNomarl //还原nomarl
	str = strings.TrimRight(str, "/")
	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getString(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeADD && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
		}
	}

	return str
}

//获取累-的计算
func (self *TagHuCostDetail) getCostString(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeCost && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
		}
	}

	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getStringFirst(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeFirst && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + TagStrType[self.Type]
		}
	}

	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getStringFont(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeFont && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + "<color=#FFFF00>" + strconv.Itoa(v.numb) + TagAddMessage[v.costType] + " </c>" + TagStrType[self.Type]
		}
	}

	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getStringPao(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypePao && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + "<color=#FFFF00>" + strconv.Itoa(v.numb) + TagAddMessage[v.costType] + " </c>" + TagStrType[self.Type]
		}
	}

	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getStringBei(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeBei && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + "<color=#FFFF00>" + strconv.Itoa(v.numb) + TagAddMessage[v.costType] + " </c>" + TagStrType[self.Type]
		}
	}

	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getStringFColor(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeFColor && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + "<color=#FFFF00>" + strconv.Itoa(v.numb) + TagAddMessage[v.costType] + " </c> " + TagStrType[self.Type]
		}
	}

	return str
}

//获取累加的计算
func (self *TagHuCostDetail) getStringFColor1(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeFColor && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + "<color=#FFFF00>" + strconv.Itoa(v.numb) + TagAddMessage[v.costType] + " </color> " + TagStrType[self.Type]
		}
	}

	return str
}

// 番数例外的计算
func (self *TagHuCostDetail) getStringOther(seat uint16) string {
	var str string

	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeOther && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
		}
	}

	return str
}

//番数的计算
func (self *TagHuCostDetail) getStringF(seat uint16) string {
	var str string

	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeF && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
		}
	}

	return str
}

//所有数据加番的计算
func (self *TagHuCostDetail) getStringAll(seat uint16) string {
	var str string

	for _, v := range self.Detail {
		if v.effect && v.costType == DetailTypeAll && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += TagHuMessage[v.id] + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
		}
	}

	return str
}

//得到多彩的字符床拼接
func (self *TagHuCostDetail) GetColourfulString(seat uint16) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			str += fmt.Sprintf(TagColourMsg[v.colour], TagHuMessage[v.id]) + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
		}
	}
	return str
}

//得到多彩的字符床拼接
func (self *TagHuCostDetail) GetColourfulString_TMMJ(seat uint16, Radix int) string {
	var str string
	for _, v := range self.Detail {
		if v.effect && (v.seat == static.INVALID_CHAIR || v.seat == seat) {
			if v.id == TagGangScore && Radix > 0 {
				score := float64(v.numb) / float64(Radix)
				str += fmt.Sprintf(TagColourMsg[v.colour], TagHuMessage[v.id]) + TagAddMessage[v.costType] + strconv.FormatFloat(score, 'f', -1, 64) + TagStrType[self.Type]
			} else {
				str += fmt.Sprintf(TagColourMsg[v.colour], TagHuMessage[v.id]) + TagAddMessage[v.costType] + strconv.Itoa(v.numb) + TagStrType[self.Type]
			}
		}
	}
	return str
}
