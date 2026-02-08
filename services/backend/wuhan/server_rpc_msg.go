package wuhan

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/sirupsen/logrus"
)

//ServerMethod RPC
type ServerMethod int

//ServerMsg Rpc 入口
func (srv *ServerMethod) ServerMsg(ctx context.Context, args *static.Rpc_Args, reply *[]byte) error {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()
	if args == nil || reply == nil {
		return errors.New("nil paramters")
	}

	head, _, _, data, ok, _, _ := static.HF_DecodeMsg(args.MsgData)
	if !ok {
		return errors.New("args err")
	}
	xlog.Logger().WithFields(logrus.Fields{
		"head": head,
		"data": string(data),
	}).Infoln("【RECEIVED RPC】")
	switch head {
	}

	return nil
}

// func (self *ServerMethod) NewServerMsg(ctx context.Context, args *[]byte, reply *[]byte) error {
// 	return self.ServerMsg(cyx, &public.Rpc_Args{MsgData: *args}, reply)
// }
