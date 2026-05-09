---
name: "Go DDD Service Rules"
description: "DDD, channel ownership, Binance streaming, and local JSON persistence rules for Go files"
applyTo: "**/*.go"
---

# Go DDD Service Rules

Apply the general rules from `../copilot-instructions.md`.

## File intent by layer

- `internal/domain/**/*.go`: pure business types and interfaces only.
- `internal/application/**/*.go`: use cases and orchestration only.
- `internal/infrastructure/**/*.go`: external API, filesystem, config, and concrete adapters.
- `cmd/**/*.go`: wiring only.

## Bar processing

- Treat Binance websocket payloads as infrastructure DTOs.
- Convert DTOs into domain `Bar` values before sending them into the application layer.
- The application layer should receive normalized bar updates from a channel and decide how they are stored or persisted.

## In-memory store

Use channel ownership instead of locks.

Preferred shape:

```go
type BarUpdate struct {
    Symbol string
    Bar    domain.Bar
}

type RAMStore struct {
    updates <-chan BarUpdate
    persist chan<- domain.Bar
}
```

One goroutine owns the `map[string]domain.Bar` and performs all writes.

## Symbol and interval rules

- Accept only `BTCUSDT` and `ETHUSDT`.
- Accept only interval `1m`.
- Reject or ignore anything outside those constraints at the boundary.

## Error handling

- Wrap infrastructure errors with operation context.
- Avoid panics for expected runtime failures.
- Bubble errors up to composition root for shutdown or retry policy.

## Testing focus

- Unit test websocket payload-to-domain mapping.
- Unit test JSON repository read/write behavior.
- Unit test the channel-driven in-memory store behavior.
