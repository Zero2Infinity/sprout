package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputSubmitMsg signals that the user submitted a message via the input.
type InputSubmitMsg struct {
	Value string
}

// InputModel provides a textarea-based input with history navigation.
type InputModel struct {
	textarea   textarea.Model
	history    []string
	historyIdx int
	width      int
}

// NewInputModel creates an input component with history tracking.
func NewInputModel() InputModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(2)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)

	bgStyle := func() lipgloss.Style {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("0")).
			Foreground(lipgloss.Color("240"))
	}

	ta.FocusedStyle.Base = bgStyle()
	ta.FocusedStyle.Placeholder = bgStyle()
	ta.FocusedStyle.CursorLine = bgStyle()
	ta.BlurredStyle.Base = bgStyle()
	ta.BlurredStyle.Placeholder = bgStyle()
	ta.BlurredStyle.CursorLine = bgStyle()

	return InputModel{
		textarea:   ta,
		history:    make([]string, 0),
		historyIdx: -1,
	}
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			val := m.textarea.Value()
			if val == "" {
				return m, nil
			}
			m.history = append(m.history, val)
			m.historyIdx = len(m.history)
			m.textarea.Reset()
			return m, func() tea.Msg {
				return InputSubmitMsg{Value: val}
			}
		case tea.KeyUp:
			if len(m.history) > 0 && m.historyIdx > 0 {
				m.historyIdx--
				m.textarea.SetValue(m.history[m.historyIdx])
				m.textarea.CursorEnd()
			}
			return m, nil
		case tea.KeyDown:
			if m.historyIdx < len(m.history)-1 {
				m.historyIdx++
				m.textarea.SetValue(m.history[m.historyIdx])
				m.textarea.CursorEnd()
			} else if m.historyIdx == len(m.history)-1 {
				m.historyIdx = len(m.history)
				m.textarea.Reset()
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// SetWidth updates the input width and re-applies background styling.
func (m *InputModel) SetWidth(w int) {
	m.width = w
	m.textarea.SetWidth(w)

	bgStyle := func() lipgloss.Style {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("0")).
			Foreground(lipgloss.Color("240")).
			Width(w)
	}

	m.textarea.FocusedStyle.Base = bgStyle()
	m.textarea.FocusedStyle.Placeholder = bgStyle()
	m.textarea.FocusedStyle.CursorLine = bgStyle()
	m.textarea.BlurredStyle.Base = bgStyle()
	m.textarea.BlurredStyle.Placeholder = bgStyle()
	m.textarea.BlurredStyle.CursorLine = bgStyle()
}

func (m InputModel) View() string {
	return m.textarea.View()
}
