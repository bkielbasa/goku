package main

import (
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type commandOpen struct {
}

func (c commandOpen) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	if len(args) == 0 {
		return m, nil
	}

	for _, filepath := range args {
		f, err := os.OpenFile(filepath, os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}

		cont, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}

		m = m.addBuffer(newBuffer(m.style, bufferStateSavedOpt, bufferWithContent(filepath, string(cont))))
		m.commandBuffer = ""
		m.mode = ModeNormal
		m.currBuffer++
	}

	return m, nil
}

func (c commandOpen) Aliases() []string {
	return []string{"open", "o", "e", "edit"}
}
