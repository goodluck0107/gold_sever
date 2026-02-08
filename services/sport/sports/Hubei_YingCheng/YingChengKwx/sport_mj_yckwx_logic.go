package YingChengKwx

import (
	"github.com/open-source/game/chess.git/pkg/static"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	"math/rand"
)

//////////////////////////////////////////////////////////////////////////
const (
	//游戏 I D
	//GAME_NAME  = "卡五星"                               //游戏名字
	GAME_JMSK_GENRE = (static.GAME_GENRE_GOLD | static.GAME_GENRE_MATCH) //游戏类型
)

//牌型设计
var k5x_strCardsMessage = [static.MAX_INDEX]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09",
	"0x11", "0x12", "0x13", "0x14", "0x15", "0x16", "0x17", "0x18", "0x19",
	"0x21", "0x22", "0x23", "0x24", "0x25", "0x26", "0x27", "0x28", "0x29",
	"0x31", "0x32", "0x33", "0x34", "0x35", "0x36", "0x37",
}
var k5x_strCardsMessage1 = [static.MAX_INDEX]string{
	"1万", "2万", "3万", "4万", "5万", "6万", "7万", "8万", "9万",
	"1条", "2条", "3条", "4条", "5条", "6条", "7条", "8条", "9条",
	"1同", "2同", "3同", "4同", "5同", "6同", "7同", "8同", "9同",
	"东风", "南风", "西风", "北风", "红中", "发财", "白板",
}

//扑克数据
var k5x_cbCardDataArray = [static.MAX_REPERTORY]byte{
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子

	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子

	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子

	0x31, 0x31, 0x31, 0x31, //东风
	0x32, 0x32, 0x32, 0x32, //南风
	0x33, 0x33, 0x33, 0x33, //西风
	0x34, 0x34, 0x34, 0x34, //北风
	0x35, 0x35, 0x35, 0x35, //中
	0x36, 0x36, 0x36, 0x36, //发
	0x37, 0x37, 0x37, 0x37, //白
}

//胡牌类型掩码数组
var k5x_KindMask = [13]int{
	static.CHK_7_DUI,
	static.CHK_7_DUI_1,
	static.CHK_7_DUI_2,
	static.CHK_7_DUI_3,
	static.CHK_QING_YI_SE,
	static.CHK_XIAO_SAN_YUAN,
	static.CHK_DA_SAN_YUAN,
	static.CHK_KA_5_XING,
	static.CHK_SHOU_ZHUA_YI,
	static.CHK_SI_GUI_YI_MING,
	static.CHK_SI_GUI_YI_AN,
	static.CHK_PENG_PENG,
	static.CHK_QIANG_GANG,
}

type SportLogicYCKWX struct {
	logic2.BaseLogic
}

//混乱扑克
func (sl *SportLogicYCKWX) RandCardData() (byte, []byte) {
	//混乱准备
	cbCardData := [static.MAX_REPERTORY]byte{}
	var cbCardDataTemp []byte
	var cbCardWanTiaoTong []byte
	var cbCardWind []byte
	var cbMaxCount byte = static.MAX_REPERTORY
	//去万
	cbMaxCount -= 4 * 9 //去万字牌时，少4*9张牌
	cbCardWanTiaoTong = make([]byte,static.MAX_REPERTORY-7*4-4*9)
	copy(cbCardWanTiaoTong, k5x_cbCardDataArray[4*9:static.MAX_REPERTORY-7*4])

	//去风
	cbMaxCount -= 4 * 4	//去风牌,少东南西北
	cbCardWind = make([]byte,static.MAX_REPERTORY-(12*9+4*4))
	copy(cbCardWind, k5x_cbCardDataArray[12*9+4*4:static.MAX_REPERTORY])

	cbCardDataTemp = append(cbCardDataTemp,cbCardWanTiaoTong...)
	cbCardDataTemp = append(cbCardDataTemp,cbCardWind...)

	//混乱扑克
	cbRandCount, cbPosition := 0, 0
	randTmp := 0
	nAccert := 0
	for {
		nAccert++
		if nAccert > 200 {
			break
		}
		randTmp = int(cbMaxCount) - cbRandCount - 1
		if randTmp > 0 {
			cbPosition = rand.Intn(randTmp)
		} else {
			cbPosition = 0
		}
		cbCardData[cbRandCount] = cbCardDataTemp[cbPosition]
		cbRandCount++
		cbCardDataTemp[cbPosition] = cbCardDataTemp[int(cbMaxCount)-cbRandCount]
		if cbRandCount >= int(cbMaxCount) {
			break
		}
	}

	//syslog.Logger().Errorln(cbCardData)

	return cbMaxCount, cbCardData[:]
}

//有效判断
func (sl *SportLogicYCKWX) IsValidCard(cbCardData byte) bool {
	cbValue := (cbCardData & static.MASK_VALUE)
	cbColor := (cbCardData & static.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	bIsValid := (((cbValue >= 1) && (cbValue <= 9) && (cbColor <= 2 && cbColor >=1)) || ((cbValue >= 5) && (cbValue <= 7) && (cbColor == 3)))
	if !bIsValid {
		return false //非法牌
	}

	return true
}

//吃牌判断
func (sl *SportLogicYCKWX) EstimateEatCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	if !sl.IsValidCard(cbCurrentCard) {
		//TODO
		return static.WIK_NULL
	}

	//过滤判断
	if cbCurrentCard >= 0x30 {
		return static.WIK_NULL
	}

	//变量定义
	cbExcursion := [3]byte{0, 1, 2}
	cbItemKind := [3]byte{static.WIK_LEFT, static.WIK_CENTER, static.WIK_RIGHT}

	//吃牌判断
	var cbEatKind, cbFirstIndex byte = 0, 0
	var cbCurrentIndex byte = sl.SwitchToCardIndex(cbCurrentCard)
	cbMgicCardIndex := sl.SwitchToCardIndex(sl.MagicCard)
	for i := 0; i < len(cbItemKind); i++ {
		var cbValueIndex byte = cbCurrentIndex % 9
		if (cbValueIndex >= cbExcursion[i]) && ((cbValueIndex - cbExcursion[i]) <= 6) {
			//吃牌判断
			cbFirstIndex = cbCurrentIndex - cbExcursion[i]
			if (cbCurrentIndex != cbFirstIndex) && (cbCardIndex[cbFirstIndex] == 0) {
				continue
			}
			if (cbCurrentIndex != (cbFirstIndex + 1)) && (cbCardIndex[cbFirstIndex+1] == 0) {
				continue
			}
			if (cbCurrentIndex != (cbFirstIndex + 2)) && (cbCardIndex[cbFirstIndex+2] == 0) {
				continue
			}

			if i == 0 && (cbMgicCardIndex == cbCurrentIndex + 1 || cbMgicCardIndex == cbCurrentIndex + 2) {
				continue
			} else if i== 1 && (cbMgicCardIndex == cbCurrentIndex - 1 || cbMgicCardIndex == cbCurrentIndex + 1) {
				continue
			} else if i == 2 && (cbMgicCardIndex == cbCurrentIndex - 1 || cbMgicCardIndex == cbCurrentIndex - 2) {
				continue
			}

			//设置类型
			cbEatKind |= cbItemKind[i]
		}
	}

	return cbEatKind
}

//杠牌分析
func (sl *SportLogicYCKWX) AnalyseGangCard(_userItem *components2.Player, cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, GangCardResult *static.TagGangCardResult, liang [4]meta2.ShowCard) int {
	//设置变量
	cbActionMask := static.WIK_NULL
	//亮牌玩家可以杠的牌有
	//1.已碰出的牌
	//2.主动暗铺、默认暗铺
	bCanGang := false
	if liang[_userItem.Seat].BIsShowCard{
		for _,card := range liang[_userItem.Seat].CbAnPuCard{
			if card != 0 {
				bCanGang = true
			}
		}
		//已碰出的牌
		for i := byte(0); i < cbWeaveCount; i++ {
			if WeaveItem[i].WeaveKind == static.WIK_PENG {
				if cbCardIndex[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
					bCanGang = true
				}
			}
		}
	}

	if liang[_userItem.Seat].BIsShowCard && !bCanGang {
		return static.WIK_NULL
	}

	//手上杠牌
	for i := byte(0); i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] == 4 {

			//玩家如果已亮牌,手牌如果还有4张,并且这张牌不是暗铺的牌,那这4张是不给杠的
			if liang[_userItem.Seat].BIsShowCard {
				//已经亮出的牌不能杠
				if sl.IsShowCard(_userItem.Seat, sl.SwitchToCardData(i), liang){
					continue
				}
				//非暗铺的牌不能杠
				if !sl.IsUserAnPuCard(_userItem.Seat, sl.SwitchToCardData(i), liang){
					continue
				}
			}

			cbActionMask |= static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = sl.SwitchToCardData(i)
			GangCardResult.CardCount++
		}
	}

	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_PENG {
			if cbCardIndex[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				//如果当前牌弃杠,不给杠
				if sl.IsGiveUpGang(_userItem, WeaveItem[i].CenterCard) {
					continue
				}
				//如果当前牌是已经亮出的牌,不给杠
				if sl.IsShowCard(_userItem.Seat, WeaveItem[i].CenterCard, liang){
					continue
				}
				//如果当前牌是其它玩家所听的牌,这时候就会点炮抢杠胡,不给杠
				if sl.IsOthersTingCard(_userItem.Seat, WeaveItem[i].CenterCard, liang){
					sl.AppendGiveUpGang(_userItem, WeaveItem[i].CenterCard)
					continue
				}

				cbActionMask |= static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = WeaveItem[i].CenterCard
				GangCardResult.CardCount++
			}
		}
	}

	return cbActionMask
}

//动作等级
func (sl *SportLogicYCKWX) GetUserActionRank(cbUserAction int) byte {
	//抢暗杠等级
	if cbUserAction&static.WIK_QIANG != 0 {
		return 5
	}

	//胡牌等级
	if cbUserAction&static.WIK_CHI_HU != 0 {
		return 4
	}

	//杠牌等级
	if cbUserAction&(static.WIK_GANG) != 0 {
		return 3
	}

	//碰牌等级
	if cbUserAction&static.WIK_PENG != 0 {
		return 2
	}

	//上牌等级
	if cbUserAction&(static.WIK_RIGHT|static.WIK_CENTER|static.WIK_LEFT) != 0 {
		return 1
	}

	return 0
}

//基本胡牌分析
func (sl *SportLogicYCKWX) AnalyseHuKind(wChiHuRight uint64, wChiHuKind uint64, cbTempCard []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte,checkk5 bool) uint64  {

	//清一色：手中的牌和组合牌都是一种花色，乱将
	if sl.IsQingYiSe(cbTempCard, WeaveItem, cbWeaveCount) == true {
		wChiHuRight |= static.CHR_QING_YI_SE
	}

	// 大三元
	if sl.IsDaSanYuan(cbTempCard, WeaveItem, cbWeaveCount){
		wChiHuRight |= static.CHR_DA_SAN_YUAN
	}
	// 小三元
	if sl.IsXiaoSanYuan(cbTempCard, WeaveItem, cbWeaveCount){
		wChiHuRight |= static.CHR_XIAO_SAN_YUAN
	}

	//7对
	if cbWeaveCount == 0{
		if  is7dui,iHaohuaNum := sl.IsQiDui(cbTempCard); is7dui{
			switch iHaohuaNum {
			case 1:
				wChiHuKind |= static.CHK_7_DUI_1
			case 2:
				wChiHuKind |= static.CHK_7_DUI_2
			case 3:
				wChiHuKind |= static.CHK_7_DUI_3
			default:
				wChiHuKind |= static.CHK_7_DUI
			}
		}
	}

	_, AnalyseItemArray := sl.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//牌型分析
		for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
			//变量定义
			bLianCard, bPengCard := false, false
			var pAnalyseItem *static.TagAnalyseItem = &AnalyseItemArray[iCount]

			//得到可胡牌类型中牌眼的值
			//cbEyeValue := pAnalyseItem.CardEye & static.MASK_VALUE
			//得到胡牌类型中对牌
			//cbEyeCard := pAnalyseItem.CardEye

			////判断牌眼是不是将将必须是2 5 8，如果是2的时候对牌不能是风牌（南风） 如果是5的时候不能是风牌（红中）
			//bSymbolEye := ((cbEyeValue == 2 && cbEyeCard != 0x32) || (cbEyeValue == 5 && cbEyeCard != 0x35) || (cbEyeValue == 8))

			//对一个胡牌分析子项进行分析，如果有一个吃牌类型，就记bLianCard为真，如果有一个碰牌类型就记bPengCard为真
			for j := 0; j < len(pAnalyseItem.WeaveKind); j++ {
				cbWeaveKind := pAnalyseItem.WeaveKind[j]
				if cbWeaveKind&(static.WIK_GANG|static.WIK_FILL|static.WIK_PENG) != 0 {
					bPengCard = true
				}

				if (cbWeaveKind & (static.WIK_LEFT | static.WIK_CENTER | static.WIK_RIGHT)) != 0 {
					bLianCard = true
				}
			}

			//1.软碰碰牌分析子项中，没有吃牌类型，必须都是碰牌类型(碰碰胡需要2 5 8做将，按小胡计算)
			if (bLianCard == false) && (bPengCard == true) {
				wChiHuKind |= static.CHK_PENG_PENG
			}

			// 有清一色的牌权并且牌型可以胡
			if (static.CHR_QING_YI_SE & wChiHuRight) != 0 {
				wChiHuKind |= static.CHK_QING_YI_SE
			}

			// 需要成牌型且要258做将的牌型
			//if bSymbolEye {
				wChiHuKind |= static.CHK_PING_HU_NOMAGIC
			//}

			//大三元
			if static.CHR_DA_SAN_YUAN & wChiHuRight != 0 {
				wChiHuKind |= static.CHK_DA_SAN_YUAN
			}
			//小三元
			if static.CHR_XIAO_SAN_YUAN & wChiHuRight != 0 {
				wChiHuKind |= static.CHK_XIAO_SAN_YUAN
			}

			//卡五星
			if checkk5&& sl.IsKa5Xing(pAnalyseItem, cbWeaveCount, cbCurrentCard){
				wChiHuKind |= static.CHK_KA_5_XING
			}

			//四归一
			if bIs4Gui1, isAn, nMingCount := sl.IsSiGuiYi_XiaoGan(cbTempCard, WeaveItem, cbWeaveCount, cbCurrentCard); bIs4Gui1{
				if isAn{
					//20200105 沈强 非清一色的暗四归扔这里
					if !checkk5{
						if wChiHuKind & static.CHK_QING_YI_SE!=0{
							wChiHuKind |= static.CHK_SI_GUI_YI_AN	//暗四归一
						}
					}else{
						wChiHuKind |= static.CHK_SI_GUI_YI_AN	//暗四归一
					}
				}else{
					//wChiHuKind |= static.CHK_SI_GUI_YI_MING	//明四归一
					if nMingCount == 0{
						wChiHuKind |= static.CHK_SI_GUI_YI_MING	//明四归一
					}
				}
				if nMingCount > 0 {
					if (wChiHuKind & static.CHK_7_DUI == 0 &&
						wChiHuKind & static.CHK_7_DUI_1 == 0 &&
						wChiHuKind & static.CHK_7_DUI_2 == 0 &&
						wChiHuKind & static.CHK_7_DUI_3 == 0)&&
						wChiHuKind & static.CHK_QING_YI_SE != 0 {
						switch nMingCount {
						case 1:
							wChiHuKind |= CHK_M4G_1
						case 2:
							wChiHuKind |= CHK_M4G_2
						case 3:
							wChiHuKind |= CHK_M4G_3
						}
					}
				}
			}

			//手抓一
			if sl.IsShouZhuaYi(cbTempCard, WeaveItem, cbWeaveCount, cbCurrentCard){
				wChiHuKind |= static.CHK_SHOU_ZHUA_YI
			}

			// 特殊情况：手抓一和碰碰胡不能同时存在
			if wChiHuKind & static.CHK_SHOU_ZHUA_YI != 0 && wChiHuKind & static.CHK_PENG_PENG != 0 {
				wChiHuKind ^= static.CHK_PENG_PENG
			}
		}
	}
	//特殊情况：当成七对牌型的时候,清一色可以不成牌型
	if (wChiHuKind & static.CHK_7_DUI != 0 ||
		wChiHuKind & static.CHK_7_DUI_1 != 0 ||
		wChiHuKind & static.CHK_7_DUI_2 != 0 ||
		wChiHuKind & static.CHK_7_DUI_3 != 0)&&
		wChiHuRight & static.CHR_QING_YI_SE != 0 {

		wChiHuKind |= static.CHK_QING_YI_SE
	}

	//特殊情况：七对牌型不能成暗四归
	if (wChiHuKind & static.CHK_7_DUI != 0 ||
		wChiHuKind & static.CHK_7_DUI_1 != 0 ||
		wChiHuKind & static.CHK_7_DUI_2 != 0 ||
		wChiHuKind & static.CHK_7_DUI_3 != 0)&&
		wChiHuKind & static.CHK_SI_GUI_YI_AN != 0 {

		wChiHuKind ^= static.CHK_SI_GUI_YI_AN
	}

	//特殊情况：七对牌型不能成卡五星
	if (wChiHuKind & static.CHK_7_DUI != 0 ||
		wChiHuKind & static.CHK_7_DUI_1 != 0 ||
		wChiHuKind & static.CHK_7_DUI_2 != 0 ||
		wChiHuKind & static.CHK_7_DUI_3 != 0)&&
		wChiHuKind & static.CHK_KA_5_XING != 0 {

		wChiHuKind ^= static.CHK_KA_5_XING
	}
//20210104 沈强 要根据情况修改
/*
卡5和碰碰，是有可能出现选碰的
 */
	if wChiHuKind & static.CHK_KA_5_XING != 0&&wChiHuKind & static.CHK_PENG_PENG != 0{
		if sl.Rule.K5x4{
			//如果选了ka5X4.就直接是卡5
			wChiHuKind ^= static.CHK_PENG_PENG
		}else{
			if sl.Rule.PPx4{
				wChiHuKind ^= static.CHK_KA_5_XING
			}else{
				wChiHuKind ^= static.CHK_PENG_PENG
			}
		}
	}

	// 杠上开花判断
	if (static.CHR_GANG_SHANG_KAI_HUA & wChiHuRight) != 0 {
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_GANG_SHANG_KAI_HUA
		}
	}
	// 杠上炮判断
	if (static.CHR_GANG_SHANG_PAO & wChiHuRight) != 0 {
		//成牌型的
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_GANG_SHANG_PAO
		}
		//特殊情况:7对不满足3n+2牌型的也可以胡杠上炮
		if (wChiHuKind & static.CHK_7_DUI != 0 ||
			wChiHuKind & static.CHK_7_DUI_1 != 0 ||
			wChiHuKind & static.CHK_7_DUI_2 != 0 ||
			wChiHuKind & static.CHK_7_DUI_3 != 0){
			wChiHuKind |= static.CHK_GANG_SHANG_PAO
		}
	}

	// 抢杠胡判断
	if sl.HuType.HAVE_QIANG_GANG_HU && (static.CHR_QIANG_GANG & wChiHuRight) != 0 {
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_QIANG_GANG
		}
	}

	// 海底捞
	if (static.CHR_HAI_DI & wChiHuRight) != 0 {
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0  {
			wChiHuKind |= static.CHK_HAI_DI
		}
	}

	// 海底炮
	if (static.CHR_HAI_DI_PAO & wChiHuRight) != 0 {
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_HAI_DI_PAO
		}
	}

	return wChiHuKind
}

//吃胡判断
func (sl *SportLogicYCKWX) AnalyseChiHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint64, ChiHuResult *static.TagChiHuResult, bNeedAdd bool,checkk5 bool) int {
	//变量定义
	wChiHuKind := uint64(static.CHK_NULL)
	//构造麻将
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

	copy(cbCardIndexTemp, cbCardIndex)

	//当前牌是不是需要带入到手牌中去,自摸来的牌发牌时已加入过,不需要再次加入,点炮来的牌是需要带入手牌计算的
	if cbCurrentCard != 0 && bNeedAdd{
		cbCardIndexTemp[sl.SwitchToCardIndex(cbCurrentCard)]++
	}

	//结果判断硬胡
	wChiHuKind |= sl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbWeaveCount, cbCurrentCard,checkk5)
	ChiHuResult.ChiHuKind = wChiHuKind
	if ChiHuResult.ChiHuKind != static.CHK_NULL {
		return static.WIK_CHI_HU
	}

	return static.WIK_NULL
}

//听牌判断,判断手上听多少张牌，0张表示没听牌
func (sl *SportLogicYCKWX) AnalyseTingCardCount(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint64,checkk5 bool) byte {
	iTingCount := 0
	//变量定义
	var ChiHuResult static.TagChiHuResult

	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]
	//听牌分析
	y := 0
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//胡牌分析
		cbCurrentCard := sl.SwitchToCardData(i)
		cbHuCardKind := sl.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult, true,checkk5)

		//结果判断
		if cbHuCardKind != static.CHK_NULL { //赖子
			y++ //计数
		}
	}

	iTingCount = y
	return byte(iTingCount)
}

// 是否为弃杠牌
func (sl *SportLogicYCKWX) IsGiveUpGang(_userItem *components2.Player, card byte) bool {
	for _, c := range _userItem.Ctx.VecGangCard {
		if c == card {
			return true
		}
	}
	return false
}

// 追加一个弃杠牌
func (sl *SportLogicYCKWX) AppendGiveUpGang(_userItem *components2.Player, card byte) {
	if sl.IsGiveUpGang(_userItem, card) {
		return
	}
	_userItem.Ctx.VecGangCard = append(_userItem.Ctx.VecGangCard, card)
}

//卡五星
func(sl *SportLogicYCKWX) IsKa5Xing(AnalyseItem *static.TagAnalyseItem, cbItemCount byte, cbCurrentCard byte) bool{
	for i:=cbItemCount;i<byte(len(AnalyseItem.WeaveKind));i++{
		if AnalyseItem.WeaveKind[i] & ( static.WIK_CENTER |static.WIK_LEFT | static.WIK_RIGHT)!= 0{
			if (AnalyseItem.CenterCard[i]+1) == sl.SwitchToCardIndex(cbCurrentCard) && (cbCurrentCard == 0x15 || cbCurrentCard == 0x25){
				return true
			}
		}
	}

	return false
}

//手抓一
func (sl *SportLogicYCKWX) IsShouZhuaYi(cbCardIndex []byte,WeaveItem []static.TagWeaveItem, cbItemCount byte, cbCurrentCard byte) bool{
	//构造麻将
	//cbCardIndexTemp := cbCardIndex[:]
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex)

	IsCurrentCardEye :=  true
	cbCurrentCardIndex := byte(0)
	if cbCurrentCard !=0 && cbCurrentCard != static.INVALID_BYTE{
		cbCurrentCardIndex = sl.SwitchToCardIndex(cbCurrentCard)
	}

	//手抓一，胡的就是这最后一张组成的将
	if cbCardIndexTemp[cbCurrentCardIndex] != 2 || sl.GetCardCount(cbCardIndexTemp) != 2{
		return false
	}

	for i:=byte(0);i<cbItemCount;i++{
		switch WeaveItem[i].WeaveKind {
		case static.WIK_PENG:
			if WeaveItem[i].CenterCard == cbCurrentCard{
				IsCurrentCardEye = false
			}
		case static.WIK_GANG:
			if WeaveItem[i].CenterCard == cbCurrentCard{
				IsCurrentCardEye = false
			}
		}
	}

	//手牌必须为2,并且摸来的牌是拿来凑将的,不满足任一条件的都不算手抓一
	if !IsCurrentCardEye{
		return false
	}

	for i:=0;i<static.MAX_INDEX;i++{
		if cbCardIndexTemp[i] == 2 && sl.SwitchToCardIndex(cbCurrentCard) == byte(i){
			return true
		}
	}

	return false
}

//小三元
func (sl *SportLogicYCKWX) IsXiaoSanYuan(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) bool {
	//构造麻将
	//cbCardIndexTemp := cbCardIndex[:]
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex)
	//数目变量
	cbFengCount :=0
	jiangCard := make(map[byte]bool, 3)
	jiangCard[0x35] = true
	jiangCard[0x36] = true
	jiangCard[0x37] = true

	//计算手牌中风牌的刻子数,这里直接从31开始算,31刚好是红中
	for i := byte(31); i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] == 3 && sl.SwitchToCardData(i) >= 0x35 {
			cbFengCount++
			jiangCard[sl.SwitchToCardData(i)] = false
		}
	}

	//对组合牌进行判断
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].CenterCard >= 0x35 {
			cbFengCount ++
			jiangCard[WeaveItem[i].CenterCard] = false
		}
	}

	//必须还有一对中发白的将牌
	bIsExist := true
	for i:=0x35;i<=0x37;i++{
		if jiangCard[byte(i)] {
			if cbCardIndexTemp[sl.SwitchToCardIndex(byte(i))] != 2{
				bIsExist = false
			}
		}
	}

	//数目验证
	if cbFengCount == 2 && bIsExist{
		return true
	}

	return false
}

//大三元
func (sl *SportLogicYCKWX) IsDaSanYuan(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) bool {
	//构造麻将
	//cbCardIndexTemp := cbCardIndex[:]
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex)
	//数目变量
	cbFengCount :=0
	//计算手牌中风牌的刻子数
	for i := byte(0); i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] == 3 && sl.SwitchToCardData(i) >= 0x35 {
			cbFengCount++
		}
	}

	//对组合牌进行判断
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].CenterCard >= 0x35 {
			cbFengCount ++
		}
	}

	//数目验证
	if cbFengCount == 3{
		return true
	}

	return false
}

//四归一
func (sl *SportLogicYCKWX) IsSiGuiYi(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte,  cbCurrentCard byte) (bool,bool) {
	//构造麻将
	//cbCardIndexTemp := cbCardIndex[:]
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

	copy(cbCardIndexTemp, cbCardIndex)
	cbCardColor := byte(cbCurrentCard & static.MASK_COLOR)
	bIsYiSe := true
	cbCurrentCardIndex := byte(0)
	if cbCurrentCard !=0 && cbCurrentCard != static.INVALID_BYTE{
		cbCurrentCardIndex = sl.SwitchToCardIndex(cbCurrentCard)
	}

	//1. 判断是不是一色牌
	//判断手牌
	for i:=0; i<static.MAX_INDEX;i++{
		if cbCardIndexTemp[i] != 0{
			if cbCardColor != (sl.SwitchToCardData(byte(i)) & static.MASK_COLOR) {
				bIsYiSe = false
				break
			}
		}
	}
	//判断组合牌
	for i:=byte(0); i<cbWeaveCount; i++{
		cbCenterCard := WeaveItem[i].CenterCard
		if (cbCenterCard & static.MASK_COLOR) != cbCardColor {
			bIsYiSe = false
			break
		}
	}

	//清一色如果胡四归一,只要手上有4个就可以
	//产品需求：见禅道bug单 http://c.kaayou.cn/index.php?m=bug&f=view&bugID=17706
	//普通牌胡四归一,必须满足条件:这4张牌不能做2句话,必须是1刻子+1句话

	//2. 判断四归一
	IsAn := false
	isSiGuiYi := false
	if bIsYiSe{
		for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[i] == 4{
				isSiGuiYi = true
				IsAn = true //暗四归一
				break
			}
		}

		//提取组合
		for i:=byte(0); i<cbWeaveCount; i++{
			switch WeaveItem[i].WeaveKind {
			//case static.WIK_LEFT:
			//	{
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]++
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard + 1)]++
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard + 2)]++
			//	}
			//case static.WIK_CENTER:
			//	{
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]++
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard + 1)]++
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard - 1)]++
			//	}
			//case static.WIK_RIGHT:
			//	{
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]++
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard - 1)]++
			//		cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard - 2)]++
			//	}
			case static.WIK_PENG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 3
			case static.WIK_GANG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 4
			}
		}

		//遍历牌
		for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[i] == 4{
				//检查是否有杠
				j:=byte(0)
				for ;j<cbWeaveCount;j++{
					if WeaveItem[j].WeaveKind == static.WIK_PENG && WeaveItem[j].CenterCard == sl.SwitchToCardData(byte(i)){
						break
					}
				}

				if j < cbWeaveCount {
					// 明四归一
					isSiGuiYi = true
					IsAn = false
					break
				}
			}
		}
	}else{
		for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[cbCurrentCardIndex] == 4{
				isSiGuiYi = true
				IsAn = true //暗四归一
				break
			}
		}

		//提取组合
		for i:=byte(0); i<cbWeaveCount; i++{
			switch WeaveItem[i].WeaveKind {
			case static.WIK_PENG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 3
			case static.WIK_GANG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 4
			}
		}

		//遍历牌
		for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[cbCurrentCardIndex] == 4{
				//检查是否有杠
				j:=byte(0)
				for ;j<cbWeaveCount;j++{
					if WeaveItem[j].WeaveKind == static.WIK_PENG && WeaveItem[j].CenterCard == sl.SwitchToCardData(byte(i)){
						break
					}
				}

				if j < cbWeaveCount {
					// 明四归一
					isSiGuiYi = true
					IsAn = false
				}
			}
		}
	}
	return isSiGuiYi, IsAn
}

//四归一
func (sl *SportLogicYCKWX) IsSiGuiYi_XiaoGan(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte,  cbCurrentCard byte) (bool,bool,int) {
	//构造麻将
	//cbCardIndexTemp := cbCardIndex[:]
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

	copy(cbCardIndexTemp, cbCardIndex)
	cbCardColor := byte(cbCurrentCard & static.MASK_COLOR)
	bIsYiSe := true
	cbCurrentCardIndex := byte(0)
	if cbCurrentCard !=0 && cbCurrentCard != static.INVALID_BYTE{
		cbCurrentCardIndex = sl.SwitchToCardIndex(cbCurrentCard)
	}

	//1. 判断是不是一色牌
	//判断手牌
	for i:=0; i<static.MAX_INDEX;i++{
		if cbCardIndexTemp[i] != 0{
			if cbCardColor != (sl.SwitchToCardData(byte(i)) & static.MASK_COLOR) {
				bIsYiSe = false
				break
			}
		}
	}
	//判断组合牌
	for i:=byte(0); i<cbWeaveCount; i++{
		cbCenterCard := WeaveItem[i].CenterCard
		if (cbCenterCard & static.MASK_COLOR) != cbCardColor {
			bIsYiSe = false
			break
		}
	}

	//清一色如果胡四归一,只要手上有4个就可以
	//产品需求：见禅道bug单 http://c.kaayou.cn/index.php?m=bug&f=view&bugID=17706
	//普通牌胡四归一,必须满足条件:这4张牌不能做2句话,必须是1刻子+1句话

	//清一色四归一,手上有4个算明四归,胡的那张刚好是第4张才算暗四归

	nMingCount := 0 //清一色明四归的个数,有多少个4个在手上就算多少个明四归
	//2. 判断四归一
	IsAn := false
	isSiGuiYi := false
	if bIsYiSe{
		bIsAnOk := false
		//这里不再用cbCardIndexTemp,经过上面的换算手牌已经变了
		cbCardIndexTemp3 := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
		copy(cbCardIndexTemp3, cbCardIndex)
		_, AnalyseItemArray := sl.AnalyseCard(cbCardIndexTemp3, WeaveItem, cbWeaveCount)
		//胡牌分析
		if len(AnalyseItemArray) > 0 {
			for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
				var pAnalyseItem *static.TagAnalyseItem = &AnalyseItemArray[iCount]
				for j := 0; j < len(pAnalyseItem.WeaveKind); j++ {
					cbWeaveKind := pAnalyseItem.WeaveKind[j]
					//判断这张牌是不是拿来组成刻子
					if cbWeaveKind&static.WIK_PENG != 0 && pAnalyseItem.CenterCard[j] == cbCurrentCardIndex && cbCardIndex[cbCurrentCardIndex] == 4{
						bIsAnOk = true
						break
					}
				}
			}
		}

		//if 4 == cbCardIndexTemp[cbCurrentCardIndex] && cbCardIndex[cbCurrentCardIndex] == 4{
		//	IsAn = true 	//暗四归一
		//}else{
		//	IsAn = false	//明四归一
		//}
		if bIsAnOk{
			IsAn = true 	//暗四归一
		}else{
			IsAn = false	//明四归一
		}

		for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[i] == 4{
				isSiGuiYi = true
			}
		}
		//if 4 == cbCardIndexTemp[cbCurrentCardIndex]{
		//	IsAn = true 	//暗四归一
		//}else{
		//	IsAn = false	//明四归一
		//}

		//提取组合
		for i:=byte(0); i<cbWeaveCount; i++{
			switch WeaveItem[i].WeaveKind {
			case static.WIK_PENG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 3
			case static.WIK_GANG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 4
			}
		}
		//遍历牌
		for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[i] == 4 {
				//if byte(i) != cbCurrentCardIndex {
				//	nMingCount++
				//}
				if !IsAn && cbCardIndex[i] != 0{
					nMingCount++
				}

				if (IsAn&& byte(i) != cbCurrentCardIndex  && cbCardIndex[i] != 0){
					nMingCount++
				}

				//bCanCount := true
				//for m:=byte(0);m<cbWeaveCount;m++{
				//	if WeaveItem[m].WeaveKind == static.WIK_PENG && WeaveItem[m].CenterCard == sl.SwitchToCardData(byte(i)) {
				//			bCanCount = false
				//	}
				//}

				//if bCanCount && cbCurrentCardIndex == byte(i){
				//	nMingCount++
				//}

				//检查是否有杠
				j:=byte(0)
				for ;j<cbWeaveCount;j++{
					if WeaveItem[j].WeaveKind == static.WIK_PENG && WeaveItem[j].CenterCard == sl.SwitchToCardData(byte(i)){
						break
					}
				}

				if j < cbWeaveCount {
					isSiGuiYi = true
					//IsAn = true
					//break

					//if cbCardIndex[i] != 4 && cbCardIndexTemp[i] == 4  {
					//	nMingCount++
					//}
				}
			}
		}

	}else{
		//for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[cbCurrentCardIndex] == 4{
				isSiGuiYi = true
				IsAn = true //暗四归一
				//break
			}
		//}

		//提取组合
		for i:=byte(0); i<cbWeaveCount; i++{
			switch WeaveItem[i].WeaveKind {
			case static.WIK_PENG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 3
			case static.WIK_GANG:
				cbCardIndexTemp[sl.SwitchToCardIndex(WeaveItem[i].CenterCard)]+= 4
			}
		}

		//遍历牌
		//for i:=0;i<static.MAX_INDEX;i++{
			if cbCardIndexTemp[cbCurrentCardIndex] == 4{
				//检查是否有4张
				j:=byte(0)
				for ;j<cbWeaveCount;j++{
					if WeaveItem[j].WeaveKind == static.WIK_PENG && WeaveItem[j].CenterCard == cbCurrentCard {// sl.SwitchToCardData(byte(cbCurrentCardIndex)){
						break
					}
				}

				if j < cbWeaveCount {
					// 明四归一
					isSiGuiYi = true
					IsAn = false
				}
			}
		//}

		//满足条件：成牌型中必须满足这4张牌不能做2句话,必须是1刻子+1句话
		bIsOk := false
		//这里不再用cbCardIndexTemp,经过上面的换算手牌已经变了
		cbCardIndexTemp2 := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
		copy(cbCardIndexTemp2, cbCardIndex)
		_, AnalyseItemArray := sl.AnalyseCard(cbCardIndexTemp2, WeaveItem, cbWeaveCount)
		//胡牌分析
		if len(AnalyseItemArray) > 0 {
			for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
				var pAnalyseItem *static.TagAnalyseItem = &AnalyseItemArray[iCount]
				for j := 0; j < len(pAnalyseItem.WeaveKind); j++ {
					cbWeaveKind := pAnalyseItem.WeaveKind[j]
					//判断这张牌是不是拿来组成刻子
					if cbWeaveKind&static.WIK_PENG != 0 && pAnalyseItem.CenterCard[j] == cbCurrentCardIndex{
						bIsOk = true
						break
					}
				}
			}
		}

		//这4张牌不能做2句话,必须是1刻子+1句话
		if !bIsOk{
			isSiGuiYi = false
		}
	}
	return isSiGuiYi, IsAn, nMingCount
}

//数坎
func(sl *SportLogicYCKWX) ShuKan(AnalyseItem *static.TagAnalyseItem, WeaveItem []static.TagWeaveItem, cbItemCount byte, cbCurrentCard byte, bZimo bool) int{
	//数坎：胡牌时手牌中刻子和已杠出的杠，每有一副+1分
	//如果当前牌是点炮来的,这张牌凑成的坎不计
	kanCount := 0
	//1.算组合牌里面的杠
	for i:=byte(0);i<cbItemCount;i++{
		if WeaveItem[i].WeaveKind & static.WIK_GANG != 0  {
			if WeaveItem[i].CenterCard == cbCurrentCard && !bZimo{
				continue
			}
			kanCount++
		}
	}
	//2.算剩下牌里面的杠、3张
	for i:=cbItemCount;i<byte(len(AnalyseItem.WeaveKind));i++{
		if AnalyseItem.WeaveKind[i] & static.WIK_PENG != 0 {
			if  AnalyseItem.CenterCard[i] == cbCurrentCard && !bZimo{
				continue
			}
			kanCount++
		}

		if AnalyseItem.WeaveKind[i] & static.WIK_GANG != 0 {
			if AnalyseItem.CenterCard[i] == cbCurrentCard && !bZimo{
				continue
			}
			kanCount++
		}
	}

	return kanCount
}

//亮牌判断
func (sl *SportLogicYCKWX) AnalyseLiangCard(cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint64, sendcard byte,checkk5 bool) (bool, []byte) {
	//变量定义
	var cbOutCard []byte
	wChiHuKind := uint64(static.CHK_NULL)
	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]

	canLiang := false
	for i:=0;i<static.MAX_INDEX;i++{
		if cbCardIndexTemp[i] != 0{
			cbCardIndexTemp[i]--
			for j:=0;j<static.MAX_INDEX;j++{
				cbCardIndexTemp[j]++
				cbCurrentCard := sl.SwitchToCardData(byte(j))
				wChiHuKind |= sl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard,checkk5)
				cbCardIndexTemp[j]--
				if wChiHuKind != static.CHK_NULL {
					canLiang =  true
					cbOutCard = append(cbOutCard, sl.SwitchToCardData(byte(i)))
					wChiHuKind = static.CHK_NULL
					break
				}
			}
			cbCardIndexTemp[i]++
			if canLiang {
				//break
			}
		}
	}

	return canLiang,cbOutCard
}

//亮牌判断
//func (self *SportLogicYCKWX) AnalyseLiangCard2(cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint64, sendcard byte) (bool, []byte) {
//	//变量定义
//	var cbOutCard []byte
//	wChiHuKind := uint64(static.CHK_NULL)
//	//构造扑克
//	cbCardIndexTemp := cbCardIndex[:]
//	//if sendcard != 0{
//	//	cbCardIndexTemp[self.SwitchToCardIndex(sendcard)]++
//	//}
//	//wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbItemCount, sendcard)
//	//if wChiHuKind != static.CHK_NULL {
//	//	return true,nil
//	//}
//	liangMap := make(map[byte][]byte, 10)
//	canLiang := false
//	for i:=0;i<static.MAX_INDEX;i++{
//		if cbCardIndexTemp[i] != 0{
//			cbCardIndexTemp[i]--
//			for j:=0;j<static.MAX_INDEX;j++{
//				cbCardIndexTemp[j]++
//				cbCurrentCard := self.SwitchToCardData(byte(j))
//				wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard)
//				cbCardIndexLiang := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
//				copy(cbCardIndexLiang, cbCardIndexTemp)
//				cbCardIndexTemp[j]--
//				if wChiHuKind != static.CHK_NULL {
//					canLiang =  true
//					cbOutCard = append(cbOutCard,self.SwitchToCardData(byte(i)))
//					wChiHuKind = static.CHK_NULL
//
//					_, AnalyseItemArray := self.AnalyseCard(cbCardIndexLiang, WeaveItem, cbItemCount)
//					if len(AnalyseItemArray) > 0 {
//						//牌型分析
//						for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
//							//变量定义
//							var pAnalyseItem *static.TagAnalyseItem = &AnalyseItemArray[iCount]
//							for i:=cbItemCount;i<byte(len(pAnalyseItem.WeaveKind));i++{
//								switch WeaveItem[i].WeaveKind {
//								case static.WIK_LEFT:
//									{
//										if WeaveItem[i].CenterCard == self.SwitchToCardData(WeaveItem[i].CenterCard){
//											liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard + 1)
//											liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard + 2)
//										}
//									}
//								case static.WIK_CENTER:
//									{
//										if WeaveItem[i].CenterCard == self.SwitchToCardData(WeaveItem[i].CenterCard){
//											liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard + 1)
//											liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard - 1)
//										}
//									}
//								case static.WIK_RIGHT:
//									{
//										if WeaveItem[i].CenterCard == self.SwitchToCardData(WeaveItem[i].CenterCard){
//											liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard - 1)
//											liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard - 2)
//										}
//									}
//								case static.WIK_PENG:
//								case static.WIK_GANG:
//									if WeaveItem[i].CenterCard == self.SwitchToCardData(WeaveItem[i].CenterCard){
//										liangMap[self.SwitchToCardData(byte(j))] = append(liangMap[self.SwitchToCardData(byte(j))], WeaveItem[i].CenterCard)
//									}
//
//								}
//							}
//						}
//					}
//					break
//				}
//			}
//			cbCardIndexTemp[i]++
//			if canLiang {
//				//break
//			}
//		}
//	}
//	fmt.Println("亮牌数据: ",liangMap)
//	return canLiang,cbOutCard
//}
//

//听牌判断,判断玩家手牌能听的最大牌型
func (sl *SportLogicYCKWX) AnalyseMAXTingType(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint64,checkk5 bool) uint64 {
	//变量定义
	var ChiHuResult static.TagChiHuResult
	cbMaxHuKind := uint64(static.CHK_NULL)
	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]

	//听牌分析
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//胡牌分析
		cbCurrentCard := sl.SwitchToCardData(i)
		cbHuAction := sl.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult, true,checkk5)
		//结果判断
		if cbHuAction != static.WIK_NULL {
			if 0 < sl.ComporeMAXTingType(cbMaxHuKind , ChiHuResult.ChiHuKind){
				cbMaxHuKind = ChiHuResult.ChiHuKind
			}
		}
	}

	return cbMaxHuKind
}

//听牌判断,比较玩家手牌能听的最大牌型，返回0表示相等，-1表示ChiHuKind1大，1表示ChiHuKind2大
func (sl *SportLogicYCKWX) ComporeMAXTingType(ChiHuKind1 uint64,ChiHuKind2 uint64) int {
	if ChiHuKind1 == 0 && ChiHuKind2 != 0{
		return 1
	}
	fenshu1 := sl.GetPlayerHuScore(ChiHuKind1)
	wFanshu1,lFanshu1 := sl.GetPeiZhuangFan(ChiHuKind1)
	fenshu2 := sl.GetPlayerHuScore(ChiHuKind2)
	wFanshu2,lFanshu2 := sl.GetPeiZhuangFan(ChiHuKind2)
	if fenshu1*wFanshu1*lFanshu1 > fenshu2*wFanshu2*lFanshu2{
		return -1
	}else if fenshu1*wFanshu1*lFanshu1 < fenshu2*wFanshu2*lFanshu2{
		return 1
	}
	return 0
}

//! 玩家胡牌分数 ，与Game_mj_k5x的一致
func (sl *SportLogicYCKWX) GetPlayerHuScore(ChiHuKind uint64) int {
	huScore := 1
	//碰碰胡
	if (ChiHuKind & static.CHK_PENG_PENG) != 0 {
		if sl.Rule.PPx4{
			//碰碰胡x4
			huScore*=4
			//fmt.Println("碰碰胡4分")
		}else{
			//碰碰胡x2
			huScore*=2
			//fmt.Println("碰碰胡2分")
		}
	}
	//卡五星
	if (ChiHuKind & static.CHK_KA_5_XING) != 0 {
		if sl.Rule.K5x4{
			//卡五星x4
			huScore*=4
			//fmt.Println("卡五星4分")
		}else{
			//卡五星x2
			huScore*=2
			//fmt.Println("卡五星2分")
		}
	}
	//明四归一x2
	if (ChiHuKind & static.CHK_SI_GUI_YI_MING) != 0 {
		huScore*=2
		//fmt.Println("明四归一2分")
	}
	//暗四归一x4
	if (ChiHuKind & static.CHK_SI_GUI_YI_AN) != 0 {
		huScore*=4
		//fmt.Println("暗四归一4分")
	}
	//手抓一x4
	if (ChiHuKind & static.CHK_SHOU_ZHUA_YI) != 0 {
		huScore*=4
		//fmt.Println("手抓一4分")
	}
	//7对x4
	if (ChiHuKind & static.CHK_7_DUI) != 0 ||
		(ChiHuKind & static.CHK_7_DUI_1) != 0 ||
		(ChiHuKind & static.CHK_7_DUI_2) != 0 ||
		(ChiHuKind & static.CHK_7_DUI_3) != 0 {
		huScore*=4
		//fmt.Println("七对4分")
	}
	//清一色x4
	if (ChiHuKind & static.CHK_QING_YI_SE) != 0 {
		huScore*=4
		//fmt.Println("清一色4分")
	}
	//小三元x4
	if (ChiHuKind & static.CHK_XIAO_SAN_YUAN) != 0 {
		huScore*=4
		//fmt.Println("小三元4分")
	}
	//大三元x8
	if (ChiHuKind & static.CHK_DA_SAN_YUAN) != 0 {
		huScore*=8
		//fmt.Println("大三元8分")
	}

	//fmt.Println(fmt.Sprintf("牌型分：%d",huScore))

	return huScore
}

//! 玩家胡牌番数，与Game_mj_k5x的一致
func (sl *SportLogicYCKWX) GetPeiZhuangFan(ChiHuKind uint64) (int,int) {

	winerFanNum := 0
	LoserFanNum := 0

	//豪华七对+1
	if ChiHuKind & static.CHK_7_DUI_1 != 0 {
		winerFanNum += 1
	}
	//双豪华七对+2
	if ChiHuKind  & static.CHK_7_DUI_2 != 0 {
		winerFanNum += 2
	}
	//三豪华七对+3
	if ChiHuKind  & static.CHK_7_DUI_3 != 0 {
		winerFanNum += 3
	}

	//清一色明四归加番
	if ChiHuKind & static.CHK_QING_YI_SE != 0 {
		if ChiHuKind &CHK_M4G_1 != 0{
			winerFanNum += 1
		}else if ChiHuKind &CHK_M4G_2 != 0{
			winerFanNum += 2
		}else if ChiHuKind &CHK_M4G_3 != 0{
			winerFanNum += 3
		}
	}

	wfan := 1
	lfan := 1
	for i:=0;i<winerFanNum;i++{
		wfan *= 2
	}
	for i:=0;i<LoserFanNum;i++{
		lfan *= 2
	}

	return wfan,lfan
}

func (sl *SportLogicYCKWX) AnalyseShukan(cbCardIndex []byte, weaveItem []static.TagWeaveItem, cbItemCount byte, cbCurrentCard byte, bZimo bool) int{
	//变量定义
	handkanCount := 0
	weaveKanCount := 0
	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]
	_, AnalyseItemArray := sl.AnalyseCard(cbCardIndexTemp, weaveItem, cbItemCount)
	//胡牌分析
	if len(AnalyseItemArray) > 0 {

		//数坎：胡牌时手牌中刻子和已杠出的杠，每有一副+1分
		//如果当前牌是点炮来的,这张牌凑成的坎不计
		//1.算组合牌里面的杠
		for i:=byte(0);i<cbItemCount;i++{
			if weaveItem[i].WeaveKind & static.WIK_GANG != 0  {
				if weaveItem[i].CenterCard == cbCurrentCard && !bZimo{
					continue
				}
				handkanCount++
			}
		}

		//牌型分析

		for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
			//变量定义
			var pAnalyseItem *static.TagAnalyseItem = &AnalyseItemArray[iCount]
			//2.算剩下牌里面的杠、3张
			for i:=cbItemCount;i<byte(len(pAnalyseItem.WeaveKind));i++{
				if pAnalyseItem.WeaveKind[i] & static.WIK_PENG != 0 {
					if  pAnalyseItem.CenterCard[i] == sl.SwitchToCardIndex(cbCurrentCard) && !bZimo{
						continue
					}
					weaveKanCount++
				}

				if pAnalyseItem.WeaveKind[i] & static.WIK_GANG != 0 {
					if pAnalyseItem.CenterCard[i] == sl.SwitchToCardIndex(cbCurrentCard) && !bZimo{
						continue
					}
					weaveKanCount++
				}
			}
			if weaveKanCount > 0 {
				break
			}
		}
	}
	kanCount := weaveKanCount + handkanCount

	return kanCount
}

func (sl *SportLogicYCKWX) IsShowCard(wCurrentUser uint16, card byte, liang [4]meta2.ShowCard) bool{
	if card != 0 {
		for _,v := range liang[wCurrentUser].CbLiangCard{
			if v != 0 && v == card{
				return true
			}
		}
	}

	return false
}

func (sl *SportLogicYCKWX) IsOthersTingCard(wCurrentUser uint16, card byte, liang [4]meta2.ShowCard) bool{
	if card != 0 {
		for i:=uint16(0);i<4;i++{
			if i == wCurrentUser{
				continue
			}
			for _,v := range liang[i].CbTingCard{
				if v != 0 && v == card{
					return true
				}
			}
		}
	}

	return false
}

func (sl *SportLogicYCKWX) IsUserAnPuCard(wCurrentUser uint16, card byte, liang [4]meta2.ShowCard) bool{
	if card != 0 {
		for _,v := range liang[wCurrentUser].CbAnPuCard{
			if v != 0 && v == card{
				return true
			}
		}
	}

	return false
}

//杠牌判断
func (sl *SportLogicYCKWX) EstimateGangCard(_userItem *components2.Player, cbCurrentCard byte,  liang [4]meta2.ShowCard) int {
	//参数效验
	if !sl.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}

	//杠牌判断
	var GangCardResult static.TagGangCardResult
	_userItem.Ctx.DispatchCard(cbCurrentCard)
	action := sl.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, nil, 0, &GangCardResult, liang)
	_userItem.Ctx.DeleteCard(cbCurrentCard)

	bIsCurCardGang := false
	for _,card := range GangCardResult.CardData{
		if card == cbCurrentCard{
			bIsCurCardGang = true
			break
		}
	}

	if action & static.WIK_GANG != 0 && !bIsCurCardGang{
		action ^= static.WIK_GANG
	}

	return action
}