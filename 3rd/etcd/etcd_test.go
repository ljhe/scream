package etcd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"testing"
)

func TestRegisterService1(t *testing.T) {
	etcd, err := NewServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		t.Error(err)
	}
	info := &ServerInfo{
		Id:    "game#9999@2@1",
		Name:  "game",
		Host:  "127.0.0.1:2701",
		Typ:   2,
		Zone:  9999,
		Index: 1,
	}
	err = etcd.RegisterService(fmt.Sprintf("%stest1", ServerPreKey), info.String())
	if err != nil {
		t.Error(err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
		fmt.Println("Program terminated.")
	}
}

func TestRegisterService2(t *testing.T) {
	etcd, err := NewServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		t.Error(err)
	}
	info := &ServerInfo{
		Id:    "game#9999@2@2",
		Name:  "game",
		Host:  "127.0.0.1:2702",
		Typ:   2,
		Zone:  9999,
		Index: 2,
	}
	err = etcd.RegisterService(fmt.Sprintf("%stest2", ServerPreKey), info.String())
	if err != nil {
		t.Error(err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
		fmt.Println("Program terminated.")
	}
}

func TestDiscoverServices(t *testing.T) {
	etcd, err := NewServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		t.Error(err)
	}
	err = etcd.DiscoverService(ServerPreKey)
	if err != nil {
		t.Error(err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
		fmt.Println("Program terminated.")
	}
}

func TestDelKey(t *testing.T) {
	etcd, err := NewServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		t.Error(err)
	}
	_, err = etcd.KV.Delete(context.TODO(), "server/9999")
	fmt.Println("etcd del key. err:", err)
}
