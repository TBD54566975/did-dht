package telemetry

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/TBD54566975/did-dht-method/config"
)

const (
	scopeName = "github.com/TBD54566975/did-dht-method"
)

var (
	tracer        trace.Tracer
	traceProvider *sdktrace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
	propagator    propagation.TextMapPropagator
)

// SetupTelemetry initializes the OpenTelemetry SDK with the appropriate exporters and propagators.
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

	// setup metrics
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return err
	}
	meterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	otel.SetMeterProvider(meterProvider)

	// setup memory metrics
	err = runtime.Start(runtime.WithMeterProvider(meterProvider), runtime.WithMinimumReadMemStatsInterval(15*time.Second))
	if err != nil {
		return err
	}

	// setup propagator
	propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)

	return nil
}

// Shutdown stops the telemetry providers and exporters safely.
func Shutdown(ctx context.Context) {
	if traceProvider != nil {
		if err := traceProvider.Shutdown(ctx); err != nil {
			logrus.WithError(err).Error("error shutting down trace provider")
		}
	}

	if meterProvider != nil {
		if err := meterProvider.Shutdown(ctx); err != nil {
			logrus.WithError(err).Error("error shutting down meter provider")
		}
	}
}

// GetTracer returns the tracer for the application. If the tracer is not yet initialized, it will be created.
func GetTracer() trace.Tracer {
	if tracer == nil {
		tracer = otel.GetTracerProvider().Tracer(scopeName, trace.WithInstrumentationVersion(config.Version))
	}
	return tracer
}
