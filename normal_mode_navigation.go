package main

import (
	"slices"
	"unicode"
	"unicode/utf16"

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

		// Only check if line is not empty and cursorX is in bounds
		if len(line) > 0 && cursorX < len(line) && cursorX >= 0 {
			// If we're on a word character, check if we've found the start of a word
			if !slices.Contains(nextWordSkipCharacters(), line[cursorX]) {
				// Check if the previous character (if it exists) is a separator
				if cursorX == 0 || slices.Contains(nextWordSkipCharacters(), line[cursorX-1]) {
					break
				}
			}
		}
	}

	b = b.SetCursorX(cursorX)
	b = b.SetCursorY(cursorY)

	m.buffers[m.currBuffer] = b

	return m, cmd
}

// commandGoToDefinition uses LSP to jump to the definition of the symbol under cursor
func (nm *normalmode) commandGoToDefinition(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	filePath := b.FileName()
	if filePath == "" {
		return m, cmd
	}

	// Set loading state and start async go-to-definition
	m.lspLoading = true
	m.lspError = ""
	
	asyncCmd := (&m).AsyncGoToDefinition(filePath, b.cursorY, b.cursorX)
	return m, asyncCmd
}

// commandGoToImplementation uses LSP to jump to the implementation of the symbol under cursor
func (nm *normalmode) commandGoToImplementation(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	filePath := b.FileName()
	if filePath == "" {
		return m, cmd
	}

	// Set loading state and start async go-to-implementation
	m.lspLoading = true
	m.lspError = ""
	
	asyncCmd := (&m).AsyncGoToImplementation(filePath, b.cursorY, b.cursorX)
	return m, asyncCmd
}

// commandGoToTypeDefinition uses LSP to jump to the type definition of the symbol under cursor
func (nm *normalmode) commandGoToTypeDefinition(m model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	b := m.buffers[m.currBuffer]
	filePath := b.FileName()
	if filePath == "" {
		return m, cmd
	}

	// Set loading state and start async go-to-type-definition
	m.lspLoading = true
	m.lspError = ""
	
	asyncCmd := (&m).AsyncGoToTypeDefinition(filePath, b.cursorY, b.cursorX)
	return m, asyncCmd
}

// utf16Index returns the number of UTF-16 code units in s[:cursorX] (where cursorX is a rune index)
func utf16Index(s string, cursorX int) int {
	runes := []rune(s)
	if cursorX > len(runes) {
		cursorX = len(runes)
	}
	return len(utf16.Encode(runes[:cursorX]))
}

// runeIndexFromUTF16 returns the rune index in s corresponding to the given UTF-16 code unit offset
func runeIndexFromUTF16(s string, utf16Offset int) int {
	runes := []rune(s)
	count := 0
	for i := 0; i < len(runes); i++ {
		codeUnits := len(utf16.Encode([]rune{runes[i]}))
		if count+codeUnits > utf16Offset {
			return i
		}
		count += codeUnits
	}
	return len(runes)
}

// extractIdentifierAt extracts the identifier at the given cursor position
func extractIdentifierAt(line string, cursorX int) string {
	runes := []rune(line)
	if cursorX >= len(runes) {
		return ""
	}
	
	// Find the start of the identifier
	start := cursorX
	for start > 0 {
		r := runes[start-1]
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		start--
	}
	
	// Find the end of the identifier
	end := cursorX
	for end < len(runes) {
		r := runes[end]
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		end++
	}
	
	return string(runes[start:end])
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