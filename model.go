package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type editorMode string

const ModeNormal editorMode = "normal"
const ModeInsert editorMode = "insert"
const ModeCommand editorMode = "command"
const ModeFloatingWindow editorMode = "floating_window"

type messageType string

const MessageInfo messageType = "info"
const MessageError messageType = "error"

type message struct {
	text    string
	msgType messageType
}

type toolInfo struct {
	Name       string
	IsInstalled bool
}

type languageSupport struct {
	Name        string
	LSPServer   toolInfo
	Formatter   toolInfo
	Highlighting toolInfo
}

type model struct {
	mode           editorMode
	normalmode     *normalmode
	commandBuffer  string // Buffer for command mode input
	commands       []command
	viewport       tea.WindowSizeMsg
	currentMessage *message

	buffers    []buffer
	currBuffer int

	style editorStyle

	// Language support info, keyed by language name or extension
	Languages map[string]languageSupport
	
	// LSP async state
	lspLoading bool
	lspError   string
	
	// Floating window
	floatingWindow *FloatingWindow
}

type modelOption func(*model)

func WithFile(filename string) modelOption {
	return func(m *model) {
		if filename != "" {
			// Try to load the file
			if loadedBuffer, err := loadFile(filename, m.style); err == nil {
				m.buffers[0] = loadedBuffer
			} else {
				// If file doesn't exist or can't be read, create a new buffer with the filename
				m.buffers[0] = newBuffer(m.style, bufferWithContent(filename, ""))
			}
		}
	}
}

func WithFiles(filenames []string) modelOption {
	return func(m *model) {
		if len(filenames) == 0 {
			return
		}
		
		// Clear the initial empty buffer
		m.buffers = []buffer{}
		
		// Create a buffer for each filename
		for _, filename := range filenames {
			if filename != "" {
				// Try to load the file
				if loadedBuffer, err := loadFile(filename, m.style); err == nil {
					m.buffers = append(m.buffers, loadedBuffer)
				} else {
					// If file doesn't exist or can't be read, create a new buffer with the filename
					m.buffers = append(m.buffers, newBuffer(m.style, bufferWithContent(filename, "")))
				}
			}
		}
		
		// Ensure we have at least one buffer
		if len(m.buffers) == 0 {
			m.buffers = []buffer{newBuffer(m.style)}
		}
	}
}

func initialModel(opts ...modelOption) model {
	s := newEditorStyle()

	m := model{
		mode:       ModeNormal,
		viewport:   tea.WindowSizeMsg{},
		normalmode: NewNormalMode(),
		commands: []command{
			&commandQuit{},
			&commandForceQuit{},
			&commandOpen{},
			&commandWrite{},
			&commandBufferNext{},
			&commandBufferPrev{},
			&commandBufferLast{},
			&commandBufferFirst{},
			&commandFindReferences{},
			&commandFileFinder{},
			&commandBufferList{},
		},
		style: s,

		buffers: []buffer{
			newBuffer(s),
		},

		Languages: make(map[string]languageSupport),
	}

	// Add default language support
	m.Languages["go"] = languageSupport{
		Name: "Go",
		LSPServer: toolInfo{Name: "gopls", IsInstalled: false},
		Formatter: toolInfo{Name: "gofmt", IsInstalled: false},
		Highlighting: toolInfo{Name: "builtin-go", IsInstalled: true},
	}
	m.Languages["py"] = languageSupport{
		Name: "Python",
		LSPServer: toolInfo{Name: "pyright", IsInstalled: false},
		Formatter: toolInfo{Name: "black", IsInstalled: false},
		Highlighting: toolInfo{Name: "builtin-python", IsInstalled: true},
	}
	m.Languages["js"] = languageSupport{
		Name: "JavaScript",
		LSPServer: toolInfo{Name: "typescript-language-server", IsInstalled: false},
		Formatter: toolInfo{Name: "prettier", IsInstalled: false},
		Highlighting: toolInfo{Name: "builtin-javascript", IsInstalled: true},
	}
	m.Languages["rs"] = languageSupport{
		Name: "Rust",
		LSPServer: toolInfo{Name: "rust-analyzer", IsInstalled: false},
		Formatter: toolInfo{Name: "rustfmt", IsInstalled: false},
		Highlighting: toolInfo{Name: "builtin-rust", IsInstalled: true},
	}
	m.Languages["c"] = languageSupport{
		Name: "C",
		LSPServer: toolInfo{Name: "clangd", IsInstalled: false},
		Formatter: toolInfo{Name: "clang-format", IsInstalled: false},
		Highlighting: toolInfo{Name: "builtin-c", IsInstalled: true},
	}

	// Check which tools are actually installed
	m.updateLanguageSupport()

	// Apply all options
	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// AsyncGoToDefinitionResult represents the result of an async go-to-definition
type asyncGoToDefinitionResult struct {
	location *location
	error    error
}

// AsyncGoToImplementationResult represents the result of an async go-to-implementation
type asyncGoToImplementationResult struct {
	location *location
	error    error
}

// asyncGoToImplementationInit represents the initialization step for implementation
type asyncGoToImplementationInit struct {
	filePath  string
	line      int
	character int
}

// asyncGoToImplementationOpenFiles represents the file opening step for implementation
type asyncGoToImplementationOpenFiles struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// asyncGoToImplementationWait represents the waiting step for implementation
type asyncGoToImplementationWait struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// asyncGoToImplementationRequest represents the actual LSP request step for implementation
type asyncGoToImplementationRequest struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// AsyncGoToTypeDefinitionResult represents the result of an async go-to-type-definition
type asyncGoToTypeDefinitionResult struct {
	location *location
	error    error
}

// asyncGoToTypeDefinitionInit represents the initialization step for type definition
type asyncGoToTypeDefinitionInit struct {
	filePath  string
	line      int
	character int
}

// asyncGoToTypeDefinitionOpenFiles represents the file opening step for type definition
type asyncGoToTypeDefinitionOpenFiles struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// asyncGoToTypeDefinitionWait represents the waiting step for type definition
type asyncGoToTypeDefinitionWait struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// asyncGoToTypeDefinitionRequest represents the actual LSP request step for type definition
type asyncGoToTypeDefinitionRequest struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// Async messages for find references
type asyncFindReferencesOpenFiles struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

type asyncFindReferencesWait struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

type asyncFindReferencesRequest struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

type goToLocationMsg struct {
	location location
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) CurrentBuffer() buffer {
	return m.buffers[m.currBuffer]
}

func (m model) EnterCommandMode() model {
	m.mode = ModeCommand
	return m
}

func (m model) EnterInsertMode() model {
	m.mode = ModeInsert
	return m
}

func (m model) ReplaceCurrentBuffer(b buffer) model {
	m.buffers[m.currBuffer] = b
	return m
}

func (m model) SetInfoMessage(text string) model {
	m.currentMessage = &message{text: text, msgType: MessageInfo}
	return m
}

func (m model) SetErrorMessage(text string) model {
	m.currentMessage = &message{text: text, msgType: MessageError}
	return m
}

func (m model) ClearMessage() model {
	m.currentMessage = nil
	return m
}

func (m model) addBuffer(b buffer) model {
	b.viewport = m.viewport
	m.buffers = append(m.buffers, b)
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Clear the message on any key event (except window resize)
	if m.currentMessage != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			m = m.ClearMessage()
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport = msg
		m.buffers[m.currBuffer].viewport = msg
		// Update floating window size/position if open
		if m.floatingWindow != nil && m.floatingWindow.IsOpen() {
			margin := 10
			windowWidth := m.viewport.Width - 2*margin
			windowHeight := m.viewport.Height - 2*margin
			if windowWidth < 20 {
				windowWidth = 20
			}
			if windowHeight < 5 {
				windowHeight = 5
			}
			posX := margin
			posY := margin
			m.floatingWindow.SetPosition(posX, posY)
			m.floatingWindow.SetSize(windowWidth, windowHeight)
		}
		return m, nil
	case asyncGoToDefinitionResult:
		m.lspLoading = false
		if msg.error != nil {
			m.lspError = msg.error.Error()
			return m, nil
		}
		if msg.location != nil {
			// Handle successful go-to-definition
			return m.handleGoToDefinitionResult(msg.location), nil
		}
		return m, nil
	case asyncGoToDefinitionInit:
		// Initialize LSP client
		client, err := m.getPersistentLSPClient(msg.filePath)
		if err != nil {
			return m, func() tea.Msg {
				return asyncGoToDefinitionResult{error: err}
			}
		}
		
		// Return command to open files
		return m, func() tea.Msg {
			return asyncGoToDefinitionOpenFiles{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    client,
			}
		}
	case asyncGoToDefinitionOpenFiles:
		// Open files in a goroutine to avoid blocking
		return m, func() tea.Msg {
			// Open files in background
			go func() {
				workspaceRoot := filepath.Dir(msg.filePath)
				absWorkspaceRoot, err := filepath.Abs(workspaceRoot)
				if err != nil {
					return
				}

				// Open all Go files in the workspace for better indexing
				filepath.Walk(absWorkspaceRoot, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(path, ".go") {
						content, err := os.ReadFile(path)
						if err != nil {
							return nil // Continue with other files
						}
						msg.client.OpenDocument(path, string(content))
					}
					return nil
				})
			}()
			
			// Return command to wait for server
			return asyncGoToDefinitionWait{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncGoToDefinitionWait:
		// Wait for server to be ready with a shorter timeout
		return m, func() tea.Msg {
			if !msg.client.WaitForReady(5 * time.Second) {
				return asyncGoToDefinitionResult{error: fmt.Errorf("LSP server not ready after 5 seconds")}
			}
			
			// Return command to make the actual request
			return asyncGoToDefinitionRequest{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncGoToDefinitionRequest:
		// Make the actual LSP request
		return m, func() tea.Msg {
			// Calculate UTF-16 offset for the cursor position
			b := m.buffers[m.currBuffer]
			lineStr := b.Line(msg.line)
			utf16Offset := utf16Index(lineStr, msg.character)

			location, err := msg.client.GoToDefinition(msg.filePath, msg.line, utf16Offset)
			return asyncGoToDefinitionResult{location: location, error: err}
		}
	case asyncGoToImplementationResult:
		m.lspLoading = false
		if msg.error != nil {
			m.lspError = msg.error.Error()
			return m, nil
		}
		if msg.location != nil {
			// Handle successful go-to-implementation
			return m.handleGoToDefinitionResult(msg.location), nil
		}
		return m, nil
	case asyncGoToImplementationInit:
		// Initialize LSP client
		client, err := m.getPersistentLSPClient(msg.filePath)
		if err != nil {
			return m, func() tea.Msg {
				return asyncGoToImplementationResult{error: err}
			}
		}
		
		// Return command to open files
		return m, func() tea.Msg {
			return asyncGoToImplementationOpenFiles{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    client,
			}
		}
	case asyncGoToImplementationOpenFiles:
		// Open files in a goroutine to avoid blocking
		return m, func() tea.Msg {
			// Open files in background
			go func() {
				workspaceRoot := filepath.Dir(msg.filePath)
				absWorkspaceRoot, err := filepath.Abs(workspaceRoot)
				if err != nil {
					return
				}

				// Open all Go files in the workspace for better indexing
				filepath.Walk(absWorkspaceRoot, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(path, ".go") {
						content, err := os.ReadFile(path)
						if err != nil {
							return nil // Continue with other files
						}
						msg.client.OpenDocument(path, string(content))
					}
					return nil
				})
			}()
			
			// Return command to wait for server
			return asyncGoToImplementationWait{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncGoToImplementationWait:
		// Wait for server to be ready with a shorter timeout
		return m, func() tea.Msg {
			if !msg.client.WaitForReady(5 * time.Second) {
				return asyncGoToImplementationResult{error: fmt.Errorf("LSP server not ready after 5 seconds")}
			}
			
			// Return command to make the actual request
			return asyncGoToImplementationRequest{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncGoToImplementationRequest:
		// Make the actual LSP request
		return m, func() tea.Msg {
			// Calculate UTF-16 offset for the cursor position
			b := m.buffers[m.currBuffer]
			lineStr := b.Line(msg.line)
			utf16Offset := utf16Index(lineStr, msg.character)

			location, err := msg.client.GoToImplementation(msg.filePath, msg.line, utf16Offset)
			return asyncGoToImplementationResult{location: location, error: err}
		}
	case asyncGoToTypeDefinitionResult:
		m.lspLoading = false
		if msg.error != nil {
			m.lspError = msg.error.Error()
			return m, nil
		}
		if msg.location != nil {
			// Handle successful go-to-type-definition
			return m.handleGoToDefinitionResult(msg.location), nil
		}
		return m, nil
	case asyncGoToTypeDefinitionInit:
		// Initialize LSP client
		client, err := m.getPersistentLSPClient(msg.filePath)
		if err != nil {
			return m, func() tea.Msg {
				return asyncGoToTypeDefinitionResult{error: err}
			}
		}
		
		// Return command to open files
		return m, func() tea.Msg {
			return asyncGoToTypeDefinitionOpenFiles{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    client,
			}
		}
	case asyncGoToTypeDefinitionOpenFiles:
		// Open files in a goroutine to avoid blocking
		return m, func() tea.Msg {
			// Open files in background
			go func() {
				workspaceRoot := filepath.Dir(msg.filePath)
				absWorkspaceRoot, err := filepath.Abs(workspaceRoot)
				if err != nil {
					return
				}

				// Open all Go files in the workspace for better indexing
				filepath.Walk(absWorkspaceRoot, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(path, ".go") {
						content, err := os.ReadFile(path)
						if err != nil {
							return nil // Continue with other files
						}
						msg.client.OpenDocument(path, string(content))
					}
					return nil
				})
			}()
			
			// Return command to wait for server
			return asyncGoToTypeDefinitionWait{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncGoToTypeDefinitionWait:
		// Wait for server to be ready with a shorter timeout
		return m, func() tea.Msg {
			if !msg.client.WaitForReady(5 * time.Second) {
				return asyncGoToTypeDefinitionResult{error: fmt.Errorf("LSP server not ready after 5 seconds")}
			}
			
			// Return command to make the actual request
			return asyncGoToTypeDefinitionRequest{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncGoToTypeDefinitionRequest:
		// Handle async go-to-type-definition request
		return m, func() tea.Msg {
			return asyncGoToTypeDefinitionResult{
				location: nil,
				error:    fmt.Errorf("go-to-type-definition not implemented yet"),
			}
		}
	case openFileMsg:
		// Handle opening a file from floating window
		newBuf, err := loadFile(msg.filePath, m.style)
		if err != nil {
			return m.SetErrorMessage(fmt.Sprintf("Failed to open file: %v", err)), nil
		}
		m.buffers = append(m.buffers, newBuf)
		m.currBuffer = len(m.buffers) - 1
		m.buffers[m.currBuffer].viewport = m.viewport
		m.floatingWindow = nil
		m.mode = ModeNormal
		return m, nil
	case switchBufferMsg:
		// Handle switching to a buffer from floating window
		if msg.bufferIndex >= 0 && msg.bufferIndex < len(m.buffers) {
			m.currBuffer = msg.bufferIndex
			m.buffers[m.currBuffer].viewport = m.viewport
		}
		m.floatingWindow = nil
		m.mode = ModeNormal
		return m, nil
	case asyncFindReferencesInit:
		// Initialize LSP client for find references
		client, err := m.getPersistentLSPClient(msg.filePath)
		if err != nil {
			return m, func() tea.Msg {
				return asyncFindReferencesResult{error: err}
			}
		}
		
		// Return command to open files
		return m, func() tea.Msg {
			return asyncFindReferencesOpenFiles{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    client,
			}
		}
	case asyncFindReferencesOpenFiles:
		// Open files in LSP client
		err := msg.client.OpenFile(msg.filePath)
		if err != nil {
			return m, func() tea.Msg {
				return asyncFindReferencesResult{error: err}
			}
		}
		
		// Return command to wait a bit for LSP to process
		return m, func() tea.Msg {
			return asyncFindReferencesWait{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncFindReferencesWait:
		// Wait a bit for LSP to process, then make the request
		return m, func() tea.Msg {
			return asyncFindReferencesRequest{
				filePath:  msg.filePath,
				line:      msg.line,
				character: msg.character,
				client:    msg.client,
			}
		}
	case asyncFindReferencesRequest:
		// Make the find references request
		locations, err := msg.client.FindReferences(msg.filePath, msg.line, msg.character)
		return m, func() tea.Msg {
			return asyncFindReferencesResult{
				locations: locations,
				error:     err,
			}
		}
	case asyncFindReferencesResult:
		if msg.error != nil {
			return m.SetErrorMessage(fmt.Sprintf("Find references failed: %v", msg.error)), nil
		}
		
		// Convert locations to floating window items
		var items []FloatingWindowItem
		for i, loc := range msg.locations {
			filePath := ""
			if strings.HasPrefix(loc.URI, "file://") {
				filePath = strings.TrimPrefix(loc.URI, "file://")
				filePath = filepath.FromSlash(filePath)
			}
			fileNameLineCol := referenceItemTitleWithCol(filePath, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
			item := FloatingWindowItem{
				ID:    fmt.Sprintf("ref_%d", i),
				Title: fileNameLineCol,
				Data:  loc,
			}
			items = append(items, item)
		}
		
		if len(items) == 0 {
			return m.SetInfoMessage("No references found"), nil
		}
		
		// Create floating window with callbacks
		fw := NewFloatingWindow("Find References", items, m.viewport, 10).Open()
		
		// Set up callbacks for location selection
		fw.SetCallbacks(
			// onSelect callback - go to selected location
			func(item FloatingWindowItem) tea.Cmd {
				if location, ok := item.Data.(location); ok {
					return func() tea.Msg {
						return goToLocationMsg{location: location}
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
	case goToLocationMsg:
		// Handle going to a location from find references
		m = m.handleGoToDefinitionResult(&msg.location)
		m.floatingWindow = nil
		m.mode = ModeNormal
		return m, nil
	case closeFloatingWindowMsg:
		m.floatingWindow = nil
		m.mode = ModeNormal
		return m, nil
	}

	switch m.mode {
	case ModeNormal:
		return m.updateNormal(msg)
	case ModeInsert:
		return m.updateInsert(msg)
	case ModeCommand:
		return m.updateCommand(msg)
	case ModeFloatingWindow:
		return m.updateFloatingWindow(msg)
	}
	return m, nil
}

func (m model) View() string {
	bufferContent := m.buffers[m.currBuffer].View()

	// Build the status bar content
	var statusBarContent string
	if m.mode == ModeCommand {
		statusBarContent = fmt.Sprintf(":%s", m.commandBuffer)
	} else {
		buf := m.buffers[m.currBuffer]
		f := fileNameLabel(buf.filename, buf.state)

		buff := fmt.Sprintf("%s ", strings.ToUpper(string(m.mode))) + f
		if len(m.buffers) > 1 {
			buff += fmt.Sprintf(" [%d/%d]", m.currBuffer+1, len(m.buffers))
		}

		lspStatus := ""
		langExt := ""
		if buf.filename != "" {
			parts := strings.Split(buf.filename, ".")
			if len(parts) > 1 {
				langExt = parts[len(parts)-1]
			}
		}
		if langExt != "" {
			if lang, ok := m.Languages[langExt]; ok {
				lspName := lang.LSPServer.Name
				if lspName != "" {
					if lang.LSPServer.IsInstalled {
						if m.lspLoading {
							lspStatus = fmt.Sprintf("%s⏳ ", lspName)
						} else {
							lspStatus = fmt.Sprintf("%s✅ ", lspName)
						}
					} else {
						lspStatus = fmt.Sprintf("%s❌ ", lspName)
					}
				}
			}
		}
		if m.lspError != "" {
			lspStatus += fmt.Sprintf("ERR: %s ", m.lspError)
		}

		posInfo := filePossitionInfo(buf.cursorY+1, buf.cursorX+1)
		width := m.CurrentBuffer().Viewport().Width
		pad := width - len(buff) - len(lspStatus) - len(posInfo)
		if pad < 1 {
			pad = 1
		}
		statusBarContent = m.style.statusBar.Render(buff + strings.Repeat(" ", pad) + lspStatus + posInfo)
	}

	var messageContent string
	if m.currentMessage != nil {
		var messageStyle lipgloss.Style
		switch m.currentMessage.msgType {
		case MessageInfo:
			messageStyle = m.style.messageInfo
		case MessageError:
			messageStyle = m.style.messageError
		}
		messageText := m.currentMessage.text
		if len(messageText) > m.viewport.Width {
			messageText = messageText[:m.viewport.Width-3] + "..."
		}
		messageContent = messageStyle.Render(messageText)
	}

	availableHeight := m.viewport.Height
	if messageContent != "" {
		availableHeight -= 1
	}
	availableHeight -= 1
	if availableHeight < 1 {
		availableHeight = 1
	}

	bufferLines := strings.Split(bufferContent, "\n")
	if len(bufferLines) > availableHeight {
		bufferLines = bufferLines[:availableHeight]
	}
	for len(bufferLines) < availableHeight {
		bufferLines = append(bufferLines, "")
	}

	// Overlay floating window if open
	if m.floatingWindow != nil && m.floatingWindow.IsOpen() {
		fwLines := strings.Split(m.floatingWindow.View(), "\n")
		fwHeight := len(fwLines)
		fwWidth := 0
		for _, l := range fwLines {
			if len(l) > fwWidth {
				fwWidth = len(l)
			}
		}
		// Center the window
		startY := (availableHeight-fwHeight)/2
		if startY < 0 {
			startY = 0
		}
		for i := 0; i < fwHeight && (startY+i) < len(bufferLines); i++ {
			// Overlay the line (replace the whole line)
			bufferLines[startY+i] = fwLines[i]
		}
	}

	content := strings.Join(bufferLines, "\n")

	var result strings.Builder
	result.WriteString(content)
	if messageContent != "" {
		result.WriteRune('\n')
		result.WriteString(messageContent)
	}
	result.WriteRune('\n')
	result.WriteString(statusBarContent)
	return result.String()
}

func fileNameLabel(filename string, s bufferState) string {
	switch s {
	case bufferStateUnnamed:
		return "[No name]"
	case bufferStateSaved:
		return filename
	case bufferStateModified:
		if filename != "" {
			return filename + "*"
		}
		return "[No name]*"
	case bufferStateReadOnly:
		return filename + " (readonly)"
	}

	return "not implemented yet"
}

func filePossitionInfo(line, cur int) string {
	return fmt.Sprintf("%d:%d", line, cur)
}

// isToolInstalled checks if a tool is available in the system PATH
func isToolInstalled(toolName string) bool {
	// For now, we'll do a simple check using exec.LookPath
	// In a real implementation, you might want more sophisticated detection
	_, err := exec.LookPath(toolName)
	return err == nil
}

// updateLanguageSupport checks and updates the installation status of tools
func (m *model) updateLanguageSupport() {
	for lang, support := range m.Languages {
		// Check LSP server
		if support.LSPServer.Name != "" {
			support.LSPServer.IsInstalled = isToolInstalled(support.LSPServer.Name)
		}
		
		// Check formatter
		if support.Formatter.Name != "" {
			support.Formatter.IsInstalled = isToolInstalled(support.Formatter.Name)
		}
		
		// Update the map
		m.Languages[lang] = support
	}
}

// getPersistentLSPClient returns a persistent LSP client for the file, using the global LSP client manager
func (m *model) getPersistentLSPClient(filePath string) (*lspClient, error) {
	// Find language
	ext := strings.ToLower(filepath.Ext(filePath))
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	lang, exists := m.Languages[ext]
	if !exists {
		return nil, fmt.Errorf("no language support found for extension: %s", ext)
	}
	if !lang.LSPServer.IsInstalled {
		return nil, fmt.Errorf("LSP server %s is not installed", lang.LSPServer.Name)
	}
	
	// Use the global LSP client manager
	rootPath := filepath.Dir(filePath)
	return lspClientManager.GetLSPClient(lang.LSPServer.Name, rootPath)
}

// AsyncGoToDefinition creates an async command for go-to-definition
func (m *model) AsyncGoToDefinition(filePath string, line, character int) tea.Cmd {
	return func() tea.Msg {
		// Start the async process by returning a command to initialize the LSP client
		return asyncGoToDefinitionInit{
			filePath: filePath,
			line:     line,
			character: character,
		}
	}
}

// asyncGoToDefinitionInit represents the initialization step
type asyncGoToDefinitionInit struct {
	filePath  string
	line      int
	character int
}

// asyncGoToDefinitionOpenFiles represents the file opening step
type asyncGoToDefinitionOpenFiles struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// asyncGoToDefinitionWait represents the waiting step
type asyncGoToDefinitionWait struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// asyncGoToDefinitionRequest represents the actual LSP request step
type asyncGoToDefinitionRequest struct {
	filePath  string
	line      int
	character int
	client    *lspClient
}

// handleGoToDefinitionResult processes the result of a go-to-definition request
func (m model) handleGoToDefinitionResult(location *location) model {
	fileURI := location.URI
	filePath := ""
	if strings.HasPrefix(fileURI, "file://") {
		filePath = strings.TrimPrefix(fileURI, "file://")
		filePath = filepath.FromSlash(filePath)
	}

	// Normalize the target file path to absolute
	absTargetPath, err := filepath.Abs(filePath)
	if err != nil {
		m.lspError = fmt.Sprintf("Failed to get absolute path: %v", err)
		return m
	}

	bufferIndex := -1
	for i, buf := range m.buffers {
		// Normalize the buffer's file path to absolute for comparison
		bufPath := buf.FileName()
		if bufPath != "" {
			absBufPath, err := filepath.Abs(bufPath)
			if err == nil && absBufPath == absTargetPath {
				bufferIndex = i
				break
			}
		}
	}

	if bufferIndex == -1 {
		newBuf, err := loadFile(filePath, m.style)
		if err != nil {
			m.lspError = fmt.Sprintf("Failed to load file: %v", err)
			return m
		}
		m.buffers = append(m.buffers, newBuf)
		bufferIndex = len(m.buffers) - 1
	}

	m.currBuffer = bufferIndex
	b := m.buffers[bufferIndex]
	// Update viewport for the new buffer
	b.viewport = m.viewport
	b = b.SetCursorY(location.Range.Start.Line)
	// Convert UTF-16 offset to rune index for cursor X
	defLine := b.Line(location.Range.Start.Line)
	runeX := runeIndexFromUTF16(defLine, location.Range.Start.Character)
	b = b.SetCursorX(runeX)
	m.buffers[bufferIndex] = b

	return m
}

// AsyncGoToImplementation creates an async command for go-to-implementation
func (m *model) AsyncGoToImplementation(filePath string, line, character int) tea.Cmd {
	return func() tea.Msg {
		// Start the async process by returning a command to initialize the LSP client
		return asyncGoToImplementationInit{
			filePath: filePath,
			line:     line,
			character: character,
		}
	}
}

// AsyncGoToTypeDefinition creates an async command for go-to-type-definition
func (m *model) AsyncGoToTypeDefinition(filePath string, line, character int) tea.Cmd {
	return func() tea.Msg {
		return asyncGoToTypeDefinitionInit{
			filePath: filePath,
			line:     line,
			character: character,
		}
	}
}

// updateFloatingWindow handles input when in floating window mode
func (m model) updateFloatingWindow(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.floatingWindow == nil {
		// If no floating window, return to normal mode
		m.mode = ModeNormal
		return m, nil
	}

	// Update the floating window
	updatedWindow, cmd := m.floatingWindow.Update(msg)
	m.floatingWindow = updatedWindow

	// If the window was closed, return to normal mode
	if !m.floatingWindow.IsOpen() {
		m.mode = ModeNormal
		m.floatingWindow = nil
	}

	return m, cmd
}

// OpenFloatingWindow opens a floating window with the given items
func (m model) OpenFloatingWindow(title string, items []FloatingWindowItem) model {
	margin := 10
	fw := NewFloatingWindow(title, items, m.viewport, margin).Open()
	m.floatingWindow = fw
	m.mode = ModeFloatingWindow
	return m
}

// CloseFloatingWindow closes the current floating window
func (m model) CloseFloatingWindow() model {
	if m.floatingWindow != nil {
		m.floatingWindow.Close()
		m.floatingWindow = nil
	}
	m.mode = ModeNormal
	return m
}

// IsFloatingWindowOpen returns true if a floating window is currently open
func (m model) IsFloatingWindowOpen() bool {
	return m.floatingWindow != nil && m.floatingWindow.IsOpen()
}
