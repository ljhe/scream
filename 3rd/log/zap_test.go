package log

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func Benchmark_ZapLog(b *testing.B) {
	_, _ = NewDefaultLogger(func(options *Options) error {
		options.OutStd = false
		return nil
	})
	str := strings.Repeat("a", 214)
	b.Run("NormalInfoF", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InfoF("this is test. str:%s", str)
		}
	})

	b.Run("NormalInfoF", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InfoSprintf("this is test. str:%s", str)
		}
	})

	b.Run("NormalInfoKV", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InfoKV("this is test.", zap.String("str", str))
		}
	})
}

func TestInfos(t *testing.T) {
	_, _ = NewDefaultLogger()
	t.Run("InfoKV", func(t *testing.T) {
		InfoKV("the color is write.")
	})
	t.Run("InfoF", func(t *testing.T) {
		InfoF("the color is write.")
	})
	t.Run("WarnF", func(t *testing.T) {
		WarnF("the color is yellow.")
	})
	t.Run("ErrorF", func(t *testing.T) {
		ErrorF("the color is yellow.")
	})
}
