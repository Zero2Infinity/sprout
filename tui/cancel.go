package tui

// CancelMsg signals that an in-flight streaming response was cancelled.
type CancelMsg struct{}

// CancelProvider allows a component to cancel the current streaming response.
type CancelProvider interface {
	Cancel() bool
}
