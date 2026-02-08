// ////////////////////////////////////////////////////////////////////////
//
//	      --
//	  --       --
//	--  纸牌游戏逻辑  --
//	  --       --
//	      --
//
// ////////////////////////////////////////////////////////////////////////
package Hubei_JianLi

//import "fmt"

import (
	"fmt"
	public "github.com/open-source/game/chess.git/pkg/static"
	modules "github.com/open-source/game/chess.git/services/sport/components"
	info "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	logic "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	"math/rand"
	"strings"
	"time"
)

type GameLogic_jl_jlkj struct {
	logic.BaseLogicDG
	gameTimer    modules.GameTime //时间计时器
	Hard510KMode bool             //纯510K是否大于四炸
}

// 20181214 每个玩法设定自己的限时操作时间
func (self *GameLogic_jl_jlkj) Setlimitetime(limitetimeOp bool) {
	//if limitetimeOp {
	//	self.Rule.limitetimeOP = GameTime_Nine
	//} else {
	//	self.Rule.limitetimeOP = 0
	//}
}

// 设置炸弹的张数
func (self *GameLogic_jl_jlkj) SetHard510KMode(byhard510k bool) {
	self.Hard510KMode = byhard510k
}

func (self *GameLogic_jl_jlkj) CreateCards() (byte, [public.ALL_CARD]byte) {
	cbCardDataTemp := [public.ALL_CARD]byte{}
	_maxCount := byte(public.ALL_CARD)
	//初始化所有牌点数
	for i := 0; i < logic.MAX_POKER_COUNTS; i++ {
		for j := 0; j < 2; j++ {
			cbCardDataTemp[i+j*logic.MAX_POKER_COUNTS] = byte(i) + 1
		}
	}
	return _maxCount, cbCardDataTemp
}

// 混乱扑克 并发牌
func (self *GameLogic_jl_jlkj) RandCardData(byAllCards [public.ALL_CARD]byte) ([meta.MAX_PLAYER]byte, [meta.MAX_PLAYER][public.MAX_CARD]byte, [meta.MAX_PLAYER]int) {
	cbCardDataTemp := [meta.MAX_PLAYER][public.MAX_CARD]byte{}
	_maxCount := [meta.MAX_PLAYER]byte{}
	_KingCount := [meta.MAX_PLAYER]int{}

	_randTmp := time.Now().Unix()
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	//洗牌
	for i := 0; i < 1000; i++ {
		rand_num := rand.Intn(1000)
		m := rand_num % (logic.MAX_POKER_COUNTS * 2)
		rand_num_2 := rand.Intn(1000)
		n := rand_num_2 % (logic.MAX_POKER_COUNTS * 2)
		zz := byAllCards[m]
		byAllCards[m] = byAllCards[n]
		byAllCards[n] = zz
	}

	//清零
	for i := 0; i < meta.MAX_PLAYER; i++ {
		for j := 0; j < public.MAX_CARD; j++ {
			cbCardDataTemp[i][j] = 0
		}
	}

	//发牌
	for i := 0; i < public.MAX_CARD && i < int(self.MaxCardCount); i++ {
		for j := 0; j < meta.MAX_PLAYER && j < int(self.MaxPlayerCount); j++ {
			if self.IsValidCard(byAllCards[int(self.MaxPlayerCount)*i+j]) {
				cbCardDataTemp[j][i] = byAllCards[int(self.MaxPlayerCount)*i+j]
				_maxCount[j]++
				if cbCardDataTemp[j][i] == logic.CARDINDEX_SMALL || cbCardDataTemp[j][i] == logic.CARDINDEX_BIG {
					_KingCount[j]++
				}
			}
		}
	}

	return _maxCount, cbCardDataTemp, _KingCount
}

// 牌有效判断
func (self *GameLogic_jl_jlkj) IsValidCard(cbCardData byte) bool {
	//校验
	if cbCardData == 0 || cbCardData > logic.CARDINDEX_BIG {
		return false
	}
	return true
}

func (self *GameLogic_jl_jlkj) Compare(typeFirst public.TCardType, typeFollow public.TCardType) bool {
	if typeFirst.Cardtype == public.TYPE_ERROR {
		return false
	}
	if typeFollow.Cardtype == public.TYPE_ERROR || typeFollow.Cardtype == public.TYPE_NULL {
		return false
	}

	//第一种情况，首出
	if typeFirst.Cardtype == public.TYPE_NULL {
		if typeFollow.Cardtype == public.TYPE_ONE || typeFollow.Cardtype == public.TYPE_TWO || typeFollow.Cardtype == public.TYPE_TWOSTR || typeFollow.Cardtype == public.TYPE_ONESTR || typeFollow.Cardtype == public.TYPE_BOMB_510K || typeFollow.Cardtype == public.TYPE_BOMB_8XI || typeFollow.Cardtype == public.TYPE_BOMB_DOUBLE_KING || typeFollow.Cardtype == public.TYPE_BOMB_FOUR_KING || typeFollow.Cardtype == public.TYPE_BOMB_NOMORL {
			return true
		} else if typeFollow.Cardtype == public.TYPE_THREE || typeFollow.Cardtype == public.TYPE_THREESTR {
			//如果王是赖子：大冶王是赖子，蕲春王不是赖子，通过王是赖子来区分是否是大冶打拱，大冶有TYPE_THREE和TYPE_THREESTR
			return true
		} else {
			return false
		}
	} else { //第二种情况，跟出的人出的不是炸弹，类型必须和首出一致//其他牌的比较，非炸弹
		if (typeFollow.Cardtype == public.TYPE_BOMB_510K) || (typeFollow.Cardtype == public.TYPE_BOMB_NOMORL) || (typeFollow.Cardtype == public.TYPE_BOMB_8XI) || (typeFollow.Cardtype == public.TYPE_BOMB_DOUBLE_KING) || (typeFollow.Cardtype == public.TYPE_BOMB_FOUR_KING) {
			//非拱笼类型的BombLevel =0,
			if typeFollow.BombLevel > typeFirst.BombLevel || (typeFollow.BombLevel == typeFirst.BombLevel && self.GetCardLevel(typeFollow.Card) > self.GetCardLevel(typeFirst.Card)) {
				return true
			} else {
				return false
			}
		} else if typeFollow.Cardtype == typeFirst.Cardtype { //跟出的人出的不是炸弹，类型必须和首出一致
			if typeFollow.Len == typeFirst.Len && self.GetCardLevel(typeFollow.Card) > self.GetCardLevel(typeFirst.Card) {
				return true
			} else {
				return false
			}
		} else { //跟出的人出的不是炸弹，类型和首出也一致
			return false
		}
	}
	return false
}

func (self *GameLogic_jl_jlkj) GetType(card_list [public.MAX_CARD]byte, cardlen int, outMagicNum byte, byType byte, outtype int) public.TCardType {
	var re public.TCardType
	re.Len = 0
	re.Card = 0
	re.Color = 0
	re.Cardtype = public.TYPE_NULL
	re.Count = 0
	re.BombLevel = 0

	card := byte(0)
	self.SortByIndex(card_list, cardlen, true)
	re.Len = int(self.GetCardNum(card_list, byte(cardlen)))
	if re.Len < 1 {
		return re
	}
	re.Card = card_list[re.Len-1]
	switch re.Len {
	case 0:
		re.Cardtype = public.TYPE_NULL
		return re
	case 1:
		re.Cardtype = public.TYPE_ONE
		////如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//if(self.IsMagic(CARDINDEX_BIG) && self.IsKing(re.Card)){
		//	re.Card = 2;//王单出算2
		//}
		return re
	case 2:
		//if !self.IsKing(card_list[0])&&!self.IsKing(card_list[1]){
		if self.GetCardLevel(card_list[0]) == self.GetCardLevel(card_list[1]) {
			re.Cardtype = public.TYPE_TWO
		} else {
			re.Cardtype = public.TYPE_ERROR
		}
		//}else if (self.IsKing(card_list[0]) && self.IsKing(card_list[1])){
		////如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//if self.IsMagic(CARDINDEX_BIG) {
		//	re.Card = 2;//大小王按2来算
		//	re.Cardtype = public.TYPE_TWO;
		//}else {
		//	//对王
		//	if (card_list[0] == card_list[1]) {
		//		if (card_list[0] == CARDINDEX_SMALL) {
		//			re.BombLevel = 60 + 1; //对小王，比6张的炸弹大一些
		//		} else
		//		{
		//			re.BombLevel = 60 + 1; //对大王，比6张的炸弹大一些
		//		}
		//	} else{
		//		// 大小王组成的一对
		//		re.BombLevel = 60 + 1;
		//	}
		//	re.Cardtype = public.TYPE_BOMB_DOUBLE_KING;
		//}
		//re.Cardtype = public.TYPE_TWO;
		//}else{
		//	//有一张王
		//	//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//	if self.IsMagic(CARDINDEX_BIG) {
		//		re.Card = card_list[1];
		//		if self.IsMagic(card_list[1]) {
		//			re.Card = card_list[0];
		//		}
		//		re.Cardtype = public.TYPE_TWO;
		//	}else {
		//		re.Cardtype = public.TYPE_ERROR;
		//	}
		//}
		return re
	case 3:
		color := byte(0)
		is510k := false
		color, is510k = self.Is510KBomb(card_list, 3)
		isBombNormal := false
		card, isBombNormal = self.IsNormalBomb(card_list, cardlen)
		if is510k {
			re.Color = color
			re.Cardtype = public.TYPE_BOMB_510K
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			//if self.IsMagic(CARDINDEX_BIG) {//王是赖子
			re.Card = 5
			if color == 0 {
				re.BombLevel = 30 + 1 //比4张的炸弹小一点
			} else {
				if self.Hard510KMode {
					//纯510K可压所有4张炸弹
					re.BombLevel = 40 + int(color)
				} else {
					re.BombLevel = 30 + int(color) + 1 //比4张的炸弹小一点，比杂510k大(+1)
				}
			}
			//}else {
			//	if(color == 0){
			//		re.BombLevel = 40 + 1;//比4张的炸弹大一点
			//	} else{
			//		re.BombLevel = 50 + int(color);//比5张的炸弹大一点
			//	}
			//}
		} else if self.IsKing(card_list[0]) && self.IsKing(card_list[1]) && self.IsKing(card_list[2]) {
			////三个王
			////如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
			//if self.IsMagic(CARDINDEX_BIG) {
			//	re.Color = 4;//3个大小王按黑桃正510k来算
			//	re.BombLevel = 30 + 4+1; //比4张的炸弹小一些，比杂510k大(+1)
			//	re.Cardtype = public.TYPE_BOMB_510K;
			//	re.Card = 5;//3个大小王按黑桃正510k来算
			//	//re.Card = 3;//3个大小王按3来算
			//	//re.Cardtype = public.TYPE_THREE;
			//}else {
			//	re.BombLevel = 60 + 4; //比6张的炸弹大一些
			//	re.Cardtype = public.TYPE_BOMB_DOUBLE_KING;
			//}
			//监利开机三个王不能出
			re.Cardtype = public.TYPE_ERROR
		} else if isBombNormal {
			re.BombLevel = re.Len * 10
			re.Card = card
			re.Cardtype = public.TYPE_BOMB_NOMORL
		} else {
			//ONESTR
			var refakepoker []public.TFakePoker
			iMagicNum := 0
			iMagicNum, refakepoker = self.GetMagicNum(card_list, cardlen, refakepoker)
			if byType != 0 { //非0表示没有赖子模式
				iMagicNum = 0
			}
			if iMagicNum <= 8 { //有1-4个王
				var ty public.FakeType
				ty = self.GetTypeByMagic(card_list, cardlen, iMagicNum, outtype)
				re.Cardtype = ty.CardType.Cardtype
				re.Card = ty.CardType.Card
			} else {
				re.Cardtype = public.TYPE_ERROR
			}
		}
		return re
	default: // >= 4张
		if re.Len == 4 {
			card = byte(0)
			if self.IsKingBomb(card_list, cardlen) { //四个王最大
				//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
				if self.IsMagic(logic.CARDINDEX_BIG) {
					re.BombLevel = 40 //4个王就是4张的炸弹
					re.Card = 2       //4个王就是4个2
					re.Cardtype = public.TYPE_BOMB_NOMORL
				} else {
					//re.BombLevel = 60 + 2000;
					re.BombLevel = 80 + 1 //比7张的炸弹大些，比8张的炸弹小
					re.Cardtype = public.TYPE_BOMB_FOUR_KING
				}
				return re
			}
		}

		card = byte(0)
		isBombNormal := false
		card, isBombNormal = self.IsNormalBomb(card_list, cardlen)
		if isBombNormal {
			re.Card = card
			if re.Len == 8 {
				re.Cardtype = public.TYPE_BOMB_8XI
			} else {
				re.Cardtype = public.TYPE_BOMB_NOMORL
			}
			re.BombLevel = re.Len * 10
			return re
		}
		//单顺，双顺等
		var refakepoker []public.TFakePoker
		iMagicNum := 0
		iMagicNum, refakepoker = self.GetMagicNum(card_list, cardlen, refakepoker)
		if byType != 0 { //非0表示没有赖子模式，给服务器用
			iMagicNum = 0
		}
		if iMagicNum <= 8 { //有1-4个王
			var ty public.FakeType
			ty = self.GetTypeByMagic(card_list, cardlen, iMagicNum, outtype)
			re.Cardtype = ty.CardType.Cardtype
			re.Card = ty.CardType.Card
		} else {
			re.Cardtype = public.TYPE_ERROR
		}
		return re
	}
	return re
}

// 是否5,10,K分牌
func (self *GameLogic_jl_jlkj) isScorePai(card byte) (bRet bool) {
	if self.GetCardLevel(card) == self.GetCardLevel(5) {
		return true
	} else if self.GetCardLevel(card) == self.GetCardLevel(10) {
		return true
	} else if self.GetCardLevel(card) == self.GetCardLevel(13) {
		return true
	}
	return false
}

func (self *GameLogic_jl_jlkj) GetWriteOutReplayRecordString(replayRecord meta.DG_Replay_Record) string {
	upd := false
	endMsgUpdateScore := [meta.MAX_PLAYER]float64{}
	ourCardStr := ""
	for k, record := range replayRecord.R_Orders {
		// 如果是分数变化 ,为什么不能写在case里面？
		if record.R_Opt == info.E_GameScore {
			flag := false
			// 如果是最后结算的update 则挪到最后追加
			for j := k; j < len(replayRecord.R_Orders); j++ {
				// 只要后面还有别的操作，说明是中途及时结算，按正常的逻辑走
				if replayRecord.R_Orders[j].R_Opt != info.E_GameScore {
					flag = true
					break
				}
			}
			// 如果是中途结算 或者没有大胡 这里就直接追加
			if flag {
				ourCardStr += fmt.Sprintf(",%d:", record.R_ChairId)
				if fs := strings.Split(fmt.Sprintf("%v", record.UserScorePL), "."); len(fs) == 1 {
					ourCardStr += fmt.Sprintf("U%s", fs[0])
				} else {
					ourCardStr += fmt.Sprintf("U%0.2f", record.UserScorePL)
				}
			} else {
				upd = true
				endMsgUpdateScore[record.R_ChairId] = record.UserScorePL
			}
			continue
		}

		switch record.R_Opt {
		case meta.DG_REPLAY_OPT_HOUPAI:
			ourCardStr += fmt.Sprintf("|H%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_QIANG:
			ourCardStr += fmt.Sprintf("|Q%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_END_HOUPAI:
			ourCardStr += fmt.Sprintf("|E%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_END_GAME:
			ourCardStr += fmt.Sprintf("|G%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_DIS_GAME:
			ourCardStr += fmt.Sprintf("|J%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_OUTCARD:
			if len(record.R_Value) > 0 {
				ourCardStr += fmt.Sprintf("|C%d:", record.R_ChairId)
			} else {
				ourCardStr += fmt.Sprintf("|C%d", record.R_ChairId)
			}
			break
		case meta.DG_REPLAY_OPT_TURN_OVER:
			if len(record.R_Opt_Ext) > 0 {
				ourCardStr += fmt.Sprintf("|T%d:", record.R_ChairId)
			} else {
				ourCardStr += fmt.Sprintf("|T%d", record.R_ChairId)
			}
			break
		case meta.DG_REPLAY_OPT_TUOGUAN:
			ourCardStr += fmt.Sprintf("|D%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_4KINGSCORE:
			ourCardStr += fmt.Sprintf("|K%d:", record.R_ChairId)
			break
		case meta.DG_REPLAY_OPT_RESTART:
			ourCardStr += fmt.Sprintf("|R%d:", record.R_ChairId)
			break
		default:
			break
		}

		if len(record.R_Value) > 0 {
			for i := 0; i < len(record.R_Value); i++ {
				ourCardStr += fmt.Sprintf("%s", self.GetStringByCard(byte(record.R_Value[i])))
			}
		}

		//打出的分牌
		if len(record.R_ScoreCard) > 0 {
			fakeStr := ""
			for i := 0; i < len(record.R_ScoreCard); i++ {
				fakeStr += fmt.Sprintf("%s", self.GetStringByCard(byte(record.R_ScoreCard[i])))
			}
			ourCardStr += fmt.Sprintf(",F%s", fakeStr)
		}

		if len(record.R_Opt_Ext) > 0 {
			for i := 0; i < len(record.R_Opt_Ext); i++ {
				if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_HOUPAI {
					ourCardStr += fmt.Sprintf(",H%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_QIANG {
					ourCardStr += fmt.Sprintf(",Q%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_TURNSCORE {
					ourCardStr += fmt.Sprintf(",S%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_GETSCORE {
					ourCardStr += fmt.Sprintf(",G%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_MINGJI {
					ourCardStr += fmt.Sprintf(",J")
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_ENDQIANG {
					ourCardStr += fmt.Sprintf(",B%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_CARDTYPE {
					ourCardStr += fmt.Sprintf(",T%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_TUOGUAN {
					ourCardStr += fmt.Sprintf(",D%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_4KINGSCORE {
					ourCardStr += fmt.Sprintf(",K%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta.DG_EXT_RESTART {
					ourCardStr += fmt.Sprintf(",R%d", record.R_Opt_Ext[i].Ext_value)
				}
			}
		}
	}

	if upd {
		for i, s := range endMsgUpdateScore {
			ourCardStr += fmt.Sprintf(",%d:", i)
			if fs := strings.Split(fmt.Sprintf("%v", s), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f", s)
			}
		}
	}

	return ourCardStr
}
