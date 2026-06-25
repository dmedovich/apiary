# apiary v1.0.0

**OpenAPI 3.1 for Go: driven by types, not comment soup.**

The first stable release. apiary type-checks your Go code and turns function
signatures and struct tags into a complete, valid OpenAPI 3.1 document, with no
schema descriptions duplicated in comments.

```go
// CreateUser registers a new account.
// apiary:operation POST /api/v1/users
// tags: users
// errors: 400,409,500
func (h *UserHandler) CreateUser(ctx context.Context, req CreateUserRequest) (UserDTO, error) {
    // business logic; apiary never touches this
}
```

The request/response types, the `operationId`, and the summary (from the godoc)
are all inferred.

---

## Highlights

- **Real type analysis.** Powered by `go/packages` + `go/types`: cross-package,
  imported, and generic types resolve semantically. In-module structs expand
  into real component schemas; well-known external types map to their formats.
- **Rich schemas for free.** `validate:"..."` tags become JSON-Schema
  constraints, pointers become nullable, enums are detected automatically, and
  `example:`/`default:` values are typed to match the field.
- **Built for clients & CI.** Stable, unique `operationId`s for code
  generators; JSON or YAML output; a `-check` mode that fails when the committed
  spec drifts.
- **Honest output.** Diagnostics (with source positions) for unsupported
  signatures, unknown annotation keys, unsupported methods, and duplicate
  path+method / operationId.
- **Live preview.** `apiary serve` serves Swagger UI that regenerates on refresh.

## What's included

**Authoring**
- `// apiary:operation METHOD /path` markers with `summary`, `description`,
  `tags`, `errors`, `security`, `request`, `response`, `operationId`.
- Summary/description fall back to the handler's Go doc comment.
- net/http and gin handlers (types via `request:`/`response:`).
- Struct tags: `json`, `doc`, `example`, `default`, `validate`, `path`,
  `query`, `header`.

**Schema generation**
- Validation constraints from validator rules (`min`/`max`/`len`/`gt`..`lte`/
  `oneof`/`email`/`uuid`/`uri`/`ip`/`hostname`) for body fields **and**
  path/query/header parameters.
- Nullable pointers: scalars as `type: [T, "null"]`, struct pointers as
  `anyOf: [{$ref}, {type: "null"}]`.
- Automatic enum detection (string and integer bases).
- Embedded structs via `allOf`.

**CLI**
- `-out`, `-title`, `-version`, `-description`, `-security`, `-server`,
  `-format yaml|json`, `-check`, `-C dir`, `-V`.
- Optional `apiary.yaml` config file (friendly to `//go:generate`).
- `apiary serve` live Swagger UI.

## Install

```bash
go install github.com/yaop-labs/apiary/cmd/apiary@v1.0.0
```

Pre-built binaries for linux/darwin/windows (amd64/arm64) are attached to this
release.

## Compatibility

apiary type-checks the packages it scans, so the code must compile (its
dependencies must be available), the same requirement as `go build`.

From v1.0.0 the annotation format, struct tags, supported handler signatures,
and CLI flags follow [Semantic Versioning](https://semver.org) and will not
change in a breaking way within the v1 series. See
[STABILITY.md](STABILITY.md).

## Documentation

Full docs and worked examples: **https://yaop-labs.github.io/apiary/**

## Links

- Changelog: [CHANGELOG.md](CHANGELOG.md)
- Stability policy: [STABILITY.md](STABILITY.md)
