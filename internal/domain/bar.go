package domain

import (
	"fmt"
)

type Symbol string

const (
	SymbolBTCUSDT Symbol = "BTCUSDT"
	SymbolETHUSDT Symbol = "ETHUSDT"
)

type Interval string

const (
	Interval1m Interval = "1m"
)

var allowedSymbols = map[Symbol]struct{}{
	SymbolBTCUSDT: {},
	SymbolETHUSDT: {},
}

type Bar struct {
	Symbol    Symbol   `json:"symbol"`
	Interval  Interval `json:"interval"`
	OpenTime  int64    `json:"open_time"`
	CloseTime int64    `json:"close_time"`
	Open      string   `json:"open"`
	High      string   `json:"high"`
	Low       string   `json:"low"`
	Close     string   `json:"close"`
	Volume    string   `json:"volume"`
	EventTime int64    `json:"event_time"`
	IsClosed  bool     `json:"is_closed"`
}

func NormalizeSymbol(raw string) (Symbol, error) {
	symbol := Symbol(raw)
	if _, ok := allowedSymbols[symbol]; !ok {
		return "", fmt.Errorf("unsupported symbol %q", raw)
	}
	return symbol, nil
}

func (bar Bar) Validate() error {
	if _, ok := allowedSymbols[bar.Symbol]; !ok {
		return fmt.Errorf("unsupported symbol %q", bar.Symbol)
	}
	if bar.Interval != Interval1m {
		return fmt.Errorf("unsupported interval %q", bar.Interval)
	}
	if bar.OpenTime <= 0 {
		return fmt.Errorf("open time must be positive")
	}
	if bar.CloseTime <= 0 {
		return fmt.Errorf("close time must be positive")
	}
	if bar.CloseTime < bar.OpenTime {
		return fmt.Errorf("close time must not be before open time")
	}
	if bar.EventTime <= 0 {
		return fmt.Errorf("event time must be positive")
	}
	if bar.Open == "" || bar.High == "" || bar.Low == "" || bar.Close == "" || bar.Volume == "" {
		return fmt.Errorf("bar prices and volume must be set")
	}
	return nil
}
