package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandOpenLineBelow(t *testing.T) {
	nm := NewNormalMode()
	
	// Create a buffer with some content
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.AppendLine("line 2")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(0)
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test o command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}, m)
	newM := newModel.(model)
	
	// Check that a new line was inserted below
	if newM.buffers[0].NoOfLines() != 3 {
		t.Errorf("Expected 3 lines after o command, got %d", newM.buffers[0].NoOfLines())
	}
	
	// Check that the new line is empty
	if newM.buffers[0].Line(1) != "" {
		t.Errorf("Expected empty line at position 1, got '%s'", newM.buffers[0].Line(1))
	}
	
	// Check that cursor moved to the new line
	if newM.buffers[0].cursorY != 1 {
		t.Errorf("Expected cursor Y to be 1, got %d", newM.buffers[0].cursorY)
	}
	
	// Check that cursor X is at beginning of line
	if newM.buffers[0].cursorX != 0 {
		t.Errorf("Expected cursor X to be 0, got %d", newM.buffers[0].cursorX)
	}
	
	// Check that mode switched to insert
	if newM.mode != ModeInsert {
		t.Errorf("Expected mode to be insert, got %s", newM.mode)
	}
}

func TestCommandOpenLineAbove(t *testing.T) {
	nm := NewNormalMode()
	
	// Create a buffer with some content
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.AppendLine("line 2")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(1)
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test O command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}}, m)
	newM := newModel.(model)
	
	// Check that a new line was inserted above
	if newM.buffers[0].NoOfLines() != 3 {
		t.Errorf("Expected 3 lines after O command, got %d", newM.buffers[0].NoOfLines())
	}
	
	// Check that the new line is empty
	if newM.buffers[0].Line(1) != "" {
		t.Errorf("Expected empty line at position 1, got '%s'", newM.buffers[0].Line(1))
	}
	
	// Check that cursor stayed at the same line
	if newM.buffers[0].cursorY != 1 {
		t.Errorf("Expected cursor Y to be 1, got %d", newM.buffers[0].cursorY)
	}
	
	// Check that cursor X is at beginning of line
	if newM.buffers[0].cursorX != 0 {
		t.Errorf("Expected cursor X to be 0, got %d", newM.buffers[0].cursorX)
	}
	
	// Check that mode switched to insert
	if newM.mode != ModeInsert {
		t.Errorf("Expected mode to be insert, got %s", newM.mode)
	}
}

func TestCommandEnterInsertMode(t *testing.T) {
	nm := NewNormalMode()
	
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test i command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}, m)
	newM := newModel.(model)
	
	// Check that mode switched to insert
	if newM.mode != ModeInsert {
		t.Errorf("Expected mode to be insert, got %s", newM.mode)
	}
}

func TestCommandEnterCommandMode(t *testing.T) {
	nm := NewNormalMode()
	
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test : command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}, m)
	newM := newModel.(model)
	
	// Check that mode switched to command
	if newM.mode != ModeCommand {
		t.Errorf("Expected mode to be command, got %s", newM.mode)
	}
}

func TestCommandClearBuffer(t *testing.T) {
	nm := NewNormalMode()
	
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test esc command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyEscape}, m)
	newM := newModel.(model)
	
	// Check that buffer was cleared
	if newM.normalmode.(normalmode).buffer != "" {
		t.Errorf("Expected buffer to be cleared, got '%s'", newM.normalmode.(normalmode).buffer)
	}
} 