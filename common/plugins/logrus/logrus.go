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

func logrusInit() {
	// 设置日志级别
	//logger.SetLevel(logrus.TraceLevel)
	logger.SetLevel(logrus.Level(conf.Log.LogLevel))

	// 设置在输出日志中添加文件名和方法信息
	// 这个设置可能会增大开销
	// Note that this does add measurable overhead - the cost will depend on the version of Go,
	// but is between 20 and 40% in recent tests with 1.6 and 1.7
	logger.SetReportCaller(true)

	// 配置日志切割
	logger.SetOutput(io.MultiWriter(&bytes.Buffer{}, os.Stdout, &lumberjack.Logger{
		Filename:   fmt.Sprintf("%v/%v_%v.log", conf.Log.SavePath, conf.Log.LogName, time.Now().Format(DateTime)),
		MaxSize:    conf.Log.MaxSize,    // 每个日志文件的最大大小(MB)
		MaxBackups: conf.Log.MaxBackups, // 保留日志文件的最大数量(maxAge可能仍然会导致它们丢失)
		MaxAge:     conf.Log.MaxAge,     // 日志文件的最大保留天数
		LocalTime:  true,
	}))

	// 设置日志格式
	logger.SetFormatter(&SelfFormatter{})
}
