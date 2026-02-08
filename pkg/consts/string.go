package consts

import "fmt"

func WealthTypeString(wt int8) string {
	switch wt {
	case WealthTypeCard:
		return "房卡"
	case WealthTypeGold:
		return "金币"
	case WealthTypeCoupon:
		return "礼券"
	case WealthTypeVitamin:
		return "比赛分"
	case WealthTypeDiamond:
		return "钻石"
	case WealthTypeCardRcd:
		return "道具-记牌器"
	default:
		return fmt.Sprintf("财富%d", wt)
	}
}
