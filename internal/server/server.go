package server

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/bjarke-xyz/fuelprices/internal/fuel"
)

// Server serves the JSON API and the embedded single-page app.
type Server struct {
	store   *fuel.Store
	fetcher *fuel.Fetcher
	jobKey  string
	dist    fs.FS
}

func New(store *fuel.Store, fetcher *fuel.Fetcher, jobKey string, dist fs.FS) *Server {
	return &Server{store: store, fetcher: fetcher, jobKey: jobKey, dist: dist}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/job", s.handleJob)
	mux.HandleFunc("GET /api/price-trends", s.handlePriceTrends)
	mux.HandleFunc("GET /api/prices", s.handlePrices)
	mux.Handle("/", spaHandler(s.dist))
	return logRequests(mux)
}

func (s *Server) handleJob(w http.ResponseWriter, r *http.Request) {
	if s.jobKey == "" || r.Header.Get("Authorization") != s.jobKey {
		http.Error(w, "missing key", http.StatusUnauthorized)
		return
	}
	s.fetcher.Fetch(r.Context())
	w.Write([]byte("success"))
}

func (s *Server) handlePriceTrends(w http.ResponseWriter, r *http.Request) {
	today := fuel.Day(time.Now())
	from := fuel.ParseDay(r.URL.Query().Get("from"), fuel.AddDays(today, -1))
	to := fuel.ParseDay(r.URL.Query().Get("to"), today)
	fuelType := fuel.ParseType(r.URL.Query().Get("type"))

	prices, err := s.store.PricesBetween(r.Context(), fuelType, from, to)
	if err != nil {
		slog.Error("get all prices failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorObj("get all prices failed"))
		return
	}

	// The frontend charts these as parallel arrays, and expects [] not null.
	trends := struct {
		Dates  []time.Time `json:"dates"`
		Prices []float64   `json:"prices"`
	}{Dates: []time.Time{}, Prices: []float64{}}
	for _, p := range prices {
		trends.Dates = append(trends.Dates, p.Date)
		trends.Prices = append(trends.Prices, p.Price)
	}
	writeJSON(w, http.StatusOK, trends)
}

func (s *Server) handlePrices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	day := fuel.ParseDay(q.Get("now"), fuel.Day(time.Now()))
	fuelType := fuel.ParseType(firstOf(q.Get("type"), q.Get("fueltype")))
	language := fuel.ParseLanguage(firstOf(q.Get("lang"), q.Get("language")))

	// Both failure paths answer 200: the iOS shortcut reads .message and would
	// otherwise treat the response as a hard failure.
	prices, err := s.store.PricesForDay(r.Context(), fuelType, day)
	if err != nil {
		slog.Error("get prices for day failed", "error", err)
		writeJSON(w, http.StatusOK, errorObj(fuel.ErrorText(language)))
		return
	}
	if prices == nil {
		writeJSON(w, http.StatusOK, errorObj(fuel.ErrorText(language)))
		return
	}

	writeJSON(w, http.StatusOK, struct {
		Message string         `json:"message"`
		Prices  fuel.DayPrices `json:"prices"`
	}{Message: fuel.Text(language, *prices, fuelType), Prices: *prices})
}

// errorObj mirrors the original response, which repeats the message in both
// fields.
func errorObj(msg string) any {
	return struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}{Message: msg, Error: msg}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json failed", "error", err)
	}
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
