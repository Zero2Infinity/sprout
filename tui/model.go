// Package tui implements the Bubble Tea terminal UI with header, chat, footer, input, and toast.
package tui

import (
	"context"
	"fmt"
	"os"

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
	content          string
	complete         bool
	hasUsage         bool
	promptTokens     int
	completionTokens int
	totalTokens      int
	err              error
}

// Model is the root Bubble Tea model composing all sub-components.
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

	streamingContent string
	streamEvents     <-chan provider.StreamEvent
	width            int
	height           int
}

// NewModel initializes the root TUI model with config, session, and agent loop.
func NewModel(cfg config.Config, sess *session.Session, loop *agent.Loop) Model {
	chat := NewChatModel(80, 20)
	if len(sess.Messages) > 0 {
		chat.LoadMessages(sess.Messages)
	}

	return Model{
		cfg:    cfg,
		sess:   sess,
		loop:   loop,
		state:  stateIdle,
		header: NewHeaderModel(mustCwd(), sess.ID[:6], cfg.Provider.Model),
		chat:   chat,
		footer: NewFooterModel().
			WithUsage(sess.TokenUsage.PromptTokens, sess.TokenUsage.CompletionTokens).
			WithContext(sess.TokenUsage.TotalTokens, provider.ModelContext(cfg.Provider.Model)),
		input: NewInputModel(),
		toast: NewToastModel(),
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
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

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

		// Block Enter during streaming (don't lose typed text), allow all other keys
		if m.state == stateStreaming && msg.Type == tea.KeyEnter {
			return m, nil
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)

	case InputSubmitMsg:
		if m.state == stateIdle {
			m.state = stateStreaming
			ctx, cancel := context.WithCancel(context.Background())
			m.cancel = cancel
			m.streamingContent = ""

			events, err := m.loop.SendMessage(ctx, msg.Value)
			if err != nil {
				m.state = stateIdle
				cmds = append(cmds, m.toast.Show(fmt.Sprintf("Error: %v", err)))
				return m, tea.Batch(cmds...)
			}

			m.streamEvents = events

			m.chat.AddMessage(message.Message{
				Role:    message.RoleUser,
				Content: msg.Value,
			})

			cmds = append(cmds, m.waitForStream())
			return m, tea.Batch(cmds...)
		}

	case streamChunkMsg:
		if msg.err != nil {
			m.state = stateIdle
			m.streamingContent = ""
			m.streamEvents = nil
			cmds = append(cmds, m.toast.Show(fmt.Sprintf("Error: %v", msg.err)))
			return m, tea.Batch(cmds...)
		}

		if msg.content != "" {
			m.streamingContent += msg.content
			m.chat.SetStreamingContent(m.streamingContent)
		}

		if msg.complete {
			m.state = stateIdle
			m.cancel = nil
			m.streamEvents = nil

			if m.streamingContent != "" {
				m.loop.Store().Add(message.Message{
					Role:    message.RoleAssistant,
					Content: m.streamingContent,
				})
				if last := m.loop.Store().LastAssistant(); last != nil {
					m.chat.AddMessage(*last)
				}
				session.SyncFromStore(m.sess, m.loop.Store())
			}

			if msg.hasUsage {
				m.footer = m.footer.WithUsage(msg.promptTokens, msg.completionTokens)
				m.footer = m.footer.WithContext(msg.totalTokens, provider.ModelContext(m.cfg.Provider.Model))
				m.sess.UpdateTokenUsage(provider.Usage{
					PromptTokens:     msg.promptTokens,
					CompletionTokens: msg.completionTokens,
					TotalTokens:      msg.totalTokens,
				})
			}

			m.chat.SetStreamingContent("")
			m.streamingContent = ""
			return m, tea.Batch(cmds...)
		}

		cmds = append(cmds, m.waitForStream())
		return m, tea.Batch(cmds...)
	}

	var chatCmd tea.Cmd
	m.chat, chatCmd = m.chat.Update(msg)
	m.header, _ = m.header.Update(msg)
	m.footer, _ = m.footer.Update(msg)
	m.toast, _ = m.toast.Update(msg)
	cmds = append(cmds, chatCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	innerWidth := m.width - 2

	header := m.header.View()
	footer := m.footer.View()
	input := m.input.View()
	toast := m.toast.View()

	chatHeight := m.height - 6
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chat.SetSize(innerWidth, chatHeight)
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

	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		chat,
		input,
		bottom,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Width(m.width - 2).
		Render(body)
}

func (m *Model) updateLayout() {
	innerWidth := m.width - 2
	chatHeight := m.height - 6
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chat.SetSize(innerWidth, chatHeight)
	m.input.SetWidth(innerWidth)
}

func (m Model) waitForStream() tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-m.streamEvents
		if !ok {
			return streamChunkMsg{complete: true}
		}
		if ev.Err != nil {
			return streamChunkMsg{err: ev.Err}
		}
		if ev.Complete {
			if ev.Usage != nil {
				return streamChunkMsg{
					complete:         true,
					hasUsage:         true,
					promptTokens:     ev.Usage.PromptTokens,
					completionTokens: ev.Usage.CompletionTokens,
					totalTokens:      ev.Usage.TotalTokens,
				}
			}
			return streamChunkMsg{complete: true}
		}
		return streamChunkMsg{content: ev.ContentDelta}
	}
}
