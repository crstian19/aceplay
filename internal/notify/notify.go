// Package notify provides desktop notification functionality
package notify

import (
	"os/exec"
)

// Notifier defines the interface for notifications
type Notifier interface {
	Notify(title, message string) error
	IsAvailable() bool
}

// LibnotifyNotifier uses libnotify (notify-send on Linux)
type LibnotifyNotifier struct{}

// NewLibnotifyNotifier creates a new libnotify notifier
func NewLibnotifyNotifier() *LibnotifyNotifier {
	return &LibnotifyNotifier{}
}

// IsAvailable checks if notify-send is available
func (n *LibnotifyNotifier) IsAvailable() bool {
	_, err := exec.LookPath("notify-send")
	return err == nil
}

// Notify sends a notification using notify-send
func (n *LibnotifyNotifier) Notify(title, message string) error {
	cmd := exec.Command("notify-send",
		"--app-name=aceplay",
		"--icon=video-x-generic",
		title,
		message,
	)
	return cmd.Run()
}

// GetNotifier returns the appropriate notifier for the operating system
func GetNotifier() Notifier {
	libnotify := NewLibnotifyNotifier()
	if libnotify.IsAvailable() {
		return libnotify
	}

	return &NoopNotifier{}
}

// NoopNotifier does nothing (fallback)
type NoopNotifier struct{}

// IsAvailable always returns true
func (n *NoopNotifier) IsAvailable() bool {
	return true
}

// Notify does nothing
func (n *NoopNotifier) Notify(title, message string) error {
	return nil
}
