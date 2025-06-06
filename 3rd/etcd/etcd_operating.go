package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/utils"
	"go.etcd.io/etcd/client/v3"
	"log"
	"strconv"
	"strings"
)

var ServerPreKey = "server/"

type ServerInfo struct {
	Id      string
	Name    string
	Host    string
	Typ     int
	Zone    int
	Index   int
	RegTime int64
}

func (e *ServerInfo) String() string {
	data, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(data)
}

func Register(node iface.INetNode) *ServerInfo {
	property := node.(iface.INodeProp)
	ed := &ServerInfo{
		Id:    utils.GenServiceId(property),
		Name:  property.GetName(),
		Host:  property.GetAddr(),
		Typ:   property.GetServerTyp(),
		Zone:  property.GetZone(),
		Index: property.GetIndex(),
	}
	ed.RegTime = utils.GetTimeSeconds()

	// 先查询是否存在该节点 如果存在不做处理(或者通过del操作关闭其他客户端)
	etcdKey := genServicePrefix(ed.Id, property.GetZone())
	resp, err := etcdDiscovery.KV.Get(context.TODO(), etcdKey)
	if err != nil {
		log.Println("etcd register error:", err)
		return nil
	}
	if resp.Count > 0 {
		fmt.Println("etcd register error: service already exist. etcdKey:", etcdKey)
		return nil
	}

	// 注册
	err = etcdDiscovery.RegisterService(etcdKey, ed.String())
	if err != nil {
		log.Println("etcd register error:", err)
		return nil
	}
	etcdDiscovery.WatchServices(etcdKey, *ed)
	logrus.Log(logrus.LogsSystem).Info("etcd register success:", ed.Id)
	return ed
}

func UnRegister(node iface.INetNode) error {
	property := node.(iface.INodeProp)
	ed := &ServerInfo{
		Id: utils.GenServiceId(property),
	}

	etcdKey := genServicePrefix(ed.Id, property.GetZone())
	return etcdDiscovery.DelServices(etcdKey)
}

func DiscoveryService(multiNode MultiServerNode, serviceName string, zone int, nodeCreator func(MultiServerNode, *ServerInfo)) iface.INetNode {
	// 如果已经存在 就停止之前正在运行的节点(注意不要配置成一样的节点信息 否则会关闭之前的连接)
	// 连接同一个zone里的服务器节点
	etcdKey := genDiscoveryServicePrefix(serviceName, zone)

	// 监测目标节点的变化
	var ch clientv3.WatchChan
	ch = etcdDiscovery.Cli.Watch(context.TODO(), etcdKey, clientv3.WithPrefix())

	go func() {
		resp, err := etcdDiscovery.KV.Get(context.TODO(), etcdKey, clientv3.WithPrefix())
		if err != nil {
			log.Println("etcd discovery error:", err)
			return
		}
		log.Printf("service[%v] node find count:%v \n", etcdKey, resp.Count)
		for _, data := range resp.Kvs {
			log.Println("etcd discovery start connect:", string(data.Key))
			var ed ServerInfo
			err = json.Unmarshal(data.Value, &ed)
			if err != nil {
				log.Printf("etcd discovery unmarshal error:%v key:%v \n", err, data.Key)
				continue
			}
			// 先停止之前的连接 再执行新的连接
			if preNode := multiNode.GetNode(ed.Id); preNode != nil {
				multiNode.DelNode(ed.Id, serviceName)
				preNode.Stop()
			}
			nodeCreator(multiNode, &ed)
		}

		for {
			select {
			case c := <-ch:
				for _, ev := range c.Events {
					switch ev.Type {
					case clientv3.EventTypePut:
						var ed ServerInfo
						err = json.Unmarshal(ev.Kv.Value, &ed)
						if err != nil {
							log.Printf("etcd discovery unmarshal error:%v key:%v \n", err, ev.Kv.Key)
							continue
						}
						log.Println("etcd discovery start connect:", string(ev.Kv.Key))
						// 先停止之前的连接 再执行新的连接
						if preNode := multiNode.GetNode(ed.Id); preNode != nil {
							multiNode.DelNode(ed.Id, serviceName)
							preNode.Stop()
							log.Println(fmt.Sprintf("del old node. id:%v", ed.Id))
						}
						nodeCreator(multiNode, &ed)
					case clientv3.EventTypeDelete:
						nodeID := getNodeId(string(ev.Kv.Key))
						if preNode := multiNode.GetNode(nodeID); preNode != nil {
							log.Println(fmt.Sprintf("del node. id:%v", nodeID))
							multiNode.DelNode(nodeID, serviceName)
							preNode.Stop()
						}
					}
				}
			}
		}
	}()
	return nil
}

func genServicePrefix(name string, zone int) string {
	//return ServerPreKey + strconv.Itoa(zone) + "/" + name
	return ServerPreKey + name
}

func genDiscoveryServicePrefix(name string, zone int) string {
	if zone > 0 {
		return ServerPreKey + name + "#" + strconv.Itoa(zone)
	}
	return ServerPreKey + name + "#"
}

func getNodeId(key string) string {
	list := strings.Split(key, "/")
	if len(list) >= 2 {
		return list[1]
	}
	return ""
}

func ParseServiceId(sid string) (typ, zone, idx int, err error) {
	str := strings.Split(sid, "#")
	if len(str) < 2 {
		err = errors.New(fmt.Sprintf("ParseServiceId sid invalid. sid:%s", sid))
		return
	} else {
		strProp := strings.Split(str[1], "@")
		if len(strProp) < 3 {
			err = errors.New(fmt.Sprintf("ParseServiceId sid invalid. sid:%s", sid))
			return
		} else {
			zone, err = utils.StrToInt(strProp[0])
			if err != nil {
				return
			}
			typ, err = utils.StrToInt(strProp[1])
			if err != nil {
				return
			}
			idx, err = utils.StrToInt(strProp[2])
			if err != nil {
				return
			}
		}
	}
	return
}
