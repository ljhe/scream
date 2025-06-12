package logrus

import (
	"github.com/ljhe/scream/utils"
)

const (
	DefaultLogName  = "test" // 日志文件名前缀
	DefaultLogLevel = 6      // 日志级别
	DefaultSavePath = "./"   // 日志保存路径
	DefaultMaxSize  = 512    // 每个日志文件的最大大小(MB)
	DefaultBackups  = 100    // 保留日志文件的最大数量(maxAge可能仍然会导致它们丢失)
	DefaultMaxAge   = 10     // 日志文件的最大保留天数
)

type Options struct {
	LogName    string
	LogLevel   uint32
	SavePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

func NewOptions(conf *logConf) *Options {
	return &Options{
		LogName:    utils.Ternary(conf.Log.LogName == "", DefaultLogName, conf.Log.LogName),
		LogLevel:   utils.Ternary(conf.Log.LogLevel == 0, DefaultLogLevel, conf.Log.LogLevel),
		SavePath:   utils.Ternary(conf.Log.SavePath == "", DefaultSavePath, conf.Log.SavePath),
		MaxSize:    utils.Ternary(conf.Log.MaxSize == 0, DefaultMaxSize, conf.Log.MaxSize),
		MaxBackups: utils.Ternary(conf.Log.MaxBackups == 0, DefaultBackups, conf.Log.MaxBackups),
		MaxAge:     utils.Ternary(conf.Log.MaxAge == 0, DefaultMaxAge, conf.Log.MaxAge),
	}
}
