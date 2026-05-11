package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"currency-converter/api"
	"currency-converter/internal/currency"
)

type Server struct {
	converter *currency.Converter
}

func NewServer() *Server {
	converter := currency.NewConverter()

	initialRates := []currency.Rate{
		{From: currency.USD, To: currency.CNY, Rate: 7.25},
		{From: currency.USD, To: currency.EUR, Rate: 0.92},
		{From: currency.EUR, To: currency.GBP, Rate: 0.85},
	}
	converter.SetRates(initialRates)

	return &Server{converter: converter}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	s.writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.converter.Convert(req.Amount, currency.Currency(req.From), currency.Currency(req.To))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp := api.ConvertResponse{
		Amount:   req.Amount,
		From:     req.From,
		To:       req.To,
		Result:   result,
		Original: req.Amount,
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCurrencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currencies := s.converter.GetCurrencies()
	apiCurrencies := make([]api.Currency, len(currencies))
	for i, c := range currencies {
		apiCurrencies[i] = api.Currency(c)
	}

	s.writeJSON(w, http.StatusOK, api.CurrenciesResponse{Currencies: apiCurrencies})
}

func (s *Server) handleRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rates := s.converter.GetRates()
	apiRates := make([]api.Rate, len(rates))
	for i, r := range rates {
		apiRates[i] = api.Rate{
			From: api.Currency(r.From),
			To:   api.Currency(r.To),
			Rate: r.Rate,
		}
	}

	s.writeJSON(w, http.StatusOK, api.RatesResponse{Rates: apiRates})
}

func (s *Server) handleSetRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SetRatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	currencyRates := make([]currency.Rate, len(req.Rates))
	for i, r := range req.Rates {
		currencyRates[i] = currency.Rate{
			From: currency.Currency(r.From),
			To:   currency.Currency(r.To),
			Rate: r.Rate,
		}
	}

	s.converter.AddRates(currencyRates)
	s.writeJSON(w, http.StatusOK, api.SetRatesResponse{Success: true})
}

func getPort() string {
	var port int
	flag.IntVar(&port, "port", 8100, "Server port")
	flag.Parse()

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			return fmt.Sprintf(":%d", p)
		}
	}

	return fmt.Sprintf(":%d", port)
}

func main() {
	server := NewServer()
	port := getPort()

	mux := http.NewServeMux()
	mux.HandleFunc("/convert", server.handleConvert)
	mux.HandleFunc("/currencies", server.handleCurrencies)
	mux.HandleFunc("/rates", server.handleRates)
	mux.HandleFunc("/rates/set", server.handleSetRates)

	fmt.Printf("Server starting on %s...\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
