package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/user/sprout/agent"
	"github.com/user/sprout/config"
	"github.com/user/sprout/message"
	"github.com/user/sprout/provider"
	"github.com/user/sprout/session"
)

type appState int

const (
	stateIdle appState = iota
	stateStreaming
	stateCancelled
)

type streamChunkMsg struct {
	content string
	complete bool
	tokens  int
	err     error
}

type Model struct {
	cfg    config.Config
	sess   *session.Session
	loop   *agent.Loop
	state  appState
	cancel context.CancelFunc

	header HeaderModel
	chat   ChatModel
	footer FooterModel
	input  InputModel
	toast  ToastModel

	width  int
	height int
}

func NewModel(cfg config.Config, sess *session.Session, loop *agent.Loop) Model {
	return Model{
		cfg:    cfg,
		sess:   sess,
		loop:   loop,
		state:  stateIdle,
		header: NewHeaderModel(mustCwd(), sess.ID[:6], cfg.Provider.Model),
		chat:   NewChatModel(80, 20),
		footer: NewFooterModel(),
		input:  NewInputModel(),
		toast:  NewToastModel(),
	}
}

func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc && m.state == stateStreaming {
			if m.cancel != nil {
				m.cancel()
			}
			m.state = stateCancelled
			cmds = append(cmds, func() tea.Msg {
				return streamChunkMsg{complete: true}
			})
			return m, tea.Batch(cmds...)
		}

		if m.state == stateStreaming {
			return m, nil
		}

		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case InputSubmitMsg:
		if m.state == stateIdle {
			m.state = stateStreaming
			ctx, cancel := context.WithCancel(context.Background())
			m.cancel = cancel

			events, err := m.loop.SendMessage(ctx, msg.Value)
			if err != nil {
				m.state = stateIdle
				cmds = append(cmds, m.toast.Show(fmt.Sprintf("Error: %v", err)))
				return m, tea.Batch(cmds...)
			}

			m.chat.AddMessage(message.Message{
				Role:    message.RoleUser,
				Content: msg.Value,
			})

			cmds = append(cmds, m.waitForStream(events))
			return m, tea.Batch(cmds...)
		}

	case streamChunkMsg:
		if msg.err != nil {
			m.state = stateIdle
			cmds = append(cmds, m.toast.Show(fmt.Sprintf("Error: %v", msg.err)))
			return m, tea.Batch(cmds...)
		}

		if msg.content != "" {
			m.footer.currentTokens += 1
			m.footer.totalTokens += 1
		}

		if msg.complete {
			m.state = stateIdle
			m.cancel = nil
			store := m.loop.Store()
			if last := store.LastAssistant(); last != nil {
				m.chat.AddMessage(*last)
			}
			return m, tea.Batch(cmds...)
		}

		return m, tea.Batch(cmds...)
	}

	if m.state == stateIdle {
		if _, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	m.chat.Update(msg)
	m.header, _ = m.header.Update(msg)
	m.footer, _ = m.footer.Update(msg)
	m.toast, _ = m.toast.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	header := m.header.View()
	footer := m.footer.View()
	input := m.input.View()
	toast := m.toast.View()

	chatHeight := m.height - 4
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chat.SetSize(m.width, chatHeight)
	chat := m.chat.View()

	stateIndicator := ""
	if m.state == stateStreaming {
		stateIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(" ● streaming")
	} else if m.state == stateCancelled {
		stateIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(" ✕ cancelled")
	}

	var bottom string
	if toast != "" {
		bottom = toast
	} else {
		bottom = footer + stateIndicator
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Repeat("─", m.width),
		chat,
		strings.Repeat("─", m.width),
		input,
		bottom,
	)
}

func (m *Model) updateLayout() {
	chatHeight := m.height - 4
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chat.SetSize(m.width, chatHeight)
}

func (m Model) waitForStream(events <-chan provider.StreamEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-events
		if !ok {
			return streamChunkMsg{complete: true}
		}
		if ev.Err != nil {
			return streamChunkMsg{err: ev.Err}
		}
		if ev.Complete && ev.Usage != nil {
			return streamChunkMsg{complete: true, tokens: ev.Usage.Tokens}
		}
		return streamChunkMsg{content: ev.ContentDelta}
	}
}
