package telemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const scopeName = "github.com/TBD54566975/did-dht-method"

var (
	traceProvider *sdktrace.TracerProvider
	tracers       = make(map[string]trace.Tracer)
	tracersLock   sync.Mutex

	meterProvider *sdkmetric.MeterProvider
)

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
	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second * 30))
	if err != nil {
		return err
	}

	return nil
}

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

func GetTracer(subpackage string) trace.Tracer {
	tracersLock.Lock()
	defer tracersLock.Unlock()

	tracer, ok := tracers[subpackage]
	if !ok {
		name := fmt.Sprintf("%s/%s", scopeName, subpackage)
		tracer = otel.GetTracerProvider().Tracer(name, trace.WithInstrumentationVersion(config.Version))
		tracers[subpackage] = tracer
	}

	return tracer
}
