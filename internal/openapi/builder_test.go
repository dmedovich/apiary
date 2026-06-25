package openapi_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/honeynil/apiary/internal/annotation"
	"github.com/honeynil/apiary/internal/openapi"
	"github.com/honeynil/apiary/internal/parser"
)

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

func TestBuild_BasicOperation(t *testing.T) {
	ops := []*parser.OperationInfo{
		{
			Annotation: &annotation.Operation{
				Method:  "POST",
				Path:    "/api/v1/auth",
				Summary: "Authenticate",
				Tags:    []string{"auth"},
				Errors:  []annotation.ErrorSpec{{Code: 400}, {Code: 401}},
			},
			RequestType:  &parser.TypeRef{Name: "AuthRequest"},
			ResponseType: &parser.TypeRef{Name: "AuthResponse"},
		},
	}

	types := map[string]*parser.TypeInfo{
		"AuthRequest": {
			Name: "AuthRequest",
			Fields: []*parser.FieldInfo{
				{
					Name:     "Token",
					JSONName: "token",
					Type:     &parser.TypeRef{Name: "string"},
					Required: true,
					Doc:      "Bearer token",
				},
			},
		},
		"AuthResponse": {
			Name: "AuthResponse",
			Fields: []*parser.FieldInfo{
				{
					Name:     "UserID",
					JSONName: "user_id",
					Type:     &parser.TypeRef{Name: "int64"},
				},
			},
		},
	}

	b := openapi.NewBuilder("Test API", "1.0.0")
	spec, err := b.Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	if spec.OpenAPI != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %s", spec.OpenAPI)
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("unexpected title: %s", spec.Info.Title)
	}

	item, ok := spec.Paths["/api/v1/auth"]
	if !ok {
		t.Fatal("path /api/v1/auth not found")
	}
	if item.Post == nil {
		t.Fatal("POST operation missing")
	}
	if item.Post.Summary != "Authenticate" {
		t.Errorf("unexpected summary: %s", item.Post.Summary)
	}
	if len(item.Post.Tags) != 1 || item.Post.Tags[0] != "auth" {
		t.Errorf("unexpected tags: %v", item.Post.Tags)
	}

	// 200 + 400 + 401
	if len(item.Post.Responses) != 3 {
		t.Errorf("expected 3 responses, got %d", len(item.Post.Responses))
	}
	if _, ok := item.Post.Responses["200"]; !ok {
		t.Error("missing 200 response")
	}
	if _, ok := item.Post.Responses["400"]; !ok {
		t.Error("missing 400 response")
	}

	if spec.Components == nil {
		t.Fatal("components is nil")
	}
	if _, ok := spec.Components.Schemas["AuthRequest"]; !ok {
		t.Error("AuthRequest not in components")
	}
	if _, ok := spec.Components.Schemas["AuthResponse"]; !ok {
		t.Error("AuthResponse not in components")
	}
	if _, ok := spec.Components.Schemas["ErrorResponse"]; !ok {
		t.Error("ErrorResponse not in components")
	}
}

func TestBuild_GetWithQueryParams(t *testing.T) {
	ops := []*parser.OperationInfo{
		{
			Annotation: &annotation.Operation{
				Method:  "GET",
				Path:    "/api/v1/users",
				Summary: "List users",
			},
			RequestType:  &parser.TypeRef{Name: "ListRequest"},
			ResponseType: &parser.TypeRef{Name: "ListResponse"},
		},
	}

	types := map[string]*parser.TypeInfo{
		"ListRequest": {
			Name: "ListRequest",
			Fields: []*parser.FieldInfo{
				{
					Name:       "Page",
					JSONName:   "page",
					QueryParam: "page",
					Type:       &parser.TypeRef{Name: "int"},
					Default:    "1",
				},
				{
					Name:       "Search",
					JSONName:   "search",
					QueryParam: "search",
					Type:       &parser.TypeRef{Name: "string"},
				},
			},
		},
		"ListResponse": {
			Name:   "ListResponse",
			Fields: []*parser.FieldInfo{},
		},
	}

	b := openapi.NewBuilder("API", "0.1.0")
	spec, err := b.Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	item := spec.Paths["/api/v1/users"]
	if item.Get == nil {
		t.Fatal("GET operation missing")
	}
	if item.Get.RequestBody != nil {
		t.Error("GET should not have a request body")
	}
	if len(item.Get.Parameters) != 2 {
		t.Errorf("expected 2 query parameters, got %d", len(item.Get.Parameters))
	}
	for _, p := range item.Get.Parameters {
		if p.In != "query" {
			t.Errorf("expected query parameter, got %s", p.In)
		}
	}
}

func TestBuild_QueryParamConstraints(t *testing.T) {
	ops := []*parser.OperationInfo{
		{
			Annotation:   &annotation.Operation{Method: "GET", Path: "/things"},
			RequestType:  &parser.TypeRef{Name: "ListReq"},
			ResponseType: &parser.TypeRef{Name: "Resp"},
		},
	}
	types := map[string]*parser.TypeInfo{
		"ListReq": {
			Name: "ListReq",
			Fields: []*parser.FieldInfo{
				{Name: "Page", JSONName: "page", QueryParam: "page", Type: &parser.TypeRef{Name: "int"}, Validate: "min=1,max=100", Default: "1"},
				{Name: "Sort", JSONName: "sort", QueryParam: "sort", Type: &parser.TypeRef{Name: "string"}, Validate: "oneof=asc desc"},
			},
		},
		"Resp": {Name: "Resp"},
	}

	spec, err := openapi.NewBuilder("API", "1.0.0").Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	params := spec.Paths["/things"].Get.Parameters
	byName := map[string]openapi.Parameter{}
	for _, p := range params {
		byName[p.Name] = p
	}
	page := byName["page"].Schema
	if page.Minimum != 1 || page.Maximum != 100 {
		t.Errorf("page bounds = %v..%v", page.Minimum, page.Maximum)
	}
	if page.Default != 1 {
		t.Errorf("page default = %v (want typed int 1)", page.Default)
	}
	if len(byName["sort"].Schema.Enum) != 2 {
		t.Errorf("sort enum = %v", byName["sort"].Schema.Enum)
	}
	// Description/example must stay on the Parameter, not duplicated onto schema.
	if page.Description != "" {
		t.Errorf("param schema must not carry description, got %q", page.Description)
	}
}

func TestBuild_PathParameters(t *testing.T) {
	ops := []*parser.OperationInfo{
		{
			Annotation: &annotation.Operation{
				Method:  "GET",
				Path:    "/api/v1/users/{id}",
				Summary: "Get user",
				Errors:  []annotation.ErrorSpec{{Code: 404}},
			},
			RequestType:  &parser.TypeRef{Name: "GetUserRequest"},
			ResponseType: &parser.TypeRef{Name: "UserResponse"},
		},
	}

	types := map[string]*parser.TypeInfo{
		"GetUserRequest": {
			Name: "GetUserRequest",
			Fields: []*parser.FieldInfo{
				{
					Name:      "ID",
					JSONName:  "id",
					PathParam: "id",
					Type:      &parser.TypeRef{Name: "int64"},
					Required:  true,
				},
			},
		},
		"UserResponse": {
			Name: "UserResponse",
			Fields: []*parser.FieldInfo{
				{Name: "ID", JSONName: "id", Type: &parser.TypeRef{Name: "int64"}},
			},
		},
	}

	b := openapi.NewBuilder("API", "0.1.0")
	spec, err := b.Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	item := spec.Paths["/api/v1/users/{id}"]
	if item.Get == nil {
		t.Fatal("GET operation missing")
	}
	if len(item.Get.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(item.Get.Parameters))
	}
	p := item.Get.Parameters[0]
	if p.In != "path" {
		t.Errorf("expected path parameter, got %s", p.In)
	}
	if p.Name != "id" {
		t.Errorf("expected parameter name 'id', got %q", p.Name)
	}
	if !p.Required {
		t.Error("path parameter must be required")
	}
}

func TestBuild_PathOnlyRequestHasNoBody(t *testing.T) {
	// DELETE /tasks/{id} bound to a request whose only field is a path param
	// must not produce an (empty) request body.
	ops := []*parser.OperationInfo{
		{
			Annotation:   &annotation.Operation{Method: "DELETE", Path: "/tasks/{id}"},
			RequestType:  &parser.TypeRef{Name: "DeleteTaskRequest"},
			ResponseType: nil,
		},
	}
	types := map[string]*parser.TypeInfo{
		"DeleteTaskRequest": {
			Name: "DeleteTaskRequest",
			Fields: []*parser.FieldInfo{
				{Name: "ID", JSONName: "id", PathParam: "id", Type: &parser.TypeRef{Name: "int64"}, Required: true},
			},
		},
	}

	spec, err := openapi.NewBuilder("API", "1.0.0").Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	del := spec.Paths["/tasks/{id}"].Delete
	if del == nil {
		t.Fatal("DELETE operation missing")
	}
	if del.RequestBody != nil {
		t.Errorf("path-only request must have no body, got %+v", del.RequestBody)
	}
	if len(del.Parameters) != 1 || del.Parameters[0].In != "path" {
		t.Errorf("expected a single path parameter, got %+v", del.Parameters)
	}
	// The empty placeholder schema must not be registered either.
	if spec.Components != nil {
		if _, ok := spec.Components.Schemas["DeleteTaskRequest"]; ok {
			t.Error("path-only request type should not be registered as a component schema")
		}
	}
}

func TestBuild_UnsupportedMethodWarnsAndSkips(t *testing.T) {
	ops := []*parser.OperationInfo{
		{
			Annotation:   &annotation.Operation{Method: "FETCH", Path: "/x"},
			ResponseType: &parser.TypeRef{Name: "R"},
		},
	}
	types := map[string]*parser.TypeInfo{"R": {Name: "R"}}

	var spec *openapi.OpenAPI
	out := captureLog(t, func() {
		var err error
		spec, err = openapi.NewBuilder("API", "1.0.0").Build(ops, types)
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}
	})
	if _, ok := spec.Paths["/x"]; ok {
		t.Error("expected unsupported-method operation to be skipped")
	}
	if !strings.Contains(out, "unsupported HTTP method") || !strings.Contains(out, "FETCH") {
		t.Errorf("expected unsupported-method warning, got: %q", out)
	}
}

func TestBuild_DuplicateOperationWarns(t *testing.T) {
	ops := []*parser.OperationInfo{
		{Annotation: &annotation.Operation{Method: "GET", Path: "/dup", Summary: "first"}, ResponseType: &parser.TypeRef{Name: "R"}},
		{Annotation: &annotation.Operation{Method: "GET", Path: "/dup", Summary: "second"}, ResponseType: &parser.TypeRef{Name: "R"}},
	}
	types := map[string]*parser.TypeInfo{"R": {Name: "R"}}

	var spec *openapi.OpenAPI
	out := captureLog(t, func() {
		var err error
		spec, err = openapi.NewBuilder("API", "1.0.0").Build(ops, types)
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}
	})
	if !strings.Contains(out, "duplicate operation") {
		t.Errorf("expected duplicate-operation warning, got: %q", out)
	}
	// Last definition wins.
	if spec.Paths["/dup"].Get == nil || spec.Paths["/dup"].Get.Summary != "second" {
		t.Errorf("expected the second definition to win, got %+v", spec.Paths["/dup"].Get)
	}
}

func TestBuild_Servers(t *testing.T) {
	ops := []*parser.OperationInfo{
		{Annotation: &annotation.Operation{Method: "GET", Path: "/x"}, ResponseType: &parser.TypeRef{Name: "R"}},
	}
	types := map[string]*parser.TypeInfo{"R": {Name: "R"}}

	spec, err := openapi.NewBuilder("API", "1.0.0").
		WithServer("https://api.example.com").
		WithServer("").
		Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(spec.Servers) != 1 {
		t.Fatalf("expected 1 server (empty ignored), got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != "https://api.example.com" {
		t.Errorf("unexpected server url: %q", spec.Servers[0].URL)
	}
}

func TestBuild_MultipleOperations(t *testing.T) {
	ops := []*parser.OperationInfo{
		{
			Annotation:   &annotation.Operation{Method: "GET", Path: "/foo"},
			RequestType:  &parser.TypeRef{Name: "FooReq"},
			ResponseType: &parser.TypeRef{Name: "FooResp"},
		},
		{
			Annotation:   &annotation.Operation{Method: "POST", Path: "/bar"},
			RequestType:  &parser.TypeRef{Name: "BarReq"},
			ResponseType: &parser.TypeRef{Name: "BarResp"},
		},
	}

	types := make(map[string]*parser.TypeInfo)
	for _, name := range []string{"FooReq", "FooResp", "BarReq", "BarResp"} {
		types[name] = &parser.TypeInfo{Name: name}
	}

	spec, err := openapi.NewBuilder("Multi", "1.0.0").Build(ops, types)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(spec.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(spec.Paths))
	}
}
