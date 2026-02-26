package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/crstian19/aceplay/internal/acestream"
	"github.com/crstian19/aceplay/internal/config"
	notify "github.com/crstian19/aceplay/internal/notify"
	"github.com/crstian19/aceplay/internal/player"
	"github.com/crstian19/aceplay/internal/ui"
	aceurl "github.com/crstian19/aceplay/pkg/acestream"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	cfg     *config.Config
	logger  ui.Logger
	styles  = ui.DefaultStyles()
)

func main() {
	install, _ := rootCmd.Flags().GetBool("install")
	if install {
		if err := runRegisterProtocol(nil, nil); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "aceplay",
	Short: "Play Ace Stream content",
	Long: `Aceplay is a modern CLI for playing Ace Stream content.

Open acestream:// links directly from your browser. The protocol handler is
registered automatically when you run: aceplay --install

Usage:
  aceplay <acestream-url>          Play a stream directly
  aceplay play <acestream-url>     Play a stream
  aceplay install                 Install protocol handler
  aceplay config                  Manage configuration`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		ui.SetLevel(logger, verbose)
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Received arguments", "args", args)
		if len(args) > 0 {
			urlStr := args[0]
			logger.Info("Processing URL", "url", urlStr)

			// Try URL decoding first (Firefox may URL-encode the protocol)
			if decoded, err := url.QueryUnescape(urlStr); err == nil {
				urlStr = decoded
				logger.Info("Decoded URL", "url", urlStr)
			}

			// If it doesn't have the prefix but looks like a content ID (40 hex chars), add prefix
			if !strings.HasPrefix(urlStr, "acestream://") && len(urlStr) >= 32 && len(urlStr) <= 40 {
				urlStr = "acestream://" + urlStr
			}

			if strings.HasPrefix(urlStr, "acestream://") {
				return runPlay(cmd, []string{urlStr})
			}
			// Maybe the URL is passed with quotes or as separate arg
			for _, arg := range args {
				if strings.Contains(arg, "acestream://") {
					return runPlay(cmd, []string{arg})
				}
			}
		}

		// First run: no config file exists yet → launch setup wizard
		if ui.IsFirstRun(cfg.ConfigPath) {
			availablePlayers := player.GetAvailablePlayers()
			selectedPlayer, useHLS, err := ui.SetupWizard(availablePlayers)
			if err != nil {
				return err
			}
			cfg.SetPlayer(selectedPlayer)
			cfg.SetHLS(useHLS)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			return nil
		}

		return cmd.Help()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	logger = ui.NewLogger(false)

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("install", false, "install and register protocol handler")
	rootCmd.PersistentFlags().StringP("player", "p", "", "video player to use")
	rootCmd.PersistentFlags().String("engine-host", "", "acestream-engine host")
	rootCmd.PersistentFlags().Int("engine-port", 0, "acestream-engine port")
	rootCmd.PersistentFlags().Duration("timeout", 0, "stream timeout")
	rootCmd.PersistentFlags().Bool("hls", false, "use HLS mode")

	_ = viper.BindPFlag("player", rootCmd.PersistentFlags().Lookup("player"))
	_ = viper.BindPFlag("engine.host", rootCmd.PersistentFlags().Lookup("engine-host"))
	_ = viper.BindPFlag("engine.port", rootCmd.PersistentFlags().Lookup("engine-port"))
	_ = viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	_ = viper.BindPFlag("hls", rootCmd.PersistentFlags().Lookup("hls"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(registerProtocolCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	configPath, _ := rootCmd.PersistentFlags().GetString("config")
	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
	}
}

var playCmd = &cobra.Command{
	Use:   "play [acestream-url]",
	Short: "Play an Ace Stream URL",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlay,
}

func runPlay(cmd *cobra.Command, args []string) error {
	urlStr := args[0]
	verbose, _ := cmd.Flags().GetBool("verbose")

	logger.Info("Parsing Ace Stream URL", "url", urlStr)

	aceURL, err := aceurl.ParseURL(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	logger.Info("Content ID extracted", "id", aceURL.ContentID)

	playerName := viper.GetString("player")
	if playerName == "" {
		playerName = cfg.Player
	}

	engineHost := viper.GetString("engine.host")
	if engineHost == "" {
		engineHost = cfg.Engine.Host
	}

	enginePort := viper.GetInt("engine.port")
	if enginePort == 0 {
		enginePort = cfg.Engine.Port
	}

	timeout := viper.GetDuration("timeout")
	if timeout == 0 {
		timeout = cfg.Timeout
	}

	logger.Info("Initializing player", "player", playerName)

	playerInstance, err := player.NewPlayer(playerName)
	if err != nil {
		return fmt.Errorf("failed to initialize player: %w", err)
	}

	logger.Info("Connecting to acestream-engine", "host", engineHost, "port", enginePort)

	acestreamClient := acestream.NewClient(
		acestream.WithHost(engineHost),
		acestream.WithPort(enginePort),
		acestream.WithTimeout(timeout),
		acestream.WithConnectTimeout(5*time.Second),
		acestream.WithAutoStart("acestreamengine"),
	)

	if err := acestreamClient.StartEngine(); err != nil {
		return fmt.Errorf("failed to start engine: %w", err)
	}
	defer func() { _ = acestreamClient.StopEngine() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("Waiting for stream to be ready...")
	streamURL, err := acestreamClient.WaitForStream(ctx, aceURL.ContentID)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	logger.Info("Starting playback", "url", streamURL)

	notifier := notify.GetNotifier()
	if notifier.IsAvailable() {
		_ = notifier.Notify("Aceplay", "Playing stream: "+aceURL.ContentID)
	}

	fmt.Println()
	fmt.Println(styles.Success.Render("✓ Stream ready! Playing..."))
	fmt.Println(styles.MutedText.Render("  URL: " + streamURL))
	fmt.Println()

	if verbose {
		logger.Info("Launching player", "player", playerInstance.Executable())
	}

	if err := playerInstance.Play(context.Background(), streamURL); err != nil {
		return fmt.Errorf("playback failed: %w", err)
	}

	return nil
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		var err error
		cfg, err = config.Load("")
		if err != nil {
			return err
		}
	}

	fmt.Println(styles.Title.Render("Aceplay Configuration"))
	fmt.Println()
	fmt.Printf("Player:     %s\n", cfg.Player)
	fmt.Printf("Engine:     %s:%d\n", cfg.Engine.Host, cfg.Engine.Port)
	fmt.Printf("Timeout:    %s\n", cfg.Timeout)
	fmt.Printf("HLS:        %v\n", cfg.HLS)
	fmt.Printf("Verbose:    %v\n", cfg.Verbose)

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	if cfg == nil {
		var err error
		cfg, err = config.Load("")
		if err != nil {
			return err
		}
	}

	switch key {
	case "player":
		cfg.SetPlayer(value)
	case "engine.host":
		cfg.SetEngineHost(value)
	case "engine.port":
		var port int
		if _, err := fmt.Sscanf(value, "%d", &port); err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.SetEnginePort(port)
	case "timeout":
		duration, _ := time.ParseDuration(value)
		cfg.SetTimeout(duration)
	case "hls":
		cfg.SetHLS(value == "true")
	case "verbose":
		cfg.SetVerbose(value == "true")
	default:
		return fmt.Errorf("unknown key: %s", key)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(styles.Success.Render("✓ Configuration updated"))

	return nil
}

var registerProtocolCmd = &cobra.Command{
	Use:   "register-protocol",
	Short: "Register acestream:// protocol handler",
	Long: `Register acestream:// as a handled protocol in your system.

This allows you to click acestream:// links in your browser and have
Aceplay automatically open and play them.

On Linux, this uses xdg-utils.`,
	RunE: runRegisterProtocol,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and register protocol handler",
	Long: `Install aceplay and register the acestream:// protocol handler.

This is the recommended way to set up aceplay for browser integration.
It creates the necessary desktop entry and registers the protocol.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRegisterProtocol(cmd, args)
	},
}

func runRegisterProtocol(cmd *cobra.Command, args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	return registerLinux(execPath)
}

func registerLinux(execPath string) error {
	desktopFile := fmt.Sprintf(`[Desktop Entry]
Name=Aceplay
Exec=%s play %%u
Type=Application
Terminal=false
MimeType=x-scheme-handler/acestream;
Categories=Network;Video;
`, execPath)

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}

	desktopDir := filepath.Join(xdgDataHome, "applications")
	if err := os.MkdirAll(desktopDir, 0755); err != nil {
		return fmt.Errorf("failed to create desktop directory: %w", err)
	}

	desktopPath := filepath.Join(desktopDir, "aceplay.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopFile), 0755); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	fmt.Println(styles.Info.Render("ℹ Running xdg-mime to register protocol handler..."))

	if err := runCmd("xdg-mime", "default", "aceplay.desktop", "x-scheme-handler/acestream"); err != nil {
		return fmt.Errorf("failed to register protocol: %w", err)
	}

	fmt.Println(styles.Success.Render("✓ Protocol handler registered successfully!"))
	fmt.Println(styles.MutedText.Render("  You can now click acestream:// links in your browser."))

	return nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(styles.Title.Render("Aceplay"))
		fmt.Printf("Version:    %s\n", version)
		fmt.Printf("Commit:     %s\n", commit)
		fmt.Printf("Date:       %s\n", date)
		fmt.Printf("Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
