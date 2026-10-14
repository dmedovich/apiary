# apiary v1.2.3

## Fixes

Type aliases now resolve to their target types when apiary is built with Go 1.27.
Previously, an alias response could become `schema: {}`, a slice of aliases
could become `items: {}`, and the target schemas could disappear from
`components.schemas`.

The fix preserves `$ref`s and component fields for aliases and alias chains,
including cross-package types, embedded aliases, and nullable pointer aliases
in annotations.

## Validation

A regression test references DTOs exclusively through aliases and verifies
response `$ref`s, array item `$ref`s, component presence, fields, and schema
count. CI now also tests Go 1.26 and Go 1.27.

## Install

```bash
go install github.com/yaop-labs/apiary/cmd/apiary@v1.2.3
```
