package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	configinfra "github.com/ranjbar-dev/otc-price/internal/infrastructure/config"
	"github.com/ranjbar-dev/otc-price/internal/interfaces/runtime"
)

func main() {
	logger := log.New(os.Stdout, "otc-price ", log.LstdFlags|log.LUTC)

	cfg, err := configinfra.Load("config/config.yml")
	if err != nil {
		logger.Printf("load config: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := runtime.NewApp(cfg, logger)
	if err != nil {
		logger.Printf("build app: %v", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		logger.Printf("run app: %v", err)
		os.Exit(1)
	}
}
