package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"sync"
)

var (
	ErrNodeRegisterInvalidParam = errors.New("system register invalid parameter")
	ErrNodeRegisterRepeat       = errors.New("system register node repeat")
	ErrNodeRegisterUnique       = errors.New("system register unique type actor")
)

type System struct {
	addressbook *AddressBook
	nodeidmap   map[string]iface.INode
	loader      iface.INodeLoader

	processID   string
	processIP   string
	processPort int

	sync.RWMutex
}

func BuildSystemWithOption(id, ip string, port int, loader iface.INodeLoader) iface.ISystem {

	sys := &System{
		nodeidmap:   make(map[string]iface.INode),
		processID:   id,
		processIP:   ip,
		processPort: port,
	}

	sys.loader = loader

	sys.addressbook = New(iface.AddressInfo{
		Process: sys.processID,
		Ip:      sys.processIP,
		Port:    sys.processPort,
	})

	if sys.processPort != 0 {

	}

	return sys
}

func (sys *System) Register(ctx context.Context, builder iface.INodeBuilder) (iface.INode, error) {
	if builder.GetID() == "" || builder.GetType() == "" {
		return nil, ErrNodeRegisterInvalidParam
	}

	sys.Lock()
	if _, ok := sys.nodeidmap[builder.GetID()]; ok {
		sys.Unlock()
		return nil, ErrNodeRegisterRepeat
	}
	sys.Unlock()

	if builder.GetGlobalQuantityLimit() != 0 {
		if builder.GetNodeUnique() {
			for _, v := range sys.nodeidmap {
				if v.Type() == builder.GetType() {
					return nil, ErrNodeRegisterUnique
				}
			}
		}

		// todo 判断节点数量
	}

	// Register first, then build
	err := sys.addressbook.Register(ctx, builder.GetType(), builder.GetID(), builder.GetWeight())
	if err != nil {
		return nil, err
	}

	var node iface.INode
	if builder.GetConstructor() != nil {
		node = builder.GetConstructor()(builder)
		node.Init(ctx)
	} else {
		panic(fmt.Errorf("system node:%v register err, constructor is nil", builder.GetType()))
	}

	sys.Lock()
	sys.nodeidmap[builder.GetID()] = node
	sys.Unlock()

	logrus.Infof("system register success node:%v typ:%v id:%v", sys.addressbook.Id, builder.GetType(), builder.GetID())
	return node, nil
}

func (sys *System) Unregister(id, ty string) error {
	// First, check if the node exists and get it
	logrus.Infof("system unregister node id:%v, node:%v, ty:%v", id, sys.addressbook.Id, ty)

	sys.RLock()
	node, exists := sys.nodeidmap[id]
	sys.RUnlock()

	if exists {
		// Call Exit on the node
		node.Exit()

		// Remove the node from the map
		sys.Lock()
		delete(sys.nodeidmap, id)
		sys.Unlock()
	}

	err := sys.addressbook.Unregister(context.TODO(), id, 0)
	if err != nil {
		// Log the error, but don't return it as the actor has already been removed locally
		logrus.Errorf("system unregister node id %s failed from address book err: %v", id, err)
	}

	logrus.Infof("system unregister node id:%s successfully", id)

	return err
}

func (sys *System) Loader(ty string) iface.INodeBuilder {
	return sys.loader.Builder(ty, sys)
}

func (sys *System) AddressBook() iface.IAddressBook {
	return sys.addressbook
}

func (sys *System) Exit(wait *sync.WaitGroup) {
	
}
