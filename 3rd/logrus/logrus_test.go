package logrus

import (
	"github.com/ljhe/scream/def"
	"testing"
)

func TestTestLog(t *testing.T) {
	Init("")
	param := map[string]interface{}{
		"name": "li",
	}
	for i := 0; i < 10; i++ {
		Log(def.LogsSystem, param).Infof("this is a tests. i:%d", i)
	}
}
