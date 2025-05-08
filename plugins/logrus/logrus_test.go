package logrus

import (
	"testing"
)

func TestTestLog(t *testing.T) {
	Init("")
	for i := 0; i < 10; i++ {
		Log(LogsSystem).Infof("this is a tests. i:%d", i)
	}
}
