package hooks

import (
	"bytes"
	"fmt"

	"github.com/open-source/game/chess.git/pkg/xlog/tools"
	"github.com/sirupsen/logrus"
)

type customHook struct {
	toDt        bool
	token       string
	errorDetail bool
	errorDeep   int
	serverShow  bool
	serverName  string
}

func (h *customHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *customHook) Fire(entry *logrus.Entry) error {
	if head, ok := entry.Data["head"]; ok {
		delete(entry.Data, "head")
		entry.Data["_head"] = head
	}
	// append wuhan eve.
	if h.serverShow {
		entry.Data["from"] = h.serverName
	}

	if entry.Level > logrus.ErrorLevel {
		return nil
	}

	// append caller file eve.
	file := entry.Data["caller"]
	if file == nil {
		file = tools.FileWithLineNum(h.errorDeep)
		if h.errorDetail {
			entry.Data["caller"] = file
		}
	}

	if entry.Level > logrus.FatalLevel {
		return nil
	}

	// send to ding talk robot.
	if !h.toDt || h.token == "" {
		return nil
	}
	var buf bytes.Buffer
	buf.WriteString(h.serverName)
	buf.WriteString(fmt.Sprintf("[%v].\n", entry.Level))
	buf.WriteString(entry.Message)
	buf.WriteString("\r\n")
	buf.WriteString(file.(string))
	return tools.CallDingRobot(h.token, buf.String(), false)
}

func NewCustomHook(dt bool, token string, errorDetail bool, errorDeep int, svrShow bool, svrName string) *customHook {
	// if not need return nil.
	if !errorDetail && !svrShow && (!dt || token == "") {
		return nil
	}
	return &customHook{
		toDt:        dt,
		token:       token,
		errorDetail: errorDetail,
		errorDeep:   errorDeep,
		serverShow:  svrShow,
		serverName:  svrName,
	}
}
