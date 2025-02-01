package main

import (
	"github.com/OblivionOcean/Goh/utils"
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
	Text        []byte
	Text2       []byte
	ConstLength int
	BufName     string
	Vars        []byte
	RawCode     []byte
	PackageName string
	Dest        string
}

func (g *Generator) New(fpath string) {
	if g.PackageName == "" {
		g.PackageName = "template"
	}
	g.Root, g.RawCode, g.DefindFunc = Parse(fpath)
	g.Generate(nil)
	file, err := os.Create(path.Join(g.Dest, path.Base(fpath)+".go"))
	if err != nil {
		panic(err.Error())
	}
	file.WriteString("// DO NOT EDIT!\n// Generate By Goh\n\n")
	file.Write(g.Text)
	execCommand("goimports -w .")
	execCommand("gofmt -w .")
	fmt.Println("\033[0;32mSuccess\033[0m", fpath)
}

func (g *Generator) Generate(b *Block) {
	if b == nil {
		g.Text = append(g.Text, Goh.String2Bytes("package ")...)
		g.Text = append(g.Text, g.PackageName...)
		g.Text = append(g.Text, Goh.String2Bytes("\nimport (\n\t\"bytes\"\n\t\"github.com/OblivionOcean/Goh/utils\"\n)\n\n")...)
		if g.DefindFunc == nil {
			return
		}
		code, name, err := g.GenerateFunc(g.DefindFunc)
		if err != nil {
			panic(err.Error())
		}
		g.BufName = name
		g.Text = append(g.Text, g.RawCode...)
		g.Text = append(g.Text, code...)
		g.Text = append(g.Text, Goh.String2Bytes(fmt.Sprintf("{\n%s.Grow(", name))...)
		b = g.Root
	}
	switch b.Type {
	case Code:
		g.Text2 = append(g.Text2, b.Content...)
		g.Text2 = append(g.Text2, '\n')
	case HTML:
		if len(bytes.Trim(b.Content, " ")) == 0 {
			break
		}
		tmp := "%s.WriteString(`%s`)\n"
		g.ConstLength += len(b.Content)
		g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf(tmp, g.BufName, b.Content))...)
	case Escape:
		b.Content = bytes.Trim(b.Content, " ")
		if len(b.Content) == 0 {
			break
		}
		switch b.VarType {
		case String:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.EscapeHTML(%s, %s)\n", b.Content, g.BufName))...)
		case Bytes:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.EscapeHTML(Goh.Bytes2String(%s), %s)\n", b.Content, g.BufName))...)
		case Int:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatInt(int64(%s), %s)\n", b.Content, g.BufName))...)
		case Uint:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatUint(uint64(%s), %s)\n", b.Content, g.BufName))...)
		case Bool:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatBool(%s, %s)\n", b.Content, g.BufName))...)
			g.ConstLength += 5
		case Any:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatAny(%s, %s)\n", b.Content, g.BufName))...)
		}
	case Val:
		b.Content = bytes.Trim(b.Content, " ")
		if len(b.Content) == 0 {
			break
		}
		switch b.VarType {
		case String:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("%s.WriteString(%s)\n", g.BufName, b.Content))...)
		case Bytes:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("%s.Write(%s)\n", g.BufName, b.Content))...)
		case Int:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatInt(%s, %s)\n", b.Content, g.BufName))...)
		case Uint:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatUint(%s, %s)\n", b.Content, g.BufName))...)
		case Bool:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatBool(%s, %s)\n", b.Content, g.BufName))...)
			g.ConstLength += 5
		case Any:
			g.Text2 = append(g.Text2, Goh.String2Bytes(fmt.Sprintf("Goh.FormatAny(%s, %s)\n", b.Content, g.BufName))...)
		}
	}
	for i := 0; i < len(b.Children); i++ {
		g.Generate(b.Children[i])
	}
	if b == g.Root {
		g.Text = append(g.Text, Goh.String2Bytes(strconv.Itoa(g.ConstLength))...)
		g.Text = append(g.Text, ')', '\n')
		g.Text = append(g.Text, g.Text2...)
		g.Text = append(g.Text, '}', '\n')
	}

}

func (g *Generator) GenerateFunc(b *Block) (code []byte, name string, err error) {
	src := []byte("package Goh\n")
	src = append(src, b.Content...)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return
	}

	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		return nil, "", errors.New("Definition is not function type")
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
