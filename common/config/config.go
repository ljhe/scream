package config

import (
	"common/plugins/logrus"
	"flag"
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
	Name                 string `yaml:"name"`
	Addr                 string `yaml:"addr"`
	Typ                  int    `yaml:"typ"`
	Zone                 int    `yaml:"zone"`
	Index                int    `yaml:"index"`
	DiscoveryServiceName string `yaml:"discoveryServiceName"`
}

var (
	ServerCmd        = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ServerConfigPath = ServerCmd.String("config", "config.yaml", "server config file")
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
