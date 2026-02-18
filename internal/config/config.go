// Package config provides configuration management for aceplay
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

// Default values
const (
	DefaultPlayer         = "mpv"
	DefaultEngineHost     = "localhost"
	DefaultEnginePort     = 6878
	DefaultTimeout        = 60 * time.Second
	DefaultConnectTimeout = 5 * time.Second
	DefaultHLS            = false
	DefaultVerbose        = false
)

// Config represents the application configuration
type Config struct {
	// Player is the video player to use (mpv, vlc, ffplay, etc.)
	Player string `mapstructure:"player"`

	// Engine contains the acestream-engine configuration
	Engine EngineConfig `mapstructure:"engine"`

	// Timeout is the maximum time to wait for a stream
	Timeout time.Duration `mapstructure:"timeout"`

	// ConnectTimeout is the maximum time to connect to the engine
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`

	// HLS enables HLS mode
	HLS bool `mapstructure:"hls"`

	// Verbose enables verbose mode with detailed logging
	Verbose bool `mapstructure:"verbose"`

	// ConfigPath is the path to the config file (not saved in file)
	ConfigPath string `mapstructure:"-"`
}

// EngineConfig contains the acestream-engine configuration
type EngineConfig struct {
	// Host of acestream-engine
	Host string `mapstructure:"host"`

	// Port of acestream-engine
	Port int `mapstructure:"port"`

	// AutoStart automatically starts the engine if not running
	AutoStart bool `mapstructure:"auto_start"`

	// AutoStartCommand is the command to start the engine
	AutoStartCommand string `mapstructure:"auto_start_command"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Player:         DefaultPlayer,
		Timeout:        DefaultTimeout,
		ConnectTimeout: DefaultConnectTimeout,
		HLS:            DefaultHLS,
		Verbose:        DefaultVerbose,
		Engine: EngineConfig{
			Host:             DefaultEngineHost,
			Port:             DefaultEnginePort,
			AutoStart:        false,
			AutoStartCommand: "",
		},
	}
}

// Load loads the configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	cfg := NewConfig()

	// Setup viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add search paths
	if configPath != "" {
		// Use specific path
		v.AddConfigPath(configPath)
		cfg.ConfigPath = configPath
	} else {
		// Search in standard locations
		configDirs := getConfigDirs()
		for _, dir := range configDirs {
			v.AddConfigPath(dir)
		}
		cfg.ConfigPath = configDirs[0]
	}

	// Setup environment variables
	v.SetEnvPrefix("ACEPLAY")
	v.AutomaticEnv()

	// Setup defaults
	setDefaults(v)

	// Try to read file
	if err := v.ReadInConfig(); err != nil {
		// Not an error if file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading configuration: %w", err)
		}
	}

	// Unmarshal configuration
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper) {
	v.SetDefault("player", DefaultPlayer)
	v.SetDefault("timeout", DefaultTimeout)
	v.SetDefault("connect_timeout", DefaultConnectTimeout)
	v.SetDefault("hls", DefaultHLS)
	v.SetDefault("verbose", DefaultVerbose)
	v.SetDefault("engine.host", DefaultEngineHost)
	v.SetDefault("engine.port", DefaultEnginePort)
	v.SetDefault("engine.auto_start", false)
}

// Save saves the configuration to file
func (c *Config) Save() error {
	if c.ConfigPath == "" {
		c.ConfigPath = getConfigDirs()[0]
	}

	// Ensure directory exists
	if err := os.MkdirAll(c.ConfigPath, 0755); err != nil {
		return fmt.Errorf("error creating configuration directory: %w", err)
	}

	configFile := filepath.Join(c.ConfigPath, "config.yaml")

	v := viper.New()
	v.SetConfigFile(configFile)

	// Set values
	v.Set("player", c.Player)
	v.Set("engine", c.Engine)
	v.Set("timeout", c.Timeout)
	v.Set("connect_timeout", c.ConnectTimeout)
	v.Set("hls", c.HLS)
	v.Set("verbose", c.Verbose)

	if err := v.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return v.SafeWriteConfig()
		}
		return fmt.Errorf("error saving configuration: %w", err)
	}

	return nil
}

// getConfigDirs returns standard configuration directories
func getConfigDirs() []string {
	var dirs []string

	// XDG_CONFIG_HOME
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		dirs = append(dirs, filepath.Join(xdgConfig, "aceplay"))
	}

	// Home directory
	home, err := os.UserHomeDir()
	if err == nil {
		switch runtime.GOOS {
		case "darwin":
			// macOS
			dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "aceplay"))
		case "windows":
			// Windows
			if appData := os.Getenv("APPDATA"); appData != "" {
				dirs = append(dirs, filepath.Join(appData, "aceplay"))
			}
		default:
			// Linux and other Unix
			dirs = append(dirs, filepath.Join(home, ".config", "aceplay"))
		}
	}

	return dirs
}

// GetEngineAddress returns the full engine address (host:port)
func (c *Config) GetEngineAddress() string {
	return fmt.Sprintf("%s:%d", c.Engine.Host, c.Engine.Port)
}

// SetPlayer configures the player
func (c *Config) SetPlayer(player string) {
	c.Player = player
}

// SetEngineHost configures the engine host
func (c *Config) SetEngineHost(host string) {
	c.Engine.Host = host
}

// SetEnginePort configures the engine port
func (c *Config) SetEnginePort(port int) {
	c.Engine.Port = port
}

// SetTimeout configures the timeout
func (c *Config) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// SetHLS enables/disables HLS
func (c *Config) SetHLS(enabled bool) {
	c.HLS = enabled
}

// SetVerbose enables/disables verbose mode
func (c *Config) SetVerbose(enabled bool) {
	c.Verbose = enabled
}
