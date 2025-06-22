package main

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type commandWrite struct{}

func (c commandWrite) Update(m model, msg tea.Msg, args []string) (model, tea.Cmd) {
	b := m.CurrentBuffer()
	if b.FileName() != "" {
		fileName := b.FileName()
		if len(args) > 0 {
			fileName = args[0]
		}

		f, err := os.OpenFile(fileName, os.O_WRONLY, 0666)
		if err != nil {
			// add error handling
			panic(err)
			return m, nil
		}

		cont := strings.Join(b.Lines(), "\n")
		_, err = f.Write([]byte(cont))
		if err != nil {
			// add error handling
			panic(err)
			return m, nil
		}
		b = b.SetFileName(fileName)
		m.commandBuffer = ""
		m.mode = ModeNormal

		return m, nil
	}

	if len(args) == 0 {
		// add error handing
		return m, nil
	}

	fileName := args[0]
	b = b.SetFileName(fileName)
	cont := strings.Join(b.Lines(), "\n")
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		// add error handling
		panic(err)
		return m, nil
	}
	_, err = f.Write([]byte(cont))
	if err != nil {
		// add error handling
		panic(err)
		return m, nil
	}

	m.commandBuffer = ""
	m.mode = ModeNormal

	return m, nil
}

func (c commandWrite) Aliases() []string {
	return []string{"write", "w"}
}
