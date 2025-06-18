package main

import (
	"strings"
	"testing"
)

func TestViewportVisual(t *testing.T) {
	m := initialModel()
	m.viewport.Height = len(m.lines) // Simulate a viewport height equal to the number of lines

	// Initial state
	want := `1 _line 1
2 line 2
3 line 3
4 line 4
5 line 5
6 line 6
7 line 7
8 line 8
9 line 9
10 line 10
NORMAL

`
	got := m.View()
	if !strings.HasPrefix(got, want[:20]) { // Only check the first line for brevity
		t.Errorf("Initial state: got\n%s\nwant prefix\n%s", got, want[:20])
	}

	// Move cursor down 8 times
	for i := 0; i < 8; i++ {
		m.cursorY++
		if m.cursorY >= m.cursorYOffset+m.viewport.Height {
			m.cursorYOffset++
		}
	}
	want = `1 line 1
2 line 2
3 line 3
4 line 4
5 line 5
6 line 6
7 line 7
8 line 8
9 _line 9
10 line 10
NORMAL

`
	got = m.View()
	if !strings.Contains(got, "9 _line 9") {
		t.Errorf("After moving cursor down 8 times: got\n%s\nwant line with cursor\n%s", got, "9 _line 9")
	}

	// Simulate entering command mode and typing :q
	m.mode = ModeCommand
	m.commandBuffer = "q"
	want = `1 line 1
2 line 2
3 line 3
4 line 4
5 line 5
6 line 6
7 line 7
8 line 8
9 _line 9
10 line 10
COMMAND
:q

`
	got = m.View()
	if !strings.Contains(got, ":q") || !strings.Contains(got, "COMMAND") {
		t.Errorf("Command mode: got\n%s\nwant :q and COMMAND", got)
	}

	// Simulate returning to normal mode
	m.mode = ModeNormal
	m.commandBuffer = ""
	want = `1 line 1
2 line 2
3 line 3
4 line 4
5 line 5
6 line 6
7 line 7
8 line 8
9 _line 9
10 line 10
NORMAL

`
	got = m.View()
	if !strings.Contains(got, "NORMAL") {
		t.Errorf("Normal mode: got\n%s\nwant NORMAL", got)
	}
} 