package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	appconfig "github.com/example/repartners-pack-calculator/internal/config"
	"github.com/example/repartners-pack-calculator/internal/packing"
)

func TestHandleAPICalculate(t *testing.T) {
	server := newTestServer(t, []int{250, 500, 1000, 2000, 5000})

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
	server := newTestServer(t, []int{250, 500})

	request := httptest.NewRequest(http.MethodPost, "/api/calculate", bytes.NewReader([]byte("{")))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestHandleGetPackConfig(t *testing.T) {
	server := newTestServer(t, []int{500, 250, 1000})

	request := httptest.NewRequest(http.MethodGet, "/api/config/packs", nil)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var cfg appconfig.AppConfig
	if err := json.NewDecoder(response.Body).Decode(&cfg); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	expected := []int{500, 250, 1000}
	if !reflect.DeepEqual(cfg.PackSizes, expected) {
		t.Fatalf("PackSizes = %#v, want %#v", cfg.PackSizes, expected)
	}
}

func TestHandlePutPackConfigUpdatesServerAndFile(t *testing.T) {
	server := newTestServer(t, []int{250, 500, 1000})

	body, err := json.Marshal(appconfig.AppConfig{PackSizes: []int{53, 23, 31}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPut, "/api/config/packs", bytes.NewReader(body))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var cfg appconfig.AppConfig
	if err := json.NewDecoder(response.Body).Decode(&cfg); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	expected := []int{23, 31, 53}
	if !reflect.DeepEqual(cfg.PackSizes, expected) {
		t.Fatalf("PackSizes = %#v, want %#v", cfg.PackSizes, expected)
	}

	persisted, err := appconfig.Load(server.configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(persisted.PackSizes, expected) {
		t.Fatalf("persisted PackSizes = %#v, want %#v", persisted.PackSizes, expected)
	}

	calculateBody, err := json.Marshal(map[string]int{"quantity": 54})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	calculateRequest := httptest.NewRequest(http.MethodPost, "/api/calculate", bytes.NewReader(calculateBody))
	calculateResponse := httptest.NewRecorder()
	server.Routes().ServeHTTP(calculateResponse, calculateRequest)

	if calculateResponse.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", calculateResponse.Code, http.StatusOK)
	}

	var result packing.Result
	if err := json.NewDecoder(calculateResponse.Body).Decode(&result); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if result.TotalItems != 54 {
		t.Fatalf("TotalItems = %d, want 54", result.TotalItems)
	}
}

func TestHandlePutPackConfigRejectsInvalidJSON(t *testing.T) {
	server := newTestServer(t, []int{250, 500})

	request := httptest.NewRequest(http.MethodPut, "/api/config/packs", bytes.NewReader([]byte("{")))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestHandlePutPackConfigRejectsInvalidPackSizes(t *testing.T) {
	server := newTestServer(t, []int{250, 500})

	body, err := json.Marshal(appconfig.AppConfig{PackSizes: []int{500, 500}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPut, "/api/config/packs", bytes.NewReader(body))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	persisted, err := appconfig.Load(server.configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expected := []int{250, 500}
	if !reflect.DeepEqual(persisted.PackSizes, expected) {
		t.Fatalf("persisted PackSizes = %#v, want %#v", persisted.PackSizes, expected)
	}
}

func TestHandleHealthz(t *testing.T) {
	server := newTestServer(t, []int{250, 500})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	if strings.TrimSpace(response.Body.String()) != "ok" {
		t.Fatalf("body = %q, want ok", response.Body.String())
	}
}

func TestHandleIndexUsesCurrentPackSizes(t *testing.T) {
	server := newTestServer(t, []int{250, 500})

	body, err := json.Marshal(appconfig.AppConfig{PackSizes: []int{100, 400}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/config/packs", bytes.NewReader(body))
	updateResponse := httptest.NewRecorder()
	server.Routes().ServeHTTP(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", updateResponse.Code, http.StatusOK)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	bodyText := response.Body.String()
	if !strings.Contains(bodyText, "100") || !strings.Contains(bodyText, "400") {
		t.Fatalf("body = %q, want updated pack sizes", bodyText)
	}
}

func newTestServer(t *testing.T, packSizes []int) *Server {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "packs.json")
	if err := appconfig.Save(configPath, appconfig.AppConfig{PackSizes: packSizes}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	calculator, err := packing.NewCalculator(packSizes)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	server, err := NewServer(calculator, configPath, packSizes)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	return server
}
