package service

import (
	"github.com/ljhe/scream/core/iface"
	"log"
	"runtime/debug"
	"sync"
)

// 事件列表容量
var queListSize = 20000

type MsgHandle struct {
	queList chan interface{}
	onError func(interface{})
	wg      sync.WaitGroup
}

func NewMsgHandle() iface.IMsgHandle {
	return &MsgHandle{
		queList: make(chan interface{}, queListSize),
		onError: func(data interface{}) {
			log.Printf("onError data:%v stack:%v \n", data, string(debug.Stack()))
			// 打印堆栈信息
			debug.PrintStack()
		},
	}
}

func GetMsgHandle() iface.IMsgHandle {
	handle := NewMsgHandle()
	handle.Start()
	return handle
}

func (m *MsgHandle) Start() iface.IMsgHandle {
	m.wg.Add(1)
	go func() {
		for {
			select {
			case msg := <-m.queList:
				switch f := msg.(type) {
				case func():
					f()
				}
			}
		}
		m.wg.Done()
	}()
	return m
}

func (m *MsgHandle) Stop() iface.IMsgHandle {
	return nil
}

func (m *MsgHandle) Wait() {
	m.wg.Wait()
}

func (m *MsgHandle) PostCb(cb func()) {
	if cb != nil {
		m.queList <- cb
	}
}
