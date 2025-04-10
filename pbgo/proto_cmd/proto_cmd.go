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

const PackageName = "pbgo"

var (
	repattern     = regexp.MustCompile(`.*[\.]{1}([^\.]+)`)
	packagekey    = regexp.MustCompile(`.*package\s+([\S]+)\s*;`)
	messageidkey  = regexp.MustCompile(`^\s*([\S]+)\s*=\s*([\d]+)\s*;\s*//\s*(.*)$`)
	messagedefkey = regexp.MustCompile(`^\s*message\s+([\S]+).*//\s*project\s+([\S]+)$`)
	messagekey    = regexp.MustCompile(`\s*message\s+([\S]+)[^\r\n]*`)
)

var protoMsgs = make(map[string][]*msg)

type msg struct {
	id        int
	name      string
	desc      string
	msgId     string
	gopackage string
	file      string
	project   string
}

type section struct {
	begin int
	end   int
}

func main() {
	projects := []string{
		"gate",
		"game",
	}

	// 导出proto文件 协议对应ID
	messageFile := "messagedef.proto"
	messageFileClient := "messagedefclient.proto"
	pbBindGo := "pbbind_gen.go"
	// 每个proto文件的消息ID分段 用来区分消息类型
	configFile := "msgidconfig.cfg"

	msgSection := getMsgSection(configFile)
	messagesDef := analysisMessageDef(messageFile)

	// 获取所有需要解析的proto文件
	fileList := getFileList("./")
	for _, file := range fileList {
		analysisProto(file)
	}

	saveOutFile(messageFile, messageFileClient, pbBindGo, projects, msgSection, messagesDef)
}

// 生成消息对应的ID映射
// 已经存在的消息映射不会改变
// messages[msgkey] = {"id":int(msgid), "desc":msgdesc}
func analysisMessageDef(fileName string) map[string]*msg {
	messageDef := make(map[string]*msg)
	f, err := os.Open(fileName)
	if err != nil {
		return messageDef
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		match := messageidkey.FindStringSubmatch(line)
		if len(match) < 1 {
			continue
		}
		id, _ := strconv.Atoi(match[2])
		messageDef[match[1]] = &msg{
			id:   id,
			desc: match[3],
		}
	}
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
		match := messagedefkey.FindStringSubmatch(line)
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
				project:   match[2],
			})
		} else {
			match = messagekey.FindStringSubmatch(line)
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
func saveOutFile(messageFile, messageFileClient, pbBindGo string, projects []string,
	msgSection map[string]*section, messagesDef map[string]*msg) {
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

			// 获取pbName中已经在messagedef.proto文件中定义的枚举最大值
			if messagesDef[data.msgId] != nil {
				msgIndex := messagesDef[data.msgId].id
				if messageIdMax[pbName] == 0 || messageIdMax[pbName] < msgIndex {
					messageIdMax[pbName] = msgIndex
				}
			}
		}
	}

	// 获取所有协议枚举的最大值
	maxIndex := 0
	for _, data := range messagesDef {
		if data.id > maxIndex {
			maxIndex = data.id
		}
	}

	// 获取配置区间的最大值
	maxMsgSection := 0
	for _, data := range msgSection {
		if data.end > maxMsgSection {
			maxMsgSection = data.end
		}
	}

	for pbName, pbData := range messageMap {
		for _, data := range pbData {
			def := &msg{
				desc:    data.desc,
				name:    data.name,
				file:    pbName,
				msgId:   data.msgId,
				project: data.project,
			}
			if messagesDef[data.msgId] == nil {
				messageIdMax[pbName] = increaseMessageId(msgSection, messageIdMax, pbName, maxIndex, maxMsgSection)
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

	// 生成pbbind_gen.go文件
	savePbBindGo(pbBindGo, messagesDef, projects)
}

// saveMessageDef 生成消息枚举定义文件messagedef.proto
func saveMessageDef(messageFile, messageFileClient string, messagesDef map[string]*msg) {
	sortMsg := make(map[int]*msg)
	for _, data := range messagesDef {
		// 协议被删除后的处理
		if data.name == "" {
			continue
		}
		sortMsg[data.id] = data
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
package %s;
enum protoMsgId{
	MSG_BEGIN	= 0;
%s
}
`, PackageName, messageText))
}

// 生成pbbind_gen.go文件
func savePbBindGo(fileName string, messagesDef map[string]*msg, projects []string) {
	sortHandleMsg := make(map[int]*msg)
	sortInitMsg := make(map[int]*msg)
	for _, data := range messagesDef {
		// 协议被删除后的处理
		if data.name == "" {
			continue
		}
		sortInitMsg[data.id] = data
		if data.project != "" {
			sortHandleMsg[data.id] = data
		}
	}
	// 获取排序后的键
	initKeys := make([]int, 0, len(sortInitMsg))
	for id := range sortInitMsg {
		initKeys = append(initKeys, id)
	}
	sort.Ints(initKeys)
	handleKeys := make([]int, 0, len(sortHandleMsg))
	for id := range sortHandleMsg {
		handleKeys = append(handleKeys, id)
	}
	sort.Ints(handleKeys)

	mhead := fmt.Sprintf(`package %s

import (
	"common"
	"common/iface"
	"log"
	"reflect"
)

func registerInfo(id uint16, msgType reflect.Type) {
	RegisterMessageInfo(&MessageInfo{ID: id, Codec: GetCodec(), Type: msgType})
}
`, PackageName)

	// 具体每个协议的定义
	mhanderDef := ""
	mhandlerDetail := ""
	mhandler := "\nfunc GetMessageHandler(sreviceName string) common.EventCallBack {\n\tswitch sreviceName { //note.serviceName must be lower words"
	for _, p := range projects {
		upper := strings.ToUpper(p)
		lower := strings.ToLower(p)
		mhanderDef += "\n//" + upper + "\nvar ("
		mhandler += "\n\tcase \"" + lower + "\":\t//" + upper + " message process part\n\t\treturn "
		mhandlerDetail = "func(e iface.IProcEvent) {\n\t\t\tswitch e.Msg().(type) {"
		for _, id := range handleKeys {
			data := sortInitMsg[id]
			if p != data.project {
				continue
			}
			mhanderDef += "\n\tHandle_" + upper + "_" + data.name + "  = func(e  iface.IProcEvent){panic(\"" + data.name + " not implements\")}"
			mhandlerDetail += "\n\t\t\tcase *" + data.name + ": Handle_" + upper + "_" + data.name + "(e)"
		}
		mhanderDef += "\n\tHandle_" + upper + "_Default		func(e  iface.IProcEvent)\n)\n"
		mhandlerDetail += "\n\t\t\tdefault:\n\t\t\t\tif Handle_" + upper + "_Default != nil {\n\t\t\t\t\tHandle_" + upper + "_Default(e)\n\t\t\t\t}\n\t\t\t}\n\t\t}\n"
		mhandler += mhandlerDetail
	}
	mhandler += "\n\tdefault: \n\t\treturn nil\n\t}\n}"

	// init部分
	minit := "\n\nfunc init() {\n\t// 协议注册\n\tlog.SetFlags(log.Lshortfile | log.LstdFlags)"
	// 格式化消息文本
	for _, id := range initKeys {
		data := sortInitMsg[id]
		minit += "\n\tregisterInfo(" + strconv.Itoa(id) + ", reflect.TypeOf((*" + data.name + ")(nil)).Elem())"
	}
	minit += "\n\tlog.Println(\"pbbind_gen.go init success\")\n}"

	messageText := mhead + mhanderDef + mhandler + minit
	saveFile(fileName, messageText)
}

func increaseMessageId(msgSection map[string]*section, messageIdMax map[string]int, pbName string, maxIndex, maxMsgSection int) int {
	increase := 1
	index := 0
	if messageIdMax[pbName] == 0 {
		if msgSection[pbName] == nil {
			if maxIndex < maxMsgSection {
				maxIndex = maxMsgSection
			}
			index = maxIndex + increase
		} else {
			index = msgSection[pbName].begin
		}
	} else {
		index = messageIdMax[pbName] + increase
		if index > maxIndex {
			index = maxIndex + increase
		}
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
