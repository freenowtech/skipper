package datadog

import (
	ot "github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
)

func InitTracer(_ []string) (tracer ot.Tracer, err error) {
	return opentracer.New(), nil
}
