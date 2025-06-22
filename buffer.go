package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bkielbasa/goku/normalmode"
	tea "github.com/charmbracelet/bubbletea"
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
		state:         bufferStateUnnamed,
		cursorY:       0,
		lines:         []string{""},
		cursorYOffset: 0,
		viewport:      tea.WindowSizeMsg{},
		style:         style,
	}

	for _, f := range ops {
		f(&b)
	}

	return b
}

func (b buffer) SetStateModified() normalmode.Buffer {
	b.state = bufferStateModified
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

		if y == m.cursorY {
			visX := visualCursorX(line, m.cursorX)
			if visX > len(visual) {
				visX = len(visual)
			}
			b.WriteString(visual[:visX])

			if visX < len(visual) {
				b.WriteString(m.style.cursor.Render(string(visual[visX])))
				b.WriteString(visual[visX+1:])
			} else {
				b.WriteString(m.style.cursor.Render(" "))
			}
		} else {
			b.WriteString(visual)
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
	return b
}

func (b buffer) SetCursorY(n int) normalmode.Buffer {
	b.cursorY = n
	return b
}

func (b buffer) IncreaseCursorX(n int) normalmode.Buffer {
	b.cursorX += n
	if b.cursorX < 0 {
		b.cursorX = 0
	}
	return b
}

func (b buffer) IncreaseCursorY(n int) normalmode.Buffer {
	b.cursorY += n
	if b.cursorY < 0 {
		b.cursorY = 0
	}

	return b
}

func (b buffer) IncreaseCursorYOffset(n int) normalmode.Buffer {
	b.cursorYOffset += n
	if b.cursorYOffset < 0 {
		b.cursorYOffset = 0
	}

	return b
}

func (b buffer) Line(n int) string {
	return b.lines[n]
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
	for _, r := range s {
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

	b := buffer{
		state:         bufferStateSaved,
		lines:         lines,
		filename:      filename,
		cursorY:       0,
		cursorYOffset: 0,
		viewport:      tea.WindowSizeMsg{},
		style:         style,
	}

	return b, nil
}
