package main

import (
	"context"
	"github.com/open-source/game/chess.git/pkg/engine"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/services/backend"
	"github.com/open-source/game/chess.git/services/backend/wuhan"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Up(ctx, wuhan.GetServer(), func(is engine.IServer) {
		worker := backend.InitRouter()
		worker.SetEncode(wuhan.GetServer().Con.Encode)
		worker.SetEncodePhpKey(wuhan.GetServer().Con.EncodePhpKey)
		router.CreateHTTP("/adminservice", worker)
	})
}
