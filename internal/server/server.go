package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmickey/stockrock/internal/service"
	"golang.org/x/exp/slog"
)

const (
	BASE_URL = "https://www.alphavantage.co"
)

type StockServer struct {
	logger   *slog.Logger
	host     string
	port     int
	mux      *http.ServeMux
	stockSvc *service.StockTickerService
}

func NewStockServer(svc *service.StockTickerService, opts ...ServerOpt) (*StockServer, error) {
	// Configure the server with sensible defaults,
	// but allow them to be overidden.
	svr := &StockServer{
		logger:   slog.Default(),
		host:     "localhost",
		port:     8080,
		mux:      http.NewServeMux(),
		stockSvc: svc,
	}

	for _, opt := range opts {
		opt(svr)
	}

	svr.configureHandlers()

	return svr, nil
}

func (s *StockServer) Run() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.logger.Debug("starting server")
	return http.ListenAndServe(addr, s.mux)
}

func (s *StockServer) configureHandlers() {
	s.mux.HandleFunc("/api/stock-info", s.getStockInfo)
	s.mux.HandleFunc("/healthz", s.getHealthStatus)
}

func (s *StockServer) getStockInfo(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With(
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("headers", r.Header),
		),
	)

	if r.Method != "GET" {
		logger.InfoCtx(r.Context(), "invalid http method",
			slog.Group("response",
				slog.Int("status_code", http.StatusMethodNotAllowed),
				slog.String("msg", http.StatusText(http.StatusMethodNotAllowed)),
			),
		)

		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	result, err := s.stockSvc.GetStockInfo(r.Context(), logger)
	if err != nil {
		logger.ErrorCtx(r.Context(), "error from stock service", err,
			slog.Group("response",
				slog.Int("status_code", http.StatusInternalServerError),
				slog.String("msg", http.StatusText(http.StatusInternalServerError)),
			),
		)

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.InfoCtx(r.Context(), "request processed",
		slog.Group("response",
			slog.Int("status_code", http.StatusOK),
			slog.Any("msg", result),
		),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *StockServer) getHealthStatus(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With(
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("headers", r.Header),
		),
	)

	if r.Method != "GET" {
		logger.InfoCtx(r.Context(), "invalid http method",
			slog.Group("response",
				slog.Int("status_code", http.StatusMethodNotAllowed),
				slog.String("msg", http.StatusText(http.StatusMethodNotAllowed)),
			),
		)

		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	logger.InfoCtx(r.Context(), "request processed",
		slog.Group("response",
			slog.Int("status_code", http.StatusOK),
			slog.Any("msg", map[string]string{"status": "ok"}),
		),
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
