package service

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// StockInfoResponse is defines the API response from the alphavantage API
type StockInfoResponse struct {
	Metadata   Metadata         `json:"Meta Data"`
	TimeSeries map[string]Entry `json:"Time Series (Daily)"`
}

type Metadata struct {
	Information  string `json:"1. Information"`
	Symbol       string `json:"2. Symbol"`
	LastRefeshed string `json:"3. Last Refreshed"`
	OutputSize   string `json:"4. Output Size"`
	Timezone     string `json:"5. Time Zone"`
}

type Entry struct {
	Date   time.Time
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"6. volume"`
}

type StockInfoTimeSeries []Entry

// Implementation methods to satisfy sort.Sort interface
func (ts StockInfoTimeSeries) Len() int {
	return len(ts)
}

func (ts StockInfoTimeSeries) Less(i, j int) bool {
	return ts[i].Date.Before(ts[j].Date)
}

func (ts StockInfoTimeSeries) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

// StockTickerResponse defines the API response from the StockTickerService
type StockTickerResponse struct {
	LastRefeshed        JSONTime        `json:"last_refeshed"`
	Days                int             `json:"days"`
	Symbol              string          `json:"symbol"`
	AverageClosingPrice decimal.Decimal `json:"average_closing_price"`
	StockTimeSeries     []ResponseEntry `json:"stock_time_series"`
}

type ResponseEntry struct {
	Date   time.Time `json:"date"`
	Open   string    `json:"open"`
	High   string    `json:"high"`
	Low    string    `json:"low"`
	Close  string    `json:"close"`
	Volume string    `json:"volume"`
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.ANSIC))), nil
}
