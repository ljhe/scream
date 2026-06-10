package iface

type IDiscover interface {
	// Loader load all node info by ETCD after the node started
	Loader()
	Close()
}
