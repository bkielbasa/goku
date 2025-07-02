package main

import (
	"fmt"
	"strings"

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

	var lines []string

	// Add title (only once)
	titleLine := fw.style.title.Render(fw.title)
	lines = append(lines, titleLine)

	// Add filter line if in filter mode
	if fw.mode == FloatingWindowFilter {
		filterLine := fw.style.filter.Render(fw.filterText)
		lines = append(lines, filterLine)
	}

	// Calculate available height for items
	availableHeight := fw.height - 2 // Account for title and status line
	if fw.mode == FloatingWindowFilter {
		availableHeight-- // Account for filter line
	}

	// Add items
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
	for i := startIndex; i < endIndex && i < len(fw.filteredItems); i++ {
		item := fw.filteredItems[i]
		var itemStyle lipgloss.Style
		if i == fw.selectedIndex {
			itemStyle = fw.style.selected
		} else {
			itemStyle = fw.style.item
		}
		title := item.Title
		if len(title) > fw.width-4 {
			title = title[:fw.width-7] + "..."
		}
		itemLine := itemStyle.Render(title)
		lines = append(lines, itemLine)
		if item.Subtitle != "" {
			subtitle := item.Subtitle
			if len(subtitle) > fw.width-6 {
				subtitle = subtitle[:fw.width-9] + "..."
			}
			subtitleLine := fw.style.subtitle.Render("  " + subtitle)
			lines = append(lines, subtitleLine)
		}
	}

	// Add status line
	statusText := fmt.Sprintf("%d/%d items", len(fw.filteredItems), len(fw.items))
	if fw.mode == FloatingWindowFilter {
		statusText += fmt.Sprintf(" (filtered: %s)", fw.filterText)
	}
	statusLine := fw.style.border.Render(statusText)
	lines = append(lines, statusLine)

	// Pad each line to fw.width-2 (account for border)
	padWidth := fw.width - 2
	for i, l := range lines {
		if len([]rune(l)) < padWidth {
			lines[i] = l + strings.Repeat(" ", padWidth-len([]rune(l)))
		} else if len([]rune(l)) > padWidth {
			lines[i] = string([]rune(l)[:padWidth])
		}
	}
	// Pad number of lines to fw.height-2
	for len(lines) < fw.height-2 {
		lines = append(lines, strings.Repeat(" ", padWidth))
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