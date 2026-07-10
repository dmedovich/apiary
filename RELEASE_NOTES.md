# apiary v1.1.0

## What's new

### Form-data request bodies

Operations can now declare a `multipart/form-data` or
`application/x-www-form-urlencoded` request body via the `content-type:`
annotation:

```go
// UploadAvatar uploads a user avatar.
// apiary:operation POST /api/v1/users/{id}/avatar
// content-type: multipart/form-data
// tags: users
// errors: 400,413
func (h *UserHandler) UploadAvatar(c *gin.Context) { ... }

type UploadAvatarRequest struct {
    ID     int64                  `path:"id"`
    Avatar *multipart.FileHeader  `form:"avatar" doc:"Image file (JPEG or PNG)"`
    Alt    string                 `form:"alt"    doc:"Alt text"`
}
```

Generated output:

```yaml
requestBody:
  required: true
  content:
    multipart/form-data:
      schema:
        $ref: '#/components/schemas/UploadAvatarRequest'
```

The `avatar` field maps to `{type: string, format: binary}` automatically.

**Shorthands** accepted by `content-type:`:

| Write | Means |
|---|---|
| `multipart` | `multipart/form-data` |
| `form` | `application/x-www-form-urlencoded` |
| `urlencoded` | `application/x-www-form-urlencoded` |

### `form` struct tag support

Fields tagged with `form:"name"` now use that name in the generated schema
when no `json` tag is present — consistent with how gin binds form fields.

### `multipart.FileHeader` schema primitive

`*multipart.FileHeader` fields map to `{type: string, format: binary}`.

---

## Install

```bash
go install github.com/yaop-labs/apiary/cmd/apiary@v1.1.0
```

Pre-built binaries for linux/darwin/windows (amd64/arm64) are attached to this
release.

## Links

- Changelog: [CHANGELOG.md](CHANGELOG.md)
- Stability policy: [STABILITY.md](STABILITY.md)
- Documentation: **https://yaop-labs.github.io/apiary/**
