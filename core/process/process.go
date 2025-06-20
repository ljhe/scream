package process

import (
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"os"
	"os/signal"
	"syscall"
)

type Process struct {
	Id    string
	Pid   int
	Host  string
	Nodes []iface.INetNode
}

func NewProcess(id, host string) iface.IProcess {
	return &Process{
		Id:    id,
		Pid:   os.Getpid(),
		Host:  host,
		Nodes: make([]iface.INetNode, 0),
	}
}

func (p *Process) ID() string {
	return p.Id
}

func (p *Process) PID() int {
	return p.Pid
}

func (p *Process) GetHost() string {
	return p.Host
}

func (p *Process) GetNode(id string) iface.INetNode {
	//TODO implement me
	panic("implement me")
}

func (p *Process) GetAllNodes() []iface.INetNode {
	//TODO implement me
	panic("implement me")
}

func (p *Process) Start() {}

func (p *Process) WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}

func (p *Process) Stop() {
	fmt.Println("process stop")
}

func (p *Process) RegisterNode(node iface.INetNode) error {
	return nil
}
