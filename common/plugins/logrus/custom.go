package logrus

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var logs logsEntry
var conf config

type logsEntry struct {
	entry map[string]*logrus.Entry
}

const (
	// LogsSystem 系统级日志
	LogsSystem = "系统级"
)

func init() {
	loadConfig()
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

// 加载配置文件
func loadConfig() {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("logrus load config readFile err:%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("logrus load config Unmarshal err: %v", err)
	}
	log.Println("logrus load config success", conf.Log)
}
