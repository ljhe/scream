package logrus

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
)

const DateTimeMS = "2006-01-02 15:04:05.000"

type ValueOnlyFormatter struct {
}

func (f *ValueOnlyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	// 获取日志级别
	b.WriteString(fmt.Sprintf("[%s] ", entry.Level.String()))

	// 获取日志时间
	b.WriteString(fmt.Sprintf("%s ", entry.Time.Format(DateTimeMS)))

	// 打印日志的位置信息
	// 需要SetReportCaller(true) 否则这里会报错
	b.WriteString(fmt.Sprintf("%s:%d ", entry.Caller.File, entry.Caller.Line))

	// 日志消息的值
	b.WriteString(fmt.Sprintf("%s\n", entry.Message))

	return b.Bytes(), nil
}