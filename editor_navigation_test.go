package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestBasicNavigation tests basic cursor movement (h, j, k, l)
func TestBasicNavigation(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		key      string
		expected struct {
			cursorX int
			cursorY int
		}
	}{
		{
			name:    "move right on single line",
			content: "hello world",
			key:     "l",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 1, cursorY: 0},
		},
		{
			name:    "move left on single line",
			content: "hello world",
			key:     "h",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 0, cursorY: 0},
		},
		{
			name:    "move down to next line",
			content: "line 1\nline 2",
			key:     "j",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 0, cursorY: 1},
		},
		{
			name:    "move up to previous line",
			content: "line 1\nline 2",
			key:     "k",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 0, cursorY: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}
			
			lines := strings.Split(tt.content, "\n")
			if len(lines) == 0 {
				lines = []string{""}
			}
			m.buffers[0].lines = lines
			m.buffers[0].viewport = m.viewport

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(keyMsg)

			buffer := newModel.(model).buffers[0]
			if buffer.cursorX != tt.expected.cursorX {
				t.Errorf("cursorX = %d, want %d", buffer.cursorX, tt.expected.cursorX)
			}
			if buffer.cursorY != tt.expected.cursorY {
				t.Errorf("cursorY = %d, want %d", buffer.cursorY, tt.expected.cursorY)
			}
		})
	}
}

// TestWordNavigation tests word-by-word navigation (w, b)
func TestWordNavigation(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		initialX int
		key      string
		expected struct {
			cursorX int
			cursorY int
		}
	}{
		{
			name:     "next word on same line",
			content:  "hello world test",
			initialX: 0,
			key:      "w",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 6, cursorY: 0},
		},
		{
			name:     "previous word on same line",
			content:  "hello world test",
			initialX: 10,
			key:      "b",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 6, cursorY: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}
			
			lines := strings.Split(tt.content, "\n")
			if len(lines) == 0 {
				lines = []string{""}
			}
			m.buffers[0].lines = lines
			m.buffers[0].cursorX = tt.initialX
			m.buffers[0].viewport = m.viewport

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(keyMsg)

			buffer := newModel.(model).buffers[0]
			if buffer.cursorX != tt.expected.cursorX {
				t.Errorf("cursorX = %d, want %d", buffer.cursorX, tt.expected.cursorX)
			}
			if buffer.cursorY != tt.expected.cursorY {
				t.Errorf("cursorY = %d, want %d", buffer.cursorY, tt.expected.cursorY)
			}
		})
	}
}

// TestBoundaryNavigation tests navigation at file boundaries
func TestBoundaryNavigation(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		initialX int
		initialY int
		key      string
		expected struct {
			cursorX int
			cursorY int
		}
	}{
		{
			name:     "move down at last line",
			content:  "line 1\nline 2\nline 3",
			initialX: 0,
			initialY: 2,
			key:      "j",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 0, cursorY: 0},
		},
		{
			name:     "move up at first line",
			content:  "line 1\nline 2\nline 3",
			initialX: 0,
			initialY: 0,
			key:      "k",
			expected: struct {
				cursorX int
				cursorY int
			}{cursorX: 0, cursorY: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}
			
			lines := strings.Split(tt.content, "\n")
			if len(lines) == 0 {
				lines = []string{""}
			}
			m.buffers[0].lines = lines
			m.buffers[0].cursorX = tt.initialX
			m.buffers[0].cursorY = tt.initialY
			m.buffers[0].viewport = m.viewport

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(keyMsg)

			buffer := newModel.(model).buffers[0]
			if buffer.cursorX != tt.expected.cursorX {
				t.Errorf("cursorX = %d, want %d", buffer.cursorX, tt.expected.cursorX)
			}
			if buffer.cursorY != tt.expected.cursorY {
				t.Errorf("cursorY = %d, want %d", buffer.cursorY, tt.expected.cursorY)
			}
		})
	}
}

// TestModeSwitching tests switching between different editor modes
func TestModeSwitching(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected editorMode
	}{
		{
			name:     "enter insert mode",
			key:      "i",
			expected: ModeInsert,
		},
		{
			name:     "enter command mode",
			key:      ":",
			expected: ModeCommand,
		},
		{
			name:     "escape clears buffer",
			key:      "esc",
			expected: ModeNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(keyMsg)

			resultModel := newModel.(model)
			if resultModel.mode != tt.expected {
				t.Errorf("mode = %s, want %s", resultModel.mode, tt.expected)
			}
		})
	}
} 