// Package telemetry provides OpenTelemetry integration for quix.
// It manages TracerProvider and MeterProvider lifecycle with unified Init/Shutdown.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds telemetry configuration.
type Config struct {
	ServiceName        string
	ServiceVersion     string
	ExporterEndpoint   string
	ResourceAttributes map[string]string
	TracesEnabled      bool
	MetricsEnabled     bool
	StdoutExporter     bool
}

// Option configures telemetry Config.
type Option func(*Config)

// WithServiceName sets the service name reported to OTel backend.
func WithServiceName(name string) Option {
	return func(c *Config) { c.ServiceName = name }
}

// WithServiceVersion sets the service version reported to OTel backend.
func WithServiceVersion(version string) Option {
	return func(c *Config) { c.ServiceVersion = version }
}

// WithExporterEndpoint sets the OTLP exporter endpoint address.
func WithExporterEndpoint(endpoint string) Option {
	return func(c *Config) { c.ExporterEndpoint = endpoint }
}

// WithResourceAttributes adds custom resource attributes (key-value pairs).
func WithResourceAttributes(kv ...string) Option {
	return func(c *Config) {
		if c.ResourceAttributes == nil {
			c.ResourceAttributes = make(map[string]string)
		}
		for i := 0; i+1 < len(kv); i += 2 {
			c.ResourceAttributes[kv[i]] = kv[i+1]
		}
	}
}

// WithTracesEnabled controls whether Traces are enabled.
func WithTracesEnabled(enabled bool) Option {
	return func(c *Config) { c.TracesEnabled = enabled }
}

// WithMetricsEnabled controls whether Metrics are enabled.
func WithMetricsEnabled(enabled bool) Option {
	return func(c *Config) { c.MetricsEnabled = enabled }
}

// WithStdoutExporter uses stdout exporters instead of OTLP (for development).
func WithStdoutExporter(enabled bool) Option {
	return func(c *Config) { c.StdoutExporter = enabled }
}

// Init initializes OTel providers and returns a unified shutdown function.
// The shutdown func flushes all providers in order: MeterProvider, TracerProvider.
func Init(ctx context.Context, opts ...Option) (func(context.Context) error, error) {
	cfg := &Config{
		ServiceName:      "unknown_service",
		ExporterEndpoint: "localhost:4317",
		TracesEnabled:    true,
		MetricsEnabled:   true,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("telemetry: create resource: %w", err)
	}

	var (
		tracerProvider *sdktrace.TracerProvider
		meterProvider  *metric.MeterProvider
	)

	if cfg.TracesEnabled {
		tp, err := newTracerProvider(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("telemetry: create tracer provider: %w", err)
		}
		tracerProvider = tp
		otel.SetTracerProvider(tp)
	}

	if cfg.MetricsEnabled {
		mp, err := newMeterProvider(ctx, res, cfg)
		if err != nil {
			// Rollback tracer provider if already created
			if tracerProvider != nil {
				_ = tracerProvider.Shutdown(ctx)
			}
			return nil, fmt.Errorf("telemetry: create meter provider: %w", err)
		}
		meterProvider = mp
		otel.SetMeterProvider(mp)
	}

	shutdown := func(ctx context.Context) error {
		var errs []error
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
			}
		}
		if tracerProvider != nil {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("telemetry shutdown errors: %v", errs)
		}
		return nil
	}

	return shutdown, nil
}

func newResource(_ context.Context, cfg *Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(cfg.ServiceVersion))
	}
	for k, v := range cfg.ResourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
	}
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}

func newTracerProvider(ctx context.Context, res *resource.Resource, cfg *Config) (*sdktrace.TracerProvider, error) {
	if cfg.StdoutExporter {
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}
		return sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(exp),
		), nil
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp),
	), nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource, cfg *Config) (*metric.MeterProvider, error) {
	if cfg.StdoutExporter {
		exp, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout metric exporter: %w", err)
		}
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(metric.NewPeriodicReader(exp)),
		), nil
	}

	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.ExporterEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp metric exporter: %w", err)
	}
	return metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)),
	), nil
}

// ExtractTraceID extracts OTel trace_id from context.
// Returns empty string if no trace context is present.
func ExtractTraceID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}
