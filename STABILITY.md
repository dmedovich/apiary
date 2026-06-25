# Stability and compatibility

Starting with **v1.0.0**, apiary commits to backward compatibility for its
public surface under [Semantic Versioning](https://semver.org). Within the v1
series, the following will not change in a breaking way:

## Stable surface

- **Annotation format.** The `// apiary:operation METHOD /path` marker and its
  keys: `summary`, `description`, `tags`, `errors`, `security`, `request`,
  `response`, `operationId`.
- **Struct tags.** `json`, `doc`, `example`, `default`, `validate`, `path`,
  `query`, `header`, and the documented `validate` to JSON-Schema constraint
  mapping.
- **Supported handler signatures.** The `(R, error)`-returning shapes, gin, and
  net/http handlers described in the README.
- **CLI flags.** `-out`, `-title`, `-version`, `-description`, `-security`,
  `-server`, `-format`, `-check`, `-C`, `-V`, the `apiary serve` subcommand, and
  the `apiary.yaml` config keys.

## Not covered by the guarantee

- The exact byte layout / field ordering of generated YAML/JSON (only its
  validity as an OpenAPI 3.1 document and the semantics above).
- Diagnostic warning wording.
- Internal packages under `internal/` (not importable anyway).
- Behavior for input Go code that does not compile.

## Deprecation policy

Deprecated features will be announced in `CHANGELOG.md`, continue to work for the
remainder of the v1 series, and be removed no earlier than the next major
version (v2).
