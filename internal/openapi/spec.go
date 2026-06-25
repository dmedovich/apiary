package openapi

import "github.com/honeynil/apiary/internal/schema"

type SecurityRequirement map[string][]string

type OpenAPI struct {
	OpenAPI    string              `yaml:"openapi"`
	Info       Info                `yaml:"info"`
	Servers    []Server            `yaml:"servers,omitempty"`
	Paths      map[string]PathItem `yaml:"paths,omitempty"`
	Security   any                 `yaml:"security,omitempty"`
	Components *Components         `yaml:"components,omitempty"`
}

type Server struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

type Info struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`
}

type PathItem struct {
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
	Patch  *Operation `yaml:"patch,omitempty"`
}

type Operation struct {
	OperationID string               `yaml:"operationId,omitempty"`
	Summary     string               `yaml:"summary,omitempty"`
	Description string               `yaml:"description,omitempty"`
	Tags        []string             `yaml:"tags,omitempty"`
	Security    any                  `yaml:"security,omitempty"`
	Parameters  []Parameter          `yaml:"parameters,omitempty"`
	RequestBody *RequestBody         `yaml:"requestBody,omitempty"`
	Responses   map[string]*Response `yaml:"responses"`
}

type Parameter struct {
	Name        string         `yaml:"name"`
	In          string         `yaml:"in"`
	Description string         `yaml:"description,omitempty"`
	Required    bool           `yaml:"required"`
	Schema      *schema.Schema `yaml:"schema"`
	Example     any            `yaml:"example,omitempty"`
}

type RequestBody struct {
	Description string                `yaml:"description,omitempty"`
	Required    bool                  `yaml:"required"`
	Content     map[string]*MediaType `yaml:"content"`
}

type MediaType struct {
	Schema *schema.Schema `yaml:"schema"`
}

type Response struct {
	Description string                `yaml:"description"`
	Content     map[string]*MediaType `yaml:"content,omitempty"`
}

type SecurityScheme struct {
	Type         string `yaml:"type"`
	Scheme       string `yaml:"scheme,omitempty"`
	BearerFormat string `yaml:"bearerFormat,omitempty"`
	In           string `yaml:"in,omitempty"`
	Name         string `yaml:"name,omitempty"`
	Description  string `yaml:"description,omitempty"`
}

type Components struct {
	Schemas         map[string]*schema.Schema  `yaml:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `yaml:"securitySchemes,omitempty"`
}
