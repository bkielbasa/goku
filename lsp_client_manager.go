package main

import (
	"path/filepath"
	"sync"
)

// LSPClientManager manages LSP clients per workspace root
var lspClientManager = &workspaceLSPManager{
	clients: make(map[string]*lspClient),
}

type workspaceLSPManager struct {
	mu      sync.Mutex
	clients map[string]*lspClient
}

// GetLSPClient returns an LSP client for the given root, creating it if necessary
func (m *workspaceLSPManager) GetLSPClient(serverName, rootPath string) (*lspClient, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if client, ok := m.clients[absRoot]; ok {
		return client, nil
	}
	
	client, err := newLSPClient(serverName, absRoot)
	if err != nil {
		return nil, err
	}
	m.clients[absRoot] = client
	return client, nil
}

// CloseAll closes all managed LSP clients
func (m *workspaceLSPManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = make(map[string]*lspClient)
} 