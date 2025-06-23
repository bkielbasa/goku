package main

import (
	"fmt"
	"strings"

	"github.com/bkielbasa/goku/normalmode"
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

type normalMode interface {
	Handle(msg tea.KeyMsg, m normalmode.EditorModel) (normalmode.NormalMode, tea.Model, tea.Cmd)
}

type model struct {
	mode           editorMode
	normalmode     normalMode
	commandBuffer  string // Buffer for command mode input
	commands       []command
	viewport       tea.WindowSizeMsg
	currentMessage *message

	buffers    []buffer
	currBuffer int

	style editorStyle
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

func initialModel(opts ...modelOption) model {
	s := newEditorStyle()

	m := model{
		mode:       ModeNormal,
		viewport:   tea.WindowSizeMsg{},
		normalmode: normalmode.New(),
		commands: []command{
			&commandQuit{},
			&commandForceQuit{},
			&commandOpen{},
			&commandWrite{},
		},
		style: s,

		buffers: []buffer{
			newBuffer(s),
		},
	}

	// Apply all options
	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) CurrentBuffer() normalmode.Buffer {
	return m.buffers[m.currBuffer]
}

func (m model) EnterCommandMode() normalmode.EditorModel {
	m.mode = ModeCommand
	return m
}

func (m model) EnterInsertMode() normalmode.EditorModel {
	m.mode = ModeInsert
	return m
}

func (m model) ReplaceCurrentBuffer(b normalmode.Buffer) normalmode.EditorModel {
	m.buffers[m.currBuffer] = b.(buffer)
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
	var b strings.Builder
	b.WriteString(m.buffers[m.currBuffer].View())

	if m.mode == ModeCommand {
		b.WriteString(fmt.Sprintf(":%s", m.commandBuffer))
	} else {
		buf := m.buffers[m.currBuffer]
		f := fileNameLabel(buf.filename, buf.state)

		buff := fmt.Sprintf("%s ", strings.ToUpper(string(m.mode))) + f
		posInfo := filePossitionInfo(buf.cursorY+1, buf.cursorX+1)
		width := m.CurrentBuffer().Viewport().Width

		pad := width - len(buff) - len(posInfo)
		if pad < 1 {
			pad = 1
		}

		b.WriteString(m.style.statusBar.Render(buff + strings.Repeat(" ", pad) + posInfo))
	}

	// Display message if present
	if m.currentMessage != nil {
		b.WriteRune('\n')
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

		b.WriteString(messageStyle.Render(messageText))
	}

	return b.String()
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
