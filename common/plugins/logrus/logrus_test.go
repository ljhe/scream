package logrus

import (
	"fmt"
	"testing"
)

func TestTestLog(t *testing.T) {
	Init()
	for i := 0; i < 10; i++ {
		logger.Info(fmt.Sprintf("this is a test. i:%d", i))
	}
}
