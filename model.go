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

// AsyncGoToDefinitionCmd represents an async go-to-definition command
type asyncGoToDefinitionCmd struct {
	filePath string
	line     int
	character int
}

// AsyncGoToDefinitionResult represents the result of an async go-to-definition
type asyncGoToDefinitionResult struct {
	location *location
	error    error
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
	}

	switch m.mode {
	case ModeNormal:
		return m.updateNormal(msg)
	case ModeInsert:
		return m.updateInsert(msg)
	case ModeCommand:
		return m.updateCommand(msg)
	}
	return m, nil
}

func (m model) View() string {
	// Get the buffer content
	bufferContent := m.buffers[m.currBuffer].View()
	
	// Build the status bar content
	var statusBarContent string
	if m.mode == ModeCommand {
		statusBarContent = fmt.Sprintf(":%s", m.commandBuffer)
	} else {
		buf := m.buffers[m.currBuffer]
		f := fileNameLabel(buf.filename, buf.state)

		buff := fmt.Sprintf("%s ", strings.ToUpper(string(m.mode))) + f
		
		// Add buffer information if there are multiple buffers
		if len(m.buffers) > 1 {
			buff += fmt.Sprintf(" [%d/%d]", m.currBuffer+1, len(m.buffers))
		}

		// LSP status info
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

		// Add LSP error if any
		if m.lspError != "" {
			lspStatus += fmt.Sprintf("ERR: %s ", m.lspError)
		}

		posInfo := filePossitionInfo(buf.cursorY+1, buf.cursorX+1)
		width := m.CurrentBuffer().Viewport().Width

		// Compose status bar: left | lspStatus | right
		pad := width - len(buff) - len(lspStatus) - len(posInfo)
		if pad < 1 {
			pad = 1
		}

		statusBarContent = m.style.statusBar.Render(buff + strings.Repeat(" ", pad) + lspStatus + posInfo)
	}

	// Build the message content if present
	var messageContent string
	if m.currentMessage != nil {
		var messageStyle lipgloss.Style
		switch m.currentMessage.msgType {
		case MessageInfo:
			messageStyle = m.style.messageInfo
		case MessageError:
			messageStyle = m.style.messageError
		}

		// Truncate message if it's too long for the viewport
		messageText := m.currentMessage.text
		if len(messageText) > m.viewport.Width {
			messageText = messageText[:m.viewport.Width-3] + "..."
		}

		messageContent = messageStyle.Render(messageText)
	}

	// Calculate available height for content (viewport height minus status bar and message)
	availableHeight := m.viewport.Height
	if messageContent != "" {
		availableHeight -= 1 // Message takes one line
	}
	availableHeight -= 1 // Status bar takes one line

	// Ensure availableHeight is at least 1 to prevent slice bounds errors
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Split buffer content into lines and ensure it fits within available height
	bufferLines := strings.Split(bufferContent, "\n")
	if len(bufferLines) > availableHeight {
		bufferLines = bufferLines[:availableHeight]
	}

	// Pad the content to fill the available height
	for len(bufferLines) < availableHeight {
		bufferLines = append(bufferLines, "")
	}

	// Join the content lines
	content := strings.Join(bufferLines, "\n")

	// Build the final layout
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
		client, err := m.getPersistentLSPClient(filePath)
		if err != nil {
			return asyncGoToDefinitionResult{error: err}
		}

		// Get the workspace root directory
		workspaceRoot := filepath.Dir(filePath)
		absWorkspaceRoot, err := filepath.Abs(workspaceRoot)
		if err != nil {
			return asyncGoToDefinitionResult{error: err}
		}

		// Open all Go files in the workspace for better indexing
		err = filepath.Walk(absWorkspaceRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				content, err := os.ReadFile(path)
				if err != nil {
					return nil // Continue with other files
				}
				err = client.OpenDocument(path, string(content))
				if err != nil {
					// Continue with other files
				}
			}
			return nil
		})
		if err != nil {
			return asyncGoToDefinitionResult{error: err}
		}
		
		// Wait much longer for the server to fully load packages
		if !client.WaitForReady(30 * time.Second) {
			return asyncGoToDefinitionResult{error: fmt.Errorf("LSP server not ready after 30 seconds")}
		}

		// Additional wait to ensure packages are fully indexed
		time.Sleep(5 * time.Second)

		// Calculate UTF-16 offset for the cursor position
		b := m.buffers[m.currBuffer]
		lineStr := b.Line(line)
		utf16Offset := utf16Index(lineStr, character)

		// Debug: show what character we're on
		if character < len([]rune(lineStr)) {
			charAtCursor := []rune(lineStr)[character]
			_ = charAtCursor // Suppress unused variable warning
		}
		
		// Debug: show the identifier at cursor position
		identifier := extractIdentifierAt(lineStr, character)
		_ = identifier // Suppress unused variable warning

		location, err := client.GoToDefinition(filePath, line, utf16Offset)
		return asyncGoToDefinitionResult{location: location, error: err}
	}
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
