package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandEnterCommandMode(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.mode = ModeCommand
	return m, cmd
}

func (nm *normalmode) commandEnterInsertMode(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.mode = ModeInsert
	return m, cmd
}

func (nm *normalmode) commandClearBuffer(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	nm.buffer = ""
	return m, cmd
} 