package xlog

import (
	"github.com/open-source/game/chess.git/pkg/xlog/gen"
	"github.com/open-source/game/chess.git/pkg/xlog/hooks"
	"github.com/sirupsen/logrus"
	"sync"
)

type ElkLogger struct {
	logger *logrus.Logger //! 日志组件

	elkIndex string //! 日志索引
}

var elkLoggerMap = make(map[string]*ElkLogger)
var elkLoggerMapLock sync.RWMutex

func ELogger(elkIndex string) *logrus.Logger {
	elkLoggerMapLock.RLock()
	elkLogger, ok := elkLoggerMap[elkIndex]
	elkLoggerMapLock.RUnlock()
	if !ok {
		elkLogger = new(ElkLogger)
		elkLogger.elkIndex = elkIndex
		elkLogger.InitLogger()
		elkLoggerMapLock.Lock()
		elkLoggerMap[elkIndex] = elkLogger
		elkLoggerMapLock.Unlock()
	}
	return elkLogger.logger
}

func (my *ElkLogger) InitLogger() {
	my.logger = logrus.New()
	my.logger.SetFormatter(gen.GenLogFormatter(false, _TableLogFormat, _TextTimeFormat))
	if RunInDocker {
		my.logger.SetOutput(gen.GenNilOutput())
	} else {
		my.logger.SetLevel(logrus.InfoLevel)
	}
	my.HooksInit()
}

func (my *ElkLogger) HooksInit() error {
	customHook := hooks.NewCustomHook(
		_DTalkSend,
		_DTalkToken,
		_ErrorDetail,
		_ErrorDeep,
		_ServerShow,
		my.elkIndex,
	)
	if customHook != nil {
		my.logger.AddHook(customHook)
	}

	fileLevel := logrus.InfoLevel
	// 	my.logger.Infof("The Current Log File Level Is [%v].", fileLevel)

	cutter, err := hooks.NewElkLogFileHook(
		my.elkIndex,
		fileLevel,
		_ElkRootDir+"/"+my.elkIndex,
		_DirFormat,
		_ElkFileDatePath,
		_ElkFileRotateTime,
		_FileMaxCount,
		_FileMaxSize,
		_FileSuffix,
		_FileFormat,
		_TextTimeFormat,
	)

	if err != nil {
		return err
	}
	if cutter != nil {
		my.logger.AddHook(cutter)
	}
	return nil
}
