package internal

import (
	"SJT/struct-validate/utils"
	"SJT/struct-validate/utils/slice"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

const (
	DefaultParseTag             = "check"
	DefaultParseCustomValidator = "// @ext:check"
	DefaultParsePath            = "// @path:"
	DefaultParsePackage         = "// @package:"
)

const (
	ErrorKeyword = "error"
)

type Generator interface {
	Parser(entity any) error
}

type Entity struct {
	EntityName   string
	PackageName  string
	PkgRelPath   string // PkgRelPath the package relative path
	Packages     []string
	ParseTag     string
	FileAbsPaths []string
	CustomFuncs  []*FuncType
	Invalid      bool
	Fields       []*Node
}

type Node struct {
	EntityName   string
	Field        string // Field's name
	Tags         []*Tag // Tags Field's tags
	Packages     []string
	Kind         string // Kind reflect.kind
	RealType     string
	Package      string
	PkgRelPath   string   // PkgRelPath package relative path
	FileAbsPaths []string // Fields 子节点
	Fields       []*Node
}

type Tag struct {
	Operator string //Operator  操作符 gt, lt, gte ,email....
	Value    any    // Value 对应的值
}

func NewEntity() *Entity {
	return &Entity{
		Fields:   make([]*Node, 0, 10),
		Packages: make([]string, 0, 5),
		ParseTag: DefaultParseTag,
	}
}

func (e *Entity) IsUseful() bool {
	var isIgnore = true
	for _, field := range e.Fields {
		for _, _ = range field.Tags {
			isIgnore = false
		}
	}

	return isIgnore != true
}

func (e *Entity) AddPackages(packages ...string) {
	for _, p := range packages {
		if strings.Trim(p, "") == "" {
			continue
		}
		if !slice.Contains[string](e.Packages, p) {
			e.Packages = append(e.Packages, p)
		}
	}
}
func (e *Entity) AddFileAbsPaths(absPaths ...string) {
	for _, ap := range absPaths {
		if strings.Trim(ap, "") == "" {
			continue
		}
		if !slice.Contains[string](e.FileAbsPaths, ap) {
			e.FileAbsPaths = append(e.FileAbsPaths, ap)
		}
	}
}

func (e *Entity) SetTag(tag string) {
	e.ParseTag = tag
}

// Parser parses entity.
func (e *Entity) Parser(entity any) error {
	if entity == nil {
		return errors.New("invalid entity")
	}

	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return errors.New("invalid entity")
	}

	e.EntityName = typ.Name()

	// 不处理main包的验证
	if typ.PkgPath() == "main" {
		return nil
	}
	relPath, pkg, _ := getRelPathAndPkg(typ.PkgPath())
	e.PkgRelPath = relPath
	e.PackageName = pkg
	return parseField(&e.Fields, typ, e.ParseTag)
}

// getRelPathAndPkg returns relative path and package name.
func getRelPathAndPkg(pkg string) (string, string, error) {
	if pkg == "" {
		return "", "", errors.New("invalid pkg")
	}
	if pkg == "main" {

		return "", "", errors.New("invalid pkg")
	}
	mod, err := utils.GetModule()
	if err != nil {
		return "", "", err
	}
	if len(pkg) < len(mod) {
		return "", "", errors.New("invalid pkg path")
	}
	rel := pkg[len(mod)+1:]
	pkgArr := strings.Split(strings.Trim(rel, "/"), "/")

	return rel, pkgArr[len(pkgArr)-1], err
}

func (n *Node) AddPackages(packages ...string) {
	for _, p := range packages {
		if strings.Trim(p, "") == "" {
			continue
		}
		if !slice.Contains[string](n.Packages, p) {
			n.Packages = append(n.Packages, p)
		}
	}
}

func (n *Node) AddFileAbsPaths(absPaths ...string) {
	for _, ap := range absPaths {
		if strings.Trim(ap, "") == "" {
			continue
		}
		if !slice.Contains[string](n.FileAbsPaths, ap) {
			n.FileAbsPaths = append(n.FileAbsPaths, ap)
		}
	}
}

func parseField(root *[]*Node, t reflect.Type, tag string) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// 禁止多重指针
		var i = 0
		fT := field.Type
		for fT.Kind() == reflect.Ptr {
			i++
			fT = fT.Elem()
			if i >= 2 {
				return errors.New(t.Name() + "." + field.Name + "只能使用一级指针")
			}
		}

		curNode := &Node{}
		curNode.Field = field.Name

		subTyp := field.Type
		curNode.Kind = subTyp.Kind().String()

		if tags, err := parseTag(field.Tag.Get(tag)); err == nil {
			curNode.Tags = tags
			for _, tag := range tags {
				if _, ok := regexpRoles[Operator(tag.Operator)]; ok {
					curNode.AddPackages("regexp")
				}
				curNode.AddPackages("errors")
			}
		}

		if subTyp.Kind() == reflect.Ptr {
			subTyp = subTyp.Elem()
		}
		curNode.RealType = subTyp.Kind().String()

		relPath, pkg, _ := getRelPathAndPkg(subTyp.PkgPath())

		curNode.PkgRelPath = relPath
		curNode.Package = pkg
		curNode.EntityName = subTyp.Name()
		if subTyp.Kind() == reflect.Struct {
			curNode.Fields = make([]*Node, 0, 10)
			err := parseField(&curNode.Fields, subTyp, tag)
			if err != nil {
				return err
			}
		}
		*root = append(*root, curNode)
	}
	return nil
}

// parseTag returns a Tag pointer slice.
func parseTag(tag string) ([]*Tag, error) {
	tag = strings.Trim(tag, ";")
	if tag == "" {
		return nil, errors.New("empty tag")
	}
	_tags := strings.Split(tag, ";")
	//if len(_tags) == 0 {
	//	return nil, errors.New("empty tag")
	//}
	tags := make([]*Tag, 0, 4)
	for _, t := range _tags {
		segs := strings.Split(strings.Trim(t, " "), " ")
		newTag := &Tag{}
		// notEmpty
		if len(segs) == 1 {
			newTag.Operator = segs[0]
		}
		// gt 0
		if len(segs) == 2 {
			newTag.Operator = segs[0]
			newTag.Value = segs[1]
		}
		tags = append(tags, newTag)
	}
	return tags, nil
}

type FuncType struct {
	Name    string    // 函数名称
	Recv    *RecvType // 函数接受者
	Returns *Returns  // 返回值
}
type Returns struct {
	Name string
	Kind string
}
type RecvType struct {
	Name  string
	Type  string // 接受者类型 值类型或者指针类型
	Value string // 接受者
}

func ParseFile(srcFiles []string) (*ParseResult, error) {
	var res ParseResult
	fts := make([]*FuncType, 0, 10)
	res.Annotations = make(map[string][]string, 10)
	res.Entities = make([]string, 0, 10)
	for _, src := range srcFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		v := &SingleFileVisitor{}
		ast.Walk(v, f)
		fts = append(fts, v.f.ft...)
		for key, val := range v.f.annotations {
			res.Annotations[key] = val
		}
		res.Entities = append(res.Entities, v.f.entities...)
		res.Pkg = v.pkg
	}
	res.FuncType = fts
	return &res, nil
}

type ParseResult struct {
	FuncType    []*FuncType
	Annotations map[string][]string
	Entities    []string
	Pkg         string
}

func (p *ParseResult) GetEntities() []string {
	return p.Entities
}

// GetPath 获取注解自定义路径
func (p *ParseResult) GetPath(entityName string) string {
	ans, ok := p.Annotations[entityName]
	if !ok {
		return ""
	}
	for _, an := range ans {
		if len(an) <= len(DefaultParsePath) {
			continue
		}
		if an[0:len(DefaultParsePath)] == DefaultParsePath {
			return strings.Trim(an[len(DefaultParsePath):], " ")
		}
	}
	return ""
}

// GetPackage 获取注解自定义包名
func (p *ParseResult) GetPackage(entityName string) string {
	ans, ok := p.Annotations[entityName]
	if !ok {
		return ""
	}
	for _, an := range ans {
		if len(an) <= len(DefaultParsePackage) {
			continue
		}
		if an[0:len(DefaultParsePackage)] == DefaultParsePackage {
			return strings.Trim(an[len(DefaultParsePackage):], " ")
		}
	}
	return ""
}

type SingleFileVisitor struct {
	f   *fileVisitor
	pkg string // package name
}

func (s *SingleFileVisitor) Visit(node ast.Node) (w ast.Visitor) {
	n, ok := node.(*ast.File)
	if ok {
		s.pkg = n.Name.Name
		s.f = &fileVisitor{ft: make([]*FuncType, 0, 3), annotations: map[string][]string{}, entities: make([]string, 0, 10)}
		return s.f
	}
	return s
}

type fileVisitor struct {
	ft          []*FuncType         // 方法注解 自定义验证方法
	annotations map[string][]string // 类型注解
	entities    []string
}

func (f *fileVisitor) Visit(node ast.Node) (w ast.Visitor) {
	typ, ok := node.(*ast.FuncDecl)
	if ok {
		var ft FuncType
		//注解 // ext:check
		if typ.Doc == nil {
			return f
		}
		var flag bool
		for _, comment := range typ.Doc.List {
			if strings.Trim(comment.Text, " ") == DefaultParseCustomValidator {
				flag = true
			}
		}
		if !flag {
			return f
		}

		// 函数名
		ft.Name = typ.Name.Name

		if typ.Type.Params.List != nil {
			fmt.Println("自定义函数签名不正确，签名应为：func() error")
			return f
		}

		// 返回值
		if typ.Type.Results != nil && len(typ.Type.Results.List) != 1 {
			fmt.Println("自定义函数签名不正确，签名应为：func() error")
			return f
		}
		if typ.Type.Results == nil {
			return f
		}

		for _, field := range typ.Type.Results.List {
			var ret Returns
			retTyp, ok := field.Type.(*ast.Ident)
			if ok {
				if retTyp.Name != ErrorKeyword {
					return f
				}
				if len(field.Names) > 0 {
					ret.Name = field.Names[0].Name
				}
				ret.Kind = retTyp.Name
				ft.Returns = &ret
			}
			//retStarTyp, ok := field.Type.(*ast.StarExpr)
		}

		// 接受者
		if typ.Recv == nil {
			return f
		}
		for _, field := range typ.Recv.List {
			var recv RecvType
			ValRecv, ok := field.Type.(*ast.Ident)
			if ok {
				recv.Value = ValRecv.Name
				ft.Recv = &recv
			}
			PtrRecv, ok := field.Type.(*ast.StarExpr)
			if ok {
				x, xok := PtrRecv.X.(*ast.Ident)
				if xok {
					recv.Value = x.Name
					ft.Recv = &recv
				}
			}
		}

		f.ft = append(f.ft, &ft)
	}

	gTyp, ok := node.(*ast.GenDecl)
	if ok {
		if gTyp.Tok == token.TYPE {
			entityName := ""
			for _, spec := range gTyp.Specs {
				s, ok := spec.(*ast.TypeSpec)
				if ok {
					entityName = s.Name.Name
					_, isStructTyp := s.Type.(*ast.StructType)
					if isStructTyp {
						f.entities = append(f.entities, entityName)
						comments := make([]string, 0, 2)
						if gTyp.Doc != nil {
							for _, comment := range gTyp.Doc.List {
								if entityName != "" {
									comments = append(comments, comment.Text)
								}
							}
						}

						if entityName != "" && len(comments) > 0 {
							f.annotations[entityName] = comments
						}
					}
				}
			}
		}
	}
	return f
}
