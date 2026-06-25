---
sidebar_position: 3
---

# Annotation format

Place the marker **directly above** the function, with no blank lines:

```
// apiary:operation METHOD /path
// summary: One-line summary
// description: Longer description (may contain colons)
// tags: tag1, tag2
// security: bearer          (optional, overrides global)
// errors: 400,401,403,500
// operationId: customId     (optional, overrides the auto-derived id)
```

## Summary from your Go doc comment

You usually don't need a `summary:` line; apiary falls back to the handler's
**godoc**. Prose lines *above* the marker become the summary (first line) and
description (the rest); a leading function name is stripped:

```go
// CreateUser registers a new account.
// It validates the email and sends a confirmation.
// apiary:operation POST /api/v1/users
func (h *UserHandler) CreateUser(ctx context.Context, req CreateUserRequest) (UserDTO, error)
```

Produces `summary: Registers a new account.` and
`description: It validates the email and sends a confirmation.`
An explicit `summary:` / `description:` always wins.

## operationId

Every operation gets a stable, unique `operationId` (great for client
generators), derived from the receiver and method name:
`(h *UserHandler) CreateUser` -> `userCreate`, `(h *CommentHandler) List` ->
`commentList`. Free functions use the function name. Override per-operation with
`operationId:`. apiary warns on collisions.

## Supported handler signatures

```go
func (h *T) A(ctx context.Context, req MyRequest) (MyResponse, error) // standard
func (h *T) B(req MyRequest) (MyResponse, error)                       // no ctx
func (h *T) C(ctx context.Context) (MyResponse, error)                 // no request body
func (h *T) D() (MyResponse, error)                                    // health-check style
func Handler(c *gin.Context)                                           // gin
func Handler(w http.ResponseWriter, r *http.Request)                   // net/http
```

## Error responses

`errors: 400,401,500` adds a response entry for each code, all sharing the
built-in `ErrorResponse` schema. Append a type name to use a custom schema for a
specific code:

```go
// errors: 400 ValidationError, 401, 500
```

Here `400` references `ValidationError`; `401` and `500` use `ErrorResponse`.
