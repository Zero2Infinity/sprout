// Package provider implements the Ollama LLM provider via its native API.
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/user/sprout/config"
	"github.com/user/sprout/message"
)

// chatMessage represents a single message in Ollama's native API format.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the request body for Ollama's /api/chat endpoint.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// chatStreamChunk is a single chunk from Ollama's streaming chat response.
type chatStreamChunk struct {
	Model              string        `json:"model"`
	Message            *chatMessage  `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration"`
	LoadDuration       int64         `json:"load_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration int64         `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       int64         `json:"eval_duration"`
}

// OllamaProvider wraps a raw HTTP client configured for Ollama's native API.
type OllamaProvider struct {
	httpClient   *http.Client
	model        string
	baseURL      string
	systemPrompt string
}

// NewOllamaProvider creates an OllamaProvider from the given application config.
// The baseURL is expected to be in the form "http://localhost:11434" (native API).
func NewOllamaProvider(cfg config.Config) *OllamaProvider {
	base := strings.TrimRight(cfg.Provider.BaseURL, "/")
	base = strings.TrimSuffix(base, "/v1")

	return &OllamaProvider{
		httpClient:   &http.Client{},
		model:        cfg.Provider.Model,
		baseURL:      base,
		systemPrompt: cfg.SystemPrompt,
	}
}

// ChatStream starts a streaming chat completion and returns a channel of events.
func (p *OllamaProvider) ChatStream(ctx context.Context, store *message.Store) (<-chan StreamEvent, error) {
	reqBody := chatRequest{
		Model:    p.model,
		Messages: p.buildMessages(store),
		Stream:   true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := p.baseURL + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, p.wrapError(err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, p.wrapError(fmt.Errorf("status %d", resp.StatusCode))
	}

	events := make(chan StreamEvent)

	go func() {
		defer close(events)
		defer resp.Body.Close()
		p.streamResponse(ctx, resp.Body, events)
	}()

	return events, nil
}

// streamResponse reads the NDJSON stream and emits events.
func (p *OllamaProvider) streamResponse(ctx context.Context, body io.Reader, events chan<- StreamEvent) {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var chunk chatStreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		// Emit content deltas
		if chunk.Message != nil && chunk.Message.Content != "" {
			events <- StreamEvent{ContentDelta: chunk.Message.Content}
		}

		// Emit completion with full metrics
		if chunk.Done {
			usage := Usage{
				PromptTokens:     chunk.PromptEvalCount,
				CompletionTokens: chunk.EvalCount,
				TotalTokens:      chunk.PromptEvalCount + chunk.EvalCount,

				TotalDurationNs:      chunk.TotalDuration,
				LoadDurationNs:       chunk.LoadDuration,
				PromptEvalDurationNs: chunk.PromptEvalDuration,
				EvalDurationNs:       chunk.EvalDuration,
			}
			events <- StreamEvent{
				Complete: true,
				Usage:    &usage,
			}
		}
	}
}

func (p *OllamaProvider) buildMessages(store *message.Store) []chatMessage {
	var msgs []chatMessage

	if p.systemPrompt != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: p.systemPrompt})
	}

	for _, m := range store.All() {
		switch m.Role {
		case message.RoleUser:
			msgs = append(msgs, chatMessage{Role: "user", Content: m.Content})
		case message.RoleAssistant:
			msgs = append(msgs, chatMessage{Role: "assistant", Content: m.Content})
		}
	}

	return msgs
}

func (p *OllamaProvider) wrapError(err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "dial tcp") {
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
