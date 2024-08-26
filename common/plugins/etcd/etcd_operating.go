package plugins

import (
	"common"
	"common/iface"
	"common/util"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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

func ETCDRegister(node iface.INetNode) {
	property := node.(common.ServerNodeProperty)
	ed := &ETCDServiceDesc{
		Id:    genServiceId(property),
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
		return
	}
	if resp.Count > 0 {
		fmt.Println("etcd register error: service already exist. etcdKey:", etcdKey)
		return
	}

	// 注册
	err = etcdDiscovery.RegisterService(etcdKey, ed.String())
	if err != nil {
		log.Println("etcd register error:", err)
		return
	}
	etcdDiscovery.WatchServices(etcdKey, *ed)
	setServiceStartupTime(property.GetZone())
	fmt.Println("etcd register success:", ed.Id)
}

func genServiceId(prop common.ServerNodeProperty) string {
	return fmt.Sprintf("%s#%d@%d@%d",
		prop.GetName(),
		prop.GetZone(),
		prop.GetServerTyp(),
		prop.GetIndex(),
	)
}

func genServicePrefix(name string, zone int) string {
	//return servicePrefixKey + strconv.Itoa(zone) + "/" + name
	return servicePrefixKey + name
}

func genServiceZonePrefix(zone int) string {
	return servicePrefixKey + strconv.Itoa(zone)
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
			log.Println("etcd setServiceStartupTime error:", err)
			return
		}
		startupTime = t
	}
	fmt.Printf("etcd setServiceStartupTime success. startupKey:%v startupTime:%v\n", startupKey, startupTime)
}
