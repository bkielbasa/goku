package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	buff := m.buffers[m.currBuffer]
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			m.mode = ModeCommand
		case "i":
			m.mode = ModeInsert
		case "j":
			if buff.cursorY < len(buff.lines)-1 {
				buff.cursorY++
				if buff.cursorY >= buff.cursorYOffset+buff.viewport.Height-2 {
					buff.cursorYOffset++
				}
			}
		case "k":
			if buff.cursorY > 0 {
				buff.cursorY--
				if buff.cursorY < buff.cursorYOffset {
					buff.cursorYOffset--
				}
			}
		case "h":
			if buff.cursorX > 0 {
				buff.cursorX--
			}
		case "l":
			if buff.cursorX < len(buff.lines[buff.cursorY]) {
				buff.cursorX++
			}
		}
	}
	m.buffers[m.currBuffer] = buff
	return m, nil
}
