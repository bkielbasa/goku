package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LSPClient handles communication with language servers
type lspClient struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	process *exec.Cmd
	ctx     context.Context
	cancel  context.CancelFunc
	nextID  int
	debugFile *os.File
	initialized bool
}

// Location represents a position in a file
type location struct {
	URI   string   `json:"uri"`
	Range lspRange `json:"range"`
}

type lspRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// JSON-RPC message structures
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InitializeParams represents LSP initialization parameters
type initializeParams struct {
	ProcessID int                    `json:"processId"`
	RootURI   string                 `json:"rootUri"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// InitializeResult represents the result of LSP initialization
type initializeResult struct {
	Capabilities map[string]interface{} `json:"capabilities"`
}

// TextDocumentPositionParams represents a position in a text document
type textDocumentPositionParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
}

type textDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentItem represents a text document
type textDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocumentParams represents the parameters for textDocument/didOpen
type didOpenTextDocumentParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

// NewLSPClient creates a new LSP client for the given language server
func newLSPClient(serverName, rootPath string) (*lspClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create debug file (append mode to preserve history)
	debugFile, err := os.OpenFile(fmt.Sprintf("/tmp/goku-lsp-debug-%s.log", serverName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create debug file: %w", err)
	}
	
	// Add a separator to indicate a new LSP client session
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	debugFile.WriteString(fmt.Sprintf("\n[%s] ===== NEW LSP CLIENT SESSION =====\n", timestamp))
	debugFile.WriteString(fmt.Sprintf("[%s] Server: %s, Root: %s\n", timestamp, serverName, rootPath))
	
	// Start the language server process
	cmd := exec.CommandContext(ctx, serverName)
	cmd.Dir = rootPath
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		debugFile.Close()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		debugFile.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		cancel()
		debugFile.Close()
		return nil, fmt.Errorf("failed to start language server: %w", err)
	}
	
	client := &lspClient{
		stdin:   stdin,
		stdout:  stdout,
		process: cmd,
		ctx:     ctx,
		cancel:  cancel,
		nextID:  1,
		debugFile: debugFile,
		initialized: false,
	}
	
	// Initialize the language server
	if err := client.initialize(rootPath); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to initialize language server: %w", err)
	}
	
	return client, nil
}

// IsReady checks if the LSP server is ready by checking if it has sent the initialized notification
func (c *lspClient) IsReady() bool {
	return c.initialized
}

// WaitForReady waits for the LSP server to be ready, with timeout
func (c *lspClient) WaitForReady(timeout time.Duration) bool {
	// If already initialized, return immediately
	if c.initialized {
		return true
	}
	
	// Wait for the initialized notification by polling
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.initialized {
			// Additional wait to ensure packages are fully loaded
			time.Sleep(3 * time.Second)
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// sendRequest sends a JSON-RPC request and waits for response
func (c *lspClient) sendRequest(method string, params interface{}) (json.RawMessage, error) {
	request := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      c.nextID,
		Method:  method,
		Params:  params,
	}
	requestID := c.nextID
	c.nextID++
	
	// Log request
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	c.debugFile.WriteString(fmt.Sprintf("[%s] REQUEST: %s\n", timestamp, string(requestData)))
	
	// Send request
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(requestData))
	_, err = c.stdin.Write([]byte(header))
	if err != nil {
		return nil, err
	}
	
	_, err = c.stdin.Write(requestData)
	if err != nil {
		return nil, err
	}
	
	// Read response - keep reading until we find the response with the correct ID
	timeout := 10 * time.Second
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		// Try to read a message
		message, err := c.readMessage()
		if err != nil {
			// If no message available, wait a bit and try again
			if err.Error() == "no message available" {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return nil, err
		}
		
		// Check if this is a response with the correct ID
		if message.ID == requestID {
			// Log response
			responseData, _ := json.Marshal(message)
			timestamp = time.Now().Format("2006-01-02 15:04:05.000")
			c.debugFile.WriteString(fmt.Sprintf("[%s] RESPONSE: %s\n", timestamp, string(responseData)))
			c.debugFile.Sync()
			
			if message.Error != nil {
				return nil, fmt.Errorf("LSP error: %s", message.Error.Message)
			}
			
			return message.Result, nil
		}
		
		// If it's a notification, log it and continue reading
		if message.Method != "" {
			notificationData, _ := json.Marshal(message)
			timestamp = time.Now().Format("2006-01-02 15:04:05.000")
			c.debugFile.WriteString(fmt.Sprintf("[%s] NOTIFICATION (while waiting for response): %s\n", timestamp, string(notificationData)))
			continue
		}
	}
	
	return nil, fmt.Errorf("timeout waiting for response with ID %d", requestID)
}

// readMessage reads a JSON-RPC message (response or notification) from stdout
func (c *lspClient) readMessage() (*jsonRPCResponse, error) {
	// Set a short timeout for reading
	timeout := 100 * time.Millisecond
	deadline := time.Now().Add(timeout)
	
	// Try to read with timeout
	for time.Now().Before(deadline) {
		// Check if there's data available to read
		if c.stdout == nil {
			return nil, fmt.Errorf("stdout is nil")
		}
		
		// Try to read Content-Length header with a short timeout
		var contentLength int
		_, err := fmt.Fscanf(c.stdout, "Content-Length: %d\r\n", &contentLength)
		if err != nil {
			// If no data available, wait a bit and try again
			time.Sleep(10 * time.Millisecond)
			continue
		}
		
		// Skip the rest of the header
		var line string
		for {
			_, err := fmt.Fscanf(c.stdout, "%s\r\n", &line)
			if err != nil || line == "" {
				break
			}
		}
		
		// Read the JSON payload
		payload := make([]byte, contentLength)
		_, err = io.ReadFull(c.stdout, payload)
		if err != nil {
			return nil, err
		}
		
		// Log raw payload
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		c.debugFile.WriteString(fmt.Sprintf("[%s] RAW PAYLOAD: %s\n", timestamp, string(payload)))
		
		var message jsonRPCResponse
		err = json.Unmarshal(payload, &message)
		if err != nil {
			return nil, err
		}
		
		return &message, nil
	}
	
	// No message available within timeout
	return nil, fmt.Errorf("no message available")
}

// sendNotification sends a JSON-RPC notification (no response expected)
func (c *lspClient) sendNotification(method string, params interface{}) error {
	notification := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	
	notificationData, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	
	// Log notification
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	c.debugFile.WriteString(fmt.Sprintf("[%s] NOTIFICATION: %s\n", timestamp, string(notificationData)))
	
	// Add Content-Length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(notificationData))
	_, err = c.stdin.Write([]byte(header))
	if err != nil {
		return err
	}
	
	_, err = c.stdin.Write(notificationData)
	c.debugFile.Sync()
	return err
}

// initialize sends the initialize request to the language server
func (c *lspClient) initialize(rootPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	rootURI := "file://" + filepath.ToSlash(absPath)
	
	params := initializeParams{
		ProcessID: 0, // We don't send our PID
		RootURI:   rootURI,
		Capabilities: map[string]interface{}{
			"workspace": map[string]interface{}{
				"workspaceFolders": true,
			},
			"textDocument": map[string]interface{}{
				"definition": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"completion": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"hover": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"signatureHelp": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"references": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"documentHighlight": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"documentSymbol": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"codeAction": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"codeLens": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"formatting": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"rangeFormatting": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"onTypeFormatting": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"rename": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"documentLink": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"executeCommand": map[string]interface{}{
					"dynamicRegistration": true,
				},
			},
		},
	}
	
	_, err = c.sendRequest("initialize", params)
	if err != nil {
		return err
	}
	
	// Send initialized notification
	err = c.sendNotification("initialized", nil)
	if err != nil {
		return err
	}
	
	// Mark as initialized - the server is now ready to handle requests
	c.initialized = true
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	c.debugFile.WriteString(fmt.Sprintf("[%s] LSP server initialized\n", timestamp))
	return nil
}

// readResponse reads a JSON-RPC response from stdout (deprecated, use readMessage)
func (c *lspClient) readResponse() (*jsonRPCResponse, error) {
	return c.readMessage()
}

// readNotification reads a JSON-RPC notification from stdout (deprecated, use readMessage)
func (c *lspClient) readNotification() (*jsonRPCRequest, error) {
	message, err := c.readMessage()
	if err != nil {
		return nil, err
	}
	
	// Convert to jsonRPCRequest for backward compatibility
	notification := &jsonRPCRequest{
		JSONRPC: message.JSONRPC,
		Method:  "", // We'll need to extract this from the raw message
	}
	
	// For notifications, we need to parse the raw payload to get the method
	// This is a bit hacky but maintains backward compatibility
	if message.Result != nil {
		// Try to extract method from the result if it's a notification
		var rawMap map[string]interface{}
		if json.Unmarshal(message.Result, &rawMap) == nil {
			if method, ok := rawMap["method"].(string); ok {
				notification.Method = method
			}
		}
	}
	
	return notification, nil
}

// GoToDefinition sends a definition request to the language server
func (c *lspClient) GoToDefinition(filePath string, line, character int) (*location, error) {
	// Get absolute path for the file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	fileURI := "file://" + filepath.ToSlash(absPath)
	
	params := textDocumentPositionParams{
		TextDocument: textDocumentIdentifier{URI: fileURI},
		Position: position{
			Line:      line,
			Character: character,
		},
	}
	
	// Log go-to-definition request specifically
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	c.debugFile.WriteString(fmt.Sprintf("[%s] GoToDefinition: file=%s, line=%d, char=%d\n", timestamp, filePath, line, character))
	
	result, err := c.sendRequest("textDocument/definition", params)
	if err != nil {
		c.debugFile.WriteString(fmt.Sprintf("[%s] GoToDefinition ERROR: %v\n", timestamp, err))
		return nil, err
	}

	// Log the raw result
	c.debugFile.WriteString(fmt.Sprintf("[%s] GoToDefinition raw result: %s\n", timestamp, string(result)))

	// Try to parse as a single location
	var loc location
	err = json.Unmarshal(result, &loc)
	if err == nil && loc.URI != "" {
		c.debugFile.WriteString(fmt.Sprintf("[%s] GoToDefinition found single location: %s\n", timestamp, loc.URI))
		return &loc, nil
	}

	// Try to parse as an array of locations
	var locs []location
	err = json.Unmarshal(result, &locs)
	if err == nil && len(locs) > 0 {
		c.debugFile.WriteString(fmt.Sprintf("[%s] GoToDefinition found %d locations\n", timestamp, len(locs)))
		return &locs[0], nil
	}

	c.debugFile.WriteString(fmt.Sprintf("[%s] GoToDefinition no definition found\n", timestamp))
	return nil, fmt.Errorf("no definition found or could not parse LSP response")
}

// OpenDocument tells the LSP server about a file we're working with
func (c *lspClient) OpenDocument(filePath, content string) error {
	// Get absolute path for the file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	fileURI := "file://" + filepath.ToSlash(absPath)
	
	// Determine language ID from file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	languageID := ""
	switch ext {
	case ".go":
		languageID = "go"
	case ".py":
		languageID = "python"
	case ".js":
		languageID = "javascript"
	case ".ts":
		languageID = "typescript"
	case ".rs":
		languageID = "rust"
	case ".c":
		languageID = "c"
	case ".cpp":
		languageID = "cpp"
	default:
		languageID = "plaintext"
	}
	
	params := didOpenTextDocumentParams{
		TextDocument: textDocumentItem{
			URI:        fileURI,
			LanguageID: languageID,
			Version:    1,
			Text:       content,
		},
	}
	
	return c.sendNotification("textDocument/didOpen", params)
}

// Close closes the LSP client and kills the language server process
func (c *lspClient) Close() error {
	c.cancel()
	
	// Give the process a moment to shut down gracefully
	done := make(chan error, 1)
	go func() {
		done <- c.process.Wait()
	}()
	
	select {
	case <-done:
		// Process exited normally
	case <-time.After(2 * time.Second):
		// Force kill if it doesn't exit within 2 seconds
		c.process.Process.Kill()
		<-done
	}
	
	c.debugFile.Close()
	return c.stdin.Close()
}

// getLSPClientForFile returns an LSP client for the given file, reusing clients per workspace root
func getLSPClientForFile(filePath string, languages map[string]languageSupport) (*lspClient, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	
	// Find the language support for this file extension
	lang, exists := languages[ext]
	if !exists {
		return nil, fmt.Errorf("no language support found for extension: %s", ext)
	}
	
	// Check if LSP server is installed
	if !lang.LSPServer.IsInstalled {
		return nil, fmt.Errorf("LSP server %s is not installed", lang.LSPServer.Name)
	}
	
	// Get the root directory (for now, just use the file's directory)
	rootPath := filepath.Dir(filePath)
	return lspClientManager.GetLSPClient(lang.LSPServer.Name, rootPath)
} 