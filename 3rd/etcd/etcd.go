package etcd

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/utils"
	"go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"time"
)

var etcdDiscovery *ServiceDiscovery

type ServiceDiscovery struct {
	Cli    *clientv3.Client
	KV     clientv3.KV
	config clientv3.Config
	mu     sync.RWMutex
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
	etcdDiscovery.KV = clientv3.NewKV(cli)
	etcdDiscovery.Cli = cli
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
	leaseResp, err := sd.Cli.Grant(context.TODO(), 3)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	rsp, err := sd.KV.Put(ctx, key, val, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	ch, err := sd.Cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for resp := range ch {
			if resp == nil {
				// 租约已失效
				log.Println("etcd keep alive channel closed")
				return
			}
			// 记录日志或处理响应
		}
	}()

	logrus.Infof("etcd register ok. key=%v clusterid=%v leaseid=%v etcdaddr=%v", key, rsp.Header.ClusterId, leaseResp.ID, sd.config)
	return nil
}

func (sd *ServiceDiscovery) DiscoverService(key string) error {
	watchChan := sd.Cli.Watch(context.TODO(), key, clientv3.WithPrefix())
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

func (sd *ServiceDiscovery) WatchServices(key string, value utils.ServerInfo) {
	watchChan := sd.Cli.Watch(context.TODO(), key, clientv3.WithPrefix())
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
						logrus.Infof("reset keep alive")
					}
				}
			}
		}
	}()
}

func (sd *ServiceDiscovery) DelServices(ctx context.Context, key string) error {
	resp, err := sd.Cli.Delete(ctx, key)
	if resp.Deleted == 0 {
		return fmt.Errorf("etcd del service key=%v deleted=0", key)
	}
	return err
}
