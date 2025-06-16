package manager

import (
	"context"
	"encoding/json"
	"github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/def"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
)

type Center struct {
	nodes sync.Map
}

func NewCenter() *Center {
	return &Center{}
}

func (c *Center) Run() {
	go c.watch()
}

func (c *Center) watch() {
	ed, err := etcd.NewServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		panic(err)
	}

	watchChan := ed.Cli.Watch(context.TODO(), etcd.ServerPreKey, clientv3.WithPrefix())
	go func() {
		for {
			select {
			case wr := <-watchChan:
				for _, event := range wr.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						var info etcd.ServerInfo
						err = json.Unmarshal(event.Kv.Value, &info)
						if err != nil {
							panic(err)
						}

						c.nodes.Store(string(event.Kv.Key), &info)

						logrus.Log(def.LogsConfigCenter, map[string]interface{}{
							"key": string(event.Kv.Key),
						}).Infof("add server")
						c.print()
					case clientv3.EventTypeDelete:
						c.nodes.Delete(string(event.Kv.Key))

						logrus.Log(def.LogsConfigCenter, map[string]interface{}{
							"key": string(event.Kv.Key),
						}).Infof("del server")
						c.print()
					}
				}
			}
		}
	}()
}

func (c *Center) print() {
	c.nodes.Range(func(key, value interface{}) bool {
		logrus.Log(def.LogsConfigCenter, map[string]interface{}{
			"key": key.(string),
			"val": value,
		}).Infof("server info change, now all server info")
		return true
	})
}
