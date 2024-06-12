package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rodrigoasouza93/otel-service-a/internal/application/dto"
	"github.com/rodrigoasouza93/otel-service-a/internal/domain/vo"
	"github.com/rodrigoasouza93/otel-service-a/internal/infra/infra/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initProvider(serviceName string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initProvider("weather-tracer")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	tracer := otel.Tracer("microservice-tracer")
	server := web.NewServer(tracer)
	router := server.CreateServer()
	go func() {
		log.Println("Starting server on port", "8080")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatal(err)
		}
	}()

	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+C pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due to other reason...")
	}

	// Create a timeout context for the graceful shutdown
	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var input dto.DTOInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cep, err := vo.NewCep(input.Cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("weather-tracer").Start(r.Context(), "get-service-b")
	defer span.End()

	var output dto.DTOOutput
	output, status, err := getInfo(cep.Value, ctx)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func getInfo(cep string, ctx context.Context) (dto.DTOOutput, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://goapp-service-b:8081/"+cep, nil)
	var output dto.DTOOutput
	if err != nil {
		return output, http.StatusInternalServerError, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return output, resp.StatusCode, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return output, resp.StatusCode, errors.New("can not find zipcode")
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return output, resp.StatusCode, err
	}
	return output, resp.StatusCode, err
}
