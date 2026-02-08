package main

import (
	"context"
	"github.com/open-source/game/chess.git/pkg/engine"
	"github.com/open-source/game/chess.git/services/user"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Up(ctx, user.GetServer())
}
