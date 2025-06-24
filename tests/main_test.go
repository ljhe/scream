package tests

import (
	"github.com/ljhe/scream/3rd/logrus"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logrus.Init("")
	os.Exit(m.Run())
}
