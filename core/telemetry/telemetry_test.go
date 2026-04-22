package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func TestWithServiceName(t *testing.T) {
	cfg := &Config{ServiceName: "default"}
	WithServiceName("myapp")(cfg)
	assert.Equal(t, "myapp", cfg.ServiceName)
}

func TestWithServiceVersion(t *testing.T) {
	cfg := &Config{ServiceVersion: ""}
	WithServiceVersion("1.0.0")(cfg)
	assert.Equal(t, "1.0.0", cfg.ServiceVersion)
}

func TestWithExporterEndpoint(t *testing.T) {
	cfg := &Config{ExporterEndpoint: "localhost:4317"}
	WithExporterEndpoint("collector:4317")(cfg)
	assert.Equal(t, "collector:4317", cfg.ExporterEndpoint)
}

func TestWithResourceAttributes(t *testing.T) {
	cfg := &Config{ResourceAttributes: nil}
	WithResourceAttributes("env", "prod", "region", "cn")(cfg)
	assert.Equal(t, "prod", cfg.ResourceAttributes["env"])
	assert.Equal(t, "cn", cfg.ResourceAttributes["region"])
}

func TestWithResourceAttributesOddArgs(t *testing.T) {
	cfg := &Config{ResourceAttributes: nil}
	WithResourceAttributes("env", "prod", "extra")(cfg)
	assert.Equal(t, "prod", cfg.ResourceAttributes["env"])
	// Odd trailing key is ignored
	_, ok := cfg.ResourceAttributes["extra"]
	assert.False(t, ok)
}

func TestWithTracesEnabled(t *testing.T) {
	cfg := &Config{TracesEnabled: true}
	WithTracesEnabled(false)(cfg)
	assert.False(t, cfg.TracesEnabled)
}

func TestWithMetricsEnabled(t *testing.T) {
	cfg := &Config{MetricsEnabled: true}
	WithMetricsEnabled(false)(cfg)
	assert.False(t, cfg.MetricsEnabled)
}

func TestWithStdoutExporter(t *testing.T) {
	cfg := &Config{StdoutExporter: false}
	WithStdoutExporter(true)(cfg)
	assert.True(t, cfg.StdoutExporter)
}

func TestInitWithStdoutExporter(t *testing.T) {
	_, shutdown, err := Init(context.Background(),
		WithServiceName("test-app"),
		WithStdoutExporter(true),
	)
	require.NoError(t, err)
	require.NotNil(t, shutdown)

	// Verify global providers are set
	assert.NotNil(t, otel.GetTracerProvider())
	assert.NotNil(t, otel.GetMeterProvider())

	// Shutdown should succeed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NoError(t, shutdown(ctx))
}

func TestInitWithTracesDisabled(t *testing.T) {
	_, shutdown, err := Init(context.Background(),
		WithServiceName("test-app"),
		WithStdoutExporter(true),
		WithTracesEnabled(false),
		WithMetricsEnabled(false),
	)
	require.NoError(t, err)
	require.NotNil(t, shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NoError(t, shutdown(ctx))
}

func TestShutdownOrder(t *testing.T) {
	_, shutdown, err := Init(context.Background(),
		WithServiceName("test-app"),
		WithStdoutExporter(true),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NoError(t, shutdown(ctx))
	// Second shutdown may return errors (providers already shut down)
	err = shutdown(ctx)
	assert.Error(t, err) // Expected: providers already shut down
}

func TestNewResource(t *testing.T) {
	cfg := &Config{
		ServiceName:    "myapp",
		ServiceVersion: "1.0.0",
		ResourceAttributes: map[string]string{
			"env": "production",
		},
	}
	res, err := newResource(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, res)

	set := res.SchemaURL()
	assert.Equal(t, semconv.SchemaURL, set)

	attrs := attribute.NewSet(res.Attributes()...)
	val, ok := attrs.Value(semconv.ServiceNameKey)
	assert.True(t, ok)
	assert.Equal(t, "myapp", val.AsString())

	val, ok = attrs.Value(semconv.ServiceVersionKey)
	assert.True(t, ok)
	assert.Equal(t, "1.0.0", val.AsString())

	val, ok = attrs.Value("env")
	assert.True(t, ok)
	assert.Equal(t, "production", val.AsString())
}

func TestDefaultConfig(t *testing.T) {
	cfg := &Config{}
	// No options applied — verify defaults are not used here
	// (defaults are set in Init, not in Config zero value)
	assert.Empty(t, cfg.ServiceName)
	assert.Empty(t, cfg.ExporterEndpoint)
	assert.False(t, cfg.TracesEnabled)
	assert.False(t, cfg.MetricsEnabled)
	assert.False(t, cfg.StdoutExporter)
}

func TestInitSetsGlobalProviders(t *testing.T) {
	// Save original providers to restore after test
	origTP := otel.GetTracerProvider()
	origMP := otel.GetMeterProvider()

	_, shutdown, err := Init(context.Background(),
		WithServiceName("provider-test"),
		WithStdoutExporter(true),
	)
	require.NoError(t, err)

	// Verify new providers are different from originals
	assert.NotEqual(t, origTP, otel.GetTracerProvider())
	assert.NotEqual(t, origMP, otel.GetMeterProvider())

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, shutdown(ctx))

	// Restore original providers
	otel.SetTracerProvider(origTP)
	otel.SetMeterProvider(origMP)
}

func TestExtractTraceID(t *testing.T) {
	// Without trace context, should return empty
	traceID := ExtractTraceID(context.Background())
	assert.Empty(t, traceID)
}

func TestExtractSpanID(t *testing.T) {
	// Without trace context, should return empty
	spanID := ExtractSpanID(context.Background())
	assert.Empty(t, spanID)
}

func TestNewResourceMerge(t *testing.T) {
	custom := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("custom-service"),
	)
	merged, err := resource.Merge(resource.Default(), custom)
	require.NoError(t, err)
	assert.NotNil(t, merged)
}
