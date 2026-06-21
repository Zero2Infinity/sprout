// Package session handles creating, loading, saving, and restoring chat sessions.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/user/sprout/message"
	"github.com/user/sprout/provider"
)

// TokenUsage tracks per-session token consumption and timing.
type TokenUsage struct {
	PromptTokens     int   `json:"promptTokens"`
	CompletionTokens int   `json:"completionTokens"`
	TotalTokens      int   `json:"totalTokens"`
	TotalDurationNs  int64 `json:"totalDurationNs,omitempty"`
	EvalDurationNs   int64 `json:"evalDurationNs,omitempty"`
}

// TokensPerSec returns session-level output tokens per second, or 0 if no eval timing.
func (t TokenUsage) TokensPerSec() float64 {
	if t.EvalDurationNs == 0 || t.CompletionTokens == 0 {
		return 0
	}
	return float64(t.CompletionTokens) / float64(t.EvalDurationNs) * 1e9
}

// Session represents a single chat session with messages, metadata, and token usage.
type Session struct {
	ID            string            `json:"id"`
	Model         string            `json:"model"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	Messages      []message.Message `json:"messages"`
	TokenUsage    TokenUsage        `json:"tokenUsage"`
}

func sessionFilePath(dataDir, id string) string {
	return filepath.Join(dataDir, id+".json")
}

// Create initializes a new session with a UUID and current timestamp.
func Create(model string) *Session {
	now := time.Now()
	return &Session{
		ID:        uuid.New().String(),
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Load reads a session from disk by ID.
func Load(dataDir, id string) (*Session, error) {
	path := sessionFilePath(dataDir, id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %s not found; run `sprout ls` to list available sessions", id)
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}
	return &sess, nil
}

// Save persists the session to a JSON file in the data directory.
func Save(dataDir string, sess *Session) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("creating sessions directory: %w", err)
	}

	sess.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	path := sessionFilePath(dataDir, sess.ID)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing session file: %w", err)
	}
	return nil
}

// LoadOrCreate returns an existing session if id is provided, otherwise creates a new one.
func LoadOrCreate(dataDir, id, model string) (*Session, error) {
	if id != "" {
		return Load(dataDir, id)
	}
	return Create(model), nil
}

// RestoreMessages loads session messages into a message store.
func RestoreMessages(sess *Session, store *message.Store) {
	for _, msg := range sess.Messages {
		store.Add(msg)
	}
}

// SyncFromStore copies messages from the in-memory store to the session for persistence.
func SyncFromStore(sess *Session, store *message.Store) {
	sess.Messages = store.All()
}

// UpdateTokenUsage accumulates token usage and timing from an API response.
func (s *Session) UpdateTokenUsage(usage provider.Usage) {
	s.TokenUsage.PromptTokens = usage.PromptTokens
	s.TokenUsage.CompletionTokens += usage.CompletionTokens
	s.TokenUsage.TotalTokens = usage.TotalTokens
	s.TokenUsage.TotalDurationNs += usage.TotalDurationNs
	s.TokenUsage.EvalDurationNs += usage.EvalDurationNs
}

// List returns all sessions in the data directory, sorted by most recently updated.
func List(dataDir string) ([]Session, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions directory: %w", err)
	}

	var sessions []Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dataDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}

	for i := 0; i < len(sessions); i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[j].UpdatedAt.After(sessions[i].UpdatedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	return sessions, nil
}
