package exchanges

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sYanXO/go-arb-bot/pkg/model"
)

// binanceResponse mirrors the exact JSON structure from Binance.
// Notice we keep fields unexported (lowercase) if possible, but JSON decoder needs them exported.
type binanceResponse struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"` // Binance sends strings!
	AskPrice string `json:"askPrice"`
}

// Binance is our exchange handler
type Binance struct {
	symbol string
	out    chan<- model.Ticker // Write-only channel (safety)
	client *http.Client
}

// NewBinance is the constructor
func NewBinance(symbol string, out chan<- model.Ticker) *Binance {
	return &Binance{
		symbol: symbol,
		out:    out,
		// Always set a timeout on HTTP clients to prevent hanging goroutines
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Start begins the polling loop
func (b *Binance) Start() {
	url := "https://api.binance.com/api/v3/ticker/bookTicker?symbol=" + b.symbol

	for {
		// 1. Fetch Data
		resp, err := b.client.Get(url)
		if err != nil {
			log.Printf("[Binance] Connection error: %v", err)
			time.Sleep(2 * time.Second) // Backoff before retrying
			continue
		}

		// 2. Decode JSON
		var apiData binanceResponse
		// Decode directly from the stream (efficient) rather than reading all bytes first
		if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
			log.Printf("[Binance] Decode error: %v", err)
			resp.Body.Close() // Don't forget to close body!
			continue
		}
		resp.Body.Close()

		// 3. Parse Strings to Floats
		bid, _ := strconv.ParseFloat(apiData.BidPrice, 64)
		ask, _ := strconv.ParseFloat(apiData.AskPrice, 64)

		// 4. Send to Main Engine
		// This is non-blocking if the channel buffer isn't full
		b.out <- model.Ticker{
			Exchange:  "Binance",
			Symbol:    apiData.Symbol,
			Bid:       bid,
			Ask:       ask,
			Timestamp: time.Now(),
		}

		// 5. Rate Limiting
		// Binance allows plenty of requests, but let's be polite
		time.Sleep(1 * time.Second)
	}
}
