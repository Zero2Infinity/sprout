package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HeaderModel displays the working directory, session ID, and active model name.
type HeaderModel struct {
	cwd       string
	sessionID string
	modelName string
}

// NewHeaderModel creates a header component with the given context info.
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
		Background(lipgloss.Color("239")).
		Foreground(lipgloss.Color("255"))
	return style.Render(fmt.Sprintf(" Dir: %s | Session: %s | Model: %s ", h.cwd, h.sessionID, h.modelName))
}
