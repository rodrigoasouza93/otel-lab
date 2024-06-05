package dto

type LocationResponse struct {
	Cep    string `json:"cep"`
	Locale string `json:"localidade"`
	Error  bool   `json:"erro"`
}
