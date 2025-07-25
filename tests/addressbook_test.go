package tests

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/process"
	"net"
	"sync"
	"testing"
	"time"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func makeNodeKey(nodid string) string {
	return fmt.Sprintf("{node:%s}", nodid)
}

func printWeight() error {

	return nil
}

/*
	func TestDynamicPicker(t *testing.T) {
		for i := 0; i < 10; i++ {
			i := i // 创建一个新的变量来捕获循环变量
			go func() {
				factory := mock.BuildNodeFactory()
				loader := node.BuildDefaultNodeLoader(factory)

				nodid := "1000_" + strconv.Itoa(i)
				port, _ := getFreePort()

				p := process.BuildProcessWithOption(
					process.WithID(nodid),
					process.WithWeight(10000),
					process.WithLoader(loader),
					process.WithFactory(factory),
					process.WithPort(port),
				)

				err := p.Init()
				if err != nil {
					panic(fmt.Errorf("node init err %v", err.Error()))
				}
			}()
		}
		time.Sleep(time.Second)

		////////////////////////////////////////////////////////////////////////////////////

		factory := mock.BuildNodeFactory()
		loader := node.BuildDefaultNodeLoader(factory)

		nodid := "1000_x"
		port, _ := getFreePort()

		p := process.BuildProcessWithOption(
			process.WithID(nodid),
			process.WithWeight(10000),
			process.WithLoader(loader),
			process.WithFactory(factory),
			process.WithPort(port),
		)

		err := p.Init()
		if err != nil {
			panic(fmt.Errorf("node init err %v", err.Error()))
		}

		time.Sleep(time.Second)

		for i := 0; i < 5000; i++ {
			err = p.System().Loader("mocka").WithID(nodid + "_" + strconv.Itoa(i)).Picker(context.TODO())
			if err != nil {
				t.Logf("picker err %v", err.Error())
			}
		}

		time.Sleep(time.Second * 10)

		// 再看下分布情况
		printWeight()
	}
*/
func TestAddressBookDel(t *testing.T) {
	p1 := process.BuildProcessWithOption(
		process.WithID("addressbook_del-1"),
		process.WithPort(8888),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p2 := process.BuildProcessWithOption(
		process.WithID("addressbook_del-2"),
		process.WithPort(7777),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	p1.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	p2.System().Loader("mockb").WithID("mockb").Register(context.TODO())

	p1.Init()
	p2.Init()
	defer func() {
		wg1 := sync.WaitGroup{}
		//wg2 := sync.WaitGroup{}
		p1.System().Exit(&wg1)
		//p2.System().Exit(&wg2)
		wg1.Wait()
		//wg2.Wait()
	}()
	time.Sleep(time.Second)
}
