package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/lib/pubsub"
	"github.com/ljhe/scream/router"
	"sync"
	"time"
)

var (
	ErrNodeRegisterInvalidParam = errors.New("system register invalid parameter")
	ErrNodeRegisterRepeat       = errors.New("system register node repeat")
	ErrNodeRegisterUnique       = errors.New("system register unique type actor")
	ErrSelfCall                 = errors.New("cannot call self node through RPC")
)

type System struct {
	addressbook *AddressBook
	nodeidmap   map[string]iface.INode
	ps          *pubsub.Pubsub
	loader      iface.INodeLoader
	factory     iface.INodeFactory

	processID   string
	processIP   string
	processPort int

	callTimeout time.Duration // sync call timeout

	sync.RWMutex
}

func BuildSystemWithOption(id, ip string, port int, loader iface.INodeLoader, factory iface.INodeFactory) iface.ISystem {

	sys := &System{
		nodeidmap:   make(map[string]iface.INode),
		processID:   id,
		processIP:   ip,
		processPort: port,
		callTimeout: time.Second * 5,
	}

	sys.loader = loader
	sys.factory = factory

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

func (sys *System) Call(idOrSymbol, actorType, event string, mw *router.Wrapper) error {
	// Set message header information
	mw.Req.Header.Event = event
	mw.Req.Header.TargetActorID = idOrSymbol
	mw.Req.Header.TargetActorType = actorType

	var info iface.AddressInfo
	//var actor iface.INode
	var err error

	if idOrSymbol == "" {
		return fmt.Errorf("system call unknown target id")
	}

	switch idOrSymbol {
	default:
		// First, check if it's a local call
		sys.RLock()
		actorp, ok := sys.nodeidmap[idOrSymbol]
		sys.RUnlock()

		if ok {
			return sys.localCall(actorp, mw)
		}

		// If not local, get from addressbook
		info, err = sys.addressbook.GetByID(mw.Ctx, idOrSymbol)
		logrus.Infof("system id call %v is not local, get from addressbook ip %v port %v err %v", idOrSymbol, info.Ip, info.Port, err)
	}

	if err != nil {
		return fmt.Errorf("system call id %v ty %v err %w", idOrSymbol, actorType, err)
	}

	if info.Ip == sys.processIP && info.Port == sys.processPort {
		if err := sys.addressbook.Unregister(mw.Ctx, info.NodeId, sys.factory.Get(actorType).Weight); err != nil {
			logrus.Warnf("system unregister stale actor record err actorTy %v NodeId %v err %v", actorType, info.NodeId, err)
		}
		logrus.Infof("system found inconsistent actor record actorTy %v NodeId %v call ev %v, cleaned up", actorType, info.NodeId, event)

		return ErrSelfCall
	}

	// At this point, we know it's a remote call
	return nil
}

func (sys *System) localCall(actorp iface.INode, mw *router.Wrapper) error {

	root := mw.GetWg().Count() == 0
	if root {
		logrus.Infof("system local call root event %v id %v", mw.Req.Header.Event, mw.Req.Header.TargetActorID)
		mw.Done = make(chan struct{})
		ready := make(chan struct{})
		go func() {
			<-ready // Wait for Received to complete

			waitCh := make(chan struct{})
			go func() {
				mw.GetWg().Wait()
				close(waitCh)
			}()

			select {
			case <-waitCh:
				// 正常完成
			case <-time.After(sys.callTimeout):
				logrus.Warnf("system wait timeout for event %v id %v, remaining tasks: %d",
					mw.Req.Header.Event, mw.Req.Header.TargetActorID, mw.GetWg().Count())
				if mw.Err == nil {
					mw.Err = fmt.Errorf("system wait timeout, some tasks did not complete")
				}
			}

			close(mw.Done)
		}()

		if err := actorp.Received(mw); err != nil {
			close(ready) // Ensure the ready channel is closed even in case of an error
			return err
		}
		close(ready) // Notify the goroutine that Received has completed

		select {
		case <-mw.Done:
			return nil
		case <-mw.Ctx.Done():
			timeoutErr := fmt.Errorf("actor %v message %v processing timed out",
				mw.Req.Header.TargetActorID, mw.Req.Header.Event)
			if mw.Err != nil {
				timeoutErr = fmt.Errorf("%w: %v", mw.Err, timeoutErr)
			}
			mw.Err = timeoutErr
			return timeoutErr
		}
	} else {
		logrus.Infof("system local call received event %v id %v", mw.Req.Header.Event, mw.Req.Header.TargetActorID)
		return actorp.Received(mw)
	}
}

func (sys *System) Loader(ty string) iface.INodeBuilder {
	return sys.loader.Builder(ty, sys)
}

func (sys *System) AddressBook() iface.IAddressBook {
	return sys.addressbook
}

func (sys *System) Pub(topic string, event string, body []byte) error {
	return sys.ps.GetTopic(topic).Pub(context.TODO(), event, body)
}

func (sys *System) Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error) {
	return sys.ps.GetOrCreateTopic(topic).Sub(context.TODO(), channel)
}

func (sys *System) Exit(wait *sync.WaitGroup) {

}
