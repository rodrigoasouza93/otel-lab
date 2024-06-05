package dto

type LocationInfo struct {
	Name string `json:"name"`
}

type CurrentInfo struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
}

type WeatherResponse struct {
	Location LocationInfo `json:"location"`
	Current  CurrentInfo  `json:"current"`
}
