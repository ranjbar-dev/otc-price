# Implementation Prompt

Build a Go application in domain-driven design that connects to the Binance public API and streams latest 1-minute kline updates for exactly two symbols only: BTCUSDT and ETHUSDT.

Requirements:

- Use Go.
- Follow DDD with clear `domain`, `application`, `infrastructure`, and `cmd` boundaries.
- Use channels for concurrency and ownership of the in-memory map.
- Do not use any mutex, RWMutex, atomic, or lock-based synchronization.
- Keep the latest bar for each symbol in RAM in a map owned by one goroutine.
- Subscribe only to Binance public topics for BTCUSDT 1m and ETHUSDT 1m.
- Persist the latest BTCUSDT bar to `data/btcusdt.json` and the latest ETHUSDT bar to `data/ethusdt.json`.
- Put all runtime configuration in `config/config.yml`.
- Create the `data` directory automatically if it does not exist.
- Keep the code simple, explicit, and production-readable.
- Do not add Docker, Compose, containers, databases, or HTTP APIs.
- Reuse the Binance Go dependency already present in `go.mod` unless replacement is explicitly necessary.

Implementation expectations:

- Create domain entities/value objects for bars.
- Create an application service/use case that receives bar updates and coordinates RAM + JSON persistence.
- Create infrastructure adapters for Binance websocket subscription, config loading, and JSON file storage.
- Make shutdown and reconnection explicit.
- Add focused tests for bar mapping, JSON persistence, and channel-driven state ownership.

Deliverables:

- `config/config.yml`
- `cmd/...` entrypoint
- `internal/domain/...`
- `internal/application/...`
- `internal/infrastructure/...`
- `data/btcusdt.json`
- `data/ethusdt.json`
- tests for the core flows