package logrus

import (
	"testing"
)

func TestTestLog(t *testing.T) {
	Init("")
	for i := 0; i < 10; i++ {
		Infof("this is test: %d", i)
	}
}
