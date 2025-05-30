package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
)

const (
	maxRetries     = 10
	initialBackoff = 1 * time.Second
	maxBackoff     = 5 * time.Minute
)

type Price struct {
	Symbol string
	Price  string
	Time   int64
}

func getPriceChannel(ctx context.Context, wg *sync.WaitGroup) chan Price {

	priceChan := make(chan Price, 100)
	var stopChan chan struct{}

	wg.Add(1)
	go func() {

		defer wg.Done()
		defer close(priceChan)

		backoff := initialBackoff
		retryCount := 0

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Create new channel for this connection attempt
				stopChan = make(chan struct{})

				wsHandler := func(event *binance.WsKlineEvent) {
					select {
					case priceChan <- Price{
						Symbol: event.Symbol,
						Price:  event.Kline.Close,
						Time:   event.Kline.StartTime,
					}:
					case <-ctx.Done():
						return
					}
				}

				errHandler := func(err error) {
					log.Printf("WebSocket error: %v\n", err)
					close(stopChan) // Signal to stop the current connection
				}

				// Start the WebSocket connection
				_, _, err := binance.WsKlineServe("BTCUSDT", "1s", wsHandler, errHandler)
				if err != nil {
					log.Printf("Error starting WebSocket: %v\n", err)
					retryCount++
					if retryCount >= maxRetries {
						log.Printf("Max retries reached. Giving up.")
						return
					}
					// Calculate next backoff duration
					backoff = time.Duration(float64(backoff) * 1.5)
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
					log.Printf("Retrying in %v... (attempt %d/%d)", backoff, retryCount, maxRetries)
					time.Sleep(backoff)
					continue
				}

				// Reset backoff on successful connection
				backoff = initialBackoff
				retryCount = 0
				log.Println("Successfully connected to Binance WebSocket")

				// Wait for either context cancellation or connection error
				select {
				case <-ctx.Done():
					return
				case <-stopChan:
					log.Println("Connection lost, attempting to reconnect...")
					// Don't sleep here as we want to retry immediately on connection loss
					continue
				}
			}
		}
	}()

	return priceChan
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	// Create context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create WaitGroup for goroutines
	var wg sync.WaitGroup

	// Handle OS signals for graceful shutdown
	forever := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel() // Cancel the context
		forever <- struct{}{}
	}()

	// Get price channel with retry logic
	priceChan := getPriceChannel(ctx, &wg)

	// Open file for writing
	file, err := os.OpenFile("data.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	otc := NewOtc(0, time.Now().Unix())

	for {
		select {
		case data := <-priceChan:

			price, err := strconv.ParseFloat(data.Price, 64)
			if err != nil {

				fmt.Printf("Error parsing price: %v\n", err)
				continue
			}

			otc.SetAndGeneratePrice(price)

			diffPercent := (otc.generatedPrice - otc.price) / otc.price * 100

			// Write the update to the file
			file.WriteString(fmt.Sprintf("%.2f,%.2f,%v\n", otc.price, otc.generatedPrice, data.Time))

			fmt.Printf("Real: %v, Generated: %v, Diff in percent: %v\n", otc.price, otc.generatedPrice, diffPercent)

		case <-forever:
			fmt.Println("Program terminated")
			wg.Wait() // Wait for all goroutines to finish
			return
		}
	}
}
