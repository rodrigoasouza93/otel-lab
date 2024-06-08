package dto

type WeatherOutput struct {
	Temp_C float64 `json:"temp_c"`
	Temp_F float64 `json:"temp_F"`
	Temp_K float64 `json:"temp_K"`
	City   string  `json:"city"`
}
