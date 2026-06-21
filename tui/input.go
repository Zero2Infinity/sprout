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
	// ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(2)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)

	style := inputBaseStyle(0)
	ta.FocusedStyle.Base = style
	ta.FocusedStyle.Placeholder = style
	ta.FocusedStyle.CursorLine = style
	ta.BlurredStyle.Base = style
	ta.BlurredStyle.Placeholder = style
	ta.BlurredStyle.CursorLine = style

	return InputModel{
		textarea:   ta,
		history:    make([]string, 0),
		historyIdx: -1,
	}
}

// inputBaseStyle returns the base lipgloss style for the textarea.
// width=0 means no width constraint; width>0 applies Width(w).
func inputBaseStyle(width int) lipgloss.Style {
	s := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("240"))
	if width > 0 {
		s = s.Width(width)
	}
	return s
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

	style := inputBaseStyle(w)
	m.textarea.FocusedStyle.Base = style
	m.textarea.FocusedStyle.Placeholder = style
	m.textarea.FocusedStyle.CursorLine = style
	m.textarea.BlurredStyle.Base = style
	m.textarea.BlurredStyle.Placeholder = style
	m.textarea.BlurredStyle.CursorLine = style
}

func (m InputModel) View() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Width(m.width)
	return style.Render(m.textarea.View())
}
