package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func createTestModel() model {
	return initialModel()
}

func TestInsertModeEnterKey(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Add some content to the buffer
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello world").(buffer)
	buff = buff.SetCursorX(5).(buffer) // Position cursor in middle
	m.buffers[0] = buff

	// Test Enter key at middle of line
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that line was split correctly
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.NoOfLines() != 2 {
		t.Errorf("Expected 2 lines after Enter, got %d", newBuff.NoOfLines())
	}
	if newBuff.Line(0) != "hello" {
		t.Errorf("Expected first line 'hello', got '%s'", newBuff.Line(0))
	}
	if newBuff.Line(1) != " world" {
		t.Errorf("Expected second line ' world', got '%s'", newBuff.Line(1))
	}
	if newBuff.CursorX() != 0 || newBuff.CursorY() != 1 {
		t.Errorf("Expected cursor at (0,1), got (%d,%d)", newBuff.CursorX(), newBuff.CursorY())
	}
}

func TestInsertModeEnterKeyAtEnd(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Add content and position cursor at end
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello world").(buffer)
	buff = buff.SetCursorX(11).(buffer) // Position cursor at end
	m.buffers[0] = buff

	// Test Enter key at end of line
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that empty line was created
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.NoOfLines() != 2 {
		t.Errorf("Expected 2 lines after Enter at end, got %d", newBuff.NoOfLines())
	}
	if newBuff.Line(0) != "hello world" {
		t.Errorf("Expected first line 'hello world', got '%s'", newBuff.Line(0))
	}
	if newBuff.Line(1) != "" {
		t.Errorf("Expected second line empty, got '%s'", newBuff.Line(1))
	}
	if newBuff.CursorX() != 0 || newBuff.CursorY() != 1 {
		t.Errorf("Expected cursor at (0,1), got (%d,%d)", newBuff.CursorX(), newBuff.CursorY())
	}
}

func TestInsertModeBackspaceNormal(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Add content and position cursor in middle
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello world").(buffer)
	buff = buff.SetCursorX(5).(buffer) // Position cursor in middle
	m.buffers[0] = buff

	// Test backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that character was deleted
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.Line(0) != "hell world" {
		t.Errorf("Expected 'hell world' after backspace, got '%s'", newBuff.Line(0))
	}
	if newBuff.CursorX() != 4 {
		t.Errorf("Expected cursor at position 4, got %d", newBuff.CursorX())
	}
}

func TestInsertModeBackspaceAtBeginning(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Add two lines with content
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello").(buffer)
	buff = buff.InsertLine(1, "world").(buffer)
	buff = buff.SetCursorX(0).(buffer) // Position cursor at beginning of second line
	buff = buff.SetCursorY(1).(buffer)
	m.buffers[0] = buff

	// Test backspace at beginning of line
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that lines were joined
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.NoOfLines() != 1 {
		t.Errorf("Expected 1 line after backspace at beginning, got %d", newBuff.NoOfLines())
	}
	if newBuff.Line(0) != "helloworld" {
		t.Errorf("Expected 'helloworld' after joining lines, got '%s'", newBuff.Line(0))
	}
	if newBuff.CursorX() != 5 || newBuff.CursorY() != 0 {
		t.Errorf("Expected cursor at (5,0), got (%d,%d)", newBuff.CursorX(), newBuff.CursorY())
	}
}

func TestInsertModeBackspaceAtBeginningOfFirstLine(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Add content and position cursor at beginning
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello world").(buffer)
	buff = buff.SetCursorX(0).(buffer) // Position cursor at beginning
	buff = buff.SetCursorY(0).(buffer)
	m.buffers[0] = buff

	// Test backspace at beginning of first line (should do nothing)
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that nothing changed
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.NoOfLines() != 1 {
		t.Errorf("Expected 1 line after backspace at beginning of first line, got %d", newBuff.NoOfLines())
	}
	if newBuff.Line(0) != "hello world" {
		t.Errorf("Expected 'hello world' unchanged, got '%s'", newBuff.Line(0))
	}
	if newBuff.CursorX() != 0 || newBuff.CursorY() != 0 {
		t.Errorf("Expected cursor at (0,0), got (%d,%d)", newBuff.CursorX(), newBuff.CursorY())
	}
}

func TestInsertModeCharacterInsertion(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Test inserting character in middle
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello world").(buffer)
	buff = buff.SetCursorX(5).(buffer) // Position cursor in middle
	m.buffers[0] = buff

	// Insert 'x'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that character was inserted
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.Line(0) != "hellox world" {
		t.Errorf("Expected 'hellox world' after insertion, got '%s'", newBuff.Line(0))
	}
	if newBuff.CursorX() != 6 {
		t.Errorf("Expected cursor at position 6, got %d", newBuff.CursorX())
	}
}

func TestInsertModeCharacterInsertionAtEnd(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Test inserting character at end
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello").(buffer)
	buff = buff.SetCursorX(5).(buffer) // Position cursor at end
	m.buffers[0] = buff

	// Insert '!'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that character was appended
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.Line(0) != "hello!" {
		t.Errorf("Expected 'hello!' after insertion at end, got '%s'", newBuff.Line(0))
	}
	if newBuff.CursorX() != 6 {
		t.Errorf("Expected cursor at position 6, got %d", newBuff.CursorX())
	}
}

func TestInsertModeSpaceKey(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Test inserting space
	buff := m.CurrentBuffer().(buffer)
	buff = buff.ReplaceLine(0, "hello").(buffer)
	buff = buff.SetCursorX(5).(buffer) // Position cursor at end
	m.buffers[0] = buff

	// Insert space
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that space was inserted
	newBuff := newM.CurrentBuffer().(buffer)
	if newBuff.Line(0) != "hello " {
		t.Errorf("Expected 'hello ' after space insertion, got '%s'", newBuff.Line(0))
	}
	if newBuff.CursorX() != 6 {
		t.Errorf("Expected cursor at position 6, got %d", newBuff.CursorX())
	}
}

func TestInsertModeEscapeKey(t *testing.T) {
	m := createTestModel()
	m.mode = ModeInsert

	// Test escape key
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.updateInsert(msg)
	newM := newModel.(model)

	// Check that mode changed to normal
	if newM.mode != ModeNormal {
		t.Errorf("Expected mode to be Normal after escape, got %s", newM.mode)
	}
	if newM.commandBuffer != "" {
		t.Errorf("Expected empty command buffer after escape, got '%s'", newM.commandBuffer)
	}
} 