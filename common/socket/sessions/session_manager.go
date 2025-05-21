package sessions

import (
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/common/util"
	"log"
	"sync"
	"time"
)

type SessionManager struct {
	genKey        int
	lastTimeStamp uint64
	sequence      uint64
	sessionMap    sync.Map
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		lastTimeStamp: util.GetCurrentTimeMs(),
	}
}

func (sm *SessionManager) Add(s iface.ISession) {
	id := sm.genSessionId()
	s.SetId(id)
	if _, ok := sm.sessionMap.Load(id); ok {
		log.Panic("session id already exists. id:", id)
	}
	sm.sessionMap.Store(id, s)
}

func (sm *SessionManager) Get(sessionId uint64) (iface.ISession, bool) {
	val, ok := sm.sessionMap.Load(sessionId)
	if !ok {
		return nil, ok
	}
	return val.(iface.ISession), ok
}

func (sm *SessionManager) Remove(s iface.ISession) {
	sm.sessionMap.Delete(s.GetId())
}

func (sm *SessionManager) CloseAllSession() {
	sm.sessionMap.Range(func(key, value interface{}) bool {
		value.(iface.ISession).Close()
		return true
	})
}

func (sm *SessionManager) SetUuidCreateKey(genKey int) {
	sm.genKey = genKey
}

// 临时使用的id 只保证同一个server内唯一 一秒钟最大产生65536个
func (sm *SessionManager) genSessionId() uint64 {
	currentTimeStamp := uint64(time.Now().Unix())
	if sm.lastTimeStamp == 0 {
		sm.lastTimeStamp = currentTimeStamp
	}

	if sm.genKey > 0xff {
		return 0
	}
	if sm.lastTimeStamp == currentTimeStamp {
		sm.sequence++
		if sm.sequence > 0xffff {
			// 一秒内尝试的数量超过上限
			return 0
		}
	} else {
		sm.lastTimeStamp = currentTimeStamp
		sm.sequence = 1
	}

	var uid uint64 = 0
	// 前40位时间戳（秒）
	uid |= sm.lastTimeStamp << 24
	// 16位序列号
	uid |= uint64(sm.sequence&0xffff) << 8
	// 左后是逻辑ID（服务器id）
	uid |= uint64(sm.genKey & 0xff)
	return uid
}
