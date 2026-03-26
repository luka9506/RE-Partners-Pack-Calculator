package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/repartners-pack-calculator/internal/packing"
)

func TestHandleAPICalculate(t *testing.T) {
	calculator, err := packing.NewCalculator([]int{250, 500, 1000, 2000, 5000})
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	server, err := NewServer(calculator, []int{250, 500, 1000, 2000, 5000})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	body, err := json.Marshal(map[string]int{"quantity": 501})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/calculate", bytes.NewReader(body))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var result packing.Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if result.TotalItems != 750 {
		t.Fatalf("TotalItems = %d, want 750", result.TotalItems)
	}
}

func TestHandleAPICalculateRejectsInvalidJSON(t *testing.T) {
	calculator, err := packing.NewCalculator([]int{250, 500})
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	server, err := NewServer(calculator, []int{250, 500})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/calculate", bytes.NewReader([]byte("{")))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
