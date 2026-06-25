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

func (p *Parser) Load(patterns ...string) error {
	return p.LoadDir("", patterns...)
}

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

func (p *Parser) typeRefFromAST(pkg *packages.Package, expr ast.Expr) *TypeRef {
	t := pkg.TypesInfo.TypeOf(expr)
	if t == nil {
		return &TypeRef{Name: "interface{}"}
	}
	return p.typeRef(t)
}

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
		return ref
	}
	resolved := p.typeRef(tn.Type())
	if ref.IsSlice {
		return &TypeRef{Name: "array", IsSlice: true, Elem: resolved}
	}
	resolved.IsPtr = ref.IsPtr
	return resolved
}

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

func (p *Parser) inModule(named *types.Named) bool {
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return false
	}
	if p.modulePath == "" {
		_, isStruct := named.Underlying().(*types.Struct)
		return isStruct
	}
	path := pkg.Path()
	return path == p.modulePath || strings.HasPrefix(path, p.modulePath+"/")
}

func qualifiedName(named *types.Named) string {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return obj.Name()
	}
	return obj.Pkg().Name() + "." + obj.Name()
}

func basicTypeName(b *types.Basic) string {
	return b.Name()
}

func basicBaseName(b *types.Basic) string {
	switch {
	case b.Info()&types.IsString != 0:
		return "string"
	case b.Info()&types.IsInteger != 0:
		return "int"
	}
	return ""
}

func constValue(v constant.Value, base string) any {
	if base == "string" {
		return constant.StringVal(v)
	}
	if n, ok := constant.Int64Val(v); ok {
		return int(n)
	}
	return v.String()
}
