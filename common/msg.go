package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

const (
	msgMaxLen = 2 // body大小 2个字节
	msgIdLen  = 2 // 包id大小 2个字节
)

var (
	systemMsgById  = map[int]*SystemMsg{}
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
	systemMsgById[sys.MsgId] = sys

	if _, ok := systemMsgByTyp[sys.typ]; ok {
		panic(fmt.Sprintf("msgType already registered. msgType: %s", sys.typ))
	}
	systemMsgByTyp[sys.typ] = sys
}

func MessageInfoById(msgId int) *SystemMsg {
	return systemMsgById[msgId]
}

func MessageInfoByMsg(msg interface{}) *SystemMsg {
	typ := reflect.TypeOf(msg)
	if typ.Kind() == reflect.Ptr {
		return systemMsgByTyp[typ.Elem()]
	}
	return systemMsgByTyp[typ]
}

func SendMessage(writer io.Writer, msg interface{}) (err error) {
	msgData, msgInfo, err := EncodeMessage(msg)
	if err != nil {
		return err
	}

	// body's len
	msgLen := len(msgData)
	msgId := uint16(msgInfo.MsgId)

	// 拼接head
	data := make([]byte, msgMaxLen+msgIdLen+msgLen)
	// msgMaxLen
	binary.BigEndian.PutUint16(data, uint16(msgLen))
	// msgIdLen
	binary.BigEndian.PutUint16(data[msgMaxLen:], msgId)

	// 拼接body
	if msgLen > 0 {
		copy(data[msgMaxLen+msgIdLen:], msgData)
	}
	return WriteFull(writer, data)
}

func WriteFull(writer io.Writer, buf []byte) error {
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
	msg, msgId, err := RcvPackageData(reader)
	if err != nil {
		return nil, err
	}

	bt, err := DecodeMessage(int(msgId), msg)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

// EncodeMessage 消息序列化
func EncodeMessage(msg interface{}) ([]byte, *SystemMsg, error) {
	info := MessageInfoByMsg(msg)
	if info == nil {
		return nil, nil, fmt.Errorf("msg not registered. msg: %v", msg)
	}
	bt, err := json.Marshal(msg)
	if err != nil {
		return nil, nil, err
	}
	return bt, info, nil
}

// DecodeMessage 消息反序列化
func DecodeMessage(msgId int, msg []byte) (interface{}, error) {
	sys := MessageInfoById(msgId)
	if sys == nil {
		return nil, fmt.Errorf("msgId not found. msgId: %d msg:%v", msgId, msg)
	}
	msgObj := reflect.New(sys.typ).Interface()
	err := json.Unmarshal(msg, msgObj)
	if err != nil {
		return nil, err
	}
	return msgObj, nil
}

// RcvPackageData 获取原始包数据
func RcvPackageData(reader io.Reader) ([]byte, uint16, error) {
	// msgLen
	bufMsgLen := make([]byte, msgMaxLen)
	_, err := io.ReadFull(reader, bufMsgLen)
	if err != nil {
		return nil, 0, err
	}
	if len(bufMsgLen) < msgMaxLen {
		return nil, 0, fmt.Errorf("msg too short. len:%v", len(bufMsgLen))
	}
	msgLen := binary.BigEndian.Uint16(bufMsgLen)

	// msgId
	bufIdLen := make([]byte, msgIdLen)
	_, err = io.ReadFull(reader, bufIdLen)
	if err != nil {
		return nil, 0, err
	}
	if len(bufIdLen) < msgIdLen {
		return nil, 0, fmt.Errorf("msg too short. len:%v", len(bufIdLen))
	}
	msgId := binary.BigEndian.Uint16(bufIdLen)

	bufMsg := make([]byte, msgLen)
	if msgLen > 0 {
		_, err = io.ReadFull(reader, bufMsg)
		if err != nil {
			return nil, 0, err
		}
		if len(bufMsg) < int(msgLen) {
			return nil, 0, fmt.Errorf("msg too short. len:%v", len(bufMsg))
		}
	}
	return bufMsg, msgId, nil
}
