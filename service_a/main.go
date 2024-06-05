package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/rodrigoasouza93/otel-service-a/internal/dto"
	"github.com/rodrigoasouza93/otel-service-a/internal/vo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	setTracing()
	http.HandleFunc("POST /", Handle)
	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
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

	ctx, span := otel.Tracer("service-a").Start(r.Context(), "get-service-b")
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

func setTracing() {
	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("Fail to create Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-a"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
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
