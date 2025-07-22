package def

const (
	// Wildcard symbol represents routing to any node of this type
	SymbolWildcard = "?"

	// Represents routing to a group of nodes
	// - Note: This symbol can only be used with the send interface (asynchronous call)
	SymbolGroup = "#"

	// Represents routing to all nodes of this type
	// - Note: This symbol can only be used with the send interface (asynchronous call)
	SymbolAll = "*"

	// Represents random routing to an node of this type, but prioritizes nodes on the current node
	// If there are no nodes of this type on the current node, it randomly selects from other nodes
	SymbolLocalFirst = "~"
)

const (
	AddressBookIDField = "addressbook.id"
	AddressBookTyField = "addressbook.ty"
)
