package ShiShou510k

import "github.com/open-source/game/chess.git/pkg/static"

type GoodCard struct {
	CardNum  int
	MagicNum int
	NoMagic  bool
}

var GoodCards = []GoodCard{
	{
		CardNum:  5,
		MagicNum: 1,
		NoMagic:  false,
	},
	{
		CardNum:  4,
		MagicNum: 2,
		NoMagic:  false,
	},
	{
		CardNum:  6,
		MagicNum: 1,
		NoMagic:  false,
	},
	{
		CardNum:  5,
		MagicNum: 2,
		NoMagic:  false,
	},
	{
		CardNum:  4,
		MagicNum: 3,
		NoMagic:  false,
	},
	{
		CardNum:  7,
		MagicNum: 1,
		NoMagic:  true,
	},
	{
		CardNum:  6,
		MagicNum: 2,
	},
	{
		CardNum:  5,
		MagicNum: 3,
		NoMagic:  true,
	},
	{
		CardNum:  4,
		MagicNum: 4,
		NoMagic:  true,
	},
}

func GetSuperManBigCards(magics []byte) ([]byte, byte, *GoodCard) {
	var goodRes []GoodCard
	for _, v := range GoodCards {
		if v.MagicNum <= len(magics) {
			goodRes = append(goodRes, v)
		}
	}
	if len(goodRes) == 0 {
		return nil, 0, nil
	}
	gc := goodRes[static.HF_GetRandom(len(goodRes))]
	points := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	point := points[static.HF_GetRandom(len(points))]
	colors := map[byte]int{
		0: 2,
		1: 2,
		2: 2,
		3: 2,
	}
	var bigCards []byte
	for i := 0; i < gc.CardNum; i++ {
		for c, n := range colors {
			if n > 0 {
				card := c*13 + point
				bigCards = append(bigCards, card)
				n--
				colors[c] = n
				break
			}
		}
	}

	for i := 0; i < gc.MagicNum; i++ {
		rdmIdx := static.HF_GetRandom(len(magics))
		bigCards = append(bigCards, magics[rdmIdx])
		magics = append(magics[:rdmIdx], magics[rdmIdx+1:]...)
	}
	return bigCards, point, &gc
}
