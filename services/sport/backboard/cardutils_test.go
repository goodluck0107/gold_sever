package backboard

//
//import (
//	"fmt"
//	"testing"
//)
//
//func Test_CreateCards(t *testing.T) {
//	//先声明是全牌
//	var cordclass int
//	fmt.Println(fmt.Sprintf("初始值%d", cordclass))
//	cordclass ^= CARDS_NOMOR
//	//去风
//	cordclass ^= CARDS_WITHOUT_WIND
//	//去白板
//	cordclass ^= CARDS_WITHOUT_BAI
//	//去发和白
//	//cordclass ^= CARDS_WITHOUT_FA | CARDS_WITHOUT_BAI
////去红中
////	cordclass ^= CARDS_WITHOUT_ZHONG
//	// //去红中 白板
//	// cordclass ^= CARDS_WITHOUT_ZHONG | CARDS_WITHOUT_BAI
//	// //去红中 发财
//	// cordclass ^= CARDS_WITHOUT_ZHONG | CARDS_WITHOUT_FA
//	// //去字牌
//	// cordclass ^= CARDS_WITHOUT_DRAGON
//	//去万
//	//  cordclass ^= CARDS_WITHOUT_WAN
//	// //添加红中
//	// var Specialcard []byte
//	// // //追加红中
//	// Specialcard = append(Specialcard, 0x35)
//	err, cards := CreateCardLibrary(cordclass, 0x37,0x37,0x38,0x31)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	// fmt.Println(fmt.Sprintf("生成牌库牌数(%d) ", len(cards)))
//	// for i := byte(0); i < byte(len(cards)); i++ {
//	// 	fmt.Print(fmt.Sprintf("(%s) ", g_CardAnother[CardToIndex(cards[i])]))
//	// }
//
//	//err, cards := CreateCards(cordclass)
//	//if err != nil {
//	//	fmt.Println(err)
//	//	return
//	//}
//	fmt.Println(fmt.Sprintf("生成牌库牌数(%d) ", len(cards)))
//	for i := byte(0); i < byte(len(cards)); i++ {
//		v, err := CardToIndex(cards[i])
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		fmt.Print(fmt.Sprintf("(%s) ", G_CardAnother[v]))
//	}
//	cardsIndex := CardsToCardIndex(cards)
//	fmt.Println(fmt.Sprintf("混牌后的牌库列表数据"))
//	Print_cards(cardsIndex[:])
//	//混乱扑克
//	num, randcards := RandCardData(cards)
//	fmt.Println(fmt.Sprintf("混牌后的数据数量（%d）", num))
//	for i := byte(0); i < byte(len(cards)); i++ {
//		if randcards[i] == 0 {
//			v, err := CardToIndex(cards[i])
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			fmt.Print(fmt.Sprintf("(%s) ", G_CardAnother[v]))
//		}
//
//	}
//	cardsIndex = CardsToCardIndex(cards)
//	fmt.Println(fmt.Sprintf("混牌后的列表数据"))
//	Print_cards(cardsIndex[:])
//}
//func Test_FindMagic(t *testing.T) {
//	cordclass := CARDS_NOMOR
//	//去风
//	cordclass ^= CARDS_WITHOUT_WIND
//	//测试边界值
//	var testcards byte = 0x34
//	card, err := FindMagicCard(cordclass, testcards)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	v, err := CardToIndex(testcards)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(fmt.Sprintf("皮子牌(%s) ", G_CardAnother[v]))
//	v, err = CardToIndex(card)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(fmt.Sprintf("赖子牌(%s) ", G_CardAnother[v]))
//
//	//检查2万
//	testcards = 0x02
//	cordclass ^= CARDS_WITHOUT_WAN
//	card, err = FindMagicCard(cordclass, testcards)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	v, err = CardToIndex(testcards)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(fmt.Sprintf("皮子牌(%s) ", G_CardAnother[v]))
//	v, err = CardToIndex(card)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(fmt.Sprintf("赖子牌(%s) ", G_CardAnother[v]))
//
//}
