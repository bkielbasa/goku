package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type commandQuit struct {
}

func (c commandQuit) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	return m, tea.Quit
}

func (c commandQuit) Aliases() []string {
	return []string{"quit", "q"}
}
