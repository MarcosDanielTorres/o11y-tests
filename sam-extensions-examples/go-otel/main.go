package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"google.golang.org/protobuf/proto"
	//"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	v1 "go.opentelemetry.io/proto/otlp/collector/logs/v1" // Ensure correct

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"

	logs "go.opentelemetry.io/proto/otlp/logs/v1" // Import for ResourceLogs
	protoresourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
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

	logger := otelslog.NewLogger("my/pkg/name", otelslog.WithLoggerProvider(loggerProvider))
	slog.SetDefault(logger)

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

func newLoggerProvider(resource *resource.Resource) (*sdklog.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(context.Background(), otlploghttp.WithInsecure(), otlploghttp.WithEndpointURL("http://localhost:4318/v1/logs"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	processor := sdklog.NewBatchProcessor(logExporter, sdklog.WithExportInterval(NearlyImmediate))
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(processor),
	)
	return loggerProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
func sendToLoki() {
	// Define the JSON structure
	data := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"label": "value",
				},
				"values": [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), "First log line"},
					{fmt.Sprintf("%d", time.Now().UnixNano()), "Second log line"},
				},
			},
		},
	}

	// Encode JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Define the POST URL
	url := "http://localhost:3100/loki/api/v1/push"

	// Create a POST request with the JSON data
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the http.DefaultClient
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("Request successful")
	} else {
		fmt.Printf("Failed to send request, status code: %d\n", resp.StatusCode)
	}
}

func sendToOtel() {
	//data := map[string]interface{}{
	//	"streams": []map[string]interface{}{
	//		{
	//			"stream": map[string]string{
	//				"label": "value",
	//			},
	//			"values": [][]string{
	//				{fmt.Sprintf("%d", time.Now().UnixNano()), "First log line"},
	//				{fmt.Sprintf("%d", time.Now().UnixNano()), "Second log line"},
	//			},
	//		},
	//	},
	//}

	request := &v1.ExportLogsServiceRequest{
		ResourceLogs: []*logs.ResourceLogs{
			{
				Resource: &protoresourcev1.Resource{
					Attributes: []*commonv1.KeyValue{
						{Key: "service.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "my-service"}}},
					},
				},
				ScopeLogs: []*logs.ScopeLogs{
					{Scope: &commonv1.InstrumentationScope{Name: "hola", Version: "1.1.1", Attributes: []*commonv1.KeyValue{
						{Key: "log.attributes", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "some-log-value"}}},
					}}, LogRecords: []*logs.LogRecord{
						{
							SeverityNumber: logs.SeverityNumber_SEVERITY_NUMBER_FATAL,
							Body:           &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: 322}},
						},
					}},
				},
			},
		},
	}

	// Encode JSON data
	jsonData, err := proto.Marshal(request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	url := "http://localhost:4318/v1/logs"

	// Create a POST request with the JSON data
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the appropriate headers
	//req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Type", "application/x-protobuf")

	// Send the request using the http.DefaultClient
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	fmt.Println("OTEL response status code: ", resp.StatusCode)
	fmt.Println("OTEL response header", resp.Header)
	fmt.Println("OTEL response body", resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Error reading body")
	}
	fmt.Println("Raw response body:", string(body))
	var response v1.ExportLogsServiceResponse
	//json.Unmarshal(body, &response)
	proto.Unmarshal(body, &response)
	fmt.Println("OTEL response body unmarshaled", response.String())

}

func main() {
	ctx := context.WithoutCancel(context.Background())
	shutdown, err := initOtel(ctx)
	if err != nil {
		log.Fatal("Error initializing OpenTelemetry: ", err)
	}

	exampleJson := []byte(`{"key1": "value1", "key2": "value2"}`)

	rec := otellog.Record{}
	rec.SetTimestamp(time.Now())
	//rec.SetBody(otellog.StringValue("Hello, World!"))
	//rec.SetBody(otellog.BytesValue(exampleJson))
	rec.SetBody(otellog.StringValue(string(exampleJson)))
	//global.GetLoggerProvider().Logger("--some-logger-name--").Emit(ctx, rec)

	//slog.Info("Info LOG")
	//slog.Debug("Debug LOG")
	//slog.Info("Info LOG")
	//slog.Warn("Warn LOG")
	//slog.Error("Error LOG")

	fmt.Println("Hello world how are you going?")

	//sendToLoki()
	sendToOtel()

	shutdown(context.Background())
}
