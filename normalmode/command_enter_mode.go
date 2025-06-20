package normalmode

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandEnterCommandMode(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m = m.EnterCommandMode()
	return m, cmd
}

func (nm *normalmode) commandEnterInsertMode(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m = m.EnterInsertMode()
	return m, cmd
}
