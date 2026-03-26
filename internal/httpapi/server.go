package httpapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"

	"github.com/example/repartners-pack-calculator/internal/packing"
)

//go:embed templates/*.html
var templateFS embed.FS

type Server struct {
	calculator *packing.Calculator
	packSizes  []int
	templates  *template.Template
}

type calculateRequest struct {
	Quantity int `json:"quantity"`
}

type pageData struct {
	PackSizes []int
	Result    *packing.Result
	Error     string
	Quantity  int
}

func NewServer(calculator *packing.Calculator, packSizes []int) (*Server, error) {
	templates, err := template.ParseFS(templateFS, path.Join("templates", "index.html"))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	server := &Server{
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

	result, err := s.calculator.Calculate(request.Quantity)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := pageData{PackSizes: s.packSizes}
		if r.Method == http.MethodPost {
			quantity, err := strconv.Atoi(r.FormValue("quantity"))
			data.Quantity = quantity
			if err != nil {
				data.Error = "Enter a whole number greater than zero."
			} else {
				result, calcErr := s.calculator.Calculate(quantity)
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
