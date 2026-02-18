package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	// Verify all styles are initialized
	assert.NotNil(t, styles.Title)
	assert.NotNil(t, styles.Subtitle)
	assert.NotNil(t, styles.Text)
	assert.NotNil(t, styles.MutedText)
	assert.NotNil(t, styles.Success)
	assert.NotNil(t, styles.Error)
	assert.NotNil(t, styles.Warning)
	assert.NotNil(t, styles.Info)
	assert.NotNil(t, styles.Box)
	assert.NotNil(t, styles.BoxRounded)
	assert.NotNil(t, styles.Button)
	assert.NotNil(t, styles.Spinner)
	assert.NotNil(t, styles.Progress)
	assert.NotNil(t, styles.Help)
}

func TestDefaultTheme(t *testing.T) {
	// Verify theme has defined colors
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Primary)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Secondary)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Success)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Error)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Warning)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Info)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Muted)
	assert.NotEqual(t, lipgloss.Color(""), DefaultTheme.Text)
}

func TestHeader(t *testing.T) {
	styles := DefaultStyles()
	result := Header("Test Title", styles)

	assert.Contains(t, result, "Test Title")
	// Result should have styles applied (contains ANSI codes)
	assert.NotEqual(t, "Test Title", result)
}

func TestSuccessMessage(t *testing.T) {
	styles := DefaultStyles()
	result := SuccessMessage("Operation completed", styles)

	assert.Contains(t, result, "✓")
	assert.Contains(t, result, "Operation completed")
}

func TestErrorMessage(t *testing.T) {
	styles := DefaultStyles()
	result := ErrorMessage("Something went wrong", styles)

	assert.Contains(t, result, "✗")
	assert.Contains(t, result, "Something went wrong")
}

func TestWarningMessage(t *testing.T) {
	styles := DefaultStyles()
	result := WarningMessage("Warning message", styles)

	assert.Contains(t, result, "⚠")
	assert.Contains(t, result, "Warning message")
}

func TestInfoMessage(t *testing.T) {
	styles := DefaultStyles()
	result := InfoMessage("Information", styles)

	assert.Contains(t, result, "ℹ")
	assert.Contains(t, result, "Information")
}

func TestBox(t *testing.T) {
	styles := DefaultStyles()
	content := "Box content"
	result := Box(content, styles)

	assert.Contains(t, result, content)
}

func TestBoxRounded(t *testing.T) {
	styles := DefaultStyles()
	content := "Rounded box content"
	result := BoxRounded(content, styles)

	assert.Contains(t, result, content)
}

func TestHelpText(t *testing.T) {
	styles := DefaultStyles()
	content := "Press q to quit"
	result := HelpText(content, styles)

	assert.Contains(t, result, content)
}

func TestLoadingSpinner(t *testing.T) {
	// Verify it returns a valid spinner character
	frame0 := LoadingSpinner(0)
	frame5 := LoadingSpinner(5)
	frame10 := LoadingSpinner(10)

	assert.NotEmpty(t, frame0)
	assert.NotEmpty(t, frame5)
	assert.NotEmpty(t, frame10)

	// Frame 10 should equal frame 0 (cycle)
	assert.Equal(t, frame0, frame10)
}

func TestLoadingSpinner_Cycle(t *testing.T) {
	// Verify cycle works correctly
	frames := []string{}
	for i := 0; i < 15; i++ {
		frames = append(frames, LoadingSpinner(i))
	}

	// First 10 frames should be different
	uniqueFrames := make(map[string]bool)
	for i := 0; i < 10; i++ {
		uniqueFrames[frames[i]] = true
	}
	assert.Len(t, uniqueFrames, 10)

	// Frame 10 should repeat frame 0
	assert.Equal(t, frames[0], frames[10])
}
