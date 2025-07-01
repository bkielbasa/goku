package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (nm normalmode) commandCenterViewport(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	n := b.cursorY - (b.viewport.Height / 2)
	b = b.SetCursorYOffset(n)
	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm normalmode) commandTopViewport(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	n := b.cursorY
	b = b.SetCursorYOffset(n)
	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm normalmode) commandBottomViewport(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	n := b.cursorY - b.viewport.Height + 3
	b = b.SetCursorYOffset(n)
	m.buffers[m.currBuffer] = b

	return m, cmd
} 