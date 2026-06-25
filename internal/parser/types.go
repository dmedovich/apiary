package parser

import (
	"reflect"
	"strings"
)

type TypeRef struct {
	Name    string
	IsPtr   bool
	IsSlice bool
	IsMap   bool
	MapKey  string
	Elem    *TypeRef
}

type FieldInfo struct {
	Name       string
	Type       *TypeRef
	JSONName   string
	Doc        string
	Example    string
	Default    string
	Required   bool
	Validate   string
	PathParam  string
	QueryParam string
	Header     string
}

type TypeInfo struct {
	Name     string
	Fields   []*FieldInfo
	Embedded []string
}

type EnumInfo struct {
	BaseType string
	Values   []any
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

func hasValidator(tag, name string) bool {
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if n, _, _ := strings.Cut(part, "="); n == name {
			return true
		}
	}
	return false
}

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
