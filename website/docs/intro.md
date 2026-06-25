---
title: Introduction
sidebar_label: Introduction
sidebar_position: 1
---

# apiary

**OpenAPI 3.1 generator for Go: driven by types, not comment soup.**

apiary generates an [OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.0) document
from annotated Go source code. It type-checks your code with `go/types`, so your
function signatures and struct types are the source of truth, with no
duplicated schema descriptions in comments.

## Before (swaggo)

Every operation needs a block of `@`-annotations that restate what the code
already says:

```go
// CreateUser godoc
// @Summary      Create a user
// @Param        body  body      CreateUserRequest  true  "Request body"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/v1/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) { ... }
```

## After (apiary)

One marker line. The request and response types come from the signature, and the
struct tags you already write become the schema:

```go
// CreateUser registers a new account.
// apiary:operation POST /api/v1/users
// tags: users
// errors: 400,409,500
func (h *UserHandler) CreateUser(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
    // business logic; apiary never touches this
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=32"`
    Email    string `json:"email"    validate:"required,email"`
    Age      *int   `json:"age,omitempty" validate:"gte=13,lte=120"`
}
```

apiary infers the request/response schemas, the `operationId`, and the summary
(from the Go doc comment). The `validate:` tags become JSON-Schema constraints
(length limits, `email` format), and the `*int` pointer becomes a nullable,
bounded `age`. Continue with [Installation](installation) and the
[Quickstart](installation#quickstart), or see the input and generated output
side by side in [Examples](examples).

## Why apiary

- **Types, not comment soup.** Signatures and struct tags are the contract.
- **Real type analysis.** Cross-package, imported, and generic types resolve via
  `go/types`.
- **Rich schemas for free.** `validate:"..."` tags become JSON-Schema
  constraints; pointers become nullable; enums are detected automatically.
- **Honest output.** Diagnostics for unsupported signatures, typos, and
  collisions; a `-check` mode keeps your committed spec from drifting.
