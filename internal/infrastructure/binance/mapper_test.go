package binanceinfra

import (
	"testing"

	binance "github.com/adshao/go-binance/v2"
	"github.com/ranjbar-dev/otc-price/internal/domain"
)

func TestMapWsKlineEvent(t *testing.T) {
	t.Parallel()

	event := &binance.WsKlineEvent{
		Time:   30,
		Symbol: "BTCUSDT",
		Kline: binance.WsKline{
			StartTime: 1,
			EndTime:   2,
			Symbol:    "BTCUSDT",
			Interval:  "1m",
			Open:      "1.0",
			Close:     "1.5",
			High:      "1.7",
			Low:       "0.9",
			Volume:    "100",
			IsFinal:   true,
		},
	}

	bar, err := MapWsKlineEvent(event)
	if err != nil {
		t.Fatalf("map event: %v", err)
	}

	if bar.Symbol != domain.SymbolBTCUSDT {
		t.Fatalf("unexpected symbol: %s", bar.Symbol)
	}
	if bar.Interval != domain.Interval1m {
		t.Fatalf("unexpected interval: %s", bar.Interval)
	}
	if !bar.IsClosed {
		t.Fatal("expected closed bar")
	}
}

func TestMapWsKlineEventRejectsUnexpectedInterval(t *testing.T) {
	t.Parallel()

	_, err := MapWsKlineEvent(&binance.WsKlineEvent{
		Time:   30,
		Symbol: "ETHUSDT",
		Kline: binance.WsKline{
			StartTime: 1,
			EndTime:   2,
			Interval:  "5m",
			Open:      "1.0",
			Close:     "1.5",
			High:      "1.7",
			Low:       "0.9",
			Volume:    "100",
		},
	})
	if err == nil {
		t.Fatal("expected error for unsupported interval")
	}
}
