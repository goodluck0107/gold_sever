//! 数据库底层
package api

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/dao"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"log"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

////////////////////////////////////////////////////////////////////////////////
//! 数据结构
type DBMgr struct {
	db_R  *dao.DB_r //! redis操作
	db_M  *gorm.DB  //! mysql操作
	Redis *redis.Client
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
		}
		// 初始化v2版本redis库
		DBMgrSingleton.Redis = static.InitRedisV2(con.Redis, consts.REDIS_DB_API_SVR, con.RedisAuth)
		//! mysql
		DBMgrSingleton.db_M, err = gorm.Open("mysql", con.DB)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("init mysql error: %s", err.Error()))
		}

		DBMgrSingleton.db_M.LogMode(true)
		DBMgrSingleton.db_M.SetLogger(xlog.DBLogger())

		DBMgrSingleton.db_M.DB().SetConnMaxLifetime(600 * time.Second)

		if err = DBMgrSingleton.ReadAllConfig(); err != nil {
			xlog.Logger().Errorln("DBMgrSingleton.ReadAllConfig:", err)
			return nil
		}
	}
	return DBMgrSingleton
}

func (self *DBMgr) GetDBmControl() *gorm.DB {
	return self.db_M
}

func (self *DBMgr) GetDBrControl() *dao.DB_r {
	return self.db_R
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

//！从数据库读取配置
func (self *DBMgr) ReadAllConfig() error {
	var err error
	xlog.Logger().Infoln("read config_games...")
	if err = self.GetDBmControl().Find(&GetServer().ConGame).Error; err != nil {
		return err
	}
	return err
}
