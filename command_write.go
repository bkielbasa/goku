package main

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type commandWrite struct {
}

func (c commandWrite) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	if len(args) == 0 {
		// Write to current buffer's filename
		buf := m.buffers[m.currBuffer]
		if buf.filename == "" {
			return m.SetErrorMessage("No filename specified"), nil
		}
		
		content := strings.Join(buf.lines, "\n")
		err := os.WriteFile(buf.filename, []byte(content), 0644)
		if err != nil {
			return m.SetErrorMessage("Failed to write file: " + err.Error()), nil
		}
		
		// Mark buffer as saved
		m.buffers[m.currBuffer] = buf.SetStateModified().(buffer)
		return m.SetInfoMessage("File written successfully"), nil
	}

	// Write to specified filename
	filename := args[0]
	buf := m.buffers[m.currBuffer]
	content := strings.Join(buf.lines, "\n")
	
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return m.SetErrorMessage("Failed to write file: " + err.Error()), nil
	}
	
	// Update buffer filename and mark as saved
	buf = buf.SetFileName(filename).(buffer)
	buf = buf.SetStateModified().(buffer)
	m.buffers[m.currBuffer] = buf
	
	m.commandBuffer = ""
	m.mode = ModeNormal
	return m.SetInfoMessage("File written successfully to " + filename), nil
}

func (c commandWrite) Aliases() []string {
	return []string{"write", "w", "save"}
} 