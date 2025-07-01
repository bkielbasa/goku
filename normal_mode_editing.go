package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandDeleteLine(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	b = b.DeleteLine(b.cursorY)
	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm normalmode) commandOpenLineBelow(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	b = b.InsertLine(b.cursorY+1, "")

	b = b.IncreaseCursorY(1)
	b = b.SetCursorX(0)

	m.buffers[m.currBuffer] = b
	m.mode = ModeInsert

	return m, cmd
}

func (nm normalmode) commandOpenLineAbove(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	b = b.InsertLine(b.cursorY, "")
	b = b.SetCursorX(0)

	m.buffers[m.currBuffer] = b
	m.mode = ModeInsert

	return m, cmd
} 