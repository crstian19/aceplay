package player

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlayer(t *testing.T) {
	tempDir := t.TempDir()
	fakeExecutable := filepath.Join(tempDir, "mpv")

	err := os.WriteFile(fakeExecutable, []byte("#!/bin/sh\necho fake mpv"), 0755)
	require.NoError(t, err)

	t.Setenv("PATH", tempDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

	player, err := NewPlayer("mpv")
	require.NoError(t, err)
	assert.Equal(t, "mpv", player.Name())
	assert.NotEmpty(t, player.Executable())
}

func TestNewPlayer_NotFound(t *testing.T) {
	// Use a name that doesn't exist
	_, err := NewPlayer("nonexistent-player-12345")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIsAvailable(t *testing.T) {
	tempDir := t.TempDir()
	fakeExecutable := filepath.Join(tempDir, "vlc")

	require.NoError(t, os.WriteFile(fakeExecutable, []byte("#!/bin/sh\necho fake vlc"), 0755))

	t.Setenv("PATH", tempDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

	assert.True(t, IsAvailable("vlc"))
	assert.False(t, IsAvailable("nonexistent"))
}

func TestGetAvailablePlayers(t *testing.T) {
	tempDir := t.TempDir()

	players := []string{"mpv", "vlc"}
	for _, p := range players {
		exe := filepath.Join(tempDir, p)
		require.NoError(t, os.WriteFile(exe, []byte("fake"), 0755))
	}

	t.Setenv("PATH", tempDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

	available := GetAvailablePlayers()
	assert.GreaterOrEqual(t, len(available), 2)
	assert.Contains(t, available, "mpv")
	assert.Contains(t, available, "vlc")
}

func TestGetDefaultArgs(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name:     PlayerMPV,
			expected: []string{"--force-window=immediate"},
		},
		{
			name:     PlayerVLC,
			expected: []string{"--play-and-exit"},
		},
		{
			name:     PlayerFFplay,
			expected: []string{"-window_title"},
		},
		{
			name:     "unknown",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := getDefaultArgs(tt.name)
			if tt.expected == nil {
				assert.Empty(t, args)
			} else {
				assert.NotEmpty(t, args)
				assert.Contains(t, args, tt.expected[0])
			}
		})
	}
}

func TestPlayer_SetArgs(t *testing.T) {
	tempDir := t.TempDir()
	fakeExecutable := filepath.Join(tempDir, "mpv")
	require.NoError(t, os.WriteFile(fakeExecutable, []byte("fake"), 0755))

	t.Setenv("PATH", tempDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

	player, err := NewPlayer("mpv")
	require.NoError(t, err)

	customArgs := []string{"--fs", "--no-border"}
	player.SetArgs(customArgs)
}

func TestPlayer_AddArgs(t *testing.T) {
	tempDir := t.TempDir()
	fakeExecutable := filepath.Join(tempDir, "mpv")
	require.NoError(t, os.WriteFile(fakeExecutable, []byte("fake"), 0755))

	t.Setenv("PATH", tempDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

	player, err := NewPlayer("mpv")
	require.NoError(t, err)

	player.AddArgs("--fs", "--no-border")
}

func TestFindExecutable(t *testing.T) {
	tempDir := t.TempDir()
	fakeExecutable := filepath.Join(tempDir, "testplayer")

	err := os.WriteFile(fakeExecutable, []byte("#!/bin/sh\necho test"), 0755)
	require.NoError(t, err)

	t.Setenv("PATH", tempDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

	found, err := findExecutable("testplayer")
	require.NoError(t, err)
	assert.Equal(t, fakeExecutable, found)
}

func TestFindExecutable_NotFound(t *testing.T) {
	_, err := findExecutable("nonexistent-player-test-123")
	require.Error(t, err)
}

func TestPlayer_Play_NotInitialized(t *testing.T) {
	player := &Player{}
	err := player.Play(context.TODO(), "http://example.com/stream")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}
