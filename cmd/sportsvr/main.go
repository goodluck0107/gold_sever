package main

import (
	"context"
	"github.com/open-source/game/chess.git/pkg/engine"
	_ "github.com/open-source/game/chess.git/services/sport/gymnasium"
	"github.com/open-source/game/chess.git/services/sport/wuhan"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Up(ctx, wuhan.GetServer())
}
