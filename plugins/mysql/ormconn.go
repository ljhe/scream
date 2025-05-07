package mysql

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"plugins/logrus"
	"sync"
)

type OrmConn struct {
	db    *sql.DB
	ormDB *gorm.DB
	mu    sync.RWMutex
}

func NewOrmConn() *OrmConn {
	return &OrmConn{}
}

func (oc *OrmConn) GetOrmDB() *gorm.DB {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return oc.ormDB
}

func (oc *OrmConn) tryConnect(dsn string) error {
	ormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("mysql connect err:%v dsn:%v \n", err, dsn)
		return err
	}
	oc.ormDB = ormDB

	db, err := oc.ormDB.DB()
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("mysql get db err:%v dsn:%v \n", err, dsn)
		return err
	}

	db.SetMaxIdleConns(GormConfig.MaxIdleConns)
	db.SetMaxOpenConns(GormConfig.MaxOpenConns)
	oc.db = db
	return nil
}

func (oc *OrmConn) Start(dsn string) error {
	return oc.tryConnect(dsn)
}
