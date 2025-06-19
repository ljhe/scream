package discover

import (
	"context"
	"github.com/ljhe/scream/3rd/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

type Discover struct {
	IdMap  map[string][]byte
	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
}

func (d *Discover) Loader() {
	d.IdMap = make(map[string][]byte)
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

func (d *Discover) setIdMap(key string, value []byte) {
	d.RLock()
	defer d.RUnlock()
	d.IdMap[key] = value
}

func (d *Discover) delIdMap(key string) {
	d.RLock()
	defer d.RUnlock()
	delete(d.IdMap, key)
}
