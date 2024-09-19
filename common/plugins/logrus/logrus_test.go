package logrus

import (
	"fmt"
	"testing"
)

func TestTestLog(t *testing.T) {
	Init()
	for i := 0; i < 10; i++ {
		entry.Info(fmt.Sprintf("this is a test. i:%d", i))
	}
}
