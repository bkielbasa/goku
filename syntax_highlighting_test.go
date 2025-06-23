package main

import (
	"strings"
	"testing"
)

func TestSyntaxHighlighting(t *testing.T) {
	// Create a simple Go code snippet
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	// Create a buffer with the code
	style := newEditorStyle()
	b := newBuffer(style, bufferWithContent("test.go", code))

	// Test that the parser is set up correctly
	if b.parser == nil {
		t.Fatal("Parser should be initialized for .go files")
	}

	if b.language == nil {
		t.Fatal("Language should be set for .go files")
	}

	// Test highlighting a line
	highlighted := b.HighlightLine(0) // First line: "package main"
	
	// The line should contain highlighting (lipgloss styles)
	if !strings.Contains(highlighted, "\x1b[") {
		t.Error("Highlighted line should contain ANSI color codes")
	}

	// Test that keywords are highlighted
	if !strings.Contains(highlighted, "package") {
		t.Error("Keyword 'package' should be present in highlighted line")
	}
}

func TestTokenStyleMapping(t *testing.T) {
	style := newEditorStyle()

	// Test keyword highlighting
	keywordStyle := style.GetTokenStyle("package")
	if keywordStyle.String() == "" {
		t.Error("Keyword style should not be empty")
	}

	// Test string highlighting
	stringStyle := style.GetTokenStyle("string_literal")
	if stringStyle.String() == "" {
		t.Error("String style should not be empty")
	}

	// Test comment highlighting
	commentStyle := style.GetTokenStyle("comment")
	if commentStyle.String() == "" {
		t.Error("Comment style should not be empty")
	}
} 