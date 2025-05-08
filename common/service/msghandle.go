package service

import (
	"github.com/ljhe/scream/common"
	"github.com/panjf2000/ants/v2"
	"log"
	"runtime/debug"
	"sync"
)

// 事件列表容量
var queListSize = 20000

type MsgHandle struct {
	queList  chan interface{}
	onError  func(interface{})
	workPool *ants.Pool
	wg       sync.WaitGroup
}

func NewMsgHandle() common.IMsgHandle {
	return &MsgHandle{
		queList: make(chan interface{}, queListSize),
		onError: func(data interface{}) {
			log.Printf("onError data:%v stack:%v \n", data, string(debug.Stack()))
			// 打印堆栈信息
			debug.PrintStack()
		},
	}
}

func GetMsgHandle(size int) common.IMsgHandle {
	handle := NewMsgHandle()
	if size > 0 {
		handle.SetWorkPool(size)
	}
	handle.Start()
	return handle
}

func (m *MsgHandle) Start() common.IMsgHandle {
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

func (m *MsgHandle) Stop() common.IMsgHandle {
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

func (m *MsgHandle) SetWorkPool(size int) {
	p, err := ants.NewPool(size)
	if err != nil {
		return
	}
	m.workPool = p
}
