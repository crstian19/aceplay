// Package player provides functionality for launching video players
package player

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Supported players
const (
	PlayerMPV    = "mpv"
	PlayerVLC    = "vlc"
	PlayerFFplay = "ffplay"
)

// Player represents a video player
type Player struct {
	name       string
	executable string
	args       []string
}

// NewPlayer creates a new player
func NewPlayer(name string) (*Player, error) {
	// Find the executable
	executable, err := findExecutable(name)
	if err != nil {
		return nil, fmt.Errorf("player '%s' not found: %w", name, err)
	}

	return &Player{
		name:       name,
		executable: executable,
		args:       getDefaultArgs(name),
	}, nil
}

// findExecutable looks for the player executable in PATH
func findExecutable(name string) (string, error) {
	// Try to find directly first
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// On Windows, add extensions
	if runtime.GOOS == "windows" {
		extensions := []string{".exe", ".cmd", ".bat"}
		for _, ext := range extensions {
			if path, err := exec.LookPath(name + ext); err == nil {
				return path, nil
			}
		}
	}

	// Common paths by operating system
	commonPaths := getCommonPaths(name)
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("executable not found")
}

// getCommonPaths returns common paths where the player might be installed
func getCommonPaths(name string) []string {
	var paths []string

	switch runtime.GOOS {
	case "darwin":
		// macOS
		paths = append(paths,
			filepath.Join("/Applications", fmt.Sprintf("%s.app", strings.Title(name)), "Contents", "MacOS", name),
			filepath.Join("/usr/local", "bin", name),
			filepath.Join("/opt", "homebrew", "bin", name),
		)
	case "windows":
		// Windows
		programFiles := os.Getenv("ProgramFiles")
		if programFiles == "" {
			programFiles = `C:\Program Files`
		}
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		if programFilesX86 == "" {
			programFilesX86 = `C:\Program Files (x86)`
		}

		switch name {
		case PlayerVLC:
			paths = append(paths,
				filepath.Join(programFiles, "VideoLAN", "VLC", "vlc.exe"),
				filepath.Join(programFilesX86, "VideoLAN", "VLC", "vlc.exe"),
			)
		case PlayerMPV:
			paths = append(paths,
				filepath.Join(programFiles, "mpv", "mpv.exe"),
				filepath.Join(programFilesX86, "mpv", "mpv.exe"),
			)
		case PlayerFFplay:
			paths = append(paths,
				filepath.Join(programFiles, "ffmpeg", "bin", "ffplay.exe"),
				filepath.Join(programFilesX86, "ffmpeg", "bin", "ffplay.exe"),
			)
		}
	default:
		// Linux and other Unix
		paths = append(paths,
			filepath.Join("/usr", "bin", name),
			filepath.Join("/usr", "local", "bin", name),
			filepath.Join("/opt", name, "bin", name),
			filepath.Join(os.Getenv("HOME"), ".local", "bin", name),
		)
	}

	return paths
}

// getDefaultArgs returns default arguments for each player
func getDefaultArgs(name string) []string {
	switch name {
	case PlayerMPV:
		return []string{
			"--force-window=immediate",
			"--cache=yes",
			"--cache-secs=30",
			"--demuxer-max-bytes=50M",
			"--demuxer-max-back-bytes=25M",
		}
	case PlayerVLC:
		return []string{
			"--play-and-exit",
			"--network-caching=3000",
		}
	case PlayerFFplay:
		return []string{
			"-window_title", "Aceplay",
			"-fflags", "nobuffer",
			"-flags", "low_delay",
			"-strict", "experimental",
		}
	default:
		return []string{}
	}
}

// IsAvailable checks if a player is available
func IsAvailable(name string) bool {
	_, err := NewPlayer(name)
	return err == nil
}

// GetAvailablePlayers returns a list of available players
func GetAvailablePlayers() []string {
	players := []string{PlayerMPV, PlayerVLC, PlayerFFplay}
	var available []string

	for _, p := range players {
		if IsAvailable(p) {
			available = append(available, p)
		}
	}

	return available
}

// Play starts playback of a URL
func (p *Player) Play(ctx context.Context, streamURL string) error {
	if p.executable == "" {
		return fmt.Errorf("player not initialized")
	}

	// Prepare arguments
	args := append(p.args, streamURL)

	// Create command
	cmd := exec.CommandContext(ctx, p.executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start playback
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting player: %w", err)
	}

	return nil
}

// PlayAndWait starts playback and waits for it to finish
func (p *Player) PlayAndWait(ctx context.Context, streamURL string, timeout time.Duration) error {
	if p.executable == "" {
		return fmt.Errorf("player not initialized")
	}

	// Prepare arguments
	args := append(p.args, streamURL)

	// Create command with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start and wait
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout waiting for playback")
		}
		return fmt.Errorf("playback error: %w", err)
	}

	return nil
}

// Name returns the player name
func (p *Player) Name() string {
	return p.name
}

// Executable returns the path to the executable
func (p *Player) Executable() string {
	return p.executable
}

// SetArgs sets custom arguments
func (p *Player) SetArgs(args []string) {
	p.args = args
}

// AddArgs adds additional arguments
func (p *Player) AddArgs(args ...string) {
	p.args = append(p.args, args...)
}
