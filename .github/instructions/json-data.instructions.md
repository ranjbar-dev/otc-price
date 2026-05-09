---
name: "Local JSON Storage Rules"
description: "Rules for files and code that manage local JSON persistence for latest bars"
applyTo: "data/**/*.json"
---

# Local JSON Storage Rules

## Scope

The repository stores only the latest 1-minute bar snapshot per symbol.

## Allowed files

- `data/btcusdt.json`
- `data/ethusdt.json`

## Content expectations

- Store a single latest-bar JSON object per file.
- Keep field names stable and predictable.
- Prefer readable formatted JSON.
- Do not store historical arrays unless the user explicitly asks for history.
