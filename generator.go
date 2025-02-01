package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func execCommand(command string) {
	parts := strings.Split(command, " ")
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

type Generator struct {
	DefindFunc  *Block
	Root        *Block
	Text        *os.File
	Text2       *bytes.Buffer
	ConstLength int
	BufName     string
	RawCode     string
	PackageName string
	Dest        string
}

func (g *Generator) New(fpath string) {
	if g.PackageName == "" {
		g.PackageName = "template"
	}
	g.Text2 = bytes.NewBuffer(nil)
	g.Root, g.RawCode, g.DefindFunc = Parse(fpath)
	file, err := os.Create(path.Join(g.Dest, path.Base(fpath)+".go"))
	if err != nil {
		panic(err.Error())
	}
	file.WriteString("// DO NOT EDIT!\n// Generate By Goh\n\n")
	g.Text = file
	g.Generate(nil)
	execCommand("goimports -w .")
	execCommand("gofmt -w .")
	fmt.Println("\033[0;32mSuccess\033[0m", fpath)
}

func (g *Generator) Generate(b *Block) {
	if b == nil {
		g.Text.WriteString("package ")
		g.Text.WriteString(g.PackageName)
		g.Text.WriteString("\nimport (\n\t\"bytes\"\n\t\"github.com/OblivionOcean/Goh/utils\"\n)\n\n")
		if g.DefindFunc == nil {
			return
		}
		code, name, err := g.GenerateFunc(g.DefindFunc)
		if err != nil {
			panic(err.Error())
		}
		g.BufName = name
		g.Text.WriteString(g.RawCode)
		g.Text.WriteString(code)
		g.Text.WriteString(fmt.Sprintf("{\n%s.Grow(", name))
		b = g.Root
	}
	switch b.Type {
	case Code:
		g.Text.WriteString(b.Content)
		g.Text.WriteString("\n")
	case HTML:
		if len(strings.Trim(b.Content, " ")) == 0 {
			break
		}
		tmp := "%s.WriteString(`%s`)\n"
		g.ConstLength += len(b.Content)
		g.Text2.WriteString(fmt.Sprintf(tmp, g.BufName, b.Content))
	case Escape:
		b.Content = strings.Trim(b.Content, " ")
		if len(b.Content) == 0 {
			break
		}
		var tmp string
		switch b.VarType {
		case String:
			tmp = fmt.Sprintf("Goh.EscapeHTML(%s, %s)\n", b.Content, g.BufName)
		case Bytes:
			tmp = fmt.Sprintf("Goh.EscapeHTML(Goh.Bytes2String(%s), %s)\n", b.Content, g.BufName)
		case Int:
			tmp = fmt.Sprintf("Goh.FormatInt(int64(%s), %s)\n", b.Content, g.BufName)
		case Uint:
			tmp = fmt.Sprintf("Goh.FormatUint(uint64(%s), %s)\n", b.Content, g.BufName)
		case Bool:
			tmp = fmt.Sprintf("Goh.FormatBool(%s, %s)\n", b.Content, g.BufName)
			g.ConstLength += 5
		case Any:
			tmp = fmt.Sprintf("Goh.FormatAny(%s, %s)\n", b.Content, g.BufName)
		}
		g.Text2.WriteString(tmp)
	case Val:
		b.Content = strings.Trim(b.Content, " ")
		if len(b.Content) == 0 {
			break
		}
		var tmp string
		switch b.VarType {
		case String:
			tmp = fmt.Sprintf("%s.WriteString(%s)\n", g.BufName, b.Content)
		case Bytes:
			tmp = fmt.Sprintf("%s.Write(%s)\n", g.BufName, b.Content)
				case Int:
			tmp = fmt.Sprintf("Goh.FormatInt(int64(%s), %s)\n", b.Content, g.BufName)
		case Uint:
			tmp = fmt.Sprintf("Goh.FormatUint(uint64(%s), %s)\n", b.Content, g.BufName)
		case Bool:
			tmp = fmt.Sprintf("Goh.FormatBool(%s, %s)\n", b.Content, g.BufName)
			g.ConstLength += 5
		case Any:
			tmp = fmt.Sprintf("Goh.FormatAny(%s, %s)\n", b.Content, g.BufName)
		}
		g.Text2.WriteString(tmp)
	}
	for i := 0; i < len(b.Children); i++ {
		g.Generate(b.Children[i])
	}
	if b == g.Root {
		g.Text.WriteString(strconv.Itoa(g.ConstLength))
			g.Text.WriteString(")\n")
		g.Text.ReadFrom(g.Text2)
		g.Text.WriteString("}\n")
	}

}

func (g *Generator) GenerateFunc(b *Block) (code string, name string, err error) {
	src := []byte("package Goh\n")
	src = append(src, b.Content...)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return
	}

	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		return "", "", errors.New("Definition is not function type")
	}

	params := funcDecl.Type.Params.List
	if len(params) == 0 {
		err = errors.New("Definition parameters should not be empty")
		return
	}

	lastParam := params[len(params)-1]
	expr := lastParam.Type
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		expr = starExpr.X
	}
	selectorExpr, ok := expr.(*ast.SelectorExpr)
	if !ok {
		err = errors.New("Definition parameters should not be empty")
		return
	}

	if selectorExpr.X.(*ast.Ident).Name != "bytes" && selectorExpr.Sel.Name != "Buffer" {
		err = errors.New("Definition parameters should not be empty")
		return
	}
	if n := len(lastParam.Names); n > 0 {
		name = lastParam.Names[n-1].Name
	}
	code = b.Content
	return
}
