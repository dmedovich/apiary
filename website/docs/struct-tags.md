---
sidebar_position: 4
---

# Struct tags

| Tag | Effect |
|-----|--------|
| `json:"name"` | JSON field name. Use `"-"` to exclude. A pointer field (`*T`) becomes **nullable**. |
| `doc:"text"` | `description` in the schema / parameter |
| `example:"val"` | `example` in the schema / parameter |
| `default:"val"` | `default` in the schema |
| `validate:"..."` | go-playground/validator rules → JSON-Schema constraints (see [Validation](validation)) |
| `path:"name"` | Path parameter — matches `{name}` in the URL |
| `query:"name"` | Query parameter |
| `header:"name"` | Header parameter (e.g. `X-Currency`, `Authorization`) |

## Parameter routing

| Tag / condition | OpenAPI location |
|---|---|
| `path:"name"` | `parameters[in=path]` — always required |
| `query:"name"` | `parameters[in=query]` |
| `header:"name"` | `parameters[in=header]` |
| Remaining fields on `GET` | implicit query parameters |
| Remaining fields on `POST`/`PUT`/`PATCH` | JSON request body |

A non-GET request type whose fields are *all* path/query/header parameters emits
no request body.

## Enums

apiary detects named types with `const` values and adds `enum` automatically — no
annotation needed. Works with string and integer base types.

```go
type ProductCategory string

const (
    CategoryElectronics ProductCategory = "electronics"
    CategoryClothing    ProductCategory = "clothing"
    CategoryFood        ProductCategory = "food"
)
```
