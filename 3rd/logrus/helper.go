package logrus

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func Tracef(format string, args ...interface{}) {
	logText(getCaller(2)).Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logText(getCaller(2)).Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logText(getCaller(2)).Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	logText(getCaller(2)).Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logText(getCaller(2)).Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logText(getCaller(2)).Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logText(getCaller(2)).Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logText(getCaller(2)).Panicf(format, args...)
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("[%s:%d]", filepath.Base(file), line)
}
