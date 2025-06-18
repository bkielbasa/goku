package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
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
		lines:         []string{"line 1", "line 2", "line 3", "line 4", "line 5", "line 6", "line 7", "line 8", "line 9", "line 10", "line 11", "line 12", "line 13", "line 14", "line 15", "line 16", "line 17", "line 18", "line 19", "line 20", "line 21", "line 22", "line 23", "line 24", "line 25", "line 26", "line 27", "line 28", "line 29", "line 30", "line 31", "line 32", "line 33", "line 34", "line 35", "line 36", "line 37", "line 38", "line 39", "line 40", "line 41", "line 42", "line 43", "line 44", "line 45", "line 46", "line 47", "line 48", "line 49", "line 50", "line 51", "line 52", "line 53", "line 54", "line 55", "line 56", "line 57", "line 58", "line 59", "line 60", "line 61", "line 62", "line 63", "line 64", "line 65", "line 66", "line 67", "line 68", "line 69", "line 70", "line 71", "line 72", "line 73", "line 74", "line 75", "line 76", "line 77", "line 78", "line 79", "line 80", "line 81", "line 82", "line 83", "line 84", "line 85", "line 86", "line 87", "line 88", "line 89", "line 90", "line 91", "line 92", "line 93", "line 94", "line 95", "line 96", "line 97", "line 98", "line 99", "line 100"},
		state:         bufferStateUnnamed,
		cursorY:       0,
		cursorYOffset: 0,
		viewport:      tea.WindowSizeMsg{},
		style:         style,
	}

	for _, f := range ops {
		f(&b)
	}

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
