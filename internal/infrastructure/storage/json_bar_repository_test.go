package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ranjbar-dev/otc-price/internal/domain"
)

func TestJSONLatestBarRepositorySaveCreatesDirectoryAndWritesJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	repository, err := NewJSONLatestBarRepository(map[domain.Symbol]string{
		domain.SymbolBTCUSDT: filepath.Join(tempDir, "data", "btcusdt.json"),
		domain.SymbolETHUSDT: filepath.Join(tempDir, "data", "ethusdt.json"),
	})
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	bar := domain.Bar{
		Symbol:    domain.SymbolBTCUSDT,
		Interval:  domain.Interval1m,
		OpenTime:  100,
		CloseTime: 200,
		Open:      "1",
		High:      "2",
		Low:       "0.5",
		Close:     "1.5",
		Volume:    "42",
		EventTime: 300,
		IsClosed:  true,
	}

	if err := repository.Save(context.Background(), bar); err != nil {
		t.Fatalf("save: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tempDir, "data", "btcusdt.json"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	var stored domain.Bar
	if err := json.Unmarshal(content, &stored); err != nil {
		t.Fatalf("unmarshal saved file: %v", err)
	}

	if stored != bar {
		t.Fatalf("stored bar mismatch: got %+v want %+v", stored, bar)
	}
}
