package log

import "go.uber.org/zap"

var dl *DefaultLogger

type DefaultLogger struct {
	Text *zap.Logger
	Json *zap.Logger
}

func NewDefaultLogger(opts ...Option) (*DefaultLogger, error) {
	var err error
	t, err := NewLogger(append(opts, WithSuffix(SuffixText))...)
	j, err := NewLogger(append(opts, WithSuffix(SuffixJson))...)
	dl = &DefaultLogger{
		Text: t,
		Json: j,
	}
	return dl, err
}

func (dl *DefaultLogger) Sync() {
	dl.Text.Sync()
	dl.Json.Sync()
}
