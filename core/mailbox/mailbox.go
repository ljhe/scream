package mailbox

import (
	"errors"
	"fmt"
	"log"
)

type MailBox struct {
	Id    int
	Queue chan interface{}
	Close chan struct{}
}

func NewMailBox(id int) *MailBox {
	mb := &MailBox{
		Id:    id,
		Queue: make(chan interface{}),
		Close: make(chan struct{}),
	}

	go mb.Start()
	return mb
}

func (mb *MailBox) Start() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("mailbox panic for player %v: %v", mb.Id, r)
		}
	}()

	for {
		select {
		case msg := <-mb.Queue:
			fmt.Println(msg)
		case <-mb.Close:
			return nil
		}
	}
}

func (mb *MailBox) Stop() {
	close(mb.Close)
	close(mb.Queue)
}

func (mb *MailBox) Push(msg interface{}) error {
	select {
	case mb.Queue <- msg:
		return nil
	default:
		return errors.New("mailbox is full")
	}
}
