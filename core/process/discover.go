package process

import (
	"context"
	"encoding/json"
	"github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/3rd/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

var idMap = make(map[string][]byte)

type Discover struct {
	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
}

func NewDiscover() *Discover {
	d := &Discover{}
	d.Loader()
	return d
}

func (d *Discover) Loader() {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
	defer cancel()
	resp, err := etcd.GetEtcdDiscovery().Cli.Get(ctx, etcd.ServerPreKey, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	for _, kv := range resp.Kvs {
		d.setIdMap(string(kv.Key), kv.Value)
	}

	go d.watch()
}

func (d *Discover) Close() {
	d.cancel()
}

func (d *Discover) watch() {
	watchChan := etcd.GetEtcdDiscovery().Cli.Watch(context.TODO(), etcd.ServerPreKey, clientv3.WithPrefix())
	for {
		select {
		case wr := <-watchChan:
			for _, event := range wr.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					d.setIdMap(string(event.Kv.Key), event.Kv.Value)
				case clientv3.EventTypeDelete:
					d.delIdMap(string(event.Kv.Key))
				}
			}
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *Discover) GetNodeByKey(key string) *etcd.ServerInfo {
	d.RLock()
	defer d.RUnlock()
	res := &etcd.ServerInfo{}
	err := json.Unmarshal(idMap[key], res)
	if err != nil {
		logrus.Errorf("[ GetNodeByKey ] Unmarshal err: %v", err)
		return nil
	}
	return res
}

func (d *Discover) setIdMap(key string, value []byte) {
	d.Lock()
	defer d.Unlock()
	idMap[key] = value
}

func (d *Discover) delIdMap(key string) {
	d.Lock()
	defer d.Unlock()
	delete(idMap, key)
}
