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
	repeatableCommands map[string]bool
}

func NewNormalMode() *normalmode {
	nm := &normalmode{}
	nm.setupCommands()
	return nm
}

func (nm *normalmode) registerCmd(key string, cmd normalCommand) {
	nm.commands[key] = cmd
}

func (nm *normalmode) registerRepeatableCmd(key string, cmd normalCommand) {
	nm.commands[key] = cmd
	nm.repeatableCommands[key] = true
}

func (nm *normalmode) setupCommands() {
	nm.commands = make(map[string]normalCommand)
	nm.repeatableCommands = make(map[string]bool)
	
	// Navigation commands
	nm.registerCmd("j", nm.commandDown)
	nm.registerCmd("down", nm.commandDown)
	nm.registerCmd("k", nm.commandUp)
	nm.registerCmd("up", nm.commandUp)
	nm.registerCmd("h", nm.commandLeft)
	nm.registerCmd("left", nm.commandLeft)
	nm.registerCmd("l", nm.commandRight)
	nm.registerCmd("right", nm.commandRight)
	nm.registerCmd("w", nm.commandNextWord)
	nm.registerCmd("b", nm.commandPrevWord)
	
	// File navigation
	nm.registerCmd("gg", nm.commandGoToBeginingOfTheFile)
	nm.registerCmd("ge", nm.commandGoToEndOfTheFile)
	nm.registerCmd("gl", nm.commandGoToLast)
	nm.registerCmd("gs", nm.commandGoToFirstNonWhiteCharacter)
	
	// Editing commands
	nm.registerRepeatableCmd("dd", nm.commandDeleteLine)
	nm.registerRepeatableCmd("o", nm.commandOpenLineBelow)
	nm.registerRepeatableCmd("O", nm.commandOpenLineAbove)
	
	// Mode switching
	nm.registerCmd("esc", nm.commandClearBuffer)
	nm.registerCmd(":", nm.commandEnterCommandMode)
	nm.registerCmd("i", nm.commandEnterInsertMode)
	
	// Viewport commands
	nm.registerCmd("zt", nm.commandTopViewport)
	nm.registerCmd("zz", nm.commandCenterViewport)
	nm.registerCmd("zb", nm.commandBottomViewport)
	
	// Repeat command
	nm.registerCmd(".", nm.commandRepeat)
}

func (nm *normalmode) Handle(msg tea.KeyMsg, m model) (*normalmode, tea.Model, tea.Cmd) {
	buff := nm.buffer + msg.String()
	nCommand, ok := nm.commands[buff]
	if ok {
		m, cmd := nCommand(m, nil)
		// Only store repeatable commands
		if nm.repeatableCommands[buff] {
			nm.lastCommand = buff
		}
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

func (nm *normalmode) commandRepeat(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	// If there's no last command, do nothing
	if nm.lastCommand == "" {
		return m, cmd
	}

	// Find and execute the last command
	if nCommand, ok := nm.commands[nm.lastCommand]; ok {
		return nCommand(m, cmd)
	}

	return m, cmd
} 