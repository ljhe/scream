package socket

import (
	"common"
	"common/iface"
	"common/plugins/logrus"
	"common/plugins/mpool"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
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
	FlagIdLen      uint16 // 加密方式
}{
	MsgBodyLen:     2,
	MsgIdLen:       2,
	MsgChunkNumLen: 2,
	MsgChunkIdLen:  2,
	Pool:           true,
	FlagIdLen:      2,
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
	flagId        uint16
}

type TcpDataPacket struct {
}

type WsDataPacket struct {
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

func (t *TcpDataPacket) ReadMessage(s iface.ISession) (interface{}, error) {
	reader, ok := s.Raw().(io.Reader)
	if !ok || reader == nil {
		return nil, fmt.Errorf("TcpDataPacket ReadMessage get io.Reader err")
	}
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

func (t *TcpDataPacket) SendMessage(s iface.ISession, msg interface{}) (err error) {
	writer, ok := s.Raw().(io.Writer)
	if !ok || writer == nil {
		return fmt.Errorf("TcpDataPacket SendMessage get io.Writer err")
	}
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

func (w *WsDataPacket) ReadMessage(s iface.ISession) (interface{}, error) {
	conn, ok := s.Raw().(*websocket.Conn)
	if !ok || conn == nil {
		return nil, fmt.Errorf("WsDataPacket ReadMessage get websocket.Conn err")
	}
	typ, bt, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("WsDataPacket ReadMessage ReadMessage err:%v", err)
	}

	switch typ {
	case websocket.BinaryMessage:
		msg, _, err := RcvPackageDataByByte(bt)

		// 测试 直接返回数据
		err = w.SendMessage(s, msg)
		return msg, err
	default:
		return nil, fmt.Errorf("WsDataPacket ReadMessage type not binary message. typ:%v", typ)
	}

}

func (w *WsDataPacket) SendMessage(s iface.ISession, msg interface{}) (err error) {
	conn, ok := s.Raw().(*websocket.Conn)
	if !ok || conn == nil {
		return fmt.Errorf("WsDataPacket SendMessage get websocket.Conn err")
	}
	mb := &msgBase{}
	buf := mb.MarshalBytes(msg.([]byte))
	err = conn.WriteMessage(websocket.BinaryMessage, buf)
	return err
}

// RcvPackageData 获取原始包数据
func RcvPackageData(reader io.Reader) ([]byte, uint16, error) {
	mb := &msgBase{}
	bufMsg, err := mb.Unmarshal(reader)
	mb.Release(bufMsg)
	return bufMsg, mb.msgId, err
}

// RcvPackageDataByByte 通过 []byte 获取原始包数据
func RcvPackageDataByByte(bt []byte) ([]byte, uint16, error) {
	mb := &msgBase{}
	bufMsg, err := mb.UnmarshalBytes(bt)
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
	binary.BigEndian.PutUint16(data, mb.msgLen)
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

// MarshalBytes 数据格式 package = MsgBodyLen + MsgIdLen + FlagIdLen + msgData
func (mb *msgBase) MarshalBytes(msgData []byte) []byte {
	msgDataLen := len(msgData)
	data := make([]byte, MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.FlagIdLen+uint16(msgDataLen))

	// header
	// MsgBodyLen
	binary.BigEndian.PutUint16(data, uint16(msgDataLen))
	// MsgIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen:], mb.msgId)
	// FlagIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen:], mb.flagId)

	// body
	if msgDataLen > 0 {
		copy(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.FlagIdLen:], msgData)
	}
	return data
}

// UnmarshalBytes 数据格式 package = MsgBodyLen + MsgIdLen + FlagIdLen + msgData
func (mb *msgBase) UnmarshalBytes(bytes []byte) (msgData []byte, err error) {
	var msgBodyLen uint16 // 请求长度

	if len(bytes) < int(MsgOptions.MsgBodyLen) {
		logrus.Log(logrus.LogsSystem).Errorf("msgBase UnmarshalBytes MsgBodyLen err. bytes'len: %d", len(bytes))
		return
	}
	msgBodyLen = binary.BigEndian.Uint16(bytes)
	bytes = bytes[MsgOptions.MsgBodyLen:]
	if msgBodyLen > 0 {
		if len(bytes) < int(MsgOptions.MsgIdLen) {
			logrus.Log(logrus.LogsSystem).Errorf("msgBase UnmarshalBytes MsgIdLen err. bytes'len: %d", len(bytes))
			return
		}
		mb.msgId = binary.BigEndian.Uint16(bytes)
		bytes = bytes[MsgOptions.MsgIdLen:]

		if len(bytes) < int(MsgOptions.FlagIdLen) {
			logrus.Log(logrus.LogsSystem).Errorf("msgBase UnmarshalBytes FlagIdLen err. bytes'len: %d", len(bytes))
			return
		}
		mb.flagId = binary.BigEndian.Uint16(bytes)
		msgData = bytes[MsgOptions.FlagIdLen:]
	}
	return msgData, nil
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
		mpool.GetMemoryPool(mpool.TCPMemoryPoolKey).Put(data)
	}
}
