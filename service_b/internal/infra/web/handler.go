package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rodrigoasouza93/otel-service-b/configs"
	"github.com/rodrigoasouza93/otel-service-b/internal/dto"
	"github.com/rodrigoasouza93/otel-service-b/internal/vo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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
	router.Get("/{cep}", we.getWeatherHandler)
	return router
}

func (we *Webserver) getWeatherHandler(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctxGlobal, spanGlobal := we.OTELTracer.Start(ctx, "get-location-weather")
	rawCep := r.PathValue("cep")
	cep, err := vo.NewCep(rawCep)
	if err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}
	_, spanLocation := we.OTELTracer.Start(ctxGlobal, "location")
	locationURL := "https://viacep.com.br/ws/" + cep.Value() + "/json"
	req, err := http.NewRequestWithContext(ctxGlobal, http.MethodGet, locationURL, nil)
	otel.GetTextMapPropagator().Inject(ctxGlobal, propagation.HeaderCarrier(req.Header))
	if err != nil {
		http.Error(w, "error creating request", http.StatusInternalServerError)
		return
	}
	respLocation, err := http.DefaultClient.Do(req)
	if err != nil || respLocation.StatusCode != http.StatusOK {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	defer respLocation.Body.Close()

	var decodedLocation dto.LocationResponse
	err = json.NewDecoder(respLocation.Body).Decode(&decodedLocation)
	if err != nil {
		http.Error(w, "error decoding location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if decodedLocation.Error {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	spanLocation.End()

	config := configs.LoadConfig(".")
	_, spanWeather := we.OTELTracer.Start(ctxGlobal, "weather")
	weatherAPIURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", config.WeatherAPIKey, url.QueryEscape(decodedLocation.Locale))
	reqWeather, err := http.NewRequestWithContext(ctxGlobal, http.MethodGet, weatherAPIURL, nil)
	otel.GetTextMapPropagator().Inject(ctxGlobal, propagation.HeaderCarrier(reqWeather.Header))
	if err != nil {
		http.Error(w, "error creating request", http.StatusInternalServerError)
		return
	}
	respWeather, err := http.DefaultClient.Do(reqWeather)
	if respWeather.StatusCode != http.StatusOK || err != nil {
		http.Error(w, "can not get weather", respWeather.StatusCode)
		return
	}
	defer respWeather.Body.Close()

	var decodedWeather dto.WeatherResponse
	if err := json.NewDecoder(respWeather.Body).Decode(&decodedWeather); err != nil {
		http.Error(w, "error decoding weather: "+err.Error(), http.StatusInternalServerError)
		return
	}
	spanWeather.End()
	spanGlobal.End()
	response := dto.WeatherOutput{
		Temp_C: decodedWeather.Current.TempC,
		Temp_F: decodedWeather.Current.TempF,
		Temp_K: getKelvinTemp(decodedWeather.Current.TempC),
		City:   decodedLocation.Locale,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getKelvinTemp(celsius float64) float64 {
	return celsius + 273.15
}
