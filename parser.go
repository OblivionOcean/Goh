package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	Goh "github.com/OblivionOcean/Goh/utils"
)

// Block type constants. A block is either a leaf (Escape, Val, HTML, Code)
// or a container (Extend) that holds nested blocks in its Children slice.
const (
	Escape = iota
	Val
	HTML
	Code
	Extend
)

// Variable-type constants for the value and escaped-value statements.
const (
	Bool = iota
	Int
	Uint
	Float
	String
	Bytes
	Any
)

var (
	ErrDefineParamsEmpty = errors.New("definition parameters should not be empty")
)

// readFile reads a template file into a string, panicking on error.
func readFile(fpath string) string {
	file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	tmp, err := io.ReadAll(file)
	if err != nil {
		panic(err.Error())
	}
	return string(tmp)
}

// Block is the unit of the template tree. It has no Before/After/AfterCount
// fields. Instead, it has a Children slice that holds the nested blocks
// (the content of an @-block). A block is either a leaf (HTML, Val, Escape,
// Code) or a container (Extend) with Children.
type Block struct {
	Type     int
	Content  string
	VarType  int
	Children Blocks
}

type Blocks []*Block

func (b *Blocks) addChild(child *Block) {
	if len(*b) > 0 && child.Type == HTML && (*b)[len(*b)-1].Type == HTML {
		if len((*b)[len(*b)-1].Content) > 0 {
			(*b)[len(*b)-1].Content += child.Content
		} else {
			(*b)[len(*b)-1] = child
		}
	} else {
		*b = append(*b, child)
	}
}

// Parser parses a template file into a block tree. It has a single
// `current` pointer that points to the block currently being filled (the
// innermost @-block), or nil if we are at the root level.
type Parser struct {
	text       string            // the remaining input, consumed left-to-right as directives are parsed
	cursor     int               // index into text where the current directive's "<%" tag begins
	endCursor  int               // index into text where the current directive's "%>" tag begins
	root       Blocks            // the top-level block list; children are appended here unless we are inside an @-block
	defineFunc *Block            // the <%: func ... %> declaration (the generated function's header), or nil
	rawCode    string            // verbatim code collected from <%! ... %> tags, emitted before the function body
	fpath      string            // path of the file currently being parsed; used to resolve relative ~ and + directives
	extends    map[string]*Block // named @-blocks known so far (seeded from the ~ layout, plus any new ones); used to route child content into the layout's slots
	current    *Block            // the @-block currently being filled, or nil when at the root level
}

// currentTarget returns the block that new children should be appended to.
// If we are inside an @-block, it's that block's Children. Otherwise, it's
// the root.
func (p *Parser) currentTarget() *Blocks {
	if p.current != nil {
		return &p.current.Children
	}
	return &p.root
}

// Parse parses a template file and returns the root block tree, the raw
// code, and the function definition.
func (p *Parser) Parse(fpath string) (Blocks, string, *Block) {
	p.text = readFile(fpath)
	p.root = Blocks{}
	p.fpath = fpath
	p.extends = map[string]*Block{}
	p.current = nil
	for {
		p.cursor = strings.Index(p.text, "<%")
		if p.cursor == -1 {
			p.currentTarget().addChild(&Block{
				Content: p.text,
				Type:    HTML,
			})
			break
		}
		p.currentTarget().addChild(&Block{
			Content: p.text[:p.cursor],
			Type:    HTML,
		})
		for {
			p.text = p.text[p.cursor+2:]
			p.endCursor = strings.Index(p.text, "%>")
			if p.endCursor == -1 {
				panic("Syntax error")
			}
			if Goh.CountByte(p.text[:p.endCursor], '`')%2 != 0 || Goh.CountByte(p.text[:p.endCursor], '"')%2 != 0 {
				p.cursor = p.endCursor
				continue
			}
			p.parse()
			break
		}
		if 2 >= len(p.text) {
			p.currentTarget().addChild(&Block{
				Content: p.text,
				Type:    HTML,
			})
			break
		}
	}
	if p.current != nil {
		panic("Unclosed extend, please check the syntax!\n\033[0;33mWarning:\033[0m This syntax is different from the Hero, please check Goh's documentation: https://github.com/OblivionOcean/Goh#syntax")
	}
	return p.root, p.rawCode, p.defineFunc
}

func (p *Parser) parse() {
	switch p.text[0] {
	case '=', '-':
		p.pHTML()
	case '#':
	case '+':
		p.pInclude()
	case ':':
		p.defineFunc = &Block{
			Content: strings.Trim(p.text[1:p.endCursor], "\n\t\r "),
		}
	case '!':
		p.rawCode += p.text[1:p.endCursor] + "\n"
	case '~':
		p.addParent()
	case '@':
		p.useBlock()
	default:
		p.currentTarget().addChild(&Block{
			Type:    Code,
			Content: p.text[:p.endCursor],
		})
	}
	p.text = p.text[p.endCursor+2:]
}

func (p *Parser) pHTML() {
	block := &Block{
		Type: Escape,
	}
	if p.text[0] == '-' {
		block.Type = Val
		p.text = p.text[1:]
		p.endCursor -= 1
	} else if p.text[1] == '=' {
		block.Type = Val
		p.text = p.text[2:]
		p.endCursor -= 2
	} else {
		p.text = p.text[1:]
		p.endCursor -= 1
	}
	// A type code (e.g. "i ", "s ", "bs ") is only present when the body is
	// long enough to hold the code plus a separating space. Guard the indexes
	// so a short body like "a" (from `<%=a %>`) does not panic.
	if len(p.text) >= 2 && (p.text[1] == ' ' || (len(p.text) >= 3 && p.text[2] == ' ')) {
		switch p.text[0] {
		case 'b':
			if p.text[1] == 's' {
				block.VarType = Bytes
			} else {
				block.VarType = Bool
			}
		case 'u':
			block.VarType = Uint
		case 'i':
			block.VarType = Int
		case 's':
			block.VarType = String
		case 'f':
			block.VarType = Float
		case 'v', 'a':
			block.VarType = Any
		default:
			block.VarType = String
		}
		if p.text[1] == 's' {
			p.text = p.text[2:]
			p.endCursor -= 2
		} else {
			p.text = p.text[1:]
			p.endCursor -= 1
		}
	} else {
		block.VarType = String
	}
	block.Content = p.text[:p.endCursor]
	p.currentTarget().addChild(block)
}

func (p *Parser) pInclude() {
	fpath := filepath.Join(filepath.Dir(p.fpath), strings.Trim(p.text[1:p.endCursor], " \"\t\n\r"))
	np := Parser{}
	tmp, rawCode, _ := np.Parse(fpath)
	p.rawCode += rawCode
	for i := 0; i < len(tmp); i++ {
		p.currentTarget().addChild(tmp[i])
	}
}

// addParent parses the layout and splices its root into the child's root.
// The layout's blocks are appended to the end of the child's root (the child
// has only the function definition so far, which is not a block in the root).
// The layout's @-blocks are registered in the extends map so the child can
// override them.
func (p *Parser) addParent() {
	fpath := filepath.Join(filepath.Dir(p.fpath), strings.Trim(p.text[1:p.endCursor], " \"\t\n\r"))
	np := &Parser{}
	np.Parse(fpath)
	for i := 0; i < len(np.root); i++ {
		p.root.addChild(np.root[i])
	}
	// Register the layout's @-blocks in the extends map.
	// Walk the layout's root and find all Extend blocks.
	var walk func(blocks Blocks)
	walk = func(blocks Blocks) {
		for _, b := range blocks {
			if b.Type == Extend {
				if _, ok := p.extends[b.Content]; !ok {
					p.extends[b.Content] = b
				}
			}
			walk(b.Children)
		}
	}
	walk(np.root)
}

// useBlock handles @-block open and close.
//   - Open: if the block is in the extends map (defined by the layout), we
//     enter that block (set p.current to it). The child's content will be
//     appended to the block's Children. If the block is not in the extends
//     map, we create a new Extend block, append it to the current target,
//     and enter it.
//   - Close: we exit the current block (set p.current back to the parent).
func (p *Parser) useBlock() {
	name := strings.Trim(p.text[1:p.endCursor], " \"\t\n\r{")
	if name == "}" {
		if p.current == nil {
			panic("Unclosed extend, please check the syntax!\n\033[0;33mWarning:\033[0m This syntax is different from the Hero, please check Goh's documentation: https://github.com/OblivionOcean/Goh#syntax")
		}
		p.current = nil
		return
	}
	if block, ok := p.extends[name]; ok {
		// The block is defined by the layout. Enter it so the child's
		// content is appended to the block's Children.
		p.current = block
		return
	}
	// The block is not defined by the layout. Create a new Extend block,
	// append it to the current target, and enter it.
	block := &Block{
		Type:    Extend,
		Content: name,
	}
	p.currentTarget().addChild(block)
	p.extends[name] = block
	p.current = block
}
