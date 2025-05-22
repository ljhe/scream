package socket

import (
	"reflect"
	"sync"
)

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
