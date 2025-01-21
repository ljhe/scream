package pbgo

import (
	"fmt"
	"reflect"
)

type Codec interface {
}

type MessageInfo struct {
	Codec Codec
	Type  reflect.Type
	ID    int
}

var (
	messageByID   = map[int]*MessageInfo{}
	messageByType = map[reflect.Type]*MessageInfo{}
	messageByName = map[string]*MessageInfo{}
)

var registerCodec Codec // 后续有别的解析部分这边可以添加

func GetCodec() Codec {
	return registerCodec
}

func RegisterMessageInfo(info *MessageInfo) {
	// 注册时统一为非指针类型
	if info.Type.Kind() == reflect.Ptr {
		info.Type = info.Type.Elem()
	}

	if info.ID == 0 {
		panic(fmt.Sprintf("message ID invalid:%v", info.Type.Name()))
	}

	if _, ok := messageByID[info.ID]; ok {
		panic(fmt.Sprintf("message ID:%v already registered", info.ID))
	} else {
		messageByID[info.ID] = info
	}

	if _, ok := messageByType[info.Type]; ok {
		panic(fmt.Sprintf("message Type:%v already registered", info.Type))
	} else {
		messageByType[info.Type] = info
	}

	if _, ok := messageByName[info.Type.Name()]; ok {
		panic(fmt.Sprintf("message Name:%v already registered", info.Type))
	} else {
		messageByName[info.Type.Name()] = info
	}
}

func MessageInfoById(msgId int) *MessageInfo {
	return messageByID[msgId]
}

func MessageInfoByMsg(msg interface{}) *MessageInfo {
	msgType := reflect.TypeOf(msg)
	if msgType.Kind() == reflect.Ptr {
		return messageByType[msgType.Elem()]
	} else {
		return messageByType[msgType]
	}
}

func MessageInfoByName(name string) *MessageInfo {
	return messageByName[name]
}

func MessageToString(msg interface{}) string {
	if msg == nil {
		return ""
	}
	if str, ok := msg.(interface{ String() string }); ok {
		return str.String()
	}
	return ""
}
