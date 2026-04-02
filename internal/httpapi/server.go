package httpapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/example/repartners-pack-calculator/internal/config"
	"github.com/example/repartners-pack-calculator/internal/packing"
)

//go:embed templates/*.html
var templateFS embed.FS

type Server struct {
	configPath string
	mu         sync.RWMutex
	calculator *packing.Calculator
	packSizes  []int
	templates  *template.Template
}

type calculateRequest struct {
	Quantity int `json:"quantity"`
}

type packConfigRequest struct {
	PackSizes []int `json:"pack_sizes"`
}

type pageData struct {
	PackSizes []int
	Result    *packing.Result
	Error     string
	Quantity  int
}

func NewServer(calculator *packing.Calculator, configPath string, packSizes []int) (*Server, error) {
	templates, err := template.ParseFS(templateFS, path.Join("templates", "index.html"))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	server := &Server{
		configPath: configPath,
		calculator: calculator,
		packSizes:  append([]int(nil), packSizes...),
		templates:  templates,
	}

	return server, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex())
	mux.HandleFunc("POST /", s.handleIndex())
	mux.HandleFunc("POST /api/calculate", s.handleAPICalculate)
	mux.HandleFunc("GET /api/config/packs", s.handleGetPackConfig)
	mux.HandleFunc("PUT /api/config/packs", s.handlePutPackConfig)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (s *Server) handleAPICalculate(w http.ResponseWriter, r *http.Request) {
	var request calculateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	calculator, _ := s.snapshot()
	result, err := calculator.Calculate(request.Quantity)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetPackConfig(w http.ResponseWriter, _ *http.Request) {
	_, packSizes := s.snapshot()
	writeJSON(w, http.StatusOK, config.AppConfig{PackSizes: packSizes})
}

func (s *Server) handlePutPackConfig(w http.ResponseWriter, r *http.Request) {
	var request packConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	nextConfig := config.AppConfig{PackSizes: append([]int(nil), request.PackSizes...)}
	if err := nextConfig.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	nextCalculator, err := packing.NewCalculator(nextConfig.PackSizes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := config.Save(s.configPath, nextConfig); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist pack configuration"})
		return
	}

	s.mu.Lock()
	s.calculator = nextCalculator
	s.packSizes = append([]int(nil), nextConfig.PackSizes...)
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, nextConfig)
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		calculator, packSizes := s.snapshot()
		data := pageData{PackSizes: packSizes}
		if r.Method == http.MethodPost {
			// The HTML form reuses the same calculator as the JSON API so both entry
			// points always follow identical packing rules.
			quantity, err := strconv.Atoi(r.FormValue("quantity"))
			data.Quantity = quantity
			if err != nil {
				data.Error = "Enter a whole number greater than zero."
			} else {
				result, calcErr := calculator.Calculate(quantity)
				if calcErr != nil {
					data.Error = calcErr.Error()
				} else {
					data.Result = &result
				}
			}
		}

		if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, "failed to render page", http.StatusInternalServerError)
		}
	}
}

func (s *Server) snapshot() (*packing.Calculator, []int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.calculator, append([]int(nil), s.packSizes...)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
