package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	//"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otelLog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

const NearlyImmediate = 100 * time.Millisecond

func initOtel(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	// its not needed for now
	//prop := newPropagator()
	//otel.SetTextMapPropagator(prop)

	resource, err := createResource()
	if err != nil {
		handleErr(err)
		return
	}

	loggerProvider, err := newLoggerProvider(resource)
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	slog.SetDefault(otelslog.NewLogger("my/pkg/name", otelslog.WithLoggerProvider(loggerProvider)))

	return
}

func createResource() (*resource.Resource, error) {
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attribute.String("service.instance.id", "go-otel")),
	)
	if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
		fmt.Println("Logging non-fatal error: ", err)
	} else if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	return res, nil
}

func newLoggerProvider(resource *resource.Resource) (*otelLog.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(context.Background(), otlploghttp.WithInsecure(), otlploghttp.WithEndpointURL("http://localhost:4318/v1/logs"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	processor := otelLog.NewBatchProcessor(logExporter, otelLog.WithExportInterval(NearlyImmediate))
	loggerProvider := otelLog.NewLoggerProvider(
		otelLog.WithResource(resource),
		otelLog.WithProcessor(processor),
	)
	return loggerProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func main() {
	ctx := context.WithoutCancel(context.Background())
	shutdown, err := initOtel(ctx)
	if err != nil {
		log.Fatal("Error initializing OpenTelemetry: ", err)
	}

	slog.Info("Info LOG")
	slog.Debug("Debug LOG")
	slog.Info("Info LOG")
	slog.Warn("Warn LOG")
	slog.Error("Error LOG")

	shutdown(context.Background())
}
