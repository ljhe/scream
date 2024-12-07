package config

import (
	"common/plugins/logrus"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var ServerConfigPath string
var SConf ScreamConfig

type ScreamConfig struct {
	Log logrus.LogConfig
}

func Init() {
	// todo 之后需要改为从参数中读取
	ServerConfigPath = "../common/config/config.yaml"
	yamlFile, err := os.ReadFile(ServerConfigPath)
	if err != nil {
		log.Fatalf("global config readFile err:%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &SConf)
	if err != nil {
		log.Fatalf("global config Unmarshal err: %v", err)
	}
	log.Println("global config load success", SConf)
}
