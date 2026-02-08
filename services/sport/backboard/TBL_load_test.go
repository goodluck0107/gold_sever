package backboard

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"testing"
	"time"
)

func BenchmarkMjDataTable_LoadTable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctime := time.Now()
		mgr := new(mjDataTable)
		mgr.LoadTable()
		speed := time.Since(ctime)
		xlog.Logger().Printf("耗时：[%0.2fms]", float64(speed.Nanoseconds())/1e4/100.00)
	}
}
