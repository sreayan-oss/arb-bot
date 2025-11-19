package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sYanXO/go-arb-bot/pkg/exchanges"
	"github.com/sYanXO/go-arb-bot/pkg/model"
)

const (
	ArbThreshold = 0.05
)

func main() {
	// 1. Setup the communication channel (Buffered to prevent blocking)
	dataStream := make(chan model.Ticker, 100)

	// 2. Initialize Exchanges
	// We will implement these constructors in the next step.
	// For now, the code compiles but won't produce data until we add these.

	binance := exchanges.NewBinance("BTCUSDT", dataStream)
	go binance.Start()

	coinbase := exchanges.NewCoinbase("BTC-USD", dataStream)
	go coinbase.Start()

	fmt.Println("🚀 Crypto Arbitrage Bot Initialized...")
	fmt.Println("Waiting for price feeds...")

	// 3. The 'State of the World'
	// Stores the latest Ticker for each exchange
	latestTickers := make(map[string]model.Ticker)

	// 4. The Event Loop (Consumer)
	for tick := range dataStream {
		// Update the registry with the new data
		latestTickers[tick.Exchange] = tick

		// Only calculate if we have data from at least 2 exchanges
		if len(latestTickers) >= 2 {
			checkArbitrage(latestTickers)
		}
	}
}

// ... inside cmd/main.go ...

const (
	// Old Reality:
	// BinanceFee  = 0.001
	// CoinbaseFee = 0.006

	// New "VIP Simulation" Mode:
	BinanceFee  = 0.0000
	CoinbaseFee = 0.0000

	UsdtPeg = 1.00
)

func checkArbitrage(tickers map[string]model.Ticker) {
	// 1. Identify Best Ask (Buy) and Best Bid (Sell)
	var bestAsk, bestBid float64
	var buyEx, sellEx string

	// Initialize with worst possible values
	bestAsk = 1e18
	bestBid = 0.0

	for _, t := range tickers {
		// Normalize Price to USD
		price := t.Ask
		if t.Exchange == "Binance" {
			price = price * UsdtPeg // Convert USDT price to USD estimated
		}

		if price < bestAsk {
			bestAsk = price
			buyEx = t.Exchange
		}
	}

	for _, t := range tickers {
		// Normalize Price to USD
		price := t.Bid
		if t.Exchange == "Binance" {
			price = price * UsdtPeg
		}

		if price > bestBid {
			bestBid = price
			sellEx = t.Exchange
		}
	}

	if bestAsk == 0 || bestBid == 0 {
		return
	}

	// 2. Calculate "Real" Cost and Revenue
	// We buy at Ask, paying the fee on top
	buyFeeRate := getFee(buyEx)
	cost := bestAsk * (1 + buyFeeRate)

	// We sell at Bid, losing the fee from the revenue
	sellFeeRate := getFee(sellEx)
	revenue := bestBid * (1 - sellFeeRate)

	// 3. Net Profit Calculation
	netProfit := revenue - cost
	roi := (netProfit / cost) * 100

	// Only print if we actually MAKE money (or if it's very close for debugging)
	if roi > 0 {
		printOpportunity(buyEx, bestAsk, sellEx, bestBid, roi, netProfit)
	} else {
		// Optional: Print negative spreads to see how "efficient" the market is
		// fmt.Printf("\r[Market Efficiency] Loss: %.2f%% (Spread: $%.2f)   ", roi, netProfit)
	}
}

// Helper to get fee based on exchange
func getFee(exchange string) float64 {
	switch exchange {
	case "Binance":
		return BinanceFee
	case "Coinbase":
		return CoinbaseFee
	default:
		return 0.0
	}
}

func printOpportunity(buyEx string, buyPrice float64, sellEx string, sellPrice float64, roi float64, netProfit float64) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	green := "\033[32m"
	reset := "\033[0m"

	fmt.Fprintf(w, "\n%s💰 REAL PROFIT FOUND 💰%s\n", green, reset)
	fmt.Fprintf(w, "------------------------------------------------\n")
	fmt.Fprintf(w, "Strategy:\tBuy %s -> Sell %s\n", buyEx, sellEx)
	fmt.Fprintf(w, "Prices:\t$%.2f -> $%.2f\n", buyPrice, sellPrice)
	fmt.Fprintf(w, "Net Profit:\t$%.2f (ROI: %.3f%%)\n", netProfit, roi)
	fmt.Fprintf(w, "------------------------------------------------\n")
	w.Flush()
}
