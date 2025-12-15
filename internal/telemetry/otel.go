package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"workctl/internal/config"
)

// Setup initializes the OpenTelemetry trace and metric providers.
// It returns a shutdown function that should be called when the application exits.
func Setup(ctx context.Context, cfg *config.OtelConfig) (func(context.Context) error, error) {
	// If no endpoints are configured, we can skip setup or set up specific no-op providers if desired.
	// For now, let's assume if config is missing critical parts we might error or just return.
	// But let's check against empty endpoints to be safe.
	if cfg.Exporter.Traces.Endpoint == "" && cfg.Exporter.Metrics.Endpoint == "" {
		// Nothing to configure
		return func(_ context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("workctl"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error

	// Shutdown function that calls all registered shutdown functions.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			if e := fn(ctx); e != nil {
				err = e
			}
		}
		return err
	}

	// 1. Trace Provider
	if cfg.Exporter.Traces.Endpoint != "" {
		traceExporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(cfg.Exporter.Traces.Endpoint),
			otlptracehttp.WithHeaders(cfg.Exporter.Traces.Headers),
			// Ideally OTLP HTTP uses https by default unless http scheme is used?
			// otlptracehttp handles scheme parsing if passed correctly?
			// otlptracehttp.WithEndpoint expects "host:port", scheme is handled by WithInsecure or implied?
			// Actually otlptracehttp.WithEndpoint docs say "The endpoint should not contain a scheme prefix".
			// If the user provided http:// or https:// we should strip it or parse it.
			// But wait, the user example showed just an endpoint.
			// Let's assume standard behavior. Typically `WithInsecure` is needed for http.
			// For now, let's just use what's given. If the user puts "https://blah", we might need to be careful.
			// Safe approach: Config usually specifies just host:port.
			// Let's assume the user knows what they are doing for now or refine later.
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})

		shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	}

	// 2. Metric Provider
	if cfg.Exporter.Metrics.Endpoint != "" {
		metricExporter, err := otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpointURL(cfg.Exporter.Metrics.Endpoint),
			otlpmetrichttp.WithHeaders(cfg.Exporter.Metrics.Headers),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}

		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(60*time.Second))),
			sdkmetric.WithResource(res),
		)
		otel.SetMeterProvider(mp)

		shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	}

	return shutdown, nil
}
