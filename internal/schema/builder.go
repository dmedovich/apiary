// Package schema converts Go type information into JSON Schema objects
// suitable for embedding in an OpenAPI 3.1 document.
package schema

import (
	"strconv"
	"strings"

	"github.com/honeynil/apiary/internal/parser"
	"gopkg.in/yaml.v3"
)

// Schema is a subset of JSON Schema Draft 2020-12 used by OpenAPI 3.1.
type Schema struct {
	Ref         string    `yaml:"$ref,omitempty"`
	AllOf       []*Schema `yaml:"allOf,omitempty"`
	AnyOf       []*Schema `yaml:"anyOf,omitempty"`
	Type        string    `yaml:"type,omitempty"`
	Format      string    `yaml:"format,omitempty"`
	Description string    `yaml:"description,omitempty"`
	Example     any       `yaml:"example,omitempty"`
	Default     any       `yaml:"default,omitempty"`
	Enum        []any     `yaml:"enum,omitempty"`
	// Validation constraints (derived from `validate:"..."` struct tags).
	Minimum          any    `yaml:"minimum,omitempty"`
	Maximum          any    `yaml:"maximum,omitempty"`
	ExclusiveMinimum any    `yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum any    `yaml:"exclusiveMaximum,omitempty"`
	MinLength        *int   `yaml:"minLength,omitempty"`
	MaxLength        *int   `yaml:"maxLength,omitempty"`
	MinItems         *int   `yaml:"minItems,omitempty"`
	MaxItems         *int   `yaml:"maxItems,omitempty"`
	Pattern          string `yaml:"pattern,omitempty"`

	Properties           map[string]*Schema `yaml:"properties,omitempty"`
	AdditionalProperties *Schema            `yaml:"additionalProperties,omitempty"`
	Items                *Schema            `yaml:"items,omitempty"`
	Required             []string           `yaml:"required,omitempty"`

	// Nullable marks a scalar schema as accepting null. It is not emitted as a
	// field; instead MarshalYAML rewrites `type` to the 3.1 form [type, "null"].
	Nullable bool `yaml:"-"`
}

// MarshalYAML renders the schema. For a non-nullable schema this is identical to
// default struct marshaling (field order preserved). For a nullable scalar
// schema it rewrites the `type` scalar into the OpenAPI 3.1 sequence form,
// e.g. `type: [string, "null"]`, leaving every other field untouched.
func (s Schema) MarshalYAML() (any, error) {
	type raw Schema // shed MarshalYAML to avoid infinite recursion

	// Nullable scalar → rewrite `type` to the 3.1 sequence form [T, "null"].
	if s.Nullable && s.Type != "" {
		var node yaml.Node
		if err := node.Encode(raw(s)); err != nil {
			return nil, err
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == "type" {
				val := node.Content[i+1]
				node.Content[i+1] = &yaml.Node{
					Kind:  yaml.SequenceNode,
					Style: yaml.FlowStyle,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: val.Value},
						{Kind: yaml.ScalarNode, Value: "null", Style: yaml.DoubleQuotedStyle},
					},
				}
				break
			}
		}
		return &node, nil
	}

	// A standalone null schema must keep `type` quoted (type: "null"), otherwise
	// YAML parses it as the null value rather than the string token.
	if s.Type == "null" {
		var node yaml.Node
		if err := node.Encode(raw(s)); err != nil {
			return nil, err
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == "type" {
				node.Content[i+1].Style = yaml.DoubleQuotedStyle
				break
			}
		}
		return &node, nil
	}

	return raw(s), nil
}

// nullSchema is the `{type: "null"}` branch used to make a $ref nullable.
func nullSchema() *Schema { return &Schema{Type: "null"} }

// CoerceScalar converts a raw struct-tag value (always a string) to a typed
// value matching the schema type, so e.g. example:"42" on an integer field
// renders as 42, not "42". Empty input yields nil; unparseable input stays a
// string.
func CoerceScalar(raw, schemaType string) any {
	if raw == "" {
		return nil
	}
	switch schemaType {
	case "integer":
		if n, err := strconv.Atoi(raw); err == nil {
			return n
		}
	case "number":
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return f
		}
	case "boolean":
		if b, err := strconv.ParseBool(raw); err == nil {
			return b
		}
	}
	return raw
}

// Builder converts parser.TypeInfo values into JSON Schema objects and tracks
// which schemas have been placed in the components/schemas section.
type Builder struct {
	types        map[string]*parser.TypeInfo
	enums        map[string]*parser.EnumInfo
	components   map[string]*Schema
	processing   map[string]bool // guards against recursive types
	unknownTypes []string        // types not found in the parsed set
}

// NewBuilder creates a Builder that can resolve the provided types.
func NewBuilder(types map[string]*parser.TypeInfo, enums map[string]*parser.EnumInfo) *Builder {
	if enums == nil {
		enums = make(map[string]*parser.EnumInfo)
	}
	return &Builder{
		types:      types,
		enums:      enums,
		components: make(map[string]*Schema),
		processing: make(map[string]bool),
	}
}

// Components returns the map of schemas that will become components/schemas.
func (b *Builder) Components() map[string]*Schema {
	return b.components
}

// UnknownTypes returns the names of types that were referenced but not found
// in the parsed source. Callers can use this to warn the user that they may
// need to broaden their scan pattern (e.g. ./... instead of ./internal/handler/...).
func (b *Builder) UnknownTypes() []string {
	return b.unknownTypes
}

// BuildSchema returns a JSON Schema for the given TypeRef. Struct types are
// registered in components and returned as a $ref.
func (b *Builder) BuildSchema(ref *parser.TypeRef) *Schema {
	if ref == nil {
		return &Schema{Type: "object"}
	}
	if ref.IsSlice {
		return &Schema{Type: "array", Items: b.BuildSchema(ref.Elem)}
	}
	if ref.IsMap {
		return &Schema{Type: "object", AdditionalProperties: b.BuildSchema(ref.Elem)}
	}
	if s := primitiveSchema(ref.Name); s != nil {
		return s
	}
	// Named enum type — inline the base type schema (with enum values added in buildFieldSchema).
	if enumInfo, ok := b.enums[ref.Name]; ok {
		if s := primitiveSchema(enumInfo.BaseType); s != nil {
			return s
		}
	}
	// Struct type — register in components, return $ref.
	b.ensureComponent(ref.Name)
	return &Schema{Ref: "#/components/schemas/" + ref.Name}
}

// BuildSchemaByName is like BuildSchema but accepts a bare type name string.
func (b *Builder) BuildSchemaByName(name string) *Schema {
	if name == "" {
		return &Schema{Type: "object"}
	}
	if s := primitiveSchema(name); s != nil {
		return s
	}
	b.ensureComponent(name)
	return &Schema{Ref: "#/components/schemas/" + name}
}

// EnsureErrorResponse registers the standard error schema in components.
func (b *Builder) EnsureErrorResponse() {
	if _, ok := b.components["ErrorResponse"]; ok {
		return
	}
	b.components["ErrorResponse"] = &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"error": {Type: "string", Description: "Human-readable error message"},
		},
		Required: []string{"error"},
	}
}

// ensureComponent builds and registers the schema for typeName if it has not
// been registered yet. Recursive types are handled safely via the processing set.
func (b *Builder) ensureComponent(name string) {
	if _, exists := b.components[name]; exists {
		return
	}
	if b.processing[name] {
		// Recursive reference — the $ref will point to a schema that will be
		// completed by the outer call; no further action needed.
		return
	}

	typeInfo, exists := b.types[name]
	if !exists {
		// Unknown / external type (e.g. from another package not scanned).
		// Emit a placeholder and record the name so callers can warn the user.
		b.components[name] = &Schema{Type: "object"}
		b.unknownTypes = append(b.unknownTypes, name)
		return
	}

	b.processing[name] = true
	defer func() { delete(b.processing, name) }()

	// Handle embedded structs via allOf.
	var allOf []*Schema
	for _, embName := range typeInfo.Embedded {
		b.ensureComponent(embName)
		allOf = append(allOf, &Schema{Ref: "#/components/schemas/" + embName})
	}

	// Own fields → properties object.
	ownSchema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}
	var required []string
	for _, field := range typeInfo.Fields {
		// Path, query and header params are represented as OpenAPI parameters,
		// not as properties of the request body schema.
		if field.PathParam != "" || field.QueryParam != "" || field.Header != "" {
			continue
		}
		fieldSchema := b.buildFieldSchema(field)
		jsonName := field.JSONName
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name[:1]) + field.Name[1:]
		}
		ownSchema.Properties[jsonName] = fieldSchema
		if field.Required {
			required = append(required, jsonName)
		}
	}
	if len(required) > 0 {
		ownSchema.Required = required
	}

	// Register before returning so recursive refs can resolve.
	if len(allOf) > 0 {
		// Merge: embedded refs + own properties (only if non-empty).
		if len(ownSchema.Properties) > 0 {
			allOf = append(allOf, ownSchema)
		}
		b.components[name] = &Schema{AllOf: allOf}
	} else {
		b.components[name] = ownSchema
	}
}

func (b *Builder) buildFieldSchema(field *parser.FieldInfo) *Schema {
	// BuildSchema always returns a freshly allocated schema, so it is safe to
	// annotate it in place.
	s := b.BuildSchema(field.Type)
	// For a body property, description and example live on the schema itself.
	if field.Doc != "" {
		s.Description = field.Doc
	}
	if field.Example != "" {
		s.Example = CoerceScalar(field.Example, s.Type)
	}
	b.applyConstraints(s, field)
	return s
}

// BuildParamSchema builds the schema for an OpenAPI parameter. It carries the
// same validation constraints, enum, default and nullability as a body field,
// but not description/example — for parameters those belong on the Parameter
// object, not on its schema.
func (b *Builder) BuildParamSchema(field *parser.FieldInfo) *Schema {
	s := b.BuildSchema(field.Type)
	b.applyConstraints(s, field)
	return s
}

// applyConstraints annotates a freshly built schema with the constraints common
// to body fields and parameters: default value, enum (named enum types) and the
// JSON-Schema constraints derived from `validate:"..."`, plus pointer nullability.
func (b *Builder) applyConstraints(s *Schema, field *parser.FieldInfo) {
	if field.Default != "" {
		s.Default = CoerceScalar(field.Default, s.Type)
	}
	// A pointer (*T) shares the same underlying type name, so IsPtr does not
	// affect the enum lookup.
	if enumInfo := b.enums[field.Type.Name]; enumInfo != nil {
		s.Enum = enumInfo.Values
	}
	if field.Validate != "" {
		applyValidators(s, field.Validate)
	}
	if field.Type.IsPtr {
		switch {
		case isScalarType(s.Type):
			// Pointer to a scalar accepts null (3.1: type: [T, "null"]).
			s.Nullable = true
		case s.Ref != "":
			// Pointer to a struct: $ref can't carry `type`, so wrap it as
			// anyOf: [{$ref}, {type: "null"}], preserving any description/example.
			*s = Schema{
				Description: s.Description,
				Example:     s.Example,
				AnyOf:       []*Schema{{Ref: s.Ref}, nullSchema()},
			}
		}
	}
}

// isScalarType reports whether t is a primitive (non-composite) JSON Schema type.
func isScalarType(t string) bool {
	switch t {
	case "string", "integer", "number", "boolean":
		return true
	}
	return false
}

// applyValidators maps the common go-playground/validator rules in a
// `validate:"..."` tag onto JSON-Schema constraints. The interpretation of
// min/max depends on the already-resolved schema type (numeric value vs string
// length vs array length). Unknown validators are ignored.
func applyValidators(s *Schema, raw string) {
	numeric := s.Type == "integer" || s.Type == "number"
	for _, part := range strings.Split(raw, ",") {
		name, val, _ := strings.Cut(strings.TrimSpace(part), "=")
		switch name {
		case "email":
			s.Format = "email"
		case "uuid", "uuid3", "uuid4", "uuid5":
			s.Format = "uuid"
		case "uri", "url", "http_url":
			s.Format = "uri"
		case "ipv4", "ip4_addr":
			s.Format = "ipv4"
		case "ipv6", "ip6_addr":
			s.Format = "ipv6"
		case "hostname", "hostname_rfc1123":
			s.Format = "hostname"
		case "oneof":
			if e := parseEnumValues(val, s.Type); len(e) > 0 {
				s.Enum = e
			}
		case "min":
			setBound(s, val, true, numeric)
		case "max":
			setBound(s, val, false, numeric)
		case "len":
			setBound(s, val, true, numeric)
			setBound(s, val, false, numeric)
		case "gt":
			if n, ok := parseNumber(val); ok && numeric {
				s.ExclusiveMinimum = n
			}
		case "gte":
			if n, ok := parseNumber(val); ok && numeric {
				s.Minimum = n
			}
		case "lt":
			if n, ok := parseNumber(val); ok && numeric {
				s.ExclusiveMaximum = n
			}
		case "lte":
			if n, ok := parseNumber(val); ok && numeric {
				s.Maximum = n
			}
		}
	}
}

// setBound applies a min (isMin) or max bound, choosing minimum/maximum for
// numbers, minItems/maxItems for arrays, or minLength/maxLength otherwise.
func setBound(s *Schema, val string, isMin, numeric bool) {
	if numeric {
		if n, ok := parseNumber(val); ok {
			if isMin {
				s.Minimum = n
			} else {
				s.Maximum = n
			}
		}
		return
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	switch s.Type {
	case "array":
		if isMin {
			s.MinItems = &n
		} else {
			s.MaxItems = &n
		}
	default: // string and other scalar-ish types use length
		if isMin {
			s.MinLength = &n
		} else {
			s.MaxLength = &n
		}
	}
}

// parseNumber parses an integer when possible (so YAML renders 1, not 1.0),
// otherwise a float.
func parseNumber(s string) (any, bool) {
	if i, err := strconv.Atoi(s); err == nil {
		return i, true
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return nil, false
}

// parseEnumValues splits a space-separated `oneof` value list, parsing integers
// when the schema type is integer.
func parseEnumValues(val, schemaType string) []any {
	fields := strings.Fields(val)
	out := make([]any, 0, len(fields))
	for _, f := range fields {
		if schemaType == "integer" {
			if i, err := strconv.Atoi(f); err == nil {
				out = append(out, i)
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

// primitiveSchema maps Go primitive type names to their JSON Schema equivalents.
// Returns nil for non-primitive (struct) types.
func primitiveSchema(name string) *Schema {
	// Strip pointer sigil if somehow present in the name.
	name = strings.TrimPrefix(name, "*")

	switch name {
	case "string":
		return &Schema{Type: "string"}
	case "bool":
		return &Schema{Type: "boolean"}
	case "int", "int8", "int16", "int32",
		"uint", "uint8", "uint16", "uint32", "byte", "rune":
		return &Schema{Type: "integer", Format: "int32"}
	case "int64", "uint64":
		return &Schema{Type: "integer", Format: "int64"}
	case "float32":
		return &Schema{Type: "number", Format: "float"}
	case "float64":
		return &Schema{Type: "number", Format: "double"}
	case "interface{}", "any":
		return &Schema{}
	case "time.Time":
		return &Schema{Type: "string", Format: "date-time"}
	case "time.Duration":
		return &Schema{Type: "integer", Format: "int64"}
	case "uuid.UUID", "uuid.NullUUID":
		return &Schema{Type: "string", Format: "uuid"}
	case "net.IP":
		return &Schema{Type: "string", Format: "ipv4"}
	case "url.URL":
		return &Schema{Type: "string", Format: "uri"}
	case "json.RawMessage":
		return &Schema{} // any
	case "sql.NullString":
		return &Schema{Type: "string"}
	case "sql.NullInt32":
		return &Schema{Type: "integer", Format: "int32"}
	case "sql.NullInt64":
		return &Schema{Type: "integer", Format: "int64"}
	case "sql.NullFloat64":
		return &Schema{Type: "number", Format: "double"}
	case "sql.NullBool":
		return &Schema{Type: "boolean"}
	case "sql.NullTime":
		return &Schema{Type: "string", Format: "date-time"}
	}

	// Unknown package-qualified type — treat as string (best-effort fallback).
	if strings.Contains(name, ".") {
		return &Schema{Type: "string"}
	}

	return nil
}
