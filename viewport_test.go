package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewportAdjustment(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		initialCursorY int
		viewportHeight int
		action         func(m model) model
		expectedOffset int
	}{
		{
			name:           "cursor moves up above viewport",
			content:        strings.Repeat("line\n", 20),
			initialCursorY: 10,
			viewportHeight: 10,
			action: func(m model) model {
				// Move cursor up 5 lines
				for i := 0; i < 5; i++ {
					keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
					newModel, _ := m.Update(keyMsg)
					m = newModel.(model)
				}
				return m
			},
			expectedOffset: 2, // Should only scroll if cursor goes above viewport
		},
		{
			name:           "cursor moves down below viewport",
			content:        strings.Repeat("line\n", 20),
			initialCursorY: 5,
			viewportHeight: 10,
			action: func(m model) model {
				// Move cursor down 8 lines
				for i := 0; i < 8; i++ {
					keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
					newModel, _ := m.Update(keyMsg)
					m = newModel.(model)
				}
				return m
			},
			expectedOffset: 6, // Should scroll down: cursorY(13) - (viewportHeight(10) - 3) = 6
		},
		{
			name:           "go to beginning of file",
			content:        strings.Repeat("line\n", 20),
			initialCursorY: 15,
			viewportHeight: 10,
			action: func(m model) model {
				keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
				newModel, _ := m.Update(keyMsg)
				m = newModel.(model)
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
				newModel, _ = m.Update(keyMsg)
				m = newModel.(model)
				return m
			},
			expectedOffset: 0, // Should scroll to top
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.viewport = tea.WindowSizeMsg{Width: 80, Height: tt.viewportHeight}
			
			lines := strings.Split(strings.TrimSpace(tt.content), "\n")
			if len(lines) == 0 {
				lines = []string{""}
			}
			m.buffers[0].lines = lines
			m.buffers[0].cursorY = tt.initialCursorY
			m.buffers[0].viewport = m.viewport

			// Perform the action
			m = tt.action(m)

			// Check if viewport offset is correct
			actualOffset := m.buffers[0].cursorYOffset
			if actualOffset != tt.expectedOffset {
				t.Errorf("viewport offset = %d, want %d", actualOffset, tt.expectedOffset)
			}

			// Verify cursor is within viewport
			cursorY := m.buffers[0].cursorY
			viewportStart := m.buffers[0].cursorYOffset
			viewportEnd := viewportStart + m.viewport.Height - 2

			if cursorY < viewportStart || cursorY >= viewportEnd {
				t.Errorf("cursor Y (%d) is outside viewport [%d, %d)", cursorY, viewportStart, viewportEnd)
			}
		})
	}
} 