---
sidebar_position: 7
---

# Frameworks: gin & net/http

Both `func(c *gin.Context)` and `func(w http.ResponseWriter, r *http.Request)`
handlers are recognised. Because the signature carries no type information,
request and response types are given via annotations:

```go
// apiary:operation POST /api/v1/tasks
// summary: Create task
// tags: tasks
// request: CreateTaskRequest   (required for gin / net-http handlers)
// response: TaskDTO
// errors: 400,401,422,500
func CreateTask(c *gin.Context) {
    var req CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil { /* ... */ }
}
```

Slice responses work too:

```go
// response: []TaskDTO
```

Path, query, and header parameters still come from struct tags: the same
`path:`, `query:`, and `header:` tags used with standard handlers.

See the bundled examples:
[`testdata/router`](https://github.com/yaop-labs/apiary/tree/main/testdata/router)
(standard) and
[`testdata/gin`](https://github.com/yaop-labs/apiary/tree/main/testdata/gin)
(gin: a nested module with the real gin dependency, shown in full on the
[Examples](examples) page).
