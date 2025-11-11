package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// This line is removed in the next change

// BinanceExporter collects cryptocurrency metrics from Binance REST API
type BinanceExporter struct {
	// Prometheus metrics
	priceGauge           *prometheus.GaugeVec
	volumeGauge          *prometheus.GaugeVec
	tradesCounter        *prometheus.CounterVec
	bidPriceGauge        *prometheus.GaugeVec
	askPriceGauge        *prometheus.GaugeVec
	spreadGauge          *prometheus.GaugeVec
	priceChangeGauge     *prometheus.GaugeVec
	priceChangePercGauge *prometheus.GaugeVec
	highPriceGauge       *prometheus.GaugeVec
	lowPriceGauge        *prometheus.GaugeVec
	quoteVolumeGauge     *prometheus.GaugeVec
	openPriceGauge       *prometheus.GaugeVec
	lastUpdateGauge      *prometheus.GaugeVec
	weightedAvgGauge     *prometheus.GaugeVec
	upGauge              *prometheus.GaugeVec

	symbols    []string
	httpClient *http.Client
	mu         sync.RWMutex
	data       map[string]*TickerData
	lastTrades map[string]int64
}

// TickerData represents the 24hr ticker statistics from Binance REST API
type TickerData struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FirstID            int64  `json:"firstId"`
	LastID             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}

// NewBinanceExporter creates a new Binance metrics exporter
func NewBinanceExporter(symbols []string) *BinanceExporter {
	return &BinanceExporter{
		symbols:    symbols,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		data:       make(map[string]*TickerData),
		lastTrades: make(map[string]int64),

		priceGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_price",
				Help: "Current price of cryptocurrency pair",
			},
			[]string{"symbol"},
		),
		volumeGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_volume",
				Help: "24h trading volume",
			},
			[]string{"symbol"},
		),
		tradesCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "binance_trades_total",
				Help: "Total number of trades in 24h",
			},
			[]string{"symbol"},
		),
		bidPriceGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_bid_price",
				Help: "Current best bid price",
			},
			[]string{"symbol"},
		),
		askPriceGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_ask_price",
				Help: "Current best ask price",
			},
			[]string{"symbol"},
		),
		spreadGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_spread",
				Help: "Bid-ask spread (ask - bid)",
			},
			[]string{"symbol"},
		),
		priceChangeGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_price_change_24h",
				Help: "24h price change in absolute value",
			},
			[]string{"symbol"},
		),
		priceChangePercGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_price_change_percent_24h",
				Help: "24h price change in percentage",
			},
			[]string{"symbol"},
		),
		highPriceGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_high_price_24h",
				Help: "24h highest price",
			},
			[]string{"symbol"},
		),
		lowPriceGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_low_price_24h",
				Help: "24h lowest price",
			},
			[]string{"symbol"},
		),
		quoteVolumeGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_quote_volume_24h",
				Help: "24h quote asset volume (e.g., USDT volume)",
			},
			[]string{"symbol"},
		),
		openPriceGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_open_price_24h",
				Help: "24h opening price",
			},
			[]string{"symbol"},
		),
		lastUpdateGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_last_update_timestamp",
				Help: "Timestamp of last update",
			},
			[]string{"symbol"},
		),
		weightedAvgGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_weighted_avg_price_24h",
				Help: "24h weighted average price",
			},
			[]string{"symbol"},
		),
		upGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "binance_up",
				Help: "1 if the exporter is successfully scraping data, 0 otherwise",
			},
			[]string{"symbol"},
		),
	}
}

// Start begins collecting metrics from Binance REST API every 20 seconds
func (e *BinanceExporter) Start(ctx context.Context) error {
	slog.Info("Starting Binance REST API exporter", slog.Any("symbols", e.symbols))

	// Fetch data immediately on start
	e.fetchData(ctx)
	e.updateMetrics()

	// Fetch and update metrics every 20 seconds
	ticker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Stopping Binance exporter")
				return
			case <-ticker.C:
				e.fetchData(ctx)
				e.updateMetrics()
			}
		}
	}()

	return nil
}

// fetchData fetches ticker data from Binance REST API for all symbols
func (e *BinanceExporter) fetchData(ctx context.Context) {
	var wg sync.WaitGroup
	// limit concurrency to avoid overwhelming the API
	sem := make(chan struct{}, 5)
	for _, symbol := range e.symbols {
		wg.Add(1)
		sym := symbol
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				e.fetchSymbolData(ctx, sym)
			case <-ctx.Done():
				return
			}
		}()
	}
	wg.Wait()
}

// fetchSymbolData fetches data for a single symbol from Binance REST API
func (e *BinanceExporter) fetchSymbolData(ctx context.Context, symbol string) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%s", symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("Failed to create request",
			slog.String("symbol", symbol),
			slog.String("error", err.Error()))
		e.upGauge.WithLabelValues(symbol).Set(0)
		return
	}
	req.Header.Set("User-Agent", "dv-binance-exporter/1.0")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		slog.Error("Failed to fetch data from Binance API",
			slog.String("symbol", symbol),
			slog.String("error", err.Error()))
		e.upGauge.WithLabelValues(symbol).Set(0)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Binance API returned error",
			slog.String("symbol", symbol),
			slog.Int("status", resp.StatusCode))
		e.upGauge.WithLabelValues(symbol).Set(0)
		return
	}

	var ticker TickerData
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		slog.Error("Failed to parse JSON response",
			slog.String("symbol", symbol),
			slog.String("error", err.Error()))
		e.upGauge.WithLabelValues(symbol).Set(0)
		return
	}

	e.mu.Lock()
	e.data[symbol] = &ticker
	e.mu.Unlock()

	e.upGauge.WithLabelValues(symbol).Set(1)

	slog.Debug("Fetched data from Binance API",
		slog.String("symbol", symbol),
		slog.String("price", ticker.LastPrice))
}

func (e *BinanceExporter) updateMetrics() {
	// Take a consistent snapshot of current data and last trade counts
	e.mu.RLock()
	dataCopy := make(map[string]TickerData, len(e.data))
	lastTradesCopy := make(map[string]int64, len(e.lastTrades))
	for s, d := range e.data {
		if d != nil {
			dataCopy[s] = *d
		}
	}
	for s, c := range e.lastTrades {
		lastTradesCopy[s] = c
	}
	e.mu.RUnlock()

	// Compute metrics and the new last trade counts without holding the lock
	newCounts := make(map[string]int64, len(dataCopy))
	for symbol, data := range dataCopy {
		var (
			price, _           = parseFloat(data.LastPrice)
			volume, _          = parseFloat(data.Volume)
			bidPrice, _        = parseFloat(data.BidPrice)
			askPrice, _        = parseFloat(data.AskPrice)
			priceChange, _     = parseFloat(data.PriceChange)
			priceChangePerc, _ = parseFloat(data.PriceChangePercent)
			highPrice, _       = parseFloat(data.HighPrice)
			lowPrice, _        = parseFloat(data.LowPrice)
			quoteVolume, _     = parseFloat(data.QuoteVolume)
			openPrice, _       = parseFloat(data.OpenPrice)
			weightedAvg, _     = parseFloat(data.WeightedAvgPrice)
		)

		e.priceGauge.WithLabelValues(symbol).Set(price)
		e.volumeGauge.WithLabelValues(symbol).Set(volume)

		// Handle counter for trades - only increment by difference using the snapshot
		lastCount := lastTradesCopy[symbol]
		if data.Count > lastCount {
			e.tradesCounter.WithLabelValues(symbol).Add(float64(data.Count - lastCount))
		}
		newCounts[symbol] = data.Count

		e.bidPriceGauge.WithLabelValues(symbol).Set(bidPrice)
		e.askPriceGauge.WithLabelValues(symbol).Set(askPrice)
		e.spreadGauge.WithLabelValues(symbol).Set(askPrice - bidPrice)
		e.priceChangeGauge.WithLabelValues(symbol).Set(priceChange)
		e.priceChangePercGauge.WithLabelValues(symbol).Set(priceChangePerc)
		e.highPriceGauge.WithLabelValues(symbol).Set(highPrice)
		e.lowPriceGauge.WithLabelValues(symbol).Set(lowPrice)
		e.quoteVolumeGauge.WithLabelValues(symbol).Set(quoteVolume)
		e.openPriceGauge.WithLabelValues(symbol).Set(openPrice)
		e.lastUpdateGauge.WithLabelValues(symbol).Set(float64(data.CloseTime))
		e.weightedAvgGauge.WithLabelValues(symbol).Set(weightedAvg)

		slog.Debug("Updated metrics",
			slog.String("symbol", symbol),
			slog.Float64("price", price),
			slog.Float64("volume", volume),
		)
	}

	// Persist the updated last trade counts with a write lock
	e.mu.Lock()
	for s, c := range newCounts {
		e.lastTrades[s] = c
	}
	e.mu.Unlock()
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
