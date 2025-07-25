package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/ljhe/scream/3rd/log"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/lib/grpc"
	"github.com/ljhe/scream/lib/pubsub"
	"github.com/ljhe/scream/msg"
	"github.com/ljhe/scream/msg/router"
	realgrpc "google.golang.org/grpc"
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
	client      *grpc.Client
	ps          *pubsub.Pubsub
	acceptor    *Acceptor
	loader      iface.INodeLoader
	factory     iface.INodeFactory

	processID   string
	processIP   string
	processPort int

	callTimeout time.Duration // sync call timeout

	sync.RWMutex
}

func BuildSystemWithOption(id, ip string, port int, loader iface.INodeLoader, factory iface.INodeFactory) iface.ISystem {
	var err error

	sys := &System{
		nodeidmap:   make(map[string]iface.INode),
		processID:   id,
		processIP:   ip,
		processPort: port,
		callTimeout: time.Second * 5,
	}

	if loader == nil || factory == nil {
		panic("system loader or factory is nil!")
	}

	var unaryInterceptors []realgrpc.UnaryClientInterceptor
	sys.client = grpc.BuildClientWithOption(grpc.ClientAppendUnaryInterceptors(unaryInterceptors...))
	sys.loader = loader
	sys.factory = factory

	sys.ps = pubsub.BuildWithOption()

	sys.addressbook = New(iface.AddressInfo{
		Process: sys.processID,
		Ip:      sys.processIP,
		Port:    sys.processPort,
	})

	if sys.processPort != 0 {
		sys.acceptor, err = NewAcceptor(sys, sys.processPort)
		if err != nil {
			panic(fmt.Errorf("system new acceptor err %v", err.Error()))
		}

		// run grpc acceptor
		sys.acceptor.server.Run()
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

	log.InfoF("system register success node:%v typ:%v id:%v", sys.addressbook.Id, builder.GetType(), builder.GetID())
	return node, nil
}

func (sys *System) Unregister(id, ty string) error {
	// First, check if the node exists and get it
	log.InfoF("system unregister node id:%v, node:%v, ty:%v", id, sys.addressbook.Id, ty)

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
		log.ErrorF("system unregister node id %s failed from address book err: %v", id, err)
	}

	log.InfoF("system unregister node id:%s successfully", id)

	return err
}

func (sys *System) Call(idOrSymbol, nodeType, event string, mw *msg.Wrapper) error {
	// Set message header information
	mw.Req.Header.Event = event
	mw.Req.Header.TargetActorID = idOrSymbol
	mw.Req.Header.TargetActorType = nodeType

	var info iface.AddressInfo
	var node iface.INode
	var err error

	if idOrSymbol == "" {
		return fmt.Errorf("system call unknown target id")
	}

	switch idOrSymbol {
	case def.SymbolWildcard:
		info, err = sys.addressbook.GetWildcardNode(mw.Ctx, nodeType)
		// Check if the wildcard actor is local
		sys.RLock()
		actor, ok := sys.nodeidmap[info.NodeId]
		sys.RUnlock()
		if ok {
			return sys.localCall(actor, mw)
		}
	case def.SymbolLocalFirst:
		node, info, err = sys.findLocalOrWildcardActor(mw.Ctx, nodeType)
		if err != nil {
			return err
		}
		if node != nil {
			// Local call
			return sys.localCall(node, mw)
		}
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
		log.InfoF("system id call %v is not local, get from addressbook ip %v port %v err %v", idOrSymbol, info.Ip, info.Port, err)
	}

	if err != nil {
		return fmt.Errorf("system call id %v ty %v err %w", idOrSymbol, nodeType, err)
	}

	if info.Ip == sys.processIP && info.Port == sys.processPort {
		if err := sys.addressbook.Unregister(mw.Ctx, info.NodeId, sys.factory.Get(nodeType).Weight); err != nil {
			log.WarnF("system unregister stale actor record err actorTy %v NodeId %v err %v", nodeType, info.NodeId, err)
		}
		log.InfoF("system found inconsistent actor record actorTy %v NodeId %v call ev %v, cleaned up", nodeType, info.NodeId, event)

		return ErrSelfCall
	}

	// At this point, we know it's a remote call
	return sys.handleRemoteCall(mw.Ctx, info, mw)
}

func (sys *System) Send(idOrSymbol, nodeType, event string, mw *msg.Wrapper) error {
	// Set message header information
	mw.Req.Header.Event = event
	mw.Req.Header.TargetActorID = idOrSymbol
	mw.Req.Header.TargetActorType = nodeType

	var info iface.AddressInfo
	//var node iface.INode
	var err error

	if idOrSymbol == "" {
		return fmt.Errorf("system send unknown target id")
	}

	switch idOrSymbol {
	case def.SymbolWildcard:
		info, err = sys.addressbook.GetWildcardNode(mw.Ctx, nodeType)
		// Check if the wildcard node is local
		sys.RLock()
		node, ok := sys.nodeidmap[info.NodeId]
		sys.RUnlock()
		if ok {
			return node.Received(mw)
		}
	default:
		// First, check if it's a local call
		sys.RLock()
		actorp, ok := sys.nodeidmap[idOrSymbol]
		sys.RUnlock()

		if ok {
			return actorp.Received(mw)
		}

		// If not local, get from addressbook
		info, err = sys.addressbook.GetByID(mw.Ctx, idOrSymbol)
	}

	if err != nil {
		return fmt.Errorf("system send id %v ty %v err %w", idOrSymbol, nodeType, err)
	}

	if info.Ip == sys.processIP && info.Port == sys.processPort {
		if err := sys.addressbook.Unregister(mw.Ctx, info.NodeId, sys.factory.Get(nodeType).Weight); err != nil {
			log.ErrorF("system unregister stale actor record err actorTy %v actorID %v err %v", nodeType, info.NodeId, err)
		}
		log.WarnF("system found inconsistent actor record actorTy %v actorID %v call ev %v, cleaned up", nodeType, info.NodeId, event)
		return ErrSelfCall
	}

	return sys.handleRemoteSend(info, mw)
}

func (sys *System) handleRemoteSend(info iface.AddressInfo, mw *msg.Wrapper) error {
	return sys.client.Call(mw.Ctx,
		fmt.Sprintf("%s:%d", info.Ip, info.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: mw.Req},
		nil) // We don't need the response for Send
}

func (sys *System) findLocalOrWildcardActor(ctx context.Context, ty string) (iface.INode, iface.AddressInfo, error) {
	sys.RLock()

	for id, node := range sys.nodeidmap {
		if node.Type() == ty {
			sys.RUnlock()
			return node, iface.AddressInfo{NodeId: id, NodeTy: ty}, nil
		}
	}
	sys.RUnlock()

	// If not found locally, use GetWildcardNode to perform a random search across the cluster
	info, err := sys.addressbook.GetWildcardNode(ctx, ty)
	return nil, info, err
}

func (sys *System) localCall(actorp iface.INode, mw *msg.Wrapper) error {

	root := mw.GetWg().Count() == 0
	if root {
		log.InfoF("system local call root event %v id %v", mw.Req.Header.Event, mw.Req.Header.TargetActorID)
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
				log.WarnF("system wait timeout for event %v id %v, remaining tasks: %d",
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
			timeoutErr := fmt.Errorf("node:%v, message:%v, processing timed out",
				mw.Req.Header.TargetActorID, mw.Req.Header.Event)
			if mw.Err != nil {
				timeoutErr = fmt.Errorf("%w: %v", mw.Err, timeoutErr)
			}
			mw.Err = timeoutErr
			return timeoutErr
		}
	} else {
		log.InfoF("system local call received event %v id %v", mw.Req.Header.Event, mw.Req.Header.TargetActorID)
		return actorp.Received(mw)
	}
}

func (sys *System) handleRemoteCall(ctx context.Context, addrinfo iface.AddressInfo, mw *msg.Wrapper) error {
	res := &router.RouteRes{}
	err := sys.client.CallWait(ctx,
		fmt.Sprintf("%s:%d", addrinfo.Ip, addrinfo.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: mw.Req},
		res)

	if err != nil {
		return err
	}

	mw.Res = res.Msg
	return nil
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
	if sys.processPort != 0 {
		wait.Add(1)
		if sys.acceptor != nil {
			sys.acceptor.Exit()
			log.InfoF("system acceptor exit")
		}
		wait.Done()
	}

	for _, node := range sys.nodeidmap {
		wait.Add(1)

		go func(n iface.INode) {
			defer wait.Done()
			err := sys.addressbook.Unregister(context.TODO(), n.ID(), 0)
			if err != nil {
				log.ErrorF("system unregister err nodeID %v nodeTy %v err %v", node.Type(), n.ID(), err)
			}
			n.Exit()
			log.InfoF("system node exit %v", n.ID())
		}(node)
	}
}
