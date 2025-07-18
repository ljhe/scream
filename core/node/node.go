package node

import (
	"context"
	"github.com/ljhe/scream/3rd/logrus"
)

type Node struct {
	Id string
	Ty string
}

func (n *Node) ID() string {
	return n.Id
}

func (n *Node) Type() string {
	return n.Ty
}

func (n *Node) Init(ctx context.Context) {

}

func (n *Node) Exit() {
	logrus.Infof("node %s exiting", n.Id)
}
