package main

import (
	"io"
	"os"
	"path"
	"strings"

	Goh "github.com/OblivionOcean/Goh/utils"
)

var Cache = map[string]*Block{}

const (
	Escape = iota
	Val
	HTML
	Code
	rawCode
)

const (
	Bool = iota
	Int
	Uint
	Float
	String
	Bytes
	Any
)

type Block struct {
	Type     int
	Children []*Block
	Parent   *Block
	Content  string
	VarType  int
}

type Parser struct {
	text       string
	cursor     int
	endCursor  int
	root       Blocks
	defindFunc *Block
	rawCode    string
	fpath      string
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

// Parse Template File
func (p *Parser) Parse(fpath string) (Blocks, string, *Block) {
	p.text = readFile(fpath)
	p.root = Blocks{}
	p.fpath = fpath
	for {
		p.cursor = strings.Index(p.text, "<%")
		if p.cursor == -1 {
			p.root.addChild(&Block{
				Content: p.text,
				Type:    HTML,
			})
			break
		}
		p.root.addChild(&Block{
			Content: p.text[:p.cursor],
			Type:    HTML,
		})
		for {
			p.text = p.text[p.cursor+2:]
			p.endCursor = strings.Index(p.text, "%>")
			if p.cursor == -1 {
				panic("Syntax error")
			}
			if Goh.CountByte(p.text[:p.endCursor], '`')%2 != 0 || Goh.CountByte(p.text[:p.endCursor], '"')%2 != 0 {
				p.cursor = p.endCursor
				continue
			} else {
				switch p.text[0] {
				case '=', '-':
					p.pHTML()
				case '#':
				case '+':
					p.pInclude()
				case ':':
					p.defindFunc = &Block{
						Content: strings.Trim(p.text[1:p.endCursor], "\n\t\r "),
					}
				case '!':
					p.rawCode += p.text[1:p.endCursor] + "\n"
				default:
					p.root.addChild(&Block{
						Type:    Code,
						Content: p.text[:p.endCursor],
					})
				}
				p.text = p.text[p.endCursor+2:]
				break
			}
		}
		if 2 >= len(p.text) {
			p.root.addChild(&Block{
				Content: p.text,
				Type:    HTML,
			})
			break
		}
	}
	return p.root, p.rawCode, p.defindFunc
}

func readFile(fpath string) string {
	file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	tmp, err := io.ReadAll(file)
	if err != nil {
		panic(err.Error())
	}
	return Goh.Byte2String(tmp)
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
	if p.text[1] == ' ' || p.text[2] == ' ' {
		switch p.text[0] { // 使用switch，底层是cmp比较，不是树遍历，速度更快
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
	p.root.addChild(block)
}

func (p *Parser) pInclude() {
	fpath := path.Join(path.Dir(p.fpath), strings.Trim(p.text[1:p.endCursor], " \"\t\n\r"))
	tmp, rawCode, _ := p.Parse(fpath)
	rawCode += rawCode
	for i := 0; i < len(tmp); i++ {
		p.root.addChild(tmp[i])
	}
}
