package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type toastTickMsg time.Time

type ToastModel struct {
	message string
	visible bool
	timer   *time.Timer
}

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
	return t.message
}
