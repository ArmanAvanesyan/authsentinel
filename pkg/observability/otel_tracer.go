package observability

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// NewOTLPTracerFromEnv returns a best-effort OTLP-backed tracer when OTEL env vars are set.
// If OTEL_EXPORTER_OTLP_ENDPOINT is missing (or initialization fails), it returns NopTracer.
//
// Env-only design: no schema/config changes required.
func NewOTLPTracerFromEnv() Tracer {
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		return NopTracer{}
	}

	protocol := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
	if protocol == "" {
		// OTel default is "grpc" for traces.
		protocol = "grpc"
	}

	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = "authsentinel"
	}

	// Parse endpoint into host/port and TLS mode.
	var (
		hostport string
		scheme   string
	)
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		hostport = u.Host
		scheme = u.Scheme
	} else {
		hostport = endpoint
	}

	// Best-effort initialization: never break the main auth flows.
	ctx := context.Background()
	var (
		exporter sdktrace.SpanExporter
		err      error
	)

	switch strings.ToLower(protocol) {
	case "http/protobuf", "http":
		httpOpts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(hostport)}
		if scheme == "http" {
			httpOpts = append(httpOpts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(ctx, httpOpts...)
	case "grpc":
		grpcOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(hostport)}
		if scheme == "http" {
			grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(ctx, grpcOpts...)
	default:
		// Fall back to grpc.
		grpcOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(hostport)}
		if scheme == "http" {
			grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(ctx, grpcOpts...)
	}
	if err != nil || exporter == nil {
		return NopTracer{}
	}

	res := resource.NewWithAttributes(
		// Schema-less resource attributes are fine for env-only usage.
		"",
		attribute.String("service.name", serviceName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	return &otelTracer{tracer: otel.Tracer(serviceName)}
}

// otelTracer adapts OpenTelemetry trace.Span/Tracer to the project observability.Tracer interface.
type otelTracer struct {
	tracer trace.Tracer
}

func (t *otelTracer) StartSpan(ctx context.Context, name string, keyvals ...any) (context.Context, Span) {
	attrs := keyvalsToAttributes(keyvals)
	ctx, sp := t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	return ctx, otelSpan{span: sp}
}

type otelSpan struct {
	span trace.Span
}

func (s otelSpan) End() { s.span.End() }

func keyvalsToAttributes(keyvals []any) []attribute.KeyValue {
	// Convention for this codebase: key/value pairs, e.g. ("operation","login").
	if len(keyvals) < 2 {
		return nil
	}
	out := make([]attribute.KeyValue, 0, len(keyvals)/2)
	for i := 0; i+1 < len(keyvals); i += 2 {
		k, ok := keyvals[i].(string)
		if !ok || strings.TrimSpace(k) == "" {
			continue
		}
		out = append(out, attributeFromAny(k, keyvals[i+1]))
	}
	return out
}

func attributeFromAny(key string, v any) attribute.KeyValue {
	switch x := v.(type) {
	case string:
		return attribute.String(key, x)
	case []byte:
		return attribute.String(key, string(x))
	case bool:
		return attribute.Bool(key, x)
	case int:
		return attribute.Int64(key, int64(x))
	case int64:
		return attribute.Int64(key, x)
	case uint64:
		return attribute.Int64(key, int64(x))
	case float64:
		return attribute.Float64(key, x)
	default:
		return attribute.String(key, fmt.Sprint(v))
	}
}

