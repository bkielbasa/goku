package normalmode

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandClearBuffer(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	nm.buffer = ""
	return m, cmd
}
