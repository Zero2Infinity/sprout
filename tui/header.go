package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Padding(0, 1)

	return style.Render(fmt.Sprintf("%s | Session: %s | %s", h.cwd, h.sessionID, h.modelName))
}
