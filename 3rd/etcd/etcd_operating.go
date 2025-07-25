package etcd

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/3rd/log"
	"go.etcd.io/etcd/client/v3"
)

func Register(ctx context.Context, key string, val []byte) error {
	// 先查询是否存在该节点 如果存在不做处理
	resp, err := etcdDiscovery.KV.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("etcd register error:%v key:%s", err, key)
	}
	if resp.Count > 0 {
		return fmt.Errorf("etcd register error: service already exist. key:%s", key)
	}

	// 注册
	err = etcdDiscovery.RegisterService(key, string(val))
	if err != nil {
		return fmt.Errorf("etcd register service error:%v key:%s", err, key)
	}
	log.InfoF("etcd register success. key:%s", key)
	return nil
}

func UnRegister(ctx context.Context, key string) error {
	return etcdDiscovery.DelServices(ctx, key)
}

func Discovery(ctx context.Context, etcdKey string, f func(string, []byte)) {
	// 监测目标节点的变化
	var ch clientv3.WatchChan
	ch = etcdDiscovery.Cli.Watch(ctx, etcdKey, clientv3.WithPrefix())

	for {
		select {
		case c := <-ch:
			for _, ev := range c.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					fallthrough
				case clientv3.EventTypeDelete:
					f(string(ev.Kv.Key), ev.Kv.Value)
				}
			}
		case <-ctx.Done():
			log.InfoF("etcd discovery exit")
			return
		}
	}
}
