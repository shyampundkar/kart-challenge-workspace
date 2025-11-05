package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Config holds telemetry configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	EnableMetrics  bool
}

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(config Config) (func(context.Context) error, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	log.Printf("OpenTelemetry tracer initialized for service: %s", config.ServiceName)

	// Return shutdown function
	return tp.Shutdown, nil
}

// InitMetrics initializes the OpenTelemetry metrics
func InitMetrics(config Config) (func(context.Context) error, error) {
	if !config.EnableMetrics {
		return func(ctx context.Context) error { return nil }, nil
	}

	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	log.Printf("OpenTelemetry metrics initialized for service: %s", config.ServiceName)

	// Return shutdown function
	return mp.Shutdown, nil
}

// InitTelemetry initializes both tracing and metrics
func InitTelemetry(config Config) (func(context.Context) error, error) {
	// Initialize tracer
	shutdownTracer, err := InitTracer(config)
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	shutdownMetrics, err := InitMetrics(config)
	if err != nil {
		return nil, err
	}

	// Return combined shutdown function
	shutdown := func(ctx context.Context) error {
		var errs []error

		if err := shutdownTracer(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer shutdown error: %w", err))
		}

		if err := shutdownMetrics(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metrics shutdown error: %w", err))
		}

		if len(errs) > 0 {
			return fmt.Errorf("telemetry shutdown errors: %v", errs)
		}

		return nil
	}

	return shutdown, nil
}

// GetConfig returns telemetry configuration from environment variables
func GetConfig(serviceName string) Config {
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces"
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	serviceVersion := os.Getenv("SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	enableMetrics := os.Getenv("ENABLE_METRICS") != "false"

	return Config{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    environment,
		JaegerEndpoint: jaegerEndpoint,
		EnableMetrics:  enableMetrics,
	}
}

// GracefulShutdown handles graceful shutdown with timeout
func GracefulShutdown(shutdownFunc func(context.Context) error, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := shutdownFunc(ctx); err != nil {
		log.Printf("Error during telemetry shutdown: %v", err)
	} else {
		log.Println("Telemetry shutdown successfully")
	}
}
