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

func Init(filepath string) {
	loadConfig(filepath)
	logrusInit()
	Log(def.LogsSystem).Infof("logrus init success. filepath:%v", filepath)
}

func Log(tag string, param ...interface{}) *logrus.Entry {
	fields := logrus.Fields{
		"tag": tag,
	}
	if len(param) > 0 {
		if p, ok := param[0].(map[string]interface{}); ok {
			for k, v := range p {
				fields[k] = v
			}
		}
	}
	return logger.WithFields(fields)
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
