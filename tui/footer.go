package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FooterModel displays token counts, timing, and keyboard shortcut hints.
type FooterModel struct {
	promptTokens     int
	completionTokens int
	contextUsed      int
	contextMax       int
	tokensPerSec     float64
	totalDurationNs  int64
}

// NewFooterModel creates a footer with default key hints.
func NewFooterModel() FooterModel {
	return FooterModel{}
}

func (f FooterModel) Update(msg tea.Msg) (FooterModel, tea.Cmd) {
	return f, nil
}

// WithUsage returns a copy of the footer with updated prompt, completion, and timing.
func (f FooterModel) WithUsage(prompt, completion int, tokensPerSec float64, totalDurationNs int64) FooterModel {
	f.promptTokens = prompt
	f.completionTokens = completion
	f.contextUsed = prompt + completion
	f.tokensPerSec = tokensPerSec
	f.totalDurationNs = totalDurationNs
	return f
}

// WithContext returns a copy of the footer with updated context window info.
func (f FooterModel) WithContext(used, max int) FooterModel {
	f.contextUsed = used
	f.contextMax = max
	return f
}

func (f FooterModel) View() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("239")).
		Foreground(lipgloss.Color("255"))

	metrics := fmt.Sprintf("Prompt: %d | Completion: %d | Context: %d/%d",
		f.promptTokens, f.completionTokens, f.contextUsed, f.contextMax)

	if f.tokensPerSec > 0 {
		durationSec := float64(f.totalDurationNs) / 1e9
		metrics += fmt.Sprintf(" | Speed: %.1f tok/s | Latency: %.1fs", f.tokensPerSec, durationSec)
	}

	return style.Render(" " + metrics + " ")
}
