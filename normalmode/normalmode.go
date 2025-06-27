package normalmode

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type Buffer interface {
	Viewport() tea.WindowSizeMsg

	CursorX() int
	SetCursorX(int) Buffer
	IncreaseCursorX(int) Buffer
	CursorY() int
	SetCursorY(int) Buffer
	IncreaseCursorY(int) Buffer

	CursorXOffset() int
	IncreaseCursorXOffset(int) Buffer
	CursorYOffset() int
	IncreaseCursorYOffset(int) Buffer

	NoOfLines() int
	Line(n int) string
	Lines() []string
	ReplaceLine(n int, s string) Buffer
	AppendLine(s string) Buffer
	InsertLine(n int, s string) Buffer
	DeleteLine(n int) Buffer

	SetStateModified() Buffer
	SetStateSaved() Buffer

	FileName() string
	SetFileName(f string) Buffer
}

type EditorModel interface {
	tea.Model

	CurrentBuffer() Buffer
	ReplaceCurrentBuffer(b Buffer) EditorModel

	EnterCommandMode() EditorModel
	EnterInsertMode() EditorModel
}

type NormalMode interface {
	Handle(msg tea.KeyMsg, m EditorModel) (NormalMode, tea.Model, tea.Cmd)
}

type command func(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd)

type normalmode struct {
	commands map[string]command
	buffer   string
}

func New() normalmode {
	nm := normalmode{}
	nm.commands = map[string]command{
		// basic navigation
		"j":     nm.commandDown,
		"down":  nm.commandDown,
		"k":     nm.commandUp,
		"up":    nm.commandUp,
		"h":     nm.commandLeft,
		"left":  nm.commandLeft,
		"l":     nm.commandRight,
		"right": nm.commandRight,
		"w":     nm.commandNextWord,
		"b":     nm.commandPrevWord,

		"gg": nm.commandGoToBeginingOfTheFile,
		"ge": nm.commandGoToEndOfTheFile,
		"gl": nm.commandGoToLast,
		"gs": nm.commandGoToFirstNonWhiteCharacter,

		"esc": nm.commandClearBuffer,
		":":   nm.commandEnterCommandMode,
		"i":   nm.commandEnterInsertMode,
		"o":   nm.commandOpenLineBelow,
		"O":   nm.commandOpenLineAbove,
	}

	return nm
}

func (nm normalmode) Handle(msg tea.KeyMsg, m EditorModel) (NormalMode, tea.Model, tea.Cmd) {
	buff := nm.buffer + msg.String()
	nCommand, ok := nm.commands[buff]
	if ok {
		m, cmd := nCommand(m, nil)
		nm.buffer = ""
		return nm, m, cmd
	}

	// let's check if there's any command that starts with the buffer
	found := false
	for k := range nm.commands {
		if strings.HasPrefix(k, buff) {
			found = true
			break
		}
	}

	if found {
		nm.buffer = buff
	} else {
		nm.buffer = ""
	}

	return nm, m, nil
}

func (nm normalmode) commandOpenLineBelow(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	buff := m.CurrentBuffer()
	
	// Insert a new empty line below the current line
	buff = buff.InsertLine(buff.CursorY()+1, "")
	
	// Move cursor to the new line
	buff = buff.IncreaseCursorY(1)
	buff = buff.SetCursorX(0)
	
	// Replace the buffer and enter insert mode
	m = m.ReplaceCurrentBuffer(buff)
	m = m.EnterInsertMode()
	
	return m, cmd
}

func (nm normalmode) commandOpenLineAbove(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	buff := m.CurrentBuffer()
	
	// Insert a new empty line above the current line
	buff = buff.InsertLine(buff.CursorY(), "")
	
	// Move cursor to the new line (cursor Y stays the same since we inserted above)
	buff = buff.SetCursorX(0)
	
	// Replace the buffer and enter insert mode
	m = m.ReplaceCurrentBuffer(buff)
	m = m.EnterInsertMode()
	
	return m, cmd
}
