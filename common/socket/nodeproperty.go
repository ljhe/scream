package socket

import (
	"github.com/ljhe/scream/common/config"
	"reflect"
	"sync"
)

type ServerNodeProperty struct {
	addr  string //
	name  string // 服务器名称
	zone  int    // 服务器区号
	typ   int    // 服务器类型
	index int    // 服务器区内的编号
}

func (n *ServerNodeProperty) SetAddr(addr string) {
	n.addr = addr
}

func (n *ServerNodeProperty) GetAddr() string {
	return n.addr
}

func (n *ServerNodeProperty) SetName(s string) {
	n.name = s
}

func (n *ServerNodeProperty) GetName() string {
	return n.name
}

func (n *ServerNodeProperty) SetZone(z int) {
	n.zone = z
}

func (n *ServerNodeProperty) GetZone() int {
	return n.zone
}

func (n *ServerNodeProperty) SetServerTyp(t int) {
	n.typ = t
}

func (n *ServerNodeProperty) GetServerTyp() int {
	return n.typ
}

func (n *ServerNodeProperty) SetIndex(i int) {
	n.index = i
}

func (n *ServerNodeProperty) GetIndex() int {
	return n.index
}

func (n *ServerNodeProperty) SetServerNodeProperty() {
	n.SetServerTyp(config.SConf.Node.Typ)
	n.SetZone(config.SConf.Node.Zone)
	n.SetIndex(config.SConf.Node.Index)
}

// ContextSet 用来记录上下文数据
type ContextSet struct {
	data map[interface{}]keyValueData
	mu   sync.RWMutex
}

type keyValueData struct {
	key   interface{}
	value interface{}
}

func (nc *ContextSet) SetContextData(key, val interface{}) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	if nc.data == nil {
		nc.data = make(map[interface{}]keyValueData)
	}
	nc.data[key] = keyValueData{key: key, value: val}
}

func (nc *ContextSet) GetContextData(key interface{}) (interface{}, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	if data, ok := nc.data[key]; ok {
		return data.value, true
	}
	return nil, false
}

func (nc *ContextSet) RawContextData(key interface{}, ptr interface{}) bool {
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
