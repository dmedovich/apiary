package parser_test

import (
	"go/version"
	"reflect"
	"runtime"
	"testing"

	"github.com/yaop-labs/apiary/internal/openapi"
	"github.com/yaop-labs/apiary/internal/schema"
)

func TestAliasResponsesKeepRefsAndComponents(t *testing.T) {
	// Reproduce the v1.2.2 regression: Alias nodes fell through to
	// interface{}, producing schema: {} or items: {} and losing components.
	t.Setenv("GODEBUG", "gotypesalias=1")
	p := loadFiles(t, map[string]string{
		"dto/types.go": `package dto
type Response struct { ID string }
type Alias = Response
`,
		"handlers.go": `package sample
import "testmod/dto"
// apiary:operation GET /one
func One() (dto.Alias, error) { return dto.Alias{}, nil }
// apiary:operation GET /many
func Many() ([]dto.Alias, error) { return nil, nil }
`,
	})
	if len(p.Operations()) != 2 {
		t.Fatalf("operations = %d, want 2", len(p.Operations()))
	}
	spec, err := openapi.NewBuilder("Alias regression", "1").Build(p.Operations(), p.Types())
	if err != nil {
		t.Fatal(err)
	}
	const wantRef = "#/components/schemas/Response"
	one := spec.Paths["/one"].Get.Responses["200"].Content["application/json"].Schema
	if one.Ref != wantRef {
		t.Errorf("alias response $ref = %q, want %q", one.Ref, wantRef)
	}
	many := spec.Paths["/many"].Get.Responses["200"].Content["application/json"].Schema
	if many.Type != "array" || many.Items == nil || many.Items.Ref != wantRef {
		t.Errorf("[]Alias must retain array items.$ref: %+v", many)
	}
	response := spec.Components.Schemas["Response"]
	if response == nil {
		t.Fatal("alias target is missing from components.schemas")
	}
	if response.Type != "object" || response.Properties["id"] == nil || response.Properties["id"].Type != "string" {
		t.Fatalf("alias target must retain its object fields: %+v", response)
	}
	// Both handlers reference only aliases, so target registration cannot be
	// accidentally supplied by a separate operation using Response directly.
	if len(spec.Components.Schemas) != 2 {
		t.Fatalf("schema count = %d, want Response and ErrorResponse", len(spec.Components.Schemas))
	}
}

func TestTypeAliases(t *testing.T) {
	// Check both representations on older toolchains. Go 1.27 always uses
	// Alias nodes and rejects gotypesalias=0 when starting the go command.
	for _, mode := range []string{"0", "1"} {
		t.Run("gotypesalias="+mode, func(t *testing.T) {
			if mode == "0" && version.Compare(version.Lang(runtime.Version()), "go1.27") >= 0 {
				t.Skip("Go 1.27 removed the legacy alias representation")
			}
			t.Setenv("GODEBUG", "gotypesalias="+mode)
			testTypeAliases(t)
		})
	}
}

func testTypeAliases(t *testing.T) {
	t.Helper()
	p := loadFiles(t, map[string]string{
		"dto/types.go": `package dto
import "time"
type Item struct { ID string }
type Alias = Item
type Chain = Alias
type Pointer = *Chain
type Items = []Chain
type Lookup = map[string]Pointer
type Fixed = [2]Chain
type Text = string
type Timestamp = time.Time
type State string
const Ready State = "ready"
type Status = State
type Envelope struct {
 Pointer
 Item Chain
 Items Items
 Lookup Lookup
 Fixed Fixed
 Text Text
 Time Timestamp
 Status Status
}
type Request struct { Name Text }
type RequestAlias = Request
`,
		"handlers.go": `package sample
import (
 "net/http"
 d "testmod/dto"
)
type Local = d.Chain
// apiary:operation POST /items
func Create(req d.RequestAlias) (Local, error) { return Local{}, nil }
// apiary:operation GET /items
func List() (d.Items, error) { return nil, nil }
// apiary:operation GET /envelope
func Get() (d.Envelope, error) { return d.Envelope{}, nil }
// apiary:operation POST /annotated
// request: d.RequestAlias
// response: []d.Chain
func Annotated(w http.ResponseWriter, r *http.Request) {}
// apiary:operation GET /pointer
// response: d.Pointer
func Pointer(w http.ResponseWriter, r *http.Request) {}
`,
	})
	if len(p.Operations()) != 5 {
		t.Fatalf("operations = %d, want 5", len(p.Operations()))
	}
	spec, err := openapi.NewBuilder("Aliases", "1").Build(p.Operations(), p.Types(), p.Enums())
	if err != nil {
		t.Fatal(err)
	}
	itemRef := &schema.Schema{Ref: "#/components/schemas/Item"}
	assertSchema := func(name string, got, want *schema.Schema) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %+v, want %+v", name, got, want)
		}
	}
	assertSchema("response alias", spec.Paths["/items"].Post.Responses["200"].Content["application/json"].Schema, itemRef)
	array := &schema.Schema{Type: "array", Items: itemRef}
	assertSchema("slice alias", spec.Paths["/items"].Get.Responses["200"].Content["application/json"].Schema, array)
	assertSchema("annotation alias", spec.Paths["/annotated"].Post.Responses["200"].Content["application/json"].Schema, array)
	pointer := &schema.Schema{AnyOf: []*schema.Schema{itemRef, {Type: "null"}}}
	assertSchema("pointer annotation alias", spec.Paths["/pointer"].Get.Responses["200"].Content["application/json"].Schema, pointer)
	for _, path := range []string{"/items", "/annotated"} {
		op := spec.Paths[path].Post
		if op.RequestBody == nil {
			t.Errorf("%s: request alias has no request body", path)
			continue
		}
		assertSchema("request alias", op.RequestBody.Content["application/json"].Schema, &schema.Schema{Ref: "#/components/schemas/Request"})
	}
	if len(spec.Components.Schemas) != 4 {
		t.Fatalf("components = %v, want Item, Request, Envelope, ErrorResponse", spec.Components.Schemas)
	}
	item := spec.Components.Schemas["Item"]
	if item == nil || item.Properties["id"] == nil || item.Properties["id"].Type != "string" {
		t.Fatal("alias target component must retain its fields")
	}
	envelope := spec.Components.Schemas["Envelope"]
	if envelope == nil || len(envelope.AllOf) != 2 {
		t.Fatalf("embedded alias must produce allOf: %+v", envelope)
	}
	assertSchema("embedded alias", envelope.AllOf[0], itemRef)
	props := envelope.AllOf[1].Properties
	for name, want := range map[string]*schema.Schema{
		"item":   itemRef,
		"items":  array,
		"fixed":  array,
		"lookup": {Type: "object", AdditionalProperties: pointer},
		"text":   {Type: "string"},
		"time":   {Type: "string", Format: "date-time"},
		"status": {Type: "string", Enum: []any{"ready"}},
	} {
		assertSchema("field "+name, props[name], want)
	}
}
