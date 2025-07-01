package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type normalCommand func(m model, cmd tea.Cmd) (tea.Model, tea.Cmd)

type normalmode struct {
	commands    map[string]normalCommand
	buffer      string
	lastCommand string
}

func NewNormalMode() normalmode {
	nm := normalmode{}
	nm.commands = map[string]normalCommand{
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

		"dd": nm.commandDeleteLine,

		"esc": nm.commandClearBuffer,
		":":   nm.commandEnterCommandMode,
		"i":   nm.commandEnterInsertMode,
		"o":   nm.commandOpenLineBelow,
		"O":   nm.commandOpenLineAbove,

		"zt": nm.commandTopViewport,
		"zz": nm.commandCenterViewport,
		"zb": nm.commandBottomViewport,
	}

	return nm
}

func (nm normalmode) Handle(msg tea.KeyMsg, m model) (normalmode, tea.Model, tea.Cmd) {
	buff := nm.buffer + msg.String()
	nCommand, ok := nm.commands[buff]
	if ok {
		m, cmd := nCommand(m, nil)
		nm.lastCommand = buff
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