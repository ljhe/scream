package iface

import (
	"context"
	"github.com/ljhe/scream/lib/pubsub"
	"github.com/ljhe/scream/msg"
)

type CreateFunc func(INodeBuilder) INode

type INode interface {
	Init(ctx context.Context)

	ID() string
	Type() string

	// Received pushes a msg into the actor's mailbox
	Received(mw *msg.Wrapper) error

	// OnEvent registers an event handling chain for the actor
	OnEvent(ev string, createChainF func(INodeContext) IChain) error

	// OnTimer registers a timer function for the actor (Note: all times used here are in milliseconds)
	//  dueTime: delay before execution, 0 for immediate execution
	//  interval: time between each tick
	//  f: callback function
	//  args: can be used to pass the actor entity to the timer callback
	OnTimer(dueTime int64, interval int64, f func(interface{}) error, args interface{}) ITimer

	// CancelTimer cancels a timer
	CancelTimer(t ITimer)

	// Sub subscribes to a message
	//  If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
	//  topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
	//  channel: Represents different categories within a topic
	//  createChainF: Callback function for successful subscription
	Sub(topic string, channel string, createChainF func(node INodeContext) IChain, opts ...pubsub.TopicOption) error

	// Call sends an event to another actor
	Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	ReenterCall(idOrSymbol, actorType, event string, mw *msg.Wrapper) IFuture

	Context() INodeContext

	Exit()
}

type INodeLoader interface {
	// Builder selects a node from the factory and provides a builder
	Builder(string, ISystem) INodeBuilder

	// Pick selects an appropriate node for the actor builder to register
	Pick(context.Context, INodeBuilder) error

	AssignToNode(IProcess)
}

type INodeBuilder interface {
	GetID() string
	GetType() string
	GetGlobalQuantityLimit() int
	GetNodeUnique() bool
	GetWeight() int
	GetOpt(key string) string
	GetOptions() map[string]string

	GetSystem() ISystem
	GetLoader() INodeLoader
	GetConstructor() CreateFunc

	WithID(string) INodeBuilder
	WithType(string) INodeBuilder
	WithOpt(string, string) INodeBuilder

	Register(context.Context) (INode, error)
	Picker(context.Context) error
}

type INodeContext interface {
	// Call performs a blocking call to target actor
	//
	// Parameters:
	//   - idOrSymbol: target actorID, or routing rule symbol to target actor
	//   - actorType: type of actor, obtained from actor template
	//   - event: event name to be handled
	//   - mw: message wrapper for routing
	Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	// ReenterCall performs a reentrant(asynchronous) call
	//
	// Parameters:
	//   - idOrSymbol: target actorID, or routing rule symbol to target actor
	//   - actorType: type of actor, obtained from actor template
	//   - event: event name to be handled
	//   - mw: message wrapper for routing
	ReenterCall(idOrSymbol, actorType, event string, mw *msg.Wrapper) IFuture

	// AddressBook actor 地址管理对象
	AddressBook() IAddressBook

	// Unregister unregisters an node
	Unregister(id, ty string) error
}

type NodeConstructor struct {
	ID   string
	Name string

	// Weight occupied by the actor, weight algorithm reference: 2c4g (pod = 2 * 4 * 1000)
	Weight int

	Dynamic bool

	// Constructor function
	Constructor CreateFunc

	// NodeUnique indicates whether this actor is unique within the current node
	NodeUnique bool

	// Global quantity limit for the current actor type that can be registered
	GlobalQuantityLimit int

	Options map[string]string
}

type INodeFactory interface {
	Get(ty string) *NodeConstructor
	GetNodes() []*NodeConstructor
}
