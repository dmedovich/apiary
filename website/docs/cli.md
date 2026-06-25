---
sidebar_position: 8
---

# CLI reference

```bash
apiary [flags] [patterns...]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-out` | `openapi.yaml` | Output file path. Use `-` for stdout. |
| `-title` | `API` | `info.title` |
| `-version` | `0.0.1` | `info.version` |
| `-description` | _(none)_ | `info.description` |
| `-security` | _(none)_ | Global scheme: `bearer`, `basic`, `apikey`, or `myName:bearer` |
| `-server` | _(none)_ | Comma-separated server URL(s) for `servers` |
| `-format` | _(by `-out` ext)_ | `yaml` or `json` |
| `-check` | `false` | Verify `-out` is up to date; exit non-zero if it differs |
| `-C` | _(cwd)_ | Load packages as if run from this directory |
| `-V` | `false` | Print the apiary version and exit |

apiary prints `warning:` diagnostics to stderr for unsupported signatures,
unknown annotation keys, unsupported HTTP methods, and duplicate path+method or
`operationId`, so silent omissions become visible.

## CI check

```bash
# Fail the build if the committed spec is stale (like `gofmt -l`)
apiary -check -out docs/openapi.yaml ./...
```

## JSON output

```bash
apiary -out openapi.json ./...     # inferred from extension
apiary -format json -out - ./...   # JSON to stdout
```

## Config file

Drop an `apiary.yaml` (or `.apiary.yaml`) in the working directory; handy with
`//go:generate apiary`. Explicit CLI flags override it.

```yaml
title: My API
version: 1.0.0
security: [bearer]
servers: [https://api.example.com]
out: docs/openapi.yaml
format: yaml
patterns: ["./internal/handler/...", "./internal/dto/..."]
```

## Live preview

```bash
apiary serve -addr :8080 ./...
```

Serves Swagger UI at `http://localhost:8080` (spec at `/openapi.yaml` and
`/openapi.json`), regenerating on every refresh. UI assets load from a CDN.
