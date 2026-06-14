package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FooterModel struct {
	currentTokens int
	totalTokens   int
	keyHints      []string
}

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
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)

	return style.Render(fmt.Sprintf("Tokens: %d/%d | %s", f.currentTokens, f.totalTokens, hints))
}
