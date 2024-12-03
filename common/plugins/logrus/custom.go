package logrus

import (
	"github.com/sirupsen/logrus"
)

var logs logsEntry

type logsEntry struct {
	entry map[string]*logrus.Entry
}

const (
	// LogsSystem 系统级日志
	LogsSystem = "系统级"
)

func init() {
	logrusInit()
	logs.entry = make(map[string]*logrus.Entry)
	logs.initSystem()
	Log(LogsSystem).Info("logrus init success")
}

// 初始化系统级日志
func (l *logsEntry) initSystem() {
	l.entry[LogsSystem] = logger.WithFields(logrus.Fields{
		"type": LogsSystem,
	})
}

func (l *logsEntry) getEntry(typ string) *logrus.Entry {
	return l.entry[typ]
}

func Log(typ string) *logrus.Entry {
	return logs.getEntry(typ)
}
