package normalmode

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Mock buffer for testing
type mockBuffer struct {
	lines   []string
	cursorX int
	cursorY int
}

func (b mockBuffer) Viewport() tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: 80, Height: 24}
}

func (b mockBuffer) CursorX() int {
	return b.cursorX
}

func (b mockBuffer) SetCursorX(n int) Buffer {
	b.cursorX = n
	return b
}

func (b mockBuffer) IncreaseCursorX(n int) Buffer {
	b.cursorX += n
	if b.cursorX < 0 {
		b.cursorX = 0
	}
	return b
}

func (b mockBuffer) CursorY() int {
	return b.cursorY
}

func (b mockBuffer) SetCursorY(n int) Buffer {
	b.cursorY = n
	return b
}

func (b mockBuffer) IncreaseCursorY(n int) Buffer {
	b.cursorY += n
	if b.cursorY < 0 {
		b.cursorY = 0
	}
	return b
}

func (b mockBuffer) CursorXOffset() int {
	return 0
}

func (b mockBuffer) IncreaseCursorXOffset(n int) Buffer {
	return b
}

func (b mockBuffer) CursorYOffset() int {
	return 0
}

func (b mockBuffer) IncreaseCursorYOffset(n int) Buffer {
	return b
}

func (b mockBuffer) NoOfLines() int {
	return len(b.lines)
}

func (b mockBuffer) Line(n int) string {
	if n >= 0 && n < len(b.lines) {
		return b.lines[n]
	}
	return ""
}

func (b mockBuffer) Lines() []string {
	return b.lines
}

func (b mockBuffer) ReplaceLine(n int, s string) Buffer {
	if n >= 0 && n < len(b.lines) {
		b.lines[n] = s
	}
	return b
}

func (b mockBuffer) AppendLine(s string) Buffer {
	b.lines = append(b.lines, s)
	return b
}

func (b mockBuffer) InsertLine(n int, s string) Buffer {
	if n < 0 {
		n = 0
	}
	if n > len(b.lines) {
		n = len(b.lines)
	}
	b.lines = append(b.lines[:n], append([]string{s}, b.lines[n:]...)...)
	return b
}

func (b mockBuffer) DeleteLine(n int) Buffer {
	if n >= 0 && n < len(b.lines) {
		b.lines = append(b.lines[:n], b.lines[n+1:]...)
	}
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	return b
}

func (b mockBuffer) SetStateModified() Buffer {
	return b
}

func (b mockBuffer) FileName() string {
	return ""
}

func (b mockBuffer) SetFileName(f string) Buffer {
	return b
}

// Mock editor model for testing
type mockEditorModel struct {
	buffer Buffer
	mode   string
}

func (m mockEditorModel) CurrentBuffer() Buffer {
	return m.buffer
}

func (m mockEditorModel) ReplaceCurrentBuffer(b Buffer) EditorModel {
	m.buffer = b
	return m
}

func (m mockEditorModel) EnterCommandMode() EditorModel {
	m.mode = "command"
	return m
}

func (m mockEditorModel) EnterInsertMode() EditorModel {
	m.mode = "insert"
	return m
}

func (m mockEditorModel) Init() tea.Cmd {
	return nil
}

func (m mockEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m mockEditorModel) View() string {
	return ""
}

func TestCommandOpenLineBelow(t *testing.T) {
	nm := New()
	
	// Create a buffer with some content
	buff := mockBuffer{
		lines:   []string{"line 1", "line 2"},
		cursorX: 3,
		cursorY: 0,
	}
	
	m := mockEditorModel{
		buffer: buff,
		mode:   "normal",
	}
	
	// Test o command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}, m)
	newM := newModel.(mockEditorModel)
	
	// Check that a new line was inserted below
	if newM.buffer.(mockBuffer).NoOfLines() != 3 {
		t.Errorf("Expected 3 lines after o command, got %d", newM.buffer.(mockBuffer).NoOfLines())
	}
	
	// Check that the new line is empty
	if newM.buffer.(mockBuffer).Line(1) != "" {
		t.Errorf("Expected empty line at position 1, got '%s'", newM.buffer.(mockBuffer).Line(1))
	}
	
	// Check that cursor moved to the new line
	if newM.buffer.(mockBuffer).CursorY() != 1 {
		t.Errorf("Expected cursor Y at 1, got %d", newM.buffer.(mockBuffer).CursorY())
	}
	
	// Check that cursor is at beginning of line
	if newM.buffer.(mockBuffer).CursorX() != 0 {
		t.Errorf("Expected cursor X at 0, got %d", newM.buffer.(mockBuffer).CursorX())
	}
	
	// Check that mode changed to insert
	if newM.mode != "insert" {
		t.Errorf("Expected mode to be insert, got %s", newM.mode)
	}
}

func TestCommandOpenLineAbove(t *testing.T) {
	nm := New()
	
	// Create a buffer with some content
	buff := mockBuffer{
		lines:   []string{"line 1", "line 2"},
		cursorX: 3,
		cursorY: 1, // Position cursor on second line
	}
	
	m := mockEditorModel{
		buffer: buff,
		mode:   "normal",
	}
	
	// Test O command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}}, m)
	newM := newModel.(mockEditorModel)
	
	// Check that a new line was inserted above
	if newM.buffer.(mockBuffer).NoOfLines() != 3 {
		t.Errorf("Expected 3 lines after O command, got %d", newM.buffer.(mockBuffer).NoOfLines())
	}
	
	// Check that the new line is empty
	if newM.buffer.(mockBuffer).Line(1) != "" {
		t.Errorf("Expected empty line at position 1, got '%s'", newM.buffer.(mockBuffer).Line(1))
	}
	
	// Check that cursor moved to the new line
	if newM.buffer.(mockBuffer).CursorY() != 1 {
		t.Errorf("Expected cursor Y at 1, got %d", newM.buffer.(mockBuffer).CursorY())
	}
	
	// Check that cursor is at beginning of line
	if newM.buffer.(mockBuffer).CursorX() != 0 {
		t.Errorf("Expected cursor X at 0, got %d", newM.buffer.(mockBuffer).CursorX())
	}
	
	// Check that mode changed to insert
	if newM.mode != "insert" {
		t.Errorf("Expected mode to be insert, got %s", newM.mode)
	}
}

func TestCommandEnterInsertMode(t *testing.T) {
	nm := New()
	
	buff := mockBuffer{
		lines:   []string{"line 1"},
		cursorX: 0,
		cursorY: 0,
	}
	
	m := mockEditorModel{
		buffer: buff,
		mode:   "normal",
	}
	
	// Test i command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}, m)
	newM := newModel.(mockEditorModel)
	
	// Check that mode changed to insert
	if newM.mode != "insert" {
		t.Errorf("Expected mode to be insert, got %s", newM.mode)
	}
}

func TestCommandEnterCommandMode(t *testing.T) {
	nm := New()
	
	buff := mockBuffer{
		lines:   []string{"line 1"},
		cursorX: 0,
		cursorY: 0,
	}
	
	m := mockEditorModel{
		buffer: buff,
		mode:   "normal",
	}
	
	// Test : command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}, m)
	newM := newModel.(mockEditorModel)
	
	// Check that mode changed to command
	if newM.mode != "command" {
		t.Errorf("Expected mode to be command, got %s", newM.mode)
	}
}

func TestCommandClearBuffer(t *testing.T) {
	nm := New()
	
	buff := mockBuffer{
		lines:   []string{"line 1"},
		cursorX: 0,
		cursorY: 0,
	}
	
	m := mockEditorModel{
		buffer: buff,
		mode:   "normal",
	}
	
	// Test esc command
	_, newModel, _ := nm.Handle(tea.KeyMsg{Type: tea.KeyEscape}, m)
	newM := newModel.(mockEditorModel)
	
	// Check that buffer was cleared (this would depend on the actual implementation)
	// For now, just check that the command was handled
	if newM.mode != "normal" {
		t.Errorf("Expected mode to remain normal, got %s", newM.mode)
	}
} 