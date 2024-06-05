package vo

import (
	"errors"
	"regexp"
	"strings"
)

type Cep struct {
	value string
}

var ErrInvalidCep = errors.New("invalid cep")

func NewCep(value string) (*Cep, error) {
	if !IsValid(value) {
		return nil, ErrInvalidCep
	}
	return &Cep{value: value}, nil
}

func (c *Cep) Value() string {
	return c.value
}

func IsValid(raw string) bool {
	cep := strings.Replace(raw, "-", "", -1)
	if len(cep) != 8 {
		return false
	}
	onlyDigits := regexp.MustCompile(`^\d+$`)
	if !onlyDigits.MatchString(cep) {
		return false
	}
	return true
}
