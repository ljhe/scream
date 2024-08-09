package plugins

import (
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
	err = etcd.RegisterService("/service/test1", "127.0.0.1:2901")
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
	err = etcd.RegisterService("/service/test2", "127.0.0.1:2902")
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
	err = etcd.DiscoverService("/service")
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
