<p align="center">
  <img src="media/logo.png" alt="apiary logo" width="200">
</p>

<h1 align="center">apiary</h1>

<p align="center">
  <em>OpenAPI 3.1 generator for Go — driven by types, not comment soup.</em>
</p>

<p align="center">
  <a href="https://github.com/honeynil/apiary/actions/workflows/ci.yml"><img src="https://github.com/honeynil/apiary/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/honeynil/apiary"><img src="https://goreportcard.com/badge/github.com/honeynil/apiary" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/honeynil/apiary"><img src="https://pkg.go.dev/badge/github.com/honeynil/apiary.svg" alt="Go Reference"></a>
  <a href="https://github.com/honeynil/apiary/releases"><img src="https://img.shields.io/github/v/release/honeynil/apiary?include_prereleases&sort=semver" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/honeynil/apiary" alt="License"></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/OpenAPI-3.1-6BA539?logo=openapiinitiative&logoColor=white" alt="OpenAPI 3.1">
  <img src="https://img.shields.io/badge/framework-net%2Fhttp%20%7C%20gin-00ADD8?logo=go&logoColor=white" alt="net/http | gin">
  <img src="https://img.shields.io/badge/types-checked%20via%20go%2Ftypes-00ADD8?logo=go&logoColor=white" alt="Type-checked via go/types">
</p>

<p align="center">
  <strong><a href="https://honeynil.github.io/apiary/">📖 Documentation</a></strong>
  ·
  <a href="https://honeynil.github.io/apiary/api/">Live API Explorer</a>
  ·
  <a href="CHANGELOG.md">Changelog</a>
</p>

---

**apiary** generates an [OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.0)
document from annotated Go source code. Your function signatures and struct types
are the source of truth — no schema descriptions duplicated in comments.

```go
// CreateUser registers a new account.
// apiary:operation POST /api/v1/users
// tags: users
// errors: 400,409,500
func (h *UserHandler) CreateUser(ctx context.Context, req CreateUserRequest) (UserDTO, error) {
    // business logic — apiary never touches this
}
```

The request/response types, the `operationId`, and the summary (from the godoc)
are all inferred. `validate:"..."` tags become JSON-Schema constraints; pointers
become nullable; enums are detected automatically.

## Install

```bash
go install github.com/honeynil/apiary/cmd/apiary@latest
```

## Quickstart

```bash
apiary ./...                                   # scan module → openapi.yaml
apiary -security bearer -out docs/api.yaml ./... # JWT default + custom output
apiary serve ./...                             # live Swagger UI on :8080
apiary -check -out docs/api.yaml ./...          # CI: fail if the spec is stale
```

## Highlights

- **Types, not comment soup** — signatures + struct tags are the contract.
- **Real `go/types` analysis** — cross-package, imported, and generic types resolve.
- **Rich schemas for free** — validator tags → constraints, nullable pointers, enums.
- **OpenAPI 3.1**, `operationId` for client codegen, JSON or YAML output.
- **Honest** — diagnostics for bad signatures, typos, and collisions.

## Documentation

Full docs, examples, and the live API explorer live at
**[honeynil.github.io/apiary](https://honeynil.github.io/apiary/)**:

- [Annotation format](https://honeynil.github.io/apiary/annotations)
- [Struct tags](https://honeynil.github.io/apiary/struct-tags) & [Validation](https://honeynil.github.io/apiary/validation)
- [Security](https://honeynil.github.io/apiary/security) & [Frameworks](https://honeynil.github.io/apiary/frameworks)
- [CLI reference](https://honeynil.github.io/apiary/cli)
- [Migrating from swaggo](https://honeynil.github.io/apiary/migrating-from-swaggo)

Stability policy: [STABILITY.md](STABILITY.md). Contributions welcome.

## License

[MIT](LICENSE)
