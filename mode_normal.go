package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nm, mod, cmd := m.normalmode.Handle(msg, m)
		m = mod.(model)
		m.normalmode = nm

		return m, cmd
	}
	return m, nil
}
