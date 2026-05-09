package application

import (
	"context"
	"fmt"

	"github.com/ranjbar-dev/otc-price/internal/domain"
)

type snapshotRequest struct {
	response chan snapshotResponse
}

type snapshotResponse struct {
	bars map[domain.Symbol]domain.Bar
	err  error
}

type BarProcessor struct {
	updates    <-chan domain.Bar
	queries    chan snapshotRequest
	repository domain.LatestBarRepository
}

func NewBarProcessor(updates <-chan domain.Bar, repository domain.LatestBarRepository) *BarProcessor {
	return &BarProcessor{
		updates:    updates,
		queries:    make(chan snapshotRequest),
		repository: repository,
	}
}

func (processor *BarProcessor) Run(ctx context.Context) error {
	latest := make(map[domain.Symbol]domain.Bar, 2)

	for {
		select {
		case <-ctx.Done():
			return nil
		case request := <-processor.queries:
			request.response <- snapshotResponse{bars: cloneBars(latest)}
		case bar, ok := <-processor.updates:
			if !ok {
				return nil
			}
			if err := bar.Validate(); err != nil {
				return fmt.Errorf("validate bar: %w", err)
			}
			latest[bar.Symbol] = bar
			if err := processor.repository.Save(ctx, bar); err != nil {
				return fmt.Errorf("persist latest bar for %s: %w", bar.Symbol, err)
			}
		}
	}
}

func (processor *BarProcessor) Snapshot(ctx context.Context) (map[domain.Symbol]domain.Bar, error) {
	response := make(chan snapshotResponse, 1)
	request := snapshotRequest{response: response}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case processor.queries <- request:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-response:
		return result.bars, result.err
	}
}

func cloneBars(source map[domain.Symbol]domain.Bar) map[domain.Symbol]domain.Bar {
	cloned := make(map[domain.Symbol]domain.Bar, len(source))
	for symbol, bar := range source {
		cloned[symbol] = bar
	}
	return cloned
}
