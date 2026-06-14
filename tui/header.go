package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type HeaderModel struct {
	cwd       string
	sessionID string
	modelName string
}

func NewHeaderModel(cwd, sessionID, modelName string) HeaderModel {
	return HeaderModel{
		cwd:       cwd,
		sessionID: sessionID,
		modelName: modelName,
	}
}

func (h HeaderModel) Update(msg tea.Msg) (HeaderModel, tea.Cmd) {
	return h, nil
}

func (h HeaderModel) View() string {
	return fmt.Sprintf(" %s | Session: %s | %s ", h.cwd, h.sessionID, h.modelName)
}
