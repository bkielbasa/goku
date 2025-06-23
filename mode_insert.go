package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateInsert(msg tea.Msg) (tea.Model, tea.Cmd) {
	buff := m.CurrentBuffer()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "alt+esc":
			m.mode = ModeNormal
			m.commandBuffer = ""
		case "enter":
			buff = buff.AppendLine("")
			buff = buff.IncreaseCursorY(1)
			buff = buff.IncreaseCursorX(-buff.CursorX())
		case "left":
			buff = buff.IncreaseCursorX(-1)
		case "right":
			if buff.CursorX() < len(buff.Line(buff.CursorY())) {
				buff = buff.IncreaseCursorX(1)
			}
		case "up":
			buff = buff.IncreaseCursorY(-1)
		case "down":
			if buff.CursorY() < buff.NoOfLines()-1 {
				buff = buff.IncreaseCursorY(1)
			}
		case "backspace":
			if buff.CursorX() > 0 {
				line := buff.Line(buff.CursorY())

				line = line[:buff.CursorX()-1] + line[buff.CursorX():]
				buff = buff.IncreaseCursorX(-1)
				buff = buff.ReplaceLine(buff.CursorY(), line)
			}
		default:
			s := msg.String()
			switch msg.Type {
			case tea.KeyRunes:
				s = string(msg.Runes)
			case tea.KeySpace:
				s = " "
			default:
				return m, nil
			}

			buff = buff.SetStateModified()
			if buff.CursorX() <= len(buff.Line(buff.CursorY())) {
				line := buff.Line(buff.CursorY())
				line = line[:buff.CursorX()] + s + line[buff.CursorX():]
				buff = buff.ReplaceLine(buff.CursorY(), line)
				buff = buff.IncreaseCursorX(1)
			} else {
				line := buff.Line(buff.CursorY())
				line += s

				buff = buff.ReplaceLine(buff.CursorY(), line)
				buff = buff.IncreaseCursorX(len(s))
			}
		}
	}

	m.buffers[m.currBuffer] = buff.(buffer)
	return m, nil
}
