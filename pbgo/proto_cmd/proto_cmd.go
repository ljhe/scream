package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	repattern  = regexp.MustCompile(`.*[\.]{1}([^\.]+)`)
	packagekey = regexp.MustCompile(`.*package\s+([\S]+)\s*;`)
	messagekey = regexp.MustCompile(`\s*message\s+([\S]+)[^\r\n]*`)
)

var protoMsgs = make(map[string][]*msg)

type msg struct {
	id        int
	name      string
	desc      string
	msgId     string
	gopackage string
	file      string
}

type section struct {
	begin int
	end   int
}

func main() {
	// 导出的proto文件协议对应ID
	messageFile := "messagedef.proto"
	messageFileClient := "messagedefclient.proto"
	// 每个proto文件的消息ID分段 用来区分消息类型
	configFile := "msgidconfig.cfg"

	msgSection := getMsgSection(configFile)
	messagesDef := analysisMessageDef(messageFile)

	// 获取所有需要解析的proto文件
	fileList := getFileList("./")
	for _, file := range fileList {
		analysisProto(file)
	}

	saveOutFile("./", messageFile, messageFileClient, msgSection, messagesDef)
}

// 生成消息对应的ID映射
// 已经存在的消息映射不会改变
// messages[msgkey] = {"id":int(msgid), "desc":msgdesc}
func analysisMessageDef(fileName string) map[string]*msg {
	messageDef := make(map[string]*msg)
	return messageDef
}

// 获得msgidconfig.cfg中定义的消息ID区间段
// 例如 login.proto 	1000-1999
func getMsgSection(fileName string) map[string]*section {
	msgSection := make(map[string]*section)
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("error opening file %q: %v\n", fileName, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalf("error closing file %q: %v\n", fileName, err)
		}
	}(f)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) <= 1 || line[:1] == "#" {
			continue
		}
		str := strings.Split(line, ":")
		if len(str) < 2 {
			continue
		}
		protoName := str[0]
		if len(protoName) <= 0 {
			continue
		}
		value := strings.Split(str[1], "-")
		if len(value) < 2 {
			continue
		}
		begin, _ := strconv.Atoi(value[0])
		end, _ := strconv.Atoi(value[1])
		msgSection[protoName] = &section{begin: begin, end: end}
	}
	return msgSection
}

// getFileList 获取指定路径下的指定类型文件列表
func getFileList(root string) []string {
	fileList := make([]string, 0)
	err := filepath.Walk(root, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("prevent panic by handling failure accessing a path %q: %v\n", p, err)
		}
		if f.IsDir() {
			return nil
		}
		match := repattern.FindStringSubmatch(f.Name())
		if len(match) > 1 {
			if match[1] == "proto" {
				fileList = append(fileList, p)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("filepath.Walk() failed with %s\n", err)
	}
	return fileList
}

// analysisProto 解析.proto文件
func analysisProto(fileName string) map[string][]*msg {
	println("analysis proto:", fileName)
	define := make([]*msg, 0)
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("error opening file %q: %v\n", fileName, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalf("error closing file %q: %v\n", fileName, err)
		}
	}(f)

	gopackage := ""
	lastline := ""
	// 按行读取内容
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) <= 0 {
			continue
		}
		if gopackage == "" {
			match := packagekey.FindStringSubmatch(line)
			if len(match) > 1 {
				gopackage = strings.TrimSpace(match[1])
			}
		}
		// 协议消息注释
		desc := ""
		if strings.HasPrefix(lastline, "//") {
			desc = strings.TrimSpace(lastline[2:])
		}
		match := messagekey.FindStringSubmatch(line)
		if len(match) > 1 {
			msgdefname := match[1]
			msgdefnameLower := strings.ToLower(msgdefname)
			if len(msgdefnameLower) < 3 {
				continue
			}
			suffix := msgdefnameLower[len(msgdefnameLower)-3:]
			if !inArray(suffix, []string{"ntf", "ack", "req"}) {
				continue
			}
			if len(desc) <= 0 {
				desc = msgdefname
			}
			define = append(define, &msg{
				name:      msgdefname,
				desc:      desc,
				msgId:     messageIdGen(msgdefname, gopackage),
				gopackage: gopackage,
			})
		}
		lastline = line
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("error scanning file %q: %v\n", fileName, err)
	}
	if len(define) > 0 {
		newName := fileName
		nameArr := strings.Split(strings.Replace(fileName, "\\", "/", -1), "/")
		if len(nameArr) > 0 {
			newName = nameArr[len(nameArr)-1]
		}
		protoMsgs[newName] = define
	}
	return nil
}

// 根据proto中定义的结构名称生成对应规则的枚举名称
// 例如CSLoginReq -> CS_LOGIN_REQ
func messageIdGen(name, packageName string) string {
	// 提取前两个字符
	result := name[:2]
	// 大写字母集
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	nameLen := len(name)
	// 遍历从第 3 个字符到倒数第 2 个字符
	for i := 2; i < nameLen-1; i++ {
		currentChar := string(name[i])
		if strings.Contains(uppercase, currentChar) {
			// 满足条件则在前面加下划线
			if i == 2 || !strings.Contains(uppercase, string(name[i-1])) || !strings.Contains(uppercase, string(name[i+1])) {
				result += "_"
			}
		}
		// 添加当前字符
		result += currentChar
	}
	// 添加最后一个字符
	result += string(name[nameLen-1])
	// 转为大写并返回
	return strings.ToUpper(result)
}

// saveOutFile 输出文件
func saveOutFile(outDir, messageFile, messageFileClient string, msgSection map[string]*section, messagesDef map[string]*msg) {
	// 所有消息的集合
	messageMap := make(map[string][]*msg)
	// 更新每一个proto文件最大的协议号
	messageIdMax := make(map[string]int)

	// pbName为proto文件名 例如login.proto
	for pbName, pbData := range protoMsgs {
		for _, data := range pbData {
			if messageMap[pbName] == nil {
				messageMap[pbName] = make([]*msg, 0)
			}
			messageMap[pbName] = append(messageMap[pbName], data)
		}
	}

	// 获取所有协议枚举的最大值
	maxIndex := 0
	for _, data := range messagesDef {
		if data.id > maxIndex {
			maxIndex = data.id
		}
	}

	for pbName, pbData := range messageMap {
		for _, data := range pbData {
			def := &msg{
				desc:  data.desc,
				name:  data.name,
				file:  pbName,
				msgId: data.msgId,
			}
			if messagesDef[data.msgId] == nil {
				messageIdMax[pbName] = increaseMessageId(msgSection, messageIdMax, pbName, maxIndex)
				maxIndex = messageIdMax[pbName]
				def.id = messageIdMax[pbName]
			} else {
				def.id = messagesDef[data.msgId].id
			}
			messagesDef[data.msgId] = def
		}
	}

	// 生成消息枚举定义文件messagedef.proto
	saveMessageDef(messageFile, messageFileClient, messagesDef)
}

// saveMessageDef 生成消息枚举定义文件messagedef.proto
func saveMessageDef(messageFile, messageFileClient string, messagesDef map[string]*msg) {
	sortMsg := make(map[int]*msg)
	for _, data := range messagesDef {
		sortMsg[data.id] = data
		fmt.Println("这里是测试 ", data)
	}
	// 获取排序后的键
	keys := make([]int, 0, len(sortMsg))
	for id := range sortMsg {
		keys = append(keys, id)
	}
	sort.Ints(keys)

	// 构建消息文本
	messageText := ""

	// 格式化消息文本
	for _, id := range keys {
		data := sortMsg[id]
		messageText += fmt.Sprintf("\t%-32s = %d;\t\t//\t%s **%s **%s **%s [%s]\n",
			data.msgId, id, data.desc, data.name, data.file, data.gopackage, data.name)
	}

	// 保存到文件
	saveFile(messageFile, fmt.Sprintf(`syntax = "proto3";
package serverproto;
enum protoMsgId{
	MSG_BEGIN	= 0;
%s
}
`, messageText))
}

func increaseMessageId(msgSection map[string]*section, messageIdMax map[string]int, pbName string, maxIndex int) int {
	increase := 1
	index := 0
	if messageIdMax[pbName] == 0 {
		if msgSection[pbName] == nil {
			index = maxIndex + increase
		} else {
			index = msgSection[pbName].begin
		}
	} else {
		index = messageIdMax[pbName] + increase
		if msgSection[pbName] != nil {
			section := msgSection[pbName]
			// pbName当前对应的区域段ID不够使用
			if index > section.end {
				log.Fatalf("error increasing message id max index: %d pbName:%v \n", index, pbName)
			}
		}
	}
	return index
}

// 保存文件
func saveFile(fileName, content string) {
	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		fmt.Println("写文件失败:", err)
	}
}

func inArray(str string, target []string) bool {
	for _, t := range target {
		if str == t {
			return true
		}
	}
	return false
}
