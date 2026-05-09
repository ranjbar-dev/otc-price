package runtime

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ranjbar-dev/otc-price/internal/application"
	"github.com/ranjbar-dev/otc-price/internal/domain"
	binanceinfra "github.com/ranjbar-dev/otc-price/internal/infrastructure/binance"
	configinfra "github.com/ranjbar-dev/otc-price/internal/infrastructure/config"
	"github.com/ranjbar-dev/otc-price/internal/infrastructure/storage"
)

type App struct {
	config configinfra.Config
	logger *log.Logger
}

func NewApp(config configinfra.Config, logger *log.Logger) (*App, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &App{config: config, logger: logger}, nil
}

func (app *App) Run(parent context.Context) error {
	binanceinfra.ConfigureEndpoints(app.config.Binance.WSURL)

	repository, err := storage.NewJSONLatestBarRepository(map[domain.Symbol]string{
		domain.SymbolBTCUSDT: app.config.Storage.BTCUSDT,
		domain.SymbolETHUSDT: app.config.Storage.ETHUSDT,
	})
	if err != nil {
		return fmt.Errorf("create json repository: %w", err)
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	updates := make(chan domain.Bar, 16)
	processor := application.NewBarProcessor(updates, repository)
	streamer := binanceinfra.NewKlineStreamer(app.logger, app.config.Binance.ReconnectDelay)

	processorErrC := make(chan error, 1)
	streamerErrC := make(chan error, 1)

	go func() {
		processorErrC <- processor.Run(ctx)
	}()

	go func() {
		streamerErrC <- streamer.Stream(ctx, updates)
		close(updates)
	}()

	select {
	case <-parent.Done():
		cancel()
		return errors.Join(ignoreCanceled(<-streamerErrC), ignoreCanceled(<-processorErrC))
	case err := <-streamerErrC:
		cancel()
		return errors.Join(ignoreCanceled(err), ignoreCanceled(<-processorErrC))
	case err := <-processorErrC:
		cancel()
		return errors.Join(ignoreCanceled(err), ignoreCanceled(<-streamerErrC))
	}
}

func ignoreCanceled(err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
