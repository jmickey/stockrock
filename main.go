package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jmickey/stockrock/internal/server"
	"github.com/jmickey/stockrock/internal/service"
	"golang.org/x/exp/slog"
)

const (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = 8080
)

type configFromEnv struct {
	host   string
	port   int
	apiKey string
	ndays  int
	symbol string
}

func main() {
	logger := newLogger()

	cfg, err := parseEnv()
	if err != nil {
		logger.Error("error parsing environment variables", err)
		os.Exit(1)
	}

	stockSvc := service.NewStockTickerService(
		cfg.apiKey,
		cfg.ndays,
		cfg.symbol,
	)

	svr, err := server.NewStockServer(
		stockSvc,
		server.WithLogger(logger),
		server.WithHost(cfg.host),
		server.WithPort(cfg.port),
	)
	if err != nil {
		logger.Error("couldn't initialise server", err)
		os.Exit(1)
	}

	if err := svr.Run(); err != nil {
		logger.Error("server closed with error", err)
	}
}

func newLogger() *slog.Logger {
	if env, ok := os.LookupEnv("ENV"); ok {
		if env == "dev" {
			return slog.New(slog.HandlerOptions{Level: slog.LevelDebug}.NewTextHandler(os.Stderr))
		}
	}

	return slog.New(slog.HandlerOptions{Level: slog.LevelInfo}.NewJSONHandler(os.Stderr))
}

func parseEnv() (*configFromEnv, error) {
	var err error

	cfg := &configFromEnv{
		host: DEFAULT_HOST,
		port: DEFAULT_PORT,
	}

	if host, ok := os.LookupEnv("HOST"); ok {
		cfg.host = host
	}

	if port, ok := os.LookupEnv("PORT"); ok {
		cfg.port, err = strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("error converting PORT to an integer: %w", err)
		}
	}

	if apiKey, ok := os.LookupEnv("API_KEY"); ok {
		cfg.apiKey = apiKey
	} else {
		return nil, fmt.Errorf("API_KEY environment variable is required and not found")
	}

	if ndays, ok := os.LookupEnv("NDAYS"); ok {
		cfg.ndays, err = strconv.Atoi(ndays)
		if err != nil || cfg.ndays < 1 {
			return nil, fmt.Errorf("NDAYS not an integer or is less than 1: %w", err)
		}
	} else {
		return nil, fmt.Errorf("NDAYS environment variable is required and not found")
	}

	if symbol, ok := os.LookupEnv("SYMBOL"); ok {
		cfg.symbol = symbol
	} else {
		return nil, fmt.Errorf("SYMBOL environment variable is required and not found")
	}

	return cfg, nil
}
