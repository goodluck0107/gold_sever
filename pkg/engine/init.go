package engine

import (
	"context"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/signal"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"golang.org/x/net/websocket"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	startupStatus int32
	stopOnce      sync.Once
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/stack", GetStackInfo)
}

func Up(ctx context.Context, svr IServer, startHooks ...func(is IServer)) {
	ok := atomic.CompareAndSwapInt32(&startupStatus, 0, 1)
	if !ok {
		xlog.Logger().Errorf("%s wuhan already started", svr.Name())
		return
	}
	newCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := svr.LoadServerConfig(newCtx)
	if err != nil {
		xlog.Logger().Error("load config error:", err)
		return
	}
	svr.SetLoggerLevel()
	go signal.Handle(
		newCtx, signal.WithTRAP(func(ctx context.Context) {
			err := svr.LoadServerConfig(ctx)
			if err != nil {
				xlog.Logger().Error(err)
			}
			err = svr.LoadConfig(ctx)
			if err != nil {
				xlog.Logger().Error(err)
			}
		}), signal.WithINT(func(ctx context.Context) {
			xlog.Logger().Warn("get INT")
			err := StopServer(ctx, svr)
			if err != nil {
				xlog.Logger().Fatal("stop wuhan error:", err)
			}
			xlog.Logger().Warn("cancel ...")
			cancel()
		}),
	)
	err = svr.Start(newCtx)
	if err != nil {
		xlog.Logger().Panic(err)
	}
	defer StopServer(newCtx, svr)
	for _, fn := range startHooks {
		fn(svr)
	}
	svr.RegisterRouter(newCtx)
	go svr.RegisterRPC(newCtx)
	if ws, ok := svr.(interface {
		GetConnectHandler() websocket.Handler
	}); ok {
		xlog.Logger().Warn("handle websocket")
		http.Handle("/", ws.GetConnectHandler())
	}
	go svr.Run(newCtx)
	xlog.Logger().Infof("%s start at :%s", svr.Name(), time.Now().Format("2006/01/02 15:04:05"))
	_, port := svr.Host()
	errCh := make(chan error, 1)
	go func(ch chan<- error) {
		ch <- http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	}(errCh)
	select {
	case listenErr := <-errCh:
		xlog.Logger().Error("ListenAndServe err:", listenErr)
	case <-newCtx.Done():
		xlog.Logger().Warnf("%s server shutdown", svr.Name())
	}
}

func StopServer(ctx context.Context, svr IServer) (err error) {
	stopOnce.Do(func() {
		xlog.Logger().Error("stop server...")
		err = svr.Stop(ctx)
		if err != nil {
			xlog.Logger().Error(err)
		}
		xlog.Logger().Error("stop server done")
	})
	return err
}
