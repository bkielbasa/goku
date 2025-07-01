package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestLSPIntegration tests the complete LSP go-to-definition functionality
func TestLSPIntegration(t *testing.T) {
	defer lspClientManager.CloseAll()
	// Skip if gopls is not installed
	if !isToolInstalled("gopls") {
		t.Skip("gopls not installed, skipping LSP integration test")
	}

	// Create a temporary test directory
	testDir, err := os.MkdirTemp("", "goku-lsp-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a simple Go module
	err = createTestGoModule(testDir)
	if err != nil {
		t.Fatalf("Failed to create test Go module: %v", err)
	}

	// Create test files
	err = createTestFiles(testDir)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	// Add a helper to open all files in the workspace with the LSP client
	m := initialModel()
	err = openAllFilesWithLSPClient(testDir, &m)
	if err != nil {
		t.Fatalf("Failed to open all files with LSP client: %v", err)
	}

	// Test LSP go-to-definition
	t.Run("GoToDefinition", func(t *testing.T) {
		testGoToDefinition(t, testDir, m)
	})
}

// createTestGoModule creates a simple Go module for testing
func createTestGoModule(testDir string) error {
	goModContent := `module github.com/test/lsp-test

go 1.21
`
	return os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goModContent), 0644)
}

// createTestFiles creates test Go files with known definitions
func createTestFiles(testDir string) error {
	// main.go - contains a function that calls another function
	mainGoContent := `package main

import "fmt"

// TestFunction is a test function that calls HelperFunction
func TestFunction() {
	fmt.Println("Calling helper function")
	HelperFunction()
}

func main() {
	TestFunction()
}
`
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(mainGoContent), 0644)
	if err != nil {
		return err
	}

	// helper.go - contains the helper function
	helperGoContent := `package main

// HelperFunction is a helper function that does something useful
func HelperFunction() {
	println("Hello from helper function")
}

// AnotherFunction is another test function
func AnotherFunction() {
	HelperFunction()
}
`
	return os.WriteFile(filepath.Join(testDir, "helper.go"), []byte(helperGoContent), 0644)
}

// testGoToDefinition tests the go-to-definition functionality
func testGoToDefinition(t *testing.T, testDir string, m model) {
	// Initialize the model with test files
	mainGoPath := filepath.Join(testDir, "main.go")

	// Load the main.go file into the existing model
	mainBuf, err := loadFile(mainGoPath, m.style)
	if err != nil {
		t.Fatalf("Failed to load main.go: %v", err)
	}
	m.buffers[0] = mainBuf
	
	// Update language support to check for gopls
	m.updateLanguageSupport()

	// Verify gopls is detected as installed
	lang, exists := m.Languages["go"]
	if !exists {
		t.Fatal("Go language support not found")
	}
	if !lang.LSPServer.IsInstalled {
		t.Skip("gopls not installed")
	}

	// Test 1: Go to definition of HelperFunction from main.go
	t.Run("HelperFunctionDefinition", func(t *testing.T) {
		// Position cursor on "HelperFunction" in main.go (line 8, around character 2)
		// The line is: HelperFunction()
		m.buffers[0] = m.buffers[0].SetCursorY(7) // Line 8 (0-indexed)
		m.buffers[0] = m.buffers[0].SetCursorX(1)  // Position on 'H' of HelperFunction

		t.Logf("Cursor position: line %d, char %d", m.buffers[0].CursorY(), m.buffers[0].CursorX())
		t.Logf("Current line: %q", m.buffers[0].Line(m.buffers[0].CursorY()))

		// Create normal mode and execute gd command
		nm := NewNormalMode()
		
		// Simulate pressing 'g' then 'd'
		nm, modelResult, cmd := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}, m)
		m = modelResult.(model)
		nm, modelResult, cmd = nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
		m = modelResult.(model)

		t.Logf("Command returned: %v", cmd != nil)

		// Execute the async command and handle the async flow
		if cmd != nil {
			msg := cmd()
			t.Logf("Async command result: %T", msg)
			
			// Handle the async flow by simulating the Update cycle
			for i := 0; i < 10; i++ { // Limit iterations to prevent infinite loop
				updateResult, updateCmd := m.Update(msg)
				m = updateResult.(model)
				
				if updateCmd != nil {
					msg = updateCmd()
					t.Logf("Update command result: %T", msg)
					
					// Check if we got the final result
					if result, ok := msg.(asyncGoToDefinitionResult); ok {
						t.Logf("Go-to-definition result: error=%v, location=%v", result.error, result.location)
						if result.error != nil {
							t.Fatalf("Go-to-definition failed: %v", result.error)
						}
						if result.location != nil {
							t.Logf("Found location: %s at line %d", result.location.URI, result.location.Range.Start.Line)
							// Handle the result
							m = m.handleGoToDefinitionResult(result.location)
						}
						break
					}
				} else {
					// No more commands to execute
					break
				}
			}
		}

		// Verify we switched to helper.go
		if len(m.buffers) < 2 {
			t.Fatal("Expected helper.go to be opened")
		}

		// Check if we're now in helper.go
		currentFile := m.buffers[m.currBuffer].FileName()
		if !strings.HasSuffix(currentFile, "helper.go") {
			t.Errorf("Expected to be in helper.go, but in %s", currentFile)
		}

		// Verify cursor is positioned on the HelperFunction definition
		// HelperFunction should be on line 4 (0-indexed)
		cursorY := m.buffers[m.currBuffer].CursorY()
		if cursorY != 3 { // Line 4 (0-indexed)
			t.Errorf("Expected cursor on line 4, but on line %d", cursorY+1)
		}
	})

	// Test 2: Go to definition of fmt.Println from main.go
	t.Run("StdLibDefinition", func(t *testing.T) {
		// Reset to main.go
		m.currBuffer = 0
		
		// Position cursor on "fmt.Println" in main.go (line 6, around character 2)
		// The line is: fmt.Println("Calling helper function")
		m.buffers[0] = m.buffers[0].SetCursorY(5) // Line 6 (0-indexed)
		m.buffers[0] = m.buffers[0].SetCursorX(4)  // Position on 'P' of Println

		// Execute gd command
		nm := NewNormalMode()
		nm, modelResult, cmd := nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}, m)
		m = modelResult.(model)
		nm, modelResult, cmd = nm.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, m)
		m = modelResult.(model)

		// Handle async flow
		if cmd != nil {
			msg := cmd()
			for i := 0; i < 10; i++ {
				updateResult, updateCmd := m.Update(msg)
				m = updateResult.(model)
				if updateCmd != nil {
					msg = updateCmd()
					if _, ok := msg.(asyncGoToDefinitionResult); ok {
						break
					}
				} else {
					break
				}
			}
		}

		// For stdlib functions, we might not be able to jump to the source
		// but we should at least not crash and should handle the response gracefully
		if m.currBuffer == 0 {
			// If we're still in main.go, that's fine for stdlib functions
			t.Log("Stayed in main.go for stdlib function (expected behavior)")
		}
	})
}

// TestLSPClientCreation tests LSP client creation and initialization
func TestLSPClientCreation(t *testing.T) {
	// Skip if gopls is not installed
	if !isToolInstalled("gopls") {
		t.Skip("gopls not installed, skipping LSP client test")
	}

	// Create a temporary directory
	testDir, err := os.MkdirTemp("", "goku-lsp-client-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a simple Go file
	testFile := filepath.Join(testDir, "test.go")
	testContent := `package main

func main() {
	println("Hello, World!")
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test LSP client creation
	client, err := getLSPClientForFile(testFile, map[string]languageSupport{
		"go": {
			Name: "Go",
			LSPServer: toolInfo{Name: "gopls", IsInstalled: true},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create LSP client: %v", err)
	}
	defer client.Close()

	// Test opening a document
	err = client.OpenDocument(testFile, testContent)
	if err != nil {
		t.Fatalf("Failed to open document: %v", err)
	}

	// Give the LSP server time to process
	time.Sleep(200 * time.Millisecond)

	// Test go-to-definition (should find main function)
	location, err := client.GoToDefinition(testFile, 2, 5) // Position on "main"
	if err != nil {
		t.Logf("GoToDefinition failed (this might be expected): %v", err)
	} else if location != nil {
		t.Logf("Found definition at: %s", location.URI)
	}
}

// TestLSPStatusBar tests that LSP status is correctly displayed in status bar
func TestLSPStatusBar(t *testing.T) {
	m := initialModel()
	m.updateLanguageSupport()

	// Test with a Go file
	m.buffers[0] = m.buffers[0].SetFileName("test.go")
	
	// Get the status bar content
	view := m.View()
	
	// Check if gopls is mentioned in the status bar
	if !strings.Contains(view, "gopls") {
		t.Error("gopls not found in status bar")
	}
	
	// Check if there's a checkmark or X for LSP status
	if !strings.Contains(view, "✅") && !strings.Contains(view, "❌") {
		t.Error("LSP status indicator (✅ or ❌) not found in status bar")
	}
}

// TestUTF16Conversion tests the UTF-16 conversion functions
func TestUTF16Conversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		runePos  int
		expected int
	}{
		{"ASCII only", "hello", 3, 3},
		{"ASCII only at end", "hello", 5, 5},
		{"Unicode character", "héllo", 2, 2}, // 'é' is 1 UTF-16 code unit
		{"Unicode at end", "héllo", 5, 5},
		{"Multiple Unicode", "héllö", 5, 5}, // 'ö' is 1 UTF-16 code unit
		{"Emoji", "he😊llo", 3, 4}, // '😊' is 2 UTF-16 code units
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utf16Index(tt.input, tt.runePos)
			if result != tt.expected {
				t.Errorf("utf16Index(%q, %d) = %d, want %d", tt.input, tt.runePos, result, tt.expected)
			}
		})
	}
}

// TestRuneIndexFromUTF16 tests the reverse UTF-16 conversion
func TestRuneIndexFromUTF16(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		utf16Pos int
		expected int
	}{
		{"ASCII only", "hello", 3, 3},
		{"ASCII only at end", "hello", 5, 5},
		{"Unicode character", "héllo", 2, 2}, // UTF-16 pos 2 = rune pos 2
		{"Unicode at end", "héllo", 5, 5},
		{"Multiple Unicode", "héllö", 5, 5}, // UTF-16 pos 5 = rune pos 5
		{"Emoji", "he😊llo", 4, 3}, // UTF-16 pos 4 = rune pos 3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runeIndexFromUTF16(tt.input, tt.utf16Pos)
			if result != tt.expected {
				t.Errorf("runeIndexFromUTF16(%q, %d) = %d, want %d", tt.input, tt.utf16Pos, result, tt.expected)
			}
		})
	}
}

// openAllFilesWithLSPClient opens all .go files in the workspace with the same LSP client
func openAllFilesWithLSPClient(testDir string, m *model) error {
	files, err := os.ReadDir(testDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			path := filepath.Join(testDir, file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			client, err := getLSPClientForFile(path, m.Languages)
			if err != nil {
				return err
			}
			err = client.OpenDocument(path, string(content))
			if err != nil {
				return err
			}
		}
	}
	return nil
} 