package config

import (
	"flag"
	"fmt"
	"github.com/ljhe/scream/3rd/db/gorm"
	"github.com/ljhe/scream/utils"
	"gopkg.in/yaml.v2"
	"os"
)

type ScreamConfig struct {
	Node Node `yaml:"node"`
}

type Node struct {
	Name    string   `yaml:"name"`
	IP      string   `yaml:"ip"`
	Port    int      `yaml:"port"`
	Typ     int      `yaml:"typ"`
	Index   int      `yaml:"index"`
	Connect []string `yaml:"connect"`
	WsAddr  string   `yaml:"ws_addr"`
	Etcd    string   `yaml:"etcd"`
}

var (
	ServerCmd        = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ServerConfigPath = ServerCmd.String("config", "", "server config file")
	OrmConnector     *gorm.Orm
)

func Init() *ScreamConfig {
	if utils.IsTesting() {
		return nil
	}
	yamlFile, err := os.ReadFile(*ServerConfigPath)
	if err != nil {
		panic(fmt.Sprintf("read config file err:%v", err))
	}
	var conf ScreamConfig
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		panic(fmt.Sprintf("global config Unmarshal err: %v", err))
	}
	return &conf
}

func GetOrm() *gorm.Orm {
	return OrmConnector
}
