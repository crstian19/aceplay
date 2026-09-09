package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Defaults(t *testing.T) {
	m := New("", true)
	assert.Equal(t, defaultCommand, m.command)
	assert.Equal(t, defaultPort, m.Port())
	assert.False(t, m.WasStarted())

	m = New("/opt/acestream/start-engine", false)
	assert.Equal(t, "/opt/acestream/start-engine", m.command)
	assert.False(t, m.autoStart)
}

func TestManager_StopWithoutStart(t *testing.T) {
	// We never started the engine, so stopping it must be a no-op and never
	// kill an engine the user had running already.
	m := New("", true)
	require.NoError(t, m.Stop())
	assert.False(t, m.WasStarted())
}

func TestManager_StartWithoutAutoStart(t *testing.T) {
	// No engine listening on the common ports and auto-start disabled: the
	// manager must refuse instead of spawning anything.
	if findRunningPort(context.Background()) != 0 {
		t.Skip("an acestream-engine is running on this machine")
	}

	m := New("", false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auto-start is disabled")
}

func TestManager_MissingBinary(t *testing.T) {
	if findRunningPort(context.Background()) != 0 {
		t.Skip("an acestream-engine is running on this machine")
	}

	m := New("acestreamengine-does-not-exist", true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in PATH")
}

func TestWaitUntilReady_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitUntilReady(ctx, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
