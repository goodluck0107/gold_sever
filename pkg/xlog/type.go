package xlog

import (
	"fmt"
	"time"
)

const (
	ErrorKind = "errorType"
	ErrorAt   = "errorAt"
	ErrorFrom = "errorFrom"
)

const (
	PanicMark = "panicMark"
	TraceMark = "trace"
	FromMark  = "from"
	panicFlag = "panicFlag"
)

const (
	GameName  = "gameName"
	GameIndex = "gameIndex"
	GameTable = "tableId"
)

const (
	TimeFormatTemplate = "2006-01-02 15:04:05"
)

const (
	DefaultMaxDay  = 7
	DefaultMaxDir  = 1
	DefaultLogPath = "./log"
)

const (
	ErrPathOnMulti  = "%s/%s/%s/%s_err.sys"
	ErrPathNotMulti = "%s/%s/%s_err.sys"

	PathOnMulti  = "%s/%s/%s/%s_%s_%s.log"
	PathNotMulti = "%s/%s/%s_%s_%s.log"

	LogDirName    = "%m-%d-%Y"
	LogFileName   = "%H"
	LogFileSuffix = "01"
)

const (
	OneHour   = time.Hour * 1
	OneDay    = OneHour * 24
	OneWeek   = OneDay * 7
	HalfMonth = OneDay * 15
)

type InitLoggerError string

func (self InitLoggerError) Error() string {
	return self.String()
}

func (self InitLoggerError) String() string {
	return fmt.Sprintf("Init Logger Error:%s", string(self))
}
