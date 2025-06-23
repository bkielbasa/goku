package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type commandQuit struct {
}

func (c commandQuit) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	// Check for unsaved buffers
	unsavedBuffers := []string{}

	for i, buf := range m.buffers {
		if buf.state == bufferStateModified {
			filename := buf.filename
			if filename == "" {
				filename = fmt.Sprintf("Buffer %d", i+1)
			}
			unsavedBuffers = append(unsavedBuffers, filename)
		}
	}

	// If there are unsaved buffers, show error and don't quit
	if len(unsavedBuffers) > 0 {
		message := fmt.Sprintf("No write since last change for %d buffer(s): %s",
			len(unsavedBuffers),
			strings.Join(unsavedBuffers, ", "))
		return m.SetErrorMessage(message), nil
	}

	// All buffers are saved, safe to quit
	return m, tea.Quit
}

func (c commandQuit) Aliases() []string {
	return []string{"quit", "q"}
}

type commandForceQuit struct {
}

func (c commandForceQuit) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	return m, tea.Quit
}

func (c commandForceQuit) Aliases() []string {
	return []string{"quit!", "q!"}
}
