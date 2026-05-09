---
applyTo: "**"
---

# Project Copilot Instructions

## Project overview

This repository is a Go application that connects to the Binance public API, subscribes to live 1-minute kline updates for BTCUSDT and ETHUSDT only, keeps the latest bars in RAM, and persists the latest state to local JSON files.

The codebase must follow domain-driven design with clear application, domain, infrastructure, and interface boundaries. Prefer simple, explicit code over abstractions that hide data flow.

## Core constraints

- Use Go modules and standard Go project conventions.
- Do not add Docker, Docker Compose, container files, or container-specific documentation.
- Put runtime configuration in `config/config.yml`.
- Persist local JSON data only to `data/btcusdt.json` and `data/ethusdt.json`.
- Subscribe to exactly two symbols only: `BTCUSDT` and `ETHUSDT`.
- Handle 1-minute bars only.
- Keep the latest bars in RAM using maps owned by a single goroutine.
- Do not use `sync.Mutex`, `sync.RWMutex`, atomics, or other lock-based coordination.
- Coordinate concurrency with channels and goroutine ownership.

## Architecture

Use DDD boundaries similar to this:

- `internal/domain`: entities, value objects, repository ports, domain services.
- `internal/application`: use cases, orchestration, command/query DTOs.
- `internal/infrastructure`: Binance client adapter, JSON storage adapter, config loader, logging.
- `internal/interfaces` or `internal/delivery`: app bootstrap, workers, transport-facing wiring.
- `cmd/<app>`: composition root only.

Keep dependencies pointing inward. Domain code must not import infrastructure packages.

## Domain model expectations

Represent a latest bar with explicit fields such as symbol, interval, open time, close time, open, high, low, close, volume, event time, and closed/final flag when provided by Binance.

Prefer strongly typed domain structs over `map[string]any`.

## Concurrency rules

- One goroutine should own the in-memory latest-bars map.
- Other goroutines communicate updates through channels.
- Avoid shared mutable state across goroutines.
- Shutdown should be explicit and channel-driven.

Correct pattern:

```go
type LatestBarStore struct {
    updates chan Bar
    queries chan chan map[string]Bar
}
```

Avoid patterns that share a map between goroutines and then protect it with locks.

## Binance integration

- Use a Binance public market data client appropriate for Go.
- Prefer the dependency already present in `go.mod` unless the user explicitly asks to replace it.
- Subscribe only to BTCUSDT and ETHUSDT 1m topics.
- Parse websocket kline payloads into domain structs immediately at the infrastructure boundary.
- Reconnection logic should be simple and explicit.

## Persistence rules

- Store the latest bar snapshot for each symbol as formatted JSON.
- Write to `data/btcusdt.json` and `data/ethusdt.json` only.
- Ensure the `data` directory is created when missing.
- Keep the persistence adapter focused on file IO and serialization.

## Configuration

All runtime configuration belongs in `config/config.yml`.

Expected configuration areas:

- Binance websocket/base settings.
- Allowed symbols.
- Interval.
- JSON file paths.
- Logging level if needed.

Do not spread configuration constants across the codebase when they belong in the YAML file.

## Style and implementation

- Prefer small packages and focused files.
- Use constructor functions for services and adapters.
- Return explicit errors with context.
- Keep `main.go` thin.
- Prefer standard library utilities unless a dependency adds clear value.
- Add tests for parsing, storage, and application coordination when implementing behavior.

## Avoid

- No mutex-based state protection.
- No support for symbols beyond BTCUSDT and ETHUSDT.
- No databases, Redis, or external persistence.
- No REST server unless explicitly requested.
- No global mutable package state.
- No Docker artifacts.
