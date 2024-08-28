package service

import (
	"common"
	plugins "common/plugins/etcd"
	"testing"
)

func TestCreateConnector(t *testing.T) {
	Init()
	CreateConnector(NetNodeParam{
		ServerTyp:            common.SocketTypTcpConnector,
		ServerName:           "test_connector",
		Typ:                  2,
		Zone:                 9999,
		Index:                1,
		DiscoveryServiceName: "test",
	}, plugins.NewMultiServerNode())
	WaitExitSignal()
}
