package main

import (
	"slices"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandDown(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	// Wrap around to first line if at last line
	if b.cursorY >= b.NoOfLines()-1 {
		b = b.SetCursorY(0)
	} else {
		b = b.IncreaseCursorY(1)
	}

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandUp(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	// Wrap around to last line if at first line
	if b.cursorY <= 0 {
		b = b.SetCursorY(b.NoOfLines() - 1)
	} else {
		b = b.IncreaseCursorY(-1)
	}

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandLeft(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	if b.cursorX > 0 {
		b = b.IncreaseCursorX(-1)
	}

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandRight(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	if b.cursorX < len(b.Line(b.cursorY))-1 {
		b = b.IncreaseCursorX(1)
	}

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandGoToBeginingOfTheFile(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	b = b.SetCursorY(0)

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandGoToEndOfTheFile(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	b = b.SetCursorY(len(b.Lines()) - 1)

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandGoToLast(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	b = b.SetCursorX(len(b.Line(b.cursorY)))

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm normalmode) commandGoToFirstNonWhiteCharacter(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	l := b.Line(b.cursorY)
	index := 0
	for i, r := range l {
		if !unicode.IsSpace(r) {
			index = i
			break
		}
	}
	b = b.SetCursorX(index)
	m.buffers[m.currBuffer] = b

	return m, cmd
}

// commandNextWord moves the cursor to beginning of the next word
func (nm *normalmode) commandNextWord(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	cursorY := b.cursorY
	cursorX := b.cursorX
	line := []rune(b.Line(cursorY))

	// If cursor is at or beyond the end of the line, move to next line
	if len(line) <= cursorX {
		cursorX = -1
		cursorY++
		if b.NoOfLines() == cursorY {
			cursorY = 0
		}
		line = []rune(b.Line(cursorY))
	}

	// Determine if we're currently in a word
	inWord := false
	if cursorX >= 0 && cursorX < len(line) {
		inWord = !slices.Contains(nextWordSkipCharacters(), line[cursorX])
	}

	for {
		cursorX++

		// we hit end of the line
		if cursorX >= len(line) {
			cursorX = -1
			cursorY++

			if b.NoOfLines() == cursorY {
				cursorY = 0
			}

			line = []rune(b.Line(cursorY))

			// Skip empty lines
			if len(line) == 0 {
				continue
			}

			continue
		}

		// TODO: detect skipping in a infinity loop

		if inWord {
			// we're waiting until we finish the word
			if !slices.Contains(nextWordSkipCharacters(), line[cursorX]) {
				continue
			}

			inWord = false
			continue
		}

		if slices.Contains(nextWordSkipCharacters(), line[cursorX]) {
			continue
		}

		b = b.SetCursorX(cursorX)
		b = b.SetCursorY(cursorY)
		break
	}

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func (nm *normalmode) commandPrevWord(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]

	cursorY := b.cursorY
	cursorX := b.cursorX
	line := []rune(b.Line(cursorY))

	// If we're at the beginning of the file, wrap to the end
	if cursorY == 0 && cursorX == 0 {
		cursorY = b.NoOfLines() - 1
		line = []rune(b.Line(cursorY))
		cursorX = len(line)
	}

	// Move back one character to start the search
	if cursorX > 0 {
		cursorX--
	} else {
		// Move to previous line
		cursorY--
		if cursorY < 0 {
			cursorY = b.NoOfLines() - 1
		}
		line = []rune(b.Line(cursorY))
		cursorX = len(line) - 1
		if cursorX < 0 {
			cursorX = 0
		}
	}

	// Find the start of the previous word
	for {
		// If we're at the beginning of the file, we're done
		if cursorY == 0 && cursorX == 0 {
			break
		}

		// If we're at the beginning of a line, go to the previous line
		if cursorX == 0 {
			cursorY--
			if cursorY < 0 {
				cursorY = b.NoOfLines() - 1
			}
			line = []rune(b.Line(cursorY))
			cursorX = len(line) - 1
			if cursorX < 0 {
				cursorX = 0
			}
			continue
		}

		// Move back one character
		cursorX--

		// If we're on a word character, check if we've found the start of a word
		if !slices.Contains(nextWordSkipCharacters(), line[cursorX]) {
			// Check if the previous character (if it exists) is a separator
			if cursorX == 0 || slices.Contains(nextWordSkipCharacters(), line[cursorX-1]) {
				break
			}
		}
	}

	b = b.SetCursorX(cursorX)
	b = b.SetCursorY(cursorY)

	m.buffers[m.currBuffer] = b

	return m, cmd
}

func nextWordSkipCharacters() []rune {
	return []rune{
		' ',
		'\t',
		'(',
		')',
		'[',
		'.',
		',',
		']',
		'\\',
		'/',
		'?',
		'"',
		'\'',
		'`',
		'|',
		'{',
		'}',
		'<',
		'>',
		'=',
	}
} 