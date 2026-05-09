---
name: "Config File Rules"
description: "Rules for YAML configuration files used by the Go service"
applyTo: "config/**/*.yml,config/**/*.yaml"
---

# Config File Rules

## Source of truth

`config/config.yml` is the primary runtime configuration file.

## Required sections

Include only settings relevant to this project, for example:

```yaml
binance:
  ws_url: wss://stream.binance.com:9443/ws
symbols:
  - BTCUSDT
  - ETHUSDT
interval: 1m
storage:
  btcusdt: data/btcusdt.json
  ethusdt: data/ethusdt.json
```

## Rules

- Keep symbol scope fixed to BTCUSDT and ETHUSDT.
- Keep interval fixed to `1m` unless the user explicitly requests a change.
- Use relative local paths for JSON persistence.
- Do not add Docker or container settings.
