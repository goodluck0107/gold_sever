package static

// type TableLoger struct {
// 	loger      *log.Logger //! 日志组件
// 	logPath    string      //! 当前日志文件路径
// 	tableId    int         //! 当前桌子日志桌子Id
// 	createTime string      //! 当前桌子创建时间
// 	tableGame  string      //! 当期桌子游戏名称
// 	cmd_output bool        //！是否在控制台实时log
// }
//
// func CreateTableLog(tableId int, tableGame string) *TableLoger {
// 	tableloger := new(TableLoger)
// 	tableloger.tableId = tableId
// 	tableloger.tableGame = tableGame
// 	tableloger.createTime = fmt.Sprintf("%02d%02d", time.Now().Hour(), time.Now().Minute())
// 	tableloger.logPath = tableloger.GetWriteDirPath()
// 	os.MkdirAll(tableloger.logPath, os.ModePerm)
//
// 	return tableloger
// }
//
// func (self *TableLoger) GetWriteDirPath() string {
// 	writePath := "./log" + "/" + self.tableGame + "/" + time.Now().Format("20060102")
// 	return writePath
// }
//
// func (self *TableLoger) GetWriteFilePath() (string, string) {
// 	writeDirPath := self.GetWriteDirPath()
// 	writeFilePath := writeDirPath + "/" + fmt.Sprintf("%d_%s.log", self.tableId, self.createTime)
//
// 	return writeDirPath, writeFilePath
// }
//
// func (self *TableLoger) Output(v ...interface{}) {
// 	self.changDay()      ///跨天则生成新文件
// 	if self.cmd_output { // 有打开控制台log开关
// 		log.Println(v)
// 	}
// 	self.loger.Output(2, fmt.Sprintf("\n%s\n", fmt.Sprintln(v...)))
// }
//
// func (self *TableLoger) changDay() { ///跨天则生成新文件
// 	logDirPath, logFilePath := self.GetWriteFilePath()
// 	_, err1 := os.Stat(logDirPath)
//
// 	// 文件夹不存在
// 	if os.IsNotExist(err1) {
// 		os.MkdirAll(logDirPath, os.ModePerm)
// 	}
//
// 	// 判断日志文件是否存在
// 	_, err2 := os.Stat(logFilePath)
//
// 	if os.IsNotExist(err2) || self.loger == nil {
// 		logfile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
// 		if err != nil {
// 			log.Println("open log file fail!", err.Error())
// 		}
// 		self.loger = log.New(logfile, "", log.Ltime|log.Lmicroseconds|log.Lshortfile)
// 	}
// }
//
// func (self *TableLoger) SetCMD(flag bool) {
// 	self.cmd_output = flag
// }
