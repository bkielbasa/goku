package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	// Check for CLI commands first
	if len(os.Args) > 1 {
		// Check if the first argument is a CLI command
		cliCommands := []string{"langs", "languages"}
		for _, cmd := range cliCommands {
			if os.Args[1] == cmd {
				handleCLICommand(os.Args[1:])
				return
			}
		}
	}
	
	var opts []modelOption
	
	// Check if filenames were provided as command line arguments
	if len(os.Args) > 1 {
		filenames := os.Args[1:]
		if len(filenames) == 1 {
			// Single file - use the existing WithFile option for backward compatibility
			opts = append(opts, WithFile(filenames[0]))
		} else {
			// Multiple files - use the new WithFiles option
			opts = append(opts, WithFiles(filenames))
		}
	}
	
	p := tea.NewProgram(initialModel(opts...), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
