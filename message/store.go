// Package message defines role types, the Message struct, and a thread-safe in-memory store.
package message

import (
	"sync"
	"time"
)

// Role identifies the sender of a message (user, assistant, or system).
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a single chat message with role, content, and metadata.
type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Store is a thread-safe, ordered collection of messages.
type Store struct {
	mu       sync.RWMutex
	messages []Message
}

// NewStore returns an empty, ready-to-use message store.
func NewStore() *Store {
	return &Store{}
}

// Add appends a message to the store in a thread-safe manner.
func (s *Store) Add(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

// All returns a defensive copy of all messages in insertion order.
func (s *Store) All() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Message, len(s.messages))
	copy(out, s.messages)
	return out
}

// LastAssistant returns the most recent assistant message, or nil.
func (s *Store) LastAssistant() *Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.messages) - 1; i >= 0; i-- {
		if s.messages[i].Role == RoleAssistant {
			return &s.messages[i]
		}
	}
	return nil
}

// Clear removes all messages from the store.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = nil
}

// Len returns the current number of messages in the store.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}
