package config

import (
	"flag"
	"github.com/ljhe/scream/3rd/db/gorm"
	"github.com/ljhe/scream/3rd/logrus"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var SConf ScreamConfig

type ScreamConfig struct {
	Node Node `yaml:"node"`
	Log  logrus.LogConfig
}

type Node struct {
	Name    string   `yaml:"name"`
	Addr    string   `yaml:"addr"`
	Typ     int      `yaml:"typ"`
	Zone    int      `yaml:"zone"`
	Index   int      `yaml:"index"`
	Connect []string `yaml:"connect"`
	WsAddr  string   `yaml:"ws_addr"`
}

var (
	ServerCmd        = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ServerConfigPath = ServerCmd.String("config", "config.yaml", "server config file")
	OrmConnector     *gorm.Orm
)

func Init() {
	yamlFile, err := os.ReadFile(*ServerConfigPath)
	if err != nil {
		log.Fatalf("global config readFile err:%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &SConf)
	if err != nil {
		log.Fatalf("global config Unmarshal err: %v", err)
	}
	log.Println("global config load success", SConf)
}

func GetOrm() *gorm.Orm {
	return OrmConnector
}
