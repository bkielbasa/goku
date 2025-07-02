package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// commandFindReferences finds all references to the symbol under cursor using LSP
type commandFindReferences struct{}

func (c commandFindReferences) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	// Get current buffer and cursor position
	buf := m.buffers[m.currBuffer]
	line := buf.cursorY
	character := buf.cursorX

	// Convert rune index to UTF-16 offset for LSP
	lineContent := buf.Line(line)
	utf16Offset := utf16Index(lineContent, character)

	// Create async command to find references
	return m, func() tea.Msg {
		return asyncFindReferencesInit{
			filePath:  buf.filename,
			line:      line,
			character: utf16Offset,
		}
	}
}

func (c commandFindReferences) Aliases() []string {
	return []string{"findref", "fr"}
}

// commandFileFinder opens a file finder floating window
type commandFileFinder struct{}

func (c commandFileFinder) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return m.SetErrorMessage("Failed to get current directory"), nil
	}

	// Find all files in the current directory and subdirectories
	var items []FloatingWindowItem
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create item
		item := FloatingWindowItem{
			ID:       path,
			Title:    filepath.Base(path),
			Subtitle: filepath.Base(filepath.Dir(path)),
			Data:     path,
		}
		items = append(items, item)

		return nil
	})

	if err != nil {
		return m.SetErrorMessage("Failed to scan directory"), nil
	}

	// Create floating window with callbacks
	fw := NewFloatingWindow("File Finder", items, m.viewport, 10).Open()
	
	// Set up callbacks for file selection
	fw.SetCallbacks(
		// onSelect callback - open selected file
		func(item FloatingWindowItem) tea.Cmd {
			if filePath, ok := item.Data.(string); ok {
				return func() tea.Msg {
					return openFileMsg{filePath: filePath}
				}
			}
			return nil
		},
		// onCancel callback - close window
		func() tea.Cmd {
			return func() tea.Msg {
				return closeFloatingWindowMsg{}
			}
		},
	)
	
	m.floatingWindow = fw
	m.mode = ModeFloatingWindow
	return m, nil
}

func (c commandFileFinder) Aliases() []string {
	return []string{"find", "f"}
}

// commandBufferList opens a buffer list floating window
type commandBufferList struct{}

func (c commandBufferList) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	var items []FloatingWindowItem

	for i, buf := range m.buffers {
		title := buf.filename
		if title == "" {
			title = "[No name]"
		}

		// Add modification status to title instead of subtitle
		if buf.state == bufferStateModified {
			title += " *"
		}

		item := FloatingWindowItem{
			ID:       fmt.Sprintf("buffer_%d", i),
			Title:    title,
			Subtitle: "", // Remove the "Buffer 1", "Buffer 2" labels
			Data:     i, // Buffer index
		}
		items = append(items, item)
	}

	// Create floating window with callbacks
	fw := NewFloatingWindow("Buffer List", items, m.viewport, 10).Open()
	
	// Set up callbacks for buffer selection
	fw.SetCallbacks(
		// onSelect callback - switch to selected buffer
		func(item FloatingWindowItem) tea.Cmd {
			if bufferIndex, ok := item.Data.(int); ok {
				return func() tea.Msg {
					return switchBufferMsg{bufferIndex: bufferIndex}
				}
			}
			return nil
		},
		// onCancel callback - close window
		func() tea.Cmd {
			return func() tea.Msg {
				return closeFloatingWindowMsg{}
			}
		},
	)
	
	m.floatingWindow = fw
	m.mode = ModeFloatingWindow
	return m, nil
}

func (c commandBufferList) Aliases() []string {
	return []string{"buffers", "ls"}
}

// Message types for floating window actions
type openFileMsg struct {
	filePath string
}

type switchBufferMsg struct {
	bufferIndex int
}

// Async messages for find references
type asyncFindReferencesInit struct {
	filePath  string
	line      int
	character int
}

type asyncFindReferencesResult struct {
	locations []location
	error     error
} 