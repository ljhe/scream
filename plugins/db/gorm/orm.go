package gorm

import (
	"database/sql"
	"github.com/ljhe/scream/plugins/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

type Orm struct {
	db    *sql.DB
	ormDB *gorm.DB
	sync.RWMutex
}

func NewOrmConn() *Orm {
	return &Orm{}
}

func (o *Orm) GetOrmDB() *gorm.DB {
	o.RLock()
	defer o.RUnlock()
	return o.ormDB
}

func (o *Orm) connect(dsn string) error {
	ormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: config.SkipDefaultTransaction,
	})
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("db connect err:%v dsn:%v \n", err, dsn)
		return err
	}
	o.ormDB = ormDB

	db, err := o.ormDB.DB()
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("db get db err:%v dsn:%v \n", err, dsn)
		return err
	}

	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	o.db = db
	return nil
}

func (o *Orm) Start(dsn string) error {
	return o.connect(dsn)
}
