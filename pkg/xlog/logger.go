package xlog

import (
	"fmt"
	"sync"

	"github.com/open-source/game/chess.git/pkg/xlog/gen"
	"github.com/open-source/game/chess.git/pkg/xlog/hooks"
	"github.com/open-source/game/chess.git/pkg/xlog/rotate"
	"github.com/open-source/game/chess.git/pkg/xlog/tools"
	"github.com/sirupsen/logrus"
)

var (
	logger              *logrus.Logger
	defaultFileLevel    = _FileLevel
	defaultConsoleLevel = _ConsoleLevel
	initOnce            sync.Once
	setLevelLock        sync.Mutex
)

func SetFileLevel(level string) {
	setLevelLock.Lock()
	defer setLevelLock.Unlock()
	defaultFileLevel = level
	if logger != nil {
		newLogger := logrus.New()
		err := hooksInit(newLogger, ServerName)
		if err != nil {
			logger.Error("SetFileLevel error:", err)
			return
		}
		logger.ReplaceHooks(newLogger.Hooks)
		newLogger = nil
	}
}

func SetConsoleLevel(level string) {
	setLevelLock.Lock()
	defer setLevelLock.Unlock()
	defaultConsoleLevel = level
	Logger().SetLevel(gen.GenLogLevel(defaultConsoleLevel))
}

// 日志指针.
// e.g..
// Logger().Debug(...interface{})
// Logger().Debugf(format string,...interface{})
// Logger().Debugln(...interface{})
func Logger() *logrus.Logger {
	if logger == nil {
		initOnce.Do(func() {
			logger = logrus.New()
			AnyQuit(createLogger(logger), "init logger error.")
		})
	}
	return logger
}

// 结构化日志输出: 官方推荐的输出方式, 打印出来的日志结构会更加清晰.
// e.g..
// WithFields(map[string]interface{}{
// 		"name":"he xu",
// 		"gender": "男",
// 		}).Info("my eve:")
// output:
// {2019/6/18 09:30:00  level=eve  msg=my eve:    name: he xu     gender=男}
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Logger().WithFields(fields)
}

func WithField(name string, value interface{}) *logrus.Entry {
	if err, ok := value.(error); ok {
		return Logger().WithError(err)
	}
	return Logger().WithField(name, value)
}

// TODO: 结构体格式化输出
// func WhitAny(obj interface{}) *logrus.Entry {
// 	switch t := obj.(type) {
// 	case fmt.Stringer:
// 		return log().WithField("obj", t.String())
// 	case error:
// 		return log().WithError(t)syslog.Logger().Infof
// 	default:
// 		obj2map, err := utils.StructToMap(obj)
// 		if err == nil {
// 			return log().WithFields(obj2map)
// 		}
// 		return log().WithFields(map[string]interface{}{
// 			"obj":   obj,
// 			"error": err,
// 		})
// 	}
// }

func WhitAny(obj interface{}) *logrus.Entry {
	return Logger().WithField("@obj", fmt.Sprintf("%+v", obj))
}

// 错误日志输出
// 如果错误为nil 则不打印；
// 如果错误不为nil 则打印错误，并附加调用者文件行号信息。
func AnyErrors(err error, args ...interface{}) {
	if err != nil {
		entry := Logger().WithError(err)
		if _ErrorDetail {
			entry = entry.WithField("caller", tools.FileWithLineNum(_ErrorDeep))
		}
		entry.Errorln(args...)
	}
}

// 程序强行退出日志输出
// 如果错误为nil 则不打印 不退出；
// 如果错误不为nil 则打印错误，并附加调用者文件行号信息，然后退出。
func AnyQuit(err error, args ...interface{}) {
	if err != nil {
		entry := Logger().WithError(err)
		if _ErrorDetail {
			entry = entry.WithField("caller", tools.FileWithLineNum(_ErrorDeep))
		}
		entry.Panic(args...)
	}
}

// 本地日志切割系统/日志文件旋转控制器
// 日志文件系统的管理及切割 是由该控制器管理
// 通过得到控制器就可以得到当前日志文件名字，路劲等信息
// 也可以主动的切割日志 e.g.. _FileControl().Rotate()
func _FileControl() *rotate.RotateLogs {
	return hooks.FileController
}

// 本地日志切割系统/日志文件旋转控制器
// 日志文件系统的管理及切割 是由该控制器管理
// 通过得到控制器就可以得到当前日志文件名字，路劲等信息
// 也可以主动的切割日志 e.g.. _FileControl().Rotate()
func _ErrorControl() *rotate.RotateLogs {
	return hooks.ErrorFileController
}

func createLogger(log *logrus.Logger) error {
	if log == nil {
		return fmt.Errorf("logger is %+v", log)
	}

	log.SetReportCaller(_CallerShow)

	log.SetFormatter(gen.GenLogFormatter(true, _ConsoleFormat, _TextTimeFormat))

	if _DeveloperModel {
		log.Infof("The Current Server Name:[%s]", ServerName)
	}

	consoleLV := gen.GenLogLevel(defaultConsoleLevel)
	if _DeveloperModel {
		log.Infof("The Current Console Log Level Is [%v].", consoleLV)
	}

	if _DeveloperModel {
		if RunInDocker {
			log.Info("Running In Docker.")
		} else {
			log.Info("Running In Development.")
		}
	}

	if err := hooksInit(log, ServerName); err != nil {
		return err
	}

	redirectAfterLoggerInit()

	if RunInDocker || consoleLV == gen.NullLevel {
		log.Info("Disable All Console Output.")
		setNull(log)
	} else {
		log.SetLevel(consoleLV)
	}
	return nil
}

func hooksInit(log *logrus.Logger, ServerName string) error {
	customHook := hooks.NewCustomHook(
		_DTalkSend,
		_DTalkToken,
		_ErrorDetail,
		_ErrorDeep,
		_ServerShow,
		ServerName,
	)
	if customHook != nil {
		log.AddHook(customHook)
	}

	fileLevel := gen.GenLogLevel(defaultFileLevel)
	// if _DeveloperModel {
	log.Warningf("The Current Log File Level Is [%s].", fileLevel)
	// }

	cutter, err := hooks.NewLogFileHook(
		ServerName,
		fileLevel,
		_RootDir,
		_DirFormat,
		_FileDatePath,
		_FileRotateTime,
		_FileMaxCount,
		_FileMaxSize,
		_FileSuffix,
		_FileFormat,
		_TextTimeFormat,
		_GenErrFile,
		_ErrDirName,
		_ErrorFileSuffix,
		_ErrFileSaveDays,
		&_EventHandler{inDocker: RunInDocker},
	)

	if err != nil {
		return err
	}
	if cutter != nil {
		log.AddHook(cutter)
	}
	return nil
}

// setnull disable console log.
func setNull(log *logrus.Logger) {
	log.SetOutput(gen.GenNilOutput())
}

func redirectAfterLoggerInit() {
	if RunInDocker {
		if c := _ErrorControl(); c != nil {
			_, _ = c.Write([]byte("【redirectAfterLoggerInit】\n"))
			redirectStderr(c.CurrentFileName())
		}
	}
}

func redirectStderr(file string) {
	Logger().Infof("redirect stderr to %s", file)
	if err := redirect(file); err != nil {
		Logger().Errorln("redirect stderr error:", err)
	}
}
