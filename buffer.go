package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bkielbasa/goku/normalmode"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
	html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	json "github.com/tree-sitter/tree-sitter-json/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

type bufferState string

const (
	bufferStateUnnamed  bufferState = "unnamed"
	bufferStateModified bufferState = "modified"
	bufferStateSaved    bufferState = "saved"
	bufferStateReadOnly bufferState = "readonly"
)

type buffer struct {
	state                        bufferState
	lines                        []string
	filename                     string
	cursorX, cursorY             int
	cursorXOffset, cursorYOffset int
	viewport                     tea.WindowSizeMsg

	style editorStyle

	parser   *tree_sitter.Parser
	language *tree_sitter.Language
}

type newBufferOps func(b *buffer)

func bufferStateSavedOpt(b *buffer) {
	b.state = bufferStateSaved
}

func bufferWithContent(f, c string) func(b *buffer) {
	return func(b *buffer) {
		b.lines = strings.Split(c, "\n")
		b.filename = f
	}
}

func newBuffer(style editorStyle, ops ...newBufferOps) buffer {
	b := buffer{
		state:    bufferStateUnnamed,
		lines:    []string{""},
		viewport: tea.WindowSizeMsg{},
		style:    style,
	}

	var parser *tree_sitter.Parser

	for _, f := range ops {
		f(&b)
	}

	if b.filename != "" {
		lang := detectLanguage(b.filename)
		if lang != nil {
			parser = tree_sitter.NewParser()
			parser.SetLanguage(lang)

			b.parser = parser
			b.language = lang
		}
	}

	return b
}

func detectLanguage(filename string) *tree_sitter.Language {
	parts := strings.Split(filename, ".")
	ext := parts[len(parts)-1]

	switch ext {
	case "go":
		return tree_sitter.NewLanguage(golang.Language())
	case "c", "h":
		return tree_sitter.NewLanguage(c.Language())
	case "cpp", "cxx", "hpp":
		return tree_sitter.NewLanguage(cpp.Language())
	case "html":
		return tree_sitter.NewLanguage(html.Language())
	case "java":
		return tree_sitter.NewLanguage(java.Language())
	case "json":
		return tree_sitter.NewLanguage(json.Language())
	case "py":
		return tree_sitter.NewLanguage(python.Language())
	case "rb":
		return tree_sitter.NewLanguage(ruby.Language())
	case "rs":
		return tree_sitter.NewLanguage(rust.Language())
	}

	return nil
}

func (b buffer) SetStateModified() normalmode.Buffer {
	b.state = bufferStateModified
	return b
}

func (b buffer) SetStateSaved() normalmode.Buffer {
	b.state = bufferStateSaved
	return b
}

func (m buffer) View() string {
	var b strings.Builder

	startY := m.cursorYOffset
	endY := startY + m.viewport.Height - 2
	endY = min(endY, len(m.lines))

	for y := startY; y < endY; y++ {
		line := m.lines[y]
		visual := expandTabs(line)
		b.WriteString(lineNumber(y+1, len(m.lines)))

		// Apply horizontal scrolling
		lineNumberWidth := len(fmt.Sprintf("%d", len(m.lines))) + 1
		availableWidth := m.viewport.Width - lineNumberWidth
		
		// Trim the line based on horizontal offset
		startX := m.cursorXOffset
		endX := startX + availableWidth
		
		if startX < len(visual) {
			if endX > len(visual) {
				endX = len(visual)
			}
			visual = visual[startX:endX]
		} else {
			visual = ""
		}

		styledChunks := m.HighlightString(visual)

		if y == m.cursorY {
			visX := visualCursorX(m.lines[y], m.cursorX) - m.cursorXOffset
			renderedCursor := false
			currentCol := 0

			for _, chunk := range styledChunks {
				chunkWidth := runewidth.StringWidth(chunk.Content)
				if !renderedCursor && visX >= currentCol && visX < currentCol+chunkWidth {
					// The cursor is in this chunk
					offsetInChunk := visX - currentCol
					var before, at, after strings.Builder
					w := 0
					found := false
					for _, r := range chunk.Content {
						runeW := runewidth.RuneWidth(r)
						if !found && w >= offsetInChunk {
							at.WriteRune(r)
							found = true
						} else if !found {
							before.WriteRune(r)
						} else {
							after.WriteRune(r)
						}
						w += runeW
					}
					b.WriteString(chunk.Style.Render(before.String()))
					b.WriteString(m.style.cursor.Render(chunk.Style.Render(at.String())))
					b.WriteString(chunk.Style.Render(after.String()))
					renderedCursor = true
				} else {
					b.WriteString(chunk.Style.Render(chunk.Content))
				}
				currentCol += chunkWidth
			}

			if !renderedCursor && visX >= 0 && visX < availableWidth {
				b.WriteString(m.style.cursor.Render(" "))
			}
		} else {
			for _, chunk := range styledChunks {
				b.WriteString(chunk.Style.Render(chunk.Content))
			}
		}
		b.WriteRune('\n')
	}

	return b.String()
}

func (b buffer) Viewport() tea.WindowSizeMsg {
	return b.viewport
}

func (b buffer) CursorX() int {
	return b.cursorX
}

func (b buffer) SetCursorX(n int) normalmode.Buffer {
	b.cursorX = n
	return b.adjustViewportForCursor()
}

func (b buffer) SetCursorY(n int) normalmode.Buffer {
	b.cursorY = n
	return b.adjustViewportForCursor()
}

func (b buffer) IncreaseCursorX(n int) normalmode.Buffer {
	b.cursorX += n
	if b.cursorX < 0 {
		b.cursorX = 0
	}
	return b.adjustViewportForCursor()
}

func (b buffer) IncreaseCursorY(n int) normalmode.Buffer {
	b.cursorY += n
	if b.cursorY < 0 {
		b.cursorY = 0
	}

	return b.adjustViewportForCursor()
}

func (b buffer) IncreaseCursorYOffset(n int) normalmode.Buffer {
	b.cursorYOffset += n
	if b.cursorYOffset < 0 {
		b.cursorYOffset = 0
	}

	return b
}

func (b buffer) IncreaseCursorXOffset(n int) normalmode.Buffer {
	b.cursorXOffset += n
	if b.cursorXOffset < 0 {
		b.cursorXOffset = 0
	}

	return b
}

// adjustViewportForCursor ensures the cursor stays within the viewport
func (b buffer) adjustViewportForCursor() normalmode.Buffer {
	// Vertical viewport adjustment
	// If cursor is above the viewport, scroll up
	if b.cursorY < b.cursorYOffset {
		b.cursorYOffset = b.cursorY
	}
	
	// If cursor is below the viewport, scroll down
	// The -2 accounts for status bar and other UI elements
	if b.cursorY >= b.cursorYOffset+b.viewport.Height-2 {
		b.cursorYOffset = b.cursorY - (b.viewport.Height - 3)
	}
	
	// Ensure viewport doesn't go below 0
	if b.cursorYOffset < 0 {
		b.cursorYOffset = 0
	}

	// Horizontal viewport adjustment
	// Calculate the visual cursor position (accounting for tabs)
	line := b.Line(b.cursorY)
	visualX := visualCursorX(line, b.cursorX)
	
	// Account for line numbers and padding
	lineNumberWidth := len(fmt.Sprintf("%d", len(b.lines))) + 1
	availableWidth := b.viewport.Width - lineNumberWidth
	
	// If cursor is to the left of the viewport, scroll left
	if visualX < b.cursorXOffset {
		b.cursorXOffset = visualX
	}
	
	// If cursor is to the right of the viewport, scroll right
	if visualX >= b.cursorXOffset+availableWidth {
		b.cursorXOffset = visualX - availableWidth + 1
	}
	
	// Ensure viewport doesn't go below 0
	if b.cursorXOffset < 0 {
		b.cursorXOffset = 0
	}
	
	return b
}

func (b buffer) Line(n int) string {
	if n >= 0 && n < len(b.lines) {
		return b.lines[n]
	}
	return ""
}

func (b buffer) Lines() []string {
	return b.lines
}

func (b buffer) NoOfLines() int {
	return len(b.lines)
}

func (b buffer) CursorXOffset() int {
	return b.cursorXOffset
}

func (b buffer) CursorY() int {
	return b.cursorY
}

func (b buffer) CursorYOffset() int {
	return b.cursorYOffset
}

func (b buffer) AppendLine(s string) normalmode.Buffer {
	b.lines = append(b.lines, s)
	return b
}

func (b buffer) InsertLine(n int, s string) normalmode.Buffer {
	// Ensure n is within bounds
	if n < 0 {
		n = 0
	}
	if n > len(b.lines) {
		n = len(b.lines)
	}
	
	// Insert the line at position n
	b.lines = append(b.lines[:n], append([]string{s}, b.lines[n:]...)...)
	return b
}

func (b buffer) DeleteLine(n int) normalmode.Buffer {
	// Ensure n is within bounds
	if n < 0 || n >= len(b.lines) {
		return b
	}
	
	// Remove the line at position n
	b.lines = append(b.lines[:n], b.lines[n+1:]...)
	
	// Ensure we always have at least one line
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	
	return b
}

func (b buffer) ReplaceLine(n int, s string) normalmode.Buffer {
	b.lines[n] = s
	return b
}

func (b buffer) FileName() string {
	return b.filename
}

func (b buffer) SetFileName(f string) normalmode.Buffer {
	b.filename = f
	return b
}

func lineNumber(n int, total int) string {
	width := len(fmt.Sprintf("%d", total))
	return fmt.Sprintf("%*d ", width, n)
}

const tabSize = 4

func expandTabs(s string) string {
	var b strings.Builder
	col := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		}

		if inEscape {
			b.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		if r == '\t' {
			spaces := tabSize - (col % tabSize)
			b.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		} else {
			b.WriteRune(r)
			col++
		}
	}
	return b.String()
}

func visualCursorX(s string, logicalX int) int {
	col := 0
	for i := 0; i < logicalX && i < len(s); i++ {
		if s[i] == '\t' {
			col += tabSize - (col % tabSize)
		} else {
			col++
		}
	}
	return col
}

func loadFile(filename string, style editorStyle) (buffer, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return buffer{}, err
	}

	// Split content into lines, but preserve empty files
	lines := strings.Split(string(content), "\n")

	// Remove trailing empty line if file ends with newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Ensure we always have at least one line
	if len(lines) == 0 {
		lines = []string{""}
	}

	b := newBuffer(style, bufferWithContent(filename, string(content)))

	return b, nil
}

func findByteIndexForVisualColumn(s string, col int) int {
	var visualCol int = 0
	inEscape := false
	for i, r := range s {
		if r == '\x1b' {
			inEscape = true
		}

		if !inEscape {
			if visualCol >= col {
				return i
			}
			visualCol += runewidth.RuneWidth(r)
		}

		if inEscape && r == 'm' {
			inEscape = false
		}
	}
	return len(s)
}
