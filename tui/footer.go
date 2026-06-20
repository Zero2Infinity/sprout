package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FooterModel displays token counts and keyboard shortcut hints.
type FooterModel struct {
	currentTokens int
	totalTokens   int
	keyHints      []string
}

// NewFooterModel creates a footer with default key hints.
func NewFooterModel() FooterModel {
	return FooterModel{
		keyHints: []string{"enter: send", "esc: cancel", "↑/↓: history"},
	}
}

func (f FooterModel) Update(msg tea.Msg) (FooterModel, tea.Cmd) {
	return f, nil
}

func (f FooterModel) View() string {
	hints := ""
	for i, h := range f.keyHints {
		if i > 0 {
			hints += " | "
		}
		hints += h
	}
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("239")).
		Foreground(lipgloss.Color("255"))
	return style.Render(fmt.Sprintf(" Tokens: %d/%d | %s ", f.currentTokens, f.totalTokens, hints))
}
