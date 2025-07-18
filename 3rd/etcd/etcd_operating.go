package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/utils"
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
	logrus.Infof("etcd register success. key:%s", key)
	return nil
}

func UnRegister(ctx context.Context, key string) error {
	return etcdDiscovery.DelServices(ctx, key)
}

func DiscoveryService(etcdKey string, nodeCreator func(*utils.ServerInfo)) iface.INetNode {
	// 如果已经存在 就停止之前正在运行的节点(注意不要配置成一样的节点信息 否则会关闭之前的连接)
	// 连接同一个zone里的服务器节点

	// 监测目标节点的变化
	var ch clientv3.WatchChan
	ch = etcdDiscovery.Cli.Watch(context.TODO(), etcdKey, clientv3.WithPrefix())

	go func() {
		resp, err := etcdDiscovery.KV.Get(context.TODO(), etcdKey, clientv3.WithPrefix())
		if err != nil {
			logrus.Errorf("etcd discovery error:%v", err)
			return
		}
		logrus.Printf("service[%v] node find count:%v", etcdKey, resp.Count)
		for _, data := range resp.Kvs {
			var ed utils.ServerInfo
			err = json.Unmarshal(data.Value, &ed)
			if err != nil {
				logrus.Printf("etcd discovery unmarshal error:%v key:%v", err, data.Key)
				continue
			}
			// todo 先停止之前的连接 再执行新的连接
			nodeCreator(&ed)
		}

		for {
			select {
			case c := <-ch:
				for _, ev := range c.Events {
					switch ev.Type {
					case clientv3.EventTypePut:
						var ed utils.ServerInfo
						err = json.Unmarshal(ev.Kv.Value, &ed)
						if err != nil {
							logrus.Printf("etcd discovery unmarshal error:%v key:%v", err, ev.Kv.Key)
							continue
						}
						logrus.Infof("etcd discovery start connect:%v", string(ev.Kv.Key))
						// todo 先停止之前的连接 再执行新的连接
						nodeCreator(&ed)
					case clientv3.EventTypeDelete:

					}
				}
			}
		}
	}()
	return nil
}
