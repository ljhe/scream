package iface

import "context"

type AddressInfo struct {
	NodeId  string `json:"node_id"`
	NodeTy  string `json:"node_ty"`
	Process string `json:"process"`
	Service string `json:"service"`
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
}

type IAddressBook interface {
	Register(context.Context, string, string, int) error
	Unregister(context.Context, string, int) error

	Watch(context.Context)

	GetByID(context.Context, string) (AddressInfo, error)
	GetByType(context.Context, string) ([]AddressInfo, error)

	GetWildcardNode(ctx context.Context, nodeType string) (AddressInfo, error)
	GetLowWeightNodeForNode(ctx context.Context, nodeType string) (AddressInfo, error)
	GetNodeTypeCount(ctx context.Context, nodeType string) (int64, error)

	Clear(context.Context) error
}
