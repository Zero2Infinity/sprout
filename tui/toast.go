package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type toastTickMsg time.Time

// ToastModel shows a transient notification that auto-hides after one second.
type ToastModel struct {
	message string
	visible bool
	timer   *time.Timer
}

// NewToastModel creates an initially hidden toast component.
func NewToastModel() ToastModel {
	return ToastModel{}
}

func (t ToastModel) Update(msg tea.Msg) (ToastModel, tea.Cmd) {
	switch msg.(type) {
	case toastTickMsg:
		t.visible = false
		return t, nil
	}
	return t, nil
}

// Show displays a message toast for a short duration (1 second).
func (t *ToastModel) Show(msg string) tea.Cmd {
	t.message = msg
	t.visible = true
	if t.timer != nil {
		t.timer.Stop()
	}
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return toastTickMsg{}
	}
}

func (t ToastModel) View() string {
	if !t.visible {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)

	return style.Render(t.message)
}
