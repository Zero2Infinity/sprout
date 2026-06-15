package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HeaderModel struct {
	width     int
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

func (h *HeaderModel) SetWidth(w int) {
	h.width = w
}

func (h HeaderModel) Update(msg tea.Msg) (HeaderModel, tea.Cmd) {
	return h, nil
}

func (h HeaderModel) View() string {
	padding := 4
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("239")).
		Foreground(lipgloss.Color("255")).
		Padding(0, 2)
	text := fmt.Sprintf(" %s │ Session: %s │ %s ", h.cwd, h.sessionID, h.modelName)
	if h.width > 0 {
		pad := h.width - lipgloss.Width(text) - padding
		if pad > 0 {
			text += strings.Repeat(" ", pad)
		}
	}
	return barStyle.Render(text)
}
