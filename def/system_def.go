package def

const (
	// SymbolWildcard Wildcard symbol represents routing to any actor of this type
	SymbolWildcard = "?"

	// SymbolGroup Represents routing to a group of actors
	// - Note: This symbol can only be used with the send interface (asynchronous call)
	SymbolGroup = "#"

	// SymbolAll Represents routing to all actors of this type
	// - Note: This symbol can only be used with the send interface (asynchronous call)
	SymbolAll = "*"

	// SymbolLocalFirst Represents random routing to an actor of this type, but prioritizes actors on the current node
	// If there are no actors of this type on the current node, it randomly selects from other nodes
	SymbolLocalFirst = "~"
)

const (
	RedisAddressbookIDField    = "addressbook.id"
	RedisAddressbookTyField    = "addressbook.ty"
	RedisAddressbookNodesField = "addressbook.nodes"
)
