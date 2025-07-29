package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
	"strings"
)

const (
	logTypText = iota
	logTypJson
)

func DebugF(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	log(logTypText, zapcore.DebugLevel, str)
}

func DebugKV(format string, v ...zap.Field) {
	log(logTypJson, zapcore.DebugLevel, format, v...)
}

func InfoF(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	log(logTypText, zapcore.InfoLevel, str)
}

func InfoKV(format string, v ...zap.Field) {
	log(logTypJson, zapcore.InfoLevel, format, v...)
}

func ErrorF(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	log(logTypText, zapcore.ErrorLevel, str)
}

func WarnF(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	log(logTypText, zapcore.WarnLevel, str)
}

func PanicF(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	log(logTypText, zapcore.PanicLevel, str)
}

func log(typ int, level zapcore.Level, msg string, v ...zap.Field) {
	logger := getLogger(typ)
	if !logger.Core().Enabled(level) {
		return
	}

	var stackTrace zapcore.Field
	if level == zapcore.ErrorLevel || level == zapcore.WarnLevel {
		stackInfo := getStackTrace()
		stackTrace = zap.String("stack_trace", stackInfo)

		// 控制台输出，使用不同颜色区分
		if level == zapcore.ErrorLevel {
			fmt.Printf("\033[31m[ERROR] %s\n%s\033[0m\n", msg, stackInfo) // 红色
		} else {
			fmt.Printf("\033[33m[WARN] %s\n%s\033[0m\n", msg, stackInfo) // 黄色
		}
	}

	switch level {
	case zapcore.DebugLevel:
		logger.Debug(msg, v...)
	case zapcore.InfoLevel:
		logger.Info(msg, v...)
	case zapcore.WarnLevel:
		logger.Warn(msg, append(v, stackTrace)...)
	case zapcore.ErrorLevel:
		logger.Error(msg, append(v, stackTrace)...)
	case zapcore.DPanicLevel:
		logger.DPanic(msg, v...)
	case zapcore.PanicLevel:
		logger.Panic(msg, v...)
	case zapcore.FatalLevel:
		logger.Fatal(msg, v...)
	}
}

func getStackTrace() string {
	// 分配调用栈空间
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	builder.WriteString("\nStack Trace:\n")

	for {
		frame, more := frames.Next()
		builder.WriteString(fmt.Sprintf("%s\n\t%s:%d\n",
			frame.Function,
			frame.File,
			frame.Line))
		if !more {
			break
		}
	}
	return builder.String()
}

func getLogger(typ int) *zap.Logger {
	if typ == logTypText {
		return dl.Text
	}
	return dl.Json
}
