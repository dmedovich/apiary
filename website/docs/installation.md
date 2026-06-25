---
sidebar_position: 2
---

# Installation

```bash
go install github.com/yaop-labs/apiary/cmd/apiary@latest
```

## Quickstart

```bash
# Scan the current module, write to openapi.yaml
apiary ./...

# With JWT security applied globally, custom metadata, and a chosen output path
apiary -security bearer -title "My API" -version "1.0.0" -out docs/openapi.yaml ./...
```

apiary type-checks the packages it scans, so the code must compile (its
dependencies must be available), the same requirement as `go build`.

See the [CLI reference](cli) for every flag, JSON output, the `-check` CI mode,
the `apiary.yaml` config file, and `apiary serve`.
