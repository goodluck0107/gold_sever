package dao

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"

	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
)

type ORM_Mysql struct {
	conn    *gorm.DB
	sqlChan chan string
}

type Model struct {
	gorm.Model
}

// Open creates a database connection, or returns an existing one if present.
func (orm *ORM_Mysql) Open(dialect, connectionString string) (*gorm.DB, error) {
	if orm.conn != nil {
		return orm.conn, nil
	}

	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		return nil, err
	}

	// Configure
	db.SetLogger(xlog.DBLogger())
	db.LogMode(true)
	db.DB().SetMaxIdleConns(100)
	db.DB().SetMaxOpenConns(0) // Unlimited
	db.DB().SetConnMaxLifetime(600 * time.Second)

	orm.conn = db

	// chan
	orm.sqlChan = make(chan string, 500)
	go func(db *gorm.DB, ch chan string) {
		for {
			sqlStr := <-ch
			if err := db.Exec(sqlStr).Error; err != nil {
				xlog.Logger().Errorf("sql = %s, err = %s", sqlStr, err.Error())
			}

		}
	}(orm.conn, orm.sqlChan)

	return orm.conn, nil
}

func (orm *ORM_Mysql) Close() {
	orm.conn.Close()
}

func (orm *ORM_Mysql) GetConn() *gorm.DB {
	return orm.conn
}

func (orm *ORM_Mysql) AutoMigrate(models []interface{}) error {
	for _, model := range models {
		if err := orm.conn.AutoMigrate(model).Error; err != nil {
			return err
		}
	}

	return nil
}

func (orm *ORM_Mysql) Exec(query string, output interface{}) *gorm.DB {
	return orm.conn.Raw(query).Scan(output)
}

// 异步执行
func (orm *ORM_Mysql) AsyncExec(sqlStr string) {
	orm.sqlChan <- sqlStr
}

func (orm *ORM_Mysql) Begin() *gorm.DB {
	return orm.conn.Begin()
}

func (orm *ORM_Mysql) Where(query interface{}, args ...interface{}) *gorm.DB {
	return orm.conn.Where(query, args...)
}

func (orm *ORM_Mysql) Create(model interface{}) *gorm.DB {
	return orm.conn.Create(model)
}

func (orm *ORM_Mysql) Delete(value interface{}, where ...interface{}) *gorm.DB {
	return orm.conn.Delete(value, where)
}

func (orm *ORM_Mysql) Save(value interface{}) *gorm.DB {
	return orm.conn.Save(value)
}

func (orm *ORM_Mysql) Update(attrs ...interface{}) *gorm.DB {
	return orm.conn.Update(attrs...)
}

func (orm *ORM_Mysql) Updates(values interface{}, ignoreProtectedAttrs ...bool) *gorm.DB {
	return orm.conn.Updates(values, ignoreProtectedAttrs...)
}

func (orm *ORM_Mysql) Model(model interface{}) *gorm.DB {
	return orm.conn.Model(model)
}

func (orm *ORM_Mysql) Find(out interface{}, where ...interface{}) *gorm.DB {
	return orm.conn.Find(out, where...)
}

func (orm *ORM_Mysql) First(model interface{}, where ...interface{}) *gorm.DB {
	return orm.conn.First(model, where...)
}

func (orm *ORM_Mysql) Last(model interface{}, where ...interface{}) *gorm.DB {
	return orm.conn.Last(model, where...)
}

func (orm *ORM_Mysql) ModelWithID(model interface{}, id uint) error {

	if exists, err := orm.ModelExistsWithID(model, id); err != nil {
		return err
	} else if !exists {
		return nil
	}

	if err := orm.First(model, id).Error; err != nil {
		return err
	}

	return nil
}

func (orm *ORM_Mysql) ModelExistsWithID(model interface{}, id uint) (bool, error) {
	var count int64

	err := orm.Model(model).Where(id).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 事务封装
func (orm *ORM_Mysql) Transaction(f func(tx *gorm.DB) error) error {
	tx := orm.Begin()

	err := f(tx)

	if err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return errors.Wrap(rbErr, "orm: transaction rollback error")
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.Wrap(err, "orm: transaction commit error")
	}

	return nil
}

func (orm *ORM_Mysql) Sum(model interface{}, column string, query string, args ...interface{}) (sum float64) {
	sql := fmt.Sprintf("SUM(%s)", column)
	err := orm.conn.Model(model).Select(sql).Where(query, args...).Row().Scan(&sum)
	if err != nil {
		xlog.Logger().Warn("sum empty record:", err)
	}
	return
}
