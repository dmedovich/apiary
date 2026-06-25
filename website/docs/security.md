---
sidebar_position: 6
---

# Security

```bash
# Define JWT Bearer as the global default
apiary -security bearer ./...
```

This adds `BearerAuth` to `components/securitySchemes` and sets it as the global
`security` requirement. Individual operations override it with the `security:`
annotation:

```go
// apiary:operation POST /api/v1/auth/login
// security: none        ← public, no token required
func (h *AuthHandler) Login(...)

// apiary:operation GET /api/v1/admin/report
// security: bearer      ← explicit (self-documenting)
func (h *AdminHandler) Report(...)
```

## Built-in scheme names

| Name | Type | Details |
|------|------|---------|
| `bearer` | `http` | `scheme: bearer`, `bearerFormat: JWT` |
| `basic` | `http` | `scheme: basic` |
| `apikey` | `apiKey` | `in: header`, `name: X-API-Key` |

Use `myName:bearer` to register a built-in scheme under a custom name.
