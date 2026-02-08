package ShiShou510k

import "testing"

func TestSportLogicSS510K_RandCardData(t *testing.T) {
	var logic SportLogicSS510K
	logic.m_playMode = 2
	logic.MaxPlayerCount = 4
	count, allCards := logic.CreateCards()
	logic.MaxCardCount = count / logic.MaxPlayerCount
	// 1, 14, 27,40
	_, playerCards, _, _ := logic.RandCardData(allCards, 100001, 2, []byte{1, 1, 14, 14}, 1)
	for i, cards := range playerCards {
		t.Logf("idx=%d, cards=%v", i, cards)
	}
}
