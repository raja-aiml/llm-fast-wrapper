package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func Init(service string) (func(context.Context) error, error) {
	exp, err := otlptracehttp.New(context.Background())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
		)),
	)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(service)
	log.Println("otel initialized")
	return tp.Shutdown, nil
}

func Tracer() trace.Tracer { return tracer }
