package socket

import (
	"sync"
	"sync/atomic"
)

type NetRuntimeTag struct {
	CloseFlag bool
	StopWg    sync.WaitGroup
	runState  int64
	mu        sync.Mutex
}

func (n *NetRuntimeTag) SetCloseFlag(b bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.CloseFlag = b
}

func (n *NetRuntimeTag) GetCloseFlag() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.CloseFlag
}

func (n *NetRuntimeTag) SetRunState(b bool) {
	if b {
		atomic.StoreInt64(&n.runState, 1)
	} else {
		atomic.StoreInt64(&n.runState, 0)
	}
}

func (n *NetRuntimeTag) GetRunState() bool {
	return atomic.LoadInt64(&n.runState) == 1
}
