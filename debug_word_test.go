package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDebugWordNavigation(t *testing.T) {
	content := "hello world test"
	
	m := initialModel()
	m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}
	
	lines := strings.Split(content, "\n")
	m.buffers[0].lines = lines
	m.buffers[0].cursorX = 10 // position at 't' in "test"
	m.buffers[0].viewport = m.viewport

	fmt.Printf("Initial position: cursorX=%d, cursorY=%d\n", m.buffers[0].cursorX, m.buffers[0].cursorY)
	fmt.Printf("Content: '%s'\n", content)
	fmt.Printf("Character at position 10: '%c'\n", []rune(content)[10])

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}
	newModel, _ := m.Update(keyMsg)

	buffer := newModel.(model).buffers[0]
	fmt.Printf("After 'b': cursorX=%d, cursorY=%d\n", buffer.cursorX, buffer.cursorY)
	
	// Print the character at the new position
	if buffer.cursorX < len(content) {
		fmt.Printf("Character at new position: '%c'\n", []rune(content)[buffer.cursorX])
	}
} 