package telemetry

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const scopeName = "github.com/TBD54566975/did-dht-method"

var (
	Tracer        trace.Tracer
	traceProvider *sdktrace.TracerProvider

	Meter         metric.Meter
	meterProvider *sdkmetric.MeterProvider

	version = "unversioned"
)

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	version = buildInfo.Main.Version
}

func SetupTelemetry(ctx context.Context) error {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(scopeName)),
	)
	if err != nil {
		return err
	}

	// setup tracing
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return err
	}
	traceProvider = sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(r))
	otel.SetTracerProvider(traceProvider)
	Tracer = traceProvider.Tracer(scopeName, trace.WithInstrumentationVersion(version))

	// setup metrics
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return err
	}
	meterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	otel.SetMeterProvider(meterProvider)
	Meter = meterProvider.Meter(scopeName, metric.WithInstrumentationVersion(version))

	// setup memory metrics
	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second * 30))
	if err != nil {
		return err
	}

	return nil
}

func Shutdown(ctx context.Context) {
	if err := traceProvider.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("error shutting down trace provider")
	}

	if err := meterProvider.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("error shutting down meter provider")
	}
}
