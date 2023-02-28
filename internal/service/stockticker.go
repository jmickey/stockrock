package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/exp/slog"
)

const (
	cacheKey   = "SERVICE_RESPONSE"
	baseAPIURL = "https://www.alphavantage.co"
)

type StockTickerService struct {
	apiKey string
	ndays  int
	symbol string
	cache  map[string]*StockTickerResponse
	mutex  *sync.RWMutex
}

type StockServiceOpt = func(*StockTickerService)

// NewStockTickerService returns a *StockTickerService with the provided configuration, along with a basic caching implementation
func NewStockTickerService(apiKey string, ndays int, symbol string) *StockTickerService {
	stockSvc := &StockTickerService{
		apiKey: apiKey,
		ndays:  ndays,
		symbol: symbol,
		cache:  make(map[string]*StockTickerResponse),
		mutex:  &sync.RWMutex{},
	}

	return stockSvc
}

func (svc *StockTickerService) GetStockInfo(ctx context.Context, log *slog.Logger) (*StockTickerResponse, error) {
	logger := log.WithGroup("stock_ticker").With(
		slog.Int("ndays", svc.ndays),
		slog.String("symbol", svc.symbol),
	)

	// Very simple cache implementation. If the cache is populated retrieve the stored response and check the age
	// If the response is less than 10 minutes old then immediately return the stored response
	svc.mutex.RLock()
	cachedResp, ok := svc.cache[cacheKey]
	svc.mutex.RUnlock()
	if ok {
		if age := time.Since(time.Time(cachedResp.LastRefeshed)); age.Seconds() < 600 {
			logger.DebugCtx(ctx, "valid cache response found", "last_refreshed", time.Time(cachedResp.LastRefeshed), "age", age.Seconds())
			return cachedResp, nil
		}
	}

	logger.DebugCtx(ctx, "cached response not found or expired, refeshing data")

	resp, err := svc.retrieveStockInfo(ctx, logger)
	if err != nil {
		return nil, err
	}

	// Sort the time series data base on date
	ts, err := getSortedTimeSeries(resp.TimeSeries, resp.Metadata.Timezone)
	if err != nil {
		return nil, err
	}

	result := &StockTickerResponse{
		LastRefeshed: JSONTime(time.Now().UTC()),
		Days:         svc.ndays,
		Symbol:       svc.symbol,
	}

	// Utilise decimal type to handle floating point arithmetic
	var total decimal.Decimal
	for _, entry := range ts[:svc.ndays] {
		result.StockTimeSeries = append(result.StockTimeSeries, ResponseEntry(entry))

		closePrice, err := decimal.NewFromString(entry.Close)
		if err != nil {
			return nil, fmt.Errorf("failed to parse close price as decimal: %s: %w", entry.Close, err)
		}

		total = total.Add(closePrice)
	}
	result.AverageClosingPrice = total.DivRound(decimal.NewFromInt(int64(svc.ndays)), 2)

	svc.mutex.Lock()
	svc.cache[cacheKey] = result
	svc.mutex.Unlock()

	return result, nil
}

func (svc *StockTickerService) retrieveStockInfo(ctx context.Context, logger *slog.Logger) (*StockInfoResponse, error) {
	client := http.Client{
		Timeout: 15 * time.Minute,
	}

	requestUrl := fmt.Sprintf("%s/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&apikey=%s", baseAPIURL, svc.symbol, svc.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete request to alphavantage API: %w", err)
	}
	defer resp.Body.Close()

	res := &StockInfoResponse{}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.DebugCtx(ctx, "decoded response from api", "api_response", res)

	return res, nil
}

// Maps in Go do not have a deterministic order, therefore we need to convert the data from the API to
// a slice and ensure the slice is sorted in reverse order.
func getSortedTimeSeries(timeseries map[string]Entry, timezone string) (StockInfoTimeSeries, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return StockInfoTimeSeries{}, fmt.Errorf("invalid timezone location: %s: %w", timezone, err)
	}

	ts := make(StockInfoTimeSeries, 0, len(timeseries))

	for dateString, data := range timeseries {
		date, err := time.ParseInLocation("2006-01-02", dateString, loc)
		if err != nil {
			return StockInfoTimeSeries{}, fmt.Errorf("error parsing date string: %s: %w", dateString, err)
		}

		data.Date = date

		ts = append(ts, data)
	}

	sort.Sort(sort.Reverse(ts))
	return ts, nil
}
