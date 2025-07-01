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
	if newM.normalmode.buffer != "" {
		t.Errorf("Expected buffer to be cleared, got '%s'", newM.normalmode.buffer)
	}
}

func TestCommandRepeat(t *testing.T) {
	nm := NewNormalMode()
	
	// Create a buffer with some content
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.AppendLine("line 2")
	buff = buff.AppendLine("line 3")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(1)
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// First, press 'd' (should buffer)
	newNM, newModel, _ := m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
	m = newModel.(model)
	m.normalmode = newNM

	// Then, press 'd' again (should execute dd)
	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
	m = newModel.(model)
	m.normalmode = newNM

	// Check that a line was deleted
	if m.buffers[0].NoOfLines() != 2 {
		t.Errorf("Expected 2 lines after dd command, got %d", m.buffers[0].NoOfLines())
	}
	// Check that lastCommand was set
	if m.normalmode.lastCommand != "dd" {
		t.Errorf("Expected lastCommand to be 'dd', got '%s'", m.normalmode.lastCommand)
	}
	// Now test the repeat command (.)
	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	// Check that another line was deleted
	if m.buffers[0].NoOfLines() != 1 {
		t.Errorf("Expected 1 line after repeat command, got %d", m.buffers[0].NoOfLines())
	}
}

func TestCommandRepeatNoLastCommand(t *testing.T) {
	nm := NewNormalMode()
	
	// Create a buffer with some content
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(0)
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test repeat command with no last command
	newNM, newModel, _ := m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}}, m)
	newM := newModel.(model)
	newM.normalmode = newNM
	// Check that nothing changed (no last command to repeat)
	if newM.buffers[0].NoOfLines() != 1 {
		t.Errorf("Expected 1 line after repeat command with no last command, got %d", newM.buffers[0].NoOfLines())
	}
	if newM.buffers[0].Line(0) != "line 1" {
		t.Errorf("Expected line to remain unchanged, got '%s'", newM.buffers[0].Line(0))
	}
}

func TestNavigationCommandsNotRepeatable(t *testing.T) {
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
	
	// Test that navigation command 'j' is not repeatable
	newNM, newModel, _ := m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that cursor moved down
	if m.buffers[0].cursorY != 1 {
		t.Errorf("Expected cursor Y to be 1 after 'j' command, got %d", m.buffers[0].cursorY)
	}
	
	// Check that lastCommand was NOT set (navigation commands are not repeatable)
	if m.normalmode.lastCommand != "" {
		t.Errorf("Expected lastCommand to be empty after navigation command, got '%s'", m.normalmode.lastCommand)
	}
	
	// Test repeat command - should do nothing since no repeatable command was executed
	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that cursor position didn't change (no repeat occurred)
	if m.buffers[0].cursorY != 1 {
		t.Errorf("Expected cursor Y to remain 1 after repeat with no repeatable command, got %d", m.buffers[0].cursorY)
	}
}

func TestEditingCommandsAreRepeatable(t *testing.T) {
	nm := NewNormalMode()
	
	// Create a buffer with some content
	style := newEditorStyle()
	buff := newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.AppendLine("line 2")
	buff = buff.AppendLine("line 3")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(1)
	
	m := model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test that editing command 'dd' is repeatable
	newNM, newModel, _ := m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
	m = newModel.(model)
	m.normalmode = newNM

	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that a line was deleted
	if m.buffers[0].NoOfLines() != 2 {
		t.Errorf("Expected 2 lines after dd command, got %d", m.buffers[0].NoOfLines())
	}
	
	// Check that lastCommand was set (editing commands are repeatable)
	if m.normalmode.lastCommand != "dd" {
		t.Errorf("Expected lastCommand to be 'dd', got '%s'", m.normalmode.lastCommand)
	}
	
	// Test repeat command - should delete another line
	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that another line was deleted
	if m.buffers[0].NoOfLines() != 1 {
		t.Errorf("Expected 1 line after repeat command, got %d", m.buffers[0].NoOfLines())
	}
}

func TestEditingCommandsMarkBufferAsModified(t *testing.T) {
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
	
	// Test that dd command marks buffer as modified
	newNM, newModel, _ := m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
	m = newModel.(model)
	m.normalmode = newNM

	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that buffer is marked as modified
	if m.buffers[0].state != bufferStateModified {
		t.Errorf("Expected buffer state to be modified after dd command, got %s", m.buffers[0].state)
	}
	
	// Reset buffer for next test
	buff = newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.AppendLine("line 2")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(0)
	
	m = model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test that o command marks buffer as modified
	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that buffer is marked as modified
	if m.buffers[0].state != bufferStateModified {
		t.Errorf("Expected buffer state to be modified after o command, got %s", m.buffers[0].state)
	}
	
	// Reset buffer for next test
	buff = newBuffer(style)
	buff = buff.ReplaceLine(0, "line 1")
	buff = buff.AppendLine("line 2")
	buff = buff.SetCursorX(3)
	buff = buff.SetCursorY(1)
	
	m = model{
		buffers:  []buffer{buff},
		currBuffer: 0,
		mode:     ModeNormal,
		normalmode: nm,
	}
	
	// Test that O command marks buffer as modified
	newNM, newModel, _ = m.normalmode.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}}, m)
	m = newModel.(model)
	m.normalmode = newNM
	
	// Check that buffer is marked as modified
	if m.buffers[0].state != bufferStateModified {
		t.Errorf("Expected buffer state to be modified after O command, got %s", m.buffers[0].state)
	}
} 