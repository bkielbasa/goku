package main

import (
	"testing"
)

func TestBufferInsertLine(t *testing.T) {
	style := newEditorStyle()
	b := newBuffer(style)

	// Test inserting at beginning
	b = b.InsertLine(0, "first line")
	if len(b.lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(b.lines))
	}
	if b.lines[0] != "first line" {
		t.Errorf("Expected 'first line', got '%s'", b.lines[0])
	}

	// Test inserting at end
	b = b.InsertLine(2, "last line")
	if len(b.lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(b.lines))
	}
	if b.lines[2] != "last line" {
		t.Errorf("Expected 'last line', got '%s'", b.lines[2])
	}

	// Test inserting in middle
	b = b.InsertLine(1, "middle line")
	if len(b.lines) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(b.lines))
	}
	if b.lines[1] != "middle line" {
		t.Errorf("Expected 'middle line', got '%s'", b.lines[1])
	}

	// Test bounds checking - negative index should insert at beginning
	b = b.InsertLine(-1, "should insert at beginning")
	if len(b.lines) != 5 {
		t.Errorf("Expected 5 lines after negative index, got %d", len(b.lines))
	}
	if b.lines[0] != "should insert at beginning" {
		t.Errorf("Expected 'should insert at beginning' at position 0, got '%s'", b.lines[0])
	}

	// Test bounds checking - index beyond length should append
	b = b.InsertLine(10, "should append")
	if len(b.lines) != 6 {
		t.Errorf("Expected 6 lines after large index, got %d", len(b.lines))
	}
	if b.lines[5] != "should append" {
		t.Errorf("Expected 'should append' at position 5, got '%s'", b.lines[5])
	}
}

func TestBufferDeleteLine(t *testing.T) {
	style := newEditorStyle()
	b := newBuffer(style)

	// Add some lines
	b = b.InsertLine(0, "line 1")
	b = b.InsertLine(1, "line 2")
	b = b.InsertLine(2, "line 3")

	// Test deleting middle line
	b = b.DeleteLine(1)
	if len(b.lines) != 3 {
		t.Errorf("Expected 3 lines after delete, got %d", len(b.lines))
	}
	if b.lines[0] != "line 1" {
		t.Errorf("Expected 'line 1', got '%s'", b.lines[0])
	}
	if b.lines[1] != "line 3" {
		t.Errorf("Expected 'line 3', got '%s'", b.lines[1])
	}

	// Test deleting first line
	b = b.DeleteLine(0)
	if len(b.lines) != 2 {
		t.Errorf("Expected 2 lines after delete first, got %d", len(b.lines))
	}
	if b.lines[0] != "line 3" {
		t.Errorf("Expected 'line 3', got '%s'", b.lines[0])
	}

	// Test deleting last line
	b = b.DeleteLine(1)
	if len(b.lines) != 1 {
		t.Errorf("Expected 1 line after delete last, got %d", len(b.lines))
	}

	// Test deleting all lines - should keep at least one empty line
	b = b.DeleteLine(0)
	if len(b.lines) != 1 {
		t.Errorf("Expected 1 line after delete all, got %d", len(b.lines))
	}
	if b.lines[0] != "" {
		t.Errorf("Expected empty line, got '%s'", b.lines[0])
	}

	// Test bounds checking - negative index
	b = b.InsertLine(0, "test line")
	b = b.DeleteLine(-1)
	if len(b.lines) != 2 {
		t.Errorf("Expected 2 lines after negative index delete, got %d", len(b.lines))
	}

	// Test bounds checking - index beyond length
	b = b.DeleteLine(10)
	if len(b.lines) != 2 {
		t.Errorf("Expected 2 lines after out of bounds delete, got %d", len(b.lines))
	}
}

func TestBufferLineOperations(t *testing.T) {
	style := newEditorStyle()
	b := newBuffer(style)

	// Test initial state
	if b.NoOfLines() != 1 {
		t.Errorf("Expected 1 line initially, got %d", b.NoOfLines())
	}
	if b.Line(0) != "" {
		t.Errorf("Expected empty line initially, got '%s'", b.Line(0))
	}

	// Test replacing line
	b = b.ReplaceLine(0, "new content")
	if b.Line(0) != "new content" {
		t.Errorf("Expected 'new content', got '%s'", b.Line(0))
	}

	// Test appending line
	b = b.AppendLine("appended line")
	if len(b.lines) != 2 {
		t.Errorf("Expected 2 lines after append, got %d", len(b.lines))
	}
	if b.lines[1] != "appended line" {
		t.Errorf("Expected 'appended line', got '%s'", b.lines[1])
	}
}

func TestBufferCursorOperations(t *testing.T) {
	style := newEditorStyle()
	b := newBuffer(style)

	// Test initial cursor position
	if b.CursorX() != 0 || b.CursorY() != 0 {
		t.Errorf("Expected cursor at (0,0), got (%d,%d)", b.CursorX(), b.CursorY())
	}

	// Test setting cursor position
	b = b.SetCursorX(5)
	// Add some lines before setting cursor Y to position 2
	b = b.AppendLine("line 1")
	b = b.AppendLine("line 2")
	b = b.SetCursorY(2)
	if b.CursorX() != 5 || b.CursorY() != 2 {
		t.Errorf("Expected cursor at (5,2), got (%d,%d)", b.CursorX(), b.CursorY())
	}

	// Test increasing cursor position
	b = b.IncreaseCursorX(3)
	b = b.IncreaseCursorY(1)
	if b.CursorX() != 8 || b.CursorY() != 3 {
		t.Errorf("Expected cursor at (8,3), got (%d,%d)", b.CursorX(), b.CursorY())
	}

	// Test decreasing cursor position
	b = b.IncreaseCursorX(-2)
	b = b.IncreaseCursorY(-1)
	if b.CursorX() != 6 || b.CursorY() != 2 {
		t.Errorf("Expected cursor at (6,2), got (%d,%d)", b.CursorX(), b.CursorY())
	}

	// Test cursor bounds - should not go below 0
	b = b.IncreaseCursorX(-10)
	b = b.IncreaseCursorY(-10)
	if b.CursorX() != 0 || b.CursorY() != 0 {
		t.Errorf("Expected cursor at (0,0) after negative bounds, got (%d,%d)", b.CursorX(), b.CursorY())
	}
} 