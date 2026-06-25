package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/ljhe/scream/3rd/log"
	trdredis "github.com/ljhe/scream/3rd/redis"
	"github.com/ljhe/scream/core"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/def"
	"github.com/redis/go-redis/v9"
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
	// Get all node infos from the set
	nodeInfoMap, err := trdredis.HGetAll(context.Background(), def.RedisAddressbookNodesField).Result()
	if err != nil {
		return fmt.Errorf("failed to get node infos: %v", err)
	}

	if len(nodeInfoMap) == 0 {
		return fmt.Errorf("no nodes found")
	}

	pipe := trdredis.Pipeline()

	// Prepare pipeline commands to get weights for all nodes
	for nodeID := range nodeInfoMap {
		pipe.HGet(context.Background(), makeNodeKey(nodeID), "total_weight")
	}

	// Execute pipeline
	cmders, err := pipe.Exec(context.Background())
	if err != nil {
		return fmt.Errorf("pipeline execution failed: %v", err)
	}

	// Process results
	i := 0
	for nodeID, nodeInfoJSON := range nodeInfoMap {
		if i >= len(cmders) {
			break
		}

		var nodeInfo core.AddressInfo
		if err := json.Unmarshal([]byte(nodeInfoJSON), &nodeInfo); err != nil {
			log.WarnF("unable to unmarshal node info: %v", err)
			i++
			continue
		}

		weightStr, err := cmders[i].(*redis.StringCmd).Result()
		if err != nil {
			log.WarnF("unable to get weight for node %s: %v", nodeID, err)
			i++
			continue
		}

		weight, _ := strconv.Atoi(weightStr)
		fmt.Println("node", nodeInfo.Node, "cur weight", weight)

		i++
	}

	return nil
}

func TestDynamicPicker(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func(i int) {
			id := "1000_" + strconv.Itoa(i)
			p, _ := getFreePort()

			nod := node.BuildProcessWithOption(
				core.NodeWithID(id),
				core.NodeWithWeight(10000),
				core.NodeWithLoader(loader),
				core.NodeWithFactory(factory),
				core.NodeWithPort(p),
			)

			err := nod.Init()
			if err != nil {
				panic(fmt.Errorf("node init err %v", err.Error()))
			}
		}(i)
	}
	time.Sleep(time.Second)

	id := "1000_x"
	p, _ := getFreePort()

	nod := node.BuildProcessWithOption(
		core.NodeWithID(id),
		core.NodeWithWeight(10000),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
		core.NodeWithPort(p),
	)

	err := nod.Init()
	if err != nil {
		panic(fmt.Errorf("node init err %v", err.Error()))
	}

	time.Sleep(time.Second)

	for i := 0; i < 5; i++ {
		err = nod.System().Loader("mocka").WithID(id + "_" + strconv.Itoa(i)).Picker(context.TODO())
		if err != nil {
			t.Logf("picker err %v", err.Error())
		}
	}

	time.Sleep(time.Second * 10)
	_ = printWeight()
}
