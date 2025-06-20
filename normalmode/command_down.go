package normalmode

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandDown(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	if b.CursorY() < b.NoOfLines()-1 {
		b = b.IncreaseCursorY(1)
		if b.CursorY() >= b.CursorYOffset()+b.Viewport().Height-2 {
			b = b.IncreaseCursorYOffset(1)
		}
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandUp(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	if b.CursorY() > 0 {
		b = b.IncreaseCursorY(-1)
		if b.CursorY() < b.CursorYOffset() {
			b = b.IncreaseCursorYOffset(-1)
		}
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandLeft(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	if b.CursorX() > 0 {
		b = b.IncreaseCursorX(-1)
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandRight(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	if b.CursorX() < len(b.Line(b.CursorY())) {
		b = b.IncreaseCursorX(1)
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}
