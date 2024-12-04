package plugins

import (
	"common"
	"common/iface"
	"common/plugins/logrus"
	"common/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"strconv"
	"strings"
)

var servicePrefixKey = "server/"

type ETCDServiceDesc struct {
	Id      string
	Name    string
	Host    string
	Typ     int
	Zone    int
	Index   int
	RegTime int64
}

func (e *ETCDServiceDesc) String() string {
	data, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(data)
}

func ETCDRegister(node iface.INetNode) *ETCDServiceDesc {
	property := node.(common.ServerNodeProperty)
	ed := &ETCDServiceDesc{
		Id:    util.GenServiceId(property),
		Name:  property.GetName(),
		Host:  property.GetAddr(),
		Typ:   property.GetServerTyp(),
		Zone:  property.GetZone(),
		Index: property.GetIndex(),
	}
	ed.RegTime = util.GetTimeSeconds()

	// 先查询是否存在该节点 如果存在不做处理(或者通过del操作关闭其他客户端)
	etcdKey := genServicePrefix(ed.Id, property.GetZone())
	resp, err := etcdDiscovery.etcdKV.Get(context.TODO(), etcdKey)
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
	setServiceStartupTime(property.GetZone())
	logrus.Log(logrus.LogsSystem).Info("etcd register success:", ed.Id)
	return ed
}

func DiscoveryService(multiNode MultiServerNode, serviceName string, zone int, nodeCreator func(MultiServerNode, *ETCDServiceDesc)) iface.INetNode {
	// 如果已经存在 就停止之前正在运行的节点(注意不要配置成一样的节点信息 否则会关闭之前的连接)
	// 连接同一个zone里的服务器节点
	etcdKey := genDiscoveryServicePrefix(serviceName, zone)

	// 监测目标节点的变化
	var ch clientv3.WatchChan
	ch = etcdDiscovery.etcdCli.Watch(context.TODO(), etcdKey, clientv3.WithPrefix())

	go func() {
		resp, err := etcdDiscovery.etcdKV.Get(context.TODO(), etcdKey, clientv3.WithPrefix())
		if err != nil {
			log.Println("etcd discovery error:", err)
			return
		}
		log.Printf("service[%v] node find count:%v \n", etcdKey, resp.Count)
		for _, data := range resp.Kvs {
			log.Println("etcd discovery start connect:", string(data.Key))
			var ed ETCDServiceDesc
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
						var ed ETCDServiceDesc
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
							log.Println(fmt.Printf("del old node. id:%v", ed.Id))
						}
						nodeCreator(multiNode, &ed)
					case clientv3.EventTypeDelete:
						nodeID := getNodeId(string(ev.Kv.Key))
						if preNode := multiNode.GetNode(nodeID); preNode != nil {
							multiNode.DelNode(nodeID, serviceName)
							preNode.Stop()
							log.Println(fmt.Printf("del node. id:%v", nodeID))
						}
					}
				}
			}
		}
	}()
	return nil
}

// setServiceStartupTime 设置服务器开服时间
func setServiceStartupTime(zone int) {
	startupKey := genServiceZonePrefix(zone)
	resp, err := etcdDiscovery.etcdKV.Get(context.TODO(), startupKey)
	if err != nil {
		log.Println("etcd setServiceStartupTime error:", err)
		return
	}
	startupTime := uint64(0)
	if resp.Count > 0 {
		startupTime, _ = strconv.ParseUint(string(resp.Kvs[0].Value), 10, 64)
	} else {
		// 注册
		t := util.GetCurrentTimeMs()
		val := strconv.FormatUint(t, 10)
		err = etcdDiscovery.RegisterService(startupKey, val)
		if err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("etcd setServiceStartupTime error:%v", err)
			return
		}
		startupTime = t
	}
	logrus.Log(logrus.LogsSystem).Infof("etcd setServiceStartupTime success. startupKey:%v startupTime:%v\n", startupKey, startupTime)
}

func genServicePrefix(name string, zone int) string {
	//return servicePrefixKey + strconv.Itoa(zone) + "/" + name
	return servicePrefixKey + name
}

func genServiceZonePrefix(zone int) string {
	return servicePrefixKey + strconv.Itoa(zone)
}

func genDiscoveryServicePrefix(name string, zone int) string {
	if zone > 0 {
		return servicePrefixKey + name + "#" + strconv.Itoa(zone)
	}
	return servicePrefixKey + name + "#"
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
		err = fmt.Errorf("ParseServiceId sid invalid. sid:" + sid)
		return
	} else {
		strProp := strings.Split(str[1], "@")
		if len(strProp) < 3 {
			err = fmt.Errorf("ParseServiceId sid invalid. sid:" + sid)
			return
		} else {
			zone, err = util.StrToInt(strProp[0])
			if err != nil {
				return
			}
			typ, err = util.StrToInt(strProp[1])
			if err != nil {
				return
			}
			idx, err = util.StrToInt(strProp[2])
			if err != nil {
				return
			}
		}
	}
	return
}
