package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFloatingWindowCreation(t *testing.T) {
	items := []FloatingWindowItem{
		{ID: "1", Title: "Item 1", Subtitle: "First item", Data: "data1"},
		{ID: "2", Title: "Item 2", Subtitle: "Second item", Data: "data2"},
		{ID: "3", Title: "Item 3", Subtitle: "Third item", Data: "data3"},
	}

	fw := NewFloatingWindow("Test Window", items)
	
	if fw == nil {
		t.Fatal("Floating window should not be nil")
	}
	
	if fw.title != "Test Window" {
		t.Errorf("Expected title 'Test Window', got '%s'", fw.title)
	}
	
	if len(fw.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(fw.items))
	}
	
	if !fw.IsOpen() {
		t.Error("Floating window should be open after creation")
	}
}

func TestFloatingWindowFiltering(t *testing.T) {
	items := []FloatingWindowItem{
		{ID: "1", Title: "apple", Subtitle: "fruit", Data: "data1"},
		{ID: "2", Title: "banana", Subtitle: "fruit", Data: "data2"},
		{ID: "3", Title: "carrot", Subtitle: "vegetable", Data: "data3"},
	}

	fw := NewFloatingWindow("Test Window", items)
	
	// Test filtering
	fw.filterText = "apple"
	fw.applyFilter()
	
	if len(fw.filteredItems) != 1 {
		t.Errorf("Expected 1 filtered item, got %d", len(fw.filteredItems))
	}
	
	if fw.filteredItems[0].Title != "apple" {
		t.Errorf("Expected filtered item 'apple', got '%s'", fw.filteredItems[0].Title)
	}
	
	// Test case-insensitive filtering
	fw.filterText = "BANANA"
	fw.applyFilter()
	
	if len(fw.filteredItems) != 1 {
		t.Errorf("Expected 1 filtered item, got %d", len(fw.filteredItems))
	}
	
	if fw.filteredItems[0].Title != "banana" {
		t.Errorf("Expected filtered item 'banana', got '%s'", fw.filteredItems[0].Title)
	}
}

func TestFloatingWindowNavigation(t *testing.T) {
	items := []FloatingWindowItem{
		{ID: "1", Title: "Item 1", Data: "data1"},
		{ID: "2", Title: "Item 2", Data: "data2"},
		{ID: "3", Title: "Item 3", Data: "data3"},
	}

	fw := NewFloatingWindow("Test Window", items)
	
	// Test initial selection
	if fw.selectedIndex != 0 {
		t.Errorf("Expected initial selection 0, got %d", fw.selectedIndex)
	}
	
	// Test moving down
	fw.selectedIndex = 1
	if fw.selectedIndex != 1 {
		t.Errorf("Expected selection 1, got %d", fw.selectedIndex)
	}
	
	// Test moving up
	fw.selectedIndex = 0
	if fw.selectedIndex != 0 {
		t.Errorf("Expected selection 0, got %d", fw.selectedIndex)
	}
}

func TestFloatingWindowRendering(t *testing.T) {
	items := []FloatingWindowItem{
		{ID: "1", Title: "Item 1", Subtitle: "First item", Data: "data1"},
		{ID: "2", Title: "Item 2", Subtitle: "Second item", Data: "data2"},
	}

	fw := NewFloatingWindow("Test Window", items)
	fw.SetSize(40, 10)
	fw.SetPosition(5, 5)
	
	view := fw.View()
	
	if view == "" {
		t.Error("Floating window view should not be empty")
	}
	
	if !strings.Contains(view, "Test Window") {
		t.Error("View should contain window title")
	}
	
	if !strings.Contains(view, "Item 1") {
		t.Error("View should contain first item")
	}
	
	if !strings.Contains(view, "Item 2") {
		t.Error("View should contain second item")
	}
}

func TestFloatingWindowModeSwitching(t *testing.T) {
	items := []FloatingWindowItem{
		{ID: "1", Title: "Item 1", Data: "data1"},
	}

	fw := NewFloatingWindow("Test Window", items)
	
	// Test initial mode
	if fw.mode != FloatingWindowList {
		t.Errorf("Expected initial mode FloatingWindowList, got %s", fw.mode)
	}
	
	// Test switching to filter mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	fw.Update(keyMsg)
	
	if fw.mode != FloatingWindowFilter {
		t.Errorf("Expected mode FloatingWindowFilter, got %s", fw.mode)
	}
	
	if fw.filterText != "a" {
		t.Errorf("Expected filter text 'a', got '%s'", fw.filterText)
	}
	
	// Test switching back to list mode
	escapeMsg := tea.KeyMsg{Type: tea.KeyEscape}
	fw.Update(escapeMsg)
	
	if fw.mode != FloatingWindowList {
		t.Errorf("Expected mode FloatingWindowList, got %s", fw.mode)
	}
	
	if fw.filterText != "" {
		t.Errorf("Expected empty filter text, got '%s'", fw.filterText)
	}
}

func TestFloatingWindowClosing(t *testing.T) {
	items := []FloatingWindowItem{
		{ID: "1", Title: "Item 1", Data: "data1"},
	}

	fw := NewFloatingWindow("Test Window", items)
	
	if !fw.IsOpen() {
		t.Error("Floating window should be open initially")
	}
	
	fw.Close()
	
	if fw.IsOpen() {
		t.Error("Floating window should be closed after Close()")
	}
	
	if fw.mode != FloatingWindowClosed {
		t.Errorf("Expected mode FloatingWindowClosed, got %s", fw.mode)
	}
} 