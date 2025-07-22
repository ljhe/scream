package msg

import (
	"encoding/binary"
	"errors"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/utils/encryption"
	"io"
	"sync"
)

const MsgMaxLen = 1024 * 40 // 40k(发送和接受字节最大数量)

const (
	MsgEncryptionNone = iota
	MsgEncryptionRSA
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

type MsgBase struct {
	MsgId         uint16
	MsgLen        uint16 // 总长度
	ChunkNum      uint16 // 分片数量
	ChunkId       uint16 // 当前片id
	SendBytes     int    // 已发送长度
	ActualDataLen int    // 实际数据长度
	ChunkSize     int    // 分片长度
	ReceivedBytes uint16 // 接受长度
	FlagId        uint16 // 加密方式
}

var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 0)
	},
}

// RcvPackageData 获取原始包数据
func RcvPackageData(reader io.Reader) ([]byte, uint16, error) {
	mb := &MsgBase{}
	bufMsg, err := mb.Unmarshal(reader)
	return bufMsg, mb.MsgId, err
}

// RcvPackageDataByByte 通过 []byte 获取原始包数据
func RcvPackageDataByByte(bt []byte) ([]byte, uint16, error) {
	mb := &MsgBase{}
	bufMsg, err := mb.UnmarshalBytes(bt)
	return bufMsg, mb.MsgId, err
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

func readUint16(reader io.Reader, byteLen uint16) (uint16, error) {
	bt := make([]byte, byteLen)
	_, err := io.ReadFull(reader, bt)
	if err != nil {
		return 0, err
	}
	btUint16 := binary.BigEndian.Uint16(bt)
	return btUint16, nil
}

func (mb *MsgBase) Marshal(msgData []byte) []byte {
	remaining := int(mb.MsgLen) - mb.SendBytes
	mb.ChunkSize = MsgMaxLen
	if remaining < mb.ChunkSize {
		mb.ChunkSize = remaining
	}
	mb.ActualDataLen = int(MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.MsgChunkNumLen+MsgOptions.MsgChunkIdLen) + mb.ChunkSize
	data := mb.Container()
	// msgBodyLen
	binary.BigEndian.PutUint16(data, mb.MsgLen)
	// msgIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen:], mb.MsgId)
	// chunkNumLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen:], mb.ChunkNum)
	// chunkIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.MsgChunkNumLen:], mb.ChunkId)
	// msgBody
	copy(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.MsgChunkNumLen+MsgOptions.MsgChunkIdLen:],
		msgData[mb.SendBytes:mb.SendBytes+mb.ChunkSize])
	return data
}

func (mb *MsgBase) Unmarshal(reader io.Reader) ([]byte, error) {
	var bufMsg []byte
	var err error
	var fId uint16 // chunkId=1时的msgId
	for {
		// msgBodyLen
		mb.MsgLen, err = readUint16(reader, MsgOptions.MsgBodyLen)
		if err != nil {
			return nil, err
		}
		// msgId
		mb.MsgId, err = readUint16(reader, MsgOptions.MsgIdLen)
		if err != nil {
			return nil, err
		}
		// chunkNum
		mb.ChunkNum, err = readUint16(reader, MsgOptions.MsgChunkNumLen)
		if err != nil {
			return nil, err
		}
		// chunkId
		mb.ChunkId, err = readUint16(reader, MsgOptions.MsgChunkIdLen)
		if err != nil {
			return nil, err
		}
		if mb.ChunkId == 1 {
			fId = mb.MsgId
		}

		if len(bufMsg) == 0 {
			bufMsg = make([]byte, mb.MsgLen)
		}
		remaining := mb.MsgLen - mb.ReceivedBytes
		mb.ChunkSize = MsgMaxLen
		if remaining < uint16(mb.ChunkSize) {
			mb.ChunkSize = int(remaining)
		}

		mb.ActualDataLen = mb.ChunkSize
		buf := mb.Container()
		// 如果使用内存池  分配的buf内存可能会大于实际数据长度 所以这里只读取有效数据的长度
		_, err = io.ReadFull(reader, buf[:mb.ChunkSize])
		if err != nil {
			return nil, err
		}
		copy(bufMsg[mb.ReceivedBytes:], buf)
		mb.Release(buf)
		mb.ReceivedBytes += uint16(mb.ChunkSize)
		if mb.ChunkId >= mb.ChunkNum {
			break
		}
		if mb.MsgId != fId {
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
func (mb *MsgBase) MarshalBytes(msgData []byte) []byte {
	msgDataLen := len(msgData)
	data := make([]byte, MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.FlagIdLen+uint16(msgDataLen))

	// header
	// MsgBodyLen
	binary.BigEndian.PutUint16(data, uint16(msgDataLen))
	// MsgIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen:], mb.MsgId)
	// FlagIdLen
	binary.BigEndian.PutUint16(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen:], mb.FlagId)

	// body
	if msgDataLen > 0 {
		copy(data[MsgOptions.MsgBodyLen+MsgOptions.MsgIdLen+MsgOptions.FlagIdLen:], msgData)
	}
	return data
}

// UnmarshalBytes 数据格式 package = MsgBodyLen + MsgIdLen + FlagIdLen + msgData
func (mb *MsgBase) UnmarshalBytes(bytes []byte) (msgData []byte, err error) {
	var msgBodyLen uint16 // 请求长度
	if len(bytes) < int(MsgOptions.MsgBodyLen) {
		logrus.Errorf("MsgBase UnmarshalBytes MsgBodyLen err. bytes'len: %d", len(bytes))
		return
	}
	msgBodyLen = binary.BigEndian.Uint16(bytes)
	mb.ActualDataLen = int(msgBodyLen)
	bytes = bytes[MsgOptions.MsgBodyLen:]

	if len(bytes) < int(MsgOptions.MsgIdLen) {
		logrus.Errorf("MsgBase UnmarshalBytes MsgIdLen err. bytes'len: %d", len(bytes))
		return
	}
	mb.MsgId = binary.BigEndian.Uint16(bytes)
	bytes = bytes[MsgOptions.MsgIdLen:]

	if len(bytes) < int(MsgOptions.FlagIdLen) {
		logrus.Errorf("MsgBase UnmarshalBytes FlagIdLen err. bytes'len: %d", len(bytes))
		return
	}
	mb.FlagId = binary.BigEndian.Uint16(bytes)
	msgData = bytes[MsgOptions.FlagIdLen:]

	switch mb.FlagId {
	case MsgEncryptionNone:
		break
	case MsgEncryptionRSA:
		msgData, err = encryption.RSADecrypt(msgData, encryption.RSAWSPrivateKey)
	default:
		logrus.Errorf("MsgBase flagId err. flagId: %d", mb.FlagId)
		return
	}

	return msgData, err
}

// 仿照go底层crypto\tls\conn中的sync.Pool
func (mb *MsgBase) Container() []byte {
	// 使用内存池
	if MsgOptions.Pool {
		outBuf := bufPool.Get().([]byte)
		_, outBuf = sliceForAppend(outBuf[:0], mb.ActualDataLen)
		return outBuf
	}
	return make([]byte, mb.ActualDataLen)
}

func (mb *MsgBase) Release(data []byte) {
	if MsgOptions.Pool {
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
