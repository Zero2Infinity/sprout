// Package agent manages the core chat loop: message dispatch, streaming, and clear.
package agent

import (
	"context"
	"fmt"

	"github.com/user/sprout/config"
	"github.com/user/sprout/message"
	"github.com/user/sprout/provider"
)

// Loop connects the LLM provider and message store for a single chat session.
type Loop struct {
	provider *provider.OllamaProvider
	store    *message.Store
}

// NewLoop creates a Loop with a fresh provider and message store.
func NewLoop(cfg config.Config) *Loop {
	return &Loop{
		provider: provider.NewOllamaProvider(cfg),
		store:    message.NewStore(),
	}
}

// Store returns the underlying message store for reading/writing messages.
func (l *Loop) Store() *message.Store {
	return l.store
}

// Provider returns the underlying LLM provider instance.
func (l *Loop) Provider() *provider.OllamaProvider {
	return l.provider
}

// SendMessage adds a user message to the store and starts a streaming response.
func (l *Loop) SendMessage(ctx context.Context, content string) (<-chan provider.StreamEvent, error) {
	l.store.Add(message.Message{
		Role:    message.RoleUser,
		Content: content,
	})

	events, err := l.provider.ChatStream(ctx, l.store)
	if err != nil {
		return nil, fmt.Errorf("starting chat: %w", err)
	}

	return events, nil
}

// Clear removes all messages from the store.
func (l *Loop) Clear() {
	l.store.Clear()
}
