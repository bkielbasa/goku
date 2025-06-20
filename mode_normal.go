package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	buff := m.buffers[m.currBuffer]
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nm, mod, cmd := m.normalmode.Handle(msg, m)
		m = mod.(model)
		m.normalmode = nm
		return m, cmd

		// case "gs":
		// 	buff.cursorX = 0
		// 	m.normalModeBuffer = ""
		// case "gl":
		// 	buff.cursorX = len(buff.lines[buff.cursorY]) - 1
		// 	m.normalModeBuffer = ""
		// case "gg":
		// 	buff.cursorY = 0
		// 	buff.cursorYOffset = 0
		// 	m.normalModeBuffer = ""
		// case "G":
		// 	buff.cursorY = len(buff.lines) - 1
		// 	buff.cursorYOffset = len(buff.lines) - buff.viewport.Height + 2
		// 	m.normalModeBuffer = ""
		// default:
		// 	m.normalModeBuffer += msg.String()
		// }
	}
	m.buffers[m.currBuffer] = buff
	return m, nil
}
