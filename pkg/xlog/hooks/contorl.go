package hooks

import (
	"fmt"
	"io"
	"time"

	"github.com/open-source/game/chess.git/pkg/xlog/gen"
	"github.com/open-source/game/chess.git/pkg/xlog/rotate"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// 日志文件切割管理者
var (
	FileController, ErrorFileController *rotate.RotateLogs
)

// NewLogFileHook 生成一个日志文件切割器
func NewLogFileHook(
	serverName string,
	fileLevel logrus.Level,
	rootDir string,
	formatDir string,
	needDatePath bool,
	rotateTime int64,
	maxCount uint,
	maxSize int64,
	suffix string,
	formatter string,
	timeFormat string,
	errGen bool,
	errDirName string,
	errFileSuffix string,
	errSaveDays uint,
	eventHandler rotate.Handler,
) (logrus.Hook, error) {
	if fileLevel == gen.NullLevel {
		return nil, nil
	}

	writer, err := rotate.New(
		gen.GenLogFileName(serverName, rootDir, formatDir, needDatePath, rotateTime),
		// baseLogPaht+".%Y-%m-%d-%H-%M-%S",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		// rotate.WithLinkName(baseLogPaht),
		// WithMaxAge和WithRotationCount二者只能设置一个，
		// rotate.WithMaxAge(maxAge),
		rotate.WithRotationCount(maxCount),
		// WithRotationTime设置日志分割的时间，这里设置为一小时分割一次
		rotate.WithRotationTime(time.Second*time.Duration(rotateTime)),
		rotate.WithLocation(time.Local),
		// 日志文件变动事件
		// rotate.WithHandler(eventHandler),
		rotate.WithMaxSize(maxSize),
		rotate.WithSuffix(suffix),
	)

	if err != nil {
		return nil, fmt.Errorf("LogCutter config local file system for logger error:%v", err)
	}

	FileController = writer

	if errGen {
		ErrorFileController, err = rotate.New(
			gen.GenErrorFileName(serverName, rootDir, errDirName, formatDir),
			rotate.WithRotationCount(errSaveDays),
			rotate.WithRotationTime(time.Hour*24),
			rotate.WithLocation(time.Local),
			rotate.WithSuffix(errFileSuffix),
			rotate.WithHandler(eventHandler),
		)
		if err != nil {
			return nil, fmt.Errorf("LogCutter local error file system for logger error:%v", err)
		}
	}

	wmap := make(map[logrus.Level]io.Writer)
	// 如果普通日志的开启等级为error 且设置生成新的error日志
	// 此时普通日志等于没用 需要处理
	if fileLevel <= logrus.ErrorLevel && errGen {
		// 释放掉原普通日志文件控制器
		FileController = nil
		for _, level := range logrus.AllLevels {
			if ErrorFileController != nil && level <= logrus.ErrorLevel {
				wmap[level] = ErrorFileController
				continue
			}
		}
	} else {
		for _, level := range logrus.AllLevels {
			if level > fileLevel {
				continue
			}
			if errGen && ErrorFileController != nil && level <= logrus.ErrorLevel {
				wmap[level] = io.MultiWriter(writer, ErrorFileController)
			} else {
				wmap[level] = writer
			}
		}
	}
	return lfshook.NewHook(lfshook.WriterMap(wmap), gen.GenLogFormatter(false, formatter, timeFormat)), nil
}

// NewLogFileHook 生成一个数据库日志文件切割器
func NewDataBaseLogFileHook(
	fileLevel logrus.Level,
	rootDir string,
	formatDir string,
	rotateTime int64,
	maxCount uint,
	maxSize int64,
	errDirName string,
	suffix string,
	errSaveDays uint,
	dbFmt logrus.Formatter,
) (logrus.Hook, error) {
	if fileLevel == gen.NullLevel || dbFmt == nil {
		return nil, nil
	}

	writer, err := rotate.New(
		gen.GenLogFileName("gorm", rootDir, formatDir, true, rotateTime),
		rotate.WithRotationCount(maxCount),
		rotate.WithRotationTime(time.Second*time.Duration(rotateTime)),
		rotate.WithLocation(time.Local),
		rotate.WithMaxSize(maxSize),
		rotate.WithSuffix(suffix),
	)

	if err != nil {
		return nil, fmt.Errorf("LogCutter config local file system for logger error:%v", err)
	}

	dbErrorFile, err := rotate.New(
		gen.GenErrorFileName("gorm", rootDir, errDirName, formatDir),
		rotate.WithRotationCount(errSaveDays),
		rotate.WithRotationTime(time.Hour*24),
		rotate.WithLocation(time.Local),
		rotate.WithSuffix(suffix),
	)
	if err != nil {
		return nil, fmt.Errorf("LogCutter local error file system for logger error:%v", err)
	}

	wmap := make(map[logrus.Level]io.Writer)
	if fileLevel <= logrus.ErrorLevel {
		// 释放掉原普通日志文件控制器
		writer = nil
		for _, level := range logrus.AllLevels {
			if dbErrorFile != nil && level <= logrus.ErrorLevel {
				wmap[level] = dbErrorFile
				continue
			}
		}
	} else {
		for _, level := range logrus.AllLevels {
			if level > fileLevel {
				continue
			}
			if dbErrorFile != nil && level <= logrus.ErrorLevel {
				wmap[level] = io.MultiWriter(writer, dbErrorFile)
				continue
			}
			wmap[level] = writer
		}
	}
	return lfshook.NewHook(lfshook.WriterMap(wmap), dbFmt), nil
}

func NewElkLogFileHook(
	serverName string,
	fileLevel logrus.Level,
	rootDir string,
	formatDir string,
	needDatePath bool,
	rotateTime int64,
	maxCount uint,
	maxSize int64,
	suffix string,
	formatter string,
	timeFormat string,
) (logrus.Hook, error) {
	if fileLevel == gen.NullLevel {
		return nil, nil
	}

	writer, err := rotate.New(
		gen.GenLogFileName(serverName, rootDir, formatDir, needDatePath, rotateTime),
		// baseLogPaht+".%Y-%m-%d-%H-%M-%S",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		// rotate.WithLinkName(baseLogPaht),
		// WithMaxAge和WithRotationCount二者只能设置一个，
		// rotate.WithMaxAge(maxAge),
		rotate.WithRotationCount(maxCount),
		// WithRotationTime设置日志分割的时间，这里设置为一小时分割一次
		rotate.WithRotationTime(time.Second*time.Duration(rotateTime)),
		rotate.WithLocation(time.Local),
		// 日志文件变动事件
		// rotate.WithHandler(eventHandler),
		rotate.WithMaxSize(maxSize),
		rotate.WithSuffix(suffix),
	)

	if err != nil {
		return nil, fmt.Errorf("LogCutter config local file system for logger error:%v", err)
	}

	wmap := make(map[logrus.Level]io.Writer)

	for _, level := range logrus.AllLevels {
		if level > fileLevel {
			continue
		}

		wmap[level] = writer

	}

	return lfshook.NewHook(lfshook.WriterMap(wmap), gen.GenLogFormatter(false, formatter, timeFormat)), nil
}
