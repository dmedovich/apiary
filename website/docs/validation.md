---
sidebar_position: 5
---

# Validation constraints

apiary maps the common [go-playground/validator](https://github.com/go-playground/validator)
rules you already write onto JSON-Schema constraints, with no extra annotations.

| `validate:` rule | JSON Schema (by field type) |
|---|---|
| `required` | adds the field to `required` |
| `min=N` / `max=N` | `minimum/maximum` (number), `minLength/maxLength` (string), `minItems/maxItems` (array) |
| `gt`/`gte`/`lt`/`lte=N` | `exclusiveMinimum`/`minimum`/`exclusiveMaximum`/`maximum` (numbers) |
| `len=N` | exact length / item count |
| `oneof=a b c` | `enum: [a, b, c]` |
| `email` | `format: email` |
| `uuid` | `format: uuid` |
| `url` / `uri` | `format: uri` |
| `ipv4` / `ipv6` | `format: ipv4` / `ipv6` |
| `hostname` | `format: hostname` |

Unknown rules are ignored. Constraints apply to request-body fields **and** to
path/query/header parameters.

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=32"`
    Email    string `json:"email"    validate:"required,email"`
    Role     string `json:"role"     validate:"oneof=admin user guest"`
    Age      *int   `json:"age,omitempty" validate:"gte=13,lte=120"` // nullable + bounded
}
```

generates:

```yaml
username: { type: string, minLength: 3, maxLength: 32 }
email:    { type: string, format: email }
role:     { type: string, enum: [admin, user, guest] }
age:      { type: [integer, "null"], minimum: 13, maximum: 120 }
required: [username, email]
```

## Go type to JSON Schema mapping

| Go type | JSON Schema |
|---------|-------------|
| `string` / `bool` | `{type: string}` / `{type: boolean}` |
| `int`, `int32` / `int64` | `{type: integer, format: int32 / int64}` |
| `float32` / `float64` | `{type: number, format: float / double}` |
| `time.Time` | `{type: string, format: date-time}` |
| `uuid.UUID` | `{type: string, format: uuid}` |
| `[]T` / `map[K]V` | array / object |
| Struct | `$ref` to a component |
| `*T` (scalar) | nullable (`type: [T, "null"]`) |
| `*Struct` | nullable (`anyOf: [{$ref}, {type: "null"}]`) |

`example:` and `default:` tag values are coerced to the field's JSON type, so
`example:"42"` on an integer field renders as `example: 42`, not `"42"`.
