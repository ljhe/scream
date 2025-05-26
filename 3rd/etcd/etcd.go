package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/utils"
	"sync"
	"time"
)

var etcdDiscovery *ServiceDiscovery

type ServiceDiscovery struct {
	config  clientv3.Config
	etcdCli *clientv3.Client
	etcdKV  clientv3.KV
	mu      sync.RWMutex
}

func NewServiceDiscovery(addr string) (*ServiceDiscovery, error) {
	etcdDiscovery = &ServiceDiscovery{
		config: clientv3.Config{
			Endpoints:   []string{addr},
			DialTimeout: 3 * time.Second,
		},
	}
	cli, err := clientv3.New(etcdDiscovery.config)
	if err != nil {
		return nil, err
	}
	etcdDiscovery.etcdKV = clientv3.NewKV(cli)
	etcdDiscovery.etcdCli = cli
	return etcdDiscovery, nil
}

func InitServiceDiscovery(addr string) error {
	if GetEtcdDiscovery() != nil {
		return nil
	}
	_, err := NewServiceDiscovery(addr)
	return err
}

func GetEtcdDiscovery() *ServiceDiscovery {
	return etcdDiscovery
}

func (sd *ServiceDiscovery) RegisterService(key, val string) error {
	leaseResp, err := sd.etcdCli.Grant(context.TODO(), 3)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	rsp, err := sd.etcdKV.Put(ctx, key, val, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	_, err = sd.etcdCli.KeepAlive(context.TODO(), leaseResp.ID)
	if err != nil {
		return err
	}
	logrus.Log(logrus.LogsSystem).Infof("etcd register ok. key=%v clusterid=%v leaseid=%v etcdaddr=%v", key, rsp.Header.ClusterId, leaseResp.ID, sd.config)
	return nil
}

func (sd *ServiceDiscovery) DiscoverService(key string) error {
	watchChan := sd.etcdCli.Watch(context.TODO(), key, clientv3.WithPrefix())
	go func() {
		for {
			select {
			case wr := <-watchChan:
				for _, event := range wr.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						fmt.Printf("etcd watch event put key=%v value=%v \n", string(event.Kv.Key), string(event.Kv.Value))
					case clientv3.EventTypeDelete:
						fmt.Printf("etcd watch event del key=%v \n", string(event.Kv.Key))
					}
				}
			}
		}
	}()
	return nil
}

func (sd *ServiceDiscovery) WatchServices(key string, value ETCDServiceDesc) {
	watchChan := sd.etcdCli.Watch(context.TODO(), key, clientv3.WithPrefix())
	go func() {
		for {
			select {
			case wr := <-watchChan:
				for _, event := range wr.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						fmt.Printf("etcd watch event put key=%v value=%v \n", string(event.Kv.Key), string(event.Kv.Value))
					case clientv3.EventTypeDelete:
						// 网络恢复后得到自己被删除的通知 重新设置key租约
						value.RegTime = utils.GetTimeSeconds()
						err := sd.RegisterService(key, value.String())
						if err != nil {
							fmt.Printf("etcd watch event del key=%v err=%v \n", string(event.Kv.Key), err)
						}
					}
				}
			}
		}
	}()
}

func (sd *ServiceDiscovery) DelServices(key string) error {
	_, err := sd.etcdCli.Delete(context.TODO(), key)
	return err
}
