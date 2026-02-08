package static

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/sirupsen/logrus"
	rdebug "runtime/debug"
)

func NO_ERROR(f func()) {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().WithFields(logrus.Fields{
				xlog.PanicMark: "NO_ERROR--->ERROR",
			}).Errorln(x, string(rdebug.Stack()))
		}
	}()
	f()
}

//可以一直循环forevery
func GOFOR(f func()) {
	go func() {
		for {
			func() {
				defer func() {
					x := recover()
					if x != nil {
						xlog.Logger().WithFields(logrus.Fields{
							xlog.PanicMark: "GOFOR--->ERROR",
						}).Errorln(x, string(rdebug.Stack()))
					}
				}()
				f()
			}()
		}
	}()
}

func GOFOR_CONTINUE(f func() bool) {
	go func() {
		for {
			ret := func() bool {
				defer func() {
					x := recover()
					if x != nil {
						xlog.Logger().WithFields(logrus.Fields{
							xlog.PanicMark: "GOFOR--->ERROR",
						}).Errorln(x, string(rdebug.Stack()))
					}
				}()
				return f()
			}()
			if !ret {
				return
			}
		}
	}()
}
