package logrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/utils"
	"github.com/sirupsen/logrus"
	"strings"
)

type SelfTextFormatter struct{}

func (f *SelfTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	// 获取日志级别
	b.WriteString(fmt.Sprintf("[%s] ", entry.Level.String()))

	// 获取日志时间
	b.WriteString(fmt.Sprintf("%s ", entry.Time.Format(utils.DateTimeMS)))

	// 添加自定义字段
	for k, v := range entry.Data {
		b.WriteString(fmt.Sprintf("%s:%v ", k, v))
	}

	// 打印日志的位置信息
	// 需要SetReportCaller(true) 否则这里会报错
	// 格式化文件名
	str := strings.Split(entry.Caller.File, "/")
	b.WriteString(fmt.Sprintf("[%s:%d] ", str[len(str)-1], entry.Caller.Line))

	// 日志消息的值
	b.WriteString(fmt.Sprintf("%s\n", entry.Message))

	return b.Bytes(), nil
}

type SelfJsonFormatter struct{}

func (j *SelfJsonFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := map[string]interface{}{
		"level":   strings.ToUpper(entry.Level.String()),
		"time":    entry.Time.Format(utils.DateTimeMS),
		"message": entry.Message,
	}

	// 添加自定义字段
	for k, v := range entry.Data {
		data[k] = v
	}

	r, _ := json.Marshal(data)

	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%s\n", r))

	return b.Bytes(), nil
}
