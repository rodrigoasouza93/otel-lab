package vo_test

import (
	"testing"

	"github.com/rodrigoasouza93/otel-service-b/internal/vo"
)

func TestNewCep(t *testing.T) {
	t.Run("should return error when cep is invalid", func(t *testing.T) {
		_, err := vo.NewCep("123456789")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when cep is empty", func(t *testing.T) {
		_, err := vo.NewCep("")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when cep is not numeric", func(t *testing.T) {
		_, err := vo.NewCep("1234a")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when cep is less than 8 characters", func(t *testing.T) {
		_, err := vo.NewCep("1234")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when cep is more than 8 characters", func(t *testing.T) {
		_, err := vo.NewCep("123456789")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when cep is valid", func(t *testing.T) {
		cep, err := vo.NewCep("12345678")
		if err != nil {
			t.Errorf("expected nil, got %s", err.Error())
		}
		if cep.Value() != "12345678" {
			t.Errorf("expected 12345678, got %s", cep.Value())
		}
	})
}
