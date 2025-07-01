package main

import (
	"fmt"
)

// CLI command handlers
func handleCLICommand(args []string) {
	if len(args) < 1 {
		return
	}

	langFilter := ""
	for i, arg := range args {
		if arg == "--lang" && i+1 < len(args) {
			langFilter = args[i+1]
		}
	}

	switch args[0] {
	case "langs", "languages":
		handleLangsCommand(langFilter)
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		fmt.Println("Available commands:")
		fmt.Println("  langs, languages [--lang <ext>] - Show language support information")
	}
}

func handleLangsCommand(langFilter string) {
	m := initialModel()
	m.updateLanguageSupport()

	for ext, support := range m.Languages {
		if langFilter != "" && ext != langFilter {
			continue
		}
		fmt.Printf("%s (%s):\n", support.Name, ext)
		fmt.Printf("  LSP: %s %s\n", checkmark(support.LSPServer.IsInstalled), toolName(support.LSPServer.Name))
		fmt.Printf("  Formatter: %s %s\n", checkmark(support.Formatter.IsInstalled), toolName(support.Formatter.Name))
		fmt.Printf("  Highlighting: %s %s\n", checkmark(support.Highlighting.IsInstalled), toolName(support.Highlighting.Name))
		fmt.Println()
	}
}

func checkmark(installed bool) string {
	if installed {
		return "✅"
	}
	return "❌"
}

func toolName(name string) string {
	if name == "" {
		return "-"
	}
	return name
} 