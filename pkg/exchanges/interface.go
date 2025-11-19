package exchanges

import "github.com/sYanXO/go-arb-bot/pkg/model"

type Fetcher interface {
	// Returns a read-only channel of Tickers
	Subscribe(symbol string) (<-chan model.Ticker, error)
}
