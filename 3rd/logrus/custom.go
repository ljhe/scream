package logrus

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

const (
	fieldsTag    = "tag"
	fieldsZLog   = "zlog"
	fieldsCaller = "caller"
)

var opt *Options
var lJson *logrus.Logger
var lText *logrus.Logger

func Init(filepath string) {
	loadConfig(filepath)
	lJson = NewLogger(&SelfJsonFormatter{})
	lText = NewLogger(&SelfTextFormatter{})
	Infof("logrus init success")
}

func LogJson(tag string, param ...interface{}) *logrus.Entry {
	fields := logrus.Fields{
		fieldsTag: tag,
	}
	if len(param) > 0 {
		fields[fieldsZLog] = param
	}
	return lJson.WithFields(fields)
}

func logText(caller string) *logrus.Entry {
	fields := logrus.Fields{
		fieldsCaller: caller,
	}
	return lText.WithFields(fields)
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
}
