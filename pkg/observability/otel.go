// Package observability wires up OpenTelemetry for the services: a single
// InitOTel call sets up distributed tracing, metrics, and logs, all exported
// over OTLP gRPC to the collector. Exporter endpoints/protocol are read from the
// standard OTEL_EXPORTER_OTLP_* environment variables.
package observability

import (
	"context"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Providers exposes the SDK providers callers may need to wire further
// instrumentation (e.g. the LoggerProvider for the Zap->OTLP bridge).
type Providers struct {
	LoggerProvider *sdklog.LoggerProvider
}

// InitOTel installs the global tracer/meter/logger providers and the W3C
// propagator, and starts Go runtime metrics collection. The returned shutdown
// function flushes and closes all providers and should be deferred in main().
func InitOTel(ctx context.Context, serviceName string) (*Providers, func(context.Context) error, error) {
	res, err := newResource(serviceName)
	if err != nil {
		return nil, nil, err
	}

	var shutdownFns []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var errs error
		for i := len(shutdownFns) - 1; i >= 0; i-- {
			errs = errors.Join(errs, shutdownFns[i](ctx))
		}
		return errs
	}
	fail := func(e error) (*Providers, func(context.Context) error, error) {
		_ = shutdown(ctx)
		return nil, nil, e
	}

	// W3C trace context + baggage propagation across the gateway -> services hop.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// ---- Traces ----
	traceExp, err := otlptracegrpc.New(ctx)
	if err != nil {
		return fail(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // dev: capture every request
	)
	otel.SetTracerProvider(tp)
	shutdownFns = append(shutdownFns, tp.Shutdown)

	// ---- Metrics ----
	metricExp, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return fail(err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(15*time.Second))),
	)
	otel.SetMeterProvider(mp)
	shutdownFns = append(shutdownFns, mp.Shutdown)

	// ---- Logs ----
	logExp, err := otlploggrpc.New(ctx)
	if err != nil {
		return fail(err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
	)
	otellogglobal.SetLoggerProvider(lp)
	shutdownFns = append(shutdownFns, lp.Shutdown)

	// ---- Go runtime metrics (GC, goroutines, memory) ----
	if err := runtime.Start(
		runtime.WithMeterProvider(mp),
		runtime.WithMinimumReadMemStatsInterval(15*time.Second),
	); err != nil {
		return fail(err)
	}

	return &Providers{LoggerProvider: lp}, shutdown, nil
}

// newResource builds the resource describing this service. The explicit
// service name/version take precedence; OTEL_RESOURCE_ATTRIBUTES and
// OTEL_SERVICE_NAME from the environment are merged in via resource.Environment.
func newResource(serviceName string) (*resource.Resource, error) {
	return resource.Merge(
		resource.Environment(),
		resource.NewSchemaless(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(getenv("OTEL_SERVICE_VERSION", "1.0.0")),
		),
	)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
