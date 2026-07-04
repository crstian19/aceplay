package acestream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		opts []ClientOption
		want struct {
			host string
			port int
		}
	}{
		{
			name: "default",
			opts: nil,
			want: struct {
				host string
				port int
			}{
				host: "localhost",
				port: 6878,
			},
		},
		{
			name: "custom host",
			opts: []ClientOption{WithHost("192.168.1.1")},
			want: struct {
				host string
				port int
			}{
				host: "192.168.1.1",
				port: 6878,
			},
		},
		{
			name: "custom port",
			opts: []ClientOption{WithPort(8080)},
			want: struct {
				host string
				port int
			}{
				host: "localhost",
				port: 8080,
			},
		},
		{
			name: "custom host and port",
			opts: []ClientOption{WithHost("10.0.0.1"), WithPort(9000)},
			want: struct {
				host string
				port int
			}{
				host: "10.0.0.1",
				port: 9000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.opts...)
			assert.Equal(t, tt.want.host, client.host)
			assert.Equal(t, tt.want.port, client.port)
		})
	}
}

func TestClient_baseURL(t *testing.T) {
	client := NewClient(WithHost("localhost"), WithPort(6878))
	assert.Equal(t, "http://localhost:6878", client.baseURL())

	client = NewClient(WithHost("192.168.1.1"), WithPort(8080))
	assert.Equal(t, "http://192.168.1.1:8080", client.baseURL())
}

func TestClient_IsRunning(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webui/app/127323294/template/api" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Extract host and port from mock server
	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	ctx := context.Background()
	assert.True(t, client.IsRunning(ctx))

	// Server that returns error - but it's still "running" if it responds
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	client.httpClient.SetBaseURL(errorServer.URL)
	// Status 500 still means the server is running, just returning an error
	assert.True(t, client.IsRunning(ctx))
}

func TestClient_GetStreamURL(t *testing.T) {
	contentID := "abcd1234abcd1234abcd1234abcd1234abcd1234"

	tests := []struct {
		name     string
		hls      bool
		response string
		wantErr  bool
	}{
		{
			name:     "HTTP stream",
			hls:      false,
			response: "ok",
			wantErr:  false,
		},
		{
			name:     "HLS stream",
			hls:      true,
			response: "ok",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/ace/getstream", r.URL.Path)
				assert.Equal(t, contentID, r.URL.Query().Get("id"))
				if tt.hls {
					assert.Equal(t, "hls", r.URL.Query().Get("format"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient()
			client.httpClient.SetBaseURL(server.URL)

			url, err := client.GetStreamURL(contentID, tt.hls)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, url, contentID)
			if tt.hls {
				assert.Contains(t, url, "manifest.m3u8")
			}
		})
	}
}

func TestClient_GetStats(t *testing.T) {
	contentID := "abcd1234abcd1234abcd1234abcd1234abcd1234"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ace/getstream", r.URL.Path)
		assert.Equal(t, contentID, r.URL.Query().Get("content_id"))
		assert.Equal(t, "get_stats", r.URL.Query().Get("method"))

		stats := StreamStats{
			Status:        StatusPrebuf,
			Progress:      50.0,
			DownloadSpeed: 1024000,
			Peers:         25,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	stats, err := client.GetStats(contentID)
	require.NoError(t, err)
	assert.Equal(t, StatusPrebuf, stats.Status)
	assert.Equal(t, 50.0, stats.Progress)
	assert.Equal(t, int64(1024000), stats.DownloadSpeed)
	assert.Equal(t, 25, stats.Peers)
}

func TestClient_GetStats_QueryString(t *testing.T) {
	contentID := "abcd1234abcd1234abcd1234abcd1234abcd1234"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate query string response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("status=dl&progress=75.5&download_speed=2048000&peers=50"))
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	stats, err := client.GetStats(contentID)
	require.NoError(t, err)
	assert.Equal(t, StatusDL, stats.Status)
	assert.Equal(t, 75.5, stats.Progress)
	assert.Equal(t, int64(2048000), stats.DownloadSpeed)
	assert.Equal(t, 50, stats.Peers)
}

func TestClient_StopStream(t *testing.T) {
	contentID := "abcd1234abcd1234abcd1234abcd1234abcd1234"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ace/getstream", r.URL.Path)
		assert.Equal(t, contentID, r.URL.Query().Get("id"))
		assert.Equal(t, "stop", r.URL.Query().Get("method"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	err := client.StopStream(contentID)
	require.NoError(t, err)
}

func TestClient_WaitForStream(t *testing.T) {
	contentID := "abcd1234"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Must check the request path
		assert.Equal(t, "/ace/manifest.m3u8", r.URL.Path)
		assert.Equal(t, contentID, r.URL.Query().Get("content_id"))
		w.Header().Set("Location", "http://localhost:6878/ace/m/session/stream.m3u8")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url, err := client.WaitForStream(ctx, contentID)
	require.NoError(t, err)
	assert.Contains(t, url, "/ace/m/")
}

func TestClient_WaitForStream_Timeout(t *testing.T) {
	contentID := "abcd1234"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := client.WaitForStream(ctx, contentID)
	assert.Error(t, err)
}

func TestClient_GetEngineInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webui/app/127323294/template/api", r.URL.Path)

		info := map[string]interface{}{
			"version":  "3.1.74",
			"platform": "linux",
		}
		_ = json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient.SetBaseURL(server.URL)

	ctx := context.Background()
	info, err := client.GetEngineInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "3.1.74", info["version"])
	assert.Equal(t, "linux", info["platform"])
}
