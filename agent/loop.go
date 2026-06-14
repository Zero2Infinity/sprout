package agent

import (
	"context"
	"fmt"

	"github.com/user/sprout/config"
	"github.com/user/sprout/message"
	"github.com/user/sprout/provider"
)

type Loop struct {
	provider *provider.OllamaProvider
	store    *message.Store
}

func NewLoop(cfg config.Config) *Loop {
	return &Loop{
		provider: provider.NewOllamaProvider(cfg),
		store:    message.NewStore(),
	}
}

func (l *Loop) Store() *message.Store {
	return l.store
}

func (l *Loop) Provider() *provider.OllamaProvider {
	return l.provider
}

func (l *Loop) SendMessage(ctx context.Context, content string) (<-chan provider.StreamEvent, error) {
	l.store.Add(message.Message{
		Role:    message.RoleUser,
		Content: content,
	})

	events, err := l.provider.ChatStream(ctx, l.store)
	if err != nil {
		return nil, fmt.Errorf("starting chat: %w", err)
	}

	var assistantContent string
	var tokens int

	go func() {
		for ev := range events {
			if ev.Err != nil {
				continue
			}
			if ev.ContentDelta != "" {
				assistantContent += ev.ContentDelta
			}
			if ev.Complete && ev.Usage != nil {
				tokens = ev.Usage.Tokens
			}
		}

		if assistantContent != "" {
			l.store.Add(message.Message{
				Role:    message.RoleAssistant,
				Content: assistantContent,
				Tokens:  tokens,
			})
		}
	}()

	return events, nil
}

func (l *Loop) Clear() {
	l.store.Clear()
}
