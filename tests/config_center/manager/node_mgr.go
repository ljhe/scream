package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/3rd/etcd"
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
						fmt.Printf("etcd watch event put key=%v value=%v \n", string(event.Kv.Key), string(event.Kv.Value))
						var info etcd.ServerInfo
						err = json.Unmarshal(event.Kv.Value, &info)
						if err != nil {
							panic(err)
						}
						c.nodes.Store(string(event.Kv.Key), &info)
						c.print()
					case clientv3.EventTypeDelete:
						fmt.Printf("etcd watch event del key=%v \n", string(event.Kv.Key))
						c.nodes.Delete(string(event.Kv.Key))
						c.print()
					}
				}
			}
		}
	}()
}

func (c *Center) print() {
	fmt.Println("--------------------------------print node info begin-------------------------------")
	c.nodes.Range(func(key, value interface{}) bool {
		fmt.Printf("key=%v value=%v\n", key, value)
		return true
	})
	fmt.Println("--------------------------------print node info  end--------------------------------")
}
