package iface

import (
	"context"
	"github.com/ljhe/scream/lib/pubsub"
	"github.com/ljhe/scream/router"
	"sync"
)

type ISystem interface {
	Register(context.Context, INodeBuilder) (INode, error)
	Unregister(id, ty string) error

	// Call sends an event to another actor
	// Synchronous call semantics (actual implementation is asynchronous, each call is in a separate goroutine)
	Call(idOrSymbol, actorType, event string, mw *router.Wrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(topic string, event string, body []byte) error

	// Sub listens to messages in a channel within a specific topic
	//  opts can be used to set initial values on first listen, such as setting the TTL for messages in this topic
	Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error)

	// Loader returns the node loader
	Loader(string) INodeBuilder

	AddressBook() IAddressBook

	Exit(*sync.WaitGroup)
}
