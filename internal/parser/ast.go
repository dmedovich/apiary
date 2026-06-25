package parser

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
	"unicode"

	"github.com/yaop-labs/apiary/internal/annotation"
	"golang.org/x/tools/go/packages"
)

type OperationInfo struct {
	Annotation   *annotation.Operation
	RequestType  *TypeRef
	ResponseType *TypeRef
}

type Parser struct {
	fset       *token.FileSet
	types      map[string]*TypeInfo
	operations []*OperationInfo
	enums      map[string]*EnumInfo
	modulePath string
	compNames  map[*types.Named]string
	usedNames  map[string]*types.Named
}

func New() *Parser {
	return &Parser{
		fset:      token.NewFileSet(),
		types:     make(map[string]*TypeInfo),
		enums:     make(map[string]*EnumInfo),
		compNames: make(map[*types.Named]string),
		usedNames: make(map[string]*types.Named),
	}
}

func (p *Parser) warnf(pos token.Pos, format string, args ...any) {
	loc := p.fset.Position(pos)
	log.Printf("apiary: warning: %s: "+format, append([]any{loc}, args...)...)
}

func (p *Parser) Operations() []*OperationInfo {
	return p.operations
}

func (p *Parser) Types() map[string]*TypeInfo {
	return p.types
}

func (p *Parser) Enums() map[string]*EnumInfo {
	return p.enums
}

func (p *Parser) parseFunction(pkg *packages.Package, fn *ast.FuncDecl) *OperationInfo {
	if fn.Doc == nil {
		return nil
	}

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

	for _, w := range op.Warnings {
		p.warnf(fn.Pos(), "%s: %s", fn.Name.Name, w)
	}

	if op.OperationID == "" {
		op.OperationID = defaultOperationID(receiverTypeName(fn), fn.Name.Name)
	}

	if op.Summary == "" {
		if summary, desc := leadingDocComment(lines, fn.Name.Name); summary != "" {
			op.Summary = summary
			if op.Description == "" {
				op.Description = desc
			}
		}
	}

	if isGinHandler(fn) || isStdlibHTTPHandler(fn) {
		return &OperationInfo{
			Annotation:   op,
			RequestType:  p.resolveAnnotationType(pkg, op.Request),
			ResponseType: p.resolveAnnotationType(pkg, op.Response),
		}
	}

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

		case 1:
			if !isContextType(params[0].Type) {
				reqRef = p.typeRefFromAST(pkg, params[0].Type)
			}

		case 2:
			if !isContextType(params[0].Type) {
				return p.warnBadSignature(fn)
			}
			reqRef = p.typeRefFromAST(pkg, params[1].Type)
		default:
			return p.warnBadSignature(fn)
		}
	}

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

func (p *Parser) warnBadSignature(fn *ast.FuncDecl) *OperationInfo {
	p.warnf(fn.Pos(),
		"%s has an apiary:operation marker but an unsupported signature; handlers must be (R, error)-returning, gin, or net/http; see README",
		fn.Name.Name)
	return nil
}

func isGinHandler(fn *ast.FuncDecl) bool {
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		return false
	}
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	param := fn.Type.Params.List[0]
	expr := param.Type

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

func isHTTPResponseWriter(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http" && sel.Sel.Name == "ResponseWriter"
}

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

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func leadingDocComment(lines []string, funcName string) (summary, description string) {
	var prose []string
	for _, line := range lines {
		if strings.HasPrefix(line, "apiary:operation ") {
			break
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

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

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

var opIDReceiverSuffixes = []string{"Handler", "Controller", "Service", "Server"}

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

func isContextType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "context" && sel.Sel.Name == "Context"
}

func isErrorType(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "error"
}
