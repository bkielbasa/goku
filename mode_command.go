package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"slices"
	"strings"
)

type command interface {
	Aliases() []string
	Update(m model, msg tea.Msg, args []string) (model, tea.Cmd)
}

func (m model) updateCommand(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.mode = ModeNormal
			m.commandBuffer = ""
		case tea.KeyEnter:
			args := strings.Split(m.commandBuffer, " ")
			cmd := strings.TrimSpace(args[0])
			for _, c := range m.commands {
				if slices.Contains(c.Aliases(), cmd) {
					return c.Update(m, msg, args[1:])
				}
			}
			m.commandBuffer = ""
			m.mode = ModeNormal
		case tea.KeyBackspace:
			if len(m.commandBuffer) > 0 {
				m.commandBuffer = m.commandBuffer[:len(m.commandBuffer)-1]
			}
		default:
			if msg.String() != "" && len(msg.String()) == 1 {
				m.commandBuffer += msg.String()
			}
		}
	}
	return m, nil
}
