// Package ui provides setup wizard functionality
package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// IsFirstRun checks if this is the first time running the application
func IsFirstRun(configPath string) bool {
	if configPath == "" {
		// Check default locations
		home, err := os.UserHomeDir()
		if err != nil {
			return true
		}
		configPath = filepath.Join(home, ".config", "aceplay")
	}

	configFile := filepath.Join(configPath, "config.yaml")
	_, err := os.Stat(configFile)
	return os.IsNotExist(err)
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) == os.ModeCharDevice
}

// SetupWizard runs the first-time setup wizard
// Returns the selected player and whether the user wants HLS by default
func SetupWizard(availablePlayers []string) (player string, hls bool, err error) {
	if len(availablePlayers) == 0 {
		return "", false, fmt.Errorf("no video players found. Please install mpv, vlc, or ffplay")
	}

	styles := DefaultStyles()

	// Welcome message
	fmt.Println()
	fmt.Println(styles.Title.Render("⚡ Welcome to Aceplay!"))
	fmt.Println()
	fmt.Println("This appears to be your first time running Aceplay.")

	// Check if we're in an interactive terminal
	if !isTerminal() {
		// Non-interactive mode - use defaults
		fmt.Println()
		fmt.Println("Running in non-interactive mode. Using default settings:")

		// Select first available player as default
		selectedPlayer := availablePlayers[0]
		if len(availablePlayers) > 0 {
			// Prefer mpv if available
			for _, p := range availablePlayers {
				if p == "mpv" {
					selectedPlayer = p
					break
				}
			}
		}

		fmt.Printf("  • Default player: %s\n", lipgloss.NewStyle().Bold(true).Render(selectedPlayer))
		fmt.Printf("  • HLS mode: false\n")
		fmt.Println()
		fmt.Println(styles.Info.Render("💡 Tip: You can change these settings later with:"))
		fmt.Println(styles.MutedText.Render("     aceplay config set player <player>"))
		fmt.Println(styles.MutedText.Render("     aceplay config show"))
		fmt.Println()

		return selectedPlayer, false, nil
	}

	fmt.Println("Let's set up your preferences.")
	fmt.Println()

	var selectedPlayer string
	var useHLS bool

	// Player selection form
	playerOptions := make([]huh.Option[string], len(availablePlayers))
	for i, p := range availablePlayers {
		desc := ""
		switch p {
		case "mpv":
			desc = "Recommended - Fast and lightweight"
		case "vlc":
			desc = "Feature-rich, supports many formats"
		case "ffplay":
			desc = "Part of FFmpeg suite"
		}
		playerOptions[i] = huh.NewOption(fmt.Sprintf("%s - %s", p, desc), p)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("First-time Setup").
				Description("Choose your default video player:\n"),

			huh.NewSelect[string]().
				Title("Default Player").
				Description("This player will be used by default when playing streams").
				Options(playerOptions...).
				Value(&selectedPlayer),

			huh.NewConfirm().
				Title("HLS Mode").
				Description("Enable HLS streaming by default? (Useful for some streams)").
				Value(&useHLS).
				Affirmative("Yes").
				Negative("No"),
		),
	)

	if err := form.Run(); err != nil {
		return "", false, err
	}

	// Show confirmation
	fmt.Println()
	fmt.Println(SuccessMessage("Configuration saved!", styles))
	fmt.Println()
	fmt.Println("Your settings:")
	fmt.Printf("  • Default player: %s\n", lipgloss.NewStyle().Bold(true).Render(selectedPlayer))
	fmt.Printf("  • HLS mode: %t\n", useHLS)
	fmt.Println()
	fmt.Println(styles.Info.Render("💡 Tip: You can change these settings anytime with:"))
	fmt.Println(styles.MutedText.Render("     aceplay config set player <player>"))
	fmt.Println(styles.MutedText.Render("     aceplay config set hls <true/false>"))
	fmt.Println(styles.MutedText.Render("     aceplay config show"))
	fmt.Println()

	return selectedPlayer, useHLS, nil
}

// ShowWelcomeReminder shows a reminder about configuration options
// This can be shown after the wizard or when the user might want to know about config
func ShowWelcomeReminder(player string) {
	styles := DefaultStyles()

	fmt.Println()
	fmt.Println(styles.BoxRounded.Render(
		fmt.Sprintf("Now playing with %s!\n\n%s\n%s\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(DefaultTheme.Primary).Render(player),
			"To change settings in the future, use:",
			"  aceplay config set player <mpv|vlc|ffplay>",
			"  aceplay config show",
		),
	))
	fmt.Println()
}
