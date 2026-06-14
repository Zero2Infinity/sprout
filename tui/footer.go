package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
	return fmt.Sprintf(" Tokens: %d/%d | %s ", f.currentTokens, f.totalTokens, hints)
}
