// Package notify provides desktop notification functionality
package notify

import (
	"fmt"
	"os/exec"
	"runtime"
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

// OSXNotifier uses osascript for macOS notifications
type OSXNotifier struct{}

// NewOSXNotifier creates a new macOS notifier
func NewOSXNotifier() *OSXNotifier {
	return &OSXNotifier{}
}

// IsAvailable checks if osascript is available (always on macOS)
func (n *OSXNotifier) IsAvailable() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath("osascript")
	return err == nil
}

// Notify sends a notification using osascript
func (n *OSXNotifier) Notify(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// WindowsNotifier uses PowerShell for Windows notifications
type WindowsNotifier struct{}

// NewWindowsNotifier creates a new Windows notifier
func NewWindowsNotifier() *WindowsNotifier {
	return &WindowsNotifier{}
}

// IsAvailable checks if PowerShell is available
func (n *WindowsNotifier) IsAvailable() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	_, err := exec.LookPath("powershell")
	return err == nil
}

// Notify sends a notification using PowerShell
func (n *WindowsNotifier) Notify(title, message string) error {
	psScript := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`[System.Windows.Forms.MessageBox]::Show("%s", "%s")`,
		message, title,
	)
	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}

// GetNotifier returns the appropriate notifier for the operating system
func GetNotifier() Notifier {
	// Try libnotify first (Linux)
	libnotify := NewLibnotifyNotifier()
	if libnotify.IsAvailable() {
		return libnotify
	}

	// Try macOS
	osx := NewOSXNotifier()
	if osx.IsAvailable() {
		return osx
	}

	// Try Windows
	windows := NewWindowsNotifier()
	if windows.IsAvailable() {
		return windows
	}

	// Fallback to noop
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
