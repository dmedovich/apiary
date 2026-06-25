package parser

import (
	"go/ast"
	"go/constant"
	"go/types"
	"log"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Load type-checks the packages matching the given patterns (the same patterns
// the go tool understands: ./..., ./pkg, import paths) and extracts operations,
// struct types and enums using full go/types information. Cross-package and
// external types are resolved properly instead of guessed from bare names.
func (p *Parser) Load(patterns ...string) error {
	return p.LoadDir("", patterns...)
}

// LoadDir is like Load but resolves patterns relative to dir (used to load a
// package that lives in a different module, e.g. nested example modules).
func (p *Parser) LoadDir(dir string, patterns ...string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports |
			packages.NeedDeps | packages.NeedModule,
		Fset: p.fset,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		if p.modulePath == "" && pkg.Module != nil {
			p.modulePath = pkg.Module.Path
		}
	}
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			log.Printf("apiary: warning: %s", e)
		}
		p.parsePackage(pkg)
	}
	return nil
}

// parsePackage scans a type-checked package for annotated handler functions.
func (p *Parser) parsePackage(pkg *packages.Package) {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if op := p.parseFunction(pkg, fn); op != nil {
				p.operations = append(p.operations, op)
			}
		}
	}
}

// typeRefFromAST resolves the static type of an AST expression (a handler
// parameter or result) via go/types and lowers it to a TypeRef.
func (p *Parser) typeRefFromAST(pkg *packages.Package, expr ast.Expr) *TypeRef {
	t := pkg.TypesInfo.TypeOf(expr)
	if t == nil {
		return &TypeRef{Name: "interface{}"}
	}
	return p.typeRef(t)
}

// resolveAnnotationType resolves a request:/response: annotation type name (e.g.
// "UserDTO", "[]UserDTO", "*UserDTO") against the package scope so its schema is
// registered. Unresolved names fall back to a bare TypeRef (placeholder).
func (p *Parser) resolveAnnotationType(pkg *packages.Package, s string) *TypeRef {
	ref := parseAnnotationTypeRef(s)
	if ref == nil {
		return nil
	}
	leaf := ref
	if ref.IsSlice && ref.Elem != nil {
		leaf = ref.Elem
	}
	obj := pkg.Types.Scope().Lookup(leaf.Name)
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return ref // not found in this package — keep bare name
	}
	resolved := p.typeRef(tn.Type())
	if ref.IsSlice {
		return &TypeRef{Name: "array", IsSlice: true, Elem: resolved}
	}
	resolved.IsPtr = ref.IsPtr
	return resolved
}

// typeRef lowers a go/types Type to a TypeRef, registering any in-module named
// struct types it encounters into p.types.
func (p *Parser) typeRef(t types.Type) *TypeRef {
	switch u := t.(type) {
	case *types.Pointer:
		inner := p.typeRef(u.Elem())
		if inner != nil {
			inner.IsPtr = true
		}
		return inner
	case *types.Slice:
		return &TypeRef{Name: "array", IsSlice: true, Elem: p.typeRef(u.Elem())}
	case *types.Array:
		return &TypeRef{Name: "array", IsSlice: true, Elem: p.typeRef(u.Elem())}
	case *types.Map:
		return &TypeRef{Name: "map", IsMap: true, MapKey: p.typeRef(u.Key()).Name, Elem: p.typeRef(u.Elem())}
	case *types.Named:
		return p.namedRef(u)
	case *types.Basic:
		return &TypeRef{Name: basicTypeName(u)}
	case *types.Interface:
		return &TypeRef{Name: "interface{}"}
	}
	return &TypeRef{Name: "interface{}"}
}

// namedRef lowers a named type. Well-known and out-of-module types map to a
// qualified name handled by the schema package; in-module structs are expanded
// into component schemas; in-module named scalars become enums (if they have
// const values) or their underlying basic type.
func (p *Parser) namedRef(named *types.Named) *TypeRef {
	if !p.inModule(named) {
		return &TypeRef{Name: qualifiedName(named)}
	}
	switch under := named.Underlying().(type) {
	case *types.Struct:
		return &TypeRef{Name: p.registerStruct(named, under)}
	case *types.Basic:
		if name := p.enumFor(named); name != "" {
			return &TypeRef{Name: name}
		}
		return &TypeRef{Name: basicTypeName(under)}
	case *types.Interface:
		return &TypeRef{Name: "interface{}"}
	default:
		return p.typeRef(named.Underlying())
	}
}

// registerStruct lowers a named struct into a TypeInfo (once) and returns its
// component name. It registers before recursing so recursive types terminate.
func (p *Parser) registerStruct(named *types.Named, st *types.Struct) string {
	name := p.componentName(named)
	if _, done := p.types[name]; done {
		return name
	}
	info := &TypeInfo{Name: name}
	p.types[name] = info
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		if field.Embedded() {
			if emb := p.embeddedComponent(field.Type()); emb != "" {
				info.Embedded = append(info.Embedded, emb)
			}
			continue
		}
		if !field.Exported() {
			continue
		}
		if fi := p.fieldInfo(field, st.Tag(i)); fi != nil {
			info.Fields = append(info.Fields, fi)
		}
	}
	return name
}

// embeddedComponent returns the component name for an embedded struct field
// (unwrapping a pointer), or "" if it is not an in-module struct.
func (p *Parser) embeddedComponent(t types.Type) string {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok || !p.inModule(named) {
		return ""
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return ""
	}
	return p.registerStruct(named, st)
}

// fieldInfo lowers a struct field plus its raw tag into a FieldInfo, or nil when
// the field is excluded (json:"-").
func (p *Parser) fieldInfo(field *types.Var, rawTag string) *FieldInfo {
	tags := parseStructTag(rawTag)
	jsonName := tags.json
	if jsonName == "" {
		jsonName = goNameToJSON(field.Name())
	}
	if jsonName == "-" {
		return nil
	}
	return &FieldInfo{
		Name:       field.Name(),
		Type:       p.typeRef(field.Type()),
		JSONName:   jsonName,
		Doc:        tags.doc,
		Example:    tags.example,
		Default:    tags.defaultV,
		Required:   hasValidator(tags.validate, "required"),
		Validate:   tags.validate,
		PathParam:  tags.path,
		QueryParam: tags.query,
		Header:     tags.header,
	}
}

// enumFor returns the component name of a named scalar type that has const
// values (an enum), registering its values in source order; "" if it is not an
// enum. Works across packages.
func (p *Parser) enumFor(named *types.Named) string {
	if name, ok := p.compNames[named]; ok {
		if _, isEnum := p.enums[name]; isEnum {
			return name
		}
	}
	basic, ok := named.Underlying().(*types.Basic)
	if !ok {
		return ""
	}
	base := basicBaseName(basic)
	if base == "" {
		return ""
	}
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return ""
	}
	var consts []*types.Const
	scope := pkg.Scope()
	for _, n := range scope.Names() {
		c, ok := scope.Lookup(n).(*types.Const)
		if ok && c.Type() == named {
			consts = append(consts, c)
		}
	}
	if len(consts) == 0 {
		return ""
	}
	sort.Slice(consts, func(i, j int) bool { return consts[i].Pos() < consts[j].Pos() })

	name := p.componentName(named)
	info := &EnumInfo{BaseType: base}
	for _, c := range consts {
		info.Values = append(info.Values, constValue(c.Val(), base))
	}
	p.enums[name] = info
	return name
}

// componentName returns a stable, unique component name for a named type,
// qualifying with the package name when two packages share a type name.
func (p *Parser) componentName(named *types.Named) string {
	if n, ok := p.compNames[named]; ok {
		return n
	}
	base := named.Obj().Name()
	name := base
	if existing, taken := p.usedNames[name]; taken && existing != named {
		if pkg := named.Obj().Pkg(); pkg != nil {
			name = capitalizeFirst(pkg.Name()) + base
		}
		for i := 2; ; i++ {
			existing2, taken2 := p.usedNames[name]
			if !taken2 || existing2 == named {
				break
			}
			name = base + strconv.Itoa(i)
		}
	}
	p.compNames[named] = name
	p.usedNames[name] = named
	return name
}

// inModule reports whether a named type belongs to the module being scanned
// (so it should be expanded into a component schema). Types from other modules
// and the standard library are treated as external.
func (p *Parser) inModule(named *types.Named) bool {
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return false
	}
	if p.modulePath == "" {
		// No module info (e.g. GOPATH) — best effort: expand any struct we can see.
		_, isStruct := named.Underlying().(*types.Struct)
		return isStruct
	}
	path := pkg.Path()
	return path == p.modulePath || strings.HasPrefix(path, p.modulePath+"/")
}

// qualifiedName returns "pkg.Type" for a named type, the form the schema package
// recognises for well-known external types (time.Time, uuid.UUID, ...).
func qualifiedName(named *types.Named) string {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return obj.Name()
	}
	return obj.Pkg().Name() + "." + obj.Name()
}

// basicTypeName returns the schema-facing name of a basic type (e.g. "int64").
func basicTypeName(b *types.Basic) string {
	return b.Name()
}

// basicBaseName classifies a basic type as an enum base ("string" or "int"),
// or "" if it cannot back an enum.
func basicBaseName(b *types.Basic) string {
	switch {
	case b.Info()&types.IsString != 0:
		return "string"
	case b.Info()&types.IsInteger != 0:
		return "int"
	}
	return ""
}

// constValue extracts a Go constant value as a string or int for an enum list.
func constValue(v constant.Value, base string) any {
	if base == "string" {
		return constant.StringVal(v)
	}
	if n, ok := constant.Int64Val(v); ok {
		return int(n)
	}
	return v.String()
}
