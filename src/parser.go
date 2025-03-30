package Goh

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	Goh "github.com/OblivionOcean/Goh/utils"
)

// Cache is a map that stores blocks by their string keys for quick retrieval.
var Cache = map[string]*Block{}

// Constants defining different types for text processing, including escape, value, HTML, code, raw code, and extend.
const (
	TypeEscape = iota
	TypeValue
	TypeHTML
	TypeCode
	typeRawCode
	TypeExtend
)

// Constants defining types for variables, including boolean, integer, unsigned integer, float, string, bytes, and any.
const (
	VarTypeBool = iota
	VarTypeInt
	VarTypeUint
	VarTypeFloat
	VarTypeString
	VarTypeBytes
	VarTypeAny
)

// Block represents a segment of content in a template, including its type, content, and relationships with other blocks.
type Block struct {
	BlockType       int
	Content         string
	VariableType    int
	PreviousBlocks  Blocks
	FollowingBlocks Blocks
	FollowingCount  int
}

// Parser is a structure for parsing template files, managing blocks, and handling the content and logic within.
type Parser struct {
	sourceText     string
	currentCursor  int
	endCursor      int
	rootBlocks     Blocks
	definedFunc    *Block
	rawCodeContent string
	filePath       string
	extendedBlocks map[string]*Block

	// extendList stores a list of Block pointers that are currently being extended in the template.
	extendList      []*Block
	unclosedExtends int
}

// Blocks is a slice of Block pointers, used to manage and organize multiple blocks in a template.
type Blocks []*Block

// addChild appends a child block to the Blocks, merging HTML content if the last block is also HTML.
func (b *Blocks) addChild(child *Block) {
	if len(*b) > 0 && child.BlockType == TypeHTML && (*b)[len(*b)-1].BlockType == TypeHTML {
		if len((*b)[len(*b)-1].Content) > 0 {
			(*b)[len(*b)-1].Content += child.Content
		} else {
			(*b)[len(*b)-1] = child
		}
	} else {
		*b = append(*b, child)
	}
}

// Parse processes the given template file, extracting and organizing blocks, and returns the root blocks, raw code content, and defined function.
func (p *Parser) Parse(filePath string) (Blocks, string, *Block) {
	p.filePath = filePath
	p.sourceText = readTemplateFile(filePath)
	p.rootBlocks = nil
	p.extendedBlocks = make(map[string]*Block)

	for {
		if nextTemplateStart := strings.Index(p.sourceText, "<%"); nextTemplateStart != -1 {
			handleNonTemplatePart(p, nextTemplateStart)
			p.parseTemplateTag()
		} else {
			finalizeParsing(p)
			break
		}
	}

	checkForUnclosedExtends(p)
	mergeExtendedBlocks(p)

	return p.rootBlocks, p.rawCodeContent, p.definedFunc
}

// handleNonTemplatePart adds a non-template part of the source text as an HTML block to the root blocks.
func handleNonTemplatePart(p *Parser, nextTemplateStart int) {
	p.rootBlocks.addChild(&Block{
		Content:   p.sourceText[:nextTemplateStart],
		BlockType: TypeHTML,
	})
}

// parseTemplateTag processes a template tag, updating the Parser's state and handling different types of tags.
func (p *Parser) parseTemplateTag() {
	p.sourceText = p.sourceText[p.currentCursor+2:]
	p.endCursor = strings.Index(p.sourceText, "%>")
	if p.currentCursor == -1 {
		panic("Syntax error")
	}
	if Goh.CountByte(p.sourceText[:p.endCursor], '`')%2 != 0 || Goh.CountByte(p.sourceText[:p.endCursor], '"')%2 != 0 {
		p.currentCursor = p.endCursor
		return
	}

	// Define a map to handle different template tag types
	var handlers = map[byte]func(*Parser){
		'=': handleEqual,
		'-': handleDash,
		'#': handleHash,
		'+': handlePlus,
		':': handleColon,
		'!': handleExclamation,
		'~': handleTilde,
		'@': handleAt,
	}

	// Get the handler function based on the first character of the source text
	if handler, exists := handlers[p.sourceText[0]]; exists {
		handler(p)
	} else {
		// Default case: add a code block
		p.rootBlocks.addChild(&Block{
			BlockType: TypeCode,
			Content:   p.sourceText[:p.endCursor],
		})
	}

	// Move the source text cursor past the current tag
	p.sourceText = p.sourceText[p.endCursor+2:]
}

// handleEqual handles the '=' tag
func handleEqual(p *Parser) {
	p.parseHTML()
}

// handleDash handles the '-' tag
func handleDash(p *Parser) {
	// Implement logic for '-' tag
}

// handleHash handles the '#' tag
func handleHash(p *Parser) {
	// Implement logic for '#' tag
}

// handlePlus handles the '+' tag
func handlePlus(p *Parser) {
	p.parseIncludeFile()
}

// handleColon handles the ':' tag
func handleColon(p *Parser) {
	p.definedFunc = &Block{
		Content: strings.Trim(p.sourceText[1:p.endCursor], "\n\t\r "),
	}
}

// handleExclamation handles the '!' tag
func handleExclamation(p *Parser) {
	p.rawCodeContent += p.sourceText[1:p.endCursor] + "\n"
}

// handleTilde handles the '~' tag
func handleTilde(p *Parser) {
	p.addParentBlock()
}

// handleAt handles the '@' tag
func handleAt(p *Parser) {
	p.useBlockDefinition()
}

// finalizeParsing adds the remaining source text as an HTML block to the rootBlocks if the source text length is at least 2.
func finalizeParsing(p *Parser) {
	if len(p.sourceText) >= 2 {
		p.rootBlocks.addChild(&Block{
			Content:   p.sourceText,
			BlockType: TypeHTML,
		})
	}
}

// checkForUnclosedExtends verifies if there are any unclosed extends in the template and panics if found.
func checkForUnclosedExtends(p *Parser) {
	if p.unclosedExtends > 0 {
		panic("Unclosed extend, please check the syntax!\n\033[0;33mWarning:\033[0m This syntax is different from the Hero, please check Goh's documentation: https://github.com/OblivionOcean/Goh#syntax")
	}
}

// mergeExtendedBlocks merges the following blocks of each extended block into its FollowingBlocks and resets FollowingCount.
func mergeExtendedBlocks(p *Parser) {
	for _, block := range p.extendedBlocks {
		if block.FollowingCount > 0 {
			block.FollowingBlocks = append(block.FollowingBlocks, p.rootBlocks[block.FollowingCount:]...)
		}
		block.FollowingCount = 0
	}
}

// readTemplateFile reads the content of a template file from the given file path and returns it as a string.
func readTemplateFile(filePath string) string {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err).Error())
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("failed to read file: %w", err).Error())
	}

	return Goh.Byte2String(data)
}

// parseHTML processes the HTML content, determines the block type and variable type, and adds the block to the root blocks.
func (p *Parser) parseHTML() {
	block := &Block{
		BlockType: TypeEscape,
	}

	// Determine the block type
	if p.sourceText[0] == '-' || (p.sourceText[0] == '=' && p.sourceText[1] == '=') {
		block.BlockType = TypeValue
		if p.sourceText[0] == '-' {
			p.sourceText = p.sourceText[1:]
			p.endCursor -= 1
		} else {
			p.sourceText = p.sourceText[2:]
			p.endCursor -= 2
		}
	} else {
		p.sourceText = p.sourceText[1:]
		p.endCursor -= 1
	}

	// Determine the variable type
	if p.sourceText[1] == ' ' || p.sourceText[2] == ' ' {
		varTypeHandlers := map[byte]func(*Parser, *Block){
			'b': handleB,
			'u': handleU,
			'i': handleI,
			's': handleS,
			'f': handleF,
			'v': handleV,
			'a': handleA,
		}

		if handler, exists := varTypeHandlers[p.sourceText[0]]; exists {
			handler(p, block)
		} else {
			block.VariableType = VarTypeString
		}

		if p.sourceText[1] == 's' {
			p.sourceText = p.sourceText[2:]
			p.endCursor -= 2
		} else {
			p.sourceText = p.sourceText[1:]
			p.endCursor -= 1
		}
	} else {
		block.VariableType = VarTypeString
	}

	// Set the block content and add it to the root blocks
	block.Content = p.sourceText[:p.endCursor]
	p.rootBlocks.addChild(block)
}

// handleB handles the 'b' character for variable type
func handleB(p *Parser, block *Block) {
	if p.sourceText[1] == 's' {
		block.VariableType = VarTypeBytes
	} else {
		block.VariableType = VarTypeBool
	}
}

// handleU handles the 'u' character for variable type
func handleU(p *Parser, block *Block) {
	block.VariableType = VarTypeUint
}

// handleI handles the 'i' character for variable type
func handleI(p *Parser, block *Block) {
	block.VariableType = VarTypeInt
}

// handleS handles the 's' character for variable type
func handleS(p *Parser, block *Block) {
	block.VariableType = VarTypeString
}

// handleF handles the 'f' character for variable type
func handleF(p *Parser, block *Block) {
	block.VariableType = VarTypeFloat
}

// handleV handles the 'v' character for variable type
func handleV(p *Parser, block *Block) {
	block.VariableType = VarTypeAny
}

// handleA handles the 'a' character for variable type
func handleA(p *Parser, block *Block) {
	block.VariableType = VarTypeAny
}

// parseIncludeFile processes an include directive, parses the included file, and merges its blocks and raw code into the current parser.
func (p *Parser) parseIncludeFile() {
	filePath := path.Join(path.Dir(p.filePath), strings.Trim(p.sourceText[1:p.endCursor], " \"\t\n\r"))
	newParser := Parser{}
	includedBlocks, rawCode, _ := newParser.Parse(filePath)
	p.rawCodeContent += rawCode
	for _, block := range includedBlocks {
		p.rootBlocks.addChild(block)
	}
}

// addParentBlock parses a parent template file and adds its extended blocks to the current parser's extendedBlocks.
func (p *Parser) addParentBlock() {
	filePath := path.Join(path.Dir(p.filePath), strings.Trim(p.sourceText[1:p.endCursor], " \"\t\n\r"))
	newParser := &Parser{}
	newParser.Parse(filePath)
	for name, block := range newParser.extendedBlocks {
		if _, exists := p.extendedBlocks[name]; !exists {
			p.extendedBlocks[name] = block
		}
	}
}

// useBlockDefinition processes the block definition, handling extends and adding blocks to the root or extended blocks.
func (p *Parser) useBlockDefinition() {
	// Trim the name once
	name := strings.Trim(p.sourceText[1:p.endCursor], " \"\t\n\r{")

	if name == "}" {
		if len(p.extendList) == 0 {
			panic("Unclosed extend, please check the syntax!\n\033[0;33mWarning:\033[0m This syntax is different from the Hero, please check Goh's documentation: https://github.com/OblivionOcean/Goh#syntax")
		}

		block := p.extendList[len(p.extendList)-1]
		if exists := p.extendedBlocks[block.Content]; exists != nil {
			for _, afterBlock := range block.FollowingBlocks {
				p.rootBlocks.addChild(afterBlock)
			}
		} else {
			p.rootBlocks.addChild(block)
			block.FollowingCount = len(p.rootBlocks)
			p.extendedBlocks[block.Content] = block
		}

		p.unclosedExtends--
		p.extendList = p.extendList[:len(p.extendList)-1]
		return
	}

	if block, exists := p.extendedBlocks[name]; exists {
		p.unclosedExtends++
		for _, beforeBlock := range block.PreviousBlocks {
			p.rootBlocks.addChild(beforeBlock)
		}
		p.extendList = append(p.extendList, block)
		return
	}

	p.unclosedExtends++
	block := &Block{
		BlockType:      TypeExtend,
		Content:        name,
		PreviousBlocks: p.rootBlocks[:len(p.rootBlocks)],
	}
	p.rootBlocks.addChild(block)
	p.extendList = append(p.extendList, block)
}
