package xlog

import (
	"github.com/open-source/game/chess.git/pkg/xlog/tools"
)

// 开发者模式
// 该模式下 go panic信息直接指向控制台，日志文件可能会漏掉一些来自golang库底层的panic信息。
// 关闭该模式会导致：所有的panic信息以及来自于该Log包的任何数据都会重定向到日志文件，不会打印在控制台。
// Docker容器中要关掉该模式，置为False；
// 开发人员本地调试时打开该模式，置为True。
// const _DeveloperModel = true
const _DeveloperModel = false

var (
	RunInDocker = tools.RunInDocker()
	ServerName  = tools.ExeName()
)

// log out level and ding talk. future migrations to configuration files.
const (
	// 控制台打印等级
	_ConsoleLevel = "warn"
	// 日志文件打印等级
	_FileLevel = "warn"
	// GORM数据库sql打印等级
	_GormConsoleLevel = "warn"
	// GORM数据库日志文件sql打印等级
	_GormFileLevel = "warn"
	// 钉钉机器人推送开关
	_DTalkSend = false
	// 钉钉机器人token
	_DTalkToken = ""
)

// gorm logger
const (
	// gorm level tag
	_GORMLOGLEVELSQL = "sql"
	// gorm log prefix
	_GORMLOGPREFIX = "\r\n"
	// gorm log tag
	_GORMLOGPREFIXTITLE = "GORM"
	// SQL speed standard
	_GORMLOGSTANDARDSPEEDFASTMS = 50
	_GORMLOGSTANDARDSPEEDSLOWMS = 100
	// SQL speed divide
	_GORMLOGPREFIXSPEEDFAST  = "FAST"
	_GORMLOGPREFIXSPEEDMEZZO = "MEZO"
	_GORMLOGPREFIXSPEEDSLOW  = "SLOW"
	// SQL level
	_GORMLOGPREFIXLEVELINFO = "INFO"
	_GORMLOGPREFIXLEVELERRO = "ERRO"

	_GORMLOGRootDir         = "./log/db"
	_GORMLOGFileSuffix      = ".log"
	_GORMLOGDirFormat       = "年月日"
	_GORMLOGErrDirName      = ""
	_GORMLOGFileRotateTime  = 60 * 60
	_GORMLOGFileMaxCount    = 168
	_GORMLOGFileMaxSize     = 1024 * 1024 * 500
	_GORMLOGErrFileSaveDays = 7
)

// log path
const (
	// 日志文件根目录
	_RootDir = "./log"
	// 控制台日志格式: 文本text/json
	_ConsoleFormat = "text"
	// 日志文件打印日志格式
	_FileFormat = "json"
	// 桌子日志文件格式
	_TableLogFormat = "json"
	// text文本日志格式下时间格式化样式
	_TextTimeFormat = "2006/01/02 15:04:05"
	// 日志文件夹格式化格式
	_DirFormat = "年月日"
	// 日志文件后缀名
	_FileSuffix = ".log"

	// 日志文件根目录
	_ElkRootDir = "./elklog"
)

// error
const (
	// 是否为错误以上级别日志再生成专门的错误日志文件
	_GenErrFile = true
	// 错误日志文件所在的目录名 e.g.. "error" => ./log/error/
	_ErrDirName = ""
	// 错误日志文件每天生成一份新的，这里设置保存的最大个数
	_ErrFileSaveDays = 3
	// 错误日志文件后缀名
	_ErrorFileSuffix = ".sys"

	//elk不需要错误日志
	_ElkGenErrFile = false
)

// log file
const (
	// 日志文件切分间隔 单位:秒
	_FileRotateTime = 60 * 60
	// _FileRotateTime = 10
	// 日志文件最大数量
	_FileMaxCount = 168
	// 日志文件最大size 单位: byte
	_FileMaxSize = 1024 * 1024 * 500
	// 日志文件是否需要创建日期目录
	_FileDatePath = true

	// 日志文件切分间隔 单位:秒
	_ElkFileRotateTime = 60 * 60 * 24
	// 日志文件是否需要创建日期目录
	_ElkFileDatePath = false
)

// custom
const (
	// windows系统下panic信息是否重定向
	_PanicRedirectWin = true
	// linux系统下panic信息是否重定向
	_PanicRedirectUnix = true
	// 是否显示日志调用者文件名及行号
	_CallerShow = true
	// 日志是否显示进程名字信息
	_ServerShow = true
	// 错误及以上级别日志显示调用者详细信息
	_ErrorDetail = true
	// 错误追踪深度
	_ErrorDeep = 2
)

const (
	ElkGpsInfo      = "GpsInfo"
	ElkProtocolInfo = "ProtocolInfo"
)
