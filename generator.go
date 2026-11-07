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

func runTool(name string, args ...string) {
	// 1. Check if the tool exists directly in the system PATH
	if _, err := exec.LookPath(name); err == nil {
		cmd := exec.Command(name, args...)
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return
	}

	// 2. Fallback: Try running via 'go tool <name>'
	// (Great for tools like goimports tracked via Go's tool directives)
	if _, err := exec.LookPath("go"); err == nil {
		goArgs := append([]string{"tool", name}, args...)
		cmd := exec.Command("go", goArgs...)
		_ = cmd.Run()
		return
	}

	// If neither works, the tool is optional and not installed—skip silently.
}

// Generator walks the block tree, emitting the Children of Extend blocks
// inline (so a child's override content lands where the layout declared it).
type Generator struct {
	DefineFunc  *Block
	Root        Blocks
	Text        *os.File
	Text2       *bytes.Buffer
	ConstLength int
	BufName     string
	RawCode     string
	PackageName string
	Dest        string
}

// New parses the template and generates the Go code.
func (g *Generator) New(fpath string) {
	if g.PackageName == "" {
		g.PackageName = "template"
	}
	if g.Dest == "" {
		g.Dest = "."
	}
	g.Text2 = bytes.NewBuffer(nil)
	g.Root, g.RawCode, g.DefineFunc = (&Parser{}).Parse(fpath)
	file, err := os.Create(path.Join(g.Dest, path.Base(fpath)+".go"))
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	file.WriteString("// DO NOT EDIT!\n// Generate By Goh\n\n")
	g.Text = file
	g.generate()
	fmt.Println("\033[0;32mSuccess\033[0m", fpath)
}

func (g *Generator) generate() {
	g.Text.WriteString("package ")
	g.Text.WriteString(g.PackageName)
	if g.DefineFunc == nil {
		return
	}
	g.Text.WriteString("\nimport (\n\t\"bytes\"\n\t\"github.com/OblivionOcean/Goh/utils\"\n)\n\n")
	code, name, err := g.gFunc(g.DefineFunc)
	if err != nil {
		panic(err.Error())
	}
	g.BufName = name
	g.Text.WriteString(g.RawCode)
	g.Text.WriteString(code)
	g.Text.WriteString(fmt.Sprintf("{\n%s.Grow(", name))
	// Walk the block tree, emitting blocks in order.
	g.emitBlocks(g.Root)
	g.Text.WriteString(strconv.Itoa(g.ConstLength))
	g.Text.WriteString(")\n")
	g.Text.ReadFrom(g.Text2)
	g.Text.WriteString("}\n")
}

// emitBlocks walks a list of blocks and emits them in order. For Extend
// blocks, it recursively emits the block's Children (the child's override
// content).
func (g *Generator) emitBlocks(blocks Blocks) {
	for _, b := range blocks {
		switch b.Type {
		case Code:
			g.Text2.WriteString(b.Content)
			g.Text2.WriteString("\n")
		case HTML:
			g.gHTML(b)
		case Escape:
			g.gEscape(b)
		case Val:
			g.gVal(b)
		case Extend:
			// Emit the block's children (the child's override content).
			g.emitBlocks(b.Children)
		}
	}
}

func (g *Generator) gFunc(b *Block) (code string, name string, err error) {
	src := []byte("package Goh\n")
	src = append(src, b.Content...)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return "", "", err
	}

	if len(file.Decls) == 0 {
		return "", "", errors.New("definition has no declarations")
	}

	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		return "", "", errors.New("definition is not function type")
	}

	params := funcDecl.Type.Params.List
	if len(params) == 0 {
		err = ErrDefineParamsEmpty
		return
	}

	lastParam := params[len(params)-1]
	expr := lastParam.Type
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		expr = starExpr.X
	}
	selectorExpr, ok := expr.(*ast.SelectorExpr)
	if !ok {
		err = ErrDefineParamsEmpty
		return
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok || ident.Name != "bytes" || selectorExpr.Sel.Name != "Buffer" {
		err = ErrDefineParamsEmpty
		return
	}
	if n := len(lastParam.Names); n > 0 {
		name = lastParam.Names[n-1].Name
	}
	code = b.Content
	return
}

func (g *Generator) gVal(b *Block) {
	b.Content = strings.Trim(b.Content, " ")
	if len(b.Content) == 0 {
		return
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
	case Float:
		tmp = fmt.Sprintf("Goh.FormatFloat(%s, %s)\n", b.Content, g.BufName)
	case Bool:
		tmp = fmt.Sprintf("Goh.FormatBool(%s, %s)\n", b.Content, g.BufName)
		g.ConstLength += 5
	case Any:
		tmp = fmt.Sprintf("Goh.FormatAny(%s, %s)\n", b.Content, g.BufName)
	}
	g.Text2.WriteString(tmp)
}

func (g *Generator) gEscape(b *Block) {
	b.Content = strings.Trim(b.Content, " ")
	if len(b.Content) == 0 {
		return
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
	case Float:
		tmp = fmt.Sprintf("Goh.FormatFloat(%s, %s)\n", b.Content, g.BufName)
	case Bool:
		tmp = fmt.Sprintf("Goh.FormatBool(%s, %s)\n", b.Content, g.BufName)
		g.ConstLength += 5
	case Any:
		tmp = fmt.Sprintf("Goh.FormatAny(%s, %s)\n", b.Content, g.BufName)
	}
	g.Text2.WriteString(tmp)
}

func (g *Generator) gHTML(b *Block) {
	if len(strings.Trim(b.Content, " ")) == 0 {
		return
	}
	tmp := "%s.WriteString(`%s`)\n"

	// Handle backtick escaping in HTML
	b.Content = strings.ReplaceAll(b.Content, "`", "\\`")

	g.ConstLength += len(b.Content)
	g.Text2.WriteString(fmt.Sprintf(tmp, g.BufName, b.Content))
}
