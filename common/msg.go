package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
)

const (
	msgBodyLen     = 2 // body大小 2个字节
	msgIdLen       = 2 // 包id大小 2个字节
	msgChunkNumLen = 2 // 分片数量大小 2个字节
	msgChunkIdLen  = 2 // 分片id大小 2个字节
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

	msgLen := len(msgData)
	msgId := uint16(msgInfo.MsgId)
	// 计算分片数量
	chunkNum := msgLen/MsgMaxLen + 1
	//chunkNum := msgLen/50 + 1
	sendBytes := 0
	chunkId := 1

	for sendBytes < msgLen {
		remaining := msgLen - sendBytes
		chunkSize := MsgMaxLen
		//chunkSize := 50
		if remaining < chunkSize {
			chunkSize = remaining
		}

		//data := make([]byte, msgBodyLen+msgIdLen+msgChunkNumLen+msgChunkIdLen+chunkSize)
		// 使用内存池
		actualDataLen := msgBodyLen + msgIdLen + msgChunkNumLen + msgChunkIdLen + chunkSize
		data := MemoryPoolObj.Get(actualDataLen)
		// msgBodyLen
		binary.BigEndian.PutUint16(data, uint16(msgLen))
		// msgIdLen
		binary.BigEndian.PutUint16(data[msgBodyLen:], msgId)
		// chunkNumLen
		binary.BigEndian.PutUint16(data[msgBodyLen+msgIdLen:], uint16(chunkNum))
		// chunkIdLen
		binary.BigEndian.PutUint16(data[msgBodyLen+msgIdLen+msgChunkNumLen:], uint16(chunkId))
		// msgBody
		copy(data[msgBodyLen+msgIdLen+msgChunkNumLen+msgChunkIdLen:], msgData[sendBytes:sendBytes+chunkSize])
		// 使用内存池 会导致每次发送的包里都会有空数据 所以写入的时候只写入有效数据的部分
		err = WriteFull(writer, data[:actualDataLen])
		if err != nil {
			return err
		}
		sendBytes += chunkSize
		chunkId++
		MemoryPoolObj.Put(data)
	}
	return nil
}

// RcvPackageData 获取原始包数据
func RcvPackageData(reader io.Reader) ([]byte, uint16, error) {
	var bufMsg = []byte{}
	msgId := uint16(0)
	fId := uint16(0) // chunkId=1时的msgId
	receivedBytes := uint16(0)
	for {
		// msgBodyLen
		msgLen, err := readUint16(reader, msgBodyLen)
		if err != nil {
			return nil, 0, err
		}
		// msgId
		msgId, err = readUint16(reader, msgIdLen)
		if err != nil {
			return nil, 0, err
		}
		// chunkNum
		bufChunkNumUint16, err := readUint16(reader, msgChunkNumLen)
		if err != nil {
			return nil, 0, err
		}
		// chunkId
		bufChunkIdUint16, err := readUint16(reader, msgChunkIdLen)
		if err != nil {
			return nil, 0, err
		}
		if bufChunkIdUint16 == 1 {
			fId = msgId
		}

		if len(bufMsg) == 0 {
			bufMsg = make([]byte, msgLen)
		}
		remaining := msgLen - receivedBytes
		chunkSize := MsgMaxLen
		//chunkSize := 50
		if remaining < uint16(chunkSize) {
			chunkSize = int(remaining)
		}

		//buf := make([]byte, chunkSize)
		// 使用内存池
		buf := MemoryPoolObj.Get(chunkSize)
		// 使用内存池  分配的buf内存可能会大于实际数据长度 所以这里只读取有效数据的长度
		_, err = io.ReadFull(reader, buf[:chunkSize])
		if err != nil {
			log.Printf("readFull err:%v \n", err)
			return nil, 0, err
		}
		copy(bufMsg[receivedBytes:], buf)
		receivedBytes += uint16(chunkSize)
		if bufChunkIdUint16 >= bufChunkNumUint16 {
			break
		}
		if msgId != fId {
			break
		}
	}

	if len(bufMsg) != 0 {
		MemoryPoolObj.Put(bufMsg)
	}
	return bufMsg, msgId, nil
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

func readUint16(reader io.Reader, byteLen int) (uint16, error) {
	bt := make([]byte, byteLen)
	_, err := io.ReadFull(reader, bt)
	if err != nil {
		return 0, err
	}
	btUint16 := binary.BigEndian.Uint16(bt)
	return btUint16, nil
}
