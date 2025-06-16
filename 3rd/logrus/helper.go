package logrus

func Tracef(format string, args ...interface{}) {
	logText.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logText.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logText.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	logText.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logText.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logText.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logText.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logText.Panicf(format, args...)
}
