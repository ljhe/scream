package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/3rd/log"
	"github.com/ljhe/scream/utils"
	"github.com/ljhe/scream/utils/request"
	"github.com/ljhe/scream/utils/template"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	confPath := "./config.yaml"
	template.Init("UTILS", confPath)

	logger, err := log.NewDefaultLogger(func(options *log.Options) error {
		options.GlobPattern = fmt.Sprintf("utils_%s", utils.GetDateOnly())
		options.OutStd = true
		return nil
	})
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	var conf Config
	err = conf.Load(confPath)
	if err != nil {
		log.ErrorF("global config Load err: %v", err)
		return
	}

	fmt.Println(template.DividingLine)
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		for _, opt := range conf.Template {
			fmt.Println(opt.Options)
		}

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		num, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("输入无效，请输入数字！")
			continue
		}

		switch num {
		case 1:
			option1(conf)
		case 0:
			fmt.Println("程序退出。")
			return
		default:
			fmt.Println("无效选项，请重新输入。")
		}
		fmt.Println(template.DividingLine)
		fmt.Println()
	}
}

func option1(conf Config) {
	fmt.Println("开始执行")

	param := map[string]interface{}{
		"sign":          conf.Options1.Sign,
		"guid":          conf.Options1.Guid,
		"local_realm":   conf.Options1.LocalRealm,
		"remote_realm":  conf.Options1.RemoteRealm,
		"target_openid": conf.Options1.TargetOpenid,
		"data":          "",
	}

	err, bt := request.Post(conf.Options1.OnlineUrl, param)
	if err != nil {
		log.ErrorF("post online_url err: %v", err)
		return
	}
	var res1 Result
	err = json.Unmarshal(bt, &res1)
	if err != nil {
		log.ErrorF("Unmarshal online_url err: %v", err)
		return
	}

	if res1.Code != http.StatusOK {
		log.ErrorF("post online_url err:%v", res1.Msg)
		return
	}

	param["data"] = res1.Data

	err, bt = request.Post(conf.Options1.LocalUrl, param)
	if err != nil {
		log.ErrorF("post local_url err: %v", err)
		return
	}
	var res2 Result
	err = json.Unmarshal(bt, &res2)
	if err != nil {
		log.ErrorF("Unmarshal local_url err: %v", err)
		return
	}

	if res2.Code != http.StatusOK {
		log.ErrorF("post online_url err:%v", res2.Msg)
		return
	}
	fmt.Println("执行完成")
}

type Result struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Msg  string `json:"msg"`
}

type Options struct {
	Options string `yaml:"options"`
}

type Options1 struct {
	OnlineUrl    string `yaml:"online_url"`
	LocalUrl     string `yaml:"local_url"`
	Sign         string `yaml:"sign"`
	Guid         string `yaml:"guid"`
	LocalRealm   int    `yaml:"local_realm"`
	RemoteRealm  int    `yaml:"remote_realm"`
	TargetOpenid string `yaml:"target_openid"`
}

type Config struct {
	Template []*Options `yaml:"template"`
	Options1 Options1   `yaml:"options1"`
}

func (c *Config) Load(filepath string) error {
	file, err := os.ReadFile(filepath)
	//file, err := os.ReadFile("E:\\workspace\\go\\scream\\utils\\template\\template_test" + filepath)
	if err != nil {
		panic(fmt.Sprintf("config filepath err. filepath:%v err:%v", filepath, err))
	}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		panic(fmt.Sprintf("config Unmarshal err: %v", err))
	}
	return nil
}
