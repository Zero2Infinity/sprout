package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type InputSubmitMsg struct {
	Value string
}

type InputModel struct {
	textarea    textarea.Model
	history     []string
	historyIdx  int
	currentLine string
}

func NewInputModel() InputModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(2)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)
	ta.FocusedStyle.CursorLine = ta.FocusedStyle.CursorLine.Foreground(nil)
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

func (m InputModel) View() string {
	return m.textarea.View()
}
