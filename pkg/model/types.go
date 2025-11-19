package model

import "time"

type Ticker struct {
	Exchange  string
	Symbol    string
	Bid       float64
	Ask       float64
	Timestamp time.Time
}
