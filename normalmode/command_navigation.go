package normalmode

import (
	"slices"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

func (nm *normalmode) commandDown(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	// Wrap around to first line if at last line
	if b.CursorY() >= b.NoOfLines()-1 {
		b = b.SetCursorY(0)
	} else {
		b = b.IncreaseCursorY(1)
	}

	// Handle viewport scrolling
	if b.CursorY() >= b.CursorYOffset()+b.Viewport().Height-2 {
		b = b.IncreaseCursorYOffset(1)
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandUp(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	// Wrap around to last line if at first line
	if b.CursorY() <= 0 {
		b = b.SetCursorY(b.NoOfLines() - 1)
	} else {
		b = b.IncreaseCursorY(-1)
	}

	// Handle viewport scrolling
	if b.CursorY() < b.CursorYOffset() {
		b = b.IncreaseCursorYOffset(-1)
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandLeft(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	if b.CursorX() > 0 {
		b = b.IncreaseCursorX(-1)
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandRight(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	if b.CursorX() < len(b.Line(b.CursorY()))-1 {
		b = b.IncreaseCursorX(1)
	}

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandGoToBeginingOfTheFile(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()
	b = b.SetCursorY(0)

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandGoToEndOfTheFile(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()
	b = b.SetCursorY(len(b.Lines()) - 1)

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandGoToLast(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()
	b = b.SetCursorX(len(b.Line(b.CursorY())))

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandGoToFirstNonWhiteCharacter(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()
	l := b.Line(b.CursorY())
	index := 0
	for i, r := range l {
		if !unicode.IsSpace(r) {
			index = i
			break
		}
	}
	b = b.SetCursorX(index)
	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

// commandNextWord moves the cursor to beginning of the next word
func (nm *normalmode) commandNextWord(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	cursorY := b.CursorY()
	cursorX := b.CursorX()
	line := []rune(b.Line(cursorY))

	if len(line) <= cursorX {
		cursorX = -1
	}

	inWord := !slices.Contains(nextWordSkipCharacters(), line[cursorX])

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

	m = m.ReplaceCurrentBuffer(b)

	return m, cmd
}

func (nm *normalmode) commandPrevWord(m EditorModel, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.CurrentBuffer()

	cursorY := b.CursorY()
	cursorX := b.CursorX()
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

	m = m.ReplaceCurrentBuffer(b)

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
