package common

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

var (
	systemMsgById  = map[int]reflect.Type{}
	systemMsgByTyp = map[reflect.Type]*SystemMsg{}
)

type SystemMsg struct {
	MsgId int
	Msg   []byte
	typ   reflect.Type
}

func RegisterSystemMsg(sys *SystemMsg) {
	if _, ok := systemMsgById[sys.MsgId]; ok {
		panic(fmt.Sprintf("msgId already registered. msgId: %d", sys.MsgId))
	}
	systemMsgById[sys.MsgId] = sys.typ

	if _, ok := systemMsgByTyp[sys.typ]; ok {
		panic(fmt.Sprintf("msgType already registered. msgType: %s", sys.typ))
	}
	systemMsgByTyp[sys.typ] = sys
}

func MessageInfoByMsg(msg interface{}) *SystemMsg {
	typ := reflect.TypeOf(msg)
	if typ.Kind() == reflect.Ptr {
		return systemMsgByTyp[typ.Elem()]
	}
	return systemMsgByTyp[typ]
}

func SendMessage(writer io.Writer, msg interface{}) (err error) {
	bt, err := EncodeMessage(msg)
	if err != nil {
		return err
	}
	return Write(writer, bt)
}

func Write(writer io.Writer, buf []byte) error {
	total := len(buf)
	for pos := 0; pos < total; {
		n, err := writer.Write(buf[pos:])
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}

func ReadMessage(reader io.Reader, maxMsgLen int) (interface{}, error) {
	msg := make([]byte, maxMsgLen)
	n, err := reader.Read(msg)
	if err != nil {
		return nil, err
	}
	if n == maxMsgLen {
		return nil, fmt.Errorf("msg too long. maxMsgLen: %d", maxMsgLen)
	}
	bt, err := DecodeMessage(msg[:n])
	if err != nil {
		return nil, err
	}
	return bt, nil
}

// EncodeMessage 消息序列化
func EncodeMessage(msg interface{}) ([]byte, error) {
	info := MessageInfoByMsg(msg)
	if info == nil {
		return nil, fmt.Errorf("msg not registered. msg: %v", msg)
	}
	bt, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	info.Msg = bt
	b, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// DecodeMessage 消息反序列化
func DecodeMessage(msg []byte) (interface{}, error) {
	data := &SystemMsg{}
	err := json.Unmarshal(msg, data)
	if err != nil {
		return nil, err
	}
	typ := systemMsgById[data.MsgId]
	if typ == nil {
		return nil, fmt.Errorf("msgId not found. msgId: %d msg:%v", data.MsgId, msg)
	}
	msgObj := reflect.New(typ).Interface()
	err = json.Unmarshal(data.Msg, msgObj)
	if err != nil {
		return nil, err
	}
	return msgObj, nil
}
