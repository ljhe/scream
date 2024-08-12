package socket

import "sync"

type NetRuntimeTag struct {
	CloseFlag bool
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
