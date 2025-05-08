package socket

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/common"
	"github.com/ljhe/scream/common/encryption"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/pbgo"
	"github.com/ljhe/scream/plugins/logrus"
	"io"
	"reflect"
	"sync"
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

var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 0)
	},
}

type TcpDataPacket struct {
}

type WsDataPacket struct {
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

	bt, err := DecodeMessage(msgId, msg)
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
		msgId:     msgInfo.ID,
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
		msg, msgId, err := RcvPackageDataByByte(bt)
		if err != nil {
			return nil, err
		}
		bt, err := DecodeMessage(msgId, msg)
		if err != nil {
			return nil, err
		}
		return bt, nil
	default:
		return nil, fmt.Errorf("WsDataPacket ReadMessage type not binary message. typ:%v", typ)
	}

}

func (w *WsDataPacket) SendMessage(s iface.ISession, msg interface{}) (err error) {
	conn, ok := s.Raw().(*websocket.Conn)
	if !ok || conn == nil {
		return fmt.Errorf("WsDataPacket SendMessage get websocket.Conn err")
	}
	msgData, msgInfo, err := EncodeMessage(msg)
	if err != nil {
		return err
	}
	msgDataLen := len(msgData)
	// todo 注意上层发包不要超过最大值 之后这里可以改成如果超过最大值 就分片发送
	opt := s.Node().(Option)
	if msgDataLen > opt.MaxMsgLen() {
		return fmt.Errorf("ws sendMessage too big. msgId=%v msglen=%v maxlen=%v", 1, msgDataLen, opt.MaxMsgLen())
	}
	mb := &msgBase{
		msgId:  msgInfo.ID,
		flagId: 1,
	}
	buf := mb.MarshalBytes(msgData)
	err = conn.WriteMessage(websocket.BinaryMessage, buf)
	return err
}

// RcvPackageData 获取原始包数据
func RcvPackageData(reader io.Reader) ([]byte, uint16, error) {
	mb := &msgBase{}
	bufMsg, err := mb.Unmarshal(reader)
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
func EncodeMessage(msg interface{}) ([]byte, *pbgo.MessageInfo, error) {
	info := pbgo.MessageInfoByMsg(msg)
	if info == nil {
		return nil, nil, ErrMsgIdNotFound
	}
	bt, err := info.Codec.Marshal(msg)
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("EncodeMessage Marshal err. msg:%v err:%v", msg, err)
		return nil, nil, err
	}
	return bt.([]byte), info, nil
}

// DecodeMessage 消息反序列化
func DecodeMessage(msgId uint16, msg []byte) (interface{}, error) {
	sys := pbgo.MessageInfoById(msgId)
	if sys == nil {
		return nil, fmt.Errorf("msgId not found. msgId: %d msg:%v", msgId, msg)
	}
	msgObj := reflect.New(sys.Type).Interface()
	err := sys.Codec.Unmarshal(msg, msgObj)
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("DecodeMessage Unmarshal err. msg:%v err:%v", msg, err)
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
		mb.Release(buf)
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
// 关于RSA加密
// RSA私钥只建议放在服务端 因为RSA私钥可以倒推出公钥 而公钥不可以倒推出私钥
// 所以对于客户端而言 只能去使用公钥来加密 而不能进行使用私钥来解密这类操作
// 所以加密 一般只是客户端到服务端来进行加密
// 如果客户端需要验证 推荐[服务端使用私钥对消息签名 客户端用公钥验证签名] 从而让客户端来验证消息的合法性
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
	mb.actualDataLen = int(msgBodyLen)
	bytes = bytes[MsgOptions.MsgBodyLen:]

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

	switch mb.flagId {
	case common.MsgEncryptionRSA:
		msgData, err = encryption.RSADecrypt(msgData, encryption.RSAWSPrivateKey)
	default:
		logrus.Log(logrus.LogsSystem).Errorf("msgBase flagId err. flagId: %d", mb.flagId)
		return
	}

	return msgData, err
}

// 仿照go底层crypto\tls\conn中的sync.Pool
func (mb *msgBase) Container() []byte {
	// 使用内存池
	if MsgOptions.Pool {
		//return mpool.GetMemoryPool(mpool.SystemMemoryPoolKey).Get(mb.actualDataLen)
		outBuf := bufPool.Get().([]byte)
		_, outBuf = sliceForAppend(outBuf[:0], mb.actualDataLen)
		return outBuf
	}
	return make([]byte, mb.actualDataLen)
}

func (mb *msgBase) Release(data []byte) {
	if MsgOptions.Pool {
		//mpool.GetMemoryPool(mpool.SystemMemoryPoolKey).Put(data)
		bufPool.Put(data[:0])
	}
}

func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}
