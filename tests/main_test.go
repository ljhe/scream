package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/ljhe/scream/3rd/redis"
	"github.com/ljhe/scream/core"
	"github.com/ljhe/scream/tests/mock"
)

var factory *mock.MockActorFactory
var loader core.IActorLoader

func TestMain(m *testing.M) {

	factory = mock.BuildActorFactory()
	loader = mock.BuildDefaultActorLoader(factory)

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	redis.BuildClientWithOption(redis.WithAddr(fmt.Sprintf("redis://%s", mr.Addr())))

	os.Exit(m.Run())
}
