package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	assert.Equal(t, DefaultPlayer, cfg.Player)
	assert.Equal(t, DefaultTimeout, cfg.Timeout)
	assert.Equal(t, DefaultConnectTimeout, cfg.ConnectTimeout)
	assert.Equal(t, DefaultHLS, cfg.HLS)
	assert.Equal(t, DefaultVerbose, cfg.Verbose)
	assert.Equal(t, DefaultEngineHost, cfg.Engine.Host)
	assert.Equal(t, DefaultEnginePort, cfg.Engine.Port)
	assert.False(t, cfg.Engine.AutoStart)
}

func TestConfig_GetEngineAddress(t *testing.T) {
	cfg := NewConfig()
	assert.Equal(t, "localhost:6878", cfg.GetEngineAddress())

	cfg.Engine.Host = "192.168.1.1"
	cfg.Engine.Port = 8080
	assert.Equal(t, "192.168.1.1:8080", cfg.GetEngineAddress())
}

func TestConfig_Setters(t *testing.T) {
	cfg := NewConfig()

	cfg.SetPlayer("vlc")
	assert.Equal(t, "vlc", cfg.Player)

	cfg.SetEngineHost("10.0.0.1")
	assert.Equal(t, "10.0.0.1", cfg.Engine.Host)

	cfg.SetEnginePort(9000)
	assert.Equal(t, 9000, cfg.Engine.Port)

	cfg.SetTimeout(120 * time.Second)
	assert.Equal(t, 120*time.Second, cfg.Timeout)

	cfg.SetHLS(true)
	assert.True(t, cfg.HLS)

	cfg.SetVerbose(true)
	assert.True(t, cfg.Verbose)
}

func TestLoad(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "aceplay")
	require.NoError(t, os.MkdirAll(configPath, 0755))

	// Create configuration file
	configContent := `
player: vlc
hls: true
verbose: true
timeout: 2m
engine:
  host: 192.168.1.100
  port: 8080
  auto_start: true
`
	configFile := filepath.Join(configPath, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load configuration
	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "vlc", cfg.Player)
	assert.True(t, cfg.HLS)
	assert.True(t, cfg.Verbose)
	assert.Equal(t, 2*time.Minute, cfg.Timeout)
	assert.Equal(t, "192.168.1.100", cfg.Engine.Host)
	assert.Equal(t, 8080, cfg.Engine.Port)
	assert.True(t, cfg.Engine.AutoStart)
}

func TestLoad_Defaults(t *testing.T) {
	// Load without file (should use defaults)
	tempDir := t.TempDir()
	cfg, err := Load(tempDir)
	require.NoError(t, err)

	assert.Equal(t, DefaultPlayer, cfg.Player)
	assert.Equal(t, DefaultTimeout, cfg.Timeout)
	assert.Equal(t, DefaultHLS, cfg.HLS)
	assert.Equal(t, DefaultVerbose, cfg.Verbose)
}

func TestSave(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	cfg := &Config{
		Player:         "vlc",
		Timeout:        120 * time.Second,
		ConnectTimeout: 10 * time.Second,
		HLS:            true,
		Verbose:        false,
		Engine: EngineConfig{
			Host:             "192.168.1.100",
			Port:             8080,
			AutoStart:        true,
			AutoStartCommand: "acestream-engine",
		},
		ConfigPath: tempDir,
	}

	// Save configuration
	err := cfg.Save()
	require.NoError(t, err)

	// Verify file exists
	configFile := filepath.Join(tempDir, "config.yaml")
	_, err = os.Stat(configFile)
	require.NoError(t, err)

	// Load and verify
	loadedCfg, err := Load(tempDir)
	require.NoError(t, err)

	assert.Equal(t, cfg.Player, loadedCfg.Player)
	assert.Equal(t, cfg.HLS, loadedCfg.HLS)
	assert.Equal(t, cfg.Engine.Host, loadedCfg.Engine.Host)
	assert.Equal(t, cfg.Engine.Port, loadedCfg.Engine.Port)
}

func TestSave_CreatesDirectory(t *testing.T) {
	// Create temporary directory that doesn't exist
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "subdir", "aceplay")

	cfg := NewConfig()
	cfg.ConfigPath = nonExistentDir

	err := cfg.Save()
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(nonExistentDir)
	require.NoError(t, err)
}

func TestLoad_InvalidConfig(t *testing.T) {
	// Create invalid file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "aceplay")
	require.NoError(t, os.MkdirAll(configPath, 0755))

	configContent := `
player: [invalid yaml structure
`
	configFile := filepath.Join(configPath, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Try to load
	_, err = Load(configPath)
	assert.Error(t, err)
}
