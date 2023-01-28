package pkg

import (
	"SJT/struct-validate/internal"
	"SJT/struct-validate/utils"
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator interface {
	Gen(entities ...any) error
}

type GenDefinition struct {
	entities []*internal.Entity
	parseTag string
}

var _ Generator = &GenDefinition{}

var createdFiles = map[string]struct{}{}

func NewGenDefinition() *GenDefinition {
	return &GenDefinition{entities: make([]*internal.Entity, 0, 10)}
}

// SetTag 设置解析标签
func (g *GenDefinition) SetTag(tag string) {
	g.parseTag = tag
}

func (g *GenDefinition) Gen(entities ...any) error {
	//fmt.Println("generating validate codes...")
	for _, entity := range entities {
		e := internal.NewEntity()
		if g.parseTag != "" {
			e.SetTag(g.parseTag)
		}
		err := e.Parser(entity)
		if err != nil {
			return err
		}

		// 过滤不需要做验证的实体
		//var isIgnore = true
		//for _, field := range e.Fields {
		//	for _, _ = range field.Tags {
		//		isIgnore = false
		//	}
		//}
		//if !isIgnore {
		//	g.entities = append(g.entities, e)
		//}
		if e.IsUseful() {
			g.entities = append(g.entities, e)
		}
	}
	g.GenValidation()
	for f, _ := range createdFiles {
		fmt.Println("created file: ", f)
	}
	//fmt.Println("generation is completed.")
	return nil
}

func genFilePath(dir, entityName string) string {
	return filepath.Join(dir, utils.UnderscoreName(entityName)+"_validate.go")
}

func (g *GenDefinition) GenValidation() {
	for _, entity := range g.entities {
		for _, field := range entity.Fields {
			entity.AddPackages(field.Packages...)
		}

		entity.CustomFuncs = make([]*internal.FuncType, 0, 10)

		//paths, err := utils.ScanFiles("./" + entity.PkgRelPath)
		//if err != nil {
		//	panic(err)
		//}
		//res, err := internal.ParseFile(paths)
		wd, err := utils.GetWorkDirectory()
		if err != nil {
			panic(err)
		}
		dir := filepath.Join(wd, entity.PkgRelPath)
		paths, err := utils.ScanFiles(dir)
		if err != nil {
			panic(err)
		}

		res, err := internal.ParseFile(paths)
		if err != nil {
			panic(err)
		}
		for _, ft := range res.FuncType {
			if entity.EntityName == ft.Recv.Value {
				entity.CustomFuncs = append(entity.CustomFuncs, ft)
			}
		}

		path := res.GetPath(entity.EntityName)
		if path != "" {
			entity.PkgRelPath = path
		}
		if pg := res.GetPackage(entity.EntityName); pg != "" {
			entity.PackageName = pg
		}

		dir = filepath.Join(wd, entity.PkgRelPath)
		if !utils.FileIsExist(dir) {
			os.MkdirAll(dir, 0666)
		}
		file := genFilePath(dir, entity.EntityName)
		var hasError bool
		f, err := os.Create(file)
		if err != nil {
			panic(err)
		}

		defer func() {
			f.Close()
			if hasError {
				os.Remove(file)
			}
		}()

		t, err := template.New(entity.EntityName + "_service").Parse(tpl)
		if err != nil {
			hasError = true
			panic(err)
		}
		if err := t.Execute(f, entity); err != nil {
			hasError = true
			panic(err)
		}
		createdFiles[file] = struct{}{}
		// 生成嵌套结构体验证
		for _, field := range entity.Fields {
			if field.Fields != nil && entity.EntityName != "" && entity.PackageName != "" {
				subGen(field.Fields, field.EntityName, field.Package, field.PkgRelPath)
			}
		}
	}

}

func subGen(subNodes []*internal.Node, entityName, packageName, pkgRelPath string) {
	sub := &internal.Entity{
		EntityName:  entityName,
		PackageName: packageName,
		PkgRelPath:  pkgRelPath,
		Fields:      subNodes,
		CustomFuncs: make([]*internal.FuncType, 0, 10),
	}

	//if !sub.IsUseful() {
	//	return
	//}

	for _, subNode := range subNodes {
		sub.AddPackages(subNode.Packages...)
	}

	wd, err := utils.GetWorkDirectory()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(wd, pkgRelPath)
	paths, err := utils.ScanFiles(dir)
	if err != nil {
		panic(err)
	}

	sub.CustomFuncs = make([]*internal.FuncType, 0, 10)
	res, err := internal.ParseFile(paths)
	for _, ft := range res.FuncType {
		if entityName == ft.Recv.Value {
			sub.CustomFuncs = append(sub.CustomFuncs, ft)
		}
	}

	path := res.GetPath(sub.EntityName)
	if path != "" {
		sub.PkgRelPath = path
	}
	if pg := res.GetPackage(sub.EntityName); pg != "" {
		sub.PackageName = pg
	}

	//exit, _ := utils.PathExist(sub.PkgRelPath)
	//if !exit {
	//	os.MkdirAll(sub.PkgRelPath, os.ModePerm)
	//}
	//file := genFilePath(sub.PkgRelPath, sub.EntityName)

	//dir := filepath.Join(wd, sub.PkgRelPath)
	dir = filepath.Join(wd, sub.PkgRelPath)
	if !utils.FileIsExist(dir) {
		os.MkdirAll(dir, 0666)
	}
	file := genFilePath(dir, sub.EntityName)
	var hasError bool
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}

	defer func() {
		f.Close()
		if hasError {
			os.Remove(file)
		}
	}()

	t, err := template.New(entityName + "_service").Parse(tpl)
	if err != nil {
		hasError = true
		panic(err)
	}
	if err := t.Execute(f, sub); err != nil {
		hasError = true
		panic(err)
	}

	createdFiles[file] = struct{}{}
	for _, subNode := range subNodes {
		if subNode.Fields != nil && subNode.EntityName != "" && subNode.Package != "" {

			subGen(subNode.Fields, subNode.EntityName, subNode.Package, subNode.PkgRelPath)
		}
	}
}

type ScanFile struct {
	Files []string
}

func (s *ScanFile) Resolver() error {
	if len(s.Files) < 1 {
		return errors.New("文件为空")
	}

	dir := filepath.Dir(s.Files[0])

	wd, err := utils.GetWorkDirectory()
	if err != nil {
		return err
	}
	if len(wd) > len(dir) {
		return fmt.Errorf("不支持的操作，work directory：%s，file directory: %s", wd, dir)
	}

	// mod
	module, err := utils.GetModule()
	if err != nil {
		return err
	}
	res, err := internal.ParseFile(s.Files)
	if err != nil {
		return err
	}
	// 不处理main 包的验证
	if res.Pkg == "main" {
		// todo
		return nil
	}

	if len(res.Entities) == 0 {
		return nil
	}

	// write temp file
	buf := bytes.Buffer{}
	buf.WriteString("package main")
	buf.WriteString("\r\n")
	buf.WriteString("import (")
	buf.WriteString("\r\n")
	buf.WriteString(`"SJT/struct-validate/pkg"`)
	buf.WriteString("\r\n")
	buf.WriteString(`"`)
	buf.WriteString(strings.Replace(filepath.Clean(filepath.Join(module, dir[len(wd):])), "\\", "/", -1))
	buf.WriteString(`"`)
	buf.WriteString("\r\n)\r\n")
	buf.WriteString("func main() {")
	buf.WriteString("\r\n")
	buf.WriteString("g := pkg.NewGenDefinition()\r\n")
	buf.WriteString("g.Gen(")
	for i, entity := range res.GetEntities() {
		buf.WriteString(res.Pkg + "." + entity + "{}")
		if i < len(res.GetEntities())-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(")")
	buf.WriteString("\r\n")
	buf.WriteString("}")

	tempFolder := utils.RandString(6)
	tempDir := filepath.Join(os.TempDir(), tempFolder)
	os.MkdirAll(tempDir, 0666)

	tempFile := filepath.Join(tempDir, "main.go")
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.RemoveAll(tempDir)
		buf.Reset()
	}()
	formatCode, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	f.Write(formatCode)

	cmd := exec.Command("go", "run", tempFile)
	var output, stderr bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
	fmt.Println(stderr.String())
	fmt.Println(output.String())
	//log.Println(string(cmdOut))
	return nil
}
