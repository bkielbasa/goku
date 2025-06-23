package main

import (
	"sort"
	"strings"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/charmbracelet/lipgloss"
)

// Token represents a highlighted token with its position and type
type Token struct {
	StartByte uint32
	EndByte   uint32
	Type      string
	Text      string
}

// SyntaxHighlighter handles syntax highlighting using tree-sitter
type SyntaxHighlighter struct {
	parser   *tree_sitter.Parser
	language *tree_sitter.Language
}

// NewSyntaxHighlighter creates a new syntax highlighter for the given language
func NewSyntaxHighlighter(lang *tree_sitter.Language) *SyntaxHighlighter {
	parser := tree_sitter.NewParser()
	parser.SetLanguage(lang)
	
	return &SyntaxHighlighter{
		parser:   parser,
		language: lang,
	}
}

// Highlight parses the content and returns a list of tokens with their types
func (sh *SyntaxHighlighter) Highlight(content string) []Token {
	if sh.parser == nil {
		return nil
	}

	tree := sh.parser.Parse([]byte(content), nil)
	root := tree.RootNode()
	
	var tokens []Token
	sh.collectTokens(root, content, &tokens)
	
	return tokens
}

// collectTokens recursively collects all tokens from the syntax tree
func (sh *SyntaxHighlighter) collectTokens(root *tree_sitter.Node, content string, tokens *[]Token) {
	if root == nil {
		return
	}
	var walk func(*tree_sitter.Node)
	walk = func(n *tree_sitter.Node) {
		if n == nil {
			return
		}
		startByte := uint32(n.StartByte())
		endByte := uint32(n.EndByte())
		if startByte < endByte {
			*tokens = append(*tokens, Token{
				StartByte: startByte, EndByte: endByte,
				Type: n.Kind(), Text: content[startByte:endByte],
			})
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(uint(i)))
		}
	}
	walk(root)
}

// GetTokenStyle returns the appropriate style for a given token type
func (s editorStyle) GetTokenStyle(tokenType string) lipgloss.Style {
	switch tokenType {
	// Go Keywords (anonymous nodes in the tree)
	case "package", "import", "func", "return", "if", "else", "for", "range", "go", "defer", "const", "var", "type", "struct", "interface", "map", "chan", "select", "switch", "case", "default", "break", "continue", "fallthrough", "goto":
		return s.keyword // orange

	// Specific Node Types from the Go Grammar
	case "comment":
		return s.comment
	case "interpreted_string_literal", "raw_string_literal", "string_literal", "interpreted_string_literal_content":
		return s.string
	case "int_literal", "float_literal":
		return s.number
	case "nil":
		return s.number // Use number style for nil constant
	case "field_identifier":
		return s.function // Use function style for fields
	case "type_identifier":
		return s.typeName
	case "package_identifier":
		return s.text // Default text for package names
	case "identifier":
		return s.function // Use function style for identifiers (makes functions yellow)
	case "escape_sequence":
		return s.string // Style escape sequences like the rest of the string

	default:
		return s.text
	}
}

// HighlightLine applies syntax highlighting to a single line
func (b buffer) HighlightLine(lineIndex int) string {
	if b.parser == nil {
		return b.lines[lineIndex]
	}

	// Get the content up to and including this line
	var content strings.Builder
	for i := 0; i <= lineIndex; i++ {
		if i > 0 {
			content.WriteString("\n")
		}
		content.WriteString(b.lines[i])
	}
	
	fullContent := content.String()
	
	// Create a highlighter and get tokens
	highlighter := NewSyntaxHighlighter(b.language)
	tokens := highlighter.Highlight(fullContent)
	
	// Find tokens that belong to this line
	lineStart := 0
	for i := 0; i < lineIndex; i++ {
		lineStart += len(b.lines[i]) + 1 // +1 for newline
	}
	lineEnd := lineStart + len(b.lines[lineIndex])
	
	// Apply highlighting to the line
	return b.applyHighlighting(b.lines[lineIndex], tokens, lineStart, lineEnd)
}

// applyHighlighting applies token highlighting to a line
func (b buffer) applyHighlighting(line string, tokens []Token, lineStart, lineEnd int) string {
	if len(tokens) == 0 {
		return line
	}

	var result strings.Builder
	currentPos := 0

	// Find tokens that overlap with this line
	for _, token := range tokens {
		tokenStart := int(token.StartByte)
		tokenEnd := int(token.EndByte)
		
		// Skip tokens that don't overlap with this line
		if tokenEnd <= lineStart || tokenStart >= lineEnd {
			continue
		}
		
		// Calculate the position within the line
		lineTokenStart := tokenStart - lineStart
		lineTokenEnd := tokenEnd - lineStart
		
		// Ensure bounds are within the line
		if lineTokenStart < 0 {
			lineTokenStart = 0
		}
		if lineTokenEnd > len(line) {
			lineTokenEnd = len(line)
		}
		
		// Add text before this token
		if lineTokenStart > currentPos {
			result.WriteString(line[currentPos:lineTokenStart])
		}
		
		// Add the highlighted token
		tokenText := line[lineTokenStart:lineTokenEnd]
		style := b.style.GetTokenStyle(token.Type)
		result.WriteString(style.Render(tokenText))
		
		currentPos = lineTokenEnd
	}
	
	// Add remaining text after the last token
	if currentPos < len(line) {
		result.WriteString(line[currentPos:])
	}
	
	return result.String()
}

type StyledChunk struct {
	Content string
	Style   lipgloss.Style
}

// HighlightString applies syntax highlighting to a given string and returns styled chunks.
func (b buffer) HighlightString(content string) []StyledChunk {
	if b.parser == nil {
		return []StyledChunk{{Content: content, Style: b.style.text}}
	}

	highlighter := NewSyntaxHighlighter(b.language)
	tokens := highlighter.Highlight(content)

	return b.applyHighlightingToLine(content, tokens)
}

// applyHighlightingToLine converts a list of tokens on a line to a series of styled chunks.
func (b buffer) applyHighlightingToLine(line string, tokens []Token) []StyledChunk {
	if len(line) == 0 {
		return nil
	}
	styles := make([]lipgloss.Style, len(line))
	for i := range styles {
		styles[i] = b.style.text
	}
	sort.SliceStable(tokens, func(i, j int) bool {
		if tokens[i].StartByte != tokens[j].StartByte {
			return tokens[i].StartByte < tokens[j].StartByte
		}
		return tokens[i].EndByte > tokens[j].EndByte
	})

	for _, token := range tokens {
		style := b.style.GetTokenStyle(token.Type)
		start := int(token.StartByte)
		end := int(token.EndByte)
		for i := start; i < end && i < len(line); i++ {
			styles[i] = style
		}
	}

	var chunks []StyledChunk
	if len(line) > 0 {
		var currentChunk strings.Builder
		currentStyle := styles[0]
		for i, r := range line {
			if styles[i].Render(" ") == currentStyle.Render(" ") {
				currentChunk.WriteRune(r)
			} else {
				chunks = append(chunks, StyledChunk{Content: currentChunk.String(), Style: currentStyle})
				currentChunk.Reset()
				currentChunk.WriteRune(r)
				currentStyle = styles[i]
			}
		}
		chunks = append(chunks, StyledChunk{Content: currentChunk.String(), Style: currentStyle})
	}
	return chunks
} 