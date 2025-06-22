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

	CursorYOffset() int
	IncreaseCursorYOffset(int) Buffer

	NoOfLines() int
	Line(n int) string
	Lines() []string
	ReplaceLine(n int, s string) Buffer
	AppendLine(s string) Buffer

	SetStateModified() Buffer

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

		"esc": nm.commandClearBuffer,
		":":   nm.commandEnterCommandMode,
		"i":   nm.commandEnterInsertMode,
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
