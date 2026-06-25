# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-06-25

### Added
- **Diagnostics.** apiary now warns (with source positions) on annotated
  functions with an unsupported signature, unknown annotation keys, unsupported
  HTTP methods, duplicate path+method, and duplicate `operationId`.
- **`operationId`.** Every operation gets a stable, unique id derived from the
  receiver and method (`TaskHandler.List` -> `taskList`); override with
  `operationId:`.
- **Summary from godoc.** When `summary:` is absent, the handler's Go doc comment
  becomes the summary/description (leading function name stripped).
- **Validation constraints.** `validate:"..."` tags map to JSON-Schema
  constraints (`min`/`max`/`len`/`gt`/`gte`/`lt`/`lte`/`oneof`/`email`/`uuid`/
  `uri`/`ipv4`/`ipv6`/`hostname`) for both body fields and parameters.
- **Nullable.** Pointer fields render as nullable: scalars as
  `type: [T, "null"]`, struct pointers as `anyOf: [{$ref}, {type: "null"}]`.
- **Typed examples & defaults.** `example:`/`default:` tag values are coerced to
  the field's JSON type (e.g. `example: 42`, not `"42"`).
- **CLI.** `-check` (CI staleness gate), `-format yaml|json` / `.json` output,
  `-server`, `-C <dir>`, `-V`, and an optional `apiary.yaml` config file.
- **`apiary serve`.** Live Swagger UI (CDN assets) that regenerates on refresh.
- **Golden snapshot tests** guard generated output against regressions.

### Changed
- **Type resolution now uses `go/packages` + `go/types`.** Types are resolved
  semantically across packages, through imports, and including generics.
  In-module types are expanded into real component schemas; well-known external
  types map to their formats. The scanned code must compile (its dependencies
  must be available), the same requirement as `go build`.
- Path/query/header parameters now carry validation constraints (previously only
  body fields did).
- A non-GET request type whose fields are all path/query/header parameters no
  longer emits an empty request body.

### Removed
- Dead `internal/schema/resolver.go` and unused field metadata.

## [0.0.1]

- Initial release: OpenAPI 3.1 generation from `// apiary:operation` annotations
  for net/http and gin handlers, enums, embedded structs, and security schemes.
