// Package provider implements the Ollama LLM provider via the OpenAI-compatible API.
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"

	"github.com/user/sprout/config"
	"github.com/user/sprout/message"
)

// OllamaProvider wraps an OpenAI-compatible client configured for Ollama.
type OllamaProvider struct {
	client      openai.Client
	model       string
	baseURL     string
	systemPrompt string
}

// NewOllamaProvider creates an OllamaProvider from the given application config.
func NewOllamaProvider(cfg config.Config) *OllamaProvider {
	opts := []option.RequestOption{
		option.WithBaseURL(cfg.Provider.BaseURL),
	}
	if cfg.Provider.APIKey != "" && cfg.Provider.APIKey != "no-key" {
		opts = append(opts, option.WithAPIKey(cfg.Provider.APIKey))
	}

	return &OllamaProvider{
		client:      openai.NewClient(opts...),
		model:       cfg.Provider.Model,
		baseURL:     cfg.Provider.BaseURL,
		systemPrompt: cfg.SystemPrompt,
	}
}

// StreamEvent carries streaming content deltas, completion signals, or errors.
type StreamEvent struct {
	ContentDelta string
	Complete     bool
	Usage        *message.Message
	Err          error
}

// ChatStream starts a streaming chat completion and returns a channel of events.
func (p *OllamaProvider) ChatStream(ctx context.Context, store *message.Store) (<-chan StreamEvent, error) {
	msgs := p.buildMessages(store)

	stream := p.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: p.model,
		Messages: msgs,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: param.Opt[bool]{Value: true},
		},
	})

	if stream.Err() != nil {
		return nil, p.wrapError(stream.Err())
	}

	events := make(chan StreamEvent)

	go func() {
		defer close(events)
		defer stream.Close()

		for stream.Next() {
			chunk := stream.Current()

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					events <- StreamEvent{ContentDelta: delta.Content}
				}
			}

			if chunk.Usage.TotalTokens > 0 {
				events <- StreamEvent{
					Complete: true,
					Usage: &message.Message{
						Role:    message.RoleAssistant,
						Tokens:  int(chunk.Usage.TotalTokens),
						Content: "",
					},
				}
			}
		}

		if err := stream.Err(); err != nil {
			events <- StreamEvent{Err: p.wrapError(err)}
		}
	}()

	return events, nil
}

func (p *OllamaProvider) buildMessages(store *message.Store) []openai.ChatCompletionMessageParamUnion {
	var msgs []openai.ChatCompletionMessageParamUnion

	if p.systemPrompt != "" {
		msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: param.Opt[string]{Value: p.systemPrompt},
				},
			},
		})
	}

	for _, m := range store.All() {
		switch m.Role {
		case message.RoleUser:
			msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: param.Opt[string]{Value: m.Content},
					},
				},
			})
		case message.RoleAssistant:
			msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: param.Opt[string]{Value: m.Content},
					},
				},
			})
		}
	}

	return msgs
}

func (p *OllamaProvider) wrapError(err error) error {
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "dial tcp") {
		return fmt.Errorf("cannot connect to LLM server at %s\nEnsure your server is running and accessible", p.baseURL)
	}
	return fmt.Errorf("LLM error: %w", err)
}

// Model returns the configured Ollama model name.
func (p *OllamaProvider) Model() string {
	return p.model
}

// BaseURL returns the configured Ollama API base URL.
func (p *OllamaProvider) BaseURL() string {
	return p.baseURL
}
