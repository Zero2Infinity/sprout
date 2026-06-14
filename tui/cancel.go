package tui

type CancelMsg struct{}

type CancelProvider interface {
	Cancel() bool
}
