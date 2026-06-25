---
sidebar_position: 10
title: Migrating from swaggo
---

# Migrating from swaggo

apiary and [swaggo](https://github.com/swaggo/swag) solve the same problem from
opposite directions: swaggo encodes the contract in comments; apiary reads it
from your **types**.

## Annotation mapping

| swaggo | apiary |
|--------|--------|
| `@Summary` / `@Description` | first/rest of the Go doc comment, or `summary:` / `description:` |
| `@Tags a,b` | `tags: a, b` |
| `@Param body body T true ""` | the request parameter type `T` in the signature (or `request: T`) |
| `@Success 200 {object} R` | the response type `R` in the signature (or `response: R`) |
| `@Failure 400 {object} E` | `errors: 400 E` |
| `@Router /p [post]` | `apiary:operation POST /p` |
| `@Security BearerAuth` | `security: bearer` (or global `-security bearer`) |
| `@ID name` | `operationId: name` (auto-derived by default) |

## What you gain

- No duplicated schema descriptions; struct tags and signatures are the truth.
- OpenAPI **3.1** output.
- Validation constraints from `validate:` tags you already write.
- Real `go/types` resolution (cross-package, generics) instead of comment parsing.

## What to check

- apiary expects one of the [supported handler signatures](annotations#supported-handler-signatures).
  For framework handlers (gin, net/http) add `request:` / `response:` annotations.
- The scanned code must compile (apiary type-checks it).
