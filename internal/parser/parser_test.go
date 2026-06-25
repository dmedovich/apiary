package parser_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaop-labs/apiary/internal/parser"
)

func loadSrc(t *testing.T, src string) *parser.Parser {
	t.Helper()
	return loadFiles(t, map[string]string{"code.go": src})
}

func loadFiles(t *testing.T, files map[string]string) *parser.Parser {
	t.Helper()
	dir := t.TempDir()
	if _, ok := files["go.mod"]; !ok {
		files["go.mod"] = "module testmod\n\ngo 1.22\n"
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p := parser.New()
	if err := p.LoadDir(dir, "./..."); err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	return p
}

func captureLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := log.Writer()
	flags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(old)
		log.SetFlags(flags)
	}()
	fn()
	return buf.String()
}

const sampleCode = `package sample

import "context"

// apiary:operation POST /api/v1/hello
// summary: Hello world
// tags: test
// errors: 400,500
func (h *Handler) Hello(ctx context.Context, req HelloRequest) (HelloResponse, error) {
	return HelloResponse{}, nil
}

type Handler struct{}

type HelloRequest struct {
	Name string ` + "`" + `json:"name" validate:"required" doc:"Your name" example:"Alice"` + "`" + `
}

type HelloResponse struct {
	Message string ` + "`" + `json:"message" doc:"Greeting message"` + "`" + `
}
`

func TestLoad_HappyPath(t *testing.T) {
	p := loadSrc(t, sampleCode)

	ops := p.Operations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	op := ops[0]
	if op.Annotation.Method != "POST" {
		t.Errorf("expected POST, got %s", op.Annotation.Method)
	}
	if op.Annotation.Path != "/api/v1/hello" {
		t.Errorf("unexpected path: %s", op.Annotation.Path)
	}
	if op.RequestType == nil || op.RequestType.Name != "HelloRequest" {
		t.Errorf("expected HelloRequest, got %v", op.RequestType)
	}
	if op.ResponseType == nil || op.ResponseType.Name != "HelloResponse" {
		t.Errorf("expected HelloResponse, got %v", op.ResponseType)
	}

	req, ok := p.Types()["HelloRequest"]
	if !ok {
		t.Fatal("HelloRequest not found in types")
	}
	if len(req.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(req.Fields))
	}
	f := req.Fields[0]
	if f.JSONName != "name" || !f.Required || f.Doc != "Your name" || f.Example != "Alice" {
		t.Errorf("unexpected field: %+v", f)
	}
}

func TestLoad_SkipsInvalidSignature(t *testing.T) {
	p := loadSrc(t, `package sample

// apiary:operation GET /api/v1/bad
// summary: Bad signature
func badFunc() error { return nil }
`)
	if len(p.Operations()) != 0 {
		t.Fatalf("expected 0 operations, got %d", len(p.Operations()))
	}
}

func TestLoad_BadSignatureWarns(t *testing.T) {
	var p *parser.Parser
	out := captureLog(t, func() {
		p = loadSrc(t, `package sample

// apiary:operation GET /api/v1/bad
// summary: Bad signature
func badFunc() error { return nil }
`)
	})
	if len(p.Operations()) != 0 {
		t.Fatalf("expected 0 operations, got %d", len(p.Operations()))
	}
	if !strings.Contains(out, "badFunc") || !strings.Contains(out, "unsupported signature") {
		t.Errorf("expected bad-signature warning, got: %q", out)
	}
}

func TestLoad_NoWarningForUnannotatedFunc(t *testing.T) {
	var p *parser.Parser
	out := captureLog(t, func() {
		p = loadSrc(t, `package sample

func helper() error { return nil }
`)
	})
	_ = p
	if out != "" {
		t.Errorf("expected no warnings, got: %q", out)
	}
}

func TestLoad_GodocSummary(t *testing.T) {
	p := loadSrc(t, `package sample

import "context"

// CreateThing creates a thing.
// It does extra work too.
// apiary:operation POST /things
func (h *Handler) CreateThing(ctx context.Context, req ThingRequest) (ThingResponse, error) {
	return ThingResponse{}, nil
}

type Handler struct{}
type ThingRequest struct{ Name string `+"`"+`json:"name"`+"`"+` }
type ThingResponse struct{ ID int64 `+"`"+`json:"id"`+"`"+` }
`)
	ann := p.Operations()[0].Annotation
	if ann.Summary != "Creates a thing." {
		t.Errorf("summary = %q", ann.Summary)
	}
	if ann.Description != "It does extra work too." {
		t.Errorf("description = %q", ann.Description)
	}
	if ann.OperationID != "createThing" {
		t.Errorf("operationId = %q", ann.OperationID)
	}
}

func TestLoad_ExplicitSummaryBeatsGodoc(t *testing.T) {
	p := loadSrc(t, `package sample

import "context"

// CreateThing creates a thing.
// apiary:operation POST /things
// summary: Explicit summary
func (h *Handler) CreateThing(ctx context.Context) (ThingResponse, error) {
	return ThingResponse{}, nil
}

type Handler struct{}
type ThingResponse struct{ ID int64 `+"`"+`json:"id"`+"`"+` }
`)
	if got := p.Operations()[0].Annotation.Summary; got != "Explicit summary" {
		t.Errorf("explicit summary must win, got %q", got)
	}
}

func TestLoad_NoCtxSignature(t *testing.T) {
	p := loadSrc(t, `package sample

// apiary:operation GET /health
// summary: Health check
// security: none
func (h *Handler) Health(req HealthRequest) (HealthResponse, error) {
	return HealthResponse{}, nil
}

type Handler struct{}
type HealthRequest struct{}
type HealthResponse struct{ Status string `+"`"+`json:"status"`+"`"+` }
`)
	ops := p.Operations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0].RequestType == nil || ops[0].RequestType.Name != "HealthRequest" {
		t.Errorf("expected HealthRequest, got %v", ops[0].RequestType)
	}
}

func TestLoad_NoParamsSignature(t *testing.T) {
	p := loadSrc(t, `package sample

// apiary:operation GET /ping
// summary: Ping
func (h *Handler) Ping() (PingResponse, error) { return PingResponse{}, nil }

type Handler struct{}
type PingResponse struct{ OK bool `+"`"+`json:"ok"`+"`"+` }
`)
	ops := p.Operations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0].RequestType != nil {
		t.Errorf("expected nil RequestType, got %v", ops[0].RequestType)
	}
	if ops[0].ResponseType == nil || ops[0].ResponseType.Name != "PingResponse" {
		t.Errorf("expected PingResponse, got %v", ops[0].ResponseType)
	}
}

func TestLoad_HeaderAndQueryTags(t *testing.T) {
	p := loadSrc(t, `package sample

import "context"

// apiary:operation GET /api/v1/products
// summary: List products
func (h *Handler) ListProducts(ctx context.Context, req ProductsRequest) (struct{}, error) {
	return struct{}{}, nil
}

type Handler struct{}
type ProductsRequest struct {
	Currency string `+"`"+`header:"X-Currency" doc:"ISO 4217 currency"`+"`"+`
	Page     int    `+"`"+`query:"page" default:"1"`+"`"+`
}
`)
	req, ok := p.Types()["ProductsRequest"]
	if !ok {
		t.Fatal("ProductsRequest not found")
	}
	var header, query *parser.FieldInfo
	for _, f := range req.Fields {
		if f.Header != "" {
			header = f
		}
		if f.QueryParam != "" {
			query = f
		}
	}
	if header == nil || header.Header != "X-Currency" {
		t.Errorf("expected X-Currency header field, got %+v", header)
	}
	if query == nil || query.QueryParam != "page" {
		t.Errorf("expected page query field, got %+v", query)
	}
}

func TestLoad_GinHandler(t *testing.T) {
	p := loadFiles(t, map[string]string{
		"gin/gin.go": "package gin\n\ntype Context struct{}\n",
		"code.go": `package sample

import "testmod/gin"

// apiary:operation GET /api/v1/users
// summary: List users
// request: ListUsersRequest
// response: []UserDTO
func ListUsers(c *gin.Context) {}

type ListUsersRequest struct{ Page int ` + "`" + `query:"page"` + "`" + ` }
type UserDTO struct {
	ID   int64  ` + "`" + `json:"id"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
}
`,
	})
	ops := p.Operations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	op := ops[0]
	if op.Annotation.Method != "GET" {
		t.Errorf("expected GET, got %s", op.Annotation.Method)
	}
	if op.RequestType == nil || op.RequestType.Name != "ListUsersRequest" {
		t.Errorf("expected ListUsersRequest, got %v", op.RequestType)
	}
	if op.ResponseType == nil || !op.ResponseType.IsSlice {
		t.Fatalf("expected slice response, got %v", op.ResponseType)
	}
	if op.ResponseType.Elem == nil || op.ResponseType.Elem.Name != "UserDTO" {
		t.Errorf("expected UserDTO elem, got %v", op.ResponseType.Elem)
	}
}

func TestLoad_StdlibHTTPHandler(t *testing.T) {
	for name, code := range map[string]string{
		"free_func": `package sample
import "net/http"
// apiary:operation POST /api/v1/auth/login
// summary: Login
// request: LoginRequest
// response: LoginResponse
func Login(w http.ResponseWriter, r *http.Request) {}
type LoginRequest struct{ User string ` + "`json:\"user\"`" + ` }
type LoginResponse struct{ Token string ` + "`json:\"token\"`" + ` }
`,
		"method": `package sample
import "net/http"
type AuthHandler struct{}
// apiary:operation POST /api/v1/auth/login
// summary: Login
// request: LoginRequest
// response: LoginResponse
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {}
type LoginRequest struct{ User string ` + "`json:\"user\"`" + ` }
type LoginResponse struct{ Token string ` + "`json:\"token\"`" + ` }
`,
	} {
		t.Run(name, func(t *testing.T) {
			p := loadSrc(t, code)
			ops := p.Operations()
			if len(ops) != 1 {
				t.Fatalf("expected 1 operation, got %d", len(ops))
			}
			if ops[0].RequestType == nil || ops[0].RequestType.Name != "LoginRequest" {
				t.Errorf("expected LoginRequest, got %v", ops[0].RequestType)
			}
			if ops[0].ResponseType == nil || ops[0].ResponseType.Name != "LoginResponse" {
				t.Errorf("expected LoginResponse, got %v", ops[0].ResponseType)
			}
		})
	}
}

func TestLoad_StdlibHTTPHandler_WrongShapes(t *testing.T) {
	cases := map[string]string{
		"returns_error": `package sample
import "net/http"
// apiary:operation GET /a
// summary: x
// response: R
func F(w http.ResponseWriter, r *http.Request) error { return nil }
type R struct{}
`,
		"one_param": `package sample
import "net/http"
// apiary:operation GET /a
// summary: x
// response: R
func F(w http.ResponseWriter) {}
type R struct{}
`,
	}
	for name, code := range cases {
		t.Run(name, func(t *testing.T) {
			var p *parser.Parser
			_ = captureLog(t, func() { p = loadSrc(t, code) })
			if ops := p.Operations(); len(ops) != 0 {
				t.Errorf("expected 0 operations, got %d", len(ops))
			}
		})
	}
}

func TestLoad_SliceResponse(t *testing.T) {
	p := loadSrc(t, `package sample

import "context"

// apiary:operation GET /api/v1/items
// summary: List items
func (h *Handler) ListItems(ctx context.Context) ([]ItemDTO, error) { return nil, nil }

type Handler struct{}
type ItemDTO struct{ ID int64 `+"`"+`json:"id"`+"`"+` }
`)
	resp := p.Operations()[0].ResponseType
	if resp == nil || !resp.IsSlice {
		t.Fatalf("expected slice ResponseType, got %v", resp)
	}
	if resp.Elem == nil || resp.Elem.Name != "ItemDTO" {
		t.Errorf("expected ItemDTO elem, got %v", resp.Elem)
	}
}

func TestLoad_EnumConsts(t *testing.T) {
	p := loadSrc(t, `package sample

type Status string

const (
	StatusActive  Status = "active"
	StatusPending Status = "pending"
	StatusClosed  Status = "closed"
)

// apiary:operation GET /items
// summary: get
func Get() (Item, error) { return Item{}, nil }

type Item struct {
	Name   string `+"`"+`json:"name"`+"`"+`
	Status Status `+"`"+`json:"status"`+"`"+`
}
`)
	info, ok := p.Enums()["Status"]
	if !ok {
		t.Fatal("Status enum not found")
	}
	if info.BaseType != "string" {
		t.Errorf("expected base type string, got %q", info.BaseType)
	}
	if len(info.Values) != 3 || info.Values[0] != "active" || info.Values[2] != "closed" {
		t.Errorf("unexpected enum values: %v", info.Values)
	}
}

func TestLoad_EnumIota(t *testing.T) {
	p := loadSrc(t, `package sample

type Role int

const (
	RoleAdmin Role = iota
	RoleUser
	RoleModerator
)

// apiary:operation GET /me
// summary: me
func Me() (U, error) { return U{}, nil }

type U struct{ Role Role `+"`"+`json:"role"`+"`"+` }
`)
	info, ok := p.Enums()["Role"]
	if !ok {
		t.Fatal("Role enum not found")
	}
	if info.BaseType != "int" {
		t.Errorf("expected base type int, got %q", info.BaseType)
	}
	if len(info.Values) != 3 || info.Values[0] != 0 || info.Values[1] != 1 || info.Values[2] != 2 {
		t.Errorf("unexpected iota values: %v", info.Values)
	}
}

func TestLoad_MultipleFiles(t *testing.T) {
	p := loadFiles(t, map[string]string{
		"a.go": sampleCode,
		"b.go": `package sample

import "context"

// apiary:operation GET /api/v1/health
// summary: Health check
func (h *Handler) Health(ctx context.Context, req HealthRequest) (HealthResponse, error) {
	return HealthResponse{}, nil
}

type HealthRequest struct{}
type HealthResponse struct{ Status string ` + "`json:\"status\"`" + ` }
`,
	})
	if len(p.Operations()) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(p.Operations()))
	}
}

func TestLoad_CrossPackageType(t *testing.T) {
	p := loadFiles(t, map[string]string{
		"dto/dto.go": `package dto

type User struct {
	ID   int64  ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`,
		"code.go": `package sample

import (
	"context"
	"testmod/dto"
)

// apiary:operation POST /users
// summary: Create
func Create(ctx context.Context, req dto.User) (dto.User, error) { return dto.User{}, nil }
`,
	})
	ops := p.Operations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0].RequestType == nil || ops[0].RequestType.Name != "User" {
		t.Fatalf("expected User request, got %v", ops[0].RequestType)
	}
	user, ok := p.Types()["User"]
	if !ok {
		t.Fatal("cross-package User type not registered as a component")
	}
	if len(user.Fields) != 2 {
		t.Errorf("expected 2 fields on User, got %d", len(user.Fields))
	}
}
