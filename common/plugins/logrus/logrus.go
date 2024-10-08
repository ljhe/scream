package logrus

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"time"
)

var logger = logrus.New()
var entry *logrus.Entry

func Init() {
	// 设置日志级别
	logger.SetLevel(logrus.TraceLevel)

	// 设置在输出日志中添加文件名和方法信息
	// 这个设置可能会增大开销
	// Note that this does add measurable overhead - the cost will depend on the version of Go,
	// but is between 20 and 40% in recent tests with 1.6 and 1.7
	logger.SetReportCaller(true)

	// 配置日志切割
	logger.SetOutput(io.MultiWriter(&bytes.Buffer{}, os.Stdout, &lumberjack.Logger{
		Filename:   fmt.Sprintf("test-%v.log", time.Now().Format(DateTime)),
		MaxSize:    512, // 每个日志文件的最大大小(MB)
		MaxBackups: 100,
		MaxAge:     10, // 日志文件的最大保留天数
		LocalTime:  true,
	}))

	// 设置日志格式
	logger.SetFormatter(&SelfFormatter{})

	// 自定义默认的输出字段
	entry = logger.WithFields(logrus.Fields{
		"IP": "127.0.0.1",
	})
}
