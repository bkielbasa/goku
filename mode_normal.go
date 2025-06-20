package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	buff := m.buffers[m.currBuffer]
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.normalModeBuffer = ""
			break
		}
		cmd := m.normalModeBuffer + msg.String()
		switch cmd {
		case ":":
			m.mode = ModeCommand
		case "i":
			m.mode = ModeInsert
		case "down":
			fallthrough
		case "j":
			if buff.cursorY < len(buff.lines)-1 {
				buff.cursorY++
				if buff.cursorY >= buff.cursorYOffset+buff.viewport.Height-2 {
					buff.cursorYOffset++
				}
			}
		case "up":
			fallthrough
		case "k":
			if buff.cursorY > 0 {
				buff.cursorY--
				if buff.cursorY < buff.cursorYOffset {
					buff.cursorYOffset--
				}
			}
		case "left":
			fallthrough
		case "h":
			if buff.cursorX > 0 {
				buff.cursorX--
			}
		case "right":
			fallthrough
		case "l":
			if buff.cursorX < len(buff.lines[buff.cursorY]) {
				buff.cursorX++
			}
		case "gs":
			buff.cursorX = 0
			m.normalModeBuffer = ""
		case "gl":
			buff.cursorX = len(buff.lines[buff.cursorY]) - 1
			m.normalModeBuffer = ""
		case "gg":
			buff.cursorY = 0
			buff.cursorYOffset = 0
			m.normalModeBuffer = ""
		case "G":
			buff.cursorY = len(buff.lines) - 1
			buff.cursorYOffset = len(buff.lines) - buff.viewport.Height + 2
			m.normalModeBuffer = ""
		default:
			m.normalModeBuffer += msg.String()
		}
	}
	m.buffers[m.currBuffer] = buff
	return m, nil
}
