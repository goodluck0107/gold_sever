package gymnasium

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/open-source/game/chess.git/services/sport/infrastructure"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_JianLi"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_JingZhou"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_JingZhou/JingZhouHuaPai"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_RunFast"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_ShiShou/ShiShou510k"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_ShiShou_AH"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_YingCheng/YingChengKwx"
)

func init() {
	infrastructure.CreateGameFunc = generateSport
}

func generateSport(kindId int) (b bool, v infrastructure.SportInterface) {
	switch kindId {
	case infrastructure.KIND_ID_JZ_CXZ: //荆州搓虾子 442
		v = new(Hubei_JingZhou.SportJZCXZ)
	case infrastructure.KIND_ID_SS_510K: //石首510K 395
		v = new(ShiShou510k.SportSS510K)
	case infrastructure.KIND_ID_MJ_SHAH: //石首捱晃 400
		v = new(Hubei_ShiShou_AH.SportSSAH)
	case infrastructure.KIND_ID_DG_PDK: //跑得快3人好友 452
		v = new(Hubei_RunFast.SportPDK)
	case infrastructure.KIND_ID_MJ_JZHZG: //荆州红中杠 419
		v = new(Hubei_JingZhou.SportJZHZG)
	case infrastructure.KIND_ID_ZP_JZHP: //荆州花牌 436
		v = new(JingZhouHuaPai.Sport_zp_jzhp)
	case infrastructure.KIND_ID_MJ_YCKWX: //荆州花牌 436
		v = new(YingChengKwx.SportYCKWX)
	case infrastructure.KIND_ID_JL_JLMJ: //监利麻将 472
		v = new(Hubei_JianLi.Game_mj_jl_jlmj)
	case infrastructure.KIND_ID_JL_JLKJ: //监利开机 464
		v = new(Hubei_JianLi.Game_jl_jlkj)
	case infrastructure.KIND_ID_MJ_JLHZLZG: //监利红中癞子杠 420
		v = new(Hubei_JianLi.Game_mj_jlhzlzg)
	default:
		xlog.Logger().Errorln("无法识别的KindId：", kindId)
		return false, nil
	}
	return true, v
}
