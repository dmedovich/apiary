# apiary v1.2.0

## What's new

### Cross-package types in `request:`/`response:` annotations

Gin and net/http handlers can now reference request/response DTOs declared
in a different package than the handler itself:

```go
import (
    "myapp/internal/dto"
    "github.com/gin-gonic/gin"
)

// apiary:operation POST /api/v1/users
// summary: Create user
// request: dto.CreateUserRequest
// response: dto.UserResponse
func (h *UserHandler) CreateUser(c *gin.Context) { ... }
```

No more keeping a duplicate types file next to your handlers just so apiary
can see the type. Renamed imports (`d "myapp/internal/dto"`) work too.

### Fixes

Previously, a qualified `request:`/`response:` reference like `dto.User`
silently fell back to `{type: string}` (it matched the same fallback rule
used for external types like `time.Time`), and an unqualified reference to
a type in another package resolved to nothing — both without any warning.
Both cases now resolve correctly, and a genuinely unresolvable type now logs
a warning naming the type and its location instead of failing silently.

> **Note:** if your project was already (unintentionally) relying on a
> cross-package `request:`/`response:` reference, your generated spec will
> change — from a stray `string`/empty schema to the correct object schema.
> Regenerate and diff before publishing.

---

## Install

```bash
go install github.com/yaop-labs/apiary/cmd/apiary@v1.2.0
```

Pre-built binaries for linux/darwin/windows (amd64/arm64) are attached to this
release.

## Links

- Changelog: [CHANGELOG.md](CHANGELOG.md)
- Stability policy: [STABILITY.md](STABILITY.md)
- Documentation: **https://yaop-labs.github.io/apiary/**