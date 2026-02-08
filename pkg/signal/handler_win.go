//+build windows

package signal

import (
	"context"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"os"
	"os/signal"
	"syscall"
)

type handler struct {
	Options
}

func Handle(ctx context.Context, opts ...Option) {
	hd := &handler{makeOptions(opts...)}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGPIPE, syscall.SIGTRAP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		select {
		case sig := <-ch:
			xlog.Logger().Warnf("got a signal = %#v", sig)
			handlerFunc := func(ctx2 context.Context) {
				xlog.Logger().Warnf("no handler func")
			}
			switch sig {
			case syscall.SIGPIPE:
				handlerFunc = hd.fnPIPE
			case syscall.SIGTRAP:
				handlerFunc = hd.fnTRAP
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM:
				handlerFunc = hd.fnINT
			}
			if handlerFunc != nil {
				handlerFunc(ctx)
			}
		case <-ctx.Done():
			xlog.Logger().Warnf("got context done")
			return
		}
	}
}
