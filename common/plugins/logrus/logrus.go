package logrus

import (
	"bytes"
	"common/util"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
)

var logger = logrus.New()

func Init() {
	// 设置日志级别
	logger.SetLevel(logrus.TraceLevel)

	// 设置在输出日志中添加文件名和方法信息
	// 这个设置可能会增大开销
	// Note that this does add measurable overhead - the cost will depend on the version of Go,
	// but is between 20 and 40% in recent tests with 1.6 and 1.7
	logger.SetReportCaller(true)

	// 重定向输出
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("create file failed: %v", err)
	}
	logger.SetOutput(io.MultiWriter(&bytes.Buffer{}, os.Stdout, file))

	// 设置日志格式
	logger.SetFormatter(&ValueOnlyFormatter{})

	// 自定义默认的输出字段
	logger.WithFields(logrus.Fields{
		"ip": util.GetIPv4(),
	})
}
