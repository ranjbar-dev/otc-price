package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ranjbar-dev/otc-price/internal/domain"
)

type JSONLatestBarRepository struct {
	pathsBySymbol map[domain.Symbol]string
}

func NewJSONLatestBarRepository(pathsBySymbol map[domain.Symbol]string) (*JSONLatestBarRepository, error) {
	if len(pathsBySymbol) != 2 {
		return nil, fmt.Errorf("expected two storage paths")
	}
	if pathsBySymbol[domain.SymbolBTCUSDT] == "" || pathsBySymbol[domain.SymbolETHUSDT] == "" {
		return nil, fmt.Errorf("storage paths for BTCUSDT and ETHUSDT are required")
	}

	return &JSONLatestBarRepository{pathsBySymbol: pathsBySymbol}, nil
}

func (repository *JSONLatestBarRepository) Save(_ context.Context, bar domain.Bar) error {
	path, ok := repository.pathsBySymbol[bar.Symbol]
	if !ok {
		return fmt.Errorf("no storage path configured for %s", bar.Symbol)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	content, err := json.MarshalIndent(bar, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bar: %w", err)
	}

	content = append(content, '\n')
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write bar file: %w", err)
	}

	return nil
}
