package parser

import (
	"reflect"
	"strings"
)

// TypeRef is a structured representation of a Go type, produced by lowering a
// go/types Type (see resolve.go).
type TypeRef struct {
	Name    string   // base type name (e.g. "string", "UserDTO", "time.Time")
	IsPtr   bool     // *T
	IsSlice bool     // []T
	IsMap   bool     // map[K]V
	MapKey  string   // key type name for maps
	Elem    *TypeRef // element type for slices and maps
}

// FieldInfo describes a single struct field relevant to API schema generation.
type FieldInfo struct {
	Name       string
	Type       *TypeRef
	JSONName   string
	Doc        string
	Example    string
	Default    string
	Required   bool
	Validate   string // raw `validate:"..."` tag, mapped to JSON-Schema constraints
	PathParam  string // non-empty when field has path:"name" tag
	QueryParam string // non-empty when field has query:"name" tag
	Header     string // non-empty when field has header:"name" tag
}

// TypeInfo describes a parsed struct type.
type TypeInfo struct {
	Name     string
	Fields   []*FieldInfo
	Embedded []string // component names of embedded (anonymous) structs, for allOf
}

// EnumInfo describes a named Go type with a set of const values.
type EnumInfo struct {
	BaseType string // underlying Go type (e.g. "string", "int")
	Values   []any  // const values in declaration order (string or int)
}

type fieldTags struct {
	json     string
	doc      string
	example  string
	defaultV string
	validate string
	path     string
	query    string
	header   string
}

func parseStructTag(raw string) fieldTags {
	st := reflect.StructTag(raw)
	tags := fieldTags{}

	jsonTag := st.Get("json")
	if jsonTag != "" {
		// Only the field-name portion before the first comma matters for the
		// schema; option flags like ",omitempty" are not used.
		tags.json, _, _ = strings.Cut(jsonTag, ",")
	}

	tags.doc = st.Get("doc")
	tags.example = st.Get("example")
	tags.defaultV = st.Get("default")
	tags.validate = st.Get("validate")
	tags.path = st.Get("path")
	tags.query = st.Get("query")
	tags.header = st.Get("header")
	return tags
}

// hasValidator reports whether the comma-separated go-playground/validator tag
// contains the given validator by name (matching the part before any '=', so
// "required" does not match "required_if=Foo").
func hasValidator(tag, name string) bool {
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if n, _, _ := strings.Cut(part, "="); n == name {
			return true
		}
	}
	return false
}

// goNameToJSON converts a Go field name to a JSON key using the same heuristic
// as encoding/json: lowercase the first letter. Additionally, pure-uppercase
// acronyms (ID, URL, UUID) are fully lowercased so "ID" → "id", not "iD".
func goNameToJSON(name string) string {
	if name == "" {
		return name
	}
	allUpper := true
	for _, r := range name {
		if r < 'A' || r > 'Z' {
			allUpper = false
			break
		}
	}
	if allUpper {
		return strings.ToLower(name)
	}
	return strings.ToLower(name[:1]) + name[1:]
}
