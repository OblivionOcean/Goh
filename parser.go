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
	Content  string
	VarType  int
	Unuse    bool
}

func (b *Block) AddChild(child *Block) {
	if len(b.Children) > 0 && child.Type == HTML && b.Children[len(b.Children)-1].Type == HTML {
		if len(b.Children[len(b.Children)-1].Content) > 0 {
			b.Children[len(b.Children)-1].Content += child.Content
		} else {
			b.Children[len(b.Children)-1] = child
		}
	} else {
		b.Children = append(b.Children, child)
		child.Parent = b
	}
}

func Parse(fpath string) (*Block, string, *Block) {
	// 将结构体变为局部变量，优化汇编代码跳转花费的时间
	var (
		Text       string
		Curser     int
		EndCurser  int
		Root       *Block
		DefindFunc *Block
		RawCode    string
	)
	file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	tmp, err := io.ReadAll(file)
	Text = Goh.Byte2String(tmp)
	if err != nil {
		panic(err.Error())
	}
	Root = &Block{}
	for {
		Curser = strings.Index(Text, "<%")
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
			EndCurser = strings.Index(Text, "%>")
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
					fpath := path.Join(path.Dir(fpath), strings.Trim(Text[1:EndCurser], " \"\t\n\r"))
					tmp, rawCode, _ := Parse(fpath)
					RawCode += rawCode
					for i := 0; i < len(tmp.Children); i++ {
						Root.AddChild(tmp.Children[i])
					}
				case ':':
					DefindFunc = &Block{
						Content: strings.Trim(Text[1:EndCurser], "\n\t\r "),
					}
				case '!':
					RawCode += Text[1:EndCurser] + "\n"
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
				Type:    HTML,
			})
			break
		}
	}
	return Root, RawCode, DefindFunc
}
