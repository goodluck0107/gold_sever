package xlog

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/open-source/game/chess.git/pkg/xlog/gen"
	"github.com/open-source/game/chess.git/pkg/xlog/hooks"
	"github.com/sirupsen/logrus"
)

type TableLoger struct {
	logger     *logrus.Logger //! 日志组件
	logPath    string         //! 当前日志文件路径
	tableId    int            //! 当前桌子日志桌子Id
	createTime string         //! 当前桌子创建时间
	tableGame  string         //! 当期桌子游戏名称
	gameIndex  string         //! 所在服务器编号
	gameNum    string         //! 当前桌子的唯一标识
	hid        int            //! 当前包厢六位编码
	floor      int            //! 当前桌子包厢楼层索引
	dHid       int64          //! 当前桌子包厢数据库Id
	dFid       int64          //! 当前桌子包厢楼层数据库Id

	line int //! 当前桌子日志行号
}

func CreateTableLog(gameIndex string, tableId int, tableGame string, gameNum string, hid int, floor int, dHid int64, dFid int64) *TableLoger {
	tableloger := new(TableLoger)
	tableloger.tableId = tableId
	tableloger.tableGame = tableGame
	tableloger.gameIndex = gameIndex
	tableloger.gameNum = gameNum
	tableloger.hid = hid
	tableloger.floor = floor
	tableloger.dHid = dHid
	tableloger.dFid = dFid

	tableloger.createTime = fmt.Sprintf("%02d%02d", time.Now().Hour(), time.Now().Minute())
	tableloger.logPath = tableloger.GetWriteDirPath()
	os.MkdirAll(tableloger.logPath, os.ModePerm)
	tableloger.initLogger()
	return tableloger
}

func (self *TableLoger) initLogger() {
	self.line = 0
	self.logger = logrus.New()
	self.logger.SetFormatter(gen.GenLogFormatter(false, _TableLogFormat, _TextTimeFormat))
	self.logger.SetLevel(logrus.InfoLevel)
	self.logger.AddHook(hooks.NewTableHook(self.tableId, self.tableGame, self.gameIndex, self.gameNum, self.hid, self.floor, self.dHid, self.dFid))
	// self.SetNull()
	self.findFile()
	self.removeOverdueFile()
}

func (self *TableLoger) SetExtraHook(uid int64, round int) {
	self.logger.AddHook(hooks.NewTableExtraHook(uid, round, self.line))
}

func (self *TableLoger) SetNull() {
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		self.logger.WithFields(logrus.Fields{
			"errinfo:": err,
		}).Errorln("table setnull error")
	}
	writer := bufio.NewWriter(src)
	self.setOutput(writer)
}

func (self *TableLoger) setOutput(out io.Writer) {
	self.logger.SetOutput(out)
}

func (self *TableLoger) GetWriteDirPath() string {
	writePath := "./log" + "/" + self.tableGame + "/" + time.Now().Format("20060102")
	return writePath
}

func (self *TableLoger) GetWriteFilePath() (string, string) {
	writeDirPath := self.GetWriteDirPath()
	writeFilePath := writeDirPath + "/" + fmt.Sprintf("%d_%s.log", self.tableId, self.createTime)

	return writeDirPath, writeFilePath
}

func (self *TableLoger) Output(playerName string, str string) {
	self.logger.Infof("玩家%s :  %s   ", playerName, str)

	self.line++
}

// 找到输出的文件
func (self *TableLoger) findFile() {
	logDirPath, logFilePath := self.GetWriteFilePath()
	_, err1 := os.Stat(logDirPath)

	// 文件夹不存在
	if os.IsNotExist(err1) {
		os.MkdirAll(logDirPath, os.ModePerm)
	}
	// 判断日志文件是否存在
	_, err := os.Stat(logFilePath)

	if os.IsNotExist(err) || self.logger == nil {
		logfile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			Logger().WithFields(logrus.Fields{
				"errinfo": err,
			}).Errorln("table log findFile output err")
		}
		self.setOutput(logfile)
	}
}

// 删除过期的文件
func (self *TableLoger) removeOverdueFile() {
	logDirs, err := ioutil.ReadDir(fmt.Sprintf("./log/%s", self.tableGame))
	if err != nil {
		Logger().WithField("errinfo:", err).Errorln("find log dir error when find at removeOverdueFile")
	}
	for _, dir := range logDirs {
		if dir.ModTime().Unix()-time.Now().Unix() > 60*60*24*30 {
			if err := os.Remove(fmt.Sprintf("./log/%s/%s", self.tableGame, dir.Name())); err != nil {
				Logger().WithField("errinfo:", err).Errorln("find log dir error when removeOverdueFile")
			}
		}
	}
}
