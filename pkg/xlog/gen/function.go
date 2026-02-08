package gen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func CallInfoRedefine(frame *runtime.Frame) (function string, file string) {
	if frame == nil {
		return "unknown func()", "unknown file"
	}
	var buf bytes.Buffer
	// 屏蔽掉方法名
	// buf.WriteString(frame.Function)
	// buf.WriteString("()")
	// function = buf.String()
	// buf.Reset()
	fileInfo := strings.Split(frame.File, "/")

	if fl := len(fileInfo); fl <= 0 {
		file = frame.File
	} else {
		buf.WriteString(fileInfo[fl-1])
		buf.WriteString(fmt.Sprintf(":%d", frame.Line))
		file = buf.String()
	}
	return
}

func GenLogFileName(serverName string, rootDir string, formatDir string, needDatePath bool, rotateTime int64) string {
	var buf bytes.Buffer
	buf.WriteString(parseRootDir(serverName, rootDir))
	if needDatePath {
		buf.WriteString(parseDate2En(formatDir))
	}
	path := buf.String()
	buf.Reset()
	buf.WriteString(parseFileFormat(serverName, rotateTime))
	name := buf.String()
	return filepath.Join(path, name)
}

// func genPanicFileName() string {
// 	var buf bytes.Buffer
// 	buf.WriteString(parseRootDir(yaml.Get().ConLog.FileRootDir))
// 	buf.WriteString("/")
// 	buf.WriteString(yaml.Get().ConLog.PanicOutDirName)
// 	path := buf.String()
// 	buf.Reset()
// 	buf.WriteString(process.CurProcess().Name())
// 	buf.WriteString(".PANIC")
// 	name := buf.String()
// 	return filepath.Join(path, name)
// }

func GenErrorFileName(serverName, rootDir, errDirName, formatDir string) string {
	var buf bytes.Buffer
	buf.WriteString(parseRootDir(serverName, rootDir))
	buf.WriteString(parseDate2En(formatDir))
	buf.WriteString("/")
	buf.WriteString(errDirName)
	path := buf.String()
	buf.Reset()
	buf.WriteString(serverName)
	buf.WriteString("_error")
	name := buf.String()
	return filepath.Join(path, name)
}

// func cleanupFile(file string) {
// 	if !yaml.Get().ConLog.PanicOutFileClear {
// 		return
// 	}
// 	_ = ioutil.WriteFile(file, []byte{}, 0644)
// }

func GenLogLevel(lv string) logrus.Level {
	llv := strings.ToLower(lv)
	switch {
	case strings.Index(llv, "p") == 0:
		return logrus.PanicLevel
	case strings.Index(llv, "f") == 0:
		return logrus.FatalLevel
	case strings.Index(llv, "e") == 0:
		return logrus.ErrorLevel
	case strings.Index(llv, "w") == 0:
		return logrus.WarnLevel
	case strings.Index(llv, "i") == 0:
		return logrus.InfoLevel
	case strings.Index(llv, "d") == 0:
		return logrus.DebugLevel
	case strings.Index(llv, "t") == 0:
		return logrus.TraceLevel
	}
	return NullLevel
}

func GenLogFormatter(console bool, format, textTimeFormat string) logrus.Formatter {
	if strings.Index(strings.ToLower(format), "j") == 0 {
		return &logrus.JSONFormatter{
			CallerPrettyfier: CallInfoRedefine,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "#time",
				logrus.FieldKeyMsg:   "_msg",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyFile:  "$file",
			},
		}
	}
	return &logrus.TextFormatter{
		DisableColors:    !console,
		ForceColors:      true,
		TimestampFormat:  textTimeFormat,
		FullTimestamp:    true,
		CallerPrettyfier: CallInfoRedefine,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "#time",
			logrus.FieldKeyMsg:   "_msg",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyFile:  "$file",
		},
	}
}

func GenNilOutput() io.Writer {
	src, _ := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	return bufio.NewWriter(src)
}
