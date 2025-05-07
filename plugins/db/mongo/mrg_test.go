package mongo

import (
	"context"
	"fmt"
	"testing"
)

type User struct {
	Name   string
	OpenId string
}

func TestNewMgrConn(t *testing.T) {
	mgr := NewMgr()
	err := mgr.Start("mongodb://localhost:27017")
	if err != nil {
		panic(fmt.Errorf("init mgr err:%s", err))
	}

	create(mgr)
}

func create(mgr *Mgr) {
	user := User{
		Name:   "ljh",
		OpenId: "123456",
	}

	coll := mgr.cli.Database("gamedb_9999").Collection("users")
	res, err := coll.InsertOne(context.TODO(), user)
	if err != nil {
		panic(fmt.Errorf("insert mgr err:%s", err))
	}
	fmt.Printf("insert result:%v\n", res)
}
