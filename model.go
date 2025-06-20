package main

import (
	"fmt"
	"strings"

	"github.com/bkielbasa/goku/normalmode"
	tea "github.com/charmbracelet/bubbletea"
)

type editorMode string

const ModeNormal editorMode = "normal"
const ModeInsert editorMode = "insert"
const ModeCommand editorMode = "command"

type normalMode interface {
	Handle(msg tea.KeyMsg, m normalmode.EditorModel) (normalmode.NormalMode, tea.Model, tea.Cmd)
}

type model struct {
	mode          editorMode
	normalmode    normalMode
	commandBuffer string // Buffer for command mode input
	commands      []command
	viewport      tea.WindowSizeMsg

	buffers    []buffer
	currBuffer int

	style editorStyle
}

func initialModel() model {
	s := newEditorStyle()
	return model{
		mode:       ModeNormal,
		viewport:   tea.WindowSizeMsg{},
		normalmode: normalmode.New(),
		commands: []command{
			&commandQuit{},
			&commandOpen{},
		},
		style: s,

		buffers: []buffer{
			newBuffer(s),
		},
	}
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

func (m model) addBuffer(b buffer) model {
	b.viewport = m.viewport
	m.buffers = append(m.buffers, b)
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
