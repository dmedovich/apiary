// Package parser type-checks Go packages via go/packages + go/types and
// extracts struct types, enums, and functions annotated with
// // apiary:operation. Type resolution is semantic: cross-package, external,
// and generic types are handled correctly rather than guessed from bare names.
package parser

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
	"unicode"

	"github.com/honeynil/apiary/internal/annotation"
	"golang.org/x/tools/go/packages"
)

// OperationInfo holds everything extracted from a single annotated function.
type OperationInfo struct {
	Annotation   *annotation.Operation
	RequestType  *TypeRef // nil if no request body
	ResponseType *TypeRef // nil if no response body schema
}

// Parser accumulates type definitions and operations from the loaded packages.
type Parser struct {
	fset       *token.FileSet
	types      map[string]*TypeInfo
	operations []*OperationInfo
	enums      map[string]*EnumInfo
	modulePath string                  // module path, for in-module detection
	compNames  map[*types.Named]string // named type → unique component/enum name
	usedNames  map[string]*types.Named // reverse index for collision detection
}

// New creates a ready-to-use Parser. Call Load to populate it.
func New() *Parser {
	return &Parser{
		fset:      token.NewFileSet(),
		types:     make(map[string]*TypeInfo),
		enums:     make(map[string]*EnumInfo),
		compNames: make(map[*types.Named]string),
		usedNames: make(map[string]*types.Named),
	}
}

// warnf logs a non-fatal diagnostic prefixed with the source position of pos.
func (p *Parser) warnf(pos token.Pos, format string, args ...any) {
	loc := p.fset.Position(pos)
	log.Printf("apiary: warning: %s: "+format, append([]any{loc}, args...)...)
}

// Operations returns all operations found so far.
func (p *Parser) Operations() []*OperationInfo {
	return p.operations
}

// Types returns all struct types found so far.
func (p *Parser) Types() map[string]*TypeInfo {
	return p.types
}

// Enums returns all enum types (named types with const values) found so far.
func (p *Parser) Enums() map[string]*EnumInfo {
	return p.enums
}

// parseFunction tries to extract an OperationInfo from a function declaration
// in the given type-checked package. Returns nil if the function is not
// annotated or has an unsupported signature.
func (p *Parser) parseFunction(pkg *packages.Package, fn *ast.FuncDecl) *OperationInfo {
	if fn.Doc == nil {
		return nil
	}

	// Strip "//" prefix from each comment line.
	var lines []string
	for _, c := range fn.Doc.List {
		text := strings.TrimPrefix(c.Text, "//")
		text = strings.TrimSpace(text)
		lines = append(lines, text)
	}

	op, ok := annotation.Parse(lines)
	if !ok {
		return nil
	}

	// Surface annotation-level diagnostics (e.g. unknown keys) with position.
	for _, w := range op.Warnings {
		p.warnf(fn.Pos(), "%s: %s", fn.Name.Name, w)
	}

	// Default operationId; an explicit `operationId:` annotation wins. A stable,
	// unique operationId lets client generators (openapi-generator, TS clients)
	// name methods sensibly.
	if op.OperationID == "" {
		op.OperationID = defaultOperationID(receiverTypeName(fn), fn.Name.Name)
	}

	// Fall back to the Go doc comment for summary/description, so a handler's
	// existing godoc doubles as its API docs — no need to repeat it in a
	// `summary:` line. Explicit annotations always win.
	if op.Summary == "" {
		if summary, desc := leadingDocComment(lines, fn.Name.Name); summary != "" {
			op.Summary = summary
			if op.Description == "" {
				op.Description = desc
			}
		}
	}

	// Framework handlers carry no type info in signature — types come
	// from annotation request:/response: fields, resolved against the package.
	if isGinHandler(fn) || isStdlibHTTPHandler(fn) {
		return &OperationInfo{
			Annotation:   op,
			RequestType:  p.resolveAnnotationType(pkg, op.Request),
			ResponseType: p.resolveAnnotationType(pkg, op.Response),
		}
	}

	// Supported signatures (results must always be (R, error)):
	//   (ctx context.Context, req T) (R, error)  — standard
	//   (req T) (R, error)                        — no ctx
	//   (ctx context.Context) (R, error)          — no request body
	//   () (R, error)                             — no ctx, no body (rare)
	if fn.Type == nil || fn.Type.Results == nil {
		return p.warnBadSignature(fn)
	}
	results := fn.Type.Results.List
	if len(results) != 2 {
		return p.warnBadSignature(fn)
	}
	if !isErrorType(results[len(results)-1].Type) {
		return p.warnBadSignature(fn)
	}

	respRef := p.typeRefFromAST(pkg, results[0].Type)

	var reqRef *TypeRef
	if fn.Type.Params != nil {
		params := fn.Type.Params.List
		switch len(params) {
		case 0:
			// () (R, error) — no request
		case 1:
			if !isContextType(params[0].Type) {
				// (req T) (R, error) — no ctx, has request
				reqRef = p.typeRefFromAST(pkg, params[0].Type)
			}
			// else: (ctx) (R, error) — ctx only, no request
		case 2:
			if !isContextType(params[0].Type) {
				return p.warnBadSignature(fn) // first param must be context when there are 2 params
			}
			reqRef = p.typeRefFromAST(pkg, params[1].Type)
		default:
			return p.warnBadSignature(fn) // more than 2 params — not supported
		}
	}

	// Annotation request:/response: fields override inferred types.
	if ann := p.resolveAnnotationType(pkg, op.Request); ann != nil {
		reqRef = ann
	}
	if ann := p.resolveAnnotationType(pkg, op.Response); ann != nil {
		respRef = ann
	}

	return &OperationInfo{
		Annotation:   op,
		RequestType:  reqRef,
		ResponseType: respRef,
	}
}

// warnBadSignature reports that fn carries an apiary:operation marker but its
// signature is not one apiary can map to an operation, then returns nil so the
// caller drops it. Without this, the operation would vanish with no explanation.
func (p *Parser) warnBadSignature(fn *ast.FuncDecl) *OperationInfo {
	p.warnf(fn.Pos(),
		"%s has an apiary:operation marker but an unsupported signature — handlers must be (R, error)-returning, gin, or net/http; see README",
		fn.Name.Name)
	return nil
}

// isGinHandler returns true when fn has the gin handler signature:
// func(...) with a single *gin.Context parameter and no return values.
func isGinHandler(fn *ast.FuncDecl) bool {
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		return false
	}
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	param := fn.Type.Params.List[0]
	expr := param.Type
	// Accept *gin.Context
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "gin" && sel.Sel.Name == "Context"
}

// isStdlibHTTPHandler returns true when fn matches:
//
//	func(w http.ResponseWriter, r *http.Request)
//
// Both method receivers and free functions are accepted.
func isStdlibHTTPHandler(fn *ast.FuncDecl) bool {
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		return false
	}
	if fn.Type.Params == nil {
		return false
	}
	var types []ast.Expr
	for _, p := range fn.Type.Params.List {
		n := len(p.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			types = append(types, p.Type)
		}
	}
	if len(types) != 2 {
		return false
	}
	return isHTTPResponseWriter(types[0]) && isHTTPRequestPtr(types[1])
}

// isHTTPResponseWriter matches `http.ResponseWriter` (no pointer).
func isHTTPResponseWriter(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http" && sel.Sel.Name == "ResponseWriter"
}

// isHTTPRequestPtr matches `*http.Request`.
func isHTTPRequestPtr(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http" && sel.Sel.Name == "Request"
}

// parseAnnotationTypeRef converts an annotation type string to a TypeRef.
// Handles "TypeName", "[]TypeName", and "*TypeName".
// Returns nil for empty input.
func parseAnnotationTypeRef(s string) *TypeRef {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if elem, ok := strings.CutPrefix(s, "[]"); ok {
		return &TypeRef{Name: "array", IsSlice: true, Elem: &TypeRef{Name: strings.TrimSpace(elem)}}
	}
	if name, ok := strings.CutPrefix(s, "*"); ok {
		return &TypeRef{Name: name, IsPtr: true}
	}
	return &TypeRef{Name: s}
}

// lowerFirst returns s with its first rune lower-cased ("CreateUser" →
// "createUser").
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// leadingDocComment extracts a summary and description from the human Go doc
// comment that precedes the apiary:operation marker. The first prose line
// becomes the summary; any following prose lines become the description. A
// leading function name is stripped per Go convention ("Login authenticates a
// user" → "Authenticates a user"). Returns empty strings when there is no prose.
func leadingDocComment(lines []string, funcName string) (summary, description string) {
	var prose []string
	for _, line := range lines {
		if strings.HasPrefix(line, "apiary:operation ") {
			break // the godoc ends where the apiary block begins
		}
		if line != "" {
			prose = append(prose, line)
		}
	}
	if len(prose) == 0 {
		return "", ""
	}
	first := prose[0]
	if rest, ok := strings.CutPrefix(first, funcName+" "); ok && rest != "" {
		first = capitalizeFirst(rest)
	}
	if len(prose) > 1 {
		description = strings.Join(prose[1:], " ")
	}
	return first, description
}

// capitalizeFirst upper-cases the first rune of s.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// receiverTypeName returns the (unqualified) receiver type name of a method, or
// "" for a free function. "func (h *TaskHandler) List" → "TaskHandler".
func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// opIDReceiverSuffixes are dropped from a receiver type before composing an
// operationId, so "TaskHandler.List" reads as "taskList", not "taskHandlerList".
var opIDReceiverSuffixes = []string{"Handler", "Controller", "Service", "Server"}

// defaultOperationID derives a default operationId from a method's receiver and
// name. Including the receiver keeps ids unique across handlers that share a
// method name (TaskHandler.List vs CommentHandler.List → taskList / commentList).
// Free functions fall back to the lower-camel function name.
func defaultOperationID(receiver, funcName string) string {
	for _, suf := range opIDReceiverSuffixes {
		if trimmed, ok := strings.CutSuffix(receiver, suf); ok {
			receiver = trimmed
			break
		}
	}
	if receiver != "" {
		return lowerFirst(receiver) + funcName
	}
	return lowerFirst(funcName)
}

// isContextType returns true when expr is context.Context.
func isContextType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "context" && sel.Sel.Name == "Context"
}

// isErrorType returns true when expr is the built-in error interface.
func isErrorType(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "error"
}
