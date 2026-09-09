// Package engine manages the lifecycle of the local acestream-engine process.
//
// It is deliberately private to aceplay: it shells out to a binary, guesses
// ports and owns an OS process, none of which belongs in a reusable library.
// Talking to the engine once it listens is the job of pkg/acestream.
package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/crstian19/aceplay/pkg/acestream"
)

const (
	defaultCommand = "acestreamengine"
	defaultPort    = 6878

	readyTimeout  = 30 * time.Second
	readyInterval = 500 * time.Millisecond
	probeTimeout  = 2 * time.Second
)

// commonPorts are the ports an already running engine is usually listening on.
var commonPorts = []int{6878, 45615, 8080, 9999}

// Manager handles the acestream-engine process lifecycle
type Manager struct {
	command    string
	process    *os.Process
	autoStart  bool
	wasStarted bool
	httpPort   int
}

// New creates a new engine manager. An empty command falls back to the
// acestream-engine binary on PATH.
func New(command string, autoStart bool) *Manager {
	if command == "" {
		command = defaultCommand
	}
	return &Manager{
		command:   command,
		autoStart: autoStart,
		httpPort:  defaultPort,
	}
}

// Port returns the HTTP port the engine is running on
func (m *Manager) Port() int {
	return m.httpPort
}

// WasStarted reports whether this process started the engine
func (m *Manager) WasStarted() bool {
	return m.wasStarted
}

// Start makes sure an engine is listening and returns the port it answers on.
//
// An engine that was already running is reused as-is (and never stopped by
// [Manager.Stop]); otherwise one is spawned when auto-start is enabled.
func (m *Manager) Start(ctx context.Context) (int, error) {
	if m.process != nil {
		return m.httpPort, nil // Already started by us
	}

	if port := findRunningPort(ctx); port != 0 {
		m.httpPort = port
		return port, nil
	}

	if !m.autoStart {
		return 0, fmt.Errorf("acestream-engine is not running and auto-start is disabled")
	}

	enginePath, err := exec.LookPath(m.command)
	if err != nil {
		return 0, fmt.Errorf("acestream-engine not found in PATH: %w", err)
	}

	cmd := exec.CommandContext(ctx, enginePath,
		"--client-console",
		"--http-port", fmt.Sprint(defaultPort),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start acestream-engine: %w", err)
	}

	m.process = cmd.Process
	m.wasStarted = true
	m.httpPort = defaultPort

	if err := waitUntilReady(ctx, defaultPort); err != nil {
		return 0, err
	}

	return defaultPort, nil
}

// Stop stops the acestream-engine, but only if we started it
func (m *Manager) Stop() error {
	if !m.wasStarted || m.process == nil {
		return nil // We didn't start it, don't stop it
	}

	if err := m.process.Kill(); err != nil {
		return fmt.Errorf("failed to stop acestream-engine: %w", err)
	}

	m.process = nil
	m.wasStarted = false
	return nil
}

// waitUntilReady polls the engine until it answers or readyTimeout elapses.
func waitUntilReady(ctx context.Context, port int) error {
	client := acestream.NewClient(acestream.WithPort(port))
	deadline := time.Now().Add(readyTimeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readyInterval):
		}

		probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
		ready := client.IsRunning(probeCtx)
		cancel()

		if ready {
			return nil
		}
	}

	return fmt.Errorf("acestream-engine failed to start within %s", readyTimeout)
}

// findRunningPort looks for an engine already listening on a common port.
func findRunningPort(ctx context.Context) int {
	for _, port := range commonPorts {
		client := acestream.NewClient(acestream.WithPort(port))

		probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
		ready := client.IsRunning(probeCtx)
		cancel()

		if ready {
			return port
		}
	}

	return 0
}
