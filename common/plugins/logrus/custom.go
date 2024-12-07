package logrus

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var logs logsEntry
var conf logConf

type logsEntry struct {
	entry map[string]*logrus.Entry
}

const (
	// LogsSystem 系统级日志
	LogsSystem = "系统级"
)

func Init(filepath string) {
	loadConfig(filepath)
	logrusInit()
	logs.entry = make(map[string]*logrus.Entry)
	logs.initSystem()
	Log(LogsSystem).Infof("logrus init success. filepath:%v", filepath)
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
func loadConfig(filepath string) {
	if filepath == "" {
		filepath = "./config.yaml"
	}
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("logrus load config readFile err:%v filepath:%v", err, filepath)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("logrus load config Unmarshal err: %v", err)
	}
	log.Println("logrus load config success", conf.Log)
}
