package socket

import (
	"common/iface"
	"common/util"
	"log"
	"sync"
	"time"
)

type SessionManager interface {
	Add(s iface.ISession)
	SetUuidCreateKey(genKey int)
}

type NetSessionManager struct {
	genKey        int
	lastTimeStamp uint64
	sequence      uint64
	sessionMap    sync.Map
}

func NewNetSessionManager() *NetSessionManager {
	return &NetSessionManager{
		lastTimeStamp: util.GetCurrentTimeMs(),
	}
}

func (n *NetSessionManager) Add(s iface.ISession) {
	id := n.genSessionId()
	s.SetId(id)
	if _, ok := n.sessionMap.Load(id); ok {
		log.Panic("session id already exists. id:", id)
	}
	n.sessionMap.Store(id, s)
}

func (n *NetSessionManager) SetUuidCreateKey(genKey int) {
	n.genKey = genKey
}

// 临时使用的id 只保证同一个server内唯一 一秒钟最大产生65536个
func (n *NetSessionManager) genSessionId() uint64 {
	currentTimeStamp := uint64(time.Now().Unix())
	if n.lastTimeStamp == 0 {
		n.lastTimeStamp = currentTimeStamp
	}

	if n.genKey > 0xff {
		return 0
	}
	if n.lastTimeStamp == currentTimeStamp {
		n.sequence++
		if n.sequence > 0xffff {
			// 一秒内尝试的数量超过上限
			return 0
		}
	} else {
		n.lastTimeStamp = currentTimeStamp
		n.sequence = 1
	}

	var uid uint64 = 0
	// 前40位时间戳（秒）
	uid |= n.lastTimeStamp << 24
	// 16位序列号
	uid |= uint64(n.sequence&0xffff) << 8
	// 左后是逻辑ID（服务器id）
	uid |= uint64(n.genKey & 0xff)
	return uid
}
