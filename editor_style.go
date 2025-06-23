package main

import (
	"github.com/charmbracelet/lipgloss"
)

type editorStyle struct {
	cursor      lipgloss.Style
	statusBar   lipgloss.Style
	keyword     lipgloss.Style
	string      lipgloss.Style
	comment     lipgloss.Style
	number      lipgloss.Style
	function    lipgloss.Style
	typeName    lipgloss.Style
	operator    lipgloss.Style
	punctuation lipgloss.Style
	text        lipgloss.Style // Plain text
}

func newEditorStyle() editorStyle {
	return editorStyle{
		cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("#383838")).Background(lipgloss.Color("#d8d8d8")), // ui.cursor.primary (grey02 on grey05)
		statusBar:   lipgloss.NewStyle().Foreground(lipgloss.Color("#b8b8b8")).Background(lipgloss.Color("#383838")),   // ui.statusline (grey04 on grey02)
		keyword:     lipgloss.NewStyle().Foreground(lipgloss.Color("#cc7832")),                                     // orange
		string:      lipgloss.NewStyle().Foreground(lipgloss.Color("#629755")),                                     // darkgreen
		comment:     lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Italic(true),                         // grey
		number:      lipgloss.NewStyle().Foreground(lipgloss.Color("#6897bb")),                                     // lightblue
		function:    lipgloss.NewStyle().Foreground(lipgloss.Color("#eedd82")),                                     // yellow
		typeName:    lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")),                                     // cyan
		operator:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8d8")),                                     // grey05
		punctuation: lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8d8")),                                     // grey05
		text:        lipgloss.NewStyle().Foreground(lipgloss.Color("#d0d0d0")),                                     // white
	}
}
