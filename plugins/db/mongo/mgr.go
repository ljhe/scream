package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
)

type Mgr struct {
	cli *mongo.Client
	sync.Mutex
}

func NewMgr() *Mgr {
	return &Mgr{}
}

func (m *Mgr) connect(uri string) error {
	opt := options.Client()
	opt.ApplyURI(uri)
	opt.SetConnectTimeout(config.connectTimeout)
	opt.SetMaxPoolSize(config.MaxPoolSize)

	cli, err := mongo.Connect(context.TODO(), opt)
	if err != nil {
		panic(fmt.Errorf("mgo connect err:%s", err))
	}

	err = cli.Ping(context.Background(), nil)
	if err != nil {
		panic(fmt.Errorf("mgo ping err:%s", err))
	}

	m.cli = cli
	return nil
}

func (m *Mgr) Start(uri string) error {
	return m.connect(uri)
}
