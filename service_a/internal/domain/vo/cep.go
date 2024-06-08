package vo

import "errors"

type Cep struct {
	Value string
}

var ErrInvalidCep = errors.New("invalid zipcode")

func NewCep(value string) (Cep, error) {
	if !validCep(value) {
		return Cep{}, ErrInvalidCep
	}
	return Cep{
		Value: value,
	}, nil
}

func validCep(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, c := range cep {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
