package binanceinfra

import (
	"context"
	"log"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/ranjbar-dev/otc-price/internal/domain"
)

type KlineStreamer struct {
	logger         *log.Logger
	reconnectDelay time.Duration
}

func NewKlineStreamer(logger *log.Logger, reconnectDelay time.Duration) *KlineStreamer {
	return &KlineStreamer{
		logger:         logger,
		reconnectDelay: reconnectDelay,
	}
}

func ConfigureEndpoints(wsURL string) {
	binance.BaseWsMainURL = wsURL
	binance.BaseCombinedMainURL = strings.TrimSuffix(wsURL, "/ws") + "/stream?streams="
}

func (streamer *KlineStreamer) Stream(ctx context.Context, output chan<- domain.Bar) error {
	pairs := map[string]string{
		string(domain.SymbolBTCUSDT): string(domain.Interval1m),
		string(domain.SymbolETHUSDT): string(domain.Interval1m),
	}

	for {
		errC := make(chan error, 1)
		handler := func(event *binance.WsKlineEvent) {
			bar, err := MapWsKlineEvent(event)
			if err != nil {
				nonBlockingSend(errC, err)
				return
			}

			select {
			case <-ctx.Done():
			case output <- bar:
			}
		}
		errHandler := func(err error) {
			if err == nil {
				return
			}
			nonBlockingSend(errC, err)
		}

		doneC, stopC, err := binance.WsCombinedKlineServe(pairs, handler, errHandler)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if waitErr := streamer.waitForReconnect(ctx); waitErr != nil {
				return waitErr
			}
			continue
		}

		for {
			select {
			case <-ctx.Done():
				close(stopC)
				<-doneC
				return nil
			case err := <-errC:
				streamer.logger.Printf("binance websocket error: %v", err)
				close(stopC)
				<-doneC
				if waitErr := streamer.waitForReconnect(ctx); waitErr != nil {
					return waitErr
				}
				goto reconnect
			case <-doneC:
				streamer.logger.Printf("binance websocket closed, reconnecting")
				if waitErr := streamer.waitForReconnect(ctx); waitErr != nil {
					return waitErr
				}
				goto reconnect
			}
		}

	reconnect:
	}
}

func (streamer *KlineStreamer) waitForReconnect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(streamer.reconnectDelay):
		return nil
	}
}

func nonBlockingSend(target chan<- error, err error) {
	select {
	case target <- err:
	default:
	}
}
