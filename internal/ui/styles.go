// Package ui provides visual components using the Charm ecosystem
package ui

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/lipgloss/v2"
)

func PrintLogo() {
	possiblePaths := []string{
		"logo.png",
		"./logo.png",
		"../logo.png",
		"/usr/share/aceplay/logo.png",
		"/usr/local/share/aceplay/logo.png",
		filepath.Join(filepath.Dir(os.Args[0]), "logo.png"),
	}

	var logoPath string
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			logoPath = p
			break
		}
	}

	if logoPath == "" {
		return
	}

	cmd := exec.Command("chafa", "-s", "30x20", logoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// Theme colors
type Theme struct {
	Primary    color.Color
	Secondary  color.Color
	Success    color.Color
	Warning    color.Color
	Error      color.Color
	Info       color.Color
	Muted      color.Color
	Text       color.Color
	Background color.Color
}

// DefaultTheme is the default aceplay theme
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("#7D56F4"), // Purple
	Secondary:  lipgloss.Color("#04B575"), // Green
	Success:    lipgloss.Color("#04B575"), // Green
	Warning:    lipgloss.Color("#FFA500"), // Orange
	Error:      lipgloss.Color("#FF6B6B"), // Red
	Info:       lipgloss.Color("#4ECDC4"), // Cyan
	Muted:      lipgloss.Color("#808080"), // Gray
	Text:       lipgloss.Color("#FAFAFA"), // White
	Background: lipgloss.Color("#1A1A2E"), // Dark background
}

// Styles contains all application styles
type Styles struct {
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Text       lipgloss.Style
	MutedText  lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
	Box        lipgloss.Style
	BoxRounded lipgloss.Style
	Button     lipgloss.Style
	Spinner    lipgloss.Style
	Progress   lipgloss.Style
	Help       lipgloss.Style
}

// DefaultStyles returns default styles
func DefaultStyles() Styles {
	t := DefaultTheme

	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Primary).
			Padding(1, 0).
			Align(lipgloss.Center),

		Subtitle: lipgloss.NewStyle().
			Foreground(t.Muted).
			Italic(true).
			Padding(0, 1),

		Text: lipgloss.NewStyle().
			Foreground(t.Text),

		MutedText: lipgloss.NewStyle().
			Foreground(t.Muted),

		Success: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Success),

		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Error),

		Warning: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Warning),

		Info: lipgloss.NewStyle().
			Foreground(t.Info),

		Box: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Primary).
			Padding(1, 2),

		BoxRounded: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(1, 2),

		Button: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.Primary).
			Padding(0, 2).
			Bold(true),

		Spinner: lipgloss.NewStyle().
			Foreground(t.Primary),

		Progress: lipgloss.NewStyle().
			Foreground(t.Primary),

		Help: lipgloss.NewStyle().
			Foreground(t.Muted).
			Padding(0, 1),
	}
}

// Header renders the application header
func Header(title string, styles Styles) string {
	return styles.Title.Render(title)
}

// SuccessMessage renders a success message
func SuccessMessage(msg string, styles Styles) string {
	return styles.Success.Render("✓ " + msg)
}

// ErrorMessage renders an error message
func ErrorMessage(msg string, styles Styles) string {
	return styles.Error.Render("✗ " + msg)
}

// WarningMessage renders a warning message
func WarningMessage(msg string, styles Styles) string {
	return styles.Warning.Render("⚠ " + msg)
}

// InfoMessage renders an informational message
func InfoMessage(msg string, styles Styles) string {
	return styles.Info.Render("ℹ " + msg)
}

// Box renders content in a box
func Box(content string, styles Styles) string {
	return styles.Box.Render(content)
}

// BoxRounded renders content in a box with rounded borders
func BoxRounded(content string, styles Styles) string {
	return styles.BoxRounded.Render(content)
}

// HelpText renders help text
func HelpText(content string, styles Styles) string {
	return styles.Help.Render(content)
}

// LoadingSpinner returns the spinner character based on frame
func LoadingSpinner(frame int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return frames[frame%len(frames)]
}

// ProgressBar generates a simple progress bar
func ProgressBar(percent float64, width int, styles Styles) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * percent / 100)
	if filled > width {
		filled = width
	}

	empty := width - filled

	bar := ""
	if filled > 0 {
		bar += lipgloss.NewStyle().
			Background(styles.Box.GetBorderTopForeground()).
			Render(string(make([]byte, filled)))
	}
	if empty > 0 {
		bar += lipgloss.NewStyle().
			Foreground(DefaultTheme.Muted).
			Render(string(make([]byte, empty)))
	}

	return bar
}

var StreamStatusStyles = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#04B575")).
	Bold(true)

var StreamStatusBuffering = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFA500")).
	Bold(true)

var StreamStatusError = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF6B6B")).
	Bold(true)

var StreamStatsLabel = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#808080"))

var StreamStatsValue = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Bold(true)

var StreamPeersValue = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#4ECDC4")).
	Bold(true)

// RenderStreamStatus renders the stream status in a nice format
func RenderStreamStatus(status string, downloadSpeed int64, uploadSpeed int64, peers int) string {
	// Format speeds
	downStr := formatSpeed(downloadSpeed)
	upStr := formatSpeed(uploadSpeed)

	// Status color
	var statusStyle lipgloss.Style
	switch status {
	case "dl", "prebuf":
		statusStyle = StreamStatusStyles
	case "wait", "check", "loading":
		statusStyle = StreamStatusBuffering
	case "error":
		statusStyle = StreamStatusError
	default:
		statusStyle = StreamStatusStyles
	}

	// Status display
	statusDisplay := statusStyle.Render("● " + status)

	return statusDisplay + "  " +
		StreamStatsLabel.Render("down:") + " " + StreamStatsValue.Render(downStr) + "  " +
		StreamStatsLabel.Render("up:") + " " + StreamStatsValue.Render(upStr) + "  " +
		StreamStatsLabel.Render("peers:") + " " + StreamPeersValue.Render(fmt.Sprintf("%d", peers))
}

func formatSpeed(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	}
	if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
	}
	return fmt.Sprintf("%.1f MB/s", float64(bytesPerSec)/(1024*1024))
}
