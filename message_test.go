package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMessageSystem(t *testing.T) {
	tests := []struct {
		name           string
		messageType    messageType
		messageText    string
		expectedInView string
	}{
		{
			name:           "info message",
			messageType:    MessageInfo,
			messageText:    "This is an info message",
			expectedInView: "This is an info message",
		},
		{
			name:           "error message",
			messageType:    MessageError,
			messageText:    "This is an error message",
			expectedInView: "This is an error message",
		},
		{
			name:           "long message truncation",
			messageType:    MessageInfo,
			messageText:    strings.Repeat("a", 100),
			expectedInView: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}

			// Set the message
			switch tt.messageType {
			case MessageInfo:
				m = m.SetInfoMessage(tt.messageText)
			case MessageError:
				m = m.SetErrorMessage(tt.messageText)
			}

			// Get the view
			view := m.View()

			// Check if the message appears in the view
			if !strings.Contains(view, tt.expectedInView) {
				t.Errorf("message '%s' not found in view", tt.expectedInView)
			}

			// Clear the message
			m = m.ClearMessage()
			viewAfterClear := m.View()

			// Check that the message is no longer in the view
			if strings.Contains(viewAfterClear, tt.messageText) {
				t.Errorf("message still appears in view after clearing")
			}
		})
	}
}

func TestMessageStyling(t *testing.T) {
	m := initialModel()
	m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}

	// Test info message styling
	m = m.SetInfoMessage("Info message")
	view := m.View()
	
	// Check that the message appears with proper styling
	if !strings.Contains(view, "Info message") {
		t.Error("info message not found in view")
	}

	// Test error message styling
	m = m.SetErrorMessage("Error message")
	view = m.View()
	
	// Check that the message appears with proper styling
	if !strings.Contains(view, "Error message") {
		t.Error("error message not found in view")
	}
}

func TestQuitWithUnsavedBuffers(t *testing.T) {
	m := initialModel()
	m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}

	// Create a buffer with unsaved changes
	m.buffers[0].state = bufferStateModified
	m.buffers[0].filename = "test.txt"

	// Try to quit
	quitCmd := commandQuit{}
	newModel, cmd := quitCmd.Update(m, nil, []string{})

	// Should not quit and should show error message
	if cmd != nil {
		t.Error("Expected no quit command when there are unsaved buffers")
	}

	// Check that error message is set
	if newModel.currentMessage == nil {
		t.Error("Expected error message when quitting with unsaved buffers")
	}

	if newModel.currentMessage.msgType != MessageError {
		t.Error("Expected error message type")
	}

	if !strings.Contains(newModel.currentMessage.text, "No write since last change") {
		t.Error("Expected error message about unsaved changes")
	}
}

func TestQuitWithSavedBuffers(t *testing.T) {
	m := initialModel()
	m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}

	// Create a buffer with saved changes
	m.buffers[0].state = bufferStateSaved
	m.buffers[0].filename = "test.txt"

	// Try to quit
	quitCmd := commandQuit{}
	newModel, cmd := quitCmd.Update(m, nil, []string{})

	// Should quit without error message
	if cmd == nil {
		t.Error("Expected quit command when all buffers are saved")
	}

	// Check that no error message is set
	if newModel.currentMessage != nil {
		t.Error("Expected no error message when quitting with saved buffers")
	}
}

func TestForceQuit(t *testing.T) {
	m := initialModel()
	m.viewport = tea.WindowSizeMsg{Width: 80, Height: 24}

	// Create a buffer with unsaved changes
	m.buffers[0].state = bufferStateModified
	m.buffers[0].filename = "test.txt"

	// Try to force quit
	quitCmd := commandForceQuit{}
	newModel, cmd := quitCmd.Update(m, nil, []string{})

	// Should quit even with unsaved buffers
	if cmd == nil {
		t.Error("Expected quit command for force quit")
	}

	// Check that no error message is set
	if newModel.currentMessage != nil {
		t.Error("Expected no error message for force quit")
	}
} 