// Package provider defines the Provider interface and shared types for LLM providers.
package provider

import (
	"context"
	"strings"

	"github.com/user/sprout/message"
)

// Usage tracks token consumption and timing for a single API request.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int

	// Timing fields (Ollama-specific, zero for other providers)
	TotalDurationNs      int64
	LoadDurationNs       int64
	PromptEvalDurationNs int64
	EvalDurationNs       int64
}

// TokensPerSec returns output tokens per second, or 0 if no eval timing is available.
func (u Usage) TokensPerSec() float64 {
	if u.EvalDurationNs == 0 || u.CompletionTokens == 0 {
		return 0
	}
	return float64(u.CompletionTokens) / float64(u.EvalDurationNs) * 1e9
}

// StreamEvent carries streaming content deltas, completion signals, or errors.
type StreamEvent struct {
	ContentDelta string
	Complete     bool
	Usage        *Usage
	Err          error
}

// Provider is the interface that all LLM providers must implement.
type Provider interface {
	ChatStream(ctx context.Context, store *message.Store) (<-chan StreamEvent, error)
	Model() string
	BaseURL() string
}

// DefaultContextWindow is the fallback context window size when a model is not recognized.
const DefaultContextWindow = 32768

type modelEntry struct {
	prefix  string
	context int
}

// modelContexts maps model name prefixes to their context window sizes.
// Entries are ordered longest-prefix-first to ensure correct matching.
var modelContexts = []modelEntry{
	{"mistral-nemo", 128000},
	{"mistral-large", 128000},
	{"command-r", 131072},
	{"llama3.2", 128000},
	{"llama3.1", 131072},
	{"deepseek", 65536},
	{"qwen3", 32768},
	{"qwen2.5", 32768},
	{"qwen2", 32768},
	{"gemma3", 131072},
	{"mistral", 32768},
	{"mixtral", 32768},
	{"codellama", 16384},
	{"llama3", 8192},
	{"gemma", 8192},
	{"phi4", 16384},
	{"phi3", 4096},
}

// ModelContext returns the context window size for a given model name.
// Falls back to DefaultContextWindow if the model prefix is not recognized.
func ModelContext(model string) int {
	m := strings.ToLower(model)
	for _, entry := range modelContexts {
		if strings.HasPrefix(m, entry.prefix) {
			return entry.context
		}
	}
	return DefaultContextWindow
}
