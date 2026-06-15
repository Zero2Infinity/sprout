package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/sprout/message"
)

type ChatModel struct {
	messages         []message.Message
	streamingContent string
	viewport         viewport.Model
	width            int
	height           int
}

func NewChatModel(width, height int) ChatModel {
	vp := viewport.New(width, height)
	return ChatModel{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ChatModel) View() string {
	return m.viewport.View()
}

func (m *ChatModel) SetStreamingContent(content string) {
	m.streamingContent = content
	m.updateContent()
}

func (m *ChatModel) AddMessage(msg message.Message) {
	m.streamingContent = ""
	m.messages = append(m.messages, msg)
	m.updateContent()
}

func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	m.updateContent()
}

func (m *ChatModel) updateContent() {
	var sb strings.Builder
	for _, msg := range m.messages {
		var styled string
		switch msg.Role {
		case message.RoleUser:
			styled = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Render("You: ") + msg.Content
		case message.RoleAssistant:
			styled = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render("Assistant: ") + msg.Content
		case message.RoleSystem:
			styled = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")).Render("System: ") + msg.Content
		}
		sb.WriteString(styled)
		sb.WriteString("\n\n")
	}
	if m.streamingContent != "" {
		styled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render("Assistant: ") + m.streamingContent
		styled += lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(" ▌")
		sb.WriteString(styled)
	}
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m ChatModel) StatusLine() string {
	return fmt.Sprintf("Messages: %d", len(m.messages))
}
