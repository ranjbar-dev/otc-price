package application

import (
	"context"
	"testing"
	"time"

	"github.com/ranjbar-dev/otc-price/internal/domain"
)

type recordingRepository struct {
	saves chan domain.Bar
	err   error
}

func (repository *recordingRepository) Save(_ context.Context, bar domain.Bar) error {
	if repository.err != nil {
		return repository.err
	}
	repository.saves <- bar
	return nil
}

func TestBarProcessorOwnsStateThroughChannels(t *testing.T) {
	t.Parallel()

	updates := make(chan domain.Bar, 1)
	repository := &recordingRepository{saves: make(chan domain.Bar, 1)}
	processor := NewBarProcessor(updates, repository)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runDone := make(chan error, 1)
	go func() {
		runDone <- processor.Run(ctx)
	}()

	bar := domain.Bar{
		Symbol:    domain.SymbolBTCUSDT,
		Interval:  domain.Interval1m,
		OpenTime:  1,
		CloseTime: 2,
		Open:      "1.0",
		High:      "1.2",
		Low:       "0.9",
		Close:     "1.1",
		Volume:    "123",
		EventTime: 3,
		IsClosed:  false,
	}

	updates <- bar

	select {
	case saved := <-repository.saves:
		if saved != bar {
			t.Fatalf("saved bar mismatch: got %+v want %+v", saved, bar)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for persistence")
	}

	snapshot, err := processor.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	if got := snapshot[domain.SymbolBTCUSDT]; got != bar {
		t.Fatalf("snapshot bar mismatch: got %+v want %+v", got, bar)
	}

	cancel()

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for processor shutdown")
	}
}
