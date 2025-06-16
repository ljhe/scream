package tests

import (
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"os"
	"testing"
)

func init() {
	fmt.Println("main test init")
}

func TestMain(m *testing.M) {
	logrus.Init("")
	code := m.Run()
	os.Exit(code)
}
