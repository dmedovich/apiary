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

## Success responses

Success responses default to `200 OK` and `application/json`. `response:`
overrides the inferred Go response type (and supplies it for gin/net/http
handlers). Set `response-status:` to any code from 200 through 299 and
`response-content-type:` to the response's media type. These annotations do
not change the request's `content-type:` or JSON error responses.

```go
// apiary:operation POST /exports
// response: ExportJobStartedResponse
// response-status: 202
func StartExport(w http.ResponseWriter, r *http.Request) { /* ... */ }
```

For a file download, use `response-format: binary`. This explicitly replaces
the inferred response schema with `type: string, format: binary`; no Go DTO
or `response:` annotation is needed. The default binary media type is
`application/octet-stream`, or you can specify the file's MIME type:

```go
// apiary:operation GET /exports/{id}/download
// response-format: binary
// response-content-type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// errors: 404,500
func DownloadExport(w http.ResponseWriter, r *http.Request) { /* ... */ }
```

Invalid status codes, media types, or formats produce warnings and are
ignored. Currently `binary` is the only supported `response-format`.
Statuses 204 and 205 must have no response body; generation fails if a
response type or binary format is specified for them. A success status cannot
also appear in `errors:`; generation fails instead of replacing the success
schema with an error schema.

## Error responses

`errors: 400,401,500` adds a response entry for each code, all sharing the
built-in `ErrorResponse` schema. Append a type name to use a custom schema for a
specific code:

```go
// errors: 400 ValidationError, 401, 500
```

Here `400` references `ValidationError`; `401` and `500` use `ErrorResponse`.
