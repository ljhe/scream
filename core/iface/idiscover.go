package iface

import "github.com/ljhe/scream/utils"

type IDiscover interface {
	// Loader load all node info by ETCD after the node started
	Loader()
	Close()
	GetNodeByKey(key string) *utils.ServerInfo
}
