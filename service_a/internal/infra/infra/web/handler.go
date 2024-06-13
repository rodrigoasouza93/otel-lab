package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rodrigoasouza93/otel-service-a/internal/application/dto"
	"github.com/rodrigoasouza93/otel-service-a/internal/domain/vo"
	"go.opentelemetry.io/otel/trace"
)

type Webserver struct {
	OTELTracer trace.Tracer
}

func NewServer(otelTracer trace.Tracer) *Webserver {
	return &Webserver{
		OTELTracer: otelTracer,
	}
}

func (we *Webserver) CreateServer() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))
	// promhttp
	router.Post("/", we.Handle)
	return router
}

func (we *Webserver) Handle(w http.ResponseWriter, r *http.Request) {
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

	ctx, span := we.OTELTracer.Start(r.Context(), "get-service-b-info")
	var output dto.DTOOutput
	output, status, err := getInfo(cep.Value, ctx)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	span.End()
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
