package engine

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"io"
	"net/http"
	"runtime"
)

// 输出栈信息
func stackInfo(w io.Writer) error {
	buf := make([]byte, 1638400)

	buf = buf[:runtime.Stack(buf, true)]

	_, err := w.Write(buf)

	return err
}

func GetStackInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                               //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") //header的类型

	xlog.Logger().Infoln("【Get Stack】")

	if err := stackInfo(w); err != nil {
		xlog.Logger().Errorln("Write Stack error:", err)
	}
}
