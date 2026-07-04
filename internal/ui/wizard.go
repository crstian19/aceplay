// Package ui provides setup wizard functionality
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/crstian19/aceplay/internal/config"
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

var playerDescriptions = map[string]string{
	"mpv":    "Recommended - Fast and lightweight",
	"vlc":    "Feature-rich, supports many formats",
	"ffplay": "Part of FFmpeg suite",
}

func ConfigEditor(availablePlayers []string, currentConfig *config.Config) error {
	styles := DefaultStyles()

	fmt.Println()
	fmt.Println(styles.Title.Render("Aceplay Configuration"))
	fmt.Println()

	if !isTerminal() {
		fmt.Println(styles.Error.Render("✗ Interactive mode requires a terminal"))
		fmt.Println(styles.MutedText.Render("  Use: aceplay config set <key> <value>"))
		return nil
	}

	cfg := &configEditorState{
		player:           currentConfig.Player,
		engineHost:       currentConfig.Engine.Host,
		enginePort:       fmt.Sprintf("%d", currentConfig.Engine.Port),
		timeout:          currentConfig.Timeout.String(),
		hls:              currentConfig.HLS,
		verbose:          currentConfig.Verbose,
		availablePlayers: availablePlayers,
	}

	playerOptions := make([]huh.Option[string], len(availablePlayers))
	for i, p := range availablePlayers {
		desc := playerDescriptions[p]
		if desc == "" {
			desc = "Video player"
		}
		playerOptions[i] = huh.NewOption(fmt.Sprintf("%s - %s", p, desc), p)
	}

	timeoutOptions := []huh.Option[string]{
		huh.NewOption("30 seconds", "30s"),
		huh.NewOption("1 minute", "1m"),
		huh.NewOption("2 minutes", "2m"),
		huh.NewOption("5 minutes", "5m"),
		huh.NewOption("10 minutes", "10m"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Video Player").
				Description("Select your preferred video player:\n"),

			huh.NewSelect[string]().
				Title("Default Player").
				Description("This player will be used by default when playing streams").
				Options(playerOptions...).
				Value(&cfg.player),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Ace Stream Engine").
				Description("Configure the acestream-engine connection:\n"),

			huh.NewInput().
				Title("Engine Host").
				Description("Hostname where acestream-engine is running").
				Placeholder("localhost").
				Value(&cfg.engineHost),

			huh.NewInput().
				Title("Engine Port").
				Description("Port number for acestream-engine").
				Placeholder("6878").
				Value(&cfg.enginePort),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Playback Options").
				Description("Configure stream playback settings:\n"),

			huh.NewSelect[string]().
				Title("Timeout").
				Description("Maximum time to wait for a stream to start").
				Options(timeoutOptions...).
				Value(&cfg.timeout),

			huh.NewConfirm().
				Title("HLS Mode").
				Description("Enable HLS streaming by default?").
				Value(&cfg.hls).
				Affirmative("Yes").
				Negative("No"),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Advanced Options").
				Description("Additional settings:\n"),

			huh.NewConfirm().
				Title("Verbose Mode").
				Description("Enable detailed logging output?").
				Value(&cfg.verbose).
				Affirmative("Yes").
				Negative("No"),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("configuration wizard error: %w", err)
	}

	currentConfig.Player = cfg.player
	currentConfig.Engine.Host = cfg.engineHost

	port := 6878
	if cfg.enginePort != "" {
		if _, err := fmt.Sscanf(cfg.enginePort, "%d", &port); err != nil {
			port = 6878
		}
	}
	currentConfig.Engine.Port = port

	if timeout, err := time.ParseDuration(cfg.timeout); err == nil {
		currentConfig.Timeout = timeout
	}

	currentConfig.HLS = cfg.hls
	currentConfig.Verbose = cfg.verbose

	fmt.Println()
	fmt.Println(styles.Success.Render("✓ Configuration updated successfully!"))
	fmt.Println()

	fmt.Println("Your settings:")
	fmt.Printf("  • Player:     %s\n", lipgloss.NewStyle().Bold(true).Render(currentConfig.Player))
	fmt.Printf("  • Engine:     %s:%d\n", currentConfig.Engine.Host, currentConfig.Engine.Port)
	fmt.Printf("  • Timeout:    %s\n", currentConfig.Timeout)
	fmt.Printf("  • HLS:        %v\n", currentConfig.HLS)
	fmt.Printf("  • Verbose:    %v\n", currentConfig.Verbose)
	fmt.Println()

	return nil
}

type configEditorState struct {
	player           string
	engineHost       string
	enginePort       string
	timeout          string
	hls              bool
	verbose          bool
	availablePlayers []string
}

func ConfigMenu(availablePlayers []string, currentConfig *config.Config, showConfig func(), editConfig func(), setConfig func()) error {
	styles := DefaultStyles()

	if !isTerminal() {
		showConfig()
		return nil
	}

	for {
		fmt.Println()
		fmt.Println(styles.Title.Render("⚙️  Aceplay Configuration"))

		currentConfigStr := fmt.Sprintf(`Current Settings:
  🎬 Player:    %s
  🔌 Engine:    %s:%d
  ⏱️  Timeout:   %s
  📺 HLS:       %s
  📝 Verbose:   %s`,
			lipgloss.NewStyle().Bold(true).Foreground(DefaultTheme.Primary).Render(currentConfig.Player),
			currentConfig.Engine.Host,
			currentConfig.Engine.Port,
			currentConfig.Timeout,
			renderBool(currentConfig.HLS),
			renderBool(currentConfig.Verbose),
		)

		var selected int

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Choose an option:").
					Description(currentConfigStr+"\n\n"),

				huh.NewSelect[int]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("✏️  Edit all settings", 0),
						huh.NewOption("🎬 Change player", 1),
						huh.NewOption("🔌 Change engine host/port", 2),
						huh.NewOption("⏱️  Change timeout", 3),
						huh.NewOption("📺 Toggle HLS", 4),
						huh.NewOption("❌ Exit", 5),
					).
					Value(&selected),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("menu error: %w", err)
		}

		fmt.Println()

		switch selected {
		case 0:
			editConfig()
			if !continuePrompt() {
				return nil
			}
		case 1:
			if err := runPlayerSelector(availablePlayers, currentConfig); err != nil {
				return err
			}
			if !continuePrompt() {
				return nil
			}
		case 2:
			if err := runEngineConfig(currentConfig); err != nil {
				return err
			}
			if !continuePrompt() {
				return nil
			}
		case 3:
			if err := runTimeoutSelector(currentConfig); err != nil {
				return err
			}
			if !continuePrompt() {
				return nil
			}
		case 4:
			currentConfig.HLS = !currentConfig.HLS
			if err := currentConfig.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Println(styles.Success.Render("✓ HLS mode: " + renderBool(currentConfig.HLS)))
			fmt.Println()
			if !continuePrompt() {
				return nil
			}
		case 5:
			fmt.Println(styles.MutedText.Render("👋 Goodbye!"))
			fmt.Println()
			return nil
		}
	}
}

func continuePrompt() bool {
	var selected int

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Continue?").
				Options(
					huh.NewOption("🔙 Back to menu", 0),
					huh.NewOption("❌ Exit", 1),
				).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return false
	}

	return selected == 0
}

func renderBool(b bool) string {
	if b {
		return "✓ Yes"
	}
	return "✗ No"
}

func runPlayerSelector(availablePlayers []string, currentConfig *config.Config) error {
	styles := DefaultStyles()
	var player string

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
			huh.NewSelect[string]().
				Title("Select Player").
				Options(playerOptions...).
				Value(&player),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	currentConfig.Player = player
	if err := currentConfig.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println(styles.Success.Render("✓ Player set to: " + player))
	fmt.Println()
	return nil
}

func runEngineConfig(currentConfig *config.Config) error {
	styles := DefaultStyles()
	var host string
	var port string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Engine Host").
				Placeholder("localhost").
				Value(&host),
			huh.NewInput().
				Title("Engine Port").
				Placeholder("6878").
				Value(&port),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if host != "" {
		currentConfig.Engine.Host = host
	}
	if port != "" {
		if _, err := fmt.Sscanf(port, "%d", &currentConfig.Engine.Port); err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
	}
	if err := currentConfig.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println(styles.Success.Render("✓ Engine: " + currentConfig.Engine.Host + ":" + fmt.Sprintf("%d", currentConfig.Engine.Port)))
	fmt.Println()
	return nil
}

func runTimeoutSelector(currentConfig *config.Config) error {
	styles := DefaultStyles()
	var timeout string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Timeout").
				Options(
					huh.NewOption("30 seconds", "30s"),
					huh.NewOption("1 minute", "1m"),
					huh.NewOption("2 minutes", "2m"),
					huh.NewOption("5 minutes", "5m"),
					huh.NewOption("10 minutes", "10m"),
				).
				Value(&timeout),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if dur, err := time.ParseDuration(timeout); err == nil {
		currentConfig.Timeout = dur
		if err := currentConfig.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}
	fmt.Println(styles.Success.Render("✓ Timeout: " + currentConfig.Timeout.String()))
	fmt.Println()
	return nil
}
