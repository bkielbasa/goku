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
			// Split the current line at cursor position
			currentLine := buff.Line(buff.CursorY())
			cursorX := buff.CursorX()
			
			// Ensure cursor position is within bounds
			if cursorX > len(currentLine) {
				cursorX = len(currentLine)
			}
			
			beforeCursor := currentLine[:cursorX]
			afterCursor := currentLine[cursorX:]
			
			// Replace current line with content before cursor
			buff = buff.ReplaceLine(buff.CursorY(), beforeCursor)
			
			// Insert new line with content after cursor
			buff = buff.InsertLine(buff.CursorY()+1, afterCursor)
			
			// Move cursor to beginning of new line
			buff = buff.IncreaseCursorY(1)
			buff = buff.SetCursorX(0)
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
				cursorX := buff.CursorX()
				
				// Ensure cursor position is within bounds
				if cursorX > len(line) {
					cursorX = len(line)
				}
				
				if cursorX > 0 {
					line = line[:cursorX-1] + line[cursorX:]
					buff = buff.IncreaseCursorX(-1)
					buff = buff.ReplaceLine(buff.CursorY(), line)
				}
			} else if buff.CursorY() > 0 {
				// At beginning of line, move to end of previous line
				currentLine := buff.Line(buff.CursorY())
				previousLine := buff.Line(buff.CursorY() - 1)
				
				// Combine previous line with current line
				combinedLine := previousLine + currentLine
				buff = buff.ReplaceLine(buff.CursorY()-1, combinedLine)
				
				// Move cursor to the end of the previous line
				buff = buff.IncreaseCursorY(-1)
				buff = buff.SetCursorX(len(previousLine))
				
				// Delete the current line (now that content has been moved)
				buff = buff.DeleteLine(buff.CursorY() + 1)
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
			cursorX := buff.CursorX()
			line := buff.Line(buff.CursorY())
			
			// Ensure cursor position is within bounds
			if cursorX > len(line) {
				cursorX = len(line)
			}
			
			if cursorX <= len(line) {
				line = line[:cursorX] + s + line[cursorX:]
				buff = buff.ReplaceLine(buff.CursorY(), line)
				buff = buff.IncreaseCursorX(1)
			} else {
				line += s
				buff = buff.ReplaceLine(buff.CursorY(), line)
				buff = buff.IncreaseCursorX(len(s))
			}
		}
	}

	m.buffers[m.currBuffer] = buff
	return m, nil
}
