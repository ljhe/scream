package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"testing"
)

type User struct {
	gorm.Model
	Name   string
	OpenId string
}

func TestNewOrmConn(t *testing.T) {
	orm := NewOrmConn()
	err := orm.Start("root:123456@(127.0.0.1:3306)/gamedb_9999?charset=utf8&loc=Asia%2FShanghai&parseTime=true")
	if err != nil {
		log.Printf("init db err:%v", err)
		return
	}

	create(orm)
}

func create(orm *OrmConn) {
	user := User{
		Name:   "ljh",
		OpenId: "123456",
	}

	// 如果没有这张表 就自动创建
	err := orm.GetOrmDB().AutoMigrate(&user)
	if err != nil {
		log.Printf("mysql create err:%v", err)
		return
	}

	res := orm.ormDB.Save(&user)
	fmt.Printf("create user res:%v \n", res)
}
