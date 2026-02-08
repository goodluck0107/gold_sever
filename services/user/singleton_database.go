//! 数据库底层
package user

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	"github.com/open-source/game/chess.git/pkg/dao"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

////////////////////////////////////////////////////////////////////////////////
//! 数据结构
type DBMgr struct {
	db_R *dao.DB_r //! redis操作
	db_M *gorm.DB  //! mysql
	db_S *gorm.DB  //! c++服务器 sqlServer数据库
}

var DBMgrSingleton *DBMgr = nil

//! 得到包厢管理单例
func GetDBMgr() *DBMgr {
	if DBMgrSingleton == nil {
		DBMgrSingleton = new(DBMgr)
		DBMgrSingleton.db_R = new(dao.DB_r)

		con := GetServer().Con
		//! redis
		err := DBMgrSingleton.db_R.Init(con.Redis, con.RedisDB, con.RedisAuth)

		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("init redis error: %s", err.Error()))
			return nil
		}
		DBMgrSingleton.db_M, err = gorm.Open("mysql", con.DB)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("init database error: %s", err.Error()))
			return nil
		}

		DBMgrSingleton.db_M.SetLogger(xlog.DBLogger())
		DBMgrSingleton.db_M.LogMode(true)
		// 设置连接可被重新使用的最大时间间隔
		DBMgrSingleton.db_M.DB().SetConnMaxLifetime(600 * time.Second)
		// 设置最大闲置的连接数
		//DBMgrSingleton.db_M.DB().SetMaxIdleConns(0)
		// 设置最大打开的连接数
		//DBMgrSingleton.db_M.DB().SetMaxOpenConns(0)
		//! basedata
		//err = DBMgrSingleton.DataInitialize()
		//if err != nil {
		//	syslog.Logger().Errorln(err)
		//	return nil
		//}
	}
	return DBMgrSingleton
}

func (self *DBMgr) GetDBmControl() *gorm.DB {
	return self.db_M
}

func (self *DBMgr) GetDBsControl() *gorm.DB {
	return self.db_S
}

func (self *DBMgr) GetDBrControl() *dao.DB_r {
	return self.db_R
}

func (self *DBMgr) ReadAllConfig() error {
	if err := self.GetDBmControl().First(&GetServer().ConServers).Error; err != nil {
		return err
	}
	xlog.Logger().Infoln(GetServer().ConServers)
	if err := self.GetDBmControl().Find(&GetServer().ConApp).Error; err != nil {
		return err
	}

	if err := self.GetDBmControl().First(&GetServer().ConQiNiu).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	xlog.Logger().Infoln(GetServer().ConQiNiu)
	if err := self.GetDBmControl().First(&GetServer().ConSqlServer).Error; err == nil {
		xlog.Logger().Infoln(GetServer().ConSqlServer)
		self.initSQLserver()
	}
	return nil
}

func (self *DBMgr) initSQLserver() {
	if GetServer().ConSqlServer.Able > 0 {
		fmt.Println(GetServer().ConSqlServer.Connect)
		var err error
		self.db_S, err = gorm.Open("mssql", GetServer().ConSqlServer.Connect)
		if err != nil {
			xlog.Logger().Panicln("sqlserver db init err:", err)
			return
		}
		self.db_S.LogMode(true)
		// 设置连接可被重新使用的最大时间间隔
		self.db_S.DB().SetConnMaxLifetime(600 * time.Second)
		xlog.Logger().Warningln("连接 sqlserver 数据库成功:", GetServer().ConSqlServer.Name)
	}
}

func (self *DBMgr) Close() error {

	if self.db_R != nil {
		self.db_R.Close()
	}

	if self.db_M != nil {
		err := self.db_M.Close()
		if err != nil {
			log.Panic("dbmap close error. ", err)
			return err
		}
	}

	return nil
}

//! 查询战绩详情
func (self *DBMgr) SelectGameRecordInfo(gamenum string) (static.Msg_S2C_GameRecordInfo, error) {
	var gametotals []models.RecordGameTotal = []models.RecordGameTotal{}
	err := self.GetDBmControl().Model(models.RecordGameTotal{}).Where("game_num = ?", gamenum).Order("seat_id ASC").Find(&gametotals).Error

	if err != nil || len(gametotals) == 0 {
		return static.Msg_S2C_GameRecordInfo{}, err
	}

	playercount := len(gametotals)
	playcount := gametotals[0].PlayCount

	var gamerounds []models.RecordGameRound = []models.RecordGameRound{}
	err = self.GetDBmControl().Model(models.RecordGameRound{}).Where("gamenum = ?", gamenum).Order("created_at ASC").Find(&gamerounds).Error
	if err != nil {
		return static.Msg_S2C_GameRecordInfo{}, err
	}

	roundRecord := new(static.Msg_S2C_GameRecordInfo)
	roundRecord.GameNum = gamenum
	roundRecord.KindId = gametotals[0].KindId
	roundRecord.RoomId = gametotals[0].RoomNum
	roundRecord.Time = time.Now().Unix()

	// 玩家列表
	for i := 0; i < playercount; i++ {
		roundRecord.UserArr = append(roundRecord.UserArr, &static.Msg_S2C_GameRecordInfoUser{
			Uid:      gametotals[i].Uid,
			Nickname: gametotals[i].UName,
			Imgurl:   "",
			Sex:      0,
			Score:    gametotals[i].GetRealScore(),
		})
	}

	// 每局积分
	scoreArr := make([][]float64, 0)
	uidArr := make([][]int64, 0)

	// 初始化二维数组
	for i := 0; i < playcount; i++ {
		arr := make([]float64, playercount)
		scoreArr = append(scoreArr, arr)

		arr2 := make([]int64, playercount)
		uidArr = append(uidArr, arr2)
	}

	replayIdArr := make([]int64, 0)
	endTimeArr := make([]int64, 0)
	startTimeArr := make([]int64, 0)
	for i := 0; i < playercount; i++ {
		list := self.SelectUserGameRecordInfo(i, gamerounds)
		for j := 0; j < len(list); j++ {

			item := list[j]
			// 测试时屏蔽, 为避免异常情况, 上线时应打开
			if j >= playcount {
				continue
			}
			scoreArr[j][i] = item.GetRealScore()
			if i == 0 {
				replayIdArr = append(replayIdArr, item.ReplayId)
				endTimeArr = append(endTimeArr, item.CreatedAt.Unix())
				startTimeArr = append(startTimeArr, item.BeginDate.Unix())
			}
			uidArr[j][i] = item.UId
		}
	}

	for i := 0; i < playcount; i++ {
		if i >= len(replayIdArr) {
			break
		}
		roundRecord.ScoreArr = append(roundRecord.ScoreArr, &static.Msg_S2C_GameRecordInfoScore{
			ReplayId:  replayIdArr[i],
			StartTime: startTimeArr[i],
			EndTime:   endTimeArr[i],
			Score:     scoreArr[i],
			Uids:      uidArr[i],
		})
	}

	return *roundRecord, nil
}

func (self *DBMgr) SelectUserGameRecordInfo(seatid int, gamerounds []models.RecordGameRound) []models.RecordGameRound {
	usergamerounds := []models.RecordGameRound{}
	for _, gameround := range gamerounds {
		if gameround.SeatId == seatid {
			usergamerounds = append(usergamerounds, gameround)
		}
	}
	return usergamerounds
}
