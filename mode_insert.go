package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateInsert(msg tea.Msg) (tea.Model, tea.Cmd) {
	buff := m.currentBuffer()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			fallthrough
		case "alt+esc":
			m.mode = ModeNormal
			m.commandBuffer = ""
		case "enter":
			buff.lines = append(buff.lines, "")
			buff.cursorY++
			buff.cursorX = 0
		case "left":
			if buff.cursorX > 0 {
				buff.cursorX--
			}
		case "right":
			if buff.cursorX < len(buff.lines[buff.cursorY]) {
				buff.cursorX++
			}
		case "up":
			if buff.cursorY > 0 {
				buff.cursorY--
			}
		case "down":
			if buff.cursorY < len(buff.lines)-1 {
				buff.cursorY++
			}
		case "backspace":
			if buff.cursorX > 0 {
				buff.lines[buff.cursorY] = buff.lines[buff.cursorY][:buff.cursorX-1] + buff.lines[buff.cursorY][buff.cursorX:]
				buff.cursorX--
			}
		default:
			s := msg.String()
			switch msg.Type {
			case tea.KeyRunes:
				s = string(msg.Runes)
			default:
				return m, nil
			}

			buff.state = bufferStateModified
			if buff.cursorX <= len(buff.lines[buff.cursorY]) {
				buff.lines[buff.cursorY] = buff.lines[buff.cursorY][:buff.cursorX] + s + buff.lines[buff.cursorY][buff.cursorX:]
				buff.cursorX++
			} else {
				buff.lines[buff.cursorY] += s
				buff.cursorX = len(buff.lines[buff.cursorY])
			}
		}
	}

	m.buffers[m.currBuffer] = buff
	return m, nil
}
