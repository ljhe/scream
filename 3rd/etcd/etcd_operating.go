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

func Register(node iface.INetNode) *utils.ServerInfo {
	property := node.(iface.INodeProp)
	ed := &utils.ServerInfo{
		Id:    utils.GenSelfServiceId(property.GetName(), property.GetServerTyp(), property.GetIndex()),
		Name:  property.GetName(),
		Host:  property.GetAddr(),
		Typ:   property.GetServerTyp(),
		Index: property.GetIndex(),
	}
	ed.RegTime = utils.GetTimeSeconds()

	// 先查询是否存在该节点 如果存在不做处理(或者通过del操作关闭其他客户端)
	etcdKey := utils.GenServicePrefix(ed.Id)
	resp, err := etcdDiscovery.KV.Get(context.TODO(), etcdKey)
	if err != nil {
		logrus.Infof("etcd register error:%v", err)
		return nil
	}
	if resp.Count > 0 {
		fmt.Println("etcd register error: service already exist. etcdKey:", etcdKey)
		return nil
	}

	// 注册
	err = etcdDiscovery.RegisterService(etcdKey, ed.String())
	if err != nil {
		logrus.Errorf("etcd register error:%v", err)
		return nil
	}
	etcdDiscovery.WatchServices(etcdKey, *ed)
	logrus.Infof("etcd register success:%s", ed.Id)
	return ed
}

func UnRegister(node iface.INetNode) error {
	property := node.(iface.INodeProp)
	ed := &utils.ServerInfo{
		Id: utils.GenSelfServiceId(property.GetName(), property.GetServerTyp(), property.GetIndex()),
	}

	etcdKey := utils.GenServicePrefix(ed.Id)
	return etcdDiscovery.DelServices(etcdKey)
}

func DiscoveryService(serviceName string, nodeCreator func(*utils.ServerInfo)) iface.INetNode {
	// 如果已经存在 就停止之前正在运行的节点(注意不要配置成一样的节点信息 否则会关闭之前的连接)
	// 连接同一个zone里的服务器节点
	etcdKey := utils.GenDiscoveryServicePrefix(serviceName)

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
