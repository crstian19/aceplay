// Package acestream provides an HTTP client for communicating with acestream-engine
package acestream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultHost           = "localhost"
	defaultPort           = 6878
	defaultTimeout        = 30 * time.Second
	defaultConnectTimeout = 5 * time.Second
	defaultEngineCmd      = "acestreamengine"
)

// StreamStatus represents the status of a stream
type StreamStatus string

const (
	StatusPrebuf  StreamStatus = "prebuf"
	StatusDL      StreamStatus = "dl"
	StatusCheck   StreamStatus = "check"
	StatusWait    StreamStatus = "wait"
	StatusIdle    StreamStatus = "idle"
	StatusLoading StreamStatus = "loading"
	StatusError   StreamStatus = "error"
)

// EngineManager handles acestream-engine process lifecycle
type EngineManager struct {
	command    string
	process    *os.Process
	autoStart  bool
	wasStarted bool
	httpPort   int
}

// NewEngineManager creates a new engine manager
func NewEngineManager(command string, autoStart bool) *EngineManager {
	if command == "" {
		command = defaultEngineCmd
	}
	return &EngineManager{
		command:   command,
		autoStart: autoStart,
		httpPort:  6878,
	}
}

// GetPort returns the HTTP port that the engine is running on
func (e *EngineManager) GetPort() int {
	return e.httpPort
}

// Start starts the acestream-engine if not already running
func (e *EngineManager) Start() error {
	if e.process != nil {
		return nil // Already running
	}

	// Try to find an already running engine on any common port
	if runningPort := findRunningEnginePort(); runningPort != 0 {
		fmt.Printf("Found acestream-engine running on port %d\n", runningPort)
		e.httpPort = runningPort
		return nil // Engine already running on another port
	}

	if !e.autoStart {
		return fmt.Errorf("acestream-engine is not running and auto-start is disabled")
	}

	// Start the engine with fixed HTTP port
	cmd := exec.Command(e.command, "--client-console", "--http-port", "6878")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start acestream-engine: %w", err)
	}

	e.process = cmd.Process
	e.wasStarted = true
	e.httpPort = 6878

	// Wait for engine to be ready (retry for up to 30 seconds)
	fmt.Println("Waiting for acestream-engine to be ready...")
	client := NewClient(WithPort(6878))
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if client.IsRunning(ctx) {
			cancel()
			return nil
		}
		cancel()
	}

	return fmt.Errorf("acestream-engine failed to start within 30 seconds")
}

// Stop stops the acestream-engine if we started it
func (e *EngineManager) Stop() error {
	if !e.wasStarted || e.process == nil {
		return nil // We didn't start it, don't stop it
	}

	if err := e.process.Kill(); err != nil {
		return fmt.Errorf("failed to stop acestream-engine: %w", err)
	}

	e.process = nil
	e.wasStarted = false
	return nil
}

// WasStarted returns true if we started the engine
func (e *EngineManager) WasStarted() bool {
	return e.wasStarted
}

// findRunningEnginePort tries to find an already running acestream-engine on common ports
func findRunningEnginePort() int {
	ports := []int{6878, 45615, 8080, 9999}

	for _, port := range ports {
		client := NewClient(WithPort(port))
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if client.IsRunning(ctx) {
			return port
		}
	}

	return 0
}

// Client is the HTTP client for communicating with acestream-engine
type Client struct {
	host           string
	port           int
	httpClient     *resty.Client
	timeout        time.Duration
	connectTimeout time.Duration
	engineManager  *EngineManager
}

// ClientOption is a function that configures the client
type ClientOption func(*Client)

// WithHost configures the engine host
func WithHost(host string) ClientOption {
	return func(c *Client) {
		c.host = host
	}
}

// WithPort configures the engine port
func WithPort(port int) ClientOption {
	return func(c *Client) {
		c.port = port
	}
}

// WithTimeout configures the general timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithConnectTimeout configures the connection timeout
func WithConnectTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.connectTimeout = timeout
	}
}

// WithAutoStart enables automatic engine start
func WithAutoStart(command string) ClientOption {
	return func(c *Client) {
		c.engineManager = NewEngineManager(command, true)
	}
}

// NewClient creates a new client for acestream-engine
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		host:           defaultHost,
		port:           defaultPort,
		timeout:        defaultTimeout,
		connectTimeout: defaultConnectTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.httpClient = resty.New().
		SetTimeout(c.timeout).
		SetBaseURL(c.baseURL()).
		SetRedirectPolicy(resty.NoRedirectPolicy())

	return c
}

// GetEnginePort returns the port used by the engine manager
func (c *Client) GetEnginePort() int {
	if c.engineManager != nil {
		return c.engineManager.GetPort()
	}
	return 0
}

// SetPort updates the HTTP port for the client
func (c *Client) SetPort(port int) {
	c.port = port
	c.httpClient.SetBaseURL(c.baseURL())
}

// StartEngine starts the acestream-engine if auto-start is enabled
func (c *Client) StartEngine() error {
	if c.engineManager != nil {
		return c.engineManager.Start()
	}
	return nil
}

// StopEngine stops the acestream-engine if we started it
func (c *Client) StopEngine() error {
	if c.engineManager != nil {
		return c.engineManager.Stop()
	}
	return nil
}

// baseURL returns the engine base URL
func (c *Client) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// IsRunning checks if acestream-engine is running
func (c *Client) IsRunning(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, c.connectTimeout)
	defer cancel()

	// Try multiple endpoints to check if engine is ready
	endpoints := []string{
		"/webui/app/127323294/template/api",
		"/",
		"/ace/getstream?content_id=test",
	}

	for _, endpoint := range endpoints {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			Get(endpoint)

		// Consider it running if we get any response (OK, NotFound, or even server errors mean it's running)
		if err == nil && (resp.StatusCode() < 500 || resp.StatusCode() == 500) {
			return true
		}
	}

	return false
}

// GetStreamURL gets the stream URL for a content ID
// If hls is true, returns HLS URL (m3u8), otherwise returns direct HTTP stream URL
func (c *Client) GetStreamURL(contentID string, hls bool) (string, error) {
	var streamURL string

	if hls {
		// HLS stream using manifest endpoint
		streamURL = fmt.Sprintf("%s/ace/manifest.m3u8?id=%s", c.baseURL(), contentID)
	} else {
		// Direct HTTP stream
		streamURL = fmt.Sprintf("%s/ace/getstream?id=%s", c.baseURL(), contentID)
	}

	return streamURL, nil
}

// StreamStats represents stream statistics
type StreamStats struct {
	Status        StreamStatus `json:"status"`
	Progress      float64      `json:"progress"`
	DownloadSpeed int64        `json:"download_speed"`
	UploadSpeed   int64        `json:"upload_speed"`
	Peers         int          `json:"peers"`
}

// GetStats gets the statistics for a stream
func (c *Client) GetStats(contentID string) (*StreamStats, error) {
	endpoint := "/ace/getstream"

	resp, err := c.httpClient.R().
		SetQueryParam("content_id", contentID).
		SetQueryParam("method", "get_stats").
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("error getting statistics: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("engine error: status %d", resp.StatusCode())
	}

	body := string(resp.Body())

	var stats StreamStats
	if err := json.Unmarshal(resp.Body(), &stats); err != nil {
		// Try manual parsing if JSON parsing fails
		return c.parseStatsManually(body)
	}

	return &stats, nil
}

// parseStatsManually tries to parse statistics when format is not standard JSON
func (c *Client) parseStatsManually(body string) (*StreamStats, error) {
	// AceStream may return data in different formats
	// Try parsing as query string
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, fmt.Errorf("error parsing statistics: %w", err)
	}

	stats := &StreamStats{
		Status: StreamStatus(values.Get("status")),
	}

	if progress := values.Get("progress"); progress != "" {
		stats.Progress, _ = strconv.ParseFloat(progress, 64)
	}

	if speed := values.Get("download_speed"); speed != "" {
		stats.DownloadSpeed, _ = strconv.ParseInt(speed, 10, 64)
	}

	if peers := values.Get("peers"); peers != "" {
		stats.Peers, _ = strconv.Atoi(peers)
	}

	return stats, nil
}

// WaitForStream waits until the stream is ready or timeout is reached
// Returns the stream URL that can be used for playback
func (c *Client) WaitForStream(ctx context.Context, contentID string) (string, error) {
	// Get the base URL from the resty client
	baseURL := c.httpClient.BaseURL
	if baseURL == "" {
		baseURL = c.baseURL()
	}

	// Use HLS manifest endpoint - it returns a redirect to the actual m3u8 URL
	url := fmt.Sprintf("%s/ace/manifest.m3u8?content_id=%s", baseURL, contentID)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Check for redirect (status 302) - stream is ready
	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location != "" {
			// If the location is a relative path, construct the full URL
			if !strings.HasPrefix(location, "http") {
				location = baseURL + location
			}
			return location, nil
		}
	}

	// If we get 200, the manifest was directly returned
	if resp.StatusCode == http.StatusOK {
		return url, nil
	}

	return "", fmt.Errorf("unexpected response: status %d", resp.StatusCode)
}

// StopStream stops an active stream
func (c *Client) StopStream(contentID string) error {
	resp, err := c.httpClient.R().
		SetQueryParam("id", contentID).
		SetQueryParam("method", "stop").
		Get("/ace/getstream")

	if err != nil {
		return fmt.Errorf("error stopping stream: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("engine error: status %d", resp.StatusCode())
	}

	return nil
}

// GetEngineInfo gets information from acestream-engine
func (c *Client) GetEngineInfo(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, c.connectTimeout)
	defer cancel()

	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get("/webui/app/127323294/template/api")

	if err != nil {
		return nil, fmt.Errorf("error getting engine information: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("engine error: status %d", resp.StatusCode())
	}

	var info map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &info); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return info, nil
}
