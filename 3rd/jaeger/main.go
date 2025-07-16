package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"time"
)

func main() {

	var cfg = jaegercfg.Configuration{
		ServiceName: "client test", // 对其发起请求的的调用链，叫什么服务
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: "http://127.0.0.1:14268/api/traces",
		},
	}

	jLogger := jaegerlog.StdLogger
	tracer, closer, _ := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
	)

	// 创建第一个 span A
	parentSpan := tracer.StartSpan("A")
	// 调用其它服务
	B(tracer, parentSpan)
	time.Sleep(1 * time.Second)
	// 结束 A
	parentSpan.Finish()
	// 结束当前 tracer
	closer.Close()
}

func B(tracer opentracing.Tracer, parentSpan opentracing.Span) {
	// 继承上下文关系，创建子 span
	childSpan := tracer.StartSpan(
		"B",
		opentracing.ChildOf(parentSpan.Context()),
	)
	time.Sleep(1 * time.Second)
	defer childSpan.Finish()
}
