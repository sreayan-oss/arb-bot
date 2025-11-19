package exchanges

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sYanXO/go-arb-bot/pkg/model"
)

// coinbaseResponse matches the https://api.exchange.coinbase.com/products/BTC-USD/ticker structure
type coinbaseResponse struct {
	Price string `json:"price"`
	Bid   string `json:"bid"`
	Ask   string `json:"ask"`
	Time  string `json:"time"`
}

type Coinbase struct {
	symbol string
	out    chan<- model.Ticker
	client *http.Client
}

func NewCoinbase(symbol string, out chan<- model.Ticker) *Coinbase {
	return &Coinbase{
		symbol: symbol,
		out:    out,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Coinbase) Start() {
	// Note: Coinbase uses dashes in symbols, e.g., "BTC-USD"
	url := "https://api.exchange.coinbase.com/products/" + c.symbol + "/ticker"

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("[Coinbase] Request creation error: %v", err)
			continue
		}

		// Coinbase sometimes requires a User-Agent to not block scripts
		req.Header.Set("User-Agent", "Go-Arb-Bot/1.0")

		resp, err := c.client.Do(req)
		if err != nil {
			log.Printf("[Coinbase] Connection error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			log.Printf("[Coinbase] API Error: Status %d", resp.StatusCode)
			resp.Body.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		var apiData coinbaseResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
			log.Printf("[Coinbase] Decode error: %v", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Parse strings to floats
		bid, _ := strconv.ParseFloat(apiData.Bid, 64)
		ask, _ := strconv.ParseFloat(apiData.Ask, 64)

		c.out <- model.Ticker{
			Exchange:  "Coinbase",
			Symbol:    c.symbol,
			Bid:       bid,
			Ask:       ask,
			Timestamp: time.Now(),
		}

		// Coinbase rate limits are stricter; sleep slightly longer
		time.Sleep(1500 * time.Millisecond)
	}
}
