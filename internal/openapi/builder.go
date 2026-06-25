package openapi

import (
	"log"
	"strconv"
	"strings"

	"github.com/honeynil/apiary/internal/parser"
	"github.com/honeynil/apiary/internal/schema"
)

// builtinSchemes maps short names to their SecurityScheme definitions.
var builtinSchemes = map[string]*SecurityScheme{
	"bearer": {
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Bearer token — pass as: Authorization: Bearer <token>",
	},
	"basic": {
		Type:        "http",
		Scheme:      "basic",
		Description: "HTTP Basic authentication",
	},
	"apikey": {
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "API key passed in the X-API-Key header",
	},
}

type securityEntry struct {
	name    string // key used in the document (like "adminAuth")
	builtin string // builtin type: "bearer", "basic", "apikey"
}

// Builder assembles an OpenAPI 3.1 document from parsed operations.
type Builder struct {
	title       string
	version     string
	description string
	servers     []Server
	security    []securityEntry
}

func NewBuilder(title, version string) *Builder {
	return &Builder{title: title, version: version}
}

func (b *Builder) WithDescription(desc string) *Builder {
	b.description = desc
	return b
}

// WithServer appends a server URL to the document's servers list. Empty URLs
// are ignored.
func (b *Builder) WithServer(url string) *Builder {
	if url != "" {
		b.servers = append(b.servers, Server{URL: url})
	}
	return b
}

func (b *Builder) WithSecurity(scheme string) *Builder {
	name, builtin, found := strings.Cut(scheme, ":")
	if !found {
		builtin = name
	}
	b.security = append(b.security, securityEntry{name: name, builtin: builtin})
	return b
}

func (b *Builder) Build(operations []*parser.OperationInfo, types map[string]*parser.TypeInfo, enums ...map[string]*parser.EnumInfo) (*OpenAPI, error) {
	var enumMap map[string]*parser.EnumInfo
	if len(enums) > 0 {
		enumMap = enums[0]
	}
	sb := schema.NewBuilder(types, enumMap)
	sb.EnsureErrorResponse()

	spec := &OpenAPI{
		OpenAPI: "3.1.0",
		Info:    Info{Title: b.title, Version: b.version, Description: b.description},
		Servers: b.servers,
		Paths:   make(map[string]PathItem),
	}

	seenOpID := make(map[string]string) // operationId → "METHOD path" of first use
	for _, opInfo := range operations {
		op, err := b.buildOperation(opInfo, sb, types)
		if err != nil {
			return nil, err
		}
		path := opInfo.Annotation.Path
		method := strings.ToUpper(opInfo.Annotation.Method)
		if op.OperationID != "" {
			if first, dup := seenOpID[op.OperationID]; dup {
				log.Printf("apiary: warning: duplicate operationId %q (first at %s, again at %s %s) — set an explicit `operationId:` to disambiguate",
					op.OperationID, first, method, path)
			} else {
				seenOpID[op.OperationID] = method + " " + path
			}
		}
		item := spec.Paths[path]
		var slot **Operation
		switch method {
		case "GET":
			slot = &item.Get
		case "POST":
			slot = &item.Post
		case "PUT":
			slot = &item.Put
		case "DELETE":
			slot = &item.Delete
		case "PATCH":
			slot = &item.Patch
		default:
			log.Printf("apiary: warning: unsupported HTTP method %q for %s — operation skipped (supported: GET, POST, PUT, DELETE, PATCH)", method, path)
			continue
		}
		if *slot != nil {
			log.Printf("apiary: warning: duplicate operation %s %s — overwriting previous definition", method, path)
		}
		*slot = op
		spec.Paths[path] = item
	}

	// Build components
	components := &Components{Schemas: sb.Components()}

	// Warn about cross-package types that were not resolved.
	for _, unknown := range sb.UnknownTypes() {
		log.Printf("apiary: warning: type %q not found — add its package to the scan pattern", unknown)
	}

	// Register security schemes.
	if len(b.security) > 0 {
		components.SecuritySchemes = make(map[string]*SecurityScheme)
		var globalReqs []SecurityRequirement
		for _, entry := range b.security {
			scheme, ok := builtinSchemes[strings.ToLower(entry.builtin)]
			if !ok {
				log.Printf("apiary: warning: unknown security scheme type %q (supported: bearer, basic, apikey)", entry.builtin)
				continue
			}
			components.SecuritySchemes[entry.name] = scheme
			globalReqs = append(globalReqs, SecurityRequirement{entry.name: {}})
		}
		if len(globalReqs) > 0 {
			spec.Security = globalReqs
		}
	}

	if len(components.Schemas) > 0 || len(components.SecuritySchemes) > 0 {
		spec.Components = components
	}
	return spec, nil
}

// buildOperation converts a single OperationInfo into an OpenAPI Operation.
func (b *Builder) buildOperation(
	opInfo *parser.OperationInfo,
	sb *schema.Builder,
	types map[string]*parser.TypeInfo,
) (*Operation, error) {
	ann := opInfo.Annotation
	method := strings.ToUpper(ann.Method)

	op := &Operation{
		OperationID: ann.OperationID,
		Summary:     ann.Summary,
		Description: ann.Description,
		Tags:        ann.Tags,
		Responses:   make(map[string]*Response),
	}

	// Per-operation security override (nil means inherit the global setting).
	if sec := operationSecurity(ann.Security); sec != nil {
		op.Security = sec
	}

	if reqRef := opInfo.RequestType; reqRef != nil {
		if reqRef.IsSlice || reqRef.IsMap {
			op.RequestBody = jsonBody(sb.BuildSchema(reqRef))
		} else {
			typeInfo := types[reqRef.Name]

			addParam := func(name, in string, required bool, f *parser.FieldInfo) {
				sch := sb.BuildParamSchema(f)
				op.Parameters = append(op.Parameters, Parameter{
					Name:        name,
					In:          in,
					Description: f.Doc,
					Required:    required,
					Schema:      sch,
					Example:     schema.CoerceScalar(f.Example, sch.Type),
				})
			}

			// Classify fields by where they belong in the OpenAPI operation.
			var pathFields, queryFields, headerFields, bodyFields []*parser.FieldInfo
			if typeInfo != nil {
				for _, f := range typeInfo.Fields {
					switch {
					case f.PathParam != "":
						pathFields = append(pathFields, f)
					case f.QueryParam != "":
						queryFields = append(queryFields, f)
					case f.Header != "":
						headerFields = append(headerFields, f)
					default:
						bodyFields = append(bodyFields, f)
					}
				}
			}

			// Emit explicit parameters grouped path → header → query.
			for _, f := range pathFields {
				addParam(f.PathParam, "path", true, f)
			}
			for _, f := range headerFields {
				addParam(f.Header, "header", f.Required, f)
			}
			for _, f := range queryFields {
				addParam(f.QueryParam, "query", f.Required, f)
			}

			if method == "GET" {
				// GET has no body — remaining fields become query parameters.
				for _, f := range bodyFields {
					addParam(f.JSONName, "query", f.Required, f)
				}
			} else if needsRequestBody(typeInfo, len(bodyFields)) {
				op.RequestBody = jsonBody(sb.BuildSchema(reqRef))
			}
		}
	}

	if opInfo.ResponseType != nil {
		op.Responses["200"] = &Response{
			Description: "OK",
			Content:     jsonContent(sb.BuildSchema(opInfo.ResponseType)),
		}
	} else {
		op.Responses["200"] = &Response{Description: "OK"}
	}

	for _, errSpec := range ann.Errors {
		schemaRef := "#/components/schemas/ErrorResponse"
		if errSpec.Schema != "" {
			sb.BuildSchemaByName(errSpec.Schema)
			schemaRef = "#/components/schemas/" + errSpec.Schema
		}
		op.Responses[strconv.Itoa(errSpec.Code)] = &Response{
			Description: httpStatusText(errSpec.Code),
			Content:     jsonContent(&schema.Schema{Ref: schemaRef}),
		}
	}

	return op, nil
}

// operationSecurity translates a per-operation `security:` annotation into the
// value for Operation.Security. It returns:
//   - nil           when no override is given (inherit the global requirement),
//   - an empty slice for `security: none` (explicitly public, overrides global),
//   - the listed schemes otherwise.
func operationSecurity(security []string) any {
	if len(security) == 0 {
		return nil
	}
	if len(security) == 1 && strings.ToLower(security[0]) == "none" {
		return []SecurityRequirement{}
	}
	reqs := make([]SecurityRequirement, 0, len(security))
	for _, s := range security {
		reqs = append(reqs, SecurityRequirement{s: {}})
	}
	return reqs
}

// needsRequestBody decides whether a non-GET operation carries a JSON body.
// A body is emitted when there are explicit body fields, or when the request
// type is opaque — not introspectable (e.g. an external type referenced by a
// gin/net-http handler), so we cannot tell its shape and assume a body.
//
// A type whose fields are exclusively path, query and/or header parameters
// carries no body: e.g. DELETE /users/{id} bound to {ID int64 `path:"id"`}
// produces no requestBody.
func needsRequestBody(typeInfo *parser.TypeInfo, bodyFieldCount int) bool {
	return bodyFieldCount > 0 || typeInfo == nil
}

// jsonContent wraps a schema as a single application/json content map, the
// shape every request body and response in this generator uses.
func jsonContent(s *schema.Schema) map[string]*MediaType {
	return map[string]*MediaType{"application/json": {Schema: s}}
}

// jsonBody wraps a schema as a required application/json request body.
func jsonBody(s *schema.Schema) *RequestBody {
	return &RequestBody{Required: true, Content: jsonContent(s)}
}

// httpStatusTexts maps the status codes apiary emits to their reason phrases.
var httpStatusTexts = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	409: "Conflict",
	422: "Unprocessable Entity",
	429: "Too Many Requests",
	500: "Internal Server Error",
	502: "Bad Gateway",
	503: "Service Unavailable",
}

func httpStatusText(code int) string {
	if t, ok := httpStatusTexts[code]; ok {
		return t
	}
	return "Error"
}
