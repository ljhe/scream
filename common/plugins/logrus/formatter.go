package logrus

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

const DateTime = "2006-01-02"
const DateTimeMS = "2006-01-02 15:04:05.000"

type SelfFormatter struct {
}

func (f *SelfFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	// 获取日志级别
	b.WriteString(fmt.Sprintf("[%s] ", entry.Level.String()))

	// 获取日志时间
	b.WriteString(fmt.Sprintf("%s ", entry.Time.Format(DateTimeMS)))

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
