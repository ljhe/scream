package socket

import (
	"github.com/ljhe/scream/common/config"
	"reflect"
	"sync"
)

type NetServerNodeProperty struct {
	addr  string //
	name  string // 服务器名称
	zone  int    // 服务器区号
	typ   int    // 服务器类型
	index int    // 服务器区内的编号
}

func (n *NetServerNodeProperty) SetAddr(addr string) {
	n.addr = addr
}

func (n *NetServerNodeProperty) GetAddr() string {
	return n.addr
}

func (n *NetServerNodeProperty) SetName(s string) {
	n.name = s
}

func (n *NetServerNodeProperty) GetName() string {
	return n.name
}

func (n *NetServerNodeProperty) SetZone(z int) {
	n.zone = z
}

func (n *NetServerNodeProperty) GetZone() int {
	return n.zone
}

func (n *NetServerNodeProperty) SetServerTyp(t int) {
	n.typ = t
}

func (n *NetServerNodeProperty) GetServerTyp() int {
	return n.typ
}

func (n *NetServerNodeProperty) SetIndex(i int) {
	n.index = i
}

func (n *NetServerNodeProperty) GetIndex() int {
	return n.index
}

func (n *NetServerNodeProperty) SetServerNodeProperty() {
	n.SetServerTyp(config.SConf.Node.Typ)
	n.SetZone(config.SConf.Node.Zone)
	n.SetIndex(config.SConf.Node.Index)
}

// NetContextSet 用来记录上下文数据
type NetContextSet struct {
	data map[interface{}]keyValueData
	mu   sync.RWMutex
}

type keyValueData struct {
	key   interface{}
	value interface{}
}

func (nc *NetContextSet) SetContextData(key, val interface{}) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	if nc.data == nil {
		nc.data = make(map[interface{}]keyValueData)
	}
	nc.data[key] = keyValueData{key: key, value: val}
}

func (nc *NetContextSet) GetContextData(key interface{}) (interface{}, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	if data, ok := nc.data[key]; ok {
		return data.value, true
	}
	return nil, false
}

func (nc *NetContextSet) RawContextData(key interface{}, ptr interface{}) bool {
	val, ok := nc.GetContextData(key)
	if !ok {
		return false
	}
	switch outValue := ptr.(type) {
	case *string:
		*outValue = val.(string)
	default:
		v := reflect.Indirect(reflect.ValueOf(ptr))
		if val != nil {
			v.Set(reflect.ValueOf(val))
		}
	}
	return true
}
