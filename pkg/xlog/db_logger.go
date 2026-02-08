package xlog

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/open-source/game/chess.git/pkg/xlog/gen"
	"github.com/open-source/game/chess.git/pkg/xlog/hooks"
	"github.com/open-source/game/chess.git/pkg/xlog/tools"
	"github.com/sirupsen/logrus"
)

var dbloggersingleton *dbLogger = nil

// 数据库sql日志
func DBLogger() *dbLogger {
	if dbloggersingleton == nil {
		dbloggersingleton = new(dbLogger)
		dbloggersingleton.init()
	}
	return dbloggersingleton
}

type dbLogger struct {
	logger *logrus.Logger
}

func (dbl *dbLogger) init() {
	dbl.logger = logrus.New()
	dbFormatter := new(TextFormatter)
	dbl.logger.SetFormatter(dbFormatter)
	dbl.logger.SetReportCaller(false)
	lv := gen.GenLogLevel(_GormConsoleLevel)
	dbl.logger.SetLevel(lv)
	dbFileHook, err := hooks.NewDataBaseLogFileHook(
		gen.GenLogLevel(_GormFileLevel),
		_GORMLOGRootDir,
		_GORMLOGDirFormat,
		_GORMLOGFileRotateTime,
		_GORMLOGFileMaxCount,
		_GORMLOGFileMaxSize,
		_GORMLOGErrDirName,
		_GORMLOGFileSuffix,
		_GORMLOGErrFileSaveDays,
		dbFormatter,
	)

	if err != nil {
		Logger().Panic("add gorm log file error", err)
		return
	}

	if dbFileHook != nil {
		dbl.logger.AddHook(dbFileHook)
	}

	if lv == gen.NullLevel || RunInDocker {
		dbl.logger.SetOutput(gen.GenNilOutput())
	}
}

// RecordError接口供程序内部调用
// 会将内容打印至控制台及日志文件
// Docker容器内自动禁用控制台输出
func (dbl *dbLogger) RecordError(values ...interface{}) {
	dbl.logger.Errorln(recordErrorLogFormatter(values...)...)
}

// Print接口供GORM官方库调用，我们自己程序内部不要调用该方法，会导致日志内容不兼容
func (dbl *dbLogger) Print(values ...interface{}) {
	lv, msg := gormLogFormatter(values...)
	if lv == logrus.ErrorLevel {
		dbl.logger.Errorln(msg...)
	} else {
		dbl.logger.Infoln(msg...)
	}
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func isSQL(level interface{}) bool {
	return level == _GORMLOGLEVELSQL
}

func gormLogPrefix(sql bool, speed float64) string {
	var buf bytes.Buffer
	buf.WriteString(_GORMLOGPREFIX)
	buf.WriteString("[")
	buf.WriteString(_GORMLOGPREFIXTITLE)
	buf.WriteString("|")
	if sql {
		buf.WriteString(genSpeedRating(speed))
	} else {
		buf.WriteString(_GORMLOGPREFIXLEVELERRO)
	}
	buf.WriteString("]")
	return buf.String()
}

func recordErrorPrefix() string {
	var buf bytes.Buffer
	buf.WriteString(_GORMLOGPREFIX)
	buf.WriteString("[USER|ERRO]")
	return buf.String()
}

func genSpeedRating(speed float64) string {
	if speed < _GORMLOGSTANDARDSPEEDFASTMS {
		return _GORMLOGPREFIXSPEEDFAST
	} else if speed < _GORMLOGSTANDARDSPEEDSLOWMS {
		return _GORMLOGPREFIXSPEEDMEZZO
	} else {
		return _GORMLOGPREFIXSPEEDSLOW
	}
}

var gormLogFormatter = func(values ...interface{}) (lv logrus.Level, messages []interface{}) {
	if len(values) > 1 {
		var (
			sql             string
			formattedValues []string
			level           = values[0]
			currentTime     = "\n" + time.Now().Format("2006-01-02 15:04:05")
			source          = fmt.Sprintf("\t%v From: %s", values[1], ServerName)
		)

		sqlFlag := isSQL(level)

		var speed float64
		if sqlFlag {
			lv = logrus.InfoLevel
			speed = float64(values[2].(time.Duration).Nanoseconds()/1e4) / 100.0
		} else {
			lv = logrus.ErrorLevel
		}

		messages = []interface{}{gormLogPrefix(sqlFlag, speed), source, currentTime}

		if sqlFlag {
			// duration
			messages = append(messages, fmt.Sprintf(" [%.2fms] ", speed))
			// sql

			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			}

			// differentiate between $n placeholders or else treat like ?
			if tools.NumericPlaceHolderRegexp.MatchString(values[3].(string)) {
				sql = values[3].(string)
				for index, value := range formattedValues {
					placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
					sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
				}
			} else {
				formattedValuesLength := len(formattedValues)
				for index, value := range tools.SqlRegexp.Split(values[3].(string), -1) {
					sql += value
					if index < formattedValuesLength {
						sql += formattedValues[index]
					}
				}
			}

			messages = append(messages, sql)
			messages = append(messages, fmt.Sprintf("\n[%v]", strconv.FormatInt(values[5].(int64), 10)+" rows affected or returned "))
		} else {
			messages = append(messages, values[2:]...)
		}
		messages = append(messages, _GORMLOGPREFIX)
	}
	return
}

var recordErrorLogFormatter = func(values ...interface{}) (messages []interface{}) {
	if len(values) > 0 {
		var (
			source      = fmt.Sprintf("\t%s From: %s", tools.FileWithLineNum(1), ServerName)
			currentTime = "\n" + time.Now().Format("2006-01-02 15:04:05")
		)
		messages = []interface{}{recordErrorPrefix(), source, currentTime}
		messages = append(messages, values...)
		messages = append(messages, _GORMLOGPREFIX)
	}
	return
}
