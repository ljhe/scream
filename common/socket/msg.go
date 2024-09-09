package socket

import (
	"common"
	"common/plugins/mpool"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var ErrMsgIdNotFound = errors.New("msgId not found")

var MsgOptions = struct {
	MsgBodyLen     uint16 // body大小 2个字节
	MsgIdLen       uint16 // 包id大小 2个字节
	MsgChunkNumLen uint16 // 分片数量大小 2个字节
	MsgChunkIdLen  uint16 // 分片id大小 2个字节
	Pool           bool   // 是否使用内存池
}{
	MsgBodyLen:     2,
	MsgIdLen:       2,
	MsgChunkNumLen: 2,
	MsgChunkIdLen:  2,
	Pool:           true,
}

var (
	systemMsgById  = map[int]*SystemMsg{}
	systemMsgByTyp = map[reflect.Type]*SystemMsg{}
)

type SystemMsg struct {
	MsgId int
	Msg   []byte
	typ   reflect.Type
}

type msgBase struct {
	msgLen        uint16
	msgId         uint16
	chunkNum      uint16
	chunkId       uint16
	sendBytes     int
	actualDataLen int
	chunkSize     int
	receivedBytes uint16
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
	mb := &msgBase{
		msgLen:    uint16(len(msgData)),
		msgId:     uint16(msgInfo.MsgId),
		chunkNum:  uint16(msgLen/common.MsgMaxLen + 1), // 计算分片数量
		chunkId:   1,
		sendBytes: 0,
	}

	for mb.sendBytes < int(mb.msgLen) {
		data := mb.Marshal(msgData)
		// 如果使用内存池 会导致每次发送的包里都会有空数据 所以写入的时候只写入有效数据的部分
		err = WriteFull(writer, data[:mb.actualDataLen])
		if err != nil {
			return err
		}
		mb.sendBytes += mb.chunkSize
		mb.chunkId++
		mpool.GetMemoryPool(mpool.TCPMemoryPoolKey).Put(data)
		mb.Release(data)
	}
	return nil
}

// RcvPackageData 获取原始包数据
func RcvPackageData(reader io.Reader) ([]byte, uint16, error) {
	mb := &msgBase{}
	bufMsg, err := mb.Unmarshal(reader)
	mb.Release(bufMsg)
	return bufMsg, mb.msgId, err
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
		return nil, nil, ErrMsgIdNotFound
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

func readUint16(reader io.Reader, byteLen uint16) (uint16, error) {
	bt := make([]byte, byteLen)
	_, err := io.ReadFull(reader, bt)
	if err != nil {
		return 0, err
	}
	btUint16 := binary.BigEndian.Uint16(bt)
	return btUint16, nil
}

func (mb *msgBase) Marshal(msgData []byte) []byte {
	remaining := int(mb.msgLen) - mb.sendBytes
	mb.chunkSize = common.MsgMaxLen
	if remaining < mb.chunkSize {
		mb.chunkSize = remaining
	}
	mb.actualDataLen = int(MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.MsgChunkNumLen+MsgOptions.MsgChunkIdLen) + mb.chunkSize
	data := mb.Container()
	// msgBodyLen
	binary.BigEndian.PutUint16(data, uint16(mb.msgLen))
	// msgIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen:], mb.msgId)
	// chunkNumLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen:], mb.chunkNum)
	// chunkIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.MsgChunkNumLen:], mb.chunkId)
	// msgBody
	copy(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.MsgChunkNumLen+MsgOptions.MsgChunkIdLen:],
		msgData[mb.sendBytes:mb.sendBytes+mb.chunkSize])
	return data
}

func (mb *msgBase) Unmarshal(reader io.Reader) ([]byte, error) {
	var bufMsg = []byte{}
	var err error
	var fId uint16 // chunkId=1时的msgId
	for {
		// msgBodyLen
		mb.msgLen, err = readUint16(reader, MsgOptions.MsgBodyLen)
		if err != nil {
			return nil, err
		}
		// msgId
		mb.msgId, err = readUint16(reader, MsgOptions.MsgIdLen)
		if err != nil {
			return nil, err
		}
		// chunkNum
		mb.chunkNum, err = readUint16(reader, MsgOptions.MsgChunkNumLen)
		if err != nil {
			return nil, err
		}
		// chunkId
		mb.chunkId, err = readUint16(reader, MsgOptions.MsgChunkIdLen)
		if err != nil {
			return nil, err
		}
		if mb.chunkId == 1 {
			fId = mb.msgId
		}

		if len(bufMsg) == 0 {
			bufMsg = make([]byte, mb.msgLen)
		}
		remaining := mb.msgLen - mb.receivedBytes
		mb.chunkSize = common.MsgMaxLen
		if remaining < uint16(mb.chunkSize) {
			mb.chunkSize = int(remaining)
		}

		mb.actualDataLen = mb.chunkSize
		buf := mb.Container()
		// 如果使用内存池  分配的buf内存可能会大于实际数据长度 所以这里只读取有效数据的长度
		_, err = io.ReadFull(reader, buf[:mb.chunkSize])
		if err != nil {
			return nil, err
		}
		copy(bufMsg[mb.receivedBytes:], buf)
		mb.receivedBytes += uint16(mb.chunkSize)
		if mb.chunkId >= mb.chunkNum {
			break
		}
		if mb.msgId != fId {
			break
		}
	}
	return bufMsg, err
}

func (mb *msgBase) Container() []byte {
	// 使用内存池
	if MsgOptions.Pool {
		return mpool.GetMemoryPool(mpool.TCPMemoryPoolKey).Get(mb.actualDataLen)
	}
	return make([]byte, mb.actualDataLen)
}

func (mb *msgBase) Release(data []byte) {
	if MsgOptions.Pool {
		data = []byte{}
		mpool.GetMemoryPool(mpool.TCPMemoryPoolKey).Put(data)
	}
}
