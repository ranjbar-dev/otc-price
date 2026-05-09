package binanceinfra

import (
	"fmt"
	"strings"

	binance "github.com/adshao/go-binance/v2"
	"github.com/ranjbar-dev/otc-price/internal/domain"
)

func MapWsKlineEvent(event *binance.WsKlineEvent) (domain.Bar, error) {
	if event == nil {
		return domain.Bar{}, fmt.Errorf("event is nil")
	}

	symbol, err := domain.NormalizeSymbol(strings.ToUpper(event.Symbol))
	if err != nil {
		return domain.Bar{}, err
	}

	if event.Kline.Interval != string(domain.Interval1m) {
		return domain.Bar{}, fmt.Errorf("unsupported interval %q", event.Kline.Interval)
	}

	bar := domain.Bar{
		Symbol:    symbol,
		Interval:  domain.Interval(event.Kline.Interval),
		OpenTime:  event.Kline.StartTime,
		CloseTime: event.Kline.EndTime,
		Open:      event.Kline.Open,
		High:      event.Kline.High,
		Low:       event.Kline.Low,
		Close:     event.Kline.Close,
		Volume:    event.Kline.Volume,
		EventTime: event.Time,
		IsClosed:  event.Kline.IsFinal,
	}

	if err := bar.Validate(); err != nil {
		return domain.Bar{}, fmt.Errorf("validate mapped bar: %w", err)
	}

	return bar, nil
}
