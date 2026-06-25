package schema

import (
	"strconv"
	"strings"

	"github.com/yaop-labs/apiary/internal/parser"
	"gopkg.in/yaml.v3"
)

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

	Nullable bool `yaml:"-"`
}

func (s Schema) MarshalYAML() (any, error) {
	type raw Schema

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

func nullSchema() *Schema { return &Schema{Type: "null"} }

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

type Builder struct {
	types        map[string]*parser.TypeInfo
	enums        map[string]*parser.EnumInfo
	components   map[string]*Schema
	processing   map[string]bool
	unknownTypes []string
}

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

func (b *Builder) Components() map[string]*Schema {
	return b.components
}

func (b *Builder) UnknownTypes() []string {
	return b.unknownTypes
}

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

	if enumInfo, ok := b.enums[ref.Name]; ok {
		if s := primitiveSchema(enumInfo.BaseType); s != nil {
			return s
		}
	}

	b.ensureComponent(ref.Name)
	return &Schema{Ref: "#/components/schemas/" + ref.Name}
}

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

func (b *Builder) ensureComponent(name string) {
	if _, exists := b.components[name]; exists {
		return
	}
	if b.processing[name] {
		return
	}

	typeInfo, exists := b.types[name]
	if !exists {
		b.components[name] = &Schema{Type: "object"}
		b.unknownTypes = append(b.unknownTypes, name)
		return
	}

	b.processing[name] = true
	defer func() { delete(b.processing, name) }()

	var allOf []*Schema
	for _, embName := range typeInfo.Embedded {
		b.ensureComponent(embName)
		allOf = append(allOf, &Schema{Ref: "#/components/schemas/" + embName})
	}

	ownSchema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}
	var required []string
	for _, field := range typeInfo.Fields {
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

	if len(allOf) > 0 {
		if len(ownSchema.Properties) > 0 {
			allOf = append(allOf, ownSchema)
		}
		b.components[name] = &Schema{AllOf: allOf}
	} else {
		b.components[name] = ownSchema
	}
}

func (b *Builder) buildFieldSchema(field *parser.FieldInfo) *Schema {
	s := b.BuildSchema(field.Type)

	if field.Doc != "" {
		s.Description = field.Doc
	}
	if field.Example != "" {
		s.Example = CoerceScalar(field.Example, s.Type)
	}
	b.applyConstraints(s, field)
	return s
}

func (b *Builder) BuildParamSchema(field *parser.FieldInfo) *Schema {
	s := b.BuildSchema(field.Type)
	b.applyConstraints(s, field)
	return s
}

func (b *Builder) applyConstraints(s *Schema, field *parser.FieldInfo) {
	if field.Default != "" {
		s.Default = CoerceScalar(field.Default, s.Type)
	}

	if enumInfo := b.enums[field.Type.Name]; enumInfo != nil {
		s.Enum = enumInfo.Values
	}
	if field.Validate != "" {
		applyValidators(s, field.Validate)
	}
	if field.Type.IsPtr {
		switch {
		case isScalarType(s.Type):

			s.Nullable = true
		case s.Ref != "":

			*s = Schema{
				Description: s.Description,
				Example:     s.Example,
				AnyOf:       []*Schema{{Ref: s.Ref}, nullSchema()},
			}
		}
	}
}

func isScalarType(t string) bool {
	switch t {
	case "string", "integer", "number", "boolean":
		return true
	}
	return false
}

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
	default:
		if isMin {
			s.MinLength = &n
		} else {
			s.MaxLength = &n
		}
	}
}

func parseNumber(s string) (any, bool) {
	if i, err := strconv.Atoi(s); err == nil {
		return i, true
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return nil, false
}

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

func primitiveSchema(name string) *Schema {
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
		return &Schema{}
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

	if strings.Contains(name, ".") {
		return &Schema{Type: "string"}
	}

	return nil
}
