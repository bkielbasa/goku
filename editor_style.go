package main

import (
	"github.com/charmbracelet/lipgloss"
)

type editorStyle struct {
	cursor    lipgloss.Style
	statusBar lipgloss.Style
}

func newEditorStyle() editorStyle {
	return editorStyle{
		cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235")),
		statusBar: lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Background(lipgloss.Color("235")),
	}
}
