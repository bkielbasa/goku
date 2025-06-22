package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	var opts []modelOption
	
	// Check if a filename was provided as a command line argument
	if len(os.Args) > 1 {
		filename := os.Args[1]
		opts = append(opts, WithFile(filename))
	}
	
	p := tea.NewProgram(initialModel(opts...), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
