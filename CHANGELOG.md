# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-08-08

### Fixed
- **Cross-package `request:`/`response:` annotations.** A qualified type
  reference (e.g. `dto.CreateUserRequest`) in a `request:`/`response:`
  annotation now resolves against the handler file's imports, including
  renamed imports (`d "myapp/internal/dto"`). Previously, any reference
  containing a `.` fell through the same fallback used for well-known
  external types (`time.Time`, `uuid.UUID`, etc.) and silently rendered as
  `{type: string}`; an unqualified reference to a type declared in another
  package resolved to nothing, producing an empty schema with no warning.
  Handlers using gin or net/http's untyped signature no longer need their
  DTOs duplicated into the handler's own package just so apiary can see
  them.
- **Unresolvable annotation types now warn.** A `request:`/`response:`
  annotation naming a type apiary can't find (typo, missing import, etc.)
  now logs a warning with the type name and source location, instead of
  silently falling back to a broken schema.

> **Note:** if a project was already (unintentionally) relying on a
> cross-package `request:`/`response:` reference, its generated spec will
> change on upgrade — from a stray `string`/empty schema to the correct
> object schema. Regenerate and diff before publishing.

## [1.1.0] - 2026-07-10

### Added
- **`content-type:` annotation.** Operations can now declare a non-JSON request
  body with `content-type: multipart/form-data` or
`content-type: application/x-www-form-urlencoded` (shorthand: `multipart`,
`form`, `urlencoded`). The generated `requestBody` uses the specified media
  type instead of `application/json`.
- **`form` struct tag.** Fields tagged with `form:"name"` (e.g. for gin's
  form binding) use that name in the generated schema when no `json` tag is
  present.
- **`multipart.FileHeader` primitive.** Fields of type `*multipart.FileHeader`
  map to `{type: string, format: binary}` in the schema.

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