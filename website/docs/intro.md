---
slug: /
title: apiary
sidebar_label: Introduction
sidebar_position: 1
---

# apiary

**OpenAPI 3.1 generator for Go — driven by types, not comment soup.**

apiary generates an [OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.0) document
from annotated Go source code. It type-checks your code with `go/types`, so your
function signatures and struct types are the source of truth — no duplicated
schema descriptions in comments.

## Before (swaggo)

```go
// TelegramAuth godoc
// @Summary      Authenticate via Telegram
// @Param        body body TelegramAuthRequest true "Request body"
// @Success      200 {object} TelegramAuthResponse
// @Failure      400 {object} ErrorResponse
// @Router       /api/v1/auth/telegram [post]
func (h *AuthHandler) TelegramAuth(w http.ResponseWriter, r *http.Request) { ... }
```

## After (apiary)

```go
// TelegramAuth verifies Telegram WebApp initData and issues a session.
// apiary:operation POST /api/v1/auth/telegram
// tags: auth
// security: none
// errors: 400,401,500
func (h *AuthHandler) TelegramAuth(ctx context.Context, req TelegramAuthRequest) (TelegramAuthResponse, error) {
    // business logic — apiary never touches this
}
```

The request and response types, the `operationId`, and even the summary (from the
Go doc comment) are inferred. Continue with [Installation](installation) and the
[Quickstart](installation#quickstart), or jump into the live
[API Explorer](/api/).

## Why apiary

- **Types, not comment soup** — signatures and struct tags are the contract.
- **Real type analysis** — cross-package, imported, and generic types resolve via
  `go/types`.
- **Rich schemas for free** — `validate:"..."` tags become JSON-Schema
  constraints; pointers become nullable; enums are detected automatically.
- **Honest output** — diagnostics for unsupported signatures, typos, and
  collisions; a `-check` mode keeps your committed spec from drifting.
