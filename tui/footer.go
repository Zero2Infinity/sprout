package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FooterModel displays token counts and keyboard shortcut hints.
type FooterModel struct {
	promptTokens     int
	completionTokens int
	contextUsed      int
	contextMax       int
	keyHints         []string
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

// WithUsage returns a copy of the footer with updated prompt and completion token counts.
func (f FooterModel) WithUsage(prompt, completion int) FooterModel {
	f.promptTokens = prompt
	f.completionTokens = completion
	f.contextUsed = prompt + completion
	return f
}

// WithContext returns a copy of the footer with updated context window info.
func (f FooterModel) WithContext(used, max int) FooterModel {
	f.contextUsed = used
	f.contextMax = max
	return f
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
	return style.Render(fmt.Sprintf(" Prompt: %d | Completion: %d | Context: %d/%d | %s ", f.promptTokens, f.completionTokens, f.contextUsed, f.contextMax, hints))
}
