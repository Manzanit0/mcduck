package trace

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type Provider struct {
	serviceName string
	*sdktrace.TracerProvider
}

func (tp Provider) TraceRequests() gin.HandlerFunc {
	return otelgin.Middleware(tp.serviceName)
}

func TracerFromEnv(ctx context.Context, service string) (*Provider, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if endpoint == "" || headers == "" {
		return nil, fmt.Errorf("missing OTEL_EXPORTER_* environment variables")
	}

	opts := NewExporterOptions(endpoint, headers)
	tp, err := InitTracer(ctx, service, opts)
	if err != nil {
		return nil, fmt.Errorf("init tracer: %s", err.Error())
	}

	return tp, nil
}

// InitTracer configures an exporter that will send spans to Honeycomb
// It should close the trace provider after.
// defer func() { _ = tp.Shutdown(ctx) }()
func InitTracer(ctx context.Context, service string, opts []otlptracegrpc.Option) (*Provider, error) {
	client := otlptracegrpc.NewClient(opts...)
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("initialize exporter: %w", err)
	}

	// The service.name attribute is required.
	resource :=
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
		)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource),
	)

	// Set the Tracer Provider and the W3C Trace Context propagator as globals
	otel.SetTracerProvider(tp)

	// Register the trace context and baggage propagators so data is propagated across services/processes.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &Provider{serviceName: service, TracerProvider: tp}, nil
}

func NewExporterOptions(endpoint, headers string) []otlptracegrpc.Option {
	headersMap := parseOTELHeaders(headers)
	return []otlptracegrpc.Option{
		otlptracegrpc.WithTimeout(5 * time.Second),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithHeaders(headersMap),
		otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})),
	}
}

func parseOTELHeaders(headers string) map[string]string {
	headersMap := make(map[string]string)
	if len(headers) > 0 {
		headerItems := strings.Split(headers, ",")
		for _, headerItem := range headerItems {
			parts := strings.Split(headerItem, "=")
			headersMap[parts[0]] = parts[1]
		}
	}
	return headersMap
}
