package logrus

import (
	"fmt"
	"github.com/ljhe/scream/def"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var opt *Options
var logJson *logrus.Logger
var logText *logrus.Logger

func Init(filepath string) {
	loadConfig(filepath)
	logJson = NewLogger(&SelfJsonFormatter{})
	logText = NewLogger(&SelfTextFormatter{})
	Log(def.LogsSystem).Infof("logrus init success. filepath:%v", filepath)
}

func Log(tag string, param ...interface{}) *logrus.Entry {
	fields := logrus.Fields{
		"tag": tag,
	}
	if len(param) > 0 {
		fields["zlog"] = param
	}
	return logJson.WithFields(fields)
}

// 加载配置文件
func loadConfig(filepath string) {
	conf := &logConf{}
	if filepath != "" {
		yamlFile, err := os.ReadFile(filepath)
		if err != nil {
			log.Fatalf("logrus load config readFile err:%v filepath:%v", err, filepath)
		}
		err = yaml.Unmarshal(yamlFile, conf)
		if err != nil {
			log.Fatalf("logrus load config Unmarshal err: %v", err)
		}
	}
	opt = NewOptions(conf)
	fmt.Printf("logrus load config success. opt:%v \n", opt)
}
