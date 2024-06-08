package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/rodrigoasouza93/otel-service-b/configs"
	"github.com/rodrigoasouza93/otel-service-b/internal/dto"
	"github.com/rodrigoasouza93/otel-service-b/internal/vo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	setTracing()
	http.HandleFunc("GET /{cep}", getWeatherHandler)
	log.Default().Printf("Listening on port 8081")
	http.ListenAndServe(":8081", nil)
}

func getWeatherHandler(w http.ResponseWriter, r *http.Request) {
	ctx, spanCep := otel.Tracer("service-b").Start(r.Context(), "get-cep")
	defer spanCep.End()
	rawCep := r.PathValue("cep")
	cep, err := vo.NewCep(rawCep)
	if err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}
	locationURL := "https://viacep.com.br/ws/" + cep.Value() + "/json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, locationURL, nil)
	if err != nil {
		http.Error(w, "error creating request", http.StatusInternalServerError)
		return
	}
	respLocation, err := http.DefaultClient.Do(req)
	if err != nil || respLocation.StatusCode != http.StatusOK {
		fmt.Println(respLocation)
		fmt.Println(err)
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

	ctx, spanWeather := otel.Tracer("service-b").Start(r.Context(), "get-wather")
	defer spanWeather.End()
	config := configs.LoadConfig(".")
	weatherAPIURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", config.WeatherAPIKey, url.QueryEscape(decodedLocation.Locale))
	reqWeather, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherAPIURL, nil)
	if err != nil {
		http.Error(w, "error creating request", http.StatusInternalServerError)
		return
	}
	respWeather, err := http.DefaultClient.Do(reqWeather)
	if respWeather.StatusCode != http.StatusOK || err != nil {
		fmt.Println(weatherAPIURL)

		http.Error(w, "can not get weather", respWeather.StatusCode)
		return
	}
	defer respWeather.Body.Close()

	var decodedWeather dto.WeatherResponse
	if err := json.NewDecoder(respWeather.Body).Decode(&decodedWeather); err != nil {
		http.Error(w, "error decoding weather: "+err.Error(), http.StatusInternalServerError)
		return
	}
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

func setTracing() {
	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("Fail to create Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-b"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}
