package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type commandBufferNext struct {
}

func (c commandBufferNext) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	if len(m.buffers) <= 1 {
		m.commandBuffer = ""
		m.mode = ModeNormal
		return m.SetInfoMessage("No other buffers"), nil
	}
	
	m.currBuffer = (m.currBuffer + 1) % len(m.buffers)
	
	// Update viewport for the new buffer
	m.buffers[m.currBuffer].viewport = m.viewport
	
	// Clear command buffer and switch to normal mode
	m.commandBuffer = ""
	m.mode = ModeNormal
	
	return m, nil
}

func (c commandBufferNext) Aliases() []string {
	return []string{"bnext", "bn"}
}

type commandBufferPrev struct {
}

func (c commandBufferPrev) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	if len(m.buffers) <= 1 {
		m.commandBuffer = ""
		m.mode = ModeNormal
		return m.SetInfoMessage("No other buffers"), nil
	}
	
	m.currBuffer = (m.currBuffer - 1 + len(m.buffers)) % len(m.buffers)
	
	// Update viewport for the new buffer
	m.buffers[m.currBuffer].viewport = m.viewport
	
	// Clear command buffer and switch to normal mode
	m.commandBuffer = ""
	m.mode = ModeNormal
	
	return m, nil
}

func (c commandBufferPrev) Aliases() []string {
	return []string{"bprev", "bp"}
}

type commandBufferLast struct {
}

func (c commandBufferLast) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	if len(m.buffers) <= 1 {
		m.commandBuffer = ""
		m.mode = ModeNormal
		return m.SetInfoMessage("No other buffers"), nil
	}
	
	m.currBuffer = len(m.buffers) - 1
	
	// Update viewport for the new buffer
	m.buffers[m.currBuffer].viewport = m.viewport
	
	// Clear command buffer and switch to normal mode
	m.commandBuffer = ""
	m.mode = ModeNormal
	
	return m, nil
}

func (c commandBufferLast) Aliases() []string {
	return []string{"blast", "bl"}
}

type commandBufferFirst struct {
}

func (c commandBufferFirst) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	if len(m.buffers) <= 1 {
		m.commandBuffer = ""
		m.mode = ModeNormal
		return m.SetInfoMessage("No other buffers"), nil
	}
	
	m.currBuffer = 0
	
	// Update viewport for the new buffer
	m.buffers[m.currBuffer].viewport = m.viewport
	
	// Clear command buffer and switch to normal mode
	m.commandBuffer = ""
	m.mode = ModeNormal
	
	return m, nil
}

func (c commandBufferFirst) Aliases() []string {
	return []string{"bfirst", "bf"}
} 