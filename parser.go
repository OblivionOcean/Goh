package main

import (
	"github.com/OblivionOcean/Goh/utils"
	"bytes"
	"io"
	"os"
	"path"
)

var Cache = map[string]*Block{}

const (
	Escape = iota
	Val
	HTML
	Code
	RawCode
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
	Content  []byte
	VarType  int
	Unuse    bool
}

func (b *Block) AddChild(child *Block) {
	if len(b.Children) > 0 && child.Type == HTML && b.Children[len(b.Children)-1].Type == HTML {
		if len(b.Children[len(b.Children)-1].Content) > 0 {
			b.Children[len(b.Children)-1].Content = append(b.Children[len(b.Children)-1].Content, child.Content...)
		} else {
			b.Children[len(b.Children)-1] = child
		}
	} else {
		b.Children = append(b.Children, child)
		child.Parent = b
	}
}

func Parse(fpath string) (*Block, []byte, *Block) {
	// 将结构体变为局部变量，优化汇编代码跳转花费的时间
	var (
		Text       []byte
		Curser     int
		EndCurser  int
		Root       *Block
		DefindFunc *Block
		RawCode    []byte
	)
	RawCode = []byte{}
	file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	Text, err = io.ReadAll(file)
	if err != nil {
		panic(err.Error())
	}
	Root = &Block{}
	for {
		Curser = bytes.Index(Text, []byte{'<', '%'})
		if Curser == -1 {
			Root.AddChild(&Block{
				Content: Text,
				Type:    HTML,
			})
			break
		}
		Root.AddChild(&Block{
			Content: Text[:Curser],
			Type:    HTML,
		})
		for {
			Text = Text[Curser+2:]
			EndCurser = bytes.Index(Text, []byte{'%', '>'})
			if Curser == -1 {
				panic("Syntax error")
			}
			if Goh.CountByte(Text[:EndCurser], '`')%2 != 0 || Goh.CountByte(Text[:EndCurser], '"')%2 != 0 {
				Curser = EndCurser
				continue
			} else {
				switch Text[0] {
				case '=', '-':
					block := &Block{
						Type: Escape,
					}
					if Text[0] == '-' {
						block.Type = Val
						Text = Text[1:]
						EndCurser -= 1
					} else if Text[1] == '=' {
						block.Type = Val
						Text = Text[2:]
						EndCurser -= 2
					} else {
						Text = Text[1:]
						EndCurser -= 1
					}
					if Text[1] == ' ' || Text[2] == ' ' {
						switch Text[0] { // 使用switch，底层是cmp比较，不是树遍历，速度更快
						case 'b':
							if Text[1] == 's' {
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
						if Text[1] == 's' {
							Text = Text[2:]
							EndCurser -= 2
						} else {
							Text = Text[1:]
							EndCurser -= 1
						}
					} else {
						block.VarType = String
					}
					block.Content = Text[:EndCurser]
					Root.AddChild(block)
				case '#':
				case '+':
					fpath := path.Join(path.Dir(fpath), Goh.Byte2String(bytes.Trim(Text[1:EndCurser], " \"\t\n\r")))
					tmp, rawCode, _ := Parse(fpath)
					RawCode = append(RawCode, rawCode...)
					for i := 0; i < len(tmp.Children); i++ {
						Root.AddChild(tmp.Children[i])
					}
				case ':':
					DefindFunc = &Block{
						Content: bytes.Trim(Text[1:EndCurser], "\n\t\r "),
					}
				case '!':
					RawCode = append(RawCode, Text[1:EndCurser]...)
					RawCode = append(RawCode, '\n')
				default:
					Root.AddChild(&Block{
						Type:    Code,
						Content: Text[:EndCurser],
					})
				}
				Text = Text[EndCurser+2:]
				break
			}
		}
		if 2 >= len(Text) {
			Root.AddChild(&Block{
				Content: Text,
			})
			break
		}
	}
	return Root, RawCode, DefindFunc
}
