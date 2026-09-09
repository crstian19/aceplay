package acestream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultHost           = "localhost"
	defaultPort           = 6878
	defaultTimeout        = 30 * time.Second
	defaultConnectTimeout = 5 * time.Second
)

// Client is the HTTP client for communicating with acestream-engine.
//
// It only speaks HTTP to an engine that is already listening: starting or
// stopping the engine process is out of scope, so a Client is safe to use
// against a remote engine.
type Client struct {
	host            string
	port            int
	baseURLOverride string
	timeout         time.Duration
	connectTimeout  time.Duration
	httpClient      *http.Client
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

// WithBaseURL overrides host and port with a full base URL (no trailing slash),
// which is handy for tests and for engines behind a reverse proxy.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURLOverride = strings.TrimSuffix(baseURL, "/")
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

// WithHTTPClient injects a custom *http.Client. The redirect policy is
// overwritten: the engine answers with a 302 that the caller needs to read.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
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

	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: c.timeout}
	}

	// The engine signals "stream ready" with a 302 to the real playlist URL,
	// so redirects must never be followed automatically.
	c.httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return c
}

// SetPort updates the HTTP port for the client
func (c *Client) SetPort(port int) {
	c.port = port
	c.baseURLOverride = ""
}

// baseURL returns the engine base URL
func (c *Client) baseURL() string {
	if c.baseURLOverride != "" {
		return c.baseURLOverride
	}
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// get issues a GET against the engine with the given path and query values.
// The caller owns the response body.
func (c *Client) get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	endpoint := c.baseURL() + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return resp, nil
}

// getBody issues a GET and returns the body, failing on any non-200 status.
func (c *Client) getBody(ctx context.Context, path string, query url.Values) ([]byte, error) {
	resp, err := c.get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("engine error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return body, nil
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
		resp, err := c.get(ctx, endpoint, nil)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()

		// Any answer at all means something is listening — even a 500.
		return true
	}

	return false
}

// GetStreamURL gets the stream URL for a content ID
// If hls is true, returns HLS URL (m3u8), otherwise returns direct HTTP stream URL
func (c *Client) GetStreamURL(contentID string, hls bool) (string, error) {
	if hls {
		// HLS stream using manifest endpoint
		return fmt.Sprintf("%s/ace/manifest.m3u8?id=%s", c.baseURL(), contentID), nil
	}

	// Direct HTTP stream
	return fmt.Sprintf("%s/ace/getstream?id=%s", c.baseURL(), contentID), nil
}

// WaitForStream waits until the stream is ready or timeout is reached
// Returns the stream URL that can be used for playback
func (c *Client) WaitForStream(ctx context.Context, contentID string) (string, error) {
	// The HLS manifest endpoint redirects to the actual m3u8 URL once ready
	resp, err := c.get(ctx, "/ace/manifest.m3u8", url.Values{"content_id": {contentID}})
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for redirect (status 302) - stream is ready
	if resp.StatusCode == http.StatusFound {
		if location := resp.Header.Get("Location"); location != "" {
			// If the location is a relative path, construct the full URL
			if !strings.HasPrefix(location, "http") {
				location = c.baseURL() + location
			}
			return location, nil
		}
	}

	// If we get 200, the manifest was directly returned
	if resp.StatusCode == http.StatusOK {
		return fmt.Sprintf("%s/ace/manifest.m3u8?content_id=%s", c.baseURL(), contentID), nil
	}

	return "", fmt.Errorf("unexpected response: status %d", resp.StatusCode)
}

// StopStream stops an active stream
func (c *Client) StopStream(ctx context.Context, contentID string) error {
	_, err := c.getBody(ctx, "/ace/getstream", url.Values{
		"id":     {contentID},
		"method": {"stop"},
	})
	if err != nil {
		return fmt.Errorf("error stopping stream: %w", err)
	}

	return nil
}

// GetEngineInfo gets information from acestream-engine
func (c *Client) GetEngineInfo(ctx context.Context) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, c.connectTimeout)
	defer cancel()

	body, err := c.getBody(ctx, "/webui/app/127323294/template/api", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting engine information: %w", err)
	}

	var info map[string]any
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return info, nil
}
