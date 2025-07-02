package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FloatingWindowMode represents the different modes a floating window can be in
type FloatingWindowMode string

const (
	FloatingWindowList    FloatingWindowMode = "list"    // Display list with selection
	FloatingWindowFilter  FloatingWindowMode = "filter"  // Filter mode (typing to search)
	FloatingWindowClosed  FloatingWindowMode = "closed"  // Window is closed
)

// FloatingWindowItem represents an item in the floating window list
type FloatingWindowItem struct {
	ID       string // Unique identifier
	Title    string // Display title
	Subtitle string // Optional subtitle
	Data     interface{} // Additional data (e.g., file path, line number)
}

// FloatingWindow represents a floating window that can display lists with filtering
type FloatingWindow struct {
	mode           FloatingWindowMode
	title          string
	items          []FloatingWindowItem
	filteredItems  []FloatingWindowItem
	filterText     string
	selectedIndex  int
	cursorX        int
	width          int
	height         int
	posX           int
	posY           int
	style          floatingWindowStyle
	onSelect       func(FloatingWindowItem) tea.Cmd
	onCancel       func() tea.Cmd
	buffers        []buffer // Add this field for buffer access
}

// floatingWindowStyle defines the styling for the floating window
type floatingWindowStyle struct {
	window     lipgloss.Style
	title      lipgloss.Style
	border     lipgloss.Style
	item       lipgloss.Style
	selected   lipgloss.Style
	filter     lipgloss.Style
	subtitle   lipgloss.Style
	previewHighlight lipgloss.Style
}

// newFloatingWindowStyle creates a new floating window style
func newFloatingWindowStyle() floatingWindowStyle {
	return floatingWindowStyle{
		window: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Padding(0, 1),
		border: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d4d4d4")).
			Padding(0, 1),
		selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#3c3c3c")).
			Padding(0, 1),
		filter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#87ceeb")).
			Padding(0, 1),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Padding(0, 1),
		previewHighlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#3c3c3c")).
			Padding(0, 1),
	}
}

// NewFloatingWindow creates a new floating window
func NewFloatingWindow(title string, items []FloatingWindowItem, viewport tea.WindowSizeMsg, margin int) *FloatingWindow {
	windowWidth := viewport.Width - 2*margin
	windowHeight := viewport.Height - 2*margin
	if windowWidth < 20 {
		windowWidth = 20
	}
	if windowHeight < 5 {
		windowHeight = 5
	}
	posX := margin
	posY := margin
	fw := &FloatingWindow{
		mode:          FloatingWindowList,
		title:         title,
		items:         items,
		filteredItems: items,
		selectedIndex: 0,
		cursorX:       0,
		width:         windowWidth,
		height:        windowHeight,
		posX:          posX,
		posY:          posY,
		style:         newFloatingWindowStyle(),
	}
	if len(fw.filteredItems) > 0 && fw.selectedIndex >= len(fw.filteredItems) {
		fw.selectedIndex = len(fw.filteredItems) - 1
	}
	return fw
}

// SetPosition sets the position of the floating window
func (fw *FloatingWindow) SetPosition(x, y int) *FloatingWindow {
	fw.posX = x
	fw.posY = y
	return fw
}

// SetSize sets the size of the floating window
func (fw *FloatingWindow) SetSize(width, height int) *FloatingWindow {
	fw.width = width
	fw.height = height
	return fw
}

// SetCallbacks sets the callback functions for selection and cancellation
func (fw *FloatingWindow) SetCallbacks(onSelect func(FloatingWindowItem) tea.Cmd, onCancel func() tea.Cmd) *FloatingWindow {
	fw.onSelect = onSelect
	fw.onCancel = onCancel
	return fw
}

// Open opens the floating window
func (fw *FloatingWindow) Open() *FloatingWindow {
	fw.mode = FloatingWindowList
	fw.filterText = ""
	fw.filteredItems = fw.items
	fw.selectedIndex = 0
	fw.cursorX = 0
	return fw
}

// Close closes the floating window
func (fw *FloatingWindow) Close() *FloatingWindow {
	fw.mode = FloatingWindowClosed
	return fw
}

// IsOpen returns true if the window is open
func (fw *FloatingWindow) IsOpen() bool {
	return fw.mode != FloatingWindowClosed
}

// Update handles input for the floating window
func (fw *FloatingWindow) Update(msg tea.Msg) (*FloatingWindow, tea.Cmd) {
	if fw.mode == FloatingWindowClosed {
		return fw, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch fw.mode {
		case FloatingWindowList:
			return fw.handleListMode(msg)
		case FloatingWindowFilter:
			return fw.handleFilterMode(msg)
		}
	}

	return fw, nil
}

type closeFloatingWindowMsg struct{}

// handleListMode handles input when in list mode
func (fw *FloatingWindow) handleListMode(msg tea.KeyMsg) (*FloatingWindow, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if fw.selectedIndex > 0 {
			fw.selectedIndex--
		}
	case tea.KeyDown:
		if fw.selectedIndex < len(fw.filteredItems)-1 {
			fw.selectedIndex++
		}
	case tea.KeyEnter:
		if len(fw.filteredItems) > 0 && fw.onSelect != nil {
			return fw, fw.onSelect(fw.filteredItems[fw.selectedIndex])
		}
	case tea.KeyEscape:
		return fw, func() tea.Msg { return closeFloatingWindowMsg{} }
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			// Start filter mode
			fw.mode = FloatingWindowFilter
			fw.filterText = string(msg.Runes)
			fw.cursorX = len(fw.filterText)
			fw.applyFilter()
		}
	}

	return fw, nil
}

// handleFilterMode handles input when in filter mode
func (fw *FloatingWindow) handleFilterMode(msg tea.KeyMsg) (*FloatingWindow, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if fw.selectedIndex > 0 {
			fw.selectedIndex--
		}
	case tea.KeyDown:
		if fw.selectedIndex < len(fw.filteredItems)-1 {
			fw.selectedIndex++
		}
	case tea.KeyEnter:
		if len(fw.filteredItems) > 0 && fw.onSelect != nil {
			return fw, fw.onSelect(fw.filteredItems[fw.selectedIndex])
		}
	case tea.KeyEscape:
		fw.mode = FloatingWindowList
		fw.filterText = ""
		fw.cursorX = 0
		fw.filteredItems = fw.items
		fw.selectedIndex = 0
		return fw, func() tea.Msg { return closeFloatingWindowMsg{} }
	case tea.KeyBackspace:
		if fw.cursorX > 0 {
			fw.filterText = fw.filterText[:fw.cursorX-1] + fw.filterText[fw.cursorX:]
			fw.cursorX--
			fw.applyFilter()
		}
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			// Insert at cursor position
			before := fw.filterText[:fw.cursorX]
			after := fw.filterText[fw.cursorX:]
			fw.filterText = before + string(msg.Runes) + after
			fw.cursorX += len(msg.Runes)
			fw.applyFilter()
		}
	case tea.KeyLeft:
		if fw.cursorX > 0 {
			fw.cursorX--
		}
	case tea.KeyRight:
		if fw.cursorX < len(fw.filterText) {
			fw.cursorX++
		}
	}

	return fw, nil
}

// applyFilter applies the current filter text to the items
func (fw *FloatingWindow) applyFilter() {
	if fw.filterText == "" {
		fw.filteredItems = fw.items
		fw.selectedIndex = 0
		return
	}

	var filtered []FloatingWindowItem
	filterLower := strings.ToLower(fw.filterText)

	for _, item := range fw.items {
		titleLower := strings.ToLower(item.Title)
		subtitleLower := strings.ToLower(item.Subtitle)
		
		if strings.Contains(titleLower, filterLower) || strings.Contains(subtitleLower, filterLower) {
			filtered = append(filtered, item)
		}
	}

	fw.filteredItems = filtered
	
	// Adjust selected index
	if len(filtered) > 0 {
		if fw.selectedIndex >= len(filtered) {
			fw.selectedIndex = len(filtered) - 1
		}
	} else {
		fw.selectedIndex = 0
	}
}

// View renders the floating window
func (fw *FloatingWindow) View() string {
	if fw.mode == FloatingWindowClosed {
		return ""
	}

	// Layout parameters
	listWidth := fw.width / 2
	previewWidth := fw.width - listWidth - 2 // -2 for border padding
	contentHeight := fw.height - 2 // -2 for status and filter
	if contentHeight < 1 {
		contentHeight = 1
	}

	var lines []string

	// Always show filter input at the top
	filterLine := fw.style.filter.Render(fw.filterText)
	if len([]rune(filterLine)) < fw.width-2 {
		filterLine += strings.Repeat(" ", fw.width-2-len([]rune(filterLine)))
	}
	lines = append(lines, filterLine)

	// Calculate available height for items
	availableHeight := contentHeight - 1 // -1 for status line
	if availableHeight < 1 {
		availableHeight = 1
	}

	// List rendering (left)
	startIndex := 0
	endIndex := len(fw.filteredItems)
	if len(fw.filteredItems) > availableHeight {
		if fw.selectedIndex >= availableHeight {
			startIndex = fw.selectedIndex - availableHeight + 1
			endIndex = fw.selectedIndex + 1
		} else {
			endIndex = availableHeight
		}
	}
	var listLines []string
	for i := startIndex; i < endIndex && i < len(fw.filteredItems); i++ {
		item := fw.filteredItems[i]
		var itemStyle lipgloss.Style
		if i == fw.selectedIndex {
			itemStyle = fw.style.selected
		} else {
			itemStyle = fw.style.item
		}
		title := item.Title
		if len(title) > listWidth-2 {
			title = title[:listWidth-5] + "..."
		}
		itemLine := itemStyle.Width(listWidth-2).Render(title)
		listLines = append(listLines, itemLine)
	}
	for len(listLines) < availableHeight {
		listLines = append(listLines, strings.Repeat(" ", listWidth-2))
	}

	// Preview rendering (right)
	var previewLines []string
	previewLines = make([]string, availableHeight)
	var previewFilePath string
	var highlightLine int = -1
	var fileBuffer *buffer = nil
	if fw.selectedIndex >= 0 && fw.selectedIndex < len(fw.filteredItems) {
		item := fw.filteredItems[fw.selectedIndex]
		// Try to extract a file path and context
		switch v := item.Data.(type) {
		case string:
			previewFilePath = v
			// Try to find a buffer for this file
			for i := range fw.buffers {
				if fw.buffers[i].filename == previewFilePath {
					fileBuffer = &fw.buffers[i]
					highlightLine = fw.buffers[i].cursorY
					break
				}
			}
		case location:
			if strings.HasPrefix(v.URI, "file://") {
				previewFilePath = strings.TrimPrefix(v.URI, "file://")
				previewFilePath = filepath.FromSlash(previewFilePath)
				highlightLine = v.Range.Start.Line
			}
		}
	}
	if previewFilePath != "" {
		const maxPreviewLines = 20
		var lines []string
		var style editorStyle
		if fileBuffer != nil {
			lines = fileBuffer.lines
			style = fileBuffer.style
		} else {
			content, err := os.ReadFile(previewFilePath)
			if err == nil {
				lines = strings.Split(string(content), "\n")
				// Remove trailing empty line if file ends with newline
				if len(lines) > 0 && lines[len(lines)-1] == "" {
					lines = lines[:len(lines)-1]
				}
				// Create a temp buffer for highlighting
				tmpBuf := newBuffer(newEditorStyle(), bufferWithContent(previewFilePath, string(content)))
				style = tmpBuf.style
				fileBuffer = &tmpBuf
			}
		}
		// Center preview on highlightLine
		startLine := 0
		if highlightLine >= 0 && len(lines) > availableHeight {
			startLine = highlightLine - availableHeight/2
			if startLine < 0 {
				startLine = 0
			}
			if startLine+availableHeight > len(lines) {
				startLine = len(lines) - availableHeight
			}
		}
		for i := 0; i < availableHeight && startLine+i < len(lines); i++ {
			lineIdx := startLine + i
			var rendered string
			if fileBuffer != nil {
				rendered = fileBuffer.HighlightLine(lineIdx)
			} else {
				rendered = lines[lineIdx]
			}
			if lineIdx == highlightLine {
				rendered = style.previewHighlight.Render(expandTabs(rendered))
			} else {
				rendered = expandTabs(rendered)
			}
			if len(rendered) > previewWidth-2 {
				rendered = rendered[:previewWidth-5] + "..."
			}
			previewLines[i] = lipgloss.NewStyle().Width(previewWidth-2).Render(rendered)
		}
	}
	for i := range previewLines {
		if previewLines[i] == "" {
			previewLines[i] = strings.Repeat(" ", previewWidth-2)
		}
	}

	// Merge list and preview columns
	for i := 0; i < availableHeight; i++ {
		row := listLines[i] + " │ " + previewLines[i]
		lines = append(lines, row)
	}

	// Add status line
	statusText := fmt.Sprintf("%d/%d items", len(fw.filteredItems), len(fw.items))
	if fw.mode == FloatingWindowFilter {
		statusText += fmt.Sprintf(" (filtered: %s)", fw.filterText)
	}
	statusLine := fw.style.border.Render(statusText)
	if len([]rune(statusLine)) < fw.width-2 {
		statusLine += strings.Repeat(" ", fw.width-2-len([]rune(statusLine)))
	}
	lines = append(lines, statusLine)

	// Pad number of lines to fw.height-2
	for len(lines) < fw.height {
		lines = append(lines, strings.Repeat(" ", fw.width-2))
	}
	contentStr := strings.Join(lines, "\n")
	windowContent := fw.style.window.Width(fw.width).Height(fw.height).Render(contentStr)
	return windowContent
}

// GetSelectedItem returns the currently selected item
func (fw *FloatingWindow) GetSelectedItem() (FloatingWindowItem, bool) {
	if len(fw.filteredItems) == 0 || fw.selectedIndex >= len(fw.filteredItems) {
		return FloatingWindowItem{}, false
	}
	return fw.filteredItems[fw.selectedIndex], true
}

// SetItems updates the items in the floating window
func (fw *FloatingWindow) SetItems(items []FloatingWindowItem) *FloatingWindow {
	fw.items = items
	fw.applyFilter()
	return fw
}

// AddItem adds an item to the floating window
func (fw *FloatingWindow) AddItem(item FloatingWindowItem) *FloatingWindow {
	fw.items = append(fw.items, item)
	fw.applyFilter()
	return fw
}

// ClearItems clears all items from the floating window
func (fw *FloatingWindow) ClearItems() *FloatingWindow {
	fw.items = nil
	fw.filteredItems = nil
	fw.selectedIndex = 0
	return fw
}

// Helper for reference items
func referenceItemTitle(path string, line int) string {
	filename := path
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		filename = path[idx+1:]
	}
	return fmt.Sprintf("%s:%d", filename, line)
}

// Helper for reference items with column
func referenceItemTitleWithCol(path string, line, col int) string {
	filename := path
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		filename = path[idx+1:]
	}
	return fmt.Sprintf("%s:%d:%d", filename, line, col)
} 